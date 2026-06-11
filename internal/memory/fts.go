package memory

import (
	"strings"
	"unicode"
)

// BuildFtsQuery converts a user query into an FTS5 query string
// Punctuation becomes separators, each alphanumeric run becomes a phrase-quoted literal
// OR-joined for broad matching with BM25 ranking
func BuildFtsQuery(query string) string {
	tokens := tokenize(query)
	if len(tokens) == 0 {
		return ""
	}

	quoted := make([]string, len(tokens))
	for i, t := range tokens {
		quoted[i] = `"` + t + `"`
	}
	return strings.Join(quoted, " OR ")
}

// tokenize splits a query into alphanumeric tokens
func tokenize(query string) []string {
	var tokens []string
	var current strings.Builder

	for _, r := range query {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				tokens = append(tokens, strings.ToLower(current.String()))
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, strings.ToLower(current.String()))
	}

	return tokens
}

// ExtractSnippet extracts a snippet from text around the first match
func ExtractSnippet(text string, query string, contextLen int) string {
	if contextLen <= 0 {
		contextLen = 64
	}

	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)

	idx := strings.Index(lowerText, lowerQuery)
	if idx == -1 {
		// Try to find any token from the query
		tokens := tokenize(query)
		for _, token := range tokens {
			idx = strings.Index(lowerText, token)
			if idx != -1 {
				break
			}
		}
	}

	if idx == -1 {
		// No match found, return beginning of text
		if len(text) > contextLen*2 {
			return text[:contextLen*2] + "..."
		}
		return text
	}

	start := idx - contextLen
	if start < 0 {
		start = 0
	}
	end := idx + len(query) + contextLen
	if end > len(text) {
		end = len(text)
	}

	snippet := text[start:end]
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(text) {
		snippet = snippet + "..."
	}

	return snippet
}
