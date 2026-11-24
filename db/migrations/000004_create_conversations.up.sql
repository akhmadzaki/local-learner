-- Migration 000004: Create the conversations and conversation_turns tables
-- for multi-turn agent chat history.

CREATE TABLE IF NOT EXISTS conversations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR(256) NOT NULL DEFAULT 'New Conversation',
    turn_count  INT          NOT NULL DEFAULT 0,
    status      VARCHAR(32)  NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conversations_status ON conversations(status);

CREATE TABLE IF NOT EXISTS conversation_turns (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id   UUID         NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role              VARCHAR(16)  NOT NULL,
    content           TEXT         NOT NULL,
    sources           JSONB        NOT NULL DEFAULT '[]',
    token_usage       JSONB        NOT NULL DEFAULT '{}',
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_turns_conversation ON conversation_turns(conversation_id);
CREATE INDEX idx_turns_created ON conversation_turns(created_at ASC);
