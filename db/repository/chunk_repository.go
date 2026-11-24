package repository

import (
	"context"
	"database/sql"

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

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO chunks (document_id, content, chunk_index, token_count, metadata) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return false
	}
	defer stmt.Close()

	for _, chunk := range chunks {
		_, err := stmt.ExecContext(ctx, chunk.DocumentID, chunk.Content, chunk.ChunkIndex, chunk.TokenCount, chunk.Metadata)
		if err != nil {
			return false
		}
	}

	return tx.Commit() == nil
}

func (r *ChunkRepository) FindByDocument(ctx context.Context, docID string) []*models.Chunk {
	rows, err := r.db.QueryContext(ctx, "SELECT id, document_id, content, chunk_index, token_count, metadata, created_at FROM chunks WHERE document_id = $1 ORDER BY chunk_index ASC", docID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var chunks []*models.Chunk
	for rows.Next() {
		chunk := &models.Chunk{}
		if err := rows.Scan(&chunk.ID, &chunk.DocumentID, &chunk.Content, &chunk.ChunkIndex, &chunk.TokenCount, &chunk.Metadata, &chunk.CreatedAt); err != nil {
			continue
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}

func (r *ChunkRepository) FindByID(ctx context.Context, id string) *models.Chunk {
	row := r.db.QueryRowContext(ctx, "SELECT id, document_id, content, chunk_index, token_count, metadata, created_at FROM chunks WHERE id = $1", id)

	chunk := &models.Chunk{}
	err := row.Scan(&chunk.ID, &chunk.DocumentID, &chunk.Content, &chunk.ChunkIndex, &chunk.TokenCount, &chunk.Metadata, &chunk.CreatedAt)
	if err != nil {
		return nil
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
