package desktop

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mimo-cli/mimo-cli/internal/agent"
	iconfig "github.com/mimo-cli/mimo-cli/internal/config"
	"github.com/mimo-cli/mimo-cli/internal/llm"
	"github.com/mimo-cli/mimo-cli/internal/safety"
)

// maskAPIKey masks the API key for display (show first 8 chars and last 4)
func maskAPIKey(key string) string {
	if len(key) <= 12 {
		return key
	}
	return key[:8] + "..." + key[len(key)-4:]
}

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
	// Provider info
	Provider string   `json:"provider"`
	Website  string   `json:"website"`
	
	// API settings
	APIBase  string   `json:"apiBase"`
	APIKey   string   `json:"apiKey"`
	
	// Model settings
	Model    string   `json:"model"`
	Models   []string `json:"models"`
	Fallback string   `json:"fallback"`
	
	// Generation parameters
	MaxTokens   int     `json:"maxTokens"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"topP"`
	
	// Features
	Streaming bool `json:"streaming"`
	Vision    bool `json:"vision"`
	Tools     bool `json:"tools"`
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
	Permission     string `json:"permission"`
	ReasoningLevel string `json:"reasoningLevel"`
	ShowTokenUsage bool   `json:"showTokenUsage"`
}

// GetConfig returns the current configuration.
func (a *App) GetConfig() AppConfigDTO {
	models := make(map[string]ModelDTO)
	for name, m := range a.cfg.Models {
		// Provide defaults for new fields if empty
		availableModels := m.Models
		if availableModels == nil {
			availableModels = []string{}
		}
		topP := m.TopP
		if topP == 0 {
			topP = 0.95
		}
		
		models[name] = ModelDTO{
			Provider:    m.Provider,
			Website:     m.Website,
			APIBase:     m.APIBase,
			APIKey:      m.APIKey,
			Model:       m.Model,
			Models:      availableModels,
			Fallback:    m.Fallback,
			MaxTokens:   m.MaxTokens,
			Temperature: m.Temperature,
			TopP:        topP,
			Streaming:   m.Streaming,
			Vision:      m.Vision,
			Tools:       m.Tools,
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
			Permission:     a.cfg.Agent.Permission,
			ReasoningLevel: a.cfg.Agent.ReasoningLevel,
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
func (a *App) AddModel(name, provider, website, apiBase, apiKey, model string, models []string, fallback string, maxTokens int, temperature float64, topP float64, streaming, vision, tools bool) error {
	a.cfg.Models[name] = iconfig.ModelConfig{
		Provider:    provider,
		Website:     website,
		APIBase:     apiBase,
		APIKey:      apiKey,
		Model:       model,
		Models:      models,
		Fallback:    fallback,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		TopP:        topP,
		Streaming:   streaming,
		Vision:      vision,
		Tools:       tools,
	}
	return iconfig.SaveUserConfig(a.cfg)
}

// RemoveModel removes a model configuration.
func (a *App) RemoveModel(name string) error {
	delete(a.cfg.Models, name)
	return iconfig.SaveUserConfig(a.cfg)
}

// ListRemoteModels fetches available models from the provider API
func (a *App) ListRemoteModels(modelName string) ([]llm.ModelInfo, error) {
	ctx := context.Background()
	return a.gateway.ListRemoteModels(ctx, modelName)
}

// ListRemoteModelsWithConfig fetches available models using provided API config.
// Uses a short-lived context (15 s) so the UI doesn't hang indefinitely.
func (a *App) ListRemoteModelsWithConfig(apiBase, apiKey string) ([]llm.ModelInfo, error) {
	if apiBase == "" || apiKey == "" {
		return nil, fmt.Errorf("API base and key are required")
	}
	
	// Create a temporary provider to fetch models
	cfg := iconfig.ModelConfig{
		APIBase: apiBase,
		APIKey:  apiKey,
	}
	
	provider := llm.NewOpenAIProvider(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return provider.ListModels(ctx)
}

// UpdateModel updates an existing model configuration.
func (a *App) UpdateModel(name, provider, website, apiBase, apiKey, model string, models []string, fallback string, maxTokens int, temperature float64, topP float64, streaming, vision, tools bool) error {
	if _, exists := a.cfg.Models[name]; !exists {
		return fmt.Errorf("model %q not found", name)
	}
	
	// If API key is masked (contains "..."), keep the original
	existing := a.cfg.Models[name]
	if strings.Contains(apiKey, "...") {
		apiKey = existing.APIKey
	}
	
	a.cfg.Models[name] = iconfig.ModelConfig{
		Provider:    provider,
		Website:     website,
		APIBase:     apiBase,
		APIKey:      apiKey,
		Model:       model,
		Models:      models,
		Fallback:    fallback,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		TopP:        topP,
		Streaming:   streaming,
		Vision:      vision,
		Tools:       tools,
	}
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

// SetPermission changes the agent permission mode (readonly, write, exec).
func (a *App) SetPermission(perm string) error {
	a.cfg.Agent.Permission = perm
	a.guardrail.SetPermission(perm)
	return iconfig.SaveUserConfig(a.cfg)
}

// SetReasoningLevel changes the reasoning effort level (low, medium, high).
func (a *App) SetReasoningLevel(level string) error {
	a.cfg.Agent.ReasoningLevel = level
	a.agent.SetReasoningLevel(level)
	return iconfig.SaveUserConfig(a.cfg)
}
