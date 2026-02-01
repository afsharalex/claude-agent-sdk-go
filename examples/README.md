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

| Example | Description | API | Difficulty |
|---------|-------------|-----|------------|
| [simple/](simple/) | One-shot query | `Query()` | Beginner |
| [streaming/](streaming/) | Interactive conversations | `Client` | Beginner |
| [mcp/](mcp/) | Custom tools with MCP | `Client` + MCP | Intermediate |
| [hooks/](hooks/) | Intercept tool calls | `Client` + Hooks | Advanced |

---

## simple/

**One-shot query to Claude Code**

The simplest way to use the SDK. Send a question, get an answer.

```bash
cd simple
go run main.go
```

**What you'll learn:**
- Using `claude.Query()` for one-shot queries
- Handling message and error channels with `select`
- Processing `AssistantMessage` and `ResultMessage` types
- Extracting text from `TextBlock` content

**Key concepts:**
```go
messages, errors := claude.Query(ctx, "What is the capital of France?")

for {
    select {
    case msg, ok := <-messages:
        // Handle messages
    case err := <-errors:
        // Handle errors
    }
}
```

---

## streaming/

**Interactive streaming client**

Build a chat interface with multi-turn conversations.

```bash
cd streaming
go run main.go
```

**What you'll learn:**
- Using `claude.NewClient()` for interactive sessions
- Connecting and disconnecting with `Connect()` / `Close()`
- Sending queries with `client.Query()`
- Processing messages in a separate goroutine
- Handling user input in a loop
- Using `/interrupt` to stop Claude

**Key concepts:**
```go
client := claude.NewClient()
client.Connect(ctx)
defer client.Close()

// Handle messages in background
go func() {
    for msg := range client.Messages() {
        handleMessage(msg)
    }
}()

// Send queries
client.Query(ctx, userInput)
```

**Commands:**
- Type any message to chat with Claude
- `/interrupt` - Stop Claude's current response
- `/quit` or `/exit` - Exit the program
- `Ctrl+C` - Force quit

---

## mcp/

**Custom tools with SDK MCP servers**

Create in-process tools that Claude can use.

```bash
cd mcp
go run main.go
```

**What you'll learn:**
- Defining tools with `claude.Tool()`
- Using `claude.SimpleInputSchema()` for input validation
- Creating MCP servers with `claude.CreateSDKMCPServer()`
- Returning results with `claude.TextResult()` and `claude.ErrorResult()`
- Configuring allowed tools with `WithAllowedTools()`

**Key concepts:**
```go
// Define a tool
addTool := claude.Tool("add", "Add two numbers",
    claude.SimpleInputSchema(map[string]string{"a": "number", "b": "number"}),
    func(ctx context.Context, args map[string]any) (claude.MCPToolResult, error) {
        a := args["a"].(float64)
        b := args["b"].(float64)
        return claude.TextResult(fmt.Sprintf("%g", a+b)), nil
    },
)

// Create server and client
server := claude.CreateSDKMCPServer("calculator", "1.0.0", []claude.MCPTool{addTool})
client := claude.NewClient(
    claude.WithMCPServers(map[string]claude.MCPServerConfig{"calc": server}),
    claude.WithAllowedTools([]string{"add"}),
)
```

**Tools in this example:**
- `add` - Add two numbers
- `multiply` - Multiply two numbers
- `sqrt` - Calculate square root

---

## hooks/

**Intercept and modify tool calls**

Use hooks for logging, security, and custom behavior.

```bash
cd hooks
go run main.go
```

**What you'll learn:**
- Defining hooks with `claude.HookMatcher`
- Implementing `claude.HookCallback` functions
- Using `PreToolUse` to allow/deny/modify tool calls
- Using `PostToolUse` to add context after execution
- Using `UserPromptSubmit` to intercept user input
- Setting hook timeouts

**Key concepts:**
```go
hooks := map[claude.HookEvent][]claude.HookMatcher{
    claude.HookEventPreToolUse: {{
        Matcher: "Bash",  // Match specific tool
        Hooks: []claude.HookCallback{
            func(ctx context.Context, input claude.HookInput, toolUseID string, hookCtx claude.HookContext) (claude.HookOutput, error) {
                pre := input.(claude.PreToolUseHookInput)

                // Allow, deny, or modify
                return claude.HookOutput{
                    HookSpecificOutput: claude.PreToolUseHookSpecificOutput{
                        HookEventName:      claude.HookEventPreToolUse,
                        PermissionDecision: claude.HookPermissionDecisionAllow,
                    },
                }, nil
            },
        },
        Timeout: 30,
    }},
}

client := claude.NewClient(claude.WithHooks(hooks))
```

**Hook events demonstrated:**
- `PreToolUse` - Before tool execution (can block)
- `PostToolUse` - After successful execution
- `UserPromptSubmit` - When user sends a prompt

---

## Running All Examples

Build and verify all examples compile:

```bash
# From repository root
go build ./examples/...
```

Run each example individually:

```bash
go run ./examples/simple
go run ./examples/streaming
go run ./examples/mcp
go run ./examples/hooks
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
