package cmd

// tui_messages.go — Agent 回调 → Bubbletea 消息桥接

import (
	"github.com/mimo-cli/mimo-cli/internal/agent"
	"github.com/mimo-cli/mimo-cli/internal/llm"
	"github.com/mimo-cli/mimo-cli/internal/safety"
)

// thinkingMsg 表示思考过程的增量更新
type thinkingMsg struct{ delta string }

// deltaMsg 表示流式文本增量
type deltaMsg struct{ text string }

// toolCallMsg 表示一次工具调用
type toolCallMsg struct {
	name string
	args string
}

// toolResultMsg 表示工具执行结果
type toolResultMsg struct {
	name   string
	result string
}

// agentDoneMsg 表示 Agent 响应完成
type agentDoneMsg struct{ response string }

// agentErrMsg 表示 Agent 运行出错
type agentErrMsg struct{ err error }

// usageMsg 表示本轮 API 调用的 token 用量
type usageMsg struct{ usage llm.Usage }

// compressingMsg 表示上下文压缩正在进行
type compressingMsg struct{}

// compressDoneMsg 表示手动压缩完成
type compressDoneMsg struct {
	before int
	after  int
}

// confirmMsg 表示需要用户确认的安全操作
type confirmMsg struct {
	action safety.Action
	result chan bool
}

// spinnerTickMsg 表示旋转动画定时刷新
type spinnerTickMsg struct{}

// planGeneratedMsg 表示计划生成完成
type planGeneratedMsg struct {
	plan *agent.Plan
}

// planStepStartMsg 表示计划步骤开始执行
type planStepStartMsg struct {
	step *agent.PlanStep
}

// planStepDoneMsg 表示计划步骤完成执行
type planStepDoneMsg struct {
	step *agent.PlanStep
}

// planningMsg 表示正在生成计划
type planningMsg struct{ message string }

// MCP 向导相关消息
type mcpWizardStepMsg struct{ step int }
type mcpInstallDoneMsg struct {
	success bool
	err     error
}
