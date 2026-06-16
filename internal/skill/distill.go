package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

	commands := d.detectCommandPatterns(sessionData)
	candidates = append(candidates, commands...)

	workflows := d.detectWorkflowPatterns(sessionData)
	candidates = append(workflows, candidates...)

	var filtered []SkillCandidate
	for _, c := range candidates {
		if c.Confidence >= d.config.MinConfidence {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) > d.config.MaxCandidates {
		filtered = filtered[:d.config.MaxCandidates]
	}

	return filtered, nil
}

func (d *Distill) detectCommandPatterns(data string) []SkillCandidate {
	var candidates []SkillCandidate
	lines := strings.Split(data, "\n")

	cmdCount := make(map[string]int)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "$ ") || strings.HasPrefix(trimmed, "> ") {
			cmd := strings.TrimPrefix(strings.TrimPrefix(trimmed, "$ "), "> ")
			cmdCount[cmd]++
		}
	}
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

	for cmd, count := range cmdCount {
		if count >= 2 {
			candidates = append(candidates, SkillCandidate{
				Name:        fmt.Sprintf("skill_%s", sanitizeName(cmd)),
				Description: fmt.Sprintf("Automated skill for: %s", cmd),
				Confidence:  float64(count) / 10.0,
				Pattern:     cmd,
				Commands:    []string{cmd},
				CreatedAt:   time.Now(),
			})
		}
	}

	return candidates
}

func (d *Distill) detectWorkflowPatterns(data string) []SkillCandidate {
	var candidates []SkillCandidate
	lines := strings.Split(data, "\n")

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

// SaveCandidates saves skill candidates to candidates.md and individual SKILL.md files.
func (d *Distill) SaveCandidates(candidates []SkillCandidate) error {
	if len(candidates) == 0 {
		return nil
	}

	skillDir := filepath.Join(d.projectDir, ".mimo", "skills")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("failed to create skill dir: %w", err)
	}

	if err := d.writeCandidatesFile(skillDir, candidates); err != nil {
		return err
	}
	if err := d.writeSkillFiles(skillDir, candidates); err != nil {
		return err
	}

	return nil
}

func (d *Distill) writeCandidatesFile(skillDir string, candidates []SkillCandidate) error {
	skillFile := filepath.Join(skillDir, "candidates.md")

	var content strings.Builder
	content.WriteString("# Skill Candidates\n\n")
	content.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))

	for _, c := range candidates {
		content.WriteString(fmt.Sprintf("## %s\n\n", normalizeSkillName(c.Name)))
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

func (d *Distill) writeSkillFiles(skillDir string, candidates []SkillCandidate) error {
	for _, candidate := range candidates {
		name := normalizeSkillName(candidate.Name)
		candidateDir := filepath.Join(skillDir, name)
		if err := os.MkdirAll(candidateDir, 0755); err != nil {
			return fmt.Errorf("failed to create skill candidate dir %s: %w", name, err)
		}

		content := renderSkillFile(name, candidate)
		skillFile := filepath.Join(candidateDir, "SKILL.md")
		if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write skill file %s: %w", name, err)
		}
	}
	return nil
}

func renderSkillFile(name string, candidate SkillCandidate) string {
	var content strings.Builder
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("name: %s\n", name))
	content.WriteString(fmt.Sprintf("description: %s\n", candidate.Description))
	content.WriteString(fmt.Sprintf("confidence: %.2f\n", candidate.Confidence))
	content.WriteString("source: distill\n")
	content.WriteString("---\n\n")
	content.WriteString("# ")
	content.WriteString(name)
	content.WriteString("\n\n")
	content.WriteString(candidate.Description)
	content.WriteString("\n\n")
	content.WriteString("## Pattern\n\n")
	if candidate.Pattern == "" {
		content.WriteString("No command pattern captured.\n\n")
	} else {
		content.WriteString("```text\n")
		content.WriteString(candidate.Pattern)
		content.WriteString("\n```\n\n")
	}
	content.WriteString("## Commands\n\n")
	if len(candidate.Commands) == 0 {
		content.WriteString("No commands captured.\n")
	} else {
		for _, cmd := range candidate.Commands {
			content.WriteString("- `")
			content.WriteString(cmd)
			content.WriteString("`\n")
		}
	}
	return content.String()
}

// ParseCandidatesMarkdown parses candidates.md generated by SaveCandidates.
func ParseCandidatesMarkdown(data []byte) []SkillCandidate {
	var candidates []SkillCandidate
	var current *SkillCandidate
	inCommands := false

	flush := func() {
		if current != nil && current.Name != "" {
			candidates = append(candidates, *current)
		}
	}

	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") {
			flush()
			current = &SkillCandidate{Name: strings.TrimSpace(strings.TrimPrefix(trimmed, "## "))}
			inCommands = false
			continue
		}
		if current == nil {
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "- **Description**:"):
			current.Description = strings.TrimSpace(strings.TrimPrefix(trimmed, "- **Description**:"))
			inCommands = false
		case strings.HasPrefix(trimmed, "- **Confidence**:"):
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "- **Confidence**:"))
			if confidence, err := strconv.ParseFloat(value, 64); err == nil {
				current.Confidence = confidence
			}
			inCommands = false
		case strings.HasPrefix(trimmed, "- **Pattern**:"):
			current.Pattern = strings.TrimSpace(strings.TrimPrefix(trimmed, "- **Pattern**:"))
			inCommands = false
		case strings.HasPrefix(trimmed, "- **Commands**:"):
			inCommands = true
		case inCommands && strings.HasPrefix(trimmed, "- `"):
			command := strings.TrimPrefix(trimmed, "- `")
			command = strings.TrimSuffix(command, "`")
			current.Commands = append(current.Commands, command)
		}
	}
	flush()

	return candidates
}

func (d *Distill) Run(sessionDir string) (int, error) {
	sessionFile := filepath.Join(sessionDir, "checkpoint.md")
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read session: %w", err)
	}

	candidates, err := d.Analyze(string(data))
	if err != nil {
		return 0, err
	}

	if err := d.SaveCandidates(candidates); err != nil {
		return 0, err
	}

	return len(candidates), nil
}

func normalizeSkillName(s string) string {
	name := sanitizeName(s)
	name = strings.Trim(name, "_")
	if name == "" {
		return "skill_candidate"
	}
	return name
}

func sanitizeName(s string) string {
	reg := strings.NewReplacer(
		" ", "_",
		"-", "_",
		"/", "_",
		"\\", "_",
		".", "_",
	)
	result := reg.Replace(s)
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}
