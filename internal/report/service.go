package report

import (
	"context"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/forecast"
	"github.com/akbarsenawijaya/solar-forecast/internal/solar"
	"github.com/google/uuid"
)

type Service interface {
	GenerateReport(ctx context.Context, userID uuid.UUID, planTier string, req ReportRequest) (*EnergyReport, error)
	GenerateReportPDF(report *EnergyReport, userName string, writer io.Writer) error
}

type service struct {
	forecastService forecast.Service
	solarService    solar.Service
}

func NewService(forecastSvc forecast.Service, solarSvc solar.Service) Service {
	return &service{
		forecastService: forecastSvc,
		solarService:    solarSvc,
	}
}

func (s *service) GenerateReport(ctx context.Context, userID uuid.UUID, planTier string, req ReportRequest) (*EnergyReport, error) {
	startDate, err := time.Parse(time.DateOnly, req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date")
	}
	endDate, err := time.Parse(time.DateOnly, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date")
	}

	if startDate.After(endDate) {
		return nil, fmt.Errorf("start_date cannot be after end_date")
	}

	diff := endDate.Sub(startDate)
	days := int(diff.Hours()/24) + 1

	var solarProfileID *uuid.UUID
	if req.SolarProfileID != "" {
		id, err := uuid.Parse(req.SolarProfileID)
		if err == nil {
			solarProfileID = &id
		}
	}

	filter := forecast.HistoryFilter{
		SolarProfileID: solarProfileID,
		StartDate:      &startDate,
		EndDate:        &endDate,
	}

	actuals, err := s.forecastService.GetActualHistory(userID, planTier, 366, filter)
	if err != nil {
		return nil, err
	}

	forecasts, err := s.forecastService.GetForecastHistory(userID, planTier, 366, filter)
	if err != nil {
		return nil, err
	}

	var totalActualKwh float64
	for _, a := range actuals.Items {
		totalActualKwh += a.ActualKwh
	}

	var totalForecastedKwh float64
	for _, f := range forecasts.Items {
		totalForecastedKwh += f.PredictedKwh
	}

	// Calculate CO2 factor based on location if profile is selected
	co2Factor := 0.78 // Generic fallback
	if solarProfileID != nil {
		profile, err := s.solarService.GetSolarProfileByIDAndUserID(*solarProfileID, userID)
		if err == nil {
			co2Factor = getEmissionFactor(profile.Lat, profile.Lng)
		}
	}

	report := &EnergyReport{
		UserID:             userID,
		SolarProfileID:    solarProfileID,
		PeriodStart:        startDate,
		PeriodEnd:          endDate,
		TotalForecastedKwh: totalForecastedKwh,
		TotalActualKwh:     totalActualKwh,
		TotalSavingsIDR:    totalActualKwh * 1444, // Fixed tariff for now
		TotalCO2AvoidedKg:  totalActualKwh * co2Factor,
		DataCoveragePct:    math.Min(100, (float64(len(actuals.Items))/float64(days))*100),
		PlanTier:           planTier,
		CreatedAt:          time.Now().UTC(),
	}

	return report, nil
}

// Re-importing getEmissionFactor from scheduler or consolidate it.
func getEmissionFactor(lat, lng float64) float64 {
	if lat >= -9.0 && lat <= -5.0 && lng >= 105.0 && lng <= 116.0 {
		return 0.85 // Java-Bali
	}
	if lat >= -6.0 && lat <= 6.0 && lng >= 95.0 && lng <= 106.0 {
		return 0.75 // Sumatra
	}
	if lat >= -4.0 && lat <= 5.0 && lng >= 108.0 && lng <= 119.0 {
		return 0.80 // Kalimantan
	}
	if lat >= -6.0 && lat <= 2.0 && lng >= 118.0 && lng <= 125.0 {
		return 0.65 // Sulawesi
	}
	if lat >= -11.0 && lat <= 0.0 && lng >= 125.0 && lng <= 141.0 {
		return 0.70 // Maluku-Papua
	}
	return 0.78
}
