package report

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/forecast"
	"github.com/akbarsenawijaya/solar-forecast/internal/rec"
	"github.com/akbarsenawijaya/solar-forecast/internal/solar"
	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/google/uuid"
)

type Service interface {
	GenerateReport(ctx context.Context, userID uuid.UUID, planTier string, req ReportRequest) (*EnergyReport, error)
	GenerateReportPDF(report *EnergyReport, userObj *user.User, writer io.Writer) error
	GenerateRECPDF(ctx context.Context, userID uuid.UUID, writer io.Writer) error
	GetESGSummary(ctx context.Context, userID uuid.UUID, planTier string, year int) (*ESGSummary, error)
	GetCO2Summary(ctx context.Context, userID uuid.UUID, planTier string, req ReportRequest) (*CO2Summary, error)
	GenerateMRVPDF(summary *CO2Summary, userObj *user.User, writer io.Writer) error
	GenerateESGReportPDF(summary *ESGSummary, userObj *user.User, year int, writer io.Writer) error
	GetRECReadinessReport(ctx context.Context, userID uuid.UUID) (*RECReadinessReport, error)
	GenerateRECReadinessReportPDF(report *RECReadinessReport, userObj *user.User, writer io.Writer) error
	GenerateCSVHistory(ctx context.Context, userID uuid.UUID, planTier string, req ReportRequest, writer io.Writer) error
}

type service struct {
	forecastService forecast.Service
	solarService    solar.Service
	recService      rec.Service
	userSvc         user.Service
}

func NewService(forecastSvc forecast.Service, solarSvc solar.Service, recSvc rec.Service, userSvc user.Service) Service {
	return &service{
		forecastService: forecastSvc,
		solarService:    solarSvc,
		recService:      recSvc,
		userSvc:         userSvc,
	}
}

func (s *service) GenerateReport(ctx context.Context, userID uuid.UUID, planTier string, req ReportRequest) (*EnergyReport, error) {
	var startDate, endDate time.Time
	var err error

	if req.IsAnnual {
		if planTier == "free" {
			return nil, fmt.Errorf("annual reports are available only for Pro/Enterprise users")
		}
		year := req.Year
		if year == 0 {
			year = time.Now().Year()
		}
		startDate = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)
	} else {
		startDate, err = time.Parse(time.DateOnly, req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date")
		}
		endDate, err = time.Parse(time.DateOnly, req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date")
		}
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

	// For annual, we fetch larger chunks
	maxRecords := 366
	actuals, err := s.forecastService.GetActualHistory(userID, planTier, maxRecords, filter)
	if err != nil {
		return nil, err
	}

	forecasts, err := s.forecastService.GetForecastHistory(userID, planTier, maxRecords, filter)
	if err != nil {
		return nil, err
	}

	var totalActualKwh float64
	monthlyMap := make(map[string]*MonthlySummary)

	for _, a := range actuals.Items {
		totalActualKwh += a.ActualKwh
		if req.IsAnnual {
			mKey := a.Date.Format("January")
			if _, ok := monthlyMap[mKey]; !ok {
				monthlyMap[mKey] = &MonthlySummary{Month: mKey}
			}
			monthlyMap[mKey].ActualKwh += a.ActualKwh
			monthlyMap[mKey].SavingsIDR += a.ActualKwh * 1444
		}
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
		TotalSavingsIDR:    totalActualKwh * 1444,
		TotalCO2AvoidedKg:  totalActualKwh * co2Factor,
		DataCoveragePct:    math.Min(100, (float64(len(actuals.Items))/float64(days))*100),
		PlanTier:           planTier,
		CreatedAt:          time.Now().UTC(),
		IsAnnual:           req.IsAnnual,
	}

	totalMwh, _ := s.recService.GetTotalMwhForUser(ctx, userID)
	report.TotalREC = int(totalMwh)

	if req.IsAnnual {
		months := []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
		for _, m := range months {
			if summ, ok := monthlyMap[m]; ok {
				report.MonthlyBreakdown = append(report.MonthlyBreakdown, *summ)
			}
		}

		if req.OfficialLetter {
			report.OfficialDetails = &OfficialDetails{
				LetterNumber: fmt.Sprintf("SOLAR/%d/%s/%04d", req.Year, userID.String()[:4], randInt(1000, 9999)),
				Signatory:    req.Signatory,
				Title:        req.Title,
				Organization: req.Organization,
				OfficialDate: time.Now(),
			}
		}
	}

	return report, nil
}

// For random letter sub-number
func randInt(min, max int) int {
	return min + time.Now().Nanosecond()%(max-min)
}

// getEmissionFactor returns the CO2 emission factor (kg CO2/kWh) based on ESDM 2023 data for Indonesia.
func getEmissionFactor(lat, lng float64) float64 {
	// Indonesia Grid-specific Factors (Ref: ESDM 2023 / Nilai Emisi GRK Sektor Pembangkit Listrik)
	
	// Jawa-Madura-Bali (JAMALI)
	if lat >= -9.0 && lat <= -5.0 && lng >= 105.0 && lng <= 116.0 {
		return 0.87 // Jamali Grid is currently the most carbon-intensive due to coal dominance
	}
	// Sumatera
	if lat >= -6.0 && lat <= 6.0 && lng >= 95.0 && lng <= 106.0 {
		return 0.81
	}
	// Kalimantan (Interconnected)
	if lat >= -4.0 && lat <= 5.0 && lng >= 108.0 && lng <= 119.0 {
		return 0.84
	}
	// Sulawesi
	if lat >= -6.0 && lat <= 2.0 && lng >= 118.0 && lng <= 127.0 {
		return 0.72
	}
	// Nusa Tenggara & Maluku & Papua
	if lat >= -11.0 && lat <= 0.0 && lng >= 116.0 && lng <= 141.0 {
		return 0.68
	}
	return 0.78 // Default National Average
}

func (s *service) GetESGSummary(ctx context.Context, userID uuid.UUID, planTier string, year int) (*ESGSummary, error) {
	if planTier != "enterprise" {
		return nil, fmt.Errorf("ESG dashboard is only available for Enterprise users")
	}

	if year == 0 {
		year = time.Now().Year()
	}

	// 1. Get all solar profiles for this user
	profiles, err := s.solarService.GetSolarProfilesByUserID(userID)
	if err != nil {
		return nil, err
	}

	summary := &ESGSummary{
		UserID:        userID,
		SiteBreakdown: make([]SiteESG, 0),
		YearlyTrend:   make([]ESGMonth, 12),
	}

	months := []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
	for i, m := range months {
		summary.YearlyTrend[i] = ESGMonth{Month: m}
	}

	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 999, time.UTC)

	// 2. Iterate through profiles and aggregate data
	for _, p := range profiles {
		filter := forecast.HistoryFilter{
			SolarProfileID: &p.ID,
			StartDate:      &startDate,
			EndDate:        &endDate,
		}

		actuals, err := s.forecastService.GetActualHistory(userID, planTier, 400, filter)
		if err != nil {
			continue
		}

		co2Factor := getEmissionFactor(p.Lat, p.Lng)
		siteActualKwh := 0.0

		for _, a := range actuals.Items {
			siteActualKwh += a.ActualKwh
			
			// Monthly aggregation for trend
			mIndex := int(a.Date.Month()) - 1
			summary.YearlyTrend[mIndex].ActualMwh += a.ActualKwh / 1000.0
			summary.YearlyTrend[mIndex].CO2SavedTon += (a.ActualKwh * co2Factor) / 1000.0
		}

		siteActualMwh := siteActualKwh / 1000.0
		siteCO2Ton := (siteActualKwh * co2Factor) / 1000.0
		
		summary.TotalActualMwh += siteActualMwh
		summary.TotalCO2SavedTon += siteCO2Ton
		summary.TotalSavingsIDR += siteActualKwh * 1444

		summary.SiteBreakdown = append(summary.SiteBreakdown, SiteESG{
			ProfileID:   p.ID,
			ProfileName: p.SiteName,
			Location:    fmt.Sprintf("%.4f, %.4f", p.Lat, p.Lng),
			ActualMwh:   siteActualMwh,
			CO2SavedTon: siteCO2Ton,
			RECReached:  int(siteActualMwh),
		})
	}

	// 3. Global Stats
	summary.TotalTreesEq = int(summary.TotalCO2SavedTon * 1000 / 20) // 20kg per tree/year
	summary.TotalRECCount = int(summary.TotalActualMwh)
	summary.CleanEnergyPct = math.Min(100, (summary.TotalActualMwh / (float64(len(profiles)) * 1.5)) * 100) // Arbitrary benchmark: 1.5MWh per site/year

	return summary, nil
}

// getGridRegionName maps lat/lng to human-readable grid region name.
func getGridRegionName(lat, lng float64) string {
	if lat >= -9.0 && lat <= -5.0 && lng >= 105.0 && lng <= 116.0 {
		return "JAMALI (Jawa-Madura-Bali)"
	}
	if lat >= -6.0 && lat <= 6.0 && lng >= 95.0 && lng <= 106.0 {
		return "Sumatera"
	}
	if lat >= -4.0 && lat <= 5.0 && lng >= 108.0 && lng <= 119.0 {
		return "Kalimantan"
	}
	if lat >= -6.0 && lat <= 2.0 && lng >= 118.0 && lng <= 127.0 {
		return "Sulawesi"
	}
	if lat >= -11.0 && lat <= 0.0 && lng >= 116.0 && lng <= 141.0 {
		return "Nusa Tenggara / Maluku / Papua"
	}
	return "Indonesia (National Average)"
}

// GetCO2Summary returns a detailed CO2 avoided summary for the requested period (Epic 4).
func (s *service) GetCO2Summary(ctx context.Context, userID uuid.UUID, planTier string, req ReportRequest) (*CO2Summary, error) {
	if planTier == "free" {
		return nil, fmt.Errorf("CO2 detailed tracker is available for Pro/Enterprise users")
	}

	var startDate, endDate time.Time
	var err error

	if req.IsAnnual {
		year := req.Year
		if year == 0 {
			year = time.Now().Year()
		}
		startDate = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate = time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)
	} else {
		startDate, err = time.Parse(time.DateOnly, req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date")
		}
		endDate, err = time.Parse(time.DateOnly, req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date")
		}
	}

	var solarProfileID *uuid.UUID
	if req.SolarProfileID != "" {
		id, err := uuid.Parse(req.SolarProfileID)
		if err == nil {
			solarProfileID = &id
		}
	}

	// Determine emission factor and grid region
	emissionFactor := 0.78
	gridRegion := "Indonesia (National Average)"
	if solarProfileID != nil {
		profile, err := s.solarService.GetSolarProfileByIDAndUserID(*solarProfileID, userID)
		if err == nil {
			emissionFactor = getEmissionFactor(profile.Lat, profile.Lng)
			gridRegion = getGridRegionName(profile.Lat, profile.Lng)
		}
	}

	filter := forecast.HistoryFilter{
		SolarProfileID: solarProfileID,
		StartDate:      &startDate,
		EndDate:        &endDate,
	}

	actuals, err := s.forecastService.GetActualHistory(userID, planTier, 400, filter)
	if err != nil {
		return nil, err
	}

	summary := &CO2Summary{
		UserID:         userID,
		PeriodStart:    startDate,
		PeriodEnd:      endDate,
		EmissionFactor: emissionFactor,
		GridRegion:     gridRegion,
		Standard:       "ESDM 2023 - Faktor Emisi GRK Sektor Pembangkit Listrik Indonesia",
		PlanTier:       planTier,
	}

	for _, a := range actuals.Items {
		co2kg := a.ActualKwh * emissionFactor
		summary.TotalActualKwh += a.ActualKwh
		summary.TotalCO2AvoidedKg += co2kg
		summary.DailyBreakdown = append(summary.DailyBreakdown, CO2Day{
			Date:         a.Date.Format("2006-01-02"),
			ActualKwh:    a.ActualKwh,
			CO2AvoidedKg: co2kg,
		})
	}

	summary.TotalCO2AvoidedTon = summary.TotalCO2AvoidedKg / 1000.0

	// Carbon credit estimates:
	// IDX Carbon voluntary market ~ Rp 30,000/ton CO2 (conserv. 2024 estimate)
	// Voluntary global market ~ USD 5/ton
	const carbonPriceIDR = 30000.0
	const carbonPriceUSD = 5.0
	summary.CarbonCreditIDR = summary.TotalCO2AvoidedTon * carbonPriceIDR
	summary.CarbonCreditUSD = summary.TotalCO2AvoidedTon * carbonPriceUSD

	return summary, nil
}

// GenerateMRVPDF delegates to the pdf.go implementation (Epic 4).
func (s *service) GenerateMRVPDF(summary *CO2Summary, userObj *user.User, writer io.Writer) error {
	return generateMRVReport(summary, userObj, writer)
}

// GenerateESGReportPDF delegates to the pdf.go implementation (Epic 5).
func (s *service) GenerateESGReportPDF(summary *ESGSummary, userObj *user.User, year int, writer io.Writer) error {
	return generateESGReport(summary, userObj, year, writer)
}

// GenerateCSVHistory generates a CSV file with daily actual production, savings, and CO2.
func (s *service) GenerateCSVHistory(ctx context.Context, userID uuid.UUID, planTier string, req ReportRequest, writer io.Writer) error {
	if planTier == "free" {
		return fmt.Errorf("CSV Export is only available for Pro and Enterprise users")
	}

	filter := forecast.HistoryFilter{
		Page:     1,
		PageSize: 1000, // Large enough for export
	}

	if req.SolarProfileID != "" {
		pid, err := uuid.Parse(req.SolarProfileID)
		if err == nil {
			filter.SolarProfileID = &pid
		}
	}
	if req.StartDate != "" {
		t, err := time.Parse(time.DateOnly, req.StartDate)
		if err == nil {
			filter.StartDate = &t
		}
	}
	if req.EndDate != "" {
		t, err := time.Parse(time.DateOnly, req.EndDate)
		if err == nil {
			filter.EndDate = &t
		}
	}

	// 1. Get history data from forecast service
	historyResp, err := s.forecastService.GetActualHistory(userID, planTier, 0, filter)
	if err != nil {
		return fmt.Errorf("failed to fetch history: %w", err)
	}

	// 2. Write CSV Header
	csvWriter := csv.NewWriter(writer)
	header := []string{"Date", "Profile ID", "Production (kWh)", "Savings (IDR)", "CO2 Avoided (kg)", "Source"}
	if err := csvWriter.Write(header); err != nil {
		return err
	}

	// 3. Write data rows
	for _, h := range historyResp.Items {
		co2 := h.ActualKwh * 0.78 // Default
		profileIDStr := "All"
		if h.SolarProfileID != nil {
			profileIDStr = h.SolarProfileID.String()
			p, err := s.solarService.GetSolarProfileByIDAndUserID(*h.SolarProfileID, userID)
			if err == nil {
				co2 = h.ActualKwh * getEmissionFactor(p.Lat, p.Lng)
			}
		}

		row := []string{
			h.Date.Format("2006-01-02"),
			profileIDStr,
			fmt.Sprintf("%.3f", h.ActualKwh),
			fmt.Sprintf("%.0f", h.ActualKwh*1500),
			fmt.Sprintf("%.3f", co2),
			h.Source,
		}
		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}

	csvWriter.Flush()
	return csvWriter.Error()
}

func (s *service) GetRECReadinessReport(ctx context.Context, userID uuid.UUID) (*RECReadinessReport, error) {
	// 1. Get total MWh from rec service
	totalMwh, err := s.recService.GetTotalMwhForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. Get site breakdown
	profiles, err := s.solarService.GetSolarProfilesByUserID(userID)
	if err != nil {
		return nil, err
	}

	var siteBreakdown []SiteREC
	for _, p := range profiles {
		// Get accumulator for this profile
		acc, err := s.recService.GetAccumulator(ctx, userID, &p.ID)
		if err == nil && acc != nil {
			siteBreakdown = append(siteBreakdown, SiteREC{
				ProfileName:    p.SiteName,
				CapacityKwp:    p.CapacityKwp,
				Location:       fmt.Sprintf("%.4f, %.4f", p.Lat, p.Lng),
				TotalActualMwh: acc.CumulativeKwh / 1000.0,
				RECContribution: int(acc.CumulativeKwh / 1000.0),
			})
		}
	}

	// 3. Get monthly history for last 12 months
	history, err := s.forecastService.GetActualHistory(userID, "pro", 365, forecast.HistoryFilter{
		Page:     1,
		PageSize: 400,
	})
	
	monthlyMap := make(map[string]float64)
	if err == nil {
		for _, h := range history.Items {
			monthKey := h.Date.Format("2006-01")
			monthlyMap[monthKey] += h.ActualKwh
		}
	}

	// Convert map to sorted list
	var monthlyHistory []RECMonth
	// Simple last 12 months
	now := time.Now()
	for i := 11; i >= 0; i-- {
		m := now.AddDate(0, -i, 0)
		key := m.Format("2006-01")
		monthlyHistory = append(monthlyHistory, RECMonth{
			Month:     m.Format("Jan 2006"),
			ActualMwh: monthlyMap[key] / 1000.0,
		})
	}

	return &RECReadinessReport{
		UserID:         userID,
		TotalActualMwh: totalMwh,
		TotalREC:       int(totalMwh),
		SiteBreakdown:  siteBreakdown,
		MonthlyHistory: monthlyHistory,
		GeneratedAt:    time.Now(),
	}, nil
}

func (s *service) GenerateRECReadinessReportPDF(report *RECReadinessReport, userObj *user.User, writer io.Writer) error {
	return generateRECReadinessReport(report, userObj, writer)
}
