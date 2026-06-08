package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mimo-cli/mimo-cli/internal/agent"
	"github.com/mimo-cli/mimo-cli/internal/backup"
	mctx "github.com/mimo-cli/mimo-cli/internal/context"
	"github.com/mimo-cli/mimo-cli/internal/ignore"
	"github.com/mimo-cli/mimo-cli/internal/llm"
	"github.com/mimo-cli/mimo-cli/internal/safety"
	"github.com/mimo-cli/mimo-cli/internal/tools"
	"github.com/spf13/cobra"
)

// 语义化退出码
const (
	ExitRunOK     = 0 // 成功
	ExitRunErr    = 1 // 执行错误
	ExitRunReview = 2 // 需要人工审核
)

// runExitError 封装带退出码的错误
type runExitError struct {
	code    int
	message string
}

func (e *runExitError) Error() string { return e.message }
func (e *runExitError) ExitCode() int { return e.code }

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run [task]",
	Short: "Execute a single task (non-interactive mode)",
	Long: `Execute a single task described in natural language. Ideal for scripts and CI/CD.
Supports pipe input: echo "fix the bug" | mimo run`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		task := strings.Join(args, " ")

		// 管道输入：当 stdin 不是终端时，从 stdin 读取
		if task == "" {
			stat, _ := os.Stdin.Stat()
			if stat != nil && (stat.Mode()&os.ModeCharDevice) == 0 {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return &runExitError{ExitRunErr, fmt.Sprintf("读取 stdin 失败: %v", err)}
				}
				task = strings.TrimSpace(string(data))
			}
		}

		if task == "" {
			return &runExitError{ExitRunErr, "请提供任务描述，例如: mimo run \"fix the bug\" 或 echo \"task\" | mimo run"}
		}

		err := runSingleTask(task)
		if err != nil {
			// 如果已经是 runExitError，直接返回
			if exitErr, ok := err.(*runExitError); ok {
				return exitErr
			}
			return &runExitError{ExitRunErr, err.Error()}
		}
		return nil
	},
}

// runResult 用于 JSON 输出
type runResult struct {
	Task       string `json:"task"`
	Response   string `json:"response"`
	Duration   string `json:"duration"`
	DurationMs int64  `json:"duration_ms"`
	Tokens     int    `json:"tokens,omitempty"`
	ExitCode   int    `json:"exit_code"`
	Error      string `json:"error,omitempty"`
}

func runSingleTask(task string) error {
	// Get model name
	modelName, _ := rootCmd.Flags().GetString("model")
	if modelName == "" {
		modelName = cfg.DefaultModel
	}

	verbose, _ := rootCmd.Flags().GetBool("verbose")
	outputFmt, _ := rootCmd.Flags().GetString("output")

	// JSON 模式下不输出流式文本
	isJSON := outputFmt == "json"
	isSilent := isJSON // JSON 模式静默流式输出

	// Create LLM gateway
	gateway := llm.NewGateway(cfg)

	// 初始化忽略规则
	wd, _ := os.Getwd()
	ign := ignore.New()
	ign.LoadFile(filepath.Join(wd, ".mimoignore"))
	ign.AddPatterns(cfg.Context.IgnorePatterns)

	// 初始化备份管理
	bm := backup.NewManager(cfg.Safety.BackupDir)

	// Create tool registry
	registry := tools.DefaultRegistry(ign, bm)

	// Create context manager
	ctxManager := mctx.NewManager(wd, cfg.Context.MaxTokens, ign)

	// Create safety guardrail
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

	// Create agent
	mimoAgent := agent.NewAgent(gateway, registry, guardrail, cfg.Context.MaxTokens, cfg.Agent.MaxIterations)

	// Set system prompt
	systemPrompt := ctxManager.BuildSystemPrompt()
	mimoAgent.SystemPrompt(systemPrompt)

	if verbose {
		fmt.Fprintf(os.Stderr, "[mimo] Model: %s\n", modelName)
		fmt.Fprintf(os.Stderr, "[mimo] Task: %s\n", task)
		fmt.Fprintf(os.Stderr, "[mimo] Working directory: %s\n", wd)
		fmt.Fprintf(os.Stderr, "[mimo] Output format: %s\n\n", outputFmt)
	}

	// 收集完整响应
	var fullResponse strings.Builder
	var usage llm.Usage
	needReview := false

	// Set up callbacks
	mimoAgent.SetStreamDeltaCallback(func(delta string) {
		fullResponse.WriteString(delta)
		if !isSilent {
			fmt.Print(delta)
		}
	})

	mimoAgent.SetStreamThinkingCallback(func(thinking string) {
		if verbose {
			fmt.Fprintf(os.Stderr, "\n💭 %s", thinking)
		}
	})

	mimoAgent.SetToolCallCallback(func(name, args string) {
		if verbose {
			fmt.Fprintf(os.Stderr, "\n[mimo] Tool call: %s(%s)\n", name, args)
		}
	})

	mimoAgent.SetToolResultCallback(func(name, result string) {
		if verbose {
			fmt.Fprintf(os.Stderr, "[mimo] Tool result: %s\n", truncateOutput(result, 200))
		}
	})

	mimoAgent.SetUsageCallback(func(u llm.Usage) {
		usage = u
	})

	// Run agent
	ctx := context.Background()
	start := time.Now()
	response, err := mimoAgent.ChatStream(ctx, task)
	elapsed := time.Since(start)

	if err != nil {
		errMsg := err.Error()
		if isJSON {
			outputJSON(task, "", elapsed, usage, ExitRunErr, errMsg)
			return &runExitError{ExitRunErr, errMsg}
		}
		fmt.Fprintf(os.Stderr, "\n[mimo] Error: %v\n", err)
		return &runExitError{ExitRunErr, errMsg}
	}

	// 流式模式下 response 已经通过回调输出，只需换行
	if !isSilent {
		fmt.Println()
	}

	finalResponse := fullResponse.String()
	if finalResponse == "" {
		finalResponse = response
	}

	// 检测是否需要人工审核（响应中包含特定标记）
	lower := strings.ToLower(finalResponse)
	if strings.Contains(lower, "需要人工审核") || strings.Contains(lower, "needs review") ||
		strings.Contains(lower, "⚠ review") {
		needReview = true
	}

	// 输出格式化
	switch outputFmt {
	case "json":
		code := ExitRunOK
		if needReview {
			code = ExitRunReview
		}
		outputJSON(task, finalResponse, elapsed, usage, code, "")
		if needReview {
			return &runExitError{ExitRunReview, "任务完成，但需要人工审核"}
		}
	case "markdown":
		outputMarkdown(task, finalResponse, elapsed, usage, modelName)
	default:
		// text 模式：流式输出已完成，仅 verbose 时输出摘要
		if verbose {
			fmt.Fprintf(os.Stderr, "\n[mimo] Done. %d tokens, %s\n", usage.TotalTokens, formatDuration(elapsed))
		}
	}

	if needReview {
		return &runExitError{ExitRunReview, "任务完成，但需要人工审核"}
	}
	return nil
}

// outputJSON 输出 JSON 格式结果
func outputJSON(task, response string, elapsed time.Duration, usage llm.Usage, exitCode int, errMsg string) {
	result := runResult{
		Task:       task,
		Response:   response,
		Duration:   formatDuration(elapsed),
		DurationMs: elapsed.Milliseconds(),
		Tokens:     usage.TotalTokens,
		ExitCode:   exitCode,
		Error:      errMsg,
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
}

// outputMarkdown 输出 Markdown 格式结果
func outputMarkdown(task, response string, elapsed time.Duration, usage llm.Usage, modelName string) {
	fmt.Printf("# MiMo CLI 执行结果\n\n")
	fmt.Printf("- **任务**: %s\n", task)
	fmt.Printf("- **模型**: %s\n", modelName)
	fmt.Printf("- **耗时**: %s\n", formatDuration(elapsed))
	if usage.TotalTokens > 0 {
		fmt.Printf("- **Token**: %d (↑%d ↓%d)\n", usage.TotalTokens, usage.PromptTokens, usage.CompletionTokens)
	}
	fmt.Printf("\n## 响应\n\n%s\n", response)
}

// truncateOutput truncates output to maxLen characters
func truncateOutput(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
