# Permissions Guide

How to control which tools Claude can use.

## Set Permission Mode

Use permission modes to control default behavior:

```go
// Prompt for dangerous tools (default)
client := claude.NewClient(
    claude.WithPermissionMode(claude.PermissionModeDefault),
)

// Auto-accept file edits
client := claude.NewClient(
    claude.WithPermissionMode(claude.PermissionModeAcceptEdits),
)

// Allow all tools (use with caution)
client := claude.NewClient(
    claude.WithPermissionMode(claude.PermissionModeBypassPermissions),
)
```

## Use Permission Callback

For fine-grained control, use `WithCanUseTool()`:

```go
client := claude.NewClient(
    claude.WithCanUseTool(func(ctx context.Context, toolName string, input map[string]any, permCtx claude.ToolPermissionContext) (claude.PermissionResult, error) {
        // Allow all Read operations
        if toolName == "Read" {
            return claude.PermissionResultAllow{}, nil
        }

        // Deny dangerous operations
        if toolName == "Bash" {
            cmd, _ := input["command"].(string)
            if strings.Contains(cmd, "rm -rf") {
                return claude.PermissionResultDeny{
                    Message: "Destructive commands not allowed",
                }, nil
            }
        }

        // Allow by default
        return claude.PermissionResultAllow{}, nil
    }),
)
```

## Allow Specific Tools

Restrict to specific tools:

```go
client := claude.NewClient(
    claude.WithAllowedTools([]string{"Read", "Glob", "Grep"}),
)
```

## Disallow Specific Tools

Block specific tools:

```go
client := claude.NewClient(
    claude.WithDisallowedTools([]string{"Bash", "Write"}),
)
```

## Modify Tool Input on Allow

Return modified input when allowing:

```go
func permCallback(ctx context.Context, toolName string, input map[string]any, permCtx claude.ToolPermissionContext) (claude.PermissionResult, error) {
    if toolName == "Write" {
        // Add a comment header to all written files
        modifiedInput := make(map[string]any)
        for k, v := range input {
            modifiedInput[k] = v
        }
        if content, ok := modifiedInput["content"].(string); ok {
            modifiedInput["content"] = "// Generated\n" + content
        }

        return claude.PermissionResultAllow{
            UpdatedInput: modifiedInput,
        }, nil
    }

    return claude.PermissionResultAllow{}, nil
}
```

## Interrupt on Deny

Stop the conversation when denying:

```go
return claude.PermissionResultDeny{
    Message:   "Operation not permitted",
    Interrupt: true, // Stop the conversation
}, nil
```

## Implement Role-Based Access

Check user roles before allowing tools:

```go
func permCallback(ctx context.Context, toolName string, input map[string]any, permCtx claude.ToolPermissionContext) (claude.PermissionResult, error) {
    // Get user role from context
    role := ctx.Value("user_role").(string)

    adminTools := map[string]bool{"Bash": true, "Write": true}
    if adminTools[toolName] && role != "admin" {
        return claude.PermissionResultDeny{
            Message: "Admin access required for this tool",
        }, nil
    }

    return claude.PermissionResultAllow{}, nil
}
```

## Validate File Paths

Restrict file operations to specific directories:

```go
func permCallback(ctx context.Context, toolName string, input map[string]any, permCtx claude.ToolPermissionContext) (claude.PermissionResult, error) {
    allowedDir := "/home/user/projects"

    if toolName == "Write" || toolName == "Read" {
        filePath, _ := input["file_path"].(string)
        if !strings.HasPrefix(filePath, allowedDir) {
            return claude.PermissionResultDeny{
                Message: fmt.Sprintf("Access restricted to %s", allowedDir),
            }, nil
        }
    }

    return claude.PermissionResultAllow{}, nil
}
```

## Log Permission Decisions

Track all permission checks:

```go
func permCallback(ctx context.Context, toolName string, input map[string]any, permCtx claude.ToolPermissionContext) (claude.PermissionResult, error) {
    // Log the request
    log.Printf("[PERMISSION] Tool=%s Input=%v", toolName, input)

    // Make decision
    allowed := checkPermission(toolName, input)

    // Log the decision
    if allowed {
        log.Printf("[PERMISSION] ALLOWED: %s", toolName)
        return claude.PermissionResultAllow{}, nil
    }

    log.Printf("[PERMISSION] DENIED: %s", toolName)
    return claude.PermissionResultDeny{
        Message: "Permission denied by policy",
    }, nil
}
```

## Change Permission Mode Mid-Session

Update permissions during a conversation:

```go
client := claude.NewClient()
client.Connect(ctx)

// Start with default permissions
client.Query(ctx, "Read the file contents")
// ... handle response

// Escalate permissions for editing
if err := client.SetPermissionMode(ctx, claude.PermissionModeAcceptEdits); err != nil {
    log.Fatal(err)
}

client.Query(ctx, "Now edit the file")
// ... handle response
```

## Combine Permission Controls

Use multiple layers of control:

```go
client := claude.NewClient(
    // Base permission mode
    claude.WithPermissionMode(claude.PermissionModeDefault),

    // Explicit allow/deny lists
    claude.WithAllowedTools([]string{"Read", "Write", "Bash"}),
    claude.WithDisallowedTools([]string{"WebFetch"}),

    // Custom callback for fine-grained control
    claude.WithCanUseTool(customPermissionCheck),
)
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "strings"

    claude "github.com/afsharalex/claude-agent-sdk-go"
)

func main() {
    ctx := context.Background()

    client := claude.NewClient(
        claude.WithAllowedTools([]string{"Read", "Write", "Bash"}),
        claude.WithCanUseTool(securityCallback),
    )

    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    client.Query(ctx, "List files and create a summary.txt")

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

func securityCallback(ctx context.Context, toolName string, input map[string]any, permCtx claude.ToolPermissionContext) (claude.PermissionResult, error) {
    // Block dangerous bash commands
    if toolName == "Bash" {
        cmd, _ := input["command"].(string)
        dangerous := []string{"rm -rf", "sudo", "> /dev/", "mkfs"}
        for _, pattern := range dangerous {
            if strings.Contains(cmd, pattern) {
                log.Printf("[BLOCKED] Dangerous command: %s", cmd)
                return claude.PermissionResultDeny{
                    Message: "Command blocked by security policy",
                }, nil
            }
        }
    }

    // Restrict Write to current directory
    if toolName == "Write" {
        filePath, _ := input["file_path"].(string)
        if strings.Contains(filePath, "..") {
            return claude.PermissionResultDeny{
                Message: "Cannot write outside current directory",
            }, nil
        }
    }

    log.Printf("[ALLOWED] %s", toolName)
    return claude.PermissionResultAllow{}, nil
}
```

## See Also

- [Hooks Guide](hooks.md) - Pre/post tool hooks
- [Streaming Guide](streaming.md) - Interactive sessions
- [API Reference](../reference.md) - Complete API
