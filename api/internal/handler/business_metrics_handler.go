package handler

import (
	"encoding/json"
	"net/http"

	"rev-saas-api/internal/middleware"
	"rev-saas-api/internal/service"
)

// BusinessMetricsHandler handles HTTP requests for business metrics.
type BusinessMetricsHandler struct {
	service *service.BusinessMetricsService
}

// NewBusinessMetricsHandler creates a new BusinessMetricsHandler.
func NewBusinessMetricsHandler(service *service.BusinessMetricsService) *BusinessMetricsHandler {
	return &BusinessMetricsHandler{
		service: service,
	}
}

type setMetricsRequest struct {
	Currency         string  `json:"currency"`
	MRR              float64 `json:"mrr"`
	Customers        int     `json:"customers"`
	MonthlyChurnRate float64 `json:"monthly_churn_rate"`
}

// Get handles GET /api/business-metrics - retrieves the current user's business metrics.
func (h *BusinessMetricsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	metrics, err := h.service.GetMetrics(r.Context(), userID)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Return null if no metrics found (consistent with "not yet set" state)
	if metrics == nil {
		w.Write([]byte("null"))
		return
	}

	json.NewEncoder(w).Encode(metrics)
}

// Set handles PUT /api/business-metrics - creates or updates the current user's business metrics.
func (h *BusinessMetricsHandler) Set(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req setMetricsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	input := service.MetricsInput{
		Currency:         req.Currency,
		MRR:              req.MRR,
		Customers:        req.Customers,
		MonthlyChurnRate: req.MonthlyChurnRate,
	}

	metrics, err := h.service.SetMetrics(r.Context(), userID, input)
	if err != nil {
		// Check if it's a validation error
		if err.Error() == "mrr must be non-negative" ||
			err.Error() == "customers must be non-negative" ||
			err.Error() == "monthly_churn_rate must be non-negative" {
			writeJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)
}

