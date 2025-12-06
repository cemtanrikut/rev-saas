package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"rev-saas-api/internal/middleware"
	"rev-saas-api/internal/model"
	"rev-saas-api/internal/service"
)

// PlanHandler handles plan-related HTTP endpoints.
type PlanHandler struct {
	service       *service.PlanService
	limitsService *service.LimitsService
}

// NewPlanHandler creates a new PlanHandler.
func NewPlanHandler(service *service.PlanService, limitsService *service.LimitsService) *PlanHandler {
	return &PlanHandler{
		service:       service,
		limitsService: limitsService,
	}
}

type createPlanRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// Create handles plan creation.
func (h *PlanHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user for limit checking
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Check limits
	result := h.limitsService.CanAddPlan(r.Context(), user)
	if !result.Allowed {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   result.ErrorCode,
			"reason":  result.Reason,
			"plan":    result.Plan,
			"limit":   result.Limit,
			"current": result.Current,
		})
		return
	}

	var req createPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	plan, err := h.service.CreatePlan(r.Context(), userID, req.Name, req.Price)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(plan)
}

// List handles listing all plans for the authenticated user.
func (h *PlanHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	plans, err := h.service.ListPlans(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to list plans", http.StatusInternalServerError)
		return
	}

	// Return empty array instead of null
	if plans == nil {
		plans = []*model.Plan{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(plans)
}

// Delete handles plan deletion.
func (h *PlanHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "missing plan id", http.StatusBadRequest)
		return
	}

	err := h.service.DeletePlan(r.Context(), userID, id)
	if err != nil {
		if err == service.ErrPlanNotFound {
			http.Error(w, "plan not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete plan", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

