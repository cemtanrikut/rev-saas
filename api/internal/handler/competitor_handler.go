package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"rev-saas-api/internal/middleware"
	"rev-saas-api/internal/model"
	"rev-saas-api/internal/service"
)

// CompetitorHandler handles competitor-related HTTP endpoints.
type CompetitorHandler struct {
	service       *service.CompetitorService
	limitsService *service.LimitsService
}

// NewCompetitorHandler creates a new CompetitorHandler.
func NewCompetitorHandler(service *service.CompetitorService, limitsService *service.LimitsService) *CompetitorHandler {
	return &CompetitorHandler{
		service:       service,
		limitsService: limitsService,
	}
}

type competitorPlanRequest struct {
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Currency     string  `json:"currency"`
	BillingCycle string  `json:"billing_cycle"`
}

type createCompetitorRequest struct {
	Name  string                  `json:"name"`
	URL   string                  `json:"url"`
	Plans []competitorPlanRequest `json:"plans"`
}

type limitErrorResponse struct {
	Error   string `json:"error"`
	Reason  string `json:"reason"`
	Plan    string `json:"plan"`
	Limit   int    `json:"limit,omitempty"`
	Current int    `json:"current,omitempty"`
}

// Create handles competitor creation.
func (h *CompetitorHandler) Create(w http.ResponseWriter, r *http.Request) {
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
	result := h.limitsService.CanAddCompetitor(r.Context(), user)
	if !result.Allowed {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(limitErrorResponse{
			Error:   result.ErrorCode,
			Reason:  result.Reason,
			Plan:    result.Plan,
			Limit:   result.Limit,
			Current: result.Current,
		})
		return
	}

	var req createCompetitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Convert request plans to service input
	planInputs := make([]service.CompetitorPlanInput, 0, len(req.Plans))
	for _, p := range req.Plans {
		planInputs = append(planInputs, service.CompetitorPlanInput{
			Name:         p.Name,
			Price:        p.Price,
			Currency:     p.Currency,
			BillingCycle: p.BillingCycle,
		})
	}

	competitor, err := h.service.CreateCompetitor(r.Context(), userID, service.CompetitorInput{
		Name:  req.Name,
		URL:   req.URL,
		Plans: planInputs,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(competitor)
}

// List handles listing all competitors for the authenticated user.
func (h *CompetitorHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	competitors, err := h.service.ListCompetitors(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to list competitors", http.StatusInternalServerError)
		return
	}

	// Return empty array instead of null
	if competitors == nil {
		competitors = []*model.Competitor{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(competitors)
}

// Delete handles competitor deletion.
func (h *CompetitorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "missing competitor id", http.StatusBadRequest)
		return
	}

	err := h.service.DeleteCompetitor(r.Context(), userID, id)
	if err != nil {
		if err == service.ErrCompetitorNotFound {
			http.Error(w, "competitor not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete competitor", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Update handles competitor updates.
func (h *CompetitorHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "missing competitor id", http.StatusBadRequest)
		return
	}

	var req createCompetitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Convert request plans to service input
	planInputs := make([]service.CompetitorPlanInput, 0, len(req.Plans))
	for _, p := range req.Plans {
		planInputs = append(planInputs, service.CompetitorPlanInput{
			Name:         p.Name,
			Price:        p.Price,
			Currency:     p.Currency,
			BillingCycle: p.BillingCycle,
		})
	}

	competitor, err := h.service.UpdateCompetitor(r.Context(), userID, id, service.CompetitorInput{
		Name:  req.Name,
		URL:   req.URL,
		Plans: planInputs,
	})
	if err != nil {
		if err == service.ErrCompetitorNotFound {
			http.Error(w, "competitor not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(competitor)
}
