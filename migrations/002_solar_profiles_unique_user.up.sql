DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'solar_profiles_user_id_key'
    ) AND NOT EXISTS (
        SELECT user_id
        FROM solar_profiles
        GROUP BY user_id
        HAVING COUNT(*) > 1
    ) THEN
        ALTER TABLE solar_profiles
        ADD CONSTRAINT solar_profiles_user_id_key UNIQUE (user_id);
    END IF;
END $$;
