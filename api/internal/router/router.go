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
	competitorV2Handler *handler.CompetitorV2Handler,
	pricingV2Handler *handler.PricingV2Handler,
	analysisHandler *handler.AnalysisHandler,
	analysisPDFHandler *handler.AnalysisPDFHandler,
	analysisV2Handler *handler.AnalysisV2Handler,
	businessMetricsHandler *handler.BusinessMetricsHandler,
	limitsHandler *handler.LimitsHandler,
	simulationHandler *handler.SimulationHandler,
	aiCreditsHandler *handler.AICreditsHandler,
	stripeHandler *handler.StripeHandler,
	billingHandler *handler.BillingHandler,
	authMiddleware *middleware.AuthMiddleware,
) http.Handler {
	r := mux.NewRouter()
	r.StrictSlash(true)

	// Health check endpoint
	r.HandleFunc("/health", healthHandler.Health).Methods(http.MethodGet)

	// Auth endpoints (public)
	r.HandleFunc("/auth/signup", authHandler.Signup).Methods(http.MethodPost)
	r.HandleFunc("/auth/login", authHandler.Login).Methods(http.MethodPost)
	r.HandleFunc("/auth/verify-email", authHandler.VerifyEmail).Methods(http.MethodGet)
	r.HandleFunc("/auth/resend-verification", authHandler.ResendVerification).Methods(http.MethodPost)

	// Public plan limits (no auth required)
	r.HandleFunc("/api/plans/limits", limitsHandler.GetPlanLimits).Methods(http.MethodGet)

	// Auth endpoints (protected)
	r.Handle("/auth/me", authMiddleware.RequireAuth(http.HandlerFunc(authHandler.Me))).
		Methods(http.MethodGet)

	// API v1 (protected)
	api := r.PathPrefix("/api").Subrouter()
	api.StrictSlash(true)
	api.Use(authMiddleware.RequireAuth)


	// Usage stats
	api.HandleFunc("/usage", limitsHandler.GetUsageStats).Methods(http.MethodGet)

	// AI Insight Credits
	api.HandleFunc("/ai-credits", aiCreditsHandler.GetCredits).Methods(http.MethodGet)

	// Plans
	api.HandleFunc("/plans", planHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/plans", planHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/plans/{id}", planHandler.Update).Methods(http.MethodPut)
	api.HandleFunc("/plans/{id}", planHandler.Delete).Methods(http.MethodDelete)

	// Competitors (V1 - legacy)
	api.HandleFunc("/competitors", competitorHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/competitors", competitorHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/competitors/{id}", competitorHandler.Update).Methods(http.MethodPut)
	api.HandleFunc("/competitors/{id}", competitorHandler.Delete).Methods(http.MethodDelete)

	// Competitors V2 (AI discovery) - use separate subrouter for v2
	apiv2 := r.PathPrefix("/api/v2").Subrouter()
	apiv2.Use(authMiddleware.RequireAuth)
	apiv2.HandleFunc("/competitors/discover", competitorV2Handler.Discover).Methods(http.MethodPost)
	apiv2.HandleFunc("/competitors/save", competitorV2Handler.Save).Methods(http.MethodPost)
	apiv2.HandleFunc("/competitors", competitorV2Handler.List).Methods(http.MethodGet)
	apiv2.HandleFunc("/competitors/{id}", competitorV2Handler.Delete).Methods(http.MethodDelete)

	// Pricing V2 (auto-import from website)
	api.HandleFunc("/pricing-v2/discover", pricingV2Handler.Discover).Methods(http.MethodPost)
	api.HandleFunc("/pricing-v2/extract", pricingV2Handler.Extract).Methods(http.MethodPost)
	api.HandleFunc("/pricing-v2/extract-from-text", pricingV2Handler.ExtractFromText).Methods(http.MethodPost)
	api.HandleFunc("/pricing-v2/save", pricingV2Handler.Save).Methods(http.MethodPost)
	api.HandleFunc("/pricing-v2", pricingV2Handler.List).Methods(http.MethodGet)
	api.HandleFunc("/pricing-v2/{id}", pricingV2Handler.Delete).Methods(http.MethodDelete)

	// Analysis (V1 - legacy)
	api.HandleFunc("/analysis/run", analysisHandler.RunAnalysis).Methods(http.MethodPost)
	api.HandleFunc("/analysis", analysisHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/analysis/{id}/export-pdf", analysisPDFHandler.ExportPDF).Methods(http.MethodGet)

	// Analysis V2 (new deterministic engine)
	api.HandleFunc("/analysis/v2", analysisV2Handler.RunAnalysisV2).Methods(http.MethodPost)
	api.HandleFunc("/analysis/v2", analysisV2Handler.ListAnalysesV2).Methods(http.MethodGet)
	api.HandleFunc("/analysis/v2/{id}", analysisV2Handler.GetAnalysisV2).Methods(http.MethodGet)
	api.HandleFunc("/analysis/v2/{id}/pdf", analysisV2Handler.ExportPDFV2).Methods(http.MethodGet)

	// Business Metrics
	api.HandleFunc("/business-metrics", businessMetricsHandler.Get).Methods(http.MethodGet)
	api.HandleFunc("/business-metrics", businessMetricsHandler.Set).Methods(http.MethodPut)

	// Simulations
	api.HandleFunc("/simulations", simulationHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/simulations", simulationHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/simulations/{id}", simulationHandler.Get).Methods(http.MethodGet)
	api.HandleFunc("/simulations/{id}/pdf", simulationHandler.ExportPDF).Methods(http.MethodGet)

	// Stripe Connect (protected endpoints)
	api.HandleFunc("/stripe/status", stripeHandler.GetStatus).Methods(http.MethodGet)
	api.HandleFunc("/stripe/connect", stripeHandler.Connect).Methods(http.MethodPost)
	api.HandleFunc("/stripe/disconnect", stripeHandler.Disconnect).Methods(http.MethodPost)
	api.HandleFunc("/stripe/sync/metrics", stripeHandler.SyncMetrics).Methods(http.MethodPost)
	api.HandleFunc("/stripe/sync/prices", stripeHandler.SyncPrices).Methods(http.MethodPost)

	// Stripe OAuth callback (public - state token validates user)
	r.HandleFunc("/api/stripe/connect/callback", stripeHandler.Callback).Methods(http.MethodGet)

	// Stripe Billing (protected endpoints)
	api.HandleFunc("/billing/status", billingHandler.GetStatus).Methods(http.MethodGet)
	api.HandleFunc("/billing/checkout-session", billingHandler.CreateCheckoutSession).Methods(http.MethodPost)
	api.HandleFunc("/billing/portal", billingHandler.CreatePortalSession).Methods(http.MethodPost)

	// Stripe Billing Webhook (public - uses Stripe signature verification, not auth)
	r.HandleFunc("/api/billing/webhook", billingHandler.HandleWebhook).Methods(http.MethodPost)

	// Apply CORS middleware to all routes
	return middleware.CORS(r)
}
