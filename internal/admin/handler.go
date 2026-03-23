package admin

import (
	"encoding/json"
	"net/http"

	"github.com/akbarsenawijaya/solar-forecast/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	adminSvc Service
}

func NewHandler(adminSvc Service) *Handler {
	return &Handler{adminSvc: adminSvc}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/admin/users", h.GetAllUsers)
	r.Put("/admin/users/{id}/tier", h.UpdateTier)
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if auth.UserRoleFromContext(r.Context()) != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	users, err := h.adminSvc.GetAllUsersWithTiers(r.Context())
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	auth.WriteJSON(w, http.StatusOK, users)
}

func (h *Handler) UpdateTier(w http.ResponseWriter, r *http.Request) {
	if auth.UserRoleFromContext(r.Context()) != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	var body struct {
		PlanTier string `json:"plan_tier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	if err := h.adminSvc.UpdateUserTier(r.Context(), userID, body.PlanTier); err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	auth.WriteJSON(w, http.StatusOK, map[string]string{"message": "tier updated"})
}
