package rec

import (
	"encoding/json"
	"net/http"

	"github.com/akbarsenawijaya/solar-forecast/internal/auth"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/accumulator/rec-readiness", h.GetRECReadiness)
}

type RECReadinessResponse struct {
	TotalMwh       float64 `json:"total_mwh"`
	TotalKwh       float64 `json:"total_kwh"`
	TotalREC       int     `json:"total_rec"`
	ProgressToNext float64 `json:"progress_to_next_pct"`
}

func (h *Handler) GetRECReadiness(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	totalMwh, err := h.service.GetTotalMwhForUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load accumulator data")
		return
	}

	totalREC := int(totalMwh)
	totalKwh := totalMwh * 1000
	progressPct := (totalMwh - float64(totalREC)) * 100

	resp := RECReadinessResponse{
		TotalMwh:       totalMwh,
		TotalKwh:       totalKwh,
		TotalREC:       totalREC,
		ProgressToNext: progressPct,
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
