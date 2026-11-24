package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/akhmadzaki/go-backend/db/models"
)

// ConversationRepository handles persistence for chat conversations and turns.
type ConversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) Create(ctx context.Context) *models.Conversation {
	row := r.db.QueryRowContext(ctx, "INSERT INTO conversations DEFAULT VALUES RETURNING id, title, turn_count, status, created_at, updated_at")

	conv := &models.Conversation{}
	err := row.Scan(&conv.ID, &conv.Title, &conv.TurnCount, &conv.Status, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		return nil
	}
	return conv
}

func (r *ConversationRepository) FindByID(ctx context.Context, id string) *models.Conversation {
	row := r.db.QueryRowContext(ctx, "SELECT id, title, turn_count, status, created_at, updated_at FROM conversations WHERE id = $1", id)

	conv := &models.Conversation{}
	err := row.Scan(&conv.ID, &conv.Title, &conv.TurnCount, &conv.Status, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		return nil
	}
	return conv
}

func (r *ConversationRepository) FindAll(ctx context.Context) []*models.Conversation {
	rows, err := r.db.QueryContext(ctx, "SELECT id, title, turn_count, status, created_at, updated_at FROM conversations ORDER BY updated_at DESC")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var convs []*models.Conversation
	for rows.Next() {
		conv := &models.Conversation{}
		if err := rows.Scan(&conv.ID, &conv.Title, &conv.TurnCount, &conv.Status, &conv.CreatedAt, &conv.UpdatedAt); err != nil {
			continue
		}
		convs = append(convs, conv)
	}
	return convs
}

func (r *ConversationRepository) Delete(ctx context.Context, id string) bool {
	_, err := r.db.ExecContext(ctx, "DELETE FROM conversations WHERE id = $1", id)
	return err == nil
}

func (r *ConversationRepository) SaveTurn(ctx context.Context, convID string, turn *models.ConversationTurn) bool {
	sourcesJSON, _ := json.Marshal(turn.Sources)
	tokenUsageJSON, _ := json.Marshal(turn.TokenUsage)

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO conversation_turns (conversation_id, role, content, sources, token_usage) VALUES ($1, $2, $3, $4, $5)",
		convID, turn.Role, turn.Content, sourcesJSON, tokenUsageJSON,
	)
	if err != nil {
		return false
	}

	r.db.ExecContext(ctx, "UPDATE conversations SET turn_count = turn_count + 1, updated_at = NOW() WHERE id = $1", convID)
	return true
}

func (r *ConversationRepository) FindTurns(ctx context.Context, convID string) []*models.ConversationTurn {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, conversation_id, role, content, sources, token_usage, created_at FROM conversation_turns WHERE conversation_id = $1 ORDER BY created_at ASC",
		convID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var turns []*models.ConversationTurn
	for rows.Next() {
		turn := &models.ConversationTurn{}
		var sourcesJSON, tokenUsageJSON []byte
		if err := rows.Scan(&turn.ID, &turn.ConversationID, &turn.Role, &turn.Content, &sourcesJSON, &tokenUsageJSON, &turn.CreatedAt); err != nil {
			continue
		}
		json.Unmarshal(sourcesJSON, &turn.Sources)
		json.Unmarshal(tokenUsageJSON, &turn.TokenUsage)
		turns = append(turns, turn)
	}
	return turns
}
