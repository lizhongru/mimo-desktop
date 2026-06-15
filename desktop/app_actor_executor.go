package desktop

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mimo-cli/mimo-cli/internal/actor"
	"github.com/mimo-cli/mimo-cli/internal/llm"
	"github.com/mimo-cli/mimo-cli/internal/tools"
)

const maxActorIterations = 10

// llmExecutor implements actor.Executor using a real LLM with tool-use loop.
type llmExecutor struct {
	gateway  *llm.Gateway
	tools    *tools.Registry
	streamCB actor.StreamCallback
}

func newLLMExecutor(gw *llm.Gateway, tr *tools.Registry) *llmExecutor {
	return &llmExecutor{gateway: gw, tools: tr}
}

func (e *llmExecutor) SetStreamCallback(cb actor.StreamCallback) {
	e.streamCB = cb
}

func (e *llmExecutor) ExecuteActor(ctx context.Context, act *actor.Actor) (string, error) {
	sysPrompt := buildActorSystemPrompt(act.Type)
	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: sysPrompt},
		{Role: llm.RoleUser, Content: act.Prompt},
	}

	// Build tool definitions
	var toolDefs []llm.ToolDefinition
	if e.tools != nil {
		for _, def := range e.tools.Definitions() {
			name, _ := def["name"].(string)
			desc, _ := def["description"].(string)
			toolDefs = append(toolDefs, llm.ToolDefinition{
				Type: "function",
				Function: llm.ToolFuncDefinition{
					Name:        name,
					Description: desc,
					Parameters:  def["parameters"],
				},
			})
		}
	}

	for i := 0; i < maxActorIterations; i++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		req := llm.ChatRequest{
			Messages: messages,
			Tools:    toolDefs,
			Stream:   true,
		}

		ch, err := e.gateway.ChatStream(ctx, req)
		if err != nil {
			return "", fmt.Errorf("LLM call failed: %w", err)
		}

		var fullContent strings.Builder
		var toolCalls []llm.ToolCall
		toolCallIndexMap := make(map[int]int)

		for chunk := range ch {
			if chunk.Error != nil {
				return "", chunk.Error
			}

			if chunk.Delta != "" {
				fullContent.WriteString(chunk.Delta)
				if e.streamCB != nil {
					e.streamCB(act.ID, chunk.Delta, false)
				}
			}

			if len(chunk.ToolCalls) > 0 {
				for _, tc := range chunk.ToolCalls {
					if tc.ID != "" {
						idx := len(toolCalls)
						toolCalls = append(toolCalls, tc)
						toolCallIndexMap[chunk.ToolCallIndex] = idx
						// Notify frontend about tool call
						if e.streamCB != nil {
							e.streamCB(act.ID, fmt.Sprintf("\n🔧 %s", tc.Function.Name), true)
						}
					} else if tc.Function.Arguments != "" {
						if pos, ok := toolCallIndexMap[chunk.ToolCallIndex]; ok {
							toolCalls[pos].Function.Arguments += tc.Function.Arguments
						}
					}
				}
			}
		}

		content := fullContent.String()

		// No tool calls → final response
		if len(toolCalls) == 0 {
			return content, nil
		}

		// Add assistant message
		messages = append(messages, llm.Message{
			Role:      llm.RoleAssistant,
			Content:   content,
			ToolCalls: toolCalls,
		})

		// Execute tool calls
		for _, tc := range toolCalls {
			var params map[string]interface{}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &params); err != nil {
				params = map[string]interface{}{}
			}

			var toolResult string
			if e.tools != nil {
				res, execErr := e.tools.Execute(ctx, tc.Function.Name, params)
				if execErr != nil {
					toolResult = fmt.Sprintf("Error: %v", execErr)
				} else if res != nil {
					toolResult = res.Output
					if res.Error != "" {
						toolResult = "Error: " + res.Error
					}
				}
			} else {
				toolResult = fmt.Sprintf("Tool %s not available", tc.Function.Name)
			}

			// Notify tool result
			if e.streamCB != nil {
				preview := toolResult
				if len(preview) > 100 {
					preview = preview[:100] + "..."
				}
				e.streamCB(act.ID, fmt.Sprintf("  └─ %s", preview), true)
			}

			messages = append(messages, llm.Message{
				Role:       llm.RoleTool,
				Content:    toolResult,
				ToolCallID: tc.ID,
				Name:       tc.Function.Name,
			})
		}
	}

	return "", fmt.Errorf("reached max iterations (%d)", maxActorIterations)
}

func buildActorSystemPrompt(t actor.ActorType) string {
	switch t {
	case actor.ActorExplore:
		return "You are an exploration agent. Investigate code, files, or systems and report findings concisely. Use file reading and search tools when needed."
	case actor.ActorGeneral:
		return "You are a general-purpose agent. Complete the given task using available tools. Report results clearly."
	case actor.ActorTitle:
		return "You are a title generation agent. Given content, produce a concise descriptive title. Return ONLY the title, under 60 characters."
	case actor.ActorSummary:
		return "You are a summarization agent. Condense the given content into a clear summary preserving essential context."
	case actor.ActorCompaction:
		return "You are a context compaction agent. Compress the current conversation into a brief summary preserving essential context."
	case actor.ActorDream:
		return "You are a reflection agent. Analyze the session and extract reusable patterns, insights, or skills."
	case actor.ActorDistill:
		return "You are a distillation agent. Extract concrete skills from candidates and format them for storage."
	case actor.ActorCheckpoint:
		return "You are a checkpoint agent. Summarize the current session state for persistence."
	default:
		return "You are a helpful assistant agent. Complete the given task."
	}
}
