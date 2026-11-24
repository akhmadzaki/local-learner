package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/akhmadzaki/go-backend/db/models"
)

// ChunkRepository handles persistence for document chunks.
type ChunkRepository struct {
	db *sql.DB
}

func NewChunkRepository(db *sql.DB) *ChunkRepository {
	return &ChunkRepository{db: db}
}

func (r *ChunkRepository) BulkCreate(ctx context.Context, chunks []*models.Chunk) bool {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false
	}
	defer tx.Rollback()

	for _, chunk := range chunks {
		// Marshal metadata to JSON bytes for the JSONB column.
		// lib/pq does not handle map[string]interface{} natively.
		metaBytes, err := json.Marshal(chunk.Metadata)
		if err != nil {
			log.Printf("ChunkRepository.BulkCreate: metadata marshal failed: %v", err)
			return false
		}

		err = tx.QueryRowContext(ctx,
			"INSERT INTO chunks (document_id, content, chunk_index, token_count, metadata) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at",
			chunk.DocumentID, chunk.Content, chunk.ChunkIndex, chunk.TokenCount, metaBytes,
		).Scan(&chunk.ID, &chunk.CreatedAt)
		if err != nil {
			log.Printf("ChunkRepository.BulkCreate: insert failed: %v", err)
			return false
		}
	}

	return tx.Commit() == nil
}

func (r *ChunkRepository) FindByDocument(ctx context.Context, docID string) []*models.Chunk {
	rows, err := r.db.QueryContext(ctx, "SELECT id, document_id, content, chunk_index, token_count, metadata, created_at FROM chunks WHERE document_id = $1 ORDER BY chunk_index ASC", docID)
	if err != nil {
		log.Printf("ChunkRepository.FindByDocument: query failed: %v", err)
		return nil
	}
	defer rows.Close()

	var chunks []*models.Chunk
	for rows.Next() {
		chunk := &models.Chunk{}
		var metaBytes []byte
		if err := rows.Scan(&chunk.ID, &chunk.DocumentID, &chunk.Content, &chunk.ChunkIndex, &chunk.TokenCount, &metaBytes, &chunk.CreatedAt); err != nil {
			log.Printf("ChunkRepository.FindByDocument: scan failed: %v", err)
			continue
		}
		if len(metaBytes) > 0 {
			json.Unmarshal(metaBytes, &chunk.Metadata)
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}

func (r *ChunkRepository) FindByID(ctx context.Context, id string) *models.Chunk {
	row := r.db.QueryRowContext(ctx, "SELECT id, document_id, content, chunk_index, token_count, metadata, created_at FROM chunks WHERE id = $1", id)

	chunk := &models.Chunk{}
	var metaBytes []byte
	err := row.Scan(&chunk.ID, &chunk.DocumentID, &chunk.Content, &chunk.ChunkIndex, &chunk.TokenCount, &metaBytes, &chunk.CreatedAt)
	if err != nil {
		log.Printf("ChunkRepository.FindByID: scan failed: %v", err)
		return nil
	}
	if len(metaBytes) > 0 {
		json.Unmarshal(metaBytes, &chunk.Metadata)
	}
	return chunk
}

func (r *ChunkRepository) DeleteByDocument(ctx context.Context, docID string) bool {
	_, err := r.db.ExecContext(ctx, "DELETE FROM chunks WHERE document_id = $1", docID)
	return err == nil
}

func (r *ChunkRepository) Delete(ctx context.Context, id string) bool {
	_, err := r.db.ExecContext(ctx, "DELETE FROM chunks WHERE id = $1", id)
	return err == nil
}
