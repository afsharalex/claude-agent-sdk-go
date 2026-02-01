package claude

import (
	"reflect"
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
		t.Run(string(tt.updateType), func(t *testing.T) {
			if string(tt.updateType) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.updateType))
			}
		})
	}
}

func TestPermissionUpdate_ToMap_AddRules(t *testing.T) {
	update := PermissionUpdate{
		Type: PermissionUpdateTypeAddRules,
		Rules: []PermissionRuleValue{
			{ToolName: "Bash", RuleContent: "allow npm commands"},
			{ToolName: "Write", RuleContent: "allow .go files"},
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

	rules, ok := result["rules"].([]map[string]any)
	if !ok {
		t.Fatalf("Expected rules to be []map[string]any, got %T", result["rules"])
	}
	if len(rules) != 2 {
		t.Fatalf("Expected 2 rules, got %d", len(rules))
	}
	if rules[0]["toolName"] != "Bash" {
		t.Errorf("Expected first rule toolName='Bash', got %v", rules[0]["toolName"])
	}
	if rules[0]["ruleContent"] != "allow npm commands" {
		t.Errorf("Expected first rule ruleContent='allow npm commands', got %v", rules[0]["ruleContent"])
	}
}

func TestPermissionUpdate_ToMap_ReplaceRules(t *testing.T) {
	update := PermissionUpdate{
		Type: PermissionUpdateTypeReplaceRules,
		Rules: []PermissionRuleValue{
			{ToolName: "Read", RuleContent: "deny /etc"},
		},
		Behavior:    PermissionBehaviorDeny,
		Destination: PermissionUpdateDestinationProjectSettings,
	}

	result := update.ToMap()

	if result["type"] != "replaceRules" {
		t.Errorf("Expected type='replaceRules', got %v", result["type"])
	}
	if result["destination"] != "projectSettings" {
		t.Errorf("Expected destination='projectSettings', got %v", result["destination"])
	}
	if result["behavior"] != "deny" {
		t.Errorf("Expected behavior='deny', got %v", result["behavior"])
	}

	rules, ok := result["rules"].([]map[string]any)
	if !ok {
		t.Fatalf("Expected rules to be []map[string]any")
	}
	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}
}

func TestPermissionUpdate_ToMap_RemoveRules(t *testing.T) {
	update := PermissionUpdate{
		Type: PermissionUpdateTypeRemoveRules,
		Rules: []PermissionRuleValue{
			{ToolName: "Bash"},
		},
	}

	result := update.ToMap()

	if result["type"] != "removeRules" {
		t.Errorf("Expected type='removeRules', got %v", result["type"])
	}

	// Should not have destination when not set
	if _, exists := result["destination"]; exists {
		t.Error("Expected destination to be absent when not set")
	}

	rules, ok := result["rules"].([]map[string]any)
	if !ok {
		t.Fatalf("Expected rules to be []map[string]any")
	}
	if rules[0]["toolName"] != "Bash" {
		t.Errorf("Expected toolName='Bash', got %v", rules[0]["toolName"])
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
	if result["destination"] != "userSettings" {
		t.Errorf("Expected destination='userSettings', got %v", result["destination"])
	}

	// Should not have rules for setMode
	if _, exists := result["rules"]; exists {
		t.Error("Expected rules to be absent for setMode")
	}
}

func TestPermissionUpdate_ToMap_SetModeWithoutMode(t *testing.T) {
	update := PermissionUpdate{
		Type: PermissionUpdateTypeSetMode,
	}

	result := update.ToMap()

	if result["type"] != "setMode" {
		t.Errorf("Expected type='setMode', got %v", result["type"])
	}
	if _, exists := result["mode"]; exists {
		t.Error("Expected mode to be absent when not set")
	}
}

func TestPermissionUpdate_ToMap_AddDirectories(t *testing.T) {
	update := PermissionUpdate{
		Type:        PermissionUpdateTypeAddDirectories,
		Directories: []string{"/home/user/project", "/tmp/work"},
		Destination: PermissionUpdateDestinationLocalSettings,
	}

	result := update.ToMap()

	if result["type"] != "addDirectories" {
		t.Errorf("Expected type='addDirectories', got %v", result["type"])
	}
	if result["destination"] != "localSettings" {
		t.Errorf("Expected destination='localSettings', got %v", result["destination"])
	}

	dirs, ok := result["directories"].([]string)
	if !ok {
		t.Fatalf("Expected directories to be []string, got %T", result["directories"])
	}
	if !reflect.DeepEqual(dirs, []string{"/home/user/project", "/tmp/work"}) {
		t.Errorf("Unexpected directories: %v", dirs)
	}

	// Should not have rules for addDirectories
	if _, exists := result["rules"]; exists {
		t.Error("Expected rules to be absent for addDirectories")
	}
}

func TestPermissionUpdate_ToMap_RemoveDirectories(t *testing.T) {
	update := PermissionUpdate{
		Type:        PermissionUpdateTypeRemoveDirectories,
		Directories: []string{"/tmp/old"},
	}

	result := update.ToMap()

	if result["type"] != "removeDirectories" {
		t.Errorf("Expected type='removeDirectories', got %v", result["type"])
	}

	dirs, ok := result["directories"].([]string)
	if !ok {
		t.Fatalf("Expected directories to be []string")
	}
	if len(dirs) != 1 || dirs[0] != "/tmp/old" {
		t.Errorf("Expected directories=['/tmp/old'], got %v", dirs)
	}
}

func TestPermissionUpdate_ToMap_NilRules(t *testing.T) {
	update := PermissionUpdate{
		Type:  PermissionUpdateTypeAddRules,
		Rules: nil,
	}

	result := update.ToMap()

	if _, exists := result["rules"]; exists {
		t.Error("Expected rules to be absent when nil")
	}
}

func TestPermissionUpdate_ToMap_NilDirectories(t *testing.T) {
	update := PermissionUpdate{
		Type:        PermissionUpdateTypeAddDirectories,
		Directories: nil,
	}

	result := update.ToMap()

	if _, exists := result["directories"]; exists {
		t.Error("Expected directories to be absent when nil")
	}
}

func TestPermissionUpdate_ToMap_EmptyBehavior(t *testing.T) {
	update := PermissionUpdate{
		Type: PermissionUpdateTypeAddRules,
		Rules: []PermissionRuleValue{
			{ToolName: "Bash"},
		},
		Behavior: "",
	}

	result := update.ToMap()

	if _, exists := result["behavior"]; exists {
		t.Error("Expected behavior to be absent when empty")
	}
}

func TestPermissionRuleValue_Fields(t *testing.T) {
	rule := PermissionRuleValue{
		ToolName:    "Bash",
		RuleContent: "allow specific commands",
	}

	if rule.ToolName != "Bash" {
		t.Errorf("Expected ToolName='Bash', got '%s'", rule.ToolName)
	}
	if rule.RuleContent != "allow specific commands" {
		t.Errorf("Expected RuleContent='allow specific commands', got '%s'", rule.RuleContent)
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
	if ctx.Suggestions[0].Type != PermissionUpdateTypeSetMode {
		t.Errorf("Expected suggestion type 'setMode', got '%s'", ctx.Suggestions[0].Type)
	}
}

func TestPermissionResult_Interface(t *testing.T) {
	// Verify both result types implement PermissionResult interface
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

func TestPermissionResultDeny_EmptyMessage(t *testing.T) {
	result := PermissionResultDeny{
		Interrupt: false,
	}

	if result.Message != "" {
		t.Errorf("Expected empty Message, got '%s'", result.Message)
	}
}

func TestPermissionUpdate_ToMap_AllUpdateTypes(t *testing.T) {
	// Test that all update types produce valid output
	updateTypes := []PermissionUpdateType{
		PermissionUpdateTypeAddRules,
		PermissionUpdateTypeReplaceRules,
		PermissionUpdateTypeRemoveRules,
		PermissionUpdateTypeSetMode,
		PermissionUpdateTypeAddDirectories,
		PermissionUpdateTypeRemoveDirectories,
	}

	for _, ut := range updateTypes {
		t.Run(string(ut), func(t *testing.T) {
			update := PermissionUpdate{Type: ut}
			result := update.ToMap()

			if result["type"] != string(ut) {
				t.Errorf("Expected type='%s', got %v", ut, result["type"])
			}
		})
	}
}

// Benchmark tests

func BenchmarkPermissionUpdate_ToMap_AddRules(b *testing.B) {
	update := PermissionUpdate{
		Type: PermissionUpdateTypeAddRules,
		Rules: []PermissionRuleValue{
			{ToolName: "Bash", RuleContent: "allow"},
			{ToolName: "Write", RuleContent: "allow .go"},
		},
		Behavior:    PermissionBehaviorAllow,
		Destination: PermissionUpdateDestinationSession,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = update.ToMap()
	}
}

func BenchmarkPermissionUpdate_ToMap_SetMode(b *testing.B) {
	update := PermissionUpdate{
		Type: PermissionUpdateTypeSetMode,
		Mode: PermissionModeDefault,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = update.ToMap()
	}
}
