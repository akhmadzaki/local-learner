package agent

import (
	"context"
	"log"
	"time"

	"github.com/akhmadzaki/go-backend/db/models"
	"github.com/akhmadzaki/go-backend/db/repository"
)

// ConversationManager handles multi-turn conversation state, persistence,
// and context window management for the agent.
type ConversationManager struct {
	convRepo   *repository.ConversationRepository
	maxHistory int // max turns to keep in context
}

func NewConversationManager(convRepo *repository.ConversationRepository, maxHistory int) *ConversationManager {
	return &ConversationManager{
		convRepo:   convRepo,
		maxHistory: maxHistory,
	}
}

// ConversationTurn represents a single exchange in a conversation.
type ConversationTurn struct {
	Role       string                   `json:"role"` // "user" | "assistant" | "system"
	Content    string                   `json:"content"`
	Sources    []map[string]interface{} `json:"sources,omitempty"`
	Timestamp  int64                    `json:"timestamp"`
	TokenCount int                      `json:"token_count"`
}

// CreateConversation initialises a new conversation session.
func (mgr *ConversationManager) CreateConversation(ctx context.Context) string {
	conv := mgr.convRepo.Create(ctx)
	if conv == nil {
		log.Printf("ConversationManager: Failed to create conversation")
		return ""
	}
	log.Printf("ConversationManager: Created conversation %s", conv.ID)
	return conv.ID
}

// GetHistory retrieves the conversation history, respecting the max context window.
func (mgr *ConversationManager) GetHistory(ctx context.Context, convID string) []ConversationTurn {
	turns := mgr.convRepo.FindTurns(ctx, convID)

	// Convert to ConversationTurn format
	var result []ConversationTurn
	for _, t := range turns {
		var sources []map[string]interface{}
		for _, s := range t.Sources {
			sources = append(sources, map[string]interface{}{
				"chunk_id":    s.ChunkID,
				"document_id": s.DocumentID,
				"score":       s.Score,
			})
		}

		result = append(result, ConversationTurn{
			Role:    t.Role,
			Content: t.Content,
			Sources: sources,
		})
	}

	// Trim to maxHistory if needed
	if len(result) > mgr.maxHistory {
		result = result[len(result)-mgr.maxHistory:]
	}

	return result
}

// SaveTurn persists a single user-assistant exchange.
func (mgr *ConversationManager) SaveTurn(ctx context.Context, convID string, query string, answer string, sources []interface{}) {
	if convID == "" {
		convID = mgr.CreateConversation(ctx)
	}

	now := time.Now().Unix()

	userTurn := models.ConversationTurn{
		Role:    "user",
		Content: query,
	}
	mgr.convRepo.SaveTurn(ctx, convID, &userTurn)

	assistantTurn := models.ConversationTurn{
		Role:    "assistant",
		Content: answer,
	}
	mgr.convRepo.SaveTurn(ctx, convID, &assistantTurn)

	_ = now
}

// DeleteConversation removes a conversation and all its turns.
func (mgr *ConversationManager) DeleteConversation(ctx context.Context, convID string) bool {
	return mgr.convRepo.Delete(ctx, convID)
}

// ListConversations returns all conversations.
func (mgr *ConversationManager) ListConversations(ctx context.Context) []map[string]interface{} {
	convs := mgr.convRepo.FindAll(ctx)
	var result []map[string]interface{}
	for _, c := range convs {
		result = append(result, map[string]interface{}{
			"id":         c.ID,
			"title":      c.Title,
			"turn_count": c.TurnCount,
			"status":     c.Status,
			"created_at": c.CreatedAt,
			"updated_at": c.UpdatedAt,
		})
	}
	return result
}
