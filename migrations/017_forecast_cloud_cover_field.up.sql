-- 017_forecast_cloud_cover_field.up.sql
-- Add cloud_cover column to forecasts table for easier reporting

ALTER TABLE forecasts ADD COLUMN IF NOT EXISTS cloud_cover INT DEFAULT 0;
