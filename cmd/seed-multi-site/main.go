package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/solar"
	"github.com/akbarsenawijaya/solar-forecast/internal/tier"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const (
	USER_ID_MULTISITE = "00000000-0000-0000-0000-000000000005"
	START_DATE        = "2025-01-01"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://akbar:akbar@localhost:5432/solar_forecast?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	uid := uuid.MustParse(USER_ID_MULTISITE)

	// 1. Setup Enterprise User
	fmt.Println("Setting up Enterprise User...")
	_, err = db.Exec(`
		INSERT INTO users (id, name, email, role, email_verified, password_hash, company_name, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		ON CONFLICT (id) DO UPDATE SET company_name = $7`,
		uid, "Multi-Site Enterprise", "enterprise-multi@example.com", "user", true, "hash", "Sinergi Global Corp", time.Now())
	if err != nil {
		log.Fatalf("failed to setup user: %v", err)
	}

	// Set plan tier in notification_preferences (where the app currently expects it)
	_, err = db.Exec(`
		INSERT INTO notification_preferences (user_id, plan_tier, primary_channel, email_enabled, timezone, preferred_send_time)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE SET plan_tier = $2`,
		uid, tier.PlanEnterprise, "email", true, "UTC", "07:00:00")
	if err != nil {
		log.Fatalf("failed to set plan tier: %v", err)
	}

	// 2. Clear existing profiles for this test user
	_, _ = db.Exec("DELETE FROM actual_daily WHERE user_id = $1", uid)
	_, _ = db.Exec("DELETE FROM solar_profiles WHERE user_id = $1", uid)

	sites := []struct {
		name string
		cap  float64
		lat  float64
		lng  float64
	}{
		{"Jakarta Head Office", 50.0, -6.2088, 106.8456},
		{"Surabaya Factory", 120.0, -7.2575, 112.7521},
		{"Bandung Warehouse", 15.0, -6.9175, 107.6191},
		{"Medan Branch", 30.0, 3.5952, 98.6722},
		{"Makassar Data Center", 80.0, -5.1476, 119.4327},
	}

	solarRepo := solar.NewRepository(db)
	solarSvc := solar.NewService(solarRepo)

	fmt.Printf("Seeding %d Sites...\n", len(sites))
	for _, s := range sites {
		p, err := solarSvc.CreateSolarProfile(ctx, solar.CreateSolarProfileRequest{
			UserID:      uid,
			SiteName:    s.name,
			CapacityKwp: s.cap,
			Lat:         s.lat,
			Lng:         s.lng,
			PlanTier:    tier.PlanEnterprise,
		})
		if err != nil {
			log.Fatalf("failed to create site %s: %v", s.name, err)
		}

		// Seed 1 year of data
		fmt.Printf("Generating 365 days of data for %s...\n", s.name)
		startDate, _ := time.Parse("2006-01-02", START_DATE)
		
		for i := 0; i < 365; i++ {
			date := startDate.AddDate(0, 0, i)
			// Random production based on capacity (avg 4 kWh/kWp/day)
			actualKwh := s.cap * (2.5 + rand.Float64()*3.0) 
			
			_, err = db.Exec(`
				INSERT INTO actual_daily (id, user_id, solar_profile_id, date, actual_kwh, source, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				uuid.New(), uid, p.ID, date, actualKwh, "iot-sim", time.Now())
			if err != nil {
				log.Fatalf("failed to insert history: %v", err)
			}
		}
	}

	fmt.Println("Multi-site seeding complete!")
	fmt.Printf("Login as: enterprise-multi@example.com / Password: (not seeded, use existing or reset hash)\n")
	fmt.Println("Recommended: Use direct API check or ESG Dashboard in UI.")
}
