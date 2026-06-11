package llm

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mimo-cli/mimo-cli/internal/config"
)

func TestAnthropicBuildAPIRequestIncludesTextAttachmentContent(t *testing.T) {
	provider := NewAnthropicProvider(config.ModelConfig{Model: "claude-test", MaxTokens: 1024})

	req := provider.buildAPIRequest(ChatRequest{
		Messages: []Message{
			{
				Role:    RoleUser,
				Content: "read this",
				Attachments: []Attachment{
					{
						Name:    "notes.md",
						Type:    "text/markdown",
						DataURL: "data:text/markdown;base64,IyBUaXRsZQoKSGVsbG8gZnJvbSBmaWxlLg==",
					},
				},
			},
		},
	}, false)

	if len(req.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(req.Messages))
	}
	if len(req.Messages[0].Content) != 2 {
		t.Fatalf("expected text plus attachment content, got %#v", req.Messages[0].Content)
	}

	text := req.Messages[0].Content[1].Text
	if !strings.Contains(text, "notes.md") {
		t.Fatalf("expected filename in attachment text, got %q", text)
	}
	if !strings.Contains(text, "# Title\n\nHello from file.") {
		t.Fatalf("expected decoded file content, got %q", text)
	}
}

func TestAnthropicBuildAPIRequestKeepsImageAttachmentAsImageContent(t *testing.T) {
	provider := NewAnthropicProvider(config.ModelConfig{Model: "claude-test", MaxTokens: 1024})

	req := provider.buildAPIRequest(ChatRequest{
		Messages: []Message{
			{
				Role:    RoleUser,
				Content: "look",
				Attachments: []Attachment{
					{
						Name:    "image.png",
						Type:    "image/png",
						DataURL: "data:image/png;base64,iVBORw0KGgo=",
					},
				},
			},
		},
	}, false)

	if len(req.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(req.Messages))
	}
	if len(req.Messages[0].Content) != 2 {
		t.Fatalf("expected text plus image content, got %#v", req.Messages[0].Content)
	}

	raw, err := json.Marshal(req.Messages[0].Content[1])
	if err != nil {
		t.Fatalf("marshal image content: %v", err)
	}
	jsonText := string(raw)
	for _, want := range []string{
		`"type":"image"`,
		`"media_type":"image/png"`,
		`"data":"iVBORw0KGgo="`,
	} {
		if !strings.Contains(jsonText, want) {
			t.Fatalf("expected image content JSON to contain %s, got %s", want, jsonText)
		}
	}
}
