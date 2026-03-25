//go:build ignore

package main

import (
	"context"
	"fmt"
	"io"
	"net/http/httptest"

	"github.com/akbarsenawijaya/solar-forecast/internal/config"
	"github.com/akbarsenawijaya/solar-forecast/internal/device"
	"github.com/akbarsenawijaya/solar-forecast/internal/forecast"
	"github.com/akbarsenawijaya/solar-forecast/internal/rec"
	"github.com/akbarsenawijaya/solar-forecast/internal/solar"
	"github.com/akbarsenawijaya/solar-forecast/internal/tier"
	"github.com/akbarsenawijaya/solar-forecast/internal/weather"
	"github.com/akbarsenawijaya/solar-forecast/internal/weatherbaseline"
	"github.com/akbarsenawijaya/solar-forecast/pkg/ctxkeys"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

	h := forecast.NewHandler(svc, "")

	var idStr string
	_ = db.QueryRow("SELECT id FROM users WHERE email='wijayasenaakbar@gmail.com'").Scan(&idStr)
	userID := uuid.MustParse(idStr)

	runReq := func(planTier string) {
		req := httptest.NewRequest("GET", "/forecast/history?start_date=2026-02-21&end_date=2026-03-23", nil)
		ctx := req.Context()
		ctx = context.WithValue(ctx, ctxkeys.UserRole, "customer")
		ctx = context.WithValue(ctx, ctxkeys.UserID, userID)
		ctx = context.WithValue(ctx, tier.TierContextKey, planTier)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		router := chi.NewRouter()
		router.Get("/forecast/history", h.GetForecastHistory)
		router.ServeHTTP(w, req)

		res := w.Result()
		body, _ := io.ReadAll(res.Body)
		fmt.Println("Tier:", planTier, "Response bytes:", len(body), "Body snippet:", string(body)[:min(100, len(body))])
	}

	runReq("free")
	runReq("pro")
	runReq("enterprise")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
