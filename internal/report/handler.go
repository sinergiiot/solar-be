package report

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/auth"
	"github.com/akbarsenawijaya/solar-forecast/internal/middleware"
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
	r.With(middleware.RequireFeature("green_report")).Get("/report/energy", h.GetEnergyReport)
	r.With(middleware.RequireFeature("green_report")).Get("/report/energy/pdf", h.DownloadEnergyReportPDF)
	r.With(middleware.RequireFeature("green_report")).Get("/report/rec", h.GetRECReadinessReport)
	r.With(middleware.RequireFeature("green_report")).Get("/report/rec/pdf", h.DownloadRECPDF)
	r.With(middleware.RequireFeature("enterprise")).Get("/report/esg", h.GetESGSummary)
	r.With(middleware.RequireFeature("enterprise")).Get("/report/esg/pdf", h.DownloadESGReportPDF)
	r.With(middleware.RequireFeature("green_report")).Get("/report/history/csv", h.DownloadHistoryCSV)
	// Epic 4: CO2 Tracker
	r.With(middleware.RequireFeature("green_report")).Get("/report/co2", h.GetCO2Summary)
	r.With(middleware.RequireFeature("green_report")).Get("/report/co2/pdf", h.DownloadMRVPDF)
}

func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	r.Get("/public/esg/{token}", h.GetPublicESGSummary)
}

func (h *Handler) GetEnergyReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	planTier := middleware.GetTierFromContext(r.Context())

	isAnnual := r.URL.Query().Get("is_annual") == "true"
	yearStr := r.URL.Query().Get("year")
	year := 0
	if yearStr != "" {
		fmt.Sscanf(yearStr, "%d", &year)
	}

	req := ReportRequest{
		SolarProfileID: r.URL.Query().Get("profile_id"),
		StartDate:      r.URL.Query().Get("start_date"),
		EndDate:        r.URL.Query().Get("end_date"),
		IsAnnual:       isAnnual,
		Year:           year,
		OfficialLetter: r.URL.Query().Get("official_letter") == "true",
		Signatory:      r.URL.Query().Get("signatory"),
		Title:          r.URL.Query().Get("title"),
		Organization:   r.URL.Query().Get("organization"),
	}

	if !isAnnual {
		if req.StartDate == "" {
			req.StartDate = time.Now().UTC().AddDate(0, 0, -30).Format(time.DateOnly)
		}
		if req.EndDate == "" {
			req.EndDate = time.Now().UTC().Format(time.DateOnly)
		}
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

	planTier := middleware.GetTierFromContext(r.Context())

	userObj, err := h.userSvc.GetUserByID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get user info")
		return
	}

	isAnnual := r.URL.Query().Get("is_annual") == "true"
	yearStr := r.URL.Query().Get("year")
	year := 0
	if yearStr != "" {
		fmt.Sscanf(yearStr, "%d", &year)
	}

	req := ReportRequest{
		SolarProfileID: r.URL.Query().Get("profile_id"),
		StartDate:      r.URL.Query().Get("start_date"),
		EndDate:        r.URL.Query().Get("end_date"),
		IsAnnual:       isAnnual,
		Year:           year,
		OfficialLetter: r.URL.Query().Get("official_letter") == "true",
		Signatory:      r.URL.Query().Get("signatory"),
		Title:          r.URL.Query().Get("title"),
		Organization:   r.URL.Query().Get("organization"),
	}

	if !isAnnual {
		if req.StartDate == "" {
			req.StartDate = time.Now().UTC().AddDate(0, 0, -30).Format(time.DateOnly)
		}
		if req.EndDate == "" {
			req.EndDate = time.Now().UTC().Format(time.DateOnly)
		}
	}

	report, err := h.service.GenerateReport(r.Context(), userID, planTier, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	
	filename := "EnergyReport"
	if isAnnual {
		filename = fmt.Sprintf("EnergyReport_Annual_%d", year)
		if req.OfficialLetter {
			filename = fmt.Sprintf("PBB_SuratKeterangan_%d", year)
		}
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pdf", filename))

	if err := h.service.GenerateReportPDF(report, userObj, w); err != nil {
		log.Printf("pdf generation failed: %v", err)
	}
}

func (h *Handler) DownloadRECPDF(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	reportType := r.URL.Query().Get("type") // "certificate" or "report"

	w.Header().Set("Content-Type", "application/pdf")
	
	if reportType == "certificate" {
		filename := fmt.Sprintf("REC_Certificate_%s", time.Now().Format("20060102"))
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pdf", filename))
		if err := h.service.GenerateRECPDF(r.Context(), userID, w); err != nil {
			log.Printf("rec certificate generation failed: %v", err)
		}
		return
	}

	// Default: REC Readiness Report
	report, err := h.service.GetRECReadinessReport(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	userObj, err := h.userSvc.GetUserByID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	filename := fmt.Sprintf("REC_Readiness_Report_%s", time.Now().Format("20060102"))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.pdf", filename))

	if err := h.service.GenerateRECReadinessReportPDF(report, userObj, w); err != nil {
		log.Printf("rec readiness report generation failed: %v", err)
	}
}

func (h *Handler) GetRECReadinessReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	report, err := h.service.GetRECReadinessReport(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func (h *Handler) GetESGSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	planTier := middleware.GetTierFromContext(r.Context())
	yearStr := r.URL.Query().Get("year")
	year := 0
	if yearStr != "" {
		fmt.Sscanf(yearStr, "%d", &year)
	}

	summary, err := h.service.GetESGSummary(r.Context(), userID, planTier, year)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// GetCO2Summary handles GET /report/co2 - Epic 4: CO2 Avoided Tracker
// DownloadESGReportPDF handles GET /report/esg/pdf
func (h *Handler) DownloadESGReportPDF(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	yearStr := r.URL.Query().Get("year")
	year := 0
	if yearStr != "" {
		fmt.Sscanf(yearStr, "%d", &year)
	}

	// 1. Get Summary
	summary, err := h.service.GetESGSummary(r.Context(), userID, "enterprise", year)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 2. Get User object for branding
	userObj, err := h.userSvc.GetUserByID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get user info")
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	filename := fmt.Sprintf("ESG_Report_%d.pdf", year)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	if err := h.service.GenerateESGReportPDF(summary, userObj, year, w); err != nil {
		log.Printf("esg pdf generation failed: %v", err)
	}
}

// DownloadHistoryCSV handles GET /report/history/csv
func (h *Handler) DownloadHistoryCSV(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	req := ReportRequest{
		SolarProfileID: r.URL.Query().Get("profile_id"),
		StartDate:      r.URL.Query().Get("start_date"),
		EndDate:        r.URL.Query().Get("end_date"),
	}

	planTier := r.URL.Query().Get("tier") // Fallback if context not enough
	if planTier == "" {
		planTier = "pro" // Assume pro if they reached here? Or check DB
	}

	w.Header().Set("Content-Type", "text/csv")
	filename := fmt.Sprintf("History_Export_%s-%s.csv", req.StartDate, req.EndDate)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	if err := h.service.GenerateCSVHistory(r.Context(), userID, planTier, req, w); err != nil {
		log.Printf("csv history export failed: %v", err)
	}
}

func (h *Handler) GetCO2Summary(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	planTier := middleware.GetTierFromContext(r.Context())

	isAnnual := r.URL.Query().Get("is_annual") == "true"
	yearStr := r.URL.Query().Get("year")
	year := 0
	if yearStr != "" {
		fmt.Sscanf(yearStr, "%d", &year)
	}

	req := ReportRequest{
		SolarProfileID: r.URL.Query().Get("profile_id"),
		IsAnnual:       isAnnual,
		Year:           year,
		StartDate:      r.URL.Query().Get("start_date"),
		EndDate:        r.URL.Query().Get("end_date"),
	}

	if !isAnnual {
		if req.StartDate == "" {
			req.StartDate = time.Now().UTC().AddDate(0, 0, -30).Format(time.DateOnly)
		}
		if req.EndDate == "" {
			req.EndDate = time.Now().UTC().Format(time.DateOnly)
		}
	}

	summary, err := h.service.GetCO2Summary(r.Context(), userID, planTier, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// DownloadMRVPDF handles GET /report/co2/pdf - Epic 4: MRV PDF Download
func (h *Handler) DownloadMRVPDF(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	planTier := middleware.GetTierFromContext(r.Context())

	userObj, err := h.userSvc.GetUserByID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get user info")
		return
	}

	isAnnual := r.URL.Query().Get("is_annual") == "true"
	yearStr := r.URL.Query().Get("year")
	year := 0
	if yearStr != "" {
		fmt.Sscanf(yearStr, "%d", &year)
	}

	req := ReportRequest{
		SolarProfileID: r.URL.Query().Get("profile_id"),
		IsAnnual:       isAnnual,
		Year:           year,
		StartDate:      r.URL.Query().Get("start_date"),
		EndDate:        r.URL.Query().Get("end_date"),
	}

	if !isAnnual {
		if req.StartDate == "" {
			req.StartDate = time.Now().UTC().AddDate(0, 0, -30).Format(time.DateOnly)
		}
		if req.EndDate == "" {
			req.EndDate = time.Now().UTC().Format(time.DateOnly)
		}
	}

	summary, err := h.service.GetCO2Summary(r.Context(), userID, planTier, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	filename := fmt.Sprintf("MRV_CO2_Report_%s.pdf", time.Now().Format("20060102"))
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)

	if err := h.service.GenerateMRVPDF(summary, userObj, w); err != nil {
		log.Printf("mrv pdf generation failed: %v", err)
	}
}

// GetPublicESGSummary handles GET /public/esg/{token}
func (h *Handler) GetPublicESGSummary(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "token required")
		return
	}

	u, err := h.userSvc.GetUserByESGShareToken(token)
	if err != nil {
		writeError(w, http.StatusNotFound, "shared report not found or disabled")
		return
	}

	yearStr := r.URL.Query().Get("year")
	year := 0
	if yearStr != "" {
		fmt.Sscanf(yearStr, "%d", &year)
	}

	// We use "enterprise" tier logic for the summary itself to bypass checks in service
	// which usually check the requestor's tier.
	summary, err := h.service.GetESGSummary(r.Context(), u.ID, "enterprise", year)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return summary + company branding info
	writeJSON(w, http.StatusOK, map[string]any{
		"summary":      summary,
		"company_name": u.CompanyName,
		"company_logo": u.CompanyLogoURL,
	})
}
