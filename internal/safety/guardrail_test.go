package safety

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/mimo-cli/mimo-cli/internal/permission"
)

func TestGuardrailPermissionDenyBlocksShell(t *testing.T) {
	guardrail := NewGuardrail(LevelAuto, NewClassifier(nil, nil, nil), "")
	guardrail.SetRuleset(permission.Ruleset{
		{Permission: "bash", Action: permission.Deny},
	})

	allowed, err := guardrail.CheckWithConfirmAll("shell", map[string]interface{}{
		"command": "echo hello",
	}, false)

	if allowed {
		t.Fatal("CheckWithConfirmAll allowed shell with bash=deny")
	}
	if err == nil || !strings.Contains(err.Error(), "permission bash denied") {
		t.Fatalf("error = %v, want permission bash denied", err)
	}
}

func TestGuardrailPermissionAskConfirmsWrite(t *testing.T) {
	guardrail := NewGuardrail(LevelAuto, NewClassifier(nil, nil, nil), "")
	guardrail.SetRuleset(permission.Ruleset{
		{Permission: "write", Action: permission.Ask},
	})

	confirmed := false
	guardrail.SetConfirmCallback(func(action Action) (bool, error) {
		confirmed = true
		return true, nil
	})

	allowed, err := guardrail.CheckWithConfirmAll("file_write", map[string]interface{}{
		"path": "notes.txt",
	}, false)

	if err != nil {
		t.Fatalf("CheckWithConfirmAll returned error: %v", err)
	}
	if !allowed {
		t.Fatal("CheckWithConfirmAll denied confirmed write")
	}
	if !confirmed {
		t.Fatal("write=ask did not invoke confirmation callback")
	}
}

func TestGuardrailPermissionConfirmAllBypassesAskNotDeny(t *testing.T) {
	guardrail := NewGuardrail(LevelAuto, NewClassifier(nil, nil, nil), "")
	guardrail.SetRuleset(permission.Ruleset{
		{Permission: "write", Action: permission.Ask},
		{Permission: "bash", Action: permission.Deny},
	})

	confirmed := false
	guardrail.SetConfirmCallback(func(action Action) (bool, error) {
		confirmed = true
		return false, nil
	})

	allowed, err := guardrail.CheckWithConfirmAll("file_write", map[string]interface{}{
		"path": "notes.txt",
	}, true)
	if err != nil {
		t.Fatalf("confirmAll write returned error: %v", err)
	}
	if !allowed {
		t.Fatal("confirmAll should allow write=ask")
	}
	if confirmed {
		t.Fatal("confirmAll should bypass ask confirmation callback")
	}

	allowed, err = guardrail.CheckWithConfirmAll("shell", map[string]interface{}{
		"command": "echo hello",
	}, true)
	if allowed {
		t.Fatal("confirmAll allowed shell with bash=deny")
	}
	if err == nil || !strings.Contains(err.Error(), "permission bash denied") {
		t.Fatalf("deny error = %v, want permission bash denied", err)
	}
}

func TestGuardrailPermissionExternalDirectoryIsCheckedBeforeToolCategory(t *testing.T) {
	workspace := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")

	guardrail := NewGuardrail(LevelAuto, NewClassifier(nil, nil, nil), "")
	guardrail.SetWorkspaceRoot(workspace)
	guardrail.SetRuleset(permission.Ruleset{
		{Permission: "external_directory", Action: permission.Deny},
		{Permission: "read", Action: permission.Allow},
	})

	allowed, err := guardrail.CheckWithConfirmAll("file_read", map[string]interface{}{
		"path": outside,
	}, true)
	if allowed {
		t.Fatal("guardrail allowed external path with external_directory=deny")
	}
	if err == nil || !strings.Contains(err.Error(), "permission external_directory denied") {
		t.Fatalf("external path error = %v, want external_directory denied", err)
	}
}

func TestGuardrailPermissionExternalAllowStillChecksToolCategory(t *testing.T) {
	workspace := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")

	guardrail := NewGuardrail(LevelAuto, NewClassifier(nil, nil, nil), "")
	guardrail.SetWorkspaceRoot(workspace)
	guardrail.SetRuleset(permission.Ruleset{
		{Permission: "external_directory", Action: permission.Allow},
		{Permission: "read", Action: permission.Deny},
	})

	allowed, err := guardrail.CheckWithConfirmAll("file_read", map[string]interface{}{
		"path": outside,
	}, true)
	if allowed {
		t.Fatal("guardrail allowed external read after external_directory allow and read=deny")
	}
	if err == nil || !strings.Contains(err.Error(), "permission read denied") {
		t.Fatalf("external read error = %v, want read denied", err)
	}
}
