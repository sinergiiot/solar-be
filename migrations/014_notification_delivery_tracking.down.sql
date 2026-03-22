ALTER TABLE notification_preferences
DROP COLUMN IF EXISTS last_daily_forecast_sent_for_date,
DROP COLUMN IF EXISTS last_daily_forecast_sent_at;
