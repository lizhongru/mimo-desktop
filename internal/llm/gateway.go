package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/mimo-cli/mimo-cli/internal/config"
)

// Gateway manages LLM providers and routes requests
type Gateway struct {
	providers map[string]Provider
	current   string
}

// NewGateway creates a new LLM Gateway from configuration
func NewGateway(cfg *config.Config) *Gateway {
	gw := &Gateway{
		providers: make(map[string]Provider),
		current:   cfg.DefaultModel,
	}

	// Register providers based on config
	for name, modelCfg := range cfg.Models {
		var provider Provider
		// Auto-detect: if API base contains "anthropic", use Anthropic provider
		if strings.Contains(strings.ToLower(modelCfg.APIBase), "anthropic") {
			provider = NewAnthropicProvider(modelCfg)
		} else {
			provider = NewOpenAIProvider(modelCfg)
		}
		gw.providers[name] = provider
	}

	return gw
}

// SetCurrentModel switches the active model
func (g *Gateway) SetCurrentModel(name string) error {
	if _, ok := g.providers[name]; !ok {
		return fmt.Errorf("model %q not available", name)
	}
	g.current = name
	return nil
}

// GetCurrentModel returns the name of the current model
func (g *Gateway) GetCurrentModel() string {
	return g.current
}

// GetProvider returns the provider for the given model (or current model)
func (g *Gateway) GetProvider(modelName string) (Provider, error) {
	if modelName == "" {
		modelName = g.current
	}
	provider, ok := g.providers[modelName]
	if !ok {
		return nil, fmt.Errorf("provider for model %q not found", modelName)
	}
	return provider, nil
}

// Chat sends a non-streaming request to the current or specified model
func (g *Gateway) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	provider, err := g.GetProvider(req.Model)
	if err != nil {
		return nil, err
	}
	return provider.Chat(ctx, req)
}

// ChatStream sends a streaming request to the current or specified model
func (g *Gateway) ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	provider, err := g.GetProvider(req.Model)
	if err != nil {
		return nil, err
	}
	return provider.ChatStream(ctx, req)
}

// ListModels returns the names of all available models
func (g *Gateway) ListModels() []string {
	models := make([]string, 0, len(g.providers))
	for name := range g.providers {
		models = append(models, name)
	}
	return models
}

// ListRemoteModels fetches available models from the specified provider API
func (g *Gateway) ListRemoteModels(ctx context.Context, modelName string) ([]ModelInfo, error) {
	provider, err := g.GetProvider(modelName)
	if err != nil {
		return nil, err
	}
	return provider.ListModels(ctx)
}
