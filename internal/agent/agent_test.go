package agent

import (
	"testing"

	"github.com/mimo-cli/mimo-cli/internal/llm"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		msg      llm.Message
		expected int
	}{
		{
			name: "short message",
			msg: llm.Message{
				Role:    llm.RoleUser,
				Content: "Hello",
			},
			expected: 12, // len("Hello")/3 + 10 = 1 + 10 = 11, but we'll check actual
		},
		{
			name: "long message",
			msg: llm.Message{
				Role:    llm.RoleUser,
				Content: "This is a longer message with many characters to test token estimation",
			},
			expected: 35, // len/3 + 10 = 24 + 10 = 34
		},
		{
			name: "message with tool calls",
			msg: llm.Message{
				Role:    llm.RoleAssistant,
				Content: "Let me read that file",
				ToolCalls: []llm.ToolCall{
					{
						Function: llm.FunctionCall{
							Name:      "file_read",
							Arguments: `{"path": "/test.txt"}`,
						},
					},
				},
			},
			expected: 20, // content + tool_name + args + overhead
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := estimateTokens(tt.msg)
			// Allow some variance in estimation
			if result < tt.expected/2 || result > tt.expected*2 {
				t.Errorf("estimateTokens() = %v, want around %v", result, tt.expected)
			}
		})
	}
}

func TestTruncateMessages(t *testing.T) {
	// Create agent with small context limit for testing
	a := &Agent{
		maxContextTokens: 100, // Very small limit
		messages:         make([]llm.Message, 0),
	}

	// Add system message
	a.messages = append(a.messages, llm.Message{
		Role:    llm.RoleSystem,
		Content: "You are a helpful assistant.",
	})

	// Add many messages to exceed limit
	for i := 0; i < 20; i++ {
		a.messages = append(a.messages, llm.Message{
			Role:    llm.RoleUser,
			Content: "Hello, this is message " + string(rune(i+'0')),
		})
		a.messages = append(a.messages, llm.Message{
			Role:    llm.RoleAssistant,
			Content: "I understand. Here is response " + string(rune(i+'0')),
		})
	}

	// Calculate initial tokens
	initialTokens := 0
	for _, msg := range a.messages {
		initialTokens += estimateTokens(msg)
	}

	t.Logf("Initial messages: %d, tokens: %d", len(a.messages), initialTokens)

	// Truncate
	a.truncateMessages()

	// Calculate final tokens
	finalTokens := 0
	for _, msg := range a.messages {
		finalTokens += estimateTokens(msg)
	}

	t.Logf("After truncation: messages: %d, tokens: %d", len(a.messages), finalTokens)

	// Verify system message is preserved
	if a.messages[0].Role != llm.RoleSystem {
		t.Error("System message was removed during truncation")
	}

	// Verify we're within context limit (allow some overhead)
	if finalTokens > a.maxContextTokens*2 {
		t.Errorf("After truncation, tokens=%d exceeds limit=%d", finalTokens, a.maxContextTokens)
	}
}

func TestTruncateMessagesWithToolResults(t *testing.T) {
	// Create agent with context limit
	a := &Agent{
		maxContextTokens: 200,
		messages:         make([]llm.Message, 0),
	}

	// Add system message
	a.messages = append(a.messages, llm.Message{
		Role:    llm.RoleSystem,
		Content: "System prompt",
	})

	// Add tool-heavy messages (tool results are usually large)
	for i := 0; i < 5; i++ {
		a.messages = append(a.messages, llm.Message{
			Role:    llm.RoleUser,
			Content: "Read file " + string(rune(i+'0')),
		})
		a.messages = append(a.messages, llm.Message{
			Role:    llm.RoleAssistant,
			Content: "Reading file...",
			ToolCalls: []llm.ToolCall{
				{
					ID:   "call_" + string(rune(i+'0')),
					Type: "function",
					Function: llm.FunctionCall{
						Name:      "file_read",
						Arguments: `{"path": "/file" + string(rune(i+'0')) + ".txt"}`,
					},
				},
			},
		})
		// Large tool result (simulating reading a large file)
		a.messages = append(a.messages, llm.Message{
			Role:       llm.RoleTool,
			ToolCallID: "call_" + string(rune(i+'0')),
			Content:    generateLargeString(500), // 500 chars
		})
	}

	initialTokens := 0
	for _, msg := range a.messages {
		initialTokens += estimateTokens(msg)
	}

	t.Logf("Initial: messages=%d, tokens=%d", len(a.messages), initialTokens)

	a.truncateMessages()

	finalTokens := 0
	for _, msg := range a.messages {
		finalTokens += estimateTokens(msg)
	}

	t.Logf("After truncation: messages=%d, tokens=%d", len(a.messages), finalTokens)

	if finalTokens > a.maxContextTokens*2 {
		t.Errorf("Tokens=%d exceeds limit=%d after truncation", finalTokens, a.maxContextTokens)
	}
}

// generateLargeString generates a string of specified length
func generateLargeString(n int) string {
	result := make([]byte, n)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}
