# Custom Tools Guide

How to create and use custom tools that Claude can invoke.

## Define a Custom Tool

Use `claude.Tool()` to define a tool:

```go
addTool := claude.Tool(
    "add",                    // Tool name
    "Add two numbers",        // Description
    map[string]any{           // JSON Schema for input
        "type": "object",
        "properties": map[string]any{
            "a": map[string]any{"type": "number"},
            "b": map[string]any{"type": "number"},
        },
        "required": []string{"a", "b"},
    },
    func(ctx context.Context, args map[string]any) (claude.MCPToolResult, error) {
        a := args["a"].(float64)
        b := args["b"].(float64)
        return claude.TextResult(fmt.Sprintf("%g + %g = %g", a, b, a+b)), nil
    },
)
```

## Use SimpleInputSchema

For common cases, use `SimpleInputSchema()`:

```go
schema := claude.SimpleInputSchema(map[string]string{
    "name": "string",
    "age":  "number",
    "active": "boolean",
})
// Equivalent to full JSON schema with required fields
```

Supported types: `"string"`, `"number"`, `"integer"`, `"boolean"`

## Create an SDK MCP Server

Bundle tools into an MCP server:

```go
tools := []claude.MCPTool{
    claude.Tool("add", "Add numbers", addSchema, addHandler),
    claude.Tool("subtract", "Subtract numbers", subSchema, subHandler),
    claude.Tool("multiply", "Multiply numbers", mulSchema, mulHandler),
}

server := claude.CreateSDKMCPServer("calculator", "1.0.0", tools)
```

## Register Tools with Client

Pass the server to the client:

```go
client := claude.NewClient(
    claude.WithMCPServers(map[string]claude.MCPServerConfig{
        "calculator": server,
    }),
    claude.WithAllowedTools([]string{"add", "subtract", "multiply"}),
)

client.Connect(ctx)
defer client.Close()
```

## Handle Tool Arguments

Arguments arrive as `map[string]any`. Type assert carefully:

```go
func handler(ctx context.Context, args map[string]any) (claude.MCPToolResult, error) {
    // String argument
    name, ok := args["name"].(string)
    if !ok {
        return claude.ErrorResult("name must be a string"), nil
    }

    // Number argument (always float64 from JSON)
    age, ok := args["age"].(float64)
    if !ok {
        return claude.ErrorResult("age must be a number"), nil
    }

    // Optional argument
    var title string
    if t, ok := args["title"].(string); ok {
        title = t
    }

    // Array argument
    tags, ok := args["tags"].([]any)
    if ok {
        for _, tag := range tags {
            if s, ok := tag.(string); ok {
                // Process tag
            }
        }
    }

    return claude.TextResult("Success"), nil
}
```

## Return Different Result Types

### Text Result
```go
return claude.TextResult("Operation completed"), nil
```

### Error Result
```go
if x < 0 {
    return claude.ErrorResult("Value must be positive"), nil
}
```

### Image Result
```go
imageData := base64.StdEncoding.EncodeToString(pngBytes)
return claude.ImageResult(imageData, "image/png"), nil
```

### Multiple Content Items
```go
return claude.MCPToolResult{
    Content: []claude.MCPContent{
        {Type: "text", Text: "Result:"},
        {Type: "text", Text: fmt.Sprintf("%v", result)},
    },
}, nil
```

## Access Context in Handlers

Use the context for cancellation and values:

```go
func handler(ctx context.Context, args map[string]any) (claude.MCPToolResult, error) {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return claude.ErrorResult("cancelled"), ctx.Err()
    default:
    }

    // Long-running operation with context
    result, err := doExpensiveOperation(ctx, args)
    if err != nil {
        return claude.ErrorResult(err.Error()), nil
    }

    return claude.TextResult(result), nil
}
```

## Combine with External MCP Servers

Mix SDK and external servers:

```go
internalServer := claude.CreateSDKMCPServer("internal", "1.0.0", internalTools)

client := claude.NewClient(
    claude.WithMCPServers(map[string]claude.MCPServerConfig{
        "internal": internalServer,
        "external": claude.MCPStdioServerConfig{
            Command: "my-external-server",
            Args:    []string{"--port", "8080"},
        },
    }),
)
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math"

    claude "github.com/afsharalex/claude-agent-sdk-go"
)

func main() {
    ctx := context.Background()

    // Define tools
    tools := []claude.MCPTool{
        claude.Tool("sqrt", "Calculate square root",
            claude.SimpleInputSchema(map[string]string{"x": "number"}),
            func(ctx context.Context, args map[string]any) (claude.MCPToolResult, error) {
                x := args["x"].(float64)
                if x < 0 {
                    return claude.ErrorResult("Cannot sqrt negative number"), nil
                }
                return claude.TextResult(fmt.Sprintf("%.4f", math.Sqrt(x))), nil
            },
        ),
        claude.Tool("pow", "Calculate power",
            claude.SimpleInputSchema(map[string]string{
                "base":     "number",
                "exponent": "number",
            }),
            func(ctx context.Context, args map[string]any) (claude.MCPToolResult, error) {
                base := args["base"].(float64)
                exp := args["exponent"].(float64)
                return claude.TextResult(fmt.Sprintf("%.4f", math.Pow(base, exp))), nil
            },
        ),
    }

    // Create server and client
    server := claude.CreateSDKMCPServer("math", "1.0.0", tools)
    client := claude.NewClient(
        claude.WithMCPServers(map[string]claude.MCPServerConfig{
            "math": server,
        }),
        claude.WithAllowedTools([]string{"sqrt", "pow"}),
    )

    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Query using tools
    client.Query(ctx, "Calculate sqrt(16) and 2^10")

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

## See Also

- [Hooks Guide](hooks.md) - Intercept tool calls
- [Permissions Guide](permissions.md) - Control tool access
- [API Reference](../reference.md) - Complete API
