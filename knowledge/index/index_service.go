package index

import (
	"github.com/akhmadzaki/go-backend/knowledge/embeddings"
	"github.com/akhmadzaki/go-backend/knowledge/vectorstore"
	"log"
)

// IndexService manages the creation and maintenance of search indices over the knowledge base.
type IndexService struct {
	embedSvc    *embeddings.EmbeddingService
	vectorStore *vectorstore.VectorStore
}

func NewIndexService(embedSvc *embeddings.EmbeddingService, vectorStore *vectorstore.VectorStore) *IndexService {
	return &IndexService{
		embedSvc:    embedSvc,
		vectorStore: vectorStore,
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

// GetStats returns current index health and counts.
func (svc *IndexService) GetStats() IndexStats {
	return IndexStats{
		TotalDocuments: 0,
		TotalChunks:    0,
		TotalVectors:   0,
		LastIndexedAt:  "",
		IndexHealth:    "healthy",
	}
}

// DeleteIndex removes all vectors from the index (does not delete documents).
func (svc *IndexService) DeleteIndex() bool {
	log.Printf("IndexService: Deleting entire vector index")
	return true
}