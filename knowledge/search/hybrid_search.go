package search

import (
	"context"
	"log"
	"math"
	"sort"
	"time"

	"github.com/akhmadzaki/go-backend/knowledge/embeddings"
	"github.com/akhmadzaki/go-backend/knowledge/vectorstore"
)

// HybridSearch combines vector similarity and keyword (BM25) scores
// using Reciprocal Rank Fusion (RRF) or weighted linear interpolation.
type HybridSearch struct {
	vectorStore   *vectorstore.VectorStore
	embeddingProv embeddings.EmbeddingProvider
	keywordSvc    *KeywordSearch
	Alpha         float64 // 0 = pure keyword, 1 = pure vector
	K             int     // RRF constant (default 60)
}

func NewHybridSearch(vectorStore *vectorstore.VectorStore, embeddingProv embeddings.EmbeddingProvider, keywordSvc *KeywordSearch) *HybridSearch {
	return &HybridSearch{
		vectorStore:   vectorStore,
		embeddingProv: embeddingProv,
		keywordSvc:    keywordSvc,
		Alpha:         0.5,
		K:             60,
	}
}

// Search performs hybrid retrieval combining vector and keyword results.
func (hs *HybridSearch) Search(ctx context.Context, req SearchRequest) SearchResponse {
	start := time.Now()
	alpha := req.HybridAlpha
	if alpha == 0 {
		alpha = hs.Alpha
	}

	// 1. Generate query embedding
	queryVector, err := hs.embeddingProv.CreateEmbedding(req.Query)
	if err != nil {
		log.Printf("HybridSearch: embedding failed: %v", err)
		queryVector = make([]float32, hs.embeddingProv.Dimension())
	}

	// 2. Vector search
	rawVector := hs.vectorStore.Search(ctx, queryVector, req.TopK)
	vectorResults := make([]SearchResult, 0, len(rawVector))
	for _, r := range rawVector {
		vectorResults = append(vectorResults, SearchResult{
			ChunkID:    r.ChunkID,
			DocumentID: r.DocumentID,
			Content:    r.Content,
			Score:      r.Score,
		})
	}

	// 3. Keyword search
	rawKeyword := hs.keywordSvc.Search(req.Query, req.TopK, req.Filters)
	keywordResults := make([]KeywordResult, 0, len(rawKeyword))
	for _, r := range rawKeyword {
		keywordResults = append(keywordResults, r)
	}

	// 4. Fuse via RRF
	fused := hs.ReciprocalRankFusion(vectorResults, keywordResults, req.TopK, alpha)

	elapsed := time.Since(start).Milliseconds()
	return SearchResponse{
		Results:     fused,
		TotalHits:   len(fused),
		QueryTimeMs: elapsed,
		Mode:        "hybrid",
	}
}

// ReciprocalRankFusion merges two ranked lists using the RRF formula:
//
//	score(d) = α * Σ(1/(k + r_v(d))) + (1-α) * Σ(1/(k + r_k(d)))
func (hs *HybridSearch) ReciprocalRankFusion(vector []SearchResult, keyword []KeywordResult, topK int, alpha float64) []SearchResult {
	scores := make(map[string]float64)
	details := make(map[string]SearchResult)

	// Score vector results
	for i, r := range vector {
		rank := i + 1
		scores[r.ChunkID] += alpha * (1.0 / float64(hs.K+rank))
		r.ScoreVector = 1.0 / float64(hs.K+rank)
		details[r.ChunkID] = r
	}

	// Score keyword results
	for i, r := range keyword {
		rank := i + 1
		scores[r.ChunkID] += (1.0 - alpha) * (1.0 / float64(hs.K+rank))
		if existing, ok := details[r.ChunkID]; ok {
			existing.ScoreKeyword = 1.0 / float64(hs.K+rank)
			details[r.ChunkID] = existing
		} else {
			details[r.ChunkID] = SearchResult{
				ChunkID:      r.ChunkID,
				DocumentID:   r.DocumentID,
				Content:      r.Content,
				ScoreKeyword: 1.0 / float64(hs.K+rank),
			}
		}
	}

	// Sort by fused score descending
	type scoredItem struct {
		id    string
		score float64
	}
	var items []scoredItem
	for id, score := range scores {
		items = append(items, scoredItem{id, score})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].score > items[j].score
	})

	// Take topK
	topK = int(math.Min(float64(topK), float64(len(items))))
	var results []SearchResult
	for i := 0; i < topK; i++ {
		r := details[items[i].id]
		r.Score = items[i].score
		results = append(results, r)
	}

	return results
}
