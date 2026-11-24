-- Migration 000002: Create the chunks table with full-text search support.

CREATE TABLE IF NOT EXISTS chunks (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id   UUID         NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    content       TEXT         NOT NULL,
    chunk_index   INT          NOT NULL DEFAULT 0,
    token_count   INT          NOT NULL DEFAULT 0,
    search_vector TSVECTOR,
    metadata      JSONB        NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chunks_document ON chunks(document_id);
CREATE INDEX idx_chunks_search ON chunks USING GIN(search_vector);

-- Trigger to auto-update the tsvector column
CREATE OR REPLACE FUNCTION chunks_search_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector := to_tsvector('english', COALESCE(NEW.content, ''));
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_chunks_search
    BEFORE INSERT OR UPDATE ON chunks
    FOR EACH ROW EXECUTE FUNCTION chunks_search_update();
