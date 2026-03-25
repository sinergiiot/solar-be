package report

import (
	"time"

	"github.com/google/uuid"
)

// CO2Summary — Epic 4: Dedicated CO2 Tracker & MRV
type CO2Summary struct {
	UserID          uuid.UUID   `json:"user_id"`
	PeriodStart     time.Time   `json:"period_start"`
	PeriodEnd       time.Time   `json:"period_end"`

	// Totals
	TotalActualKwh   float64 `json:"total_actual_kwh"`
	TotalCO2AvoidedKg float64 `json:"total_co2_avoided_kg"`
	TotalCO2AvoidedTon float64 `json:"total_co2_avoided_ton"`

	// Carbon credit market estimate (IDX Carbon / voluntary)
	CarbonCreditIDR float64 `json:"carbon_credit_idr"` // Rp per ton CO2 × tonnes
	CarbonCreditUSD float64 `json:"carbon_credit_usd"` // USD equivalent

	// Grid factor metadata
	EmissionFactor  float64 `json:"emission_factor_kg_per_kwh"` // kg CO2 / kWh
	GridRegion      string  `json:"grid_region"`                 // JAMALI / Sumatera / etc.
	Standard        string  `json:"methodology_standard"`        // ESDM 2023

	// Period breakdown
	DailyBreakdown []CO2Day `json:"daily_breakdown,omitempty"`
	PlanTier       string   `json:"plan_tier"`
}

type CO2Day struct {
	Date        string  `json:"date"`
	ActualKwh   float64 `json:"actual_kwh"`
	CO2AvoidedKg float64 `json:"co2_avoided_kg"`
}

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

	// Epic 2: Annual & Official metadata
	IsAnnual        bool             `json:"is_annual"`
	MonthlyBreakdown []MonthlySummary `json:"monthly_breakdown,omitempty"`
	OfficialDetails  *OfficialDetails `json:"official_details,omitempty"`
	TotalREC         int              `json:"total_rec"`
}

type MonthlySummary struct {
	Month     string  `json:"month"`
	ActualKwh float64 `json:"actual_kwh"`
	SavingsIDR float64 `json:"savings_idr"`
}

type OfficialDetails struct {
	LetterNumber string `json:"letter_number"`
	Signatory    string `json:"signatory"`    // Nama pejabat (misal: Bambang)
	Title        string `json:"title"`        // Jabatan (misal: Ketua RT / Lurah)
	Organization string `json:"organization"` // Instansi
	OfficialDate time.Time `json:"official_date"`
}

type ReportRequest struct {
	SolarProfileID string `json:"solar_profile_id,omitempty"`
	StartDate      string `json:"start_date"` // YYYY-MM-DD
	EndDate        string `json:"end_date"`   // YYYY-MM-DD
	
	// Epic 2
	IsAnnual       bool   `json:"is_annual"`
	Year           int    `json:"year"`
	OfficialLetter bool   `json:"official_letter"` 
	Signatory      string `json:"signatory"`
	Title          string `json:"title"`
	Organization   string `json:"organization"`
}

// Epic 5: ESG Dashboard Models
type ESGSummary struct {
	UserID           uuid.UUID   `json:"user_id"`
	TotalActualMwh   float64     `json:"total_actual_mwh"`
	TotalCO2SavedTon float64     `json:"total_co2_saved_ton"`
	TotalTreesEq     int         `json:"total_trees_eq"`
	TotalRECCount    int         `json:"total_rec_count"`
	TotalSavingsIDR  float64     `json:"total_savings_idr"`
	CleanEnergyPct   float64     `json:"clean_energy_pct"` // Percentage of target met
	
	SiteBreakdown    []SiteESG   `json:"site_breakdown"`
	YearlyTrend      []ESGMonth  `json:"yearly_trend"`
}

type SiteESG struct {
	ProfileID   uuid.UUID `json:"profile_id"`
	ProfileName string    `json:"profile_name"`
	Location    string    `json:"location"`
	ActualMwh   float64   `json:"actual_mwh"`
	CO2SavedTon float64   `json:"co2_saved_ton"`
	RECReached  int       `json:"rec_reached"`
}

type ESGMonth struct {
	Month       string  `json:"month"`
	ActualMwh   float64 `json:"actual_mwh"`
	CO2SavedTon float64 `json:"co2_saved_ton"`
}

// Epic 3: REC Readiness
type RECReadinessReport struct {
	UserID         uuid.UUID   `json:"user_id"`
	TotalActualMwh float64     `json:"total_actual_mwh"`
	TotalREC       int         `json:"total_rec"`
	SiteBreakdown  []SiteREC   `json:"site_breakdown"`
	MonthlyHistory []RECMonth  `json:"monthly_history"`
	GeneratedAt    time.Time   `json:"generated_at"`
}

type SiteREC struct {
	ProfileName    string  `json:"profile_name"`
	CapacityKwp    float64 `json:"capacity_kwp"`
	Location       string  `json:"location"`
	TotalActualMwh float64 `json:"total_actual_mwh"`
	RECContribution int     `json:"rec_contribution"`
}

type RECMonth struct {
	Month     string  `json:"month"`
	ActualMwh float64 `json:"actual_mwh"`
}
