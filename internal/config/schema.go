package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// Config represents the complete MiMo CLI configuration
type Config struct {
	// Basic
	DefaultModel string `yaml:"default_model" mapstructure:"default_model"`
	Language     string `yaml:"language" mapstructure:"language"`
	Theme        string `yaml:"theme" mapstructure:"theme"`
	Editor       string `yaml:"editor" mapstructure:"editor"`
	Shell        string `yaml:"shell" mapstructure:"shell"`
	UserName     string `yaml:"user_name" mapstructure:"user_name"`

	// Models
	Models map[string]ModelConfig `yaml:"models" mapstructure:"models"`

	// Safety
	Safety SafetyConfig `yaml:"safety" mapstructure:"safety"`

	// Agent behavior
	Agent AgentConfig `yaml:"agent" mapstructure:"agent"`

	// MCP servers
	MCP MCPConfig `yaml:"mcp" mapstructure:"mcp"`

	// Context
	Context ContextConfig `yaml:"context" mapstructure:"context"`
}

// ModelConfig represents a single model provider configuration
// ModelConfig represents a single model provider configuration
type ModelConfig struct {
	// Provider info
	Provider    string `yaml:"provider" mapstructure:"provider"`       // e.g. "OpenAI", "MiMo", "Custom"
	Website     string `yaml:"website" mapstructure:"website"`         // Provider website URL
	
	// API settings
	APIBase     string  `yaml:"api_base" mapstructure:"api_base"`
	APIKey      string  `yaml:"api_key" mapstructure:"api_key"`
	
	// Model settings
	Model       string  `yaml:"model" mapstructure:"model"`            // Default model ID
	Models      []string `yaml:"models" mapstructure:"models"`         // Available model IDs
	Fallback    string   `yaml:"fallback" mapstructure:"fallback"`     // Fallback model ID
	
	// Generation parameters
	MaxTokens   int     `yaml:"max_tokens" mapstructure:"max_tokens"`
	Temperature float64 `yaml:"temperature" mapstructure:"temperature"`
	TopP        float64 `yaml:"top_p" mapstructure:"top_p"`
	
	// Features
	Streaming   bool    `yaml:"streaming" mapstructure:"streaming"`    // Support streaming
	Vision      bool    `yaml:"vision" mapstructure:"vision"`          // Support vision/images
	Tools       bool    `yaml:"tools" mapstructure:"tools"`            // Support function calling
}

// SafetyConfig represents safety-related configuration
type SafetyConfig struct {
	Level             string   `yaml:"level" mapstructure:"level"`
	BackupBeforeWrite bool     `yaml:"backup_before_write" mapstructure:"backup_before_write"`
	BackupDir         string   `yaml:"backup_dir" mapstructure:"backup_dir"`
	AuditLog          string   `yaml:"audit_log" mapstructure:"audit_log"`
	BlockedCommands   []string `yaml:"blocked_commands" mapstructure:"blocked_commands"`
	ProtectedFiles    []string `yaml:"protected_files" mapstructure:"protected_files"`
	ProtectedBranches []string `yaml:"protected_branches" mapstructure:"protected_branches"`
}

// AgentConfig represents agent behavior configuration
type AgentConfig struct {
	MaxIterations      int    `yaml:"max_iterations" mapstructure:"max_iterations"`
	MaxParallelTools   int    `yaml:"max_parallel_tools" mapstructure:"max_parallel_tools"`
	PlanningMode       string `yaml:"planning_mode" mapstructure:"planning_mode"`
	Permission         string `yaml:"permission" mapstructure:"permission"`
	ReasoningLevel     string `yaml:"reasoning_level" mapstructure:"reasoning_level"`
	AutoConfirmLowRisk bool   `yaml:"auto_confirm_low_risk" mapstructure:"auto_confirm_low_risk"`
	ShowTokenUsage     bool   `yaml:"show_token_usage" mapstructure:"show_token_usage"`
	ShowCost           bool   `yaml:"show_cost" mapstructure:"show_cost"`
	Verbose            bool   `yaml:"verbose" mapstructure:"verbose"`
}

// ContextConfig represents context management configuration
type ContextConfig struct {
	MaxTokens          int      `yaml:"max_tokens" mapstructure:"max_tokens"`
	DirectoryTreeDepth int      `yaml:"directory_tree_depth" mapstructure:"directory_tree_depth"`
	IgnorePatterns     []string `yaml:"ignore_patterns" mapstructure:"ignore_patterns"`
}

// MCPConfig represents MCP server configurations
type MCPConfig struct {
	Servers map[string]MCPServerConfig `yaml:"servers" mapstructure:"servers"`
}

// MCPServerConfig represents a single MCP server
type MCPServerConfig struct {
	Command string            `yaml:"command" mapstructure:"command"`
	Args    []string          `yaml:"args" mapstructure:"args"`
	URL     string            `yaml:"url" mapstructure:"url"`
	Env     map[string]string `yaml:"env" mapstructure:"env"`
	Enabled bool              `yaml:"enabled" mapstructure:"enabled"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultModel: "mimo",
		Language:     "zh-CN",
		Theme:        "dark",
		Editor:       "code",
		Shell:        defaultShell(),
		Models: map[string]ModelConfig{
			"mimo": {
				APIBase:     "https://api.mimo.xiaomi.com/v1",
				Model:       "mimo-v2",
				MaxTokens:   32768,
				Temperature: 0.3,
				TopP:        0.95,
			},
		},
		Safety: SafetyConfig{
			Level:             "confirm",
			BackupBeforeWrite: true,
			BackupDir:         ".mimo/backups",
			AuditLog:          "~/.mimo/audit.log",
			BlockedCommands:   []string{"sudo", "chmod 777"},
			ProtectedFiles:    []string{".env", "*.pem", "*.key", "id_rsa*"},
			ProtectedBranches: []string{"main", "master", "release/*"},
		},
		Agent: AgentConfig{
		MaxIterations:      50,
			MaxParallelTools:   5,
			PlanningMode:       "auto",
			Permission:         "exec",
			ReasoningLevel:     "medium",
			AutoConfirmLowRisk: true,
			ShowTokenUsage:     true,
			ShowCost:           true,
			Verbose:            false,
		},
		Context: ContextConfig{
			MaxTokens:          128000,
			DirectoryTreeDepth: 3,
			IgnorePatterns: []string{
				"node_modules", ".git", "__pycache__", "dist",
				"build", ".venv", "vendor",
			},
		},
		MCP: MCPConfig{
			Servers: make(map[string]MCPServerConfig),
		},
	}
}

// defaultShell returns the default shell based on OS
func defaultShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}

	switch runtime.GOOS {
	case "windows":
		if _, err := os.Stat("C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe"); err == nil {
			return "powershell"
		}
		return "cmd"
	default:
		return "/bin/sh"
	}
}

// GetConfigDir returns the MiMo config directory (~/.mimo)
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".mimo"), nil
}

// GetConfigFilePath returns the path to the config file
func GetConfigFilePath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}
