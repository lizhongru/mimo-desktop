package cmd

// interactive.go — Bubbletea 入口，初始化 Agent + 启动 TUI

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mimo-cli/mimo-cli/internal/agent"
	"github.com/mimo-cli/mimo-cli/internal/backup"
	mctx "github.com/mimo-cli/mimo-cli/internal/context"
	"github.com/mimo-cli/mimo-cli/internal/ignore"
	"github.com/mimo-cli/mimo-cli/internal/llm"
	"github.com/mimo-cli/mimo-cli/internal/safety"
	"github.com/mimo-cli/mimo-cli/internal/session"
	"github.com/mimo-cli/mimo-cli/internal/tools"
	"github.com/mimo-cli/mimo-cli/internal/version"
	"github.com/mimo-cli/mimo-cli/internal/mcp"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runInteractive()
	}
	rootCmd.AddCommand(interactiveCmd)
}

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Enter interactive mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInteractive()
	},
}

func runInteractive() error {
	modelName, _ := rootCmd.Flags().GetString("model")
	if modelName == "" {
		modelName = cfg.DefaultModel
	}
	modelCfg, _ := cfg.GetModelConfig(modelName)
	displayModel := modelCfg.Model
	if displayModel == "" {
		displayModel = modelName
	}

	// 初始化忽略规则
	wd, _ := os.Getwd()
	ign := ignore.New()
	ign.LoadFile(filepath.Join(wd, ".mimoignore"))
	ign.AddPatterns(cfg.Context.IgnorePatterns)

	// 初始化备份管理
	bm := backup.NewManager(cfg.Safety.BackupDir)

	// 初始化 LLM Gateway
	gateway := llm.NewGateway(cfg)
	registry := tools.DefaultRegistry(ign, bm)

	// 初始化 MCP 管理器
	mcpManager := mcp.NewManager()
	fmt.Fprintf(os.Stderr, "MCP: 配置中的服务器数量: %d\n", len(cfg.MCP.Servers))
	if len(cfg.MCP.Servers) > 0 {
		for name, srv := range cfg.MCP.Servers {
			fmt.Fprintf(os.Stderr, "MCP: 服务器 %q enabled=%v command=%q url=%q\n", name, srv.Enabled, srv.Command, srv.URL)
		}
		mcpErrs := mcpManager.ConnectAll(cfg.MCP)
		for _, err := range mcpErrs {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
		// 注册 MCP 工具到 Registry
		mcpManager.RegisterTools(registry)
		fmt.Fprintf(os.Stderr, "MCP: 已连接 %d 个服务器, 注册 %d 个工具\n", len(mcpManager.ServerNames()), len(mcpManager.GetTools()))
		defer mcpManager.CloseAll()
	} else {
		fmt.Fprintf(os.Stderr, "MCP: 未配置任何服务器\n")
	}


	// 安全防护
	safetyLevel := safety.LevelConfirm
	if level, _ := rootCmd.Flags().GetString("safety"); level != "" {
		safetyLevel = safety.SafetyLevel(level)
	}
	classifier := safety.NewClassifier(
		cfg.Safety.BlockedCommands,
		cfg.Safety.ProtectedFiles,
		cfg.Safety.ProtectedBranches,
	)
	auditLogPath := cfg.Safety.AuditLog
	guardrail := safety.NewGuardrail(safetyLevel, classifier, auditLogPath)
	guardrail.SetPermission(cfg.Agent.Permission)

	// 上下文管理
	ctxManager := mctx.NewManager(wd, cfg.Context.MaxTokens, ign)

	// 初始化 Agent
	mimoAgent := agent.NewAgent(gateway, registry, guardrail, cfg.Context.MaxTokens, cfg.Agent.MaxIterations)
	// 构建系统提示词，附加 MCP 工具描述
	systemPrompt := ctxManager.BuildSystemPrompt()
	if mcpTools := mcpManager.GetTools(); len(mcpTools) > 0 {
		systemPrompt += "\n\n## MCP Tools (Model Context Protocol)\n\n"
		systemPrompt += "You have access to additional tools from MCP servers. Use them when appropriate:\n\n"
		for _, t := range mcpTools {
			systemPrompt += fmt.Sprintf("- %s: %s\n", t.Name(), t.Description())
		}
		systemPrompt += "\nMCP tools are prefixed with the server name (e.g., filesystem__list_directory). Use them like any other tool.\n"
	}
	mimoAgent.SystemPrompt(systemPrompt)

	// 设置规划模式
	if cfg.Agent.PlanningMode != "" {
		switch cfg.Agent.PlanningMode {
		case "react":
			mimoAgent.SetPlanningMode(agent.ModeReact)
		case "plan-execute":
			mimoAgent.SetPlanningMode(agent.ModePlanExecute)
		case "auto":
			mimoAgent.SetPlanningMode(agent.ModeAuto)
		default:
			mimoAgent.SetPlanningMode(agent.ModeAuto)
		}
	}

	appVersion := "v" + version.Version

	// 加载主题设置
	if cfg.Theme != "" {
		SetTheme(cfg.Theme)
	}

	// 初始化会话存储
	sessionStore, err := session.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cannot open session store: %v\n", err)
	}
	if sessionStore != nil {
		defer sessionStore.Close()
	}

	// 创建 Bubbletea Model
	model := newTUIModel(mimoAgent, bm, mcpManager, displayModel, appVersion, cfg.UserName, sessionStore)

	// 设置安全确认回调
	guardrail.SetConfirmCallback(func(action safety.Action) (bool, error) {
		resultChan := make(chan bool, 1)
		model.streamChan <- confirmMsg{action: action, result: resultChan}
		result := <-resultChan
		return result, nil
	})

	// 注册 Agent 回调 → channel 消息桥接
	mimoAgent.SetStreamThinkingCallback(func(delta string) {
		model.streamChan <- thinkingMsg{delta: delta}
	})
	mimoAgent.SetStreamDeltaCallback(func(text string) {
		model.streamChan <- deltaMsg{text: text}
	})
	mimoAgent.SetToolCallCallback(func(name, args string) {
		model.streamChan <- toolCallMsg{name: name, args: args}
	})
	mimoAgent.SetToolResultCallback(func(name, result string) {
		model.streamChan <- toolResultMsg{name: name, result: result}
	})
	mimoAgent.SetErrorCallback(func(err error) {
		errMsg := err.Error()
		// 忽略参数校验错误（流式中常见）
		if errMsg == "invalid parameters" || errMsg == "missing required parameter" {
			return
		}
		model.streamChan <- agentErrMsg{err: err}
	})
	mimoAgent.SetUsageCallback(func(u llm.Usage) {
		model.streamChan <- usageMsg{usage: u}
	})
	mimoAgent.SetCompressingCallback(func() {
		model.streamChan <- compressingMsg{}
	})
	mimoAgent.SetPlanningCallback(func(message string) {
		model.streamChan <- planningMsg{message: message}
	})

	// 启动 Bubbletea（欢迎页会在第一个 WindowSizeMsg 时生成）
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 优雅退出提示
	// 退出后打印对话历史到普通终端（alt-screen 内容会消失）
	if m, ok := finalModel.(tuiModel); ok {
		printSessionHistory(m)
	}
	fmt.Println("\n  Goodbye! 👋")

	return nil
}







