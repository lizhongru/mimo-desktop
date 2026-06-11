package tools

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// --- git_status ---

type GitStatusTool struct{}

func NewGitStatusTool() *GitStatusTool { return &GitStatusTool{} }

func (t *GitStatusTool) Name() string                { return "git_status" }
func (t *GitStatusTool) Description() string         { return "Show the working tree status" }
func (t *GitStatusTool) GetSafetyLevel() SafetyLevel { return SafetyLow }
func (t *GitStatusTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}
}
func (t *GitStatusTool) Validate(params map[string]interface{}) error            { return nil }
func (t *GitStatusTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *GitStatusTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	out, err := runGit(ctx, "status", "--porcelain")
	if err != nil {
		return ToolError("git status failed: %v", err), nil
	}
	if out == "" {
		return &ToolResult{Output: "✓ Working tree clean"}, nil
	}
	// 格式化 porcelain 输出
	var sb strings.Builder
	var staged, unstaged, untracked []string
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 4 {
			continue
		}
		code := line[:2]
		file := strings.TrimSpace(line[3:])
		switch {
		case code[0] != ' ' && code[0] != '?':
			staged = append(staged, fmt.Sprintf("  %c  %s", code[0], file))
		case code[1] != ' ' && code[1] != '?':
			unstaged = append(unstaged, fmt.Sprintf("  %c  %s", code[1], file))
		case code[0] == '?' && code[1] == '?':
			untracked = append(untracked, fmt.Sprintf("  %s", file))
		}
	}
	if len(staged) > 0 {
		sb.WriteString(fmt.Sprintf("Staged (%d):\n", len(staged)))
		sb.WriteString(strings.Join(staged, "\n"))
		sb.WriteString("\n")
	}
	if len(unstaged) > 0 {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("Unstaged (%d):\n", len(unstaged)))
		sb.WriteString(strings.Join(unstaged, "\n"))
		sb.WriteString("\n")
	}
	if len(untracked) > 0 {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("Untracked (%d):\n", len(untracked)))
		sb.WriteString(strings.Join(untracked, "\n"))
		sb.WriteString("\n")
	}
	return &ToolResult{Output: strings.TrimSpace(sb.String())}, nil
}

// --- git_diff ---

type GitDiffTool struct{}

func NewGitDiffTool() *GitDiffTool { return &GitDiffTool{} }

func (t *GitDiffTool) Name() string                { return "git_diff" }
func (t *GitDiffTool) Description() string         { return "Show changes between commits or working tree" }
func (t *GitDiffTool) GetSafetyLevel() SafetyLevel { return SafetyLow }
func (t *GitDiffTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File or directory to diff (optional)",
			},
			"staged": map[string]interface{}{
				"type":        "boolean",
				"description": "Show staged changes (default: false)",
			},
		},
		"required": []string{},
	}
}
func (t *GitDiffTool) Validate(params map[string]interface{}) error            { return nil }
func (t *GitDiffTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *GitDiffTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	args := []string{"diff"}
	isStaged := false
	if v, ok := params["staged"].(bool); ok && v {
		args = append(args, "--staged")
		isStaged = true
	}
	if v, err := StringParam(params, "path"); err == nil && v != "" {
		args = append(args, v)
	}
	out, err := runGit(ctx, args...)
	if err != nil {
		return ToolError("git diff failed: %v", err), nil
	}
	if out == "" {
		if isStaged {
			return &ToolResult{Output: "No staged changes"}, nil
		}
		return &ToolResult{Output: "No changes"}, nil
	}
	// 获取 diffstat 摘要
	statArgs := append([]string{}, args...)
	statArgs = append(statArgs, "--stat")
	statOut, _ := runGit(ctx, statArgs...)
	var sb strings.Builder
	if statOut != "" {
		sb.WriteString("Diff summary:\n")
		sb.WriteString(statOut)
		sb.WriteString("\n\n")
	}
	// 限制 diff 内容长度，防止上下文爆炸
	maxLen := 4000
	if len(out) > maxLen {
		sb.WriteString(fmt.Sprintf("Diff (truncated, showing first %d of %d chars):\n", maxLen, len(out)))
		sb.WriteString(out[:maxLen])
		sb.WriteString("\n... (truncated)")
	} else {
		sb.WriteString("Diff:\n")
		sb.WriteString(out)
	}
	return &ToolResult{Output: sb.String()}, nil
}

// --- git_log ---

type GitLogTool struct{}

func NewGitLogTool() *GitLogTool { return &GitLogTool{} }

func (t *GitLogTool) Name() string                { return "git_log" }
func (t *GitLogTool) Description() string         { return "Show commit logs" }
func (t *GitLogTool) GetSafetyLevel() SafetyLevel { return SafetyLow }
func (t *GitLogTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"count": map[string]interface{}{
				"type":        "integer",
				"description": "Number of commits to show (default: 10)",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File or directory to filter (optional)",
			},
		},
		"required": []string{},
	}
}
func (t *GitLogTool) Validate(params map[string]interface{}) error            { return nil }
func (t *GitLogTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *GitLogTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	count := 10
	if v, err := IntParam(params, "count"); err == nil && v > 0 {
		count = v
	}
	// 使用详细格式：hash | date | author | message
	args := []string{"log", fmt.Sprintf("-n%d", count), "--format=%h | %ad | %an | %s", "--date=short"}
	if v, err := StringParam(params, "path"); err == nil && v != "" {
		args = append(args, "--", v)
	}
	out, err := runGit(ctx, args...)
	if err != nil {
		return ToolError("git log failed: %v", err), nil
	}
	if out == "" {
		return &ToolResult{Output: "No commits found"}, nil
	}
	return &ToolResult{Output: out}, nil
}

// --- git_commit ---

type GitCommitTool struct{}

func NewGitCommitTool() *GitCommitTool { return &GitCommitTool{} }

func (t *GitCommitTool) Name() string                { return "git_commit" }
func (t *GitCommitTool) Description() string         { return "Stage files and create a git commit" }
func (t *GitCommitTool) GetSafetyLevel() SafetyLevel { return SafetyMedium }
func (t *GitCommitTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Commit message",
			},
			"files": map[string]interface{}{
				"type":        "array",
				"description": "Files to stage (optional)",
				"items":       map[string]interface{}{"type": "string"},
			},
			"all": map[string]interface{}{
				"type":        "boolean",
				"description": "Stage all modified files (-a flag, default: false)",
			},
		},
		"required": []string{"message"},
	}
}
func (t *GitCommitTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "message")
	return err
}
func (t *GitCommitTool) RequiresConfirmation(params map[string]interface{}) bool { return true }

func (t *GitCommitTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	message, _ := StringParam(params, "message")

	// Stage files
	if files, ok := params["files"].([]interface{}); ok && len(files) > 0 {
		args := []string{"add"}
		for _, f := range files {
			if s, ok := f.(string); ok {
				args = append(args, s)
			}
		}
		if _, err := runGit(ctx, args...); err != nil {
			return ToolError("git add failed: %v", err), nil
		}
	}

	// Commit
	commitArgs := []string{"commit", "-m", message}
	if v, ok := params["all"].(bool); ok && v {
		commitArgs = append(commitArgs, "-a")
	}
	out, err := runGit(ctx, commitArgs...)
	if err != nil {
		return ToolError("git commit failed: %v", err), nil
	}
	// 获取新提交的信息
	logOut, _ := runGit(ctx, "log", "-1", "--format=%h | %ad | %s", "--date=short")
	if logOut != "" {
		return &ToolResult{Output: fmt.Sprintf("Committed:\n%s\n%s", logOut, out)}, nil
	}
	return &ToolResult{Output: out}, nil
}

// --- helper ---

func runGit(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = WorkingDir(ctx)
	out, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(out))
	if err != nil {
		if result != "" {
			return result, fmt.Errorf("%s", result)
		}
		return "", err
	}
	return result, nil
}

func init() {
	_ = filepath.Clean // keep import
}

// --- git_branch ---

type GitBranchTool struct{}

func NewGitBranchTool() *GitBranchTool { return &GitBranchTool{} }

func (t *GitBranchTool) Name() string                { return "git_branch" }
func (t *GitBranchTool) Description() string         { return "List, create, or delete git branches" }
func (t *GitBranchTool) GetSafetyLevel() SafetyLevel { return SafetyLow }
func (t *GitBranchTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action: list (default), create, delete",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Branch name (for create/delete)",
			},
		},
		"required": []string{},
	}
}
func (t *GitBranchTool) Validate(params map[string]interface{}) error            { return nil }
func (t *GitBranchTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *GitBranchTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	action := OptionalStringParam(params, "action", "list")
	name, _ := StringParam(params, "name")

	switch action {
	case "list":
		out, err := runGit(ctx, "branch", "-a")
		if err != nil {
			return ToolError("git branch failed: %v", err), nil
		}
		if out == "" {
			return &ToolResult{Output: "No branches found"}, nil
		}
		return &ToolResult{Output: out}, nil

	case "create":
		if name == "" {
			return ToolError("branch name required for create"), nil
		}
		out, err := runGit(ctx, "branch", name)
		if err != nil {
			return ToolError("git branch create failed: %v", err), nil
		}
		return &ToolResult{Output: "Branch created: " + name + "\n" + out}, nil

	case "delete":
		if name == "" {
			return ToolError("branch name required for delete"), nil
		}
		out, err := runGit(ctx, "branch", "-d", name)
		if err != nil {
			return ToolError("git branch delete failed: %v", err), nil
		}
		return &ToolResult{Output: "Branch deleted: " + name + "\n" + out}, nil

	default:
		return ToolError("unknown action: %s (use list/create/delete)", action), nil
	}
}

// --- git_checkout ---

type GitCheckoutTool struct{}

func NewGitCheckoutTool() *GitCheckoutTool { return &GitCheckoutTool{} }

func (t *GitCheckoutTool) Name() string                { return "git_checkout" }
func (t *GitCheckoutTool) Description() string         { return "Switch branches or restore files" }
func (t *GitCheckoutTool) GetSafetyLevel() SafetyLevel { return SafetyMedium }
func (t *GitCheckoutTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"branch": map[string]interface{}{
				"type":        "string",
				"description": "Branch name to checkout",
			},
			"create": map[string]interface{}{
				"type":        "boolean",
				"description": "Create branch if it doesn't exist (-b flag)",
			},
		},
		"required": []string{"branch"},
	}
}
func (t *GitCheckoutTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "branch")
	return err
}
func (t *GitCheckoutTool) RequiresConfirmation(params map[string]interface{}) bool { return true }

func (t *GitCheckoutTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	branch, _ := StringParam(params, "branch")
	args := []string{"checkout"}
	if v, ok := params["create"].(bool); ok && v {
		args = append(args, "-b")
	}
	args = append(args, branch)
	out, err := runGit(ctx, args...)
	if err != nil {
		return ToolError("git checkout failed: %v", err), nil
	}
	return &ToolResult{Output: out}, nil
}

// --- git_merge ---

type GitMergeTool struct{}

func NewGitMergeTool() *GitMergeTool { return &GitMergeTool{} }

func (t *GitMergeTool) Name() string                { return "git_merge" }
func (t *GitMergeTool) Description() string         { return "Merge a branch into the current branch" }
func (t *GitMergeTool) GetSafetyLevel() SafetyLevel { return SafetyMedium }
func (t *GitMergeTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"branch": map[string]interface{}{
				"type":        "string",
				"description": "Branch to merge",
			},
			"no_ff": map[string]interface{}{
				"type":        "boolean",
				"description": "Create a merge commit even for fast-forward (--no-ff)",
			},
		},
		"required": []string{"branch"},
	}
}
func (t *GitMergeTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "branch")
	return err
}
func (t *GitMergeTool) RequiresConfirmation(params map[string]interface{}) bool { return true }

func (t *GitMergeTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	branch, _ := StringParam(params, "branch")
	args := []string{"merge"}
	if v, ok := params["no_ff"].(bool); ok && v {
		args = append(args, "--no-ff")
	}
	args = append(args, branch)
	out, err := runGit(ctx, args...)
	if err != nil {
		return ToolError("git merge failed: %v", err), nil
	}
	return &ToolResult{Output: out}, nil
}
