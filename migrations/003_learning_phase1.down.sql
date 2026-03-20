DROP TABLE IF EXISTS actual_daily;

ALTER TABLE users
DROP COLUMN IF EXISTS forecast_efficiency;
