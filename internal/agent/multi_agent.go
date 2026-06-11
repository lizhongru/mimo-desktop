package agent

import (
	"fmt"
	"sync"
)

// AgentMode represents the mode of an agent
type AgentMode string

const (
	ModeBuild    AgentMode = "build"
	ModePlan     AgentMode = "plan"
	ModeCompose  AgentMode = "compose"
	ModeSubagent AgentMode = "subagent"
)

// AgentConfig defines configuration for a specific agent type
type AgentConfig struct {
	Name          string    `json:"name"`
	Mode          AgentMode `json:"mode"`
	Color         string    `json:"color"`
	Description   string    `json:"description"`
	Prompt        string    `json:"prompt"`
	ToolAllowlist []string  `json:"tool_allowlist,omitempty"`
	ToolDenylist  []string  `json:"tool_denylist,omitempty"`
	MaxTokens     int       `json:"max_tokens,omitempty"`
	Temperature   float64   `json:"temperature,omitempty"`
}

// DefaultAgentConfigs returns the default agent configurations
func DefaultAgentConfigs() map[string]*AgentConfig {
	return map[string]*AgentConfig{
		"build": {
			Name:        "Build",
			Mode:        ModeBuild,
			Color:       "#10b981",
			Description: "Build and implement code changes",
			Prompt:      "You are a build agent. Focus on implementing code changes, writing tests, and fixing bugs.",
		},
		"plan": {
			Name:        "Plan",
			Mode:        ModePlan,
			Color:       "#3b82f6",
			Description: "Plan and design solutions",
			Prompt:      "You are a planning agent. Focus on analyzing requirements, designing solutions, and creating implementation plans.",
		},
		"compose": {
			Name:        "Compose",
			Mode:        ModeCompose,
			Color:       "#8b5cf6",
			Description: "Compose and orchestrate tasks",
			Prompt:      "You are a compose agent. Focus on orchestrating multiple sub-agents and coordinating complex workflows.",
		},
	}
}

// MultiAgentManager manages multiple agent configurations
type MultiAgentManager struct {
	mu      sync.RWMutex
	configs map[string]*AgentConfig
	current string
}

// NewMultiAgentManager creates a new multi-agent manager
func NewMultiAgentManager() *MultiAgentManager {
	return &MultiAgentManager{
		configs: DefaultAgentConfigs(),
		current: "build",
	}
}

// GetConfig returns an agent configuration by name
func (m *MultiAgentManager) GetConfig(name string) (*AgentConfig, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	config, ok := m.configs[name]
	return config, ok
}

// SetConfig sets an agent configuration
func (m *MultiAgentManager) SetConfig(name string, config *AgentConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.configs[name] = config
}

// ListConfigs returns all agent configurations
func (m *MultiAgentManager) ListConfigs() map[string]*AgentConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]*AgentConfig)
	for k, v := range m.configs {
		result[k] = v
	}
	return result
}

// SetCurrent sets the current active agent
func (m *MultiAgentManager) SetCurrent(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.configs[name]; !ok {
		return fmt.Errorf("agent %s not found", name)
	}
	m.current = name
	return nil
}

// GetCurrent returns the current active agent configuration
func (m *MultiAgentManager) GetCurrent() *AgentConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.configs[m.current]
}

// GetCurrentName returns the name of the current active agent
func (m *MultiAgentManager) GetCurrentName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}
