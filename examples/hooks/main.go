// Example: Using hooks to intercept and modify tool calls
//
// This example demonstrates using hooks to:
// - Log tool usage before and after execution
// - Optionally modify or block tool calls
// - Add context to Claude based on tool results
//
// Run with: go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	claude "github.com/afsharalex/claude-agent-sdk-go"
)

func main() {
	// Create a context that cancels on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nInterrupted, shutting down...")
		cancel()
	}()

	// Define hooks
	hooks := map[claude.HookEvent][]claude.HookMatcher{
		// PreToolUse hook - runs before any tool is executed
		claude.HookEventPreToolUse: {
			{
				// Match all tools (nil matcher matches everything)
				Matcher: "",
				Hooks: []claude.HookCallback{
					preToolUseHook,
				},
				Timeout: 30, // 30 second timeout
			},
		},

		// PostToolUse hook - runs after successful tool execution
		claude.HookEventPostToolUse: {
			{
				Matcher: "",
				Hooks: []claude.HookCallback{
					postToolUseHook,
				},
			},
		},

		// UserPromptSubmit hook - runs when user submits a prompt
		claude.HookEventUserPromptSubmit: {
			{
				Matcher: "",
				Hooks: []claude.HookCallback{
					userPromptSubmitHook,
				},
			},
		},
	}

	// Create client with hooks
	client := claude.NewClient(
		claude.WithHooks(hooks),
	)

	// Connect
	fmt.Println("Connecting to Claude Code with hooks...")
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer func() { _ = client.Close() }()

	// Send a query that will trigger tool use
	fmt.Println("Sending query...")
	if err := client.Query(ctx, "What files are in the current directory?"); err != nil {
		log.Fatalf("Failed to query: %v\n", err)
	}

	// Process messages
	for msg := range client.ReceiveResponse(ctx) {
		switch m := msg.(type) {
		case *claude.AssistantMessage:
			for _, block := range m.Content {
				if text, ok := block.(claude.TextBlock); ok {
					fmt.Print(text.Text)
				}
			}
			fmt.Println()

		case *claude.ResultMessage:
			fmt.Println()
			if m.TotalCostUSD != nil {
				fmt.Printf("Cost: $%.4f\n", *m.TotalCostUSD)
			}
		}
	}
}

// preToolUseHook is called before a tool is executed.
// It can allow, deny, or modify the tool call.
func preToolUseHook(ctx context.Context, input claude.HookInput, toolUseID string, hookCtx claude.HookContext) (claude.HookOutput, error) {
	preToolUse, ok := input.(claude.PreToolUseHookInput)
	if !ok {
		return claude.HookOutput{}, nil
	}

	fmt.Printf("\n[Hook] PreToolUse: %s\n", preToolUse.ToolName)
	fmt.Printf("  Input: %v\n", preToolUse.ToolInput)

	// Example: Block certain tools
	if preToolUse.ToolName == "DangerousTool" {
		return claude.HookOutput{
			HookSpecificOutput: claude.PreToolUseHookSpecificOutput{
				HookEventName:            claude.HookEventPreToolUse,
				PermissionDecision:       claude.HookPermissionDecisionDeny,
				PermissionDecisionReason: "This tool is blocked by policy",
			},
		}, nil
	}

	// Example: Modify tool input
	// if preToolUse.ToolName == "SomeTool" {
	// 	modifiedInput := maps.Clone(preToolUse.ToolInput)
	// 	modifiedInput["extra_param"] = "added by hook"
	// 	return claude.HookOutput{
	// 		HookSpecificOutput: claude.PreToolUseHookSpecificOutput{
	// 			HookEventName:      claude.HookEventPreToolUse,
	// 			PermissionDecision: claude.HookPermissionDecisionAllow,
	// 			UpdatedInput:       modifiedInput,
	// 		},
	// 	}, nil
	// }

	// Allow the tool to proceed
	return claude.HookOutput{
		HookSpecificOutput: claude.PreToolUseHookSpecificOutput{
			HookEventName:      claude.HookEventPreToolUse,
			PermissionDecision: claude.HookPermissionDecisionAllow,
		},
	}, nil
}

// postToolUseHook is called after a tool is successfully executed.
// It can add additional context for Claude to consider.
func postToolUseHook(ctx context.Context, input claude.HookInput, toolUseID string, hookCtx claude.HookContext) (claude.HookOutput, error) {
	postToolUse, ok := input.(claude.PostToolUseHookInput)
	if !ok {
		return claude.HookOutput{}, nil
	}

	fmt.Printf("\n[Hook] PostToolUse: %s\n", postToolUse.ToolName)

	// Example: Add context based on tool result
	return claude.HookOutput{
		HookSpecificOutput: claude.PostToolUseHookSpecificOutput{
			HookEventName:     claude.HookEventPostToolUse,
			AdditionalContext: "Note: This tool was executed via SDK hooks.",
		},
	}, nil
}

// userPromptSubmitHook is called when the user submits a prompt.
func userPromptSubmitHook(ctx context.Context, input claude.HookInput, toolUseID string, hookCtx claude.HookContext) (claude.HookOutput, error) {
	promptSubmit, ok := input.(claude.UserPromptSubmitHookInput)
	if !ok {
		return claude.HookOutput{}, nil
	}

	fmt.Printf("\n[Hook] UserPromptSubmit: %s\n", promptSubmit.Prompt)

	// Allow the prompt to proceed
	return claude.HookOutput{}, nil
}
