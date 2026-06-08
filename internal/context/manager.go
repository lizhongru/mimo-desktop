package context

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mimo-cli/mimo-cli/internal/ignore"
)

// Manager manages the context for the agent
type Manager struct {
	projectDir string
	maxTokens  int
	ignore     *ignore.Matcher
}

// NewManager creates a new context manager
func NewManager(projectDir string, maxTokens int, ign *ignore.Matcher) *Manager {
	return &Manager{
		projectDir: projectDir,
		maxTokens:  maxTokens,
		ignore:     ign,
	}
}

// ProjectInfo contains information about the current project
type ProjectInfo struct {
	RootDir       string
	ProjectType   string
	GitBranch     string
	GitStatus     string
	HasAgentMD    bool
	AgentMD       string
	AgentMDParsed *AgentMDSections
	HasReadme     bool
	Readme        string
	TreeStructure string
	HasExtraRules bool
	ExtraRules    string
}

// AgentMDSections holds structured sections extracted from AGENT.md
type AgentMDSections struct {
	ProjectOverview string            // 项目概述
	TechStack       string            // 技术栈
	CodeConventions string            // 编码规范
	ProjectStructure string           // 项目结构
	CommonCommands  string            // 常用命令
	AgentRules      string            // 代理规则 (critical)
	ForbiddenOps    string            // 禁止操作 (critical)
	CustomSections  map[string]string // 其他自定义 section
}

// CollectProjectInfo gathers information about the current project
func (m *Manager) CollectProjectInfo() (*ProjectInfo, error) {
	info := &ProjectInfo{
		RootDir: m.projectDir,
	}

	// Detect project type
	info.ProjectType = detectProjectType(m.projectDir)

	// Get Git info
	info.GitBranch, info.GitStatus = getGitInfo(m.projectDir)

	// Check for AGENT.md
	agentPath := filepath.Join(m.projectDir, "AGENT.md")
	if data, err := os.ReadFile(agentPath); err == nil {
		info.HasAgentMD = true
		info.AgentMD = string(data)
		info.AgentMDParsed = parseAgentMD(string(data))
	}

	// Check for README.md
	readmePath := filepath.Join(m.projectDir, "README.md")
	if data, err := os.ReadFile(readmePath); err == nil {
		info.HasReadme = true
		info.Readme = truncate(string(data), 2000)
	}

	// Check for .mimo/rules.md (额外规则)
	rulesPath := filepath.Join(m.projectDir, ".mimo", "rules.md")
	if data, err := os.ReadFile(rulesPath); err == nil {
		info.HasExtraRules = true
		info.ExtraRules = string(data)
	}

	// Get directory tree
	info.TreeStructure = m.getDirectoryTree(m.projectDir, 3)

	return info, nil
}

// BuildSystemPrompt builds the system prompt with project context
func (m *Manager) BuildSystemPrompt() string {
	info, err := m.CollectProjectInfo()
	if err != nil {
		return defaultSystemPrompt()
	}

	var sb strings.Builder
	sb.WriteString(defaultSystemPrompt())

	sb.WriteString("\n\n## Current Project Context\n\n")

	// 操作系统信息，让 AI 知道该用什么命令
	osName := runtime.GOOS
	switch osName {
	case "windows":
		sb.WriteString("Operating System: Windows\n")
		sb.WriteString("Shell: cmd.exe (use Windows commands like dir, type, del, copy, move, findstr)\n")
		sb.WriteString("IMPORTANT: You are running on Windows. Always use Windows-native commands. Do NOT use Unix commands like ls, cat, grep, chmod, etc.\n")
	case "darwin":
		sb.WriteString("Operating System: macOS\n")
		sb.WriteString("Shell: /bin/sh (use Unix commands like ls, cat, grep, chmod)\n")
	case "linux":
		sb.WriteString("Operating System: Linux\n")
		sb.WriteString("Shell: /bin/sh (use Unix commands like ls, cat, grep, chmod)\n")
	default:
		sb.WriteString(fmt.Sprintf("Operating System: %s\n", osName))
	}

	sb.WriteString(fmt.Sprintf("Project Directory: %s\n", info.RootDir))
	sb.WriteString(fmt.Sprintf("Project Type: %s\n", info.ProjectType))

	// Git 信息
	if info.GitBranch != "" {
		sb.WriteString(fmt.Sprintf("Git Branch: %s\n", info.GitBranch))
	}
	if info.GitStatus != "" {
		sb.WriteString(fmt.Sprintf("Git Status:\n%s\n", info.GitStatus))
	}

	// .mimo/rules.md — 用户自定义规则，最高优先级注入
	if info.HasExtraRules {
		sb.WriteString("\n## ⚠ CRITICAL RULES — YOU MUST FOLLOW THESE IN EVERY RESPONSE\n")
		sb.WriteString("These rules were set by the project owner. They override all other instructions.\n\n")
		sb.WriteString(info.ExtraRules)
		sb.WriteString("\n")
	}

	// AGENT.md 结构化解析
	if info.HasAgentMD && info.AgentMDParsed != nil {
		parsed := info.AgentMDParsed

		// 代理规则 - 最高优先级，醒目展示
		if parsed.AgentRules != "" {
			sb.WriteString("\n### ⚠ MUST-FOLLOW Agent Rules\n")
			sb.WriteString(parsed.AgentRules)
			sb.WriteString("\n")
		}

		// 禁止操作 - 最高优先级
		if parsed.ForbiddenOps != "" {
			sb.WriteString("\n### 🚫 FORBIDDEN Operations (NEVER DO THESE)\n")
			sb.WriteString(parsed.ForbiddenOps)
			sb.WriteString("\n")
		}

		// 编码规范
		if parsed.CodeConventions != "" {
			sb.WriteString("\n### Code Conventions\n")
			sb.WriteString(parsed.CodeConventions)
			sb.WriteString("\n")
		}

		// 常用命令
		if parsed.CommonCommands != "" {
			sb.WriteString("\n### Common Commands\n")
			sb.WriteString(parsed.CommonCommands)
			sb.WriteString("\n")
		}

		// 项目概述
		if parsed.ProjectOverview != "" {
			sb.WriteString("\n### Project Overview\n")
			sb.WriteString(parsed.ProjectOverview)
			sb.WriteString("\n")
		}

		// 技术栈
		if parsed.TechStack != "" {
			sb.WriteString("\n### Tech Stack\n")
			sb.WriteString(parsed.TechStack)
			sb.WriteString("\n")
		}

		// 项目结构
		if parsed.ProjectStructure != "" {
			sb.WriteString("\n### Project Structure (from AGENT.md)\n")
			sb.WriteString(parsed.ProjectStructure)
			sb.WriteString("\n")
		}

		// 自定义 sections
		for title, content := range parsed.CustomSections {
			sb.WriteString(fmt.Sprintf("\n### %s\n", title))
			sb.WriteString(content)
			sb.WriteString("\n")
		}
	} else if info.HasAgentMD {
		// fallback: 无法解析时直接注入原文
		sb.WriteString("\n### Project Rules (AGENT.md)\n")
		sb.WriteString(info.AgentMD)
		sb.WriteString("\n")
	}



	if info.HasReadme {
		sb.WriteString("\n### Project README\n")
		sb.WriteString(info.Readme)
		sb.WriteString("\n")
	}

	sb.WriteString("\n### Project Structure\n```\n")
	sb.WriteString(info.TreeStructure)
	sb.WriteString("\n```\n")

	return sb.String()
}

// parseAgentMD parses AGENT.md into structured sections
func parseAgentMD(content string) *AgentMDSections {
	sections := &AgentMDSections{
		CustomSections: make(map[string]string),
	}

	lines := strings.Split(content, "\n")
	var currentSection string
	var currentContent strings.Builder

	flushSection := func() {
		if currentSection == "" {
			return
		}
		content := strings.TrimSpace(currentContent.String())
		if content == "" {
			return
		}

		lower := strings.ToLower(currentSection)

		switch {
		case strings.Contains(lower, "规则") || strings.Contains(lower, "代理规则") ||
			strings.Contains(lower, "agent rules") || strings.Contains(lower, "代理指令"):
			sections.AgentRules += content + "\n"
		case strings.Contains(lower, "禁止") || strings.Contains(lower, "forbidden") ||
			strings.Contains(lower, "不要") || strings.Contains(lower, "do not"):
			sections.ForbiddenOps += content + "\n"
		case strings.Contains(lower, "编码规范") || strings.Contains(lower, "coding") ||
			strings.Contains(lower, "convention") || strings.Contains(lower, "style"):
			sections.CodeConventions += content + "\n"
		case strings.Contains(lower, "常用命令") || strings.Contains(lower, "commands") ||
			strings.Contains(lower, "命令"):
			sections.CommonCommands += content + "\n"
		case strings.Contains(lower, "项目概述") || strings.Contains(lower, "overview") ||
			strings.Contains(lower, "概述"):
			sections.ProjectOverview += content + "\n"
		case strings.Contains(lower, "技术栈") || strings.Contains(lower, "tech stack") ||
			strings.Contains(lower, "stack"):
			sections.TechStack += content + "\n"
		case strings.Contains(lower, "项目结构") || strings.Contains(lower, "structure") ||
			strings.Contains(lower, "目录"):
			sections.ProjectStructure += content + "\n"
		default:
			// 保存为自定义 section
			sections.CustomSections[currentSection] = content + "\n"
		}

		currentContent.Reset()
	}

	for _, line := range lines {
		// 检测 Markdown 标题 (## 或 ###)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			flushSection()
			currentSection = strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))
			currentContent.Reset()
			continue
		}
		if strings.HasPrefix(trimmed, "### ") {
			flushSection()
			currentSection = strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))
			currentContent.Reset()
			continue
		}
		if currentSection != "" {
			currentContent.WriteString(line + "\n")
		}
	}
	flushSection()

	return sections
}

// getGitInfo returns the current git branch and status
func getGitInfo(dir string) (branch, status string) {
	// 获取分支名
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	if out, err := cmd.Output(); err == nil {
		branch = strings.TrimSpace(string(out))
	}

	// 获取简洁状态
	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	if out, err := cmd.Output(); err == nil {
		s := strings.TrimSpace(string(out))
		if s == "" {
			status = "clean"
		} else {
			lines := strings.Split(s, "\n")
			if len(lines) > 10 {
				status = fmt.Sprintf("%d files changed (showing first 10)\n", len(lines))
				for i := 0; i < 10; i++ {
					status += lines[i] + "\n"
				}
			} else {
				status = s
			}
		}
	}

	return branch, status
}

// getDirectoryTree generates a directory tree string
func (m *Manager) getDirectoryTree(dir string, depth int) string {
	if depth <= 0 {
		return ""
	}

	var sb strings.Builder
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	count := 0
	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(dir, name)

		// Use ignore matcher
		if m.ignore != nil && m.ignore.ShouldIgnore(name, entry.IsDir(), dir, fullPath) {
			continue
		}

		if entry.IsDir() {
			sb.WriteString(fmt.Sprintf("%s/\n", name))
			subtree := m.getDirectoryTree(filepath.Join(dir, name), depth-1)
			for _, line := range strings.Split(subtree, "\n") {
				if line != "" {
					sb.WriteString("  " + line + "\n")
				}
			}
		} else {
			sb.WriteString(name + "\n")
		}

		count++
		if count > 50 {
			sb.WriteString("... (truncated)\n")
			break
		}
	}

	return sb.String()
}

// detectProjectType detects the project type based on files present
func detectProjectType(dir string) string {
	signatures := map[string]string{
		"package.json":     "node",
		"pyproject.toml":   "python",
		"setup.py":         "python",
		"requirements.txt": "python",
		"go.mod":           "go",
		"Cargo.toml":       "rust",
		"pom.xml":          "java",
		"build.gradle":     "java",
		"composer.json":    "php",
		"Gemfile":          "ruby",
		"pubspec.yaml":     "flutter",
		"Dockerfile":       "docker",
	}

	for file, ptype := range signatures {
		if _, err := os.Stat(filepath.Join(dir, file)); err == nil {
			return ptype
		}
	}

	return "unknown"
}

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n... (truncated)"
}

// defaultSystemPrompt returns the default system prompt
func defaultSystemPrompt() string {
	return `You are MiMo, an AI-powered command-line development assistant.

## Core Behavior
- Be concise and direct. Answer questions immediately without unnecessary preamble.
- Always use tools when the user asks to read, list, search, analyze, edit, or create files/directories.
- Only reply with text (no tools) for pure greetings, abstract questions, or explanations that don't need file access.
- When you do need to make changes, read only the specific files involved, not the entire project.

## Rules
- If a command might be dangerous, warn the user first
- When writing code, follow the project's existing style and conventions
- Do not read files unless explicitly asked or absolutely necessary for the task
- If "CRITICAL RULES" are defined below, you MUST follow them in EVERY response without exception

## Available Tools
- shell: Execute shell commands
- file_read: Read file contents (with optional offset/limit)
- file_write: Write/create files
- file_edit: Edit files (exact text replacement)
- file_delete: Delete a file or empty directory (requires confirmation)
- file_diff: Compare two files and show differences
- dir_list: List directory contents (supports recursive, pattern filter)
- dir_create: Create a directory (including parent directories)
- glob: Find files by glob pattern (e.g. "**/*.go")
- search: Search text in files by regex pattern
- json_query: Read and query JSON/YAML files using dot-notation paths
- git_status: Show git working tree status
- git_diff: Show git changes (working tree or staged)
- git_log: Show git commit history
- git_commit: Stage files and create a git commit
- git_branch: List, create, or delete branches
- git_checkout: Switch branches or restore files
- git_merge: Merge a branch into current branch
- web_fetch: Fetch content from a URL
- web_search: Search the web using DuckDuckGo
- http_request: Send HTTP requests (GET/POST/PUT/DELETE)
- clipboard: Read from or write to the system clipboard
- process: List running processes or kill a process
- env: Read or list environment variables
- dependency: Manage project dependencies (auto-detects npm/pip/go/cargo)`
}
