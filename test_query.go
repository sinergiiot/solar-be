package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/akbarsenawijaya/solar-forecast/internal/forecast"
	"github.com/akbarsenawijaya/solar-forecast/internal/config"
)

func main() {
	cfg := config.Load()
	db := config.OpenDB(cfg.DBUrl)
	defer db.Close()

	repo := forecast.NewRepository(db)

	userID := uuid.MustParse("00000000-0000-0000-0000-000000000000")
	var idStr string
	err := db.QueryRow("SELECT id FROM users WHERE email='wijayasenaakbar@gmail.com'").Scan(&idStr)
	if err != nil {
		panic(err)
	}
	userID = uuid.MustParse(idStr)

	start, _ := time.Parse("2006-01-02", "2026-02-21")
	end, _ := time.Parse("2006-01-02", "2026-03-23")

	filter := forecast.HistoryFilter{
		StartDate: &start,
		EndDate:   &end,
	}

	forecasts, err := repo.GetForecastHistoryByUser(userID, 90, filter)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Forecasts count:", len(forecasts))

	forecastsFree, err := repo.GetForecastHistoryByUser(userID, 7, filter)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Forecasts count (free 7 days):", len(forecastsFree))
}
