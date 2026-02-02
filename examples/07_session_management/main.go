// Example: Session management - resume and fork
//
// This example demonstrates session management features:
// - WithResume: Continue a previous session
// - WithForkSession: Create a new branch from a session
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

	// First, create a new session and get its ID
	fmt.Println("=== Creating initial session ===")
	sessionID := createInitialSession(ctx)
	if sessionID == "" {
		log.Fatal("Failed to get session ID")
	}
	fmt.Printf("Session ID: %s\n\n", sessionID)

	// Resume the session and continue the conversation
	fmt.Println("=== Resuming session ===")
	resumeSession(ctx, sessionID)

	// Fork the session to create a new branch
	fmt.Println("\n=== Forking session ===")
	forkSession(ctx, sessionID)
}

func createInitialSession(ctx context.Context) string {
	client := claude.NewClient()

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer func() { _ = client.Close() }()

	// Send initial query
	if err := client.Query(ctx, "Remember that my favorite color is blue."); err != nil {
		log.Fatalf("Failed to query: %v\n", err)
	}

	var sessionID string
	for msg := range client.ReceiveResponse(ctx) {
		switch m := msg.(type) {
		case *claude.AssistantMessage:
			for _, block := range m.Content {
				if text, ok := block.(claude.TextBlock); ok {
					fmt.Printf("Claude: %s\n", text.Text)
				}
			}
		case *claude.ResultMessage:
			sessionID = m.SessionID
			fmt.Printf("(Session: %s)\n", sessionID)
		}
	}

	return sessionID
}

func resumeSession(ctx context.Context, sessionID string) {
	// Create client that resumes the previous session
	client := claude.NewClient(
		claude.WithResume(sessionID),
	)

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer func() { _ = client.Close() }()

	// Ask about the previous context
	if err := client.Query(ctx, "What is my favorite color?"); err != nil {
		log.Fatalf("Failed to query: %v\n", err)
	}

	for msg := range client.ReceiveResponse(ctx) {
		switch m := msg.(type) {
		case *claude.AssistantMessage:
			for _, block := range m.Content {
				if text, ok := block.(claude.TextBlock); ok {
					fmt.Printf("Claude (resumed): %s\n", text.Text)
				}
			}
		case *claude.ResultMessage:
			fmt.Printf("(Same session: %s)\n", m.SessionID)
		}
	}
}

func forkSession(ctx context.Context, sessionID string) {
	// Create client that forks from the session
	// This creates a new session ID but keeps the conversation history
	client := claude.NewClient(
		claude.WithResume(sessionID),
		claude.WithForkSession(true), // Fork to new session
	)

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer func() { _ = client.Close() }()

	// This forked session still has the context
	if err := client.Query(ctx, "What color did I mention earlier? Now pretend I said green instead."); err != nil {
		log.Fatalf("Failed to query: %v\n", err)
	}

	for msg := range client.ReceiveResponse(ctx) {
		switch m := msg.(type) {
		case *claude.AssistantMessage:
			for _, block := range m.Content {
				if text, ok := block.(claude.TextBlock); ok {
					fmt.Printf("Claude (forked): %s\n", text.Text)
				}
			}
		case *claude.ResultMessage:
			fmt.Printf("(New forked session: %s)\n", m.SessionID)
		}
	}
}
