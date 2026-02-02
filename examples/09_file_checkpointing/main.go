// Example: File checkpointing
//
// This example demonstrates file checkpointing, which tracks file changes
// made by Claude. This allows you to review, revert, or manage changes
// Claude makes to your files.
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

	// Create a temporary directory for this example
	tmpDir, err := os.MkdirTemp("", "checkpointing-example-*")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v\n", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	fmt.Printf("Working directory: %s\n\n", tmpDir)

	// Create client with file checkpointing enabled
	client := claude.NewClient(
		claude.WithCwd(tmpDir),
		claude.WithFileCheckpointing(), // Enable checkpointing
		// Also available as: claude.WithEnableFileCheckpointing(true)
	)

	fmt.Println("Connecting to Claude Code with file checkpointing enabled...")
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer func() { _ = client.Close() }()

	// Ask Claude to create a file
	prompt := "Create a file called hello.txt with the content 'Hello, World!'"
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
					fmt.Printf("\n[Tool: %s]\n", b.Name)
				}
			}
			fmt.Println()

		case *claude.UserMessage:
			// Tool results show what Claude did
			if blocks := m.GetContentBlocks(); blocks != nil {
				for _, block := range blocks {
					if result, ok := block.(claude.ToolResultBlock); ok {
						if content, ok := result.Content.(string); ok {
							fmt.Printf("[Result: %s]\n", content)
						}
					}
				}
			}

		case *claude.SystemMessage:
			// System messages may include checkpoint information
			fmt.Printf("[System: %s]\n", m.Subtype)
			if m.Data != nil {
				if checkpoint, ok := m.Data["checkpoint"]; ok {
					fmt.Printf("  Checkpoint: %v\n", checkpoint)
				}
			}

		case *claude.ResultMessage:
			fmt.Println()
			if m.TotalCostUSD != nil {
				fmt.Printf("Cost: $%.4f\n", *m.TotalCostUSD)
			}
		}
	}

	// Verify the file was created
	content, err := os.ReadFile(tmpDir + "/hello.txt")
	if err != nil {
		fmt.Printf("\nFile not created (expected if Claude didn't have permission)\n")
	} else {
		fmt.Printf("\nFile content: %s\n", string(content))
	}

	fmt.Println("\nWith file checkpointing enabled, you can use the Claude Code CLI")
	fmt.Println("to view and revert file changes made during this session.")
}
