package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// jsonEncoder wraps a writer with JSON encoding
type jsonEncoder struct {
	w *bufio.Writer
}

func (e *jsonEncoder) Encode(v interface{}) error {
	if err := json.NewEncoder(e.w).Encode(v); err != nil {
		return err
	}
	return e.w.Flush()
}

// jsonDecoder wraps a reader with JSON decoding
type jsonDecoder struct {
	r *bufio.Reader
}

func (d *jsonDecoder) Decode(v interface{}) error {
	return json.NewDecoder(d.r).Decode(v)
}

// Start starts the subprocess and sets up stdin/stdout pipes
func (t *StdioTransport) Start() error {
	cmd := exec.Command(t.cmd, t.args...)

	// Set environment variables
	cmd.Env = os.Environ()
	for _, e := range t.env {
		cmd.Env = append(cmd.Env, e)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command %q: %w", t.cmd, err)
	}

	t.stdin = stdin
	t.stdout = stdout
	t.stderr = stderr
	t.proc = cmd.Process
	t.enc = &jsonEncoder{w: bufio.NewWriter(stdin)}
	t.dec = &jsonDecoder{r: bufio.NewReader(stdout)}

	// Drain stderr in background to avoid blocking
	go func() {
		buf := make([]byte, 4096)
		for {
			_, err := stderr.Read(buf)
			if err != nil {
				return
			}
		}
	}()

	return nil
}

// Send sends a JSON-RPC message to the server
func (t *StdioTransport) Send(msg interface{}) error {
	return t.enc.Encode(msg)
}

// Receive reads a JSON-RPC response from the server
func (t *StdioTransport) Receive() (*Response, error) {
	var resp Response
	if err := t.dec.Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &resp, nil
}

// Close closes the subprocess
func (t *StdioTransport) Close() error {
	var errs []string

	if t.stdin != nil {
		if err := t.stdin.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("stdin close: %v", err))
		}
	}

	if t.proc != nil {
		if err := t.proc.Kill(); err != nil {
			errs = append(errs, fmt.Sprintf("process kill: %v", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("transport close errors: %s", strings.Join(errs, "; "))
	}
	return nil
}
