-- Drop triggers and functions
DROP TRIGGER IF EXISTS trigger_update_account_lockouts_updated_at ON account_lockouts;
DROP FUNCTION IF EXISTS update_account_lockouts_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_account_lockouts_locked_until;
DROP INDEX IF EXISTS idx_account_lockouts_user_id;
DROP INDEX IF EXISTS idx_failed_attempts_attempted_at;
DROP INDEX IF EXISTS idx_failed_attempts_user_id;
DROP INDEX IF EXISTS idx_failed_attempts_email;

-- Drop tables
DROP TABLE IF EXISTS account_lockouts;
DROP TABLE IF EXISTS failed_login_attempts;
