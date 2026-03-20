package notification

import (
	"encoding/json"
	"net/http"

	"github.com/akbarsenawijaya/solar-forecast/internal/auth"
	"github.com/go-chi/chi/v5"
)

// Handler serves HTTP endpoints for notification preferences.
type Handler struct {
	service Service
}

// NewHandler creates a notification HTTP handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes wires notification preference routes.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/notifications/preferences", h.GetPreference)
	r.Put("/notifications/preferences", h.UpsertPreference)
}

// GetPreference handles GET /notifications/preferences.
func (h *Handler) GetPreference(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	pref, err := h.service.GetPreference(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, pref)
}

// UpsertPreference handles PUT /notifications/preferences.
func (h *Handler) UpsertPreference(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req UpsertPreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	pref, err := h.service.UpsertPreference(userID, req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, pref)
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
