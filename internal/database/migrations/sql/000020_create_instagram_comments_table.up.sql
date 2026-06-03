CREATE TABLE IF NOT EXISTS instagram_comments (
    id VARCHAR(255) PRIMARY KEY, -- This is the Instagram comment ID
    instagram_media_id VARCHAR(255) NOT NULL,
    account_id UUID NOT NULL,
    username VARCHAR(255) NOT NULL,
    text TEXT NOT NULL,
    is_hidden BOOLEAN DEFAULT false,
    is_flagged BOOLEAN DEFAULT false,
    moderation_reason VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT fk_account FOREIGN KEY (account_id) REFERENCES instagram_accounts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_instagram_comments_media_id ON instagram_comments(instagram_media_id);
CREATE INDEX IF NOT EXISTS idx_instagram_comments_account_id ON instagram_comments(account_id);
