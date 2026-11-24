-- Rollback 000003: Drop the embeddings table.
-- The pgvector `vector` extension is kept so other tables can keep using it.

DROP TABLE IF EXISTS embeddings CASCADE;
