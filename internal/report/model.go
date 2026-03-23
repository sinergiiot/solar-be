package report

import (
	"time"

	"github.com/google/uuid"
)

type EnergyReport struct {
	UserID         uuid.UUID  `json:"user_id"`
	SolarProfileID *uuid.UUID `json:"solar_profile_id,omitempty"`
	PeriodStart    time.Time  `json:"period_start"`
	PeriodEnd      time.Time  `json:"period_end"`
	
	TotalForecastedKwh float64 `json:"total_forecasted_kwh"`
	TotalActualKwh     float64 `json:"total_actual_kwh"`
	
	TotalSavingsIDR    float64 `json:"total_savings_idr"`
	TotalCO2AvoidedKg  float64 `json:"total_co2_avoided_kg"`
	
	DataCoveragePct    float64 `json:"data_coverage_pct"` // % of days with actual data
	PlanTier           string  `json:"plan_tier"`
	CreatedAt          time.Time `json:"created_at"`
}

type ReportRequest struct {
	SolarProfileID string `json:"solar_profile_id,omitempty"`
	StartDate      string `json:"start_date"` // YYYY-MM-DD
	EndDate        string `json:"end_date"`   // YYYY-MM-DD
}
