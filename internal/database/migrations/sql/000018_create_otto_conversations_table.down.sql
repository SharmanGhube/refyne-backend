-- Drop otto_conversations table

DROP TRIGGER IF EXISTS update_otto_conversations_updated_at ON otto_conversations;
DROP TABLE IF EXISTS otto_conversations;
