// Example: Sandbox mode for bash commands
//
// This example demonstrates sandbox mode, which isolates bash commands
// for improved security. When sandboxing is enabled, commands run in
// a restricted environment with limited file and network access.
//
// Note: Sandboxing is only available on macOS and Linux.
//
// Run with: go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
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

	// Check platform compatibility
	if runtime.GOOS == "windows" {
		fmt.Println("Note: Sandboxing is not available on Windows.")
		fmt.Println("This example will run without sandbox restrictions.")
	}

	// Create sandbox configuration
	sandbox := &claude.SandboxSettings{
		// Enable sandboxing
		Enabled: true,

		// Auto-approve bash commands when sandboxed (safe due to isolation)
		AutoAllowBashIfSandboxed: true,

		// Commands that should bypass the sandbox
		ExcludedCommands: []string{
			"git", // Allow git to run outside sandbox for proper repo access
		},

		// Network configuration
		Network: &claude.SandboxNetworkConfig{
			// Allow binding to localhost (useful for development servers)
			AllowLocalBinding: true,
		},
	}

	// Create client with sandbox enabled
	client := claude.NewClient(
		claude.WithSandbox(sandbox),
	)

	fmt.Println("Connecting to Claude Code with sandboxing enabled...")
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer func() { _ = client.Close() }()

	// Example query that will use bash commands
	prompt := "Run 'pwd' to show the current directory, then list files with 'ls -la'."
	fmt.Printf("Query: %s\n\n", prompt)

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
					fmt.Printf("\n[Sandboxed Tool: %s]\n", b.Name)
					if b.Name == "Bash" {
						if cmd, ok := b.Input["command"].(string); ok {
							fmt.Printf("  Command: %s\n", cmd)
						}
					}
				}
			}
			fmt.Println()

		case *claude.UserMessage:
			// Show tool results
			if blocks := m.GetContentBlocks(); blocks != nil {
				for _, block := range blocks {
					if result, ok := block.(claude.ToolResultBlock); ok {
						if content, ok := result.Content.(string); ok {
							// Truncate long output
							if len(content) > 500 {
								content = content[:500] + "..."
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

	fmt.Println("\n--- Sandbox Configuration Summary ---")
	fmt.Printf("Enabled: %v\n", sandbox.Enabled)
	fmt.Printf("Auto-allow when sandboxed: %v\n", sandbox.AutoAllowBashIfSandboxed)
	fmt.Printf("Excluded commands: %v\n", sandbox.ExcludedCommands)
}
