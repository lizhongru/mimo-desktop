package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load loads the configuration from files and environment variables
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// 1. Load system config
	systemPath := "/etc/mimo/config.yaml"
	loadFromFile(cfg, systemPath)

	// 2. Load user config
	userPath, err := GetConfigFilePath()
	if err == nil {
		loadFromFile(cfg, userPath)
	}

	// 3. Load project config (./mimo.yaml)
	loadFromFile(cfg, "mimo.yaml")

	// 4. Expand environment variables in API keys
	expandEnvVars(cfg)

	return cfg, nil
}

// LoadFromFile loads configuration from a specific file path
func LoadFromFile(path string) (*Config, error) {
	cfg := DefaultConfig()
	if err := loadFromFile(cfg, path); err != nil {
		return nil, err
	}
	expandEnvVars(cfg)
	return cfg, nil
}

// SaveUserConfig saves the configuration to the user config file
func SaveUserConfig(cfg *Config) error {
	configPath, err := GetConfigFilePath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Ensure config directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// loadFromFile loads config from a YAML file, merging into existing config
func loadFromFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, skip
		}
		return err
	}

	var fileCfg Config
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	mergeConfig(cfg, &fileCfg)
	return nil
}

// mergeConfig merges src into dst with deep merge semantics:
// - Scalar fields: src overrides dst only if non-zero
// - Map fields: src entries are added to dst (existing dst keys preserved)
// - Slice fields: if src has a non-nil slice, it replaces dst's slice
func mergeConfig(dst, src *Config) {
	// Scalar fields
	if src.DefaultModel != "" {
		dst.DefaultModel = src.DefaultModel
	}
	if src.Language != "" {
		dst.Language = src.Language
	}
	if src.Theme != "" {
		dst.Theme = src.Theme
	}
	if src.Editor != "" {
		dst.Editor = src.Editor
	}
	if src.Shell != "" {
		dst.Shell = src.Shell
	}
	if src.UserName != "" {
		dst.UserName = src.UserName
	}

	// Models map — merge entries
	for k, v := range src.Models {
		if dst.Models == nil {
			dst.Models = make(map[string]ModelConfig)
		}
		dst.Models[k] = v
	}

	// Safety — merge fields
	if src.Safety.Level != "" {
		dst.Safety.Level = src.Safety.Level
	}
	if src.Safety.BackupDir != "" {
		dst.Safety.BackupDir = src.Safety.BackupDir
	}
	if src.Safety.AuditLog != "" {
		dst.Safety.AuditLog = src.Safety.AuditLog
	}
	if src.Safety.BlockedCommands != nil {
		dst.Safety.BlockedCommands = src.Safety.BlockedCommands
	}
	if src.Safety.ProtectedFiles != nil {
		dst.Safety.ProtectedFiles = src.Safety.ProtectedFiles
	}
	if src.Safety.ProtectedBranches != nil {
		dst.Safety.ProtectedBranches = src.Safety.ProtectedBranches
	}

	// Agent — merge fields
	if src.Agent.MaxIterations != 0 {
		dst.Agent.MaxIterations = src.Agent.MaxIterations
	}
	if src.Agent.MaxParallelTools != 0 {
		dst.Agent.MaxParallelTools = src.Agent.MaxParallelTools
	}
	if src.Agent.PlanningMode != "" {
		dst.Agent.PlanningMode = src.Agent.PlanningMode
	}
	if src.Agent.Permission != "" {
		dst.Agent.Permission = src.Agent.Permission
	}
	// Bool fields — always override (can't distinguish "not set" from "false")
	dst.Agent.AutoConfirmLowRisk = src.Agent.AutoConfirmLowRisk
	dst.Agent.ShowTokenUsage = src.Agent.ShowTokenUsage
	dst.Agent.ShowCost = src.Agent.ShowCost
	dst.Agent.Verbose = src.Agent.Verbose


	// MCP — merge servers
	if src.MCP.Servers != nil {
		if dst.MCP.Servers == nil {
			dst.MCP.Servers = make(map[string]MCPServerConfig)
		}
		for k, v := range src.MCP.Servers {
			dst.MCP.Servers[k] = v
		}
	}
	// Context — merge fields
	if src.Context.MaxTokens != 0 {
		dst.Context.MaxTokens = src.Context.MaxTokens
	}
	if src.Context.DirectoryTreeDepth != 0 {
		dst.Context.DirectoryTreeDepth = src.Context.DirectoryTreeDepth
	}
	if src.Context.IgnorePatterns != nil {
		dst.Context.IgnorePatterns = src.Context.IgnorePatterns
	}
}

// expandEnvVars expands ${VAR} patterns in API keys
func expandEnvVars(cfg *Config) {
	for name, model := range cfg.Models {
		model.APIKey = expandString(model.APIKey)
		cfg.Models[name] = model
	}
}

// expandString expands ${VAR} patterns in a string
func expandString(s string) string {
	if !strings.Contains(s, "${") {
		return s
	}
	return os.Expand(s, func(key string) string {
		return os.Getenv(key)
	})
}

// GetModelConfig returns the config for a specific model, or the default model
func (c *Config) GetModelConfig(modelName string) (ModelConfig, error) {
	if modelName == "" {
		modelName = c.DefaultModel
	}
	mc, ok := c.Models[modelName]
	if !ok {
		return ModelConfig{}, fmt.Errorf("model %q not found in config", modelName)
	}
	return mc, nil
}

// GetShell returns the configured shell
func (c *Config) GetShell() string {
	if c.Shell != "" {
		return c.Shell
	}
	return defaultShell()
}
