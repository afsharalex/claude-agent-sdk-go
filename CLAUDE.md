# Workflow

```bash
# Format code
go fmt ./...

# Lint (if golangci-lint is installed)
golangci-lint run

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test file
go test -v ./internal/protocol/

# Run tests with coverage
go test -cover ./...

# Build all packages
go build ./...

# Tidy dependencies
go mod tidy
```

# Codebase Structure

- `claude.go` - Main SDK entry point with `Query()` and `QueryStreaming()` functions
- `client.go` - `Client` for interactive sessions
- `options.go` - Configuration options and `With*` functional option functions
- `types.go` - All public type definitions (messages, content blocks, hooks, permissions, MCP configs)
- `mcp.go` - MCP helper functions (`Tool()`, `TextResult()`, `ErrorResult()`, etc.)
- `errors.go` - Error types
- `messages.go` - Message parsing logic
- `internal/` - Internal implementation details
  - `protocol/` - Control protocol handling
    - `query.go` - Query handler with hooks and MCP support
    - `types.go` - Protocol message types
  - `transport/` - CLI subprocess management
    - `subprocess.go` - Subprocess transport implementation
    - `mock.go` - Mock transport for testing
  - `types/` - Internal type definitions
    - `hooks.go` - Hook types
    - `mcp.go` - MCP types
    - `permissions.go` - Permission types
- `examples/` - Example applications
  - `simple/` - Basic one-shot query
  - `streaming/` - Interactive streaming client
  - `mcp/` - Custom MCP tools example
  - `hooks/` - Hooks example

# Key Design Patterns

## Functional Options

Configuration uses the functional options pattern:

```go
client := claude.NewClient(
    claude.WithCwd("/path"),
    claude.WithModel("claude-sonnet-4-5"),
)
```

## Channel-Based Communication

Messages are delivered via channels for concurrent processing:

```go
messages, errors := claude.Query(ctx, "prompt")
for msg := range messages {
    // Process messages
}
```

## Interface-Based Message Types

Messages implement the `Message` interface, with concrete types:
- `*AssistantMessage`
- `*UserMessage`
- `*SystemMessage`
- `*ResultMessage`

Content blocks implement the `ContentBlock` interface.

# Testing

Tests are co-located with source files using the `_test.go` suffix. The `internal/transport` package provides a `MockTransport` for testing without spawning actual CLI processes.

```go
mock := transport.NewMockTransport()
q := protocol.NewQuery(protocol.QueryConfig{
    Transport:       mock,
    IsStreamingMode: true,
})
```
