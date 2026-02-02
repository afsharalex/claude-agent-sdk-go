# Error Handling Guide

How to handle errors from the Claude Agent SDK.

## Use Error Helper Functions

The SDK provides convenient helper functions for checking and extracting error types. These follow Go idioms and work with wrapped errors:

```go
messages, errors := claude.Query(ctx, "Hello")

select {
case err := <-errors:
    if claude.IsCLINotFoundError(err) {
        fmt.Println("Claude Code CLI is not installed")
        fmt.Println("Install with: curl -fsSL https://claude.ai/install.sh | bash")
    } else if claude.IsConnectionError(err) {
        fmt.Println("Failed to connect to Claude Code")
    } else if claude.IsProcessError(err) {
        if procErr, ok := claude.AsProcessError(err); ok {
            fmt.Printf("Process failed with exit code: %d\n", procErr.ExitCode)
            if procErr.Stderr != "" {
                fmt.Printf("Stderr: %s\n", procErr.Stderr)
            }
        }
    } else {
        fmt.Printf("Unknown error: %v\n", err)
    }
case msg := <-messages:
    // Handle message
}
```

### Available Helper Functions

| Function | Description |
|----------|-------------|
| `IsConnectionError(err)` | Returns true if err is a `CLIConnectionError` |
| `IsCLINotFoundError(err)` | Returns true if err is a `CLINotFoundError` |
| `IsProcessError(err)` | Returns true if err is a `ProcessError` |
| `AsConnectionError(err)` | Extracts `*CLIConnectionError` from err |
| `AsCLINotFoundError(err)` | Extracts `*CLINotFoundError` from err |
| `AsProcessError(err)` | Extracts `*ProcessError` from err |

These helpers use `errors.As` internally and work correctly with wrapped errors.

## Handle CLI Not Found

Check if Claude Code is installed:

```go
messages, errors := claude.Query(ctx, "Hello")

select {
case err := <-errors:
    if notFound, ok := err.(*claude.CLINotFoundError); ok {
        fmt.Println("Claude Code CLI is not installed.")
        fmt.Printf("Expected at: %s\n", notFound.CLIPath)
        fmt.Println("Install with: curl -fsSL https://claude.ai/install.sh | bash")
        return
    }
case msg := <-messages:
    // Handle message
}
```

## Handle Connection Failures

Catch connection issues:

```go
client := claude.NewClient()

if err := client.Connect(ctx); err != nil {
    if connErr, ok := err.(*claude.CLIConnectionError); ok {
        fmt.Printf("Failed to connect: %s\n", connErr.Message)

        // Check if it's specifically a "not found" error
        if _, ok := err.(*claude.CLINotFoundError); ok {
            fmt.Println("CLI not installed")
        }
        return
    }
    log.Fatal(err)
}
```

## Handle Process Errors

Capture CLI process failures:

```go
messages, errors := claude.Query(ctx, "Hello")

for {
    select {
    case err := <-errors:
        if procErr, ok := err.(*claude.ProcessError); ok {
            fmt.Printf("Process failed with exit code: %d\n", procErr.ExitCode)
            if procErr.Stderr != "" {
                fmt.Printf("Stderr: %s\n", procErr.Stderr)
            }
            return
        }
    case msg, ok := <-messages:
        if !ok {
            return
        }
        // Handle message
    }
}
```

## Handle JSON Decode Errors

Catch malformed responses:

```go
messages, errors := claude.Query(ctx, "Hello")

for {
    select {
    case err := <-errors:
        if jsonErr, ok := err.(*claude.JSONDecodeError); ok {
            fmt.Printf("Failed to parse response: %s\n", jsonErr.Message)
            fmt.Printf("Raw line: %s\n", jsonErr.Line)
            // Log for debugging
            log.Printf("JSON error: %v", jsonErr.OriginalError)
            return
        }
    case msg, ok := <-messages:
        if !ok {
            return
        }
        // Handle message
    }
}
```

## Implement Retry Logic

Retry failed operations:

```go
func queryWithRetry(ctx context.Context, prompt string, maxRetries int) error {
    var lastErr error

    for attempt := 0; attempt < maxRetries; attempt++ {
        if attempt > 0 {
            // Exponential backoff
            delay := time.Duration(1<<attempt) * time.Second
            time.Sleep(delay)
        }

        messages, errors := claude.Query(ctx, prompt)

        success := true
        for {
            select {
            case err := <-errors:
                if err != nil {
                    lastErr = err
                    success = false
                    // Check if retryable
                    if isRetryable(err) {
                        break
                    }
                    return err // Non-retryable error
                }
            case msg, ok := <-messages:
                if !ok {
                    if success {
                        return nil
                    }
                    break
                }
                handleMessage(msg)
            }
        }
    }

    return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

func isRetryable(err error) bool {
    // Connection errors are often transient
    if claude.IsConnectionError(err) {
        return true
    }
    // Process errors might be transient
    if procErr, ok := claude.AsProcessError(err); ok {
        // Retry on specific exit codes
        return procErr.ExitCode == 1
    }
    return false
}
```

## Handle Context Cancellation

Properly handle timeouts and cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

messages, errors := claude.Query(ctx, "Long running query")

for {
    select {
    case <-ctx.Done():
        if ctx.Err() == context.DeadlineExceeded {
            fmt.Println("Query timed out")
        } else if ctx.Err() == context.Canceled {
            fmt.Println("Query was canceled")
        }
        return
    case err := <-errors:
        if err != nil {
            log.Fatal(err)
        }
    case msg, ok := <-messages:
        if !ok {
            return
        }
        handleMessage(msg)
    }
}
```

## Wrap Errors with Context

Add context to errors:

```go
func processQuery(ctx context.Context, prompt string) error {
    messages, errors := claude.Query(ctx, prompt)

    for {
        select {
        case err := <-errors:
            if err != nil {
                return fmt.Errorf("query failed for prompt %q: %w", prompt, err)
            }
        case msg, ok := <-messages:
            if !ok {
                return nil
            }
            if err := handleMessage(msg); err != nil {
                return fmt.Errorf("failed to handle message: %w", err)
            }
        }
    }
}
```

## Check Error Types

Use type switches for comprehensive handling:

```go
func handleError(err error) {
    switch e := err.(type) {
    case *claude.CLINotFoundError:
        fmt.Printf("CLI not found at: %s\n", e.CLIPath)
        fmt.Println("Install Claude Code to continue")

    case *claude.CLIConnectionError:
        fmt.Printf("Connection error: %s\n", e.Message)
        fmt.Println("Check that Claude Code is running")

    case *claude.ProcessError:
        fmt.Printf("Process error (exit code %d)\n", e.ExitCode)
        if e.Stderr != "" {
            fmt.Printf("Details: %s\n", e.Stderr)
        }

    case *claude.JSONDecodeError:
        fmt.Println("Response parsing error")
        log.Printf("Raw: %s, Error: %v", e.Line, e.OriginalError)

    case *claude.MessageParseError:
        fmt.Println("Message format error")
        log.Printf("Data: %v", e.Data)

    case *claude.ClaudeSDKError:
        fmt.Printf("SDK error: %s\n", e.Message)
        if e.Cause != nil {
            fmt.Printf("Caused by: %v\n", e.Cause)
        }

    default:
        fmt.Printf("Unknown error: %v\n", err)
    }
}
```

## Log Errors for Debugging

Create a logging wrapper:

```go
func queryWithLogging(ctx context.Context, prompt string) {
    log.Printf("[QUERY] Starting: %s", prompt)

    messages, errors := claude.Query(ctx, prompt)

    for {
        select {
        case err := <-errors:
            if err != nil {
                log.Printf("[ERROR] %T: %v", err, err)
                return
            }
        case msg, ok := <-messages:
            if !ok {
                log.Printf("[QUERY] Completed")
                return
            }
            log.Printf("[MESSAGE] %T received", msg)
            handleMessage(msg)
        }
    }
}
```

## Graceful Degradation

Fall back when errors occur:

```go
func getResponse(ctx context.Context, prompt string) string {
    messages, errors := claude.Query(ctx, prompt,
        claude.WithMaxTurns(1),
    )

    for {
        select {
        case err := <-errors:
            if err != nil {
                log.Printf("Query failed: %v", err)
                return fallbackResponse(prompt)
            }
        case msg, ok := <-messages:
            if !ok {
                return fallbackResponse(prompt)
            }
            if m, ok := msg.(*claude.AssistantMessage); ok {
                for _, block := range m.Content {
                    if text, ok := block.(claude.TextBlock); ok {
                        return text.Text
                    }
                }
            }
        }
    }
}

func fallbackResponse(prompt string) string {
    return "Sorry, I couldn't process your request. Please try again."
}
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    claude "github.com/afsharalex/claude-agent-sdk-go"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    err := runQuery(ctx, "What is 2 + 2?")
    if err != nil {
        handleError(err)
        return
    }
    fmt.Println("Query completed successfully")
}

func runQuery(ctx context.Context, prompt string) error {
    messages, errors := claude.Query(ctx, prompt)

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()

        case err := <-errors:
            if err != nil {
                return err
            }

        case msg, ok := <-messages:
            if !ok {
                return nil
            }

            switch m := msg.(type) {
            case *claude.AssistantMessage:
                for _, block := range m.Content {
                    if text, ok := block.(claude.TextBlock); ok {
                        fmt.Println(text.Text)
                    }
                }
            case *claude.ResultMessage:
                if m.IsError {
                    return claude.NewClaudeSDKError("query returned error result")
                }
            }
        }
    }
}

func handleError(err error) {
    switch e := err.(type) {
    case *claude.CLINotFoundError:
        fmt.Println("Error: Claude Code is not installed.")
        fmt.Println("Install it with: curl -fsSL https://claude.ai/install.sh | bash")

    case *claude.CLIConnectionError:
        fmt.Printf("Error: Could not connect to Claude Code: %s\n", e.Message)

    case *claude.ProcessError:
        fmt.Printf("Error: Claude Code process failed (exit code %d)\n", e.ExitCode)

    default:
        fmt.Printf("Error: %v\n", err)
    }
}
```

## See Also

- [Getting Started](../getting-started.md) - Basic setup
- [Sessions Guide](sessions.md) - Managing sessions
- [API Reference](../reference.md) - Error types reference
