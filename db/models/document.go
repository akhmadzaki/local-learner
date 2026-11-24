package models

import "time"

// Document represents a knowledge-base document in the relational store.
type Document struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	Source     string                 `json:"source"` // "upload", "web", "api", "scrape"
	Status     string                 `json:"status"` // "pending", "processing", "indexed", "failed"
	ChunkCount int                    `json:"chunk_count"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}
