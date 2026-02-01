package protocol

import (
	"context"
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
	q.Close()
}

func TestNewQuery_WithTimeouts(t *testing.T) {
	mock := transport.NewMockTransport()

	q := NewQuery(QueryConfig{
		Transport:          mock,
		IsStreamingMode:    true,
		InitializeTimeout:  30 * time.Second,
		StreamCloseTimeout: 45 * time.Second,
	})
	defer q.Close()

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
	defer q.Close()

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
	defer q.Close()

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
	defer q.Close()

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
	defer q.Close()

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
	defer q.Close()

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
	defer q.Close()

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
	defer q.Close()

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
		q.Close()
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
