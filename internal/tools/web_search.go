package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// WebSearchTool searches the web (via Bing)
type WebSearchTool struct{}

func NewWebSearchTool() *WebSearchTool { return &WebSearchTool{} }

func (t *WebSearchTool) Name() string        { return "web_search" }
func (t *WebSearchTool) Description() string  { return "Search the web using Bing" }
func (t *WebSearchTool) GetSafetyLevel() SafetyLevel { return SafetyLow }
func (t *WebSearchTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query",
			},
			"count": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results (default: 5)",
			},
		},
		"required": []string{"query"},
	}
}
func (t *WebSearchTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "query")
	return err
}
func (t *WebSearchTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *WebSearchTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	query, _ := StringParam(params, "query")
	count := 5
	if c, err := IntParam(params, "count"); err == nil && c > 0 {
		count = c
	}

	searchURL := "https://www.bing.com/search?q=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return ToolError("failed to create request: %v", err), nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ToolError("search failed: %v", err), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ToolError("search returned status %d", resp.StatusCode), nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 200000))
	if err != nil {
		return ToolError("failed to read response: %v", err), nil
	}

	html := string(body)
	results := parseBingResults(html, count)
	if len(results) == 0 {
		return &ToolResult{Output: "No results found for: " + query}, nil
	}

	var sb strings.Builder
	sb.WriteString("Search results for: " + query + "\n\n")
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("%d. %s\n   %s\n", i+1, r.title, r.url))
		if r.snippet != "" {
			sb.WriteString("   " + r.snippet + "\n")
		}
		sb.WriteString("\n")
	}

	return &ToolResult{Output: sb.String()}, nil
}

type searchResult struct {
	title   string
	url     string
	snippet string
}

func parseBingResults(html string, maxCount int) []searchResult {
	var results []searchResult

	// Bing: <li class="b_algo"><h2><a href="url">title</a></h2><p>snippet</p>
	blocks := strings.Split(html, "b_algo")
	for i := 1; i < len(blocks) && len(results) < maxCount; i++ {
		block := blocks[i]

		// Extract URL
		hrefIdx := strings.Index(block, "href=")
		if hrefIdx == -1 {
			continue
		}
		hrefIdx += 6
		hrefEnd := strings.Index(block[hrefIdx:], "\"")
		if hrefEnd == -1 {
			continue
		}
		resultURL := block[hrefIdx : hrefIdx+hrefEnd]

		// Extract title
		tagEnd := strings.Index(block, ">")
		if tagEnd == -1 {
			continue
		}
		titleContent := block[tagEnd+1:]
		titleEnd := strings.Index(titleContent, "</a>")
		if titleEnd == -1 {
			continue
		}
		title := stripHTMLTags(titleContent[:titleEnd])

		// Extract snippet
		snippet := ""
		pIdx := strings.Index(block, "<p")
		if pIdx != -1 {
			pBlock := block[pIdx:]
			pTagEnd := strings.Index(pBlock, ">")
			if pTagEnd != -1 {
				pContent := pBlock[pTagEnd+1:]
				pEnd := strings.Index(pContent, "</p>")
				if pEnd == -1 {
					pEnd = 200
				}
				if pEnd > len(pContent) {
					pEnd = len(pContent)
				}
				snippet = stripHTMLTags(pContent[:pEnd])
			}
		}

		if title != "" && strings.HasPrefix(resultURL, "http") {
			results = append(results, searchResult{title: title, url: resultURL, snippet: snippet})
		}
	}

	return results
}

func stripHTMLTags(s string) string {
	var sb strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			sb.WriteRune(r)
		}
	}
	return strings.TrimSpace(sb.String())
}
