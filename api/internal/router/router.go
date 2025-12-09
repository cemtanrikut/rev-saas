package router

import (
	"net/http"

	"github.com/gorilla/mux"

	"rev-saas-api/internal/handler"
	"rev-saas-api/internal/middleware"
)

// NewRouter creates and configures a new HTTP router.
func NewRouter(
	healthHandler *handler.HealthHandler,
	authHandler *handler.AuthHandler,
	planHandler *handler.PlanHandler,
	competitorHandler *handler.CompetitorHandler,
	analysisHandler *handler.AnalysisHandler,
	analysisPDFHandler *handler.AnalysisPDFHandler,
	businessMetricsHandler *handler.BusinessMetricsHandler,
	limitsHandler *handler.LimitsHandler,
	simulationHandler *handler.SimulationHandler,
	authMiddleware *middleware.AuthMiddleware,
) http.Handler {
	r := mux.NewRouter()

	// Health check endpoint
	r.HandleFunc("/health", healthHandler.Health).Methods(http.MethodGet)

	// Auth endpoints (public)
	r.HandleFunc("/auth/signup", authHandler.Signup).Methods(http.MethodPost)
	r.HandleFunc("/auth/login", authHandler.Login).Methods(http.MethodPost)

	// Public plan limits (no auth required)
	r.HandleFunc("/api/plans/limits", limitsHandler.GetPlanLimits).Methods(http.MethodGet)

	// Auth endpoints (protected)
	r.Handle("/auth/me", authMiddleware.RequireAuth(http.HandlerFunc(authHandler.Me))).
		Methods(http.MethodGet)

	// API v1 (protected)
	api := r.PathPrefix("/api").Subrouter()
	api.Use(authMiddleware.RequireAuth)

	// Usage stats
	api.HandleFunc("/usage", limitsHandler.GetUsageStats).Methods(http.MethodGet)

	// Plans
	api.HandleFunc("/plans", planHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/plans", planHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/plans/{id}", planHandler.Delete).Methods(http.MethodDelete)

	// Competitors
	api.HandleFunc("/competitors", competitorHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/competitors", competitorHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/competitors/{id}", competitorHandler.Update).Methods(http.MethodPut)
	api.HandleFunc("/competitors/{id}", competitorHandler.Delete).Methods(http.MethodDelete)

	// Analysis
	api.HandleFunc("/analysis/run", analysisHandler.RunAnalysis).Methods(http.MethodPost)
	api.HandleFunc("/analysis", analysisHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/analysis/{id}/export-pdf", analysisPDFHandler.ExportPDF).Methods(http.MethodGet)

	// Business Metrics
	api.HandleFunc("/business-metrics", businessMetricsHandler.Get).Methods(http.MethodGet)
	api.HandleFunc("/business-metrics", businessMetricsHandler.Set).Methods(http.MethodPut)

	// Simulations
	api.HandleFunc("/simulations", simulationHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/simulations", simulationHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/simulations/{id}", simulationHandler.Get).Methods(http.MethodGet)
	api.HandleFunc("/simulations/{id}/pdf", simulationHandler.ExportPDF).Methods(http.MethodGet)

	// Apply CORS middleware to all routes
	return middleware.CORS(r)
}
