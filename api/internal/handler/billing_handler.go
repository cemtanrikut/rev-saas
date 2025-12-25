package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"rev-saas-api/internal/config"
	"rev-saas-api/internal/middleware"
	"rev-saas-api/internal/model"
	"rev-saas-api/internal/service"
)

// BillingHandler handles billing-related HTTP requests.
type BillingHandler struct {
	billingService *service.BillingService
	cfg            *config.Config
}

// NewBillingHandler creates a new billing handler.
func NewBillingHandler(billingService *service.BillingService, cfg *config.Config) *BillingHandler {
	return &BillingHandler{
		billingService: billingService,
		cfg:            cfg,
	}
}

// GetStatus returns the user's current billing status.
// GET /api/billing/status
func (h *BillingHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	status, err := h.billingService.GetBillingStatus(r.Context(), userID)
	if err != nil {
		log.Printf("[billing] error getting status: %v", err)
		writeJSONError(w, "failed to get billing status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// CreateCheckoutSession creates a Stripe Checkout session.
// POST /api/billing/checkout-session
func (h *BillingHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	userEmail := middleware.UserEmailFromContext(r.Context())
	if userID == "" {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.billingService.IsConfigured() {
		writeJSONError(w, "billing not configured", http.StatusServiceUnavailable)
		return
	}

	var req model.CheckoutSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate plan key
	validPlans := map[string]bool{"starter": true, "growth": true, "enterprise": true}
	if !validPlans[req.PlanKey] {
		writeJSONError(w, "invalid plan_key", http.StatusBadRequest)
		return
	}

	url, err := h.billingService.CreateCheckoutSession(r.Context(), userID, userEmail, req.PlanKey)
	if err != nil {
		if err == service.ErrInvalidPlanKey {
			writeJSONError(w, "plan not available", http.StatusBadRequest)
			return
		}
		log.Printf("[billing] error creating checkout session: %v", err)
		writeJSONError(w, "failed to create checkout session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.CheckoutSessionResponse{URL: url})
}

// CreatePortalSession creates a Stripe Customer Portal session.
// POST /api/billing/portal
func (h *BillingHandler) CreatePortalSession(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.billingService.IsConfigured() {
		writeJSONError(w, "billing not configured", http.StatusServiceUnavailable)
		return
	}

	url, err := h.billingService.CreatePortalSession(r.Context(), userID)
	if err != nil {
		if err == service.ErrNoStripeCustomer {
			writeJSONError(w, "no billing account found. Please subscribe to a plan first.", http.StatusBadRequest)
			return
		}
		log.Printf("[billing] error creating portal session: %v", err)
		writeJSONError(w, "failed to create portal session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.PortalSessionResponse{URL: url})
}

// HandleWebhook handles Stripe webhook events.
// POST /api/billing/webhook
// This endpoint does NOT use auth middleware - it uses Stripe signature verification.
func (h *BillingHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if !h.billingService.IsConfigured() {
		http.Error(w, "billing not configured", http.StatusServiceUnavailable)
		return
	}

	// Read raw body for signature verification
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	// Verify Stripe signature
	sigHeader := r.Header.Get("Stripe-Signature")
	if sigHeader == "" {
		http.Error(w, "missing stripe signature", http.StatusBadRequest)
		return
	}

	if !h.verifyStripeSignature(body, sigHeader) {
		log.Printf("[billing] webhook signature verification failed")
		http.Error(w, "signature verification failed", http.StatusBadRequest)
		return
	}

	// Parse event
	var event struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Livemode bool   `json:"livemode"`
		Data     struct {
			Object json.RawMessage `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("[billing] failed to parse webhook event: %v", err)
		http.Error(w, "failed to parse event", http.StatusBadRequest)
		return
	}

	log.Printf("[billing] received webhook: type=%s id=%s livemode=%v", event.Type, event.ID, event.Livemode)

	// CRITICAL: Validate livemode matches environment
	// Production MUST only process live events, staging/local MUST only process test events
	expectedLivemode := h.cfg.Environment.IsProduction()
	if event.Livemode != expectedLivemode {
		log.Printf("[billing] REJECTED webhook: livemode mismatch - event.livemode=%v, expected=%v (env=%s)",
			event.Livemode, expectedLivemode, h.cfg.Environment)
		http.Error(w, "livemode mismatch", http.StatusBadRequest)
		return
	}

	// Process event
	if err := h.billingService.HandleWebhookEvent(r.Context(), event.ID, event.Type, event.Data.Object); err != nil {
		log.Printf("[billing] error processing webhook: %v", err)
		// Return 200 anyway to prevent Stripe from retrying
		// (we've already logged the error and can investigate)
	}

	// Always return 200 quickly to acknowledge receipt
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"received": true}`))
}

// verifyStripeSignature verifies the Stripe webhook signature.
// See: https://stripe.com/docs/webhooks/signatures
func (h *BillingHandler) verifyStripeSignature(payload []byte, sigHeader string) bool {
	// Parse signature header
	parts := strings.Split(sigHeader, ",")
	var timestamp string
	var signatures []string

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			timestamp = kv[1]
		case "v1":
			signatures = append(signatures, kv[1])
		}
	}

	if timestamp == "" || len(signatures) == 0 {
		return false
	}

	// Check timestamp is within tolerance (5 minutes)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	if time.Now().Unix()-ts > 300 {
		log.Printf("[billing] webhook timestamp too old: %d", ts)
		return false
	}

	// Compute expected signature
	signedPayload := fmt.Sprintf("%s.%s", timestamp, string(payload))
	mac := hmac.New(sha256.New, []byte(h.cfg.StripeWebhookSecret))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	// Compare with provided signatures
	for _, sig := range signatures {
		if hmac.Equal([]byte(expectedSig), []byte(sig)) {
			return true
		}
	}

	return false
}



