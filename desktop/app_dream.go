package desktop

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mimo-cli/mimo-cli/internal/memory"
	"github.com/mimo-cli/mimo-cli/internal/session"
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
	Explanation string   `json:"explanation"`
	Confidence  float64  `json:"confidence"`
	Pattern     string   `json:"pattern,omitempty"`
	Commands    []string `json:"commands,omitempty"`
	Enabled     bool     `json:"enabled"`
}

// DreamRun runs the dream process on the current session
func (a *App) DreamRun() DreamResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentSessionID == "" {
		return DreamResult{Success: false, Message: ""}
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
	sessionID := a.currentSessionID
	a.mu.Unlock()

	if sessionID == "" {
		return DreamResult{Success: false, Message: ""}
	}

	wd, _ := os.Getwd()
	distill := skill.NewDistill(skill.DefaultDistillConfig(), wd)
	if a.sessionStore != nil {
		_, messages, err := a.sessionStore.LoadSession(sessionID)
		if err == nil && len(messages) > 0 {
			count, err := distill.RunText(renderSessionMessagesForDistill(messages))
			if err != nil {
				return DreamResult{Success: false, Message: err.Error()}
			}
			return DreamResult{
				Success: true,
				Message: "",
				Count:   count,
			}
		}
	}

	sessionDir := filepath.Join(wd, ".mimo", "memory", "sessions", sessionID)
	count, err := distill.Run(sessionDir)
	if err != nil {
		return DreamResult{Success: false, Message: err.Error()}
	}

	return DreamResult{
		Success: true,
		Message: "",
		Count:   count,
	}
}

func renderSessionMessagesForDistill(messages []session.Message) string {
	var content strings.Builder
	for _, message := range messages {
		content.WriteString(message.Role)
		content.WriteString(":\n")
		content.WriteString(message.Content)
		content.WriteString("\n\n")
		if message.Thinking != "" {
			content.WriteString("thinking:\n")
			content.WriteString(message.Thinking)
			content.WriteString("\n\n")
		}
		for _, line := range message.ToolLines {
			content.WriteString(line)
			content.WriteByte('\n')
		}
	}
	return content.String()
}

// DistillListCandidates returns skill candidates
func (a *App) DistillListCandidates() []SkillCandidateInfo {
	wd, _ := os.Getwd()
	skillFile := filepath.Join(wd, ".mimo", "skills", "candidates.md")

	data, err := os.ReadFile(skillFile)
	if err != nil {
		return []SkillCandidateInfo{}
	}

	distill := skill.NewDistill(skill.DefaultDistillConfig(), wd)
	enabledNames, err := distill.ListEnabledCandidates()
	enabled := make(map[string]bool, len(enabledNames))
	if err == nil {
		for _, name := range enabledNames {
			enabled[name] = true
		}
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
			Enabled:     enabled[candidate.Name],
		})
	}

	return candidates
}

// DistillEnableCandidate enables a generated skill candidate
func (a *App) DistillEnableCandidate(name string) DreamResult {
	wd, _ := os.Getwd()
	distill := skill.NewDistill(skill.DefaultDistillConfig(), wd)
	if err := distill.EnableCandidate(name); err != nil {
		return DreamResult{Success: false, Message: err.Error()}
	}
	return DreamResult{Success: true, Message: "", Count: 1}
}

// DistillDeleteCandidate deletes a generated skill candidate
func (a *App) DistillDeleteCandidate(name string) DreamResult {
	wd, _ := os.Getwd()
	distill := skill.NewDistill(skill.DefaultDistillConfig(), wd)
	if err := distill.DeleteCandidate(name); err != nil {
		return DreamResult{Success: false, Message: err.Error()}
	}
	return DreamResult{Success: true, Message: "", Count: 1}
}
