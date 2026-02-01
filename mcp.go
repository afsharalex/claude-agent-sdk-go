package claude

import "context"

// MCPServerConfig is the interface for MCP server configurations.
type MCPServerConfig interface {
	mcpServerConfig()
	// GetType returns the server type.
	GetType() string
}

// MCPStdioServerConfig configures an MCP server that communicates via stdio.
type MCPStdioServerConfig struct {
	Type    string            `json:"type,omitempty"` // Optional, defaults to "stdio"
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

func (MCPStdioServerConfig) mcpServerConfig() {}
func (c MCPStdioServerConfig) GetType() string {
	if c.Type == "" {
		return "stdio"
	}
	return c.Type
}

// MCPSSEServerConfig configures an MCP server that communicates via Server-Sent Events.
type MCPSSEServerConfig struct {
	Type    string            `json:"type"` // Must be "sse"
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

func (MCPSSEServerConfig) mcpServerConfig() {}
func (c MCPSSEServerConfig) GetType() string { return "sse" }

// MCPHTTPServerConfig configures an MCP server that communicates via HTTP.
type MCPHTTPServerConfig struct {
	Type    string            `json:"type"` // Must be "http"
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

func (MCPHTTPServerConfig) mcpServerConfig() {}
func (c MCPHTTPServerConfig) GetType() string { return "http" }

// MCPSDKServerConfig configures an in-process SDK MCP server.
type MCPSDKServerConfig struct {
	Type    string     `json:"type"` // Must be "sdk"
	Name    string     `json:"name"`
	Server  *MCPServer `json:"-"` // The server instance, not serialized
}

func (MCPSDKServerConfig) mcpServerConfig() {}
func (c MCPSDKServerConfig) GetType() string { return "sdk" }

// MCPServer represents an in-process MCP server.
type MCPServer struct {
	name    string
	version string
	tools   []MCPTool
}

// MCPTool represents a tool that can be called via MCP.
type MCPTool struct {
	Name        string
	Description string
	InputSchema map[string]any
	Handler     MCPToolHandler
}

// MCPToolHandler is the function signature for MCP tool handlers.
type MCPToolHandler func(ctx context.Context, args map[string]any) (MCPToolResult, error)

// MCPToolResult represents the result of an MCP tool call.
type MCPToolResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"is_error,omitempty"`
}

// MCPContent represents content in an MCP tool result.
type MCPContent struct {
	Type     string `json:"type"` // "text" or "image"
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

// NewMCPServer creates a new in-process MCP server.
func NewMCPServer(name, version string, tools []MCPTool) *MCPServer {
	return &MCPServer{
		name:    name,
		version: version,
		tools:   tools,
	}
}

// Name returns the server name.
func (s *MCPServer) Name() string { return s.name }

// Version returns the server version.
func (s *MCPServer) Version() string { return s.version }

// Tools returns the list of tools.
func (s *MCPServer) Tools() []MCPTool { return s.tools }

// CreateSDKMCPServer creates an MCPSDKServerConfig with an in-process MCP server.
func CreateSDKMCPServer(name, version string, tools []MCPTool) MCPSDKServerConfig {
	server := NewMCPServer(name, version, tools)
	return MCPSDKServerConfig{
		Type:   "sdk",
		Name:   name,
		Server: server,
	}
}

// SettingSource defines where settings are loaded from.
type SettingSource string

const (
	SettingSourceUser    SettingSource = "user"
	SettingSourceProject SettingSource = "project"
	SettingSourceLocal   SettingSource = "local"
)

// AgentDefinition defines a custom agent.
type AgentDefinition struct {
	Description string   `json:"description"`
	Prompt      string   `json:"prompt"`
	Tools       []string `json:"tools,omitempty"`
	Model       string   `json:"model,omitempty"` // "sonnet", "opus", "haiku", "inherit"
}

// SandboxNetworkConfig configures network access for sandbox.
type SandboxNetworkConfig struct {
	// AllowUnixSockets specifies Unix socket paths accessible in sandbox.
	AllowUnixSockets []string `json:"allowUnixSockets,omitempty"`
	// AllowAllUnixSockets allows all Unix sockets (less secure).
	AllowAllUnixSockets bool `json:"allowAllUnixSockets,omitempty"`
	// AllowLocalBinding allows binding to localhost ports (macOS only).
	AllowLocalBinding bool `json:"allowLocalBinding,omitempty"`
	// HTTPProxyPort is the HTTP proxy port if bringing your own proxy.
	HTTPProxyPort int `json:"httpProxyPort,omitempty"`
	// SOCKSProxyPort is the SOCKS5 proxy port if bringing your own proxy.
	SOCKSProxyPort int `json:"socksProxyPort,omitempty"`
}

// SandboxIgnoreViolations configures which violations to ignore.
type SandboxIgnoreViolations struct {
	// File paths for which violations should be ignored.
	File []string `json:"file,omitempty"`
	// Network hosts for which violations should be ignored.
	Network []string `json:"network,omitempty"`
}

// SandboxSettings configures how Claude Code sandboxes bash commands.
type SandboxSettings struct {
	// Enabled enables bash sandboxing (macOS/Linux only).
	Enabled bool `json:"enabled,omitempty"`
	// AutoAllowBashIfSandboxed auto-approves bash commands when sandboxed.
	AutoAllowBashIfSandboxed bool `json:"autoAllowBashIfSandboxed,omitempty"`
	// ExcludedCommands are commands that should run outside the sandbox.
	ExcludedCommands []string `json:"excludedCommands,omitempty"`
	// AllowUnsandboxedCommands allows commands to bypass sandbox via dangerouslyDisableSandbox.
	AllowUnsandboxedCommands bool `json:"allowUnsandboxedCommands,omitempty"`
	// Network is the network configuration for sandbox.
	Network *SandboxNetworkConfig `json:"network,omitempty"`
	// IgnoreViolations specifies violations to ignore.
	IgnoreViolations *SandboxIgnoreViolations `json:"ignoreViolations,omitempty"`
	// EnableWeakerNestedSandbox enables weaker sandbox for unprivileged Docker environments.
	EnableWeakerNestedSandbox bool `json:"enableWeakerNestedSandbox,omitempty"`
}

// SdkPluginConfig configures an SDK plugin.
type SdkPluginConfig struct {
	Type string `json:"type"` // Currently only "local" is supported
	Path string `json:"path"`
}

// SdkBeta represents beta features that can be enabled.
type SdkBeta string

const (
	SdkBetaContext1M SdkBeta = "context-1m-2025-08-07"
)

// SystemPromptPreset configures a system prompt preset.
type SystemPromptPreset struct {
	Type   string `json:"type"`   // "preset"
	Preset string `json:"preset"` // "claude_code"
	Append string `json:"append,omitempty"`
}

// ToolsPreset configures a tools preset.
type ToolsPreset struct {
	Type   string `json:"type"`   // "preset"
	Preset string `json:"preset"` // "claude_code"
}
