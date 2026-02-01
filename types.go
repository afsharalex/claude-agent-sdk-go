package claude

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
