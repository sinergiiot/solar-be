package notification

import (
	"time"

	"github.com/google/uuid"
)

const (
	// PlanFree routes notifications through free channels only.
	PlanFree = "free"
	// PlanPro is the standard paid tier.
	PlanPro = "pro"
	// PlanEnterprise is the highest tier with all features.
	PlanEnterprise = "enterprise"
	// PlanPaid allows WhatsApp as premium channel (deprecated, use PlanPro).
	PlanPaid = "paid"

	// ChannelEmail routes to SMTP email delivery.
	ChannelEmail = "email"
	// ChannelTelegram routes to Telegram bot delivery.
	ChannelTelegram = "telegram"
	// ChannelWhatsApp routes to WhatsApp delivery.
	ChannelWhatsApp = "whatsapp"
)

// EmailPayload holds all data needed to compose a forecast email
type EmailPayload struct {
	ToName             string
	ToEmail            string
	Date               string
	PredictedKwh       float64
	CloudCover         int
	BaselineType       string
	WeatherFactor      float64
	Efficiency         float64
	SolarProfileName   string
	EstimatedCost      float64
	EstimatedCO2Kg     float64
	DeviationPct       *float64
	ReferenceLabel     string
	WeatherRisk        string
	Lat                float64
	Lng                float64
	ConditionLabel     string
	ConditionImpact    string
}

// DispatchPayload holds one forecast notification payload independent of delivery channel.
type DispatchPayload struct {
	UserID           uuid.UUID
	ToName           string
	ToEmail          string
	Date             string
	PredictedKwh     float64
	CloudCover       int
	BaselineType     string
	WeatherFactor    float64
	Efficiency       float64
	SolarProfileName string
	WeatherRisk      string
	EstimatedCost    float64
	EstimatedCO2Kg   float64
	DeviationPct     *float64
	ReferenceLabel   string
	Lat              float64
	Lng              float64
	ConditionLabel   string
	ConditionImpact  string
}

// NotificationPreference stores per-user channel settings.
type NotificationPreference struct {
	UserID            uuid.UUID `json:"user_id"`
	PlanTier          string    `json:"plan_tier"`
	PrimaryChannel    string    `json:"primary_channel"`
	EmailEnabled      bool      `json:"email_enabled"`
	TelegramEnabled   bool      `json:"telegram_enabled"`
	WhatsAppEnabled   bool      `json:"whatsapp_enabled"`
	TelegramChatID    string    `json:"telegram_chat_id,omitempty"`
	WhatsAppPhoneE164 string    `json:"whatsapp_phone_e164,omitempty"`
	WhatsAppOptedIn   bool      `json:"whatsapp_opted_in"`
	Timezone          string    `json:"timezone"`
	PreferredSendTime string    `json:"preferred_send_time"`
	PlanExpiresAt     *time.Time `json:"plan_expires_at,omitempty"`
	LastDailyForecastSentAt      *time.Time `json:"last_daily_forecast_sent_at,omitempty"`
	LastDailyForecastSentForDate *time.Time `json:"last_daily_forecast_sent_for_date,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type UpsertPreferenceRequest struct {
	PrimaryChannel    string `json:"primary_channel"`
	EmailEnabled      bool   `json:"email_enabled"`
	TelegramEnabled   bool   `json:"telegram_enabled"`
	WhatsAppEnabled   bool   `json:"whatsapp_enabled"`
	TelegramChatID    string `json:"telegram_chat_id"`
	WhatsAppPhoneE164 string `json:"whatsapp_phone_e164"`
	WhatsAppOptedIn   bool   `json:"whatsapp_opted_in"`
	Timezone          string `json:"timezone"`
	PreferredSendTime string `json:"preferred_send_time"`
}
