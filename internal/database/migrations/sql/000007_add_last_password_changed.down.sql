-- Drop indexes
DROP INDEX IF EXISTS idx_users_token_version;
DROP INDEX IF EXISTS idx_users_last_password_changed;

-- Remove columns
ALTER TABLE users
DROP COLUMN IF EXISTS token_version;

ALTER TABLE users
DROP COLUMN IF EXISTS last_password_changed_at;
