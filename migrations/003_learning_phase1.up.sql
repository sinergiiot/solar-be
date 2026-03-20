ALTER TABLE users
ADD COLUMN IF NOT EXISTS forecast_efficiency FLOAT NOT NULL DEFAULT 0.8;

CREATE TABLE IF NOT EXISTS actual_daily (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    actual_kwh FLOAT NOT NULL,
    source VARCHAR(50) NOT NULL DEFAULT 'manual',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, date)
);
