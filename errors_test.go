package claude

import (
	"errors"
	"strings"
	"testing"
)

func TestClaudeSDKError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ClaudeSDKError
		expected string
	}{
		{
			name:     "without cause",
			err:      &ClaudeSDKError{Message: "simple error"},
			expected: "simple error",
		},
		{
			name:     "with cause",
			err:      &ClaudeSDKError{Message: "wrapped error", Cause: errors.New("underlying cause")},
			expected: "wrapped error: underlying cause",
		},
		{
			name:     "empty message without cause",
			err:      &ClaudeSDKError{Message: ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestClaudeSDKError_Unwrap(t *testing.T) {
	cause := errors.New("original error")
	err := &ClaudeSDKError{Message: "wrapped", Cause: cause}

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("Expected unwrapped error to be %v, got %v", cause, unwrapped)
	}

	// Test nil cause
	errNoCause := &ClaudeSDKError{Message: "no cause"}
	if errNoCause.Unwrap() != nil {
		t.Error("Expected nil when unwrapping error without cause")
	}
}

func TestNewClaudeSDKError(t *testing.T) {
	err := NewClaudeSDKError("test error message")

	if err.Message != "test error message" {
		t.Errorf("Expected message 'test error message', got '%s'", err.Message)
	}
	if err.Cause != nil {
		t.Error("Expected Cause to be nil")
	}
}

func TestWrapClaudeSDKError(t *testing.T) {
	cause := errors.New("root cause")
	err := WrapClaudeSDKError("context message", cause)

	if err.Message != "context message" {
		t.Errorf("Expected message 'context message', got '%s'", err.Message)
	}
	if err.Cause != cause {
		t.Errorf("Expected Cause to be %v, got %v", cause, err.Cause)
	}
}

func TestCLIConnectionError(t *testing.T) {
	err := NewCLIConnectionError("failed to connect")

	if err.Message != "failed to connect" {
		t.Errorf("Expected message 'failed to connect', got '%s'", err.Message)
	}
	if err.Error() != "failed to connect" {
		t.Errorf("Expected Error() 'failed to connect', got '%s'", err.Error())
	}
}

func TestWrapCLIConnectionError(t *testing.T) {
	cause := errors.New("network timeout")
	err := WrapCLIConnectionError("connection failed", cause)

	if err.Message != "connection failed" {
		t.Errorf("Expected message 'connection failed', got '%s'", err.Message)
	}
	if err.Cause != cause {
		t.Errorf("Expected Cause to be %v, got %v", cause, err.Cause)
	}
	if !strings.Contains(err.Error(), "connection failed") {
		t.Errorf("Expected Error() to contain 'connection failed', got '%s'", err.Error())
	}
	if !strings.Contains(err.Error(), "network timeout") {
		t.Errorf("Expected Error() to contain 'network timeout', got '%s'", err.Error())
	}
}

func TestCLINotFoundError_WithPath(t *testing.T) {
	err := NewCLINotFoundError("claude not found", "/usr/local/bin/claude")

	if !strings.Contains(err.Message, "claude not found") {
		t.Errorf("Expected message to contain 'claude not found', got '%s'", err.Message)
	}
	if !strings.Contains(err.Message, "/usr/local/bin/claude") {
		t.Errorf("Expected message to contain path, got '%s'", err.Message)
	}
	if err.CLIPath != "/usr/local/bin/claude" {
		t.Errorf("Expected CLIPath '/usr/local/bin/claude', got '%s'", err.CLIPath)
	}
}

func TestCLINotFoundError_WithoutPath(t *testing.T) {
	err := NewCLINotFoundError("claude not found", "")

	if err.Message != "claude not found" {
		t.Errorf("Expected message 'claude not found', got '%s'", err.Message)
	}
	if err.CLIPath != "" {
		t.Errorf("Expected empty CLIPath, got '%s'", err.CLIPath)
	}
}

func TestProcessError_WithExitCodeAndStderr(t *testing.T) {
	err := NewProcessError("process failed", 1, "permission denied")

	if !strings.Contains(err.Message, "process failed") {
		t.Errorf("Expected message to contain 'process failed', got '%s'", err.Message)
	}
	if !strings.Contains(err.Message, "exit code: 1") {
		t.Errorf("Expected message to contain 'exit code: 1', got '%s'", err.Message)
	}
	if !strings.Contains(err.Message, "permission denied") {
		t.Errorf("Expected message to contain stderr, got '%s'", err.Message)
	}
	if err.ExitCode != 1 {
		t.Errorf("Expected ExitCode 1, got %d", err.ExitCode)
	}
	if err.Stderr != "permission denied" {
		t.Errorf("Expected Stderr 'permission denied', got '%s'", err.Stderr)
	}
}

func TestProcessError_WithZeroExitCode(t *testing.T) {
	err := NewProcessError("process failed", 0, "warning message")

	// Should not include exit code when it's 0
	if strings.Contains(err.Message, "exit code: 0") {
		t.Errorf("Expected message to not contain 'exit code: 0', got '%s'", err.Message)
	}
	if !strings.Contains(err.Message, "warning message") {
		t.Errorf("Expected message to contain stderr, got '%s'", err.Message)
	}
}

func TestProcessError_WithoutStderr(t *testing.T) {
	err := NewProcessError("process failed", 2, "")

	if !strings.Contains(err.Message, "exit code: 2") {
		t.Errorf("Expected message to contain 'exit code: 2', got '%s'", err.Message)
	}
	if strings.Contains(err.Message, "Error output") {
		t.Errorf("Expected message to not contain 'Error output', got '%s'", err.Message)
	}
}

func TestProcessError_MessageOnly(t *testing.T) {
	err := NewProcessError("simple error", 0, "")

	if err.Message != "simple error" {
		t.Errorf("Expected message 'simple error', got '%s'", err.Message)
	}
}

func TestJSONDecodeError(t *testing.T) {
	originalErr := errors.New("unexpected character")
	err := NewJSONDecodeError(`{"invalid": json}`, originalErr)

	if !strings.Contains(err.Message, "Failed to decode JSON") {
		t.Errorf("Expected message to contain 'Failed to decode JSON', got '%s'", err.Message)
	}
	if err.Line != `{"invalid": json}` {
		t.Errorf("Expected Line to be the original line, got '%s'", err.Line)
	}
	if err.OriginalError != originalErr {
		t.Errorf("Expected OriginalError to be %v, got %v", originalErr, err.OriginalError)
	}
	if err.Cause != originalErr {
		t.Errorf("Expected Cause to be %v, got %v", originalErr, err.Cause)
	}
}

func TestJSONDecodeError_LongLine(t *testing.T) {
	longLine := strings.Repeat("x", 200)
	err := NewJSONDecodeError(longLine, nil)

	// Should truncate to ~103 chars (100 + "...")
	if !strings.Contains(err.Message, "...") {
		t.Error("Expected truncated line to contain '...'")
	}
	if len(err.Message) > 150 {
		t.Errorf("Expected truncated message, got length %d", len(err.Message))
	}

	// Original line should be preserved
	if err.Line != longLine {
		t.Error("Expected Line to preserve original line")
	}
}

func TestJSONDecodeError_ShortLine(t *testing.T) {
	shortLine := "short json"
	err := NewJSONDecodeError(shortLine, nil)

	if strings.Contains(err.Message, "...") {
		t.Error("Expected short line to not be truncated")
	}
	if !strings.Contains(err.Message, shortLine) {
		t.Errorf("Expected message to contain '%s', got '%s'", shortLine, err.Message)
	}
}

func TestMessageParseError(t *testing.T) {
	data := map[string]any{"type": "unknown", "field": "value"}
	err := NewMessageParseError("unknown message type", data)

	if err.Message != "unknown message type" {
		t.Errorf("Expected message 'unknown message type', got '%s'", err.Message)
	}
	if err.Data == nil {
		t.Error("Expected Data to be non-nil")
	}
	if err.Data["type"] != "unknown" {
		t.Errorf("Expected Data['type']='unknown', got '%v'", err.Data["type"])
	}
}

func TestMessageParseError_NilData(t *testing.T) {
	err := NewMessageParseError("parsing failed", nil)

	if err.Message != "parsing failed" {
		t.Errorf("Expected message 'parsing failed', got '%s'", err.Message)
	}
	if err.Data != nil {
		t.Error("Expected Data to be nil")
	}
}

func TestErrorHierarchy(t *testing.T) {
	// Verify error type hierarchy
	cliNotFound := NewCLINotFoundError("not found", "/path")
	cliConnection := NewCLIConnectionError("connection error")
	sdkError := NewClaudeSDKError("sdk error")

	// CLINotFoundError embeds CLIConnectionError
	var _ = &cliNotFound.CLIConnectionError

	// CLIConnectionError embeds ClaudeSDKError
	var _ = &cliConnection.ClaudeSDKError

	// All implement error interface
	var _ error = sdkError
	var _ error = cliConnection
	var _ error = cliNotFound
}

func TestErrorsIs(t *testing.T) {
	cause := errors.New("original")
	wrapped := WrapClaudeSDKError("wrapped", cause)

	// Test that errors.Is works with wrapped errors
	if !errors.Is(wrapped, cause) {
		t.Error("Expected errors.Is to find the cause")
	}
}

func TestErrorsAs(t *testing.T) {
	err := NewProcessError("process failed", 1, "stderr")

	// Test that errors.As works for ProcessError
	var processErr *ProcessError
	if !errors.As(err, &processErr) {
		t.Error("Expected errors.As to match ProcessError")
	}

	// Test CLIConnectionError chain
	connErr := NewCLIConnectionError("connection failed")
	var connTarget *CLIConnectionError
	if !errors.As(connErr, &connTarget) {
		t.Error("Expected errors.As to match CLIConnectionError")
	}

	// Test CLINotFoundError chain
	notFoundErr := NewCLINotFoundError("not found", "/usr/bin/claude")
	var notFoundTarget *CLINotFoundError
	if !errors.As(notFoundErr, &notFoundTarget) {
		t.Error("Expected errors.As to match CLINotFoundError")
	}
}

// Benchmark tests

func BenchmarkNewClaudeSDKError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewClaudeSDKError("benchmark error")
	}
}

func BenchmarkNewProcessError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewProcessError("process failed", 1, "stderr output here")
	}
}

func BenchmarkNewJSONDecodeError_LongLine(b *testing.B) {
	longLine := strings.Repeat("x", 200)
	err := errors.New("parse error")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewJSONDecodeError(longLine, err)
	}
}
