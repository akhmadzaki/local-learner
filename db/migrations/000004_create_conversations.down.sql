-- Rollback 000004: Drop the conversation_turns and conversations tables.

DROP TABLE IF EXISTS conversation_turns CASCADE;
DROP TABLE IF EXISTS conversations CASCADE;
