package chunks

import (
	"context"
	"log"

	"github.com/akhmadzaki/go-backend/db/models"
	"github.com/akhmadzaki/go-backend/db/repository"
)

// ChunkService splits document content into manageable pieces for embedding and retrieval.
type ChunkService struct {
	chunkRepo *repository.ChunkRepository
	strategy  ChunkStrategy
	config    ChunkConfig
}

func NewChunkService(chunkRepo *repository.ChunkRepository, strategy ChunkStrategy, config ChunkConfig) *ChunkService {
	return &ChunkService{
		chunkRepo: chunkRepo,
		strategy:  strategy,
		config:    config,
	}
}

// ChunkStrategy defines the interface for different chunking algorithms.
type ChunkStrategy interface {
	Split(text string, docID string, meta map[string]interface{}) []*models.Chunk
}

// ChunkConfig holds parameters that control chunking behaviour.
type ChunkConfig struct {
	Strategy       string `json:"strategy"`         // "fixed", "sentence", "paragraph"
	ChunkSize      int    `json:"chunk_size"`       // target token count per chunk
	ChunkOverlap   int    `json:"chunk_overlap"`    // overlap between consecutive chunks
	MinChunkLength int    `json:"min_chunk_length"` // discard chunks shorter than this
}

// SplitContent applies the configured strategy and persists chunks.
func (svc *ChunkService) SplitContent(ctx context.Context, content string, docID string) []*models.Chunk {
	chunks := svc.strategy.Split(content, docID, nil)

	// Filter out chunks that are too short
	var filtered []*models.Chunk
	for _, ch := range chunks {
		if len(ch.Content) >= svc.config.MinChunkLength {
			filtered = append(filtered, ch)
		}
	}

	// Ensure metadata is never nil (JSONB NOT NULL constraint)
	for _, ch := range filtered {
		if ch.Metadata == nil {
			ch.Metadata = make(map[string]interface{})
		}
	}

	// Persist chunks to database
	if !svc.chunkRepo.BulkCreate(ctx, filtered) {
		log.Printf("ChunkService: Failed to persist chunks for document %s", docID)
	}

	log.Printf("ChunkService: Split document %s into %d chunks", docID, len(filtered))
	return filtered
}

// DeleteChunksByDocument removes all chunks for a given document.
func (svc *ChunkService) DeleteChunksByDocument(ctx context.Context, docID string) bool {
	return svc.chunkRepo.DeleteByDocument(ctx, docID)
}

// GetChunksByDocument retrieves all chunks for a document.
func (svc *ChunkService) GetChunksByDocument(ctx context.Context, docID string) []*models.Chunk {
	return svc.chunkRepo.FindByDocument(ctx, docID)
}
