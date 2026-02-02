// Example: Programmatic agent definitions
//
// This example demonstrates defining custom agents programmatically.
// Agents are specialized Claude instances with custom prompts and tools
// that can be invoked for specific tasks.
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

	// Define custom agents using WithAgent (can be called multiple times)
	// or WithAgents for all at once

	// Example 1: Using WithAgent for individual agents
	fmt.Println("=== Example 1: Individual Agent Definitions ===")
	individualAgentsExample(ctx)

	// Example 2: Using WithAgents for bulk definition
	fmt.Println("\n=== Example 2: Bulk Agent Definitions ===")
	bulkAgentsExample(ctx)
}

func individualAgentsExample(ctx context.Context) {
	// Create client with individually defined agents
	client := claude.NewClient(
		// Define a code reviewer agent
		claude.WithAgent("code-reviewer", claude.AgentDefinition{
			Description: "Reviews code for quality, security, and best practices",
			Prompt: `You are a senior code reviewer. When reviewing code:
1. Check for security vulnerabilities
2. Identify performance issues
3. Suggest improvements for readability
4. Point out any violations of best practices

Be constructive and specific in your feedback.`,
			Tools: []string{"Read", "Glob", "Grep"}, // Read-only tools
			Model: "sonnet",                         // Use Sonnet for efficiency
		}),

		// Define a documentation writer agent
		claude.WithAgent("doc-writer", claude.AgentDefinition{
			Description: "Writes clear, comprehensive documentation",
			Prompt: `You are a technical writer specializing in software documentation.
Write clear, well-structured documentation that includes:
- Purpose and overview
- Usage examples
- API reference (if applicable)
- Common pitfalls and troubleshooting`,
			Tools: []string{"Read", "Write", "Edit"},
			Model: "sonnet",
		}),
	)

	fmt.Println("Connecting with individual agent definitions...")
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer func() { _ = client.Close() }()

	// Query that might use the code reviewer
	if err := client.Query(ctx, "What agents are available?"); err != nil {
		log.Fatalf("Failed to query: %v\n", err)
	}

	processResponse(ctx, client)
}

func bulkAgentsExample(ctx context.Context) {
	// Define all agents at once using WithAgents
	agents := map[string]claude.AgentDefinition{
		"test-writer": {
			Description: "Writes comprehensive unit and integration tests",
			Prompt: `You are a QA engineer specialized in test automation.
When writing tests:
- Cover edge cases and error conditions
- Use descriptive test names
- Follow the Arrange-Act-Assert pattern
- Include both positive and negative tests`,
			Tools: []string{"Read", "Write", "Bash"},
			Model: "sonnet",
		},

		"refactor-helper": {
			Description: "Helps refactor code for better maintainability",
			Prompt: `You are a software architect helping with code refactoring.
Focus on:
- Reducing code duplication
- Improving separation of concerns
- Simplifying complex logic
- Maintaining backward compatibility`,
			Tools: []string{"Read", "Write", "Edit", "Glob", "Grep"},
			Model: "opus", // Use Opus for complex refactoring
		},

		"quick-helper": {
			Description: "Quick helper for simple questions",
			Prompt:      "You provide quick, concise answers to simple questions.",
			Model:       "haiku", // Use Haiku for speed
		},
	}

	client := claude.NewClient(
		claude.WithAgents(agents),
	)

	fmt.Println("Connecting with bulk agent definitions...")
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer func() { _ = client.Close() }()

	// Show what agents we defined
	fmt.Println("\nDefined agents:")
	for name, agent := range agents {
		fmt.Printf("  - %s: %s (model: %s)\n", name, agent.Description, agent.Model)
	}

	// Simple query
	if err := client.Query(ctx, "Say hello in one word."); err != nil {
		log.Fatalf("Failed to query: %v\n", err)
	}

	processResponse(ctx, client)
}

func processResponse(ctx context.Context, client *claude.Client) {
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
}
