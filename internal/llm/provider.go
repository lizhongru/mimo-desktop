package llm

import (
	"context"
)

// Provider defines the interface for LLM providers
type Provider interface {
	// Name returns the provider name
	Name() string

	// Chat sends a non-streaming chat request
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	// ChatStream sends a streaming chat request, returns a channel of chunks
	ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)

	// IsAvailable checks if the provider is properly configured
	IsAvailable() error

	// ListModels fetches available models from the provider API
	ListModels(ctx context.Context) ([]ModelInfo, error)
}
