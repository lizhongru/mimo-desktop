package tools

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// FileDiffTool compares two files or shows file content with line numbers
type FileDiffTool struct{}

func NewFileDiffTool() *FileDiffTool { return &FileDiffTool{} }

func (t *FileDiffTool) Name() string        { return "file_diff" }
func (t *FileDiffTool) Description() string  { return "Compare two files and show differences" }
func (t *FileDiffTool) GetSafetyLevel() SafetyLevel { return SafetyLow }
func (t *FileDiffTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path_a": map[string]interface{}{
				"type":        "string",
				"description": "First file path",
			},
			"path_b": map[string]interface{}{
				"type":        "string",
				"description": "Second file path",
			},
		},
		"required": []string{"path_a", "path_b"},
	}
}
func (t *FileDiffTool) Validate(params map[string]interface{}) error {
	if _, err := StringParam(params, "path_a"); err != nil {
		return err
	}
	_, err := StringParam(params, "path_b")
	return err
}
func (t *FileDiffTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *FileDiffTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	pathA, _ := StringParam(params, "path_a")
	pathB, _ := StringParam(params, "path_b")

	dataA, err := os.ReadFile(pathA)
	if err != nil {
		return ToolError("cannot read %s: %v", pathA, err), nil
	}
	dataB, err := os.ReadFile(pathB)
	if err != nil {
		return ToolError("cannot read %s: %v", pathB, err), nil
	}

	linesA := strings.Split(string(dataA), "\n")
	linesB := strings.Split(string(dataB), "\n")

	var sb strings.Builder
	maxLines := len(linesA)
	if len(linesB) > maxLines {
		maxLines = len(linesB)
	}

	diffCount := 0
	for i := 0; i < maxLines; i++ {
		var a, b string
		if i < len(linesA) {
			a = linesA[i]
		}
		if i < len(linesB) {
			b = linesB[i]
		}
		if a != b {
			diffCount++
			sb.WriteString(fmt.Sprintf("L%d:\n", i+1))
			if i < len(linesA) {
				sb.WriteString(fmt.Sprintf("  - %s\n", a))
			}
			if i < len(linesB) {
				sb.WriteString(fmt.Sprintf("  + %s\n", b))
			}
		}
	}

	if diffCount == 0 {
		return &ToolResult{Output: "Files are identical"}, nil
	}

	result := fmt.Sprintf("%d differences found:\n%s", diffCount, sb.String())
	if len(result) > 4000 {
		result = result[:4000] + "\n... (truncated)"
	}
	return &ToolResult{Output: result}, nil
}
