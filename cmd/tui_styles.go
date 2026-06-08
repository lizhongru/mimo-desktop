package cmd

// tui_styles.go — Lipgloss 样式定义（支持 dark/light 主题）

import "github.com/charmbracelet/lipgloss"

// 主题变量
var (
	themeBg         lipgloss.Color
	themeFg         lipgloss.Color
	themeFgDim      lipgloss.Color
	themeAccent     lipgloss.Color
	themeLabel      lipgloss.Color
	themeError      lipgloss.Color
	themeSuccess    lipgloss.Color
	themeWarning    lipgloss.Color
	themeBorder     lipgloss.Color
	themeHighlight  lipgloss.Color
)

// 样式变量
var (
	userPrefixStyle      lipgloss.Style
	userTextStyle        lipgloss.Style
	assistantPrefixStyle lipgloss.Style
	assistantTextStyle   lipgloss.Style
	toolDotStyle         lipgloss.Style
	toolMCPStyle         lipgloss.Style
	toolNameStyle        lipgloss.Style
	toolArgStyle         lipgloss.Style
	toolResultStyle      lipgloss.Style
	toolErrorStyle       lipgloss.Style
	errorStyle           lipgloss.Style
	thinkingStyle        lipgloss.Style
	thinkingContentStyle lipgloss.Style
	statusBarStyle       lipgloss.Style
	statusModelStyle     lipgloss.Style
	statusDotStyle       lipgloss.Style
	statusThinkingStyle  lipgloss.Style
	promptStyle          lipgloss.Style
	separatorStyle       lipgloss.Style
	helpHintStyle        lipgloss.Style
	confirmBorderStyle   lipgloss.Style
	confirmTextStyle     lipgloss.Style
	confirmOptionStyle   lipgloss.Style
	confirmSelectedStyle lipgloss.Style
	confirmHintStyle     lipgloss.Style
	bannerBorderStyle    lipgloss.Style
	bannerTitleStyle     lipgloss.Style
	bannerInfoStyle      lipgloss.Style
	bannerDimStyle       lipgloss.Style
	selectedMenuStyle    lipgloss.Style
	warningStyle         lipgloss.Style
	toolTimeStyle        lipgloss.Style
	welcomeTitleStyle    lipgloss.Style
	welcomeSubtitleStyle lipgloss.Style
	statsStyle           lipgloss.Style
	tokenLabelStyle      lipgloss.Style
	inputBoxStyle        lipgloss.Style
)

// SetTheme 设置主题
func SetTheme(theme string) {
	switch theme {
	case "dark":
		setDarkTheme()
	case "light":
		setLightTheme()
	default:
		setDarkTheme()
	}
	initStyles()
}

// GetTheme 获取当前主题
func GetTheme() string {
	return cfg.Theme
}

// setDarkTheme 设置暗色主题 — 小米橙 + 深灰
func setDarkTheme() {
	themeBg = lipgloss.Color("#121212")
	themeFg = lipgloss.Color("#E8E8E8")
	themeFgDim = lipgloss.Color("#6C6C6C")
	themeAccent = lipgloss.Color("#be8367")
	themeError = lipgloss.Color("#FF3B30")
	themeSuccess = lipgloss.Color("#30D158")
	themeWarning = lipgloss.Color("#FF9F0A")
	themeBorder = lipgloss.Color("#2C2C2E")
	themeHighlight = lipgloss.Color("#1C1C1E")
	themeLabel = lipgloss.Color("#79cbcb")
}

// setLightTheme 设置亮色主题 — 小米橙 + 浅灰
func setLightTheme() {
	themeBg = lipgloss.Color("#FFFFFF")
	themeFg = lipgloss.Color("#333333")
	themeFgDim = lipgloss.Color("#666666")
	themeAccent = lipgloss.Color("#be8367")
	themeError = lipgloss.Color("#CC0000")
	themeSuccess = lipgloss.Color("#006600")
	themeWarning = lipgloss.Color("#CC6600")
	themeBorder = lipgloss.Color("#CCCCCC")
	themeHighlight = lipgloss.Color("#E6E6E6")
	themeLabel = lipgloss.Color("#79cbcb")
}

// initStyles 初始化样式
func initStyles() {
	userPrefixStyle = lipgloss.NewStyle().Foreground(themeAccent).Bold(true)
	userTextStyle = lipgloss.NewStyle().Foreground(themeFg).Background(themeHighlight).Padding(0, 1)
	assistantPrefixStyle = lipgloss.NewStyle().Foreground(themeAccent).Bold(true)
	assistantTextStyle = lipgloss.NewStyle().Foreground(themeFg)
	toolDotStyle = lipgloss.NewStyle().Foreground(themeFgDim)
	toolMCPStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#a78bfa")).Bold(true)
	toolNameStyle = lipgloss.NewStyle().Foreground(themeFg)
	toolArgStyle = lipgloss.NewStyle().Foreground(themeFgDim)
	toolResultStyle = lipgloss.NewStyle().Foreground(themeFg)
	toolErrorStyle = lipgloss.NewStyle().Foreground(themeError)
	errorStyle = lipgloss.NewStyle().Foreground(themeError)
	thinkingStyle = lipgloss.NewStyle().Foreground(themeFgDim).Italic(true)
	thinkingContentStyle = lipgloss.NewStyle().Foreground(themeFgDim)
	statusBarStyle = lipgloss.NewStyle().Foreground(themeBorder)
	statusModelStyle = lipgloss.NewStyle().Foreground(themeAccent).Bold(true)
	statusDotStyle = lipgloss.NewStyle().Foreground(themeSuccess)
	statusThinkingStyle = lipgloss.NewStyle().Foreground(themeFgDim)
	promptStyle = lipgloss.NewStyle().Foreground(themeAccent).Bold(true)
	separatorStyle = lipgloss.NewStyle().Foreground(themeAccent)
	helpHintStyle = lipgloss.NewStyle().Foreground(themeFg).Italic(true)
	confirmBorderStyle = lipgloss.NewStyle().Foreground(themeAccent).Bold(true)
	confirmTextStyle = lipgloss.NewStyle().Foreground(themeFg)
	confirmOptionStyle = lipgloss.NewStyle().Foreground(themeFgDim)
	confirmSelectedStyle = lipgloss.NewStyle().Foreground(themeAccent).Bold(true)
	confirmHintStyle = lipgloss.NewStyle().Foreground(themeFg)
	bannerBorderStyle = lipgloss.NewStyle().Foreground(themeAccent)
	bannerTitleStyle = lipgloss.NewStyle().Foreground(themeAccent).Bold(true)
	bannerInfoStyle = lipgloss.NewStyle().Foreground(themeFg)
	bannerDimStyle = lipgloss.NewStyle().Foreground(themeFgDim)
	selectedMenuStyle = lipgloss.NewStyle().Foreground(themeAccent).Bold(true)
	warningStyle = lipgloss.NewStyle().Foreground(themeWarning).Bold(true)
	toolTimeStyle = lipgloss.NewStyle().Foreground(themeFg)
	welcomeTitleStyle = lipgloss.NewStyle().Foreground(themeAccent).Bold(true).MarginTop(1)
	welcomeSubtitleStyle = lipgloss.NewStyle().Foreground(themeFgDim)
	statsStyle = lipgloss.NewStyle().Foreground(themeFgDim)
	tokenLabelStyle = lipgloss.NewStyle().Foreground(themeLabel)
	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(themeAccent).
			Padding(0, 1)
}

// statsDotStyle 状态点样式 - 固定颜色
var statsDotStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#5fd7d7"))

// statsTextStyle 状态文字样式 - 固定颜色，小字体
var statsTextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#5fd7d7")).Faint(true)

// helpFooterStyle 底部帮助样式 - 固定灰色
var helpFooterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

func init() {
	SetTheme("dark")
}
