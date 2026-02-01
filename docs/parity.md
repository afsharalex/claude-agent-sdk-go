# Python SDK Parity

This document compares the Go SDK with the [Python Claude Agent SDK](https://github.com/anthropics/claude-agent-sdk-python), showing feature mappings and behavioral differences.

## Feature Comparison

| Feature | Python SDK | Go SDK | Notes |
|---------|------------|--------|-------|
| One-shot queries | `query()` | `Query()` | Both return iterators/channels |
| Interactive client | `ClaudeSDKClient` | `Client` | Context manager vs explicit Connect/Close |
| Custom tools | `@tool` decorator | `Tool()` function | Go uses function, not decorator |
| SDK MCP servers | `create_sdk_mcp_server()` | `CreateSDKMCPServer()` | Identical functionality |
| Hooks | `HookMatcher` | `HookMatcher` | Same structure |
| Permission callbacks | `can_use_tool` | `WithCanUseTool()` | Callback-based |
| Streaming | `async for` | `for range` | Channel-based in Go |
| Partial messages | Supported | `WithIncludePartialMessages()` | Opt-in |
| Session resume | `resume` | `WithResume()` | Same behavior |
| Subagents | `agents` | `WithAgents()` | Same configuration |
| Sandbox | Supported | `WithSandbox()` | Same settings |
| File checkpointing | Supported | `WithEnableFileCheckpointing()` | Same behavior |

## API Name Mappings

### Naming Conventions

| Python | Go |
|--------|-----|
| `snake_case` functions | `PascalCase` functions |
| `snake_case` parameters | `PascalCase` struct fields |
| `snake_case` options | `WithPascalCase()` functions |

### Core Functions

| Python | Go |
|--------|-----|
| `query(prompt, options)` | `Query(ctx, prompt, opts...)` |
| `ClaudeSDKClient(options)` | `NewClient(opts...)` |
| `client.query(prompt)` | `client.Query(ctx, prompt)` |
| `client.receive_response()` | `client.ReceiveResponse(ctx)` |

### Options

| Python | Go |
|--------|-----|
| `ClaudeAgentOptions(...)` | `NewOptions(opts...)` (internal) |
| `system_prompt=` | `WithSystemPrompt()` |
| `max_turns=` | `WithMaxTurns()` |
| `max_budget_usd=` | `WithMaxBudgetUSD()` |
| `model=` | `WithModel()` |
| `cwd=` | `WithCwd()` |
| `allowed_tools=` | `WithAllowedTools()` |
| `disallowed_tools=` | `WithDisallowedTools()` |
| `permission_mode=` | `WithPermissionMode()` |
| `mcp_servers=` | `WithMCPServers()` |
| `hooks=` | `WithHooks()` |
| `can_use_tool=` | `WithCanUseTool()` |
| `sandbox=` | `WithSandbox()` |
| `agents=` | `WithAgents()` |
| `env=` | `WithEnv()` |

### Message Types

| Python | Go |
|--------|-----|
| `AssistantMessage` | `*AssistantMessage` |
| `UserMessage` | `*UserMessage` |
| `SystemMessage` | `*SystemMessage` |
| `ResultMessage` | `*ResultMessage` |
| `TextBlock` | `TextBlock` |
| `ToolUseBlock` | `ToolUseBlock` |
| `ToolResultBlock` | `ToolResultBlock` |
| `ThinkingBlock` | `ThinkingBlock` |

### Error Types

| Python | Go |
|--------|-----|
| `ClaudeSDKError` | `*ClaudeSDKError` |
| `CLINotFoundError` | `*CLINotFoundError` |
| `CLIConnectionError` | `*CLIConnectionError` |
| `ProcessError` | `*ProcessError` |
| `CLIJSONDecodeError` | `*JSONDecodeError` |

### Hook Events

| Python | Go |
|--------|-----|
| `"PreToolUse"` | `HookEventPreToolUse` |
| `"PostToolUse"` | `HookEventPostToolUse` |
| `"PostToolUseFailure"` | `HookEventPostToolUseFailed` |
| `"UserPromptSubmit"` | `HookEventUserPromptSubmit` |
| `"Stop"` | `HookEventStop` |
| `"SubagentStop"` | `HookEventSubagentStop` |
| `"PreCompact"` | `HookEventPreCompact` |

## Behavioral Differences

### Async vs Channels

**Python:**
```python
async for message in query(prompt="Hello"):
    print(message)
```

**Go:**
```go
messages, errors := claude.Query(ctx, "Hello")
for msg := range messages {
    fmt.Println(msg)
}
```

Go uses channels instead of async iterators. This requires handling both message and error channels, typically with `select`.

### Context Manager vs Explicit Lifecycle

**Python:**
```python
async with ClaudeSDKClient(options) as client:
    await client.query("Hello")
```

**Go:**
```go
client := claude.NewClient(opts...)
client.Connect(ctx)
defer client.Close()
client.Query(ctx, "Hello")
```

Go uses explicit `Connect()` and `Close()` calls instead of Python's context manager.

### Tool Definition

**Python:**
```python
@tool("add", "Add numbers", {"a": int, "b": int})
async def add(args):
    return {"content": [{"type": "text", "text": str(args["a"] + args["b"])}]}
```

**Go:**
```go
claude.Tool("add", "Add numbers",
    claude.SimpleInputSchema(map[string]string{"a": "number", "b": "number"}),
    func(ctx context.Context, args map[string]any) (claude.MCPToolResult, error) {
        a := args["a"].(float64)
        b := args["b"].(float64)
        return claude.TextResult(fmt.Sprintf("%g", a+b)), nil
    },
)
```

Go requires explicit type assertions when accessing arguments.

### Options Configuration

**Python:**
```python
options = ClaudeAgentOptions(
    model="claude-sonnet-4-5",
    max_turns=5,
    system_prompt="Be helpful"
)
query(prompt="Hello", options=options)
```

**Go:**
```go
claude.Query(ctx, "Hello",
    claude.WithModel("claude-sonnet-4-5"),
    claude.WithMaxTurns(5),
    claude.WithSystemPrompt("Be helpful"),
)
```

Go uses functional options pattern instead of a configuration class.

### Error Handling

**Python:**
```python
try:
    async for message in query(prompt="Hello"):
        pass
except CLINotFoundError:
    print("CLI not found")
```

**Go:**
```go
messages, errors := claude.Query(ctx, "Hello")
select {
case err := <-errors:
    if _, ok := err.(*claude.CLINotFoundError); ok {
        fmt.Println("CLI not found")
    }
case msg := <-messages:
    // handle message
}
```

Go separates messages and errors into different channels.

### Type Checking

**Python:**
```python
if isinstance(message, AssistantMessage):
    for block in message.content:
        if isinstance(block, TextBlock):
            print(block.text)
```

**Go:**
```go
if m, ok := msg.(*claude.AssistantMessage); ok {
    for _, block := range m.Content {
        if text, ok := block.(claude.TextBlock); ok {
            fmt.Println(text.Text)
        }
    }
}
```

Go uses type assertions instead of `isinstance()`.

## Go-Specific Additions

The Go SDK includes some conveniences not in the Python SDK:

- `SimpleInputSchema()` - Creates JSON schemas from simple type mappings
- `TextResult()`, `ErrorResult()`, `ImageResult()` - Helper functions for tool results
- `ReceiveResponse()` - Convenience method for single-response workflows

## Migration Examples

### Simple Query

**Python:**
```python
async for message in query(prompt="Hello"):
    if isinstance(message, AssistantMessage):
        print(message.content[0].text)
```

**Go:**
```go
messages, _ := claude.Query(ctx, "Hello")
for msg := range messages {
    if m, ok := msg.(*claude.AssistantMessage); ok {
        if text, ok := m.Content[0].(claude.TextBlock); ok {
            fmt.Println(text.Text)
        }
    }
}
```

### Interactive Client

**Python:**
```python
async with ClaudeSDKClient(options) as client:
    await client.query("Hello")
    async for msg in client.receive_response():
        print(msg)
```

**Go:**
```go
client := claude.NewClient(opts...)
client.Connect(ctx)
defer client.Close()

client.Query(ctx, "Hello")
for msg := range client.ReceiveResponse(ctx) {
    fmt.Println(msg)
}
```

### Custom Tool

**Python:**
```python
@tool("greet", "Greet user", {"name": str})
async def greet(args):
    return {"content": [{"type": "text", "text": f"Hello, {args['name']}!"}]}

server = create_sdk_mcp_server("tools", "1.0.0", [greet])
```

**Go:**
```go
greet := claude.Tool("greet", "Greet user",
    claude.SimpleInputSchema(map[string]string{"name": "string"}),
    func(ctx context.Context, args map[string]any) (claude.MCPToolResult, error) {
        name := args["name"].(string)
        return claude.TextResult("Hello, " + name + "!"), nil
    },
)

server := claude.CreateSDKMCPServer("tools", "1.0.0", []claude.MCPTool{greet})
```
