# Streaming Guide

How to handle streaming responses and build interactive conversations.

## Handle Streaming Responses

Use the `Client` type for streaming conversations:

```go
client := claude.NewClient()

if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Close()

// Send a query
if err := client.Query(ctx, "Tell me a story"); err != nil {
    log.Fatal(err)
}

// Process messages as they arrive
for msg := range client.Messages() {
    switch m := msg.(type) {
    case *claude.AssistantMessage:
        for _, block := range m.Content {
            if text, ok := block.(claude.TextBlock); ok {
                fmt.Print(text.Text)
            }
        }
    case *claude.ResultMessage:
        fmt.Printf("\nDone. Cost: $%.4f\n", *m.TotalCostUSD)
    }
}
```

## Process Messages Concurrently

Handle messages in a separate goroutine while accepting user input:

```go
client := claude.NewClient()
client.Connect(ctx)
defer client.Close()

// Start message handler
go func() {
    for {
        select {
        case msg, ok := <-client.Messages():
            if !ok {
                return
            }
            handleMessage(msg)
        case err := <-client.Errors():
            if err != nil {
                log.Println("Error:", err)
            }
        }
    }
}()

// Accept user input
reader := bufio.NewReader(os.Stdin)
for {
    fmt.Print("You: ")
    input, _ := reader.ReadString('\n')
    input = strings.TrimSpace(input)

    if input == "/quit" {
        break
    }

    client.Query(ctx, input)
}
```

## Wait for Complete Response

Use `ReceiveResponse()` to wait for a complete response:

```go
client.Query(ctx, "What is 2 + 2?")

// Wait for all messages until ResultMessage
for msg := range client.ReceiveResponse(ctx) {
    if m, ok := msg.(*claude.AssistantMessage); ok {
        for _, block := range m.Content {
            if text, ok := block.(claude.TextBlock); ok {
                fmt.Println(text.Text)
            }
        }
    }
}
// Channel closes after ResultMessage
```

## Display Partial Content

Enable partial message streaming for real-time display:

```go
client := claude.NewClient(
    claude.WithIncludePartialMessages(true),
)

client.Connect(ctx)
client.Query(ctx, "Write a poem")

for msg := range client.Messages() {
    switch m := msg.(type) {
    case *claude.StreamEvent:
        // Handle partial updates
        event := m.Event
        if eventType, ok := event["type"].(string); ok {
            if eventType == "content_block_delta" {
                if delta, ok := event["delta"].(map[string]any); ok {
                    if text, ok := delta["text"].(string); ok {
                        fmt.Print(text) // Print as it arrives
                    }
                }
            }
        }
    case *claude.AssistantMessage:
        // Complete message also arrives
    }
}
```

## Handle Multiple Queries

Send multiple queries in the same session:

```go
client := claude.NewClient()
client.Connect(ctx)
defer client.Close()

queries := []string{
    "What is Go?",
    "What are its main features?",
    "Show me a simple example",
}

for _, q := range queries {
    fmt.Printf("\n> %s\n", q)

    if err := client.Query(ctx, q); err != nil {
        log.Fatal(err)
    }

    // Wait for this response to complete
    for msg := range client.ReceiveResponse(ctx) {
        if m, ok := msg.(*claude.AssistantMessage); ok {
            for _, block := range m.Content {
                if text, ok := block.(claude.TextBlock); ok {
                    fmt.Println(text.Text)
                }
            }
        }
    }
}
```

## Interrupt Long Operations

Send an interrupt signal to stop Claude:

```go
client := claude.NewClient()
client.Connect(ctx)

client.Query(ctx, "Write a very long story")

// In another goroutine or after timeout
go func() {
    time.Sleep(5 * time.Second)
    if err := client.Interrupt(ctx); err != nil {
        log.Println("Interrupt failed:", err)
    }
}()

for msg := range client.Messages() {
    // Handle messages...
}
```

## Handle Tool Use in Streaming

Process tool use blocks as they appear:

```go
for msg := range client.Messages() {
    switch m := msg.(type) {
    case *claude.AssistantMessage:
        for _, block := range m.Content {
            switch b := block.(type) {
            case claude.TextBlock:
                fmt.Println(b.Text)
            case claude.ToolUseBlock:
                fmt.Printf("[Tool: %s]\n", b.Name)
                fmt.Printf("  Input: %v\n", b.Input)
            }
        }
    case *claude.UserMessage:
        // Tool results appear as UserMessages
        if blocks := m.GetContentBlocks(); blocks != nil {
            for _, block := range blocks {
                if result, ok := block.(claude.ToolResultBlock); ok {
                    fmt.Printf("[Result: %v]\n", result.Content)
                }
            }
        }
    }
}
```

## Use with Context Cancellation

Properly handle context cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

client := claude.NewClient()
if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Close()

client.Query(ctx, "Hello")

for {
    select {
    case <-ctx.Done():
        fmt.Println("Timeout reached")
        return
    case msg, ok := <-client.Messages():
        if !ok {
            return
        }
        // Handle message
    case err := <-client.Errors():
        if err != nil {
            log.Fatal(err)
        }
    }
}
```

## See Also

- [Getting Started](../getting-started.md) - Basic setup
- [Custom Tools Guide](custom-tools.md) - Adding tools
- [API Reference](../reference.md) - Complete API
