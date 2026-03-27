package middleware

import (
	"context"
	"net/http"

	"github.com/akbarsenawijaya/solar-forecast/internal/tier"
	"github.com/google/uuid"
)

type contextKey string

const (
	TierContextKey contextKey = "user_tier"
)

// TierProvider fetches plan_tier from database.
type TierProvider interface {
	GetPlanTier(userID uuid.UUID) (string, error)
}

// UserIDGetter extracts userID from another context (e.g. from auth middleware).
type UserIDGetter func(ctx context.Context) (uuid.UUID, bool)

// TierMiddleware injects the user's plan tier into the request context.
func TierMiddleware(provider TierProvider, getID UserIDGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := getID(r.Context())
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			planTier, err := provider.GetPlanTier(userID)
			if err != nil {
				planTier = tier.Free
			}

			ctx := context.WithValue(r.Context(), TierContextKey, planTier)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireFeature gates access based on tier.CanAccess.
func RequireFeature(feature string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userTier := GetTierFromContext(r.Context())
			if !tier.CanAccess(userTier, feature) {
				http.Error(w, "Upgrade to Pro/Enterprise to unlock this feature.", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetTierFromContext extracts tier injected by TierMiddleware.
func GetTierFromContext(ctx context.Context) string {
	if t, ok := ctx.Value(TierContextKey).(string); ok {
		return t
	}
	return tier.Free
}
