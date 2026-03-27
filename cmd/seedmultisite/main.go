package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/device"
	"github.com/akbarsenawijaya/solar-forecast/internal/forecast"
	"github.com/akbarsenawijaya/solar-forecast/internal/notification"
	"github.com/akbarsenawijaya/solar-forecast/internal/rec"
	"github.com/akbarsenawijaya/solar-forecast/internal/solar"
	"github.com/akbarsenawijaya/solar-forecast/internal/tier"
	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/akbarsenawijaya/solar-forecast/internal/weather"
	"github.com/akbarsenawijaya/solar-forecast/internal/weatherbaseline"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const USER_ID = "00000000-0000-0000-0000-000000000001"

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	uid := uuid.MustParse(USER_ID)

	// 1. Setup Enterprise User
	_, _ = db.Exec("INSERT INTO users (id, name, email, role, email_verified, password_hash, forecast_efficiency, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (id) DO UPDATE SET role = $4", 
		uid, "Enterprise Test User", "enterprise@example.com", "user", true, "hash", 0.8, time.Now())

	// Force tier to enterprise in notification_preferences
	_, _ = db.Exec("INSERT INTO notification_preferences (user_id, plan_tier, email_enabled, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (user_id) DO UPDATE SET plan_tier = $2",
		uid, tier.Enterprise, true, time.Now(), time.Now())

	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo)
	solarRepo := solar.NewRepository(db)
	solarSvc := solar.NewService(solarRepo)
	forecastRepo := forecast.NewRepository(db)
	weatherRepo := weather.NewRepository(db)
	weatherSvc := weather.NewService(weatherRepo, "https://api.open-meteo.com/v1")
	notifRepo := notification.NewRepository(db)
	notifSvc := notification.NewService(notifRepo, "localhost", "587", "user", "pass", "from@example.com", "", "", "", "", "")
	recRepo := rec.NewRepository(db)
	recSvc := rec.NewService(recRepo, userSvc, notifSvc)
	wbRepo := weatherbaseline.NewRepository(db)
	wbSvc := weatherbaseline.NewService(wbRepo, "https://archive-api.open-meteo.com/v1")
	deviceRepo := device.NewRepository(db)
	deviceSvc := device.NewService(deviceRepo)
	forecastSvc := forecast.NewService(forecastRepo, solarSvc, deviceSvc, weatherSvc, recSvc, wbSvc)

	sites := []struct {
		Name string
		Lat  float64
		Lng  float64
		Cap  float64
	}{
		{"Jakarta HQ", -6.2000, 106.8166, 50.0},
		{"Surabaya Factory", -7.2575, 112.7521, 120.0},
		{"Bandung DC", -6.9175, 107.6191, 30.0},
		{"Medan Branch", 3.5952, 98.6722, 45.0},
		{"Bali Resort", -8.4095, 115.1889, 80.0},
	}

	fmt.Println("Seeding 5 Enterprise sites...")
	for _, s := range sites {
		p, err := solarSvc.CreateSolarProfile(ctx, solar.CreateSolarProfileRequest{
			UserID:      uid,
			SiteName:    s.Name,
			CapacityKwp: s.Cap,
			Lat:         s.Lat,
			Lng:         s.Lng,
			PlanTier:    tier.Enterprise,
		})
		if err != nil {
			fmt.Printf("Skipping %s (already exists or error: %v)\n", s.Name, err)
			// Try to find existing
			existing, _ := solarSvc.GetSolarProfilesByUserID(uid)
			for _, ex := range existing {
				if ex.SiteName == s.Name {
					p = ex
					break
				}
			}
		}

		if p == nil { continue }

		fmt.Printf("Seeding data for %s (%s)...\n", p.SiteName, p.ID)
		seedSiteData(db, forecastSvc, uid, p.ID, p.CapacityKwp)
	}

	fmt.Println("Seed complete! You can now view the ESG Dashboard for the Enterprise Test User.")
}

func seedSiteData(db *sql.DB, s forecast.Service, userID, profileID uuid.UUID, capacity float64) {
	now := time.Now().UTC().Truncate(24 * time.Hour)
	for i := 0; i < 90; i++ {
		date := now.AddDate(0, 0, -i)
		// Random daily production between 3.5 - 5.5 PSH equivalent
		psh := 3.5 + rand.Float64()*2.0
		actualKwh := capacity * psh * 0.8 // 80% efficiency
		
		// Record Actual
		_, err := s.RecordActualDaily(forecast.RecordActualRequest{
			UserID:         userID,
			SolarProfileID: profileID.String(),
			Date:           date.Format(time.DateOnly),
			ActualKwh:      actualKwh,
			Source:         "simulated",
		})
		if err != nil {
			// fmt.Printf("Error seeding actual for %s: %v\n", date.Format(time.DateOnly), err)
		}

		// Record dummy forecast too (simpler via SQL)
		_, _ = db.Exec("INSERT INTO forecasts (id, user_id, solar_profile_id, date, predicted_kwh, cloud_cover, shortwave_radiation, efficiency_used, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) ON CONFLICT (user_id, solar_profile_id, date) DO NOTHING",
			uuid.New(), userID, profileID, date, actualKwh * (0.9 + rand.Float64()*0.2), 30, 15, 0.8, time.Now())
	}
}
