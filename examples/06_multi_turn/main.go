// Example: Structured multi-turn conversation
//
// This example demonstrates having a structured multi-turn conversation
// with Claude. It shows how to maintain context across multiple queries
// within a single session.
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

	// Create and connect client
	client := claude.NewClient()

	fmt.Println("Connecting to Claude Code...")
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer func() { _ = client.Close() }()

	// Define a sequence of queries that build on each other
	queries := []string{
		"My name is Alice. Please remember that.",
		"What's my name?",
		"Please create a simple greeting function in Go that uses my name.",
	}

	for i, query := range queries {
		fmt.Printf("\n--- Turn %d ---\n", i+1)
		fmt.Printf("You: %s\n", query)

		// Send the query
		if err := client.Query(ctx, query); err != nil {
			log.Fatalf("Failed to send query: %v\n", err)
		}

		// Process the response
		fmt.Print("Claude: ")
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
				if m.TotalCostUSD != nil {
					fmt.Printf("(Cost: $%.4f, Turns: %d)\n", *m.TotalCostUSD, m.NumTurns)
				}
			}
		}
	}

	fmt.Println("\n--- Conversation Complete ---")
}
