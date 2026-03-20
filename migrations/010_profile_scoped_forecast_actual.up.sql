ALTER TABLE forecasts
ADD COLUMN IF NOT EXISTS solar_profile_id UUID REFERENCES solar_profiles(id) ON DELETE SET NULL;

WITH latest_profile AS (
    SELECT DISTINCT ON (user_id) id, user_id
    FROM solar_profiles
    ORDER BY user_id, created_at DESC
)
UPDATE forecasts f
SET solar_profile_id = lp.id
FROM latest_profile lp
WHERE f.user_id = lp.user_id
  AND f.solar_profile_id IS NULL;

ALTER TABLE forecasts
DROP CONSTRAINT IF EXISTS forecasts_user_id_date_key;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE table_name = 'forecasts'
          AND constraint_name = 'uq_forecasts_user_profile_date'
    ) THEN
        ALTER TABLE forecasts
        ADD CONSTRAINT uq_forecasts_user_profile_date UNIQUE (user_id, solar_profile_id, date);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_forecasts_user_profile_date
ON forecasts(user_id, solar_profile_id, date DESC);

ALTER TABLE actual_daily
ADD COLUMN IF NOT EXISTS solar_profile_id UUID REFERENCES solar_profiles(id) ON DELETE SET NULL;

WITH latest_profile AS (
    SELECT DISTINCT ON (user_id) id, user_id
    FROM solar_profiles
    ORDER BY user_id, created_at DESC
)
UPDATE actual_daily a
SET solar_profile_id = lp.id
FROM latest_profile lp
WHERE a.user_id = lp.user_id
  AND a.solar_profile_id IS NULL;

ALTER TABLE actual_daily
DROP CONSTRAINT IF EXISTS actual_daily_user_id_date_key;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE table_name = 'actual_daily'
          AND constraint_name = 'uq_actual_daily_user_profile_date'
    ) THEN
        ALTER TABLE actual_daily
        ADD CONSTRAINT uq_actual_daily_user_profile_date UNIQUE (user_id, solar_profile_id, date);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_actual_daily_user_profile_date
ON actual_daily(user_id, solar_profile_id, date DESC);
