DROP INDEX IF EXISTS idx_devices_solar_profile_id;

ALTER TABLE devices
DROP COLUMN IF EXISTS solar_profile_id;
