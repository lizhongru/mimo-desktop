# UI 修复计划

## 一、MIMO 横幅边框改成小米橙

**文件**: `cmd/tui_styles.go`

**现状**: `bannerBorderStyle` 使用的是 `themeBorder`（深灰色 `#2C2C2E`），横幅边框几乎看不见。

**修改**: 将 `bannerBorderStyle` 的前景色从 `themeBorder` 改为 `themeAccent`（小米橙 `#be8367`）。

```go
// 修改前
bannerBorderStyle = lipgloss.NewStyle().Foreground(themeBorder)

// 修改后
bannerBorderStyle = lipgloss.NewStyle().Foreground(themeAccent)
```

---

## 二、/resume 选择列表改为固定高度

**文件**: `cmd/tui_model.go` + `cmd/tui_view.go`

**现状**: 选择列表（resume/theme/plan/confirm）渲染在 `renderMessages()` 里，内容写入 viewport。当聊天消息很多时，选择列表会随 viewport 滚动消失（"一直往上就没了"）。

**修改思路**: 将选择列表从 viewport 中移出，渲染在 `View()` 的输入区域位置（viewport 下方），用带边框的固定高度容器。

### 2.1 `renderMessages()` 中删除选择列表渲染

**文件**: `cmd/tui_model.go`，`renderMessages()` 函数（约第 744~764 行）

删除以下整个 if-else 块：
```go
// 删除这段 ↓
if m.themeSelecting {
    lines = append(lines, "", confirmBorderStyle.Render("  选择主题 ..."))
    for j, t := range m.themeOptions { ... }
} else if m.planSelecting {
    lines = append(lines, "", confirmBorderStyle.Render("  选择规划模式 ..."))
    for j, mode := range m.planOptions { ... }
} else if m.resuming && m.sessionStore != nil {
    sessions, _ := m.sessionStore.ListSessions(20)
    lines = append(lines, "", confirmBorderStyle.Render("  选择要恢复的会话 ..."))
    for j, sid := range m.resumeFiles { ... }
} else if m.resuming {
    lines = append(lines, "", confirmTextStyle.Render("  没有可恢复的会话"))
} else if m.confirming {
    lines = append(lines, "", confirmBorderStyle.Render("  ⚠️  安全确认"))
    ...
}
```

### 2.2 `View()` 中添加选择列表渲染

**文件**: `cmd/tui_view.go`，`View()` 函数

在渲染 `inputView` 的位置，判断是否处于选择模式：

```go
var inputView string
if m.resuming || m.themeSelecting || m.planSelecting || m.confirming {
    // 渲染选择列表到固定容器（带边框）
    var selLines []string
    if m.themeSelecting {
        selLines = append(selLines, confirmBorderStyle.Render("  选择主题 (↑↓ 选择, Enter 确认, Esc 取消):"))
        for j, t := range m.themeOptions {
            current := ""; if t == GetTheme() { current = " (当前)" }
            if j == m.themeIdx {
                selLines = append(selLines, confirmSelectedStyle.Render("  ▶ "+t+current))
            } else {
                selLines = append(selLines, confirmTextStyle.Render("    "+t+current))
            }
        }
    } else if m.planSelecting {
        // 类似 themeSelecting，渲染 planOptions
    } else if m.resuming && m.sessionStore != nil {
        // 固定窗口渲染 resume 列表（最多显示 7 项）
        const winSize = 7
        sessions, _ := m.sessionStore.ListSessions(20)
        selLines = append(selLines, confirmBorderStyle.Render("  选择要恢复的会话 (↑↓ 选择, Enter 确认, Esc 取消):"))
        start := m.resumeIdx - winSize/2
        if start < 0 { start = 0 }
        if start+winSize > len(m.resumeFiles) { start = len(m.resumeFiles) - winSize }
        if start < 0 { start = 0 }
        if start > 0 { selLines = append(selLines, confirmTextStyle.Render("    ↑ ...")) }
        for j := start; j < start+winSize && j < len(m.resumeFiles); j++ {
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
                selLines = append(selLines, confirmSelectedStyle.Render("  ▶ "+label))
            } else {
                selLines = append(selLines, confirmTextStyle.Render("    "+label))
            }
        }
        if start+winSize < len(m.resumeFiles) {
            selLines = append(selLines, confirmTextStyle.Render("    ↓ ..."))
        }
    } else if m.resuming {
        selLines = append(selLines, confirmTextStyle.Render("  没有可恢复的会话"))
    } else if m.confirming {
        // 渲染安全确认对话框
    }
    inputView = inputBoxStyle.Render(strings.Join(selLines, "\n"))
} else if m.mcpWizard || m.nameSetup {
    inputView = m.cmdInput.View()
} else {
    inputView = m.ta.View()
}
```

### 2.3 `calcInputAreaHeight()` 适配

**文件**: `cmd/tui_model.go`

选择模式下使用固定高度，非选择模式使用 textarea 高度：

```go
func (m *tuiModel) calcInputAreaHeight() int {
    if m.resuming || m.themeSelecting || m.planSelecting || m.confirming {
        // 选择模式固定高度：边框2 + 内容区8 + 分隔线/状态栏/底部 = 15
        return 15
    }
    lines := len(strings.Split(m.ta.Value(), "\n"))
    if lines < 1 { lines = 1 }
    if lines > 8 { lines = 8 }
    m.ta.SetHeight(lines)
    return lines + 6 // textarea + 边框2 + 状态栏 + 分隔线 + 底部
}
```

---

## 三、输入框加边框

**文件**: `cmd/tui_styles.go` + `cmd/tui_view.go`

### 3.1 新增样式

**文件**: `cmd/tui_styles.go`

在样式变量声明区添加：
```go
var inputBoxStyle lipgloss.Style
```

在 `initStyles()` 中添加：
```go
inputBoxStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(themeAccent)
```

### 3.2 View() 中用 inputBoxStyle 包裹输入框

非选择模式时，用 `inputBoxStyle.Render(inputView)` 渲染 textarea。

---

## 四、修复 handleThemeKey / handlePlanKey

**文件**: `cmd/tui_model.go`

**现状**: 之前修改把 `inputFocus()` 加到了公共 return 路径，导致 up/down 按键也会调用 `inputFocus()`（在选择模式下不应该聚焦 textarea）。

### 4.1 handleThemeKey（约第 359~371 行）

```go
func (m tuiModel) handleThemeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "up", "k": if m.themeIdx > 0 { m.themeIdx-- } else { m.themeIdx = len(m.themeOptions) - 1 }
    case "down", "j": if m.themeIdx < len(m.themeOptions)-1 { m.themeIdx++ } else { m.themeIdx = 0 }
    case "enter": SetTheme(m.themeOptions[m.themeIdx]); cfg.Theme = m.themeOptions[m.themeIdx]; go config.SaveUserConfig(cfg); m.themeSelecting = false; m.messages = append(m.messages, confirmTextStyle.Render("  主题已切换为: "+m.themeOptions[m.themeIdx]))
    case "esc": m.themeSelecting = false
    }
    m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
    if !m.themeSelecting { focusCmd := m.inputFocus(); return m, focusCmd }
    return m, nil
}
```

关键点：只有 `!m.themeSelecting`（即 esc/enter 退出选择模式）时才调用 `inputFocus()`。

### 4.2 handlePlanKey 同理

```go
func (m tuiModel) handlePlanKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "up", "k": ...
    case "down", "j": ...
    case "enter": ...; m.planSelecting = false; ...
    case "esc": m.planSelecting = false
    }
    m.viewport.SetContent(m.renderMessages()); m.viewport.GotoBottom()
    if !m.planSelecting { focusCmd := m.inputFocus(); return m, focusCmd }
    return m, nil
}
```

---

## 五、其他调用 inputFocus() 的修改（已做，需保留）

| 位置 | 修改内容 |
|------|---------|
| `inputFocus()` 函数签名 | `func (m *tuiModel) inputFocus()` → `func (m *tuiModel) inputFocus() tea.Cmd`，返回 `tea.Cmd` |
| `agentDoneMsg` 处理 (3处) | `m.inputFocus(); return m, m.listenStream()` → `focusCmd := m.inputFocus(); return m, tea.Batch(focusCmd, m.listenStream())` |
| `loadResumeSession` | `m.inputFocus(); return m, nil` → `focusCmd := m.inputFocus(); return m, focusCmd` |
| `mcpWizard` 完成 (2处) | 捕获 focusCmd 并在 return 中使用 |
| `handleResumeKey` esc 分支 | 加入 `inputFocus()` + viewport 高度重算 |

---

## 六、修改顺序建议

1. `tui_styles.go`: bannerBorderColor + inputBoxStyle 变量声明和初始化
2. `tui_model.go`: fix handleThemeKey/handlePlanKey（只在退出选择时调 inputFocus）
3. `tui_model.go`: 删除 renderMessages() 中的选择列表渲染块
4. `tui_model.go`: 修改 calcInputAreaHeight()，选择模式返回固定高度
5. `tui_view.go`: View() 中添加选择列表渲染 + 输入框边框包裹
6. `go build ./...` 编译验证
