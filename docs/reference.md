# API Reference

Complete reference documentation for the Claude Agent SDK for Go.

## Functions

### Query

```go
func Query(ctx context.Context, prompt string, opts ...Option) (<-chan Message, <-chan error)
```

Performs a one-shot query to Claude Code. Returns channels for messages and errors. The message channel closes when the query completes or encounters an error.

**Parameters:**
- `ctx` - Context for cancellation and timeouts
- `prompt` - The prompt to send to Claude
- `opts` - Optional configuration options

**Returns:**
- `<-chan Message` - Channel receiving response messages
- `<-chan error` - Channel receiving errors

**Example:**
```go
messages, errors := claude.Query(ctx, "Hello Claude",
    claude.WithMaxTurns(1),
)
```

---

### QueryStreaming

```go
func QueryStreaming(ctx context.Context, inputCh <-chan map[string]any, opts ...Option) (<-chan Message, <-chan error)
```

Performs a streaming query with an input channel for sending multiple messages. Used for more complex streaming scenarios.

**Parameters:**
- `ctx` - Context for cancellation
- `inputCh` - Channel for sending input messages
- `opts` - Optional configuration options

---

### WithClient

```go
func WithClient(ctx context.Context, fn func(*Client) error, opts ...Option) error
```

Creates a client, executes the function, and ensures cleanup. This is the recommended pattern for short-lived client operations.

**Parameters:**
- `ctx` - Context for cancellation and timeouts
- `fn` - Function to execute with the client
- `opts` - Optional configuration options

**Returns:**
- `error` - Error from connection or the executed function

**Example:**
```go
err := claude.WithClient(ctx, func(c *claude.Client) error {
    c.Query(ctx, "Hello")
    for msg := range c.ReceiveResponse(ctx) {
        // handle messages
    }
    return nil
}, claude.WithCwd("/path"))
```

---

### QueryWithSession

```go
func QueryWithSession(ctx context.Context, sessionID, prompt string, opts ...Option) (<-chan Message, <-chan error)
```

Performs a one-shot query with a specific session ID. This is useful for maintaining conversation context across multiple independent queries without using the full Client interface.

**Parameters:**
- `ctx` - Context for cancellation and timeouts
- `sessionID` - Session ID to resume (empty string creates new session)
- `prompt` - The prompt to send to Claude
- `opts` - Optional configuration options

**Returns:**
- `<-chan Message` - Channel receiving response messages
- `<-chan error` - Channel receiving errors

**Example:**
```go
// First query creates a session
messages, errors := claude.QueryWithSession(ctx, "", "Remember: my name is Alice")

// Later queries can resume the session
messages, errors := claude.QueryWithSession(ctx, "session-123", "What is my name?")
```

---

### NewClient

```go
func NewClient(opts ...Option) *Client
```

Creates a new Client for interactive conversations. The client must be connected before use.

**Parameters:**
- `opts` - Optional configuration options

**Returns:**
- `*Client` - New client instance

**Example:**
```go
client := claude.NewClient(
    claude.WithCwd("/path/to/project"),
    claude.WithModel("claude-sonnet-4-5"),
)
```

---

### NewMCPServer

```go
func NewMCPServer(name, version string, tools []MCPTool) *MCPServer
```

Creates a new in-process MCP server with the specified tools.

---

### CreateSDKMCPServer

```go
func CreateSDKMCPServer(name, version string, tools []MCPTool) MCPSDKServerConfig
```

Creates an MCP server configuration ready for use with `WithMCPServers()`.

---

### Tool

```go
func Tool(name, description string, inputSchema map[string]any, handler MCPToolHandler) MCPTool
```

Creates an MCP tool definition. Helper function for defining custom tools.

**Example:**
```go
tool := claude.Tool("add", "Add two numbers",
    map[string]any{
        "type": "object",
        "properties": map[string]any{
            "a": map[string]any{"type": "number"},
            "b": map[string]any{"type": "number"},
        },
    },
    func(ctx context.Context, args map[string]any) (claude.MCPToolResult, error) {
        a := args["a"].(float64)
        b := args["b"].(float64)
        return claude.TextResult(fmt.Sprintf("%g", a+b)), nil
    },
)
```

---

### SimpleInputSchema

```go
func SimpleInputSchema(fields map[string]string) map[string]any
```

Creates a JSON schema from simple field type mappings.

**Supported types:** `"string"`, `"number"`, `"integer"`, `"boolean"`

**Example:**
```go
schema := claude.SimpleInputSchema(map[string]string{
    "name": "string",
    "age":  "number",
})
```

---

### TextResult

```go
func TextResult(text string) MCPToolResult
```

Creates an MCP tool result with text content.

---

### ErrorResult

```go
func ErrorResult(message string) MCPToolResult
```

Creates an MCP tool result indicating an error.

---

### ImageResult

```go
func ImageResult(data, mimeType string) MCPToolResult
```

Creates an MCP tool result with image content.

---

## Types

### Client

```go
type Client struct {
    // unexported fields
}
```

Provides bidirectional, interactive conversations with Claude Code.

#### Methods

##### Connect

```go
func (c *Client) Connect(ctx context.Context) error
```

Establishes connection to Claude Code. Must be called before sending queries.

##### Query

```go
func (c *Client) Query(ctx context.Context, prompt string) error
```

Sends a query to Claude.

##### QueryMessage

```go
func (c *Client) QueryMessage(ctx context.Context, message map[string]any) error
```

Sends a structured message to Claude.

##### Messages

```go
func (c *Client) Messages() <-chan Message
```

Returns channel for receiving messages.

##### Errors

```go
func (c *Client) Errors() <-chan error
```

Returns channel for receiving errors.

##### ReceiveResponse

```go
func (c *Client) ReceiveResponse(ctx context.Context) <-chan Message
```

Returns a channel that yields messages until a ResultMessage is received.

##### Interrupt

```go
func (c *Client) Interrupt(ctx context.Context) error
```

Sends an interrupt signal to Claude.

##### SetPermissionMode

```go
func (c *Client) SetPermissionMode(ctx context.Context, mode PermissionMode) error
```

Changes the permission mode during conversation.

##### SetModel

```go
func (c *Client) SetModel(ctx context.Context, model string) error
```

Changes the AI model during conversation.

##### RewindFiles

```go
func (c *Client) RewindFiles(ctx context.Context, userMessageID string) error
```

Rewinds tracked files to their state at a specific user message.

##### GetMCPStatus

```go
func (c *Client) GetMCPStatus(ctx context.Context) (map[string]any, error)
```

Gets current MCP server connection status.

##### GetServerInfo

```go
func (c *Client) GetServerInfo() map[string]any
```

Returns server initialization info.

##### Close

```go
func (c *Client) Close() error
```

Disconnects from Claude Code.

##### SetSessionID

```go
func (c *Client) SetSessionID(sessionID string)
```

Sets the session ID for subsequent queries.

---

### Message Interface

```go
type Message interface {
    message()
}
```

Interface implemented by all message types.

**Implementations:**
- `*AssistantMessage`
- `*UserMessage`
- `*SystemMessage`
- `*ResultMessage`
- `*StreamEvent`

---

### AssistantMessage

```go
type AssistantMessage struct {
    Content         []ContentBlock        // Response content blocks
    Model           string                // Model that generated the response
    ParentToolUseID string                // Parent tool use ID (for nested calls)
    Error           AssistantMessageError // Error type if applicable
}
```

Represents Claude's response.

---

### UserMessage

```go
type UserMessage struct {
    Content         any            // string or []ContentBlock
    UUID            string         // Message UUID
    ParentToolUseID string         // Parent tool use ID
    ToolUseResult   map[string]any // Tool result data
}
```

Represents user input or tool results.

**Methods:**
- `GetContentBlocks() []ContentBlock` - Returns content as blocks
- `GetContentString() string` - Returns content as string

---

### SystemMessage

```go
type SystemMessage struct {
    Subtype string         // Message subtype
    Data    map[string]any // Message data
}
```

Represents system notifications.

---

### ResultMessage

```go
type ResultMessage struct {
    Subtype          string         // Result subtype
    DurationMs       int            // Total duration in milliseconds
    DurationAPIMs    int            // API call duration
    IsError          bool           // Whether the result indicates an error
    NumTurns         int            // Number of conversation turns
    SessionID        string         // Session identifier
    TotalCostUSD     *float64       // Total cost in USD
    Usage            map[string]any // Token usage details
    Result           string         // Text result summary
    StructuredOutput any            // Structured output data
}
```

Represents query completion with cost and usage information.

---

### StreamEvent

```go
type StreamEvent struct {
    UUID            string         // Event UUID
    SessionID       string         // Session ID
    Event           map[string]any // Raw API stream event
    ParentToolUseID string         // Parent tool use ID
}
```

Represents streaming events for partial message updates.

---

### ContentBlock Interface

```go
type ContentBlock interface {
    contentBlock()
}
```

Interface implemented by all content block types.

---

### TextBlock

```go
type TextBlock struct {
    Text string // The text content
}
```

Plain text content.

---

### ThinkingBlock

```go
type ThinkingBlock struct {
    Thinking  string // The thinking content
    Signature string // Verification signature
}
```

Extended thinking content.

---

### ToolUseBlock

```go
type ToolUseBlock struct {
    ID    string         // Tool use ID
    Name  string         // Tool name
    Input map[string]any // Tool input parameters
}
```

Tool invocation request.

---

### ToolResultBlock

```go
type ToolResultBlock struct {
    ToolUseID string // Corresponding tool use ID
    Content   any    // Result content
    IsError   *bool  // Whether the result is an error
}
```

Tool execution result.

---

## Options

### WithTools

```go
func WithTools(tools []string) Option
```

Sets the base set of tools to use.

---

### WithAllowedTools

```go
func WithAllowedTools(tools []string) Option
```

Sets additional tools to allow.

---

### WithDisallowedTools

```go
func WithDisallowedTools(tools []string) Option
```

Sets tools that should not be used.

---

### WithSystemPrompt

```go
func WithSystemPrompt(prompt string) Option
```

Sets the system prompt.

---

### WithAppendSystemPrompt

```go
func WithAppendSystemPrompt(text string) Option
```

Appends text to the system prompt. If no system prompt is set, this becomes the system prompt. Can be called multiple times to append additional text.

**Example:**
```go
client := claude.NewClient(
    claude.WithSystemPrompt("You are a helpful assistant."),
    claude.WithAppendSystemPrompt("Always respond in JSON format."),
)
```

---

### WithMCPServers

```go
func WithMCPServers(servers map[string]MCPServerConfig) Option
```

Configures MCP servers.

---

### WithMCPConfigPath

```go
func WithMCPConfigPath(path string) Option
```

Sets path to an MCP config file.

---

### WithPermissionMode

```go
func WithPermissionMode(mode PermissionMode) Option
```

Sets the permission mode.

**Values:**
- `PermissionModeDefault` - Prompts for dangerous tools
- `PermissionModeAcceptEdits` - Auto-accepts file edits
- `PermissionModePlan` - Plan mode operations
- `PermissionModeBypassPermissions` - Allows all tools

---

### WithContinueConversation

```go
func WithContinueConversation(cont bool) Option
```

Enables continuing from the last conversation.

---

### WithResume

```go
func WithResume(sessionID string) Option
```

Sets a session ID to resume.

---

### WithMaxTurns

```go
func WithMaxTurns(turns int) Option
```

Limits the number of agentic turns.

---

### WithMaxBudgetUSD

```go
func WithMaxBudgetUSD(budget float64) Option
```

Sets the maximum cost budget in USD.

---

### WithModel

```go
func WithModel(model string) Option
```

Sets the AI model to use.

---

### WithFallbackModel

```go
func WithFallbackModel(model string) Option
```

Sets a fallback model.

---

### WithCwd

```go
func WithCwd(cwd string) Option
```

Sets the working directory.

---

### WithCLIPath

```go
func WithCLIPath(path string) Option
```

Sets the path to the Claude CLI.

---

### WithEnv

```go
func WithEnv(env map[string]string) Option
```

Sets additional environment variables.

---

### WithEnvVar

```go
func WithEnvVar(key, value string) Option
```

Sets a single environment variable. Can be called multiple times to set multiple variables.

**Example:**
```go
client := claude.NewClient(
    claude.WithEnvVar("API_KEY", "secret"),
    claude.WithEnvVar("DEBUG", "true"),
)
```

---

### WithDebugStderr

```go
func WithDebugStderr() Option
```

Enables stderr output to os.Stderr for debugging. This is a convenience wrapper around `WithStderr` that prints CLI stderr output to standard error.

**Example:**
```go
// Enable debug output
client := claude.NewClient(
    claude.WithDebugStderr(),
)
```

---

### WithHooks

```go
func WithHooks(hooks map[HookEvent][]HookMatcher) Option
```

Registers hook configurations.

---

### WithCanUseTool

```go
func WithCanUseTool(callback CanUseToolFunc) Option
```

Sets the tool permission callback.

---

### WithSandbox

```go
func WithSandbox(sandbox *SandboxSettings) Option
```

Configures sandbox settings.

---

### WithAgents

```go
func WithAgents(agents map[string]AgentDefinition) Option
```

Defines custom subagents.

---

### WithIncludePartialMessages

```go
func WithIncludePartialMessages(include bool) Option
```

Enables partial message streaming.

---

### WithEnableFileCheckpointing

```go
func WithEnableFileCheckpointing(enable bool) Option
```

Enables file checkpointing.

---

## Hook Types

### HookEvent

```go
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
```

---

### HookMatcher

```go
type HookMatcher struct {
    Matcher string          // Pattern to match tool names
    Hooks   []HookCallback  // Callbacks to run
    Timeout float64         // Timeout in seconds
}
```

---

### HookCallback

```go
type HookCallback func(ctx context.Context, input HookInput, toolUseID string, hookCtx HookContext) (HookOutput, error)
```

---

### HookInput Types

- `PreToolUseHookInput` - Before tool execution
- `PostToolUseHookInput` - After successful tool execution
- `PostToolUseFailureHookInput` - After failed tool execution
- `UserPromptSubmitHookInput` - When user submits a prompt
- `StopHookInput` - When conversation stops
- `SubagentStopHookInput` - When subagent stops
- `PreCompactHookInput` - Before context compaction

---

### HookOutput

```go
type HookOutput struct {
    Async              bool              // Defer hook execution
    AsyncTimeout       int               // Async timeout in ms
    Continue           *bool             // Continue after hook
    SuppressOutput     bool              // Hide stdout
    StopReason         string            // Message when Continue is false
    Decision           HookDecision      // "block" to block
    SystemMessage      string            // Warning message
    Reason             string            // Feedback for Claude
    HookSpecificOutput HookSpecificOutput // Event-specific controls
}
```

---

## Permission Types

### PermissionResult

```go
type PermissionResult interface {
    permissionResult()
}
```

**Implementations:**
- `PermissionResultAllow` - Allow the tool
- `PermissionResultDeny` - Deny the tool

---

### CanUseToolFunc

```go
type CanUseToolFunc func(ctx context.Context, toolName string, input map[string]any, permCtx ToolPermissionContext) (PermissionResult, error)
```

Callback for tool permission decisions.

---

## MCP Types

### MCPServerConfig

```go
type MCPServerConfig interface {
    mcpServerConfig()
    GetType() string
}
```

**Implementations:**
- `MCPStdioServerConfig` - Subprocess via stdio
- `MCPSSEServerConfig` - Server-Sent Events
- `MCPHTTPServerConfig` - HTTP transport
- `MCPSDKServerConfig` - In-process SDK server

---

### MCPTool

```go
type MCPTool struct {
    Name        string
    Description string
    InputSchema map[string]any
    Handler     MCPToolHandler
}
```

---

### MCPToolResult

```go
type MCPToolResult struct {
    Content []MCPContent
    IsError bool
}
```

---

### MCPContent

```go
type MCPContent struct {
    Type     string // "text" or "image"
    Text     string
    Data     string
    MimeType string
}
```

---

## Error Types

### Error Helper Functions

#### IsConnectionError

```go
func IsConnectionError(err error) bool
```

Reports whether err is a `CLIConnectionError`. Works with wrapped errors.

---

#### IsCLINotFoundError

```go
func IsCLINotFoundError(err error) bool
```

Reports whether err is a `CLINotFoundError`. Works with wrapped errors.

---

#### IsProcessError

```go
func IsProcessError(err error) bool
```

Reports whether err is a `ProcessError`. Works with wrapped errors.

---

#### AsConnectionError

```go
func AsConnectionError(err error) (*CLIConnectionError, bool)
```

Extracts a `CLIConnectionError` from err. Returns the error and true if found, nil and false otherwise.

---

#### AsCLINotFoundError

```go
func AsCLINotFoundError(err error) (*CLINotFoundError, bool)
```

Extracts a `CLINotFoundError` from err. Returns the error and true if found, nil and false otherwise.

---

#### AsProcessError

```go
func AsProcessError(err error) (*ProcessError, bool)
```

Extracts a `ProcessError` from err. Returns the error and true if found, nil and false otherwise.

**Example:**
```go
if procErr, ok := claude.AsProcessError(err); ok {
    fmt.Printf("Exit code: %d\n", procErr.ExitCode)
    fmt.Printf("Stderr: %s\n", procErr.Stderr)
}
```

---

### ClaudeSDKError

```go
type ClaudeSDKError struct {
    Message string
    Cause   error
}
```

Base error type. Implements `error` and `Unwrap()`.

---

### CLIConnectionError

```go
type CLIConnectionError struct {
    ClaudeSDKError
}
```

Raised when unable to connect to Claude Code.

---

### CLINotFoundError

```go
type CLINotFoundError struct {
    CLIConnectionError
    CLIPath string
}
```

Raised when Claude Code CLI is not found.

---

### ProcessError

```go
type ProcessError struct {
    ClaudeSDKError
    ExitCode int
    Stderr   string
}
```

Raised when the CLI process fails.

---

### JSONDecodeError

```go
type JSONDecodeError struct {
    ClaudeSDKError
    Line          string
    OriginalError error
}
```

Raised when unable to decode JSON from CLI output.

---

### MessageParseError

```go
type MessageParseError struct {
    ClaudeSDKError
    Data map[string]any
}
```

Raised when unable to parse a message.

---

## Constants

### Version

```go
const Version = "0.1.0"
```

SDK version string.
