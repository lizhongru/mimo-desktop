package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mimo-cli/mimo-cli/internal/config"
)

// AnthropicProvider implements the Provider interface for Anthropic-compatible APIs
type AnthropicProvider struct {
	apiBase     string
	apiKey      string
	model       string
	maxTokens   int
	temperature float64
	httpClient  *http.Client
}

// NewAnthropicProvider creates a new Anthropic-compatible provider
func NewAnthropicProvider(cfg config.ModelConfig) *AnthropicProvider {
	return &AnthropicProvider{
		apiBase:     strings.TrimSuffix(cfg.APIBase, "/"),
		apiKey:      cfg.APIKey,
		model:       cfg.Model,
		maxTokens:   cfg.MaxTokens,
		temperature: cfg.Temperature,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

func (p *AnthropicProvider) Name() string {
	return "anthropic-compatible"
}

func (p *AnthropicProvider) IsAvailable() error {
	if p.apiBase == "" {
		return fmt.Errorf("API base URL is not configured")
	}
	if p.apiKey == "" {
		return fmt.Errorf("API key is not configured")
	}
	return nil
}

// ListModels returns a static list of known Anthropic models
func (p *AnthropicProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	// Anthropic does not have a public models API, return known models
	return []ModelInfo{
		{ID: "claude-3-opus-20240229", OwnedBy: "anthropic"},
		{ID: "claude-3-sonnet-20240229", OwnedBy: "anthropic"},
		{ID: "claude-3-haiku-20240307", OwnedBy: "anthropic"},
		{ID: "claude-2.1", OwnedBy: "anthropic"},
		{ID: "claude-2.0", OwnedBy: "anthropic"},
		{ID: "claude-instant-1.2", OwnedBy: "anthropic"},
	}, nil
}

func (p *AnthropicProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	apiReq := p.buildAPIRequest(req, false)

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Retry logic for rate limiting
	maxRetries := 3
	var resp *http.Response

	for attempt := 0; attempt <= maxRetries; attempt++ {
		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.apiBase+"/v1/messages", bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		p.setHeaders(httpReq)

		resp, err = p.httpClient.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}

		// If rate limited, wait and retry
		if resp.StatusCode == 429 && attempt < maxRetries {
			resp.Body.Close()
			// Exponential backoff: 1s, 2s, 4s
			waitTime := time.Duration(1<<attempt) * time.Second
			time.Sleep(waitTime)
			continue
		}

		break
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var apiResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := &ChatResponse{
		FinishReason: apiResp.StopReason,
		Usage: Usage{
			PromptTokens:     apiResp.Usage.InputTokens,
			CompletionTokens: apiResp.Usage.OutputTokens,
			TotalTokens:      apiResp.Usage.InputTokens + apiResp.Usage.OutputTokens,
		},
	}

	// Extract content
	for _, block := range apiResp.Content {
		switch block.Type {
		case "text":
			result.Content += block.Text
		case "thinking":
			result.Thinking += block.Thinking
		case "tool_use":
			argsJSON, _ := json.Marshal(block.Input)
			result.ToolCalls = append(result.ToolCalls, ToolCall{
				ID:   block.ID,
				Type: "function",
				Function: FunctionCall{
					Name:      block.Name,
					Arguments: string(argsJSON),
				},
			})
		}
	}

	return result, nil
}

func (p *AnthropicProvider) ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	apiReq := p.buildAPIRequest(req, true)

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Retry logic for rate limiting
	maxRetries := 3
	var resp *http.Response

	for attempt := 0; attempt <= maxRetries; attempt++ {
		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.apiBase+"/v1/messages", bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		p.setHeaders(httpReq)

		resp, err = p.httpClient.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}

		// If rate limited, wait and retry
		if resp.StatusCode == 429 && attempt < maxRetries {
			resp.Body.Close()
			// Exponential backoff: 1s, 2s, 4s
			waitTime := time.Duration(1<<attempt) * time.Second
			time.Sleep(waitTime)
			continue
		}

		break
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	ch := make(chan StreamChunk, 100)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		var inputTokens int // 从 message_start 中提取

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			// SSE format: "event: xxx" or "data: xxx"
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")

				var event anthropicStreamEvent
				if err := json.Unmarshal([]byte(data), &event); err != nil {
					continue
				}

				switch event.Type {
				case "message_start":
					// 提取 input_tokens
					if event.Message != nil {
						inputTokens = event.Message.Usage.InputTokens
					}
				case "content_block_delta":
					switch event.Delta.Type {
					case "text_delta":
						ch <- StreamChunk{Delta: event.Delta.Text}
					case "input_json_delta":
						// Tool call argument streaming - include index for proper merging
						ch <- StreamChunk{
							ToolCallIndex: event.Index,
							ToolCalls: []ToolCall{
								{
									Function: FunctionCall{
										Arguments: event.Delta.PartialJSON,
									},
								},
							},
						}
					case "thinking_delta":
						ch <- StreamChunk{Thinking: event.Delta.Thinking}
					}

				case "content_block_start":
					if event.ContentBlock.Type == "tool_use" {
						ch <- StreamChunk{
							ToolCallIndex: event.Index,
							ToolCalls: []ToolCall{
								{
									ID:   event.ContentBlock.ID,
									Type: "function",
									Function: FunctionCall{
										Name: event.ContentBlock.Name,
									},
								},
							},
						}
					}

				case "message_delta":
					// Contains stop_reason and final output usage
					chunk := StreamChunk{FinishReason: event.Delta.StopReason}
					// 合并 message_start 的 input_tokens + message_delta 的 output_tokens
					outputTokens := 0
					if event.Usage != nil {
						outputTokens = event.Usage.OutputTokens
						// 如果 message_delta 中有 input_tokens，使用它（覆盖 message_start 的值）
						if event.Usage.InputTokens > 0 {
							inputTokens = event.Usage.InputTokens
						}
					}
					if inputTokens > 0 || outputTokens > 0 {
						chunk.Usage = &Usage{
							PromptTokens:     inputTokens,
							CompletionTokens: outputTokens,
							TotalTokens:      inputTokens + outputTokens,
						}
					}
					ch <- chunk
					return

				case "message_stop":
					ch <- StreamChunk{FinishReason: "stop"}
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- StreamChunk{Error: fmt.Errorf("stream read error: %w", err)}
		}
	}()

	return ch, nil
}

func (p *AnthropicProvider) buildAPIRequest(req ChatRequest, stream bool) anthropicRequest {
	apiReq := anthropicRequest{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
		Stream:    stream,
	}

	if apiReq.Model == "" {
		apiReq.Model = p.model
	}
	if apiReq.MaxTokens == 0 {
		apiReq.MaxTokens = p.maxTokens
	}
	if req.Temperature != nil {
		apiReq.Temperature = *req.Temperature
	} else {
		apiReq.Temperature = p.temperature
	}

	// Convert messages
	var systemPrompt string
	for _, msg := range req.Messages {
		if msg.Role == RoleSystem {
			systemPrompt = msg.Content
			continue
		}

		if msg.Role == RoleTool {
			// Tool result message
			apiReq.Messages = append(apiReq.Messages, anthropicMessage{
				Role: "user",
				Content: []anthropicContent{
					{
						Type:      "tool_result",
						ToolUseID: msg.ToolCallID,
						Content:   msg.Content,
					},
				},
			})
			continue
		}

		if msg.Role == RoleAssistant && len(msg.ToolCalls) > 0 {
			// Assistant message with tool calls
			var content []anthropicContent
			if msg.Content != "" {
				content = append(content, anthropicContent{
					Type: "text",
					Text: msg.Content,
				})
			}
			for _, tc := range msg.ToolCalls {
				var input map[string]interface{}
				if tc.Function.Arguments != "" {
					json.Unmarshal([]byte(tc.Function.Arguments), &input)
				}
				if input == nil {
					input = map[string]interface{}{}
				}
				content = append(content, anthropicContent{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Function.Name,
					Input: input,
				})
			}
			apiReq.Messages = append(apiReq.Messages, anthropicMessage{
				Role:    "assistant",
				Content: content,
			})
			continue
		}

		// Normal user/assistant message
		content := buildAnthropicContent(msg.Content, msg.Attachments)
		apiReq.Messages = append(apiReq.Messages, anthropicMessage{
			Role:    string(msg.Role),
			Content: content,
		})
	}

	apiReq.System = systemPrompt

	// Convert tools
	if len(req.Tools) > 0 {
		apiReq.Tools = make([]anthropicTool, len(req.Tools))
		for i, t := range req.Tools {
			apiReq.Tools[i] = anthropicTool{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				InputSchema: t.Function.Parameters,
			}
		}
	}

	return apiReq
}

func buildAnthropicContent(text string, attachments []Attachment) []anthropicContent {
	content := []anthropicContent{}
	if text != "" || len(attachments) == 0 {
		content = append(content, anthropicContent{
			Type: "text",
			Text: text,
		})
	}
	for _, att := range attachments {
		if strings.HasPrefix(att.Type, "image/") {
			if source, ok := parseAnthropicImageSource(att); ok {
				content = append(content, anthropicContent{
					Type:   "image",
					Source: source,
				})
			}
			continue
		}
		content = append(content, anthropicContent{
			Type: "text",
			Text: formatTextAttachment(att),
		})
	}
	return content
}

func parseAnthropicImageSource(att Attachment) (*anthropicImageSource, bool) {
	const marker = ";base64,"
	idx := strings.Index(att.DataURL, marker)
	if idx < 0 {
		return nil, false
	}
	mediaType := strings.TrimPrefix(att.DataURL[:idx], "data:")
	if mediaType == "" {
		mediaType = att.Type
	}
	return &anthropicImageSource{
		Type:      "base64",
		MediaType: mediaType,
		Data:      att.DataURL[idx+len(marker):],
	}, true
}

func (p *AnthropicProvider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
}

// Anthropic API types (internal)

type anthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature,omitempty"`
	System      string             `json:"system,omitempty"`
	Messages    []anthropicMessage `json:"messages"`
	Tools       []anthropicTool    `json:"tools,omitempty"`
	Stream      bool               `json:"stream"`
}

type anthropicMessage struct {
	Role    string             `json:"role"`
	Content []anthropicContent `json:"content"`
}

type anthropicContent struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
	Source    *anthropicImageSource  `json:"source,omitempty"`
	ToolUseID string                 `json:"tool_use_id,omitempty"`
	Content   string                 `json:"content,omitempty"`
	Thinking  string                 `json:"thinking,omitempty"`
}

type anthropicImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// toolUseContent is used for marshaling tool_use blocks with required input field
type toolUseContent struct {
	Type  string                 `json:"type"`
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// textContent is used for marshaling text and other content blocks without input field
type textContent struct {
	Type      string                `json:"type"`
	Text      string                `json:"text,omitempty"`
	Thinking  string                `json:"thinking,omitempty"`
	ToolUseID string                `json:"tool_use_id,omitempty"`
	Content   string                `json:"content,omitempty"`
	Source    *anthropicImageSource `json:"source,omitempty"`
}

// MarshalJSON custom marshaler to handle the Input field correctly
func (c anthropicContent) MarshalJSON() ([]byte, error) {
	if c.Type == "tool_use" {
		input := c.Input
		if input == nil {
			input = map[string]interface{}{}
		}
		return json.Marshal(toolUseContent{
			Type:  c.Type,
			ID:    c.ID,
			Name:  c.Name,
			Input: input,
		})
	}
	return json.Marshal(textContent{
		Type:      c.Type,
		Text:      c.Text,
		Thinking:  c.Thinking,
		ToolUseID: c.ToolUseID,
		Content:   c.Content,
		Source:    c.Source,
	})
}

type anthropicTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"input_schema"`
}

type anthropicResponse struct {
	ID         string             `json:"id"`
	Content    []anthropicContent `json:"content"`
	StopReason string             `json:"stop_reason"`
	Usage      anthropicUsage     `json:"usage"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicStreamEvent struct {
	Type         string                `json:"type"`
	Index        int                   `json:"index"`
	Delta        anthropicDelta        `json:"delta"`
	ContentBlock anthropicContentBlock `json:"content_block"`
	Message      *anthropicResponse    `json:"message"`
	Usage        *anthropicUsage       `json:"usage"`
}

type anthropicDelta struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	PartialJSON string `json:"partial_json"`
	StopReason  string `json:"stop_reason"`
	Thinking    string `json:"thinking"`
	Signature   string `json:"signature"`
}

type anthropicContentBlock struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}
