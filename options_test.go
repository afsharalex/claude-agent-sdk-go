package claude

import (
	"context"
	"testing"
)

func TestNewOptions_Defaults(t *testing.T) {
	opts := NewOptions()

	if opts == nil {
		t.Fatal("Expected non-nil Options")
	}
	if opts.Env == nil {
		t.Error("Expected Env to be initialized")
	}
	if opts.ExtraArgs == nil {
		t.Error("Expected ExtraArgs to be initialized")
	}
}

func TestNewOptions_WithMultipleOptions(t *testing.T) {
	opts := NewOptions(
		WithModel("claude-3-opus"),
		WithCwd("/home/user"),
		WithMaxTurns(10),
	)

	if opts.Model != "claude-3-opus" {
		t.Errorf("Expected Model 'claude-3-opus', got '%s'", opts.Model)
	}
	if opts.Cwd != "/home/user" {
		t.Errorf("Expected Cwd '/home/user', got '%s'", opts.Cwd)
	}
	if opts.MaxTurns != 10 {
		t.Errorf("Expected MaxTurns 10, got %d", opts.MaxTurns)
	}
}

func TestWithTools(t *testing.T) {
	opts := NewOptions(WithTools([]string{"read_file", "write_file"}))

	tools, ok := opts.Tools.([]string)
	if !ok {
		t.Fatalf("Expected Tools to be []string, got %T", opts.Tools)
	}
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
}

func TestWithToolsPreset(t *testing.T) {
	preset := &ToolsPreset{Type: "preset", Preset: "claude_code"}
	opts := NewOptions(WithToolsPreset(preset))

	p, ok := opts.Tools.(*ToolsPreset)
	if !ok {
		t.Fatalf("Expected Tools to be *ToolsPreset, got %T", opts.Tools)
	}
	if p.Preset != "claude_code" {
		t.Errorf("Expected preset 'claude_code', got '%s'", p.Preset)
	}
}

func TestWithAllowedTools(t *testing.T) {
	opts := NewOptions(WithAllowedTools([]string{"Bash", "Read"}))

	if len(opts.AllowedTools) != 2 {
		t.Errorf("Expected 2 allowed tools, got %d", len(opts.AllowedTools))
	}
	if opts.AllowedTools[0] != "Bash" {
		t.Errorf("Expected first tool 'Bash', got '%s'", opts.AllowedTools[0])
	}
}

func TestWithSystemPrompt(t *testing.T) {
	opts := NewOptions(WithSystemPrompt("You are a helpful assistant"))

	prompt, ok := opts.SystemPrompt.(string)
	if !ok {
		t.Fatalf("Expected SystemPrompt to be string, got %T", opts.SystemPrompt)
	}
	if prompt != "You are a helpful assistant" {
		t.Errorf("Expected prompt to match")
	}
}

func TestWithSystemPromptPreset(t *testing.T) {
	preset := &SystemPromptPreset{Type: "preset", Preset: "claude_code", Append: "Extra"}
	opts := NewOptions(WithSystemPromptPreset(preset))

	p, ok := opts.SystemPrompt.(*SystemPromptPreset)
	if !ok {
		t.Fatalf("Expected SystemPrompt to be *SystemPromptPreset, got %T", opts.SystemPrompt)
	}
	if p.Append != "Extra" {
		t.Errorf("Expected append 'Extra', got '%s'", p.Append)
	}
}

func TestWithMCPServers(t *testing.T) {
	servers := map[string]MCPServerConfig{
		"test": MCPStdioServerConfig{Command: "npx"},
	}
	opts := NewOptions(WithMCPServers(servers))

	s, ok := opts.MCPServers.(map[string]MCPServerConfig)
	if !ok {
		t.Fatalf("Expected MCPServers to be map, got %T", opts.MCPServers)
	}
	if len(s) != 1 {
		t.Errorf("Expected 1 server, got %d", len(s))
	}
}

func TestWithMCPConfigPath(t *testing.T) {
	opts := NewOptions(WithMCPConfigPath("/path/to/config.json"))

	path, ok := opts.MCPServers.(string)
	if !ok {
		t.Fatalf("Expected MCPServers to be string, got %T", opts.MCPServers)
	}
	if path != "/path/to/config.json" {
		t.Errorf("Expected path to match")
	}
}

func TestWithPermissionMode(t *testing.T) {
	tests := []struct {
		mode     PermissionMode
		expected PermissionMode
	}{
		{PermissionModeDefault, PermissionModeDefault},
		{PermissionModeAcceptEdits, PermissionModeAcceptEdits},
		{PermissionModePlan, PermissionModePlan},
		{PermissionModeBypassPermissions, PermissionModeBypassPermissions},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			opts := NewOptions(WithPermissionMode(tt.mode))
			if opts.PermissionMode != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, opts.PermissionMode)
			}
		})
	}
}

func TestWithContinueConversation(t *testing.T) {
	opts := NewOptions(WithContinueConversation(true))

	if !opts.ContinueConversation {
		t.Error("Expected ContinueConversation to be true")
	}
}

func TestWithResume(t *testing.T) {
	opts := NewOptions(WithResume("session-123"))

	if opts.Resume != "session-123" {
		t.Errorf("Expected Resume 'session-123', got '%s'", opts.Resume)
	}
}

func TestWithMaxTurns(t *testing.T) {
	opts := NewOptions(WithMaxTurns(5))

	if opts.MaxTurns != 5 {
		t.Errorf("Expected MaxTurns 5, got %d", opts.MaxTurns)
	}
}

func TestWithMaxBudgetUSD(t *testing.T) {
	opts := NewOptions(WithMaxBudgetUSD(1.50))

	if opts.MaxBudgetUSD == nil {
		t.Fatal("Expected MaxBudgetUSD to be non-nil")
	}
	if *opts.MaxBudgetUSD != 1.50 {
		t.Errorf("Expected MaxBudgetUSD 1.50, got %f", *opts.MaxBudgetUSD)
	}
}

func TestWithDisallowedTools(t *testing.T) {
	opts := NewOptions(WithDisallowedTools([]string{"Execute", "Shell"}))

	if len(opts.DisallowedTools) != 2 {
		t.Errorf("Expected 2 disallowed tools, got %d", len(opts.DisallowedTools))
	}
}

func TestWithModel(t *testing.T) {
	opts := NewOptions(WithModel("claude-3-opus"))

	if opts.Model != "claude-3-opus" {
		t.Errorf("Expected Model 'claude-3-opus', got '%s'", opts.Model)
	}
}

func TestWithFallbackModel(t *testing.T) {
	opts := NewOptions(WithFallbackModel("claude-3-sonnet"))

	if opts.FallbackModel != "claude-3-sonnet" {
		t.Errorf("Expected FallbackModel 'claude-3-sonnet', got '%s'", opts.FallbackModel)
	}
}

func TestWithBetas(t *testing.T) {
	opts := NewOptions(WithBetas([]SdkBeta{SdkBetaContext1M}))

	if len(opts.Betas) != 1 {
		t.Fatalf("Expected 1 beta, got %d", len(opts.Betas))
	}
	if opts.Betas[0] != SdkBetaContext1M {
		t.Errorf("Expected SdkBetaContext1M")
	}
}

func TestWithPermissionPromptToolName(t *testing.T) {
	opts := NewOptions(WithPermissionPromptToolName("stdio"))

	if opts.PermissionPromptToolName != "stdio" {
		t.Errorf("Expected PermissionPromptToolName 'stdio', got '%s'", opts.PermissionPromptToolName)
	}
}

func TestWithCwd(t *testing.T) {
	opts := NewOptions(WithCwd("/home/user/project"))

	if opts.Cwd != "/home/user/project" {
		t.Errorf("Expected Cwd '/home/user/project', got '%s'", opts.Cwd)
	}
}

func TestWithCLIPath(t *testing.T) {
	opts := NewOptions(WithCLIPath("/usr/local/bin/claude"))

	if opts.CLIPath != "/usr/local/bin/claude" {
		t.Errorf("Expected CLIPath '/usr/local/bin/claude', got '%s'", opts.CLIPath)
	}
}

func TestWithSettings(t *testing.T) {
	opts := NewOptions(WithSettings(`{"key": "value"}`))

	if opts.Settings != `{"key": "value"}` {
		t.Errorf("Expected Settings to match")
	}
}

func TestWithAddDirs(t *testing.T) {
	opts := NewOptions(WithAddDirs([]string{"/dir1", "/dir2"}))

	if len(opts.AddDirs) != 2 {
		t.Errorf("Expected 2 dirs, got %d", len(opts.AddDirs))
	}
}

func TestWithEnv(t *testing.T) {
	env := map[string]string{"KEY": "VALUE"}
	opts := NewOptions(WithEnv(env))

	if opts.Env["KEY"] != "VALUE" {
		t.Errorf("Expected Env['KEY']='VALUE', got '%s'", opts.Env["KEY"])
	}
}

func TestWithExtraArg(t *testing.T) {
	value := "flagvalue"
	opts := NewOptions(
		WithExtraArg("custom-flag", &value),
		WithExtraArg("boolean-flag", nil),
	)

	if opts.ExtraArgs["custom-flag"] == nil || *opts.ExtraArgs["custom-flag"] != "flagvalue" {
		t.Error("Expected custom-flag with value")
	}
	if _, exists := opts.ExtraArgs["boolean-flag"]; !exists {
		t.Error("Expected boolean-flag to exist")
	}
}

func TestWithExtraArg_InitializesMap(t *testing.T) {
	// Start with nil ExtraArgs
	opts := &Options{ExtraArgs: nil}
	opt := WithExtraArg("flag", nil)
	opt(opts)

	if opts.ExtraArgs == nil {
		t.Error("Expected ExtraArgs to be initialized")
	}
}

func TestWithMaxBufferSize(t *testing.T) {
	opts := NewOptions(WithMaxBufferSize(4096))

	if opts.MaxBufferSize != 4096 {
		t.Errorf("Expected MaxBufferSize 4096, got %d", opts.MaxBufferSize)
	}
}

func TestWithStderr(t *testing.T) {
	called := false
	callback := func(line string) { called = true }
	opts := NewOptions(WithStderr(callback))

	if opts.Stderr == nil {
		t.Fatal("Expected Stderr callback to be set")
	}
	opts.Stderr("test")
	if !called {
		t.Error("Expected callback to be called")
	}
}

func TestWithCanUseTool(t *testing.T) {
	called := false
	callback := func(ctx context.Context, toolName string, input map[string]any, permCtx ToolPermissionContext) (PermissionResult, error) {
		called = true
		return PermissionResultAllow{}, nil
	}
	opts := NewOptions(WithCanUseTool(callback))

	if opts.CanUseTool == nil {
		t.Fatal("Expected CanUseTool callback to be set")
	}
	_, _ = opts.CanUseTool(context.Background(), "test", nil, ToolPermissionContext{})
	if !called {
		t.Error("Expected callback to be called")
	}
}

func TestWithHooks(t *testing.T) {
	hooks := map[HookEvent][]HookMatcher{
		HookEventPreToolUse: {{Matcher: "Bash"}},
	}
	opts := NewOptions(WithHooks(hooks))

	if opts.Hooks == nil {
		t.Fatal("Expected Hooks to be set")
	}
	if len(opts.Hooks[HookEventPreToolUse]) != 1 {
		t.Errorf("Expected 1 matcher, got %d", len(opts.Hooks[HookEventPreToolUse]))
	}
}

func TestWithUser(t *testing.T) {
	opts := NewOptions(WithUser("testuser"))

	if opts.User != "testuser" {
		t.Errorf("Expected User 'testuser', got '%s'", opts.User)
	}
}

func TestWithIncludePartialMessages(t *testing.T) {
	opts := NewOptions(WithIncludePartialMessages(true))

	if !opts.IncludePartialMessages {
		t.Error("Expected IncludePartialMessages to be true")
	}
}

func TestWithForkSession(t *testing.T) {
	opts := NewOptions(WithForkSession(true))

	if !opts.ForkSession {
		t.Error("Expected ForkSession to be true")
	}
}

func TestWithAgents(t *testing.T) {
	agents := map[string]AgentDefinition{
		"test-agent": {
			Description: "Test agent",
			Prompt:      "You are a test",
		},
	}
	opts := NewOptions(WithAgents(agents))

	if opts.Agents == nil {
		t.Fatal("Expected Agents to be set")
	}
	if opts.Agents["test-agent"].Description != "Test agent" {
		t.Error("Expected agent description to match")
	}
}

func TestWithSettingSources(t *testing.T) {
	sources := []SettingSource{SettingSourceUser, SettingSourceProject}
	opts := NewOptions(WithSettingSources(sources))

	if len(opts.SettingSources) != 2 {
		t.Errorf("Expected 2 setting sources, got %d", len(opts.SettingSources))
	}
}

func TestWithSandbox(t *testing.T) {
	sandbox := &SandboxSettings{
		Enabled: true,
		Network: &SandboxNetworkConfig{
			AllowLocalBinding: true,
		},
	}
	opts := NewOptions(WithSandbox(sandbox))

	if opts.Sandbox == nil {
		t.Fatal("Expected Sandbox to be set")
	}
	if !opts.Sandbox.Enabled {
		t.Error("Expected Sandbox.Enabled to be true")
	}
	if opts.Sandbox.Network == nil || !opts.Sandbox.Network.AllowLocalBinding {
		t.Error("Expected Sandbox.Network.AllowLocalBinding to be true")
	}
}

func TestWithPlugins(t *testing.T) {
	plugins := []SdkPluginConfig{
		{Type: "local", Path: "/path/to/plugin"},
	}
	opts := NewOptions(WithPlugins(plugins))

	if len(opts.Plugins) != 1 {
		t.Errorf("Expected 1 plugin, got %d", len(opts.Plugins))
	}
	if opts.Plugins[0].Path != "/path/to/plugin" {
		t.Error("Expected plugin path to match")
	}
}

func TestWithMaxThinkingTokens(t *testing.T) {
	opts := NewOptions(WithMaxThinkingTokens(1000))

	if opts.MaxThinkingTokens != 1000 {
		t.Errorf("Expected MaxThinkingTokens 1000, got %d", opts.MaxThinkingTokens)
	}
}

func TestWithOutputFormat(t *testing.T) {
	format := map[string]any{
		"type":   "json_schema",
		"schema": map[string]any{"type": "object"},
	}
	opts := NewOptions(WithOutputFormat(format))

	if opts.OutputFormat == nil {
		t.Fatal("Expected OutputFormat to be set")
	}
	if opts.OutputFormat["type"] != "json_schema" {
		t.Error("Expected type 'json_schema'")
	}
}

func TestWithEnableFileCheckpointing(t *testing.T) {
	opts := NewOptions(WithEnableFileCheckpointing(true))

	if !opts.EnableFileCheckpointing {
		t.Error("Expected EnableFileCheckpointing to be true")
	}
}

// Test option chaining and composition
func TestOptionsChaining(t *testing.T) {
	opts := NewOptions(
		WithModel("claude-3-opus"),
		WithCwd("/home/user"),
		WithMaxTurns(10),
		WithPermissionMode(PermissionModeDefault),
		WithTools([]string{"read_file"}),
		WithEnv(map[string]string{"KEY": "VALUE"}),
	)

	if opts.Model != "claude-3-opus" {
		t.Error("Expected Model to be set")
	}
	if opts.Cwd != "/home/user" {
		t.Error("Expected Cwd to be set")
	}
	if opts.MaxTurns != 10 {
		t.Error("Expected MaxTurns to be set")
	}
	if opts.PermissionMode != PermissionModeDefault {
		t.Error("Expected PermissionMode to be set")
	}
	if opts.Tools == nil {
		t.Error("Expected Tools to be set")
	}
	if opts.Env["KEY"] != "VALUE" {
		t.Error("Expected Env to be set")
	}
}

// Benchmark tests

func BenchmarkNewOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewOptions()
	}
}

func BenchmarkNewOptions_WithOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewOptions(
			WithModel("claude-3-opus"),
			WithCwd("/home/user"),
			WithMaxTurns(10),
		)
	}
}

// Tests for new convenience options

func TestWithAppendSystemPrompt(t *testing.T) {
	t.Run("sets prompt when none exists", func(t *testing.T) {
		opts := NewOptions(WithAppendSystemPrompt("You are helpful"))

		if opts.AppendSystemPrompt != "You are helpful" {
			t.Errorf("Expected AppendSystemPrompt 'You are helpful', got '%s'", opts.AppendSystemPrompt)
		}
	})

	t.Run("appends to existing append prompt", func(t *testing.T) {
		opts := NewOptions(
			WithAppendSystemPrompt("First line"),
			WithAppendSystemPrompt("Second line"),
		)

		expected := "First line\nSecond line"
		if opts.AppendSystemPrompt != expected {
			t.Errorf("Expected AppendSystemPrompt '%s', got '%s'", expected, opts.AppendSystemPrompt)
		}
	})

	t.Run("multiple appends with newlines", func(t *testing.T) {
		opts := NewOptions(
			WithAppendSystemPrompt("Line 1"),
			WithAppendSystemPrompt("Line 2"),
			WithAppendSystemPrompt("Line 3"),
		)

		expected := "Line 1\nLine 2\nLine 3"
		if opts.AppendSystemPrompt != expected {
			t.Errorf("Expected AppendSystemPrompt '%s', got '%s'", expected, opts.AppendSystemPrompt)
		}
	})
}

func TestWithAppendSystemPrompt_CombinesWithSystemPrompt(t *testing.T) {
	opts := NewOptions(
		WithSystemPrompt("Base prompt"),
		WithAppendSystemPrompt("Additional instructions"),
	)

	if opts.SystemPrompt != "Base prompt" {
		t.Errorf("Expected SystemPrompt 'Base prompt', got '%v'", opts.SystemPrompt)
	}
	if opts.AppendSystemPrompt != "Additional instructions" {
		t.Errorf("Expected AppendSystemPrompt 'Additional instructions', got '%s'", opts.AppendSystemPrompt)
	}
}

func TestWithDebugStderr(t *testing.T) {
	opts := NewOptions(WithDebugStderr())

	if opts.Stderr == nil {
		t.Fatal("Expected Stderr callback to be set")
	}

	// The callback should not panic when called
	opts.Stderr("test output")
}

func TestWithDebugStderr_OverridesExisting(t *testing.T) {
	customCalled := false
	customCallback := func(line string) { customCalled = true }

	opts := NewOptions(
		WithStderr(customCallback),
		WithDebugStderr(),
	)

	// Call the callback - it should be the debug stderr, not the custom one
	opts.Stderr("test")

	if customCalled {
		t.Error("Expected custom callback to be overridden by WithDebugStderr")
	}
}

func TestWithEnvVar(t *testing.T) {
	t.Run("sets single env var", func(t *testing.T) {
		opts := NewOptions(WithEnvVar("MY_KEY", "my_value"))

		if opts.Env["MY_KEY"] != "my_value" {
			t.Errorf("Expected Env['MY_KEY']='my_value', got '%s'", opts.Env["MY_KEY"])
		}
	})

	t.Run("sets multiple env vars", func(t *testing.T) {
		opts := NewOptions(
			WithEnvVar("KEY1", "value1"),
			WithEnvVar("KEY2", "value2"),
			WithEnvVar("KEY3", "value3"),
		)

		if opts.Env["KEY1"] != "value1" {
			t.Errorf("Expected Env['KEY1']='value1', got '%s'", opts.Env["KEY1"])
		}
		if opts.Env["KEY2"] != "value2" {
			t.Errorf("Expected Env['KEY2']='value2', got '%s'", opts.Env["KEY2"])
		}
		if opts.Env["KEY3"] != "value3" {
			t.Errorf("Expected Env['KEY3']='value3', got '%s'", opts.Env["KEY3"])
		}
	})

	t.Run("overwrites existing key", func(t *testing.T) {
		opts := NewOptions(
			WithEnvVar("KEY", "original"),
			WithEnvVar("KEY", "updated"),
		)

		if opts.Env["KEY"] != "updated" {
			t.Errorf("Expected Env['KEY']='updated', got '%s'", opts.Env["KEY"])
		}
	})

	t.Run("initializes nil Env map", func(t *testing.T) {
		opts := &Options{Env: nil}
		opt := WithEnvVar("KEY", "value")
		opt(opts)

		if opts.Env == nil {
			t.Error("Expected Env to be initialized")
		}
		if opts.Env["KEY"] != "value" {
			t.Errorf("Expected Env['KEY']='value', got '%s'", opts.Env["KEY"])
		}
	})
}

func TestWithEnvVar_CombinesWithWithEnv(t *testing.T) {
	opts := NewOptions(
		WithEnv(map[string]string{"EXISTING": "value"}),
		WithEnvVar("NEW", "new_value"),
	)

	// WithEnv replaces the map, so EXISTING should exist
	if opts.Env["EXISTING"] != "value" {
		t.Errorf("Expected Env['EXISTING']='value', got '%s'", opts.Env["EXISTING"])
	}
	if opts.Env["NEW"] != "new_value" {
		t.Errorf("Expected Env['NEW']='new_value', got '%s'", opts.Env["NEW"])
	}
}
