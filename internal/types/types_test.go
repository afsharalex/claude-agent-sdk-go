package types

import (
	"testing"
)

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
		t.Run(tt.expected, func(t *testing.T) {
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
		t.Run(tt.expected, func(t *testing.T) {
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
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.behavior) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.behavior))
			}
		})
	}
}

func TestPermissionUpdateType_Constants(t *testing.T) {
	tests := []struct {
		updateType PermissionUpdateType
		expected   string
	}{
		{PermissionUpdateTypeAddRules, "addRules"},
		{PermissionUpdateTypeReplaceRules, "replaceRules"},
		{PermissionUpdateTypeRemoveRules, "removeRules"},
		{PermissionUpdateTypeSetMode, "setMode"},
		{PermissionUpdateTypeAddDirectories, "addDirectories"},
		{PermissionUpdateTypeRemoveDirectories, "removeDirectories"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.updateType) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.updateType))
			}
		})
	}
}

func TestPermissionUpdate_ToMap(t *testing.T) {
	update := PermissionUpdate{
		Type: PermissionUpdateTypeAddRules,
		Rules: []PermissionRuleValue{
			{ToolName: "Bash", RuleContent: "allow npm"},
		},
		Behavior:    PermissionBehaviorAllow,
		Destination: PermissionUpdateDestinationSession,
	}

	result := update.ToMap()

	if result["type"] != "addRules" {
		t.Errorf("Expected type 'addRules', got '%v'", result["type"])
	}
	if result["destination"] != "session" {
		t.Errorf("Expected destination 'session', got '%v'", result["destination"])
	}
	if result["behavior"] != "allow" {
		t.Errorf("Expected behavior 'allow', got '%v'", result["behavior"])
	}

	rules, ok := result["rules"].([]map[string]any)
	if !ok {
		t.Fatal("Expected rules to be []map[string]any")
	}
	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}
	if rules[0]["toolName"] != "Bash" {
		t.Errorf("Expected toolName 'Bash', got '%v'", rules[0]["toolName"])
	}
}

func TestPermissionResult_Interface(t *testing.T) {
	// Verify both types implement PermissionResult
	var _ PermissionResult = PermissionResultAllow{}
	var _ PermissionResult = PermissionResultDeny{}
}

func TestPermissionResultAllow_IsAllow(t *testing.T) {
	result := PermissionResultAllow{}
	if !result.IsAllow() {
		t.Error("Expected IsAllow to return true")
	}
}

func TestPermissionResultDeny_IsAllow(t *testing.T) {
	result := PermissionResultDeny{}
	if result.IsAllow() {
		t.Error("Expected IsAllow to return false")
	}
}

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
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.event) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.event))
			}
		})
	}
}

func TestHookInput_Interface(t *testing.T) {
	// Verify all hook input types implement HookInput
	var _ HookInput = PreToolUseHookInput{}
	var _ HookInput = PostToolUseHookInput{}
	var _ HookInput = PostToolUseFailureHookInput{}
	var _ HookInput = UserPromptSubmitHookInput{}
	var _ HookInput = StopHookInput{}
	var _ HookInput = SubagentStopHookInput{}
	var _ HookInput = PreCompactHookInput{}
}

func TestBaseHookInput_GetSessionID(t *testing.T) {
	base := BaseHookInput{SessionID: "session-123"}
	if base.GetSessionID() != "session-123" {
		t.Errorf("Expected 'session-123', got '%s'", base.GetSessionID())
	}
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

func TestPreCompactTrigger_Constants(t *testing.T) {
	if PreCompactTriggerManual != "manual" {
		t.Errorf("Expected 'manual', got '%s'", PreCompactTriggerManual)
	}
	if PreCompactTriggerAuto != "auto" {
		t.Errorf("Expected 'auto', got '%s'", PreCompactTriggerAuto)
	}
}

func TestHookPermissionDecision_Constants(t *testing.T) {
	tests := []struct {
		decision HookPermissionDecision
		expected string
	}{
		{HookPermissionDecisionAllow, "allow"},
		{HookPermissionDecisionDeny, "deny"},
		{HookPermissionDecisionAsk, "ask"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.decision) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.decision))
			}
		})
	}
}

func TestHookDecision_Constants(t *testing.T) {
	if HookDecisionBlock != "block" {
		t.Errorf("Expected 'block', got '%s'", HookDecisionBlock)
	}
}

func TestHookOutput_ToMap_Async(t *testing.T) {
	output := HookOutput{
		Async:        true,
		AsyncTimeout: 5000,
	}

	result := output.ToMap()

	if result["async"] != true {
		t.Errorf("Expected async=true, got %v", result["async"])
	}
	if result["asyncTimeout"] != 5000 {
		t.Errorf("Expected asyncTimeout=5000, got %v", result["asyncTimeout"])
	}
}

func TestHookOutput_ToMap_NonAsync(t *testing.T) {
	cont := true
	output := HookOutput{
		Continue:       &cont,
		SuppressOutput: true,
		StopReason:     "test",
		Decision:       HookDecisionBlock,
		SystemMessage:  "warning",
		Reason:         "blocked",
	}

	result := output.ToMap()

	if result["continue"] != true {
		t.Errorf("Expected continue=true, got %v", result["continue"])
	}
	if result["suppressOutput"] != true {
		t.Errorf("Expected suppressOutput=true, got %v", result["suppressOutput"])
	}
	if result["stopReason"] != "test" {
		t.Errorf("Expected stopReason='test', got %v", result["stopReason"])
	}
	if result["decision"] != "block" {
		t.Errorf("Expected decision='block', got %v", result["decision"])
	}
	if result["systemMessage"] != "warning" {
		t.Errorf("Expected systemMessage='warning', got %v", result["systemMessage"])
	}
	if result["reason"] != "blocked" {
		t.Errorf("Expected reason='blocked', got %v", result["reason"])
	}
}

func TestHookOutput_ToMap_WithHookSpecificOutput(t *testing.T) {
	output := HookOutput{
		HookEventName:      HookEventPreToolUse,
		PermissionDecision: HookPermissionDecisionAllow,
		UpdatedInput:       map[string]any{"key": "value"},
		AdditionalContext:  "context",
	}

	result := output.ToMap()

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

func TestMCPServer_Fields(t *testing.T) {
	server := MCPServer{
		Name:    "test-server",
		Version: "1.0.0",
		Tools:   []MCPTool{},
	}

	if server.Name != "test-server" {
		t.Errorf("Expected Name 'test-server', got '%s'", server.Name)
	}
	if server.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got '%s'", server.Version)
	}
}

func TestMCPTool_Fields(t *testing.T) {
	tool := MCPTool{
		Name:        "my-tool",
		Description: "A tool",
		InputSchema: map[string]any{"type": "object"},
	}

	if tool.Name != "my-tool" {
		t.Errorf("Expected Name 'my-tool', got '%s'", tool.Name)
	}
	if tool.Description != "A tool" {
		t.Errorf("Expected Description 'A tool', got '%s'", tool.Description)
	}
}

func TestMCPContent_Fields(t *testing.T) {
	content := MCPContent{
		Type:     "text",
		Text:     "Hello",
		Data:     "base64",
		MimeType: "image/png",
	}

	if content.Type != "text" {
		t.Errorf("Expected Type 'text', got '%s'", content.Type)
	}
	if content.Text != "Hello" {
		t.Errorf("Expected Text 'Hello', got '%s'", content.Text)
	}
}

// Benchmark tests

func BenchmarkHookOutput_ToMap(b *testing.B) {
	cont := true
	output := HookOutput{
		Continue:           &cont,
		HookEventName:      HookEventPreToolUse,
		PermissionDecision: HookPermissionDecisionAllow,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = output.ToMap()
	}
}

func BenchmarkPermissionUpdate_ToMap(b *testing.B) {
	update := PermissionUpdate{
		Type: PermissionUpdateTypeAddRules,
		Rules: []PermissionRuleValue{
			{ToolName: "Bash", RuleContent: "allow"},
		},
		Behavior: PermissionBehaviorAllow,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = update.ToMap()
	}
}
