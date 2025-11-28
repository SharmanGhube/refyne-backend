-- Rollback subscription fields from users table

-- Drop indexes
DROP INDEX IF EXISTS idx_users_subscription_expires_at;
DROP INDEX IF EXISTS idx_users_paddle_subscription_id;
DROP INDEX IF EXISTS idx_users_paddle_customer_id;
DROP INDEX IF EXISTS idx_users_subscription_status;
DROP INDEX IF EXISTS idx_users_subscription_tier;

-- Drop columns
ALTER TABLE users DROP COLUMN IF EXISTS onboarding_completed;
ALTER TABLE users DROP COLUMN IF EXISTS paddle_subscription_id;
ALTER TABLE users DROP COLUMN IF EXISTS paddle_customer_id;
ALTER TABLE users DROP COLUMN IF EXISTS subscription_expires_at;
ALTER TABLE users DROP COLUMN IF EXISTS subscription_status;
ALTER TABLE users DROP COLUMN IF EXISTS subscription_tier;
