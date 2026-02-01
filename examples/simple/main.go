// Example: Simple one-shot query to Claude Code
//
// This example demonstrates the most basic usage of the Claude Agent SDK.
// It sends a simple question and prints the response.
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

	// Simple query - ask Claude a question
	prompt := "What is the capital of France? Please give a brief answer."

	fmt.Printf("Sending query: %s\n\n", prompt)

	// Query returns two channels: messages and errors
	messages, errors := claude.Query(ctx, prompt)

	// Process messages as they arrive
	for {
		select {
		case msg, ok := <-messages:
			if !ok {
				// Messages channel closed, we're done
				return
			}

			switch m := msg.(type) {
			case *claude.AssistantMessage:
				// Print text content from assistant messages
				for _, block := range m.Content {
					if text, ok := block.(claude.TextBlock); ok {
						fmt.Print(text.Text)
					}
				}
				fmt.Println()

			case *claude.ResultMessage:
				// Print cost information when the query completes
				fmt.Println()
				if m.TotalCostUSD != nil {
					fmt.Printf("Total cost: $%.4f\n", *m.TotalCostUSD)
				}
				fmt.Printf("Duration: %dms\n", m.DurationMs)
			}

		case err := <-errors:
			if err != nil {
				log.Fatalf("Error: %v\n", err)
			}
		}
	}
}
