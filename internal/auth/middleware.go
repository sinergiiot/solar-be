package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/akbarsenawijaya/solar-forecast/pkg/ctxkeys"
	"github.com/google/uuid"
)

// APIKeyValidator defines the interface needed to validate API keys without circular dependency
type APIKeyValidator interface {
	ValidateKey(key string) (uuid.UUID, string, error)
}

// ServiceBridge implements APIKeyValidator for the middleware
type ServiceBridge struct {
	ValidateFn func(key string) (uuid.UUID, string, error)
}

func (b *ServiceBridge) ValidateKey(key string) (uuid.UUID, string, error) {
	return b.ValidateFn(key)
}

// Middleware validates bearer tokens or API keys and injects user info into request context.
func Middleware(authService Service, apiKeyValidator APIKeyValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Try API Key first (X-API-Key header)
			apiKey := r.Header.Get("X-API-Key")
			if apiKey != "" && apiKeyValidator != nil {
				userID, role, err := apiKeyValidator.ValidateKey(apiKey)
				if err == nil {
					ctx := context.WithValue(r.Context(), ctxkeys.UserID, userID)
					ctx = context.WithValue(ctx, ctxkeys.UserRole, role)
					ctx = context.WithValue(ctx, ctxkeys.IsAPIKey, true)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				// If API key is provided but invalid, we fail early
				WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid api key"})
				return
			}

			// 2. Try Bearer Token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization header or api key"})
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid authorization header"})
				return
			}

			userID, claims, err := authService.ParseToken(parts[1])
			if err != nil {
				WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
				return
			}

			ctx := context.WithValue(r.Context(), ctxkeys.UserID, userID)
			role, _ := claims["role"].(string)
			ctx = context.WithValue(ctx, ctxkeys.UserRole, role)
			ctx = context.WithValue(ctx, ctxkeys.IsAPIKey, false)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts authenticated user id from request context.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(ctxkeys.UserID).(uuid.UUID)
	return userID, ok
}

// UserRoleFromContext extracts authenticated user role from request context.
func UserRoleFromContext(ctx context.Context) string {
	role, _ := ctx.Value(ctxkeys.UserRole).(string)
	return role
}
// RequireAdmin ensures only users with role 'admin' can access certain routes.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := UserRoleFromContext(r.Context())
		if role != "admin" {
			WriteJSON(w, http.StatusForbidden, map[string]string{"error": "admin access required"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
