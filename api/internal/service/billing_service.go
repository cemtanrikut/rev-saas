package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"rev-saas-api/internal/config"
	"rev-saas-api/internal/model"
	"rev-saas-api/internal/repository/mongo"
)

// Stripe API endpoints
const (
	stripeCustomersURL       = "https://api.stripe.com/v1/customers"
	stripeCheckoutSessionURL = "https://api.stripe.com/v1/checkout/sessions"
	stripePortalSessionURL   = "https://api.stripe.com/v1/billing_portal/sessions"
	stripeSubscriptionsURL   = "https://api.stripe.com/v1/subscriptions"
)

// Billing errors
var (
	ErrBillingNotConfigured   = errors.New("billing not configured")
	ErrInvalidPlanKey         = errors.New("invalid plan key")
	ErrNoStripeCustomer       = errors.New("no Stripe customer found for this user")
	ErrSubscriptionNotFound   = errors.New("subscription not found")
	ErrWebhookSignatureInvalid = errors.New("webhook signature verification failed")
)

// BillingService handles Stripe billing operations.
type BillingService struct {
	cfg           *config.Config
	billingRepo   *mongo.BillingSubscriptionRepository
	webhookRepo   *mongo.WebhookEventRepository
	userRepo      *mongo.UserRepository
	aiUsageRepo   *mongo.AIUsageRepository
}

// NewBillingService creates a new billing service.
func NewBillingService(
	cfg *config.Config,
	billingRepo *mongo.BillingSubscriptionRepository,
	webhookRepo *mongo.WebhookEventRepository,
	userRepo *mongo.UserRepository,
	aiUsageRepo *mongo.AIUsageRepository,
) *BillingService {
	return &BillingService{
		cfg:         cfg,
		billingRepo: billingRepo,
		webhookRepo: webhookRepo,
		userRepo:    userRepo,
		aiUsageRepo: aiUsageRepo,
	}
}

// IsConfigured returns true if billing is properly configured.
func (s *BillingService) IsConfigured() bool {
	return s.cfg.IsBillingEnabled()
}

// GetBillingStatus returns the user's current billing status.
func (s *BillingService) GetBillingStatus(ctx context.Context, userID string) (*model.BillingStatusResponse, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}

	sub, err := s.billingRepo.GetByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}

	// No subscription = free tier
	if sub == nil {
		return &model.BillingStatusResponse{
			PlanKey: model.PlanKeyFree,
			Status:  "",
		}, nil
	}

	// Only return period end for active/trialing subscriptions
	// For canceled subscriptions, don't show renewal date
	var periodEnd *time.Time
	isActiveOrTrialing := sub.Status == model.SubscriptionStatusActive || sub.Status == model.SubscriptionStatusTrialing
	if !sub.CurrentPeriodEnd.IsZero() && isActiveOrTrialing {
		periodEnd = &sub.CurrentPeriodEnd
	}

	// For canceled subscriptions, return free plan
	planKey := sub.PlanKey
	if sub.Status == model.SubscriptionStatusCanceled {
		planKey = model.PlanKeyFree
	}

	return &model.BillingStatusResponse{
		PlanKey:              planKey,
		Status:               sub.Status,
		CancelAtPeriodEnd:    sub.CancelAtPeriodEnd,
		CurrentPeriodEnd:     periodEnd,
		StripeSubscriptionID: sub.StripeSubscriptionID,
		StripeCustomerID:     sub.StripeCustomerID,
	}, nil
}

// CreateCheckoutSession creates a Stripe Checkout session for subscription purchase.
// If user already has an active subscription, it updates the existing one instead.
func (s *BillingService) CreateCheckoutSession(ctx context.Context, userID, userEmail, planKey string) (string, error) {
	if !s.IsConfigured() {
		return "", ErrBillingNotConfigured
	}

	// Validate plan key and get price ID
	priceID := s.cfg.GetPriceIDForPlan(planKey)
	if priceID == "" {
		return "", ErrInvalidPlanKey
	}

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return "", fmt.Errorf("invalid user id: %w", err)
	}

	// Check if user already has an active subscription
	existingSub, err := s.billingRepo.GetByUserID(ctx, uid)
	if err != nil {
		return "", fmt.Errorf("failed to check existing subscription: %w", err)
	}

	// If user has an active subscription, update it instead of creating new one
	if existingSub != nil && existingSub.StripeSubscriptionID != "" && existingSub.Status == model.SubscriptionStatusActive {
		log.Printf("[billing] user %s has active subscription %s, updating instead of creating new",
			userID, existingSub.StripeSubscriptionID)
		return s.updateSubscription(ctx, existingSub.StripeSubscriptionID, priceID, userID)
	}

	// Get or create Stripe customer
	customerID, err := s.getOrCreateStripeCustomer(ctx, uid, userEmail)
	if err != nil {
		return "", fmt.Errorf("failed to get/create Stripe customer: %w", err)
	}

	// Create Checkout Session for new subscription
	data := url.Values{}
	data.Set("customer", customerID)
	data.Set("mode", "subscription")
	data.Set("line_items[0][price]", priceID)
	data.Set("line_items[0][quantity]", "1")
	data.Set("success_url", s.cfg.StripeBillingSuccessURL)
	data.Set("cancel_url", s.cfg.StripeBillingCancelURL)
	data.Set("allow_promotion_codes", "true")
	// Store user ID in metadata for webhook processing
	data.Set("subscription_data[metadata][user_id]", userID)
	data.Set("client_reference_id", userID)

	req, err := http.NewRequestWithContext(ctx, "POST", stripeCheckoutSessionURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(s.cfg.StripeBillingSecretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("stripe checkout session error: %s", string(body))
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.URL, nil
}

// updateSubscription updates an existing Stripe subscription to a new price.
// Returns the success URL directly since no checkout is needed.
func (s *BillingService) updateSubscription(ctx context.Context, subscriptionID, newPriceID, userID string) (string, error) {
	// First, get the current subscription to find the subscription item ID
	getReq, err := http.NewRequestWithContext(ctx, "GET", stripeSubscriptionsURL+"/"+subscriptionID, nil)
	if err != nil {
		return "", err
	}
	getReq.SetBasicAuth(s.cfg.StripeBillingSecretKey, "")

	getResp, err := http.DefaultClient.Do(getReq)
	if err != nil {
		return "", err
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(getResp.Body)
		return "", fmt.Errorf("failed to get subscription: %s", string(body))
	}

	var subscription struct {
		Items struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		} `json:"items"`
	}
	if err := json.NewDecoder(getResp.Body).Decode(&subscription); err != nil {
		return "", err
	}

	if len(subscription.Items.Data) == 0 {
		return "", fmt.Errorf("subscription has no items")
	}

	subscriptionItemID := subscription.Items.Data[0].ID

	// Update the subscription with the new price
	data := url.Values{}
	data.Set("items[0][id]", subscriptionItemID)
	data.Set("items[0][price]", newPriceID)
	data.Set("proration_behavior", "create_prorations") // Prorate the change
	data.Set("metadata[user_id]", userID)

	updateReq, err := http.NewRequestWithContext(ctx, "POST", stripeSubscriptionsURL+"/"+subscriptionID, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	updateReq.SetBasicAuth(s.cfg.StripeBillingSecretKey, "")
	updateReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	updateResp, err := http.DefaultClient.Do(updateReq)
	if err != nil {
		return "", err
	}
	defer updateResp.Body.Close()

	if updateResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(updateResp.Body)
		return "", fmt.Errorf("failed to update subscription: %s", string(body))
	}

	log.Printf("[billing] subscription %s updated to price %s for user %s", subscriptionID, newPriceID, userID)

	// Return success URL - the webhook will handle the database update
	return s.cfg.StripeBillingSuccessURL, nil
}

// CreatePortalSession creates a Stripe Customer Portal session.
func (s *BillingService) CreatePortalSession(ctx context.Context, userID string) (string, error) {
	if !s.IsConfigured() {
		return "", ErrBillingNotConfigured
	}

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return "", fmt.Errorf("invalid user id: %w", err)
	}

	// Get existing subscription to find customer ID
	sub, err := s.billingRepo.GetByUserID(ctx, uid)
	if err != nil {
		return "", err
	}

	if sub == nil || sub.StripeCustomerID == "" {
		return "", ErrNoStripeCustomer
	}

	// Create Portal Session
	data := url.Values{}
	data.Set("customer", sub.StripeCustomerID)
	data.Set("return_url", s.cfg.StripeCustomerPortalReturnURL)

	req, err := http.NewRequestWithContext(ctx, "POST", stripePortalSessionURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(s.cfg.StripeBillingSecretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("stripe portal session error: %s", string(body))
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.URL, nil
}

// getOrCreateStripeCustomer gets or creates a Stripe customer for the user.
func (s *BillingService) getOrCreateStripeCustomer(ctx context.Context, userID primitive.ObjectID, email string) (string, error) {
	// Check if user already has a subscription with customer ID
	sub, err := s.billingRepo.GetByUserID(ctx, userID)
	if err != nil {
		return "", err
	}

	if sub != nil && sub.StripeCustomerID != "" {
		return sub.StripeCustomerID, nil
	}

	// Create new Stripe customer
	data := url.Values{}
	data.Set("email", email)
	data.Set("metadata[user_id]", userID.Hex())

	req, err := http.NewRequestWithContext(ctx, "POST", stripeCustomersURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(s.cfg.StripeBillingSecretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("stripe customer creation error: %s", string(body))
	}

	var customer struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&customer); err != nil {
		return "", err
	}

	// Store the customer ID in a new subscription record (will be updated by webhook)
	newSub := &model.BillingSubscription{
		UserID:           userID,
		StripeCustomerID: customer.ID,
		PlanKey:          model.PlanKeyFree,
		Status:           model.SubscriptionStatusIncomplete,
	}
	if err := s.billingRepo.Upsert(ctx, newSub); err != nil {
		log.Printf("[billing] warning: failed to store customer ID: %v", err)
	}

	return customer.ID, nil
}

// HandleWebhookEvent processes a Stripe webhook event.
func (s *BillingService) HandleWebhookEvent(ctx context.Context, eventID, eventType string, eventData json.RawMessage) error {
	// Check idempotency
	processed, err := s.webhookRepo.IsProcessed(ctx, eventID)
	if err != nil {
		return fmt.Errorf("idempotency check failed: %w", err)
	}
	if processed {
		log.Printf("[billing] event %s already processed, skipping", eventID)
		return nil
	}

	// Process based on event type
	switch eventType {
	case "checkout.session.completed":
		err = s.handleCheckoutCompleted(ctx, eventData)
	case "customer.subscription.created":
		err = s.handleSubscriptionCreatedOrUpdated(ctx, eventData)
	case "customer.subscription.updated":
		err = s.handleSubscriptionCreatedOrUpdated(ctx, eventData)
	case "customer.subscription.deleted":
		err = s.handleSubscriptionDeleted(ctx, eventData)
	case "invoice.paid":
		err = s.handleInvoicePaid(ctx, eventData)
	case "invoice.payment_failed":
		err = s.handlePaymentFailed(ctx, eventData)
	default:
		log.Printf("[billing] unhandled event type: %s", eventType)
	}

	if err != nil {
		return err
	}

	// Mark event as processed
	if err := s.webhookRepo.MarkProcessed(ctx, eventID, eventType); err != nil {
		log.Printf("[billing] warning: failed to mark event processed: %v", err)
	}

	return nil
}

// handleCheckoutCompleted handles checkout.session.completed event.
func (s *BillingService) handleCheckoutCompleted(ctx context.Context, data json.RawMessage) error {
	var session struct {
		Customer          string `json:"customer"`
		Subscription      string `json:"subscription"`
		ClientReferenceID string `json:"client_reference_id"`
	}
	if err := json.Unmarshal(data, &session); err != nil {
		return err
	}

	log.Printf("[billing] checkout completed: customer=%s subscription=%s user=%s",
		session.Customer, session.Subscription, session.ClientReferenceID)

	// The subscription.created event will handle the actual subscription setup
	return nil
}

// handleSubscriptionCreatedOrUpdated handles subscription created/updated events.
func (s *BillingService) handleSubscriptionCreatedOrUpdated(ctx context.Context, data json.RawMessage) error {
	var subscription struct {
		ID                 string `json:"id"`
		Customer           string `json:"customer"`
		Status             string `json:"status"`
		CancelAtPeriodEnd  bool   `json:"cancel_at_period_end"`
		CancelAt           int64  `json:"cancel_at"` // Stripe may use this instead of cancel_at_period_end
		CurrentPeriodEnd   int64  `json:"current_period_end"`
		Items              struct {
			Data []struct {
				Price struct {
					ID string `json:"id"`
				} `json:"price"`
				CurrentPeriodEnd int64 `json:"current_period_end"` // Also check items level
			} `json:"data"`
		} `json:"items"`
		Metadata map[string]string `json:"metadata"`
	}
	if err := json.Unmarshal(data, &subscription); err != nil {
		return err
	}

	// Fallback: if root level current_period_end is 0, try to get from items
	if subscription.CurrentPeriodEnd == 0 && len(subscription.Items.Data) > 0 {
		subscription.CurrentPeriodEnd = subscription.Items.Data[0].CurrentPeriodEnd
		log.Printf("[billing] using current_period_end from items: %d", subscription.CurrentPeriodEnd)
	}

	log.Printf("[billing] subscription event: id=%s customer=%s status=%s period_end=%d cancel_at_period_end=%v cancel_at=%d",
		subscription.ID, subscription.Customer, subscription.Status, subscription.CurrentPeriodEnd, 
		subscription.CancelAtPeriodEnd, subscription.CancelAt)

	// Get price ID from first item
	var priceID string
	if len(subscription.Items.Data) > 0 {
		priceID = subscription.Items.Data[0].Price.ID
	}

	// Determine plan key from price ID
	planKey := model.PlanKey(s.cfg.GetPlanKeyForPriceID(priceID))

	// Try to find user ID from metadata or existing subscription
	var userID primitive.ObjectID
	var existingSub *model.BillingSubscription
	if uid, ok := subscription.Metadata["user_id"]; ok {
		userID, _ = primitive.ObjectIDFromHex(uid)
	}

	// If no user ID in metadata, try to find by customer ID
	if userID.IsZero() {
		existingSub, _ = s.billingRepo.GetByStripeCustomerID(ctx, subscription.Customer)
		if existingSub != nil {
			userID = existingSub.UserID
		}
	} else {
		// Get existing subscription to check for plan change
		existingSub, _ = s.billingRepo.GetByUserID(ctx, userID)
	}

	if userID.IsZero() {
		log.Printf("[billing] warning: could not determine user_id for subscription %s", subscription.ID)
		return nil // Don't fail, just log
	}

	// Check if plan changed - if so, reset credits
	oldPlanKey := model.PlanKeyFree
	if existingSub != nil {
		oldPlanKey = existingSub.PlanKey
	}
	planChanged := oldPlanKey != planKey

	// Parse current_period_end
	var periodEnd time.Time
	if subscription.CurrentPeriodEnd > 0 {
		periodEnd = time.Unix(subscription.CurrentPeriodEnd, 0)
	}

	// Determine if subscription is scheduled for cancellation
	// Stripe may use either cancel_at_period_end OR cancel_at
	cancelScheduled := subscription.CancelAtPeriodEnd || subscription.CancelAt > 0

	// Upsert subscription
	sub := &model.BillingSubscription{
		UserID:               userID,
		StripeCustomerID:     subscription.Customer,
		StripeSubscriptionID: subscription.ID,
		StripePriceID:        priceID,
		PlanKey:              planKey,
		Status:               model.SubscriptionStatus(subscription.Status),
		CancelAtPeriodEnd:    cancelScheduled,
		CurrentPeriodEnd:     periodEnd,
	}

	if err := s.billingRepo.Upsert(ctx, sub); err != nil {
		return fmt.Errorf("failed to upsert subscription: %w", err)
	}

	// Update user's plan field
	if err := s.userRepo.UpdatePlan(ctx, userID, string(planKey)); err != nil {
		log.Printf("[billing] warning: failed to update user plan: %v", err)
	}

	// Reset credits when plan changes (upgrade or downgrade)
	if planChanged {
		monthKey := time.Now().Format("2006-01")
		if err := s.aiUsageRepo.ResetCredits(ctx, userID.Hex(), monthKey); err != nil {
			log.Printf("[billing] warning: failed to reset credits on plan change: %v", err)
		} else {
			log.Printf("[billing] credits reset on plan change: user=%s old_plan=%s new_plan=%s",
				userID.Hex(), oldPlanKey, planKey)
		}
	}

	log.Printf("[billing] subscription upserted: user=%s plan=%s status=%s period_end=%s",
		userID.Hex(), planKey, subscription.Status, periodEnd.Format(time.RFC3339))

	return nil
}

// handleSubscriptionDeleted handles subscription.deleted event.
func (s *BillingService) handleSubscriptionDeleted(ctx context.Context, data json.RawMessage) error {
	var subscription struct {
		ID       string `json:"id"`
		Customer string `json:"customer"`
	}
	if err := json.Unmarshal(data, &subscription); err != nil {
		return err
	}

	// Find subscription by customer ID
	existingSub, err := s.billingRepo.GetByStripeCustomerID(ctx, subscription.Customer)
	if err != nil {
		return err
	}

	if existingSub == nil {
		log.Printf("[billing] subscription deleted but not found: %s", subscription.ID)
		return nil
	}

	// Update to canceled/free
	existingSub.Status = model.SubscriptionStatusCanceled
	existingSub.PlanKey = model.PlanKeyFree
	existingSub.StripeSubscriptionID = ""

	if err := s.billingRepo.Upsert(ctx, existingSub); err != nil {
		return fmt.Errorf("failed to update canceled subscription: %w", err)
	}

	// Update user's plan to free
	if err := s.userRepo.UpdatePlan(ctx, existingSub.UserID, string(model.PlanKeyFree)); err != nil {
		log.Printf("[billing] warning: failed to update user plan to free: %v", err)
	}

	log.Printf("[billing] subscription canceled: user=%s", existingSub.UserID.Hex())

	return nil
}

// handleInvoicePaid handles invoice.paid event - resets credits.
func (s *BillingService) handleInvoicePaid(ctx context.Context, data json.RawMessage) error {
	var invoice struct {
		ID           string `json:"id"`
		Customer     string `json:"customer"`
		Subscription string `json:"subscription"`
		BillingReason string `json:"billing_reason"` // "subscription_cycle", "subscription_create", etc.
	}
	if err := json.Unmarshal(data, &invoice); err != nil {
		return err
	}

	// Only reset credits for subscription cycle payments
	if invoice.BillingReason != "subscription_cycle" && invoice.BillingReason != "subscription_create" {
		return nil
	}

	// Find user by customer ID
	sub, err := s.billingRepo.GetByStripeCustomerID(ctx, invoice.Customer)
	if err != nil || sub == nil {
		log.Printf("[billing] invoice paid but subscription not found: customer=%s", invoice.Customer)
		return nil
	}

	// Reset AI credits for this user
	monthKey := time.Now().Format("2006-01")
	if err := s.aiUsageRepo.ResetCredits(ctx, sub.UserID.Hex(), monthKey); err != nil {
		log.Printf("[billing] warning: failed to reset credits for user %s: %v", sub.UserID.Hex(), err)
	} else {
		log.Printf("[billing] credits reset for user %s (invoice %s)", sub.UserID.Hex(), invoice.ID)
	}

	return nil
}

// handlePaymentFailed handles invoice.payment_failed event.
func (s *BillingService) handlePaymentFailed(ctx context.Context, data json.RawMessage) error {
	var invoice struct {
		ID           string `json:"id"`
		Customer     string `json:"customer"`
		Subscription string `json:"subscription"`
	}
	if err := json.Unmarshal(data, &invoice); err != nil {
		return err
	}

	// Find subscription by customer ID
	sub, err := s.billingRepo.GetByStripeCustomerID(ctx, invoice.Customer)
	if err != nil || sub == nil {
		log.Printf("[billing] payment failed but subscription not found: customer=%s", invoice.Customer)
		return nil
	}

	// Update status to past_due
	sub.Status = model.SubscriptionStatusPastDue
	if err := s.billingRepo.Upsert(ctx, sub); err != nil {
		return fmt.Errorf("failed to update past_due status: %w", err)
	}

	log.Printf("[billing] payment failed, marked past_due: user=%s", sub.UserID.Hex())

	return nil
}

