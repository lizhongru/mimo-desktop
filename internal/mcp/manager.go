package mcp

import (
	"fmt"
	"os"
	"strings"
	"path/filepath"

	"github.com/mimo-cli/mimo-cli/internal/config"
	"github.com/mimo-cli/mimo-cli/internal/tools"
)

// Manager manages multiple MCP server connections
type Manager struct {
	clients map[string]*Client
	tools   []*MCPToolAdapter
}

// NewManager creates a new MCP manager
func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*Client),
	}
}

// ConnectAll connects to all enabled MCP servers from config
func (m *Manager) ConnectAll(cfg config.MCPConfig) []error {
	var errs []error

	for name, serverCfg := range cfg.Servers {
		if !serverCfg.Enabled {
			continue
		}

		if err := m.ConnectServer(name, serverCfg); err != nil {
			errs = append(errs, fmt.Errorf("MCP server %q: %w", name, err))
		}
	}

	return errs
}

// ConnectServer connects to a single MCP server
func (m *Manager) ConnectServer(name string, cfg config.MCPServerConfig) error {
	var transport Transport

	if cfg.URL != "" {
		// SSE transport for remote servers
		transport = NewSSETransport(cfg.URL)
	} else if cfg.Command != "" {
		// Stdio transport for local subprocess
		env := make([]string, 0, len(cfg.Env))
		for k, v := range cfg.Env {
			env = append(env, k+"="+v)
		}
		// 自动设置 npm 缓存目录（避免权限问题）
		if cfg.Command == "npx" || cfg.Command == "npm" {
			cwd, _ := os.Getwd()
			npmCache := filepath.Join(cwd, ".npm-cache")
			env = append(env, "npm_config_cache="+npmCache)
		}
		// filesystem 服务器：如果 args 里没有目录参数，自动加上当前工作目录
		args := cfg.Args
		if len(args) == 1 && strings.Contains(args[0], "server-filesystem") {
			cwd, _ := os.Getwd()
			args = append(args, cwd)
		}
		transport = NewStdioTransport(cfg.Command, args, env)
	} else {
		return fmt.Errorf("server %q: must specify either 'command' or 'url'", name)
	}

	client := NewClient(transport)
	client.SetServerName(name)

	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}

	m.clients[name] = client

	// Discover tools
	tools, err := client.ListTools()
	if err != nil {
		// Still connected, just no tools
		fmt.Fprintf(os.Stderr, "Warning: MCP server %q: failed to list tools: %v\n", name, err)
		return nil
	}

	// Create tool adapters
	for _, tool := range tools {
		adapter := NewMCPToolAdapter(client, tool, name)
		m.tools = append(m.tools, adapter)
	}

	fmt.Fprintf(os.Stderr, "MCP: connected to %q (%d tools)\n", name, len(tools))
	return nil
}

// RegisterTools registers all discovered MCP tools into a registry
func (m *Manager) RegisterTools(registry *tools.Registry) {
	for _, adapter := range m.tools {
		registry.Register(adapter)
	}
}

// GetTools returns all discovered MCP tool adapters
func (m *Manager) GetTools() []*MCPToolAdapter {
	return m.tools
}

// GetClient returns a client by server name
func (m *Manager) GetClient(name string) (*Client, bool) {
	c, ok := m.clients[name]
	return c, ok
}

// CloseAll closes all MCP server connections
func (m *Manager) CloseAll() {
	var errs []string
	for name, client := range m.clients {
		if err := client.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", name, err))
		}
	}
	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: errors closing MCP servers: %s\n", strings.Join(errs, "; "))
	}
}

// ServerNames returns the names of all connected servers
func (m *Manager) ServerNames() []string {
	names := make([]string, 0, len(m.clients))
	for name := range m.clients {
		names = append(names, name)
	}
	return names
}
