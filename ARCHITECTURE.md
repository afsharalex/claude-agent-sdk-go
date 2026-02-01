# Architecture

This document explains the design decisions and internal architecture of the Claude Agent SDK for Go.

## Design Philosophy

The SDK follows Go idioms while maintaining conceptual alignment with the Python SDK:

1. **Channel-based concurrency** - Messages flow through channels rather than async iterators, enabling native Go patterns like `select` statements and goroutine composition.

2. **Functional options** - Configuration uses the functional options pattern (`WithModel()`, `WithCwd()`) for type-safe, extensible, and self-documenting APIs.

3. **Interface-based polymorphism** - Message and content types implement interfaces (`Message`, `ContentBlock`) enabling type switches for handling different message types.

4. **Internal packages** - Implementation details live in `internal/` to keep the public API surface clean while allowing internal refactoring.

## Three-Layer Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Public API Layer                         │
│  claude.go  │  client.go  │  options.go  │  types.go  │  mcp.go │
├─────────────────────────────────────────────────────────────────┤
│                       Protocol Layer                            │
│                    internal/protocol/                           │
│         Query handler, hooks, MCP routing, control protocol     │
├─────────────────────────────────────────────────────────────────┤
│                       Transport Layer                           │
│                    internal/transport/                          │
│              Subprocess management, I/O handling                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌───────────────────┐
                    │   Claude Code CLI  │
                    └───────────────────┘
```

### Public API Layer

The public API provides two primary interfaces:

- **`Query()`** - One-shot function for simple queries. Returns message and error channels, spawns a subprocess, and cleans up when complete.

- **`Client`** - Stateful struct for interactive sessions. Maintains connection state, supports bidirectional messaging, custom tools, and hooks.

Supporting components:
- **`options.go`** - Functional options and the `Options` struct
- **`types.go`** - All public types (messages, content blocks, hooks, permissions, MCP)
- **`mcp.go`** - MCP helper functions (`Tool()`, `TextResult()`, etc.)
- **`errors.go`** - Error type hierarchy
- **`messages.go`** - Message parsing from JSON

### Protocol Layer

Located in `internal/protocol/`, this layer handles the control protocol between the SDK and CLI:

- **`query.go`** - Core `Query` struct that:
  - Manages bidirectional control protocol messages
  - Handles hook callbacks (PreToolUse, PostToolUse, etc.)
  - Routes MCP tool calls to in-process SDK servers
  - Tracks pending request/response pairs
  - Provides interrupt and control methods

- **`control.go`** - Control protocol types and constants

The protocol layer abstracts the streaming JSON protocol, handling initialization, request/response correlation, and message routing.

### Transport Layer

Located in `internal/transport/`, this layer handles subprocess lifecycle and I/O:

- **`transport.go`** - `Transport` interface definition
- **`subprocess.go`** - `SubprocessTransport` implementation that:
  - Spawns Claude Code CLI as a subprocess
  - Manages stdin/stdout/stderr pipes
  - Handles graceful shutdown and cleanup
  - Buffers and parses JSON lines

- **`mock_transport.go`** - Mock implementation for testing

## Message Flow

### Query Flow

```
Query("prompt")
      │
      ▼
┌─────────────┐    spawn     ┌──────────────┐
│  Options    │─────────────▶│  Subprocess  │
└─────────────┘              │   Transport  │
      │                      └──────────────┘
      │                            │
      ▼                            │ stdin
┌─────────────┐              ┌─────▼────────┐
│   Query     │◀────JSON─────│ Claude Code  │
│  Handler    │              │     CLI      │
└─────────────┘              └──────────────┘
      │
      │ channels
      ▼
┌─────────────┐
│  Caller's   │
│  for loop   │
└─────────────┘
```

### Client Interactive Flow

```
Client.Connect()
      │
      ▼
┌─────────────┐    spawn     ┌──────────────┐
│  Transport  │─────────────▶│ Claude Code  │
└─────────────┘              │  (streaming) │
      │                      └──────────────┘
      ▼                            │
┌─────────────┐                    │
│   Query     │◀───────────────────┘
│  Handler    │
└─────────────┘
      │
      ├─────────────────┬─────────────────┐
      ▼                 ▼                 ▼
┌───────────┐    ┌───────────┐    ┌───────────┐
│  Messages │    │   Hooks   │    │    MCP    │
│  Channel  │    │ Callbacks │    │  Servers  │
└───────────┘    └───────────┘    └───────────┘
```

## Key Design Decisions

### Why Channels Instead of Iterators?

Go lacks native async iterators. Channels provide:
- Natural integration with `select` for timeouts and cancellation
- Easy composition with other goroutines
- Standard Go concurrency patterns

The tradeoff is that channels require explicit closing and careful goroutine management, which the SDK handles internally.

### Why Functional Options?

The functional options pattern offers:
- Type safety at compile time
- Optional parameters with sensible defaults
- Backward-compatible API evolution
- Self-documenting code (`WithModel("claude-sonnet-4-5")`)

Alternative approaches (config structs, builder pattern) were considered but functional options align best with Go conventions for optional configuration.

### Why Internal Packages?

Placing `protocol/` and `transport/` under `internal/`:
- Keeps the public API surface minimal
- Allows internal refactoring without breaking changes
- Clearly separates what users should depend on
- Enables testing with mock implementations

### Why In-Process MCP Servers?

SDK MCP servers run tools as Go functions within the same process:
- **Zero IPC overhead** - Direct function calls vs subprocess spawning
- **Single binary deployment** - No external server processes to manage
- **Unified debugging** - Breakpoints work across tool calls
- **Type safety** - Go compiler validates tool implementations

External MCP servers are still supported for interoperability.

## Type Conversion

The SDK maintains parallel type hierarchies:
- Public types in the `claude` package (e.g., `claude.HookInput`)
- Internal types in `internal/types/` (e.g., `types.HookInput`)

Conversion functions in `claude.go` bridge these:
- `toInternalHooks()` - Public hooks → internal hooks
- `toPublicHookInput()` - Internal input → public input
- `toInternalMCPServers()` - Public MCP config → internal servers

This separation allows the public API to remain stable while internal representations can evolve.

## Concurrency Model

The SDK uses several goroutines:

1. **Message reader** - Started by `Query.Start()`, reads from transport and routes messages
2. **Input streamer** - For `QueryStreaming()`, forwards input channel to transport
3. **Hook handlers** - Control requests spawn goroutines for callback execution
4. **User code** - The caller's goroutine(s) consuming message channels

All goroutines respect context cancellation and coordinate shutdown via `sync.WaitGroup`.

## Testing Strategy

- **Unit tests** - Co-located `*_test.go` files test individual components
- **Mock transport** - `internal/transport/mock_transport.go` enables testing without CLI
- **Integration tests** - Examples in `examples/` serve as integration tests

The mock transport simulates CLI responses, allowing protocol and handler testing without subprocess overhead.
