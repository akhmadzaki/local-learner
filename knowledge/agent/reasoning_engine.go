package agent

import (
	"fmt"
	"log"
	"strings"

	"github.com/akhmadzaki/go-backend/helpers"
	"github.com/akhmadzaki/go-backend/knowledge/search"
)

// ReasoningEngine implements multi-turn reasoning strategies.
type ReasoningEngine struct {
	LLMProvider string
	ModelName   string
	MaxTokens   int
	Temperature float64
}

func NewReasoningEngine(llmProvider, modelName string, maxTokens int, temperature float64) *ReasoningEngine {
	return &ReasoningEngine{
		LLMProvider: llmProvider,
		ModelName:   modelName,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}
}

// ReasoningInput is the full context provided to the reasoning engine.
type ReasoningInput struct {
	Query   string                `json:"query"`
	Context []search.SearchResult `json:"context"`
	History []ConversationTurn    `json:"history"`
	Mode    string                `json:"mode"`
}

// ReasoningOutput is the structured result of the reasoning process.
type ReasoningOutput struct {
	Answer        string         `json:"answer"`
	Trace         string         `json:"trace"`
	Confidence    float64        `json:"confidence"`
	RequiresTools bool           `json:"requires_tools"`
	ToolCalls     []ToolCall     `json:"tool_calls,omitempty"`
	TokenUsage    map[string]int `json:"token_usage"`
}

// Reason runs the reasoning loop based on the selected mode.
func (eng *ReasoningEngine) Reason(input ReasoningInput) ReasoningOutput {
	switch input.Mode {
	case "direct":
		return eng.DirectReason(input)
	case "reasoned":
		return eng.ChainOfThought(input)
	case "deep":
		return eng.DeepReason(input)
	default:
		return eng.DirectReason(input)
	}
}

// DirectReason performs simple context-augmented Q&A.
func (eng *ReasoningEngine) DirectReason(input ReasoningInput) ReasoningOutput {
	context := eng.FormatContext(helpers.ToInterfaceSlice(input.Context))
	prompt := fmt.Sprintf("Answer the question based on the provided context.\n\nContext:\n%s\n\nQuestion: %s", context, input.Query)

	answer, usage := eng.CallLLM(prompt)

	return ReasoningOutput{
		Answer:     answer,
		Trace:      "",
		Confidence: 0.9,
		TokenUsage: usage,
	}
}

// ChainOfThought performs step-by-step reasoning with explicit trace.
func (eng *ReasoningEngine) ChainOfThought(input ReasoningInput) ReasoningOutput {
	context := eng.FormatContext(helpers.ToInterfaceSlice(input.Context))
	prompt := fmt.Sprintf("Think step by step to answer the question.\n\nContext:\n%s\n\nQuestion: %s\n\nLet me reason through this step by step.", context, input.Query)

	answer, usage := eng.CallLLM(prompt)

	return ReasoningOutput{
		Answer:     answer,
		Trace:      answer,
		Confidence: 0.85,
		TokenUsage: usage,
	}
}

// DeepReason performs multi-step reasoning with tool calls and reflection.
func (eng *ReasoningEngine) DeepReason(input ReasoningInput) ReasoningOutput {
	log.Printf("ReasoningEngine: Deep reasoning mode activated")

	context := eng.FormatContext(helpers.ToInterfaceSlice(input.Context))
	prompt := fmt.Sprintf("You are a deep reasoning agent. Analyse the question thoroughly.\n\nContext:\n%s\n\nQuestion: %s", context, input.Query)

	answer, usage := eng.CallLLM(prompt)

	return ReasoningOutput{
		Answer:        answer,
		Trace:         "Deep reasoning trace",
		Confidence:    0.8,
		RequiresTools: false,
		ToolCalls:     nil,
		TokenUsage:    usage,
	}
}

// FormatContext serialises search results into a prompt-friendly string.
func (eng *ReasoningEngine) FormatContext(results []interface{}) string {
	var buf strings.Builder
	for i, r := range results {
		result, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		content, _ := result["content"].(string)
		buf.WriteString(fmt.Sprintf("[%d] %s\n---\n", i+1, content))
	}
	return buf.String()
}

// CallLLM is a stub for LLM invocation (OpenAI, Anthropic, local, etc.).
func (eng *ReasoningEngine) CallLLM(prompt string) (string, map[string]int) {
	log.Printf("ReasoningEngine: LLM call (stub) - prompt length: %d", len(prompt))
	return fmt.Sprintf("This is a stub response. Configure an LLM provider (local/Ollama recommended). Model: %s", eng.ModelName),
		map[string]int{"prompt_tokens": len(prompt) / 4, "completion_tokens": 50, "total_tokens": len(prompt)/4 + 50}
}
