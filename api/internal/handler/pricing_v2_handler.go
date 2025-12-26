package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"rev-saas-api/internal/middleware"
	"rev-saas-api/internal/model"
	"rev-saas-api/internal/service"
)

// PricingV2Handler handles pricing v2 API requests
type PricingV2Handler struct {
	service *service.PricingV2Service
}

// NewPricingV2Handler creates a new PricingV2Handler
func NewPricingV2Handler(svc *service.PricingV2Service) *PricingV2Handler {
	return &PricingV2Handler{service: svc}
}

// writeJSONPV2 writes a JSON response
func writeJSONPV2(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Discover handles POST /api/pricing-v2/discover
func (h *PricingV2Handler) Discover(w http.ResponseWriter, r *http.Request) {
	var req model.PricingDiscoverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONPV2(w, model.PricingDiscoverResponse{
			Error: "invalid request body",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.service.DiscoverPricingPage(r.Context(), req.WebsiteURL)
	if err != nil {
		log.Printf("[pricing-v2-handler] discover error: %v", err)
		writeJSONPV2(w, model.PricingDiscoverResponse{
			Error: err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	status := http.StatusOK
	if result.Error != "" {
		status = http.StatusBadRequest
	}
	writeJSONPV2(w, result, status)
}

// Extract handles POST /api/pricing-v2/extract
func (h *PricingV2Handler) Extract(w http.ResponseWriter, r *http.Request) {
	var req model.PricingExtractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONPV2(w, model.PricingExtractResponse{
			Error: "invalid request body",
		}, http.StatusBadRequest)
		return
	}

	if req.PricingURL == "" {
		writeJSONPV2(w, model.PricingExtractResponse{
			Error: "pricing_url is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.service.ExtractPricing(r.Context(), req.PricingURL)
	if err != nil {
		log.Printf("[pricing-v2-handler] extract error: %v", err)
		writeJSONPV2(w, model.PricingExtractResponse{
			Error: err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	status := http.StatusOK
	if result.Error != "" {
		status = http.StatusBadRequest
	}
	writeJSONPV2(w, result, status)
}

// Save handles POST /api/pricing-v2/save
func (h *PricingV2Handler) Save(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONPV2(w, model.PricingSaveResponse{
			Error: "unauthorized",
		}, http.StatusUnauthorized)
		return
	}

	var req model.PricingSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONPV2(w, model.PricingSaveResponse{
			Error: "invalid request body",
		}, http.StatusBadRequest)
		return
	}

	if len(req.Plans) == 0 {
		writeJSONPV2(w, model.PricingSaveResponse{
			Error: "at least one plan is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.service.SavePlans(r.Context(), user.ID.Hex(), req)
	if err != nil {
		log.Printf("[pricing-v2-handler] save error: %v", err)
		writeJSONPV2(w, model.PricingSaveResponse{
			Error: err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	status := http.StatusOK
	if result.Error != "" {
		status = http.StatusBadRequest
	}
	writeJSONPV2(w, result, status)
}

// List handles GET /api/pricing-v2
func (h *PricingV2Handler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONPV2(w, map[string]string{"error": "unauthorized"}, http.StatusUnauthorized)
		return
	}

	result, err := h.service.GetSavedPlans(r.Context(), user.ID.Hex())
	if err != nil {
		log.Printf("[pricing-v2-handler] list error: %v", err)
		writeJSONPV2(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	writeJSONPV2(w, result, http.StatusOK)
}

// Delete handles DELETE /api/pricing-v2/{id}
func (h *PricingV2Handler) Delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONPV2(w, map[string]string{"error": "unauthorized"}, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	planID := vars["id"]
	if planID == "" {
		writeJSONPV2(w, map[string]string{"error": "plan ID required"}, http.StatusBadRequest)
		return
	}

	err := h.service.DeletePlan(r.Context(), user.ID.Hex(), planID)
	if err != nil {
		log.Printf("[pricing-v2-handler] delete error: %v", err)
		writeJSONPV2(w, map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		return
	}

	writeJSONPV2(w, map[string]string{"status": "deleted"}, http.StatusOK)
}

// ExtractFromText handles POST /api/pricing-v2/extract-from-text (paste mode fallback)
func (h *PricingV2Handler) ExtractFromText(w http.ResponseWriter, r *http.Request) {
	var req model.PricingExtractFromTextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONPV2(w, model.PricingExtractFromTextResponse{
			Error: "invalid request body",
		}, http.StatusBadRequest)
		return
	}

	if req.MonthlyText == "" && req.YearlyText == "" {
		writeJSONPV2(w, model.PricingExtractFromTextResponse{
			Error: "at least one of monthly_text or yearly_text is required",
		}, http.StatusBadRequest)
		return
	}

	result, err := h.service.ExtractFromText(r.Context(), req)
	if err != nil {
		log.Printf("[pricing-v2-handler] extract-from-text error: %v", err)
		writeJSONPV2(w, model.PricingExtractFromTextResponse{
			Error: err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	status := http.StatusOK
	if result.Error != "" {
		status = http.StatusBadRequest
	}
	writeJSONPV2(w, result, status)
}
