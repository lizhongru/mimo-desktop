package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DistillConfig holds distill configuration
type DistillConfig struct {
	Enabled       bool    `yaml:"enabled"`
	MinConfidence float64 `yaml:"min_confidence"`
	MaxCandidates int     `yaml:"max_candidates"`
}

// DefaultDistillConfig returns default distill configuration
func DefaultDistillConfig() DistillConfig {
	return DistillConfig{
		Enabled:       true,
		MinConfidence: 0.6,
		MaxCandidates: 10,
	}
}

// SkillCandidate represents a potential skill to be distilled
type SkillCandidate struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Confidence  float64   `json:"confidence"`
	Pattern     string    `json:"pattern"`
	Commands    []string  `json:"commands,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// Distill extracts reusable skills from session history
type Distill struct {
	config     DistillConfig
	projectDir string
}

// NewDistill creates a new distill instance
func NewDistill(config DistillConfig, projectDir string) *Distill {
	return &Distill{
		config:     config,
		projectDir: projectDir,
	}
}

// Analyze analyzes session data for skill candidates
func (d *Distill) Analyze(sessionData string) ([]SkillCandidate, error) {
	if !d.config.Enabled {
		return nil, nil
	}

	var candidates []SkillCandidate

	// Detect repeated command patterns
	commands := d.detectCommandPatterns(sessionData)
	candidates = append(candidates, commands...)

	// Detect workflow patterns
	workflows := d.detectWorkflowPatterns(sessionData)
	candidates = append(workflows, candidates...)

	// Filter by confidence
	var filtered []SkillCandidate
	for _, c := range candidates {
		if c.Confidence >= d.config.MinConfidence {
			filtered = append(filtered, c)
		}
	}

	// Limit candidates
	if len(filtered) > d.config.MaxCandidates {
		filtered = filtered[:d.config.MaxCandidates]
	}

	return filtered, nil
}

// detectCommandPatterns detects repeated command patterns
func (d *Distill) detectCommandPatterns(data string) []SkillCandidate {
	var candidates []SkillCandidate
	lines := strings.Split(data, "\n")

	// Count command occurrences
	cmdCount := make(map[string]int)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "$ ") || strings.HasPrefix(trimmed, "> ") {
			cmd := strings.TrimPrefix(strings.TrimPrefix(trimmed, "$ "), "> ")
			cmdCount[cmd]++
		}
	}
	// Also check for lines that look like commands (start with common command prefixes)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "go ") ||
			strings.HasPrefix(trimmed, "npm ") ||
			strings.HasPrefix(trimmed, "yarn ") ||
			strings.HasPrefix(trimmed, "pip ") ||
			strings.HasPrefix(trimmed, "cargo ") {
			cmdCount[trimmed]++
		}
	}

	// Create candidates for repeated commands
	for cmd, count := range cmdCount {
		if count >= 2 {
			candidates = append(candidates, SkillCandidate{
				Name:        fmt.Sprintf("skill_%s", sanitizeName(cmd)),
				Description: fmt.Sprintf("Automated skill for: %s", cmd),
				Confidence:  float64(count) / 10.0, // Simple confidence score
				Pattern:     cmd,
				Commands:    []string{cmd},
				CreatedAt:   time.Now(),
			})
		}
	}

	return candidates
}

// detectWorkflowPatterns detects workflow patterns
func (d *Distill) detectWorkflowPatterns(data string) []SkillCandidate {
	var candidates []SkillCandidate
	lines := strings.Split(data, "\n")

	// Detect common workflows
	workflowPatterns := []struct {
		keywords []string
		name     string
		desc     string
	}{
		{
			keywords: []string{"test", "run", "pass"},
			name:     "test_workflow",
			desc:     "Test execution workflow",
		},
		{
			keywords: []string{"build", "compile", "make"},
			name:     "build_workflow",
			desc:     "Build workflow",
		},
		{
			keywords: []string{"deploy", "push", "release"},
			name:     "deploy_workflow",
			desc:     "Deployment workflow",
		},
	}

	for _, pattern := range workflowPatterns {
		count := 0
		for _, line := range lines {
			lower := strings.ToLower(line)
			for _, keyword := range pattern.keywords {
				if strings.Contains(lower, keyword) {
					count++
					break
				}
			}
		}

		if count >= 2 {
			candidates = append(candidates, SkillCandidate{
				Name:        pattern.name,
				Description: pattern.desc,
				Confidence:  float64(count) / 10.0,
				CreatedAt:   time.Now(),
			})
		}
	}

	return candidates
}

// SaveCandidates saves skill candidates to a file
func (d *Distill) SaveCandidates(candidates []SkillCandidate) error {
	if len(candidates) == 0 {
		return nil
	}

	skillDir := filepath.Join(d.projectDir, ".mimo", "skills")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("failed to create skill dir: %w", err)
	}

	skillFile := filepath.Join(skillDir, "candidates.md")

	var content strings.Builder
	content.WriteString("# Skill Candidates\n\n")
	content.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))

	for _, c := range candidates {
		content.WriteString(fmt.Sprintf("## %s\n\n", c.Name))
		content.WriteString(fmt.Sprintf("- **Description**: %s\n", c.Description))
		content.WriteString(fmt.Sprintf("- **Confidence**: %.2f\n", c.Confidence))
		if c.Pattern != "" {
			content.WriteString(fmt.Sprintf("- **Pattern**: %s\n", c.Pattern))
		}
		if len(c.Commands) > 0 {
			content.WriteString("- **Commands**:\n")
			for _, cmd := range c.Commands {
				content.WriteString(fmt.Sprintf("  - `%s`\n", cmd))
			}
		}
		content.WriteString("\n")
	}

	return os.WriteFile(skillFile, []byte(content.String()), 0644)
}

// Run executes the distill process on session history
func (d *Distill) Run(sessionDir string) (int, error) {
	// Read session files
	sessionFile := filepath.Join(sessionDir, "checkpoint.md")
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read session: %w", err)
	}

	// Analyze for candidates
	candidates, err := d.Analyze(string(data))
	if err != nil {
		return 0, err
	}

	// Save candidates
	if err := d.SaveCandidates(candidates); err != nil {
		return 0, err
	}

	return len(candidates), nil
}

// sanitizeName sanitizes a string for use as a name
func sanitizeName(s string) string {
	// Replace special characters with underscores
	reg := strings.NewReplacer(
		" ", "_",
		"-", "_",
		"/", "_",
		"\\", "_",
		".", "_",
	)
	result := reg.Replace(s)
	// Truncate if too long
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}
