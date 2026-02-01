// Package claude provides a Go SDK for interacting with Claude Code.
package claude

import (
	"fmt"
)

// ClaudeSDKError is the base error type for all Claude SDK errors.
type ClaudeSDKError struct {
	Message string
	Cause   error
}

func (e *ClaudeSDKError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *ClaudeSDKError) Unwrap() error {
	return e.Cause
}

// NewClaudeSDKError creates a new ClaudeSDKError with the given message.
func NewClaudeSDKError(message string) *ClaudeSDKError {
	return &ClaudeSDKError{Message: message}
}

// WrapClaudeSDKError wraps an error with a ClaudeSDKError.
func WrapClaudeSDKError(message string, cause error) *ClaudeSDKError {
	return &ClaudeSDKError{Message: message, Cause: cause}
}

// CLIConnectionError is raised when unable to connect to Claude Code.
type CLIConnectionError struct {
	ClaudeSDKError
}

// NewCLIConnectionError creates a new CLIConnectionError.
func NewCLIConnectionError(message string) *CLIConnectionError {
	return &CLIConnectionError{ClaudeSDKError{Message: message}}
}

// WrapCLIConnectionError wraps an error with a CLIConnectionError.
func WrapCLIConnectionError(message string, cause error) *CLIConnectionError {
	return &CLIConnectionError{ClaudeSDKError{Message: message, Cause: cause}}
}

// CLINotFoundError is raised when Claude Code is not found or not installed.
type CLINotFoundError struct {
	CLIConnectionError
	CLIPath string
}

// NewCLINotFoundError creates a new CLINotFoundError.
func NewCLINotFoundError(message string, cliPath string) *CLINotFoundError {
	if cliPath != "" {
		message = fmt.Sprintf("%s: %s", message, cliPath)
	}
	return &CLINotFoundError{
		CLIConnectionError: CLIConnectionError{ClaudeSDKError{Message: message}},
		CLIPath:            cliPath,
	}
}

// ProcessError is raised when the CLI process fails.
type ProcessError struct {
	ClaudeSDKError
	ExitCode int
	Stderr   string
}

// NewProcessError creates a new ProcessError.
func NewProcessError(message string, exitCode int, stderr string) *ProcessError {
	fullMessage := message
	if exitCode != 0 {
		fullMessage = fmt.Sprintf("%s (exit code: %d)", fullMessage, exitCode)
	}
	if stderr != "" {
		fullMessage = fmt.Sprintf("%s\nError output: %s", fullMessage, stderr)
	}
	return &ProcessError{
		ClaudeSDKError: ClaudeSDKError{Message: fullMessage},
		ExitCode:       exitCode,
		Stderr:         stderr,
	}
}

// JSONDecodeError is raised when unable to decode JSON from CLI output.
type JSONDecodeError struct {
	ClaudeSDKError
	Line          string
	OriginalError error
}

// NewJSONDecodeError creates a new JSONDecodeError.
func NewJSONDecodeError(line string, originalError error) *JSONDecodeError {
	truncatedLine := line
	if len(truncatedLine) > 100 {
		truncatedLine = truncatedLine[:100] + "..."
	}
	return &JSONDecodeError{
		ClaudeSDKError: ClaudeSDKError{
			Message: fmt.Sprintf("Failed to decode JSON: %s", truncatedLine),
			Cause:   originalError,
		},
		Line:          line,
		OriginalError: originalError,
	}
}

// MessageParseError is raised when unable to parse a message from CLI output.
type MessageParseError struct {
	ClaudeSDKError
	Data map[string]any
}

// NewMessageParseError creates a new MessageParseError.
func NewMessageParseError(message string, data map[string]any) *MessageParseError {
	return &MessageParseError{
		ClaudeSDKError: ClaudeSDKError{Message: message},
		Data:           data,
	}
}
