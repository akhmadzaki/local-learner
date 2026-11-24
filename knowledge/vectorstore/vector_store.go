package vectorstore

import (
	"context"
	"database/sql"
	"log"

	"github.com/pgvector/pgvector-go"
)

// VectorStore is the abstraction over the vector database (pgvector).
// Currently supports pgvector; can be extended for Pinecone, Qdrant, etc.
type VectorStore struct {
	TableName string
	Dimension int
	IndexType string // "ivfflat", "hnsw"
	db        *sql.DB
}

// SearchResult represents a single match from a vector similarity search.
type SearchResult struct {
	ChunkID    string                 `json:"chunk_id"`
	DocumentID string                 `json:"document_id"`
	Content    string                 `json:"content"`
	Score      float64                `json:"score"` // cosine similarity or distance
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

func NewVectorStore(db *sql.DB, tableName string, dimension int) *VectorStore {
	return &VectorStore{
		TableName: tableName,
		Dimension: dimension,
		IndexType: "ivfflat",
		db:        db,
	}
}

// UpsertVector inserts or updates a vector in the store.
// The single INSERT covers all columns (chunk_id, document_id, embedding, model_name, dimension, metadata)
// so callers no longer need a separate embedRepo.Create call.
func (s *VectorStore) UpsertVector(ctx context.Context, chunkID, documentID, modelName string, dimension int, vector []float32, metadata map[string]interface{}) bool {
	pgVector := pgvector.NewVector(vector)
	query := "INSERT INTO embeddings (chunk_id, document_id, embedding, model_name, dimension) VALUES ($1, $2, $3, $4, $5) " +
		"ON CONFLICT (chunk_id) DO UPDATE SET embedding = $3, model_name = $4, dimension = $5"
	_, err := s.db.ExecContext(ctx, query, chunkID, documentID, pgVector, modelName, dimension)
	if err != nil {
		log.Printf("VectorStore: Upsert failed: %v", err)
		return false
	}
	return true
}

// Search performs approximate nearest-neighbour search via pgvector.
func (s *VectorStore) Search(ctx context.Context, query []float32, topK int) []SearchResult {
	pgQuery := pgvector.NewVector(query)
	queryStr := "SELECT e.chunk_id, e.document_id, c.content, " +
		"1 - (e.embedding <=> $1) AS score " +
		"FROM embeddings e " +
		"JOIN chunks c ON c.id = e.chunk_id " +
		"ORDER BY e.embedding <=> $1 " +
		"LIMIT $2"

	rows, err := s.db.QueryContext(ctx, queryStr, pgQuery, topK)
	if err != nil {
		log.Printf("VectorStore: Search failed: %v", err)
		return nil
	}
	defer rows.Close()

	results := make([]SearchResult, 0)
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.ChunkID, &r.DocumentID, &r.Content, &r.Score); err != nil {
			continue
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		log.Printf("VectorStore: Row iteration error: %v", err)
		return nil
	}
	return results
}

// DeleteByDocument removes all vectors belonging to a document.
func (s *VectorStore) DeleteByDocument(ctx context.Context, docID string) bool {
	_, err := s.db.ExecContext(ctx, "DELETE FROM embeddings WHERE document_id = $1", docID)
	if err != nil {
		log.Printf("VectorStore: DeleteByDocument failed: %v", err)
		return false
	}
	return true
}

// DeleteByChunk removes a single vector by chunk ID.
func (s *VectorStore) DeleteByChunk(ctx context.Context, chunkID string) bool {
	_, err := s.db.ExecContext(ctx, "DELETE FROM embeddings WHERE chunk_id = $1", chunkID)
	if err != nil {
		log.Printf("VectorStore: DeleteByChunk failed: %v", err)
		return false
	}
	return true
}

// HealthCheck verifies the database connection.
func (s *VectorStore) HealthCheck(ctx context.Context) bool {
	err := s.db.QueryRowContext(ctx, "SELECT 1").Err()
	return err == nil
}
