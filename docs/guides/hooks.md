# Hooks Guide

How to intercept and modify Claude's behavior at specific points.

## Intercept Tool Execution

Use `PreToolUse` hooks to intercept tool calls before execution:

```go
hooks := map[claude.HookEvent][]claude.HookMatcher{
    claude.HookEventPreToolUse: {
        {
            Matcher: "", // Empty matches all tools
            Hooks: []claude.HookCallback{
                func(ctx context.Context, input claude.HookInput, toolUseID string, hookCtx claude.HookContext) (claude.HookOutput, error) {
                    pre := input.(claude.PreToolUseHookInput)
                    fmt.Printf("Tool: %s\n", pre.ToolName)
                    fmt.Printf("Input: %v\n", pre.ToolInput)

                    return claude.HookOutput{
                        HookSpecificOutput: claude.PreToolUseHookSpecificOutput{
                            HookEventName:      claude.HookEventPreToolUse,
                            PermissionDecision: claude.HookPermissionDecisionAllow,
                        },
                    }, nil
                },
            },
        },
    },
}

client := claude.NewClient(claude.WithHooks(hooks))
```

## Allow or Deny Tool Usage

Block specific tools or commands:

```go
func preToolHook(ctx context.Context, input claude.HookInput, toolUseID string, hookCtx claude.HookContext) (claude.HookOutput, error) {
    pre := input.(claude.PreToolUseHookInput)

    // Block dangerous commands
    if pre.ToolName == "Bash" {
        cmd, _ := pre.ToolInput["command"].(string)
        if strings.Contains(cmd, "rm -rf") {
            return claude.HookOutput{
                HookSpecificOutput: claude.PreToolUseHookSpecificOutput{
                    HookEventName:            claude.HookEventPreToolUse,
                    PermissionDecision:       claude.HookPermissionDecisionDeny,
                    PermissionDecisionReason: "Destructive commands are blocked",
                },
            }, nil
        }
    }

    // Allow everything else
    return claude.HookOutput{
        HookSpecificOutput: claude.PreToolUseHookSpecificOutput{
            HookEventName:      claude.HookEventPreToolUse,
            PermissionDecision: claude.HookPermissionDecisionAllow,
        },
    }, nil
}
```

## Modify Tool Input

Modify arguments before tool execution:

```go
func modifyInputHook(ctx context.Context, input claude.HookInput, toolUseID string, hookCtx claude.HookContext) (claude.HookOutput, error) {
    pre := input.(claude.PreToolUseHookInput)

    if pre.ToolName == "Write" {
        // Add a header to all written files
        modifiedInput := make(map[string]any)
        for k, v := range pre.ToolInput {
            modifiedInput[k] = v
        }
        if content, ok := modifiedInput["content"].(string); ok {
            modifiedInput["content"] = "// Auto-generated\n" + content
        }

        return claude.HookOutput{
            HookSpecificOutput: claude.PreToolUseHookSpecificOutput{
                HookEventName:      claude.HookEventPreToolUse,
                PermissionDecision: claude.HookPermissionDecisionAllow,
                UpdatedInput:       modifiedInput,
            },
        }, nil
    }

    return claude.HookOutput{
        HookSpecificOutput: claude.PreToolUseHookSpecificOutput{
            HookEventName:      claude.HookEventPreToolUse,
            PermissionDecision: claude.HookPermissionDecisionAllow,
        },
    }, nil
}
```

## Add Context After Tool Runs

Use `PostToolUse` to add information for Claude:

```go
hooks := map[claude.HookEvent][]claude.HookMatcher{
    claude.HookEventPostToolUse: {
        {
            Matcher: "Bash",
            Hooks: []claude.HookCallback{
                func(ctx context.Context, input claude.HookInput, toolUseID string, hookCtx claude.HookContext) (claude.HookOutput, error) {
                    post := input.(claude.PostToolUseHookInput)

                    // Add context based on tool result
                    return claude.HookOutput{
                        HookSpecificOutput: claude.PostToolUseHookSpecificOutput{
                            HookEventName:     claude.HookEventPostToolUse,
                            AdditionalContext: "Note: This command was run in a sandboxed environment.",
                        },
                    }, nil
                },
            },
        },
    },
}
```

## Log All Tool Activity

Create a logging hook:

```go
func loggingHook(ctx context.Context, input claude.HookInput, toolUseID string, hookCtx claude.HookContext) (claude.HookOutput, error) {
    switch v := input.(type) {
    case claude.PreToolUseHookInput:
        log.Printf("[PRE] Tool=%s Input=%v", v.ToolName, v.ToolInput)
    case claude.PostToolUseHookInput:
        log.Printf("[POST] Tool=%s Response=%v", v.ToolName, v.ToolResponse)
    }

    // Allow to proceed
    if _, ok := input.(claude.PreToolUseHookInput); ok {
        return claude.HookOutput{
            HookSpecificOutput: claude.PreToolUseHookSpecificOutput{
                HookEventName:      claude.HookEventPreToolUse,
                PermissionDecision: claude.HookPermissionDecisionAllow,
            },
        }, nil
    }
    return claude.HookOutput{}, nil
}

hooks := map[claude.HookEvent][]claude.HookMatcher{
    claude.HookEventPreToolUse: {
        {Matcher: "", Hooks: []claude.HookCallback{loggingHook}},
    },
    claude.HookEventPostToolUse: {
        {Matcher: "", Hooks: []claude.HookCallback{loggingHook}},
    },
}
```

## Match Specific Tools

Use the `Matcher` field to target specific tools:

```go
hooks := map[claude.HookEvent][]claude.HookMatcher{
    claude.HookEventPreToolUse: {
        // Match only Bash
        {
            Matcher: "Bash",
            Hooks:   []claude.HookCallback{bashHook},
        },
        // Match file editing tools
        {
            Matcher: "Write|Edit|MultiEdit",
            Hooks:   []claude.HookCallback{fileHook},
        },
        // Match all tools (fallback)
        {
            Matcher: "",
            Hooks:   []claude.HookCallback{defaultHook},
        },
    },
}
```

## Set Hook Timeouts

Configure timeout for hook execution:

```go
hooks := map[claude.HookEvent][]claude.HookMatcher{
    claude.HookEventPreToolUse: {
        {
            Matcher: "Bash",
            Hooks:   []claude.HookCallback{myHook},
            Timeout: 30, // 30 seconds
        },
    },
}
```

## Handle User Prompt Submit

Intercept user input before processing:

```go
hooks := map[claude.HookEvent][]claude.HookMatcher{
    claude.HookEventUserPromptSubmit: {
        {
            Matcher: "",
            Hooks: []claude.HookCallback{
                func(ctx context.Context, input claude.HookInput, toolUseID string, hookCtx claude.HookContext) (claude.HookOutput, error) {
                    prompt := input.(claude.UserPromptSubmitHookInput)
                    log.Printf("User submitted: %s", prompt.Prompt)
                    return claude.HookOutput{}, nil
                },
            },
        },
    },
}
```

## Handle Stop Events

React when conversation stops:

```go
hooks := map[claude.HookEvent][]claude.HookMatcher{
    claude.HookEventStop: {
        {
            Matcher: "",
            Hooks: []claude.HookCallback{
                func(ctx context.Context, input claude.HookInput, toolUseID string, hookCtx claude.HookContext) (claude.HookOutput, error) {
                    stop := input.(claude.StopHookInput)
                    log.Printf("Conversation stopped. Hook active: %v", stop.StopHookActive)
                    return claude.HookOutput{}, nil
                },
            },
        },
    },
}
```

## Block and Display Warning

Show a warning message when blocking:

```go
return claude.HookOutput{
    Decision:      claude.HookDecisionBlock,
    SystemMessage: "This operation is not permitted",
    Reason:        "Policy violation: destructive command detected",
    HookSpecificOutput: claude.PreToolUseHookSpecificOutput{
        HookEventName:            claude.HookEventPreToolUse,
        PermissionDecision:       claude.HookPermissionDecisionDeny,
        PermissionDecisionReason: "Blocked by security policy",
    },
}, nil
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

    hooks := map[claude.HookEvent][]claude.HookMatcher{
        claude.HookEventPreToolUse: {
            {
                Matcher: "Bash",
                Hooks:   []claude.HookCallback{bashSecurityHook},
                Timeout: 30,
            },
        },
        claude.HookEventPostToolUse: {
            {
                Matcher: "",
                Hooks:   []claude.HookCallback{loggingPostHook},
            },
        },
    }

    client := claude.NewClient(claude.WithHooks(hooks))

    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    client.Query(ctx, "List files in the current directory")

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

func bashSecurityHook(ctx context.Context, input claude.HookInput, toolUseID string, hookCtx claude.HookContext) (claude.HookOutput, error) {
    pre := input.(claude.PreToolUseHookInput)
    cmd, _ := pre.ToolInput["command"].(string)

    blocked := []string{"rm -rf", "sudo", "chmod 777"}
    for _, pattern := range blocked {
        if strings.Contains(cmd, pattern) {
            return claude.HookOutput{
                HookSpecificOutput: claude.PreToolUseHookSpecificOutput{
                    HookEventName:            claude.HookEventPreToolUse,
                    PermissionDecision:       claude.HookPermissionDecisionDeny,
                    PermissionDecisionReason: fmt.Sprintf("Command contains blocked pattern: %s", pattern),
                },
            }, nil
        }
    }

    return claude.HookOutput{
        HookSpecificOutput: claude.PreToolUseHookSpecificOutput{
            HookEventName:      claude.HookEventPreToolUse,
            PermissionDecision: claude.HookPermissionDecisionAllow,
        },
    }, nil
}

func loggingPostHook(ctx context.Context, input claude.HookInput, toolUseID string, hookCtx claude.HookContext) (claude.HookOutput, error) {
    post := input.(claude.PostToolUseHookInput)
    log.Printf("[AUDIT] Tool %s completed", post.ToolName)
    return claude.HookOutput{}, nil
}
```

## See Also

- [Permissions Guide](permissions.md) - Permission callbacks
- [Custom Tools Guide](custom-tools.md) - Creating tools
- [API Reference](../reference.md) - Complete API
