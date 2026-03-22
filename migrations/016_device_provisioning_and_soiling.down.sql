ALTER TABLE solar_profiles
DROP COLUMN IF EXISTS soiling_alert_active,
DROP COLUMN IF EXISTS soiling_alert_last_checked;

ALTER TABLE devices
DROP COLUMN IF EXISTS claim_pin;

-- Note: We do not restore NOT NULL to user_id safely here as it could fail 
-- if there are orphaned factory devices during a downgrade.
