-- Create otto_messages table for AI assistant conversation messages

CREATE TABLE IF NOT EXISTS otto_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL,
    user_id UUID NOT NULL,

    -- Message content
    role VARCHAR(50) NOT NULL, -- "user" or "assistant"
    content TEXT NOT NULL,

    -- AI metadata
    tokens_used INT DEFAULT 0,
    model_used VARCHAR(255), -- Which AI model generated the response

    -- Message metadata (JSON for flexible fields)
    metadata JSONB,

    -- User feedback
    is_liked BOOLEAN,
    feedback_notes TEXT,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Foreign key constraints
    CONSTRAINT fk_otto_messages_conversation_id FOREIGN KEY (conversation_id) REFERENCES otto_conversations(id) ON DELETE CASCADE,
    CONSTRAINT fk_otto_messages_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_otto_messages_conversation_id ON otto_messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_otto_messages_user_id ON otto_messages(user_id);
CREATE INDEX IF NOT EXISTS idx_otto_messages_created_at ON otto_messages(created_at);
CREATE INDEX IF NOT EXISTS idx_otto_messages_conversation_created ON otto_messages(conversation_id, created_at);
CREATE INDEX IF NOT EXISTS idx_otto_messages_role ON otto_messages(role);
CREATE INDEX IF NOT EXISTS idx_otto_messages_is_liked ON otto_messages(is_liked) WHERE is_liked IS NOT NULL;

-- Add comments for documentation
COMMENT ON TABLE otto_messages IS 'Messages within AI assistant conversations';
COMMENT ON COLUMN otto_messages.conversation_id IS 'Parent conversation (FK to otto_conversations.id)';
COMMENT ON COLUMN otto_messages.user_id IS 'User who sent/received the message (FK to users.id)';
COMMENT ON COLUMN otto_messages.role IS 'Message sender: "user" for user messages or "assistant" for AI responses';
COMMENT ON COLUMN otto_messages.content IS 'The message text content';
COMMENT ON COLUMN otto_messages.tokens_used IS 'Number of AI tokens consumed by this message';
COMMENT ON COLUMN otto_messages.model_used IS 'AI model that generated assistant responses';
COMMENT ON COLUMN otto_messages.metadata IS 'Additional JSON metadata for extensibility';
COMMENT ON COLUMN otto_messages.is_liked IS 'User feedback: true for liked, false for disliked, null for no feedback';
COMMENT ON COLUMN otto_messages.feedback_notes IS 'Additional user feedback or notes';
