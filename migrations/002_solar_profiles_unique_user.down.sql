DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'solar_profiles_user_id_key'
    ) THEN
        ALTER TABLE solar_profiles
        DROP CONSTRAINT solar_profiles_user_id_key;
    END IF;
END $$;
