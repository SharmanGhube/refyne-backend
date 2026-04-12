-- Create instagram_accounts table for storing connected Instagram accounts

CREATE TABLE IF NOT EXISTS instagram_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    instagram_user_id VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(255) NOT NULL,

    -- Token management
    access_token TEXT NOT NULL, -- Encrypted
    refresh_token TEXT, -- Encrypted, long-lived token
    token_expires_at TIMESTAMP WITH TIME ZONE NOT NULL,

    -- Account metadata
    profile_picture_url TEXT,
    biography TEXT,
    followers_count INT DEFAULT 0,

    -- Account status
    connected_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_sync_at TIMESTAMP WITH TIME ZONE,
    sync_status VARCHAR(50) DEFAULT 'idle', -- idle, syncing, or error
    sync_error_message TEXT,

    -- Timestamps for creation and updates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Foreign key constraint
    CONSTRAINT fk_instagram_accounts_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_instagram_accounts_user_id ON instagram_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_instagram_accounts_instagram_user_id ON instagram_accounts(instagram_user_id);
CREATE INDEX IF NOT EXISTS idx_instagram_accounts_deleted_at ON instagram_accounts(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_instagram_accounts_created_at ON instagram_accounts(created_at);
CREATE INDEX IF NOT EXISTS idx_instagram_accounts_sync_status ON instagram_accounts(sync_status);
CREATE INDEX IF NOT EXISTS idx_instagram_accounts_token_expires ON instagram_accounts(token_expires_at);

-- Create trigger for updating updated_at timestamp
CREATE TRIGGER update_instagram_accounts_updated_at
    BEFORE UPDATE ON instagram_accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE instagram_accounts IS 'Connected Instagram business accounts for users';
COMMENT ON COLUMN instagram_accounts.user_id IS 'Owner of the Instagram account (FK to users.id)';
COMMENT ON COLUMN instagram_accounts.instagram_user_id IS 'Unique Instagram user ID from the platform';
COMMENT ON COLUMN instagram_accounts.access_token IS 'OAuth access token for API calls (encrypted)';
COMMENT ON COLUMN instagram_accounts.refresh_token IS 'Long-lived refresh token for token renewal (encrypted)';
COMMENT ON COLUMN instagram_accounts.token_expires_at IS 'Expiration time of the current access token';
COMMENT ON COLUMN instagram_accounts.sync_status IS 'Current sync state: idle, syncing, or error';
