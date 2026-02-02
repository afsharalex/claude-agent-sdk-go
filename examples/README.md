# Examples

This directory contains working examples demonstrating the Claude Agent SDK for Go.

## Prerequisites

Before running any example:

1. **Install Go 1.25+**
   ```bash
   go version
   ```

2. **Install Claude Code CLI**
   ```bash
   curl -fsSL https://claude.ai/install.sh | bash
   claude --version
   ```

3. **Authenticate with Claude Code**
   ```bash
   claude auth login
   ```

## Examples Overview

| Example | Description | Key APIs | Difficulty |
|---------|-------------|----------|------------|
| [01_simple](01_simple/) | One-shot query | `Query()` | Beginner |
| [02_streaming](02_streaming/) | Interactive conversations | `Client`, `Messages()` | Beginner |
| [03_mcp](03_mcp/) | Custom tools with MCP | `MCPServer`, `Tool()` | Intermediate |
| [04_hooks](04_hooks/) | Intercept tool calls | `WithHooks()`, `HookCallback` | Intermediate |
| [05_context_manager](05_context_manager/) | Automatic resource cleanup | `WithClient()` | Beginner |
| [06_multi_turn](06_multi_turn/) | Multi-turn conversations | `Client.Query()` | Beginner |
| [07_session_management](07_session_management/) | Resume and fork sessions | `WithResume()`, `WithForkSession()` | Intermediate |
| [08_structured_output](08_structured_output/) | JSON Schema output | `WithJSONSchema()` | Intermediate |
| [09_file_checkpointing](09_file_checkpointing/) | Track file changes | `WithFileCheckpointing()` | Intermediate |
| [10_error_handling](10_error_handling/) | Typed error handling | `Is*Error()`, `As*Error()` | Intermediate |
| [11_partial_streaming](11_partial_streaming/) | Real-time updates | `WithPartialStreaming()` | Intermediate |
| [12_sandbox](12_sandbox/) | Bash command isolation | `WithSandbox()` | Advanced |
| [13_debugging](13_debugging/) | Debug output | `WithDebugStderr()`, `WithStderr()` | Beginner |
| [14_permission_callback](14_permission_callback/) | Tool permission control | `WithCanUseTool()` | Advanced |
| [15_subagents](15_subagents/) | Custom agent definitions | `WithAgent()`, `WithAgents()` | Advanced |
| [16_plugins](16_plugins/) | Plugin loading | `WithLocalPlugin()`, `WithPlugin()` | Advanced |

## Difficulty Progression

**Beginner** - Start here to learn the basics:
1. `01_simple` - Understand basic query/response flow
2. `02_streaming` - Learn interactive client usage
3. `05_context_manager` - Automatic resource cleanup
4. `06_multi_turn` - Multi-turn conversations
5. `13_debugging` - Debug output for troubleshooting

**Intermediate** - Build on fundamentals:
1. `03_mcp` - Create custom tools
2. `04_hooks` - Intercept and modify tool calls
3. `07_session_management` - Session resume and forking
4. `08_structured_output` - JSON Schema constrained output
5. `09_file_checkpointing` - Track file changes
6. `10_error_handling` - Proper error type handling
7. `11_partial_streaming` - Real-time streaming updates

**Advanced** - Full control and customization:
1. `12_sandbox` - Secure bash execution
2. `14_permission_callback` - Fine-grained permission control
3. `15_subagents` - Programmatic agent definitions
4. `16_plugins` - Load and use plugins

---

## 01_simple/

**One-shot query to Claude Code**

The simplest way to use the SDK. Send a question, get an answer.

```bash
cd 01_simple
go run main.go
```

**What you'll learn:**
- Using `claude.Query()` for one-shot queries
- Handling message and error channels with `select`
- Processing `AssistantMessage` and `ResultMessage` types
- Extracting text from `TextBlock` content

---

## 02_streaming/

**Interactive streaming client**

Build a chat interface with multi-turn conversations.

```bash
cd 02_streaming
go run main.go
```

**What you'll learn:**
- Using `claude.NewClient()` for interactive sessions
- Connecting and disconnecting with `Connect()` / `Close()`
- Sending queries with `client.Query()`
- Processing messages in a separate goroutine

---

## 03_mcp/

**Custom tools with SDK MCP servers**

Create in-process tools that Claude can use.

```bash
cd 03_mcp
go run main.go
```

**What you'll learn:**
- Defining tools with `claude.Tool()`
- Using `claude.SimpleInputSchema()` for input validation
- Creating MCP servers with `claude.CreateSDKMCPServer()`
- Returning results with `claude.TextResult()` and `claude.ErrorResult()`

---

## 04_hooks/

**Intercept and modify tool calls**

Use hooks for logging, security, and custom behavior.

```bash
cd 04_hooks
go run main.go
```

**What you'll learn:**
- Defining hooks with `claude.HookMatcher`
- Using `PreToolUse` to allow/deny/modify tool calls
- Using `PostToolUse` to add context after execution

---

## 05_context_manager/

**Automatic resource cleanup with WithClient**

Use the `WithClient` pattern to ensure resources are cleaned up automatically.

```bash
cd 05_context_manager
go run main.go
```

**What you'll learn:**
- Using `claude.WithClient()` for automatic cleanup
- Simplifying error handling with callbacks

---

## 06_multi_turn/

**Structured multi-turn conversations**

Build conversations that maintain context across multiple exchanges.

```bash
cd 06_multi_turn
go run main.go
```

**What you'll learn:**
- Maintaining conversation context
- Sequential query patterns

---

## 07_session_management/

**Session resume and fork**

Continue previous sessions or create branches.

```bash
cd 07_session_management
go run main.go
```

**What you'll learn:**
- Using `WithResume()` to continue sessions
- Using `WithForkSession()` to branch conversations
- Managing session IDs

---

## 08_structured_output/

**JSON Schema constrained output**

Get structured JSON output that conforms to your schema.

```bash
cd 08_structured_output
go run main.go
```

**What you'll learn:**
- Using `WithJSONSchema()` for structured output
- Parsing structured responses into Go types

---

## 09_file_checkpointing/

**Track and manage file changes**

Enable checkpointing to track what files Claude modifies.

```bash
cd 09_file_checkpointing
go run main.go
```

**What you'll learn:**
- Using `WithFileCheckpointing()` to track changes
- Managing file modifications

---

## 10_error_handling/

**Comprehensive error type handling**

Handle different error types appropriately.

```bash
cd 10_error_handling
go run main.go
```

**What you'll learn:**
- Using `IsJSONDecodeError()`, `IsMessageParseError()`, etc.
- Using `As*Error()` to extract error details
- Building robust error handling

---

## 11_partial_streaming/

**Real-time progressive updates**

Get updates as Claude generates its response.

```bash
cd 11_partial_streaming
go run main.go
```

**What you'll learn:**
- Using `WithPartialStreaming()` for real-time updates
- Processing `StreamEvent` messages

---

## 12_sandbox/

**Secure bash command execution**

Isolate bash commands for improved security.

```bash
cd 12_sandbox
go run main.go
```

**What you'll learn:**
- Configuring `SandboxSettings`
- Using `WithSandbox()` for isolation
- Excluding specific commands from sandbox

---

## 13_debugging/

**Debug output and diagnostics**

Capture and analyze CLI output for troubleshooting.

```bash
cd 13_debugging
go run main.go
```

**What you'll learn:**
- Using `WithDebugStderr()` for quick debugging
- Using `WithStderr()` for custom handling

---

## 14_permission_callback/

**Fine-grained tool permission control**

Programmatically control which tools Claude can use.

```bash
cd 14_permission_callback
go run main.go
```

**What you'll learn:**
- Using `WithCanUseTool()` for permission callbacks
- Implementing allow/deny logic per tool
- Modifying tool inputs

---

## 15_subagents/

**Programmatic agent definitions**

Define custom specialized agents.

```bash
cd 15_subagents
go run main.go
```

**What you'll learn:**
- Using `WithAgent()` to define individual agents
- Using `WithAgents()` for bulk definitions
- Configuring agent prompts, tools, and models

---

## 16_plugins/

**Load and use plugins**

Load plugins to extend Claude's capabilities with custom commands.

```bash
cd 16_plugins
go run main.go
```

**What you'll learn:**
- Using `WithLocalPlugin()` to load plugins from a path
- Using `WithPlugin()` with `SdkPluginConfig` for more control
- Plugin directory structure and configuration
- Inspecting `SystemMessage.Data` for plugin information

---

## Running All Examples

Build and verify all examples compile:

```bash
# From repository root
go build ./examples/...
```

Run each example individually:

```bash
go run ./examples/01_simple
go run ./examples/02_streaming
go run ./examples/03_mcp
# ... etc
```

---

## Common Patterns

### Signal Handling

All examples include proper signal handling:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
go func() {
    <-sigCh
    cancel()
}()
```

### Error Handling

Check error channels alongside message channels:

```go
select {
case msg, ok := <-messages:
    if !ok {
        return // Channel closed
    }
    // Handle message
case err := <-errors:
    if err != nil {
        log.Fatal(err)
    }
}
```

### Resource Cleanup

Always close clients:

```go
client := claude.NewClient()
if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Close()
```

Or use `WithClient` for automatic cleanup:

```go
err := claude.WithClient(ctx, func(client *claude.Client) error {
    // Use client - automatically closed when function returns
    return nil
})
```

---

## Next Steps

After exploring these examples:

1. Read the [Getting Started Guide](../docs/getting-started.md)
2. Explore the [How-to Guides](../docs/guides/)
3. Check the [API Reference](../docs/reference.md)
4. Review the [Architecture](../ARCHITECTURE.md) for deeper understanding

---

## Troubleshooting

### "Claude Code not found"

Install Claude Code:
```bash
curl -fsSL https://claude.ai/install.sh | bash
```

### "Not authenticated"

Log in to Claude Code:
```bash
claude auth login
```

### "Permission denied"

Some examples may need permission mode:
```go
claude.WithPermissionMode(claude.PermissionModeAcceptEdits)
```

### Examples won't compile

Ensure dependencies are up to date:
```bash
go mod tidy
```
