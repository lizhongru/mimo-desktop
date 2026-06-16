package desktop

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mimo-cli/mimo-cli/internal/memory"
	"github.com/mimo-cli/mimo-cli/internal/skill"
)

// DreamResult represents the result of a dream operation
type DreamResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Count   int    `json:"count"`
}

// SkillCandidateInfo represents a skill candidate for the frontend
type SkillCandidateInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Confidence  float64  `json:"confidence"`
	Pattern     string   `json:"pattern,omitempty"`
	Commands    []string `json:"commands,omitempty"`
}

// DreamRun runs the dream process on the current session
func (a *App) DreamRun() DreamResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentSessionID == "" {
		return DreamResult{Success: false, Message: "No active session"}
	}

	wd, _ := os.Getwd()
	dream := memory.NewDream(memory.DefaultDreamConfig(), wd)

	sessionDir := filepath.Join(wd, ".mimo", "memory", "sessions", a.currentSessionID)
	count, err := dream.Run(sessionDir)
	if err != nil {
		return DreamResult{Success: false, Message: fmt.Sprintf("Dream failed: %v", err)}
	}

	return DreamResult{
		Success: true,
		Message: fmt.Sprintf("Extracted %d memory entries", count),
		Count:   count,
	}
}

// DistillRun runs the distill process on the current session
func (a *App) DistillRun() DreamResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentSessionID == "" {
		return DreamResult{Success: false, Message: "No active session"}
	}

	wd, _ := os.Getwd()
	distill := skill.NewDistill(skill.DefaultDistillConfig(), wd)

	sessionDir := filepath.Join(wd, ".mimo", "memory", "sessions", a.currentSessionID)
	count, err := distill.Run(sessionDir)
	if err != nil {
		return DreamResult{Success: false, Message: fmt.Sprintf("Distill failed: %v", err)}
	}

	return DreamResult{
		Success: true,
		Message: fmt.Sprintf("Found %d skill candidates", count),
		Count:   count,
	}
}

// DistillListCandidates returns skill candidates
func (a *App) DistillListCandidates() []SkillCandidateInfo {
	wd, _ := os.Getwd()
	skillFile := filepath.Join(wd, ".mimo", "skills", "candidates.md")

	data, err := os.ReadFile(skillFile)
	if err != nil {
		return []SkillCandidateInfo{}
	}

	parsed := skill.ParseCandidatesMarkdown(data)
	candidates := make([]SkillCandidateInfo, 0, len(parsed))
	for _, candidate := range parsed {
		candidates = append(candidates, SkillCandidateInfo{
			Name:        candidate.Name,
			Description: candidate.Description,
			Confidence:  candidate.Confidence,
			Pattern:     candidate.Pattern,
			Commands:    candidate.Commands,
		})
	}

	return candidates
}
