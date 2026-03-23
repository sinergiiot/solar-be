CREATE TABLE IF NOT EXISTS user_api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    api_key_hash TEXT NOT NULL UNIQUE,
    api_key_preview VARCHAR(16) NOT NULL, -- e.g. "sk_live_...xxxx"
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_api_key_hash ON user_api_keys(api_key_hash);
CREATE INDEX IF NOT EXISTS idx_user_api_key_user_id ON user_api_keys(user_id);
