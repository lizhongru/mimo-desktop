package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/mimo-cli/mimo-cli/internal/agent"
	"github.com/mimo-cli/mimo-cli/internal/backup"
	"github.com/mimo-cli/mimo-cli/internal/config"
	"github.com/mimo-cli/mimo-cli/internal/llm"
	"github.com/mimo-cli/mimo-cli/internal/safety"
	"github.com/mimo-cli/mimo-cli/internal/mcp"
	"github.com/mimo-cli/mimo-cli/internal/session"
)

type chatMessage struct {
	role, content, tool, args, result string
	thinking                         string  // 保存 thinking 内容
	tokens, toolCalls                int
	duration                         time.Duration
	toolLines  []string  // 工具调用和思考过程的行
}
type cmdEntry struct{ name, desc string }
var commands = []cmdEntry{
	{"/clear", "清屏"}, {"/compress", "压缩上下文"}, {"/confirm", "重新启用安全确认"},
	{"/copy", "复制到剪贴板"}, {"/exit", "退出"}, {"/export", "导出对话"},
	{"/help", "帮助"}, {"/model", "切换模型"}, {"/mcp", "MCP 管理 (add/remove/status)"},
	{"/name", "改昵称"}, {"/plan", "切换规划模式"}, {"/resume", "恢复会话"},
	{"/rollback", "回滚备份"}, {"/theme", "切换主题"},
}

// MCPRecommendedServer 推荐的 MCP 服务器
type MCPRecommendedServer struct {
	Name        string
	Package     string
	Description string
	Category    string
}

// mcpRecommendedServers 推荐的 MCP 服务器列表
var mcpRecommendedServers = []MCPRecommendedServer{
	{"filesystem", "@modelcontextprotocol/server-filesystem", "文件系统操作 (读写、搜索、目录管理)", "文件"},
	{"sqlite", "@modelcontextprotocol/server-sqlite", "SQLite 数据库查询和管理", "数据库"},
	{"postgres", "@modelcontextprotocol/server-postgres", "PostgreSQL 数据库查询", "数据库"},
	{"github", "@modelcontextprotocol/server-github", "GitHub 仓库、Issue、PR 管理", "开发"},
	{"gitlab", "@modelcontextprotocol/server-gitlab", "GitLab 项目管理", "开发"},
	{"git", "@modelcontextprotocol/server-git", "Git 仓库操作 (提交、分支、差异)", "开发"},
	{"brave-search", "@modelcontextprotocol/server-brave-search", "Brave 搜索引擎", "搜索"},
	{"google-search", "@modelcontextprotocol/server-google-search", "Google 搜索引擎", "搜索"},
	{"puppeteer", "@modelcontextprotocol/server-puppeteer", "浏览器自动化 (截图、点击、表单)", "浏览器"},
	{"playwright", "@anthropic/mcp-server-playwright", "Playwright 浏览器自动化", "浏览器"},
	{"memory", "@modelcontextprotocol/server-memory", "持久化记忆存储", "工具"},
	{"fetch", "@anthropic/mcp-server-fetch", "HTTP 请求和网页抓取", "工具"},
	{"slack", "@anthropic/mcp-server-slack", "Slack 消息发送和频道管理", "协作"},
	{"notion", "@anthropic/mcp-server-notion", "Notion 文档读写", "协作"},
}
type tuiModel struct {
	cmdInput textinput.Model; ta textarea.Model; viewport viewport.Model
	chatMessages []chatMessage; messages []string
	history []string; historyIdx int
	rawStream string; streamLines []string; rawThinking string
	thinking bool; thinkStart time.Time; responding bool
	busy bool; busyStart time.Time; busyEnd time.Time
	compressing bool; compressStart time.Time
	planning bool; planStart time.Time
	quitting bool; cmdMatches []cmdEntry; cmdSelectIdx int
	currentTool string; toolStart time.Time
	currentPlan *agent.Plan; planActive bool
	resuming bool; resumeFiles []string; resumeIdx int
	themeSelecting bool; themeIdx int; themeOptions []string
	planSelecting bool; planIdx int; planOptions []string
	confirming bool; confirmAction safety.Action; confirmResult chan bool; confirmChoice int; confirmAll bool
	mcpWizard bool; mcpWizardStep int; mcpWizardData map[string]string; mcpWizardSelectIdx int; mcpWizardOptions []MCPRecommendedServer; mcpDirSelectIdx int; mcpDirOptions []string; mcpTypeSelectIdx int
	messageQueue []string; cancelPending bool; showToolCalls bool; userScrolledUp bool; spinnerFrame int
	msgStartTokens int; msgToolCalls int; msgTokens int; msgPromptTokens int; msgCompletionTokens int; totalUsage llm.Usage
	width, height int; welcomeShown bool; userName string; nameSetup bool
	mcpManager *mcp.Manager; agent *agent.Agent; bm *backup.Manager; streamChan chan tea.Msg
	modelName, appVersion, sessionId string; sessionStore *session.Store; ready bool
}
var inputAreaHeight = 5

// calcInputAreaHeight 根据 textarea 内容行数动态计算输入区域高度
// 选择模式（resume/theme/plan/confirm）使用固定高度 15，确保容器不被压扁
func (m *tuiModel) calcInputAreaHeight() int {
	if m.resuming || m.themeSelecting || m.planSelecting || m.confirming {
		return 15 // 边框2 + 内容区8 + 状态栏/分隔线/底部
	}
	lines := len(strings.Split(m.ta.Value(), "\n"))
	if lines < 1 { lines = 1 }
	if lines > 8 { lines = 8 }
	m.ta.SetHeight(lines)
	return lines + 4 // textarea + 状态栏 + 2个分隔线 + 底部提示
}

func newTUIModel(ag *agent.Agent, bm *backup.Manager, mcpMgr *mcp.Manager, modelName, appVersion, userName string, sessionStore *session.Store) tuiModel {
	ti := textinput.New(); ti.Placeholder = "选择时使用 ↑↓, Enter 确认"; ti.CharLimit = 0; ti.Width = 80; ti.Prompt = "❯ "
	ta := textarea.New(); ta.Placeholder = "输入消息... (Ctrl+J 换行, Enter 发送)"; ta.Focus(); ta.CharLimit = 0; ta.SetWidth(80); ta.SetHeight(1); ta.ShowLineNumbers = false; ta.KeyMap.InsertNewline.SetKeys("ctrl+j")
	vp := viewport.New(80, 20); vp.KeyMap = viewport.KeyMap{}; vp.MouseWheelEnabled = true
	return tuiModel{cmdInput: ti, ta: ta, viewport: vp, messages: make([]string, 0), history: loadHistory(),
		historyIdx: -1, cmdSelectIdx: -1, streamChan: make(chan tea.Msg, 256), agent: ag, bm: bm,
		mcpManager: mcpMgr, modelName: modelName, appVersion: appVersion, userName: userName, sessionStore: sessionStore, nameSetup: userName == "", showToolCalls: true, sessionId: time.Now().Format("20060102_150405")}
}
func (m tuiModel) Init() tea.Cmd { return tea.Batch(textarea.Blink, m.listenStream()) }
func (m tuiModel) listenStream() tea.Cmd { return func() tea.Msg { return <-m.streamChan } }
func (m tuiModel) tickSpinner() tea.Cmd { return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return spinnerTickMsg{} }) }
func (m *tuiModel) followBottom() { if !m.userScrolledUp { m.viewport.GotoBottom() } }
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height, m.ready = msg.Width, msg.Height, true
		if !m.welcomeShown { m.welcomeShown = true }
		m.cmdInput.Width = m.width - 3; m.ta.SetWidth(m.width - 3)
		vpHeight := m.height - m.calcInputAreaHeight(); if vpHeight < 5 { vpHeight = 5 }
		m.viewport.Width = m.width; m.viewport.Height = vpHeight
		m.rebuildMessages()
		if m.busy && len(m.streamLines) > 0 { m.messages = append(m.messages, m.streamLines...) }
		m.viewport.SetContent(m.renderMessages()); m.followBottom()
		// 确保 resize 后 stream 监听不中断（busy 时必须持续监听 channel）
		if m.busy || m.compressing || m.responding || m.thinking {
			return m, m.listenStream()
		}
		return m, nil
	case tea.KeyMsg: return m.handleKey(msg)
	case tea.MouseMsg: return m.handleMouse(msg)
	case thinkingMsg:
		if !m.thinking { m.thinking, m.thinkStart, m.busy = true, time.Now(), true }
		m.rawThinking += msg.delta
		m.viewport.SetContent(m.renderMessages()); m.followBottom()
		return m, m.listenStream()
	case planningMsg:
		if !m.planning { m.planning, m.planStart, m.busy = true, time.Now(), true }
		if m.rawThinking != "" { m.rawThinking += "\n" + msg.message } else { m.rawThinking = msg.message }
		m.viewport.SetContent(m.renderMessages()); m.followBottom()
		return m, m.listenStream()
	case deltaMsg:
		if m.thinking { m.thinking = false }
		m.responding, m.busy = true, true; m.rawStream += msg.text
		m.viewport.SetContent(m.renderMessages()); m.followBottom()
		return m, m.listenStream()
	case toolCallMsg:
		m.responding, m.currentTool, m.toolStart, m.msgToolCalls, m.planning = false, msg.name, time.Now(), m.msgToolCalls+1, false
		m.streamLines = append(m.streamLines, "")
		displayName := msg.name
		if strings.Contains(msg.name, "__") {
			parts := strings.SplitN(msg.name, "__", 2)
			displayName = toolMCPStyle.Render("[MCP:"+parts[0]+"]") + " " + toolNameStyle.Render(parts[1])
		} else {
			displayName = toolNameStyle.Render(msg.name)
		}
		m.streamLines = append(m.streamLines, statsDotStyle.Render("●")+" "+displayName)
		if msg.args != "" {
			var args map[string]interface{}
			if json.Unmarshal([]byte(msg.args), &args) == nil {
				if cmd, ok := args["command"].(string); ok { m.streamLines = append(m.streamLines, "    "+toolArgStyle.Render("command: ")+cmd) }
				if path, ok := args["path"].(string); ok { m.streamLines = append(m.streamLines, "    "+toolArgStyle.Render("path: ")+path) }
				if content, ok := args["content"].(string); ok && len(content) > 0 {
					m.streamLines = append(m.streamLines, "    "+toolArgStyle.Render("content ("+fmt.Sprintf("%d", len(content))+" chars):"))
					lines := strings.Split(content, "\n")
					maxLines := 20
					if len(lines) > maxLines {
						for i := 0; i < maxLines; i++ { m.streamLines = append(m.streamLines, "      "+thinkingContentStyle.Render(lines[i])) }
						m.streamLines = append(m.streamLines, "      "+thinkingContentStyle.Render("... ("+fmt.Sprintf("%d", len(lines)-maxLines)+" more lines)"))
					} else { for _, line := range lines { m.streamLines = append(m.streamLines, "      "+thinkingContentStyle.Render(line)) } }
				}
			}
		}
		m.viewport.SetContent(m.renderMessages()); m.followBottom()
		return m, m.listenStream()
	case toolResultMsg:
		m.currentTool = ""
		for i, l := range strings.Split(collapseToolResult(msg.result, 3, 100), "\n") {
			if i == 0 { m.streamLines = append(m.streamLines, "  ⎿  "+toolResultStyle.Render(l)) } else { m.streamLines = append(m.streamLines, "     "+toolResultStyle.Render(l)) }
		}
		m.viewport.SetContent(m.renderMessages()); m.followBottom()
		return m, m.listenStream()
	case agentDoneMsg:
		m.thinking, m.responding, m.busy, m.busyEnd, m.compressing, m.planning = false, false, false, time.Now(), false, false
		if m.cancelPending {
			m.cancelPending, m.streamLines, m.rawStream = false, nil, ""
			q := findLastUserMsg(m.chatMessages)
			if q != "" { m.messages = append(m.messages, errorStyle.Render("  「"+truncStr(q, 40)+"」已终止")) } else { m.messages = append(m.messages, errorStyle.Render("  已终止")) }
			m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
			if len(m.messageQueue) > 0 { savedInput := m.inputValue(); next := m.messageQueue[0]; m.messageQueue = m.messageQueue[1:]; m2, cmd := m.sendMessage(next); tm := m2.(tuiModel); tm.inputSetValue(savedInput); return tm, cmd }
			focusCmd := m.inputFocus(); return m, tea.Batch(focusCmd, m.listenStream())
		}
		m.finalizeResponse(msg.response); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
		if len(m.messageQueue) > 0 { savedInput := m.inputValue(); next := m.messageQueue[0]; m.messageQueue = m.messageQueue[1:]; m2, cmd := m.sendMessage(next); tm := m2.(tuiModel); tm.inputSetValue(savedInput); return tm, cmd }
		focusCmd := m.inputFocus(); return m, tea.Batch(focusCmd, m.listenStream())
	case agentErrMsg:
		m.thinking, m.responding, m.busy, m.busyEnd, m.compressing, m.planning = false, false, false, time.Now(), false, false
		m.streamLines, m.rawStream = nil, ""
		if m.cancelPending {
			m.cancelPending = false
			q := findLastUserMsg(m.chatMessages)
			if q != "" { m.messages = append(m.messages, errorStyle.Render("  「"+truncStr(q, 40)+"」已终止")) } else { m.messages = append(m.messages, errorStyle.Render("  已终止")) }
		} else { m.messages = append(m.messages, "  "+toolErrorStyle.Render("✖ "+msg.err.Error())) }
		m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
		if len(m.messageQueue) > 0 { savedInput := m.inputValue(); next := m.messageQueue[0]; m.messageQueue = m.messageQueue[1:]; m2, cmd := m.sendMessage(next); tm := m2.(tuiModel); tm.inputSetValue(savedInput); return tm, cmd }
		focusCmd := m.inputFocus(); return m, tea.Batch(focusCmd, m.listenStream())
	case usageMsg:
		d := msg.usage.TotalTokens - (m.msgStartTokens + m.msgTokens); if d > 0 { m.msgTokens += d }
		m.msgPromptTokens = msg.usage.PromptTokens; m.msgCompletionTokens = msg.usage.CompletionTokens
		m.totalUsage = msg.usage; m.viewport.SetContent(m.renderMessages())
		return m, m.listenStream()
	case compressingMsg:
		m.compressing, m.compressStart = true, time.Now()
		m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
		return m, m.listenStream()
	case compressDoneMsg:
		m.compressing, m.busy = false, false
		elapsed := time.Since(m.compressStart); saved := msg.before - msg.after; pct := 0.0
		if msg.before > 0 { pct = float64(saved) / float64(msg.before) * 100 }
		bar := strings.Repeat("█", 30) + " 100%"
		m.messages = append(m.messages, "")
		m.messages = append(m.messages, assistantPrefixStyle.Render("◈ 压缩完成"))
		m.messages = append(m.messages, confirmTextStyle.Render("  "+bar))
		m.messages = append(m.messages, confirmTextStyle.Render(fmt.Sprintf("  %s → %s  (-%s, %.1f%%)", formatTokens(msg.before), formatTokens(msg.after), formatTokens(saved), pct)))
		m.messages = append(m.messages, thinkingStyle.Render(fmt.Sprintf("  耗时 %s", formatDuration(elapsed))))
	case mcpInstallDoneMsg:
		if msg.success {
			m.messages = append(m.messages, confirmTextStyle.Render("  ✓ MCP 服务器配置完成"))
			m.messages = append(m.messages, helpHintStyle.Render("  重启后生效，使用 /mcp 查看状态"))
		} else {
			m.messages = append(m.messages, errorStyle.Render(fmt.Sprintf("  ✗ 配置失败: %v", msg.err)))
		}
		m.mcpWizard = false
		m.busy, m.responding, m.thinking = false, false, false
		m.inputReset(); focusCmd := m.inputFocus()
		m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
		return m, tea.Batch(focusCmd, m.listenStream())
	case confirmMsg:
		if m.confirmAll { msg.result <- true; return m, m.listenStream() }
		m.confirming, m.confirmAction, m.confirmResult, m.confirmChoice = true, msg.action, msg.result, 0
		m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
		return m, nil
	case planGeneratedMsg:
		m.currentPlan = msg.plan; m.planActive = true; m.planning = false
		m.messages = append(m.messages, "")
		m.messages = append(m.messages, confirmBorderStyle.Render("  执行计划已生成 ("+fmt.Sprintf("%d", msg.plan.TotalSteps)+" 步)"))
		m.messages = append(m.messages, thinkingStyle.Render("  "+msg.plan.FormatPlan()))
		m.messages = append(m.messages, "")
		m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
		return m, m.listenStream()
	case planStepStartMsg:
		m.messages = append(m.messages, confirmTextStyle.Render(fmt.Sprintf("  执行步骤 %d: %s", msg.step.ID, msg.step.Description)))
		m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
		return m, m.listenStream()
	case planStepDoneMsg:
		if msg.step.Status == agent.StepCompleted { m.messages = append(m.messages, confirmTextStyle.Render(fmt.Sprintf("  步骤 %d 完成", msg.step.ID)))
		} else if msg.step.Status == agent.StepFailed { m.messages = append(m.messages, errorStyle.Render(fmt.Sprintf("  步骤 %d 失败: %s", msg.step.ID, msg.step.Error))) }
		if m.currentPlan != nil { m.messages = append(m.messages, thinkingStyle.Render("  "+m.currentPlan.FormatProgress())) }
		m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
		return m, m.listenStream()
	case spinnerTickMsg:
		if m.busy { m.spinnerFrame++; m.viewport.SetContent(m.renderMessages()); m.followBottom(); return m, m.tickSpinner() }
		return m, nil
	}
	return m, tea.Batch(cmds...)
}

func findLastUserMsg(chatMessages []chatMessage) string {
	for i := len(chatMessages) - 1; i >= 0; i-- { if chatMessages[i].role == "user" { return chatMessages[i].content } }
	return ""
}
func (m tuiModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mcpWizard { return m.handleMCPWizardKey(msg) }
	if m.themeSelecting { return m.handleThemeKey(msg) }
	if m.planSelecting { return m.handlePlanKey(msg) }
	if m.resuming { return m.handleResumeKey(msg) }
	if m.confirming { return m.handleConfirmKey(msg) }
	if m.nameSetup {
		if msg.String() == "enter" {
			name := strings.TrimSpace(m.inputValue()); if name == "" { return m, nil }
			m.userName, m.nameSetup = name, false; m.inputReset()
			cfg.UserName = name; go config.SaveUserConfig(cfg)
			m.viewport.SetContent(m.renderMessages()); return m, nil
		}
		if msg.String() == "ctrl+c" { saveHistory(m.history); return m, tea.Quit }
		var cmd tea.Cmd; m.cmdInput, cmd = m.cmdInput.Update(msg); return m, cmd
	}
	if m.quitting {
		switch msg.String() { case "ctrl+c", "esc", "y", "Y": saveHistory(m.history); return m, tea.Quit; default: m.quitting = false; m.viewport.SetContent(m.renderMessages()); return m, nil }
	}
	switch msg.String() {
	case "ctrl+c", "esc":
		if m.isCmdMode() && msg.String() == "esc" { m.cmdMatches, m.cmdSelectIdx = nil, -1; m.viewport.SetContent(m.renderMessages()); return m, nil }
		if m.busy { m.agent.Cancel(); m.busy, m.thinking, m.responding, m.rawStream, m.streamLines, m.messageQueue, m.cancelPending = false, false, false, "", nil, nil, true; m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil }
		if m.quitting { saveHistory(m.history); return m, tea.Quit }
		m.quitting = true; m.viewport.SetContent(m.renderMessages()); return m, nil
	case "ctrl+l": m.messages = nil; m.viewport.SetContent(""); return m, nil
	case "ctrl+t":
		m.showToolCalls = !m.showToolCalls
		m.rebuildMessages()
		m.viewport.SetContent(m.renderMessages())
		return m, nil
	case "tab":
		if m.isCmdMode() && len(m.cmdMatches) > 0 {
			if m.cmdSelectIdx < 0 { m.cmdSelectIdx = 0 } else { m.cmdSelectIdx = (m.cmdSelectIdx + 1) % len(m.cmdMatches) }
		}
		m.viewport.SetContent(m.renderMessages()); return m, nil
	case "enter":
		if m.cmdSelectIdx >= 0 && m.cmdSelectIdx < len(m.cmdMatches) { m.acceptCmdSuggestion(); m.viewport.SetContent(m.renderMessages()); return m, nil }
		if len(m.cmdMatches) == 1 { m.acceptCmdSuggestion(); m.viewport.SetContent(m.renderMessages()); return m, nil }
		m.cmdMatches, m.cmdSelectIdx = nil, -1
		input := strings.TrimSpace(m.ta.Value())
		if input == "" { return m, nil }
		if input == "exit" || input == "quit" { saveHistory(m.history); return m, tea.Quit }
		if m.busy {
			if !strings.HasPrefix(input, "/") { m.messageQueue = append(m.messageQueue, input); m.chatMessages = append(m.chatMessages, chatMessage{role: "user", content: input}) }
			if len(m.history) == 0 || m.history[len(m.history)-1] != input { m.history = append(m.history, input); if len(m.history) > 100 { m.history = m.history[len(m.history)-100:] } }
			m.historyIdx = -1; m.ta.Reset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil
		}
		return m.sendMessage(input)
	case "up":
		if m.isCmdMode() { if len(m.cmdMatches) > 0 { if m.cmdSelectIdx <= 0 { m.cmdSelectIdx = len(m.cmdMatches) - 1 } else { m.cmdSelectIdx-- } }; m.viewport.SetContent(m.renderMessages()); return m, nil }
		if len(m.history) > 0 { if m.historyIdx < 0 { m.historyIdx = len(m.history) - 1 } else if m.historyIdx > 0 { m.historyIdx-- }; m.ta.SetValue(m.history[m.historyIdx]) }
		return m, nil
	case "down":
		if m.isCmdMode() { if len(m.cmdMatches) > 0 { if m.cmdSelectIdx < 0 { m.cmdSelectIdx = 0 } else { m.cmdSelectIdx = (m.cmdSelectIdx + 1) % len(m.cmdMatches) } }; m.viewport.SetContent(m.renderMessages()); return m, nil }
		if m.historyIdx >= 0 { m.historyIdx++; if m.historyIdx >= len(m.history) { m.historyIdx = -1; m.ta.Reset() } else { m.ta.SetValue(m.history[m.historyIdx]) } }
		return m, nil
	case "pgup": m.viewport.HalfViewUp(); m.userScrolledUp = true; return m, nil
	case "pgdown": m.viewport.HalfViewDown(); if m.viewport.AtBottom() { m.userScrolledUp = false }; return m, nil
	default:
		if m.isCmdMode() { m.cmdMatches, m.cmdSelectIdx = nil, -1 }
		var cmd tea.Cmd; m.ta, cmd = m.ta.Update(msg); m.updateCmdSuggestions()
		vpH := m.height - m.calcInputAreaHeight(); if vpH < 5 { vpH = 5 }; m.viewport.Height = vpH
		return m, cmd
	}
}








func (m tuiModel) handleResumeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k": if m.resumeIdx > 0 { m.resumeIdx-- } else { m.resumeIdx = len(m.resumeFiles) - 1 }
	case "down", "j": if m.resumeIdx < len(m.resumeFiles)-1 { m.resumeIdx++ } else { m.resumeIdx = 0 }
	case "enter": return m.loadResumeSession(m.resumeFiles[m.resumeIdx])
	case "esc": m.resuming = false; m.resumeFiles = nil; vpH := m.height - m.calcInputAreaHeight(); if vpH < 5 { vpH = 5 }; m.viewport.Height = vpH; m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); focusCmd := m.inputFocus(); return m, focusCmd
	}
	m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil
}
func (m tuiModel) handleThemeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k": if m.themeIdx > 0 { m.themeIdx-- } else { m.themeIdx = len(m.themeOptions) - 1 }
	case "down", "j": if m.themeIdx < len(m.themeOptions)-1 { m.themeIdx++ } else { m.themeIdx = 0 }
	case "enter": SetTheme(m.themeOptions[m.themeIdx]); cfg.Theme = m.themeOptions[m.themeIdx]; go config.SaveUserConfig(cfg); m.themeSelecting = false; m.messages = append(m.messages, confirmTextStyle.Render("  主题已切换为: "+m.themeOptions[m.themeIdx]))
	case "esc": m.themeSelecting = false
	}
	m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
	// 关键：只有退出选择模式（esc/enter）时才重新聚焦输入框
	// up/down 导航时 textarea 不可见，不能调 inputFocus()
	if !m.themeSelecting { focusCmd := m.inputFocus(); return m, focusCmd }
	return m, nil
}
func (m tuiModel) handlePlanKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k": if m.planIdx > 0 { m.planIdx-- } else { m.planIdx = len(m.planOptions) - 1 }
	case "down", "j": if m.planIdx < len(m.planOptions)-1 { m.planIdx++ } else { m.planIdx = 0 }
	case "enter":
		newMode := m.planOptions[m.planIdx]; cfg.Agent.PlanningMode = newMode; go config.SaveUserConfig(cfg)
		switch newMode { case "react": m.agent.SetPlanningMode(agent.ModeReact); case "plan-execute": m.agent.SetPlanningMode(agent.ModePlanExecute); case "auto": m.agent.SetPlanningMode(agent.ModeAuto) }
		m.planSelecting = false; m.messages = append(m.messages, confirmTextStyle.Render("  已切换到: "+newMode+" 模式"))
	case "esc": m.planSelecting = false
	}
	m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
	// 关键：只有退出选择模式时才重新聚焦输入框
	if !m.planSelecting { focusCmd := m.inputFocus(); return m, focusCmd }
	return m, nil
}
func (m tuiModel) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k": if m.confirmChoice > 0 { m.confirmChoice-- }
	case "down", "j": if m.confirmChoice < 2 { m.confirmChoice++ }
	case "y", "Y": m.confirmResult <- true; m.confirming = false
	case "n", "N": m.confirmResult <- false; m.confirming = false
	case "a", "A": m.confirmAll = true; m.agent.SetConfirmAll(true); m.confirmResult <- true; m.confirming = false
	case "enter":
		if m.confirmChoice == 0 { m.confirmResult <- true } else if m.confirmChoice == 1 { m.confirmResult <- false } else { m.confirmAll = true; m.agent.SetConfirmAll(true); m.confirmResult <- true }
		m.confirming = false
	case "esc": m.confirmResult <- false; m.confirming = false
	}
	m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, m.listenStream()
}
func (m tuiModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Button {
	case tea.MouseButtonWheelUp: m.viewport.HalfViewUp(); m.userScrolledUp = true
	case tea.MouseButtonWheelDown: m.viewport.HalfViewDown(); if m.viewport.AtBottom() { m.userScrolledUp = false }
	}
	return m, nil
}

func (m tuiModel) sendMessage(text string) (tea.Model, tea.Cmd) {
	if m.mcpWizard {
		m.chatMessages = append(m.chatMessages, chatMessage{role: "user", content: text})
		m.messages = append(m.messages, userPrefixStyle.Render("❯ "+m.userName))
		renderUserText(text, &m.messages, m.width)
		m.handleMCPWizardInput(text)
		m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
		return m, nil
	}
	if len(m.history) == 0 || m.history[len(m.history)-1] != text { m.history = append(m.history, text); if len(m.history) > 100 { m.history = m.history[len(m.history)-100:] } }
	if strings.HasPrefix(text, "/") { return m.handleCommand(text) }
	m.historyIdx = -1
	m.chatMessages = append(m.chatMessages, chatMessage{role: "user", content: text})
	m.messages = append(m.messages, userPrefixStyle.Render("❯ "+m.userName))
	renderUserText(text, &m.messages, m.width)
	m.messages = append(m.messages, "", "")
	m.ta.Reset(); m.busy, m.busyStart = true, time.Now(); m.userScrolledUp = false
	m.msgStartTokens, m.msgToolCalls, m.msgTokens, m.msgPromptTokens, m.msgCompletionTokens = m.totalUsage.TotalTokens, 0, 0, 0, 0
	m.rawStream, m.streamLines, m.rawThinking = "", nil, ""
	m.currentPlan = nil; m.planActive = false
	m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
	go m.runAgent(text)
	return m, tea.Batch(m.listenStream(), m.tickSpinner())
}

func (m tuiModel) handleCommand(text string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(text); cmd := parts[0]
	switch cmd {
	case "/copy":
		var textToCopy string; arg := ""; if len(parts) > 1 { arg = parts[1] }
		if arg == "all" {
			var all []string; for _, msg := range m.chatMessages { if msg.role == "assistant" && msg.content != "" { all = append(all, msg.content) } }
			if len(all) == 0 { m.messages = append(m.messages, errorStyle.Render("  没有可复制的内容")) } else { textToCopy = strings.Join(all, "\n\n---\n\n") }
		} else if arg == "user" {
			for i := len(m.chatMessages) - 1; i >= 0; i-- { if m.chatMessages[i].role == "user" { textToCopy = m.chatMessages[i].content; break } }
			if textToCopy == "" { m.messages = append(m.messages, errorStyle.Render("  没有可复制的内容")) }
		} else if arg != "" {
			n := 0; fmt.Sscanf(arg, "%d", &n)
			if n <= 0 { m.messages = append(m.messages, errorStyle.Render("  用法: /copy [N|all|user]")) } else {
				count := 0; for _, msg := range m.chatMessages { if msg.role == "assistant" && msg.content != "" { count++; if count == n { textToCopy = msg.content; break } } }
				if textToCopy == "" { m.messages = append(m.messages, errorStyle.Render(fmt.Sprintf("  只有 %d 条 assistant 消息", count))) }
			}
		} else {
			for i := len(m.chatMessages) - 1; i >= 0; i-- { if m.chatMessages[i].role == "assistant" && m.chatMessages[i].content != "" { textToCopy = m.chatMessages[i].content; break } }
			if textToCopy == "" { m.messages = append(m.messages, errorStyle.Render("  没有可复制的内容")) }
		}
		if textToCopy != "" { if err := copyToClipboard(textToCopy); err != nil { m.messages = append(m.messages, errorStyle.Render("  复制失败: "+err.Error())) } else { m.messages = append(m.messages, confirmTextStyle.Render("  已复制到剪贴板")) } }
		m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil
	case "/help":
		m.messages = append(m.messages, "", confirmBorderStyle.Render("  可用命令:"),
			confirmTextStyle.Render("  /copy      - 复制响应到剪贴板（/copy N/all/user）"),
			confirmTextStyle.Render("  /compress  - 压缩上下文，释放 token 空间"),
			confirmTextStyle.Render("  /rollback  - 回滚文件到备份（/rollback N）"),
			confirmTextStyle.Render("  /clear     - 清空屏幕"), confirmTextStyle.Render("  /help      - 显示此帮助"),
			confirmTextStyle.Render("  /name      - 修改昵称（/name 昵称）"), confirmTextStyle.Render("  /export    - 导出对话记录"),
			confirmTextStyle.Render("  /model     - 查看/切换模型（/model 名称）"), confirmTextStyle.Render("  /resume    - 恢复历史会话"),
			confirmTextStyle.Render("  /theme     - 切换主题（/theme dark/light）"), confirmTextStyle.Render("  /plan      - 切换规划模式（/plan react|plan-execute|auto）"),
			confirmTextStyle.Render("  /mcp       - MCP 管理 (add/remove/status)"), confirmTextStyle.Render("  /exit      - 退出程序"), "")
		m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil
	case "/mcp":
		if len(parts) >= 2 && parts[1] == "add" {
			m.mcpWizard = true; m.mcpWizardStep = 0; m.mcpWizardData = make(map[string]string)
			m.messages = append(m.messages, "", confirmBorderStyle.Render("  ═══ MCP 服务器添加向导 ═══"), "")
			m.messages = append(m.messages, confirmTextStyle.Render("  步骤 1/4: 请输入服务器名称"))
			m.messages = append(m.messages, helpHintStyle.Render("  直接回车使用默认名，或输入自定义名称"), "")
		} else if len(parts) >= 2 && parts[1] == "remove" && len(parts) >= 3 { m.removeMCPServer(parts[2]) } else { m.showMCPStatus() }
		m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil
	case "/plan":
		if len(parts) >= 2 {
			newMode := parts[1]; modes := []string{"react", "plan-execute", "auto"}; valid := false
			for _, mm := range modes { if mm == newMode { valid = true; break } }
			if !valid { m.messages = append(m.messages, errorStyle.Render("  无效模式: "+newMode)) } else {
				cfg.Agent.PlanningMode = newMode; go config.SaveUserConfig(cfg)
				switch newMode { case "react": m.agent.SetPlanningMode(agent.ModeReact); case "plan-execute": m.agent.SetPlanningMode(agent.ModePlanExecute); case "auto": m.agent.SetPlanningMode(agent.ModeAuto) }
				m.messages = append(m.messages, confirmTextStyle.Render("  已切换到: "+newMode+" 模式"))
			}
		} else { m.planOptions = []string{"react", "plan-execute", "auto"}; m.planSelecting = true; m.planIdx = 0; for i, mode := range m.planOptions { if mode == cfg.Agent.PlanningMode { m.planIdx = i; break } } }
		m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil
	case "/clear": m.messages, m.chatMessages = nil, nil; m.inputReset(); m.viewport.SetContent(""); return m, nil
	case "/rollback":
		if m.bm == nil { m.messages = append(m.messages, errorStyle.Render("  备份管理未启用")) } else {
			backups, err := m.bm.List()
			if err != nil { m.messages = append(m.messages, errorStyle.Render("  读取备份失败: "+err.Error())) } else if len(backups) == 0 { m.messages = append(m.messages, confirmTextStyle.Render("  没有可用的备份")) } else if len(parts) >= 2 {
				n := 0; fmt.Sscanf(parts[1], "%d", &n)
				if n > 0 && n <= len(backups) { t := backups[n-1]; if err := m.bm.Restore(t.ID); err != nil { m.messages = append(m.messages, errorStyle.Render("  回滚失败: "+err.Error())) } else { m.messages = append(m.messages, confirmTextStyle.Render(fmt.Sprintf("  ✓ 已回滚: %s", t.OrigPath))) } } else { m.messages = append(m.messages, errorStyle.Render(fmt.Sprintf("  无效序号，可选范围: 1-%d", len(backups)))) }
			} else {
				m.messages = append(m.messages, "", confirmBorderStyle.Render("  可用备份:"))
				for i, bk := range backups { m.messages = append(m.messages, confirmTextStyle.Render(fmt.Sprintf("  [%d] %s  %s  → %s", i+1, bk.Timestamp.Format("01-02 15:04:05"), bk.Description, filepath.Base(bk.OrigPath)))) }
				m.messages = append(m.messages, "", helpHintStyle.Render("  使用 /rollback N 回滚第 N 个备份"))
			}
		}
		m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil
	case "/confirm": m.confirmAll = false; m.agent.SetConfirmAll(false); m.messages = append(m.messages, confirmTextStyle.Render("  已重新启用安全确认")); m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil
	case "/exit", "/quit": saveHistory(m.history); return m, tea.Quit
	case "/compress":
		if m.busy { m.messages = append(m.messages, errorStyle.Render("  请等待当前操作完成")) } else if m.agent.GetMessageCount() <= 3 { m.messages = append(m.messages, confirmTextStyle.Render("  对话太短，无需压缩")) } else { m.busy, m.compressing = true, true; go m.runCompress(); return m, tea.Batch(m.listenStream(), m.tickSpinner()) }
		m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil
	case "/name":
		if len(parts) < 2 { m.messages = append(m.messages, errorStyle.Render("  用法: /name <昵称>")) } else { m.userName = strings.Join(parts[1:], " "); cfg.UserName = m.userName; go config.SaveUserConfig(cfg); m.messages = append(m.messages, confirmTextStyle.Render("  昵称已改为: "+m.userName)) }
		m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil
	case "/model":
		models := cfg.Models
		if len(parts) < 2 {
			m.messages = append(m.messages, confirmBorderStyle.Render("  可用模型:"))
			for name, mc := range models { marker := "  "; if name == cfg.DefaultModel { marker = "  ▶ " }; m.messages = append(m.messages, confirmTextStyle.Render(marker+name+" ("+mc.Model+")")) }
			m.messages = append(m.messages, "", helpHintStyle.Render("  使用 /model <名称> 切换"))
		} else {
			modelName := parts[1]
			if _, ok := models[modelName]; !ok { m.messages = append(m.messages, errorStyle.Render("  模型不存在: "+modelName)) } else { cfg.DefaultModel = modelName; go config.SaveUserConfig(cfg); m.messages = append(m.messages, confirmTextStyle.Render("  已切换到: "+modelName)) }
		}
		m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil
	case "/export":
		if len(m.chatMessages) == 0 { m.messages = append(m.messages, errorStyle.Render("  没有可导出的对话")) } else {
			home, _ := os.UserHomeDir(); dir := filepath.Join(home, ".mimo", "exports"); os.MkdirAll(dir, 0755)
			fpath := filepath.Join(dir, fmt.Sprintf("export_%s.md", time.Now().Format("20060102_150405")))
			var sb strings.Builder; sb.WriteString("# MiMo CLI 对话导出\n\n"); sb.WriteString(fmt.Sprintf("- 时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))); sb.WriteString(fmt.Sprintf("- 模型: %s\n\n", m.modelName))
			for _, msg := range m.chatMessages { switch msg.role { case "user": sb.WriteString(fmt.Sprintf("## ❯ %s\n\n%s\n\n", m.userName, msg.content)); case "assistant": if msg.content != "" { sb.WriteString(fmt.Sprintf("## ◈ MiMo\n\n%s\n\n", msg.content)) }; if msg.tokens > 0 { sb.WriteString(fmt.Sprintf("> tokens: %s · time: %s · tools: %d\n\n", formatTokens(msg.tokens), formatDuration(msg.duration), msg.toolCalls)) } } }
			os.WriteFile(fpath, []byte(sb.String()), 0644)
			m.messages = append(m.messages, confirmTextStyle.Render("  已导出到: "+fpath))
		}
		m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil
	case "/resume":
		if m.sessionStore == nil { m.messages = append(m.messages, errorStyle.Render("  会话存储未启用")); m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil }
		sessions, err := m.sessionStore.ListSessions(20)
		if err != nil || len(sessions) == 0 { m.messages = append(m.messages, confirmTextStyle.Render("  没有可恢复的会话")); m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil }
		m.resuming = true; m.resumeFiles = nil; for _, s := range sessions { m.resumeFiles = append(m.resumeFiles, s.ID) }; m.resumeIdx = 0
		m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil
	case "/theme":
		m.themeOptions = []string{"dark", "light"}; m.themeSelecting = true; m.themeIdx = 0
		for i, t := range m.themeOptions { if t == GetTheme() { m.themeIdx = i; break } }
		m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil
	default: return m, nil
	}
}

func (m *tuiModel) finalizeResponse(response string) {
	elapsed := m.busyEnd.Sub(m.busyStart); tu := m.msgTokens; if tu < 0 { tu = 0 }; if tu == 0 && response != "" { tu = len(response) / 2 }
	allLines := make([]string, 0)
	allLines = append(allLines, assistantPrefixStyle.Render("◈ MiMo"))
	allLines = append(allLines, m.streamLines...)
	if response != "" { for _, line := range strings.Split(renderMarkdownSafe(response, m.width), "\n") { allLines = append(allLines, "  "+line) } } else if len(m.streamLines) == 0 { allLines = append(allLines, "  "+thinkingStyle.Render("(no response)")) }
	allLines = append(allLines, "")
	var sp []string
	displayTokens := tu
	if m.msgCompletionTokens > 0 { displayTokens = m.msgCompletionTokens }
	if displayTokens <= 0 && response != "" { displayTokens = len(response) / 2 }
	sp = append(sp, tokenLabelStyle.Render("tokens: "+formatTokens(displayTokens)))
	sp = append(sp, tokenLabelStyle.Render("time: "+formatDuration(elapsed)))
	if m.msgToolCalls > 0 { sp = append(sp, tokenLabelStyle.Render("tools: "+fmt.Sprintf("%d", m.msgToolCalls))) }
	allLines = append(allLines, "  "+statsDotStyle.Render("●")+" "+strings.Join(sp, tokenLabelStyle.Render(" · ")))
	m.messages = append(m.messages, allLines...)
	m.messages = append(m.messages, "", "")
	m.chatMessages = append(m.chatMessages, chatMessage{role: "assistant", content: response, tokens: tu, duration: elapsed, toolCalls: m.msgToolCalls, thinking: m.rawThinking, toolLines: m.streamLines})
	m.streamLines, m.rawStream = nil, ""
}
func (m tuiModel) loadResumeSession(sessionID string) (tea.Model, tea.Cmd) {
	if m.sessionStore == nil { m.messages = append(m.messages, errorStyle.Render("  会话存储未启用")); m.resuming = false; m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil }
	sess, msgs, err := m.sessionStore.LoadSession(sessionID)
	if err != nil { m.messages = append(m.messages, errorStyle.Render("  读取会话失败: "+err.Error())); m.resuming = false; m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); return m, nil }
	m.chatMessages = make([]chatMessage, len(msgs))
	for i, msg := range msgs { m.chatMessages[i] = chatMessage{role: msg.Role, content: msg.Content, tokens: msg.Tokens, toolCalls: msg.ToolCalls, duration: time.Duration(msg.DurationMs) * time.Millisecond, thinking: msg.Thinking, toolLines: msg.ToolLines} }
	if sess.UserName != "" { m.userName = sess.UserName }
	var agentMsgs []llm.Message
	for _, cm := range m.chatMessages { switch cm.role { case "user": agentMsgs = append(agentMsgs, llm.Message{Role: llm.RoleUser, Content: cm.content}); case "assistant": agentMsgs = append(agentMsgs, llm.Message{Role: llm.RoleAssistant, Content: cm.content}) } }
	m.agent.LoadMessages(agentMsgs)
	m.resuming = false; m.resumeFiles = nil; m.sessionId = sessionID
	m.rebuildMessages()
	m.ta.Reset(); m.ta.SetHeight(1)
	vpH := m.height - m.calcInputAreaHeight(); if vpH < 5 { vpH = 5 }; m.viewport.Height = vpH; m.viewport.Width = m.width
	m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom(); focusCmd := m.inputFocus(); return m, focusCmd
}
func (m *tuiModel) rebuildMessages() {
	m.messages = make([]string, 0)
	if m.welcomeShown { m.messages = append(m.messages, renderWelcomeBanner(m.appVersion, m.modelName, m.width)...) }
	for _, msg := range m.chatMessages {
		switch msg.role {
		case "user":
			m.messages = append(m.messages, userPrefixStyle.Render("❯ "+m.userName))
			renderUserText(msg.content, &m.messages, m.width)
			m.messages = append(m.messages, "", "")
		case "assistant":
			m.messages = append(m.messages, assistantPrefixStyle.Render("◈ MiMo"))
			if msg.thinking != "" { m.messages = append(m.messages, thinkingStyle.Render("  [思考过程]")); for _, l := range strings.Split(msg.thinking, "\n") { if strings.TrimSpace(l) != "" { m.messages = append(m.messages, thinkingContentStyle.Render("    "+l)) } }; m.messages = append(m.messages, "") }
 			if msg.content != "" { for _, line := range strings.Split(renderMarkdownSafe(msg.content, m.width), "\n") { m.messages = append(m.messages, "  "+line) } }
 			if m.showToolCalls && len(msg.toolLines) > 0 { m.messages = append(m.messages, msg.toolLines...) }
			{
				var parts []string; tok := msg.tokens; if tok <= 0 && msg.content != "" { tok = len(msg.content) / 2 }
				parts = append(parts, tokenLabelStyle.Render("tokens: "+formatTokens(tok)))
				parts = append(parts, tokenLabelStyle.Render("time: "+formatDuration(msg.duration)))
				if msg.toolCalls > 0 { parts = append(parts, tokenLabelStyle.Render("tools: "+fmt.Sprintf("%d", msg.toolCalls))) }
				m.messages = append(m.messages, "", "  "+statsDotStyle.Render("●")+" "+strings.Join(parts, tokenLabelStyle.Render(" · ")))
			}
			m.messages = append(m.messages, "", "")
		}
	}
}
func (m *tuiModel) runAgent(input string) {
	ctx := context.Background()
	m.agent.SetPlanningCallback(func(message string) { m.streamChan <- planningMsg{message: message} })
	m.agent.SetPlanGeneratedCallback(func(plan *agent.Plan) { m.streamChan <- planGeneratedMsg{plan: plan} })
	m.agent.SetPlanStepStartCallback(func(step *agent.PlanStep) { m.streamChan <- planStepStartMsg{step: step} })
	m.agent.SetPlanStepDoneCallback(func(step *agent.PlanStep) { m.streamChan <- planStepDoneMsg{step: step} })
	response, err := m.agent.ChatStream(ctx, input)
	if err != nil { m.streamChan <- agentErrMsg{err: err}; return }
	m.streamChan <- agentDoneMsg{response: response}
}
func (m *tuiModel) runCompress() {
	ctx := context.Background(); before, after, err := m.agent.CompressContext(ctx)
	if err != nil { m.streamChan <- agentErrMsg{err: err}; return }
	m.streamChan <- compressDoneMsg{before: before, after: after}
}
func (m *tuiModel) updateCmdSuggestions() {
	input := m.inputValue(); if !strings.HasPrefix(input, "/") { m.cmdMatches, m.cmdSelectIdx = nil, -1; return }
	lower := strings.ToLower(input); m.cmdMatches = nil
	for _, c := range commands { if strings.HasPrefix(c.name, lower) { m.cmdMatches = append(m.cmdMatches, c) } }
	if len(m.cmdMatches) == 0 { m.cmdSelectIdx = -1 } else if m.cmdSelectIdx < 0 || m.cmdSelectIdx >= len(m.cmdMatches) { m.cmdSelectIdx = 0 }
}
func (m *tuiModel) acceptCmdSuggestion() {
	idx := m.cmdSelectIdx; if idx < 0 && len(m.cmdMatches) > 0 { idx = 0 }
	if idx >= 0 && idx < len(m.cmdMatches) { m.inputSetValue(m.cmdMatches[idx].name + " ") }
	m.cmdMatches, m.cmdSelectIdx = nil, -1
}
func (m tuiModel) isCmdMode() bool { return len(m.cmdMatches) > 0 }

func printSessionHistory(m tuiModel) {
	if len(m.chatMessages) == 0 { return }
	fmt.Println("\n  ── 会话记录 ──\n")
	for _, msg := range m.chatMessages {
		switch msg.role {
		case "user": fmt.Printf("  ❯ %s: %s\n", m.userName, msg.content)
		case "assistant":
			if msg.content != "" { for _, line := range strings.Split(renderMarkdownSimple(msg.content, 80), "\n") { fmt.Printf("  %s\n", line) } }
			if msg.tokens > 0 || msg.duration > 0 { parts := []string{}; if msg.tokens > 0 { parts = append(parts, fmt.Sprintf("tokens: %s", formatTokens(msg.tokens))) }; if msg.duration > 0 { parts = append(parts, fmt.Sprintf("time: %s", formatDuration(msg.duration))) }; if msg.toolCalls > 0 { parts = append(parts, fmt.Sprintf("tools: %d", msg.toolCalls)) }; fmt.Printf("    ── %s\n", strings.Join(parts, " · ")) }
			fmt.Println()
		}
	}
	saveSessionToFile(m)
}
func saveSessionToFile(m tuiModel) {
	if len(m.chatMessages) == 0 { return }
	if m.sessionStore != nil {
		msgs := make([]session.Message, len(m.chatMessages))
		for i, cm := range m.chatMessages { msgs[i] = session.Message{Role: cm.role, Content: cm.content, Tokens: cm.tokens, ToolCalls: cm.toolCalls, DurationMs: cm.duration.Milliseconds(), CreatedAt: time.Now(), Thinking: cm.thinking, ToolLines: cm.toolLines} }
		wd, _ := os.Getwd()
		m.sessionStore.SaveSession(m.sessionId, m.modelName, m.userName, wd, msgs)
	}
	sessionDir := ".mimo/sessions"; if home, err := os.UserHomeDir(); err == nil { sessionDir = filepath.Join(home, ".mimo", "sessions") }
	os.MkdirAll(sessionDir, 0755)
	path := filepath.Join(sessionDir, fmt.Sprintf("session_%s.md", m.sessionId))
	var sb strings.Builder
	sb.WriteString("# MiMo CLI 会话记录\n\n" + fmt.Sprintf("- 时间: %s\n", time.Now().Format("2006-01-02 15:04:05")) + fmt.Sprintf("- 模型: %s\n\n", m.modelName))
	for _, msg := range m.chatMessages { switch msg.role { case "user": sb.WriteString(fmt.Sprintf("## ❯ %s\n\n%s\n\n", m.userName, msg.content)); case "assistant": if msg.content != "" { sb.WriteString(fmt.Sprintf("## ◈ MiMo\n\n%s\n\n", msg.content)) }; if msg.tokens > 0 { sb.WriteString(fmt.Sprintf("> tokens: %s · time: %s · tools: %d\n\n", formatTokens(msg.tokens), formatDuration(msg.duration), msg.toolCalls)) } } }
	os.WriteFile(path, []byte(sb.String()), 0644)
}
func loadHistory() []string {
	home, err := os.UserHomeDir(); if err != nil { return make([]string, 0) }
	path := filepath.Join(home, ".mimo", "history"); data, err := os.ReadFile(path); if err != nil { return make([]string, 0) }
	lines := strings.Split(strings.TrimSpace(string(data)), "\n"); result := make([]string, 0, len(lines))
	for _, l := range lines { if strings.TrimSpace(l) != "" { result = append(result, l) } }
	if len(result) > 100 { result = result[len(result)-100:] }
	return result
}
func saveHistory(history []string) {
	home, err := os.UserHomeDir(); if err != nil { return }; dir := filepath.Join(home, ".mimo"); os.MkdirAll(dir, 0755)
	if len(history) > 100 { history = history[len(history)-100:] }
	os.WriteFile(filepath.Join(dir, "history"), []byte(strings.Join(history, "\n")), 0644)
}
func getSpinnerChar() string { spinners := []string{"⠋","⠙","⠹","⠸","⠼","⠴","⠦","⠧","⠇","⠏"}; return spinners[int(time.Now().UnixMilli()/100)%len(spinners)] }
func formatToolCall(name, args string) (display, arg string) {
	nameMap := map[string]string{"file_read":"Read","file_write":"Write","file_edit":"Update","shell":"Bash"}; display = name; if v, ok := nameMap[name]; ok { display = v }
	argKeyMap := map[string]string{"file_read":"path","file_write":"path","file_edit":"path","shell":"command"}
	if args != "" { var m map[string]interface{}; if json.Unmarshal([]byte(args), &m) == nil { if k, ok := argKeyMap[name]; ok { if v, ok := m[k]; ok { arg = fmt.Sprintf("%v", v) } } } }
	return
}
func collapseToolResult(result string, maxLines, maxLineLen int) string {
	if result == "" { return "" }; lines := strings.Split(result, "\n"); total := len(lines)
	for total > 0 && strings.TrimSpace(lines[total-1]) == "" { lines = lines[:total-1]; total-- }; if total == 0 { return "" }
	for i, l := range lines { runes := []rune(l); if len(runes) > maxLineLen { lines[i] = string(runes[:maxLineLen]) + "..." } }
	if total <= maxLines { return strings.Join(lines, "\n") }; return strings.Join(lines[:maxLines], "\n") + fmt.Sprintf("\n  ⎿  ... +%d lines", total-maxLines)
}
func renderMarkdownSafe(content string, width int) string { if width < 20 { width = 80 }; return renderMarkdownSimple(content, width) }
func copyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows": cmd = exec.Command("powershell", "-NoProfile", "-Command", "$input | Set-Clipboard"); cmd.Stdin = strings.NewReader(text)
	case "darwin": cmd = exec.Command("pbcopy"); cmd.Stdin = strings.NewReader(text)
	default: cmd = exec.Command("xclip", "-selection", "clipboard"); cmd.Stdin = strings.NewReader(text)
	}
	return cmd.Run()
}

func renderUserText(text string, msgs *[]string, width int) {
	maxW := width - 4
	if maxW < 20 { maxW = 80 }
	for _, line := range strings.Split(text, "\n") {
		if displayWidth(line) > maxW {
			for _, wl := range wrapText(line, maxW) {
				*msgs = append(*msgs, userTextStyle.Render("  "+wl))
			}
		} else {
			*msgs = append(*msgs, userTextStyle.Render("  "+line))
		}
	}
}

// displayWidth returns the display width of a string (CJK chars count as 2)
func displayWidth(s string) int { return runewidth.StringWidth(s) }

// 输入辅助函数：选择模式 (resume/theme/plan/confirm) 优先路由到 cmdInput，确保焦点/输入正确
// 普通模式路由到 textarea
func (m *tuiModel) isUsingCmdInput() bool {
	return m.mcpWizard || m.nameSetup || m.resuming || m.themeSelecting || m.planSelecting || m.confirming
}
func (m *tuiModel) inputReset() {
	if m.isUsingCmdInput() { m.cmdInput.SetValue("") } else { m.ta.Reset(); m.ta.SetHeight(1) }
}
func (m tuiModel) inputValue() string {
	if m.isUsingCmdInput() { return m.cmdInput.Value() }
	return m.ta.Value()
}
func (m *tuiModel) inputFocus() tea.Cmd {
	if m.isUsingCmdInput() { return m.cmdInput.Focus() }
	return m.ta.Focus()
}
func (m *tuiModel) inputSetValue(v string) {
	if m.isUsingCmdInput() { m.cmdInput.SetValue(v) } else { m.ta.SetValue(v) }
}

func (m tuiModel) renderMessages() string {
	lines := make([]string, len(m.messages)); copy(lines, m.messages)
	if m.busy {
		lines = append(lines, assistantPrefixStyle.Render(getSpinnerChar()+" MiMo"))
		// 只显示最近的工具调用（而不是全部 streamLines）
		if len(m.streamLines) > 0 {
			// 找最后一个工具调用的起始位置（● 开头的行）
			start := 0
			for i := len(m.streamLines) - 1; i >= 0; i-- {
				if strings.Contains(m.streamLines[i], "●") { start = i; break }
			}
			lines = append(lines, m.streamLines[start:]...)
		}
		if m.planActive && m.currentPlan != nil { completed := 0; for _, step := range m.currentPlan.Steps { if step.Status == agent.StepCompleted { completed++ } }; lines = append(lines, thinkingStyle.Render(fmt.Sprintf("  计划进度: %d/%d", completed, m.currentPlan.TotalSteps))) }
		if m.planning {
			elapsed := int(time.Since(m.planStart).Seconds()); lines = append(lines, thinkingStyle.Render(fmt.Sprintf("  ▎ planning... %ds", elapsed)))
			if m.rawThinking != "" { lines = append(lines, thinkingStyle.Render("  [规划中]")); for _, l := range strings.Split(m.rawThinking, "\n") { if strings.TrimSpace(l) != "" { lines = append(lines, thinkingContentStyle.Render("    "+l)) } } }
		} else if m.thinking {
			elapsed := int(time.Since(m.thinkStart).Seconds()); lines = append(lines, thinkingStyle.Render(fmt.Sprintf("  ▎ thinking... %ds", elapsed)))
			if m.rawThinking != "" { tl := strings.Split(m.rawThinking, "\n"); var last []string; for i := len(tl) - 1; i >= 0 && len(last) < 5; i-- { if strings.TrimSpace(tl[i]) != "" { last = append([]string{tl[i]}, last...) } }; for _, l := range last { for _, wl := range wrapText(l, m.width-6) { lines = append(lines, thinkingContentStyle.Render("  "+wl)) } } }
		}
		if m.responding && m.rawStream != "" { for _, line := range strings.Split(renderMarkdownSafe(m.rawStream, m.width), "\n") { lines = append(lines, "  "+line) } }
		if m.compressing {
			elapsedSec := int(time.Since(m.compressStart).Seconds()); progress := elapsedSec * 15; if progress > 95 { progress = 95 }
			filled := progress * 20 / 100; empty := 20 - filled
			lines = append(lines, "", assistantPrefixStyle.Render(getSpinnerChar()+" 正在压缩上下文..."))
			lines = append(lines, confirmTextStyle.Render("  "+strings.Repeat("█", filled)+strings.Repeat("░", empty)+fmt.Sprintf(" %d%%", progress)))
			lines = append(lines, thinkingStyle.Render(fmt.Sprintf("  已用时 %ds", elapsedSec)))
		}
	}
	// 选择列表（theme/plan/resume/confirm）已从 viewport 移至 View() 中的 inputBoxStyle 容器，
	// 不再渲染到消息区，避免聊天内容滚动时把选择列表带出视口。
	return strings.Join(lines, "\n")
}

func (m tuiModel) handleMCPWizardKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 步骤 0：名称 — 直接回车使用默认名
	if m.mcpWizardStep == 0 {
		if msg.String() == "enter" {
			name := strings.TrimSpace(m.inputValue())
			if name == "" { name = "mcp-server" }
			m.mcpWizardData["name"] = name; m.mcpWizardStep = 1
			m.mcpTypeSelectIdx = 0
			m.messages = append(m.messages, confirmTextStyle.Render("  名称: "+name), "")
			m.renderMCPTypeSelect()
			m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
			return m, nil
		}
		if msg.String() == "esc" {
			m.mcpWizard = false; m.messages = append(m.messages, confirmTextStyle.Render("  已取消"))
			m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
			return m, nil
		}
		var cmd tea.Cmd; m.cmdInput, cmd = m.cmdInput.Update(msg); return m, cmd
	}
	// 步骤 1：类型选择（↑↓ 导航）
	if m.mcpWizardStep == 1 {
		switch msg.String() {
		case "up", "k":
			if m.mcpTypeSelectIdx > 0 { m.mcpTypeSelectIdx-- }
			m.renderMCPTypeSelect(); return m, nil
		case "down", "j":
			if m.mcpTypeSelectIdx < 2 { m.mcpTypeSelectIdx++ }
			m.renderMCPTypeSelect(); return m, nil
		case "enter":
			types := []string{"npm", "command", "url"}
			labels := []string{"npm 包", "本地命令", "远程 URL (SSE)"}
			selected := types[m.mcpTypeSelectIdx]
			m.mcpWizardData["type"] = selected
			m.messages = append(m.messages, confirmTextStyle.Render("  类型: "+labels[m.mcpTypeSelectIdx]))
			if selected == "npm" {
				m.mcpWizardStep = 2
				m.mcpWizardOptions = mcpRecommendedServers; m.mcpWizardSelectIdx = 0
				m.showMCPRecommendList()
			} else {
				m.mcpWizardStep = 2; m.mcpWizardOptions = nil
				m.messages = append(m.messages, "")
				if selected == "command" {
					m.messages = append(m.messages, confirmTextStyle.Render("  请输入命令和参数"))
					m.messages = append(m.messages, helpHintStyle.Render("  例如: python -m mcp_server"))
				} else {
					m.messages = append(m.messages, confirmTextStyle.Render("  请输入服务器 URL"))
					m.messages = append(m.messages, helpHintStyle.Render("  例如: https://example.com/mcp/sse"))
				}
			}
			m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
			return m, nil
		case "esc":
			m.mcpWizardStep = 0
			m.messages = append(m.messages, helpHintStyle.Render("  返回上一步"))
			m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
			return m, nil
		default: return m, nil
		}
	}
	// 步骤 3：目录选择（↑↓ 导航）
	if m.mcpWizardStep == 3 {
		switch msg.String() {
		case "up", "k":
			if m.mcpDirSelectIdx > 0 { m.mcpDirSelectIdx-- }
			m.renderMCPDirSelect(); return m, nil
		case "down", "j":
			if m.mcpDirSelectIdx < len(m.mcpDirOptions)-1 { m.mcpDirSelectIdx++ }
			m.renderMCPDirSelect(); return m, nil
		case "enter":
			if m.mcpDirSelectIdx == len(m.mcpDirOptions)-1 {
				// "自定义输入..."
				m.mcpWizardStep = 4
				m.messages = append(m.messages, confirmTextStyle.Render("  选择了: 自定义输入"))
				m.messages = append(m.messages, helpHintStyle.Render("  请输入目录路径，多个用逗号分隔"))
				m.inputReset(); focusCmd := m.inputFocus()
				m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
				return m, focusCmd
			}
			dir := m.mcpDirOptions[m.mcpDirSelectIdx]
			m.messages = append(m.messages, confirmTextStyle.Render("  允许访问: "+dir))
			m.messages = append(m.messages, "", confirmTextStyle.Render("  正在安装并配置..."))
			m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
			m.busy = true; go m.installAndConfigureMCP(m.mcpWizardData["name"], m.mcpWizardData["pkg"], []string{dir})
			return m, m.tickSpinner()
		case "esc":
			m.mcpWizard = false; m.messages = append(m.messages, confirmTextStyle.Render("  已取消"))
			m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
			return m, nil
		default: return m, nil
		}
	}
	// 步骤 2：推荐服务器列表（↑↓ 导航）— 仅当有推荐选项时
	if m.mcpWizardStep == 2 && len(m.mcpWizardOptions) > 0 {
		switch msg.String() {
		case "up", "k":
			if m.mcpWizardSelectIdx > 0 { m.mcpWizardSelectIdx--; m.messages = m.messages[:len(m.messages)-10]; m.showMCPRecommendList(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom() }
			return m, nil
		case "down", "j":
			if m.mcpWizardSelectIdx < len(m.mcpWizardOptions)-1 { m.mcpWizardSelectIdx++; m.messages = m.messages[:len(m.messages)-10]; m.showMCPRecommendList(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom() }
			return m, nil
		case "enter":
			// 输入框有内容 → 当作自定义包名
			if input := strings.TrimSpace(m.inputValue()); input != "" {
				m.mcpWizardData["pkg"] = input
				m.mcpWizardOptions = nil
				if strings.Contains(input, "filesystem") {
					wd, _ := os.Getwd(); home, _ := os.UserHomeDir()
					m.mcpWizardStep = 3; m.mcpDirOptions = []string{wd, home, "自定义输入..."}; m.mcpDirSelectIdx = 0
					m.messages = append(m.messages, confirmTextStyle.Render("  包名: "+input))
					m.messages = append(m.messages, "", confirmBorderStyle.Render("  文件系统权限配置"))
					m.messages = append(m.messages, confirmTextStyle.Render("  选择允许访问的目录:"))
					m.renderMCPDirSelectLines()
				} else {
					m.messages = append(m.messages, confirmTextStyle.Render("  包名: "+input), "", confirmTextStyle.Render("  正在安装并配置..."))
					m.busy = true; go m.installAndConfigureMCP(m.mcpWizardData["name"], input, nil)
				}
				m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
				if m.busy { return m, m.tickSpinner() }
				return m, nil
			}
			// 输入框为空 → 从列表选择
			srv := m.mcpWizardOptions[m.mcpWizardSelectIdx]
			m.messages = append(m.messages, confirmTextStyle.Render(fmt.Sprintf("  选择了: %s (%s)", srv.Name, srv.Package)))
			if strings.Contains(srv.Package, "filesystem") {
				wd, _ := os.Getwd(); home, _ := os.UserHomeDir()
				m.mcpWizardData["pkg"] = srv.Package
				m.mcpWizardStep = 3
				m.mcpDirOptions = []string{wd, home, "自定义输入..."}
				m.mcpDirSelectIdx = 0
				m.messages = append(m.messages, "", confirmBorderStyle.Render("  文件系统权限配置"))
				m.messages = append(m.messages, confirmTextStyle.Render("  选择允许访问的目录:"))
				m.renderMCPDirSelectLines()
				m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
				return m, nil
			}
			m.messages = append(m.messages, "", confirmTextStyle.Render("  正在安装并配置..."))
			m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
			m.busy = true; go m.installAndConfigureMCP(m.mcpWizardData["name"], srv.Package, nil)
			return m, m.tickSpinner()
		case "esc":
			m.mcpWizard = false; m.messages = append(m.messages, confirmTextStyle.Render("  已取消"))
			m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
			return m, nil
		}
	}
	// 步骤 0/1/2(文本输入)/4(自定义路径)：透传给输入框，Enter 提交
	if msg.String() == "enter" {
		input := strings.TrimSpace(m.inputValue())
		if input == "" { return m, nil }
		return m.sendMessage(input)
	}
	if msg.String() == "esc" {
		m.mcpWizard = false; m.messages = append(m.messages, confirmTextStyle.Render("  已取消"))
		m.inputReset(); m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
		return m, nil
	}
	var cmd tea.Cmd; m.cmdInput, cmd = m.cmdInput.Update(msg); return m, cmd
}

// renderMCPTypeSelect 重绘类型选择界面
func (m *tuiModel) renderMCPTypeSelect() {
	cutoff := len(m.messages)
	for i := len(m.messages) - 1; i >= 0; i-- {
		if strings.Contains(m.messages[i], "选择服务器类型") { cutoff = i; break }
	}
	m.messages = m.messages[:cutoff]
	m.messages = append(m.messages, confirmBorderStyle.Render("  选择服务器类型:"))
	types := []string{"npm 包 (如 filesystem, github, sqlite)", "本地命令", "远程 URL (SSE)"}
	for i, label := range types {
		if i == m.mcpTypeSelectIdx {
			m.messages = append(m.messages, confirmSelectedStyle.Render("  ▶ "+label))
		} else {
			m.messages = append(m.messages, confirmTextStyle.Render("    "+label))
		}
	}
	m.messages = append(m.messages, "", helpHintStyle.Render("  ↑↓ 选择 | Enter 确认"))
	m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
}

// renderMCPDirSelectLines 生成目录选择列表行（追加到 m.messages）
func (m *tuiModel) renderMCPDirSelectLines() {
	labels := []string{"当前目录", "用户目录", "自定义输入..."}
	for i, opt := range m.mcpDirOptions {
		label := labels[i]
		if i < 2 { label += ": " + opt }
		if i == m.mcpDirSelectIdx {
			m.messages = append(m.messages, confirmSelectedStyle.Render("  ▶ "+label))
		} else {
			m.messages = append(m.messages, confirmTextStyle.Render("    "+label))
		}
	}
	m.messages = append(m.messages, "", helpHintStyle.Render("  ↑↓ 选择 | Enter 确认"))
}

// renderMCPDirSelect 重绘目录选择界面
func (m *tuiModel) renderMCPDirSelect() {
	// 保留到最后一条 assistantPrefixStyle 之前（即标题行之前的所有内容）
	cutoff := len(m.messages)
	for i := len(m.messages) - 1; i >= 0; i-- {
		if strings.Contains(m.messages[i], "文件系统权限配置") {
			cutoff = i; break
		}
	}
	m.messages = m.messages[:cutoff]
	m.messages = append(m.messages, confirmBorderStyle.Render("  文件系统权限配置"))
	m.messages = append(m.messages, confirmTextStyle.Render("  选择允许访问的目录:"))
	m.renderMCPDirSelectLines()
	m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
}
func (m *tuiModel) showMCPRecommendList() {
	m.messages = append(m.messages, confirmTextStyle.Render("  类型: npm 包"), "")
	m.messages = append(m.messages, confirmBorderStyle.Render("  推荐的 MCP 服务器:"), "")
	categories := []string{"文件", "数据库", "开发", "搜索", "浏览器", "工具", "协作"}; globalIdx := 0
	for _, cat := range categories {
		m.messages = append(m.messages, helpHintStyle.Render(fmt.Sprintf("  [%s]", cat)))
		for _, srv := range mcpRecommendedServers { if srv.Category == cat { prefix := "    "; if globalIdx == m.mcpWizardSelectIdx { prefix = "  ▶ "; m.messages = append(m.messages, selectedMenuStyle.Render(fmt.Sprintf("%s%d. %-20s %s", prefix, globalIdx+1, srv.Name, srv.Description))) } else { m.messages = append(m.messages, confirmTextStyle.Render(fmt.Sprintf("%s%d. %-20s %s", prefix, globalIdx+1, srv.Name, srv.Description))) }; globalIdx++ } }
	}
	m.messages = append(m.messages, "", helpHintStyle.Render("  ↑↓ 上下选择 | Enter 确认 | 直接输入包名"), "")
}
func (m *tuiModel) showMCPStatus() {
	if m.mcpManager == nil { m.messages = append(m.messages, errorStyle.Render("  MCP 未启用")); return }
	servers := m.mcpManager.ServerNames()
	if len(servers) == 0 { m.messages = append(m.messages, confirmTextStyle.Render("  无已连接的 MCP 服务器"), "", helpHintStyle.Render("  使用 /mcp add 添加新的 MCP 服务器")) } else {
		m.messages = append(m.messages, confirmBorderStyle.Render("  MCP 服务器:"))
		for _, name := range servers { client, _ := m.mcpManager.GetClient(name); status := "✓ 已连接"; if client != nil && !client.IsConnected() { status = "✗ 断开" }; m.messages = append(m.messages, confirmTextStyle.Render(fmt.Sprintf("  ▶ %s (%s)", name, status))) }
		tools := m.mcpManager.GetTools(); m.messages = append(m.messages, "", confirmTextStyle.Render(fmt.Sprintf("  共 %d 个 MCP 工具可用", len(tools))))
		if len(tools) > 0 { m.messages = append(m.messages, helpHintStyle.Render("  工具列表:")); for _, t := range tools { m.messages = append(m.messages, confirmTextStyle.Render(fmt.Sprintf("    - %s", t.Name()))) } }
		m.messages = append(m.messages, "", helpHintStyle.Render("  /mcp add 添加 | /mcp remove <名称> 移除"))
	}
}
func (m *tuiModel) removeMCPServer(name string) {
	if cfg.MCP.Servers != nil { delete(cfg.MCP.Servers, name); config.SaveUserConfig(cfg) }
	m.messages = append(m.messages, confirmTextStyle.Render(fmt.Sprintf("  已移除 MCP 服务器: %s", name)), helpHintStyle.Render("  重启后生效"))
}
func (m *tuiModel) handleMCPWizardInput(input string) {
	switch m.mcpWizardStep {
	case 0:
		m.mcpWizardData["name"] = input; m.mcpWizardStep = 1; m.mcpTypeSelectIdx = 0
		m.messages = append(m.messages, confirmTextStyle.Render(fmt.Sprintf("  名称: %s", input)), "")
		m.renderMCPTypeSelect()
	case 2:
		serverType := m.mcpWizardData["type"]; name := m.mcpWizardData["name"]
		switch serverType {
		case "npm":
			// 解析包名（数字选推荐列表，否则直接用输入）
			pkgName := input
			if idx, err := strconv.Atoi(input); err == nil && idx >= 1 && idx <= len(mcpRecommendedServers) { srv := mcpRecommendedServers[idx-1]; pkgName = srv.Package; m.messages = append(m.messages, confirmTextStyle.Render(fmt.Sprintf("  选择了: %s (%s)", srv.Name, srv.Package))) } else { m.messages = append(m.messages, confirmTextStyle.Render(fmt.Sprintf("  包名: %s", pkgName))) }
			m.mcpWizardData["pkg"] = pkgName
			// filesystem 类型：询问允许访问的目录
			if strings.Contains(pkgName, "filesystem") {
				wd, _ := os.Getwd(); home, _ := os.UserHomeDir()
				m.mcpWizardStep = 3
				m.mcpDirOptions = []string{wd, home, "自定义输入..."}
				m.mcpDirSelectIdx = 0
				m.messages = append(m.messages, "", confirmBorderStyle.Render("  文件系统权限配置"))
				m.messages = append(m.messages, confirmTextStyle.Render("  选择允许访问的目录:"))
				m.messages = append(m.messages, confirmSelectedStyle.Render("  ▶ 当前目录: "+wd))
				m.messages = append(m.messages, confirmTextStyle.Render("    用户目录: "+home))
				m.messages = append(m.messages, confirmTextStyle.Render("    自定义输入..."))
				m.messages = append(m.messages, "", helpHintStyle.Render("  ↑↓ 选择 | Enter 确认"))
			} else {
				m.messages = append(m.messages, "", confirmTextStyle.Render("  正在安装并配置..."))
				m.busy = true; go m.installAndConfigureMCP(name, pkgName, nil)
			}
		case "command": m.mcpWizardData["command"] = input; m.saveMCPConfig(name, "command", input, ""); m.messages = append(m.messages, confirmTextStyle.Render("  配置已保存"), helpHintStyle.Render("  重启后生效")); m.mcpWizard = false
		case "url": m.saveMCPConfig(name, "url", "", input); m.messages = append(m.messages, confirmTextStyle.Render("  配置已保存"), helpHintStyle.Render("  重启后生效")); m.mcpWizard = false
		}
	case 4:
		// filesystem 自定义路径输入
		name := m.mcpWizardData["name"]; pkgName := m.mcpWizardData["pkg"]
		var allowedDirs []string
		for _, d := range strings.Split(input, ",") { if s := strings.TrimSpace(d); s != "" { allowedDirs = append(allowedDirs, s) } }
		if len(allowedDirs) == 0 {
			m.messages = append(m.messages, errorStyle.Render("  请输入至少一个目录路径"))
			m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
			return
		}
		m.messages = append(m.messages, confirmTextStyle.Render(fmt.Sprintf("  允许访问: %s", strings.Join(allowedDirs, ", "))))
		m.messages = append(m.messages, "", confirmTextStyle.Render("  正在安装并配置..."))
		m.busy = true; go m.installAndConfigureMCP(name, pkgName, allowedDirs)
	}
}
func (m *tuiModel) saveMCPConfig(name, serverType, command, url string) {
	if cfg.MCP.Servers == nil { cfg.MCP.Servers = make(map[string]config.MCPServerConfig) }
	serverCfg := config.MCPServerConfig{Enabled: true}
	if serverType == "command" { parts := strings.Fields(command); if len(parts) > 0 { serverCfg.Command = parts[0]; if len(parts) > 1 { serverCfg.Args = parts[1:] } } } else if serverType == "url" { serverCfg.URL = url } else if serverType == "npm" { return }
	cfg.MCP.Servers[name] = serverCfg; config.SaveUserConfig(cfg)
}
func (m *tuiModel) installAndConfigureMCP(name, pkgName string, allowedDirs []string) {
	home, _ := os.UserHomeDir()
	mcpDir := filepath.Join(home, ".mimo", "mcp")
	if err := os.MkdirAll(mcpDir, 0755); err != nil {
		m.streamChan <- mcpInstallDoneMsg{success: false, err: fmt.Errorf("创建目录失败: %v", err)}; return
	}

	// 初始化 package.json（如果不存在）
	if _, err := os.Stat(filepath.Join(mcpDir, "package.json")); os.IsNotExist(err) {
		initCmd := exec.Command("npm", "init", "-y"); initCmd.Dir = mcpDir
		if output, err := initCmd.CombinedOutput(); err != nil {
			m.streamChan <- mcpInstallDoneMsg{success: false, err: fmt.Errorf("npm init 失败: %v\n%s", err, string(output))}; return
		}
	}

	packageDir := filepath.Join(mcpDir, "node_modules", pkgName)
	if _, err := os.Stat(packageDir); os.IsNotExist(err) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute); defer cancel()
		cmd := exec.CommandContext(ctx, "npm", "install", pkgName); cmd.Dir = mcpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				m.streamChan <- mcpInstallDoneMsg{success: false, err: fmt.Errorf("安装超时（超过 3 分钟）")}
			} else {
				m.streamChan <- mcpInstallDoneMsg{success: false, err: fmt.Errorf("安装失败: %v\n%s", err, string(output))}
			}
			return
		}
	}

	entryPoint := findNPMPackageEntryPoint(packageDir)
	if entryPoint == "" { m.streamChan <- mcpInstallDoneMsg{success: false, err: fmt.Errorf("找不到包的入口文件")}; return }
	if cfg.MCP.Servers == nil { cfg.MCP.Servers = make(map[string]config.MCPServerConfig) }
	args := []string{entryPoint}
	args = append(args, allowedDirs...)
	cfg.MCP.Servers[name] = config.MCPServerConfig{Command: "node", Args: args, Enabled: true}
	if err := config.SaveUserConfig(cfg); err != nil { m.streamChan <- mcpInstallDoneMsg{success: false, err: fmt.Errorf("保存配置失败: %v", err)}; return }
	m.streamChan <- mcpInstallDoneMsg{success: true, err: nil}
}
func findNPMPackageEntryPoint(packageDir string) string {
	pkgJsonPath := filepath.Join(packageDir, "package.json"); data, err := os.ReadFile(pkgJsonPath); if err != nil { return "" }
	var pkg struct { Main string `json:"main"`; Bin interface{} `json:"bin"` }
	if err := json.Unmarshal(data, &pkg); err != nil { return "" }
	if pkg.Main != "" { return filepath.Join(packageDir, pkg.Main) }
	for _, entry := range []string{"dist/index.js", "src/index.js", "index.js"} { p := filepath.Join(packageDir, entry); if _, err := os.Stat(p); err == nil { return p } }
	return ""
}

func renderMarkdownSimple(content string, width int) string {
	lines := strings.Split(content, "\n"); var result []string
	for _, line := range lines {
		if strings.HasPrefix(line, "### ") { result = append(result, bannerTitleStyle.Render("  "+strings.TrimPrefix(line, "### "))); continue }
		if strings.HasPrefix(line, "## ") { result = append(result, bannerTitleStyle.Render("  "+strings.TrimPrefix(line, "## "))); continue }
		if strings.HasPrefix(line, "# ") { result = append(result, bannerTitleStyle.Render("  "+strings.TrimPrefix(line, "# "))); continue }
		if strings.HasPrefix(line, "```") { result = append(result, thinkingContentStyle.Render("  ─────")); continue }
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			indent := len(line) - len(strings.TrimLeft(line, " ")); prefix := strings.Repeat(" ", indent)
			text := strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* "); text = renderInlineMarkdown(text)
			result = append(result, prefix+toolDotStyle.Render("•")+" "+text); continue
		}
		if len(trimmed) > 2 && trimmed[0] >= '0' && trimmed[0] <= '9' && strings.Contains(trimmed[:min(4, len(trimmed))], ". ") {
			idx := strings.Index(trimmed, ". "); num := trimmed[:idx]; text := trimmed[idx+2:]; text = renderInlineMarkdown(text)
			result = append(result, "  "+tokenLabelStyle.Render(num+".")+" "+text); continue
		}
		if strings.TrimSpace(line) == "" { result = append(result, ""); continue }
		if displayWidth(line) > width { wrapped := wrapText(line, width); for _, wl := range wrapped { result = append(result, "  "+renderInlineMarkdown(wl)) } } else { result = append(result, "  "+renderInlineMarkdown(line)) }
	}
	return strings.Join(result, "\n")
}
func renderInlineMarkdown(text string) string {
	text = replaceInline(text, "**", "**", func(s string) string { return lipgloss.NewStyle().Bold(true).Render(s) })
	text = replaceInline(text, "*", "*", func(s string) string { return lipgloss.NewStyle().Italic(true).Render(s) })
	text = replaceInline(text, "`", "`", func(s string) string { return lipgloss.NewStyle().Foreground(themeLabel).Render(s) })
	return text
}
func replaceInline(text, open, close string, style func(string) string) string {
	for {
		start := strings.Index(text, open); if start == -1 { break }
		end := strings.Index(text[start+len(open):], close); if end == -1 { break }
		end += start + len(open); inner := text[start+len(open) : end]; styled := style(inner)
		text = text[:start] + styled + text[end+len(close):]
	}
	return text
}
func wrapText(text string, width int) []string {
	if width <= 0 { return []string{text} }
	var lines []string
	var currentLine strings.Builder
	currentWidth := 0
	for _, r := range text {
		rw := runewidth.RuneWidth(r)
		if r == '\n' || currentWidth+rw > width {
			if currentWidth > 0 { lines = append(lines, currentLine.String()); currentLine.Reset(); currentWidth = 0 }
			if r != '\n' { currentLine.WriteRune(r); currentWidth = rw }
		} else { currentLine.WriteRune(r); currentWidth += rw }
	}
	if currentWidth > 0 { lines = append(lines, currentLine.String()) }
	return lines
}
