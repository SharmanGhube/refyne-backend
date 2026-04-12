-- Create instagram_insights table for analytics data

CREATE TABLE IF NOT EXISTS instagram_insights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL,

    -- Date for which insights are recorded
    metric_date DATE NOT NULL,

    -- Account-level metrics
    impressions INT DEFAULT 0,
    reach INT DEFAULT 0,
    profile_views INT DEFAULT 0,
    follower_count INT DEFAULT 0,
    follower_growth INT DEFAULT 0,
    engagement_rate DECIMAL(10, 4) DEFAULT 0,
    growth_rate DECIMAL(10, 4) DEFAULT 0,

    -- Demographics (JSON stored if available)
    demographics JSONB,

    -- Timestamps
    synced_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Timestamps for creation and updates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Foreign key constraint
    CONSTRAINT fk_instagram_insights_account_id FOREIGN KEY (account_id) REFERENCES instagram_accounts(id) ON DELETE CASCADE,
    CONSTRAINT uk_instagram_insights_account_date UNIQUE (account_id, metric_date)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_instagram_insights_account_id ON instagram_insights(account_id);
CREATE INDEX IF NOT EXISTS idx_instagram_insights_account_date ON instagram_insights(account_id, metric_date DESC);
CREATE INDEX IF NOT EXISTS idx_instagram_insights_metric_date ON instagram_insights(metric_date DESC);
CREATE INDEX IF NOT EXISTS idx_instagram_insights_synced_at ON instagram_insights(synced_at);
CREATE INDEX IF NOT EXISTS idx_instagram_insights_deleted_at ON instagram_insights(deleted_at) WHERE deleted_at IS NULL;

-- Create trigger for updating updated_at timestamp
CREATE TRIGGER update_instagram_insights_updated_at
    BEFORE UPDATE ON instagram_insights
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE instagram_insights IS 'Daily analytics data for Instagram accounts';
COMMENT ON COLUMN instagram_insights.account_id IS 'Account the insights belong to (FK to instagram_accounts.id)';
COMMENT ON COLUMN instagram_insights.metric_date IS 'Date for which insights are recorded';
COMMENT ON COLUMN instagram_insights.engagement_rate IS 'Engagement rate as a percentage (0-100)';
COMMENT ON COLUMN instagram_insights.demographics IS 'JSON object with audience demographics if available';
