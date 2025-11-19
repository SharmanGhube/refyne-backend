-- Drop trigger
DROP TRIGGER IF EXISTS password_reset_tokens_updated_at ON password_reset_tokens;

-- Drop function
DROP FUNCTION IF EXISTS update_password_reset_tokens_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_password_reset_tokens_expires_at;
DROP INDEX IF EXISTS idx_password_reset_tokens_token;
DROP INDEX IF EXISTS idx_password_reset_tokens_user_id;

-- Drop table
DROP TABLE IF EXISTS password_reset_tokens;
