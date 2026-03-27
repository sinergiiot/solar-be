-- Migration 028: Add plan_expires_at and last_sent fields to notification_preferences
ALTER TABLE notification_preferences
  ADD COLUMN IF NOT EXISTS plan_expires_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN IF NOT EXISTS last_daily_forecast_sent_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN IF NOT EXISTS last_daily_forecast_sent_for_date TIMESTAMP WITH TIME ZONE;
