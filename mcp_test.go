package claude

import (
	"context"
	"testing"
)

func TestMCPServerConfig_Interface(t *testing.T) {
	// Verify all MCP server config types implement MCPServerConfig interface
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
		{
			name:     "default type",
			config:   MCPStdioServerConfig{Command: "npx"},
			expected: "stdio",
		},
		{
			name:     "explicit type",
			config:   MCPStdioServerConfig{Type: "stdio", Command: "npx"},
			expected: "stdio",
		},
		{
			name:     "custom type (should still work)",
			config:   MCPStdioServerConfig{Type: "custom", Command: "npx"},
			expected: "custom",
		},
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
	config := MCPSSEServerConfig{
		Type: "sse",
		URL:  "https://example.com/sse",
	}

	if config.GetType() != "sse" {
		t.Errorf("Expected type 'sse', got '%s'", config.GetType())
	}
}

func TestMCPSSEServerConfig_Fields(t *testing.T) {
	config := MCPSSEServerConfig{
		Type:    "sse",
		URL:     "https://example.com/sse",
		Headers: map[string]string{"Authorization": "Bearer token"},
	}

	if config.URL != "https://example.com/sse" {
		t.Errorf("Expected URL to match")
	}
	if config.Headers["Authorization"] != "Bearer token" {
		t.Errorf("Expected Authorization header to match")
	}
}

func TestMCPHTTPServerConfig_GetType(t *testing.T) {
	config := MCPHTTPServerConfig{
		Type: "http",
		URL:  "https://example.com/api",
	}

	if config.GetType() != "http" {
		t.Errorf("Expected type 'http', got '%s'", config.GetType())
	}
}

func TestMCPHTTPServerConfig_Fields(t *testing.T) {
	config := MCPHTTPServerConfig{
		Type:    "http",
		URL:     "https://example.com/api",
		Headers: map[string]string{"X-API-Key": "secret"},
	}

	if config.URL != "https://example.com/api" {
		t.Errorf("Expected URL to match")
	}
	if config.Headers["X-API-Key"] != "secret" {
		t.Errorf("Expected X-API-Key header to match")
	}
}

func TestMCPSDKServerConfig_GetType(t *testing.T) {
	config := MCPSDKServerConfig{
		Type: "sdk",
		Name: "test-server",
	}

	if config.GetType() != "sdk" {
		t.Errorf("Expected type 'sdk', got '%s'", config.GetType())
	}
}

func TestMCPSDKServerConfig_Fields(t *testing.T) {
	server := NewMCPServer("test", "1.0.0", nil)
	config := MCPSDKServerConfig{
		Type:   "sdk",
		Name:   "test-server",
		Server: server,
	}

	if config.Name != "test-server" {
		t.Errorf("Expected Name 'test-server', got '%s'", config.Name)
	}
	if config.Server == nil {
		t.Error("Expected Server to be non-nil")
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

func TestMCPServer_Getters(t *testing.T) {
	server := NewMCPServer("test-server", "2.0.0", []MCPTool{
		{Name: "tool1"},
		{Name: "tool2"},
	})

	if server.Name() != "test-server" {
		t.Errorf("Expected Name 'test-server', got '%s'", server.Name())
	}
	if server.Version() != "2.0.0" {
		t.Errorf("Expected Version '2.0.0', got '%s'", server.Version())
	}
	if len(server.Tools()) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(server.Tools()))
	}
}

func TestMCPTool_Fields(t *testing.T) {
	handler := func(ctx context.Context, args map[string]any) (MCPToolResult, error) {
		return TextResult("result"), nil
	}

	tool := MCPTool{
		Name:        "my-tool",
		Description: "Does something useful",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"arg1": map[string]any{"type": "string"},
			},
		},
		Handler: handler,
	}

	if tool.Name != "my-tool" {
		t.Errorf("Expected Name 'my-tool', got '%s'", tool.Name)
	}
	if tool.Description != "Does something useful" {
		t.Errorf("Expected Description to match")
	}
	if tool.InputSchema == nil {
		t.Error("Expected InputSchema to be non-nil")
	}
	if tool.Handler == nil {
		t.Error("Expected Handler to be non-nil")
	}

	// Test handler
	result, err := tool.Handler(context.Background(), nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(result.Content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(result.Content))
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
	content := MCPContent{
		Type: "text",
		Text: "Hello, world!",
	}

	if content.Type != "text" {
		t.Errorf("Expected Type 'text', got '%s'", content.Type)
	}
	if content.Text != "Hello, world!" {
		t.Errorf("Expected Text 'Hello, world!', got '%s'", content.Text)
	}
}

func TestMCPContent_Image(t *testing.T) {
	content := MCPContent{
		Type:     "image",
		Data:     "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
		MimeType: "image/png",
	}

	if content.Type != "image" {
		t.Errorf("Expected Type 'image', got '%s'", content.Type)
	}
	if content.MimeType != "image/png" {
		t.Errorf("Expected MimeType 'image/png', got '%s'", content.MimeType)
	}
}

func TestCreateSDKMCPServer(t *testing.T) {
	tools := []MCPTool{
		{Name: "add", Description: "Add numbers"},
	}

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
	if config.Server.Name() != "math-server" {
		t.Errorf("Expected Server.Name 'math-server', got '%s'", config.Server.Name())
	}
	if config.Server.Version() != "1.0.0" {
		t.Errorf("Expected Server.Version '1.0.0', got '%s'", config.Server.Version())
	}
	if len(config.Server.Tools()) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(config.Server.Tools()))
	}
}

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
	if agent.Prompt != "You are a test agent" {
		t.Errorf("Expected Prompt to match")
	}
	if len(agent.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(agent.Tools))
	}
	if agent.Model != "sonnet" {
		t.Errorf("Expected Model 'sonnet', got '%s'", agent.Model)
	}
}

func TestSandboxSettings_Fields(t *testing.T) {
	settings := SandboxSettings{
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
	}

	if !settings.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if !settings.AutoAllowBashIfSandboxed {
		t.Error("Expected AutoAllowBashIfSandboxed to be true")
	}
	if len(settings.ExcludedCommands) != 1 {
		t.Errorf("Expected 1 excluded command, got %d", len(settings.ExcludedCommands))
	}
	if settings.Network == nil {
		t.Fatal("Expected Network to be non-nil")
	}
	if settings.Network.HTTPProxyPort != 8080 {
		t.Errorf("Expected HTTPProxyPort 8080, got %d", settings.Network.HTTPProxyPort)
	}
	if settings.IgnoreViolations == nil {
		t.Fatal("Expected IgnoreViolations to be non-nil")
	}
	if len(settings.IgnoreViolations.File) != 1 {
		t.Errorf("Expected 1 file violation, got %d", len(settings.IgnoreViolations.File))
	}
}

func TestSdkPluginConfig_Fields(t *testing.T) {
	config := SdkPluginConfig{
		Type: "local",
		Path: "/path/to/plugin",
	}

	if config.Type != "local" {
		t.Errorf("Expected Type 'local', got '%s'", config.Type)
	}
	if config.Path != "/path/to/plugin" {
		t.Errorf("Expected Path '/path/to/plugin', got '%s'", config.Path)
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
	if preset.Append != "Additional instructions" {
		t.Errorf("Expected Append to match")
	}
}

func TestToolsPreset_Fields(t *testing.T) {
	preset := ToolsPreset{
		Type:   "preset",
		Preset: "claude_code",
	}

	if preset.Type != "preset" {
		t.Errorf("Expected Type 'preset', got '%s'", preset.Type)
	}
	if preset.Preset != "claude_code" {
		t.Errorf("Expected Preset 'claude_code', got '%s'", preset.Preset)
	}
}

// Benchmark tests

func BenchmarkNewMCPServer(b *testing.B) {
	tools := []MCPTool{
		{Name: "tool1", Description: "Tool 1"},
		{Name: "tool2", Description: "Tool 2"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewMCPServer("server", "1.0.0", tools)
	}
}

func BenchmarkCreateSDKMCPServer(b *testing.B) {
	tools := []MCPTool{
		{Name: "tool1", Description: "Tool 1"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CreateSDKMCPServer("server", "1.0.0", tools)
	}
}
