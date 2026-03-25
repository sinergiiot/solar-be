ALTER TABLE users
  DROP COLUMN IF EXISTS esg_share_token,
  DROP COLUMN IF EXISTS esg_share_enabled;
