-- Prepare devices table for factory pre-provisioning (claiming system)
-- Devices are produced at factory without an assigned user yet.
ALTER TABLE devices
ALTER COLUMN user_id DROP NOT NULL;

-- Adding claim_pin for secure QR code claiming
ALTER TABLE devices
ADD COLUMN IF NOT EXISTS claim_pin VARCHAR(32);

-- Add soiling alert status to solar profiles to persist algorithms warnings
ALTER TABLE solar_profiles
ADD COLUMN IF NOT EXISTS soiling_alert_active BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS soiling_alert_last_checked TIMESTAMPTZ;
