-- Drop trigger
DROP TRIGGER IF EXISTS trigger_update_verification_tokens_updated_at ON verification_tokens;

-- Drop function
DROP FUNCTION IF EXISTS update_verification_tokens_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_verification_tokens_is_valid;
DROP INDEX IF EXISTS idx_verification_tokens_expires_at;
DROP INDEX IF EXISTS idx_verification_tokens_token;
DROP INDEX IF EXISTS idx_verification_tokens_user_id;

-- Drop table
DROP TABLE IF EXISTS verification_tokens;
