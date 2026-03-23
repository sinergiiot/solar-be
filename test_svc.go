package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/akbarsenawijaya/solar-forecast/internal/config"
	"github.com/akbarsenawijaya/solar-forecast/internal/device"
	"github.com/akbarsenawijaya/solar-forecast/internal/forecast"
	"github.com/akbarsenawijaya/solar-forecast/internal/solar"
	"github.com/akbarsenawijaya/solar-forecast/internal/weather"
	"github.com/akbarsenawijaya/solar-forecast/internal/rec"
	"github.com/akbarsenawijaya/solar-forecast/internal/weatherbaseline"
)

func main() {
	cfg := config.Load()
	db := config.OpenDB(cfg.DBUrl)
	defer db.Close()

	repo := forecast.NewRepository(db)
	solarSvc := solar.NewService(solar.NewRepository(db))
	deviceSvc := device.NewService(device.NewRepository(db))
	weatherSvc := weather.NewService(weather.NewRepository(db), "")
	wb := weatherbaseline.NewService(weatherbaseline.NewRepository(db), "")
	recSvc := rec.NewService(rec.NewRepository(db), nil, nil)

	svc := forecast.NewService(repo, solarSvc, deviceSvc, weatherSvc, recSvc, wb)

	var idStr string
	_ = db.QueryRow("SELECT id FROM users WHERE email='wijayasenaakbar@gmail.com'").Scan(&idStr)
	userID := uuid.MustParse(idStr)

	start, _ := time.Parse("2006-01-02", "2026-02-21")
	end, _ := time.Parse("2006-01-02", "2026-03-23")

	filter := forecast.HistoryFilter{StartDate: &start, EndDate: &end}

	f1, _ := svc.GetForecastHistory(userID, "pro", 90, filter)
	f2, _ := svc.GetForecastHistory(userID, "free", 90, filter)
	f3, _ := svc.GetForecastHistory(userID, "enterprise", 90, filter)

	fmt.Println("Pro:", len(f1.Items))
	fmt.Println("Free:", len(f2.Items))
	fmt.Println("Enterprise:", len(f3.Items))
}
