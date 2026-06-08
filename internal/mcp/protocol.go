package mcp

import "encoding/json"

// JSON-RPC 2.0 message types

// Request represents a JSON-RPC 2.0 request
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response represents a JSON-RPC 2.0 response
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC 2.0 error
type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return e.Message
}

// Notification represents a JSON-RPC 2.0 notification (no ID)
type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// MCP Protocol types

// InitializeParams represents the params for initialize request
type InitializeParams struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

// ClientCapabilities represents what the client supports
type ClientCapabilities struct {
	Roots    *RootsCapability    `json:"roots,omitempty"`
	Sampling *SamplingCapability `json:"sampling,omitempty"`
}

// RootsCapability represents roots support
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability represents sampling support
type SamplingCapability struct{}

// ClientInfo represents information about the client
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult represents the result of initialize
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

// ServerCapabilities represents what the server supports
type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
	Logging   *LoggingCapability   `json:"logging,omitempty"`
}

// ToolsCapability represents tools support
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability represents resources support
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptsCapability represents prompts support
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// LoggingCapability represents logging support
type LoggingCapability struct{}

// ServerInfo represents information about the server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ListToolsResult represents the result of tools/list
type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

// CallToolParams represents the params for tools/call
type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// CallToolResult represents the result of tools/call
type CallToolResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// Content represents a content block in MCP
type Content struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

// MCP protocol constants
const (
	ProtocolVersion = "2024-11-05"

	MethodInitialize     = "initialize"
	MethodInitialized    = "notifications/initialized"
	MethodToolsList      = "tools/list"
	MethodToolsCall      = "tools/call"
	MethodPing           = "ping"
)

// ClientInfo for MiMo CLI
var MiMoClientInfo = ClientInfo{
	Name:    "mimo-cli",
	Version: "0.1.0",
}
