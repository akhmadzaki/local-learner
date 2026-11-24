package models

// Embedding stores the vector metadata alongside the chunk reference.
// The actual vector lives in the pgvector-enabled table.
type Embedding struct {
	ID         string `json:"id"`
	ChunkID    string `json:"chunk_id"`
	DocumentID string `json:"document_id"`
	ModelName  string `json:"model_name"`
	Dimension  int    `json:"dimension"`
	CreatedAt  string `json:"created_at"`
}