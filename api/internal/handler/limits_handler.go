package handler

import (
	"encoding/json"
	"net/http"

	"rev-saas-api/internal/middleware"
	"rev-saas-api/internal/service"
)

// LimitsHandler handles HTTP requests for usage limits.
type LimitsHandler struct {
	limitsService *service.LimitsService
}

// NewLimitsHandler creates a new LimitsHandler.
func NewLimitsHandler(limitsService *service.LimitsService) *LimitsHandler {
	return &LimitsHandler{
		limitsService: limitsService,
	}
}

// GetUsageStats handles GET /api/usage - returns current usage stats for the user.
func (h *LimitsHandler) GetUsageStats(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.limitsService.GetUserUsageStats(r.Context(), user)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

// GetPlanLimits handles GET /api/plans/limits - returns all plan limits (public info).
func (h *LimitsHandler) GetPlanLimits(w http.ResponseWriter, r *http.Request) {
	limits := service.GetAllPlanLimits()

	// Convert to a friendlier JSON format
	response := make(map[string]interface{})
	for plan, limit := range limits {
		response[plan] = map[string]interface{}{
			"max_competitors":        limit.MaxCompetitors,
			"max_plans":              limit.MaxPlans,
			"max_analyses_per_month": limit.MaxAnalysesPerMonth,
			"max_analyses_total":     limit.MaxAnalysesTotal,
			"trial_days":             limit.TrialDays,
			"is_unlimited":           limit.IsUnlimited,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

