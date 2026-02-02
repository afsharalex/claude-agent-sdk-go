// Example: Debugging output and diagnostics
//
// This example demonstrates debugging features:
// - WithDebugStderr: Quick way to output CLI stderr to os.Stderr
// - WithStderr: Custom callback for stderr handling
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
	"sync"
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

	fmt.Println("=== Example 1: Simple Debug Output ===")
	simpleDebugExample(ctx)

	fmt.Println("\n=== Example 2: Custom Stderr Handler ===")
	customStderrExample(ctx)
}

func simpleDebugExample(ctx context.Context) {
	// WithDebugStderr is a convenience function that outputs all CLI
	// stderr to os.Stderr. This is useful for quick debugging.
	client := claude.NewClient(
		claude.WithDebugStderr(), // Enable debug output
	)

	fmt.Println("Connecting with debug stderr enabled...")
	fmt.Println("(CLI stderr will appear below)")
	fmt.Println("---")

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer func() { _ = client.Close() }()

	// Simple query
	if err := client.Query(ctx, "What is 1+1? One word answer."); err != nil {
		log.Fatalf("Failed to query: %v\n", err)
	}

	for msg := range client.ReceiveResponse(ctx) {
		switch m := msg.(type) {
		case *claude.AssistantMessage:
			for _, block := range m.Content {
				if text, ok := block.(claude.TextBlock); ok {
					fmt.Printf("Claude: %s\n", text.Text)
				}
			}
		case *claude.ResultMessage:
			fmt.Println("---")
			fmt.Println("Query complete")
		}
	}
}

func customStderrExample(ctx context.Context) {
	// Collect and analyze stderr output
	var stderrLines []string
	var mu sync.Mutex

	// WithStderr lets you process each line of stderr output
	client := claude.NewClient(
		claude.WithStderr(func(line string) {
			mu.Lock()
			defer mu.Unlock()

			// Filter or transform stderr output
			if strings.Contains(line, "error") || strings.Contains(line, "Error") {
				// Highlight errors
				fmt.Printf("[ERROR] %s\n", line)
			} else if strings.Contains(line, "warn") || strings.Contains(line, "Warn") {
				// Highlight warnings
				fmt.Printf("[WARN] %s\n", line)
			}
			// Collect all lines for analysis
			stderrLines = append(stderrLines, line)
		}),
	)

	fmt.Println("Connecting with custom stderr handler...")

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer func() { _ = client.Close() }()

	// Simple query
	if err := client.Query(ctx, "What is 2+2? One word answer."); err != nil {
		log.Fatalf("Failed to query: %v\n", err)
	}

	for msg := range client.ReceiveResponse(ctx) {
		switch m := msg.(type) {
		case *claude.AssistantMessage:
			for _, block := range m.Content {
				if text, ok := block.(claude.TextBlock); ok {
					fmt.Printf("Claude: %s\n", text.Text)
				}
			}
		case *claude.ResultMessage:
			// Done
		}
	}

	// Print diagnostic summary
	mu.Lock()
	defer mu.Unlock()

	fmt.Println("\n--- Stderr Summary ---")
	fmt.Printf("Total stderr lines: %d\n", len(stderrLines))

	// Count error/warning lines
	var errors, warnings int
	for _, line := range stderrLines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "error") {
			errors++
		}
		if strings.Contains(lower, "warn") {
			warnings++
		}
	}
	fmt.Printf("Errors: %d, Warnings: %d\n", errors, warnings)
}
