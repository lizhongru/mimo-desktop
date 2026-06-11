package llm

// Role represents the role of a message sender
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message represents a single message in the conversation
type Message struct {
	Role        Role         `json:"role"`
	Content     string       `json:"content"`
	Attachments []Attachment `json:"attachments,omitempty"`
	ToolCalls   []ToolCall   `json:"tool_calls,omitempty"`
	ToolCallID  string       `json:"tool_call_id,omitempty"`
	Name        string       `json:"name,omitempty"`
}

// Attachment represents a file attachment (image or document)
type Attachment struct {
	Name    string `json:"name"`
	Type    string `json:"type"`    // MIME type, e.g. "image/png", "application/pdf"
	DataURL string `json:"dataUrl"` // data:... base64 URL
}

// ToolCall represents a tool call from the model
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall represents a function call within a tool call
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolDefinition represents a tool available to the model
type ToolDefinition struct {
	Type     string             `json:"type"`
	Function ToolFuncDefinition `json:"function"`
}

// ToolFuncDefinition describes a function tool
type ToolFuncDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// StreamChunk represents a single chunk from a streaming response
type StreamChunk struct {
	Delta        string    `json:"delta"`
	Thinking     string    `json:"thinking,omitempty"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	ToolCallIndex int      `json:"tool_call_index,omitempty"` // Index of the tool call in the response
	FinishReason string    `json:"finish_reason,omitempty"`
	Usage        *Usage    `json:"usage,omitempty"` // Only set on final chunk
	Error        error     `json:"-"`
}

// ChatRequest represents a request to the chat API
type ChatRequest struct {
	Model       string           `json:"model"`
	Messages    []Message        `json:"messages"`
	Tools       []ToolDefinition `json:"tools,omitempty"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Temperature *float64         `json:"temperature,omitempty"`
	TopP        *float64         `json:"top_p,omitempty"`
	Stream          bool           `json:"stream"`
	ReasoningEffort string         `json:"reasoning_effort,omitempty"` // low, medium, high
}

// Usage represents token usage statistics
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatResponse represents a non-streaming response
type ChatResponse struct {
	Content      string     `json:"content"`
	Thinking     string     `json:"thinking,omitempty"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	FinishReason string     `json:"finish_reason"`
	Usage        Usage      `json:"usage"`
}

// ModelInfo represents information about an available model
type ModelInfo struct {
	ID           string   `json:"id"`
	OwnedBy      string   `json:"owned_by,omitempty"`
	Description  string   `json:"description,omitempty"`
	ContextWindow int     `json:"context_window,omitempty"`
	MaxOutput    int      `json:"max_output,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// ModelCapabilities represents the capabilities of a model
type ModelCapabilities struct {
	ContextWindow int    `json:"context_window"` // Max context tokens
	MaxOutput     int    `json:"max_output"`     // Max output tokens
	SupportsVision bool  `json:"supports_vision"`
	SupportsTools  bool  `json:"supports_tools"`
	SupportsStreaming bool `json:"supports_streaming"`
}
