// Example: Structured output with JSON Schema
//
// This example demonstrates how to get structured JSON output from Claude
// using JSON Schema validation. The output is guaranteed to conform to
// your specified schema.
//
// Run with: go run main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	claude "github.com/afsharalex/claude-agent-sdk-go"
)

// Person represents the structured output we expect
type Person struct {
	Name       string   `json:"name"`
	Age        int      `json:"age"`
	Occupation string   `json:"occupation"`
	Skills     []string `json:"skills"`
}

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

	// Define the JSON schema for our expected output
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "The person's full name",
			},
			"age": map[string]any{
				"type":        "integer",
				"description": "The person's age in years",
			},
			"occupation": map[string]any{
				"type":        "string",
				"description": "The person's job or profession",
			},
			"skills": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
				"description": "List of skills the person has",
			},
		},
		"required": []string{"name", "age", "occupation", "skills"},
	}

	// Create client with JSON schema output format
	client := claude.NewClient(
		claude.WithJSONSchema(schema),
	)

	fmt.Println("Connecting to Claude Code...")
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer func() { _ = client.Close() }()

	// Query Claude for structured data
	prompt := `Create a fictional person profile for a software developer
named Sarah Chen who is 32 years old. Include 3-4 relevant skills.`

	fmt.Printf("Query: %s\n\n", prompt)

	if err := client.Query(ctx, prompt); err != nil {
		log.Fatalf("Failed to query: %v\n", err)
	}

	// Process messages and extract structured output
	for msg := range client.ReceiveResponse(ctx) {
		switch m := msg.(type) {
		case *claude.AssistantMessage:
			// The response text might contain the JSON
			for _, block := range m.Content {
				if text, ok := block.(claude.TextBlock); ok {
					fmt.Printf("Raw response: %s\n\n", text.Text)
				}
			}

		case *claude.ResultMessage:
			// The structured output is available here
			if m.StructuredOutput != nil {
				fmt.Println("Structured Output:")

				// Convert to JSON for display
				jsonBytes, err := json.MarshalIndent(m.StructuredOutput, "", "  ")
				if err != nil {
					log.Fatalf("Failed to marshal structured output: %v\n", err)
				}
				fmt.Println(string(jsonBytes))

				// Parse into our Go struct
				var person Person
				jsonBytes, _ = json.Marshal(m.StructuredOutput)
				if err := json.Unmarshal(jsonBytes, &person); err != nil {
					log.Fatalf("Failed to unmarshal to Person: %v\n", err)
				}

				fmt.Printf("\nParsed Person:\n")
				fmt.Printf("  Name: %s\n", person.Name)
				fmt.Printf("  Age: %d\n", person.Age)
				fmt.Printf("  Occupation: %s\n", person.Occupation)
				fmt.Printf("  Skills: %v\n", person.Skills)
			}

			if m.TotalCostUSD != nil {
				fmt.Printf("\nCost: $%.4f\n", *m.TotalCostUSD)
			}
		}
	}
}
