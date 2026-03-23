package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	CreateSubscription(ctx context.Context, sub *Subscription) error
	GetLatestSubscription(ctx context.Context, userID uuid.UUID) (*Subscription, error)
	UpdateSubscription(ctx context.Context, sub *Subscription) error
	// For scheduler/cleanup
	GetPastDueSubscriptions(ctx context.Context, now time.Time) ([]*Subscription, error)
}

type Service interface {
	InitiateCheckout(ctx context.Context, userID uuid.UUID, req CheckoutRequest) (*CheckoutResponse, error)
	HandleWebhook(ctx context.Context, externalID string, payload map[string]interface{}) error
	GetSubscriptionStatus(ctx context.Context, userID uuid.UUID) (*Subscription, error)
	CleanupExpiredSubscriptions(ctx context.Context) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) InitiateCheckout(ctx context.Context, userID uuid.UUID, req CheckoutRequest) (*CheckoutResponse, error) {
	// In a real world, calling Midtrans/Xendit to get a payment link.
	// We'll generate a fake checkout URL for now.
	checkoutID := uuid.New().String()
	fakeURL := fmt.Sprintf("https://mock-payment.sinergi-iot.id/checkout/%s", checkoutID)

	return &CheckoutResponse{
		CheckoutURL: fakeURL,
		ID:          checkoutID,
	}, nil
}

func (s *service) GetSubscriptionStatus(ctx context.Context, userID uuid.UUID) (*Subscription, error) {
	return s.repo.GetLatestSubscription(ctx, userID)
}

func (s *service) HandleWebhook(ctx context.Context, externalID string, payload map[string]interface{}) error {
	// Logic to mark subscription as active once payment is confirmed.
	return nil
}

func (s *service) CleanupExpiredSubscriptions(ctx context.Context) error {
	// Downgrade users who exceed grace period.
	return nil
}
