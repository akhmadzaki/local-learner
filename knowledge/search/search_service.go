package search

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/akhmadzaki/go-backend/knowledge/embeddings"
	"github.com/akhmadzaki/go-backend/knowledge/vectorstore"
)

// SearchService orchestrates hybrid search combining vector similarity
// and keyword (full-text) retrieval with configurable weighting.
type SearchService struct {
	embedSvc    *embeddings.EmbeddingService
	vectorStore *vectorstore.VectorStore
	hybridSvc   *HybridSearch
	keywordSvc  *KeywordSearch
}

func NewSearchService(embedSvc *embeddings.EmbeddingService, vectorStore *vectorstore.VectorStore, db *sql.DB) *SearchService {
	return &SearchService{
		embedSvc:    embedSvc,
		vectorStore: vectorStore,
		hybridSvc:   NewHybridSearch(vectorStore),
		keywordSvc:  NewKeywordSearch(db),
	}
}

// SearchRequest captures all parameters for a search query.
type SearchRequest struct {
	Query       string                 `json:"query"`
	Mode        string                 `json:"mode"` // "vector", "keyword", "hybrid"
	TopK        int                    `json:"top_k"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	MinScore    float64                `json:"min_score"`
	HybridAlpha float64                `json:"hybrid_alpha"` // weight for vector vs keyword
}

// SearchResult is a single ranked result from any search mode.
type SearchResult struct {
	ChunkID       string                 `json:"chunk_id"`
	DocumentID    string                 `json:"document_id"`
	DocumentTitle string                 `json:"document_title"`
	Content       string                 `json:"content"`
	Score         float64                `json:"score"`
	ScoreVector   float64                `json:"score_vector"`
	ScoreKeyword  float64                `json:"score_keyword"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// SearchResponse wraps the full set of results and metadata.
type SearchResponse struct {
	Results     []SearchResult `json:"results"`
	TotalHits   int            `json:"total_hits"`
	QueryTimeMs int64          `json:"query_time_ms"`
	Mode        string         `json:"mode"`
}

// Search dispatches to the appropriate search strategy.
func (svc *SearchService) Search(ctx context.Context, req SearchRequest) SearchResponse {
	start := time.Now()

	var resp SearchResponse
	switch req.Mode {
	case "vector":
		resp = svc.vectorSearch(ctx, req)
	case "keyword":
		resp = svc.keywordSearch(req)
	case "hybrid":
		resp = svc.hybridSearch(ctx, req)
	default:
		resp = svc.hybridSearch(ctx, req)
	}

	resp.QueryTimeMs = time.Since(start).Milliseconds()
	return resp
}

// vectorSearch performs pure embedding-based similarity search.
func (svc *SearchService) vectorSearch(ctx context.Context, req SearchRequest) SearchResponse {
	queryVector, err := svc.embedSvc.Provider.CreateEmbedding(req.Query)
	if err != nil {
		log.Printf("SearchService: Embedding failed: %v", err)
		return SearchResponse{Mode: "vector"}
	}

	results := svc.vectorStore.Search(ctx, queryVector, req.TopK)
	return svc.toSearchResponse(results, "vector")
}

// keywordSearch performs full-text / BM25 search via PostgreSQL tsvector.
func (svc *SearchService) keywordSearch(req SearchRequest) SearchResponse {
	results := svc.keywordSvc.Search(req.Query, req.TopK, req.Filters)

	var searchResults []SearchResult
	for _, r := range results {
		searchResults = append(searchResults, SearchResult{
			ChunkID:       r.ChunkID,
			DocumentID:    r.DocumentID,
			Content:       r.Content,
			Score:         r.Rank,
			ScoreKeyword:  r.Rank,
		})
	}

	return SearchResponse{
		Results:   searchResults,
		TotalHits: len(searchResults),
		Mode:      "keyword",
	}
}

// hybridSearch combines vector and keyword scores.
func (svc *SearchService) hybridSearch(ctx context.Context, req SearchRequest) SearchResponse {
	return svc.hybridSvc.Search(ctx, req)
}

// toSearchResponse converts raw vector store results to the unified format.
func (svc *SearchService) toSearchResponse(raw []vectorstore.SearchResult, mode string) SearchResponse {
	var results []SearchResult
	for _, r := range raw {
		results = append(results, SearchResult{
			ChunkID:    r.ChunkID,
			DocumentID: r.DocumentID,
			Content:    r.Content,
			Score:      r.Score,
			Metadata:   r.Metadata,
		})
	}
	return SearchResponse{
		Results:   results,
		TotalHits: len(results),
		Mode:      mode,
	}
}
