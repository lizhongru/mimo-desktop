package tools

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

// DependencyTool manages project dependencies (npm, pip, go, cargo)
type DependencyTool struct{}

func NewDependencyTool() *DependencyTool { return &DependencyTool{} }

func (t *DependencyTool) Name() string        { return "dependency" }
func (t *DependencyTool) Description() string  { return "Manage project dependencies (auto-detects npm/pip/go/cargo)" }
func (t *DependencyTool) GetSafetyLevel() SafetyLevel { return SafetyMedium }
func (t *DependencyTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action: install, list, add, remove",
				"enum":        []string{"install", "list", "add", "remove"},
			},
			"package": map[string]interface{}{
				"type":        "string",
				"description": "Package name (for add/remove actions)",
			},
		},
		"required": []string{"action"},
	}
}
func (t *DependencyTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "action")
	return err
}
func (t *DependencyTool) RequiresConfirmation(params map[string]interface{}) bool {
	action, _ := StringParam(params, "action")
	return action == "install" || action == "add" || action == "remove"
}

func (t *DependencyTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	action, _ := StringParam(params, "action")
	pkg, _ := StringParam(params, "package")

	wd, _ := os.Getwd()
	pm := detectPackageManager(wd)
	if pm == "" {
		return ToolError("no package manager detected (checked package.json, requirements.txt, go.mod, Cargo.toml)"), nil
	}

	var cmd *exec.Cmd
	switch pm {
	case "npm":
		switch action {
		case "install":
			cmd = exec.CommandContext(ctx, "npm", "install")
		case "list":
			cmd = exec.CommandContext(ctx, "npm", "list", "--depth=0")
		case "add":
			if pkg == "" { return ToolError("package name required for add"), nil }
			cmd = exec.CommandContext(ctx, "npm", "install", pkg)
		case "remove":
			if pkg == "" { return ToolError("package name required for remove"), nil }
			cmd = exec.CommandContext(ctx, "npm", "uninstall", pkg)
		}
	case "pip":
		switch action {
		case "install":
			cmd = exec.CommandContext(ctx, "pip", "install", "-r", "requirements.txt")
		case "list":
			cmd = exec.CommandContext(ctx, "pip", "list")
		case "add":
			if pkg == "" { return ToolError("package name required for add"), nil }
			cmd = exec.CommandContext(ctx, "pip", "install", pkg)
		case "remove":
			if pkg == "" { return ToolError("package name required for remove"), nil }
			cmd = exec.CommandContext(ctx, "pip", "uninstall", "-y", pkg)
		}
	case "go":
		switch action {
		case "install":
			cmd = exec.CommandContext(ctx, "go", "mod", "download")
		case "list":
			cmd = exec.CommandContext(ctx, "go", "list", "-m", "all")
		case "add":
			if pkg == "" { return ToolError("package path required for add"), nil }
			cmd = exec.CommandContext(ctx, "go", "get", pkg)
		case "remove":
			if pkg == "" { return ToolError("package path required for remove"), nil }
			cmd = exec.CommandContext(ctx, "go", "get", pkg+"@none")
		}
	case "cargo":
		switch action {
		case "install":
			cmd = exec.CommandContext(ctx, "cargo", "build")
		case "list":
			cmd = exec.CommandContext(ctx, "cargo", "tree", "--depth=1")
		case "add":
			if pkg == "" { return ToolError("package name required for add"), nil }
			cmd = exec.CommandContext(ctx, "cargo", "add", pkg)
		case "remove":
			if pkg == "" { return ToolError("package name required for remove"), nil }
			cmd = exec.CommandContext(ctx, "cargo", "remove", pkg)
		}
	}

	if cmd == nil {
		return ToolError("unsupported action: %s", action), nil
	}

	cmd.Dir = wd
	out, err := cmd.CombinedOutput()
	result := string(out)
	if len(result) > 4000 {
		result = result[:4000] + "\n... (truncated)"
	}
	if err != nil {
		return ToolError("%s failed: %v\n%s", action, err, result), nil
	}
	return &ToolResult{Output: result}, nil
}

func detectPackageManager(dir string) string {
	if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
		return "npm"
	}
	if _, err := os.Stat(filepath.Join(dir, "requirements.txt")); err == nil {
		return "pip"
	}
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		return "go"
	}
	if _, err := os.Stat(filepath.Join(dir, "Cargo.toml")); err == nil {
		return "cargo"
	}
	return ""
}
