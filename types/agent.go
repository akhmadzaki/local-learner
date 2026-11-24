package types

// AgentMode defines the reasoning depth for the agent.
type AgentMode string

const (
	AgentModeDirect   AgentMode = "direct"   // Simple Q&A with context
	AgentModeReasoned AgentMode = "reasoned" // Chain-of-thought
	AgentModeDeep     AgentMode = "deep"     // Multi-step with tools
)

// AgentCapability enumerates what the agent can do.
type AgentCapability string

const (
	CapRetrieve  AgentCapability = "retrieve"
	CapReason    AgentCapability = "reason"
	CapSummarise AgentCapability = "summarise"
	CapCompare   AgentCapability = "compare"
	CapCodeExec  AgentCapability = "code_execution"
	CapWebSearch AgentCapability = "web_search"
)

// TokenUsage tracks LLM token consumption.
type TokenUsage struct {
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	TotalTokens      int    `json:"total_tokens"`
	ModelName        string `json:"model_name"`
}
