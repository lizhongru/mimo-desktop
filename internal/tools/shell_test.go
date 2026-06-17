package tools

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestShellToolTimeoutReturnsPromptly(t *testing.T) {
	command := "sleep 3"
	if runtime.GOOS == "windows" {
		command = "ping -n 6 127.0.0.1 > nul"
	}

	start := time.Now()
	result, err := NewShellTool().Execute(context.Background(), map[string]interface{}{
		"command": command,
		"timeout": 1,
	})
	if err != nil {
		t.Fatalf("execute shell: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 4*time.Second {
		t.Fatalf("timeout returned too slowly: %s", elapsed)
	}
	if result.ExitCode != -1 {
		t.Fatalf("exit code = %d, want -1; result = %#v", result.ExitCode, result)
	}
	if !strings.Contains(result.Error, "timed out after 1s") {
		t.Fatalf("timeout error = %q", result.Error)
	}
}
