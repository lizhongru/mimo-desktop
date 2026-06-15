package desktop

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mimo-cli/mimo-cli/internal/actor"
	"github.com/mimo-cli/mimo-cli/internal/llm"
	"github.com/mimo-cli/mimo-cli/internal/tools"
)

const maxActorIterations = 10

// llmExecutor implements actor.Executor using a real LLM with tool-use loop.
type llmExecutor struct {
	gateway *llm.Gateway
	tools   *tools.Registry
}

func newLLMExecutor(gw *llm.Gateway, tr *tools.Registry) *llmExecutor {
	return &llmExecutor{gateway: gw, tools: tr}
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
			Stream:   false,
		}

		resp, err := e.gateway.Chat(ctx, req)
		if err != nil {
			return "", fmt.Errorf("LLM call failed: %w", err)
		}

		if len(resp.ToolCalls) == 0 {
			return resp.Content, nil
		}

		messages = append(messages, llm.Message{
			Role:      llm.RoleAssistant,
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		for _, tc := range resp.ToolCalls {
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
		return "You are a summarization agent. Condense the given content into a clear summary preserving key information, decisions, and action items."
	case actor.ActorCompaction:
		return "You are a context compaction agent. Compress the given conversation into a brief summary preserving essential context."
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