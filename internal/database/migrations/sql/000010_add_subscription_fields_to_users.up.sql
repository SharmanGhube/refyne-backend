-- Add subscription fields to users table for Paddle integration

-- Add subscription tier
ALTER TABLE users ADD COLUMN IF NOT EXISTS subscription_tier VARCHAR(20) NOT NULL DEFAULT 'starter' 
    CHECK (subscription_tier IN ('starter', 'professional', 'business', 'enterprise'));

-- Add subscription status
ALTER TABLE users ADD COLUMN IF NOT EXISTS subscription_status VARCHAR(20) NOT NULL DEFAULT 'inactive'
    CHECK (subscription_status IN ('active', 'cancelled', 'past_due', 'trialing', 'paused', 'inactive'));

-- Add subscription expiration date
ALTER TABLE users ADD COLUMN IF NOT EXISTS subscription_expires_at TIMESTAMP WITH TIME ZONE;

-- Add Paddle customer identifiers
ALTER TABLE users ADD COLUMN IF NOT EXISTS paddle_customer_id VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS paddle_subscription_id VARCHAR(255);

-- Add onboarding tracking
ALTER TABLE users ADD COLUMN IF NOT EXISTS onboarding_completed BOOLEAN NOT NULL DEFAULT FALSE;

-- Create indexes for subscription queries
CREATE INDEX IF NOT EXISTS idx_users_subscription_tier ON users(subscription_tier);
CREATE INDEX IF NOT EXISTS idx_users_subscription_status ON users(subscription_status);
CREATE INDEX IF NOT EXISTS idx_users_paddle_customer_id ON users(paddle_customer_id);
CREATE INDEX IF NOT EXISTS idx_users_paddle_subscription_id ON users(paddle_subscription_id);
CREATE INDEX IF NOT EXISTS idx_users_subscription_expires_at ON users(subscription_expires_at) WHERE subscription_expires_at IS NOT NULL;

-- Add comment for documentation
COMMENT ON COLUMN users.subscription_tier IS 'User subscription tier: starter, professional, business, or enterprise';
COMMENT ON COLUMN users.subscription_status IS 'Current subscription status from Paddle';
COMMENT ON COLUMN users.paddle_customer_id IS 'Paddle customer ID for API calls';
COMMENT ON COLUMN users.paddle_subscription_id IS 'Paddle subscription ID for management';
