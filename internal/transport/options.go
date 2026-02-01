package transport

import "io"

// Options configures SubprocessTransport behavior.
type Options struct {
	Tools                    any // []string | *ToolsPreset
	AllowedTools             []string
	SystemPrompt             any // string | *SystemPromptPreset
	MCPServers               any // map[string]MCPServerConfig | string
	PermissionMode           string
	ContinueConversation     bool
	Resume                   string
	MaxTurns                 int
	MaxBudgetUSD             *float64
	DisallowedTools          []string
	Model                    string
	FallbackModel            string
	Betas                    []string
	PermissionPromptToolName string
	Cwd                      string
	CLIPath                  string
	Settings                 string
	AddDirs                  []string
	Env                      map[string]string
	ExtraArgs                map[string]*string
	MaxBufferSize            int
	DebugStderr              io.Writer
	Stderr                   func(line string)
	User                     string
	IncludePartialMessages   bool
	ForkSession              bool
	Agents                   map[string]AgentDefinition
	SettingSources           []string
	Sandbox                  *SandboxSettings
	Plugins                  []PluginConfig
	MaxThinkingTokens        int
	OutputFormat             map[string]any
	EnableFileCheckpointing  bool
}

// AgentDefinition defines a custom agent.
type AgentDefinition struct {
	Description string   `json:"description"`
	Prompt      string   `json:"prompt"`
	Tools       []string `json:"tools,omitempty"`
	Model       string   `json:"model,omitempty"`
}

// SandboxSettings configures bash command sandboxing.
type SandboxSettings struct {
	Enabled                   bool                     `json:"enabled,omitempty"`
	AutoAllowBashIfSandboxed  bool                     `json:"autoAllowBashIfSandboxed,omitempty"`
	ExcludedCommands          []string                 `json:"excludedCommands,omitempty"`
	AllowUnsandboxedCommands  bool                     `json:"allowUnsandboxedCommands,omitempty"`
	Network                   *SandboxNetworkConfig    `json:"network,omitempty"`
	IgnoreViolations          *SandboxIgnoreViolations `json:"ignoreViolations,omitempty"`
	EnableWeakerNestedSandbox bool                     `json:"enableWeakerNestedSandbox,omitempty"`
}

// SandboxNetworkConfig configures network access for sandbox.
type SandboxNetworkConfig struct {
	AllowUnixSockets    []string `json:"allowUnixSockets,omitempty"`
	AllowAllUnixSockets bool     `json:"allowAllUnixSockets,omitempty"`
	AllowLocalBinding   bool     `json:"allowLocalBinding,omitempty"`
	HTTPProxyPort       int      `json:"httpProxyPort,omitempty"`
	SOCKSProxyPort      int      `json:"socksProxyPort,omitempty"`
}

// SandboxIgnoreViolations configures which violations to ignore.
type SandboxIgnoreViolations struct {
	File    []string `json:"file,omitempty"`
	Network []string `json:"network,omitempty"`
}

// PluginConfig configures an SDK plugin.
type PluginConfig struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

// SystemPromptPreset configures a system prompt preset.
type SystemPromptPreset struct {
	Type   string `json:"type"`
	Preset string `json:"preset"`
	Append string `json:"append,omitempty"`
}

// MCPServerConfig is a minimal interface for MCP server configs.
type MCPServerConfig interface {
	GetType() string
}

// MCPSDKServerConfig represents an SDK MCP server.
type MCPSDKServerConfig struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

func (MCPSDKServerConfig) GetType() string { return "sdk" }
