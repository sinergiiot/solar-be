package report

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/auth"
	"github.com/akbarsenawijaya/solar-forecast/internal/tier"
	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Service
	userSvc user.Service
}

func NewHandler(service Service, userSvc user.Service) *Handler {
	return &Handler{service: service, userSvc: userSvc}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	// Gated for Pro/Enterprise users
	r.With(tier.RequireFeature("green_report")).Get("/report/energy", h.GetEnergyReport)
	r.With(tier.RequireFeature("green_report")).Get("/report/energy/pdf", h.DownloadEnergyReportPDF)
}

func (h *Handler) GetEnergyReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	planTier := tier.GetTierFromContext(r.Context())

	req := ReportRequest{
		SolarProfileID: r.URL.Query().Get("profile_id"),
		StartDate:      r.URL.Query().Get("start_date"),
		EndDate:        r.URL.Query().Get("end_date"),
	}

	// Default to last 30 days if not provided
	if req.StartDate == "" {
		req.StartDate = time.Now().UTC().AddDate(0, 0, -30).Format(time.DateOnly)
	}
	if req.EndDate == "" {
		req.EndDate = time.Now().UTC().Format(time.DateOnly)
	}

	report, err := h.service.GenerateReport(r.Context(), userID, planTier, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func (h *Handler) DownloadEnergyReportPDF(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	planTier := tier.GetTierFromContext(r.Context())

	userObj, err := h.userSvc.GetUserByID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get user info")
		return
	}

	req := ReportRequest{
		SolarProfileID: r.URL.Query().Get("profile_id"),
		StartDate:      r.URL.Query().Get("start_date"),
		EndDate:        r.URL.Query().Get("end_date"),
	}

	// Default to last 30 days if not provided
	if req.StartDate == "" {
		req.StartDate = time.Now().UTC().AddDate(0, 0, -30).Format(time.DateOnly)
	}
	if req.EndDate == "" {
		req.EndDate = time.Now().UTC().Format(time.DateOnly)
	}

	report, err := h.service.GenerateReport(r.Context(), userID, planTier, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=EnergyReport_%s.pdf", time.Now().Format("20060102")))

	if err := h.service.GenerateReportPDF(report, userObj.Name, w); err != nil {
		// Since we already set headers, we can't easily writeError here if it's already started
		// but for PDF generation it usually happens all at once
		log.Printf("pdf generation failed: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
