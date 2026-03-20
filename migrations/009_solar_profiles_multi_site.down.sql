DROP INDEX IF EXISTS idx_solar_profiles_user_created_at;

DELETE FROM solar_profiles sp
USING solar_profiles newer
WHERE sp.user_id = newer.user_id
  AND sp.created_at < newer.created_at;

ALTER TABLE solar_profiles
ADD CONSTRAINT solar_profiles_user_id_key UNIQUE (user_id);

ALTER TABLE solar_profiles
DROP COLUMN IF EXISTS site_name;
