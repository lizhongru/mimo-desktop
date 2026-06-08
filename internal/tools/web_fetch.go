package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// WebFetchTool fetches content from a URL
type WebFetchTool struct{}

func NewWebFetchTool() *WebFetchTool { return &WebFetchTool{} }

func (t *WebFetchTool) Name() string        { return "web_fetch" }
func (t *WebFetchTool) Description() string  { return "Fetch content from a URL" }
func (t *WebFetchTool) GetSafetyLevel() SafetyLevel { return SafetyLow }

func (t *WebFetchTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "URL to fetch",
			},
			"max_length": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum content length in characters (default: 5000)",
			},
		},
		"required": []string{"url"},
	}
}

func (t *WebFetchTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "url")
	return err
}

func (t *WebFetchTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *WebFetchTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	url, _ := StringParam(params, "url")

	maxLen := 5000
	if v, err := IntParam(params, "max_length"); err == nil && v > 0 {
		maxLen = v
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return ToolError("invalid URL: %v", err), nil
	}
	req.Header.Set("User-Agent", "MiMo-CLI/0.2.0")

	resp, err := client.Do(req)
	if err != nil {
		return ToolError("request failed: %v", err), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ToolError("HTTP %d: %s", resp.StatusCode, resp.Status), nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxLen*2)))
	if err != nil {
		return ToolError("failed to read response: %v", err), nil
	}

	content := string(body)
	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "html") {
		content = stripHTML(content)
	}

	if len(content) > maxLen {
		content = content[:maxLen] + "\n... (truncated)"
	}

	return &ToolResult{Output: fmt.Sprintf("URL: %s\nContent-Type: %s\n\n%s", url, ct, content)}, nil
}

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)
var htmlScriptRe = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
var htmlStyleRe = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)

func stripHTML(html string) string {
	html = htmlScriptRe.ReplaceAllString(html, "")
	html = htmlStyleRe.ReplaceAllString(html, "")
	html = htmlTagRe.ReplaceAllString(html, "")
	// Clean up whitespace
	lines := strings.Split(html, "\n")
	var cleaned []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			cleaned = append(cleaned, l)
		}
	}
	return strings.Join(cleaned, "\n")
}
