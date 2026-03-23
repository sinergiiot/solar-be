package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDContextKey contextKey = "auth_user_id"
	RoleContextKey   contextKey = "auth_user_role"
)

// Middleware validates bearer tokens and injects user id into request context.
func Middleware(authService Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
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

			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			// Note: We only have userID from token. 
			// To avoid DB call in middleware for every request, we SHOULD bake role into JWT.
			// Let's update buildAccessToken to include role.
			role, _ := claims["role"].(string)
			ctx = context.WithValue(ctx, RoleContextKey, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts authenticated user id from request context.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(uuid.UUID)
	return userID, ok
}

// UserRoleFromContext extracts authenticated user role from request context.
func UserRoleFromContext(ctx context.Context) string {
	role, _ := ctx.Value(RoleContextKey).(string)
	return role
}
