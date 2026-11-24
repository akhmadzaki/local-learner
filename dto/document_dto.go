package dto

// ── Request DTOs ─────────────────────────────────────────────────────────────

type IngestDocumentRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Source  string `json:"source"`
}

type UpdateDocumentRequest struct {
	Title  string `json:"title,omitempty"`
	Status string `json:"status,omitempty"`
}

// ── Response DTOs ────────────────────────────────────────────────────────────

type DocumentResponse struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Source     string `json:"source"`
	Status     string `json:"status"`
	ChunkCount int    `json:"chunk_count"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type DocumentListResponse struct {
	Documents []DocumentResponse `json:"documents"`
	Total     int                `json:"total"`
	Offset    int                `json:"offset"`
	Limit     int                `json:"limit"`
}
