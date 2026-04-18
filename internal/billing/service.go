package billing

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/notification"
	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/google/uuid"
)

var ErrDOKUNotConfigured = errors.New("payment gateway is not configured: DOKU_CLIENT_ID or DOKU_SECRET_KEY is empty")

const dokuCheckoutPath = "/checkout/v1/payment"
const dokuOrderStatusPathPrefix = "/orders/v1/status/"

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
	repo         Repository
	notifSvc     notification.Service
	userSvc      user.Service
	httpClient   *http.Client
	appBaseURL   string
	dokuClientID string
	dokuSecret   string
	dokuBaseURL  string
}

// NewService creates billing service with DOKU Checkout API configuration.
func NewService(repo Repository, notifSvc notification.Service, userSvc user.Service, clientID, secretKey, baseURL, appBaseURL string) Service {
	trimmedBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")

	return &service{
		repo:         repo,
		notifSvc:     notifSvc,
		userSvc:      userSvc,
		httpClient:   &http.Client{Timeout: 20 * time.Second},
		appBaseURL:   appBaseURL,
		dokuClientID: strings.TrimSpace(clientID),
		dokuSecret:   strings.TrimSpace(secretKey),
		dokuBaseURL:  trimmedBaseURL,
	}
}

// requireDOKUConfig validates payment gateway readiness before API calls.
func (s *service) requireDOKUConfig() error {
	if s.dokuClientID == "" || s.dokuSecret == "" {
		return ErrDOKUNotConfigured
	}
	if s.dokuBaseURL == "" {
		return errors.New("payment gateway is not configured: DOKU_BASE_URL is empty")
	}

	return nil
}

// InitiateCheckout creates one DOKU checkout transaction and stores pending subscription.
func (s *service) InitiateCheckout(ctx context.Context, userID uuid.UUID, req CheckoutRequest) (*CheckoutResponse, error) {
	if err := s.requireDOKUConfig(); err != nil {
		return nil, err
	}

	// 1. Determine price
	amount := int64(99000) // Default Pro Monthly
	plan := strings.ToLower(req.PlanTier)
	if plan == "enterprise" {
		amount = 499000
	}

	orderID := fmt.Sprintf("SUB-%s-%d", userID.String()[:8], time.Now().Unix())

	orderPayload := map[string]any{
		"amount":         amount,
		"invoice_number": orderID,
		"currency":       "IDR",
	}

	if shouldSendDOKUCallbackURL(s.appBaseURL) {
		orderPayload["callback_url"] = s.appBaseURL + "/dashboard?payment=success"
		orderPayload["callback_url_cancel"] = s.appBaseURL + "/dashboard?payment=cancel"
		orderPayload["callback_url_result"] = s.appBaseURL + "/dashboard?payment=success"
		orderPayload["auto_redirect"] = true
	}

	dokuReq := map[string]any{
		"order": orderPayload,
		"payment": map[string]any{
			"payment_due_date": 60,
		},
	}

	bodyBytes, err := json.Marshal(dokuReq)
	if err != nil {
		return nil, fmt.Errorf("failed to build doku checkout request: %w", err)
	}

	requestID := uuid.NewString()
	requestTimestamp := time.Now().UTC().Format(time.RFC3339)
	digest := generateDigest(bodyBytes)
	signature := generateSignature(
		s.dokuClientID,
		requestID,
		requestTimestamp,
		dokuCheckoutPath,
		digest,
		s.dokuSecret,
	)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.dokuBaseURL+dokuCheckoutPath, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create doku request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Client-Id", s.dokuClientID)
	httpReq.Header.Set("Request-Id", requestID)
	httpReq.Header.Set("Request-Timestamp", requestTimestamp)
	httpReq.Header.Set("Signature", signature)

	httpResp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create doku transaction: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read doku response: %w", err)
	}

	var dokuResp struct {
		Message       []string `json:"message"`
		ErrorMessages []string `json:"error_messages"`
		Response      struct {
			Payment struct {
				URL string `json:"url"`
			} `json:"payment"`
		} `json:"response"`
	}
	if err := json.Unmarshal(respBody, &dokuResp); err != nil {
		return nil, fmt.Errorf("failed to parse doku response: %w", err)
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to create doku transaction: status %d, details: %s", httpResp.StatusCode, extractDOKUErrorDetails(dokuResp.Message, dokuResp.ErrorMessages, respBody))
	}

	checkoutURL := strings.TrimSpace(dokuResp.Response.Payment.URL)
	if checkoutURL == "" {
		return nil, errors.New("failed to create doku transaction: missing payment url")
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
		CheckoutURL: checkoutURL,
		ID:          orderID,
	}, nil
}

// GetSubscriptionStatus returns latest subscription and optionally syncs pending payment status.
func (s *service) GetSubscriptionStatus(ctx context.Context, userID uuid.UUID) (*Subscription, error) {
	sub, err := s.repo.GetLatestSubscription(ctx, userID)
	if err != nil {
		return nil, err
	}

	if sub != nil && sub.Status == "pending" {
		if err := s.syncPendingSubscriptionStatus(ctx, sub); err != nil {
			// Keep endpoint responsive even if upstream payment status check fails.
			log.Printf("billing: failed to sync pending subscription %s status: %v", sub.ExternalCheckoutID, err)
			return sub, nil
		}
	}

	return sub, nil
}

// syncPendingSubscriptionStatus checks DOKU order status and activates subscription on success.
func (s *service) syncPendingSubscriptionStatus(ctx context.Context, sub *Subscription) error {
	if err := s.requireDOKUConfig(); err != nil {
		return err
	}

	requestID := uuid.NewString()
	requestTimestamp := time.Now().UTC().Format(time.RFC3339)
	requestTarget := dokuOrderStatusPathPrefix + url.PathEscape(sub.ExternalCheckoutID)
	signature := generateSignatureForGet(s.dokuClientID, requestID, requestTimestamp, requestTarget, s.dokuSecret)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, s.dokuBaseURL+requestTarget, nil)
	if err != nil {
		return fmt.Errorf("failed to create doku status request: %w", err)
	}
	httpReq.Header.Set("Client-Id", s.dokuClientID)
	httpReq.Header.Set("Request-Id", requestID)
	httpReq.Header.Set("Request-Timestamp", requestTimestamp)
	httpReq.Header.Set("Signature", signature)

	httpResp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to fetch doku order status: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read doku order status response: %w", err)
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return fmt.Errorf("failed to fetch doku order status: status %d, body: %s", httpResp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var statusPayload map[string]interface{}
	if err := json.Unmarshal(respBody, &statusPayload); err != nil {
		return fmt.Errorf("failed to parse doku order status response: %w", err)
	}

	if !isSuccessfulPayment(flattenStatusPayload(statusPayload)) {
		log.Printf("billing: pending subscription %s still unpaid or status not recognized", sub.ExternalCheckoutID)
		return nil
	}

	log.Printf("billing: pending subscription %s reported SUCCESS by DOKU, activating", sub.ExternalCheckoutID)
	return s.activateSubscription(ctx, sub)
}

// HandleWebhook validates payment result and activates subscription when paid.
func (s *service) HandleWebhook(ctx context.Context, externalID string, payload map[string]interface{}) error {
	if err := s.requireDOKUConfig(); err != nil {
		return err
	}

	if !isSuccessfulPayment(payload) {
		return nil
	}

	sub, err := s.repo.GetSubscriptionByExternalID(ctx, externalID)
	if err != nil {
		return err
	}
	if sub == nil {
		return fmt.Errorf("subscription not found for order id %s", externalID)
	}

	return s.activateSubscription(ctx, sub)
}

// activateSubscription marks subscription active and syncs tier + confirmation email.
func (s *service) activateSubscription(ctx context.Context, sub *Subscription) error {
	if sub.Status == "active" {
		log.Printf("billing: subscription %s already active, skip activation", sub.ExternalCheckoutID)
		return nil
	}

	sub.Status = "active"
	sub.LastPaymentAt = ptrTime(time.Now())
	sub.ExpiresAt = time.Now().AddDate(0, 1, 0)
	sub.GracePeriodUntil = ptrTime(sub.ExpiresAt.AddDate(0, 0, 7))
	if err := s.repo.UpdateSubscription(ctx, sub); err != nil {
		return err
	}

	if err := s.notifSvc.SetPlanTier(sub.UserID, sub.PlanTier, &sub.ExpiresAt); err != nil {
		return fmt.Errorf("failed to sync tier with notification service: %w", err)
	}

	if u, err := s.userSvc.GetUserByID(sub.UserID); err == nil {
		s.notifSvc.SendUpgradeConfirmationEmail(u.Email, u.Name, sub.PlanTier, sub.ExpiresAt)
	}

	log.Printf("billing: subscription %s activated with tier %s", sub.ExternalCheckoutID, sub.PlanTier)

	return nil
}

// CleanupExpiredSubscriptions downgrades subscriptions past grace period.
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

// NotifyExpiringSubscriptions sends reminder email for subscriptions expiring in 7 days.
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

// CancelSubscription marks active subscription as cancelled and resets user tier.
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

// generateDigest returns base64 SHA-256 digest from request body bytes.
func generateDigest(body []byte) string {
	sum := sha256.Sum256(body)
	return base64.StdEncoding.EncodeToString(sum[:])
}

// generateSignature creates DOKU non-SNAP HMACSHA256 signature.
func generateSignature(clientID, requestID, timestamp, requestTarget, digest, secret string) string {
	component := fmt.Sprintf("Client-Id:%s\nRequest-Id:%s\nRequest-Timestamp:%s\nRequest-Target:%s\nDigest:%s", clientID, requestID, timestamp, requestTarget, digest)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(component))
	encoded := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return "HMACSHA256=" + encoded
}

// generateSignatureForGet creates DOKU non-SNAP signature for GET request (no Digest line).
func generateSignatureForGet(clientID, requestID, timestamp, requestTarget, secret string) string {
	component := fmt.Sprintf("Client-Id:%s\nRequest-Id:%s\nRequest-Timestamp:%s\nRequest-Target:%s", clientID, requestID, timestamp, requestTarget)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(component))
	encoded := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return "HMACSHA256=" + encoded
}

// flattenStatusPayload normalizes common DOKU status response locations into one map for status checks.
func flattenStatusPayload(payload map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for k, v := range payload {
		result[k] = v
	}

	if responseMap, ok := payload["response"].(map[string]interface{}); ok {
		for k, v := range responseMap {
			if _, exists := result[k]; !exists {
				result[k] = v
			}
		}

		if txMap, ok := responseMap["transaction"].(map[string]interface{}); ok {
			result["transaction"] = txMap
		}
		if paymentMap, ok := responseMap["payment"].(map[string]interface{}); ok {
			result["payment"] = paymentMap
		}
	}

	return result
}

// isSuccessfulPayment checks common DOKU and generic payment success markers.
func isSuccessfulPayment(payload map[string]interface{}) bool {
	candidates := []string{
		strings.ToUpper(extractString(payload, "transaction", "status")),
		strings.ToUpper(extractString(payload, "transaction_status")),
		strings.ToUpper(extractString(payload, "status")),
		strings.ToUpper(extractString(payload, "latest_transaction_status")),
		strings.ToUpper(extractString(payload, "payment", "status")),
	}
	for _, status := range candidates {
		switch status {
		case "SUCCESS", "PAID", "SETTLEMENT", "CAPTURE", "COMPLETED":
			return true
		}
	}
	return false
}

// extractString safely gets nested string value from map payload.
func extractString(payload map[string]interface{}, keys ...string) string {
	var current interface{} = payload
	for _, key := range keys {
		asMap, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}
		current = asMap[key]
	}
	asString, ok := current.(string)
	if !ok {
		return ""
	}
	return asString
}

// shouldSendDOKUCallbackURL ensures callback URL is HTTPS and valid before sending to DOKU.
func shouldSendDOKUCallbackURL(appBaseURL string) bool {
	u, err := url.Parse(strings.TrimSpace(appBaseURL))
	if err != nil || u == nil {
		return false
	}
	if u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}

// extractDOKUErrorDetails returns best-effort error detail from DOKU response.
func extractDOKUErrorDetails(messages, errorMessages []string, rawBody []byte) string {
	if len(errorMessages) > 0 {
		return strings.Join(errorMessages, ", ")
	}
	if len(messages) > 0 {
		return strings.Join(messages, ", ")
	}
	body := strings.TrimSpace(string(rawBody))
	if body == "" {
		return "upstream returned empty error body"
	}
	if len(body) > 500 {
		return body[:500] + "..."
	}
	return body
}

// ptrTime returns pointer value for a time object.
func ptrTime(t time.Time) *time.Time { return &t }
