ALTER TABLE solar_profiles
ADD COLUMN IF NOT EXISTS site_name VARCHAR(255);

UPDATE solar_profiles
SET site_name = 'Main Site'
WHERE site_name IS NULL OR TRIM(site_name) = '';

ALTER TABLE solar_profiles
ALTER COLUMN site_name SET NOT NULL;

ALTER TABLE solar_profiles
DROP CONSTRAINT IF EXISTS solar_profiles_user_id_key;

CREATE INDEX IF NOT EXISTS idx_solar_profiles_user_created_at
ON solar_profiles(user_id, created_at DESC);
