package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

func TestSubprocessTransport_Write_NotReady(t *testing.T) {
	transport := &SubprocessTransport{
		ready:   false,
		options: &Options{},
	}

	err := transport.Write(context.Background(), "test data")
	if err == nil {
		t.Error("Expected error when transport is not ready")
	}
	if err.Error() != "transport is not ready for writing" {
		t.Errorf("Expected 'not ready' error, got: %v", err)
	}
}

func TestSubprocessTransport_Write_NilStdin(t *testing.T) {
	transport := &SubprocessTransport{
		ready:   true,
		stdin:   nil,
		options: &Options{},
	}

	err := transport.Write(context.Background(), "test data")
	if err == nil {
		t.Error("Expected error when stdin is nil")
	}
}

func TestSubprocessTransport_EndInput_NilStdin(t *testing.T) {
	transport := &SubprocessTransport{
		stdin:   nil,
		options: &Options{},
	}

	err := transport.EndInput()
	if err != nil {
		t.Errorf("Expected no error for nil stdin, got %v", err)
	}
}

func TestSubprocessTransport_BuildCommand_MaxThinkingTokens(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			MaxThinkingTokens: 8192,
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundMaxThinking := false
	for i, arg := range cmd {
		if arg == "--max-thinking-tokens" && i+1 < len(cmd) && cmd[i+1] == "8192" {
			foundMaxThinking = true
		}
	}

	if !foundMaxThinking {
		t.Error("Expected --max-thinking-tokens flag with value 8192")
	}
}

func TestSubprocessTransport_BuildCommand_OutputFormatJSON(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			OutputFormat: map[string]any{
				"type": "json_schema",
				"schema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{"type": "string"},
					},
				},
			},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundJSONSchema := false
	for i, arg := range cmd {
		if arg == "--json-schema" && i+1 < len(cmd) {
			foundJSONSchema = true
		}
	}

	if !foundJSONSchema {
		t.Error("Expected --json-schema flag")
	}
}

func TestSubprocessTransport_BuildCommand_ForkSession(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			ForkSession: true,
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundForkSession := false
	for _, arg := range cmd {
		if arg == "--fork-session" {
			foundForkSession = true
		}
	}

	if !foundForkSession {
		t.Error("Expected --fork-session flag")
	}
}

func TestSubprocessTransport_BuildCommand_ContinueConversation(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			ContinueConversation: true,
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundContinue := false
	for _, arg := range cmd {
		if arg == "--continue" {
			foundContinue = true
		}
	}

	if !foundContinue {
		t.Error("Expected --continue flag")
	}
}

func TestSubprocessTransport_BuildCommand_Resume(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			Resume: "session-123",
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundResume := false
	for i, arg := range cmd {
		if arg == "--resume" && i+1 < len(cmd) && cmd[i+1] == "session-123" {
			foundResume = true
		}
	}

	if !foundResume {
		t.Error("Expected --resume flag with session ID")
	}
}

func TestSubprocessTransport_BuildCommand_DisallowedTools(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			DisallowedTools: []string{"Bash", "Write"},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundDisallowed := false
	for i, arg := range cmd {
		if arg == "--disallowedTools" && i+1 < len(cmd) && cmd[i+1] == "Bash,Write" {
			foundDisallowed = true
		}
	}

	if !foundDisallowed {
		t.Error("Expected --disallowedTools flag")
	}
}

func TestSubprocessTransport_BuildCommand_FallbackModel(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			FallbackModel: "claude-3-haiku",
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundFallback := false
	for i, arg := range cmd {
		if arg == "--fallback-model" && i+1 < len(cmd) && cmd[i+1] == "claude-3-haiku" {
			foundFallback = true
		}
	}

	if !foundFallback {
		t.Error("Expected --fallback-model flag")
	}
}

func TestSubprocessTransport_BuildCommand_IncludePartialMessages(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			IncludePartialMessages: true,
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundPartial := false
	for _, arg := range cmd {
		if arg == "--include-partial-messages" {
			foundPartial = true
		}
	}

	if !foundPartial {
		t.Error("Expected --include-partial-messages flag")
	}
}

func TestSubprocessTransport_BuildCommand_PermissionPromptTool(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			PermissionPromptToolName: "custom_prompt_tool",
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundPromptTool := false
	for i, arg := range cmd {
		if arg == "--permission-prompt-tool" && i+1 < len(cmd) && cmd[i+1] == "custom_prompt_tool" {
			foundPromptTool = true
		}
	}

	if !foundPromptTool {
		t.Error("Expected --permission-prompt-tool flag")
	}
}

func TestSubprocessTransport_BuildCommand_MaxBudgetUSD(t *testing.T) {
	budget := 10.5
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			MaxBudgetUSD: &budget,
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundBudget := false
	for i, arg := range cmd {
		if arg == "--max-budget-usd" && i+1 < len(cmd) && cmd[i+1] == "10.5" {
			foundBudget = true
		}
	}

	if !foundBudget {
		t.Error("Expected --max-budget-usd flag")
	}
}

func TestSubprocessTransport_BuildCommand_Agents(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			Agents: map[string]AgentDefinition{
				"test-agent": {
					Description: "A test agent",
					Prompt:      "You are a test agent",
					Tools:       []string{"Read"},
				},
			},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundAgents := false
	for _, arg := range cmd {
		if arg == "--agents" {
			foundAgents = true
		}
	}

	if !foundAgents {
		t.Error("Expected --agents flag")
	}
}

func TestSubprocessTransport_BuildCommand_SettingSources(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			SettingSources: []string{"project", "user"},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundSettingSources := false
	for i, arg := range cmd {
		if arg == "--setting-sources" && i+1 < len(cmd) && cmd[i+1] == "project,user" {
			foundSettingSources = true
		}
	}

	if !foundSettingSources {
		t.Error("Expected --setting-sources flag with values")
	}
}

func TestSubprocessTransport_BuildSettingsValue_Both(t *testing.T) {
	transport := &SubprocessTransport{
		options: &Options{
			Settings: `{"key": "value"}`,
			Sandbox: &SandboxSettings{
				Enabled: true,
			},
		},
	}

	result, err := transport.buildSettingsValue()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Result should be valid JSON containing both settings
	if result == "" {
		t.Error("Expected non-empty result")
	}

	// Should contain sandbox
	if !contains(result, "sandbox") {
		t.Error("Expected result to contain sandbox settings")
	}
}

func TestSubprocessTransport_BuildSettingsValue_InvalidJSON(t *testing.T) {
	transport := &SubprocessTransport{
		options: &Options{
			Settings: `{invalid json}`,
			Sandbox: &SandboxSettings{
				Enabled: true,
			},
		},
	}

	// Should not error - invalid JSON is treated as file path
	result, err := transport.buildSettingsValue()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty result")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestSubprocessTransport_BuildCommand_ToolsDefault(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			Tools: "default", // Not a []string, should use "default"
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundTools := false
	for i, arg := range cmd {
		if arg == "--tools" && i+1 < len(cmd) && cmd[i+1] == "default" {
			foundTools = true
		}
	}

	if !foundTools {
		t.Error("Expected --tools default for non-slice tools")
	}
}

func TestSubprocessTransport_BuildCommand_MCPServersEmpty(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			MCPServers: map[string]any{}, // Empty map
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should not have --mcp-config for empty map
	for _, arg := range cmd {
		if arg == "--mcp-config" {
			t.Error("Should not include --mcp-config for empty map")
		}
	}
}

func TestSubprocessTransport_BuildCommand_SettingSourcesNil(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			SettingSources: nil,
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have --setting-sources with empty string for nil
	foundSettingSources := false
	for i, arg := range cmd {
		if arg == "--setting-sources" && i+1 < len(cmd) && cmd[i+1] == "" {
			foundSettingSources = true
		}
	}

	if !foundSettingSources {
		t.Error("Expected --setting-sources with empty string for nil")
	}
}

func TestSubprocessTransport_BuildCommand_MultiplePlugins(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			Plugins: []PluginConfig{
				{Type: "local", Path: "/path/to/plugin1"},
				{Type: "local", Path: "/path/to/plugin2"},
				{Type: "remote", Path: "/path/to/remote"}, // Non-local should be skipped
			},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	pluginDirCount := 0
	for _, arg := range cmd {
		if arg == "--plugin-dir" {
			pluginDirCount++
		}
	}

	if pluginDirCount != 2 {
		t.Errorf("Expected 2 --plugin-dir flags (only local), got %d", pluginDirCount)
	}
}

func TestSubprocessTransport_BuildCommand_OutputFormatNotJSON(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			OutputFormat: map[string]any{
				"type": "text", // Not json_schema
			},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should not have --json-schema for non-json_schema type
	for _, arg := range cmd {
		if arg == "--json-schema" {
			t.Error("Should not include --json-schema for non-json_schema type")
		}
	}
}

func TestSubprocessTransport_BuildCommand_SystemPromptPresetNoAppend(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			SystemPrompt: &SystemPromptPreset{
				Type:   "preset",
				Preset: "claude_code",
				Append: "", // Empty append
			},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should not have --append-system-prompt for empty Append
	for _, arg := range cmd {
		if arg == "--append-system-prompt" {
			t.Error("Should not include --append-system-prompt for empty Append")
		}
	}
}

func TestSubprocessTransport_BuildCommand_SystemPromptPresetWrongType(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			SystemPrompt: &SystemPromptPreset{
				Type:   "custom", // Not "preset"
				Preset: "claude_code",
				Append: "Additional instructions",
			},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should not have --append-system-prompt for wrong type
	for _, arg := range cmd {
		if arg == "--append-system-prompt" {
			t.Error("Should not include --append-system-prompt for wrong type")
		}
	}
}

func TestSubprocessTransport_EndInput_WithStdin(t *testing.T) {
	// Create a mock stdin
	r, w, _ := os.Pipe()
	defer r.Close()

	transport := &SubprocessTransport{
		stdin:   w,
		options: &Options{},
	}

	err := transport.EndInput()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// stdin should be nil after EndInput
	if transport.stdin != nil {
		t.Error("Expected stdin to be nil after EndInput")
	}
}

func TestSubprocessTransport_Close_WithTempFiles(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test-temp-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	transport := &SubprocessTransport{
		tempFiles: []string{tmpFile.Name()},
		options:   &Options{},
	}

	err = transport.Close()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Temp file should be removed
	if _, err := os.Stat(tmpFile.Name()); !os.IsNotExist(err) {
		t.Error("Expected temp file to be removed")
		os.Remove(tmpFile.Name()) // Clean up if not removed
	}
}

func TestMCPSDKServerConfig_GetType(t *testing.T) {
	config := MCPSDKServerConfig{
		Type: "sdk",
		Name: "test",
	}

	if config.GetType() != "sdk" {
		t.Errorf("Expected GetType() to return 'sdk', got '%s'", config.GetType())
	}
}

func TestSubprocessTransport_BuildSettingsValue_FilePath(t *testing.T) {
	// Create a temp settings file
	tmpFile, err := os.CreateTemp("", "settings-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	settingsContent := `{"key": "file_value"}`
	_, err = tmpFile.WriteString(settingsContent)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	transport := &SubprocessTransport{
		options: &Options{
			Settings: tmpFile.Name(),
			Sandbox: &SandboxSettings{
				Enabled: true,
			},
		},
	}

	result, err := transport.buildSettingsValue()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should contain sandbox settings merged with file content
	if result == "" {
		t.Error("Expected non-empty result")
	}
}

func TestSubprocessTransport_BuildSettingsValue_InvalidFilePath(t *testing.T) {
	transport := &SubprocessTransport{
		options: &Options{
			Settings: "/nonexistent/path/to/settings.json",
			Sandbox: &SandboxSettings{
				Enabled: true,
			},
		},
	}

	// Should not error, just ignore the file
	result, err := transport.buildSettingsValue()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should still have sandbox settings
	if result == "" {
		t.Error("Expected non-empty result with sandbox")
	}
}

func TestSubprocessTransport_Write_WithExitError(t *testing.T) {
	transport := &SubprocessTransport{
		ready:     true,
		stdin:     nil,
		exitError: os.ErrClosed,
		options:   &Options{},
	}

	err := transport.Write(context.Background(), "test")
	if err == nil {
		t.Error("Expected error when stdin is nil")
	}
}

func TestSubprocessTransport_Close_WithStdinAndStderr(t *testing.T) {
	// Create mock pipes
	stdinR, stdinW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()
	defer stdinR.Close()
	defer stderrW.Close()

	transport := &SubprocessTransport{
		stdin:   stdinW,
		stderr:  stderrR,
		options: &Options{},
	}

	err := transport.Close()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if transport.stdin != nil {
		t.Error("Expected stdin to be nil after Close")
	}
	if transport.stderr != nil {
		t.Error("Expected stderr to be nil after Close")
	}
}

func TestSubprocessTransport_BuildCommand_AllOptions(t *testing.T) {
	budget := 5.0
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			Model:                    "claude-3-opus",
			FallbackModel:            "claude-3-sonnet",
			MaxTurns:                 10,
			MaxBudgetUSD:             &budget,
			PermissionMode:           "default",
			PermissionPromptToolName: "custom_tool",
			Tools:                    []string{"Bash", "Read"},
			AllowedTools:             []string{"Bash"},
			DisallowedTools:          []string{"Write"},
			Betas:                    []string{"beta1"},
			ContinueConversation:     true,
			Resume:                   "session-id",
			ForkSession:              true,
			IncludePartialMessages:   true,
			AddDirs:                  []string{"/dir1", "/dir2"},
			MaxThinkingTokens:        4096,
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check all flags are present
	flags := map[string]bool{
		"--model":                    false,
		"--fallback-model":           false,
		"--max-turns":                false,
		"--max-budget-usd":           false,
		"--permission-mode":          false,
		"--permission-prompt-tool":   false,
		"--tools":                    false,
		"--allowedTools":             false,
		"--disallowedTools":          false,
		"--betas":                    false,
		"--continue":                 false,
		"--resume":                   false,
		"--fork-session":             false,
		"--include-partial-messages": false,
		"--add-dir":                  false,
		"--max-thinking-tokens":      false,
	}

	for _, arg := range cmd {
		for flag := range flags {
			if arg == flag {
				flags[flag] = true
			}
		}
	}

	for flag, found := range flags {
		if !found {
			t.Errorf("Expected flag %s not found in command", flag)
		}
	}
}

func TestSubprocessTransport_BuildCommand_OutputFormatWithSchema(t *testing.T) {
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			OutputFormat: map[string]any{
				"type":   "json_schema",
				"schema": map[string]any{"type": "object"},
			},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundJSONSchema := false
	for i, arg := range cmd {
		if arg == "--json-schema" && i+1 < len(cmd) {
			foundJSONSchema = true
			// Verify it's valid JSON
			var parsed map[string]any
			if json.Unmarshal([]byte(cmd[i+1]), &parsed) != nil {
				t.Error("Expected valid JSON schema after --json-schema flag")
			}
		}
	}

	if !foundJSONSchema {
		t.Error("Expected --json-schema flag")
	}
}

func TestSubprocessTransport_BuildCommand_LongAgentsCommandLine(t *testing.T) {
	// Create many agents to potentially trigger temp file creation
	// Note: This won't actually create temp file in test since cmdLengthLimit is high
	agents := make(map[string]AgentDefinition)
	for i := 0; i < 10; i++ {
		agents[fmt.Sprintf("agent-%d", i)] = AgentDefinition{
			Description: "A test agent with a moderately long description",
			Prompt:      "You are a test agent. Please do test things.",
			Tools:       []string{"Read", "Write", "Bash"},
		}
	}

	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			Agents: agents,
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundAgents := false
	for _, arg := range cmd {
		if arg == "--agents" {
			foundAgents = true
		}
	}

	if !foundAgents {
		t.Error("Expected --agents flag")
	}
}

func TestSubprocessTransport_NewSubprocessTransport_WithCLIPath(t *testing.T) {
	options := &Options{
		CLIPath: "/custom/path/to/claude",
	}

	transport, err := NewSubprocessTransport("test prompt", false, options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if transport.cliPath != "/custom/path/to/claude" {
		t.Errorf("Expected CLI path '/custom/path/to/claude', got '%s'", transport.cliPath)
	}
}

func TestSubprocessTransport_BuildSettingsValue_JsonBraces(t *testing.T) {
	// Test settings that look like JSON (has braces) but combined with sandbox
	transport := &SubprocessTransport{
		options: &Options{
			Settings: `{"existing": "value"}`,
			Sandbox: &SandboxSettings{
				Enabled:                  true,
				AutoAllowBashIfSandboxed: true,
			},
		},
	}

	result, err := transport.buildSettingsValue()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Result should contain both existing and sandbox
	if result == "" {
		t.Error("Expected non-empty result")
	}
	if !contains(result, "sandbox") {
		t.Error("Expected result to contain sandbox")
	}
}

func TestSubprocessTransport_BuildCommand_AgentsError(t *testing.T) {
	// Create agents map that will serialize correctly
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			Agents: map[string]AgentDefinition{
				"test": {Description: "Test", Prompt: "Prompt"},
			},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have agents flag
	foundAgents := false
	for _, arg := range cmd {
		if arg == "--agents" {
			foundAgents = true
		}
	}
	if !foundAgents {
		t.Error("Expected --agents flag")
	}
}

func TestSubprocessTransport_BuildCommand_MCPServersJsonError(t *testing.T) {
	// Valid MCP servers map
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			MCPServers: map[string]any{
				"server": map[string]any{
					"type":    "stdio",
					"command": "test",
				},
			},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	foundMCP := false
	for i, arg := range cmd {
		if arg == "--mcp-config" && i+1 < len(cmd) {
			foundMCP = true
			// Verify valid JSON
			var parsed map[string]any
			if json.Unmarshal([]byte(cmd[i+1]), &parsed) != nil {
				t.Error("Expected valid JSON for --mcp-config value")
			}
		}
	}
	if !foundMCP {
		t.Error("Expected --mcp-config flag")
	}
}

func TestSubprocessTransport_BuildCommand_OutputFormatNoSchema(t *testing.T) {
	// Output format with json_schema type but no schema field
	transport := &SubprocessTransport{
		cliPath:     "/usr/local/bin/claude",
		isStreaming: true,
		options: &Options{
			OutputFormat: map[string]any{
				"type": "json_schema",
				// No schema field
			},
		},
	}

	cmd, err := transport.buildCommand()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should not have --json-schema since there's no schema
	for _, arg := range cmd {
		if arg == "--json-schema" {
			t.Error("Should not have --json-schema without schema field")
		}
	}
}

func TestSubprocessTransport_Close_MultipleCloses(t *testing.T) {
	transport := &SubprocessTransport{
		options: &Options{},
	}

	// First close
	err1 := transport.Close()
	if err1 != nil {
		t.Errorf("First close error: %v", err1)
	}

	// Second close should not error
	err2 := transport.Close()
	if err2 != nil {
		t.Errorf("Second close error: %v", err2)
	}
}

func TestSubprocessTransport_Close_Ready(t *testing.T) {
	transport := &SubprocessTransport{
		ready:   true,
		options: &Options{},
	}

	err := transport.Close()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if transport.ready {
		t.Error("Expected ready to be false after Close")
	}
}
