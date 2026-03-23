CREATE TABLE IF NOT EXISTS mwh_accumulators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    solar_profile_id UUID REFERENCES solar_profiles(id) ON DELETE CASCADE,
    cumulative_kwh DOUBLE PRECISION DEFAULT 0,
    last_updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    milestone_reached BOOLEAN DEFAULT FALSE,
    UNIQUE(user_id, solar_profile_id)
);

CREATE INDEX IF NOT EXISTS idx_mwh_accumulators_user_id ON mwh_accumulators(user_id);
