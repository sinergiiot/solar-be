package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/akbarsenawijaya/solar-forecast/internal/forecast"
	"github.com/akbarsenawijaya/solar-forecast/internal/report"
	"github.com/akbarsenawijaya/solar-forecast/internal/solar"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	db, _ := sql.Open("postgres", dbURL)
	defer db.Close()

	solarRepo := solar.NewRepository(db)
	solarSvc := solar.NewService(solarRepo)
	forecastRepo := forecast.NewRepository(db)
	forecastSvc := forecast.NewService(forecastRepo, solarSvc, nil, nil, nil, nil)
	reportSvc := report.NewService(forecastSvc, solarSvc, nil, nil)

	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	summary, err := reportSvc.GetESGSummary(context.Background(), uid, "enterprise", 2026)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ESG Summary for %d sites:\n", len(summary.SiteBreakdown))
	fmt.Printf("Total Energy: %.2f MWh\n", summary.TotalActualMwh)
	fmt.Printf("Total CO2 Offset: %.2f Ton\n", summary.TotalCO2SavedTon)
	fmt.Printf("Total Savings: IDR %.0f\n", summary.TotalSavingsIDR)
	fmt.Printf("Trees Equivalent: %d\n", summary.TotalTreesEq)
}
