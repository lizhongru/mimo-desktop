package mcp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// SSETransport implements Transport using HTTP with Server-Sent Events
type SSETransport struct {
	url        string
	httpClient *http.Client
	sessionID  string
	msgURL     string // URL to send messages to (received via SSE endpoint event)
	mu         sync.Mutex
	respCh     chan *Response
	closed     bool
}

// NewSSETransport creates a new SSE transport
func NewSSETransport(url string) *SSETransport {
	return &SSETransport{
		url: url,
		httpClient: &http.Client{
			Timeout: 0, // No timeout for long-lived SSE connections
		},
		respCh: make(chan *Response, 10),
	}
}

// Start connects to the SSE endpoint and starts listening for events
func (t *SSETransport) Start() error {
	req, err := http.NewRequest("GET", t.url, nil)
	if err != nil {
		return fmt.Errorf("failed to create SSE request: %w", err)
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to SSE endpoint: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("SSE endpoint returned status %d", resp.StatusCode)
	}

	// Start reading SSE events in background
	go t.readSSEEvents(resp.Body)

	return nil
}

// readSSEEvents reads Server-Sent Events from the response body
func (t *SSETransport) readSSEEvents(body io.ReadCloser) {
	defer body.Close()

	scanner := bufio.NewScanner(body)
	var eventType, data string

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line = end of event, process it
			if data != "" {
				t.handleEvent(eventType, data)
			}
			eventType = ""
			data = ""
			continue
		}

		if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			data = strings.TrimPrefix(line, "data: ")
		}
	}

	// Connection closed
	t.mu.Lock()
	if !t.closed {
		close(t.respCh)
	}
	t.mu.Unlock()
}

// handleEvent processes a single SSE event
func (t *SSETransport) handleEvent(eventType, data string) {
	switch eventType {
	case "endpoint":
		// Server tells us where to send messages
		t.mu.Lock()
		t.msgURL = data
		t.mu.Unlock()

	case "message":
		// JSON-RPC response from server
		var resp Response
		if err := json.Unmarshal([]byte(data), &resp); err == nil {
			t.mu.Lock()
			if !t.closed {
				select {
				case t.respCh <- &resp:
				default:
					// Channel full, drop message
				}
			}
			t.mu.Unlock()
		}
	}
}

// Send sends a JSON-RPC request to the server
func (t *SSETransport) Send(msg interface{}) error {
	t.mu.Lock()
	msgURL := t.msgURL
	t.mu.Unlock()

	if msgURL == "" {
		return fmt.Errorf("message URL not yet received from server")
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", msgURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if t.sessionID != "" {
		req.Header.Set("Mcp-Session-Id", t.sessionID)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	// Capture session ID if provided
	if sid := resp.Header.Get("Mcp-Session-Id"); sid != "" {
		t.mu.Lock()
		t.sessionID = sid
		t.mu.Unlock()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// Receive reads the next JSON-RPC response
func (t *SSETransport) Receive() (*Response, error) {
	select {
	case resp, ok := <-t.respCh:
		if !ok {
			return nil, fmt.Errorf("SSE connection closed")
		}
		return resp, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response")
	}
}

// Close closes the SSE transport
func (t *SSETransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	return nil
}
