package forecast

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Repository defines data access methods for forecasts
type Repository interface {
	// CountValidActualDays returns the number of days with actual_kwh > 0 for a user/profile
	CountValidActualDays(ctx context.Context, userID uuid.UUID, profileID uuid.UUID) (int, error)
	SaveForecast(f *Forecast) error
	GetForecastByUserAndDate(userID uuid.UUID, solarProfileID uuid.UUID, date time.Time) (*Forecast, error)
	GetAllForecastsByDate(date time.Time) ([]*Forecast, error)
	SaveActualDaily(a *ActualDaily) error
	GetActualDailyByUserAndDate(userID uuid.UUID, date time.Time) (*ActualDaily, error)
	GetUserEfficiency(userID uuid.UUID) (float64, error)
	UpdateUserEfficiency(userID uuid.UUID, efficiency float64) error
	GetForecastHistoryByUser(userID uuid.UUID, days int, filter HistoryFilter) ([]*Forecast, error)
	GetActualHistoryByUser(userID uuid.UUID, days int, filter HistoryFilter) ([]*ActualDaily, error)
	CountForecastHistoryByUser(userID uuid.UUID, days int, filter HistoryFilter) (int, error)
	CountActualHistoryByUser(userID uuid.UUID, days int, filter HistoryFilter) (int, error)
	HasAnyActualData(userID uuid.UUID) (bool, error)
	GetPaginatedActuals(userID uuid.UUID, req ListActualsRequest) ([]ActualWithProfile, int, error)
	GetPaginatedForecasts(userID uuid.UUID, req ListActualsRequest) ([]ForecastWithProfile, int, error)
}
// HasAnyActualData returns true if the user has any actual daily data recorded.
func (r *repository) HasAnyActualData(userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM actual_daily WHERE user_id = $1 LIMIT 1)`
	var exists bool
	err := r.db.QueryRow(query, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check actual data: %w", err)
	}
	return exists, nil
}

type repository struct {
	db *sql.DB
}

// NewRepository creates a new forecast repository
func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// SaveForecast inserts or updates a forecast for a user on a given date
func (r *repository) SaveForecast(f *Forecast) error {
	query := `
		INSERT INTO forecasts (id, user_id, solar_profile_id, date, predicted_kwh, weather_factor, cloud_cover, efficiency, delta_wf, baseline_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (user_id, solar_profile_id, date) DO UPDATE
			SET predicted_kwh  = EXCLUDED.predicted_kwh,
			    weather_factor = EXCLUDED.weather_factor,
                cloud_cover    = EXCLUDED.cloud_cover,
			    efficiency     = EXCLUDED.efficiency,
                delta_wf       = EXCLUDED.delta_wf,
                baseline_type  = EXCLUDED.baseline_type
	`
	normalizedDate := normalizeDate(f.Date)
	_, err := r.db.Exec(query, f.ID, f.UserID, f.SolarProfileID, normalizedDate, f.PredictedKwh, f.WeatherFactor, f.CloudCover, f.Efficiency, f.DeltaWF, f.BaselineType, f.CreatedAt)
	if err != nil {
		return fmt.Errorf("save forecast: %w", err)
	}
	return nil
}

// GetForecastByUserAndDate retrieves the forecast for a specific user and date
func (r *repository) GetForecastByUserAndDate(userID uuid.UUID, solarProfileID uuid.UUID, date time.Time) (*Forecast, error) {
	query := `
		SELECT id, user_id, solar_profile_id, date, predicted_kwh, weather_factor, cloud_cover, efficiency, delta_wf, baseline_type, created_at
		FROM forecasts WHERE user_id = $1 AND solar_profile_id = $2 AND date = $3
	`
	row := r.db.QueryRow(query, userID, solarProfileID, normalizeDate(date))

	f := &Forecast{}
	if err := row.Scan(&f.ID, &f.UserID, &f.SolarProfileID, &f.Date, &f.PredictedKwh, &f.WeatherFactor, &f.CloudCover, &f.Efficiency, &f.DeltaWF, &f.BaselineType, &f.CreatedAt); err != nil {
		return nil, fmt.Errorf("get forecast: %w", err)
	}
	return f, nil
}

// GetAllForecastsByDate returns all forecasts generated for a given date
func (r *repository) GetAllForecastsByDate(date time.Time) ([]*Forecast, error) {
	query := `
		SELECT id, user_id, solar_profile_id, date, predicted_kwh, weather_factor, cloud_cover, efficiency, delta_wf, baseline_type, created_at
		FROM forecasts WHERE date = $1
	`
	rows, err := r.db.Query(query, normalizeDate(date))
	if err != nil {
		return nil, fmt.Errorf("get forecasts by date: %w", err)
	}
	defer rows.Close()

	var forecasts []*Forecast
	for rows.Next() {
		f := &Forecast{}
		if err := rows.Scan(&f.ID, &f.UserID, &f.SolarProfileID, &f.Date, &f.PredictedKwh, &f.WeatherFactor, &f.CloudCover, &f.Efficiency, &f.DeltaWF, &f.BaselineType, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan forecast: %w", err)
		}
		forecasts = append(forecasts, f)
	}
	return forecasts, nil
}

// SaveActualDaily inserts or updates actual daily energy for a user/date.
func (r *repository) SaveActualDaily(a *ActualDaily) error {
	query := `
		INSERT INTO actual_daily (id, user_id, solar_profile_id, date, actual_kwh, source, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, solar_profile_id, date) DO UPDATE
			SET actual_kwh = EXCLUDED.actual_kwh,
			    source = EXCLUDED.source
	`

	_, err := r.db.Exec(query, a.ID, a.UserID, a.SolarProfileID, normalizeDate(a.Date), a.ActualKwh, a.Source, a.CreatedAt)
	if err != nil {
		return fmt.Errorf("save actual daily: %w", err)
	}

	return nil
}

// GetActualDailyByUserAndDate returns actual daily energy for one user and date.
func (r *repository) GetActualDailyByUserAndDate(userID uuid.UUID, date time.Time) (*ActualDaily, error) {
	query := `
		SELECT id, user_id, solar_profile_id, date, actual_kwh, source, created_at
		FROM actual_daily
		WHERE user_id = $1 AND date = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := r.db.QueryRow(query, userID, normalizeDate(date))
	a := &ActualDaily{}
	if err := row.Scan(&a.ID, &a.UserID, &a.SolarProfileID, &a.Date, &a.ActualKwh, &a.Source, &a.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("get actual daily: %w", err)
	}

	return a, nil
}

// GetUserEfficiency returns the current adaptive efficiency value for a user.
func (r *repository) GetUserEfficiency(userID uuid.UUID) (float64, error) {
	query := `SELECT forecast_efficiency FROM users WHERE id = $1`

	var efficiency float64
	if err := r.db.QueryRow(query, userID).Scan(&efficiency); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, sql.ErrNoRows
		}
		return 0, fmt.Errorf("get user efficiency: %w", err)
	}

	return efficiency, nil
}

// UpdateUserEfficiency persists a new adaptive efficiency value for a user.
func (r *repository) UpdateUserEfficiency(userID uuid.UUID, efficiency float64) error {
	query := `UPDATE users SET forecast_efficiency = $2 WHERE id = $1`

	res, err := r.db.Exec(query, userID, efficiency)
	if err != nil {
		return fmt.Errorf("update user efficiency: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected for update user efficiency: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetForecastHistoryByUser returns recent forecasts for a user (last N days).
func (r *repository) GetForecastHistoryByUser(userID uuid.UUID, days int, filter HistoryFilter) ([]*Forecast, error) {
	args := []any{userID, days}
	b := strings.Builder{}
	b.WriteString(`
		SELECT id, user_id, solar_profile_id, date, predicted_kwh, weather_factor, cloud_cover, efficiency, delta_wf, baseline_type, created_at
		FROM forecasts
		WHERE user_id = $1 AND date >= NOW() - INTERVAL '1 day' * $2
	`)

	if filter.SolarProfileID != nil {
		args = append(args, *filter.SolarProfileID)
		b.WriteString(fmt.Sprintf(" AND solar_profile_id = $%d", len(args)))
	}
	if filter.StartDate != nil {
		args = append(args, normalizeDate(*filter.StartDate))
		b.WriteString(fmt.Sprintf(" AND date >= $%d", len(args)))
	}
	if filter.EndDate != nil {
		args = append(args, normalizeDate(*filter.EndDate))
		b.WriteString(fmt.Sprintf(" AND date <= $%d", len(args)))
	}
	b.WriteString(" ORDER BY date DESC")

	if filter.PageSize > 0 {
		b.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2))
		page := filter.Page
		if page < 1 {
			page = 1
		}
		offset := (page - 1) * filter.PageSize
		args = append(args, filter.PageSize, offset)
	}

	rows, err := r.db.Query(b.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("get forecast history: %w", err)
	}
	defer rows.Close()

	forecasts := []*Forecast{}
	for rows.Next() {
		f := &Forecast{}
		if err := rows.Scan(&f.ID, &f.UserID, &f.SolarProfileID, &f.Date, &f.PredictedKwh, &f.WeatherFactor, &f.CloudCover, &f.Efficiency, &f.DeltaWF, &f.BaselineType, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan forecast row: %w", err)
		}
		forecasts = append(forecasts, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("forecast rows error: %w", err)
	}

	return forecasts, nil
}

// GetActualHistoryByUser returns recent actual measurements for a user (last N days).
func (r *repository) GetActualHistoryByUser(userID uuid.UUID, days int, filter HistoryFilter) ([]*ActualDaily, error) {
	args := []any{userID}
	b := strings.Builder{}
	b.WriteString(`
		SELECT id, user_id, solar_profile_id, date, actual_kwh, source, created_at
		FROM actual_daily
		WHERE user_id = $1
	`)

	if filter.StartDate != nil {
		args = append(args, normalizeDate(*filter.StartDate))
		b.WriteString(fmt.Sprintf(" AND date >= $%d", len(args)))
	} else if days > 0 {
		args = append(args, days)
		b.WriteString(fmt.Sprintf(" AND date >= NOW() - INTERVAL '1 day' * $%d", len(args)))
	}

	if filter.SolarProfileID != nil {
		args = append(args, *filter.SolarProfileID)
		b.WriteString(fmt.Sprintf(" AND solar_profile_id = $%d", len(args)))
	}
	if filter.EndDate != nil {
		args = append(args, normalizeDate(*filter.EndDate))
		b.WriteString(fmt.Sprintf(" AND date <= $%d", len(args)))
	}
	b.WriteString(" ORDER BY date DESC")

	if filter.PageSize > 0 {
		b.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2))
		page := filter.Page
		if page < 1 {
			page = 1
		}
		offset := (page - 1) * filter.PageSize
		args = append(args, filter.PageSize, offset)
	}

	rows, err := r.db.Query(b.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("get actual history: %w", err)
	}
	defer rows.Close()

	actuals := []*ActualDaily{}
	for rows.Next() {
		a := &ActualDaily{}
		if err := rows.Scan(&a.ID, &a.UserID, &a.SolarProfileID, &a.Date, &a.ActualKwh, &a.Source, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan actual row: %w", err)
		}
		actuals = append(actuals, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("actual rows error: %w", err)
	}

	return actuals, nil
}

// CountValidActualDays returns the number of days with actual_kwh > 0 for a user/profile
func (r *repository) CountValidActualDays(ctx context.Context, userID uuid.UUID, profileID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM actual_daily WHERE user_id = $1 AND solar_profile_id = $2 AND actual_kwh > 0`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID, profileID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count valid actual days: %w", err)
	}
	return count, nil
}

// normalizeDate strips the time component before persisting or querying DATE columns.
func normalizeDate(input time.Time) time.Time {
	return time.Date(input.Year(), input.Month(), input.Day(), 0, 0, 0, 0, time.UTC)
}

func (r *repository) CountForecastHistoryByUser(userID uuid.UUID, days int, filter HistoryFilter) (int, error) {
	args := []any{userID, days}
	b := strings.Builder{}
	b.WriteString(`
		SELECT COUNT(*)
		FROM forecasts
		WHERE user_id = $1 AND date >= NOW() - INTERVAL '1 day' * $2
	`)

	if filter.SolarProfileID != nil {
		args = append(args, *filter.SolarProfileID)
		b.WriteString(fmt.Sprintf(" AND solar_profile_id = $%d", len(args)))
	}
	if filter.StartDate != nil {
		args = append(args, normalizeDate(*filter.StartDate))
		b.WriteString(fmt.Sprintf(" AND date >= $%d", len(args)))
	}
	if filter.EndDate != nil {
		args = append(args, normalizeDate(*filter.EndDate))
		b.WriteString(fmt.Sprintf(" AND date <= $%d", len(args)))
	}

	var count int
	if err := r.db.QueryRow(b.String(), args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count forecast history: %w", err)
	}
	return count, nil
}

func (r *repository) CountActualHistoryByUser(userID uuid.UUID, days int, filter HistoryFilter) (int, error) {
	args := []any{userID}
	b := strings.Builder{}
	b.WriteString(`
		SELECT COUNT(*)
		FROM actual_daily
		WHERE user_id = $1
	`)

	if filter.StartDate != nil {
		args = append(args, normalizeDate(*filter.StartDate))
		b.WriteString(fmt.Sprintf(" AND date >= $%d", len(args)))
	} else if days > 0 {
		args = append(args, days)
		b.WriteString(fmt.Sprintf(" AND date >= NOW() - INTERVAL '1 day' * $%d", len(args)))
	}

	if filter.SolarProfileID != nil {
		args = append(args, *filter.SolarProfileID)
		b.WriteString(fmt.Sprintf(" AND solar_profile_id = $%d", len(args)))
	}
	if filter.EndDate != nil {
		args = append(args, normalizeDate(*filter.EndDate))
		b.WriteString(fmt.Sprintf(" AND date <= $%d", len(args)))
	}

	var count int
	if err := r.db.QueryRow(b.String(), args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count actual history: %w", err)
	}
	return count, nil
}

func (r *repository) GetPaginatedActuals(userID uuid.UUID, req ListActualsRequest) ([]ActualWithProfile, int, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page < 1 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.Limit

	// Sort and Order validation
	allowedSorts := map[string]bool{"created_at": true, "date": true, "actual_kwh": true}
	if !allowedSorts[req.Sort] {
		req.Sort = "date"
	}
	if strings.ToLower(req.Order) != "asc" {
		req.Order = "desc"
	}

	// 1. Get Total Count
	countQuery := `SELECT COUNT(*) FROM actual_daily a WHERE a.user_id = $1`
	var countArgs = []any{userID}
	if req.StartDate != "" {
		countQuery += " AND a.date >= $2"
		countArgs = append(countArgs, req.StartDate)
	}
	if req.EndDate != "" {
		countQuery += fmt.Sprintf(" AND a.date <= $%d", len(countArgs)+1)
		countArgs = append(countArgs, req.EndDate)
	}

	var totalCount int
	if err := r.db.QueryRow(countQuery, countArgs...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("count actuals: %w", err)
	}

	// 2. Get Data with Join
	whereClause := "WHERE a.user_id = $1"
	dataArgs := []any{userID}
	if req.StartDate != "" {
		whereClause += " AND a.date >= $2"
		dataArgs = append(dataArgs, req.StartDate)
	}
	if req.EndDate != "" {
		whereClause += fmt.Sprintf(" AND a.date <= $%d", len(dataArgs)+1)
		dataArgs = append(dataArgs, req.EndDate)
	}

	dataQuery := fmt.Sprintf(`
		SELECT a.id, a.user_id, a.solar_profile_id, a.date, a.actual_kwh, a.source, a.created_at, 
		       COALESCE(sp.site_name, 'Unknown Site') as site_name
		FROM actual_daily a
		LEFT JOIN solar_profiles sp ON a.solar_profile_id = sp.id
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, req.Sort, req.Order, len(dataArgs)+1, len(dataArgs)+2)

	dataArgs = append(dataArgs, req.Limit, offset)
	rows, err := r.db.Query(dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query paginated actuals: %w", err)
	}
	defer rows.Close()

	var results []ActualWithProfile
	for rows.Next() {
		var ap ActualWithProfile
		if err := rows.Scan(
			&ap.ID, &ap.UserID, &ap.SolarProfileID, &ap.Date, &ap.ActualKwh, &ap.Source, &ap.CreatedAt, 
			&ap.SiteName,
		); err != nil {
			return nil, 0, fmt.Errorf("scan actual with profile: %w", err)
		}
		results = append(results, ap)
	}

	return results, totalCount, nil
}

func (r *repository) GetPaginatedForecasts(userID uuid.UUID, req ListActualsRequest) ([]ForecastWithProfile, int, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page < 1 {
		req.Page = 1
	}
	offset := (req.Page - 1) * req.Limit

	// Sort and Order validation
	allowedSorts := map[string]bool{"created_at": true, "date": true, "predicted_kwh": true, "efficiency": true}
	if !allowedSorts[req.Sort] {
		req.Sort = "date"
	}
	if strings.ToLower(req.Order) != "asc" {
		req.Order = "desc"
	}

	// 1. Get Total Count
	countQuery := `SELECT COUNT(*) FROM forecasts f WHERE f.user_id = $1`
	var countArgs = []any{userID}
	if req.StartDate != "" {
		countQuery += " AND f.date >= $2"
		countArgs = append(countArgs, req.StartDate)
	}
	if req.EndDate != "" {
		countQuery += fmt.Sprintf(" AND f.date <= $%d", len(countArgs)+1)
		countArgs = append(countArgs, req.EndDate)
	}

	var totalCount int
	if err := r.db.QueryRow(countQuery, countArgs...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("count forecasts: %w", err)
	}

	// 2. Get Data with Join
	whereClause := "WHERE f.user_id = $1"
	dataArgs := []any{userID}
	if req.StartDate != "" {
		whereClause += " AND f.date >= $2"
		dataArgs = append(dataArgs, req.StartDate)
	}
	if req.EndDate != "" {
		whereClause += fmt.Sprintf(" AND f.date <= $%d", len(dataArgs)+1)
		dataArgs = append(dataArgs, req.EndDate)
	}

	dataQuery := fmt.Sprintf(`
		SELECT f.id, f.user_id, f.solar_profile_id, f.date, f.predicted_kwh, f.weather_factor, 
		       f.cloud_cover, f.efficiency, f.delta_wf, f.baseline_type, f.created_at,
		       COALESCE(sp.site_name, 'Unknown Site') as site_name
		FROM forecasts f
		LEFT JOIN solar_profiles sp ON f.solar_profile_id = sp.id
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, req.Sort, req.Order, len(dataArgs)+1, len(dataArgs)+2)

	dataArgs = append(dataArgs, req.Limit, offset)
	rows, err := r.db.Query(dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query paginated forecasts: %w", err)
	}
	defer rows.Close()

	var results []ForecastWithProfile
	for rows.Next() {
		var fp ForecastWithProfile
		if err := rows.Scan(
			&fp.ID, &fp.UserID, &fp.SolarProfileID, &fp.Date, &fp.PredictedKwh, &fp.WeatherFactor,
			&fp.CloudCover, &fp.Efficiency, &fp.DeltaWF, &fp.BaselineType, &fp.CreatedAt,
			&fp.SiteName,
		); err != nil {
			return nil, 0, fmt.Errorf("scan forecast with profile: %w", err)
		}
		results = append(results, fp)
	}

	return results, totalCount, nil
}
