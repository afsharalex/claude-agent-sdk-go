package claude

import (
	"reflect"
	"testing"
)

func TestHookEvent_Constants(t *testing.T) {
	// Verify hook event constants have expected values
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
	// Verify all hook input types implement HookInput interface
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
	// When Async is true, only async fields should be included
	output := HookOutput{
		Async:        true,
		AsyncTimeout: 5000,
		Continue:     boolPtr(false), // Should be ignored when Async is true
		StopReason:   "test",         // Should be ignored when Async is true
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

func TestHookOutput_ToMap_AsyncWithoutTimeout(t *testing.T) {
	output := HookOutput{
		Async: true,
	}

	result := output.ToMap()

	if result["async"] != true {
		t.Errorf("Expected async=true, got %v", result["async"])
	}
	if _, exists := result["asyncTimeout"]; exists {
		t.Error("Expected 'asyncTimeout' to be absent when not set")
	}
}

func TestHookOutput_ToMap_Continue(t *testing.T) {
	tests := []struct {
		name     string
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
			if tt.suppress && result["suppressOutput"] != true {
				t.Errorf("Expected suppressOutput=true, got %v", result["suppressOutput"])
			}
		})
	}
}

func TestHookOutput_ToMap_StopReason(t *testing.T) {
	tests := []struct {
		name     string
		reason   string
		expected any
	}{
		{"non-empty", "User cancelled", "User cancelled"},
		{"empty", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := HookOutput{StopReason: tt.reason}
			result := output.ToMap()

			if tt.expected == nil {
				if _, exists := result["stopReason"]; exists {
					t.Error("Expected 'stopReason' to be absent when empty")
				}
			} else {
				if result["stopReason"] != tt.expected {
					t.Errorf("Expected stopReason=%v, got %v", tt.expected, result["stopReason"])
				}
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
			if tt.exists && result["decision"] != string(tt.decision) {
				t.Errorf("Expected decision=%s, got %v", tt.decision, result["decision"])
			}
		})
	}
}

func TestHookOutput_ToMap_SystemMessage(t *testing.T) {
	output := HookOutput{SystemMessage: "Warning: operation may fail"}
	result := output.ToMap()

	if result["systemMessage"] != "Warning: operation may fail" {
		t.Errorf("Expected systemMessage='Warning: operation may fail', got %v", result["systemMessage"])
	}

	// Empty should not be present
	output = HookOutput{SystemMessage: ""}
	result = output.ToMap()
	if _, exists := result["systemMessage"]; exists {
		t.Error("Expected 'systemMessage' to be absent when empty")
	}
}

func TestHookOutput_ToMap_Reason(t *testing.T) {
	output := HookOutput{Reason: "Security check failed"}
	result := output.ToMap()

	if result["reason"] != "Security check failed" {
		t.Errorf("Expected reason='Security check failed', got %v", result["reason"])
	}

	// Empty should not be present
	output = HookOutput{Reason: ""}
	result = output.ToMap()
	if _, exists := result["reason"]; exists {
		t.Error("Expected 'reason' to be absent when empty")
	}
}

func TestHookOutput_ToMap_PreToolUseHookSpecificOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   PreToolUseHookSpecificOutput
		expected map[string]any
	}{
		{
			name: "full output",
			output: PreToolUseHookSpecificOutput{
				HookEventName:            HookEventPreToolUse,
				PermissionDecision:       HookPermissionDecisionAllow,
				PermissionDecisionReason: "Approved by policy",
				UpdatedInput:             map[string]any{"path": "/safe/path"},
			},
			expected: map[string]any{
				"hookEventName":            "PreToolUse",
				"permissionDecision":       "allow",
				"permissionDecisionReason": "Approved by policy",
				"updatedInput":             map[string]any{"path": "/safe/path"},
			},
		},
		{
			name: "minimal output",
			output: PreToolUseHookSpecificOutput{
				HookEventName: HookEventPreToolUse,
			},
			expected: map[string]any{
				"hookEventName": "PreToolUse",
			},
		},
		{
			name: "deny decision",
			output: PreToolUseHookSpecificOutput{
				HookEventName:      HookEventPreToolUse,
				PermissionDecision: HookPermissionDecisionDeny,
			},
			expected: map[string]any{
				"hookEventName":      "PreToolUse",
				"permissionDecision": "deny",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hookOutput := HookOutput{HookSpecificOutput: tt.output}
			result := hookOutput.ToMap()

			specific, ok := result["hookSpecificOutput"].(map[string]any)
			if !ok {
				t.Fatal("Expected hookSpecificOutput to be map[string]any")
			}

			if !reflect.DeepEqual(specific, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, specific)
			}
		})
	}
}

func TestHookOutput_ToMap_PostToolUseHookSpecificOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   PostToolUseHookSpecificOutput
		expected map[string]any
	}{
		{
			name: "with context",
			output: PostToolUseHookSpecificOutput{
				HookEventName:     HookEventPostToolUse,
				AdditionalContext: "Tool completed successfully",
			},
			expected: map[string]any{
				"hookEventName":     "PostToolUse",
				"additionalContext": "Tool completed successfully",
			},
		},
		{
			name: "minimal",
			output: PostToolUseHookSpecificOutput{
				HookEventName: HookEventPostToolUse,
			},
			expected: map[string]any{
				"hookEventName": "PostToolUse",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hookOutput := HookOutput{HookSpecificOutput: tt.output}
			result := hookOutput.ToMap()

			specific, ok := result["hookSpecificOutput"].(map[string]any)
			if !ok {
				t.Fatal("Expected hookSpecificOutput to be map[string]any")
			}

			if !reflect.DeepEqual(specific, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, specific)
			}
		})
	}
}

func TestHookOutput_ToMap_PostToolUseFailureHookSpecificOutput(t *testing.T) {
	output := PostToolUseFailureHookSpecificOutput{
		HookEventName:     HookEventPostToolUseFailed,
		AdditionalContext: "Retry with different parameters",
	}

	hookOutput := HookOutput{HookSpecificOutput: output}
	result := hookOutput.ToMap()

	specific, ok := result["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatal("Expected hookSpecificOutput to be map[string]any")
	}

	expected := map[string]any{
		"hookEventName":     "PostToolUseFailure",
		"additionalContext": "Retry with different parameters",
	}

	if !reflect.DeepEqual(specific, expected) {
		t.Errorf("Expected %v, got %v", expected, specific)
	}
}

func TestHookOutput_ToMap_UserPromptSubmitHookSpecificOutput(t *testing.T) {
	output := UserPromptSubmitHookSpecificOutput{
		HookEventName:     HookEventUserPromptSubmit,
		AdditionalContext: "Prompt validated",
	}

	hookOutput := HookOutput{HookSpecificOutput: output}
	result := hookOutput.ToMap()

	specific, ok := result["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatal("Expected hookSpecificOutput to be map[string]any")
	}

	expected := map[string]any{
		"hookEventName":     "UserPromptSubmit",
		"additionalContext": "Prompt validated",
	}

	if !reflect.DeepEqual(specific, expected) {
		t.Errorf("Expected %v, got %v", expected, specific)
	}
}

func TestHookOutput_ToMap_NilHookSpecificOutput(t *testing.T) {
	output := HookOutput{}
	result := output.ToMap()

	if _, exists := result["hookSpecificOutput"]; exists {
		t.Error("Expected 'hookSpecificOutput' to be absent when nil")
	}
}

func TestHookOutput_ToMap_FullOutput(t *testing.T) {
	// Test a complete output with multiple fields
	output := HookOutput{
		Continue:       boolPtr(true),
		SuppressOutput: true,
		StopReason:     "Test stop",
		Decision:       HookDecisionBlock,
		SystemMessage:  "System warning",
		Reason:         "Block reason",
		HookSpecificOutput: PreToolUseHookSpecificOutput{
			HookEventName:      HookEventPreToolUse,
			PermissionDecision: HookPermissionDecisionDeny,
		},
	}

	result := output.ToMap()

	if result["continue"] != true {
		t.Errorf("Expected continue=true, got %v", result["continue"])
	}
	if result["suppressOutput"] != true {
		t.Errorf("Expected suppressOutput=true, got %v", result["suppressOutput"])
	}
	if result["stopReason"] != "Test stop" {
		t.Errorf("Expected stopReason='Test stop', got %v", result["stopReason"])
	}
	if result["decision"] != "block" {
		t.Errorf("Expected decision='block', got %v", result["decision"])
	}
	if result["systemMessage"] != "System warning" {
		t.Errorf("Expected systemMessage='System warning', got %v", result["systemMessage"])
	}
	if result["reason"] != "Block reason" {
		t.Errorf("Expected reason='Block reason', got %v", result["reason"])
	}
	if _, exists := result["hookSpecificOutput"]; !exists {
		t.Error("Expected 'hookSpecificOutput' to be present")
	}
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

func TestHookSpecificOutput_Interface(t *testing.T) {
	// Verify all hook specific output types implement HookSpecificOutput interface
	var _ HookSpecificOutput = PreToolUseHookSpecificOutput{}
	var _ HookSpecificOutput = PostToolUseHookSpecificOutput{}
	var _ HookSpecificOutput = PostToolUseFailureHookSpecificOutput{}
	var _ HookSpecificOutput = UserPromptSubmitHookSpecificOutput{}
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}
