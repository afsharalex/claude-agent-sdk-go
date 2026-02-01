// Package transport provides transport implementations for Claude SDK.
package transport

import "context"

// ReadResult represents a message read from the transport.
type ReadResult struct {
	Data  map[string]any
	Error error
}

// Transport is the interface for Claude communication.
//
// WARNING: This internal API is exposed for custom transport implementations
// (e.g., remote Claude Code connections). The Claude Code team may change or
// remove this interface in any future release. Custom implementations must be
// updated to match interface changes.
//
// This is a low-level transport interface that handles raw I/O with the Claude
// process or service. The Query handler builds on top of this to implement the
// control protocol and message routing.
type Transport interface {
	// Connect establishes the transport connection.
	// For subprocess transports, this starts the process.
	// For network transports, this establishes the connection.
	Connect(ctx context.Context) error

	// Write sends raw data to the transport.
	// data is typically JSON + newline.
	Write(ctx context.Context, data string) error

	// ReadMessages returns a channel that receives parsed JSON messages.
	// The channel is closed when the transport is closed or an error occurs.
	ReadMessages(ctx context.Context) <-chan ReadResult

	// Close closes the transport connection and cleans up resources.
	Close() error

	// IsReady returns true if the transport is ready for communication.
	IsReady() bool

	// EndInput ends the input stream (closes stdin for process transports).
	EndInput() error
}
