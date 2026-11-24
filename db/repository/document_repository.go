package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/akhmadzaki/go-backend/db/models"
)

// DocumentRepository handles persistence for knowledge-base documents.
type DocumentRepository struct {
	db *sql.DB
}

func NewDocumentRepository(db *sql.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Create(ctx context.Context, title string, source string, status string) *models.Document {
	query := "INSERT INTO documents (title, source, status) VALUES ($1, $2, $3) RETURNING id, title, source, status, chunk_count, metadata, created_at, updated_at"
	row := r.db.QueryRowContext(ctx, query, title, source, status)

	doc := &models.Document{}
	var metaBytes []byte
	err := row.Scan(&doc.ID, &doc.Title, &doc.Source, &doc.Status, &doc.ChunkCount, &metaBytes, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		log.Printf("DocumentRepository.Create: scan failed: %v", err)
		return nil
	}
	if len(metaBytes) > 0 {
		json.Unmarshal(metaBytes, &doc.Metadata)
	}
	return doc
}

func (r *DocumentRepository) FindByID(ctx context.Context, id string) *models.Document {
	row := r.db.QueryRowContext(ctx, "SELECT id, title, source, status, chunk_count, metadata, created_at, updated_at FROM documents WHERE id = $1", id)

	doc := &models.Document{}
	var metaBytes []byte
	err := row.Scan(&doc.ID, &doc.Title, &doc.Source, &doc.Status, &doc.ChunkCount, &metaBytes, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		log.Printf("DocumentRepository.FindByID: scan failed: %v", err)
		return nil
	}
	if len(metaBytes) > 0 {
		json.Unmarshal(metaBytes, &doc.Metadata)
	}
	return doc
}

func (r *DocumentRepository) FindAll(ctx context.Context, offset int, limit int) []*models.Document {
	rows, err := r.db.QueryContext(ctx, "SELECT id, title, source, status, chunk_count, metadata, created_at, updated_at FROM documents ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		log.Printf("DocumentRepository.FindAll: query failed: %v", err)
		return nil
	}
	defer rows.Close()

	docs := make([]*models.Document, 0)
	for rows.Next() {
		doc := &models.Document{}
		var metaBytes []byte
		if err := rows.Scan(&doc.ID, &doc.Title, &doc.Source, &doc.Status, &doc.ChunkCount, &metaBytes, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			log.Printf("DocumentRepository.FindAll: scan failed: %v", err)
			continue
		}
		if len(metaBytes) > 0 {
			json.Unmarshal(metaBytes, &doc.Metadata)
		}
		docs = append(docs, doc)
	}
	return docs
}

func (r *DocumentRepository) UpdateStatus(ctx context.Context, id string, status string) bool {
	_, err := r.db.ExecContext(ctx, "UPDATE documents SET status = $1, updated_at = NOW() WHERE id = $2", status, id)
	return err == nil
}

func (r *DocumentRepository) UpdateChunkCount(ctx context.Context, id string, count int) bool {
	_, err := r.db.ExecContext(ctx, "UPDATE documents SET chunk_count = $1, updated_at = NOW() WHERE id = $2", count, id)
	return err == nil
}

func (r *DocumentRepository) Delete(ctx context.Context, id string) bool {
	_, err := r.db.ExecContext(ctx, "DELETE FROM documents WHERE id = $1", id)
	return err == nil
}

func (r *DocumentRepository) Count(ctx context.Context) int {
	var count int
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM documents").Scan(&count)
	return count
}
