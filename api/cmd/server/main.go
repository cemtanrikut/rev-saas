package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rev-saas-api/internal/config"
	"rev-saas-api/internal/handler"
	"rev-saas-api/internal/middleware"
	mongorepo "rev-saas-api/internal/repository/mongo"
	"rev-saas-api/internal/router"
	"rev-saas-api/internal/service"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize MongoDB connection
	mongoClient, err := mongorepo.NewClient(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := mongoClient.Close(ctx); err != nil {
			log.Printf("error closing Mongo client: %v", err)
		}
	}()

	// Get database instance
	db := mongoClient.DB()

	// Initialize repositories
	userRepo := mongorepo.NewUserRepository(db)
	companyRepo := mongorepo.NewCompanyRepository(db)
	userMetadataRepo := mongorepo.NewUserMetadataRepository(db)
	planRepo := mongorepo.NewPlanRepository(db)
	competitorRepo := mongorepo.NewCompetitorRepository(db)
	analysisRepo := mongorepo.NewAnalysisRepository(db)
	businessMetricsRepo := mongorepo.NewBusinessMetricsRepository(db)

	// Initialize services
	jwtService := service.NewJWTService(cfg.JWTSecret)
	authService := service.NewAuthService(userRepo, companyRepo, userMetadataRepo, jwtService)
	planService := service.NewPlanService(planRepo)
	competitorService := service.NewCompetitorService(competitorRepo)
	analysisService := service.NewAnalysisService(analysisRepo, planRepo, competitorRepo, businessMetricsRepo)
	businessMetricsService := service.NewBusinessMetricsService(businessMetricsRepo)
	limitsService := service.NewLimitsService(userRepo, planRepo, competitorRepo, analysisRepo)
	aiPricingService := service.NewAIPricingService(cfg.OpenAIAPIKey)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService, userRepo)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler()
	authHandler := handler.NewAuthHandler(authService)
	planHandler := handler.NewPlanHandler(planService, limitsService)
	competitorHandler := handler.NewCompetitorHandler(competitorService, limitsService)
	analysisHandler := handler.NewAnalysisHandler(analysisService, limitsService, aiPricingService)
	analysisPDFHandler := handler.NewAnalysisPDFHandler(analysisService, businessMetricsRepo)
	businessMetricsHandler := handler.NewBusinessMetricsHandler(businessMetricsService)
	limitsHandler := handler.NewLimitsHandler(limitsService)

	// Create router
	r := router.NewRouter(healthHandler, authHandler, planHandler, competitorHandler, analysisHandler, analysisPDFHandler, businessMetricsHandler, limitsHandler, authMiddleware)

	// Configure HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on :%s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline to wait for current operations to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	log.Println("Server stopped gracefully")
}
