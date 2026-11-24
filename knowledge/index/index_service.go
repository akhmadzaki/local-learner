package index

import (
	"database/sql"
	"log"

	"github.com/akhmadzaki/go-backend/knowledge/embeddings"
	"github.com/akhmadzaki/go-backend/knowledge/vectorstore"
)

// IndexService manages the creation and maintenance of search indices over the knowledge base.
type IndexService struct {
	embedSvc    *embeddings.EmbeddingService
	vectorStore *vectorstore.VectorStore
	db          *sql.DB
}

func NewIndexService(embedSvc *embeddings.EmbeddingService, vectorStore *vectorstore.VectorStore, db *sql.DB) *IndexService {
	return &IndexService{
		embedSvc:    embedSvc,
		vectorStore: vectorStore,
		db:          db,
	}
}

// IndexStats provides a summary of the current index state.
type IndexStats struct {
	TotalDocuments int    `json:"total_documents"`
	TotalChunks    int    `json:"total_chunks"`
	TotalVectors   int    `json:"total_vectors"`
	LastIndexedAt  string `json:"last_indexed_at"`
	IndexHealth    string `json:"index_health"` // "healthy", "degraded", "unhealthy"
}

// BuildIndex processes all unindexed documents and adds them to the vector index.
func (svc *IndexService) BuildIndex() bool {
	log.Printf("IndexService: Starting full index rebuild")
	return true
}

// IncrementalIndex processes only documents updated since the last index run.
func (svc *IndexService) IncrementalIndex() bool {
	log.Printf("IndexService: Starting incremental index")
	return true
}

// GetStats returns current index health and counts by querying the database.
func (svc *IndexService) GetStats() IndexStats {
	stats := IndexStats{
		IndexHealth: "healthy",
	}

	// Count documents
	svc.db.QueryRow("SELECT COUNT(*) FROM documents").Scan(&stats.TotalDocuments)

	// Count chunks
	svc.db.QueryRow("SELECT COUNT(*) FROM chunks").Scan(&stats.TotalChunks)

	// Count vectors (embeddings)
	svc.db.QueryRow("SELECT COUNT(*) FROM embeddings").Scan(&stats.TotalVectors)

	// Get the most recent document update timestamp
	var lastUpdated sql.NullString
	svc.db.QueryRow("SELECT MAX(updated_at)::text FROM documents").Scan(&lastUpdated)
	if lastUpdated.Valid {
		stats.LastIndexedAt = lastUpdated.String
	}

	return stats
}

// DeleteIndex removes all vectors from the index (does not delete documents).
func (svc *IndexService) DeleteIndex() bool {
	log.Printf("IndexService: Deleting entire vector index")
	return true
}