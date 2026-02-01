package claude

import (
	"context"
	"reflect"
	"testing"
)

// =============================================================================
// Content Block Tests
// =============================================================================

func TestContentBlock_Interface(t *testing.T) {
	// Verify all content block types implement ContentBlock interface
	var _ ContentBlock = TextBlock{}
	var _ ContentBlock = ThinkingBlock{}
	var _ ContentBlock = ToolUseBlock{}
	var _ ContentBlock = ToolResultBlock{}
}

func TestMessage_Interface(t *testing.T) {
	// Verify all message types implement Message interface
	var _ Message = &UserMessage{}
	var _ Message = &AssistantMessage{}
	var _ Message = &SystemMessage{}
	var _ Message = &ResultMessage{}
	var _ Message = &StreamEvent{}
}

func TestTextBlock(t *testing.T) {
	block := TextBlock{Text: "Hello, world!"}

	if block.Text != "Hello, world!" {
		t.Errorf("Expected text 'Hello, world!', got '%s'", block.Text)
	}
}

func TestThinkingBlock(t *testing.T) {
	block := ThinkingBlock{
		Thinking:  "Let me think about this...",
		Signature: "sig-abc-123",
	}

	if block.Thinking != "Let me think about this..." {
		t.Errorf("Expected thinking text to match")
	}
	if block.Signature != "sig-abc-123" {
		t.Errorf("Expected signature 'sig-abc-123', got '%s'", block.Signature)
	}
}

func TestToolUseBlock(t *testing.T) {
	block := ToolUseBlock{
		ID:    "tool-123",
		Name:  "read_file",
		Input: map[string]any{"path": "/tmp/test.txt"},
	}

	if block.ID != "tool-123" {
		t.Errorf("Expected ID 'tool-123', got '%s'", block.ID)
	}
	if block.Name != "read_file" {
		t.Errorf("Expected Name 'read_file', got '%s'", block.Name)
	}
	if block.Input["path"] != "/tmp/test.txt" {
		t.Errorf("Expected Input path '/tmp/test.txt', got '%v'", block.Input["path"])
	}
}

func TestToolResultBlock(t *testing.T) {
	isErr := true
	block := ToolResultBlock{
		ToolUseID: "tool-456",
		Content:   "File contents here",
		IsError:   &isErr,
	}

	if block.ToolUseID != "tool-456" {
		t.Errorf("Expected ToolUseID 'tool-456', got '%s'", block.ToolUseID)
	}
	if block.Content != "File contents here" {
		t.Errorf("Expected Content to match")
	}
	if block.IsError == nil || !*block.IsError {
		t.Error("Expected IsError to be true")
	}
}

func TestToolResultBlock_NilIsError(t *testing.T) {
	block := ToolResultBlock{
		ToolUseID: "tool-789",
		Content:   "result",
	}

	if block.IsError != nil {
		t.Error("Expected IsError to be nil")
	}
}

// =============================================================================
// Message Tests
// =============================================================================

func TestUserMessage_GetContentString(t *testing.T) {
	msg := &UserMessage{Content: "Hello, Claude!"}

	if msg.GetContentString() != "Hello, Claude!" {
		t.Errorf("Expected 'Hello, Claude!', got '%s'", msg.GetContentString())
	}
}

func TestUserMessage_GetContentString_NotString(t *testing.T) {
	msg := &UserMessage{Content: []ContentBlock{TextBlock{Text: "Hello"}}}

	if msg.GetContentString() != "" {
		t.Errorf("Expected empty string for non-string content, got '%s'", msg.GetContentString())
	}
}

func TestUserMessage_GetContentBlocks(t *testing.T) {
	blocks := []ContentBlock{
		TextBlock{Text: "Hello"},
		ToolResultBlock{ToolUseID: "t1"},
	}
	msg := &UserMessage{Content: blocks}

	result := msg.GetContentBlocks()
	if len(result) != 2 {
		t.Fatalf("Expected 2 blocks, got %d", len(result))
	}

	_, ok := result[0].(TextBlock)
	if !ok {
		t.Errorf("Expected first block to be TextBlock")
	}
}

func TestUserMessage_GetContentBlocks_NotBlocks(t *testing.T) {
	msg := &UserMessage{Content: "string content"}

	result := msg.GetContentBlocks()
	if result != nil {
		t.Errorf("Expected nil for string content, got %v", result)
	}
}

func TestUserMessage_Fields(t *testing.T) {
	msg := &UserMessage{
		Content:         "Test",
		UUID:            "uuid-123",
		ParentToolUseID: "parent-456",
		ToolUseResult:   map[string]any{"success": true},
	}

	if msg.UUID != "uuid-123" {
		t.Errorf("Expected UUID 'uuid-123', got '%s'", msg.UUID)
	}
	if msg.ParentToolUseID != "parent-456" {
		t.Errorf("Expected ParentToolUseID 'parent-456', got '%s'", msg.ParentToolUseID)
	}
	if msg.ToolUseResult["success"] != true {
		t.Error("Expected ToolUseResult['success'] to be true")
	}
}

func TestAssistantMessage_Fields(t *testing.T) {
	msg := &AssistantMessage{
		Content: []ContentBlock{
			TextBlock{Text: "Hello!"},
		},
		Model:           "claude-3-opus",
		ParentToolUseID: "parent-123",
		Error:           AssistantMessageErrorRateLimit,
	}

	if len(msg.Content) != 1 {
		t.Errorf("Expected 1 content block, got %d", len(msg.Content))
	}
	if msg.Model != "claude-3-opus" {
		t.Errorf("Expected Model 'claude-3-opus', got '%s'", msg.Model)
	}
	if msg.ParentToolUseID != "parent-123" {
		t.Errorf("Expected ParentToolUseID 'parent-123', got '%s'", msg.ParentToolUseID)
	}
	if msg.Error != AssistantMessageErrorRateLimit {
		t.Errorf("Expected Error 'rate_limit', got '%s'", msg.Error)
	}
}

func TestAssistantMessageError_Constants(t *testing.T) {
	tests := []struct {
		err      AssistantMessageError
		expected string
	}{
		{AssistantMessageErrorAuthenticationFailed, "authentication_failed"},
		{AssistantMessageErrorBillingError, "billing_error"},
		{AssistantMessageErrorRateLimit, "rate_limit"},
		{AssistantMessageErrorInvalidRequest, "invalid_request"},
		{AssistantMessageErrorServerError, "server_error"},
		{AssistantMessageErrorUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.err) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.err))
			}
		})
	}
}

func TestSystemMessage_Fields(t *testing.T) {
	msg := &SystemMessage{
		Subtype: "init",
		Data: map[string]any{
			"session_id": "session-123",
			"tools":      []string{"read_file"},
		},
	}

	if msg.Subtype != "init" {
		t.Errorf("Expected Subtype 'init', got '%s'", msg.Subtype)
	}
	if msg.Data == nil {
		t.Error("Expected Data to be non-nil")
	}
	if msg.Data["session_id"] != "session-123" {
		t.Errorf("Expected session_id 'session-123', got '%v'", msg.Data["session_id"])
	}
}

func TestResultMessage_Fields(t *testing.T) {
	cost := 0.0045
	msg := &ResultMessage{
		Subtype:       "success",
		DurationMs:    1500,
		DurationAPIMs: 1200,
		IsError:       false,
		NumTurns:      3,
		SessionID:     "session-789",
		TotalCostUSD:  &cost,
		Usage: map[string]any{
			"input_tokens":  100,
			"output_tokens": 250,
		},
		Result:           "completed",
		StructuredOutput: map[string]any{"status": "ok"},
	}

	if msg.Subtype != "success" {
		t.Errorf("Expected Subtype 'success', got '%s'", msg.Subtype)
	}
	if msg.DurationMs != 1500 {
		t.Errorf("Expected DurationMs 1500, got %d", msg.DurationMs)
	}
	if msg.DurationAPIMs != 1200 {
		t.Errorf("Expected DurationAPIMs 1200, got %d", msg.DurationAPIMs)
	}
	if msg.IsError {
		t.Error("Expected IsError to be false")
	}
	if msg.NumTurns != 3 {
		t.Errorf("Expected NumTurns 3, got %d", msg.NumTurns)
	}
	if msg.SessionID != "session-789" {
		t.Errorf("Expected SessionID 'session-789', got '%s'", msg.SessionID)
	}
	if msg.TotalCostUSD == nil || *msg.TotalCostUSD != 0.0045 {
		t.Errorf("Expected TotalCostUSD 0.0045, got %v", msg.TotalCostUSD)
	}
	if msg.Usage == nil {
		t.Error("Expected Usage to be non-nil")
	}
	if msg.Result != "completed" {
		t.Errorf("Expected Result 'completed', got '%s'", msg.Result)
	}
	if msg.StructuredOutput == nil {
		t.Error("Expected StructuredOutput to be non-nil")
	}
}

func TestResultMessage_NilOptionalFields(t *testing.T) {
	msg := &ResultMessage{
		Subtype: "success",
	}

	if msg.TotalCostUSD != nil {
		t.Error("Expected TotalCostUSD to be nil")
	}
	if msg.Usage != nil {
		t.Error("Expected Usage to be nil")
	}
	if msg.StructuredOutput != nil {
		t.Error("Expected StructuredOutput to be nil")
	}
}

func TestStreamEvent_Fields(t *testing.T) {
	msg := &StreamEvent{
		UUID:      "stream-001",
		SessionID: "session-stream",
		Event: map[string]any{
			"type": "content_block_delta",
			"delta": map[string]any{
				"text": "Hello",
			},
		},
		ParentToolUseID: "parent-123",
	}

	if msg.UUID != "stream-001" {
		t.Errorf("Expected UUID 'stream-001', got '%s'", msg.UUID)
	}
	if msg.SessionID != "session-stream" {
		t.Errorf("Expected SessionID 'session-stream', got '%s'", msg.SessionID)
	}
	if msg.Event == nil {
		t.Error("Expected Event to be non-nil")
	}
	if msg.Event["type"] != "content_block_delta" {
		t.Errorf("Expected Event type 'content_block_delta', got '%v'", msg.Event["type"])
	}
	if msg.ParentToolUseID != "parent-123" {
		t.Errorf("Expected ParentToolUseID 'parent-123', got '%s'", msg.ParentToolUseID)
	}
}

// =============================================================================
// Hook Tests
// =============================================================================

func TestHookEvent_Constants(t *testing.T) {
	tests := []struct {
		event    HookEvent
		expected string
	}{
		{HookEventPreToolUse, "PreToolUse"},
		{HookEventPostToolUse, "PostToolUse"},
		{HookEventPostToolUseFailed, "PostToolUseFailure"},
		{HookEventUserPromptSubmit, "UserPromptSubmit"},
		{HookEventStop, "Stop"},
		{HookEventSubagentStop, "SubagentStop"},
		{HookEventPreCompact, "PreCompact"},
	}

	for _, tt := range tests {
		t.Run(string(tt.event), func(t *testing.T) {
			if string(tt.event) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.event))
			}
		})
	}
}

func TestBaseHookInput_Getters(t *testing.T) {
	base := BaseHookInput{
		SessionID:      "session-123",
		TranscriptPath: "/path/to/transcript",
		Cwd:            "/home/user",
		PermissionMode: "default",
	}

	if base.GetSessionID() != "session-123" {
		t.Errorf("Expected session ID 'session-123', got '%s'", base.GetSessionID())
	}
	if base.GetTranscriptPath() != "/path/to/transcript" {
		t.Errorf("Expected transcript path '/path/to/transcript', got '%s'", base.GetTranscriptPath())
	}
	if base.GetCwd() != "/home/user" {
		t.Errorf("Expected cwd '/home/user', got '%s'", base.GetCwd())
	}
}

func TestHookInputTypes_Interface(t *testing.T) {
	var _ HookInput = PreToolUseHookInput{}
	var _ HookInput = PostToolUseHookInput{}
	var _ HookInput = PostToolUseFailureHookInput{}
	var _ HookInput = UserPromptSubmitHookInput{}
	var _ HookInput = StopHookInput{}
	var _ HookInput = SubagentStopHookInput{}
	var _ HookInput = PreCompactHookInput{}
}

func TestHookInputTypes_GetHookEventName(t *testing.T) {
	tests := []struct {
		name     string
		input    HookInput
		expected HookEvent
	}{
		{"PreToolUse", PreToolUseHookInput{}, HookEventPreToolUse},
		{"PostToolUse", PostToolUseHookInput{}, HookEventPostToolUse},
		{"PostToolUseFailure", PostToolUseFailureHookInput{}, HookEventPostToolUseFailed},
		{"UserPromptSubmit", UserPromptSubmitHookInput{}, HookEventUserPromptSubmit},
		{"Stop", StopHookInput{}, HookEventStop},
		{"SubagentStop", SubagentStopHookInput{}, HookEventSubagentStop},
		{"PreCompact", PreCompactHookInput{}, HookEventPreCompact},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input.GetHookEventName() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.input.GetHookEventName())
			}
		})
	}
}

func TestHookOutput_ToMap_Async(t *testing.T) {
	output := HookOutput{
		Async:        true,
		AsyncTimeout: 5000,
		Continue:     boolPtr(false),
		StopReason:   "test",
	}

	result := output.ToMap()

	if result["async"] != true {
		t.Errorf("Expected async=true, got %v", result["async"])
	}
	if result["asyncTimeout"] != 5000 {
		t.Errorf("Expected asyncTimeout=5000, got %v", result["asyncTimeout"])
	}
	if _, exists := result["continue"]; exists {
		t.Error("Expected 'continue' to be absent when Async is true")
	}
	if _, exists := result["stopReason"]; exists {
		t.Error("Expected 'stopReason' to be absent when Async is true")
	}
}

func TestHookOutput_ToMap_Continue(t *testing.T) {
	tests := []struct {
		name      string
		continue_ *bool
		expected  any
	}{
		{"true", boolPtr(true), true},
		{"false", boolPtr(false), false},
		{"nil", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := HookOutput{Continue: tt.continue_}
			result := output.ToMap()

			if tt.expected == nil {
				if _, exists := result["continue"]; exists {
					t.Error("Expected 'continue' to be absent when nil")
				}
			} else {
				if result["continue"] != tt.expected {
					t.Errorf("Expected continue=%v, got %v", tt.expected, result["continue"])
				}
			}
		})
	}
}

func TestHookOutput_ToMap_SuppressOutput(t *testing.T) {
	tests := []struct {
		name     string
		suppress bool
		exists   bool
	}{
		{"true", true, true},
		{"false", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := HookOutput{SuppressOutput: tt.suppress}
			result := output.ToMap()

			_, exists := result["suppressOutput"]
			if exists != tt.exists {
				t.Errorf("Expected suppressOutput exists=%v, got %v", tt.exists, exists)
			}
		})
	}
}

func TestHookOutput_ToMap_Decision(t *testing.T) {
	tests := []struct {
		name     string
		decision HookDecision
		exists   bool
	}{
		{"block", HookDecisionBlock, true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := HookOutput{Decision: tt.decision}
			result := output.ToMap()

			_, exists := result["decision"]
			if exists != tt.exists {
				t.Errorf("Expected decision exists=%v, got %v", tt.exists, exists)
			}
		})
	}
}

func TestHookOutput_ToMap_PreToolUseHookSpecificOutput(t *testing.T) {
	output := PreToolUseHookSpecificOutput{
		HookEventName:            HookEventPreToolUse,
		PermissionDecision:       HookPermissionDecisionAllow,
		PermissionDecisionReason: "Approved by policy",
		UpdatedInput:             map[string]any{"path": "/safe/path"},
	}

	hookOutput := HookOutput{HookSpecificOutput: output}
	result := hookOutput.ToMap()

	specific, ok := result["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatal("Expected hookSpecificOutput to be map[string]any")
	}

	if specific["hookEventName"] != "PreToolUse" {
		t.Errorf("Expected hookEventName='PreToolUse', got %v", specific["hookEventName"])
	}
	if specific["permissionDecision"] != "allow" {
		t.Errorf("Expected permissionDecision='allow', got %v", specific["permissionDecision"])
	}
}

func TestHookSpecificOutput_Interface(t *testing.T) {
	var _ HookSpecificOutput = PreToolUseHookSpecificOutput{}
	var _ HookSpecificOutput = PostToolUseHookSpecificOutput{}
	var _ HookSpecificOutput = PostToolUseFailureHookSpecificOutput{}
	var _ HookSpecificOutput = UserPromptSubmitHookSpecificOutput{}
}

func TestPreCompactTrigger_Constants(t *testing.T) {
	if PreCompactTriggerManual != "manual" {
		t.Errorf("Expected PreCompactTriggerManual='manual', got '%s'", PreCompactTriggerManual)
	}
	if PreCompactTriggerAuto != "auto" {
		t.Errorf("Expected PreCompactTriggerAuto='auto', got '%s'", PreCompactTriggerAuto)
	}
}

func TestHookPermissionDecision_Constants(t *testing.T) {
	if HookPermissionDecisionAllow != "allow" {
		t.Errorf("Expected HookPermissionDecisionAllow='allow', got '%s'", HookPermissionDecisionAllow)
	}
	if HookPermissionDecisionDeny != "deny" {
		t.Errorf("Expected HookPermissionDecisionDeny='deny', got '%s'", HookPermissionDecisionDeny)
	}
	if HookPermissionDecisionAsk != "ask" {
		t.Errorf("Expected HookPermissionDecisionAsk='ask', got '%s'", HookPermissionDecisionAsk)
	}
}

func TestHookDecision_Constants(t *testing.T) {
	if HookDecisionBlock != "block" {
		t.Errorf("Expected HookDecisionBlock='block', got '%s'", HookDecisionBlock)
	}
}

// =============================================================================
// Permission Tests
// =============================================================================

func TestPermissionMode_Constants(t *testing.T) {
	tests := []struct {
		mode     PermissionMode
		expected string
	}{
		{PermissionModeDefault, "default"},
		{PermissionModeAcceptEdits, "acceptEdits"},
		{PermissionModePlan, "plan"},
		{PermissionModeBypassPermissions, "bypassPermissions"},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			if string(tt.mode) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.mode))
			}
		})
	}
}

func TestPermissionUpdateDestination_Constants(t *testing.T) {
	tests := []struct {
		dest     PermissionUpdateDestination
		expected string
	}{
		{PermissionUpdateDestinationUserSettings, "userSettings"},
		{PermissionUpdateDestinationProjectSettings, "projectSettings"},
		{PermissionUpdateDestinationLocalSettings, "localSettings"},
		{PermissionUpdateDestinationSession, "session"},
	}

	for _, tt := range tests {
		t.Run(string(tt.dest), func(t *testing.T) {
			if string(tt.dest) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.dest))
			}
		})
	}
}

func TestPermissionBehavior_Constants(t *testing.T) {
	tests := []struct {
		behavior PermissionBehavior
		expected string
	}{
		{PermissionBehaviorAllow, "allow"},
		{PermissionBehaviorDeny, "deny"},
		{PermissionBehaviorAsk, "ask"},
	}

	for _, tt := range tests {
		t.Run(string(tt.behavior), func(t *testing.T) {
			if string(tt.behavior) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.behavior))
			}
		})
	}
}

func TestPermissionUpdate_ToMap_AddRules(t *testing.T) {
	update := PermissionUpdate{
		Type: PermissionUpdateTypeAddRules,
		Rules: []PermissionRuleValue{
			{ToolName: "Bash", RuleContent: "allow npm commands"},
		},
		Behavior:    PermissionBehaviorAllow,
		Destination: PermissionUpdateDestinationSession,
	}

	result := update.ToMap()

	if result["type"] != "addRules" {
		t.Errorf("Expected type='addRules', got %v", result["type"])
	}
	if result["destination"] != "session" {
		t.Errorf("Expected destination='session', got %v", result["destination"])
	}
	if result["behavior"] != "allow" {
		t.Errorf("Expected behavior='allow', got %v", result["behavior"])
	}
}

func TestPermissionUpdate_ToMap_SetMode(t *testing.T) {
	update := PermissionUpdate{
		Type:        PermissionUpdateTypeSetMode,
		Mode:        PermissionModeBypassPermissions,
		Destination: PermissionUpdateDestinationUserSettings,
	}

	result := update.ToMap()

	if result["type"] != "setMode" {
		t.Errorf("Expected type='setMode', got %v", result["type"])
	}
	if result["mode"] != "bypassPermissions" {
		t.Errorf("Expected mode='bypassPermissions', got %v", result["mode"])
	}
}

func TestPermissionResult_Interface(t *testing.T) {
	var _ PermissionResult = PermissionResultAllow{}
	var _ PermissionResult = PermissionResultDeny{}
}

func TestPermissionResultAllow_Fields(t *testing.T) {
	result := PermissionResultAllow{
		UpdatedInput: map[string]any{"path": "/safe/path"},
		UpdatedPermissions: []PermissionUpdate{
			{Type: PermissionUpdateTypeAddRules},
		},
	}

	if result.UpdatedInput["path"] != "/safe/path" {
		t.Errorf("Expected UpdatedInput path='/safe/path', got %v", result.UpdatedInput["path"])
	}
	if len(result.UpdatedPermissions) != 1 {
		t.Errorf("Expected 1 updated permission, got %d", len(result.UpdatedPermissions))
	}
}

func TestPermissionResultDeny_Fields(t *testing.T) {
	result := PermissionResultDeny{
		Message:   "Operation not allowed",
		Interrupt: true,
	}

	if result.Message != "Operation not allowed" {
		t.Errorf("Expected Message='Operation not allowed', got '%s'", result.Message)
	}
	if !result.Interrupt {
		t.Error("Expected Interrupt to be true")
	}
}

func TestToolPermissionContext_Fields(t *testing.T) {
	ctx := ToolPermissionContext{
		Signal: nil,
		Suggestions: []PermissionUpdate{
			{Type: PermissionUpdateTypeSetMode, Mode: PermissionModeDefault},
		},
	}

	if ctx.Signal != nil {
		t.Error("Expected Signal to be nil")
	}
	if len(ctx.Suggestions) != 1 {
		t.Errorf("Expected 1 suggestion, got %d", len(ctx.Suggestions))
	}
}

// =============================================================================
// MCP Config Tests
// =============================================================================

func TestMCPServerConfig_Interface(t *testing.T) {
	var _ MCPServerConfig = MCPStdioServerConfig{}
	var _ MCPServerConfig = MCPSSEServerConfig{}
	var _ MCPServerConfig = MCPHTTPServerConfig{}
	var _ MCPServerConfig = MCPSDKServerConfig{}
}

func TestMCPStdioServerConfig_GetType(t *testing.T) {
	tests := []struct {
		name     string
		config   MCPStdioServerConfig
		expected string
	}{
		{"default type", MCPStdioServerConfig{Command: "npx"}, "stdio"},
		{"explicit type", MCPStdioServerConfig{Type: "stdio", Command: "npx"}, "stdio"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.GetType() != tt.expected {
				t.Errorf("Expected type '%s', got '%s'", tt.expected, tt.config.GetType())
			}
		})
	}
}

func TestMCPStdioServerConfig_Fields(t *testing.T) {
	config := MCPStdioServerConfig{
		Type:    "stdio",
		Command: "npx",
		Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
		Env:     map[string]string{"PATH": "/usr/bin"},
	}

	if config.Command != "npx" {
		t.Errorf("Expected Command 'npx', got '%s'", config.Command)
	}
	if len(config.Args) != 2 {
		t.Errorf("Expected 2 Args, got %d", len(config.Args))
	}
	if config.Env["PATH"] != "/usr/bin" {
		t.Errorf("Expected Env['PATH']='/usr/bin', got '%s'", config.Env["PATH"])
	}
}

func TestMCPSSEServerConfig_GetType(t *testing.T) {
	config := MCPSSEServerConfig{Type: "sse", URL: "https://example.com/sse"}
	if config.GetType() != "sse" {
		t.Errorf("Expected type 'sse', got '%s'", config.GetType())
	}
}

func TestMCPHTTPServerConfig_GetType(t *testing.T) {
	config := MCPHTTPServerConfig{Type: "http", URL: "https://example.com/api"}
	if config.GetType() != "http" {
		t.Errorf("Expected type 'http', got '%s'", config.GetType())
	}
}

func TestMCPSDKServerConfig_GetType(t *testing.T) {
	config := MCPSDKServerConfig{Type: "sdk", Name: "test-server"}
	if config.GetType() != "sdk" {
		t.Errorf("Expected type 'sdk', got '%s'", config.GetType())
	}
}

func TestNewMCPServer(t *testing.T) {
	tools := []MCPTool{
		{Name: "test-tool", Description: "A test tool"},
	}
	server := NewMCPServer("my-server", "1.0.0", tools)

	if server == nil {
		t.Fatal("Expected non-nil server")
	}
	if server.Name() != "my-server" {
		t.Errorf("Expected name 'my-server', got '%s'", server.Name())
	}
	if server.Version() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", server.Version())
	}
	if len(server.Tools()) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(server.Tools()))
	}
}

func TestMCPTool_Fields(t *testing.T) {
	handler := func(ctx context.Context, args map[string]any) (MCPToolResult, error) {
		return TextResult("result"), nil
	}

	tool := MCPTool{
		Name:        "my-tool",
		Description: "Does something useful",
		InputSchema: map[string]any{"type": "object"},
		Handler:     handler,
	}

	if tool.Name != "my-tool" {
		t.Errorf("Expected Name 'my-tool', got '%s'", tool.Name)
	}
	if tool.Description != "Does something useful" {
		t.Errorf("Expected Description to match")
	}
	if tool.Handler == nil {
		t.Error("Expected Handler to be non-nil")
	}
}

func TestMCPToolResult_Fields(t *testing.T) {
	result := MCPToolResult{
		Content: []MCPContent{
			{Type: "text", Text: "Hello"},
			{Type: "image", Data: "base64data", MimeType: "image/png"},
		},
		IsError: true,
	}

	if len(result.Content) != 2 {
		t.Errorf("Expected 2 content items, got %d", len(result.Content))
	}
	if !result.IsError {
		t.Error("Expected IsError to be true")
	}
}

func TestMCPContent_Text(t *testing.T) {
	content := MCPContent{Type: "text", Text: "Hello, world!"}
	if content.Type != "text" {
		t.Errorf("Expected Type 'text', got '%s'", content.Type)
	}
	if content.Text != "Hello, world!" {
		t.Errorf("Expected Text 'Hello, world!', got '%s'", content.Text)
	}
}

func TestMCPContent_Image(t *testing.T) {
	content := MCPContent{Type: "image", Data: "base64==", MimeType: "image/png"}
	if content.Type != "image" {
		t.Errorf("Expected Type 'image', got '%s'", content.Type)
	}
	if content.MimeType != "image/png" {
		t.Errorf("Expected MimeType 'image/png', got '%s'", content.MimeType)
	}
}

func TestCreateSDKMCPServer(t *testing.T) {
	tools := []MCPTool{{Name: "add", Description: "Add numbers"}}
	config := CreateSDKMCPServer("math-server", "1.0.0", tools)

	if config.Type != "sdk" {
		t.Errorf("Expected Type 'sdk', got '%s'", config.Type)
	}
	if config.Name != "math-server" {
		t.Errorf("Expected Name 'math-server', got '%s'", config.Name)
	}
	if config.Server == nil {
		t.Fatal("Expected Server to be non-nil")
	}
}

// =============================================================================
// Configuration Type Tests
// =============================================================================

func TestSettingSource_Constants(t *testing.T) {
	tests := []struct {
		source   SettingSource
		expected string
	}{
		{SettingSourceUser, "user"},
		{SettingSourceProject, "project"},
		{SettingSourceLocal, "local"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.source) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.source))
			}
		})
	}
}

func TestAgentDefinition_Fields(t *testing.T) {
	agent := AgentDefinition{
		Description: "A test agent",
		Prompt:      "You are a test agent",
		Tools:       []string{"read_file", "write_file"},
		Model:       "sonnet",
	}

	if agent.Description != "A test agent" {
		t.Errorf("Expected Description to match")
	}
	if len(agent.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(agent.Tools))
	}
}

func TestSandboxSettings_Fields(t *testing.T) {
	settings := SandboxSettings{
		Enabled:                  true,
		AutoAllowBashIfSandboxed: true,
		ExcludedCommands:         []string{"docker"},
		Network: &SandboxNetworkConfig{
			HTTPProxyPort: 8080,
		},
	}

	if !settings.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if settings.Network.HTTPProxyPort != 8080 {
		t.Errorf("Expected HTTPProxyPort 8080, got %d", settings.Network.HTTPProxyPort)
	}
}

func TestSdkBeta_Constants(t *testing.T) {
	if SdkBetaContext1M != "context-1m-2025-08-07" {
		t.Errorf("Expected SdkBetaContext1M 'context-1m-2025-08-07', got '%s'", SdkBetaContext1M)
	}
}

func TestSystemPromptPreset_Fields(t *testing.T) {
	preset := SystemPromptPreset{
		Type:   "preset",
		Preset: "claude_code",
		Append: "Additional instructions",
	}

	if preset.Type != "preset" {
		t.Errorf("Expected Type 'preset', got '%s'", preset.Type)
	}
	if preset.Preset != "claude_code" {
		t.Errorf("Expected Preset 'claude_code', got '%s'", preset.Preset)
	}
}

func TestToolsPreset_Fields(t *testing.T) {
	preset := ToolsPreset{Type: "preset", Preset: "claude_code"}
	if preset.Type != "preset" {
		t.Errorf("Expected Type 'preset', got '%s'", preset.Type)
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkUserMessage_GetContentString(b *testing.B) {
	msg := &UserMessage{Content: "Hello, Claude!"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.GetContentString()
	}
}

func BenchmarkUserMessage_GetContentBlocks(b *testing.B) {
	blocks := []ContentBlock{TextBlock{Text: "Hello"}, TextBlock{Text: "World"}}
	msg := &UserMessage{Content: blocks}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.GetContentBlocks()
	}
}

func BenchmarkNewMCPServer(b *testing.B) {
	tools := []MCPTool{{Name: "tool1"}, {Name: "tool2"}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewMCPServer("server", "1.0.0", tools)
	}
}

func BenchmarkPermissionUpdate_ToMap(b *testing.B) {
	update := PermissionUpdate{
		Type:        PermissionUpdateTypeAddRules,
		Rules:       []PermissionRuleValue{{ToolName: "Bash", RuleContent: "allow"}},
		Behavior:    PermissionBehaviorAllow,
		Destination: PermissionUpdateDestinationSession,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = update.ToMap()
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func boolPtr(b bool) *bool {
	return &b
}

// Suppress unused import warning for reflect
var _ = reflect.DeepEqual
