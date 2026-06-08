package mcp

import (
	"io"
	"os"
)

// Transport defines the interface for MCP communication
type Transport interface {
	// Start starts the transport
	Start() error

	// Send sends a message (request or notification)
	Send(msg interface{}) error

	// Receive reads the next message from the server
	Receive() (*Response, error)

	// Close closes the transport
	Close() error
}

// StdioTransport implements Transport using stdin/stdout of a subprocess
type StdioTransport struct {
	cmd    string
	args   []string
	env    []string
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	proc   *os.Process
	enc    *jsonEncoder
	dec    *jsonDecoder
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport(command string, args []string, env []string) *StdioTransport {
	return &StdioTransport{
		cmd:  command,
		args: args,
		env:  env,
	}
}
