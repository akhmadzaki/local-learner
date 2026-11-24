package models

// Chunk represents a single text segment extracted from a document.
// Each chunk is independently embedded and searchable.
type Chunk struct {
	ID          string                 `json:"id"`
	DocumentID  string                 `json:"document_id"`
	Content     string                 `json:"content"`
	ChunkIndex  int                    `json:"chunk_index"`
	TokenCount  int                    `json:"token_count"`
	SearchVector string                `json:"-"` // PostgreSQL tsvector (populated by DB trigger)
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   string                 `json:"created_at"`
}