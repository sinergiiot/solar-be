package solar

import (
	"encoding/json"
	"net/http"

	"github.com/akbarsenawijaya/solar-forecast/internal/auth"
	"github.com/akbarsenawijaya/solar-forecast/internal/tier"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the solar profile domain
type Handler struct {
	service Service
}

// NewHandler creates a new solar profile HTTP handler
func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

// RegisterRoutes wires all solar profile endpoints onto the given router
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/solar-profiles", h.CreateSolarProfile)
	r.Get("/solar-profiles", h.ListMySolarProfiles)
	r.Get("/solar-profiles/{profileID}", h.GetMySolarProfileByID)
	r.Get("/solar-profiles/me", h.GetMySolarProfile)
	r.Put("/solar-profiles/{profileID}", h.UpdateSolarProfile)
	r.Delete("/solar-profiles/{profileID}", h.DeleteSolarProfile)
}

// CreateSolarProfile handles POST /solar-profiles
func (h *Handler) CreateSolarProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateSolarProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.UserID = userID
	req.PlanTier = tier.GetTierFromContext(r.Context())

	profile, err := h.service.CreateSolarProfile(r.Context(), req)
	if err != nil {
		if limitErr, ok := err.(*tier.LimitError); ok {
			writeJSON(w, http.StatusForbidden, limitErr)
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, profile)
}

// GetMySolarProfile handles GET /solar-profiles/me for authenticated user.
func (h *Handler) GetMySolarProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profile, err := h.service.GetSolarProfileByUserID(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "solar profile not found")
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

// ListMySolarProfiles handles GET /solar-profiles for authenticated user.
func (h *Handler) ListMySolarProfiles(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profiles, err := h.service.GetSolarProfilesByUserID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load solar profiles")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"profiles": profiles,
		"count":    len(profiles),
	})
}

// GetMySolarProfileByID handles GET /solar-profiles/{profileID} for authenticated user.
func (h *Handler) GetMySolarProfileByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profileID, err := uuid.Parse(chi.URLParam(r, "profileID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid profile id")
		return
	}

	profile, err := h.service.GetSolarProfileByIDAndUserID(profileID, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "solar profile not found")
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// UpdateSolarProfile handles PUT /solar-profiles/{profileID}.
func (h *Handler) UpdateSolarProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profileID, err := uuid.Parse(chi.URLParam(r, "profileID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid profile id")
		return
	}

	var req UpdateSolarProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.UserID = userID

	updated, err := h.service.UpdateSolarProfile(profileID, req)
	if err != nil {
		if err.Error() == "solar profile not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// DeleteSolarProfile handles DELETE /solar-profiles/{profileID}.
func (h *Handler) DeleteSolarProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profileID, err := uuid.Parse(chi.URLParam(r, "profileID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid profile id")
		return
	}

	if err := h.service.DeleteSolarProfile(profileID, userID); err != nil {
		if err.Error() == "solar profile not found" {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "solar profile deleted"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
