// Package types provides shared type definitions for the Claude SDK.
package types

import "context"

// PermissionMode defines the permission mode for tool execution.
type PermissionMode string

const (
	PermissionModeDefault           PermissionMode = "default"
	PermissionModeAcceptEdits       PermissionMode = "acceptEdits"
	PermissionModePlan              PermissionMode = "plan"
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

// PermissionRuleValue represents a permission rule.
type PermissionRuleValue struct {
	ToolName    string `json:"tool_name"`
	RuleContent string `json:"rule_content,omitempty"`
}

// PermissionUpdate represents a permission update configuration.
type PermissionUpdate struct {
	Type        PermissionUpdateType        `json:"type"`
	Rules       []PermissionRuleValue       `json:"rules,omitempty"`
	Behavior    PermissionBehavior          `json:"behavior,omitempty"`
	Mode        PermissionMode              `json:"mode,omitempty"`
	Directories []string                    `json:"directories,omitempty"`
	Destination PermissionUpdateDestination `json:"destination,omitempty"`
}

// ToMap converts PermissionUpdate to a map format.
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

// ToolPermissionContext provides context for tool permission callbacks.
type ToolPermissionContext struct {
	Signal      any
	Suggestions []PermissionUpdate
}

// PermissionResult is the interface for permission callback results.
type PermissionResult interface {
	IsAllow() bool
}

// PermissionResultAllow represents an allow permission result.
type PermissionResultAllow struct {
	UpdatedInput       map[string]any
	UpdatedPermissions []PermissionUpdate
}

func (PermissionResultAllow) IsAllow() bool { return true }

// PermissionResultDeny represents a deny permission result.
type PermissionResultDeny struct {
	Message   string
	Interrupt bool
}

func (PermissionResultDeny) IsAllow() bool { return false }

// CanUseToolFunc is the callback type for tool permission requests.
type CanUseToolFunc func(ctx context.Context, toolName string, input map[string]any, permCtx ToolPermissionContext) (PermissionResult, error)

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
	GetSessionID() string
	GetHookEventName() HookEvent
}

// BaseHookInput contains common fields.
type BaseHookInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
	PermissionMode string `json:"permission_mode,omitempty"`
}

func (b BaseHookInput) GetSessionID() string { return b.SessionID }

// PreToolUseHookInput is the input for PreToolUse hook events.
type PreToolUseHookInput struct {
	BaseHookInput
	ToolName  string         `json:"tool_name"`
	ToolInput map[string]any `json:"tool_input"`
}

func (PreToolUseHookInput) GetHookEventName() HookEvent { return HookEventPreToolUse }

// PostToolUseHookInput is the input for PostToolUse hook events.
type PostToolUseHookInput struct {
	BaseHookInput
	ToolName     string         `json:"tool_name"`
	ToolInput    map[string]any `json:"tool_input"`
	ToolResponse any            `json:"tool_response"`
}

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

func (PostToolUseFailureHookInput) GetHookEventName() HookEvent { return HookEventPostToolUseFailed }

// UserPromptSubmitHookInput is the input for UserPromptSubmit hook events.
type UserPromptSubmitHookInput struct {
	BaseHookInput
	Prompt string `json:"prompt"`
}

func (UserPromptSubmitHookInput) GetHookEventName() HookEvent { return HookEventUserPromptSubmit }

// StopHookInput is the input for Stop hook events.
type StopHookInput struct {
	BaseHookInput
	StopHookActive bool `json:"stop_hook_active"`
}

func (StopHookInput) GetHookEventName() HookEvent { return HookEventStop }

// SubagentStopHookInput is the input for SubagentStop hook events.
type SubagentStopHookInput struct {
	BaseHookInput
	StopHookActive bool `json:"stop_hook_active"`
}

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

func (PreCompactHookInput) GetHookEventName() HookEvent { return HookEventPreCompact }

// HookPermissionDecision represents the permission decision for PreToolUse hooks.
type HookPermissionDecision string

const (
	HookPermissionDecisionAllow HookPermissionDecision = "allow"
	HookPermissionDecisionDeny  HookPermissionDecision = "deny"
	HookPermissionDecisionAsk   HookPermissionDecision = "ask"
)

// HookDecision represents the decision field in hook output.
type HookDecision string

const (
	HookDecisionBlock HookDecision = "block"
)

// HookOutput represents the output from a hook callback.
type HookOutput struct {
	Async              bool                   `json:"async,omitempty"`
	AsyncTimeout       int                    `json:"asyncTimeout,omitempty"`
	Continue           *bool                  `json:"continue,omitempty"`
	SuppressOutput     bool                   `json:"suppressOutput,omitempty"`
	StopReason         string                 `json:"stopReason,omitempty"`
	Decision           HookDecision           `json:"decision,omitempty"`
	SystemMessage      string                 `json:"systemMessage,omitempty"`
	Reason             string                 `json:"reason,omitempty"`
	HookEventName      HookEvent              `json:"hookEventName,omitempty"`
	PermissionDecision HookPermissionDecision `json:"permissionDecision,omitempty"`
	UpdatedInput       map[string]any         `json:"updatedInput,omitempty"`
	AdditionalContext  string                 `json:"additionalContext,omitempty"`
}

// ToMap converts HookOutput to a map for JSON serialization.
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

	// Hook specific output
	if h.HookEventName != "" {
		specific := map[string]any{"hookEventName": string(h.HookEventName)}
		if h.PermissionDecision != "" {
			specific["permissionDecision"] = string(h.PermissionDecision)
		}
		if h.UpdatedInput != nil {
			specific["updatedInput"] = h.UpdatedInput
		}
		if h.AdditionalContext != "" {
			specific["additionalContext"] = h.AdditionalContext
		}
		result["hookSpecificOutput"] = specific
	}

	return result
}

// HookContext provides context for hook callbacks.
type HookContext struct {
	Signal any
}

// HookCallback is the function signature for hook callbacks.
type HookCallback func(ctx context.Context, input HookInput, toolUseID string, hookCtx HookContext) (HookOutput, error)

// HookMatcher configures which hooks to run for specific tool patterns.
type HookMatcher struct {
	Matcher string
	Hooks   []HookCallback
	Timeout float64
}

// MCPServer represents an in-process MCP server.
type MCPServer struct {
	Name    string
	Version string
	Tools   []MCPTool
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
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}
