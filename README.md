# Claude Agent SDK for Go

Go SDK for Claude Agent. This SDK provides a native Go interface for interacting with Claude Code.

## Installation

```bash
go get github.com/afsharalex/claude-agent-sdk-go
```

**Prerequisites:**

- Go 1.21+
- Claude Code CLI installed: `curl -fsSL https://claude.ai/install.sh | bash`

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    claude "github.com/afsharalex/claude-agent-sdk-go"
)

func main() {
    ctx := context.Background()

    messages, errors := claude.Query(ctx, "What is 2 + 2?")

    for {
        select {
        case msg, ok := <-messages:
            if !ok {
                return
            }
            switch m := msg.(type) {
            case *claude.AssistantMessage:
                for _, block := range m.Content {
                    if text, ok := block.(claude.TextBlock); ok {
                        fmt.Println(text.Text)
                    }
                }
            case *claude.ResultMessage:
                if m.TotalCostUSD != nil {
                    fmt.Printf("Cost: $%.4f\n", *m.TotalCostUSD)
                }
            }
        case err := <-errors:
            if err != nil {
                log.Fatal(err)
            }
        }
    }
}
```

## Basic Usage: Query()

`Query()` is a function for one-shot queries to Claude Code. It returns two channels: one for messages and one for errors.

```go
import claude "github.com/afsharalex/claude-agent-sdk-go"

// Simple query
messages, errors := claude.Query(ctx, "Hello Claude")

// With options
messages, errors := claude.Query(ctx, "Tell me a joke",
    claude.WithSystemPrompt("You are a helpful assistant"),
    claude.WithMaxTurns(1),
)
```

### Using Tools

```go
messages, errors := claude.Query(ctx, "Create a hello.go file",
    claude.WithAllowedTools([]string{"Read", "Write", "Bash"}),
    claude.WithPermissionMode(claude.PermissionModeAcceptEdits),
)
```

### Working Directory

```go
messages, errors := claude.Query(ctx, "What files are here?",
    claude.WithCwd("/path/to/project"),
)
```

## Client

`Client` supports bidirectional, interactive conversations with Claude Code.

Unlike `Query()`, `Client` additionally enables **custom tools** and **hooks**, both defined as Go functions.

```go
client := claude.NewClient(
    claude.WithCwd("/path/to/project"),
    claude.WithModel("claude-sonnet-4-5"),
)

if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Close()

if err := client.Query(ctx, "Help me understand this codebase"); err != nil {
    log.Fatal(err)
}

for msg := range client.Messages() {
    // Handle messages...
}
```

### Custom Tools (SDK MCP Servers)

Custom tools are implemented as in-process MCP servers that run directly within your Go application.

```go
// Define tools
tools := []claude.MCPTool{
    claude.Tool("add", "Add two numbers",
        claude.SimpleInputSchema(map[string]string{
            "a": "number",
            "b": "number",
        }),
        func(ctx context.Context, args map[string]any) (claude.MCPToolResult, error) {
            a, _ := args["a"].(float64)
            b, _ := args["b"].(float64)
            return claude.TextResult(fmt.Sprintf("%g + %g = %g", a, b, a+b)), nil
        },
    ),
}

// Create SDK MCP server
server := claude.CreateSDKMCPServer("calculator", "1.0.0", tools)

// Use with client
client := claude.NewClient(
    claude.WithMCPServers(map[string]claude.MCPServerConfig{
        "calculator": server,
    }),
    claude.WithAllowedTools([]string{"add"}),
)
```

#### Benefits Over External MCP Servers

- **No subprocess management** - Runs in the same process as your application
- **Better performance** - No IPC overhead for tool calls
- **Simpler deployment** - Single binary instead of multiple processes
- **Easier debugging** - All code runs in the same process
- **Type safety** - Direct Go function calls

#### Mixed Server Support

You can use both SDK and external MCP servers together:

```go
client := claude.NewClient(
    claude.WithMCPServers(map[string]claude.MCPServerConfig{
        "internal": sdkServer,  // In-process SDK server
        "external": claude.MCPStdioServerConfig{  // External subprocess
            Command: "external-server",
            Args:    []string{"--port", "8080"},
        },
    }),
)
```

### Hooks

Hooks are Go functions that Claude Code invokes at specific points of the agent loop. They provide deterministic processing and automated feedback.

```go
hooks := map[claude.HookEvent][]claude.HookMatcher{
    claude.HookEventPreToolUse: {
        {
            Matcher: "Bash",
            Hooks: []claude.HookCallback{
                func(ctx context.Context, input claude.HookInput, toolUseID string, hookCtx claude.HookContext) (claude.HookOutput, error) {
                    preToolUse := input.(claude.PreToolUseHookInput)
                    fmt.Printf("Tool: %s\n", preToolUse.ToolName)

                    // Allow the tool
                    return claude.HookOutput{
                        HookSpecificOutput: claude.PreToolUseHookSpecificOutput{
                            HookEventName:      claude.HookEventPreToolUse,
                            PermissionDecision: claude.HookPermissionDecisionAllow,
                        },
                    }, nil
                },
            },
            Timeout: 30,
        },
    },
}

client := claude.NewClient(
    claude.WithHooks(hooks),
)
```

### Permission Callbacks

For fine-grained tool permission control:

```go
client := claude.NewClient(
    claude.WithCanUseTool(func(ctx context.Context, toolName string, input map[string]any, permCtx claude.ToolPermissionContext) (claude.PermissionResult, error) {
        if toolName == "Bash" {
            cmd, _ := input["command"].(string)
            if strings.Contains(cmd, "rm -rf") {
                return claude.PermissionResultDeny{
                    Message: "Destructive commands are not allowed",
                }, nil
            }
        }
        return claude.PermissionResultAllow{}, nil
    }),
)
```

## Types

Key types defined in this package:

- `Options` - Configuration options (use `With*` functions)
- `AssistantMessage`, `UserMessage`, `SystemMessage`, `ResultMessage` - Message types
- `TextBlock`, `ToolUseBlock`, `ToolResultBlock`, `ThinkingBlock` - Content blocks
- `HookEvent`, `HookMatcher`, `HookCallback` - Hook types
- `MCPTool`, `MCPToolResult` - MCP tool types

## Error Handling

```go
import claude "github.com/afsharalex/claude-agent-sdk-go"

messages, errors := claude.Query(ctx, "Hello")

select {
case err := <-errors:
    switch e := err.(type) {
    case *claude.CLINotFoundError:
        fmt.Println("Please install Claude Code")
    case *claude.CLIConnectionError:
        fmt.Printf("Connection failed: %s\n", e.Message)
    case *claude.ProcessError:
        fmt.Printf("Process failed with exit code: %d\n", e.ExitCode)
    case *claude.CLIJSONDecodeError:
        fmt.Printf("Failed to parse response: %s\n", e.Message)
    case *claude.ClaudeSDKError:
        fmt.Printf("SDK error: %s\n", e.Message)
    }
case msg := <-messages:
    // Handle message
}
```

## Available Options

| Option | Description |
|--------|-------------|
| `WithCwd(path)` | Set working directory |
| `WithModel(model)` | Set AI model |
| `WithSystemPrompt(prompt)` | Set system prompt |
| `WithMaxTurns(n)` | Limit conversation turns |
| `WithMaxBudgetUSD(amount)` | Set cost budget |
| `WithPermissionMode(mode)` | Set permission mode |
| `WithAllowedTools(tools)` | Specify allowed tools |
| `WithDisallowedTools(tools)` | Specify disallowed tools |
| `WithMCPServers(servers)` | Configure MCP servers |
| `WithHooks(hooks)` | Register hooks |
| `WithCanUseTool(callback)` | Set permission callback |
| `WithSandbox(settings)` | Configure sandbox |
| `WithAgents(agents)` | Define subagents |
| `WithEnv(env)` | Set environment variables |

See `options.go` for all available options.

## Examples

See the `examples/` directory for complete working examples:

- [examples/simple/](examples/simple/) - Basic one-shot query
- [examples/streaming/](examples/streaming/) - Interactive streaming client
- [examples/mcp/](examples/mcp/) - Custom MCP tools
- [examples/hooks/](examples/hooks/) - Using hooks

Run an example:

```bash
cd examples/simple
go run main.go
```

## Feature Parity with Python SDK

This Go SDK implements full feature parity with the [Python Claude Agent SDK](https://github.com/anthropics/claude-agent-sdk-python), with Go-idiomatic adaptations:

| Python | Go |
|--------|-----|
| `@tool` decorator | `claude.Tool()` function |
| `async with client` | `client.Connect()` / `client.Close()` |
| `async for message` | `for msg := range messages` |
| `snake_case` | `camelCase` / `PascalCase` |
| `dataclass` | Go structs with JSON tags |

## License

Use of this SDK is governed by Anthropic's [Commercial Terms of Service](https://www.anthropic.com/legal/commercial-terms).
