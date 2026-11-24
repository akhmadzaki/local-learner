package agent

import "log"

// Tool is the interface for all agent-callable tools.
type Tool interface {
	Name() string
	Description() string
	Execute(args map[string]interface{}) map[string]interface{}
}

// ToolCall represents a tool invocation request from the reasoning engine.
type ToolCall struct {
	ToolName string                 `json:"tool_name"`
	Args     map[string]interface{} `json:"args"`
	Result   map[string]interface{} `json:"result"`
}

// ToolDefinition is used for registration and LLM function-calling schemas.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolRegistry manages the set of tools the agent can invoke during reasoning.
type ToolRegistry struct {
	tools map[string]Tool
}

func NewToolRegistry() *ToolRegistry {
	reg := &ToolRegistry{
		tools: make(map[string]Tool),
	}
	reg.Register(&WebSearchTool{})
	reg.Register(&CalculatorTool{})
	return reg
}

// Register adds a tool to the registry.
func (reg *ToolRegistry) Register(tool Tool) {
	reg.tools[tool.Name()] = tool
	log.Printf("ToolRegistry: Registered tool %s", tool.Name())
}

// ExecuteTools runs a batch of tool calls.
func (reg *ToolRegistry) ExecuteTools(calls []ToolCall) {
	for _, call := range calls {
		tool, ok := reg.tools[call.ToolName]
		if !ok {
			log.Printf("ToolRegistry: Unknown tool %s", call.ToolName)
			continue
		}
		call.Result = tool.Execute(call.Args)
	}
}

// GetDefinitions returns all tool definitions for LLM function-calling.
func (reg *ToolRegistry) GetDefinitions() []ToolDefinition {
	var defs []ToolDefinition
	for _, tool := range reg.tools {
		defs = append(defs, ToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  make(map[string]interface{}),
		})
	}
	return defs
}

// ── Built-in tools ───────────────────────────────────────────────────────────

type WebSearchTool struct{}

func (t *WebSearchTool) Name() string        { return "web_search" }
func (t *WebSearchTool) Description() string { return "Search the web for current information" }
func (t *WebSearchTool) Execute(args map[string]interface{}) map[string]interface{} {
	log.Printf("WebSearchTool: stub - args: %v", args)
	return map[string]interface{}{"results": []interface{}{}}
}

type CalculatorTool struct{}

func (t *CalculatorTool) Name() string        { return "calculator" }
func (t *CalculatorTool) Description() string { return "Perform mathematical calculations" }
func (t *CalculatorTool) Execute(args map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"result": 0}
}
