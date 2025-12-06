package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"rev-saas-api/internal/model"
	mongorepo "rev-saas-api/internal/repository/mongo"
	"rev-saas-api/internal/service"
)

type contextKey string

const (
	userIDContextKey contextKey = "userID"
	userContextKey   contextKey = "user"
)

// AuthMiddleware handles JWT authentication.
type AuthMiddleware struct {
	jwt      *service.JWTService
	userRepo *mongorepo.UserRepository
}

// NewAuthMiddleware creates a new AuthMiddleware.
func NewAuthMiddleware(jwt *service.JWTService, userRepo *mongorepo.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		jwt:      jwt,
		userRepo: userRepo,
	}
}

// RequireAuth is a standard HTTP middleware that enforces JWT auth.
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			unauthorizedJSON(w, "missing Authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			unauthorizedJSON(w, "invalid Authorization header format")
			return
		}

		tokenStr := strings.TrimSpace(parts[1])
		if tokenStr == "" {
			unauthorizedJSON(w, "empty token")
			return
		}

		claims, err := m.jwt.ParseToken(tokenStr)
		if err != nil {
			unauthorizedJSON(w, "invalid or expired token")
			return
		}

		userID := claims.UserID
		if userID == "" {
			unauthorizedJSON(w, "invalid token: missing user_id")
			return
		}

		// Store user ID in context
		ctx := context.WithValue(r.Context(), userIDContextKey, userID)

		// Fetch and store full user object for limit checking
		if m.userRepo != nil {
			user, err := m.userRepo.GetByIDString(ctx, userID)
			if err == nil && user != nil {
				ctx = context.WithValue(ctx, userContextKey, user)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func unauthorizedJSON(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// UserIDFromContext returns the user ID stored by the auth middleware, or "" if not present.
func UserIDFromContext(ctx context.Context) string {
	v := ctx.Value(userIDContextKey)
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// UserFromContext returns the full user object stored by the auth middleware, or nil if not present.
func UserFromContext(ctx context.Context) *model.User {
	v := ctx.Value(userContextKey)
	if v == nil {
		return nil
	}
	if u, ok := v.(*model.User); ok {
		return u
	}
	return nil
}

