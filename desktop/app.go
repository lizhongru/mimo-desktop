package desktop

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"sync"
	"time"

	"github.com/mimo-cli/mimo-cli/internal/agent"
	"github.com/mimo-cli/mimo-cli/internal/backup"
	iconfig "github.com/mimo-cli/mimo-cli/internal/config"
	mctx "github.com/mimo-cli/mimo-cli/internal/context"
	"github.com/mimo-cli/mimo-cli/internal/ignore"
	"github.com/mimo-cli/mimo-cli/internal/llm"
	"github.com/mimo-cli/mimo-cli/internal/mcp"
	"github.com/mimo-cli/mimo-cli/internal/memory"
	"github.com/mimo-cli/mimo-cli/internal/safety"
	"github.com/mimo-cli/mimo-cli/internal/session"
	"github.com/mimo-cli/mimo-cli/internal/skill"
	"github.com/mimo-cli/mimo-cli/internal/task"
	"github.com/mimo-cli/mimo-cli/internal/tools"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main Wails binding object exposed to the frontend.
type App struct {
	ctx context.Context

	// Infrastructure (reused from internal/)
	agent         *agent.Agent
	gateway       *llm.Gateway
	registry      *tools.Registry
	guardrail     *safety.Guardrail
	cfg           *iconfig.Config
	sessionStore  *session.Store
	mcpManager    *mcp.Manager
	backupMgr     *backup.Manager
	ignoreMatcher *ignore.Matcher
	ctxManager    *mctx.Manager
	memorySvc     *memory.Service // Memory system service

	// State
	currentSessionID string
	isBusy           bool
	confirmAll       bool
	cancelChat       context.CancelFunc
	mu               sync.Mutex

	// Safety confirmation bridge — agent goroutine blocks on this channel
	confirmChan chan bool
}

// ConfigType is the internal config type, exposed for frontend binding.
type ConfigType = iconfig.Config

// NewApp creates and initializes the App with all infrastructure.
func NewApp() (*App, error) {
	a := &App{
		confirmChan: make(chan bool, 1),
	}

	// 1. Load config
	c, err := iconfig.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	a.cfg = c

	// 2. Initialize ignore rules
	wd, _ := os.Getwd()
	ign := ignore.New()
	ign.LoadFile(filepath.Join(wd, ".mimoignore"))
	ign.AddPatterns(c.Context.IgnorePatterns)
	a.ignoreMatcher = ign

	// 3. Initialize backup manager
	bm := backup.NewManager(c.Safety.BackupDir)
	a.backupMgr = bm

	// 4. Initialize LLM Gateway
	gateway := llm.NewGateway(c)
	a.gateway = gateway

	// 5. Create tool registry with built-in tools
	registry := tools.DefaultRegistry(ign, bm)
	a.registry = registry

	// 6. Initialize MCP
	mcpManager := mcp.NewManager()
	a.mcpManager = mcpManager
	if len(c.MCP.Servers) > 0 {
		mcpErrs := mcpManager.ConnectAll(c.MCP)
		for _, err := range mcpErrs {
			fmt.Fprintf(os.Stderr, "MCP warning: %v\n", err)
		}
		mcpManager.RegisterTools(registry)
		fmt.Fprintf(os.Stderr, "MCP: connected %d servers, registered %d tools\n",
			len(mcpManager.ServerNames()), len(mcpManager.GetTools()))
	}

	// 7. Safety guardrail
	safetyLevel := safety.SafetyLevel(c.Safety.Level)
	if safetyLevel == "" {
		safetyLevel = safety.LevelConfirm
	}
	classifier := safety.NewClassifier(
		c.Safety.BlockedCommands,
		c.Safety.ProtectedFiles,
		c.Safety.ProtectedBranches,
	)
	guardrail := safety.NewGuardrail(safetyLevel, classifier, c.Safety.AuditLog)
	guardrail.SetPermission(c.Agent.Permission)
	guardrail.SetRuleset(permissionRulesetFromConfig(c.Permission.Rules))
	guardrail.SetWorkspaceRoot(wd)
	a.guardrail = guardrail

	// 8. Context manager
	ctxManager := mctx.NewManager(wd, c.Context.MaxTokens, ign)
	a.ctxManager = ctxManager

	// 9. Create Agent
	mimoAgent := agent.NewAgent(gateway, registry, guardrail, c.Context.MaxTokens, c.Agent.MaxIterations)

	mimoAgent.SystemPrompt(a.buildSystemPrompt(wd))

	// Set planning mode
	switch c.Agent.PlanningMode {
	case "react":
		mimoAgent.SetPlanningMode(agent.ModeReact)
	case "plan-execute":
		mimoAgent.SetPlanningMode(agent.ModePlanExecute)
	default:
		mimoAgent.SetPlanningMode(agent.ModeAuto)
	}

	// Set reasoning level
	if c.Agent.ReasoningLevel != "" {
		mimoAgent.SetReasoningLevel(c.Agent.ReasoningLevel)
	}

	a.agent = mimoAgent

	// 10. Open session store
	sessionStore, err := session.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cannot open session store: %v\n", err)
	} else {
		a.sessionStore = sessionStore
	}

	// 11. Initialize memory service (lazy init, will be created on first use)
	// Memory service is initialized lazily in memoryService() method
	registry.RegisterMemoryTools(a.memoryService)
	registry.RegisterTaskTools(
		func() *task.Registry {
			if a.sessionStore == nil {
				return nil
			}
			return task.NewRegistry(a.sessionStore.DB())
		},
		func() string {
			a.mu.Lock()
			defer a.mu.Unlock()
			return a.currentSessionID
		},
	)
	registry.RegisterActorTools(
		getActorRegistry,
		func() string {
			a.mu.Lock()
			defer a.mu.Unlock()
			return a.currentSessionID
		},
	)

	// Initialize actor registry with real LLM executor
	initActorRegistry(a)

	return a, nil
}

type enabledProjectSkillsFile struct {
	Skills []string `json:"skills"`
}

func buildEnabledProjectSkillsContext(projectDir string, selectedSkills []string) string {
	data, err := os.ReadFile(filepath.Join(projectDir, ".mimo", "skills", "enabled.json"))
	if err != nil {
		return ""
	}

	var enabled enabledProjectSkillsFile
	if err := json.Unmarshal(data, &enabled); err != nil || len(enabled.Skills) == 0 {
		return ""
	}

	allowed := make(map[string]bool, len(enabled.Skills))
	for _, name := range enabled.Skills {
		normalized, err := skill.SafeCandidateName(name)
		if err == nil {
			allowed[normalized] = true
		}
	}

	var names []string
	for _, name := range selectedSkills {
		normalized, err := skill.SafeCandidateName(name)
		if err == nil && allowed[normalized] {
			names = append(names, normalized)
		}
	}
	if len(names) == 0 {
		return ""
	}

	var content strings.Builder
	var selectedCommands []string
	for _, name := range names {
		normalized, err := skill.SafeCandidateName(name)
		if err != nil {
			continue
		}
		skillData, err := os.ReadFile(filepath.Join(projectDir, ".mimo", "skills", normalized, "SKILL.md"))
		if err != nil {
			continue
		}
		selectedCommands = append(selectedCommands, extractSkillCommands(string(skillData))...)
		if content.Len() == 0 {
			content.WriteString("## Selected Project Skills\n\n")
			content.WriteString("The user explicitly selected the following project skills for this turn. Treat them as active instructions for the current request, not as optional background context.\n")
			content.WriteString("The selected skills override prior conversation interpretations of similar user phrases. If the user repeats a previous request, reinterpret it according to the skills selected for this current turn.\n")
			content.WriteString("Apply the selected skill workflow before doing broad project exploration. If a selected skill contains commands, run those exact commands unless they are unsafe or impossible, and explain any deviation to the user.\n")
		}
		content.WriteString("\n### ")
		content.WriteString(normalized)
		content.WriteString("\n\n")
		content.Write(skillData)
		content.WriteString("\n")
	}
	if len(selectedCommands) > 0 {
		content.WriteString("\n## Current Turn Selected Skill Commands\n\n")
		content.WriteString("Execute these selected Skill commands first, before substituting broader project tests or reusing prior interpretations:\n")
		for _, command := range selectedCommands {
			content.WriteString("- ")
			content.WriteString(command)
			content.WriteString("\n")
		}
	}
	return strings.TrimSpace(content.String())
}

func extractSkillCommands(skillMarkdown string) []string {
	var commands []string
	inCommands := false
	seen := make(map[string]bool)
	for _, line := range strings.Split(skillMarkdown, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			inCommands = strings.EqualFold(trimmed, "## Commands")
			continue
		}
		if !inCommands || !strings.HasPrefix(trimmed, "- `") || !strings.HasSuffix(trimmed, "`") {
			continue
		}
		command := strings.TrimSuffix(strings.TrimPrefix(trimmed, "- `"), "`")
		if command != "" && !seen[command] {
			commands = append(commands, command)
			seen[command] = true
		}
	}
	return commands
}

func (a *App) buildSystemPrompt(projectDir string, selectedSkills ...string) string {
	if projectDir == "" {
		projectDir, _ = os.Getwd()
	}
	ctxManager := mctx.NewManager(projectDir, a.cfg.Context.MaxTokens, a.ignoreMatcher)
	systemPrompt := ctxManager.BuildSystemPrompt()
	if skillContext := buildEnabledProjectSkillsContext(projectDir, selectedSkills); skillContext != "" {
		systemPrompt += "\n\n" + skillContext
	}
	if config := getMultiAgentManager().GetCurrent(); config != nil {
		systemPrompt += "\n\n## Current Agent Mode\n\n"
		systemPrompt += fmt.Sprintf("Active agent: %s (%s)\n", config.Name, config.Mode)
		if config.Description != "" {
			systemPrompt += fmt.Sprintf("Agent purpose: %s\n", config.Description)
		}
		if config.Prompt != "" {
			systemPrompt += "\nAgent-specific instructions:\n"
			systemPrompt += config.Prompt
			systemPrompt += "\n"
		}
		if len(config.ToolAllowlist) > 0 {
			systemPrompt += "\nOnly these tools are available in this mode:\n"
			for _, toolName := range config.ToolAllowlist {
				systemPrompt += fmt.Sprintf("- %s\n", toolName)
			}
		}
	}
	if a.mcpManager == nil {
		return systemPrompt
	}
	if mcpTools := a.mcpManager.GetTools(); len(mcpTools) > 0 {
		systemPrompt += "\n\n## MCP Tools (Model Context Protocol)\n\n"
		systemPrompt += "You have access to additional tools from MCP servers. Use them when appropriate:\n\n"
		for _, t := range mcpTools {
			systemPrompt += fmt.Sprintf("- %s: %s\n", t.Name(), t.Description())
		}
		systemPrompt += "\nMCP tools are prefixed with the server name (e.g., filesystem__list_directory). Use them like any other tool.\n"
	}
	return systemPrompt
}

// Startup is called when the Wails application starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.registerCallbacks()
	a.registerConfirmCallback()
}

// Shutdown is called when the Wails application shuts down.
func (a *App) Shutdown(ctx context.Context) {
	if a.mcpManager != nil {
		a.mcpManager.CloseAll()
	}
	if a.sessionStore != nil {
		a.sessionStore.Close()
	}
}

// registerCallbacks wires agent callbacks to Wails events.
func (a *App) registerCallbacks() {
	a.agent.SetStreamThinkingCallback(func(delta string) {
		runtime.EventsEmit(a.ctx, EventThinking, delta)
	})
	a.agent.SetStreamDeltaCallback(func(text string) {
		runtime.EventsEmit(a.ctx, EventDelta, text)
	})
	a.agent.SetToolCallCallback(func(name, args string) {
		runtime.EventsEmit(a.ctx, EventToolCall, name, args)
	})
	a.agent.SetToolResultCallback(func(name, result string) {
		runtime.EventsEmit(a.ctx, EventToolResult, name, result)
	})
	a.agent.SetErrorCallback(func(err error) {
		errMsg := err.Error()
		if errMsg == "invalid parameters" || errMsg == "missing required parameter" {
			return
		}
		runtime.EventsEmit(a.ctx, EventError, errMsg)
	})
	a.agent.SetUsageCallback(func(u llm.Usage) {
		runtime.EventsEmit(a.ctx, EventUsage, u)
	})
	a.agent.SetCompressingCallback(func() {
		runtime.EventsEmit(a.ctx, EventCompressing)
	})
	a.agent.SetPlanningCallback(func(message string) {
		runtime.EventsEmit(a.ctx, EventPlanning, message)
	})
	a.agent.SetPlanGeneratedCallback(func(plan *agent.Plan) {
		runtime.EventsEmit(a.ctx, EventPlanGenerated, plan)
	})
	a.agent.SetPlanStepStartCallback(func(step *agent.PlanStep) {
		runtime.EventsEmit(a.ctx, EventPlanStepStart, step)
	})
	a.agent.SetPlanStepDoneCallback(func(step *agent.PlanStep) {
		runtime.EventsEmit(a.ctx, EventPlanStepDone, step)
	})
}

// Window controls for frameless window

func (a *App) WindowMinimise() {
	runtime.WindowMinimise(a.ctx)
}

func (a *App) WindowMaximise() {
	runtime.WindowToggleMaximise(a.ctx)
}

func (a *App) WindowClose() {
	runtime.Quit(a.ctx)
}

func (a *App) WindowIsMaximised() bool {
	return runtime.WindowIsMaximised(a.ctx)
}

// OpenInExplorer opens the given path in the system file explorer.
func (a *App) OpenInExplorer(path string) error {
	if path == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine working directory: %w", err)
		}
		path = wd
	}
	switch goruntime.GOOS {
	case "windows":
		return exec.Command("explorer", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}

// registerConfirmCallback sets up the safety confirmation bridge.
func (a *App) registerConfirmCallback() {
	a.guardrail.SetConfirmCallback(func(action safety.Action) (bool, error) {
		runtime.EventsEmit(a.ctx, EventSafetyConfirm, action)
		// Block until frontend responds (with timeout)
		select {
		case result := <-a.confirmChan:
			return result, nil
		case <-time.After(60 * time.Second):
			return false, fmt.Errorf("safety confirmation timed out after 60s")
		}
	})
}
