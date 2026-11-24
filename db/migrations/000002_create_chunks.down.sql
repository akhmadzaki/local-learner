-- Rollback 000002: Drop the chunks trigger, function, and table.

DROP TRIGGER IF EXISTS trg_chunks_search ON chunks;
DROP FUNCTION IF EXISTS chunks_search_update();
DROP TABLE IF EXISTS chunks CASCADE;
