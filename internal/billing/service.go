package billing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/notification"
	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
)

type Repository interface {
	CreateSubscription(ctx context.Context, sub *Subscription) error
	GetLatestSubscription(ctx context.Context, userID uuid.UUID) (*Subscription, error)
	UpdateSubscription(ctx context.Context, sub *Subscription) error
	GetSubscriptionByExternalID(ctx context.Context, extID string) (*Subscription, error)
	// For scheduler/cleanup
	GetPastDueSubscriptions(ctx context.Context, now time.Time) ([]*Subscription, error)
	GetExpiringSubscriptions(ctx context.Context, start, end time.Time) ([]*Subscription, error)
}

type Service interface {
	InitiateCheckout(ctx context.Context, userID uuid.UUID, req CheckoutRequest) (*CheckoutResponse, error)
	HandleWebhook(ctx context.Context, externalID string, payload map[string]interface{}) error
	GetSubscriptionStatus(ctx context.Context, userID uuid.UUID) (*Subscription, error)
	CleanupExpiredSubscriptions(ctx context.Context) error
	NotifyExpiringSubscriptions(ctx context.Context) error
	CancelSubscription(ctx context.Context, userID uuid.UUID) error
}

type service struct {
	repo       Repository
	notifSvc   notification.Service
	userSvc    user.Service
	snapClient snap.Client
	coreClient coreapi.Client
	appBaseURL string
}

func NewService(repo Repository, notifSvc notification.Service, userSvc user.Service, serverKey string, isProd bool, appBaseURL string) Service {
	sc := snap.Client{}
	env := midtrans.Sandbox
	if isProd {
		env = midtrans.Production
	}
	sc.New(serverKey, env)

	cc := coreapi.Client{}
	cc.New(serverKey, env)

	return &service{
		repo:       repo,
		notifSvc:   notifSvc,
		userSvc:    userSvc,
		snapClient: sc,
		coreClient: cc,
		appBaseURL: appBaseURL,
	}
}

func (s *service) InitiateCheckout(ctx context.Context, userID uuid.UUID, req CheckoutRequest) (*CheckoutResponse, error) {
	// 1. Determine price
	amount := int64(99000) // Default Pro Monthly
	plan := strings.ToLower(req.PlanTier)
	if plan == "enterprise" {
		amount = 499000
	}

	orderID := fmt.Sprintf("SUB-%s-%d", userID.String()[:8], time.Now().Unix())

	// 2. Create Midtrans Snap request
	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: amount,
		},
		EnabledPayments: []snap.SnapPaymentType{
			snap.PaymentTypeGopay, snap.PaymentTypeShopeepay, snap.PaymentTypeCreditCard,
			snap.PaymentTypeBankTransfer, snap.PaymentTypeOtherVA,
		},
		Items: &[]midtrans.ItemDetails{
			{
				ID:    plan,
				Price: amount,
				Qty:   1,
				Name:  fmt.Sprintf("Solar Forecast %s Plan", strings.Title(plan)),
			},
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: "User", // Ideally fetch name/email from user service
		},
		Callbacks: &snap.Callbacks{
			Finish: s.appBaseURL + "/dashboard?payment=success",
		},
	}

	// 3. Get Snap URL
	snapResp, err := s.snapClient.CreateTransaction(snapReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create midtrans transaction: %w", err)
	}

	// 4. Record pending subscription in DB
	sub := &Subscription{
		ID:                 uuid.New(),
		UserID:             userID,
		PlanTier:           plan,
		Status:             "pending", // Wait for webhook to activate
		BillingCycle:       req.BillingCycle,
		Amount:             amount,
		Currency:           "IDR",
		ExternalCheckoutID: orderID,
		ExpiresAt:          time.Now().AddDate(0, 1, 0), // Default 1 month
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		return nil, err
	}

	return &CheckoutResponse{
		CheckoutURL: snapResp.RedirectURL,
		ID:          orderID,
	}, nil
}

func (s *service) GetSubscriptionStatus(ctx context.Context, userID uuid.UUID) (*Subscription, error) {
	sub, err := s.repo.GetLatestSubscription(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Local development fallback: if pending, check Midtrans directly
	// because webhooks cannot reach localhost.
	if sub != nil && sub.Status == "pending" {
		status, err := s.coreClient.CheckTransaction(sub.ExternalCheckoutID)
		if err == nil {
			if status.TransactionStatus == "settlement" || status.TransactionStatus == "capture" {
				// Activate now!
				sub.Status = "active"
				sub.LastPaymentAt = ptrTime(time.Now())
				sub.ExpiresAt = time.Now().AddDate(0, 1, 0)
				sub.GracePeriodUntil = ptrTime(sub.ExpiresAt.AddDate(0, 0, 7))
				s.repo.UpdateSubscription(ctx, sub)
				s.notifSvc.SetPlanTier(sub.UserID, sub.PlanTier, &sub.ExpiresAt)

				// Send confirmation email
				if u, err := s.userSvc.GetUserByID(sub.UserID); err == nil {
					s.notifSvc.SendUpgradeConfirmationEmail(u.Email, u.Name, sub.PlanTier, sub.ExpiresAt)
				}
			}
		}
	}

	return sub, nil
}

func (s *service) HandleWebhook(ctx context.Context, externalID string, payload map[string]interface{}) error {
	// SECURE: Do not trust the payload from public internet.
	// Instead, ask Midtrans directly about this OrderID using the server key.
	statusResp, mErr := s.coreClient.CheckTransaction(externalID)
	if mErr != nil {
		return fmt.Errorf("failed to verify transaction with midtrans: %w", mErr)
	}

	status := statusResp.TransactionStatus
	if status != "settlement" && status != "capture" {
		return nil
	}

	sub, err := s.repo.GetSubscriptionByExternalID(ctx, externalID)
	if err != nil {
		return err
	}
	if sub == nil {
		return fmt.Errorf("subscription not found for order id %s", externalID)
	}

	// Activate subscription
	sub.Status = "active"
	sub.LastPaymentAt = ptrTime(time.Now())
	sub.ExpiresAt = time.Now().AddDate(0, 1, 0) // Align with billing cycle
	sub.GracePeriodUntil = ptrTime(sub.ExpiresAt.AddDate(0, 0, 7))
	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		return err
	}

	// Propagate tier change to notification/tier engine
	if err := s.notifSvc.SetPlanTier(sub.UserID, sub.PlanTier, &sub.ExpiresAt); err != nil {
		return fmt.Errorf("failed to sync tier with notification service: %w", err)
	}

	// Send confirmation email
	if u, err := s.userSvc.GetUserByID(sub.UserID); err == nil {
		s.notifSvc.SendUpgradeConfirmationEmail(u.Email, u.Name, sub.PlanTier, sub.ExpiresAt)
	}

	return nil
}

func (s *service) CleanupExpiredSubscriptions(ctx context.Context) error {
	now := time.Now()
	expired, err := s.repo.GetPastDueSubscriptions(ctx, now)
	if err != nil {
		return err
	}

	for _, sub := range expired {
		sub.Status = "expired"
		if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
			continue
		}
		// Reset to free
		s.notifSvc.SetPlanTier(sub.UserID, "free", nil)
	}

	return nil
}

func (s *service) NotifyExpiringSubscriptions(ctx context.Context) error {
	// Notify users whose subscription expires in exactly 7 days
	targetDate := time.Now().AddDate(0, 0, 7)
	start := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)

	expiring, err := s.repo.GetExpiringSubscriptions(ctx, start, end)
	if err != nil {
		return err
	}

	for _, sub := range expiring {
		u, err := s.userSvc.GetUserByID(sub.UserID)
		if err != nil {
			continue
		}
		s.notifSvc.SendSubscriptionExpiringEmail(u.Email, u.Name, sub.PlanTier, sub.ExpiresAt)
	}

	return nil
}

func (s *service) CancelSubscription(ctx context.Context, userID uuid.UUID) error {
	sub, err := s.repo.GetLatestSubscription(ctx, userID)
	if err != nil {
		return err
	}
	if sub == nil || sub.Status != "active" {
		return fmt.Errorf("no active subscription found")
	}

	sub.Status = "cancelled"
	sub.UpdatedAt = time.Now()
	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		return err
	}

	// Reset tier to free immediately
	return s.notifSvc.SetPlanTier(userID, "free", nil)
}

func ptrTime(t time.Time) *time.Time { return &t }
