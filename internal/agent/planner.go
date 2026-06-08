package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/mimo-cli/mimo-cli/internal/llm"
)

// PlanningMode defines how the agent plans and executes tasks
type PlanningMode string

const (
	// ModeReact uses the standard ReAct loop (Think → Act → Observe)
	ModeReact PlanningMode = "react"
	// ModePlanExecute generates a plan first, then executes step by step
	ModePlanExecute PlanningMode = "plan-execute"
	// ModeAuto lets the agent decide which mode to use based on task complexity
	ModeAuto PlanningMode = "auto"
)

// PlanStepStatus represents the execution status of a plan step
type PlanStepStatus string

const (
	StepPending    PlanStepStatus = "pending"
	StepInProgress PlanStepStatus = "in_progress"
	StepCompleted  PlanStepStatus = "completed"
	StepFailed     PlanStepStatus = "failed"
	StepSkipped    PlanStepStatus = "skipped"
)

// PlanStep represents a single step in an execution plan
type PlanStep struct {
	ID          int            `json:"id"`
	Description string         `json:"description"`
	ToolName    string         `json:"tool_name,omitempty"`
	ToolArgs    string         `json:"tool_args,omitempty"`
	Status      PlanStepStatus `json:"status"`
	Result      string         `json:"result,omitempty"`
	Error       string         `json:"error,omitempty"`
}

// Plan represents a complete execution plan
type Plan struct {
	Goal        string     `json:"goal"`
	Steps       []PlanStep `json:"steps"`
	CurrentStep int        `json:"current_step"`
	TotalSteps  int        `json:"total_steps"`
}

// Planner handles task decomposition and plan generation
type Planner struct {
	gateway *llm.Gateway
	mu      sync.RWMutex
}

// NewPlanner creates a new Planner instance
func NewPlanner(gateway *llm.Gateway) *Planner {
	return &Planner{
		gateway: gateway,
	}
}

// GeneratePlan uses LLM to decompose a complex task into executable steps
func (p *Planner) GeneratePlan(ctx context.Context, task string, toolDefs []llm.ToolDefinition) (*Plan, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Build tool list for the prompt
	var toolList strings.Builder
	for _, td := range toolDefs {
		toolList.WriteString(fmt.Sprintf("- %s: %s\n", td.Function.Name, td.Function.Description))
	}

	planPrompt := fmt.Sprintf(`你是一个任务规划专家。用户会给你一个任务，你需要将其分解为可执行的步骤。

可用工具：
%s

任务：%s

请将这个任务分解为清晰的步骤。每个步骤应该：
1. 描述要做什么
2. 如果需要使用工具，指定工具名称和参数

请严格按照以下 JSON 格式输出，不要输出其他内容：
{
  "goal": "任务目标的简短描述",
  "steps": [
    {
      "id": 1,
      "description": "步骤描述",
      "tool_name": "工具名称（如果需要）",
      "tool_args": "JSON格式的工具参数（如果需要）"
    }
  ]
}

注意：
- 步骤数量控制在 3-10 个
- 每个步骤应该是原子性的
- 步骤之间可以有依赖关系
- 如果任务简单到不需要工具，直接描述即可
- tool_name 和 tool_args 可以为空字符串`, toolList.String(), task)

	req := llm.ChatRequest{
		Messages: []llm.Message{
			{
				Role:    llm.RoleSystem,
				Content: "你是一个任务规划专家，擅长将复杂任务分解为可执行的步骤。请始终使用中文回复。",
			},
			{
				Role:    llm.RoleUser,
				Content: planPrompt,
			},
		},
		Stream: false,
	}

	resp, err := p.gateway.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %w", err)
	}

	// Parse the plan from JSON
	plan, err := parsePlanFromResponse(resp.Content)
	if err != nil {
		// Fallback: create a simple single-step plan
		return &Plan{
			Goal: task,
			Steps: []PlanStep{
				{
					ID:          1,
					Description: task,
					Status:      StepPending,
				},
			},
			CurrentStep: 0,
			TotalSteps:  1,
		}, nil
	}

	return plan, nil
}

// parsePlanFromResponse parses the LLM response into a Plan struct
func parsePlanFromResponse(content string) (*Plan, error) {
	// Try to extract JSON from the response
	content = strings.TrimSpace(content)

	// Find JSON block
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no valid JSON found in response")
	}

	jsonStr := content[start : end+1]

	var plan Plan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
	}

	// Validate and set defaults
	if len(plan.Steps) == 0 {
		return nil, fmt.Errorf("plan has no steps")
	}

	plan.TotalSteps = len(plan.Steps)
	plan.CurrentStep = 0

	// Set initial status for all steps
	for i := range plan.Steps {
		plan.Steps[i].Status = StepPending
	}

	return &plan, nil
}

// FormatPlan returns a human-readable string representation of the plan
func (p *Plan) FormatPlan() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("📋 执行计划: %s\n", p.Goal))
	sb.WriteString(fmt.Sprintf("   共 %d 个步骤\n\n", p.TotalSteps))

	for _, step := range p.Steps {
		var statusIcon string
		switch step.Status {
		case StepPending:
			statusIcon = "○"
		case StepInProgress:
			statusIcon = "◉"
		case StepCompleted:
			statusIcon = "✓"
		case StepFailed:
			statusIcon = "✗"
		case StepSkipped:
			statusIcon = "⊘"
		}

		sb.WriteString(fmt.Sprintf("  %s %d. %s", statusIcon, step.ID, step.Description))
		if step.ToolName != "" {
			sb.WriteString(fmt.Sprintf(" [%s]", step.ToolName))
		}
		sb.WriteString("\n")

		// Show result/error for completed/failed steps
		if step.Status == StepCompleted && step.Result != "" {
			result := step.Result
			if len(result) > 100 {
				result = result[:100] + "..."
			}
			sb.WriteString(fmt.Sprintf("     └─ %s\n", result))
		} else if step.Status == StepFailed && step.Error != "" {
			sb.WriteString(fmt.Sprintf("     └─ 错误: %s\n", step.Error))
		}
	}

	return sb.String()
}

// FormatProgress returns a compact progress representation
func (p *Plan) FormatProgress() string {
	completed := 0
	for _, step := range p.Steps {
		if step.Status == StepCompleted {
			completed++
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("进度: %d/%d ", completed, p.TotalSteps))

	// Progress bar
	barWidth := 20
	filled := completed * barWidth / p.TotalSteps
	sb.WriteString("[")
	sb.WriteString(strings.Repeat("█", filled))
	sb.WriteString(strings.Repeat("░", barWidth-filled))
	sb.WriteString("]")

	return sb.String()
}

// GetCurrentStep returns the current pending/in-progress step
func (p *Plan) GetCurrentStep() *PlanStep {
	for i := range p.Steps {
		if p.Steps[i].Status == StepPending || p.Steps[i].Status == StepInProgress {
			return &p.Steps[i]
		}
	}
	return nil
}

// MarkStepCompleted marks a step as completed with its result
func (p *Plan) MarkStepCompleted(stepID int, result string) {
	for i := range p.Steps {
		if p.Steps[i].ID == stepID {
			p.Steps[i].Status = StepCompleted
			p.Steps[i].Result = result
			p.CurrentStep = i + 1
			break
		}
	}
}

// MarkStepFailed marks a step as failed with the error
func (p *Plan) MarkStepFailed(stepID int, err string) {
	for i := range p.Steps {
		if p.Steps[i].ID == stepID {
			p.Steps[i].Status = StepFailed
			p.Steps[i].Error = err
			break
		}
	}
}

// IsComplete returns true if all steps are completed or skipped
func (p *Plan) IsComplete() bool {
	for _, step := range p.Steps {
		if step.Status != StepCompleted && step.Status != StepSkipped {
			return false
		}
	}
	return true
}

// HasFailures returns true if any step has failed
func (p *Plan) HasFailures() bool {
	for _, step := range p.Steps {
		if step.Status == StepFailed {
			return true
		}
	}
	return false
}
