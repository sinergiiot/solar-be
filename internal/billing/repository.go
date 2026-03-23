package billing

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateSubscription(ctx context.Context, sub *Subscription) error {
	query := `
		INSERT INTO subscriptions (
			id, user_id, plan_tier, status, billing_cycle,
			amount, currency, external_checkout_id, expires_at,
			next_billing_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.ExecContext(ctx, query,
		sub.ID, sub.UserID, sub.PlanTier, sub.Status, sub.BillingCycle,
		sub.Amount, sub.Currency, sub.ExternalCheckoutID, sub.ExpiresAt,
		sub.NextBillingAt, sub.CreatedAt, sub.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}
	return nil
}

func (r *repository) GetLatestSubscription(ctx context.Context, userID uuid.UUID) (*Subscription, error) {
	query := `
		SELECT id, user_id, plan_tier, status, billing_cycle,
			amount, currency, external_checkout_id, expires_at,
			next_billing_at, last_payment_at, grace_period_until,
			created_at, updated_at
		FROM subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	sub := &Subscription{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&sub.ID, &sub.UserID, &sub.PlanTier, &sub.Status, &sub.BillingCycle,
		&sub.Amount, &sub.Currency, &sub.ExternalCheckoutID, &sub.ExpiresAt,
		&sub.NextBillingAt, &sub.LastPaymentAt, &sub.GracePeriodUntil,
		&sub.CreatedAt, &sub.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get latest subscription: %w", err)
	}
	return sub, nil
}

func (r *repository) UpdateSubscription(ctx context.Context, sub *Subscription) error {
	query := `
		UPDATE subscriptions
		SET plan_tier = $2, status = $3, expires_at = $4,
			next_billing_at = $5, last_payment_at = $6,
			grace_period_until = $7, updated_at = $8
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		sub.ID, sub.PlanTier, sub.Status, sub.ExpiresAt,
		sub.NextBillingAt, sub.LastPaymentAt,
		sub.GracePeriodUntil, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}
	return nil
}

func (r *repository) GetPastDueSubscriptions(ctx context.Context, now time.Time) ([]*Subscription, error) {
	query := `
		SELECT id, user_id, plan_tier, status, expires_at
		FROM subscriptions
		WHERE status != 'free' AND grace_period_until < $1
	`
	rows, err := r.db.QueryContext(ctx, query, now)
	if err != nil {
		return nil, fmt.Errorf("get past due subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []*Subscription
	for rows.Next() {
		sub := &Subscription{}
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.PlanTier, &sub.Status, &sub.ExpiresAt); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}
