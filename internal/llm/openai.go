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

// OpenAIProvider implements the Provider interface for OpenAI-compatible APIs
type OpenAIProvider struct {
	apiBase     string
	apiKey      string
	model       string
	maxTokens   int
	temperature float64
	topP        float64
	httpClient  *http.Client
}

// NewOpenAIProvider creates a new OpenAI-compatible provider
func NewOpenAIProvider(cfg config.ModelConfig) *OpenAIProvider {
	return &OpenAIProvider{
		apiBase:     strings.TrimSuffix(cfg.APIBase, "/"),
		apiKey:      cfg.APIKey,
		model:       cfg.Model,
		maxTokens:   cfg.MaxTokens,
		temperature: cfg.Temperature,
		topP:        cfg.TopP,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai-compatible"
}

func (p *OpenAIProvider) IsAvailable() error {
	if p.apiBase == "" {
		return fmt.Errorf("API base URL is not configured")
	}
	if p.apiKey == "" {
		return fmt.Errorf("API key is not configured")
	}
	return nil
}

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	apiReq := p.buildAPIRequest(req)
	apiReq.Stream = false

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Retry logic for rate limiting
	maxRetries := 3
	var resp *http.Response

	for attempt := 0; attempt <= maxRetries; attempt++ {
		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.apiBase+"/chat/completions", bytes.NewReader(body))
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

	var apiResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := apiResp.Choices[0]
	result := &ChatResponse{
		Content:      choice.Message.Content,
		FinishReason: choice.FinishReason,
		Usage: Usage{
			PromptTokens:     apiResp.Usage.PromptTokens,
			CompletionTokens: apiResp.Usage.CompletionTokens,
			TotalTokens:      apiResp.Usage.TotalTokens,
		},
	}

	if len(choice.Message.ToolCalls) > 0 {
		result.ToolCalls = make([]ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			result.ToolCalls[i] = ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}

	return result, nil
}

func (p *OpenAIProvider) ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	apiReq := p.buildAPIRequest(req)
	apiReq.Stream = true

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Retry logic for rate limiting
	maxRetries := 3
	var resp *http.Response

	for attempt := 0; attempt <= maxRetries; attempt++ {
		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.apiBase+"/chat/completions", bytes.NewReader(body))
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
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines
			if line == "" {
				continue
			}

			// Handle SSE data lines
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")

				// End of stream
				if data == "[DONE]" {
					ch <- StreamChunk{FinishReason: "stop"}
					return
				}

				var chunk openAIStreamChunk
				if err := json.Unmarshal([]byte(data), &chunk); err != nil {
					ch <- StreamChunk{Error: fmt.Errorf("failed to parse chunk: %w", err)}
					return
				}

				if len(chunk.Choices) > 0 {
					delta := chunk.Choices[0].Delta
					sc := StreamChunk{
						Delta:        delta.Content,
						FinishReason: chunk.Choices[0].FinishReason,
					}

					// Handle tool calls in streaming
					if len(delta.ToolCalls) > 0 {
						for _, tc := range delta.ToolCalls {
							sc.ToolCallIndex = tc.Index
							sc.ToolCalls = append(sc.ToolCalls, ToolCall{
								ID:   tc.ID,
								Type: tc.Type,
								Function: FunctionCall{
									Name:      tc.Function.Name,
									Arguments: tc.Function.Arguments,
								},
							})
						}
					}

					ch <- sc
				}

				// Extract usage from final chunk (when stream_options.include_usage=true)
				if chunk.Usage != nil {
					ch <- StreamChunk{
						Usage: &Usage{
							PromptTokens:     chunk.Usage.PromptTokens,
							CompletionTokens: chunk.Usage.CompletionTokens,
							TotalTokens:      chunk.Usage.TotalTokens,
						},
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- StreamChunk{Error: fmt.Errorf("stream read error: %w", err)}
		}
	}()

	return ch, nil
}

func (p *OpenAIProvider) buildAPIRequest(req ChatRequest) openAIRequest {
	apiReq := openAIRequest{
		Model:       req.Model,
		Messages:    make([]openAIMessage, len(req.Messages)),
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
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
	if req.TopP != nil {
		apiReq.TopP = *req.TopP
	} else {
		apiReq.TopP = p.topP
	}

	// Request usage in streaming responses
	if apiReq.Stream {
		apiReq.StreamOptions = &openAIStreamOptions{IncludeUsage: true}
	}

	for i, msg := range req.Messages {
		apiReq.Messages[i] = openAIMessage{
			Role:       string(msg.Role),
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
			Name:       msg.Name,
		}
		if len(msg.ToolCalls) > 0 {
			apiReq.Messages[i].ToolCalls = make([]openAIToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				apiReq.Messages[i].ToolCalls[j] = openAIToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: openAIFunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}
	}

	if len(req.Tools) > 0 {
		apiReq.Tools = make([]openAITool, len(req.Tools))
		for i, t := range req.Tools {
			apiReq.Tools[i] = openAITool{
				Type: t.Type,
				Function: openAIToolFunction{
					Name:        t.Function.Name,
					Description: t.Function.Description,
					Parameters:  t.Function.Parameters,
				},
			}
		}
	}

	return apiReq
}

func (p *OpenAIProvider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Accept", "text/event-stream")
}

// OpenAI API types (internal)

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Tools       []openAITool    `json:"tools,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	TopP        float64         `json:"top_p,omitempty"`
	Stream      bool            `json:"stream"`
	StreamOptions *openAIStreamOptions `json:"stream_options,omitempty"`
}

type openAIStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
	Name       string           `json:"name,omitempty"`
}

type openAIToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Index    int                `json:"index,omitempty"`
	Function openAIFunctionCall `json:"function"`
}

type openAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAITool struct {
	Type     string              `json:"type"`
	Function openAIToolFunction  `json:"function"`
}

type openAIToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type openAIResponse struct {
	ID      string           `json:"id"`
	Choices []openAIChoice   `json:"choices"`
	Usage   openAIUsage      `json:"usage"`
}

type openAIChoice struct {
	Message      openAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type openAIStreamChunk struct {
	ID      string               `json:"id"`
	Choices []openAIStreamChoice `json:"choices"`
	Usage   *openAIUsage         `json:"usage,omitempty"` // Only on final chunk with stream_options
}

type openAIStreamChoice struct {
	Delta        openAIStreamDelta `json:"delta"`
	FinishReason string            `json:"finish_reason"`
}

type openAIStreamDelta struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []openAIToolCall `json:"tool_calls,omitempty"`
}
