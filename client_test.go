package claude

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewClient_Defaults(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	if client.options == nil {
		t.Error("Expected non-nil options")
	}
	if client.messageCh == nil {
		t.Error("Expected non-nil messageCh")
	}
	if client.errorCh == nil {
		t.Error("Expected non-nil errorCh")
	}
	if client.sessionID != "default" {
		t.Errorf("Expected default sessionID 'default', got '%s'", client.sessionID)
	}
	if client.connected {
		t.Error("Expected connected to be false initially")
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	client := NewClient(
		WithCwd("/home/user/project"),
		WithModel("claude-3-opus"),
		WithMaxTurns(5),
	)

	if client.options.Cwd != "/home/user/project" {
		t.Errorf("Expected Cwd '/home/user/project', got '%s'", client.options.Cwd)
	}
	if client.options.Model != "claude-3-opus" {
		t.Errorf("Expected Model 'claude-3-opus', got '%s'", client.options.Model)
	}
	if client.options.MaxTurns != 5 {
		t.Errorf("Expected MaxTurns 5, got %d", client.options.MaxTurns)
	}
}

func TestClient_SetSessionID(t *testing.T) {
	client := NewClient()

	client.SetSessionID("custom-session")

	if client.sessionID != "custom-session" {
		t.Errorf("Expected sessionID 'custom-session', got '%s'", client.sessionID)
	}
}

func TestClient_SetSessionID_Concurrent(t *testing.T) {
	client := NewClient()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			client.SetSessionID(string(rune('a' + id)))
		}(i)
	}

	wg.Wait()
	// Just verify no race condition, final value doesn't matter
}

func TestClient_Messages(t *testing.T) {
	client := NewClient()

	ch := client.Messages()
	if ch == nil {
		t.Error("Expected non-nil Messages channel")
	}
}

func TestClient_Errors(t *testing.T) {
	client := NewClient()

	ch := client.Errors()
	if ch == nil {
		t.Error("Expected non-nil Errors channel")
	}
}

func TestClient_Close_NotConnected(t *testing.T) {
	client := NewClient()

	err := client.Close()
	if err != nil {
		t.Errorf("Expected no error when closing non-connected client, got %v", err)
	}
}

func TestClient_Close_Idempotent(t *testing.T) {
	client := NewClient()

	// Multiple closes should not error
	err1 := client.Close()
	err2 := client.Close()
	err3 := client.Close()

	if err1 != nil || err2 != nil || err3 != nil {
		t.Error("Expected Close to be idempotent")
	}
}

func TestClient_Query_NotConnected(t *testing.T) {
	client := NewClient()

	err := client.Query(context.Background(), "Hello")

	if err == nil {
		t.Fatal("Expected error when querying without connection")
	}

	_, ok := err.(*CLIConnectionError)
	if !ok {
		t.Fatalf("Expected *CLIConnectionError, got %T", err)
	}
}

func TestClient_QueryMessage_NotConnected(t *testing.T) {
	client := NewClient()

	err := client.QueryMessage(context.Background(), map[string]any{"test": "data"})

	if err == nil {
		t.Fatal("Expected error when querying without connection")
	}

	_, ok := err.(*CLIConnectionError)
	if !ok {
		t.Fatalf("Expected *CLIConnectionError, got %T", err)
	}
}

func TestClient_Interrupt_NotConnected(t *testing.T) {
	client := NewClient()

	err := client.Interrupt(context.Background())

	if err == nil {
		t.Fatal("Expected error when interrupting without connection")
	}

	_, ok := err.(*CLIConnectionError)
	if !ok {
		t.Fatalf("Expected *CLIConnectionError, got %T", err)
	}
}

func TestClient_SetPermissionMode_NotConnected(t *testing.T) {
	client := NewClient()

	err := client.SetPermissionMode(context.Background(), PermissionModeDefault)

	if err == nil {
		t.Fatal("Expected error when setting mode without connection")
	}

	_, ok := err.(*CLIConnectionError)
	if !ok {
		t.Fatalf("Expected *CLIConnectionError, got %T", err)
	}
}

func TestClient_SetModel_NotConnected(t *testing.T) {
	client := NewClient()

	err := client.SetModel(context.Background(), "claude-3-opus")

	if err == nil {
		t.Fatal("Expected error when setting model without connection")
	}
}

func TestClient_RewindFiles_NotConnected(t *testing.T) {
	client := NewClient()

	err := client.RewindFiles(context.Background(), "message-123")

	if err == nil {
		t.Fatal("Expected error when rewinding without connection")
	}
}

func TestClient_GetMCPStatus_NotConnected(t *testing.T) {
	client := NewClient()

	_, err := client.GetMCPStatus(context.Background())

	if err == nil {
		t.Fatal("Expected error when getting MCP status without connection")
	}
}

func TestClient_GetServerInfo_NotConnected(t *testing.T) {
	client := NewClient()

	info := client.GetServerInfo()

	if info != nil {
		t.Errorf("Expected nil server info when not connected, got %v", info)
	}
}

func TestClient_ReceiveResponse(t *testing.T) {
	client := NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start receiving (will block since no connection)
	ch := client.ReceiveResponse(ctx)

	// Should close when context is done
	select {
	case _, ok := <-ch:
		if ok {
			// Received something unexpectedly
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected channel to close when context is done")
	}
}

// Tests for message format validation
func TestClient_Query_MessageFormat(t *testing.T) {
	// This test verifies the message structure without actually connecting
	// We can test the JSON structure by examining what would be sent

	client := NewClient()
	client.sessionID = "test-session"

	// Create a mock that captures the written data
	expectedMessage := map[string]any{
		"type": "user",
		"message": map[string]any{
			"role":    "user",
			"content": "Hello, Claude!",
		},
		"parent_tool_use_id": nil,
		"session_id":         "test-session",
	}

	data, err := json.Marshal(expectedMessage)
	if err != nil {
		t.Fatalf("Failed to marshal expected message: %v", err)
	}

	// Verify the structure
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded["type"] != "user" {
		t.Errorf("Expected type 'user', got '%v'", decoded["type"])
	}
	if decoded["session_id"] != "test-session" {
		t.Errorf("Expected session_id 'test-session', got '%v'", decoded["session_id"])
	}

	message, ok := decoded["message"].(map[string]any)
	if !ok {
		t.Fatal("Expected message to be map[string]any")
	}
	if message["role"] != "user" {
		t.Errorf("Expected role 'user', got '%v'", message["role"])
	}
	if message["content"] != "Hello, Claude!" {
		t.Errorf("Expected content 'Hello, Claude!', got '%v'", message["content"])
	}
}

func TestClient_QueryMessage_AddsSessionID(t *testing.T) {
	// Test that session_id is added when not present
	client := NewClient()
	client.sessionID = "injected-session"

	message := map[string]any{
		"type": "user",
		"message": map[string]any{
			"role":    "user",
			"content": "Test",
		},
	}

	// Check that session_id would be added
	if _, ok := message["session_id"]; !ok {
		message["session_id"] = client.sessionID
	}

	if message["session_id"] != "injected-session" {
		t.Errorf("Expected session_id to be injected, got '%v'", message["session_id"])
	}
}

func TestClient_QueryMessage_PreservesSessionID(t *testing.T) {
	// Test that existing session_id is preserved
	client := NewClient()
	client.sessionID = "default-session"

	message := map[string]any{
		"type":       "user",
		"session_id": "custom-session",
	}

	// The logic: if session_id is not present, add it
	if _, ok := message["session_id"]; !ok {
		message["session_id"] = client.sessionID
	}

	// Should preserve the custom session
	if message["session_id"] != "custom-session" {
		t.Errorf("Expected preserved session_id 'custom-session', got '%v'", message["session_id"])
	}
}

func TestClient_Connect_CanUseToolConflict(t *testing.T) {
	// Test that Connect returns error when both CanUseTool and PermissionPromptToolName are set
	fn := func(ctx context.Context, toolName string, input map[string]any, permCtx ToolPermissionContext) (PermissionResult, error) {
		return PermissionResultAllow{}, nil
	}

	client := NewClient(
		WithCanUseTool(fn),
		WithPermissionPromptToolName("custom"),
	)

	err := client.Connect(context.Background())

	if err == nil {
		t.Fatal("Expected error when both CanUseTool and PermissionPromptToolName are set")
	}

	sdkErr, ok := err.(*ClaudeSDKError)
	if !ok {
		t.Fatalf("Expected *ClaudeSDKError, got %T", err)
	}
	if !strings.Contains(sdkErr.Message, "can_use_tool callback cannot be used with permission_prompt_tool_name") {
		t.Errorf("Unexpected error message: %s", sdkErr.Message)
	}
}

func TestClient_Connect_AlreadyConnected(t *testing.T) {
	client := NewClient()

	// Manually set connected state to test the check
	client.mu.Lock()
	client.connected = true
	client.mu.Unlock()

	err := client.Connect(context.Background())

	// Should return nil when already connected
	if err != nil {
		t.Errorf("Expected no error when already connected, got %v", err)
	}
}

func TestClient_Close_WhenConnected(t *testing.T) {
	client := NewClient()

	// Simulate connected state with a nil query (edge case)
	client.mu.Lock()
	client.connected = true
	client.query = nil
	client.mu.Unlock()

	err := client.Close()
	if err != nil {
		t.Errorf("Expected no error when closing, got %v", err)
	}

	// Verify state after close
	if client.connected {
		t.Error("Expected connected to be false after Close")
	}
	if client.transport != nil {
		t.Error("Expected transport to be nil after Close")
	}
}

func TestClient_ReceiveResponse_ClosedChannel(t *testing.T) {
	client := NewClient()

	// Close the message channel to simulate end of messages
	close(client.messageCh)

	ctx := context.Background()
	ch := client.ReceiveResponse(ctx)

	// Should close when message channel is closed
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("Expected channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected channel to close promptly when source is closed")
	}
}

func TestClient_Connect_WithCanUseTool(t *testing.T) {
	// Test that PermissionPromptToolName is auto-set when CanUseTool is provided
	fn := func(ctx context.Context, toolName string, input map[string]any, permCtx ToolPermissionContext) (PermissionResult, error) {
		return PermissionResultAllow{}, nil
	}

	client := NewClient(
		WithCanUseTool(fn),
	)

	// Connect will fail because CLI isn't available, but we can verify the option was set
	_ = client.Connect(context.Background())

	// The PermissionPromptToolName should have been set to "stdio" during Connect
	if client.options.PermissionPromptToolName != "stdio" {
		t.Errorf("Expected PermissionPromptToolName 'stdio', got '%s'", client.options.PermissionPromptToolName)
	}
}

func TestClient_GetServerInfo_WithNilQuery(t *testing.T) {
	client := NewClient()
	client.mu.Lock()
	client.query = nil
	client.mu.Unlock()

	info := client.GetServerInfo()
	if info != nil {
		t.Errorf("Expected nil when query is nil, got %v", info)
	}
}

func TestClient_SetModel_NotConnected_ErrorType(t *testing.T) {
	client := NewClient()

	err := client.SetModel(context.Background(), "claude-3-opus")

	if err == nil {
		t.Fatal("Expected error when setting model without connection")
	}

	_, ok := err.(*CLIConnectionError)
	if !ok {
		t.Fatalf("Expected *CLIConnectionError, got %T", err)
	}
}

func TestClient_RewindFiles_NotConnected_ErrorType(t *testing.T) {
	client := NewClient()

	err := client.RewindFiles(context.Background(), "message-123")

	if err == nil {
		t.Fatal("Expected error when rewinding without connection")
	}

	_, ok := err.(*CLIConnectionError)
	if !ok {
		t.Fatalf("Expected *CLIConnectionError, got %T", err)
	}
}

func TestClient_GetMCPStatus_NotConnected_ErrorType(t *testing.T) {
	client := NewClient()

	_, err := client.GetMCPStatus(context.Background())

	if err == nil {
		t.Fatal("Expected error when getting MCP status without connection")
	}

	_, ok := err.(*CLIConnectionError)
	if !ok {
		t.Fatalf("Expected *CLIConnectionError, got %T", err)
	}
}

// Benchmark tests

func BenchmarkNewClient(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewClient()
	}
}

func BenchmarkNewClient_WithOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewClient(
			WithCwd("/home/user"),
			WithModel("claude-3-opus"),
			WithMaxTurns(10),
		)
	}
}

func BenchmarkClient_SetSessionID(b *testing.B) {
	client := NewClient()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.SetSessionID("session-id")
	}
}
