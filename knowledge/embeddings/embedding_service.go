package embeddings

import (
	"context"
	"log"

	"github.com/akhmadzaki/go-backend/db/models"
	"github.com/akhmadzaki/go-backend/db/repository"
	"github.com/akhmadzaki/go-backend/knowledge/vectorstore"
)

// EmbeddingService orchestrates embedding generation and vector storage.
type EmbeddingService struct {
	Provider    EmbeddingProvider
	vectorStore *vectorstore.VectorStore
	embedRepo   *repository.EmbeddingRepository
	dimension   int
}

func NewEmbeddingService(provider EmbeddingProvider, vectorStore *vectorstore.VectorStore, embedRepo *repository.EmbeddingRepository) *EmbeddingService {
	return &EmbeddingService{
		Provider:    provider,
		vectorStore: vectorStore,
		embedRepo:   embedRepo,
		dimension:   provider.Dimension(),
	}
}

// EmbeddingResult pairs a chunk with its generated vector and metadata.
type EmbeddingResult struct {
	ChunkID    string    `json:"chunk_id"`
	DocumentID string    `json:"document_id"`
	Vector     []float32 `json:"vector"`
	ModelName  string    `json:"model_name"`
	Dimension  int       `json:"dimension"`
}

// GenerateEmbeddings produces vectors for a list of chunks.
func (svc *EmbeddingService) GenerateEmbeddings(ctx context.Context, chunks []*models.Chunk) []EmbeddingResult {
	var results []EmbeddingResult

	for _, chunk := range chunks {
		vector, err := svc.Provider.CreateEmbedding(chunk.Content)
		if err != nil {
			log.Printf("EmbeddingService: Failed to embed chunk %s: %v", chunk.ID, err)
			continue
		}

		results = append(results, EmbeddingResult{
			ChunkID:    chunk.ID,
			DocumentID: chunk.DocumentID,
			Vector:     vector,
			ModelName:  svc.Provider.ModelName(),
			Dimension:  len(vector),
		})
	}

	return results
}

// StoreVectors writes embedding vectors into the vector store for similarity search.
// The single UpsertVector call handles all columns (chunk_id, document_id, embedding,
// model_name, dimension, metadata) so a separate embedRepo.Create is no longer needed.
// Returns false if any UpsertVector call fails.
func (svc *EmbeddingService) StoreVectors(ctx context.Context, results []EmbeddingResult) bool {
	for _, r := range results {
		if !svc.vectorStore.UpsertVector(ctx, r.ChunkID, r.DocumentID, r.ModelName, r.Dimension, r.Vector, nil) {
			log.Printf("EmbeddingService: Failed to store vector for chunk %s", r.ChunkID)
			return false
		}
	}
	return true
}

// DeleteVectorsByDocument removes all vectors belonging to a document.
func (svc *EmbeddingService) DeleteVectorsByDocument(ctx context.Context, docID string) {
	svc.vectorStore.DeleteByDocument(ctx, docID)
	svc.embedRepo.DeleteByDocument(ctx, docID)
}

// SearchVectors performs approximate nearest-neighbour search.
func (svc *EmbeddingService) SearchVectors(ctx context.Context, queryVector []float32, topK int) []vectorstore.SearchResult {
	return svc.vectorStore.Search(ctx, queryVector, topK)
}
