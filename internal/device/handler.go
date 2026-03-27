package device

import (
	"encoding/json"
	"net/http"

	"github.com/akbarsenawijaya/solar-forecast/internal/auth"
	"github.com/akbarsenawijaya/solar-forecast/internal/middleware"
	"github.com/akbarsenawijaya/solar-forecast/internal/tier"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles HTTP operations for device integration.
type Handler struct {
	service Service
}

// NewHandler creates a new device handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterProtectedRoutes wires user-authenticated routes for device management.
func (h *Handler) RegisterProtectedRoutes(r chi.Router) {
	r.Get("/devices", h.ListDevices)
	r.Get("/devices/heartbeat-summary", h.GetHeartbeatSummary)
	r.Post("/devices", h.CreateDevice)
	r.Put("/devices/{deviceID}", h.UpdateDevice)
	r.Delete("/devices/{deviceID}", h.DeleteDevice)
	r.Post("/devices/{deviceID}/rotate-key", h.RotateDeviceKey)
}

// RegisterPublicRoutes wires public ingestion route for field devices.
func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	r.Post("/ingest/telemetry", h.IngestTelemetry)
}

// ListDevices handles GET /devices.
func (h *Handler) ListDevices(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	devices, err := h.service.ListDevices(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"devices": devices, "count": len(devices)})
}

// GetHeartbeatSummary handles GET /devices/heartbeat-summary.
func (h *Handler) GetHeartbeatSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	summary, err := h.service.GetHeartbeatSummary(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// CreateDevice handles POST /devices.
func (h *Handler) CreateDevice(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.PlanTier = middleware.GetTierFromContext(r.Context())
 
 	res, err := h.service.CreateDevice(userID, req)
 	if err != nil {
		if limitErr, ok := err.(*tier.LimitError); ok {
			writeJSON(w, http.StatusForbidden, limitErr)
			return
		}
 		writeError(w, http.StatusBadRequest, err.Error())
 		return
 	}

	writeJSON(w, http.StatusCreated, res)
}

// RotateDeviceKey handles POST /devices/{deviceID}/rotate-key.
func (h *Handler) RotateDeviceKey(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	deviceID, err := uuid.Parse(chi.URLParam(r, "deviceID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid device id")
		return
	}

	res, err := h.service.RotateDeviceKey(userID, deviceID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// UpdateDevice handles PUT /devices/{deviceID}.
func (h *Handler) UpdateDevice(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	deviceID, err := uuid.Parse(chi.URLParam(r, "deviceID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid device id")
		return
	}

	var req UpdateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updated, err := h.service.UpdateDevice(userID, deviceID, req)
	if err != nil {
		if err.Error() == "device not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// DeleteDevice handles DELETE /devices/{deviceID}.
func (h *Handler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	deviceID, err := uuid.Parse(chi.URLParam(r, "deviceID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid device id")
		return
	}

	if err := h.service.DeleteDevice(userID, deviceID); err != nil {
		if err.Error() == "device not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "device deleted"})
}

// IngestTelemetry handles POST /ingest/telemetry.
func (h *Handler) IngestTelemetry(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("X-Device-Key")
	if apiKey == "" {
		writeError(w, http.StatusUnauthorized, "missing X-Device-Key header")
		return
	}

	var req IngestTelemetryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	res, err := h.service.IngestTelemetry(apiKey, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, res)
}

// writeJSON writes one JSON response payload.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// writeError writes one consistent JSON error payload.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
