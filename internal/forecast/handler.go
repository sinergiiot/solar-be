package forecast

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the forecast domain
type Handler struct {
	service    Service
	debugToken string
}

// NewHandler creates a new forecast HTTP handler
func NewHandler(s Service, debugToken string) *Handler {
	return &Handler{service: s, debugToken: strings.TrimSpace(debugToken)}
}

// RegisterRoutes wires all forecast endpoints onto the given router
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/forecast/today", h.GetTodayForecast)
	r.Post("/forecast/actual", h.RecordActualDaily)
	r.Get("/forecast/history", h.GetForecastHistory)
	r.Get("/forecast/actuals/history", h.GetActualHistory)
	r.Get("/forecast/summary", h.GetSummary)
	r.Get("/forecast/debug/calculate", h.GetForecastDebugBreakdown)
}

// GetTodayForecast handles GET /forecast/today for authenticated user.
func (h *Handler) GetTodayForecast(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profileID, err := parseOptionalUUIDQuery(r, "profile_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	f, err := h.service.GetTodayForecastForUser(userID, profileID)
	if err != nil {
		// Return 400 if data missing, 500 for internal error
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "weather data missing") || strings.Contains(err.Error(), "cloud_cover=0") {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, map[string]any{"error": err.Error()})
		return
	}

	deltaWfVal := float64(f.DeltaWF)
	if deltaWfVal == 0 {
		deltaWfVal = float64(f.WeatherFactor)
	}

	bType := f.BaselineType
	if bType == "" {
		bType = "synthetic"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":          f.UserID,
		"solar_profile_id": f.SolarProfileID,
		"date":             f.Date.Format(time.DateOnly),
		"predicted_kwh":    f.PredictedKwh,
		"cloud_cover":      float64(f.CloudCover), // always float64 for JSON
		"cloud_cover_mean": float64(f.CloudCover), // required by frontend
		"weather_factor":   deltaWfVal,
		"transmittance":    deltaWfVal,            // required by frontend
		"delta_wf":         deltaWfVal,
		"baseline_type":    bType,
		"efficiency":       f.Efficiency,
	})
}

// RecordActualDaily handles POST /forecast/actual.
func (h *Handler) RecordActualDaily(w http.ResponseWriter, r *http.Request) {
	var req RecordActualRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	req.UserID = userID

	actual, err := h.service.RecordActualDaily(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":               actual.ID,
		"user_id":          actual.UserID,
		"solar_profile_id": actual.SolarProfileID,
		"date":             actual.Date.Format(time.DateOnly),
		"actual_kwh":       actual.ActualKwh,
		"source":           actual.Source,
	})
}

// GetForecastHistory handles GET /forecast/history for authenticated user.
func (h *Handler) GetForecastHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	days := parseOptionalIntQuery(r, "days", 90)
	filter, err := parseHistoryFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	forecasts, err := h.service.GetForecastHistory(userID, days, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var result []map[string]any
	for _, f := range forecasts {
		deltaWfVal := f.DeltaWF
		if deltaWfVal == 0 {
			deltaWfVal = f.WeatherFactor
		}
		bType := f.BaselineType
		if bType == "" {
			bType = "synthetic"
		}
		result = append(result, map[string]any{
			"date":             f.Date.Format(time.DateOnly),
			"solar_profile_id": f.SolarProfileID,
			"predicted_kwh":    f.PredictedKwh,
			"cloud_cover_mean": f.CloudCover,    // percent (0-100)
			"cloud_cover":      f.CloudCover,    // add just in case
			"weather_factor":   deltaWfVal,      // delta_wf was previously mapped to weather_factor
			"transmittance":    deltaWfVal,      // delta_wf (0.5-1.5), actual weather factor used
			"delta_wf":         deltaWfVal,
			"baseline_type":    bType,           // synthetic/site/blended
			"efficiency":       f.Efficiency,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"forecasts": result,
		"count":     len(result),
	})
}

// GetActualHistory handles GET /forecast/actuals/history for authenticated user.
func (h *Handler) GetActualHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	days := parseOptionalIntQuery(r, "days", 90)
	filter, err := parseHistoryFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	actuals, err := h.service.GetActualHistory(userID, days, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var result []map[string]any
	for _, a := range actuals {
		result = append(result, map[string]any{
			"date":             a.Date.Format(time.DateOnly),
			"solar_profile_id": a.SolarProfileID,
			"actual_kwh":       a.ActualKwh,
			"source":           a.Source,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"actuals": result,
		"count":   len(result),
	})
}

// GetSummary handles GET /forecast/summary for authenticated user.
func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	summary, err := h.service.GetDashboardSummary(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]any{
		"total_forecasted_kwh": summary.TotalForecastedKwh,
		"total_actual_kwh":     summary.TotalActualKwh,
		"average_forecast_kwh": summary.AverageForecastKwh,
		"average_actual_kwh":   summary.AverageActualKwh,
		"current_efficiency":   summary.CurrentEfficiency,
		"accuracy_percent":     summary.AccuracyPercent,
		"forecast_count":       summary.ForecastCount,
		"actual_count":         summary.ActualCount,
	}

	if summary.LastForecastDate != nil {
		response["last_forecast_date"] = summary.LastForecastDate.Format(time.DateOnly)
	}
	if summary.LastActualDate != nil {
		response["last_actual_date"] = summary.LastActualDate.Format(time.DateOnly)
	}

	writeJSON(w, http.StatusOK, response)
}

// GetForecastDebugBreakdown handles GET /forecast/debug/calculate for internal audit use.
func (h *Handler) GetForecastDebugBreakdown(w http.ResponseWriter, r *http.Request) {
	if h.debugToken == "" {
		writeError(w, http.StatusForbidden, "debug endpoint disabled")
		return
	}

	providedToken := strings.TrimSpace(r.Header.Get("X-Debug-Token"))
	if providedToken == "" || providedToken != h.debugToken {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	userID, err := parseRequiredUUIDQuery(r, "user_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	date, err := parseRequiredDateQuery(r, "date")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	profileID, err := parseOptionalUUIDQuery(r, "profile_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	payload, err := h.service.GetForecastDebugBreakdown(*userID, profileID, *date)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, payload)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// parseHistoryFilter parses optional profile and date range filters from query params.
func parseHistoryFilter(r *http.Request) (HistoryFilter, error) {
	profileID, err := parseOptionalUUIDQuery(r, "profile_id")
	if err != nil {
		return HistoryFilter{}, err
	}

	startDate, err := parseOptionalDateQuery(r, "start_date")
	if err != nil {
		return HistoryFilter{}, err
	}

	endDate, err := parseOptionalDateQuery(r, "end_date")
	if err != nil {
		return HistoryFilter{}, err
	}

	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		return HistoryFilter{}, errInvalid("start_date must be before or equal to end_date")
	}

	return HistoryFilter{
		SolarProfileID: profileID,
		StartDate:      startDate,
		EndDate:        endDate,
	}, nil
}

// parseOptionalDateQuery reads one YYYY-MM-DD query param and returns nil when absent.
func parseOptionalDateQuery(r *http.Request, key string) (*time.Time, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		return nil, errInvalid(key + " must use YYYY-MM-DD")
	}

	normalized := parsed.UTC()
	return &normalized, nil
}

// parseRequiredDateQuery reads one required YYYY-MM-DD query param.
func parseRequiredDateQuery(r *http.Request, key string) (*time.Time, error) {
	parsed, err := parseOptionalDateQuery(r, key)
	if err != nil {
		return nil, err
	}
	if parsed == nil {
		return nil, errInvalid(key + " is required and must use YYYY-MM-DD")
	}

	return parsed, nil
}

// parseOptionalUUIDQuery reads one UUID query param and returns nil when absent.
func parseOptionalUUIDQuery(r *http.Request, key string) (*uuid.UUID, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, nil
	}

	parsed, err := uuid.Parse(raw)
	if err != nil {
		return nil, errInvalid(key + " must be a valid UUID")
	}

	return &parsed, nil
}

// parseRequiredUUIDQuery reads one required UUID query param.
func parseRequiredUUIDQuery(r *http.Request, key string) (*uuid.UUID, error) {
	parsed, err := parseOptionalUUIDQuery(r, key)
	if err != nil {
		return nil, err
	}
	if parsed == nil {
		return nil, errInvalid(key + " is required and must be a valid UUID")
	}

	return parsed, nil
}

// parseOptionalIntQuery parses one integer query param and falls back to default value.
func parseOptionalIntQuery(r *http.Request, key string, fallback int) int {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

// errInvalid creates a reusable bad-request error message.
func errInvalid(message string) error {
	return &validationError{message: message}
}

type validationError struct {
	message string
}

func (e *validationError) Error() string {
	return e.message
}
