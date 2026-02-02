// Example: Fine-grained tool permission control
//
// This example demonstrates WithCanUseTool for programmatic control
// over which tools Claude can use. The callback is invoked for each
// tool use request, allowing you to approve, deny, or modify calls.
//
// Run with: go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
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

	// Define allowed directories for file operations
	allowedDirs := []string{"/tmp", os.TempDir()}

	// Create client with permission callback
	client := claude.NewClient(
		claude.WithCanUseTool(func(ctx context.Context, toolName string, input map[string]any, permCtx claude.ToolPermissionContext) (claude.PermissionResult, error) {
			fmt.Printf("\n[Permission Request]\n")
			fmt.Printf("  Tool: %s\n", toolName)
			fmt.Printf("  Input: %v\n", input)

			// Example: Allow read operations, be careful with writes
			switch toolName {
			case "Read":
				// Always allow reading files
				fmt.Println("  Decision: ALLOW (read operations are safe)")
				return claude.PermissionResultAllow{}, nil

			case "Write", "Edit":
				// Only allow writes to allowed directories
				filePath, _ := input["file_path"].(string)
				if filePath == "" {
					filePath, _ = input["path"].(string)
				}

				for _, dir := range allowedDirs {
					if strings.HasPrefix(filePath, dir) {
						fmt.Printf("  Decision: ALLOW (path %s is in allowed directory %s)\n", filePath, dir)
						return claude.PermissionResultAllow{}, nil
					}
				}

				fmt.Printf("  Decision: DENY (path %s not in allowed directories)\n", filePath)
				return claude.PermissionResultDeny{
					Message: fmt.Sprintf("Write operations only allowed in: %v", allowedDirs),
				}, nil

			case "Bash":
				// Check for dangerous commands
				command, _ := input["command"].(string)
				dangerousPatterns := []string{"rm -rf", "sudo", "chmod", "chown"}

				for _, pattern := range dangerousPatterns {
					if strings.Contains(command, pattern) {
						fmt.Printf("  Decision: DENY (command contains dangerous pattern: %s)\n", pattern)
						return claude.PermissionResultDeny{
							Message:   fmt.Sprintf("Command blocked: contains '%s'", pattern),
							Interrupt: false, // Don't stop the conversation
						}, nil
					}
				}

				fmt.Println("  Decision: ALLOW (command appears safe)")
				return claude.PermissionResultAllow{}, nil

			default:
				// Allow other tools by default
				fmt.Println("  Decision: ALLOW (default policy)")
				return claude.PermissionResultAllow{}, nil
			}
		}),
	)

	fmt.Println("Connecting to Claude Code with permission callback...")
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer func() { _ = client.Close() }()

	// Send a query that will trigger tool use
	prompt := "List the files in the current directory, then try to create a file in /tmp/test.txt"
	fmt.Printf("Query: %s\n", prompt)

	if err := client.Query(ctx, prompt); err != nil {
		log.Fatalf("Failed to query: %v\n", err)
	}

	// Process messages
	for msg := range client.ReceiveResponse(ctx) {
		switch m := msg.(type) {
		case *claude.AssistantMessage:
			for _, block := range m.Content {
				switch b := block.(type) {
				case claude.TextBlock:
					fmt.Print(b.Text)
				case claude.ToolUseBlock:
					fmt.Printf("\n[Tool Use: %s]\n", b.Name)
				}
			}
			fmt.Println()

		case *claude.UserMessage:
			// Show tool results
			if blocks := m.GetContentBlocks(); blocks != nil {
				for _, block := range blocks {
					if result, ok := block.(claude.ToolResultBlock); ok {
						if content, ok := result.Content.(string); ok {
							if len(content) > 200 {
								content = content[:200] + "..."
							}
							fmt.Printf("[Result: %s]\n", content)
						}
					}
				}
			}

		case *claude.ResultMessage:
			fmt.Println()
			if m.TotalCostUSD != nil {
				fmt.Printf("Cost: $%.4f\n", *m.TotalCostUSD)
			}
		}
	}
}
