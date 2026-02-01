package claude

import "io"

// Options configures Claude SDK behavior.
type Options struct {
	// Tools specifies the base set of tools to use.
	// Can be a list of tool names or a ToolsPreset.
	Tools any // []string | *ToolsPreset

	// AllowedTools specifies additional tools to allow.
	AllowedTools []string

	// SystemPrompt is the system prompt to use.
	// Can be a string or a SystemPromptPreset.
	SystemPrompt any // string | *SystemPromptPreset

	// MCPServers configures MCP servers to use.
	// Can be a map of server configs or a path to a config file.
	MCPServers any // map[string]MCPServerConfig | string

	// PermissionMode controls tool execution permissions.
	PermissionMode PermissionMode

	// ContinueConversation continues from the last conversation.
	ContinueConversation bool

	// Resume specifies a session ID to resume.
	Resume string

	// MaxTurns limits the number of agentic turns.
	MaxTurns int

	// MaxBudgetUSD limits the cost in USD.
	MaxBudgetUSD *float64

	// DisallowedTools specifies tools that should not be used.
	DisallowedTools []string

	// Model specifies the model to use.
	Model string

	// FallbackModel specifies a fallback model.
	FallbackModel string

	// Betas specifies beta features to enable.
	Betas []SdkBeta

	// PermissionPromptToolName specifies the tool for permission prompts.
	PermissionPromptToolName string

	// Cwd specifies the working directory.
	Cwd string

	// CLIPath specifies the path to the Claude CLI.
	CLIPath string

	// Settings specifies settings as JSON string or file path.
	Settings string

	// AddDirs specifies additional directories to add.
	AddDirs []string

	// Env specifies additional environment variables.
	Env map[string]string

	// ExtraArgs specifies additional CLI flags.
	ExtraArgs map[string]*string

	// MaxBufferSize limits the buffer size for CLI stdout.
	MaxBufferSize int

	// DebugStderr is deprecated: use Stderr callback instead.
	DebugStderr io.Writer

	// Stderr is a callback for stderr output from CLI.
	Stderr func(line string)

	// CanUseTool is a callback for tool permission requests.
	CanUseTool CanUseToolFunc

	// Hooks configures hook callbacks.
	Hooks map[HookEvent][]HookMatcher

	// User specifies the user to run the CLI as.
	User string

	// IncludePartialMessages enables partial message streaming.
	IncludePartialMessages bool

	// ForkSession forks resumed sessions to a new session ID.
	ForkSession bool

	// Agents defines custom agent configurations.
	Agents map[string]AgentDefinition

	// SettingSources specifies which setting sources to load.
	SettingSources []SettingSource

	// Sandbox configures bash command sandboxing.
	Sandbox *SandboxSettings

	// Plugins configures SDK plugins.
	Plugins []SdkPluginConfig

	// MaxThinkingTokens limits tokens for thinking blocks.
	MaxThinkingTokens int

	// OutputFormat specifies structured output format.
	// Expected: {"type": "json_schema", "schema": {...}}
	OutputFormat map[string]any

	// EnableFileCheckpointing enables file checkpointing.
	EnableFileCheckpointing bool
}

// Option is a functional option for configuring Options.
type Option func(*Options)

// NewOptions creates a new Options with defaults applied.
func NewOptions(opts ...Option) *Options {
	o := &Options{
		Env:       make(map[string]string),
		ExtraArgs: make(map[string]*string),
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// WithTools sets the tools to use.
func WithTools(tools []string) Option {
	return func(o *Options) {
		o.Tools = tools
	}
}

// WithToolsPreset sets a tools preset.
func WithToolsPreset(preset *ToolsPreset) Option {
	return func(o *Options) {
		o.Tools = preset
	}
}

// WithAllowedTools sets additional allowed tools.
func WithAllowedTools(tools []string) Option {
	return func(o *Options) {
		o.AllowedTools = tools
	}
}

// WithSystemPrompt sets the system prompt.
func WithSystemPrompt(prompt string) Option {
	return func(o *Options) {
		o.SystemPrompt = prompt
	}
}

// WithSystemPromptPreset sets a system prompt preset.
func WithSystemPromptPreset(preset *SystemPromptPreset) Option {
	return func(o *Options) {
		o.SystemPrompt = preset
	}
}

// WithMCPServers sets MCP server configurations.
func WithMCPServers(servers map[string]MCPServerConfig) Option {
	return func(o *Options) {
		o.MCPServers = servers
	}
}

// WithMCPConfigPath sets the path to an MCP config file.
func WithMCPConfigPath(path string) Option {
	return func(o *Options) {
		o.MCPServers = path
	}
}

// WithPermissionMode sets the permission mode.
func WithPermissionMode(mode PermissionMode) Option {
	return func(o *Options) {
		o.PermissionMode = mode
	}
}

// WithContinueConversation enables continuing from the last conversation.
func WithContinueConversation(cont bool) Option {
	return func(o *Options) {
		o.ContinueConversation = cont
	}
}

// WithResume sets the session ID to resume.
func WithResume(sessionID string) Option {
	return func(o *Options) {
		o.Resume = sessionID
	}
}

// WithMaxTurns sets the maximum number of turns.
func WithMaxTurns(turns int) Option {
	return func(o *Options) {
		o.MaxTurns = turns
	}
}

// WithMaxBudgetUSD sets the maximum budget in USD.
func WithMaxBudgetUSD(budget float64) Option {
	return func(o *Options) {
		o.MaxBudgetUSD = &budget
	}
}

// WithDisallowedTools sets tools that should not be used.
func WithDisallowedTools(tools []string) Option {
	return func(o *Options) {
		o.DisallowedTools = tools
	}
}

// WithModel sets the model to use.
func WithModel(model string) Option {
	return func(o *Options) {
		o.Model = model
	}
}

// WithFallbackModel sets the fallback model.
func WithFallbackModel(model string) Option {
	return func(o *Options) {
		o.FallbackModel = model
	}
}

// WithBetas sets the beta features to enable.
func WithBetas(betas []SdkBeta) Option {
	return func(o *Options) {
		o.Betas = betas
	}
}

// WithPermissionPromptToolName sets the permission prompt tool name.
func WithPermissionPromptToolName(name string) Option {
	return func(o *Options) {
		o.PermissionPromptToolName = name
	}
}

// WithCwd sets the working directory.
func WithCwd(cwd string) Option {
	return func(o *Options) {
		o.Cwd = cwd
	}
}

// WithCLIPath sets the path to the Claude CLI.
func WithCLIPath(path string) Option {
	return func(o *Options) {
		o.CLIPath = path
	}
}

// WithSettings sets settings as JSON string or file path.
func WithSettings(settings string) Option {
	return func(o *Options) {
		o.Settings = settings
	}
}

// WithAddDirs sets additional directories to add.
func WithAddDirs(dirs []string) Option {
	return func(o *Options) {
		o.AddDirs = dirs
	}
}

// WithEnv sets additional environment variables.
func WithEnv(env map[string]string) Option {
	return func(o *Options) {
		o.Env = env
	}
}

// WithExtraArg adds an extra CLI argument.
func WithExtraArg(flag string, value *string) Option {
	return func(o *Options) {
		if o.ExtraArgs == nil {
			o.ExtraArgs = make(map[string]*string)
		}
		o.ExtraArgs[flag] = value
	}
}

// WithMaxBufferSize sets the max buffer size for CLI stdout.
func WithMaxBufferSize(size int) Option {
	return func(o *Options) {
		o.MaxBufferSize = size
	}
}

// WithStderr sets the stderr callback.
func WithStderr(callback func(string)) Option {
	return func(o *Options) {
		o.Stderr = callback
	}
}

// WithCanUseTool sets the tool permission callback.
func WithCanUseTool(callback CanUseToolFunc) Option {
	return func(o *Options) {
		o.CanUseTool = callback
	}
}

// WithHooks sets hook configurations.
func WithHooks(hooks map[HookEvent][]HookMatcher) Option {
	return func(o *Options) {
		o.Hooks = hooks
	}
}

// WithUser sets the user to run the CLI as.
func WithUser(user string) Option {
	return func(o *Options) {
		o.User = user
	}
}

// WithIncludePartialMessages enables partial message streaming.
func WithIncludePartialMessages(include bool) Option {
	return func(o *Options) {
		o.IncludePartialMessages = include
	}
}

// WithForkSession enables forking resumed sessions.
func WithForkSession(fork bool) Option {
	return func(o *Options) {
		o.ForkSession = fork
	}
}

// WithAgents sets custom agent definitions.
func WithAgents(agents map[string]AgentDefinition) Option {
	return func(o *Options) {
		o.Agents = agents
	}
}

// WithSettingSources sets which setting sources to load.
func WithSettingSources(sources []SettingSource) Option {
	return func(o *Options) {
		o.SettingSources = sources
	}
}

// WithSandbox sets sandbox configuration.
func WithSandbox(sandbox *SandboxSettings) Option {
	return func(o *Options) {
		o.Sandbox = sandbox
	}
}

// WithPlugins sets SDK plugin configurations.
func WithPlugins(plugins []SdkPluginConfig) Option {
	return func(o *Options) {
		o.Plugins = plugins
	}
}

// WithMaxThinkingTokens sets the max tokens for thinking blocks.
func WithMaxThinkingTokens(tokens int) Option {
	return func(o *Options) {
		o.MaxThinkingTokens = tokens
	}
}

// WithOutputFormat sets the structured output format.
func WithOutputFormat(format map[string]any) Option {
	return func(o *Options) {
		o.OutputFormat = format
	}
}

// WithEnableFileCheckpointing enables file checkpointing.
func WithEnableFileCheckpointing(enable bool) Option {
	return func(o *Options) {
		o.EnableFileCheckpointing = enable
	}
}
