package scheduler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/billing"
	"github.com/akbarsenawijaya/solar-forecast/internal/forecast"
	"github.com/akbarsenawijaya/solar-forecast/internal/notification"
	"github.com/akbarsenawijaya/solar-forecast/internal/solar"
	"github.com/akbarsenawijaya/solar-forecast/internal/tier"
	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

// Scheduler runs periodic jobs for the application.
type Scheduler struct {
	cron            *cron.Cron
	db              *sql.DB
	userService     user.Service
	solarService    solar.Service
	forecastService forecast.Service
	notifService    notification.Service
	billingService  billing.Service
}

// New creates a new scheduler with all required service dependencies.
func New(
	db *sql.DB,
	userSvc user.Service,
	solarSvc solar.Service,
	forecastSvc forecast.Service,
	notifSvc notification.Service,
	billingSvc billing.Service,
) *Scheduler {
	return &Scheduler{
		cron:            cron.New(cron.WithSeconds()),
		db:              db,
		userService:     userSvc,
		solarService:    solarSvc,
		forecastService: forecastSvc,
		notifService:    notifSvc,
		billingService:  billingSvc,
	}
}

// Start registers all cron jobs and starts the scheduler.
func (s *Scheduler) Start() {
	_, err := s.cron.AddFunc("0 */5 * * * *", func() {
		s.logRun("Daily Forecast Delivery", s.runScheduledForecastChecks)
	})
	if err != nil {
		log.Printf("failed to register daily forecast cron: %v", err)
	}

	_, err = s.cron.AddFunc("0 0 * * * *", func() {
		s.logRun("Subscription Cleanup", s.runSubscriptionCleanup)
	})
	if err != nil {
		log.Printf("failed to register subscription cleanup cron: %v", err)
	}

	// Run once daily to avoid sending duplicate expiry reminders throughout the same day.
	_, err = s.cron.AddFunc("0 1 0 * * *", func() {
		s.logRun("Subscription Expiry Notice", s.runSubscriptionExpiryNotice)
	})
	if err != nil {
		log.Printf("failed to register subscription expiry notice cron: %v", err)
	}

	s.cron.Start()
	log.Println("Scheduler started: background jobs active")
}

// Stop gracefully stops the scheduler.
func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("Scheduler stopped")
}

func (s *Scheduler) logRun(jobName string, fn func()) {
	startedAt := time.Now()
	
	// Use a recover to catch panics within the job
	var errStr string
	func() {
		defer func() {
			if r := recover(); r != nil {
				errStr = fmt.Sprintf("panic: %v", r)
				log.Printf("CRITICAL: job %s panicked: %v", jobName, r)
			}
		}()
		fn()
	}()

	finishedAt := time.Now()
	duration := finishedAt.Sub(startedAt)
	status := "success"
	if errStr != "" {
		status = "failed"
	}

	query := `
		INSERT INTO scheduler_runs (job_name, status, duration_ms, error_message, started_at, finished_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := s.db.Exec(query, jobName, status, duration.Milliseconds(), sql.NullString{String: errStr, Valid: errStr != ""}, startedAt, finishedAt)
	if err != nil {
		log.Printf("failed to log scheduler run for %s: %v", jobName, err)
	}
}

// runScheduledForecastChecks checks every verified user and only sends for those due in the current 5-minute window.
func (s *Scheduler) runScheduledForecastChecks() {
	log.Println("Running scheduled forecast delivery check...")
	nowUTC := time.Now().UTC()
	prefs, err := s.notifService.GetAllPreferences()
	if err != nil {
		log.Printf("scheduled forecast: failed to fetch notification preferences: %v", err)
		return
	}

	prefMap := make(map[uuid.UUID]*notification.NotificationPreference, len(prefs))
	for _, pref := range prefs {
		prefMap[pref.UserID] = pref
	}

	users, err := s.userService.GetAllUsers()
	if err != nil {
		log.Printf("scheduled forecast: failed to fetch users: %v", err)
		return
	}

	dueCount := 0

	for _, currentUser := range users {
		if !currentUser.EmailVerified {
			log.Printf("scheduled forecast: skipped for user %s (%s): email not verified", currentUser.Name, currentUser.ID)
			continue
		}

		pref := prefMap[currentUser.ID]
		if pref == nil {
			pref, err = s.notifService.GetPreference(currentUser.ID)
			if err != nil {
				log.Printf("scheduled forecast: failed to resolve preference for user %s (%s): %v", currentUser.Name, currentUser.ID, err)
				continue
			}
		}

		if pref == nil {
			log.Printf("scheduled forecast: skipped for user %s (%s): preference unavailable", currentUser.Name, currentUser.ID)
			continue
		}

		localNow, localToday, due, reason := dueForDispatch(pref, nowUTC)
		if !due {
			if reason != "outside delivery window" {
				log.Printf("scheduled forecast: skipped for user %s (%s): %s", currentUser.Name, currentUser.ID, reason)
			}
			continue
		}

		dueCount++
		yesterday := localToday.AddDate(0, 0, -1)

		if err := s.calibrateUserEfficiency(currentUser.ID, yesterday); err != nil {
			log.Printf("scheduled calibration: failed for user %s (%s): %v", currentUser.Name, currentUser.ID, err)
		}

		if err := s.processUserForecast(currentUser.ID, currentUser.Name, currentUser.Email, localToday, pref); err != nil {
			if isMissingSolarProfileError(err) {
				log.Printf("scheduled forecast: skipped for user %s (%s): solar profile not available", currentUser.Name, currentUser.ID)
				continue
			}
			log.Printf("scheduled forecast: failed for user %s (%s): %v", currentUser.Name, currentUser.ID, err)
			continue
		}

		if err := s.notifService.MarkDailyForecastSent(currentUser.ID, localToday, nowUTC); err != nil {
			log.Printf("scheduled forecast: failed to mark sent for user %s (%s): %v", currentUser.Name, currentUser.ID, err)
			continue
		}

		log.Printf("scheduled forecast: sent for user %s (%s) at local %s", currentUser.Name, currentUser.ID, localNow.Format(time.RFC3339))
	}

	log.Printf("Scheduled forecast delivery check completed: %d due users processed out of %d", dueCount, len(users))
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
func (s *Scheduler) processUserForecast(userID uuid.UUID, name, email string, date time.Time, pref *notification.NotificationPreference) error {
	result, err := s.forecastService.GenerateForecastForUser(userID, date)
	if err != nil {
		return fmt.Errorf("generate forecast: %w", err)
	}

	solarProfileName := "-"
	var lat, lng float64
	if result.SolarProfileID != nil {
		if profile, profileErr := s.solarService.GetSolarProfileByIDAndUserID(*result.SolarProfileID, userID); profileErr == nil {
			solarProfileName = profile.SiteName
			lat = profile.Lat
			lng = profile.Lng
		}
	}

	var deviationPct *float64
	referenceLabel := "actual referensi"
	if result.SolarProfileID != nil {
		planTier := tier.Free
		if pref != nil {
			planTier = pref.PlanTier
		}
		actualHistory, historyErr := s.forecastService.GetActualHistory(userID, planTier, 90, forecast.HistoryFilter{SolarProfileID: result.SolarProfileID})
		if historyErr == nil {
			referenceActual := findReferenceActual(actualHistory.Items, date)
			if referenceActual != nil && referenceActual.ActualKwh > 0 {
				value := ((result.PredictedKwh - referenceActual.ActualKwh) / referenceActual.ActualKwh) * 100
				deviationPct = &value
				referenceLabel = fmt.Sprintf("actual %s", referenceActual.Date.Format("02 Jan 2006"))
			}
		}
	}

	weatherRisk := forecast.DetermineWeatherRisk(int(result.CloudCover), result.DeltaWF)

	conditionLabel, conditionImpact := getConditionText(result.WeatherFactor)
	baselineType := result.BaselineType
	if baselineType == "" {
		baselineType = "synthetic"
	}

	payload := notification.DispatchPayload{
		UserID:           userID,
		ToName:           name,
		ToEmail:          email,
		Date:             date.Format("2006-01-02"),
		PredictedKwh:     result.PredictedKwh,
		CloudCover:       result.CloudCover,
		BaselineType:     baselineType,
		WeatherFactor:    result.WeatherFactor,
		Efficiency:       result.Efficiency,
		SolarProfileName: solarProfileName,
		WeatherRisk:      weatherRisk,
		EstimatedCost:    result.PredictedKwh * 1444,
		EstimatedCO2Kg:   result.PredictedKwh * getEmissionFactor(lat, lng),
		DeviationPct:     deviationPct,
		ReferenceLabel:   referenceLabel,
		Lat:              lat,
		Lng:              lng,
		ConditionLabel:   conditionLabel,
		ConditionImpact:  conditionImpact,
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

func dueForDispatch(pref *notification.NotificationPreference, nowUTC time.Time) (time.Time, time.Time, bool, string) {
	locationName := strings.TrimSpace(pref.Timezone)
	if locationName == "" {
		locationName = "UTC"
	}

	location, err := time.LoadLocation(locationName)
	if err != nil {
		return time.Time{}, time.Time{}, false, "invalid timezone"
	}

	preferredClock := strings.TrimSpace(pref.PreferredSendTime)
	if preferredClock == "" {
		preferredClock = "06:00:00"
	}

	hour, minute, second, err := parseClock(preferredClock)
	if err != nil {
		return time.Time{}, time.Time{}, false, "invalid preferred send time"
	}

	localNow := nowUTC.In(location)
	localToday := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, location)
	targetTime := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), hour, minute, second, 0, location)
	windowStart := targetTime
	windowEnd := targetTime.Add(5 * time.Minute)

	if localNow.Before(windowStart) || !localNow.Before(windowEnd) {
		return localNow, localToday, false, "outside delivery window"
	}

	if pref.LastDailyForecastSentForDate != nil {
		lastSentDate := *pref.LastDailyForecastSentForDate
		if sameCalendarDay(lastSentDate, localToday) {
			return localNow, localToday, false, "already sent for local date"
		}
	}

	return localNow, localToday, true, "due"
}

func parseClock(value string) (int, int, int, error) {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("clock must use HH:MM:SS")
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, 0, fmt.Errorf("invalid hour")
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, 0, 0, fmt.Errorf("invalid minute")
	}

	second, err := strconv.Atoi(parts[2])
	if err != nil || second < 0 || second > 59 {
		return 0, 0, 0, fmt.Errorf("invalid second")
	}

	return hour, minute, second, nil
}

func sameCalendarDay(left, right time.Time) bool {
	return left.Year() == right.Year() && left.Month() == right.Month() && left.Day() == right.Day()
}

func getEmissionFactor(lat, lng float64) float64 {
	if lat >= -9.0 && lat <= -5.0 && lng >= 105.0 && lng <= 116.0 {
		return 0.85
	}
	if lat >= -6.0 && lat <= 6.0 && lng >= 95.0 && lng <= 106.0 {
		return 0.75
	}
	if lat >= -4.0 && lat <= 5.0 && lng >= 108.0 && lng <= 119.0 {
		return 0.80
	}
	if lat >= -6.0 && lat <= 2.0 && lng >= 118.0 && lng <= 125.0 {
		return 0.65
	}
	if lat >= -11.0 && lat <= 0.0 && lng >= 125.0 && lng <= 141.0 {
		return 0.70
	}
	return 0.78
}

func getConditionText(wf float64) (string, string) {
	if wf >= 0.9 {
		return "cerah", "dampak ke produksi rendah, panel berpotensi menghasilkan energi optimal"
	}
	if wf >= 0.75 {
		return "berawan", "dampak ke produksi ringan, output masih cukup baik"
	}
	if wf >= 0.6 {
		return "mendung", "dampak ke produksi sedang, output cenderung turun dibanding hari cerah"
	}
	return "mendung tebal", "dampak ke produksi tinggi, output berpotensi turun signifikan"
}

// runSubscriptionCleanup checks for past-due subscriptions and downgrades users.
func (s *Scheduler) runSubscriptionCleanup() {
	log.Println("Running background subscription cleanup...")
	ctx := context.Background()
	if err := s.billingService.CleanupExpiredSubscriptions(ctx); err != nil {
		log.Printf("cleanup: failed to process expired subscriptions: %v", err)
	}
}

func (s *Scheduler) runSubscriptionExpiryNotice() {
	log.Println("Running background subscription expiry notice...")
	ctx := context.Background()
	if err := s.billingService.NotifyExpiringSubscriptions(ctx); err != nil {
		log.Printf("expiry notice: failed to notify users: %v", err)
	}
}
