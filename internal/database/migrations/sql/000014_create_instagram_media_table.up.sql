-- Create instagram_media table for caching Instagram posts

CREATE TABLE IF NOT EXISTS instagram_media (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL,
    instagram_media_id VARCHAR(255) NOT NULL,

    -- Media metadata
    media_type VARCHAR(50) NOT NULL, -- PHOTO, VIDEO, CAROUSEL, REELS, STORY
    caption TEXT,
    media_url TEXT NOT NULL,
    permalink VARCHAR(500),
    thumbnail_url TEXT,

    -- Engagement metrics (cached)
    like_count INT DEFAULT 0,
    comment_count INT DEFAULT 0,
    shares_count INT DEFAULT 0,
    impressions INT DEFAULT 0,
    reach INT DEFAULT 0,

    -- Timestamps
    posted_at TIMESTAMP WITH TIME ZONE NOT NULL,
    synced_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- AI analysis result (JSON stored)
    ai_analysis JSONB,

    -- Timestamps for creation and updates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Foreign key constraint
    CONSTRAINT fk_instagram_media_account_id FOREIGN KEY (account_id) REFERENCES instagram_accounts(id) ON DELETE CASCADE,
    CONSTRAINT uk_instagram_media_account_id_media_id UNIQUE (account_id, instagram_media_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_instagram_media_account_id ON instagram_media(account_id);
CREATE INDEX IF NOT EXISTS idx_instagram_media_account_posted ON instagram_media(account_id, posted_at DESC);
CREATE INDEX IF NOT EXISTS idx_instagram_media_synced_at ON instagram_media(synced_at);
CREATE INDEX IF NOT EXISTS idx_instagram_media_media_type ON instagram_media(account_id, media_type);
CREATE INDEX IF NOT EXISTS idx_instagram_media_posted_at ON instagram_media(posted_at DESC);
CREATE INDEX IF NOT EXISTS idx_instagram_media_deleted_at ON instagram_media(deleted_at) WHERE deleted_at IS NULL;

-- Create trigger for updating updated_at timestamp
CREATE TRIGGER update_instagram_media_updated_at
    BEFORE UPDATE ON instagram_media
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE instagram_media IS 'Cached Instagram posts/media for analysis and tracking';
COMMENT ON COLUMN instagram_media.account_id IS 'Account the media belongs to (FK to instagram_accounts.id)';
COMMENT ON COLUMN instagram_media.instagram_media_id IS 'Unique Instagram media ID from the platform';
COMMENT ON COLUMN instagram_media.media_type IS 'Type of media: PHOTO, VIDEO, CAROUSEL, REELS, STORY';
COMMENT ON COLUMN instagram_media.ai_analysis IS 'JSON object with AI-generated analysis and recommendations';
COMMENT ON COLUMN instagram_media.synced_at IS 'Last time media data was synced from Instagram API';
