package billing

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID                 uuid.UUID  `json:"id"`
	UserID             uuid.UUID  `json:"user_id"`
	PlanTier           string     `json:"plan_tier"`           // 'free', 'pro', 'enterprise'
	Status             string     `json:"status"`              // 'active', 'inactive', 'past_due', 'canceled'
	BillingCycle       string     `json:"billing_cycle"`       // 'monthly', 'yearly'
	Amount             int64      `json:"amount"`
	Currency           string     `json:"currency"`
	ExternalCheckoutID string     `json:"external_checkout_id"`
	ExpiresAt          time.Time  `json:"expires_at"`
	NextBillingAt     *time.Time `json:"next_billing_at,omitempty"`
	LastPaymentAt      *time.Time `json:"last_payment_at,omitempty"`
	GracePeriodUntil   *time.Time `json:"grace_period_until,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type CheckoutRequest struct {
	PlanTier     string `json:"plan_tier" validate:"required,oneof=pro enterprise"`
	BillingCycle string `json:"billing_cycle" fallback:"monthly"`
}

type CheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
	ID          string `json:"id"`
}
