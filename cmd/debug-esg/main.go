package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/akbarsenawijaya/solar-forecast/internal/device"
	"github.com/akbarsenawijaya/solar-forecast/internal/forecast"
	"github.com/akbarsenawijaya/solar-forecast/internal/notification"
	"github.com/akbarsenawijaya/solar-forecast/internal/rec"
	"github.com/akbarsenawijaya/solar-forecast/internal/report"
	"github.com/akbarsenawijaya/solar-forecast/internal/solar"
	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/akbarsenawijaya/solar-forecast/internal/weather"
	"github.com/akbarsenawijaya/solar-forecast/internal/weatherbaseline"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const USER_ID = "00000000-0000-0000-0000-000000000005"

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	db, _ := sql.Open("postgres", dbURL)
	defer db.Close()

	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo)
	solarRepo := solar.NewRepository(db)
	solarSvc := solar.NewService(solarRepo)
	deviceRepo := device.NewRepository(db)
	deviceSvc := device.NewService(deviceRepo)
	weatherRepo := weather.NewRepository(db)
	weatherSvc := weather.NewService(weatherRepo, "")
	recRepo := rec.NewRepository(db)
	notifRepo := notification.NewRepository(db)
	notifSvc := notification.NewService(notifRepo, "", "0", "", "", "", "", "", "", "", "")
	recSvc := rec.NewService(recRepo, userSvc, notifSvc)
	weatherBaselineRepo := weatherbaseline.NewRepository(db)
	weatherBaselineSvc := weatherbaseline.NewService(weatherBaselineRepo, "")
	forecastRepo := forecast.NewRepository(db)
	forecastSvc := forecast.NewService(forecastRepo, solarSvc, deviceSvc, weatherSvc, recSvc, weatherBaselineSvc)
	
	reportSvc := report.NewService(forecastSvc, solarSvc, recSvc, userSvc)

	uid := uuid.MustParse(USER_ID)
	summary, err := reportSvc.GetESGSummary(context.Background(), uid, "enterprise", 2025)
	if err != nil {
		log.Fatalf("failed to get ESG summary: %v", err)
	}

	fmt.Printf("ESG Summary for %d sites:\n", len(summary.SiteBreakdown))
	fmt.Printf("Total Energy: %.3f MWh\n", summary.TotalActualMwh)
	fmt.Printf("Total CO2 Saved: %.2f Tons\n", summary.TotalCO2SavedTon)
	fmt.Printf("Total Trees Eq: %d\n", summary.TotalTreesEq)
	fmt.Printf("Total REC: %d\n", summary.TotalRECCount)
	
	fmt.Println("\nSite Breakdown:")
	for _, s := range summary.SiteBreakdown {
		fmt.Printf("- %s (%s): %.3f MWh, %.2f Tons CO2, %d REC\n", s.ProfileName, s.Location, s.ActualMwh, s.CO2SavedTon, s.RECReached)
	}
}
