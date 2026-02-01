// Example: Using SDK MCP servers with custom tools
//
// This example demonstrates creating in-process MCP servers that provide
// custom tools for Claude to use. These tools run in your Go application,
// giving you full control and better performance than external servers.
//
// Run with: go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"math"
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

	// Create custom tools
	tools := []claude.MCPTool{
		// Calculator tool - adds two numbers
		claude.Tool("add", "Add two numbers together",
			claude.SimpleInputSchema(map[string]string{
				"a": "number",
				"b": "number",
			}),
			func(ctx context.Context, args map[string]any) (claude.MCPToolResult, error) {
				a, _ := args["a"].(float64)
				b, _ := args["b"].(float64)
				result := a + b
				return claude.TextResult(fmt.Sprintf("The sum of %g and %g is %g", a, b, result)), nil
			},
		),

		// Calculator tool - multiplies two numbers
		claude.Tool("multiply", "Multiply two numbers together",
			claude.SimpleInputSchema(map[string]string{
				"a": "number",
				"b": "number",
			}),
			func(ctx context.Context, args map[string]any) (claude.MCPToolResult, error) {
				a, _ := args["a"].(float64)
				b, _ := args["b"].(float64)
				result := a * b
				return claude.TextResult(fmt.Sprintf("The product of %g and %g is %g", a, b, result)), nil
			},
		),

		// Calculator tool - calculates square root
		claude.Tool("sqrt", "Calculate the square root of a number",
			claude.SimpleInputSchema(map[string]string{
				"x": "number",
			}),
			func(ctx context.Context, args map[string]any) (claude.MCPToolResult, error) {
				x, _ := args["x"].(float64)
				if x < 0 {
					return claude.ErrorResult("Cannot calculate square root of negative number"), nil
				}
				result := math.Sqrt(x)
				return claude.TextResult(fmt.Sprintf("The square root of %g is %g", x, result)), nil
			},
		),
	}

	// Create the SDK MCP server
	calculatorServer := claude.CreateSDKMCPServer("calculator", "1.0.0", tools)

	// Create client with MCP server
	client := claude.NewClient(
		claude.WithMCPServers(map[string]claude.MCPServerConfig{
			"calculator": calculatorServer,
		}),
		claude.WithAllowedTools([]string{"add", "multiply", "sqrt"}),
	)

	// Connect
	fmt.Println("Connecting to Claude Code with MCP calculator server...")
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer client.Close()

	// Send a query that will use the calculator tools
	fmt.Println("Sending query...")
	if err := client.Query(ctx, "Calculate (25 * 4) + sqrt(144). Show your work step by step."); err != nil {
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
					fmt.Printf("\n[Using tool: %s with args: %v]\n", b.Name, b.Input)
				}
			}
			fmt.Println()

		case *claude.UserMessage:
			// Show tool results
			if blocks := m.GetContentBlocks(); blocks != nil {
				for _, block := range blocks {
					if result, ok := block.(claude.ToolResultBlock); ok {
						if content, ok := result.Content.(string); ok {
							fmt.Printf("[Tool result: %s]\n", content)
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
