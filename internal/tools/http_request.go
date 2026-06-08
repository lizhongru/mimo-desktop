package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPRequestTool sends HTTP requests
type HTTPRequestTool struct{}

func NewHTTPRequestTool() *HTTPRequestTool { return &HTTPRequestTool{} }

func (t *HTTPRequestTool) Name() string        { return "http_request" }
func (t *HTTPRequestTool) Description() string  { return "Send HTTP requests (GET, POST, PUT, DELETE)" }
func (t *HTTPRequestTool) GetSafetyLevel() SafetyLevel { return SafetyMedium }
func (t *HTTPRequestTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "Request URL",
			},
			"method": map[string]interface{}{
				"type":        "string",
				"description": "HTTP method (default: GET)",
				"enum":        []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
			},
			"body": map[string]interface{}{
				"type":        "string",
				"description": "Request body (for POST/PUT/PATCH)",
			},
			"headers": map[string]interface{}{
				"type":        "object",
				"description": "Request headers as key-value pairs",
			},
		},
		"required": []string{"url"},
	}
}
func (t *HTTPRequestTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "url")
	return err
}
func (t *HTTPRequestTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *HTTPRequestTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	url, _ := StringParam(params, "url")
	method := OptionalStringParam(params, "method", "GET")
	body, _ := StringParam(params, "body")

	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return ToolError("failed to create request: %v", err), nil
	}

	// Set headers
	if headers, ok := params["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if s, ok := v.(string); ok {
				req.Header.Set(k, s)
			}
		}
	}
	if body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ToolError("request failed: %v", err), nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 10000))
	if err != nil {
		return ToolError("failed to read response: %v", err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Status: %d %s\n", resp.StatusCode, resp.Status))
	sb.WriteString("Headers:\n")
	for k, v := range resp.Header {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", k, strings.Join(v, ", ")))
	}
	sb.WriteString("\nBody:\n")
	sb.WriteString(string(respBody))

	result := sb.String()
	if len(result) > 6000 {
		result = result[:6000] + "\n... (truncated)"
	}
	return &ToolResult{Output: result}, nil
}
