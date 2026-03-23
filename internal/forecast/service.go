package forecast

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/solar"
	"github.com/akbarsenawijaya/solar-forecast/internal/weather"
	"github.com/google/uuid"
)

const (
	defaultEfficiency = 0.8
	minEfficiency     = 0.6
	maxEfficiency     = 0.95
	learningRate      = 0.2
)

var (
	// ErrSolarProfileNotFound indicates a user has not configured any solar profile yet.
	ErrSolarProfileNotFound = errors.New("solar profile not found")
)

// Service defines business logic for generating solar forecasts
type Service interface {
	GenerateForecastForUser(userID uuid.UUID, date time.Time) (*Forecast, error)
	GetTodayForecastForUser(userID uuid.UUID, solarProfileID *uuid.UUID) (*Forecast, error)
	GetForecastDebugBreakdown(userID uuid.UUID, solarProfileID *uuid.UUID, date time.Time) (*ForecastDebugBreakdown, error)
	RecordActualDaily(req RecordActualRequest) (*ActualDaily, error)
	CalibrateEfficiencyForUser(userID uuid.UUID, date time.Time) (*CalibrationResult, error)
	GetForecastHistory(userID uuid.UUID, days int, filter HistoryFilter) ([]*Forecast, error)
	GetActualHistory(userID uuid.UUID, days int, filter HistoryFilter) ([]*ActualDaily, error)
	GetDashboardSummary(userID uuid.UUID) (*DashboardSummary, error)
}


// service implements forecast business logic, now with weatherbaseline
type service struct {
	repo           Repository
	solarService   solar.Service
	weatherService weather.Service
	weatherBaselineService interface {
        GetSyntheticBaseline(ctx context.Context, profileID, userID string, lat, lng float64) (float64, int, error)
        GetSiteBaseline(ctx context.Context, profileID, userID string) (float64, int, error)
    }
}

// NewService creates a new forecast service with weather baseline
func NewService(repo Repository, solarSvc solar.Service, weatherSvc weather.Service, weatherBaselineSvc interface {
	GetSyntheticBaseline(ctx context.Context, profileID, userID string, lat, lng float64) (float64, int, error)
	GetSiteBaseline(ctx context.Context, profileID, userID string) (float64, int, error)
}) Service {
	return &service{
		repo:           repo,
		solarService:   solarSvc,
		weatherService: weatherSvc,
		weatherBaselineService: weatherBaselineSvc,
	}
}

// GenerateForecastForUser fetches weather and solar data, calculates and saves a forecast
func (s *service) GenerateForecastForUser(userID uuid.UUID, date time.Time) (*Forecast, error) {
	profile, err := s.solarService.GetSolarProfileByUserID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w for user %s", ErrSolarProfileNotFound, userID)
		}
		return nil, fmt.Errorf("get solar profile for user %s: %w", userID, err)
	}
	return s.generateForecastForProfile(userID, profile.ID, date)
}

// generateForecastForProfile calculates and stores one forecast for a selected profile using ΔWF
func (s *service) generateForecastForProfile(userID uuid.UUID, solarProfileID uuid.UUID, date time.Time) (*Forecast, error) {
	// Get the user's solar profile
	profile, err := s.solarService.GetSolarProfileByIDAndUserID(solarProfileID, userID)
	if err != nil {
		log.Println("[DEBUG] ERROR: get solar profile", solarProfileID, userID, err)
		return nil, fmt.Errorf("get solar profile %s for user %s: %w", solarProfileID, userID, err)
	}

	// Fetch weather for the profile's location
	w, err := s.weatherService.FetchWeatherForDate(profile.Lat, profile.Lng, date)
	if err != nil {
		log.Println("[ERROR] fetch weather failed", profile.Lat, profile.Lng, date, err)
		return nil, fmt.Errorf("fetch weather for user %s: %w", userID, err)
	}
	if w == nil || w.CloudCover == 0 {
		log.Println("[ERROR] weather data missing or cloud_cover=0", profile.Lat, profile.Lng, date, w)
		return nil, fmt.Errorf("weather data missing or cloud_cover=0 for user %s, profile %s, date %s", userID, profile.ID, date.Format("2006-01-02"))
	}

	psh, err := getPSHFromWeather(w)
	if err != nil {
		log.Println("[ERROR] getPSHFromWeather", w, err)
		return nil, fmt.Errorf("derive psh from weather data: %w", err)
	}

	efficiency, err := s.repo.GetUserEfficiency(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			efficiency = defaultEfficiency
		} else {
			log.Println("[DEBUG] ERROR: getUserEfficiency", userID, err)
			return nil, fmt.Errorf("get user efficiency for user %s: %w", userID, err)
		}
	}

	// Calculate deltaWF and baselineType using weatherbaseline
	deltaRes, err := ComputeDeltaWF(
	    context.Background(),
	    s.repo,
	    s.weatherBaselineService,
	    profile.ID.String(),
	    userID.String(),
	    float64(w.CloudCover),
	    profile.Lat,
	    profile.Lng,
	)
	if err != nil {
		log.Println("[ERROR] ComputeDeltaWF failed", profile.ID, userID, w.CloudCover, err)
	    return nil, fmt.Errorf("compute deltaWF: %w", err)
	}

	// DEBUG LOG: print all calculation inputs and outputs
	predictedKwh := profile.CapacityKwp * psh * efficiency * deltaRes.DeltaWF
	// HARD CAP: cannot exceed 95% of theoretical max
	theoreticalMax := profile.CapacityKwp * psh * 0.95
	predictedKwh = math.Min(predictedKwh, theoreticalMax)
	if predictedKwh < 0 {
		predictedKwh = 0
	}

	log.Println("[DEBUG] Forecast calculation:")
	log.Println("  user_id:", userID)
	log.Println("  profile_id:", profile.ID)
	log.Println("  date:", date.Format("2006-01-02"))
	log.Println("  capacity_kWp:", profile.CapacityKwp)
	log.Println("  shortwave_MJ:", w.ShortwaveRadiationMJ)
	log.Println("  psh:", psh)
	log.Println("  efficiency:", efficiency)
	log.Println("  cloud_cover:", w.CloudCover)
	log.Println("  baseline_type:", deltaRes.BaselineType)
	log.Println("  delta_wf:", deltaRes.DeltaWF)
	log.Println("  predicted_kwh:", predictedKwh)

f := &Forecast{
	ID:             uuid.New(),
	UserID:         userID,
	SolarProfileID: &profile.ID,
	Date:           date,
	PredictedKwh:   predictedKwh,
	WeatherFactor:  deltaRes.WeatherFactor, // correct weather factor (not deltaWF)
	CloudCover:     int(w.CloudCover),      // store as percent int
	Efficiency:     efficiency,
	DeltaWF:        deltaRes.DeltaWF,
	BaselineType:   deltaRes.BaselineType,
	CreatedAt:      time.Now().UTC(),
}

	if err := s.repo.SaveForecast(f); err != nil {
		return nil, err
	}
	return f, nil
}

// RecordActualDaily validates and stores one actual daily production record.
func (s *service) RecordActualDaily(req RecordActualRequest) (*ActualDaily, error) {
	if req.ActualKwh <= 0 {
		return nil, fmt.Errorf("actual_kwh must be greater than 0")
	}

	source := strings.TrimSpace(strings.ToLower(req.Source))
	if source == "" {
		source = "manual"
	}

	var parsedDate time.Time
	var err error
	if strings.TrimSpace(req.Date) == "" {
		parsedDate = time.Now().UTC().Truncate(24 * time.Hour)
	} else {
		parsedDate, err = time.Parse(time.DateOnly, req.Date)
		if err != nil {
			return nil, fmt.Errorf("date must use YYYY-MM-DD")
		}
	}

	var solarProfileID *uuid.UUID
	if strings.TrimSpace(req.SolarProfileID) != "" {
		parsedProfileID, err := uuid.Parse(strings.TrimSpace(req.SolarProfileID))
		if err != nil {
			return nil, fmt.Errorf("solar_profile_id must be a valid UUID")
		}

		if _, err := s.solarService.GetSolarProfileByIDAndUserID(parsedProfileID, req.UserID); err != nil {
			return nil, fmt.Errorf("solar profile not found for user")
		}

		solarProfileID = &parsedProfileID
	}

	a := &ActualDaily{
		ID:             uuid.New(),
		UserID:         req.UserID,
		SolarProfileID: solarProfileID,
		Date:           parsedDate,
		ActualKwh:      req.ActualKwh,
		Source:         source,
		CreatedAt:      time.Now().UTC(),
	}

	if err := s.repo.SaveActualDaily(a); err != nil {
		return nil, err
	}

	return a, nil
}

// CalibrateEfficiencyForUser updates user efficiency using actual vs predicted for one date.
func (s *service) CalibrateEfficiencyForUser(userID uuid.UUID, date time.Time) (*CalibrationResult, error) {
	currentEfficiency, err := s.repo.GetUserEfficiency(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			currentEfficiency = defaultEfficiency
		} else {
			return nil, fmt.Errorf("get user efficiency for calibration: %w", err)
		}
	}

	actual, err := s.repo.GetActualDailyByUserAndDate(userID, date)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &CalibrationResult{
				UserID:        userID,
				Date:          normalizeDate(date),
				OldEfficiency: currentEfficiency,
				NewEfficiency: currentEfficiency,
				Updated:       false,
				Message:       "learning skipped: actual data not available",
			}, nil
		}
		return nil, fmt.Errorf("get actual daily for calibration: %w", err)
	}

	profile, profileErr := s.solarService.GetSolarProfileByUserID(userID)
	if profileErr != nil {
		if errors.Is(profileErr, sql.ErrNoRows) {
			return &CalibrationResult{
				UserID:        userID,
				Date:          normalizeDate(date),
				OldEfficiency: currentEfficiency,
				NewEfficiency: currentEfficiency,
				ActualKwh:     actual.ActualKwh,
				Updated:       false,
				Message:       "learning skipped: solar profile not available",
			}, nil
		}
		return nil, fmt.Errorf("get solar profile for calibration: %w", profileErr)
	}

	forecastForDate, err := s.repo.GetForecastByUserAndDate(userID, profile.ID, date)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &CalibrationResult{
				UserID:        userID,
				Date:          normalizeDate(date),
				OldEfficiency: currentEfficiency,
				NewEfficiency: currentEfficiency,
				ActualKwh:     actual.ActualKwh,
				Updated:       false,
				Message:       "learning skipped: forecast data not available",
			}, nil
		}
		return nil, fmt.Errorf("get forecast for calibration: %w", err)
	}

	if forecastForDate.PredictedKwh <= 0 {
		return &CalibrationResult{
			UserID:        userID,
			Date:          normalizeDate(date),
			OldEfficiency: currentEfficiency,
			NewEfficiency: currentEfficiency,
			PredictedKwh:  forecastForDate.PredictedKwh,
			ActualKwh:     actual.ActualKwh,
			Updated:       false,
			Message:       "learning skipped: predicted_kwh must be greater than 0",
		}, nil
	}

	correctionRate := actual.ActualKwh / forecastForDate.PredictedKwh
	targetEfficiency := currentEfficiency * correctionRate
	nextEfficiency := (1-learningRate)*currentEfficiency + learningRate*targetEfficiency
	nextEfficiency = clamp(nextEfficiency, minEfficiency, maxEfficiency)

	if err := s.repo.UpdateUserEfficiency(userID, nextEfficiency); err != nil {
		return nil, fmt.Errorf("update user efficiency for calibration: %w", err)
	}

	return &CalibrationResult{
		UserID:         userID,
		Date:           normalizeDate(date),
		OldEfficiency:  currentEfficiency,
		NewEfficiency:  nextEfficiency,
		PredictedKwh:   forecastForDate.PredictedKwh,
		ActualKwh:      actual.ActualKwh,
		CorrectionRate: correctionRate,
		Updated:        true,
		Message:        "learning applied",
	}, nil
}

// GetTodayForecastForUser returns the cached forecast for today, or generates one
func (s *service) GetTodayForecastForUser(userID uuid.UUID, solarProfileID *uuid.UUID) (*Forecast, error) {
	// Allow override for testing (via TEST_DATE env var in YYYY-MM-DD format)
	var today time.Time
	if testDateStr := os.Getenv("TEST_DATE"); testDateStr != "" {
		parsed, err := time.Parse("2006-01-02", testDateStr)
		if err == nil {
			today = parsed
		} else {
			now := time.Now().UTC()
			today = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		}
	} else {
		now := time.Now().UTC()
		today = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	}

	selectedProfileID := solarProfileID

	if selectedProfileID == nil {
		profile, err := s.solarService.GetSolarProfileByUserID(userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("%w for user %s", ErrSolarProfileNotFound, userID)
			}
			return nil, fmt.Errorf("get solar profile for user %s: %w", userID, err)
		}
		selectedProfileID = &profile.ID
	}

	if _, err := s.solarService.GetSolarProfileByIDAndUserID(*selectedProfileID, userID); err != nil {
		return nil, fmt.Errorf("solar profile not found for user")
	}

	cached, err := s.repo.GetForecastByUserAndDate(userID, *selectedProfileID, today)
	if err == nil {
		return cached, nil
	}

	// No forecast yet — generate one on-demand
	return s.generateForecastForProfile(userID, *selectedProfileID, today)
}

// GetForecastDebugBreakdown returns a read-only audit payload for one forecast calculation.
func (s *service) GetForecastDebugBreakdown(userID uuid.UUID, solarProfileID *uuid.UUID, date time.Time) (*ForecastDebugBreakdown, error) {
	targetDate := normalizeDate(date)

	selectedProfileID := solarProfileID
	if selectedProfileID == nil {
		profile, err := s.solarService.GetSolarProfileByUserID(userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("%w for user %s", ErrSolarProfileNotFound, userID)
			}
			return nil, fmt.Errorf("get solar profile for user %s: %w", userID, err)
		}
		selectedProfileID = &profile.ID
	}

	profile, err := s.solarService.GetSolarProfileByIDAndUserID(*selectedProfileID, userID)
	if err != nil {
		return nil, fmt.Errorf("get solar profile %s for user %s: %w", *selectedProfileID, userID, err)
	}

	w, err := s.weatherService.FetchWeatherForDate(profile.Lat, profile.Lng, targetDate)
	if err != nil {
		return nil, fmt.Errorf("fetch weather for user %s: %w", userID, err)
	}

	psh, err := getPSHFromWeather(w)
	if err != nil {
		return nil, fmt.Errorf("derive psh from weather data: %w", err)
	}

	efficiency, err := s.repo.GetUserEfficiency(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			efficiency = defaultEfficiency
		} else {
			return nil, fmt.Errorf("get user efficiency for user %s: %w", userID, err)
		}
	}

	weatherFactor := getWeatherFactor(w.CloudCover)
	predictedKwh := CalculateForecast(profile.CapacityKwp, psh, efficiency)

	inputs := ForecastDebugInputs{
		UserID:               userID,
		SolarProfileID:       selectedProfileID,
		Date:                 targetDate.Format(time.DateOnly),
		CapacityKwp:          profile.CapacityKwp,
		CloudCoverPercent:    w.CloudCover,
		TemperatureC:         w.Temperature,
		ShortwaveRadiationMJ: w.ShortwaveRadiationMJ,
		PSH:                  psh,
		WeatherFactor:        weatherFactor,
		ForecastEfficiency:   efficiency,
	}

	if existingForecast, err := s.repo.GetForecastByUserAndDate(userID, *selectedProfileID, targetDate); err == nil {
		inputs.ExistingForecastFound = true
		inputs.ExistingForecastKwh = &existingForecast.PredictedKwh
	}

	if existingActual, err := s.repo.GetActualDailyByUserAndDate(userID, targetDate); err == nil {
		inputs.ExistingActualFound = true
		inputs.ExistingActualKwh = &existingActual.ActualKwh
	}

	return &ForecastDebugBreakdown{
		Inputs:       inputs,
		Formula:      "predicted_kwh = capacity_kwp * psh * forecast_efficiency, with psh = shortwave_radiation_mj / 3.6 and weather_factor = 1 - cloud_cover_percent/100",
		PredictedKwh: predictedKwh,
	}, nil
}

// CalculateForecast computes solar energy (kWh) from capacity, PSH, and efficiency.
func CalculateForecast(capacityKwp float64, psh float64, efficiency float64) float64 {
	return capacityKwp * psh * efficiency
}

// CalculateForecastWithWeatherFactor computes solar energy (kWh) for cold start (no actual data), including Weather Factor.
func CalculateForecastWithWeatherFactor(capacityKwp float64, psh float64, efficiency float64, weatherFactor float64) float64 {
	return capacityKwp * psh * efficiency * weatherFactor
}

// GetForecastHistory returns recent forecasts for a user.
func (s *service) GetForecastHistory(userID uuid.UUID, days int, filter HistoryFilter) ([]*Forecast, error) {
	return s.repo.GetForecastHistoryByUser(userID, days, filter)
}

// GetActualHistory returns recent actual measurements for a user.
func (s *service) GetActualHistory(userID uuid.UUID, days int, filter HistoryFilter) ([]*ActualDaily, error) {
	return s.repo.GetActualHistoryByUser(userID, days, filter)
}

// GetDashboardSummary computes key performance indicators for the user dashboard.
func (s *service) GetDashboardSummary(userID uuid.UUID) (*DashboardSummary, error) {
	forecasts, err := s.repo.GetForecastHistoryByUser(userID, 90, HistoryFilter{})
	if err != nil {
		return nil, fmt.Errorf("get forecast history for summary: %w", err)
	}

	actuals, err := s.repo.GetActualHistoryByUser(userID, 90, HistoryFilter{})
	if err != nil {
		return nil, fmt.Errorf("get actual history for summary: %w", err)
	}

	summary := &DashboardSummary{
		ForecastCount: len(forecasts),
		ActualCount:   len(actuals),
	}

	// Calculate totals and averages from forecasts
	if len(forecasts) > 0 {
		for _, f := range forecasts {
			summary.TotalForecastedKwh += f.PredictedKwh
		}
		summary.AverageForecastKwh = summary.TotalForecastedKwh / float64(len(forecasts))
		summary.LastForecastDate = &forecasts[0].Date
	}

	// Calculate totals and averages from actuals
	if len(actuals) > 0 {
		for _, a := range actuals {
			summary.TotalActualKwh += a.ActualKwh
		}
		summary.AverageActualKwh = summary.TotalActualKwh / float64(len(actuals))
		summary.LastActualDate = &actuals[0].Date
	}

	// Get current efficiency
	efficiency, err := s.repo.GetUserEfficiency(userID)
	if err == nil {
		summary.CurrentEfficiency = efficiency
	}

	// Calculate accuracy if both forecasts and actuals exist
	if len(forecasts) > 0 && len(actuals) > 0 && summary.TotalForecastedKwh > 0 {
		accuracy := (summary.TotalActualKwh / summary.TotalForecastedKwh) * 100
		if accuracy > 100 {
			accuracy = 100
		}
		summary.AccuracyPercent = accuracy
	}

	return summary, nil
}

// getWeatherFactor returns a continuous transmittance factor based on cloud cover percentage.
func getWeatherFactor(cloudCover int) float64 {
	cc := clamp(float64(cloudCover), 0, 100)
	return 1 - (cc / 100)
}

// getPSHFromWeather converts shortwave radiation (MJ/m2/day) to PSH (kWh/m2/day).
func getPSHFromWeather(w *weather.WeatherDaily) (float64, error) {
	if w == nil {
		return 0, fmt.Errorf("weather data is nil")
	}
	if w.ShortwaveRadiationMJ <= 0 {
		return 0, fmt.Errorf("shortwave_radiation_mj must be greater than 0")
	}

	// 1 kWh = 3.6 MJ
	return w.ShortwaveRadiationMJ / 3.6, nil
}

// ...existing code... (removed duplicate clamp)
