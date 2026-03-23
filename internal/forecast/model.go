package forecast

import (
	"time"

	"github.com/google/uuid"
)

// Forecast represents a daily energy forecast for a user
type Forecast struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	SolarProfileID *uuid.UUID `json:"solar_profile_id,omitempty"`
	Date           time.Time  `json:"date"`
	PredictedKwh   float64    `json:"predicted_kwh"`
	WeatherFactor  float64    `json:"weather_factor"`
	CloudCover     int        `json:"cloud_cover"`
	Efficiency     float64    `json:"efficiency"`
	DeltaWF        float64    `json:"delta_wf"`
	BaselineType      string     `json:"baseline_type"`
	WeatherRiskStatus string     `json:"weather_risk_status"`
	CreatedAt         time.Time  `json:"created_at"`
}

// ActualDaily represents actual measured energy for one user and one date.
type ActualDaily struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	SolarProfileID *uuid.UUID `json:"solar_profile_id,omitempty"`
	Date           time.Time  `json:"date"`
	ActualKwh      float64    `json:"actual_kwh"`
	Source         string     `json:"source"`
	CreatedAt      time.Time  `json:"created_at"`
}

// RecordActualRequest contains payload for storing actual daily energy.
type RecordActualRequest struct {
	UserID         uuid.UUID `json:"user_id"`
	SolarProfileID string    `json:"solar_profile_id"`
	Date           string    `json:"date"`
	ActualKwh      float64   `json:"actual_kwh"`
	Source         string    `json:"source"`
}

// HistoryFilter defines optional profile and date-range filters for history endpoints.
type HistoryFilter struct {
	SolarProfileID *uuid.UUID
	StartDate      *time.Time
	EndDate        *time.Time
}

// CalibrationResult summarizes one adaptive efficiency update attempt.
type CalibrationResult struct {
	UserID         uuid.UUID `json:"user_id"`
	Date           time.Time `json:"date"`
	OldEfficiency  float64   `json:"old_efficiency"`
	NewEfficiency  float64   `json:"new_efficiency"`
	PredictedKwh   float64   `json:"predicted_kwh"`
	ActualKwh      float64   `json:"actual_kwh"`
	CorrectionRate float64   `json:"correction_rate"`
	Updated        bool      `json:"updated"`
	Message        string    `json:"message"`
}

// ForecastWithActual combines a forecast with its actual value (if recorded).
type ForecastWithActual struct {
	Forecast *Forecast    `json:"forecast"`
	Actual   *ActualDaily `json:"actual,omitempty"`
	Accuracy float64      `json:"accuracy,omitempty"`
}

// DashboardSummary provides KPI stats for the user dashboard.
type DashboardSummary struct {
	TotalForecastedKwh float64    `json:"total_forecasted_kwh"`
	TotalActualKwh     float64    `json:"total_actual_kwh"`
	AverageForecastKwh float64    `json:"average_forecast_kwh"`
	AverageActualKwh   float64    `json:"average_actual_kwh"`
	CurrentEfficiency  float64    `json:"current_efficiency"`
	AccuracyPercent    float64    `json:"accuracy_percent"`
	ForecastCount      int        `json:"forecast_count"`
	ActualCount        int        `json:"actual_count"`
	LastForecastDate   *time.Time `json:"last_forecast_date,omitempty"`
	LastActualDate     *time.Time `json:"last_actual_date,omitempty"`
}

// ForecastDebugInputs captures the source variables used in one forecast calculation.
type ForecastDebugInputs struct {
	UserID                uuid.UUID  `json:"user_id"`
	SolarProfileID        *uuid.UUID `json:"solar_profile_id,omitempty"`
	Date                  string     `json:"date"`
	CapacityKwp           float64    `json:"capacity_kwp"`
	CloudCoverPercent     int        `json:"cloud_cover_percent"`
	TemperatureC          float64    `json:"temperature_c"`
	ShortwaveRadiationMJ  float64    `json:"shortwave_radiation_mj"`
	PSH                   float64    `json:"psh"`
	WeatherFactor         float64    `json:"weather_factor"`
	ForecastEfficiency    float64    `json:"forecast_efficiency"`
	ExistingForecastKwh   *float64   `json:"existing_forecast_kwh,omitempty"`
	ExistingActualKwh     *float64   `json:"existing_actual_kwh,omitempty"`
	ExistingForecastFound bool       `json:"existing_forecast_found"`
	ExistingActualFound   bool       `json:"existing_actual_found"`
}

// ForecastDebugBreakdown provides one read-only audit payload for forecast calculation.
type ForecastDebugBreakdown struct {
	Inputs       ForecastDebugInputs `json:"inputs"`
	Formula      string              `json:"formula"`
	PredictedKwh float64             `json:"predicted_kwh"`
}
