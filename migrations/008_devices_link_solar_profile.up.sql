ALTER TABLE devices
ADD COLUMN IF NOT EXISTS solar_profile_id UUID REFERENCES solar_profiles(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_devices_solar_profile_id ON devices(solar_profile_id);

-- Backfill existing devices with the user's newest solar profile (currently usually one profile per user)
UPDATE devices d
SET solar_profile_id = sp.id
FROM (
    SELECT DISTINCT ON (user_id) id, user_id
    FROM solar_profiles
    ORDER BY user_id, created_at DESC
) sp
WHERE d.user_id = sp.user_id
  AND d.solar_profile_id IS NULL;
