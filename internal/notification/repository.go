package notification

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// Repository defines persistence operations for notification preferences.
type Repository interface {
	GetPreference(userID uuid.UUID) (*NotificationPreference, error)
	UpsertPreference(pref *NotificationPreference) error
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
			created_at,
			updated_at
		FROM notification_preferences
		WHERE user_id = $1
	`

	pref := &NotificationPreference{}
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
		&pref.CreatedAt,
		&pref.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get notification preference: %w", err)
	}

	return pref, nil
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
			preferred_send_time
		) VALUES (
			$1, $2, $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, ''), $9, $10, $11::time
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
	)
	if err != nil {
		return fmt.Errorf("upsert notification preference: %w", err)
	}

	return nil
}
