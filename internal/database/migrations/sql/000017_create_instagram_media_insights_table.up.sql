-- Create instagram_media_insights table for media-level analytics

CREATE TABLE IF NOT EXISTS instagram_media_insights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_id UUID NOT NULL,
    account_id UUID NOT NULL,

    -- Engagement metrics per media
    impressions INT DEFAULT 0,
    reach INT DEFAULT 0,
    profile_views INT DEFAULT 0,
    shares INT DEFAULT 0,
    saves INT DEFAULT 0,
    clicks INT DEFAULT 0,
    engagement_rate DECIMAL(10, 4) DEFAULT 0,

    -- Date metrics were collected for
    metric_date DATE NOT NULL,

    -- Timestamps
    collected_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Timestamps for creation and updates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Foreign key constraints
    CONSTRAINT fk_instagram_media_insights_media_id FOREIGN KEY (media_id) REFERENCES instagram_media(id) ON DELETE CASCADE,
    CONSTRAINT fk_instagram_media_insights_account_id FOREIGN KEY (account_id) REFERENCES instagram_accounts(id) ON DELETE CASCADE,
    CONSTRAINT uk_instagram_media_insights_media_date UNIQUE (media_id, metric_date)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_instagram_media_insights_media_id ON instagram_media_insights(media_id);
CREATE INDEX IF NOT EXISTS idx_instagram_media_insights_account_id ON instagram_media_insights(account_id);
CREATE INDEX IF NOT EXISTS idx_instagram_media_insights_account_date ON instagram_media_insights(account_id, metric_date DESC);
CREATE INDEX IF NOT EXISTS idx_instagram_media_insights_metric_date ON instagram_media_insights(metric_date DESC);
CREATE INDEX IF NOT EXISTS idx_instagram_media_insights_collected_at ON instagram_media_insights(collected_at);
CREATE INDEX IF NOT EXISTS idx_instagram_media_insights_deleted_at ON instagram_media_insights(deleted_at) WHERE deleted_at IS NULL;

-- Create trigger for updating updated_at timestamp
CREATE TRIGGER update_instagram_media_insights_updated_at
    BEFORE UPDATE ON instagram_media_insights
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE instagram_media_insights IS 'Per-media analytics data collected from Instagram';
COMMENT ON COLUMN instagram_media_insights.media_id IS 'Media the insights belong to (FK to instagram_media.id)';
COMMENT ON COLUMN instagram_media_insights.account_id IS 'Account the media belongs to (FK to instagram_accounts.id)';
COMMENT ON COLUMN instagram_media_insights.metric_date IS 'Date for which insights are recorded';
COMMENT ON COLUMN instagram_media_insights.engagement_rate IS 'Engagement rate as a percentage (0-100)';
