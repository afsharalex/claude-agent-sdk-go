package claude

import "context"

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

func (PreToolUseHookInput) hookInput()                   {}
func (PreToolUseHookInput) GetHookEventName() HookEvent  { return HookEventPreToolUse }

// PostToolUseHookInput is the input for PostToolUse hook events.
type PostToolUseHookInput struct {
	BaseHookInput
	ToolName     string         `json:"tool_name"`
	ToolInput    map[string]any `json:"tool_input"`
	ToolResponse any            `json:"tool_response"`
}

func (PostToolUseHookInput) hookInput()                   {}
func (PostToolUseHookInput) GetHookEventName() HookEvent  { return HookEventPostToolUse }

// PostToolUseFailureHookInput is the input for PostToolUseFailure hook events.
type PostToolUseFailureHookInput struct {
	BaseHookInput
	ToolName    string         `json:"tool_name"`
	ToolInput   map[string]any `json:"tool_input"`
	ToolUseID   string         `json:"tool_use_id"`
	Error       string         `json:"error"`
	IsInterrupt bool           `json:"is_interrupt,omitempty"`
}

func (PostToolUseFailureHookInput) hookInput()                   {}
func (PostToolUseFailureHookInput) GetHookEventName() HookEvent  { return HookEventPostToolUseFailed }

// UserPromptSubmitHookInput is the input for UserPromptSubmit hook events.
type UserPromptSubmitHookInput struct {
	BaseHookInput
	Prompt string `json:"prompt"`
}

func (UserPromptSubmitHookInput) hookInput()                   {}
func (UserPromptSubmitHookInput) GetHookEventName() HookEvent  { return HookEventUserPromptSubmit }

// StopHookInput is the input for Stop hook events.
type StopHookInput struct {
	BaseHookInput
	StopHookActive bool `json:"stop_hook_active"`
}

func (StopHookInput) hookInput()                   {}
func (StopHookInput) GetHookEventName() HookEvent  { return HookEventStop }

// SubagentStopHookInput is the input for SubagentStop hook events.
type SubagentStopHookInput struct {
	BaseHookInput
	StopHookActive bool `json:"stop_hook_active"`
}

func (SubagentStopHookInput) hookInput()                   {}
func (SubagentStopHookInput) GetHookEventName() HookEvent  { return HookEventSubagentStop }

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

func (PreCompactHookInput) hookInput()                   {}
func (PreCompactHookInput) GetHookEventName() HookEvent  { return HookEventPreCompact }

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
