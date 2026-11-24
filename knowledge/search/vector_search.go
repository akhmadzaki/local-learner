package search

import (
	"context"

	"github.com/akhmadzaki/go-backend/knowledge/vectorstore"
)

// VectorSearch provides pure embedding-based similarity search.
// This is a thin wrapper around the vector store's search method.
type VectorSearch struct {
	vectorStore *vectorstore.VectorStore
}

func NewVectorSearch(vectorStore *vectorstore.VectorStore) *VectorSearch {
	return &VectorSearch{vectorStore: vectorStore}
}

func (vs *VectorSearch) Search(ctx context.Context, queryVector []float32, topK int, filters map[string]interface{}) []SearchResult {
	results := vs.vectorStore.Search(ctx, queryVector, topK)

	var converted []SearchResult
	for _, r := range results {
		converted = append(converted, SearchResult{
			ChunkID:    r.ChunkID,
			DocumentID: r.DocumentID,
			Content:    r.Content,
			Score:      r.Score,
			Metadata:   r.Metadata,
		})
	}
	return converted
}
