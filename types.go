package claude

import "context"

// =============================================================================
// Content Blocks
// =============================================================================

// ContentBlock represents a block of content in a message.
// It can be one of: TextBlock, ThinkingBlock, ToolUseBlock, ToolResultBlock.
type ContentBlock interface {
	contentBlock()
}

// TextBlock represents text content.
type TextBlock struct {
	Text string `json:"text"`
}

func (TextBlock) contentBlock() {}

// ThinkingBlock represents thinking content from extended thinking.
type ThinkingBlock struct {
	Thinking  string `json:"thinking"`
	Signature string `json:"signature"`
}

func (ThinkingBlock) contentBlock() {}

// ToolUseBlock represents a tool use request.
type ToolUseBlock struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

func (ToolUseBlock) contentBlock() {}

// ToolResultBlock represents the result of a tool use.
type ToolResultBlock struct {
	ToolUseID string `json:"tool_use_id"`
	Content   any    `json:"content,omitempty"` // string | []map[string]any | nil
	IsError   *bool  `json:"is_error,omitempty"`
}

func (ToolResultBlock) contentBlock() {}

// =============================================================================
// Messages
// =============================================================================

// Message represents a message in the conversation.
// It can be one of: UserMessage, AssistantMessage, SystemMessage, ResultMessage, StreamEvent.
type Message interface {
	message()
}

// AssistantMessageError represents possible error types from assistant messages.
type AssistantMessageError string

const (
	AssistantMessageErrorAuthenticationFailed AssistantMessageError = "authentication_failed"
	AssistantMessageErrorBillingError         AssistantMessageError = "billing_error"
	AssistantMessageErrorRateLimit            AssistantMessageError = "rate_limit"
	AssistantMessageErrorInvalidRequest       AssistantMessageError = "invalid_request"
	AssistantMessageErrorServerError          AssistantMessageError = "server_error"
	AssistantMessageErrorUnknown              AssistantMessageError = "unknown"
)

// UserMessage represents a user message.
type UserMessage struct {
	// Content can be a string or a list of ContentBlocks
	Content         any            `json:"content"`
	UUID            string         `json:"uuid,omitempty"`
	ParentToolUseID string         `json:"parent_tool_use_id,omitempty"`
	ToolUseResult   map[string]any `json:"tool_use_result,omitempty"`
}

func (UserMessage) message() {}

// GetContentBlocks returns the content as a slice of ContentBlocks if applicable.
// Returns nil if content is a string.
func (m *UserMessage) GetContentBlocks() []ContentBlock {
	if blocks, ok := m.Content.([]ContentBlock); ok {
		return blocks
	}
	return nil
}

// GetContentString returns the content as a string if applicable.
// Returns empty string if content is blocks.
func (m *UserMessage) GetContentString() string {
	if s, ok := m.Content.(string); ok {
		return s
	}
	return ""
}

// AssistantMessage represents an assistant message with content blocks.
type AssistantMessage struct {
	Content         []ContentBlock        `json:"content"`
	Model           string                `json:"model"`
	ParentToolUseID string                `json:"parent_tool_use_id,omitempty"`
	Error           AssistantMessageError `json:"error,omitempty"`
}

func (AssistantMessage) message() {}

// SystemMessage represents a system message with metadata.
type SystemMessage struct {
	Subtype string         `json:"subtype"`
	Data    map[string]any `json:"data"`
}

func (SystemMessage) message() {}

// ResultMessage represents a result message with cost and usage information.
type ResultMessage struct {
	Subtype          string         `json:"subtype"`
	DurationMs       int            `json:"duration_ms"`
	DurationAPIMs    int            `json:"duration_api_ms"`
	IsError          bool           `json:"is_error"`
	NumTurns         int            `json:"num_turns"`
	SessionID        string         `json:"session_id"`
	TotalCostUSD     *float64       `json:"total_cost_usd,omitempty"`
	Usage            map[string]any `json:"usage,omitempty"`
	Result           string         `json:"result,omitempty"`
	StructuredOutput any            `json:"structured_output,omitempty"`
}

func (ResultMessage) message() {}

// StreamEvent represents a stream event for partial message updates during streaming.
type StreamEvent struct {
	UUID            string         `json:"uuid"`
	SessionID       string         `json:"session_id"`
	Event           map[string]any `json:"event"` // The raw Anthropic API stream event
	ParentToolUseID string         `json:"parent_tool_use_id,omitempty"`
}

func (StreamEvent) message() {}

// =============================================================================
// Hooks
// =============================================================================

// HookEvent represents the type of hook event.
type HookEvent string

const (
	HookEventPreToolUse        HookEvent = "PreToolUse"
	HookEventPostToolUse       HookEvent = "PostToolUse"
	HookEventPostToolUseFailed HookEvent = "PostToolUseFailure"
	HookEventUserPromptSubmit  HookEvent = "UserPromptSubmit"
	HookEventStop              HookEvent = "Stop"
	HookEventSubagentStop      HookEvent = "SubagentStop"
	HookEventPreCompact        HookEvent = "PreCompact"
)

// HookInput is the interface for all hook input types.
type HookInput interface {
	hookInput()
	GetSessionID() string
	GetTranscriptPath() string
	GetCwd() string
	GetHookEventName() HookEvent
}

// BaseHookInput contains common fields present across many hook events.
type BaseHookInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
	PermissionMode string `json:"permission_mode,omitempty"`
}

func (b BaseHookInput) GetSessionID() string      { return b.SessionID }
func (b BaseHookInput) GetTranscriptPath() string { return b.TranscriptPath }
func (b BaseHookInput) GetCwd() string            { return b.Cwd }

// PreToolUseHookInput is the input for PreToolUse hook events.
type PreToolUseHookInput struct {
	BaseHookInput
	ToolName  string         `json:"tool_name"`
	ToolInput map[string]any `json:"tool_input"`
}

func (PreToolUseHookInput) hookInput()                  {}
func (PreToolUseHookInput) GetHookEventName() HookEvent { return HookEventPreToolUse }

// PostToolUseHookInput is the input for PostToolUse hook events.
type PostToolUseHookInput struct {
	BaseHookInput
	ToolName     string         `json:"tool_name"`
	ToolInput    map[string]any `json:"tool_input"`
	ToolResponse any            `json:"tool_response"`
}

func (PostToolUseHookInput) hookInput()                  {}
func (PostToolUseHookInput) GetHookEventName() HookEvent { return HookEventPostToolUse }

// PostToolUseFailureHookInput is the input for PostToolUseFailure hook events.
type PostToolUseFailureHookInput struct {
	BaseHookInput
	ToolName    string         `json:"tool_name"`
	ToolInput   map[string]any `json:"tool_input"`
	ToolUseID   string         `json:"tool_use_id"`
	Error       string         `json:"error"`
	IsInterrupt bool           `json:"is_interrupt,omitempty"`
}

func (PostToolUseFailureHookInput) hookInput()                  {}
func (PostToolUseFailureHookInput) GetHookEventName() HookEvent { return HookEventPostToolUseFailed }

// UserPromptSubmitHookInput is the input for UserPromptSubmit hook events.
type UserPromptSubmitHookInput struct {
	BaseHookInput
	Prompt string `json:"prompt"`
}

func (UserPromptSubmitHookInput) hookInput()                  {}
func (UserPromptSubmitHookInput) GetHookEventName() HookEvent { return HookEventUserPromptSubmit }

// StopHookInput is the input for Stop hook events.
type StopHookInput struct {
	BaseHookInput
	StopHookActive bool `json:"stop_hook_active"`
}

func (StopHookInput) hookInput()                  {}
func (StopHookInput) GetHookEventName() HookEvent { return HookEventStop }

// SubagentStopHookInput is the input for SubagentStop hook events.
type SubagentStopHookInput struct {
	BaseHookInput
	StopHookActive bool `json:"stop_hook_active"`
}

func (SubagentStopHookInput) hookInput()                  {}
func (SubagentStopHookInput) GetHookEventName() HookEvent { return HookEventSubagentStop }

// PreCompactTrigger represents what triggered the pre-compact hook.
type PreCompactTrigger string

const (
	PreCompactTriggerManual PreCompactTrigger = "manual"
	PreCompactTriggerAuto   PreCompactTrigger = "auto"
)

// PreCompactHookInput is the input for PreCompact hook events.
type PreCompactHookInput struct {
	BaseHookInput
	Trigger            PreCompactTrigger `json:"trigger"`
	CustomInstructions *string           `json:"custom_instructions,omitempty"`
}

func (PreCompactHookInput) hookInput()                  {}
func (PreCompactHookInput) GetHookEventName() HookEvent { return HookEventPreCompact }

// HookPermissionDecision represents the permission decision for PreToolUse hooks.
type HookPermissionDecision string

const (
	HookPermissionDecisionAllow HookPermissionDecision = "allow"
	HookPermissionDecisionDeny  HookPermissionDecision = "deny"
	HookPermissionDecisionAsk   HookPermissionDecision = "ask"
)

// PreToolUseHookSpecificOutput contains output specific to PreToolUse events.
type PreToolUseHookSpecificOutput struct {
	HookEventName            HookEvent              `json:"hookEventName"`
	PermissionDecision       HookPermissionDecision `json:"permissionDecision,omitempty"`
	PermissionDecisionReason string                 `json:"permissionDecisionReason,omitempty"`
	UpdatedInput             map[string]any         `json:"updatedInput,omitempty"`
}

// PostToolUseHookSpecificOutput contains output specific to PostToolUse events.
type PostToolUseHookSpecificOutput struct {
	HookEventName     HookEvent `json:"hookEventName"`
	AdditionalContext string    `json:"additionalContext,omitempty"`
}

// PostToolUseFailureHookSpecificOutput contains output specific to PostToolUseFailure events.
type PostToolUseFailureHookSpecificOutput struct {
	HookEventName     HookEvent `json:"hookEventName"`
	AdditionalContext string    `json:"additionalContext,omitempty"`
}

// UserPromptSubmitHookSpecificOutput contains output specific to UserPromptSubmit events.
type UserPromptSubmitHookSpecificOutput struct {
	HookEventName     HookEvent `json:"hookEventName"`
	AdditionalContext string    `json:"additionalContext,omitempty"`
}

// HookSpecificOutput is the interface for hook-specific output types.
type HookSpecificOutput interface {
	hookSpecificOutput()
}

func (PreToolUseHookSpecificOutput) hookSpecificOutput()        {}
func (PostToolUseHookSpecificOutput) hookSpecificOutput()       {}
func (PostToolUseFailureHookSpecificOutput) hookSpecificOutput() {}
func (UserPromptSubmitHookSpecificOutput) hookSpecificOutput()  {}

// HookDecision represents the decision field in hook output.
type HookDecision string

const (
	HookDecisionBlock HookDecision = "block"
)

// HookOutput represents the output from a hook callback.
// See https://docs.anthropic.com/en/docs/claude-code/hooks#advanced%3A-json-output
type HookOutput struct {
	// Async indicates this hook should be deferred.
	// When true, the hook returns immediately and AsyncTimeout applies.
	Async bool `json:"async,omitempty"`
	// AsyncTimeout is the timeout in milliseconds for async operations.
	AsyncTimeout int `json:"asyncTimeout,omitempty"`

	// Continue indicates whether Claude should proceed after hook execution.
	// Defaults to true.
	Continue *bool `json:"continue,omitempty"`
	// SuppressOutput hides stdout from transcript mode.
	SuppressOutput bool `json:"suppressOutput,omitempty"`
	// StopReason is the message shown when Continue is false.
	StopReason string `json:"stopReason,omitempty"`

	// Decision is set to "block" to indicate blocking behavior.
	Decision HookDecision `json:"decision,omitempty"`
	// SystemMessage is a warning message displayed to the user.
	SystemMessage string `json:"systemMessage,omitempty"`
	// Reason is feedback message for Claude about the decision.
	Reason string `json:"reason,omitempty"`

	// HookSpecificOutput contains event-specific controls.
	HookSpecificOutput HookSpecificOutput `json:"hookSpecificOutput,omitempty"`
}

// ToMap converts HookOutput to a map for JSON serialization.
// This handles the conversion of Go field names to CLI-expected names.
func (h *HookOutput) ToMap() map[string]any {
	result := make(map[string]any)

	if h.Async {
		result["async"] = true
		if h.AsyncTimeout > 0 {
			result["asyncTimeout"] = h.AsyncTimeout
		}
		return result
	}

	if h.Continue != nil {
		result["continue"] = *h.Continue
	}
	if h.SuppressOutput {
		result["suppressOutput"] = true
	}
	if h.StopReason != "" {
		result["stopReason"] = h.StopReason
	}
	if h.Decision != "" {
		result["decision"] = string(h.Decision)
	}
	if h.SystemMessage != "" {
		result["systemMessage"] = h.SystemMessage
	}
	if h.Reason != "" {
		result["reason"] = h.Reason
	}
	if h.HookSpecificOutput != nil {
		// Convert hook specific output based on type
		switch v := h.HookSpecificOutput.(type) {
		case PreToolUseHookSpecificOutput:
			specific := map[string]any{"hookEventName": string(v.HookEventName)}
			if v.PermissionDecision != "" {
				specific["permissionDecision"] = string(v.PermissionDecision)
			}
			if v.PermissionDecisionReason != "" {
				specific["permissionDecisionReason"] = v.PermissionDecisionReason
			}
			if v.UpdatedInput != nil {
				specific["updatedInput"] = v.UpdatedInput
			}
			result["hookSpecificOutput"] = specific
		case PostToolUseHookSpecificOutput:
			specific := map[string]any{"hookEventName": string(v.HookEventName)}
			if v.AdditionalContext != "" {
				specific["additionalContext"] = v.AdditionalContext
			}
			result["hookSpecificOutput"] = specific
		case PostToolUseFailureHookSpecificOutput:
			specific := map[string]any{"hookEventName": string(v.HookEventName)}
			if v.AdditionalContext != "" {
				specific["additionalContext"] = v.AdditionalContext
			}
			result["hookSpecificOutput"] = specific
		case UserPromptSubmitHookSpecificOutput:
			specific := map[string]any{"hookEventName": string(v.HookEventName)}
			if v.AdditionalContext != "" {
				specific["additionalContext"] = v.AdditionalContext
			}
			result["hookSpecificOutput"] = specific
		}
	}

	return result
}

// HookContext provides context information for hook callbacks.
type HookContext struct {
	// Signal is reserved for future abort signal support.
	Signal any
}

// HookCallback is the function signature for hook callbacks.
type HookCallback func(ctx context.Context, input HookInput, toolUseID string, hookCtx HookContext) (HookOutput, error)

// HookMatcher configures which hooks to run for specific tool patterns.
type HookMatcher struct {
	// Matcher is a pattern to match tool names.
	// For PreToolUse, this can be a tool name like "Bash" or
	// a combination like "Write|MultiEdit|Edit".
	// See https://docs.anthropic.com/en/docs/claude-code/hooks#structure
	Matcher string

	// Hooks is the list of callbacks to run when the matcher matches.
	Hooks []HookCallback

	// Timeout is the timeout in seconds for all hooks in this matcher.
	// Defaults to 60 seconds.
	Timeout float64
}

// =============================================================================
// Permissions
// =============================================================================

// PermissionMode defines the permission mode for tool execution.
type PermissionMode string

const (
	// PermissionModeDefault prompts for dangerous tools.
	PermissionModeDefault PermissionMode = "default"
	// PermissionModeAcceptEdits auto-accepts file edits.
	PermissionModeAcceptEdits PermissionMode = "acceptEdits"
	// PermissionModePlan is for plan mode operations.
	PermissionModePlan PermissionMode = "plan"
	// PermissionModeBypassPermissions allows all tools (use with caution).
	PermissionModeBypassPermissions PermissionMode = "bypassPermissions"
)

// PermissionUpdateDestination defines where permission updates are stored.
type PermissionUpdateDestination string

const (
	PermissionUpdateDestinationUserSettings    PermissionUpdateDestination = "userSettings"
	PermissionUpdateDestinationProjectSettings PermissionUpdateDestination = "projectSettings"
	PermissionUpdateDestinationLocalSettings   PermissionUpdateDestination = "localSettings"
	PermissionUpdateDestinationSession         PermissionUpdateDestination = "session"
)

// PermissionBehavior defines the permission behavior.
type PermissionBehavior string

const (
	PermissionBehaviorAllow PermissionBehavior = "allow"
	PermissionBehaviorDeny  PermissionBehavior = "deny"
	PermissionBehaviorAsk   PermissionBehavior = "ask"
)

// PermissionRuleValue represents a permission rule.
type PermissionRuleValue struct {
	ToolName    string `json:"tool_name"`
	RuleContent string `json:"rule_content,omitempty"`
}

// PermissionUpdateType defines the type of permission update.
type PermissionUpdateType string

const (
	PermissionUpdateTypeAddRules          PermissionUpdateType = "addRules"
	PermissionUpdateTypeReplaceRules      PermissionUpdateType = "replaceRules"
	PermissionUpdateTypeRemoveRules       PermissionUpdateType = "removeRules"
	PermissionUpdateTypeSetMode           PermissionUpdateType = "setMode"
	PermissionUpdateTypeAddDirectories    PermissionUpdateType = "addDirectories"
	PermissionUpdateTypeRemoveDirectories PermissionUpdateType = "removeDirectories"
)

// PermissionUpdate represents a permission update configuration.
type PermissionUpdate struct {
	Type        PermissionUpdateType        `json:"type"`
	Rules       []PermissionRuleValue       `json:"rules,omitempty"`
	Behavior    PermissionBehavior          `json:"behavior,omitempty"`
	Mode        PermissionMode              `json:"mode,omitempty"`
	Directories []string                    `json:"directories,omitempty"`
	Destination PermissionUpdateDestination `json:"destination,omitempty"`
}

// ToMap converts PermissionUpdate to a map format matching TypeScript control protocol.
func (p *PermissionUpdate) ToMap() map[string]any {
	result := map[string]any{
		"type": string(p.Type),
	}

	if p.Destination != "" {
		result["destination"] = string(p.Destination)
	}

	switch p.Type {
	case PermissionUpdateTypeAddRules, PermissionUpdateTypeReplaceRules, PermissionUpdateTypeRemoveRules:
		if p.Rules != nil {
			rules := make([]map[string]any, len(p.Rules))
			for i, rule := range p.Rules {
				rules[i] = map[string]any{
					"toolName":    rule.ToolName,
					"ruleContent": rule.RuleContent,
				}
			}
			result["rules"] = rules
		}
		if p.Behavior != "" {
			result["behavior"] = string(p.Behavior)
		}

	case PermissionUpdateTypeSetMode:
		if p.Mode != "" {
			result["mode"] = string(p.Mode)
		}

	case PermissionUpdateTypeAddDirectories, PermissionUpdateTypeRemoveDirectories:
		if p.Directories != nil {
			result["directories"] = p.Directories
		}
	}

	return result
}

// ToolPermissionContext provides context information for tool permission callbacks.
type ToolPermissionContext struct {
	// Signal is reserved for future abort signal support.
	Signal any
	// Suggestions contains permission suggestions from CLI.
	Suggestions []PermissionUpdate
}

// PermissionResult is the interface for permission callback results.
type PermissionResult interface {
	permissionResult()
}

// PermissionResultAllow represents an allow permission result.
type PermissionResultAllow struct {
	UpdatedInput       map[string]any     `json:"updated_input,omitempty"`
	UpdatedPermissions []PermissionUpdate `json:"updated_permissions,omitempty"`
}

func (PermissionResultAllow) permissionResult() {}

// PermissionResultDeny represents a deny permission result.
type PermissionResultDeny struct {
	Message   string `json:"message,omitempty"`
	Interrupt bool   `json:"interrupt,omitempty"`
}

func (PermissionResultDeny) permissionResult() {}

// CanUseToolFunc is the callback type for tool permission requests.
// It receives the tool name, input parameters, and context, and returns a permission result.
type CanUseToolFunc func(ctx context.Context, toolName string, input map[string]any, permCtx ToolPermissionContext) (PermissionResult, error)

// =============================================================================
// MCP (Model Context Protocol)
// =============================================================================

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

func (MCPSSEServerConfig) mcpServerConfig()   {}
func (c MCPSSEServerConfig) GetType() string { return "sse" }

// MCPHTTPServerConfig configures an MCP server that communicates via HTTP.
type MCPHTTPServerConfig struct {
	Type    string            `json:"type"` // Must be "http"
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

func (MCPHTTPServerConfig) mcpServerConfig()   {}
func (c MCPHTTPServerConfig) GetType() string { return "http" }

// MCPSDKServerConfig configures an in-process SDK MCP server.
type MCPSDKServerConfig struct {
	Type   string     `json:"type"` // Must be "sdk"
	Name   string     `json:"name"`
	Server *MCPServer `json:"-"` // The server instance, not serialized
}

func (MCPSDKServerConfig) mcpServerConfig()   {}
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

// =============================================================================
// Configuration Types
// =============================================================================

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
