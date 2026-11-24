package models

// Conversation represents a multi-turn chat session with the agent.
type Conversation struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	TurnCount int    `json:"turn_count"`
	Status    string `json:"status"` // "active", "archived"
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ConversationTurn represents a single exchange within a conversation.
type ConversationTurn struct {
	ID             string           `json:"id"`
	ConversationID string           `json:"conversation_id"`
	Role           string           `json:"role"` // "user" | "assistant" | "system"
	Content        string           `json:"content"`
	Sources        []SourceRef      `json:"sources"` // JSON-encoded references
	TokenUsage     map[string]int   `json:"token_usage"`
	CreatedAt      string           `json:"created_at"`
}

// SourceRef links an assistant response back to the retrieved chunks.
type SourceRef struct {
	ChunkID    string  `json:"chunk_id"`
	DocumentID string  `json:"document_id"`
	Score      float64 `json:"score"`
}