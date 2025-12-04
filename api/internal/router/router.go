package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"rev-saas-api/internal/handler"
)

// NewRouter creates and configures a new HTTP router.
func NewRouter(healthHandler *handler.HealthHandler) http.Handler {
	r := mux.NewRouter()

	// Health check endpoint
	r.HandleFunc("/health", healthHandler.Health).Methods(http.MethodGet)

	// API v1 routes will be mounted here in the future
	// apiV1 := r.PathPrefix("/api/v1").Subrouter()
	// apiV1.HandleFunc("/users", ...).Methods(http.MethodGet)

	return r
}

