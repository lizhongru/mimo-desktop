package safety

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mimo-cli/mimo-cli/internal/permission"
)

// SafetyLevel defines the overall safety mode
type SafetyLevel string

const (
	LevelLockdown SafetyLevel = "lockdown"
	LevelConfirm  SafetyLevel = "confirm"
	LevelAuto     SafetyLevel = "auto"
)

// Guardrail enforces safety policies on tool executions
type Guardrail struct {
	level      SafetyLevel
	classifier *Classifier
	auditLog   *AuditLog
	confirmFn  func(action Action) (bool, error) // Callback for confirmation
	permission string                            // readonly, write, exec
	ruleset    permission.Ruleset
	workspace  string
}

// NewGuardrail creates a new safety guardrail
func NewGuardrail(level SafetyLevel, classifier *Classifier, auditLogPath string) *Guardrail {
	wd, _ := os.Getwd()
	return &Guardrail{
		level:      level,
		classifier: classifier,
		auditLog:   NewAuditLog(auditLogPath),
		permission: "exec", // default: allow all
		ruleset:    permission.DefaultRuleset(),
		workspace:  wd,
	}
}

// SetPermission sets the permission level (readonly, write, exec)
func (g *Guardrail) SetPermission(p string) {
	g.permission = p
}

// SetRuleset sets the runtime permission rules used before safety checks.
func (g *Guardrail) SetRuleset(ruleset permission.Ruleset) {
	if len(ruleset) == 0 {
		g.ruleset = permission.DefaultRuleset()
		return
	}
	g.ruleset = ruleset
}

// SetWorkspaceRoot sets the directory used for external_directory checks.
func (g *Guardrail) SetWorkspaceRoot(root string) {
	if root == "" {
		return
	}
	if abs, err := filepath.Abs(root); err == nil {
		g.workspace = filepath.Clean(abs)
		return
	}
	g.workspace = filepath.Clean(root)
}

// SetConfirmCallback sets the callback for confirmation dialogs
func (g *Guardrail) SetConfirmCallback(fn func(action Action) (bool, error)) {
	g.confirmFn = fn
}

// Check verifies if an action is allowed and asks for confirmation if needed
func (g *Guardrail) Check(toolName string, params map[string]interface{}) (bool, error) {
	action := g.classifier.Classify(toolName, params)

	// Permission check
	if !g.checkPermission(action.Level) {
		return false, fmt.Errorf("blocked by permission mode (%s): %s", g.permission, action.Description)
	}

	// Log the action
	g.auditLog.Log(action)

	if err := g.checkRuleset(action, false); err != nil {
		return false, err
	}

	switch action.Level {
	case ActionCritical:
		return false, fmt.Errorf("CRITICAL: operation blocked — %s", action.Description)

	case ActionHigh:
		switch g.level {
		case LevelLockdown:
			return false, fmt.Errorf("blocked in lockdown mode: %s", action.Description)
		case LevelAuto:
			return true, nil
		default: // confirm
			return g.confirmAction(action)
		}

	case ActionMedium:
		// Medium risk - only confirm in lockdown mode, auto-approve otherwise
		switch g.level {
		case LevelLockdown:
			return false, fmt.Errorf("blocked in lockdown mode: %s", action.Description)
		default:
			return true, nil
		}

	case ActionLow:
		return true, nil
	}

	return true, nil
}

// CheckWithConfirmAll checks with confirm-all support
func (g *Guardrail) CheckWithConfirmAll(toolName string, params map[string]interface{}, confirmAll bool) (bool, error) {
	action := g.classifier.Classify(toolName, params)

	// Permission check
	if !g.checkPermission(action.Level) {
		return false, fmt.Errorf("blocked by permission mode (%s): %s", g.permission, action.Description)
	}

	// Log the action
	g.auditLog.Log(action)

	if err := g.checkRuleset(action, confirmAll); err != nil {
		return false, err
	}

	switch action.Level {
	case ActionCritical:
		return false, fmt.Errorf("CRITICAL: operation blocked — %s", action.Description)

	case ActionHigh:
		switch g.level {
		case LevelLockdown:
			return false, fmt.Errorf("blocked in lockdown mode: %s", action.Description)
		case LevelAuto:
			return true, nil
		default: // confirm
			if confirmAll {
				return true, nil
			}
			return g.confirmAction(action)
		}

	case ActionMedium:
		// Medium risk - only confirm in lockdown mode, auto-approve otherwise
		switch g.level {
		case LevelLockdown:
			return false, fmt.Errorf("blocked in lockdown mode: %s", action.Description)
		default:
			return true, nil
		}

	case ActionLow:
		return true, nil
	}

	return true, nil
}

func (g *Guardrail) checkRuleset(action Action, confirmAll bool) error {
	category := permission.PermissionForTool(action.Tool)
	if category == "" {
		category = action.Tool
	}

	if g.hasExternalPath(action.Params) {
		if err := g.checkPermissionDecision("external_directory", action, confirmAll); err != nil {
			return err
		}
	}

	return g.checkPermissionDecision(category, action, confirmAll)
}

func (g *Guardrail) checkPermissionDecision(category string, action Action, confirmAll bool) error {
	decision := g.ruleset.Evaluate(category, action.Params)
	switch decision {
	case permission.Allow:
		return nil
	case permission.Deny:
		return fmt.Errorf("permission %s denied tool %s", category, action.Tool)
	case permission.Ask:
		if confirmAll {
			return nil
		}
		allowed, err := g.confirmAction(action)
		if err != nil {
			return err
		}
		if !allowed {
			return fmt.Errorf("permission %s declined for tool %s", category, action.Tool)
		}
		return nil
	default:
		return fmt.Errorf("permission %s has invalid action %q for tool %s", category, decision, action.Tool)
	}
}

func (g *Guardrail) hasExternalPath(params map[string]interface{}) bool {
	if len(params) == 0 {
		return false
	}
	for _, key := range []string{"path", "file_path", "path_a", "path_b", "dir", "directory"} {
		value, ok := params[key].(string)
		if !ok || value == "" {
			continue
		}
		if g.isExternalPath(value) {
			return true
		}
	}
	return false
}

func (g *Guardrail) isExternalPath(path string) bool {
	workspace := g.workspace
	if workspace == "" {
		workspace, _ = os.Getwd()
	}
	workspaceAbs, err := filepath.Abs(workspace)
	if err != nil {
		return false
	}

	target := path
	if !filepath.IsAbs(target) {
		target = filepath.Join(workspaceAbs, target)
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return false
	}

	rel, err := filepath.Rel(filepath.Clean(workspaceAbs), filepath.Clean(targetAbs))
	if err != nil {
		return true
	}
	return rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// confirmAction asks the user for confirmation
func (g *Guardrail) confirmAction(action Action) (bool, error) {
	// Use callback if available
	if g.confirmFn != nil {
		return g.confirmFn(action)
	}

	// Fallback to stdin
	fmt.Printf("\n⚠️  Safety confirmation required\n")
	fmt.Printf("   Action: %s\n", action.Description)
	fmt.Printf("   Tool:   %s\n", action.Tool)

	if cmd, ok := action.Params["command"].(string); ok {
		fmt.Printf("   Command: %s\n", cmd)
	}
	if path, ok := action.Params["path"].(string); ok {
		fmt.Printf("   File:    %s\n", path)
	}

	fmt.Printf("\n   Proceed? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes", nil
}

// checkPermission checks if the action level is allowed by the current permission mode
func (g *Guardrail) checkPermission(level ActionLevel) bool {
	switch g.permission {
	case "readonly":
		return level == ActionLow
	case "write":
		return level == ActionLow || level == ActionMedium
	default: // "exec" or empty
		return true
	}
}

// AuditLog records all tool executions
type AuditLog struct {
	path string
}

// NewAuditLog creates a new audit log
func NewAuditLog(path string) *AuditLog {
	return &AuditLog{path: path}
}

// Log records an action to the audit log
func (l *AuditLog) Log(action Action) {
	if l.path == "" {
		return
	}

	// Ensure directory exists
	// In production, this would be append-only with proper locking
	entry := fmt.Sprintf("[%s] %s | tool=%s | level=%s | desc=%s\n",
		time.Now().Format(time.RFC3339),
		"EXECUTE",
		action.Tool,
		action.Level,
		action.Description,
	)

	// Append to log file (simplified)
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return // Silently fail for audit log
	}
	defer f.Close()
	f.WriteString(entry)
}
