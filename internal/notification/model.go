package notification

import (
	"time"

	"github.com/google/uuid"
)

const (
	// PlanFree routes notifications through free channels only.
	PlanFree = "free"
	// PlanPaid allows WhatsApp as premium channel.
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
	WeatherFactor      float64
	Efficiency         float64
	SolarProfileName   string
	EstimatedCost      float64
	EstimatedCO2Kg     float64
	DeviationPct       *float64
	ReferenceLabel     string
	WeatherRisk        string
}

// DispatchPayload holds one forecast notification payload independent of delivery channel.
type DispatchPayload struct {
	UserID           uuid.UUID
	ToName           string
	ToEmail          string
	Date             string
	PredictedKwh     float64
	WeatherFactor    float64
	Efficiency       float64
	SolarProfileName string
	WeatherRisk      string
	EstimatedCost    float64
	EstimatedCO2Kg   float64
	DeviationPct     *float64
	ReferenceLabel   string
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
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// UpsertPreferenceRequest contains editable fields for user notification preferences.
type UpsertPreferenceRequest struct {
	PlanTier          string `json:"plan_tier"`
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
