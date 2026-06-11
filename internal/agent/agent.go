package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/mimo-cli/mimo-cli/internal/llm"
	"github.com/mimo-cli/mimo-cli/internal/safety"
	"github.com/mimo-cli/mimo-cli/internal/tools"
)

// Agent is the core AI agent that orchestrates LLM + Tools
type Agent struct {
	gateway   *llm.Gateway
	registry  *tools.Registry
	guardrail *safety.Guardrail
	planner   *Planner
	messages  []llm.Message
	mu        sync.Mutex

	// Cancellation support
	cancelFunc context.CancelFunc

	// Callbacks for UI updates
	onStreamDelta    func(delta string)
	onStreamThinking func(thinking string)
	onToolCall       func(name string, args string)
	onToolResult     func(name string, result string)
	onError          func(err error)
	onUsage          func(usage llm.Usage)
	onCompressing    func() // 上下文压缩开始时回调
	onPlanning       func(message string) // 正在生成计划回调
	onPlanGenerated  func(plan *Plan) // 计划生成完成回调
	onPlanStepStart  func(step *PlanStep) // 计划步骤开始执行回调
	onPlanStepDone   func(step *PlanStep) // 计划步骤完成回调

	// Configuration
	maxIterations    int
	maxContextTokens int
	maxParallelTools int // 最大并行工具数
	planningMode     PlanningMode // 规划模式
	reasoningLevel   string       // reasoning effort level: low, medium, high
	verbose          bool
	confirmAll       bool // 是否确认所有操作
}

// NewAgent creates a new Agent
func NewAgent(gateway *llm.Gateway, registry *tools.Registry, guardrail *safety.Guardrail, maxContextTokens, maxIterations int) *Agent {
	if maxContextTokens <= 0 {
		maxContextTokens = 128000 // default
	}
	return &Agent{
		gateway:         gateway,
		registry:        registry,
		guardrail:       guardrail,
		planner:         NewPlanner(gateway),
		maxIterations:   maxIterations,
		maxContextTokens: maxContextTokens,
		maxParallelTools: 5,
		planningMode:    ModeAuto,
	}
}

// SetStreamDeltaCallback sets the callback for streaming delta updates
func (a *Agent) SetStreamDeltaCallback(fn func(string)) {
	a.onStreamDelta = fn
}

// SetStreamThinkingCallback sets the callback for streaming thinking updates
func (a *Agent) SetStreamThinkingCallback(fn func(string)) {
	a.onStreamThinking = fn
}

// SetToolCallCallback sets the callback for tool call notifications
func (a *Agent) SetToolCallCallback(fn func(string, string)) {
	a.onToolCall = fn
}

// SetToolResultCallback sets the callback for tool result notifications
func (a *Agent) SetToolResultCallback(fn func(string, string)) {
	a.onToolResult = fn
}

// SetErrorCallback sets the callback for error notifications
func (a *Agent) SetErrorCallback(fn func(error)) {
	a.onError = fn
}

// SetUsageCallback sets the callback for token usage updates
func (a *Agent) SetUsageCallback(fn func(llm.Usage)) {
	a.onUsage = fn
}

// SetCompressingCallback sets the callback for context compression notifications
func (a *Agent) SetCompressingCallback(fn func()) {
	a.onCompressing = fn
}

// SetPlanningCallback sets the callback for planning status notifications
func (a *Agent) SetPlanningCallback(fn func(string)) {
	a.onPlanning = fn
}

// SetPlanGeneratedCallback sets the callback for plan generation
func (a *Agent) SetPlanGeneratedCallback(fn func(*Plan)) {
	a.onPlanGenerated = fn
}

// SetPlanStepStartCallback sets the callback for plan step start
func (a *Agent) SetPlanStepStartCallback(fn func(*PlanStep)) {
	a.onPlanStepStart = fn
}

// SetPlanStepDoneCallback sets the callback for plan step completion
func (a *Agent) SetPlanStepDoneCallback(fn func(*PlanStep)) {
	a.onPlanStepDone = fn
}

// SetPlanningMode sets the planning mode
func (a *Agent) SetPlanningMode(mode PlanningMode) {
	a.planningMode = mode
}

// SetReasoningLevel sets the reasoning effort level (low, medium, high)
func (a *Agent) SetReasoningLevel(level string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.reasoningLevel = level
}

// CompressContext manually triggers context compression.
// Returns token counts before and after compression.
func (a *Agent) CompressContext(ctx context.Context) (before int, after int, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for _, msg := range a.messages {
		before += estimateTokens(msg)
	}

	a.compressContextForce(ctx)

	for _, msg := range a.messages {
		after += estimateTokens(msg)
	}

	return before, after, nil
}

// GetMessageCount returns the number of messages in the conversation
func (a *Agent) LoadMessages(msgs []llm.Message) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.messages = msgs
}

func (a *Agent) GetMessageCount() int {

	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.messages)
}

// Cancel cancels the current operation
func (a *Agent) Cancel() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancelFunc != nil {
		a.cancelFunc()
		a.cancelFunc = nil
	}
}

// SetConfirmAll sets the confirm-all flag
func (a *Agent) SetConfirmAll(confirmAll bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.confirmAll = confirmAll
}

// SystemPrompt sets the system prompt for the agent
func (a *Agent) SystemPrompt(prompt string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Replace or add system message
	for i, msg := range a.messages {
		if msg.Role == llm.RoleSystem {
			a.messages[i].Content = prompt
			return
		}
	}
	// Prepend system message
	a.messages = append([]llm.Message{{
		Role:    llm.RoleSystem,
		Content: prompt,
	}}, a.messages...)
}

// Chat sends a message and runs the agent loop
func (a *Agent) Chat(ctx context.Context, userMessage string, attachments []llm.Attachment) (string, error) {
	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	a.mu.Lock()
	a.cancelFunc = cancel
	a.messages = append(a.messages, llm.Message{
		Role:        llm.RoleUser,
		Content:     userMessage,
		Attachments: attachments,
	})
	a.compressContext(ctx)
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		a.cancelFunc = nil
		a.mu.Unlock()
		cancel()
	}()

	return a.runLoop(ctx)
}

// ChatStream sends a message and streams the response
func (a *Agent) ChatStream(ctx context.Context, userMessage string, attachments []llm.Attachment) (string, error) {
	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	a.mu.Lock()
	a.cancelFunc = cancel
	a.messages = append(a.messages, llm.Message{
		Role:        llm.RoleUser,
		Content:     userMessage,
		Attachments: attachments,
	})
	a.compressContext(ctx)
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		a.cancelFunc = nil
		a.mu.Unlock()
		cancel()
	}()

	// Check planning mode
	if a.planningMode == ModePlanExecute || (a.planningMode == ModeAuto && a.isComplexTask(userMessage)) {
		// Use plan-execute mode for complex tasks
		response, err := a.runPlanExecuteLoop(ctx, userMessage)
		if err == nil {
			return response, nil
		}
		// Fallback to ReAct if planning fails
		if a.onError != nil {
			a.onError(fmt.Errorf("planning failed, falling back to ReAct: %w", err))
		}
	}

	return a.runStreamLoop(ctx)
}


// isComplexTask determines if a task needs planning based on keywords and structure
func (a *Agent) isComplexTask(task string) bool {
	task = strings.ToLower(strings.TrimSpace(task))
	
	// Simple patterns that don't need planning (single-step tasks)
	simplePatterns := []string{
		// Greetings
		"你好", "hi", "hello", "嗨", "hey",
		// Simple questions
		"什么是", "解释", "how", "what", "why",
		// Simple commands
		"帮助", "help", "谢谢", "thanks",
		// Single file operations (NOT complex)
		"创建一个", "新建一个", "写一个", "生成一个", "写入一个",
		"create a", "write a", "generate a",
	}
	
	for _, pattern := range simplePatterns {
		if strings.HasPrefix(task, pattern) {
			return false
		}
	}
	
	// Complex task indicators (require multiple steps)
	complexIndicators := []string{
		// Project-level tasks
		"项目", "project", "系统", "system", "平台", "platform",
		// Multi-file refactoring
		"重构所有", "优化整个", "整理代码", "refactor all", "optimize entire",
		// Build/deploy with multiple components
		"搭建环境", "部署服务", "setup environment", "deploy service",
	}
	
	// Check for complex task keywords
	for _, indicator := range complexIndicators {
		if strings.Contains(task, indicator) {
			return true
		}
	}
	
	// Check task length (very long tasks are usually complex)
	if len(task) > 200 {
		return true
	}
	
	// Check for multiple explicit requests (3+ sentences)
	if strings.Count(task, "。") >= 2 || strings.Count(task, "\n") >= 3 {
		return true
	}
	
	return false
}
// runPlanExecuteLoop runs the Plan-Execute loop
func (a *Agent) runPlanExecuteLoop(ctx context.Context, task string) (string, error) {
	// 发送规划状态
	if a.onPlanning != nil {
		a.onPlanning("正在分析任务，生成执行计划...")
	}

	// Generate plan
	toolDefs := a.toolDefinitions()
	plan, err := a.planner.GeneratePlan(ctx, task, toolDefs)
	if err != nil {
		return "", fmt.Errorf("failed to generate plan: %w", err)
	}

	// Notify UI about the plan
	if a.onPlanGenerated != nil {
		a.onPlanGenerated(plan)
	}

	// Execute each step
	for i := range plan.Steps {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		step := &plan.Steps[i]
		step.Status = StepInProgress

		// 发送规划状态
		if a.onPlanning != nil {
			a.onPlanning(fmt.Sprintf("执行步骤 %d/%d: %s", i+1, len(plan.Steps), step.Description))
		}

		// Notify UI about step start
		if a.onPlanStepStart != nil {
			a.onPlanStepStart(step)
		}

		// Execute the step
		var result string
		var stepErr error

		if step.ToolName != "" {
			// Execute tool call
			params := make(map[string]interface{})
			if step.ToolArgs != "" && step.ToolArgs != "{}" {
				if err := json.Unmarshal([]byte(step.ToolArgs), &params); err != nil {
					stepErr = fmt.Errorf("failed to parse tool args: %w", err)
				}
			}

			if stepErr == nil {
				if a.onToolCall != nil {
					a.onToolCall(step.ToolName, step.ToolArgs)
				}

				// Safety check
				if a.guardrail != nil {
					allowed, err := a.guardrail.CheckWithConfirmAll(step.ToolName, params, a.confirmAll)
					if !allowed {
						stepErr = fmt.Errorf("blocked by safety: %w", err)
					}
				}

				if stepErr == nil {
					toolResult, err := a.registry.Execute(ctx, step.ToolName, params)
					if err != nil {
						stepErr = err
					} else {
						result = toolResult.Output
						if toolResult.Error != "" {
							result = fmt.Sprintf("Error: %s\nOutput: %s", toolResult.Error, toolResult.Output)
						}
					}
				}
			}
		} else {
			// Use LLM to describe what should be done
			req := llm.ChatRequest{
				ReasoningEffort: a.reasoningLevel,
				Messages: []llm.Message{
					{Role: llm.RoleSystem, Content: "你是一个任务执行助手。请执行以下步骤并描述结果。"},
					{Role: llm.RoleUser, Content: fmt.Sprintf("任务: %s\n当前步骤: %s", task, step.Description)},
				},
				Stream: false,
			}

			resp, err := a.gateway.Chat(ctx, req)
			if err != nil {
				stepErr = err
			} else {
				result = resp.Content
			}
		}

		// Update step status
		if stepErr != nil {
			step.Status = StepFailed
			step.Error = stepErr.Error()
			if a.onToolResult != nil {
				a.onToolResult(step.ToolName, "Error: "+stepErr.Error())
			}
		} else {
			step.Status = StepCompleted
			step.Result = result
			if a.onToolResult != nil {
				a.onToolResult(step.ToolName, result)
			}
		}

		// Notify UI about step completion
		if a.onPlanStepDone != nil {
			a.onPlanStepDone(step)
		}

		// Stop if a step failed
		if stepErr != nil {
			return "", fmt.Errorf("step %d failed: %w", step.ID, stepErr)
		}
	}

	// 将执行结果添加到消息历史
	a.mu.Lock()
	// 添加计划执行的摘要消息
	planSummary := fmt.Sprintf("执行计划完成:\n%s", plan.FormatPlan())
	a.messages = append(a.messages, llm.Message{
		Role:    llm.RoleAssistant,
		Content: planSummary,
	})
	a.mu.Unlock()

	// Generate final summary
	summaryPrompt := fmt.Sprintf("任务完成。执行计划:\n%s\n\n请总结执行结果。", plan.FormatPlan())
	req := llm.ChatRequest{
		ReasoningEffort: a.reasoningLevel,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: "请用中文简洁地总结任务执行结果。"},
			{Role: llm.RoleUser, Content: summaryPrompt},
		},
		Stream: true,
	}

	// Stream the summary
	ch, err := a.gateway.ChatStream(ctx, req)
	if err != nil {
		return plan.FormatPlan(), nil
	}

	var summary strings.Builder
	for chunk := range ch {
		if chunk.Error != nil {
			continue
		}
		if chunk.Delta != "" {
			summary.WriteString(chunk.Delta)
			if a.onStreamDelta != nil {
				a.onStreamDelta(chunk.Delta)
			}
		}
	}

	return summary.String(), nil
}

// runLoop runs the ReAct loop (non-streaming)
func (a *Agent) runLoop(ctx context.Context) (string, error) {
	var lastContent string

	for i := 0; i < a.maxIterations; i++ {
		// Call LLM
		req := llm.ChatRequest{
			Messages:        a.messages,
			Tools:           a.toolDefinitions(),
			Stream:          false,
			ReasoningEffort: a.reasoningLevel,
		}

		resp, err := a.gateway.Chat(ctx, req)
		if err != nil {
			return "", fmt.Errorf("LLM call failed: %w", err)
		}

		// Add assistant message to history
		assistantMsg := llm.Message{
			Role:    llm.RoleAssistant,
			Content: resp.Content,
		}

		if len(resp.ToolCalls) > 0 {
			assistantMsg.ToolCalls = resp.ToolCalls
		}

		a.mu.Lock()
		a.messages = append(a.messages, assistantMsg)
		a.mu.Unlock()

		// No tool calls → final response
		if len(resp.ToolCalls) == 0 {
			return resp.Content, nil
		}

		if resp.Content != "" {
			lastContent = resp.Content
		}

		// Execute tool calls in parallel
		results := a.executeToolCallsParallel(ctx, resp.ToolCalls)
		for i, tc := range resp.ToolCalls {
			a.mu.Lock()
			a.messages = append(a.messages, llm.Message{
				Role:       llm.RoleTool,
				ToolCallID: tc.ID,
				Content:    results[i],
			})
			a.mu.Unlock()
		}
	}

	if lastContent != "" {
		return lastContent, nil
	}
	return "", fmt.Errorf("agent exceeded maximum iterations (%d)", a.maxIterations)
}

// runStreamLoop runs the ReAct loop with streaming
func (a *Agent) runStreamLoop(ctx context.Context) (string, error) {
	var lastContent string

	for i := 0; i < a.maxIterations; i++ {
		req := llm.ChatRequest{
			Messages:        a.messages,
			Tools:           a.toolDefinitions(),
			Stream:          true,
			ReasoningEffort: a.reasoningLevel,
		}

		ch, err := a.gateway.ChatStream(ctx, req)
		if err != nil {
			return "", fmt.Errorf("LLM stream failed: %w", err)
		}

		// Collect streamed response
		var fullContent strings.Builder
		var fullThinking strings.Builder
		var toolCalls []llm.ToolCall
		var toolCallIndexMap map[int]int // maps stream index → position in toolCalls
		var finishReason string

		for chunk := range ch {
			if chunk.Error != nil {
				if a.onError != nil {
					a.onError(chunk.Error)
				}
				return "", chunk.Error
			}

			if chunk.Delta != "" {
				fullContent.WriteString(chunk.Delta)
				if a.onStreamDelta != nil {
					a.onStreamDelta(chunk.Delta)
				}
			}

			if chunk.Thinking != "" {
				fullThinking.WriteString(chunk.Thinking)
				if a.onStreamThinking != nil {
					a.onStreamThinking(chunk.Thinking)
				}
			}

			if len(chunk.ToolCalls) > 0 {
				// Use ToolCallIndex to properly track and merge streaming tool calls
				if toolCallIndexMap == nil {
					toolCallIndexMap = make(map[int]int)
				}
				for _, tc := range chunk.ToolCalls {
					if tc.ID != "" {
						// New tool call with ID and name (from content_block_start)
						idx := len(toolCalls)
						toolCalls = append(toolCalls, tc)
						toolCallIndexMap[chunk.ToolCallIndex] = idx
					} else if tc.Function.Arguments != "" {
						// Continuation: partial JSON arguments (from input_json_delta)
						if pos, ok := toolCallIndexMap[chunk.ToolCallIndex]; ok {
							toolCalls[pos].Function.Arguments += tc.Function.Arguments
						}
					}
				}
			}

			if chunk.FinishReason != "" {
				finishReason = chunk.FinishReason
			}

			if chunk.Usage != nil && a.onUsage != nil {
				a.onUsage(*chunk.Usage)
			}
		}

		// Add assistant message to history
		assistantMsg := llm.Message{
			Role:    llm.RoleAssistant,
			Content: fullContent.String(),
		}

		if len(toolCalls) > 0 {
			// Tool calls already properly merged in streaming loop via index tracking
			assistantMsg.ToolCalls = toolCalls
		}

		a.mu.Lock()
		a.messages = append(a.messages, assistantMsg)
		a.mu.Unlock()

		// No tool calls → final response, we're done
		if len(toolCalls) == 0 {
			return fullContent.String(), nil
		}

		// Has both text and finish=stop but NO tool calls → model wants to stop
		// 注意：不能在有 toolCalls 时提前返回，否则工具调用会被丢弃
		if finishReason == "stop" && len(toolCalls) == 0 {
			return fullContent.String(), nil
		}

		// Remember last non-empty content in case we hit max iterations
		if fullContent.String() != "" {
			lastContent = fullContent.String()
		}

		// Execute tool calls in parallel (already merged in streaming loop)
		results := a.executeToolCallsParallel(ctx, toolCalls)
		for i, tc := range toolCalls {
			a.mu.Lock()
			a.messages = append(a.messages, llm.Message{
				Role:       llm.RoleTool,
				ToolCallID: tc.ID,
				Content:    results[i],
			})
			a.mu.Unlock()
		}
	}

	// Reached max iterations — return last accumulated content instead of erroring
	if lastContent != "" {
		return lastContent, nil
	}
	return "", fmt.Errorf("agent exceeded maximum iterations (%d)", a.maxIterations)
}

// executeToolCall executes a single tool call
func (a *Agent) executeToolCall(ctx context.Context, tc llm.ToolCall) (string, error) {
	// Parse arguments
	params := make(map[string]interface{})
	if tc.Function.Arguments != "" && tc.Function.Arguments != "{}" {
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &params); err != nil {
			// Try to handle non-JSON arguments gracefully
			if a.onError != nil {
				a.onError(fmt.Errorf("failed to parse tool arguments for %s: %w", tc.Function.Name, err))
			}
			return "", fmt.Errorf("failed to parse tool arguments: %w", err)
		}
	}

	if a.onToolCall != nil {
		a.onToolCall(tc.Function.Name, tc.Function.Arguments)
	}

	// Safety check before execution
	if a.guardrail != nil {
		allowed, err := a.guardrail.CheckWithConfirmAll(tc.Function.Name, params, a.confirmAll)
		if !allowed {
			if a.onError != nil {
				a.onError(fmt.Errorf("operation blocked by safety guardrail: %w", err))
			}
			return "", fmt.Errorf("operation blocked by safety guardrail: %w", err)
		}
	}

	// Execute the tool
	result, err := a.registry.Execute(ctx, tc.Function.Name, params)
	if err != nil {
		return "", fmt.Errorf("tool execution failed: %w", err)
	}

	output := result.Output
	if result.Error != "" {
		output = fmt.Sprintf("Error: %s\nOutput: %s", result.Error, result.Output)
	}

	if a.onToolResult != nil {
		a.onToolResult(tc.Function.Name, output)
	}

	return output, nil
}

// executeToolCallsParallel executes multiple tool calls in parallel with semaphore limiting
func (a *Agent) executeToolCallsParallel(ctx context.Context, toolCalls []llm.ToolCall) []string {
	results := make([]string, len(toolCalls))
	var wg sync.WaitGroup

	// Semaphore to limit concurrency
	sem := make(chan struct{}, a.maxParallelTools)

	for i, tc := range toolCalls {
		wg.Add(1)
		go func(idx int, toolCall llm.ToolCall) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result, err := a.executeToolCall(ctx, toolCall)
			if err != nil {
				results[idx] = "Error: " + err.Error()
			} else {
				results[idx] = result
			}
		}(i, tc)
	}
	wg.Wait()
	return results
}

// toolDefinitions returns LLM tool definitions for all registered tools
func (a *Agent) toolDefinitions() []llm.ToolDefinition {
	defs := a.registry.Definitions()
	result := make([]llm.ToolDefinition, len(defs))
	for i, d := range defs {
		fn, _ := d["function"].(map[string]interface{})
		if fn == nil {
			continue
		}
		name, _ := fn["name"].(string)
		desc, _ := fn["description"].(string)
		result[i] = llm.ToolDefinition{
			Type: "function",
			Function: llm.ToolFuncDefinition{
				Name:        name,
				Description: desc,
				Parameters:  fn["parameters"],
			},
		}
	}
	return result
}

// estimateTokens estimates the number of tokens for a message
// Simple estimation: 1 token ≈ 3 characters (conservative for Chinese + English mix)
func estimateTokens(msg llm.Message) int {
	total := 0

	// Estimate content tokens
	total += len(msg.Content) / 3

	// Estimate tool call tokens
	for _, tc := range msg.ToolCalls {
		total += len(tc.Function.Name) / 3
		total += len(tc.Function.Arguments) / 3
	}

	// Overhead per message
	total += 10

	return total
}

// compressContext compresses conversation history using LLM summarization
// Called with a.mu held. Falls back to truncation if summarization fails.
func (a *Agent) compressContext(ctx context.Context) {
	if len(a.messages) <= 1 {
		return
	}

	totalTokens := 0
	for _, msg := range a.messages {
		totalTokens += estimateTokens(msg)
	}

	// Trigger compression at 75% of max context
	threshold := int(float64(a.maxContextTokens) * 0.75)
	if totalTokens <= threshold {
		return
	}

	// Separate system prompt from conversation messages
	var systemMessages []llm.Message
	var convMessages []llm.Message
	for _, msg := range a.messages {
		if msg.Role == llm.RoleSystem {
			systemMessages = append(systemMessages, msg)
		} else {
			convMessages = append(convMessages, msg)
		}
	}

	// Keep recent 6 turns (12 messages: user+assistant pairs)
	const recentKeep = 12
	if len(convMessages) <= recentKeep {
		// Not enough messages to compress, fall back to truncation
		a.truncateMessages()
		return
	}

	oldMessages := convMessages[:len(convMessages)-recentKeep]
	recentMessages := convMessages[len(convMessages)-recentKeep:]

	// Format old messages for summarization
	var sb strings.Builder
	for _, msg := range oldMessages {
		switch msg.Role {
		case llm.RoleUser:
			sb.WriteString(fmt.Sprintf("用户: %s\n", msg.Content))
		case llm.RoleAssistant:
			if msg.Content != "" {
				sb.WriteString(fmt.Sprintf("助手: %s\n", msg.Content))
			}
			for _, tc := range msg.ToolCalls {
				sb.WriteString(fmt.Sprintf("助手调用工具: %s(%s)\n", tc.Function.Name, tc.Function.Arguments))
			}
		case llm.RoleTool:
			// Truncate long tool results for the summary request
			content := msg.Content
			if len(content) > 500 {
				content = content[:500] + "... (截断)"
			}
			sb.WriteString(fmt.Sprintf("工具结果[%s]: %s\n", msg.Name, content))
		}
	}

	summaryPrompt := fmt.Sprintf(`请将以下对话历史压缩为简洁的摘要，保留：
1. 用户的主要需求和目标
2. 已完成的关键操作和结果
3. 重要的技术决策和发现
4. 未完成的任务和当前上下文状态

对话历史：
%s

请用中文输出摘要，控制在 500 字以内。只输出摘要内容，不要加标题或前缀。`, sb.String())

	// Notify UI
	if a.onCompressing != nil {
		a.onCompressing()
	}

	// Call LLM for summarization (non-streaming, lightweight)
	req := llm.ChatRequest{
		ReasoningEffort: a.reasoningLevel,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: "你是一个对话摘要助手，擅长将长对话压缩为简洁的摘要。"},
			{Role: llm.RoleUser, Content: summaryPrompt},
		},
		Stream: false,
	}

	resp, err := a.gateway.Chat(ctx, req)
	if err != nil {
		// Fallback to truncation
		a.truncateMessages()
		return
	}

	summary := resp.Content
	if summary == "" {
		a.truncateMessages()
		return
	}

	// Rebuild messages: system + summary + recent
	compressed := make([]llm.Message, 0, len(systemMessages)+2+len(recentMessages))
	compressed = append(compressed, systemMessages...)
	compressed = append(compressed, llm.Message{
		Role:    llm.RoleUser,
		Content: "[上下文压缩]\n之前对话的摘要：\n" + summary,
	}, llm.Message{
		Role:    llm.RoleAssistant,
		Content: "已了解之前对话的上下文，继续。",
	})
	compressed = append(compressed, recentMessages...)

	a.messages = compressed
}

// compressContextForce is like compressContext but skips the 75% threshold check.
// Used by manual /compress command.
func (a *Agent) compressContextForce(ctx context.Context) {
	if len(a.messages) <= 1 {
		return
	}

	var systemMessages []llm.Message
	var convMessages []llm.Message
	for _, msg := range a.messages {
		if msg.Role == llm.RoleSystem {
			systemMessages = append(systemMessages, msg)
		} else {
			convMessages = append(convMessages, msg)
		}
	}

	const recentKeep = 12
	if len(convMessages) <= recentKeep {
		a.truncateMessages()
		return
	}

	oldMessages := convMessages[:len(convMessages)-recentKeep]
	recentMessages := convMessages[len(convMessages)-recentKeep:]

	var sb strings.Builder
	for _, msg := range oldMessages {
		switch msg.Role {
		case llm.RoleUser:
			sb.WriteString(fmt.Sprintf("用户: %s\n", msg.Content))
		case llm.RoleAssistant:
			if msg.Content != "" {
				sb.WriteString(fmt.Sprintf("助手: %s\n", msg.Content))
			}
			for _, tc := range msg.ToolCalls {
				sb.WriteString(fmt.Sprintf("助手调用工具: %s(%s)\n", tc.Function.Name, tc.Function.Arguments))
			}
		case llm.RoleTool:
			content := msg.Content
			if len(content) > 500 {
				content = content[:500] + "... (截断)"
			}
			sb.WriteString(fmt.Sprintf("工具结果[%s]: %s\n", msg.Name, content))
		}
	}

	summaryPrompt := fmt.Sprintf(`请将以下对话历史压缩为简洁的摘要，保留：
1. 用户的主要需求和目标
2. 已完成的关键操作和结果
3. 重要的技术决策和发现
4. 未完成的任务和当前上下文状态

对话历史：
%s

请用中文输出摘要，控制在 500 字以内。只输出摘要内容，不要加标题或前缀。`, sb.String())

	if a.onCompressing != nil {
		a.onCompressing()
	}

	req := llm.ChatRequest{
		ReasoningEffort: a.reasoningLevel,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: "你是一个对话摘要助手，擅长将长对话压缩为简洁的摘要。"},
			{Role: llm.RoleUser, Content: summaryPrompt},
		},
		Stream: false,
	}

	resp, err := a.gateway.Chat(ctx, req)
	if err != nil {
		a.truncateMessages()
		return
	}

	summary := resp.Content
	if summary == "" {
		a.truncateMessages()
		return
	}

	compressed := make([]llm.Message, 0, len(systemMessages)+2+len(recentMessages))
	compressed = append(compressed, systemMessages...)
	compressed = append(compressed, llm.Message{
		Role:    llm.RoleUser,
		Content: "[上下文压缩]\n之前对话的摘要：\n" + summary,
	}, llm.Message{
		Role:    llm.RoleAssistant,
		Content: "已了解之前对话的上下文，继续。",
	})
	compressed = append(compressed, recentMessages...)
	a.messages = compressed
}

// truncateMessages truncates message history to fit within maxContextTokens
// Strategy: preserve system prompt + recent messages, remove old tool results first
func (a *Agent) truncateMessages() {
	if len(a.messages) <= 1 {
		return
	}

	// Calculate current total tokens
	totalTokens := 0
	for _, msg := range a.messages {
		totalTokens += estimateTokens(msg)
	}

	// If within limit, no truncation needed
	if totalTokens <= a.maxContextTokens {
		return
	}

	// Target: keep within 80% of max to leave room for response
	targetTokens := int(float64(a.maxContextTokens) * 0.8)

	// Strategy:
	// 1. Always keep system prompt
	// 2. Keep recent messages (last 10 user/assistant turns)
	// 3. Remove old tool results (they're usually large)
	// 4. If still too large, remove old assistant messages

	// Separate messages by type
	var systemMessages []llm.Message
	var toolMessages []llm.Message
	var recentMessages []llm.Message

	for i, msg := range a.messages {
		if msg.Role == llm.RoleSystem {
			systemMessages = append(systemMessages, msg)
			continue
		}

		if msg.Role == llm.RoleTool {
			toolMessages = append(toolMessages, msg)
			continue
		}

		// Keep recent user/assistant messages (last 10 pairs)
		if i >= len(a.messages)-20 {
			recentMessages = append(recentMessages, msg)
		}
	}

	// Rebuild messages with truncation
	truncated := make([]llm.Message, 0, len(a.messages))

	// Add system messages
	truncated = append(truncated, systemMessages...)

	// Add recent messages
	truncated = append(truncated, recentMessages...)

	// Check if we need to remove more messages
	currentTokens := 0
	for _, msg := range truncated {
		currentTokens += estimateTokens(msg)
	}

	// If still too large, keep removing from the beginning of recentMessages
	for currentTokens > targetTokens && len(recentMessages) > 2 {
		// Remove the oldest message (first non-system message)
		removedMsg := recentMessages[0]
		recentMessages = recentMessages[1:]
		currentTokens -= estimateTokens(removedMsg)

		// Rebuild truncated
		truncated = append(systemMessages, recentMessages...)
	}

	// Replace messages
	a.messages = truncated
}

// mergeToolCalls merges streaming tool call chunks into complete tool calls
func mergeToolCalls(chunks []llm.ToolCall) []llm.ToolCall {
	if len(chunks) == 0 {
		return nil
	}

	merged := make(map[string]*llm.ToolCall)
	for _, tc := range chunks {
		if tc.ID != "" {
			// New tool call
			merged[tc.ID] = &llm.ToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: llm.FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		} else if tc.Function.Arguments != "" {
			// Continuation of existing tool call (streaming arguments)
			for id, existing := range merged {
				if existing.Function.Name == tc.Function.Name || existing.Function.Name == "" {
					existing.Function.Arguments += tc.Function.Arguments
					merged[id] = existing
					break
				}
			}
		}
	}

	result := make([]llm.ToolCall, 0, len(merged))
	for _, tc := range merged {
		result = append(result, *tc)
	}
	return result
}




