# Sessions Guide

How to maintain conversation context and manage sessions.

## Maintain Conversation Context

The `Client` maintains context across multiple queries:

```go
client := claude.NewClient()
client.Connect(ctx)
defer client.Close()

// First query
client.Query(ctx, "My name is Alice")
for msg := range client.ReceiveResponse(ctx) {
    // Handle response
}

// Second query - Claude remembers context
client.Query(ctx, "What is my name?")
for msg := range client.ReceiveResponse(ctx) {
    // Claude will respond "Alice"
}
```

## Resume a Previous Session

Use `WithResume()` to continue a previous conversation:

```go
// First session
client1 := claude.NewClient()
client1.Connect(ctx)

client1.Query(ctx, "Remember: the secret code is 12345")
for msg := range client1.ReceiveResponse(ctx) {
    if result, ok := msg.(*claude.ResultMessage); ok {
        sessionID := result.SessionID
        // Save sessionID for later
        saveSession(sessionID)
    }
}
client1.Close()

// Later, resume the session
savedID := loadSession()
client2 := claude.NewClient(
    claude.WithResume(savedID),
)
client2.Connect(ctx)

client2.Query(ctx, "What is the secret code?")
for msg := range client2.ReceiveResponse(ctx) {
    // Claude will respond "12345"
}
```

## Fork a Session

Create a new session branching from a resumed one:

```go
client := claude.NewClient(
    claude.WithResume(existingSessionID),
    claude.WithForkSession(true), // Create a new session ID
)
client.Connect(ctx)

// This creates a new session with the context from existingSessionID
// Changes here won't affect the original session
```

## Continue Last Conversation

Resume the most recent conversation:

```go
client := claude.NewClient(
    claude.WithContinueConversation(true),
)
client.Connect(ctx)

// Continues from whatever conversation was last active
```

## Set Session ID Manually

Assign a custom session identifier:

```go
client := claude.NewClient()
client.Connect(ctx)

// Set a custom session ID
client.SetSessionID("my-custom-session-123")

client.Query(ctx, "Hello")
// Messages will be tagged with this session ID
```

## Get Session Info from Results

Extract session information from results:

```go
for msg := range client.Messages() {
    if result, ok := msg.(*claude.ResultMessage); ok {
        fmt.Printf("Session ID: %s\n", result.SessionID)
        fmt.Printf("Duration: %dms\n", result.DurationMs)
        fmt.Printf("Turns: %d\n", result.NumTurns)
    }
}
```

## Isolate Sessions

Create isolated sessions for different contexts:

```go
// Session for user A
clientA := claude.NewClient(
    claude.WithCwd("/home/userA/project"),
)
clientA.Connect(ctx)
clientA.SetSessionID("user-a-session")

// Session for user B (completely separate)
clientB := claude.NewClient(
    claude.WithCwd("/home/userB/project"),
)
clientB.Connect(ctx)
clientB.SetSessionID("user-b-session")

// These sessions have no shared context
```

## Rewind Files in Session

Revert files to a previous state within a session:

```go
client := claude.NewClient(
    claude.WithEnableFileCheckpointing(true),
    claude.WithExtraArg("replay-user-messages", nil),
)
client.Connect(ctx)

var userMessageIDs []string

// Track user message IDs
go func() {
    for msg := range client.Messages() {
        if user, ok := msg.(*claude.UserMessage); ok {
            userMessageIDs = append(userMessageIDs, user.UUID)
        }
    }
}()

// Make some file changes
client.Query(ctx, "Create test.txt with 'version 1'")
// ... wait for response

client.Query(ctx, "Update test.txt to 'version 2'")
// ... wait for response

// Rewind to after first query
if len(userMessageIDs) > 0 {
    err := client.RewindFiles(ctx, userMessageIDs[0])
    // test.txt now contains 'version 1'
}
```

## Multiple Concurrent Sessions

Run multiple sessions simultaneously:

```go
var wg sync.WaitGroup

sessions := []string{"session-1", "session-2", "session-3"}

for _, sessionID := range sessions {
    wg.Add(1)
    go func(id string) {
        defer wg.Done()

        client := claude.NewClient()
        client.Connect(ctx)
        defer client.Close()

        client.SetSessionID(id)
        client.Query(ctx, fmt.Sprintf("Session %s checking in", id))

        for msg := range client.ReceiveResponse(ctx) {
            // Handle response for this session
        }
    }(sessionID)
}

wg.Wait()
```

## Session with Custom Settings

Load specific settings for a session:

```go
client := claude.NewClient(
    claude.WithSettings("/path/to/session-settings.json"),
    claude.WithSettingSources([]claude.SettingSource{
        claude.SettingSourceLocal, // Only load local settings
    }),
)
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    claude "github.com/afsharalex/claude-agent-sdk-go"
)

const sessionFile = "session.txt"

func main() {
    ctx := context.Background()

    // Check for existing session
    savedSession, _ := os.ReadFile(sessionFile)
    sessionID := string(savedSession)

    var client *claude.Client
    if sessionID != "" {
        fmt.Println("Resuming previous session...")
        client = claude.NewClient(
            claude.WithResume(sessionID),
        )
    } else {
        fmt.Println("Starting new session...")
        client = claude.NewClient()
    }

    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Send a query
    client.Query(ctx, "Hello! What have we discussed so far?")

    // Process response and save session ID
    for msg := range client.ReceiveResponse(ctx) {
        switch m := msg.(type) {
        case *claude.AssistantMessage:
            for _, block := range m.Content {
                if text, ok := block.(claude.TextBlock); ok {
                    fmt.Println(text.Text)
                }
            }
        case *claude.ResultMessage:
            // Save session for next time
            os.WriteFile(sessionFile, []byte(m.SessionID), 0644)
            fmt.Printf("\nSession saved: %s\n", m.SessionID)
        }
    }
}
```

## See Also

- [Streaming Guide](streaming.md) - Interactive conversations
- [Error Handling Guide](error-handling.md) - Handling failures
- [API Reference](../reference.md) - Complete API
