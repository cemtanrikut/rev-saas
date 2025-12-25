package handler

import (
	"encoding/json"
	"net/http"

	"rev-saas-api/internal/middleware"
	"rev-saas-api/internal/service"
)

// AICreditsHandler handles HTTP requests for AI credits.
type AICreditsHandler struct {
	aiCreditsService *service.AICreditsService
}

// NewAICreditsHandler creates a new AICreditsHandler.
func NewAICreditsHandler(aiCreditsService *service.AICreditsService) *AICreditsHandler {
	return &AICreditsHandler{
		aiCreditsService: aiCreditsService,
	}
}

// GetCredits handles GET /api/ai-credits - returns current AI credits info.
func (h *AICreditsHandler) GetCredits(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	planType := user.GetEffectivePlan()
	creditsInfo, err := h.aiCreditsService.GetCreditsInfo(r.Context(), userID, planType)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(creditsInfo)
}

// WriteAIQuotaExceededError writes a standardized AI quota exceeded error response.
func WriteAIQuotaExceededError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusPaymentRequired) // 402
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    service.AIQuotaExceededCode,
		"message": "You have used all your AI Insight Credits for this month on your current plan.",
	})
}

// WriteSimulationNotAvailableError writes a standardized simulation not available error response.
func WriteSimulationNotAvailableError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden) // 403
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    service.SimulationNotAvailableCode,
		"message": "Pricing simulations are only available on Growth and Enterprise plans.",
	})
}





