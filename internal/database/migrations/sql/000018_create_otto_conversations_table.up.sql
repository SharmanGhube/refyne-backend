-- Create otto_conversations table for AI assistant conversations

CREATE TABLE IF NOT EXISTS otto_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    workspace_id UUID NOT NULL,

    -- Conversation metadata
    title VARCHAR(255) NOT NULL,
    description TEXT,
    context TEXT NOT NULL, -- JSON context (accounts, metrics, documents)

    -- Conversation state
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, archived, deleted
    is_bookmarked BOOLEAN DEFAULT FALSE,

    -- Message tracking
    message_count INT DEFAULT 0,
    last_message_at TIMESTAMP WITH TIME ZONE,

    -- Timestamps for creation and updates
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Foreign key constraints
    CONSTRAINT fk_otto_conversations_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_otto_conversations_workspace_id FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_otto_conversations_user_id ON otto_conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_otto_conversations_workspace_id ON otto_conversations(workspace_id);
CREATE INDEX IF NOT EXISTS idx_otto_conversations_status ON otto_conversations(status) WHERE status != 'deleted';
CREATE INDEX IF NOT EXISTS idx_otto_conversations_updated_at ON otto_conversations(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_otto_conversations_created_at ON otto_conversations(created_at);
CREATE INDEX IF NOT EXISTS idx_otto_conversations_is_bookmarked ON otto_conversations(is_bookmarked) WHERE is_bookmarked = TRUE;
CREATE INDEX IF NOT EXISTS idx_otto_conversations_user_workspace ON otto_conversations(user_id, workspace_id, status) WHERE status != 'deleted';

-- Create trigger for updating updated_at timestamp
CREATE TRIGGER update_otto_conversations_updated_at
    BEFORE UPDATE ON otto_conversations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE otto_conversations IS 'AI assistant conversation threads';
COMMENT ON COLUMN otto_conversations.user_id IS 'User who owns the conversation (FK to users.id)';
COMMENT ON COLUMN otto_conversations.workspace_id IS 'Workspace context for the conversation (FK to workspaces.id)';
COMMENT ON COLUMN otto_conversations.context IS 'JSON-encoded conversation context (accounts, metrics, documents)';
COMMENT ON COLUMN otto_conversations.status IS 'Conversation state: active, archived, or deleted';
COMMENT ON COLUMN otto_conversations.message_count IS 'Total number of messages in the conversation';
COMMENT ON COLUMN otto_conversations.last_message_at IS 'Timestamp of the most recent message';
