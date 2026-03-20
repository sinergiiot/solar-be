CREATE TABLE IF NOT EXISTS notification_preferences (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  plan_tier TEXT NOT NULL DEFAULT 'free' CHECK (plan_tier IN ('free', 'paid')),
  primary_channel TEXT NOT NULL DEFAULT 'email' CHECK (primary_channel IN ('email', 'telegram', 'whatsapp')),
  email_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  telegram_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  whatsapp_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  telegram_chat_id TEXT,
  whatsapp_phone_e164 TEXT,
  whatsapp_opted_in BOOLEAN NOT NULL DEFAULT FALSE,
  timezone TEXT NOT NULL DEFAULT 'UTC',
  preferred_send_time TIME NOT NULL DEFAULT '06:00:00',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notification_preferences_plan_tier
  ON notification_preferences (plan_tier);

CREATE INDEX IF NOT EXISTS idx_notification_preferences_primary_channel
  ON notification_preferences (primary_channel);
