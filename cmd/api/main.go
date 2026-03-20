package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/auth"
	"github.com/akbarsenawijaya/solar-forecast/internal/config"
	"github.com/akbarsenawijaya/solar-forecast/internal/device"
	"github.com/akbarsenawijaya/solar-forecast/internal/forecast"
	"github.com/akbarsenawijaya/solar-forecast/internal/notification"
	"github.com/akbarsenawijaya/solar-forecast/internal/scheduler"
	"github.com/akbarsenawijaya/solar-forecast/internal/solar"
	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/akbarsenawijaya/solar-forecast/internal/weather"
	"github.com/akbarsenawijaya/solar-forecast/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// main boots the API server, scheduler, and supporting infrastructure.
func main() {
	// Load configuration from environment variables
	cfg := config.Load()

	// Connect to the database
	db := config.OpenDB(cfg.DBUrl)
	defer db.Close()

	// Run database migrations
	migrationsDir := resolveMigrationsDir()
	if err := utils.RunMigrations(db, migrationsDir); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Wire repositories
	userRepo := user.NewRepository(db)
	solarRepo := solar.NewRepository(db)
	weatherRepo := weather.NewRepository(db)
	forecastRepo := forecast.NewRepository(db)
	deviceRepo := device.NewRepository(db)
	notifRepo := notification.NewRepository(db)

	// Wire services
	userSvc := user.NewService(userRepo)
	solarSvc := solar.NewService(solarRepo)
	weatherSvc := weather.NewService(weatherRepo, cfg.Weather.BaseURL)
	forecastSvc := forecast.NewService(forecastRepo, solarSvc, weatherSvc)
	deviceSvc := device.NewService(deviceRepo)
	authSvc := auth.NewService(db, userSvc, cfg.Auth.JWTSecret, cfg.Auth.TokenExpiryHrs, cfg.Auth.RefreshTokenExpiryDays)
	notifSvc := notification.NewService(
		notifRepo,
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
		cfg.SMTP.From,
		cfg.Telegram.BotToken,
		cfg.WhatsApp.Token,
		cfg.WhatsApp.PhoneNumberID,
		cfg.WhatsApp.TemplateName,
		cfg.WhatsApp.LanguageCode,
	)

	// Start the daily forecast scheduler
	sched := scheduler.New(userSvc, solarSvc, forecastSvc, notifSvc)
	sched.Start()
	defer sched.Stop()

	// Wire HTTP handlers and router
	r := chi.NewRouter()
	r.Use(corsMiddleware(cfg.Frontend.AllowedOrigin))
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	authHandler := auth.NewHandler(authSvc, userSvc)
	deviceHandler := device.NewHandler(deviceSvc)
	notifHandler := notification.NewHandler(notifSvc)
	authHandler.RegisterPublicRoutes(r)
	deviceHandler.RegisterPublicRoutes(r)

	r.Group(func(protected chi.Router) {
		protected.Use(auth.Middleware(authSvc))
		authHandler.RegisterProtectedRoutes(protected)
		solar.NewHandler(solarSvc).RegisterRoutes(protected)
		forecast.NewHandler(forecastSvc, cfg.Debug.ForecastToken).RegisterRoutes(protected)
		deviceHandler.RegisterProtectedRoutes(protected)
		notifHandler.RegisterRoutes(protected)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	// Start HTTP server and wait for OS signal to shut down
	addr := ":" + cfg.Port
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		log.Printf("Server running on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}

// resolveMigrationsDir finds the migrations directory for both local runs and VPS service execution.
func resolveMigrationsDir() string {
	if _, err := os.Stat("migrations"); err == nil {
		return "migrations"
	}

	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to resolve executable path: %v", err)
	}

	baseDir := filepath.Dir(execPath)
	candidates := []string{
		filepath.Join(baseDir, "migrations"),
		filepath.Join(baseDir, "..", "migrations"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	log.Fatalf("Migrations directory not found")
	return ""
}

// corsMiddleware enables frontend access from the configured React origin.
func corsMiddleware(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
