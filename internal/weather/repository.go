package weather

import (
	"database/sql"
	"fmt"
	"time"
)

// Repository defines data access methods for weather data
type Repository interface {
	SaveWeatherDaily(w *WeatherDaily) error
	GetWeatherByDateAndLocation(date time.Time, lat, lng float64) (*WeatherDaily, error)
}

type repository struct {
	db *sql.DB
}

// NewRepository creates a new weather repository
func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// SaveWeatherDaily inserts or updates a weather record for a given date and location
func (r *repository) SaveWeatherDaily(w *WeatherDaily) error {
	query := `
		INSERT INTO weather_daily (id, date, lat, lng, cloud_cover, cloud_cover_mean, temperature, shortwave_radiation_mj, raw_json, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (date, lat, lng) DO UPDATE
			SET cloud_cover = EXCLUDED.cloud_cover,
			    cloud_cover_mean = EXCLUDED.cloud_cover_mean,
			    temperature = EXCLUDED.temperature,
			    shortwave_radiation_mj = EXCLUDED.shortwave_radiation_mj,
			    raw_json    = EXCLUDED.raw_json
	`
	_, err := r.db.Exec(query, w.ID, normalizeDate(w.Date), w.Lat, w.Lng, w.CloudCover, w.CloudCoverMean, w.Temperature, w.ShortwaveRadiationMJ, w.RawJSON, w.CreatedAt)
	if err != nil {
		return fmt.Errorf("save weather daily: %w", err)
	}
	return nil
}

// GetWeatherByDateAndLocation retrieves cached weather data for a specific date and location
func (r *repository) GetWeatherByDateAndLocation(date time.Time, lat, lng float64) (*WeatherDaily, error) {
	query := `
		SELECT id, date, lat, lng, cloud_cover, COALESCE(cloud_cover_mean, cloud_cover), temperature, COALESCE(shortwave_radiation_mj, 0), raw_json, created_at
		FROM weather_daily
		WHERE date = $1 AND lat = $2 AND lng = $3
	`
	row := r.db.QueryRow(query, normalizeDate(date), lat, lng)

	w := &WeatherDaily{}
	if err := row.Scan(&w.ID, &w.Date, &w.Lat, &w.Lng, &w.CloudCover, &w.CloudCoverMean, &w.Temperature, &w.ShortwaveRadiationMJ, &w.RawJSON, &w.CreatedAt); err != nil {
		return nil, err
	}
	return w, nil
}

// normalizeDate strips the time component before persisting or querying DATE columns.
func normalizeDate(input time.Time) time.Time {
	return time.Date(input.Year(), input.Month(), input.Day(), 0, 0, 0, 0, time.UTC)
}
