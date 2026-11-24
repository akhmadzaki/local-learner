package dto

// ── Request DTOs ─────────────────────────────────────────────────────────────

type ChatRequestDTO struct {
	Message        string `json:"message"`
	ConversationID string `json:"conversation_id,omitempty"`
	Mode           string `json:"mode"` // "direct", "reasoned", "deep"
	StreamResponse bool   `json:"stream_response,omitempty"`
}

// ── Response DTOs ────────────────────────────────────────────────────────────

type ChatResponseDTO struct {
	Answer         string            `json:"answer"`
	Sources        []SearchResultDTO `json:"sources"`
	ReasoningTrace string            `json:"reasoning_trace"`
	Confidence     float64           `json:"confidence"`
	ConversationID string            `json:"conversation_id"`
	TokenUsage     map[string]int    `json:"token_usage"`
}

type ConversationDTO struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	TurnCount int    `json:"turn_count"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ConversationListDTO struct {
	Conversations []ConversationDTO `json:"conversations"`
	Total         int               `json:"total"`
}
