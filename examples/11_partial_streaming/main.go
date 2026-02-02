// Example: Partial message streaming
//
// This example demonstrates partial streaming, which provides real-time
// updates as Claude generates its response. This is useful for showing
// typing indicators or progressive rendering.
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

	// Create client with partial streaming enabled
	client := claude.NewClient(
		claude.WithPartialStreaming(), // Enable partial message streaming
		// Also available as: claude.WithIncludePartialMessages(true)
	)

	fmt.Println("Connecting to Claude Code with partial streaming enabled...")
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer func() { _ = client.Close() }()

	// Send a query that will generate a longer response
	prompt := "Write a haiku about programming."
	fmt.Printf("Query: %s\n\n", prompt)
	fmt.Println("Response (with streaming):")

	if err := client.Query(ctx, prompt); err != nil {
		log.Fatalf("Failed to query: %v\n", err)
	}

	// Track streaming state
	var lastTextLen int

	// Process messages including partial stream events
	for msg := range client.ReceiveResponse(ctx) {
		switch m := msg.(type) {
		case *claude.StreamEvent:
			// StreamEvent provides raw API stream events
			// These give real-time updates as Claude generates text
			if event := m.Event; event != nil {
				eventType, _ := event["type"].(string)

				switch eventType {
				case "content_block_delta":
					// Extract delta text for progressive display
					if delta, ok := event["delta"].(map[string]any); ok {
						if text, ok := delta["text"].(string); ok {
							// Print new text as it arrives
							fmt.Print(text)
						}
					}

				case "content_block_start":
					// A new content block is starting
					fmt.Print("[start]")

				case "content_block_stop":
					// A content block is complete
					fmt.Print("[stop]")
				}
			}

		case *claude.AssistantMessage:
			// AssistantMessage contains the current accumulated state
			// This may be partial or complete depending on streaming state
			for _, block := range m.Content {
				if text, ok := block.(claude.TextBlock); ok {
					newText := text.Text[lastTextLen:]
					if newText != "" {
						// Only print new content we haven't seen
						// (StreamEvent typically handles this)
						lastTextLen = len(text.Text)
					}
				}
			}

		case *claude.ResultMessage:
			// Query is complete
			fmt.Println()
			if m.TotalCostUSD != nil {
				fmt.Printf("Cost: $%.4f\n", *m.TotalCostUSD)
			}
			fmt.Printf("Duration: %dms\n", m.DurationMs)
		}
	}
}
