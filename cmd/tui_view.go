package cmd

// tui_view.go — 视图渲染

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// View 渲染视图
func (m tuiModel) View() string {
	if !m.ready {
		return "  Loading..."
	}
	var b strings.Builder
	b.WriteString(m.viewport.View())
	b.WriteString("\n")
	b.WriteString(m.renderStatusBar())
	b.WriteString("\n")
	sep := strings.Repeat("─", m.width)
	if len(m.messageQueue) > 0 {
		b.WriteString("\n")
		for _, q := range m.messageQueue {
			b.WriteString(userPrefixStyle.Render("❯ "+m.userName) + "  " + userTextStyle.Render(truncStr(q, 60)) + errorStyle.Render("  排队中") + "\n")
		}
		b.WriteString("\n")
	}
	b.WriteString(separatorStyle.Render(sep))
	b.WriteString("\n")
	if m.nameSetup {
		b.WriteString(confirmTextStyle.Render("  请输入你的昵称，按 Enter 确认") + "\n")
	}
	// 注意：confirming 提示已内联到 renderSelectorLines() 中，避免重复显示
	// 输入区 / 选择列表渲染
	// 选择模式（resume/theme/plan/confirm）使用固定高度带边框容器，固定在输入区位置
	// 其他模式使用 textarea，外层加 inputBoxStyle 边框
	var inputView string
	if m.resuming || m.themeSelecting || m.planSelecting || m.confirming {
		// 选择模式：固定 13 行高（外加上下边框 2 行 = 15），让容器真占满 calcInputAreaHeight 返回的高度
		selLines := m.renderSelectorLines()
		boxStyle := inputBoxStyle.Width(m.width - 2).Height(13)
		inputView = boxStyle.Render(strings.Join(selLines, "\n"))
	} else if m.mcpWizard || m.nameSetup {
		boxStyle := inputBoxStyle.Width(m.width - 2).Height(3)
		inputView = boxStyle.Render(m.cmdInput.View())
	} else {
		boxStyle := inputBoxStyle.Width(m.width - 2)
		inputView = boxStyle.Render(m.ta.View())
	}
	b.WriteString(inputView)
	b.WriteString("\n" + separatorStyle.Render(sep) + "\n")
	if m.isCmdMode() {
		maxLen := 0
		for _, c := range m.cmdMatches {
			if len(c.name) > maxLen {
				maxLen = len(c.name)
			}
		}
		maxLen += 2
		winSize := 5
		total := len(m.cmdMatches)
		start := 0
		if total > winSize {
			start = m.cmdSelectIdx - winSize + 1
			if start < 0 {
				start = 0
			}
			if start+winSize > total {
				start = total - winSize
			}
		}
		end := start + winSize
		if end > total {
			end = total
		}
		if start > 0 {
			b.WriteString(confirmTextStyle.Render("    ...") + "\n")
		}
		for i := start; i < end; i++ {
			c := m.cmdMatches[i]
			pad := strings.Repeat(" ", maxLen-len(c.name))
			if i == m.cmdSelectIdx {
				b.WriteString(confirmSelectedStyle.Render("  ▶ "+c.name+pad+c.desc))
			} else {
				b.WriteString(confirmTextStyle.Render("    "+c.name+pad+c.desc))
			}
			if i < end-1 {
				b.WriteString("\n")
			}
		}
		if end < total {
			b.WriteString("\n" + confirmTextStyle.Render("    ..."))
		}
		b.WriteString("\n")
	}
	if m.quitting {
		b.WriteString("  " + confirmSelectedStyle.Render("确定要退出吗？") + "  " + helpHintStyle.Render("y: 退出  n: 取消"))
	} else if m.confirming {
		b.WriteString("  " + helpHintStyle.Render("y: 确认  n: 取消"))
	} else if m.isCmdMode() {
		b.WriteString("  " + helpHintStyle.Render("↑↓ 选择  Tab 循环  y 确认  n/Esc 关闭  Enter 直接选中"))
	} else {
		b.WriteString("  " + helpFooterStyle.Render("/help for commands"))
	}
	b.WriteString("\n")
	return b.String()
}

// renderStatusBar 渲染状态栏
func (m tuiModel) renderStatusBar() string {
	bar := statusDotStyle.Render("●") + " " + statusModelStyle.Render(m.modelName)
	if cfg.Agent.PlanningMode != "" {
		bar += " [" + cfg.Agent.PlanningMode + "]"
	}
	if !m.busyStart.IsZero() {
		var d time.Duration
		if m.busy {
			d = time.Since(m.busyStart)
		} else {
			d = m.busyEnd.Sub(m.busyStart)
		}
		tokStr := formatTokens(m.totalUsage.TotalTokens)
		if m.totalUsage.TotalTokens > 96000 {
			tokStr = "⚠" + tokStr
		}
		bar += " " + thinkingStyle.Render(fmt.Sprintf("(%s · %s)", tokStr, formatDuration(d)))
	}
	var right string
	if m.compressing {
		right = fmt.Sprintf("compressing... %ds", int(time.Since(m.compressStart).Seconds()))
	} else if m.planning {
		if m.currentPlan != nil {
			right = fmt.Sprintf("planning... %d/%d", m.currentPlan.CurrentStep, m.currentPlan.TotalSteps)
		} else {
			right = fmt.Sprintf("generating plan... %ds", int(time.Since(m.planStart).Seconds()))
		}
	} else if m.thinking {
		right = fmt.Sprintf("thinking... %ds", int(time.Since(m.thinkStart).Seconds()))
	} else if m.busy {
		if m.currentTool != "" {
			right = fmt.Sprintf("executing %s... %ds", m.currentTool, int(time.Since(m.toolStart).Seconds()))
		} else if m.responding {
			right = "responding..."
		} else {
			right = "processing..."
		}
	} else {
		right = "ready"
	}
	gap := m.width - lipgloss.Width(bar) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}
	bar += strings.Repeat(" ", gap)
	if m.totalUsage.TotalTokens > 96000 {
		bar += statusThinkingStyle.Render(right)
	} else if m.thinking || m.compressing || m.planning {
		bar += statusThinkingStyle.Render(right)
	} else {
		bar += thinkingStyle.Render(right)
	}
	return bar
}

// renderWelcomeBanner 渲染欢迎横幅
func renderWelcomeBanner(appVersion, modelName string, width int) []string {
	inner := 60
	if width > 0 && width-4 < inner {
		inner = width - 4
	}
	if inner < 30 {
		inner = 30
	}
	bar := strings.Repeat("─", inner)
	return []string{
		bannerBorderStyle.Render("╭"+bar+"╮"),
		bannerBorderStyle.Render("│") + bannerPad("", inner) + bannerBorderStyle.Render("│"),
		bannerBorderStyle.Render("│") + bannerPadCenter(bannerTitleStyle.Render("███╗   ███╗ ██╗ ███╗   ███╗  ██████╗"), inner) + bannerBorderStyle.Render("│"),
		bannerBorderStyle.Render("│") + bannerPadCenter(bannerTitleStyle.Render("████╗ ████║ ██║ ████╗ ████║ ██╔═══██╗"), inner) + bannerBorderStyle.Render("│"),
		bannerBorderStyle.Render("│") + bannerPadCenter(bannerTitleStyle.Render("██╔████╔██║ ██║ ██╔████╔██║ ██║   ██║"), inner) + bannerBorderStyle.Render("│"),
		bannerBorderStyle.Render("│") + bannerPadCenter(bannerTitleStyle.Render("██║╚██╔╝██║ ██║ ██║╚██╔╝██║ ██║   ██║"), inner) + bannerBorderStyle.Render("│"),
		bannerBorderStyle.Render("│") + bannerPadCenter(bannerTitleStyle.Render("██║ ╚═╝ ██║ ██║ ██║ ╚═╝ ██║ ╚██████╔╝"), inner) + bannerBorderStyle.Render("│"),
		bannerBorderStyle.Render("│") + bannerPadCenter(bannerTitleStyle.Render("╚═╝     ╚═╝ ╚═╝ ╚═╝     ╚═╝  ╚═════╝"), inner) + bannerBorderStyle.Render("│"),
		bannerBorderStyle.Render("│") + bannerPad("", inner) + bannerBorderStyle.Render("│"),
		bannerBorderStyle.Render("│") + bannerPadCenter(bannerInfoStyle.Render("MIMO CLI")+" "+bannerDimStyle.Render(appVersion), inner) + bannerBorderStyle.Render("│"),
		bannerBorderStyle.Render("│") + bannerPadCenter(bannerDimStyle.Render("Model:")+" "+bannerInfoStyle.Render(modelName), inner) + bannerBorderStyle.Render("│"),
		bannerBorderStyle.Render("│") + bannerPad("", inner) + bannerBorderStyle.Render("│"),
		bannerBorderStyle.Render("╰"+bar+"╯"),
		"",
		"  " + bannerDimStyle.Render("输入消息后按 Enter 发送，'exit' 退出。"),
		"",
	}
}

// bannerPad 横幅填充
func bannerPad(content string, width int) string {
	dw := lipgloss.Width(content)
	if dw >= width {
		return content
	}
	return content + strings.Repeat(" ", width-dw)
}

// bannerPadCenter 横幅居中填充
func bannerPadCenter(content string, width int) string {
	dw := lipgloss.Width(content)
	if dw >= width {
		return content
	}
	left := (width - dw) / 2
	return strings.Repeat(" ", left) + content + strings.Repeat(" ", width-dw-left)
}

// formatTokens 格式化 token 数量
func formatTokens(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

// formatDuration 格式化时间
func formatDuration(d time.Duration) string {
	total := int(d.Seconds())
	if total < 60 {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%dm%02ds", total/60, total%60)
}

// truncStr 截断字符串
func truncStr(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

// renderSelectorLines 渲染选择列表行（用于 resume/theme/plan/confirm）
// 这些列表是固定的，不会随消息滚动而消失
func (m tuiModel) renderSelectorLines() []string {
	var lines []string
	if m.themeSelecting {
		lines = append(lines, confirmBorderStyle.Render("  选择主题 (↑↓ 选择, Enter 确认, Esc 取消):"))
		for j, t := range m.themeOptions {
			current := ""
			if t == GetTheme() { current = " (当前)" }
			if j == m.themeIdx {
				lines = append(lines, confirmSelectedStyle.Render("  ▶ "+t+current))
			} else {
				lines = append(lines, confirmTextStyle.Render("    "+t+current))
			}
		}
	} else if m.planSelecting {
		lines = append(lines, confirmBorderStyle.Render("  选择规划模式 (↑↓ 选择, Enter 确认, Esc 取消):"))
		for j, mode := range m.planOptions {
			current := ""
			if mode == cfg.Agent.PlanningMode { current = " (当前)" }
			desc := ""
			switch mode {
			case "react": desc = " - ReAct 循环"
			case "plan-execute": desc = " - 先生成计划再执行"
			case "auto": desc = " - 自动判断"
			}
			if j == m.planIdx {
				lines = append(lines, confirmSelectedStyle.Render("  ▶ "+mode+desc+current))
			} else {
				lines = append(lines, confirmTextStyle.Render("    "+mode+desc+current))
			}
		}
	} else if m.resuming && m.sessionStore != nil {
		// 固定窗口 7 项，超出滚动
		const winSize = 7
		sessions, _ := m.sessionStore.ListSessions(20)
		lines = append(lines, confirmBorderStyle.Render("  选择要恢复的会话 (↑↓ 选择, Enter 确认, Esc 取消):"))
		total := len(m.resumeFiles)
		start := m.resumeIdx - winSize/2
		if start < 0 { start = 0 }
		if start+winSize > total { start = total - winSize }
		if start < 0 { start = 0 }
		if start > 0 { lines = append(lines, confirmTextStyle.Render("    ↑ ...")) }
		for j := start; j < start+winSize && j < total; j++ {
			sid := m.resumeFiles[j]
			label := sid
			for _, s := range sessions {
				if s.ID == sid {
					label = truncStr(s.LastMessage, 50)
					if label == "" { label = sid }
					label += "  (" + s.UpdatedAt.Format("01-02 15:04") + ")"
					break
				}
			}
			if j == m.resumeIdx {
				lines = append(lines, confirmSelectedStyle.Render("  ▶ "+label))
			} else {
				lines = append(lines, confirmTextStyle.Render("    "+label))
			}
		}
		if start+winSize < total {
			lines = append(lines, confirmTextStyle.Render("    ↓ ..."))
		}
	} else if m.resuming {
		lines = append(lines, confirmTextStyle.Render("  没有可恢复的会话"))
	} else if m.confirming {
		lines = append(lines, confirmBorderStyle.Render("  ⚠️  安全确认"))
		lines = append(lines, confirmTextStyle.Render("  操作: "+m.confirmAction.Description))
		lines = append(lines, confirmTextStyle.Render("  工具: "+m.confirmAction.Tool))
		if cmd, ok := m.confirmAction.Params["command"].(string); ok {
			lines = append(lines, confirmTextStyle.Render("  命令: "+cmd))
		}
		if path, ok := m.confirmAction.Params["path"].(string); ok {
			lines = append(lines, confirmTextStyle.Render("  文件: "+path))
		}
		ys, ns, als := confirmOptionStyle, confirmOptionStyle, confirmOptionStyle
		if m.confirmChoice == 0 { ys = confirmSelectedStyle } else if m.confirmChoice == 1 { ns = confirmSelectedStyle } else { als = confirmSelectedStyle }
		lines = append(lines, "", ys.Render("  ▶ 是 (y)"), ns.Render("    否 (n)"), als.Render("    全部确认 (a)"))
		lines = append(lines, confirmHintStyle.Render("  使用 ↑↓ 选择，Enter 确认"))
	}
	return lines
}
