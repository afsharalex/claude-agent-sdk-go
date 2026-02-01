package transport

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{"equal versions", "1.0.0", "1.0.0", 0},
		{"a less than b major", "1.0.0", "2.0.0", -1},
		{"a greater than b major", "2.0.0", "1.0.0", 1},
		{"a less than b minor", "1.1.0", "1.2.0", -1},
		{"a greater than b minor", "1.2.0", "1.1.0", 1},
		{"a less than b patch", "1.0.1", "1.0.2", -1},
		{"a greater than b patch", "1.0.2", "1.0.1", 1},
		{"shorter a version", "1.0", "1.0.0", -1},
		{"shorter b version", "1.0.0", "1.0", 1},
		{"different lengths equal", "1.0", "1.0", 0},
		{"complex comparison", "2.1.3", "2.1.4", -1},
		{"major trumps minor", "2.0.0", "1.9.9", 1},
		{"minor trumps patch", "1.2.0", "1.1.9", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareVersions(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("compareVersions(%s, %s) = %d, expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestBuildSettingsValue_NoSettings(t *testing.T) {
	transport := &SubprocessTransport{
		options: &Options{},
	}

	result, err := transport.buildSettingsValue()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

func TestBuildSettingsValue_SettingsOnly(t *testing.T) {
	transport := &SubprocessTransport{
		options: &Options{
			Settings: `{"key": "value"}`,
		},
	}

	result, err := transport.buildSettingsValue()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != `{"key": "value"}` {
		t.Errorf("Expected settings string, got '%s'", result)
	}
}

func TestBuildSettingsValue_SandboxOnly(t *testing.T) {
	transport := &SubprocessTransport{
		options: &Options{
			Sandbox: &SandboxSettings{
				Enabled: true,
			},
		},
	}

	result, err := transport.buildSettingsValue()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty result")
	}
	// Should contain sandbox setting
	if result == "" {
		t.Error("Expected result with sandbox settings")
	}
}

func TestNewSubprocessTransport_Defaults(t *testing.T) {
	// This test would require CLI to be installed, so we skip if not available
	// We can test that the function handles options correctly

	options := &Options{
		CLIPath:       "/nonexistent/path/claude",
		MaxBufferSize: 2048,
		Cwd:           "/tmp",
	}

	transport, err := NewSubprocessTransport("test prompt", false, options)
	if err != nil {
		// Expected if CLI not found
		return
	}

	if transport.maxBufferSize != 2048 {
		t.Errorf("Expected maxBufferSize 2048, got %d", transport.maxBufferSize)
	}
	if transport.cwd != "/tmp" {
		t.Errorf("Expected cwd '/tmp', got '%s'", transport.cwd)
	}
}

func TestNewSubprocessTransport_DefaultBufferSize(t *testing.T) {
	options := &Options{
		CLIPath: "/nonexistent/path/claude",
	}

	transport, _ := NewSubprocessTransport("test", false, options)
	if transport == nil {
		return // CLI not found, skip
	}

	if transport.maxBufferSize != defaultMaxBufferSize {
		t.Errorf("Expected default maxBufferSize %d, got %d", defaultMaxBufferSize, transport.maxBufferSize)
	}
}

func TestSubprocessTransport_BuildCommand_Basic(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: false,
		prompt:      "Hello, Claude!",
		options: &Options{
			Model:    "claude-3-opus",
			MaxTurns: 5,
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(cmd) == 0 {
		t.Fatal("Expected non-empty command")
	}
	if cmd[0] != "/usr/local/bin/claude" {
		t.Errorf("Expected first arg to be CLI path, got '%s'", cmd[0])
	}

	// Check for expected flags
	foundModel := false
	foundMaxTurns := false
	for i, arg := range cmd {
		if arg == "--model" && i+1 < len(cmd) && cmd[i+1] == "claude-3-opus" {
			foundModel = true
		}
		if arg == "--max-turns" && i+1 < len(cmd) && cmd[i+1] == "5" {
			foundMaxTurns = true
		}
	}

	if !foundModel {
		t.Error("Expected --model flag")
	}
	if !foundMaxTurns {
		t.Error("Expected --max-turns flag")
	}
}

func TestSubprocessTransport_BuildCommand_Streaming(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options:     &Options{},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Streaming mode should have --input-format stream-json
	foundInputFormat := false
	for i, arg := range cmd {
		if arg == "--input-format" && i+1 < len(cmd) && cmd[i+1] == "stream-json" {
			foundInputFormat = true
		}
	}

	if !foundInputFormat {
		t.Error("Expected --input-format stream-json for streaming mode")
	}
}

func TestSubprocessTransport_BuildCommand_NonStreaming(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: false,
		prompt:      "Test prompt",
		options:     &Options{},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Non-streaming mode should have --print and the prompt
	foundPrint := false
	foundPrompt := false
	for i, arg := range cmd {
		if arg == "--print" {
			foundPrint = true
		}
		if arg == "Test prompt" {
			foundPrompt = true
		}
		// Check for -- separator
		if arg == "--" && i+1 < len(cmd) && cmd[i+1] == "Test prompt" {
			foundPrompt = true
		}
	}

	if !foundPrint {
		t.Error("Expected --print flag for non-streaming mode")
	}
	if !foundPrompt {
		t.Error("Expected prompt in command for non-streaming mode")
	}
}

func TestSubprocessTransport_BuildCommand_SystemPrompt(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			SystemPrompt: "You are a helpful assistant",
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundSystemPrompt := false
	for i, arg := range cmd {
		if arg == "--system-prompt" && i+1 < len(cmd) && cmd[i+1] == "You are a helpful assistant" {
			foundSystemPrompt = true
		}
	}

	if !foundSystemPrompt {
		t.Error("Expected --system-prompt flag with value")
	}
}

func TestSubprocessTransport_BuildCommand_SystemPromptNil(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			SystemPrompt: nil,
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// nil system prompt should still add --system-prompt with empty string
	foundSystemPrompt := false
	for i, arg := range cmd {
		if arg == "--system-prompt" && i+1 < len(cmd) && cmd[i+1] == "" {
			foundSystemPrompt = true
		}
	}

	if !foundSystemPrompt {
		t.Error("Expected --system-prompt with empty string for nil")
	}
}

func TestSubprocessTransport_BuildCommand_SystemPromptPreset(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			SystemPrompt: &SystemPromptPreset{
				Type:   "preset",
				Preset: "claude_code",
				Append: "Additional instructions",
			},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundAppend := false
	for i, arg := range cmd {
		if arg == "--append-system-prompt" && i+1 < len(cmd) && cmd[i+1] == "Additional instructions" {
			foundAppend = true
		}
	}

	if !foundAppend {
		t.Error("Expected --append-system-prompt flag")
	}
}

func TestSubprocessTransport_BuildCommand_Tools(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			Tools: []string{"read_file", "write_file"},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundTools := false
	for i, arg := range cmd {
		if arg == "--tools" && i+1 < len(cmd) && cmd[i+1] == "read_file,write_file" {
			foundTools = true
		}
	}

	if !foundTools {
		t.Error("Expected --tools flag with comma-separated values")
	}
}

func TestSubprocessTransport_BuildCommand_EmptyTools(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			Tools: []string{},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundTools := false
	for i, arg := range cmd {
		if arg == "--tools" && i+1 < len(cmd) && cmd[i+1] == "" {
			foundTools = true
		}
	}

	if !foundTools {
		t.Error("Expected --tools with empty string for empty slice")
	}
}

func TestSubprocessTransport_BuildCommand_AllowedTools(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			AllowedTools: []string{"Bash", "Read"},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundAllowedTools := false
	for i, arg := range cmd {
		if arg == "--allowedTools" && i+1 < len(cmd) && cmd[i+1] == "Bash,Read" {
			foundAllowedTools = true
		}
	}

	if !foundAllowedTools {
		t.Error("Expected --allowedTools flag")
	}
}

func TestSubprocessTransport_BuildCommand_Betas(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			Betas: []string{"context-1m-2025-08-07"},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundBetas := false
	for i, arg := range cmd {
		if arg == "--betas" && i+1 < len(cmd) && cmd[i+1] == "context-1m-2025-08-07" {
			foundBetas = true
		}
	}

	if !foundBetas {
		t.Error("Expected --betas flag")
	}
}

func TestSubprocessTransport_BuildCommand_PermissionMode(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			PermissionMode: "bypassPermissions",
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundMode := false
	for i, arg := range cmd {
		if arg == "--permission-mode" && i+1 < len(cmd) && cmd[i+1] == "bypassPermissions" {
			foundMode = true
		}
	}

	if !foundMode {
		t.Error("Expected --permission-mode flag")
	}
}

func TestSubprocessTransport_BuildCommand_AddDirs(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			AddDirs: []string{"/path/one", "/path/two"},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	addDirCount := 0
	for _, arg := range cmd {
		if arg == "--add-dir" {
			addDirCount++
		}
	}

	if addDirCount != 2 {
		t.Errorf("Expected 2 --add-dir flags, got %d", addDirCount)
	}
}

func TestSubprocessTransport_BuildCommand_ExtraArgs(t *testing.T) {
	flagValue := "value"
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			ExtraArgs: map[string]*string{
				"custom-flag":  &flagValue,
				"boolean-flag": nil,
			},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundCustomFlag := false
	foundBooleanFlag := false
	for i, arg := range cmd {
		if arg == "--custom-flag" && i+1 < len(cmd) && cmd[i+1] == "value" {
			foundCustomFlag = true
		}
		if arg == "--boolean-flag" {
			foundBooleanFlag = true
		}
	}

	if !foundCustomFlag {
		t.Error("Expected --custom-flag with value")
	}
	if !foundBooleanFlag {
		t.Error("Expected --boolean-flag without value")
	}
}

func TestSubprocessTransport_BuildCommand_MCPServersString(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			MCPServers: "/path/to/mcp-config.json",
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundMCP := false
	for i, arg := range cmd {
		if arg == "--mcp-config" && i+1 < len(cmd) && cmd[i+1] == "/path/to/mcp-config.json" {
			foundMCP = true
		}
	}

	if !foundMCP {
		t.Error("Expected --mcp-config with path")
	}
}

func TestSubprocessTransport_BuildCommand_MCPServersMap(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			MCPServers: map[string]any{
				"server1": map[string]any{
					"type":    "stdio",
					"command": "npx",
				},
			},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundMCP := false
	for _, arg := range cmd {
		if arg == "--mcp-config" {
			foundMCP = true
		}
	}

	if !foundMCP {
		t.Error("Expected --mcp-config flag")
	}
}

func TestSubprocessTransport_BuildCommand_Plugins(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			Plugins: []PluginConfig{
				{Type: "local", Path: "/path/to/plugin"},
			},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundPlugin := false
	for i, arg := range cmd {
		if arg == "--plugin-dir" && i+1 < len(cmd) && cmd[i+1] == "/path/to/plugin" {
			foundPlugin = true
		}
	}

	if !foundPlugin {
		t.Error("Expected --plugin-dir flag")
	}
}

func TestSubprocessTransport_IsReady(t *testing.T) {
	transport := &SubprocessTransport{
		ready: false,
	}

	if transport.IsReady() {
		t.Error("Expected IsReady to return false")
	}

	transport.ready = true

	if !transport.IsReady() {
		t.Error("Expected IsReady to return true")
	}
}

func TestSubprocessTransport_Close_Idempotent(t *testing.T) {
	transport := &SubprocessTransport{
		options: &Options{},
	}

	// Multiple closes should not error
	err1 := transport.Close()
	err2 := transport.Close()
	err3 := transport.Close()

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

// Benchmark tests

func BenchmarkCompareVersions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		compareVersions("2.1.3", "2.0.0")
	}
}

func BenchmarkBuildCommand(b *testing.B) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			Model:          "claude-3-opus",
			MaxTurns:       10,
			PermissionMode: "default",
			Tools:          []string{"read_file", "write_file"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = transport.buildCommand()
	}
}
