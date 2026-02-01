# Getting Started

This tutorial guides you through your first interaction with the Claude Agent SDK for Go. By the end, you'll have a working program that queries Claude and displays the response.

## What We'll Build

We'll create a simple Go program that:
1. Sends a question to Claude Code
2. Receives and displays the response
3. Shows the cost of the query

## Prerequisites

Before we begin, verify you have:

- [ ] Go 1.25 or later installed (`go version`)
- [ ] Claude Code CLI installed (`claude --version`)

If you need to install Claude Code:
```bash
curl -fsSL https://claude.ai/install.sh | bash
```

## Step 1: Create a New Project

Create a new directory and initialize a Go module:

```bash
mkdir claude-demo
cd claude-demo
go mod init claude-demo
```

## Step 2: Install the SDK

Add the Claude Agent SDK as a dependency:

```bash
go get github.com/afsharalex/claude-agent-sdk-go
```

## Step 3: Write Your First Query

Create a file named `main.go`:

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

    // Send a query to Claude
    messages, errors := claude.Query(ctx, "What is the capital of France?")

    // Process the response
    for {
        select {
        case msg, ok := <-messages:
            if !ok {
                // Channel closed, we're done
                return
            }

            switch m := msg.(type) {
            case *claude.AssistantMessage:
                // Print Claude's response
                for _, block := range m.Content {
                    if text, ok := block.(claude.TextBlock); ok {
                        fmt.Println(text.Text)
                    }
                }

            case *claude.ResultMessage:
                // Print cost when complete
                if m.TotalCostUSD != nil {
                    fmt.Printf("\nCost: $%.4f\n", *m.TotalCostUSD)
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

## Step 4: Run the Program

Execute your program:

```bash
go run main.go
```

You should see output like:

```
The capital of France is Paris.

Cost: $0.0012
```

Congratulations! You've successfully queried Claude using the Go SDK.

## Understanding the Response

The `Query()` function returns two channels:

- **`messages`** - Receives different message types as Claude responds
- **`errors`** - Receives any errors that occur

Messages come in several types:

| Type | Description |
|------|-------------|
| `*AssistantMessage` | Claude's response, containing content blocks |
| `*ResultMessage` | Query completion with cost and timing info |
| `*UserMessage` | User input or tool results |
| `*SystemMessage` | System notifications |

Content blocks within `AssistantMessage` include:

| Type | Description |
|------|-------------|
| `TextBlock` | Plain text response |
| `ToolUseBlock` | When Claude uses a tool |
| `ThinkingBlock` | Extended thinking (if enabled) |

## Step 5: Add Configuration

Let's enhance our program with configuration options:

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

    // Query with options
    messages, errors := claude.Query(ctx, "Write a haiku about Go programming",
        claude.WithSystemPrompt("You are a creative poet who writes concise verse."),
        claude.WithMaxTurns(1),
    )

    for {
        select {
        case msg, ok := <-messages:
            if !ok {
                return
            }

            if m, ok := msg.(*claude.AssistantMessage); ok {
                for _, block := range m.Content {
                    if text, ok := block.(claude.TextBlock); ok {
                        fmt.Println(text.Text)
                    }
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

Run it:

```bash
go run main.go
```

## Common Options

| Option | Purpose |
|--------|---------|
| `WithSystemPrompt(prompt)` | Set Claude's behavior |
| `WithMaxTurns(n)` | Limit conversation turns |
| `WithModel(model)` | Choose the AI model |
| `WithCwd(path)` | Set working directory |
| `WithMaxBudgetUSD(amount)` | Set cost limit |

## Next Steps

Now that you have the basics working:

- **[Streaming Guide](guides/streaming.md)** - Build interactive conversations
- **[Custom Tools Guide](guides/custom-tools.md)** - Create tools Claude can use
- **[Hooks Guide](guides/hooks.md)** - Intercept and modify tool calls
- **[API Reference](reference.md)** - Complete API documentation

## Troubleshooting

### "Claude Code not found"

Ensure Claude Code is installed and in your PATH:
```bash
claude --version
```

### "Connection failed"

Check that Claude Code is properly authenticated:
```bash
claude auth status
```

### "Permission denied"

Some operations require explicit permission. Use `WithPermissionMode()`:
```go
claude.Query(ctx, "Create a file",
    claude.WithPermissionMode(claude.PermissionModeAcceptEdits),
)
```
