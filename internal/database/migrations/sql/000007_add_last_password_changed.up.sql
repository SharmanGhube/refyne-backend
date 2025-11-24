-- Add last_password_changed_at column to track password changes
ALTER TABLE users
ADD COLUMN last_password_changed_at TIMESTAMPTZ;

-- Add token_version column for invalidating tokens on password change
ALTER TABLE users
ADD COLUMN token_version INTEGER NOT NULL DEFAULT 1;

-- Set initial value to created_at for existing users
UPDATE users
SET last_password_changed_at = created_at
WHERE last_password_changed_at IS NULL;

-- Create indexes for efficient lookup
CREATE INDEX idx_users_last_password_changed ON users(last_password_changed_at);
CREATE INDEX idx_users_token_version ON users(id, token_version);

COMMENT ON COLUMN users.last_password_changed_at IS 'Timestamp of last password change - used to invalidate tokens issued before password change';
COMMENT ON COLUMN users.token_version IS 'Version counter incremented on password change - used to invalidate all existing tokens';
