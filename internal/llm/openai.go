package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	// Extract content string from raw message (may be string or array)
	var contentStr string
	if len(choice.Message.Content) > 0 {
		if err := json.Unmarshal(choice.Message.Content, &contentStr); err != nil {
			var parts []map[string]interface{}
			if json.Unmarshal(choice.Message.Content, &parts) == nil {
				var texts []string
				for _, p := range parts {
					if t, ok := p["text"].(string); ok {
						texts = append(texts, t)
					}
				}
				contentStr = strings.Join(texts, "\n")
			}
		}
	}
	result := &ChatResponse{
		Content:      contentStr,
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

const maxTextAttachmentChars = 200000

// buildOpenAIContent builds the content field for OpenAI API.
// When there are no attachments, it returns a JSON-quoted string.
// When attachments exist, it returns a content array with text + image_url parts.
func buildOpenAIContent(text string, attachments []Attachment) json.RawMessage {
	if len(attachments) == 0 {
		// Simple string content
		raw, _ := json.Marshal(text)
		return raw
	}

	// Build content array: [text_part, image_part, ...]
	parts := []map[string]interface{}{}
	if text != "" {
		parts = append(parts, map[string]interface{}{
			"type": "text",
			"text": text,
		})
	}
	for _, att := range attachments {
		if strings.HasPrefix(att.Type, "image/") {
			parts = append(parts, map[string]interface{}{
				"type": "image_url",
				"image_url": map[string]string{
					"url": att.DataURL,
				},
			})
		} else {
			attachmentText := formatTextAttachment(att)
			parts = append(parts, map[string]interface{}{
				"type": "text",
				"text": attachmentText,
			})
		}
	}
	raw, _ := json.Marshal(parts)
	return raw
}

func formatTextAttachment(att Attachment) string {
	decoded, ok := decodeTextAttachment(att)
	if !ok {
		return fmt.Sprintf("[Attached file: %s (%s)]", att.Name, att.Type)
	}
	if len(decoded) > maxTextAttachmentChars {
		decoded = decoded[:maxTextAttachmentChars] + "\n[Attachment truncated]"
	}
	return fmt.Sprintf("[Attached file: %s (%s)]\n%s", att.Name, att.Type, decoded)
}

func decodeTextAttachment(att Attachment) (string, bool) {
	if !isTextAttachment(att) {
		return "", false
	}
	const marker = ";base64,"
	if idx := strings.Index(att.DataURL, marker); idx >= 0 {
		decoded, err := base64.StdEncoding.DecodeString(att.DataURL[idx+len(marker):])
		if err != nil {
			return "", false
		}
		return string(decoded), true
	}
	if idx := strings.Index(att.DataURL, ","); idx >= 0 {
		decoded, err := url.QueryUnescape(att.DataURL[idx+1:])
		if err != nil {
			return "", false
		}
		return decoded, true
	}
	return "", false
}

func isTextAttachment(att Attachment) bool {
	if strings.HasPrefix(att.Type, "text/") {
		return true
	}
	switch strings.ToLower(att.Type) {
	case "application/json", "application/javascript", "application/xml", "application/yaml", "application/x-yaml":
		return true
	}
	name := strings.ToLower(att.Name)
	for _, ext := range []string{
		".md", ".markdown", ".txt", ".json", ".csv", ".tsv", ".yaml", ".yml", ".xml",
		".js", ".jsx", ".ts", ".tsx", ".css", ".scss", ".html", ".go", ".py", ".java",
		".c", ".cc", ".cpp", ".h", ".hpp", ".rs", ".toml", ".ini", ".env", ".sql",
	} {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}
func (p *OpenAIProvider) buildAPIRequest(req ChatRequest) openAIRequest {
	apiReq := openAIRequest{
		Model:     req.Model,
		Messages:  make([]openAIMessage, len(req.Messages)),
		MaxTokens: req.MaxTokens,
		Stream:    req.Stream,
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

	// Set reasoning effort
	if req.ReasoningEffort != "" {
		apiReq.ReasoningEffort = req.ReasoningEffort
	}

	// Request usage in streaming responses
	if apiReq.Stream {
		apiReq.StreamOptions = &openAIStreamOptions{IncludeUsage: true}
	}

	for i, msg := range req.Messages {
		apiReq.Messages[i] = openAIMessage{
			Role:       string(msg.Role),
			Content:    buildOpenAIContent(msg.Content, msg.Attachments),
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

// MiMo model capabilities (static data from platform documentation)
var mimoModelCapabilities = map[string]ModelInfo{
	"mimo-v2.5-pro": {
		ID: "mimo-v2.5-pro", Description: "复杂推理、深度分析、长文档处理",
		ContextWindow: 1000000, MaxOutput: 128000,
		Capabilities: []string{"text_generation", "deep_thinking", "streaming", "function_calling", "structured_output", "web_search"},
	},
	"mimo-v2-pro": {
		ID: "mimo-v2-pro", Description: "复杂推理、深度分析、长文档处理",
		ContextWindow: 1000000, MaxOutput: 128000,
		Capabilities: []string{"text_generation", "deep_thinking", "streaming", "function_calling", "structured_output", "web_search"},
	},
	"mimo-v2.5": {
		ID: "mimo-v2.5", Description: "图片、音频、视频内容理解",
		ContextWindow: 1000000, MaxOutput: 128000,
		Capabilities: []string{"text_generation", "multimodal", "deep_thinking", "streaming", "function_calling", "structured_output", "web_search"},
	},
	"mimo-v2-omni": {
		ID: "mimo-v2-omni", Description: "图片、音频、视频内容理解",
		ContextWindow: 256000, MaxOutput: 128000,
		Capabilities: []string{"text_generation", "multimodal", "deep_thinking", "streaming", "function_calling", "structured_output", "web_search"},
	},
	"mimo-v2-flash": {
		ID: "mimo-v2-flash", Description: "低成本、快速响应",
		ContextWindow: 256000, MaxOutput: 64000,
		Capabilities: []string{"text_generation", "deep_thinking", "streaming", "function_calling", "structured_output", "web_search"},
	},
	"mimo-v2.5-asr": {
		ID: "mimo-v2.5-asr", Description: "语音转文字（支持中英双语）",
		ContextWindow: 8000, MaxOutput: 2000,
		Capabilities: []string{"speech_recognition"},
	},
	"mimo-v2.5-tts": {
		ID: "mimo-v2.5-tts", Description: "文字转语音（标准预置音色）",
		ContextWindow: 8000, MaxOutput: 8000,
		Capabilities: []string{"speech_synthesis"},
	},
	"mimo-v2.5-tts-voiceclone": {
		ID: "mimo-v2.5-tts-voiceclone", Description: "声音克隆（上传音频样本）",
		ContextWindow: 8000, MaxOutput: 8000,
		Capabilities: []string{"speech_synthesis", "voice_clone"},
	},
	"mimo-v2.5-tts-voicedesign": {
		ID: "mimo-v2.5-tts-voicedesign", Description: "自定义音色设计",
		ContextWindow: 8000, MaxOutput: 8000,
		Capabilities: []string{"speech_synthesis", "voice_design"},
	},
	"mimo-v2-tts": {
		ID: "mimo-v2-tts", Description: "文字转语音",
		ContextWindow: 8000, MaxOutput: 8000,
		Capabilities: []string{"speech_synthesis"},
	},
}

// enrichModelInfo adds capabilities to model info if available
func enrichModelInfo(model ModelInfo) ModelInfo {
	if caps, ok := mimoModelCapabilities[model.ID]; ok {
		model.Description = caps.Description
		model.ContextWindow = caps.ContextWindow
		model.MaxOutput = caps.MaxOutput
		model.Capabilities = caps.Capabilities
	}
	return model
}

// Known compatible suffixes for API endpoints
var knownCompatSuffixes = []string{
	"/api/claudecode",
	"/api/anthropic",
	"/apps/anthropic",
	"/api/coding",
	"/claudecode",
	"/anthropic",
	"/step_plan",
	"/coding",
	"/claude",
}

// modelsEndpoint constructs the models endpoint URL from a base URL
func modelsEndpoint(baseURL string) string {
	cleaned := strings.TrimSuffix(baseURL, "/")

	// If already ends with /models, return as-is
	if strings.HasSuffix(cleaned, "/models") {
		return cleaned
	}

	// If ends with /v1, append /models
	if strings.HasSuffix(cleaned, "/v1") {
		return cleaned + "/models"
	}

	// Default: append /v1/models
	return cleaned + "/v1/models"
}

// buildModelsURLCandidates builds candidate URLs for the models endpoint
func buildModelsURLCandidates(baseURL string) []string {
	candidates := []string{
		modelsEndpoint(baseURL),
	}

	// Check for known compat suffixes and try alternatives
	baseURL = strings.TrimSuffix(baseURL, "/")
	for _, suffix := range knownCompatSuffixes {
		if strings.HasSuffix(baseURL, suffix) {
			withoutSuffix := strings.TrimSuffix(baseURL, suffix)
			candidates = append(candidates, withoutSuffix+"/v1/models")
			candidates = append(candidates, withoutSuffix+"/models")
			break
		}
	}

	return candidates
}

// ListModels fetches available models from the API
func (p *OpenAIProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	candidates := buildModelsURLCandidates(p.apiBase)

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no valid URL candidates")
	}

	var lastErr error

	for _, url := range candidates {
		models, err := p.fetchModelsFromURL(ctx, url)
		if err == nil {
			return models, nil
		}
		lastErr = err

		// If it's a 404/405, try next candidate
		if strings.Contains(err.Error(), "status 404") || strings.Contains(err.Error(), "status 405") {
			continue
		}
		// For other errors (auth, network), fail immediately
		return nil, err
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all endpoints failed: %w", lastErr)
	}
	return nil, fmt.Errorf("no models found")
}

// fetchModelsFromURL attempts to fetch models from a specific URL
func (p *OpenAIProvider) fetchModelsFromURL(ctx context.Context, url string) ([]ModelInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 || resp.StatusCode == 405 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, fmt.Errorf("authentication failed (status %d), please check your API key", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Try to parse response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse the payload using flexible parsing
	var payload interface{}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	modelIDs := parseModelPayload(payload)
	if len(modelIDs) == 0 {
		return nil, fmt.Errorf("no models found in response")
	}

	// Convert to ModelInfo and enrich with capabilities
	models := make([]ModelInfo, len(modelIDs))
	for i, id := range modelIDs {
		model := ModelInfo{ID: id}
		models[i] = enrichModelInfo(model)
	}

	return models, nil
}

// parseModelPayload extracts model IDs from various response formats
// Supports: string array, object array, data/models/items nested, single object
func parseModelPayload(payload interface{}) []string {
	var ids []string

	switch v := payload.(type) {
	case []interface{}:
		// Array at top level
		for _, item := range v {
			ids = append(ids, extractModelID(item)...)
		}
	case map[string]interface{}:
		// Object - try known keys
		for _, key := range []string{"data", "models", "items"} {
			if arr, ok := v[key].([]interface{}); ok {
				for _, item := range arr {
					ids = append(ids, extractModelID(item)...)
				}
				break
			}
		}
		// If no array found, try single object
		if len(ids) == 0 {
			ids = append(ids, extractModelID(v)...)
		}
	case string:
		// Single string
		ids = append(ids, v)
	}

	return deduplicateStrings(ids)
}

// extractModelID extracts model ID from an item (string or object)
func extractModelID(item interface{}) []string {
	var ids []string

	switch v := item.(type) {
	case string:
		// Direct string
		if v != "" {
			ids = append(ids, v)
		}
	case map[string]interface{}:
		// Object - try known keys in priority order
		for _, key := range []string{"id", "model", "name", "slug"} {
			if val, ok := v[key].(string); ok && val != "" {
				ids = append(ids, val)
				break
			}
		}
	}

	return ids
}

// deduplicateStrings removes duplicate strings
func deduplicateStrings(input []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, s := range input {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

func (p *OpenAIProvider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Accept", "text/event-stream")
}

// OpenAI API types (internal)

type openAIRequest struct {
	Model           string               `json:"model"`
	Messages        []openAIMessage      `json:"messages"`
	Tools           []openAITool         `json:"tools,omitempty"`
	MaxTokens       int                  `json:"max_tokens,omitempty"`
	Temperature     float64              `json:"temperature,omitempty"`
	TopP            float64              `json:"top_p,omitempty"`
	Stream          bool                 `json:"stream"`
	StreamOptions   *openAIStreamOptions `json:"stream_options,omitempty"`
	ReasoningEffort string               `json:"reasoning_effort,omitempty"`
}

type openAIStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    json.RawMessage  `json:"content"`
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
	Type     string             `json:"type"`
	Function openAIToolFunction `json:"function"`
}

type openAIToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

type openAIResponse struct {
	ID      string         `json:"id"`
	Choices []openAIChoice `json:"choices"`
	Usage   openAIUsage    `json:"usage"`
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
