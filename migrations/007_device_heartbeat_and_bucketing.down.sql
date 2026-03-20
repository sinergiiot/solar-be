DROP INDEX IF EXISTS idx_devices_user_last_seen;
DROP INDEX IF EXISTS idx_telemetry_raw_device_bucket;

ALTER TABLE telemetry_raw
DROP COLUMN IF EXISTS bucket_start;

ALTER TABLE devices
DROP COLUMN IF EXISTS last_telemetry_at,
DROP COLUMN IF EXISTS last_seen_at;
