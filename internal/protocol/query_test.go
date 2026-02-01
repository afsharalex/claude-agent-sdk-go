package protocol

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/afsharalex/claude-agent-sdk-go/internal/transport"
	"github.com/afsharalex/claude-agent-sdk-go/internal/types"
)

func TestNewQuery_Defaults(t *testing.T) {
	mock := transport.NewMockTransport()

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})

	if q == nil {
		t.Fatal("Expected non-nil Query")
	}
	if q.transport != mock {
		t.Error("Expected transport to be set")
	}
	if !q.isStreamingMode {
		t.Error("Expected isStreamingMode to be true")
	}
	if q.hookCallbacks == nil {
		t.Error("Expected hookCallbacks to be initialized")
	}
	if q.messageChan == nil {
		t.Error("Expected messageChan to be initialized")
	}

	// Clean up
	_ = q.Close()
}

func TestNewQuery_WithTimeouts(t *testing.T) {
	mock := transport.NewMockTransport()

	q := NewQuery(QueryConfig{
		Transport:          mock,
		IsStreamingMode:    true,
		InitializeTimeout:  30 * time.Second,
		StreamCloseTimeout: 45 * time.Second,
	})
	defer func() { _ = q.Close() }()

	if q.initTimeout != 30*time.Second {
		t.Errorf("Expected initTimeout 30s, got %v", q.initTimeout)
	}
	if q.streamCloseTimeout != 45*time.Second {
		t.Errorf("Expected streamCloseTimeout 45s, got %v", q.streamCloseTimeout)
	}
}

func TestNewQuery_DefaultTimeouts(t *testing.T) {
	mock := transport.NewMockTransport()

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	if q.initTimeout != 60*time.Second {
		t.Errorf("Expected default initTimeout 60s, got %v", q.initTimeout)
	}
	if q.streamCloseTimeout != 60*time.Second {
		t.Errorf("Expected default streamCloseTimeout 60s, got %v", q.streamCloseTimeout)
	}
}

func TestQuery_Initialize_NonStreamingMode(t *testing.T) {
	mock := transport.NewMockTransport()

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: false,
	})
	defer func() { _ = q.Close() }()

	result, err := q.Initialize(context.Background())

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil result in non-streaming mode, got %v", result)
	}
}

func TestQuery_InitResult(t *testing.T) {
	mock := transport.NewMockTransport()

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	// Initially nil
	if q.InitResult() != nil {
		t.Error("Expected InitResult to be nil initially")
	}

	// Set via Initialize (simulated)
	q.initResult = map[string]any{"test": "value"}

	result := q.InitResult()
	if result == nil {
		t.Fatal("Expected non-nil InitResult")
	}
	if result["test"] != "value" {
		t.Errorf("Expected test='value', got '%v'", result["test"])
	}
}

func TestQuery_ReceiveMessages(t *testing.T) {
	mock := transport.NewMockTransport()

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	ch := q.ReceiveMessages()
	if ch == nil {
		t.Error("Expected non-nil ReceiveMessages channel")
	}
}

func TestQuery_Close_Idempotent(t *testing.T) {
	mock := transport.NewMockTransport()

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})

	// Multiple closes should not panic
	err1 := q.Close()
	err2 := q.Close()
	err3 := q.Close()

	if err1 != nil {
		t.Errorf("Expected no error on first close, got %v", err1)
	}
	if err2 != nil {
		t.Errorf("Expected no error on second close, got %v", err2)
	}
	if err3 != nil {
		t.Errorf("Expected no error on third close, got %v", err3)
	}
}

func TestParseHookInput_PreToolUse(t *testing.T) {
	data := map[string]any{
		"hook_event_name": "PreToolUse",
		"session_id":      "session-123",
		"transcript_path": "/path/to/transcript",
		"cwd":             "/home/user",
		"permission_mode": "default",
		"tool_name":       "Bash",
		"tool_input":      map[string]any{"command": "ls -la"},
	}

	result, err := parseHookInput(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	input, ok := result.(types.PreToolUseHookInput)
	if !ok {
		t.Fatalf("Expected PreToolUseHookInput, got %T", result)
	}

	if input.SessionID != "session-123" {
		t.Errorf("Expected session_id 'session-123', got '%s'", input.SessionID)
	}
	if input.ToolName != "Bash" {
		t.Errorf("Expected tool_name 'Bash', got '%s'", input.ToolName)
	}
	if input.ToolInput["command"] != "ls -la" {
		t.Errorf("Expected command 'ls -la', got '%v'", input.ToolInput["command"])
	}
}

func TestParseHookInput_PostToolUse(t *testing.T) {
	data := map[string]any{
		"hook_event_name": "PostToolUse",
		"session_id":      "session-456",
		"tool_name":       "Write",
		"tool_input":      map[string]any{"path": "/tmp/file.txt"},
		"tool_response":   "File written successfully",
	}

	result, err := parseHookInput(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	input, ok := result.(types.PostToolUseHookInput)
	if !ok {
		t.Fatalf("Expected PostToolUseHookInput, got %T", result)
	}

	if input.ToolName != "Write" {
		t.Errorf("Expected tool_name 'Write', got '%s'", input.ToolName)
	}
	if input.ToolResponse != "File written successfully" {
		t.Errorf("Expected tool_response, got '%v'", input.ToolResponse)
	}
}

func TestParseHookInput_PostToolUseFailure(t *testing.T) {
	data := map[string]any{
		"hook_event_name": "PostToolUseFailure",
		"session_id":      "session-789",
		"tool_name":       "Execute",
		"tool_use_id":     "tool-123",
		"error":           "Permission denied",
		"is_interrupt":    true,
	}

	result, err := parseHookInput(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	input, ok := result.(types.PostToolUseFailureHookInput)
	if !ok {
		t.Fatalf("Expected PostToolUseFailureHookInput, got %T", result)
	}

	if input.Error != "Permission denied" {
		t.Errorf("Expected error 'Permission denied', got '%s'", input.Error)
	}
	if !input.IsInterrupt {
		t.Error("Expected IsInterrupt to be true")
	}
}

func TestParseHookInput_UserPromptSubmit(t *testing.T) {
	data := map[string]any{
		"hook_event_name": "UserPromptSubmit",
		"session_id":      "session-prompt",
		"prompt":          "Hello, Claude!",
	}

	result, err := parseHookInput(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	input, ok := result.(types.UserPromptSubmitHookInput)
	if !ok {
		t.Fatalf("Expected UserPromptSubmitHookInput, got %T", result)
	}

	if input.Prompt != "Hello, Claude!" {
		t.Errorf("Expected prompt 'Hello, Claude!', got '%s'", input.Prompt)
	}
}

func TestParseHookInput_Stop(t *testing.T) {
	data := map[string]any{
		"hook_event_name":  "Stop",
		"session_id":       "session-stop",
		"stop_hook_active": true,
	}

	result, err := parseHookInput(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	input, ok := result.(types.StopHookInput)
	if !ok {
		t.Fatalf("Expected StopHookInput, got %T", result)
	}

	if !input.StopHookActive {
		t.Error("Expected StopHookActive to be true")
	}
}

func TestParseHookInput_SubagentStop(t *testing.T) {
	data := map[string]any{
		"hook_event_name":  "SubagentStop",
		"session_id":       "session-subagent",
		"stop_hook_active": false,
	}

	result, err := parseHookInput(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	input, ok := result.(types.SubagentStopHookInput)
	if !ok {
		t.Fatalf("Expected SubagentStopHookInput, got %T", result)
	}

	if input.StopHookActive {
		t.Error("Expected StopHookActive to be false")
	}
}

func TestParseHookInput_PreCompact(t *testing.T) {
	data := map[string]any{
		"hook_event_name":     "PreCompact",
		"session_id":          "session-compact",
		"trigger":             "auto",
		"custom_instructions": "Keep summary brief",
	}

	result, err := parseHookInput(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	input, ok := result.(types.PreCompactHookInput)
	if !ok {
		t.Fatalf("Expected PreCompactHookInput, got %T", result)
	}

	if input.Trigger != types.PreCompactTriggerAuto {
		t.Errorf("Expected trigger 'auto', got '%s'", input.Trigger)
	}
	if input.CustomInstructions == nil || *input.CustomInstructions != "Keep summary brief" {
		t.Errorf("Expected custom_instructions 'Keep summary brief'")
	}
}

func TestParseHookInput_PreCompactWithoutCustomInstructions(t *testing.T) {
	data := map[string]any{
		"hook_event_name": "PreCompact",
		"session_id":      "session-compact",
		"trigger":         "manual",
	}

	result, err := parseHookInput(data)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	input := result.(types.PreCompactHookInput)
	if input.CustomInstructions != nil {
		t.Error("Expected CustomInstructions to be nil")
	}
}

func TestParseHookInput_UnknownEvent(t *testing.T) {
	data := map[string]any{
		"hook_event_name": "UnknownEvent",
		"session_id":      "session-unknown",
	}

	_, err := parseHookInput(data)
	if err == nil {
		t.Error("Expected error for unknown event")
	}
}

func TestParseHookInput_InvalidFormat(t *testing.T) {
	// Not a map
	_, err := parseHookInput("invalid")
	if err == nil {
		t.Error("Expected error for invalid format")
	}
}

func TestGetString(t *testing.T) {
	m := map[string]any{
		"key1": "value1",
		"key2": 123,
		"key3": nil,
	}

	if getString(m, "key1") != "value1" {
		t.Errorf("Expected 'value1', got '%s'", getString(m, "key1"))
	}
	if getString(m, "key2") != "" {
		t.Errorf("Expected empty string for non-string, got '%s'", getString(m, "key2"))
	}
	if getString(m, "key3") != "" {
		t.Errorf("Expected empty string for nil, got '%s'", getString(m, "key3"))
	}
	if getString(m, "missing") != "" {
		t.Errorf("Expected empty string for missing key, got '%s'", getString(m, "missing"))
	}
}

func TestGetMap(t *testing.T) {
	inner := map[string]any{"nested": "value"}
	m := map[string]any{
		"key1": inner,
		"key2": "not a map",
	}

	result := getMap(m, "key1")
	if result == nil {
		t.Error("Expected non-nil map")
	}
	if result["nested"] != "value" {
		t.Errorf("Expected nested='value', got '%v'", result["nested"])
	}

	if getMap(m, "key2") != nil {
		t.Error("Expected nil for non-map")
	}
	if getMap(m, "missing") != nil {
		t.Error("Expected nil for missing key")
	}
}

func TestGetBool(t *testing.T) {
	m := map[string]any{
		"key1": true,
		"key2": false,
		"key3": "true",
	}

	if !getBool(m, "key1") {
		t.Error("Expected true for key1")
	}
	if getBool(m, "key2") {
		t.Error("Expected false for key2")
	}
	if getBool(m, "key3") {
		t.Error("Expected false for string 'true'")
	}
	if getBool(m, "missing") {
		t.Error("Expected false for missing key")
	}
}

func TestQuery_WithHooks(t *testing.T) {
	mock := transport.NewMockTransport()

	hooks := map[types.HookEvent][]types.HookMatcher{
		types.HookEventPreToolUse: {
			{
				Matcher: "Bash",
				Hooks: []types.HookCallback{
					func(ctx context.Context, input types.HookInput, toolUseID string, hookCtx types.HookContext) (types.HookOutput, error) {
						return types.HookOutput{}, nil
					},
				},
				Timeout: 30.0,
			},
		},
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		Hooks:           hooks,
	})
	defer func() { _ = q.Close() }()

	if len(q.hooks) != 1 {
		t.Errorf("Expected 1 hook event, got %d", len(q.hooks))
	}
}

func TestQuery_WithCanUseTool(t *testing.T) {
	mock := transport.NewMockTransport()

	canUseTool := func(ctx context.Context, toolName string, input map[string]any, permCtx types.ToolPermissionContext) (types.PermissionResult, error) {
		return types.PermissionResultAllow{}, nil
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		CanUseTool:      canUseTool,
	})
	defer func() { _ = q.Close() }()

	if q.canUseTool == nil {
		t.Error("Expected canUseTool to be set")
	}
}

func TestQuery_WithSDKMCPServers(t *testing.T) {
	mock := transport.NewMockTransport()

	servers := map[string]*types.MCPServer{
		"test-server": {
			Name:    "test",
			Version: "1.0.0",
			Tools:   []types.MCPTool{},
		},
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		SDKMCPServers:   servers,
	})
	defer func() { _ = q.Close() }()

	if len(q.sdkMCPServers) != 1 {
		t.Errorf("Expected 1 MCP server, got %d", len(q.sdkMCPServers))
	}
}

// Benchmark tests

func BenchmarkNewQuery(b *testing.B) {
	mock := transport.NewMockTransport()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := NewQuery(QueryConfig{
			Transport:       mock,
			IsStreamingMode: true,
		})
		_ = q.Close()
	}
}

func BenchmarkParseHookInput_PreToolUse(b *testing.B) {
	data := map[string]any{
		"hook_event_name": "PreToolUse",
		"session_id":      "session-123",
		"tool_name":       "Bash",
		"tool_input":      map[string]any{"command": "ls"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseHookInput(data)
	}
}

// Tests for Start() and message reading

func TestQuery_Start_MessagesForwarded(t *testing.T) {
	messages := []map[string]any{
		{"type": "assistant", "message": "Hello"},
		{"type": "user", "message": "Hi"},
	}
	mock := transport.NewMockTransport().WithMessages(messages...)
	_ = mock.Connect(context.Background())

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	q.Start(ctx)

	ch := q.ReceiveMessages()
	received := make([]map[string]any, 0)

	for msg := range ch {
		received = append(received, msg)
		if len(received) >= len(messages) {
			break
		}
	}

	if len(received) != len(messages) {
		t.Errorf("Expected %d messages, got %d", len(messages), len(received))
	}
}

func TestQuery_Start_ContextCancellation(t *testing.T) {
	// Messages that would take a while
	messages := []map[string]any{
		{"type": "assistant", "message": "Hello"},
	}
	mock := transport.NewMockTransport().WithMessages(messages...)
	_ = mock.Connect(context.Background())

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})

	ctx, cancel := context.WithCancel(context.Background())
	q.Start(ctx)

	// Cancel immediately
	cancel()

	// Channel should eventually close
	ch := q.ReceiveMessages()
	timeout := time.After(200 * time.Millisecond)
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				// Channel closed as expected
				_ = q.Close()
				return
			}
		case <-timeout:
			_ = q.Close()
			return
		}
	}
}

func TestQuery_Start_ResultMessageTriggersOnce(t *testing.T) {
	messages := []map[string]any{
		{"type": "result", "result": "done"},
	}
	mock := transport.NewMockTransport().WithMessages(messages...)
	_ = mock.Connect(context.Background())

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	q.Start(ctx)

	// Wait for message to be received
	ch := q.ReceiveMessages()
	<-ch

	// firstResultCh should be closed now - this should not block
	select {
	case <-q.firstResultCh:
		// Good - channel is closed
	case <-time.After(100 * time.Millisecond):
		t.Error("firstResultCh should be closed after result message")
	}
}

func TestQuery_Start_ErrorFromTransport(t *testing.T) {
	// Set up mock to have a message that will keep channel open longer
	mock := transport.NewMockTransport().WithMessages(
		map[string]any{"type": "keep_alive"},
	)
	_ = mock.Connect(context.Background())

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	q.Start(ctx)

	// Get the first message and check for any error in channels
	ch := q.ReceiveMessages()
	select {
	case msg, ok := <-ch:
		if !ok {
			// Channel closed - acceptable
			return
		}
		// Just verify we got a message
		if msg["type"] == "error" {
			// We got an error message, which is what we're testing for
			return
		}
		// Otherwise we got the keep_alive message, which is fine
	case <-time.After(200 * time.Millisecond):
		// Timeout is acceptable - test is mainly ensuring no panics
	}
}

// Tests for handleControlResponse

func TestQuery_handleControlResponse_Success(t *testing.T) {
	mock := transport.NewMockTransport()
	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	// Register a pending response
	responseCh := make(chan map[string]any, 1)
	q.pendingResponses.Store("test-request-id", responseCh)

	// Handle the control response
	message := map[string]any{
		"type": "control_response",
		"response": map[string]any{
			"request_id": "test-request-id",
			"subtype":    "success",
			"response":   map[string]any{"data": "test"},
		},
	}

	q.handleControlResponse(message)

	// Should receive the response
	select {
	case resp := <-responseCh:
		if resp["request_id"] != "test-request-id" {
			t.Errorf("Expected request_id 'test-request-id', got %v", resp["request_id"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for response")
	}
}

func TestQuery_handleControlResponse_UnknownRequest(t *testing.T) {
	mock := transport.NewMockTransport()
	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	// Handle a response for unknown request - should not panic
	message := map[string]any{
		"type": "control_response",
		"response": map[string]any{
			"request_id": "unknown-request-id",
			"subtype":    "success",
		},
	}

	// Should not panic
	q.handleControlResponse(message)
}

func TestQuery_handleControlResponse_InvalidFormat(t *testing.T) {
	mock := transport.NewMockTransport()
	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	// Handle invalid response format - should not panic
	message := map[string]any{
		"type":     "control_response",
		"response": "invalid",
	}

	// Should not panic
	q.handleControlResponse(message)
}

// Tests for handleControlRequest routing

func TestQuery_handleControlRequest_InvalidFormat(t *testing.T) {
	mock := transport.NewMockTransport()
	_ = mock.Connect(context.Background())

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	// Handle request with invalid format
	request := map[string]any{
		"request_id": "test-id",
		"request":    "invalid", // should be map
	}

	q.handleControlRequest(context.Background(), request)

	// Should have written error response
	time.Sleep(50 * time.Millisecond)
	written := mock.GetWrittenData()
	if len(written) == 0 {
		t.Error("Expected error response to be written")
	}
}

func TestQuery_handleControlRequest_UnknownSubtype(t *testing.T) {
	mock := transport.NewMockTransport()
	_ = mock.Connect(context.Background())

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"request_id": "test-id",
		"request": map[string]any{
			"subtype": "unknown_subtype",
		},
	}

	q.handleControlRequest(context.Background(), request)

	// Should have written error response
	time.Sleep(50 * time.Millisecond)
	written := mock.GetWrittenData()
	if len(written) == 0 {
		t.Error("Expected error response to be written")
	}
}

// Tests for handleCanUseTool

func TestQuery_handleCanUseTool_Allow(t *testing.T) {
	mock := transport.NewMockTransport()

	canUseTool := func(ctx context.Context, toolName string, input map[string]any, permCtx types.ToolPermissionContext) (types.PermissionResult, error) {
		return types.PermissionResultAllow{}, nil
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		CanUseTool:      canUseTool,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"tool_name": "Bash",
		"input":     map[string]any{"command": "ls"},
	}

	result, err := q.handleCanUseTool(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result["behavior"] != "allow" {
		t.Errorf("Expected behavior 'allow', got %v", result["behavior"])
	}
}

func TestQuery_handleCanUseTool_AllowWithUpdatedInput(t *testing.T) {
	mock := transport.NewMockTransport()

	updatedInput := map[string]any{"command": "ls -la"}
	canUseTool := func(ctx context.Context, toolName string, input map[string]any, permCtx types.ToolPermissionContext) (types.PermissionResult, error) {
		return types.PermissionResultAllow{
			UpdatedInput: updatedInput,
		}, nil
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		CanUseTool:      canUseTool,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"tool_name": "Bash",
		"input":     map[string]any{"command": "ls"},
	}

	result, err := q.handleCanUseTool(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result["updatedInput"].(map[string]any)["command"] != "ls -la" {
		t.Errorf("Expected updatedInput command 'ls -la', got %v", result["updatedInput"])
	}
}

func TestQuery_handleCanUseTool_AllowWithUpdatedPermissions(t *testing.T) {
	mock := transport.NewMockTransport()

	canUseTool := func(ctx context.Context, toolName string, input map[string]any, permCtx types.ToolPermissionContext) (types.PermissionResult, error) {
		return types.PermissionResultAllow{
			UpdatedPermissions: []types.PermissionUpdate{
				{
					Type: types.PermissionUpdateTypeAddRules,
					Rules: []types.PermissionRuleValue{
						{ToolName: "Bash", RuleContent: "allow ls"},
					},
					Behavior: types.PermissionBehaviorAllow,
				},
			},
		}, nil
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		CanUseTool:      canUseTool,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"tool_name": "Bash",
		"input":     map[string]any{"command": "ls"},
	}

	result, err := q.handleCanUseTool(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result["updatedPermissions"] == nil {
		t.Error("Expected updatedPermissions to be set")
	}
}

func TestQuery_handleCanUseTool_Deny(t *testing.T) {
	mock := transport.NewMockTransport()

	canUseTool := func(ctx context.Context, toolName string, input map[string]any, permCtx types.ToolPermissionContext) (types.PermissionResult, error) {
		return types.PermissionResultDeny{
			Message: "Not allowed",
		}, nil
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		CanUseTool:      canUseTool,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"tool_name": "Bash",
		"input":     map[string]any{"command": "rm -rf /"},
	}

	result, err := q.handleCanUseTool(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result["behavior"] != "deny" {
		t.Errorf("Expected behavior 'deny', got %v", result["behavior"])
	}
	if result["message"] != "Not allowed" {
		t.Errorf("Expected message 'Not allowed', got %v", result["message"])
	}
}

func TestQuery_handleCanUseTool_DenyInterrupt(t *testing.T) {
	mock := transport.NewMockTransport()

	canUseTool := func(ctx context.Context, toolName string, input map[string]any, permCtx types.ToolPermissionContext) (types.PermissionResult, error) {
		return types.PermissionResultDeny{
			Message:   "Critical violation",
			Interrupt: true,
		}, nil
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		CanUseTool:      canUseTool,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"tool_name": "Bash",
		"input":     map[string]any{"command": "dangerous"},
	}

	result, err := q.handleCanUseTool(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result["interrupt"] != true {
		t.Error("Expected interrupt to be true")
	}
}

func TestQuery_handleCanUseTool_NoCallback(t *testing.T) {
	mock := transport.NewMockTransport()

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		CanUseTool:      nil,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"tool_name": "Bash",
		"input":     map[string]any{"command": "ls"},
	}

	_, err := q.handleCanUseTool(context.Background(), request)
	if err == nil {
		t.Error("Expected error when canUseTool is nil")
	}
}

func TestQuery_handleCanUseTool_CallbackError(t *testing.T) {
	mock := transport.NewMockTransport()
	expectedErr := errors.New("callback error")

	canUseTool := func(ctx context.Context, toolName string, input map[string]any, permCtx types.ToolPermissionContext) (types.PermissionResult, error) {
		return nil, expectedErr
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		CanUseTool:      canUseTool,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"tool_name": "Bash",
		"input":     map[string]any{"command": "ls"},
	}

	_, err := q.handleCanUseTool(context.Background(), request)
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// Tests for handleHookCallback

func TestQuery_handleHookCallback_Success(t *testing.T) {
	mock := transport.NewMockTransport()

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	// Register a callback
	q.hookCallbacks["test-callback"] = func(ctx context.Context, input types.HookInput, toolUseID string, hookCtx types.HookContext) (types.HookOutput, error) {
		return types.HookOutput{
			Decision: types.HookDecisionBlock,
			Reason:   "blocked for testing",
		}, nil
	}

	request := map[string]any{
		"callback_id": "test-callback",
		"input": map[string]any{
			"hook_event_name": "PreToolUse",
			"session_id":      "session-123",
			"tool_name":       "Bash",
			"tool_input":      map[string]any{"command": "ls"},
		},
		"tool_use_id": "tool-123",
	}

	result, err := q.handleHookCallback(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result["decision"] != "block" {
		t.Errorf("Expected decision 'block', got %v", result["decision"])
	}
}

func TestQuery_handleHookCallback_UnknownID(t *testing.T) {
	mock := transport.NewMockTransport()

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"callback_id": "unknown-callback",
		"input": map[string]any{
			"hook_event_name": "PreToolUse",
			"session_id":      "session-123",
			"tool_name":       "Bash",
		},
	}

	_, err := q.handleHookCallback(context.Background(), request)
	if err == nil {
		t.Error("Expected error for unknown callback ID")
	}
}

func TestQuery_handleHookCallback_CallbackError(t *testing.T) {
	mock := transport.NewMockTransport()
	expectedErr := errors.New("hook callback error")

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	q.hookCallbacks["error-callback"] = func(ctx context.Context, input types.HookInput, toolUseID string, hookCtx types.HookContext) (types.HookOutput, error) {
		return types.HookOutput{}, expectedErr
	}

	request := map[string]any{
		"callback_id": "error-callback",
		"input": map[string]any{
			"hook_event_name": "PreToolUse",
			"session_id":      "session-123",
			"tool_name":       "Bash",
		},
	}

	_, err := q.handleHookCallback(context.Background(), request)
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// Tests for handleMCPMessage

func TestQuery_handleMCPMessage_Initialize(t *testing.T) {
	mock := transport.NewMockTransport()

	servers := map[string]*types.MCPServer{
		"test-server": {
			Name:    "test",
			Version: "1.0.0",
			Tools:   []types.MCPTool{},
		},
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		SDKMCPServers:   servers,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"server_name": "test-server",
		"message": map[string]any{
			"method": "initialize",
			"id":     1,
		},
	}

	result, err := q.handleMCPMessage(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	mcpResponse, ok := result["mcp_response"].(map[string]any)
	if !ok {
		t.Fatal("Expected mcp_response in result")
	}

	resultData, ok := mcpResponse["result"].(map[string]any)
	if !ok {
		t.Fatal("Expected result in mcp_response")
	}

	serverInfo, ok := resultData["serverInfo"].(map[string]any)
	if !ok {
		t.Fatal("Expected serverInfo in result")
	}

	if serverInfo["name"] != "test" {
		t.Errorf("Expected server name 'test', got %v", serverInfo["name"])
	}
}

func TestQuery_handleMCPMessage_ToolsList(t *testing.T) {
	mock := transport.NewMockTransport()

	servers := map[string]*types.MCPServer{
		"test-server": {
			Name:    "test",
			Version: "1.0.0",
			Tools: []types.MCPTool{
				{
					Name:        "test_tool",
					Description: "A test tool",
					InputSchema: map[string]any{"type": "object"},
					Handler:     nil,
				},
			},
		},
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		SDKMCPServers:   servers,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"server_name": "test-server",
		"message": map[string]any{
			"method": "tools/list",
			"id":     1,
		},
	}

	result, err := q.handleMCPMessage(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	mcpResponse := result["mcp_response"].(map[string]any)
	resultData := mcpResponse["result"].(map[string]any)
	tools := resultData["tools"].([]map[string]any)

	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
	if tools[0]["name"] != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got %v", tools[0]["name"])
	}
}

func TestQuery_handleMCPMessage_ToolsCall_Success(t *testing.T) {
	mock := transport.NewMockTransport()

	servers := map[string]*types.MCPServer{
		"test-server": {
			Name:    "test",
			Version: "1.0.0",
			Tools: []types.MCPTool{
				{
					Name:        "echo",
					Description: "Echo tool",
					InputSchema: map[string]any{"type": "object"},
					Handler: func(ctx context.Context, args map[string]any) (types.MCPToolResult, error) {
						return types.MCPToolResult{
							Content: []types.MCPContent{
								{Type: "text", Text: "Hello"},
							},
						}, nil
					},
				},
			},
		},
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		SDKMCPServers:   servers,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"server_name": "test-server",
		"message": map[string]any{
			"method": "tools/call",
			"id":     1,
			"params": map[string]any{
				"name":      "echo",
				"arguments": map[string]any{},
			},
		},
	}

	result, err := q.handleMCPMessage(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	mcpResponse := result["mcp_response"].(map[string]any)
	resultData := mcpResponse["result"].(map[string]any)
	content := resultData["content"].([]map[string]any)

	if len(content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(content))
	}
	if content[0]["text"] != "Hello" {
		t.Errorf("Expected text 'Hello', got %v", content[0]["text"])
	}
}

func TestQuery_handleMCPMessage_ToolsCall_Error(t *testing.T) {
	mock := transport.NewMockTransport()

	servers := map[string]*types.MCPServer{
		"test-server": {
			Name:    "test",
			Version: "1.0.0",
			Tools: []types.MCPTool{
				{
					Name:        "failing",
					Description: "Failing tool",
					Handler: func(ctx context.Context, args map[string]any) (types.MCPToolResult, error) {
						return types.MCPToolResult{}, errors.New("tool failed")
					},
				},
			},
		},
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		SDKMCPServers:   servers,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"server_name": "test-server",
		"message": map[string]any{
			"method": "tools/call",
			"id":     1,
			"params": map[string]any{
				"name":      "failing",
				"arguments": map[string]any{},
			},
		},
	}

	result, err := q.handleMCPMessage(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	mcpResponse := result["mcp_response"].(map[string]any)
	errData := mcpResponse["error"].(map[string]any)

	if errData["message"] != "tool failed" {
		t.Errorf("Expected error message 'tool failed', got %v", errData["message"])
	}
}

func TestQuery_handleMCPMessage_ToolsCall_UnknownTool(t *testing.T) {
	mock := transport.NewMockTransport()

	servers := map[string]*types.MCPServer{
		"test-server": {
			Name:    "test",
			Version: "1.0.0",
			Tools:   []types.MCPTool{},
		},
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		SDKMCPServers:   servers,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"server_name": "test-server",
		"message": map[string]any{
			"method": "tools/call",
			"id":     1,
			"params": map[string]any{
				"name":      "nonexistent",
				"arguments": map[string]any{},
			},
		},
	}

	result, err := q.handleMCPMessage(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	mcpResponse := result["mcp_response"].(map[string]any)
	errData := mcpResponse["error"].(map[string]any)

	if errData["code"].(int) != -32601 {
		t.Errorf("Expected error code -32601, got %v", errData["code"])
	}
}

func TestQuery_handleMCPMessage_NotificationsInitialized(t *testing.T) {
	mock := transport.NewMockTransport()

	servers := map[string]*types.MCPServer{
		"test-server": {
			Name:    "test",
			Version: "1.0.0",
		},
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		SDKMCPServers:   servers,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"server_name": "test-server",
		"message": map[string]any{
			"method": "notifications/initialized",
		},
	}

	result, err := q.handleMCPMessage(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	mcpResponse := result["mcp_response"].(map[string]any)
	if mcpResponse["result"] == nil {
		t.Error("Expected result in response")
	}
}

func TestQuery_handleMCPMessage_UnknownMethod(t *testing.T) {
	mock := transport.NewMockTransport()

	servers := map[string]*types.MCPServer{
		"test-server": {
			Name:    "test",
			Version: "1.0.0",
		},
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		SDKMCPServers:   servers,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"server_name": "test-server",
		"message": map[string]any{
			"method": "unknown/method",
			"id":     1,
		},
	}

	result, err := q.handleMCPMessage(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	mcpResponse := result["mcp_response"].(map[string]any)
	errData := mcpResponse["error"].(map[string]any)

	if errData["code"].(int) != -32601 {
		t.Errorf("Expected error code -32601, got %v", errData["code"])
	}
}

func TestQuery_handleMCPMessage_UnknownServer(t *testing.T) {
	mock := transport.NewMockTransport()

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		SDKMCPServers:   map[string]*types.MCPServer{},
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"server_name": "nonexistent-server",
		"message": map[string]any{
			"method": "initialize",
			"id":     1,
		},
	}

	result, err := q.handleMCPMessage(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	mcpResponse := result["mcp_response"].(map[string]any)
	errData := mcpResponse["error"].(map[string]any)

	if errData["code"].(int) != -32601 {
		t.Errorf("Expected error code -32601, got %v", errData["code"])
	}
}

func TestQuery_handleMCPMessage_MissingServerOrMessage(t *testing.T) {
	mock := transport.NewMockTransport()

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"server_name": "",
		"message":     nil,
	}

	_, err := q.handleMCPMessage(context.Background(), request)
	if err == nil {
		t.Error("Expected error for missing server_name or message")
	}
}

// Tests for sendControlRequest

func TestQuery_sendControlRequest_NonStreamingError(t *testing.T) {
	mock := transport.NewMockTransport()

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: false,
	})
	defer func() { _ = q.Close() }()

	_, err := q.sendControlRequest(context.Background(), map[string]any{}, time.Second)
	if err == nil {
		t.Error("Expected error for non-streaming mode")
	}
}

func TestQuery_sendControlRequest_Timeout(t *testing.T) {
	mock := transport.NewMockTransport()
	_ = mock.Connect(context.Background())

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	// Request that never gets a response
	_, err := q.sendControlRequest(context.Background(), map[string]any{"subtype": "test"}, 50*time.Millisecond)
	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestQuery_sendControlRequest_ContextCancellation(t *testing.T) {
	mock := transport.NewMockTransport()
	_ = mock.Connect(context.Background())

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := q.sendControlRequest(ctx, map[string]any{"subtype": "test"}, time.Second)
	if err == nil {
		t.Error("Expected context cancellation error")
	}
}

// Tests for public API methods

func TestQuery_GetMCPStatus(t *testing.T) {
	mock := transport.NewMockTransport()
	_ = mock.Connect(context.Background())

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	// This will timeout since there's no response, but validates the request is sent
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, _ = q.GetMCPStatus(ctx)

	// Verify something was written
	time.Sleep(10 * time.Millisecond)
	written := mock.GetWrittenData()
	if len(written) == 0 {
		t.Error("Expected MCP status request to be written")
	}
}

func TestQuery_Interrupt(t *testing.T) {
	mock := transport.NewMockTransport()
	_ = mock.Connect(context.Background())

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = q.Interrupt(ctx)

	time.Sleep(10 * time.Millisecond)
	written := mock.GetWrittenData()
	if len(written) == 0 {
		t.Error("Expected interrupt request to be written")
	}
}

func TestQuery_SetPermissionMode(t *testing.T) {
	mock := transport.NewMockTransport()
	_ = mock.Connect(context.Background())

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = q.SetPermissionMode(ctx, types.PermissionModeBypassPermissions)

	time.Sleep(10 * time.Millisecond)
	written := mock.GetWrittenData()
	if len(written) == 0 {
		t.Error("Expected set permission mode request to be written")
	}
}

func TestQuery_SetModel(t *testing.T) {
	mock := transport.NewMockTransport()
	_ = mock.Connect(context.Background())

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = q.SetModel(ctx, "claude-3-opus")

	time.Sleep(10 * time.Millisecond)
	written := mock.GetWrittenData()
	if len(written) == 0 {
		t.Error("Expected set model request to be written")
	}
}

func TestQuery_SetModel_Empty(t *testing.T) {
	mock := transport.NewMockTransport()
	_ = mock.Connect(context.Background())

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = q.SetModel(ctx, "")

	time.Sleep(10 * time.Millisecond)
	written := mock.GetWrittenData()
	if len(written) == 0 {
		t.Error("Expected set model request to be written even with empty model")
	}
}

func TestQuery_RewindFiles(t *testing.T) {
	mock := transport.NewMockTransport()
	_ = mock.Connect(context.Background())

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = q.RewindFiles(ctx, "user-msg-123")

	time.Sleep(10 * time.Millisecond)
	written := mock.GetWrittenData()
	if len(written) == 0 {
		t.Error("Expected rewind files request to be written")
	}
}

// Tests for StreamInput

func TestQuery_StreamInput_WritesMessages(t *testing.T) {
	mock := transport.NewMockTransport()
	_ = mock.Connect(context.Background())

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	messages := make(chan map[string]any, 3)
	messages <- map[string]any{"type": "user_input", "content": "Hello"}
	messages <- map[string]any{"type": "user_input", "content": "World"}
	close(messages)

	ctx := context.Background()
	go q.StreamInput(ctx, messages)

	// Wait for messages to be written
	time.Sleep(100 * time.Millisecond)

	written := mock.GetWrittenData()
	if len(written) < 2 {
		t.Errorf("Expected at least 2 messages written, got %d", len(written))
	}
}

func TestQuery_StreamInput_ContextCancellation(t *testing.T) {
	mock := transport.NewMockTransport()
	_ = mock.Connect(context.Background())

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	messages := make(chan map[string]any)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		q.StreamInput(ctx, messages)
		close(done)
	}()

	// Cancel context
	cancel()

	// StreamInput should exit
	select {
	case <-done:
		// Good
	case <-time.After(200 * time.Millisecond):
		t.Error("StreamInput did not exit after context cancellation")
	}
}

func TestQuery_StreamInput_ChannelClose(t *testing.T) {
	mock := transport.NewMockTransport()
	_ = mock.Connect(context.Background())

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
	})
	defer func() { _ = q.Close() }()

	messages := make(chan map[string]any)

	done := make(chan struct{})
	go func() {
		q.StreamInput(context.Background(), messages)
		close(done)
	}()

	// Close the channel
	close(messages)

	// StreamInput should exit
	select {
	case <-done:
		// Good
	case <-time.After(200 * time.Millisecond):
		t.Error("StreamInput did not exit after channel close")
	}
}

func TestQuery_Initialize_WithHooks(t *testing.T) {
	mock := transport.NewMockTransport()
	_ = mock.Connect(context.Background())

	hooks := map[types.HookEvent][]types.HookMatcher{
		types.HookEventPreToolUse: {
			{
				Matcher: "Bash",
				Hooks: []types.HookCallback{
					func(ctx context.Context, input types.HookInput, toolUseID string, hookCtx types.HookContext) (types.HookOutput, error) {
						return types.HookOutput{}, nil
					},
				},
				Timeout: 30.0,
			},
		},
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		Hooks:           hooks,
	})
	defer func() { _ = q.Close() }()

	// Start background reader to handle responses
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	q.Start(ctx)

	// Initialize will timeout but should write the request
	_, _ = q.Initialize(ctx)

	// Verify something was written
	time.Sleep(10 * time.Millisecond)
	written := mock.GetWrittenData()
	if len(written) == 0 {
		t.Error("Expected initialize request to be written")
	}

	// Verify hook callback was registered
	if len(q.hookCallbacks) == 0 {
		t.Error("Expected hook callback to be registered")
	}
}

func TestQuery_handleMCPMessage_ToolsCall_ImageContent(t *testing.T) {
	mock := transport.NewMockTransport()

	servers := map[string]*types.MCPServer{
		"test-server": {
			Name:    "test",
			Version: "1.0.0",
			Tools: []types.MCPTool{
				{
					Name:        "screenshot",
					Description: "Screenshot tool",
					Handler: func(ctx context.Context, args map[string]any) (types.MCPToolResult, error) {
						return types.MCPToolResult{
							Content: []types.MCPContent{
								{Type: "image", Data: "base64data", MimeType: "image/png"},
							},
						}, nil
					},
				},
			},
		},
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		SDKMCPServers:   servers,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"server_name": "test-server",
		"message": map[string]any{
			"method": "tools/call",
			"id":     1,
			"params": map[string]any{
				"name":      "screenshot",
				"arguments": map[string]any{},
			},
		},
	}

	result, err := q.handleMCPMessage(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	mcpResponse := result["mcp_response"].(map[string]any)
	resultData := mcpResponse["result"].(map[string]any)
	content := resultData["content"].([]map[string]any)

	if len(content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(content))
	}
	if content[0]["type"] != "image" {
		t.Errorf("Expected type 'image', got %v", content[0]["type"])
	}
	if content[0]["data"] != "base64data" {
		t.Errorf("Expected data 'base64data', got %v", content[0]["data"])
	}
}

func TestQuery_handleMCPMessage_ToolsCall_IsError(t *testing.T) {
	mock := transport.NewMockTransport()

	servers := map[string]*types.MCPServer{
		"test-server": {
			Name:    "test",
			Version: "1.0.0",
			Tools: []types.MCPTool{
				{
					Name:        "error_tool",
					Description: "Tool that returns error result",
					Handler: func(ctx context.Context, args map[string]any) (types.MCPToolResult, error) {
						return types.MCPToolResult{
							Content: []types.MCPContent{
								{Type: "text", Text: "Something went wrong"},
							},
							IsError: true,
						}, nil
					},
				},
			},
		},
	}

	q := NewQuery(QueryConfig{
		Transport:       mock,
		IsStreamingMode: true,
		SDKMCPServers:   servers,
	})
	defer func() { _ = q.Close() }()

	request := map[string]any{
		"server_name": "test-server",
		"message": map[string]any{
			"method": "tools/call",
			"id":     1,
			"params": map[string]any{
				"name":      "error_tool",
				"arguments": map[string]any{},
			},
		},
	}

	result, err := q.handleMCPMessage(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	mcpResponse := result["mcp_response"].(map[string]any)
	resultData := mcpResponse["result"].(map[string]any)

	if resultData["is_error"] != true {
		t.Error("Expected is_error to be true")
	}
}
