ALTER TABLE notification_preferences
ADD COLUMN IF NOT EXISTS last_daily_forecast_sent_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS last_daily_forecast_sent_for_date DATE;
