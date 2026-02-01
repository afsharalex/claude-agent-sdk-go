package claude

import (
	"context"
	"testing"

	"github.com/afsharalex/claude-agent-sdk-go/internal/types"
)

func TestToTransportOptions_BasicFields(t *testing.T) {
	opts := &Options{
		Tools:        []string{"read_file", "write_file"},
		AllowedTools: []string{"read_file"},
		SystemPrompt: "You are a helpful assistant",
		Model:        "claude-3-opus",
		MaxTurns:     10,
		Cwd:          "/home/user/project",
	}

	result := toTransportOptions(opts)

	tools, ok := result.Tools.([]string)
	if !ok {
		t.Fatalf("Expected Tools to be []string, got %T", result.Tools)
	}
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
	if len(result.AllowedTools) != 1 {
		t.Errorf("Expected 1 allowed tool, got %d", len(result.AllowedTools))
	}
	if result.SystemPrompt != "You are a helpful assistant" {
		t.Errorf("Expected system prompt to match")
	}
	if result.Model != "claude-3-opus" {
		t.Errorf("Expected model 'claude-3-opus', got '%s'", result.Model)
	}
	if result.MaxTurns != 10 {
		t.Errorf("Expected MaxTurns 10, got %d", result.MaxTurns)
	}
	if result.Cwd != "/home/user/project" {
		t.Errorf("Expected Cwd '/home/user/project', got '%s'", result.Cwd)
	}
}

func TestToTransportOptions_Betas(t *testing.T) {
	opts := &Options{
		Betas: []SdkBeta{SdkBetaContext1M},
	}

	result := toTransportOptions(opts)

	if len(result.Betas) != 1 {
		t.Fatalf("Expected 1 beta, got %d", len(result.Betas))
	}
	if result.Betas[0] != "context-1m-2025-08-07" {
		t.Errorf("Expected beta 'context-1m-2025-08-07', got '%s'", result.Betas[0])
	}
}

func TestToTransportOptions_SettingSources(t *testing.T) {
	opts := &Options{
		SettingSources: []SettingSource{SettingSourceUser, SettingSourceProject},
	}

	result := toTransportOptions(opts)

	if len(result.SettingSources) != 2 {
		t.Fatalf("Expected 2 setting sources, got %d", len(result.SettingSources))
	}
	if result.SettingSources[0] != "user" {
		t.Errorf("Expected 'user', got '%s'", result.SettingSources[0])
	}
	if result.SettingSources[1] != "project" {
		t.Errorf("Expected 'project', got '%s'", result.SettingSources[1])
	}
}

func TestToTransportOptions_Sandbox(t *testing.T) {
	opts := &Options{
		Sandbox: &SandboxSettings{
			Enabled:                   true,
			AutoAllowBashIfSandboxed:  true,
			ExcludedCommands:          []string{"docker"},
			AllowUnsandboxedCommands:  false,
			EnableWeakerNestedSandbox: true,
			Network: &SandboxNetworkConfig{
				AllowUnixSockets:    []string{"/var/run/docker.sock"},
				AllowAllUnixSockets: false,
				AllowLocalBinding:   true,
				HTTPProxyPort:       8080,
				SOCKSProxyPort:      1080,
			},
			IgnoreViolations: &SandboxIgnoreViolations{
				File:    []string{"/tmp"},
				Network: []string{"localhost"},
			},
		},
	}

	result := toTransportOptions(opts)

	if result.Sandbox == nil {
		t.Fatal("Expected Sandbox to be non-nil")
	}
	if !result.Sandbox.Enabled {
		t.Error("Expected Sandbox.Enabled to be true")
	}
	if !result.Sandbox.AutoAllowBashIfSandboxed {
		t.Error("Expected Sandbox.AutoAllowBashIfSandboxed to be true")
	}
	if len(result.Sandbox.ExcludedCommands) != 1 {
		t.Errorf("Expected 1 excluded command, got %d", len(result.Sandbox.ExcludedCommands))
	}
	if !result.Sandbox.EnableWeakerNestedSandbox {
		t.Error("Expected EnableWeakerNestedSandbox to be true")
	}

	// Network config
	if result.Sandbox.Network == nil {
		t.Fatal("Expected Sandbox.Network to be non-nil")
	}
	if len(result.Sandbox.Network.AllowUnixSockets) != 1 {
		t.Errorf("Expected 1 unix socket, got %d", len(result.Sandbox.Network.AllowUnixSockets))
	}
	if result.Sandbox.Network.HTTPProxyPort != 8080 {
		t.Errorf("Expected HTTPProxyPort 8080, got %d", result.Sandbox.Network.HTTPProxyPort)
	}

	// Ignore violations
	if result.Sandbox.IgnoreViolations == nil {
		t.Fatal("Expected Sandbox.IgnoreViolations to be non-nil")
	}
	if len(result.Sandbox.IgnoreViolations.File) != 1 {
		t.Errorf("Expected 1 file violation, got %d", len(result.Sandbox.IgnoreViolations.File))
	}
}

func TestToTransportOptions_SandboxNilSubstructs(t *testing.T) {
	opts := &Options{
		Sandbox: &SandboxSettings{
			Enabled: true,
		},
	}

	result := toTransportOptions(opts)

	if result.Sandbox.Network != nil {
		t.Error("Expected Sandbox.Network to be nil")
	}
	if result.Sandbox.IgnoreViolations != nil {
		t.Error("Expected Sandbox.IgnoreViolations to be nil")
	}
}

func TestToTransportOptions_Agents(t *testing.T) {
	opts := &Options{
		Agents: map[string]AgentDefinition{
			"test-agent": {
				Description: "A test agent",
				Prompt:      "You are a test agent",
				Tools:       []string{"read_file"},
				Model:       "sonnet",
			},
		},
	}

	result := toTransportOptions(opts)

	if len(result.Agents) != 1 {
		t.Fatalf("Expected 1 agent, got %d", len(result.Agents))
	}
	agent := result.Agents["test-agent"]
	if agent.Description != "A test agent" {
		t.Errorf("Expected description 'A test agent', got '%s'", agent.Description)
	}
	if agent.Prompt != "You are a test agent" {
		t.Errorf("Expected prompt to match")
	}
}

func TestToTransportOptions_Plugins(t *testing.T) {
	opts := &Options{
		Plugins: []SdkPluginConfig{
			{Type: "local", Path: "/path/to/plugin"},
		},
	}

	result := toTransportOptions(opts)

	if len(result.Plugins) != 1 {
		t.Fatalf("Expected 1 plugin, got %d", len(result.Plugins))
	}
	if result.Plugins[0].Type != "local" {
		t.Errorf("Expected type 'local', got '%s'", result.Plugins[0].Type)
	}
	if result.Plugins[0].Path != "/path/to/plugin" {
		t.Errorf("Expected path '/path/to/plugin', got '%s'", result.Plugins[0].Path)
	}
}

func TestToTransportOptions_MCPServers_Map(t *testing.T) {
	opts := &Options{
		MCPServers: map[string]MCPServerConfig{
			"stdio-server": MCPStdioServerConfig{
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
				Env:     map[string]string{"PATH": "/usr/bin"},
			},
			"sse-server": MCPSSEServerConfig{
				URL:     "https://example.com/sse",
				Headers: map[string]string{"Authorization": "Bearer token"},
			},
			"http-server": MCPHTTPServerConfig{
				URL: "https://example.com/api",
			},
			"sdk-server": MCPSDKServerConfig{
				Name: "my-sdk-server",
			},
		},
	}

	result := toTransportOptions(opts)

	servers, ok := result.MCPServers.(map[string]any)
	if !ok {
		t.Fatalf("Expected MCPServers to be map[string]any, got %T", result.MCPServers)
	}

	// Stdio server
	stdio, ok := servers["stdio-server"].(map[string]any)
	if !ok {
		t.Fatal("Expected stdio-server to be map[string]any")
	}
	if stdio["type"] != "stdio" {
		t.Errorf("Expected type 'stdio', got '%v'", stdio["type"])
	}
	if stdio["command"] != "npx" {
		t.Errorf("Expected command 'npx', got '%v'", stdio["command"])
	}

	// SSE server
	sse, ok := servers["sse-server"].(map[string]any)
	if !ok {
		t.Fatal("Expected sse-server to be map[string]any")
	}
	if sse["type"] != "sse" {
		t.Errorf("Expected type 'sse', got '%v'", sse["type"])
	}
	if sse["url"] != "https://example.com/sse" {
		t.Errorf("Expected URL to match")
	}

	// HTTP server
	http, ok := servers["http-server"].(map[string]any)
	if !ok {
		t.Fatal("Expected http-server to be map[string]any")
	}
	if http["type"] != "http" {
		t.Errorf("Expected type 'http', got '%v'", http["type"])
	}

	// SDK server
	sdk, ok := servers["sdk-server"].(map[string]any)
	if !ok {
		t.Fatal("Expected sdk-server to be map[string]any")
	}
	if sdk["type"] != "sdk" {
		t.Errorf("Expected type 'sdk', got '%v'", sdk["type"])
	}
}

func TestToTransportOptions_MCPServers_StringPath(t *testing.T) {
	opts := &Options{
		MCPServers: "/path/to/mcp-config.json",
	}

	result := toTransportOptions(opts)

	if result.MCPServers != "/path/to/mcp-config.json" {
		t.Errorf("Expected MCPServers to be path string, got %v", result.MCPServers)
	}
}

func TestToTransportOptions_PermissionMode(t *testing.T) {
	opts := &Options{
		PermissionMode: PermissionModeBypassPermissions,
	}

	result := toTransportOptions(opts)

	if result.PermissionMode != "bypassPermissions" {
		t.Errorf("Expected 'bypassPermissions', got '%s'", result.PermissionMode)
	}
}

func TestToInternalHooks_Nil(t *testing.T) {
	result := toInternalHooks(nil)
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestToInternalHooks_Basic(t *testing.T) {
	called := false
	hooks := map[HookEvent][]HookMatcher{
		HookEventPreToolUse: {
			{
				Matcher: "Bash",
				Hooks: []HookCallback{
					func(ctx context.Context, input HookInput, toolUseID string, hookCtx HookContext) (HookOutput, error) {
						called = true
						return HookOutput{}, nil
					},
				},
				Timeout: 30.0,
			},
		},
	}

	result := toInternalHooks(hooks)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if len(result) != 1 {
		t.Fatalf("Expected 1 hook event, got %d", len(result))
	}

	matchers, ok := result[types.HookEventPreToolUse]
	if !ok {
		t.Fatal("Expected PreToolUse matchers")
	}
	if len(matchers) != 1 {
		t.Fatalf("Expected 1 matcher, got %d", len(matchers))
	}
	if matchers[0].Matcher != "Bash" {
		t.Errorf("Expected matcher 'Bash', got '%s'", matchers[0].Matcher)
	}
	if matchers[0].Timeout != 30.0 {
		t.Errorf("Expected timeout 30.0, got %f", matchers[0].Timeout)
	}

	// Test that the wrapped callback works
	internalInput := types.PreToolUseHookInput{
		BaseHookInput: types.BaseHookInput{SessionID: "test"},
		ToolName:      "Bash",
	}
	_, err := matchers[0].Hooks[0](context.Background(), internalInput, "tool-123", types.HookContext{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !called {
		t.Error("Expected callback to be called")
	}
}

func TestToPublicHookInput_PreToolUse(t *testing.T) {
	internal := types.PreToolUseHookInput{
		BaseHookInput: types.BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/path/to/transcript",
			Cwd:            "/home/user",
			PermissionMode: "default",
		},
		ToolName:  "Bash",
		ToolInput: map[string]any{"command": "ls"},
	}

	result := toPublicHookInput(internal)

	public, ok := result.(PreToolUseHookInput)
	if !ok {
		t.Fatalf("Expected PreToolUseHookInput, got %T", result)
	}
	if public.SessionID != "session-123" {
		t.Errorf("Expected session ID 'session-123', got '%s'", public.SessionID)
	}
	if public.ToolName != "Bash" {
		t.Errorf("Expected tool name 'Bash', got '%s'", public.ToolName)
	}
	if public.ToolInput["command"] != "ls" {
		t.Errorf("Expected command 'ls', got '%v'", public.ToolInput["command"])
	}
}

func TestToPublicHookInput_PostToolUse(t *testing.T) {
	internal := types.PostToolUseHookInput{
		BaseHookInput: types.BaseHookInput{SessionID: "session-456"},
		ToolName:      "Write",
		ToolInput:     map[string]any{"path": "/tmp/file.txt"},
		ToolResponse:  "File written successfully",
	}

	result := toPublicHookInput(internal)

	public, ok := result.(PostToolUseHookInput)
	if !ok {
		t.Fatalf("Expected PostToolUseHookInput, got %T", result)
	}
	if public.ToolName != "Write" {
		t.Errorf("Expected tool name 'Write', got '%s'", public.ToolName)
	}
	if public.ToolResponse != "File written successfully" {
		t.Errorf("Expected tool response to match")
	}
}

func TestToPublicHookInput_UserPromptSubmit(t *testing.T) {
	internal := types.UserPromptSubmitHookInput{
		BaseHookInput: types.BaseHookInput{SessionID: "session-789"},
		Prompt:        "Hello, Claude!",
	}

	result := toPublicHookInput(internal)

	public, ok := result.(UserPromptSubmitHookInput)
	if !ok {
		t.Fatalf("Expected UserPromptSubmitHookInput, got %T", result)
	}
	if public.Prompt != "Hello, Claude!" {
		t.Errorf("Expected prompt 'Hello, Claude!', got '%s'", public.Prompt)
	}
}

func TestToPublicHookInput_Stop(t *testing.T) {
	internal := types.StopHookInput{
		BaseHookInput:  types.BaseHookInput{SessionID: "session-stop"},
		StopHookActive: true,
	}

	result := toPublicHookInput(internal)

	public, ok := result.(StopHookInput)
	if !ok {
		t.Fatalf("Expected StopHookInput, got %T", result)
	}
	if !public.StopHookActive {
		t.Error("Expected StopHookActive to be true")
	}
}

func TestToPublicHookInput_Unknown(t *testing.T) {
	// Use a type that's not handled
	internal := types.SubagentStopHookInput{
		BaseHookInput: types.BaseHookInput{SessionID: "test"},
	}

	result := toPublicHookInput(internal)

	// Current implementation returns nil for unhandled types
	if result != nil {
		t.Errorf("Expected nil for unhandled type, got %T", result)
	}
}

func TestToInternalHookOutput_Basic(t *testing.T) {
	cont := true
	output := HookOutput{
		Async:          false,
		AsyncTimeout:   5000,
		Continue:       &cont,
		SuppressOutput: true,
		StopReason:     "test stop",
		Decision:       HookDecisionBlock,
		SystemMessage:  "System warning",
		Reason:         "Block reason",
	}

	result := toInternalHookOutput(output)

	if result.Async != false {
		t.Error("Expected Async to be false")
	}
	if result.AsyncTimeout != 5000 {
		t.Errorf("Expected AsyncTimeout 5000, got %d", result.AsyncTimeout)
	}
	if result.Continue == nil || !*result.Continue {
		t.Error("Expected Continue to be true")
	}
	if !result.SuppressOutput {
		t.Error("Expected SuppressOutput to be true")
	}
	if result.StopReason != "test stop" {
		t.Errorf("Expected StopReason 'test stop', got '%s'", result.StopReason)
	}
	if result.Decision != types.HookDecisionBlock {
		t.Errorf("Expected Decision 'block', got '%s'", result.Decision)
	}
}

func TestToInternalHookOutput_PreToolUseSpecific(t *testing.T) {
	output := HookOutput{
		HookSpecificOutput: PreToolUseHookSpecificOutput{
			HookEventName:      HookEventPreToolUse,
			PermissionDecision: HookPermissionDecisionAllow,
			UpdatedInput:       map[string]any{"modified": true},
		},
	}

	result := toInternalHookOutput(output)

	if result.HookEventName != types.HookEventPreToolUse {
		t.Errorf("Expected HookEventName 'PreToolUse', got '%s'", result.HookEventName)
	}
	if result.PermissionDecision != types.HookPermissionDecisionAllow {
		t.Errorf("Expected PermissionDecision 'allow', got '%s'", result.PermissionDecision)
	}
	if result.UpdatedInput["modified"] != true {
		t.Errorf("Expected UpdatedInput['modified'] to be true")
	}
}

func TestToInternalHookOutput_PostToolUseSpecific(t *testing.T) {
	output := HookOutput{
		HookSpecificOutput: PostToolUseHookSpecificOutput{
			HookEventName:     HookEventPostToolUse,
			AdditionalContext: "Extra context",
		},
	}

	result := toInternalHookOutput(output)

	if result.HookEventName != types.HookEventPostToolUse {
		t.Errorf("Expected HookEventName 'PostToolUse', got '%s'", result.HookEventName)
	}
	if result.AdditionalContext != "Extra context" {
		t.Errorf("Expected AdditionalContext 'Extra context', got '%s'", result.AdditionalContext)
	}
}

func TestToInternalCanUseTool_Nil(t *testing.T) {
	result := toInternalCanUseTool(nil)
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestToInternalCanUseTool_Allow(t *testing.T) {
	fn := func(ctx context.Context, toolName string, input map[string]any, permCtx ToolPermissionContext) (PermissionResult, error) {
		return PermissionResultAllow{
			UpdatedInput: map[string]any{"sanitized": true},
		}, nil
	}

	internal := toInternalCanUseTool(fn)
	if internal == nil {
		t.Fatal("Expected non-nil internal function")
	}

	result, err := internal(context.Background(), "Bash", map[string]any{}, types.ToolPermissionContext{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	allow, ok := result.(types.PermissionResultAllow)
	if !ok {
		t.Fatalf("Expected PermissionResultAllow, got %T", result)
	}
	if allow.UpdatedInput["sanitized"] != true {
		t.Errorf("Expected UpdatedInput['sanitized'] to be true")
	}
}

func TestToInternalCanUseTool_Deny(t *testing.T) {
	fn := func(ctx context.Context, toolName string, input map[string]any, permCtx ToolPermissionContext) (PermissionResult, error) {
		return PermissionResultDeny{
			Message:   "Operation not allowed",
			Interrupt: true,
		}, nil
	}

	internal := toInternalCanUseTool(fn)
	result, err := internal(context.Background(), "Bash", map[string]any{}, types.ToolPermissionContext{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	deny, ok := result.(types.PermissionResultDeny)
	if !ok {
		t.Fatalf("Expected PermissionResultDeny, got %T", result)
	}
	if deny.Message != "Operation not allowed" {
		t.Errorf("Expected message 'Operation not allowed', got '%s'", deny.Message)
	}
	if !deny.Interrupt {
		t.Error("Expected Interrupt to be true")
	}
}

func TestToInternalMCPServers_Nil(t *testing.T) {
	result := toInternalMCPServers(nil)
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestToInternalMCPServers_NonSDK(t *testing.T) {
	servers := map[string]MCPServerConfig{
		"stdio": MCPStdioServerConfig{Command: "npx"},
	}

	result := toInternalMCPServers(servers)

	// Non-SDK servers should not be included
	if len(result) != 0 {
		t.Errorf("Expected 0 servers (non-SDK should be filtered), got %d", len(result))
	}
}

func TestToInternalMCPServers_SDK(t *testing.T) {
	handler := func(ctx context.Context, args map[string]any) (MCPToolResult, error) {
		return TextResult("test result"), nil
	}

	server := NewMCPServer("test-server", "1.0.0", []MCPTool{
		{
			Name:        "test-tool",
			Description: "A test tool",
			InputSchema: SimpleInputSchema(map[string]string{"arg": "string"}),
			Handler:     handler,
		},
	})

	servers := map[string]MCPServerConfig{
		"sdk-server": MCPSDKServerConfig{
			Name:   "test-server",
			Server: server,
		},
	}

	result := toInternalMCPServers(servers)

	if len(result) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(result))
	}

	srv := result["sdk-server"]
	if srv == nil {
		t.Fatal("Expected sdk-server to be present")
	}
	if srv.Name != "test-server" {
		t.Errorf("Expected name 'test-server', got '%s'", srv.Name)
	}
	if srv.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", srv.Version)
	}
	if len(srv.Tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(srv.Tools))
	}

	// Test the wrapped handler
	toolResult, err := srv.Tools[0].Handler(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(toolResult.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(toolResult.Content))
	}
	if toolResult.Content[0].Text != "test result" {
		t.Errorf("Expected text 'test result', got '%s'", toolResult.Content[0].Text)
	}
}

func TestToInternalMCPServers_SDKNilServer(t *testing.T) {
	servers := map[string]MCPServerConfig{
		"sdk-server": MCPSDKServerConfig{
			Name:   "test-server",
			Server: nil, // nil server should be filtered
		},
	}

	result := toInternalMCPServers(servers)

	if len(result) != 0 {
		t.Errorf("Expected 0 servers (nil Server should be filtered), got %d", len(result))
	}
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Expected Version to be non-empty")
	}
}

func TestToPublicHookInput_Unknown_ReturnsNil(t *testing.T) {
	// Types not handled by toPublicHookInput should return nil
	internal := types.SubagentStopHookInput{
		BaseHookInput:  types.BaseHookInput{SessionID: "session-subagent"},
		StopHookActive: true,
	}

	result := toPublicHookInput(internal)

	if result != nil {
		t.Errorf("Expected nil for unhandled type, got %T", result)
	}
}

func TestToInternalCanUseTool_WithContext(t *testing.T) {
	fn := func(ctx context.Context, toolName string, input map[string]any, permCtx ToolPermissionContext) (PermissionResult, error) {
		// Return Allow as it's the most permissive default
		return PermissionResultAllow{}, nil
	}

	internal := toInternalCanUseTool(fn)
	result, err := internal(context.Background(), "Test", map[string]any{}, types.ToolPermissionContext{
		Signal:      nil,
		Suggestions: []types.PermissionUpdate{},
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
}

func TestToInternalHooks_AllEvents(t *testing.T) {
	hooks := map[HookEvent][]HookMatcher{
		HookEventPreToolUse:        {{Matcher: "*"}},
		HookEventPostToolUse:       {{Matcher: "*"}},
		HookEventPostToolUseFailed: {{Matcher: "*"}},
		HookEventUserPromptSubmit:  {{Matcher: "*"}},
		HookEventStop:              {{Matcher: "*"}},
		HookEventSubagentStop:      {{Matcher: "*"}},
		HookEventPreCompact:        {{Matcher: "*"}},
	}

	result := toInternalHooks(hooks)

	if len(result) != 7 {
		t.Errorf("Expected 7 hook events, got %d", len(result))
	}
}

func TestToInternalHooks_CallbackError(t *testing.T) {
	expectedErr := NewClaudeSDKError("hook error")
	hooks := map[HookEvent][]HookMatcher{
		HookEventPreToolUse: {
			{
				Matcher: "Bash",
				Hooks: []HookCallback{
					func(ctx context.Context, input HookInput, toolUseID string, hookCtx HookContext) (HookOutput, error) {
						return HookOutput{}, expectedErr
					},
				},
			},
		},
	}

	result := toInternalHooks(hooks)

	// Call the wrapped callback to test error handling
	internalInput := types.PreToolUseHookInput{
		BaseHookInput: types.BaseHookInput{SessionID: "test"},
		ToolName:      "Bash",
	}

	output, err := result[types.HookEventPreToolUse][0].Hooks[0](context.Background(), internalInput, "tool-123", types.HookContext{})

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	// Output should be empty when there's an error
	if output.Continue != nil || output.StopReason != "" {
		t.Error("Expected empty output on error")
	}
}

func TestToTransportOptions_EmptyOptions(t *testing.T) {
	opts := &Options{}

	result := toTransportOptions(opts)

	if result.Cwd != "" {
		t.Errorf("Expected empty Cwd, got '%s'", result.Cwd)
	}
	if result.Model != "" {
		t.Errorf("Expected empty Model, got '%s'", result.Model)
	}
	if result.MaxTurns != 0 {
		t.Errorf("Expected MaxTurns 0, got %d", result.MaxTurns)
	}
}

func TestToInternalCanUseTool_Error(t *testing.T) {
	expectedErr := NewClaudeSDKError("test error")
	fn := func(ctx context.Context, toolName string, input map[string]any, permCtx ToolPermissionContext) (PermissionResult, error) {
		return nil, expectedErr
	}

	internal := toInternalCanUseTool(fn)
	result, err := internal(context.Background(), "Test", map[string]any{}, types.ToolPermissionContext{})

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if result != nil {
		t.Errorf("Expected nil result on error, got %v", result)
	}
}

// mockPermissionResult is a custom type that doesn't implement the expected cases
type mockPermissionResult struct{}

func (mockPermissionResult) permissionResult() {}

func TestToInternalCanUseTool_InvalidResultType(t *testing.T) {
	fn := func(ctx context.Context, toolName string, input map[string]any, permCtx ToolPermissionContext) (PermissionResult, error) {
		return mockPermissionResult{}, nil
	}

	internal := toInternalCanUseTool(fn)
	result, err := internal(context.Background(), "Test", map[string]any{}, types.ToolPermissionContext{})

	if err == nil {
		t.Fatal("Expected error for invalid result type")
	}
	if result != nil {
		t.Errorf("Expected nil result, got %v", result)
	}
	sdkErr, ok := err.(*ClaudeSDKError)
	if !ok {
		t.Fatalf("Expected *ClaudeSDKError, got %T", err)
	}
	if sdkErr.Message != "invalid permission result type" {
		t.Errorf("Expected error message 'invalid permission result type', got '%s'", sdkErr.Message)
	}
}

func TestToTransportOptions_OutputFormat(t *testing.T) {
	opts := &Options{
		OutputFormat: map[string]any{"type": "json"},
	}

	result := toTransportOptions(opts)

	if result.OutputFormat == nil {
		t.Fatal("Expected OutputFormat to be non-nil")
	}
	if result.OutputFormat["type"] != "json" {
		t.Errorf("Expected OutputFormat['type']='json', got %v", result.OutputFormat["type"])
	}
}

func TestToTransportOptions_MaxThinkingTokens(t *testing.T) {
	opts := &Options{
		MaxThinkingTokens: 1000,
	}

	result := toTransportOptions(opts)

	if result.MaxThinkingTokens != 1000 {
		t.Errorf("Expected MaxThinkingTokens 1000, got %d", result.MaxThinkingTokens)
	}
}

func TestToTransportOptions_EnableFileCheckpointing(t *testing.T) {
	opts := &Options{
		EnableFileCheckpointing: true,
	}

	result := toTransportOptions(opts)

	if !result.EnableFileCheckpointing {
		t.Error("Expected EnableFileCheckpointing to be true")
	}
}

func TestToTransportOptions_User(t *testing.T) {
	opts := &Options{
		User: "test-user",
	}

	result := toTransportOptions(opts)

	if result.User != "test-user" {
		t.Errorf("Expected User 'test-user', got '%s'", result.User)
	}
}

func TestToTransportOptions_IncludePartialMessages(t *testing.T) {
	opts := &Options{
		IncludePartialMessages: true,
	}

	result := toTransportOptions(opts)

	if !result.IncludePartialMessages {
		t.Error("Expected IncludePartialMessages to be true")
	}
}

func TestToTransportOptions_ForkSession(t *testing.T) {
	opts := &Options{
		ForkSession: true,
	}

	result := toTransportOptions(opts)

	if !result.ForkSession {
		t.Error("Expected ForkSession to be true")
	}
}

func TestToTransportOptions_PermissionPromptToolName(t *testing.T) {
	opts := &Options{
		PermissionPromptToolName: "custom-tool",
	}

	result := toTransportOptions(opts)

	if result.PermissionPromptToolName != "custom-tool" {
		t.Errorf("Expected PermissionPromptToolName 'custom-tool', got '%s'", result.PermissionPromptToolName)
	}
}

func TestToTransportOptions_CLIPath(t *testing.T) {
	opts := &Options{
		CLIPath: "/custom/path/claude",
	}

	result := toTransportOptions(opts)

	if result.CLIPath != "/custom/path/claude" {
		t.Errorf("Expected CLIPath '/custom/path/claude', got '%s'", result.CLIPath)
	}
}

func TestToTransportOptions_Settings(t *testing.T) {
	opts := &Options{
		Settings: "/path/to/settings.json",
	}

	result := toTransportOptions(opts)

	if result.Settings != "/path/to/settings.json" {
		t.Errorf("Expected Settings '/path/to/settings.json', got '%s'", result.Settings)
	}
}

func TestToTransportOptions_AddDirs(t *testing.T) {
	dirs := []string{"/home/user", "/tmp"}
	opts := &Options{
		AddDirs: dirs,
	}

	result := toTransportOptions(opts)

	if len(result.AddDirs) != 2 {
		t.Fatalf("Expected 2 add dirs, got %d", len(result.AddDirs))
	}
	if result.AddDirs[0] != "/home/user" {
		t.Errorf("Expected first dir '/home/user', got '%s'", result.AddDirs[0])
	}
}

func TestToTransportOptions_Env(t *testing.T) {
	env := map[string]string{"FOO": "bar"}
	opts := &Options{
		Env: env,
	}

	result := toTransportOptions(opts)

	if result.Env == nil {
		t.Fatal("Expected Env to be non-nil")
	}
	if result.Env["FOO"] != "bar" {
		t.Errorf("Expected Env['FOO']='bar', got '%s'", result.Env["FOO"])
	}
}

func TestToTransportOptions_ExtraArgs(t *testing.T) {
	val := "true"
	opts := &Options{
		ExtraArgs: map[string]*string{"verbose": &val},
	}

	result := toTransportOptions(opts)

	if result.ExtraArgs == nil {
		t.Fatal("Expected ExtraArgs to be non-nil")
	}
	if result.ExtraArgs["verbose"] == nil || *result.ExtraArgs["verbose"] != "true" {
		t.Errorf("Expected ExtraArgs['verbose']='true', got %v", result.ExtraArgs["verbose"])
	}
}

func TestToTransportOptions_MaxBufferSize(t *testing.T) {
	opts := &Options{
		MaxBufferSize: 1024 * 1024,
	}

	result := toTransportOptions(opts)

	if result.MaxBufferSize != 1024*1024 {
		t.Errorf("Expected MaxBufferSize 1048576, got %d", result.MaxBufferSize)
	}
}

func TestToTransportOptions_FallbackModel(t *testing.T) {
	opts := &Options{
		FallbackModel: "claude-3-sonnet",
	}

	result := toTransportOptions(opts)

	if result.FallbackModel != "claude-3-sonnet" {
		t.Errorf("Expected FallbackModel 'claude-3-sonnet', got '%s'", result.FallbackModel)
	}
}

func TestToTransportOptions_Resume(t *testing.T) {
	opts := &Options{
		Resume: "session-abc",
	}

	result := toTransportOptions(opts)

	if result.Resume != "session-abc" {
		t.Errorf("Expected Resume 'session-abc', got '%s'", result.Resume)
	}
}

func TestToTransportOptions_ContinueConversation(t *testing.T) {
	opts := &Options{
		ContinueConversation: true,
	}

	result := toTransportOptions(opts)

	if !result.ContinueConversation {
		t.Error("Expected ContinueConversation to be true")
	}
}

func TestToTransportOptions_MaxBudgetUSD(t *testing.T) {
	budget := 10.0
	opts := &Options{
		MaxBudgetUSD: &budget,
	}

	result := toTransportOptions(opts)

	if result.MaxBudgetUSD == nil || *result.MaxBudgetUSD != 10.0 {
		t.Errorf("Expected MaxBudgetUSD 10.0, got %v", result.MaxBudgetUSD)
	}
}

func TestToTransportOptions_DisallowedTools(t *testing.T) {
	opts := &Options{
		DisallowedTools: []string{"Bash"},
	}

	result := toTransportOptions(opts)

	if len(result.DisallowedTools) != 1 {
		t.Fatalf("Expected 1 disallowed tool, got %d", len(result.DisallowedTools))
	}
	if result.DisallowedTools[0] != "Bash" {
		t.Errorf("Expected DisallowedTools[0]='Bash', got '%s'", result.DisallowedTools[0])
	}
}
