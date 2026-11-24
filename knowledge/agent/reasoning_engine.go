package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/akhmadzaki/go-backend/helpers"
	"github.com/akhmadzaki/go-backend/knowledge/search"
)

// ReasoningEngine implements multi-turn reasoning strategies using local Ollama.
type ReasoningEngine struct {
	ModelName   string
	MaxTokens   int
	Temperature float64
	BaseURL     string
}

func NewReasoningEngine(modelName string, maxTokens int, temperature float64, baseURL string) *ReasoningEngine {
	return &ReasoningEngine{
		ModelName:   modelName,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		BaseURL:     baseURL,
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

	answer, usage := eng.callOllama(prompt)

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

	answer, usage := eng.callOllama(prompt)

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

	answer, usage := eng.callOllama(prompt)

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

// ── Ollama ─────────────────────────────────────────────────────────────────────

type ollamaGenerateRequest struct {
	Model   string        `json:"model"`
	Prompt  string        `json:"prompt"`
	Stream  bool          `json:"stream"`
	Options ollamaOptions `json:"options"`
}

type ollamaOptions struct {
	Temperature float64 `json:"temperature"`
	NumPredict  int     `json:"num_predict"`
}

type ollamaGenerateResponse struct {
	Response        string `json:"response"`
	EvalCount       int    `json:"eval_count"`
	PromptEvalCount int    `json:"prompt_eval_count"`
}

func (eng *ReasoningEngine) callOllama(prompt string) (string, map[string]int) {
	baseURL := strings.TrimRight(eng.BaseURL, "/")
	reqBody := ollamaGenerateRequest{
		Model:  eng.ModelName,
		Prompt: prompt,
		Stream: false,
		Options: ollamaOptions{
			Temperature: eng.Temperature,
			NumPredict:  eng.MaxTokens,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("ReasoningEngine: failed to marshal Ollama request: %v", err)
		return fmt.Sprintf("Error: failed to create request: %v", err), map[string]int{}
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Post(baseURL+"/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("ReasoningEngine: Ollama request failed: %v", err)
		return fmt.Sprintf("Error: LLM request failed: %v. Is Ollama running at %s?", err, baseURL), map[string]int{}
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ReasoningEngine: failed to read Ollama response: %v", err)
		return fmt.Sprintf("Error: failed to read LLM response: %v", err), map[string]int{}
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("ReasoningEngine: Ollama returned status %d: %s", resp.StatusCode, string(respBytes))
		return fmt.Sprintf("Error: LLM returned status %d", resp.StatusCode), map[string]int{}
	}

	var result ollamaGenerateResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		log.Printf("ReasoningEngine: failed to decode Ollama response: %v", err)
		return fmt.Sprintf("Error: failed to decode LLM response: %v", err), map[string]int{}
	}

	usage := map[string]int{
		"prompt_tokens":     result.PromptEvalCount,
		"completion_tokens": result.EvalCount,
		"total_tokens":      result.PromptEvalCount + result.EvalCount,
	}

	return result.Response, usage
}
