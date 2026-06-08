package mcp

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Client represents an MCP client connected to a single server
type Client struct {
	transport    Transport
	serverName   string
	serverInfo   ServerInfo
	capabilities ServerCapabilities
	nextID       atomic.Int64
	mu           sync.Mutex
	connected    bool
}

// NewClient creates a new MCP client
func NewClient(transport Transport) *Client {
	return &Client{
		transport: transport,
	}
}

// Connect starts the transport and performs the initialize handshake
func (c *Client) Connect() error {
	if err := c.transport.Start(); err != nil {
		return fmt.Errorf("transport start failed: %w", err)
	}

	// Send initialize request
	params := InitializeParams{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ClientCapabilities{
			Roots: &RootsCapability{ListChanged: false},
		},
		ClientInfo: MiMoClientInfo,
	}

	result, err := c.sendRequest(MethodInitialize, params)
	if err != nil {
		c.transport.Close()
		return fmt.Errorf("initialize failed: %w", err)
	}

	var initResult InitializeResult
	if err := json.Unmarshal(result, &initResult); err != nil {
		c.transport.Close()
		return fmt.Errorf("failed to parse initialize result: %w", err)
	}

	c.serverInfo = initResult.ServerInfo
	c.capabilities = initResult.Capabilities

	// Send initialized notification
	if err := c.sendNotification(MethodInitialized, nil); err != nil {
		c.transport.Close()
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}

	c.connected = true
	return nil
}

// ServerName returns the name of the connected server
func (c *Client) ServerName() string {
	if c.serverInfo.Name != "" {
		return c.serverInfo.Name
	}
	return c.serverName
}

// SetServerName sets a display name for this server
func (c *Client) SetServerName(name string) {
	c.serverName = name
}

// IsConnected returns whether the client is connected
func (c *Client) IsConnected() bool {
	return c.connected
}

// ListTools returns the list of tools available on the server
func (c *Client) ListTools() ([]Tool, error) {
	result, err := c.sendRequest(MethodToolsList, nil)
	if err != nil {
		return nil, fmt.Errorf("tools/list failed: %w", err)
	}

	var listResult ListToolsResult
	if err := json.Unmarshal(result, &listResult); err != nil {
		return nil, fmt.Errorf("failed to parse tools list: %w", err)
	}

	return listResult.Tools, nil
}

// CallTool calls a tool on the server
func (c *Client) CallTool(name string, arguments map[string]interface{}) (*CallToolResult, error) {
	params := CallToolParams{
		Name:      name,
		Arguments: arguments,
	}

	result, err := c.sendRequest(MethodToolsCall, params)
	if err != nil {
		return nil, fmt.Errorf("tools/call failed: %w", err)
	}

	var callResult CallToolResult
	if err := json.Unmarshal(result, &callResult); err != nil {
		return nil, fmt.Errorf("failed to parse tool call result: %w", err)
	}

	return &callResult, nil
}

// Ping sends a ping to the server
func (c *Client) Ping() error {
	_, err := c.sendRequest(MethodPing, nil)
	return err
}

// Close closes the client and the underlying transport
func (c *Client) Close() error {
	c.connected = false
	return c.transport.Close()
}

// sendRequest sends a JSON-RPC request and waits for the response
// 跳过 id=0 的通知消息，只返回真正的响应
func (c *Client) sendRequest(method string, params interface{}) (json.RawMessage, error) {
	id := int(c.nextID.Add(1))

	var paramsJSON json.RawMessage
	if params != nil {
		var err error
		paramsJSON, err = json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
	}

	req := Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  paramsJSON,
	}

	// 锁住整个 Send+Receive 周期，防止并发请求响应错位
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.transport.Send(req); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// 循环读取，跳过 id=0 的通知消息，直到收到真正的响应
	type recvResult struct {
		resp *Response
		err  error
	}
	ch := make(chan recvResult, 1)
	go func() {
		for {
			resp, err := c.transport.Receive()
			if err != nil {
				ch <- recvResult{nil, err}
				return
			}
			// id=0 是通知消息，跳过继续读
			if resp.ID != 0 {
				ch <- recvResult{resp, nil}
				return
			}
		}
	}()

	select {
	case <-time.After(60 * time.Second):
		return nil, fmt.Errorf("MCP server timeout (60s) waiting for response to %s", method)
	case result := <-ch:
		if result.err != nil {
			return nil, fmt.Errorf("failed to receive response: %w", result.err)
		}
		if result.resp.Error != nil {
			return nil, fmt.Errorf("server error (%d): %s", result.resp.Error.Code, result.resp.Error.Message)
		}
		return result.resp.Result, nil
	}
}

// sendNotification sends a JSON-RPC notification (no response expected)
func (c *Client) sendNotification(method string, params interface{}) error {
	var paramsJSON json.RawMessage
	if params != nil {
		var err error
		paramsJSON, err = json.Marshal(params)
		if err != nil {
			return fmt.Errorf("failed to marshal params: %w", err)
		}
	}

	notif := Notification{
		JSONRPC: "2.0",
		Method:  method,
		Params:  paramsJSON,
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.transport.Send(notif)
}
