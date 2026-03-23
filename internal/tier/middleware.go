package tier

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const TierContextKey contextKey = "plan_tier"

// Repository interface for the middleware to fetch user tier.
type TierProvider interface {
	GetPlanTier(userID uuid.UUID) (string, error)
}

// TierMiddleware injects the user's plan tier into the request context.
func TierMiddleware(provider TierProvider, userIDGetter func(context.Context) (uuid.UUID, bool)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := userIDGetter(r.Context())
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			planTier, err := provider.GetPlanTier(userID)
			if err != nil {
				// On error, default to free to be safe.
				planTier = PlanFree
			}

			ctx := context.WithValue(r.Context(), TierContextKey, planTier)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireFeature returns a middleware that checks if the current user tier has access to all specified features.
func RequireFeature(features ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userTier := GetTierFromContext(r.Context())

			for _, feature := range features {
				if !CanAccess(userTier, feature) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error":         "feature_not_available",
						"message":       "Fitur ini tidak tersedia di paket Anda.",
						"tier":          userTier,
						"required_feature": feature,
						"upgrade_url":   "/pricing",
					})
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetTierFromContext extracts the injected plan tier from the request context.
func GetTierFromContext(ctx context.Context) string {
	if tierVal, ok := ctx.Value(TierContextKey).(string); ok {
		return strings.ToLower(tierVal)
	}
	return PlanFree
}
