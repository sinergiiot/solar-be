-- 015_weather_baselines_and_delta_wf.up.sql
-- Add weather_baselines table and new columns for delta_wf and baseline_type

CREATE TABLE IF NOT EXISTS weather_baselines (
    id              BIGSERIAL PRIMARY KEY,
    profile_id      UUID NOT NULL,
    user_id         UUID NOT NULL,
    baseline_type   VARCHAR(20) NOT NULL, -- 'synthetic' | 'site' | 'blended'
    baseline_value  FLOAT NOT NULL,
    sample_count    INT NOT NULL,
    valid_from      DATE NOT NULL,
    valid_to        DATE NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    UNIQUE(profile_id, user_id, baseline_type)
);

ALTER TABLE weather_daily ADD COLUMN IF NOT EXISTS cloud_cover_mean FLOAT;

ALTER TABLE forecasts ADD COLUMN IF NOT EXISTS delta_wf FLOAT;
ALTER TABLE forecasts ADD COLUMN IF NOT EXISTS baseline_type VARCHAR(20);
