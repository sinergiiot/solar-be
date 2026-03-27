package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	r.Group(func(r chi.Router) {
		// Midddleware is already applied in main.go for /admin path, 
		// but we double check inside handlers just in case.
		r.Get("/admin/users", h.GetAllUsers)
		r.Put("/admin/users/{id}/tier", h.UpdateTier)
		r.Post("/admin/users/{id}/impersonate", h.ImpersonateUser)
		r.Get("/admin/subscriptions/expiring", h.GetExpiringSubscriptions)
		r.Get("/admin/scheduler/status", h.GetSchedulerStatus)
		r.Get("/admin/stats", h.GetStats)

		// Sprint B: Data Quality
		r.Get("/admin/forecast-quality", h.GetForecastQuality)
		r.Get("/admin/cold-start-monitor", h.GetColdStartSites)
		r.Get("/admin/notification-logs", h.GetNotificationLogs)
		r.Get("/admin/data-anomalies", h.GetDataAnomalies)

		// Sprint C: Business Intelligence
		r.Get("/admin/analytics/aggregate", h.GetAggregateAnalytics)
		r.Get("/admin/analytics/ranking", h.GetSiteRankings)
		r.Get("/admin/analytics/tier-distribution", h.GetTierDistribution)

		// Sprint C Extension
		r.Get("/admin/weather/health", h.GetWeatherHealth)
		r.Get("/admin/audit-logs", h.GetAuditLogs)
	})
}

func (h *Handler) checkAdmin(w http.ResponseWriter, r *http.Request) bool {
	if auth.UserRoleFromContext(r.Context()) != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}
	return true
}

func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if !h.checkAdmin(w, r) {
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
	if !h.checkAdmin(w, r) {
		return
	}
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	var input struct {
		PlanTier string `json:"plan_tier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid input"})
		return
	}

	adminID, _ := auth.UserIDFromContext(r.Context())
	ip := r.RemoteAddr

	err = h.adminSvc.UpdateUserTier(r.Context(), userID, input.PlanTier, adminID, ip)
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	auth.WriteJSON(w, http.StatusOK, map[string]string{"message": "tier updated"})
}

func (h *Handler) GetWeatherHealth(w http.ResponseWriter, r *http.Request) {
	if !h.checkAdmin(w, r) {
		return
	}
	health, err := h.adminSvc.GetWeatherAPIHealth(r.Context())
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	auth.WriteJSON(w, http.StatusOK, health)
}

func (h *Handler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	if !h.checkAdmin(w, r) {
		return
	}
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	logs, err := h.adminSvc.GetAuditLogs(r.Context(), limit)
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	auth.WriteJSON(w, http.StatusOK, logs)
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	if !h.checkAdmin(w, r) {
		return
	}

	stats, err := h.adminSvc.GetSystemStats(r.Context())
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	auth.WriteJSON(w, http.StatusOK, stats)
}

func (h *Handler) ImpersonateUser(w http.ResponseWriter, r *http.Request) {
	if !h.checkAdmin(w, r) {
		return
	}

	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	token, err := h.adminSvc.GenerateImpersonationToken(r.Context(), userID)
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	auth.WriteJSON(w, http.StatusOK, map[string]string{
		"impersonation_token": token,
	})
}

func (h *Handler) GetExpiringSubscriptions(w http.ResponseWriter, r *http.Request) {
	if !h.checkAdmin(w, r) {
		return
	}

	daysStr := r.URL.Query().Get("days")
	days := 7
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil {
			days = d
		}
	}

	subs, err := h.adminSvc.GetExpiringSubscriptions(r.Context(), days)
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	auth.WriteJSON(w, http.StatusOK, subs)
}

func (h *Handler) GetSchedulerStatus(w http.ResponseWriter, r *http.Request) {
	if !h.checkAdmin(w, r) {
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	runs, err := h.adminSvc.GetSchedulerRuns(r.Context(), limit)
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	auth.WriteJSON(w, http.StatusOK, runs)
}

func (h *Handler) GetForecastQuality(w http.ResponseWriter, r *http.Request) {
	if !h.checkAdmin(w, r) {
		return
	}
	quality, err := h.adminSvc.GetForecastQuality(r.Context())
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	auth.WriteJSON(w, http.StatusOK, quality)
}

func (h *Handler) GetColdStartSites(w http.ResponseWriter, r *http.Request) {
	if !h.checkAdmin(w, r) {
		return
	}
	sites, err := h.adminSvc.GetColdStartSites(r.Context())
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	auth.WriteJSON(w, http.StatusOK, sites)
}

func (h *Handler) GetNotificationLogs(w http.ResponseWriter, r *http.Request) {
	if !h.checkAdmin(w, r) {
		return
	}
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	logs, err := h.adminSvc.GetNotificationLogs(r.Context(), limit)
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	auth.WriteJSON(w, http.StatusOK, logs)
}

func (h *Handler) GetDataAnomalies(w http.ResponseWriter, r *http.Request) {
	if !h.checkAdmin(w, r) {
		return
	}
	anomalies, err := h.adminSvc.GetDataAnomalies(r.Context())
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	auth.WriteJSON(w, http.StatusOK, anomalies)
}

func (h *Handler) GetAggregateAnalytics(w http.ResponseWriter, r *http.Request) {
	if !h.checkAdmin(w, r) {
		return
	}
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil {
			days = d
		}
	}
	stats, err := h.adminSvc.GetAggregateAnalytics(r.Context(), days)
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	auth.WriteJSON(w, http.StatusOK, stats)
}

func (h *Handler) GetSiteRankings(w http.ResponseWriter, r *http.Request) {
	if !h.checkAdmin(w, r) {
		return
	}
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	rankings, err := h.adminSvc.GetSiteRankings(r.Context(), limit)
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	auth.WriteJSON(w, http.StatusOK, rankings)
}

func (h *Handler) GetTierDistribution(w http.ResponseWriter, r *http.Request) {
	if !h.checkAdmin(w, r) {
		return
	}
	dist, err := h.adminSvc.GetTierDistribution(r.Context())
	if err != nil {
		auth.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	auth.WriteJSON(w, http.StatusOK, dist)
}
