// Example: Comprehensive error handling
//
// This example demonstrates proper error handling with the SDK's typed errors.
// It shows how to use Is* and As* helper functions to identify and extract
// specific error types.
//
// Run with: go run main.go
package main

import (
	"context"
	"errors"
	"fmt"
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

	// Demonstrate error handling with intentionally wrong CLI path
	fmt.Println("=== Testing CLI Not Found Error ===")
	testCLINotFound(ctx)

	fmt.Println("\n=== Testing Normal Operation ===")
	testNormalOperation(ctx)

	fmt.Println("\n=== Testing Error Type Checking ===")
	testErrorTypeChecking()
}

func testCLINotFound(ctx context.Context) {
	// Try to connect with a non-existent CLI path
	client := claude.NewClient(
		claude.WithCLIPath("/nonexistent/path/to/claude"),
	)

	err := client.Connect(ctx)
	if err != nil {
		handleError(err)
	}
}

func testNormalOperation(ctx context.Context) {
	client := claude.NewClient()

	if err := client.Connect(ctx); err != nil {
		handleError(err)
		return
	}
	defer func() { _ = client.Close() }()

	fmt.Println("Connected successfully!")

	// Send a simple query
	if err := client.Query(ctx, "Say 'Hello' in one word."); err != nil {
		handleError(err)
		return
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
			if m.IsError {
				fmt.Printf("Query returned with error flag\n")
			}
		}
	}

	// Check for errors from the error channel
	select {
	case err := <-client.Errors():
		if err != nil {
			handleError(err)
		}
	default:
		fmt.Println("No errors occurred")
	}
}

func testErrorTypeChecking() {
	// Create sample errors for demonstration
	errors := []error{
		claude.NewCLIConnectionError("connection failed"),
		claude.NewCLINotFoundError("CLI not found", "/usr/bin/claude"),
		claude.NewProcessError("process failed", 1, "permission denied"),
		claude.NewJSONDecodeError(`{"invalid": json}`, fmt.Errorf("parse error")),
		claude.NewMessageParseError("unknown message type", map[string]any{"type": "unknown"}),
	}

	for _, err := range errors {
		fmt.Printf("\nError: %v\n", err)
		identifyError(err)
	}
}

func handleError(err error) {
	fmt.Printf("Error occurred: %v\n", err)
	identifyError(err)
}

func identifyError(err error) {
	// Check for specific error types using Is* functions
	switch {
	case claude.IsCLINotFoundError(err):
		if notFoundErr, ok := claude.AsCLINotFoundError(err); ok {
			fmt.Printf("  Type: CLI Not Found\n")
			fmt.Printf("  Path: %s\n", notFoundErr.CLIPath)
			fmt.Println("  Solution: Install Claude Code or specify correct path with WithCLIPath()")
		}

	case claude.IsConnectionError(err):
		if connErr, ok := claude.AsConnectionError(err); ok {
			fmt.Printf("  Type: Connection Error\n")
			fmt.Printf("  Message: %s\n", connErr.Message)
			fmt.Println("  Solution: Ensure Claude Code is running and accessible")
		}

	case claude.IsProcessError(err):
		if procErr, ok := claude.AsProcessError(err); ok {
			fmt.Printf("  Type: Process Error\n")
			fmt.Printf("  Exit Code: %d\n", procErr.ExitCode)
			fmt.Printf("  Stderr: %s\n", procErr.Stderr)
			fmt.Println("  Solution: Check CLI output for details")
		}

	case claude.IsJSONDecodeError(err):
		if jsonErr, ok := claude.AsJSONDecodeError(err); ok {
			fmt.Printf("  Type: JSON Decode Error\n")
			fmt.Printf("  Line: %s\n", jsonErr.Line)
			fmt.Println("  Solution: This may indicate a CLI communication issue")
		}

	case claude.IsMessageParseError(err):
		if parseErr, ok := claude.AsMessageParseError(err); ok {
			fmt.Printf("  Type: Message Parse Error\n")
			fmt.Printf("  Data: %v\n", parseErr.Data)
			fmt.Println("  Solution: This may indicate an SDK/CLI version mismatch")
		}

	default:
		// Standard error handling
		var sdkErr *claude.ClaudeSDKError
		if errors.As(err, &sdkErr) {
			fmt.Printf("  Type: General SDK Error\n")
			fmt.Printf("  Message: %s\n", sdkErr.Message)
		} else {
			fmt.Printf("  Type: Unknown Error\n")
		}
	}
}
