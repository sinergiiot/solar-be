package notification

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/akbarsenawijaya/solar-forecast/internal/tier"
	"github.com/google/uuid"
)

// Repository defines persistence operations for notification preferences.
type Repository interface {
	GetPreference(userID uuid.UUID) (*NotificationPreference, error)
	GetAllPreferences() ([]*NotificationPreference, error)
	UpsertPreference(pref *NotificationPreference) error
	MarkDailyForecastSent(userID uuid.UUID, forecastDate time.Time, sentAt time.Time) error
	GetPlanTier(userID uuid.UUID) (string, error)
	LogNotification(userID uuid.UUID, channel, status, errMsg string) error
}

type repository struct {
	db *sql.DB
}

// NewRepository creates a notification repository backed by postgres.
func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// GetPreference fetches one user's notification preference.
func (r *repository) GetPreference(userID uuid.UUID) (*NotificationPreference, error) {
	query := `
		SELECT
			user_id,
			plan_tier,
			primary_channel,
			email_enabled,
			telegram_enabled,
			whatsapp_enabled,
			COALESCE(telegram_chat_id, ''),
			COALESCE(whatsapp_phone_e164, ''),
			whatsapp_opted_in,
			timezone,
			to_char(preferred_send_time, 'HH24:MI:SS'),
			last_daily_forecast_sent_at,
			last_daily_forecast_sent_for_date,
			plan_expires_at,
			created_at,
			updated_at
		FROM notification_preferences
		WHERE user_id = $1
	`

	pref := &NotificationPreference{}
	var lastSentAt sql.NullTime
	var lastSentForDate sql.NullTime
	if err := r.db.QueryRow(query, userID).Scan(
		&pref.UserID,
		&pref.PlanTier,
		&pref.PrimaryChannel,
		&pref.EmailEnabled,
		&pref.TelegramEnabled,
		&pref.WhatsAppEnabled,
		&pref.TelegramChatID,
		&pref.WhatsAppPhoneE164,
		&pref.WhatsAppOptedIn,
		&pref.Timezone,
		&pref.PreferredSendTime,
		&lastSentAt,
		&lastSentForDate,
		&pref.PlanExpiresAt,
		&pref.CreatedAt,
		&pref.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get notification preference: %w", err)
	}

	if lastSentAt.Valid {
		pref.LastDailyForecastSentAt = &lastSentAt.Time
	}
	if lastSentForDate.Valid {
		pref.LastDailyForecastSentForDate = &lastSentForDate.Time
	}

	return pref, nil
}

// GetAllPreferences fetches all stored notification preferences in one query.
func (r *repository) GetAllPreferences() ([]*NotificationPreference, error) {
	query := `
		SELECT
			user_id,
			plan_tier,
			primary_channel,
			email_enabled,
			telegram_enabled,
			whatsapp_enabled,
			COALESCE(telegram_chat_id, ''),
			COALESCE(whatsapp_phone_e164, ''),
			whatsapp_opted_in,
			timezone,
			to_char(preferred_send_time, 'HH24:MI:SS'),
			last_daily_forecast_sent_at,
			last_daily_forecast_sent_for_date,
			plan_expires_at,
			created_at,
			updated_at
		FROM notification_preferences
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("get all notification preferences: %w", err)
	}
	defer rows.Close()

	prefs := []*NotificationPreference{}
	for rows.Next() {
		pref := &NotificationPreference{}
		var lastSentAt sql.NullTime
		var lastSentForDate sql.NullTime

		if err := rows.Scan(
			&pref.UserID,
			&pref.PlanTier,
			&pref.PrimaryChannel,
			&pref.EmailEnabled,
			&pref.TelegramEnabled,
			&pref.WhatsAppEnabled,
			&pref.TelegramChatID,
			&pref.WhatsAppPhoneE164,
			&pref.WhatsAppOptedIn,
			&pref.Timezone,
			&pref.PreferredSendTime,
			&lastSentAt,
			&lastSentForDate,
			&pref.PlanExpiresAt,
			&pref.CreatedAt,
			&pref.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification preference: %w", err)
		}

		if lastSentAt.Valid {
			pref.LastDailyForecastSentAt = &lastSentAt.Time
		}
		if lastSentForDate.Valid {
			pref.LastDailyForecastSentForDate = &lastSentForDate.Time
		}

		prefs = append(prefs, pref)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notification preferences: %w", err)
	}

	return prefs, nil
}

// UpsertPreference inserts or updates a user's notification preference.
func (r *repository) UpsertPreference(pref *NotificationPreference) error {
	query := `
		INSERT INTO notification_preferences (
			user_id,
			plan_tier,
			primary_channel,
			email_enabled,
			telegram_enabled,
			whatsapp_enabled,
			telegram_chat_id,
			whatsapp_phone_e164,
			whatsapp_opted_in,
			timezone,
			preferred_send_time,
			plan_expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, ''), $9, $10, $11::time, $12
		)
		ON CONFLICT (user_id) DO UPDATE SET
			plan_tier = EXCLUDED.plan_tier,
			primary_channel = EXCLUDED.primary_channel,
			email_enabled = EXCLUDED.email_enabled,
			telegram_enabled = EXCLUDED.telegram_enabled,
			whatsapp_enabled = EXCLUDED.whatsapp_enabled,
			telegram_chat_id = EXCLUDED.telegram_chat_id,
			whatsapp_phone_e164 = EXCLUDED.whatsapp_phone_e164,
			whatsapp_opted_in = EXCLUDED.whatsapp_opted_in,
			timezone = EXCLUDED.timezone,
			preferred_send_time = EXCLUDED.preferred_send_time,
			plan_expires_at = EXCLUDED.plan_expires_at,
			updated_at = NOW()
	`

	_, err := r.db.Exec(
		query,
		pref.UserID,
		strings.TrimSpace(pref.PlanTier),
		strings.TrimSpace(pref.PrimaryChannel),
		pref.EmailEnabled,
		pref.TelegramEnabled,
		pref.WhatsAppEnabled,
		strings.TrimSpace(pref.TelegramChatID),
		strings.TrimSpace(pref.WhatsAppPhoneE164),
		pref.WhatsAppOptedIn,
		strings.TrimSpace(pref.Timezone),
		strings.TrimSpace(pref.PreferredSendTime),
		pref.PlanExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("upsert notification preference: %w", err)
	}

	return nil
}

// MarkDailyForecastSent stores last successful daily forecast delivery markers.
func (r *repository) MarkDailyForecastSent(userID uuid.UUID, forecastDate time.Time, sentAt time.Time) error {
	query := `
		UPDATE notification_preferences
		SET last_daily_forecast_sent_at = $2,
			last_daily_forecast_sent_for_date = $3,
			updated_at = NOW()
		WHERE user_id = $1
	`

	_, err := r.db.Exec(query, userID, sentAt.UTC(), normalizeDateOnly(forecastDate))
	if err != nil {
		return fmt.Errorf("mark daily forecast sent: %w", err)
	}

	return nil
}

func normalizeDateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

// GetPlanTier returns only the plan_tier string for a user, defaulting to 'free' if no record exists.
func (r *repository) GetPlanTier(userID uuid.UUID) (string, error) {
	query := `SELECT plan_tier FROM notification_preferences WHERE user_id = $1`
	var planTier string
	err := r.db.QueryRow(query, userID).Scan(&planTier)
	if err != nil {
		if err == sql.ErrNoRows {
			return tier.Free, nil
		}
		return tier.Free, fmt.Errorf("get plan tier: %w", err)
	}
	return planTier, nil
}

func (r *repository) LogNotification(userID uuid.UUID, channel, status, errMsg string) error {
	query := `INSERT INTO notification_logs (user_id, channel, status, error_message) VALUES ($1, $2, $3, $4)`
	var errVal sql.NullString
	if errMsg != "" {
		errVal = sql.NullString{String: errMsg, Valid: true}
	}
	_, err := r.db.Exec(query, userID, channel, status, errVal)
	return err
}
