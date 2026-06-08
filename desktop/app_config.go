package desktop

import (
	"github.com/mimo-cli/mimo-cli/internal/agent"
	iconfig "github.com/mimo-cli/mimo-cli/internal/config"
	"github.com/mimo-cli/mimo-cli/internal/safety"
)

// AppConfigDTO is a frontend-friendly config representation.
type AppConfigDTO struct {
	DefaultModel string              `json:"defaultModel"`
	Language     string              `json:"language"`
	Theme        string              `json:"theme"`
	UserName     string              `json:"userName"`
	Models       map[string]ModelDTO `json:"models"`
	Safety       SafetyDTO           `json:"safety"`
	Agent        AgentDTO            `json:"agent"`
}

// ModelDTO is a frontend-friendly model config.
type ModelDTO struct {
	APIBase     string  `json:"apiBase"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"maxTokens"`
	Temperature float64 `json:"temperature"`
}

// SafetyDTO is frontend-friendly safety config.
type SafetyDTO struct {
	Level      string `json:"level"`
	Permission string `json:"permission"`
}

// AgentDTO is frontend-friendly agent config.
type AgentDTO struct {
	MaxIterations  int    `json:"maxIterations"`
	PlanningMode   string `json:"planningMode"`
	ShowTokenUsage bool   `json:"showTokenUsage"`
}

// GetConfig returns the current configuration.
func (a *App) GetConfig() AppConfigDTO {
	models := make(map[string]ModelDTO)
	for name, m := range a.cfg.Models {
		models[name] = ModelDTO{
			APIBase:     m.APIBase,
			Model:       m.Model,
			MaxTokens:   m.MaxTokens,
			Temperature: m.Temperature,
		}
	}
	return AppConfigDTO{
		DefaultModel: a.cfg.DefaultModel,
		Language:     a.cfg.Language,
		Theme:        a.cfg.Theme,
		UserName:     a.cfg.UserName,
		Models:       models,
		// Permission is stored on AgentConfig but exposed in SafetyDTO
		// because it controls the safety guardrail's permission level.
		Safety: SafetyDTO{
			Level:      a.cfg.Safety.Level,
			Permission: a.cfg.Agent.Permission,
		},
		Agent: AgentDTO{
			MaxIterations:  a.cfg.Agent.MaxIterations,
			PlanningMode:   a.cfg.Agent.PlanningMode,
			ShowTokenUsage: a.cfg.Agent.ShowTokenUsage,
		},
	}
}

// SetTheme changes the theme and saves config.
func (a *App) SetTheme(theme string) error {
	a.cfg.Theme = theme
	return iconfig.SaveUserConfig(a.cfg)
}

// SetLanguage changes the language and saves config.
func (a *App) SetLanguage(lang string) error {
	a.cfg.Language = lang
	return iconfig.SaveUserConfig(a.cfg)
}

// SetDefaultModel changes the default model and saves config.
func (a *App) SetDefaultModel(name string) error {
	a.cfg.DefaultModel = name
	if err := a.gateway.SetCurrentModel(name); err != nil {
		return err
	}
	return iconfig.SaveUserConfig(a.cfg)
}

// AddModel adds a new model configuration.
func (a *App) AddModel(name, apiBase, apiKey, model string, maxTokens int, temperature float64) error {
	a.cfg.Models[name] = iconfig.ModelConfig{
		APIBase:     apiBase,
		APIKey:      apiKey,
		Model:       model,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}
	return iconfig.SaveUserConfig(a.cfg)
}

// RemoveModel removes a model configuration.
func (a *App) RemoveModel(name string) error {
	delete(a.cfg.Models, name)
	return iconfig.SaveUserConfig(a.cfg)
}

// SetSafetyLevel changes the safety level.
func (a *App) SetSafetyLevel(level string) error {
	a.cfg.Safety.Level = level
	// Re-create guardrail with new level
	classifier := safety.NewClassifier(
		a.cfg.Safety.BlockedCommands,
		a.cfg.Safety.ProtectedFiles,
		a.cfg.Safety.ProtectedBranches,
	)
	a.guardrail = safety.NewGuardrail(safety.SafetyLevel(level), classifier, a.cfg.Safety.AuditLog)
	a.guardrail.SetPermission(a.cfg.Agent.Permission)
	a.registerConfirmCallback()
	return iconfig.SaveUserConfig(a.cfg)
}

// SetPlanningMode changes the agent planning mode.
func (a *App) SetPlanningMode(mode string) error {
	a.cfg.Agent.PlanningMode = mode
	switch mode {
	case "react":
		a.agent.SetPlanningMode(agent.ModeReact)
	case "plan-execute":
		a.agent.SetPlanningMode(agent.ModePlanExecute)
	default:
		a.agent.SetPlanningMode(agent.ModeAuto)
	}
	return iconfig.SaveUserConfig(a.cfg)
}
