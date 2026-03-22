DROP INDEX IF EXISTS idx_email_verification_codes_code;
DROP INDEX IF EXISTS idx_email_verification_codes_user_id;
DROP TABLE IF EXISTS email_verification_codes;

ALTER TABLE users
DROP COLUMN IF EXISTS email_verified_at;

ALTER TABLE users
DROP COLUMN IF EXISTS email_verified;
