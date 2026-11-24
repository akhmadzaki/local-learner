package dto

// ── Request DTOs ─────────────────────────────────────────────────────────────

type SearchRequestDTO struct {
	Query       string                 `json:"query"`
	Mode        string                 `json:"mode"` // "vector", "keyword", "hybrid"
	TopK        int                    `json:"top_k"`
	MinScore    float64                `json:"min_score"`
	HybridAlpha float64                `json:"hybrid_alpha"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
}

// ── Response DTOs ────────────────────────────────────────────────────────────

type SearchResultDTO struct {
	ChunkID       string  `json:"chunk_id"`
	DocumentID    string  `json:"document_id"`
	DocumentTitle string  `json:"document_title"`
	Content       string  `json:"content"`
	Score         float64 `json:"score"`
	ScoreVector   float64 `json:"score_vector"`
	ScoreKeyword  float64 `json:"score_keyword"`
}

type SearchResponseDTO struct {
	Results     []SearchResultDTO `json:"results"`
	TotalHits   int               `json:"total_hits"`
	QueryTimeMs int               `json:"query_time_ms"`
	Mode        string            `json:"mode"`
}
