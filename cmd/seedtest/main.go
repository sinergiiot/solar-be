package main

import (
	"context"
	"database/sql"
	"flag"
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

const (
	USER_ID        = "00000000-0000-0000-0000-000000000001"
	PROFILE_NAME   = "Seed Test Site"
	CAPACITY_KWP   = 10.0
	LAT            = -8.0971
	LNG            = 112.1489
	START_DAYS_AGO = 60
)

func main() {
	reset := flag.Bool("reset", false, "clear existing seed data")
	verbose := flag.Bool("verbose", false, "reseed with fixes applied")
	dryRun := flag.Bool("dry-run", false, "run simulation without saving to DB")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer dbClose(db)

	ctx := context.Background()
	uid := uuid.MustParse(USER_ID)

	if *reset {
		fmt.Println("Resetting seed data...")
		_, _ = db.Exec("DELETE FROM actual_daily WHERE user_id = $1", uid)
		_, _ = db.Exec("DELETE FROM forecasts WHERE user_id = $1", uid)
		_, _ = db.Exec("DELETE FROM weather_baselines WHERE user_id = $1", uid)
		_, _ = db.Exec("DELETE FROM solar_profiles WHERE user_id = $1", uid)
		_, _ = db.Exec("UPDATE users SET forecast_efficiency = 0.8 WHERE id = $1", uid)
		if !*verbose {
			fmt.Println("Reset complete.")
			return
		}
	}

	// 1. Setup User and Profile
	_, _ = db.Exec("INSERT INTO users (id, name, email, role, email_verified, password_hash, forecast_efficiency, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (id) DO UPDATE SET forecast_efficiency = $7", 
		uid, "Seed Test User", "seedtest@example.com", "user", true, "hash", 0.8, time.Now())

	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo)
	solarRepo := solar.NewRepository(db)
	solarSvc := solar.NewService(solarRepo)

	var profile *solar.SolarProfile
	existingProfiles, err := solarSvc.GetSolarProfilesByUserID(uid)
	if err == nil && len(existingProfiles) > 0 {
		profile = existingProfiles[0]
		if profile.CapacityKwp != CAPACITY_KWP {
			profile.CapacityKwp = CAPACITY_KWP
			_, _ = db.Exec("UPDATE solar_profiles SET capacity_kwp = $1 WHERE id = $2", CAPACITY_KWP, profile.ID)
		}
	} else {
		profile, err = solarSvc.CreateSolarProfile(ctx, solar.CreateSolarProfileRequest{
			UserID:      uid,
			SiteName:    PROFILE_NAME,
			CapacityKwp: CAPACITY_KWP,
			Lat:         LAT,
			Lng:         LNG,
			PlanTier:    tier.Enterprise,
		})
		if err != nil {
			log.Fatalf("failed to create profile: %v", err)
		}
	}
	fmt.Printf("Using profile: %s (%s) with capacity %.2f kWp\n", profile.SiteName, profile.ID, profile.CapacityKwp)

	// 2. Initialize Services
	weatherRepo := weather.NewRepository(db)
	weatherSvc := weather.NewService(weatherRepo, "https://api.open-meteo.com/v1")
	
	wbRepo := weatherbaseline.NewRepository(db)
	wbSvc := weatherbaseline.NewService(wbRepo, "https://archive-api.open-meteo.com/v1")
	
	forecastRepo := forecast.NewRepository(db)
	deviceRepo := device.NewRepository(db)
	deviceSvc := device.NewService(deviceRepo)
	
	notifRepo := notification.NewRepository(db)
	notifSvc := notification.NewService(notifRepo, "localhost", "587", "user", "pass", "from@example.com", "", "", "", "", "")
	
	recRepo := rec.NewRepository(db)
	recSvc := rec.NewService(recRepo, userSvc, notifSvc)

	forecastSvc := forecast.NewService(forecastRepo, solarSvc, deviceSvc, weatherSvc, recSvc, wbSvc)

	// 3. Simulation Loop
	fmt.Println("\nStarting 60-day simulation...")
	fmt.Printf("%-12s | %-10s | %-5s | %-8s | %-8s | %-8s | %-8s\n", "Date", "Phase", "CC", "Pred", "Actual", "η (Old)", "η (New)")
	fmt.Println(string(make([]byte, 85)))

	now := time.Now().UTC().Truncate(24 * time.Hour)
	startDate := now.AddDate(0, 0, -START_DAYS_AGO-1)

	for i := 0; i <= START_DAYS_AGO; i++ {
		simDate := startDate.AddDate(0, 0, i)
		
		// 3.1. Compute Baseline - dihitung SETELAH semua actual di-seed untuk site baseline, 
		// tapi engine butuh baseline saat eksekusi. Kita panggil GetSiteBaseline di sini 
		// hanya untuk memastikan cache ter-refresh jika ada data baru.
		if !*dryRun {
			_, _, _ = wbSvc.GetSiteBaseline(ctx, profile.ID.String(), uid.String())
		}

		// 3.2. Generate Forecast
		oldEta, _ := forecastRepo.GetUserEfficiency(uid)
		
		f, err := forecastSvc.GenerateForecastForUser(uid, simDate)
		if err != nil {
			log.Fatalf("failed to generate forecast for %s: %v", simDate.Format("2006-01-02"), err)
		}

		// 3.3. Simulate Actual Energy
		w, _ := weatherSvc.FetchWeatherForDate(LAT, LNG, simDate)
		
		// Solusi alternatif: cap cloud_cover_mean di 90 untuk seed data (test data)
		cc := float64(w.CloudCover)
		if cc >= 95 {
			cc = 90
		}
		
		noise := 0.90 + rand.Float64()*0.20
		// Recalculate based on capped CC for simulation stability
		psh := w.ShortwaveRadiationMJ / 3.6
		simulatedActual := profile.CapacityKwp * psh * (1 - cc/100) * 0.75 * noise // Using target η ~0.75

		if simulatedActual <= 0 {
			simulatedActual = 0.01
		}

		if !*dryRun {
			_, err = forecastSvc.RecordActualDaily(forecast.RecordActualRequest{
				UserID:         uid,
				SolarProfileID: profile.ID.String(),
				Date:           simDate.Format("2006-01-02"),
				ActualKwh:      simulatedActual,
				Source:         "simulated",
			})
			if err != nil {
				log.Fatalf("failed to record actual for %s: %v", simDate.Format("2006-01-02"), err)
			}

			// 3.4. Calibrate
			res, err := forecastSvc.CalibrateEfficiencyForUser(uid, simDate)
			if err != nil {
				log.Printf("warning: failed to calibrate for %s: %v", simDate.Format("2006-01-02"), err)
			}
			newEta := oldEta
			if res != nil && res.Updated {
				newEta = res.NewEfficiency
			}

			phase := "Cold"
			if i >= 14 {
				phase = "Calibrated"
			} else if i > 0 {
				phase = "Transition"
			}

			if *verbose || i%5 == 0 || i > START_DAYS_AGO-5 {
				fmt.Printf("%-12s | %-10s | %-5.0f | %-8.2f | %-8.2f | %-8.4f | %-8.4f\n", 
					simDate.Format("2006-01-02"), phase, cc, f.PredictedKwh, simulatedActual, oldEta, newEta)
			}
		} else {
			fmt.Printf("%-12s | %-10s | %-5.0f | %-8.2f | %-8.2f | %-8.4f | (DRY)\n", 
				simDate.Format("2006-01-02"), "DryRun", cc, f.PredictedKwh, simulatedActual, oldEta)
		}
	}
	
	// Final Site Baseline Refresh (After all data seeded)
	if !*dryRun {
		fmt.Println("\nCalculating final Site Baseline after full seeding...")
		avg, count, _ := wbSvc.GetSiteBaseline(ctx, profile.ID.String(), uid.String())
		fmt.Printf("Final Site Baseline: %.2f%% (based on %d samples)\n", avg, count)
	}

	finalEta, _ := forecastRepo.GetUserEfficiency(uid)
	fmt.Printf("\nSimulation complete.\nFinal Efficiency (η): %.4f\n", finalEta)
}

func dbClose(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("error closing db: %v", err)
	}
}
