-- Create otto_settings table

CREATE TABLE IF NOT EXISTS otto_settings (
    user_id UUID PRIMARY KEY,
    auto_reply_enabled BOOLEAN NOT NULL DEFAULT false,
    tone VARCHAR(50) NOT NULL DEFAULT 'professional',
    guardrails TEXT,
    custom_instructions TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_otto_settings_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create trigger for updating updated_at timestamp
CREATE TRIGGER update_otto_settings_updated_at
    BEFORE UPDATE ON otto_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE otto_settings IS 'AI configuration settings for Otto';
COMMENT ON COLUMN otto_settings.user_id IS 'User who owns the settings (FK to users.id)';
COMMENT ON COLUMN otto_settings.auto_reply_enabled IS 'Whether Otto is allowed to automatically reply to safe comments';
COMMENT ON COLUMN otto_settings.tone IS 'The personality tone Otto should use when generating replies';
COMMENT ON COLUMN otto_settings.guardrails IS 'Rules for things Otto should never say';
COMMENT ON COLUMN otto_settings.custom_instructions IS 'Specific brand guidelines or FAQs for Otto to reference';
