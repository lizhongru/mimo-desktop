package desktop

import (
	"github.com/mimo-cli/mimo-cli/internal/agent"
)

// AgentConfigInfo represents agent configuration for the frontend
type AgentConfigInfo struct {
	Name          string   `json:"name"`
	Mode          string   `json:"mode"`
	Color         string   `json:"color"`
	Description   string   `json:"description"`
	Prompt        string   `json:"prompt"`
	ToolAllowlist []string `json:"tool_allowlist,omitempty"`
}

// AgentSwitchResult represents the result of switching agents
type AgentSwitchResult struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Agent   *AgentConfigInfo `json:"agent,omitempty"`
}

// multiAgentManager holds the multi-agent manager (singleton)
var multiAgentManager *agent.MultiAgentManager

func getMultiAgentManager() *agent.MultiAgentManager {
	if multiAgentManager == nil {
		multiAgentManager = agent.NewMultiAgentManager()
	}
	return multiAgentManager
}

// AgentListConfigs returns all agent configurations
func (a *App) AgentListConfigs() []AgentConfigInfo {
	manager := getMultiAgentManager()
	configs := manager.ListConfigs()

	var result []AgentConfigInfo
	for _, config := range configs {
		result = append(result, AgentConfigInfo{
			Name:          config.Name,
			Mode:          string(config.Mode),
			Color:         config.Color,
			Description:   config.Description,
			Prompt:        config.Prompt,
			ToolAllowlist: config.ToolAllowlist,
		})
	}
	return result
}

// AgentGetCurrent returns the current active agent
func (a *App) AgentGetCurrent() *AgentConfigInfo {
	manager := getMultiAgentManager()
	config := manager.GetCurrent()
	if config == nil {
		return nil
	}
	return agentConfigInfoFromConfig(config)
}

// AgentSwitch switches to a different agent
func (a *App) AgentSwitch(name string) AgentSwitchResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	manager := getMultiAgentManager()
	if err := manager.SetCurrent(name); err != nil {
		return AgentSwitchResult{Success: false, Message: err.Error()}
	}

	config := manager.GetCurrent()

	if a.agent != nil && config != nil {
		a.agent.SetToolAllowlist(config.ToolAllowlist)
	}

	return AgentSwitchResult{
		Success: true,
		Message: "Switched to " + name + " agent",
		Agent:   agentConfigInfoFromConfig(config),
	}
}

// AgentUpdateConfig updates an agent configuration
func (a *App) AgentUpdateConfig(name string, config AgentConfigInfo) AgentSwitchResult {
	manager := getMultiAgentManager()
	manager.SetConfig(name, &agent.AgentConfig{
		Name:          config.Name,
		Mode:          agent.AgentMode(config.Mode),
		Color:         config.Color,
		Description:   config.Description,
		Prompt:        config.Prompt,
		ToolAllowlist: config.ToolAllowlist,
	})
	return AgentSwitchResult{Success: true, Message: "Agent config updated"}
}

func agentConfigInfoFromConfig(config *agent.AgentConfig) *AgentConfigInfo {
	if config == nil {
		return nil
	}
	return &AgentConfigInfo{
		Name:          config.Name,
		Mode:          string(config.Mode),
		Color:         config.Color,
		Description:   config.Description,
		Prompt:        config.Prompt,
		ToolAllowlist: config.ToolAllowlist,
	}
}
