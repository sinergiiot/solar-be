package scheduler

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/forecast"
	"github.com/akbarsenawijaya/solar-forecast/internal/notification"
	"github.com/akbarsenawijaya/solar-forecast/internal/solar"
	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

// Scheduler runs periodic jobs for the application.
type Scheduler struct {
	cron            *cron.Cron
	userService     user.Service
	solarService    solar.Service
	forecastService forecast.Service
	notifService    notification.Service
}

// New creates a new scheduler with all required service dependencies.
func New(
	userSvc user.Service,
	solarSvc solar.Service,
	forecastSvc forecast.Service,
	notifSvc notification.Service,
) *Scheduler {
	return &Scheduler{
		cron:            cron.New(cron.WithSeconds()),
		userService:     userSvc,
		solarService:    solarSvc,
		forecastService: forecastSvc,
		notifService:    notifSvc,
	}
}

// Start registers all cron jobs and starts the scheduler.
func (s *Scheduler) Start() {
	_, err := s.cron.AddFunc("0 36 15 * * *", s.runDailyForecastJob)
	if err != nil {
		log.Fatalf("failed to register daily forecast cron: %v", err)
	}

	s.cron.Start()
	log.Println("Scheduler started: daily forecast job runs at 15:36 UTC")
}

// Stop gracefully stops the scheduler.
func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("Scheduler stopped")
}

// runDailyForecastJob generates and delivers forecasts for all users.
func (s *Scheduler) runDailyForecastJob() {
	log.Println("Running daily forecast job...")
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
	yesterday := today.AddDate(0, 0, -1)

	users, err := s.userService.GetAllUsers()
	if err != nil {
		log.Printf("daily forecast: failed to fetch users: %v", err)
		return
	}

	for _, currentUser := range users {
		if err := s.calibrateUserEfficiency(currentUser.ID, yesterday); err != nil {
			log.Printf("daily calibration: failed for user %s (%s): %v", currentUser.Name, currentUser.ID, err)
		}

		if err := s.processUserForecast(currentUser.ID, currentUser.Name, currentUser.Email, today); err != nil {
			if isMissingSolarProfileError(err) {
				log.Printf("daily forecast: skipped for user %s (%s): solar profile not available", currentUser.Name, currentUser.ID)
				continue
			}
			log.Printf("daily forecast: failed for user %s (%s): %v", currentUser.Name, currentUser.ID, err)
		}
	}

	log.Printf("Daily forecast job completed for %d users", len(users))
}

// calibrateUserEfficiency updates one user's efficiency from yesterday actual data when available.
func (s *Scheduler) calibrateUserEfficiency(userID uuid.UUID, date time.Time) error {
	result, err := s.forecastService.CalibrateEfficiencyForUser(userID, date)
	if err != nil {
		return fmt.Errorf("calibrate efficiency: %w", err)
	}

	if !result.Updated {
		log.Printf("Calibration skipped for %s on %s: %s", userID, date.Format("2006-01-02"), result.Message)
		return nil
	}

	log.Printf(
		"Calibration applied for %s on %s: eff %.4f -> %.4f (pred=%.2f, actual=%.2f)",
		userID,
		date.Format("2006-01-02"),
		result.OldEfficiency,
		result.NewEfficiency,
		result.PredictedKwh,
		result.ActualKwh,
	)

	return nil
}

// processUserForecast generates the forecast and sends the notification for one user.
func (s *Scheduler) processUserForecast(userID uuid.UUID, name, email string, date time.Time) error {
	result, err := s.forecastService.GenerateForecastForUser(userID, date)
	if err != nil {
		return fmt.Errorf("generate forecast: %w", err)
	}

	solarProfileName := "-"
	if result.SolarProfileID != nil {
		if profile, profileErr := s.solarService.GetSolarProfileByIDAndUserID(*result.SolarProfileID, userID); profileErr == nil {
			solarProfileName = profile.SiteName
		}
	}

	var deviationPct *float64
	referenceLabel := "actual referensi"
	if result.SolarProfileID != nil {
		actualHistory, historyErr := s.forecastService.GetActualHistory(userID, 90, forecast.HistoryFilter{SolarProfileID: result.SolarProfileID})
		if historyErr == nil {
			referenceActual := findReferenceActual(actualHistory, date)
			if referenceActual != nil && referenceActual.ActualKwh > 0 {
				value := ((result.PredictedKwh - referenceActual.ActualKwh) / referenceActual.ActualKwh) * 100
				deviationPct = &value
				referenceLabel = fmt.Sprintf("actual %s", referenceActual.Date.Format("02 Jan 2006"))
			}
		}
	}

	weatherRisk := "Risiko Cuaca Tinggi"
	if result.WeatherFactor >= 0.9 {
		weatherRisk = "Risiko Cuaca Rendah"
	} else if result.WeatherFactor >= 0.7 {
		weatherRisk = "Risiko Cuaca Sedang"
	}

	payload := notification.DispatchPayload{
		UserID:           userID,
		ToName:           name,
		ToEmail:          email,
		Date:             date.Format("2006-01-02"),
		PredictedKwh:     result.PredictedKwh,
		WeatherFactor:    result.WeatherFactor,
		Efficiency:       result.Efficiency,
		SolarProfileName: solarProfileName,
		WeatherRisk:      weatherRisk,
		EstimatedCost:    result.PredictedKwh * 1444,
		EstimatedCO2Kg:   result.PredictedKwh * 0.85,
		DeviationPct:     deviationPct,
		ReferenceLabel:   referenceLabel,
	}

	if err := s.notifService.DispatchDailyForecast(payload); err != nil {
		return fmt.Errorf("send notification: %w", err)
	}

	log.Printf("Forecast sent to %s: %.2f kWh", email, result.PredictedKwh)
	return nil
}

// findReferenceActual chooses one latest actual reference, preferring the latest date before forecast date.
func findReferenceActual(actuals []*forecast.ActualDaily, forecastDate time.Time) *forecast.ActualDaily {
	if len(actuals) == 0 {
		return nil
	}

	forecastDay := time.Date(forecastDate.Year(), forecastDate.Month(), forecastDate.Day(), 0, 0, 0, 0, time.UTC)
	for _, actual := range actuals {
		actualDay := time.Date(actual.Date.Year(), actual.Date.Month(), actual.Date.Day(), 0, 0, 0, 0, time.UTC)
		if actualDay.Before(forecastDay) {
			return actual
		}
	}

	return actuals[0]
}

// isMissingSolarProfileError detects scheduler-safe cases where a user has not configured solar data yet.
func isMissingSolarProfileError(err error) bool {
	return err != nil && (containsForecastProfileNotFound(err) || containsNoRowsProfileLookup(err))
}

// containsForecastProfileNotFound checks for the forecast sentinel error.
func containsForecastProfileNotFound(err error) bool {
	return errors.Is(err, forecast.ErrSolarProfileNotFound)
}

// containsNoRowsProfileLookup covers wrapped legacy errors coming from repository/service layers.
func containsNoRowsProfileLookup(err error) bool {
	message := err.Error()
	return message != "" &&
		contains(message, "get solar profile") &&
		contains(message, "sql: no rows in result set")
}

// contains wraps strings.Contains for explicit log-skip helpers.
func contains(text, fragment string) bool {
	return strings.Contains(text, fragment)
}
