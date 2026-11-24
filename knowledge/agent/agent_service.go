package agent

import (
	"context"
	"log"

	"github.com/akhmadzaki/go-backend/helpers"
	"github.com/akhmadzaki/go-backend/knowledge/search"
)

// AgentService implements the agentic RAG loop: plan → retrieve → reason → respond.
type AgentService struct {
	searchSvc    *search.SearchService
	reasonEngine *ReasoningEngine
	toolReg      *ToolRegistry
	convManager  *ConversationManager
}

func NewAgentService(searchSvc *search.SearchService, convManager *ConversationManager, llmProvider, modelName string, maxTokens int, temperature float64) *AgentService {
	return &AgentService{
		searchSvc:    searchSvc,
		reasonEngine: NewReasoningEngine(llmProvider, modelName, maxTokens, temperature),
		toolReg:      NewToolRegistry(),
		convManager:  convManager,
	}
}

// AgentRequest captures the user's query and conversation context.
type AgentRequest struct {
	Query          string `json:"query"`
	ConversationID string `json:"conversation_id,omitempty"`
	Mode           string `json:"mode"` // "direct", "reasoned", "deep"
	StreamResponse bool   `json:"stream_response,omitempty"`
}

// AgentResponse is the final answer with provenance and reasoning trace.
type AgentResponse struct {
	Answer         string                `json:"answer"`
	Sources        []search.SearchResult `json:"sources"`
	ReasoningTrace string                `json:"reasoning_trace"`
	Confidence     float64               `json:"confidence"`
	ConversationID string                `json:"conversation_id"`
	TokenUsage     map[string]int        `json:"token_usage"`
}

// ProcessMessage is the main entry point for agentic reasoning.
func (svc *AgentService) ProcessMessage(ctx context.Context, req AgentRequest) AgentResponse {
	log.Printf("AgentService: Processing message: %s", req.Query)

	// 1. Load conversation history
	history := svc.convManager.GetHistory(ctx, req.ConversationID)

	// 2. Retrieve relevant context (hybrid search)
	searchReq := search.SearchRequest{
		Query: req.Query,
		Mode:  "hybrid",
		TopK:  10,
	}
	searchResp := svc.searchSvc.Search(ctx, searchReq)

	// 3. Run reasoning engine
	reasoning := svc.reasonEngine.Reason(ReasoningInput{
		Query:   req.Query,
		Context: searchResp.Results,
		History: history,
		Mode:    req.Mode,
	})

	// 4. Optionally invoke tools
	if reasoning.RequiresTools {
		svc.toolReg.ExecuteTools(reasoning.ToolCalls)
	}

	// 5. Persist conversation
	svc.convManager.SaveTurn(ctx, req.ConversationID, req.Query, reasoning.Answer, helpers.ToInterfaceSlice(searchResp.Results))

	return AgentResponse{
		Answer:         reasoning.Answer,
		Sources:        searchResp.Results,
		ReasoningTrace: reasoning.Trace,
		Confidence:     reasoning.Confidence,
		ConversationID: req.ConversationID,
	}
}

// NewConversation creates a fresh conversation session.
func (svc *AgentService) NewConversation(ctx context.Context) string {
	return svc.convManager.CreateConversation(ctx)
}

// GetHistory retrieves the full conversation history.
func (svc *AgentService) GetHistory(ctx context.Context, convID string) []ConversationTurn {
	return svc.convManager.GetHistory(ctx, convID)
}

// ListConversations returns all conversations.
func (svc *AgentService) ListConversations(ctx context.Context) []map[string]interface{} {
	return svc.convManager.ListConversations(ctx)
}

// DeleteConversation removes a conversation.
func (svc *AgentService) DeleteConversation(ctx context.Context, convID string) bool {
	return svc.convManager.DeleteConversation(ctx, convID)
}
