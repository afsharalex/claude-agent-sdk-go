// Example: Context manager pattern with WithClient
//
// This example demonstrates the WithClient pattern for automatic resource
// cleanup. The client is automatically closed when the callback returns,
// even if an error occurs or a panic happens.
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

	// Use WithClient for automatic resource cleanup
	// The client is automatically closed when the callback returns
	err := claude.WithClient(ctx, func(client *claude.Client) error {
		fmt.Println("Connected to Claude Code")

		// Send a query
		if err := client.Query(ctx, "What is 2 + 2? Give a one-word answer."); err != nil {
			return fmt.Errorf("failed to query: %w", err)
		}

		// Process messages
		for msg := range client.ReceiveResponse(ctx) {
			switch m := msg.(type) {
			case *claude.AssistantMessage:
				for _, block := range m.Content {
					if text, ok := block.(claude.TextBlock); ok {
						fmt.Printf("Claude: %s\n", text.Text)
					}
				}

			case *claude.ResultMessage:
				if m.TotalCostUSD != nil {
					fmt.Printf("Cost: $%.4f\n", *m.TotalCostUSD)
				}
			}
		}

		return nil
	})

	// Client is automatically closed at this point

	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	fmt.Println("Done! Client was automatically closed.")
}
