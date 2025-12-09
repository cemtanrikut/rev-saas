package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"rev-saas-api/internal/middleware"
	"rev-saas-api/internal/model"
	"rev-saas-api/internal/service"
)

// SimulationHandler handles HTTP requests for pricing simulations.
type SimulationHandler struct {
	service   *service.SimulationService
	aiService *service.AIPricingService
}

// NewSimulationHandler creates a new SimulationHandler.
func NewSimulationHandler(service *service.SimulationService, aiService *service.AIPricingService) *SimulationHandler {
	return &SimulationHandler{
		service:   service,
		aiService: aiService,
	}
}

// createSimulationRequest represents the request body for creating a simulation.
type createSimulationRequest struct {
	PlanID                string  `json:"plan_id"`
	CurrentPrice          float64 `json:"current_price"`
	NewPrice              float64 `json:"new_price"`
	Currency              string  `json:"currency"`
	ActiveCustomersOnPlan int     `json:"active_customers_on_plan"`
	GlobalMRR             float64 `json:"global_mrr"`
	GlobalChurnRate       float64 `json:"global_churn_rate"`
	PricingGoal           string  `json:"pricing_goal"`
}

// Create handles POST /api/simulations - runs a new pricing simulation.
func (h *SimulationHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	// Parse request body
	var req createSimulationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.PlanID == "" {
		writeJSONError(w, "plan_id is required", http.StatusBadRequest)
		return
	}
	if req.CurrentPrice <= 0 {
		writeJSONError(w, "current_price must be greater than zero", http.StatusBadRequest)
		return
	}
	if req.NewPrice < 0 {
		writeJSONError(w, "new_price cannot be negative", http.StatusBadRequest)
		return
	}
	if req.ActiveCustomersOnPlan < 0 {
		writeJSONError(w, "active_customers_on_plan cannot be negative", http.StatusBadRequest)
		return
	}
	if req.GlobalMRR < 0 {
		writeJSONError(w, "global_mrr cannot be negative", http.StatusBadRequest)
		return
	}
	if req.GlobalChurnRate < 0 || req.GlobalChurnRate > 100 {
		writeJSONError(w, "global_churn_rate must be between 0 and 100", http.StatusBadRequest)
		return
	}

	// Default currency
	if req.Currency == "" {
		req.Currency = "USD"
	}

	// Build simulation request
	simReq := model.SimulationRequest{
		PlanID:                req.PlanID,
		CurrentPrice:          req.CurrentPrice,
		NewPrice:              req.NewPrice,
		Currency:              req.Currency,
		ActiveCustomersOnPlan: req.ActiveCustomersOnPlan,
		GlobalMRR:             req.GlobalMRR,
		GlobalChurnRate:       req.GlobalChurnRate,
		PricingGoal:           req.PricingGoal,
	}

	// Run simulation
	result, err := h.service.RunPricingSimulation(r.Context(), userID, simReq)
	if err != nil {
		// Check for specific errors
		switch err {
		case service.ErrInvalidPriceChange:
			writeJSONError(w, err.Error(), http.StatusBadRequest)
		case service.ErrNoBucketFound:
			writeJSONError(w, "price change percentage out of supported range", http.StatusBadRequest)
		case service.ErrNoProfileFound:
			writeJSONError(w, "invalid pricing goal", http.StatusBadRequest)
		default:
			if err.Error() == "plan not found or access denied" {
				writeJSONError(w, err.Error(), http.StatusNotFound)
			} else {
				writeJSONError(w, err.Error(), http.StatusInternalServerError)
			}
		}
		return
	}

	// Generate AI narrative (optional, non-blocking)
	if h.aiService != nil && h.aiService.IsEnabled() {
		narrative, err := h.aiService.GenerateSimulationNarrative(r.Context(), result)
		if err == nil && narrative != "" {
			result.AINarrative = narrative
		}
		// If AI fails, we just skip it
	}

	// Save simulation to database
	if err := h.service.SaveSimulation(r.Context(), result); err != nil {
		writeJSONError(w, "failed to save simulation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

// List handles GET /api/simulations - lists simulations for the current user.
func (h *SimulationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	planID := r.URL.Query().Get("plan_id")
	limitStr := r.URL.Query().Get("limit")
	
	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			if limit > 100 {
				limit = 100 // cap at 100
			}
		}
	}

	simulations, err := h.service.ListSimulations(r.Context(), userID, planID, limit)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return empty array instead of null
	if simulations == nil {
		simulations = []*model.SimulationResult{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(simulations)
}

// Get handles GET /api/simulations/{id} - gets a specific simulation.
func (h *SimulationHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		writeJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	simulationID := vars["id"]
	if simulationID == "" {
		writeJSONError(w, "missing simulation id", http.StatusBadRequest)
		return
	}

	simulation, err := h.service.GetSimulationByID(r.Context(), userID, simulationID)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if simulation == nil {
		writeJSONError(w, "simulation not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(simulation)
}

// ExportPDF handles GET /api/simulations/{id}/pdf - exports a simulation as PDF.
func (h *SimulationHandler) ExportPDF(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	simulationID := vars["id"]
	if simulationID == "" {
		http.Error(w, "missing simulation id", http.StatusBadRequest)
		return
	}

	// Fetch the simulation (ensures it belongs to the user)
	simulation, err := h.service.GetSimulationByID(r.Context(), userID, simulationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if simulation == nil {
		http.Error(w, "simulation not found", http.StatusNotFound)
		return
	}

	// Generate PDF
	pdfData := service.SimulationPDFData{
		Simulation: simulation,
	}

	pdfBuffer, err := service.GenerateSimulationPDF(pdfData)
	if err != nil {
		http.Error(w, "failed to generate PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate filename with date
	filename := fmt.Sprintf("pricing-simulation-%s-%s.pdf",
		sanitizeFilename(simulation.PlanName),
		simulation.CreatedAt.Format("20060102"),
	)

	// Set headers for PDF download
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", pdfBuffer.Len()))

	// Write PDF to response
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBuffer.Bytes())
}

// sanitizeFilename removes or replaces characters that shouldn't be in filenames
func sanitizeFilename(name string) string {
	// Replace spaces with dashes
	name = strings.ReplaceAll(name, " ", "-")
	// Remove or replace other problematic characters
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return -1 // Remove character
	}, name)
	// Lowercase
	name = strings.ToLower(name)
	// Limit length
	if len(name) > 50 {
		name = name[:50]
	}
	return name
}

