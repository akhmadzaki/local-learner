package repository

import (
	"context"
	"database/sql"

	"github.com/akhmadzaki/go-backend/db/models"
)

// EmbeddingRepository handles persistence for embedding metadata.
type EmbeddingRepository struct {
	db *sql.DB
}

func NewEmbeddingRepository(db *sql.DB) *EmbeddingRepository {
	return &EmbeddingRepository{db: db}
}

func (r *EmbeddingRepository) Create(ctx context.Context, chunkID string, dimension int, modelName string) bool {
	query := "INSERT INTO embeddings (chunk_id, document_id, model_name, dimension) " +
		"VALUES ($1, (SELECT document_id FROM chunks WHERE id = $1), $2, $3)"
	_, err := r.db.ExecContext(ctx, query, chunkID, modelName, dimension)
	return err == nil
}

func (r *EmbeddingRepository) FindByChunk(ctx context.Context, chunkID string) *models.Embedding {
	row := r.db.QueryRowContext(ctx, "SELECT id, chunk_id, document_id, model_name, dimension, created_at FROM embeddings WHERE chunk_id = $1", chunkID)

	emb := &models.Embedding{}
	err := row.Scan(&emb.ID, &emb.ChunkID, &emb.DocumentID, &emb.ModelName, &emb.Dimension, &emb.CreatedAt)
	if err != nil {
		return nil
	}
	return emb
}

func (r *EmbeddingRepository) DeleteByDocument(ctx context.Context, docID string) bool {
	_, err := r.db.ExecContext(ctx, "DELETE FROM embeddings WHERE document_id = $1", docID)
	return err == nil
}

func (r *EmbeddingRepository) DeleteByChunk(ctx context.Context, chunkID string) bool {
	_, err := r.db.ExecContext(ctx, "DELETE FROM embeddings WHERE chunk_id = $1", chunkID)
	return err == nil
}
