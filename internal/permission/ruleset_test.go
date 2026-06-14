package permission

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultRuleset(t *testing.T) {
	ruleset := DefaultRuleset()
	if len(ruleset) == 0 {
		t.Error("default ruleset should not be empty")
	}
}

func TestEvaluate(t *testing.T) {
	ruleset := Ruleset{
		{Permission: "read", Action: Allow},
		{Permission: "write", Action: Ask},
		{Permission: "bash", Action: Deny},
	}

	tests := []struct {
		tool     string
		params   map[string]interface{}
		expected PermissionAction
	}{
		{"read", nil, Allow},
		{"write", nil, Ask},
		{"bash", nil, Deny},
		{"unknown", nil, Ask}, // default
	}

	for _, tt := range tests {
		result := ruleset.Evaluate(tt.tool, tt.params)
		if result != tt.expected {
			t.Errorf("Evaluate(%s) = %v, want %v", tt.tool, result, tt.expected)
		}
	}
}

func TestEvaluateWithPattern(t *testing.T) {
	ruleset := Ruleset{
		{Permission: "write", Action: Allow, Pattern: "/tmp/*"},
		{Permission: "write", Action: Deny, Pattern: "/etc/*"},
	}

	// Test matching pattern
	result := ruleset.Evaluate("write", map[string]interface{}{"path": "/tmp/test.txt"})
	if result != Allow {
		t.Errorf("expected Allow for /tmp/test.txt, got %v", result)
	}

	// Test non-matching pattern
	result = ruleset.Evaluate("write", map[string]interface{}{"path": "/etc/passwd"})
	if result != Deny {
		t.Errorf("expected Deny for /etc/passwd, got %v", result)
	}
}

func TestMerge(t *testing.T) {
	a := Ruleset{
		{Permission: "read", Action: Allow},
	}
	b := Ruleset{
		{Permission: "write", Action: Deny},
	}

	merged := Merge(a, b)
	if len(merged) != 2 {
		t.Errorf("expected 2 rules, got %d", len(merged))
	}
}

func TestSaveAndLoadRuleset(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "rules.json")

	ruleset := Ruleset{
		{Permission: "read", Action: Allow},
		{Permission: "write", Action: Ask},
	}

	// Save
	if err := SaveRuleset(path, ruleset); err != nil {
		t.Fatalf("failed to save ruleset: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("ruleset file should exist")
	}

	// Load
	loaded, err := LoadRuleset(path)
	if err != nil {
		t.Fatalf("failed to load ruleset: %v", err)
	}

	if len(loaded) != 2 {
		t.Errorf("expected 2 rules, got %d", len(loaded))
	}
}

func TestPermissionForToolMapsCoreTools(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"file_read", "read"},
		{"dir_list", "read"},
		{"search", "read"},
		{"glob", "read"},
		{"git_status", "read"},
		{"file_write", "write"},
		{"file_delete", "write"},
		{"dir_create", "write"},
		{"file_edit", "edit"},
		{"file_diff", "edit"},
		{"shell", "bash"},
		{"process", "bash"},
	}

	for _, tt := range tests {
		if got := PermissionForTool(tt.name); got != tt.want {
			t.Fatalf("PermissionForTool(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestRulesetFromConfigUsesDefaultsWhenRulesAreInvalid(t *testing.T) {
	ruleset := RulesetFromConfig([]PermissionRule{
		{Permission: "bash", Action: "launch"},
		{Permission: "", Action: Deny},
	})

	if got := ruleset.Evaluate("read", nil); got != Allow {
		t.Fatalf("read action = %q, want default %q", got, Allow)
	}
	if got := ruleset.Evaluate("bash", nil); got != Ask {
		t.Fatalf("bash action = %q, want default %q", got, Ask)
	}
}

func TestRulesetFromConfigKeepsValidRules(t *testing.T) {
	ruleset := RulesetFromConfig([]PermissionRule{
		{Permission: "bash", Action: Deny},
		{Permission: "write", Action: Allow, Pattern: "/tmp/*"},
	})

	if got := ruleset.Evaluate("bash", nil); got != Deny {
		t.Fatalf("bash action = %q, want %q", got, Deny)
	}
	if got := ruleset.Evaluate("write", map[string]interface{}{"path": "/tmp/file.txt"}); got != Allow {
		t.Fatalf("write action = %q, want %q", got, Allow)
	}
}
