package billing

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/akbarsenawijaya/solar-forecast/internal/auth"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	r.Post("/billing/webhook", h.Webhook)
}

func (h *Handler) RegisterProtectedRoutes(r chi.Router) {
	r.Post("/billing/checkout", h.Checkout)
	r.Get("/billing/subscription", h.GetSubscription)
	r.Post("/billing/subscription/cancel", h.CancelSubscription)
}

func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	res, err := h.service.InitiateCheckout(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, ErrDOKUNotConfigured) {
			writeError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	sub, err := h.service.GetSubscriptionStatus(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if sub == nil {
		writeJSON(w, http.StatusOK, map[string]string{"plan_tier": "free", "status": "active"})
		return
	}

	writeJSON(w, http.StatusOK, sub)
}

func (h *Handler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.service.CancelSubscription(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "subscription cancelled and tier reset to free"})
}

func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderID := extractWebhookOrderID(payload)
	if orderID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.HandleWebhook(r.Context(), orderID, payload); err != nil {
		if errors.Is(err, ErrDOKUNotConfigured) {
			writeError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// extractWebhookOrderID supports multiple provider payload shapes.
func extractWebhookOrderID(payload map[string]interface{}) string {
	if orderID, ok := payload["order_id"].(string); ok && orderID != "" {
		return orderID
	}
	if invoiceNumber, ok := payload["invoice_number"].(string); ok && invoiceNumber != "" {
		return invoiceNumber
	}

	if orderMap, ok := payload["order"].(map[string]interface{}); ok {
		if invoiceNumber, ok := orderMap["invoice_number"].(string); ok && invoiceNumber != "" {
			return invoiceNumber
		}
		if invoiceNumber, ok := orderMap["invoiceNumber"].(string); ok && invoiceNumber != "" {
			return invoiceNumber
		}
	}

	if txMap, ok := payload["transaction"].(map[string]interface{}); ok {
		if invoiceNumber, ok := txMap["invoice_number"].(string); ok && invoiceNumber != "" {
			return invoiceNumber
		}
		if orderID, ok := txMap["order_id"].(string); ok && orderID != "" {
			return orderID
		}
	}

	return ""
}
