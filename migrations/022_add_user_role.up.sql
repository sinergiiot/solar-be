ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'user';

-- Set existing specific user to admin if present
UPDATE users SET role = 'admin' WHERE email = 'wijayasenaakbar@gmail.com';
