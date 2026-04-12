-- Create instagram_ai_recommendations table for storing AI-generated insights

CREATE TABLE IF NOT EXISTS instagram_ai_recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    media_id UUID NOT NULL,

    -- AI-generated recommendations (stored as JSON for flexibility)
    sentiment_analysis JSONB, -- {overall, score, factors}
    content_analysis JSONB, -- {themes, content_type, quality_score, engagement_potential}
    hashtag_recommendations TEXT[], -- Array of recommended hashtags
    caption_suggestions JSONB, -- [{caption, reason}, ...]
    emoji_recommendations TEXT[], -- Array of recommended emojis
    cta_suggestions TEXT[], -- Array of call-to-action suggestions

    -- Posting strategy
    best_posting_times JSONB, -- {day: [...], time_utc: [...]}
    predicted_reach_multiplier DECIMAL(10, 2),
    growth_strategy TEXT,
    competitor_response TEXT,
    trend_alignment INT, -- 0-100

    -- Model metadata
    model_used VARCHAR(100), -- e.g., "gemini-2.0-flash"
    tokens_used INT,

    -- Validity
    generated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,

    -- Timestamps for creation and updates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Foreign key constraint
    CONSTRAINT fk_instagram_ai_recommendations_media_id FOREIGN KEY (media_id) REFERENCES instagram_media(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_instagram_ai_recommendations_media_id ON instagram_ai_recommendations(media_id);
CREATE INDEX IF NOT EXISTS idx_instagram_ai_recommendations_generated_at ON instagram_ai_recommendations(generated_at DESC);
CREATE INDEX IF NOT EXISTS idx_instagram_ai_recommendations_expires_at ON instagram_ai_recommendations(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_instagram_ai_recommendations_deleted_at ON instagram_ai_recommendations(deleted_at) WHERE deleted_at IS NULL;

-- Create trigger for updating updated_at timestamp
CREATE TRIGGER update_instagram_ai_recommendations_updated_at
    BEFORE UPDATE ON instagram_ai_recommendations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE instagram_ai_recommendations IS 'AI-generated recommendations for Instagram media using Gemini API';
COMMENT ON COLUMN instagram_ai_recommendations.media_id IS 'Media the recommendations are for (FK to instagram_media.id)';
COMMENT ON COLUMN instagram_ai_recommendations.sentiment_analysis IS 'JSON object with sentiment analysis results';
COMMENT ON COLUMN instagram_ai_recommendations.content_analysis IS 'JSON object with content analysis results';
COMMENT ON COLUMN instagram_ai_recommendations.best_posting_times IS 'Recommended days and times to post similar content';
COMMENT ON COLUMN instagram_ai_recommendations.tokens_used IS 'Gemini API tokens consumed for this analysis';
COMMENT ON COLUMN instagram_ai_recommendations.expires_at IS 'When recommendations expire and need to be regenerated';
