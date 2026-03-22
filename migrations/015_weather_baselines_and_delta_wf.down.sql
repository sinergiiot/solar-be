-- 015_weather_baselines_and_delta_wf.down.sql
-- Rollback for weather_baselines and new columns

ALTER TABLE forecasts DROP COLUMN IF EXISTS delta_wf;
ALTER TABLE forecasts DROP COLUMN IF EXISTS baseline_type;
ALTER TABLE weather_daily DROP COLUMN IF EXISTS cloud_cover_mean;
DROP TABLE IF EXISTS weather_baselines;
