package llm

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildOpenAIContentIncludesTextAttachmentContent(t *testing.T) {
	raw := buildOpenAIContent("read this", []Attachment{
		{
			Name:    "notes.md",
			Type:    "text/markdown",
			DataURL: "data:text/markdown;base64,IyBUaXRsZQoKSGVsbG8gZnJvbSBmaWxlLg==",
		},
	})

	var parts []map[string]interface{}
	if err := json.Unmarshal(raw, &parts); err != nil {
		t.Fatalf("unmarshal content parts: %v", err)
	}
	if len(parts) != 2 {
		t.Fatalf("expected 2 content parts, got %d: %#v", len(parts), parts)
	}

	text, ok := parts[1]["text"].(string)
	if !ok {
		t.Fatalf("expected second part text, got %#v", parts[1])
	}
	if !strings.Contains(text, "notes.md") {
		t.Fatalf("expected filename in attachment text, got %q", text)
	}
	if !strings.Contains(text, "# Title\n\nHello from file.") {
		t.Fatalf("expected decoded file content, got %q", text)
	}
}

func TestBuildOpenAIContentKeepsImageAttachmentAsImageURL(t *testing.T) {
	const imageDataURL = "data:image/png;base64,iVBORw0KGgo="
	raw := buildOpenAIContent("look", []Attachment{
		{
			Name:    "image.png",
			Type:    "image/png",
			DataURL: imageDataURL,
		},
	})

	var parts []map[string]interface{}
	if err := json.Unmarshal(raw, &parts); err != nil {
		t.Fatalf("unmarshal content parts: %v", err)
	}
	if len(parts) != 2 {
		t.Fatalf("expected 2 content parts, got %d: %#v", len(parts), parts)
	}

	imageURL, ok := parts[1]["image_url"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected image_url part, got %#v", parts[1])
	}
	if got := imageURL["url"]; got != imageDataURL {
		t.Fatalf("expected image data URL %q, got %#v", imageDataURL, got)
	}
}
