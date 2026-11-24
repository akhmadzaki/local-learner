-- Migration 000001: Create the documents table for the knowledge base.

CREATE TABLE IF NOT EXISTS documents (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR(512) NOT NULL,
    source      VARCHAR(64)  NOT NULL DEFAULT 'upload',
    status      VARCHAR(32)  NOT NULL DEFAULT 'pending',
    chunk_count INT          NOT NULL DEFAULT 0,
    metadata    JSONB        NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_documents_status ON documents(status);
CREATE INDEX idx_documents_source ON documents(source);
CREATE INDEX idx_documents_created ON documents(created_at DESC);
