package types

// SearchMode defines the available retrieval strategies.
type SearchMode string

const (
	SearchModeVector  SearchMode = "vector"
	SearchModeKeyword SearchMode = "keyword"
	SearchModeHybrid  SearchMode = "hybrid"
)

// SearchFilter defines constraints applied during retrieval.
type SearchFilter struct {
	DocumentIDs []string `json:"document_ids,omitempty"`
	Sources     []string `json:"sources,omitempty"`
	DateFrom    string   `json:"date_from,omitempty"`
	DateTo      string   `json:"date_to,omitempty"`
	MinScore    float64  `json:"min_score"`
	MaxResults  int      `json:"max_results"`
}

// SortOrder for search results.
type SortOrder string

const (
	SortByRelevance SortOrder = "relevance"
	SortByDate      SortOrder = "date"
	SortBySource    SortOrder = "source"
)
