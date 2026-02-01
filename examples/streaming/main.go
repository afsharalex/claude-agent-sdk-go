// Example: Interactive streaming client
//
// This example demonstrates using the Client for interactive conversations.
// It allows you to have a multi-turn conversation with Claude.
//
// Run with: go run main.go
package main

import (
	"bufio"
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

	// Create client with options
	client := claude.NewClient(
		// claude.WithCwd("/path/to/project"), // Uncomment to set working directory
		// claude.WithModel("claude-sonnet-4-5"), // Uncomment to set model
	)

	// Connect to Claude
	fmt.Println("Connecting to Claude Code...")
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer client.Close()

	fmt.Println("Connected! Type your messages (Ctrl+C to quit)")

	// Start a goroutine to handle incoming messages
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-client.Messages():
				if !ok {
					return
				}
				handleMessage(msg)
			case err := <-client.Errors():
				if err != nil {
					fmt.Printf("\nError: %v\n", err)
				}
			}
		}
	}()

	// Read user input in a loop
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\nYou: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Handle special commands
		if input == "/quit" || input == "/exit" {
			break
		}

		if input == "/interrupt" {
			if err := client.Interrupt(ctx); err != nil {
				fmt.Printf("Failed to interrupt: %v\n", err)
			}
			continue
		}

		// Send the query
		if err := client.Query(ctx, input); err != nil {
			fmt.Printf("Failed to send query: %v\n", err)
			continue
		}
	}

	fmt.Println("\nGoodbye!")
}

func handleMessage(msg claude.Message) {
	switch m := msg.(type) {
	case *claude.AssistantMessage:
		fmt.Print("\nClaude: ")
		for _, block := range m.Content {
			switch b := block.(type) {
			case claude.TextBlock:
				fmt.Print(b.Text)
			case claude.ToolUseBlock:
				fmt.Printf("\n[Using tool: %s]\n", b.Name)
			}
		}
		fmt.Println()

	case *claude.UserMessage:
		// Echo user messages (useful for seeing tool results)
		if blocks := m.GetContentBlocks(); blocks != nil {
			for _, block := range blocks {
				if result, ok := block.(claude.ToolResultBlock); ok {
					fmt.Printf("\n[Tool result for %s]\n", result.ToolUseID)
				}
			}
		}

	case *claude.SystemMessage:
		fmt.Printf("\n[System: %s]\n", m.Subtype)

	case *claude.ResultMessage:
		fmt.Println()
		if m.TotalCostUSD != nil {
			fmt.Printf("Cost: $%.4f | ", *m.TotalCostUSD)
		}
		fmt.Printf("Duration: %dms | Turns: %d\n", m.DurationMs, m.NumTurns)
	}
}
