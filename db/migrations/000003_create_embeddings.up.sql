-- Migration 000003: Create the vector embeddings table using pgvector.
-- Requires the pgvector extension to be installed.

CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS embeddings (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chunk_id      UUID         NOT NULL REFERENCES chunks(id) ON DELETE CASCADE,
    document_id   UUID         NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    model_name    VARCHAR(128) NOT NULL DEFAULT 'local',
    dimension     INT          NOT NULL DEFAULT 768,
    embedding     VECTOR(768)  NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Unique constraint enables ON CONFLICT (chunk_id) in UpsertVector
CREATE UNIQUE INDEX idx_embeddings_chunk ON embeddings(chunk_id);
CREATE INDEX idx_embeddings_document ON embeddings(document_id);

-- IVFFlat index for approximate nearest-neighbour search
CREATE INDEX idx_embeddings_vector ON embeddings
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);
