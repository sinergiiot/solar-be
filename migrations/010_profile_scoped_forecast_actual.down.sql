DROP INDEX IF EXISTS idx_actual_daily_user_profile_date;
ALTER TABLE actual_daily
DROP CONSTRAINT IF EXISTS uq_actual_daily_user_profile_date;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE table_name = 'actual_daily'
          AND constraint_name = 'actual_daily_user_id_date_key'
    ) THEN
        ALTER TABLE actual_daily
        ADD CONSTRAINT actual_daily_user_id_date_key UNIQUE (user_id, date);
    END IF;
END $$;

ALTER TABLE actual_daily
DROP COLUMN IF EXISTS solar_profile_id;

DROP INDEX IF EXISTS idx_forecasts_user_profile_date;
ALTER TABLE forecasts
DROP CONSTRAINT IF EXISTS uq_forecasts_user_profile_date;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints
        WHERE table_name = 'forecasts'
          AND constraint_name = 'forecasts_user_id_date_key'
    ) THEN
        ALTER TABLE forecasts
        ADD CONSTRAINT forecasts_user_id_date_key UNIQUE (user_id, date);
    END IF;
END $$;

ALTER TABLE forecasts
DROP COLUMN IF EXISTS solar_profile_id;
