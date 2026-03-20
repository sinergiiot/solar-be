package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Config holds all application configuration loaded from environment variables
type Config struct {
	DBUrl    string
	Port     string
	Auth     AuthConfig
	Debug    DebugConfig
	Frontend FrontendConfig
	SMTP     SMTPConfig
	Telegram TelegramConfig
	WhatsApp WhatsAppConfig
	Weather  WeatherConfig
}

// DebugConfig holds access control values for internal debug endpoints.
type DebugConfig struct {
	ForecastToken string
}

// AuthConfig holds authentication and token settings.
type AuthConfig struct {
	JWTSecret              string
	TokenExpiryHrs         int
	RefreshTokenExpiryDays int
}

// FrontendConfig holds browser client integration settings.
type FrontendConfig struct {
	AllowedOrigin string
}

// SMTPConfig holds email notification configuration
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// TelegramConfig holds Telegram bot integration settings.
type TelegramConfig struct {
	BotToken string
}

// WhatsAppConfig holds WhatsApp Cloud API settings.
type WhatsAppConfig struct {
	Token         string
	PhoneNumberID string
	TemplateName  string
	LanguageCode  string
}

// WeatherConfig holds external weather API configuration
type WeatherConfig struct {
	BaseURL string
}

// Load reads environment variables and returns a Config struct
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		DBUrl: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/solar_forecast?sslmode=disable"),
		Port:  getEnv("PORT", "8080"),
		Auth: AuthConfig{
			JWTSecret:              getEnv("AUTH_JWT_SECRET", "dev-secret-change-me"),
			TokenExpiryHrs:         getEnvAsInt("AUTH_TOKEN_EXPIRY_HOURS", 24),
			RefreshTokenExpiryDays: getEnvAsInt("AUTH_REFRESH_TOKEN_EXPIRY_DAYS", 7),
		},
		Debug: DebugConfig{
			ForecastToken: getEnv("FORECAST_DEBUG_TOKEN", ""),
		},
		Frontend: FrontendConfig{
			AllowedOrigin: getEnv("FRONTEND_ORIGIN", "http://localhost:5173"),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:     getEnv("SMTP_PORT", "587"),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@solar-forecast.com"),
		},
		Telegram: TelegramConfig{
			BotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		},
		WhatsApp: WhatsAppConfig{
			Token:         getEnv("WHATSAPP_ACCESS_TOKEN", ""),
			PhoneNumberID: getEnv("WHATSAPP_PHONE_NUMBER_ID", ""),
			TemplateName:  getEnv("WHATSAPP_TEMPLATE_NAME", "solar_forecast_daily"),
			LanguageCode:  getEnv("WHATSAPP_TEMPLATE_LANG", "id"),
		},
		Weather: WeatherConfig{
			BaseURL: getEnv("WEATHER_BASE_URL", "https://api.open-meteo.com/v1"),
		},
	}
}

// getEnvAsInt returns an int environment value or fallback.
func getEnvAsInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}

	return parsed
}

// OpenDB establishes a database connection and returns the *sql.DB instance
func OpenDB(dbURL string) *sql.DB {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Database connected successfully")
	return db
}

// getEnv returns the value of an environment variable or a fallback default
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
