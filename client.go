package claude

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/afsharalex/claude-agent-sdk-go/internal/protocol"
	"github.com/afsharalex/claude-agent-sdk-go/internal/transport"
	"github.com/afsharalex/claude-agent-sdk-go/internal/types"
)

// Client provides bidirectional, interactive conversations with Claude Code.
//
// This client provides full control over the conversation flow with support
// for streaming, interrupts, and dynamic message sending. For simple one-shot
// queries, consider using the Query function instead.
//
// Key features:
//   - Bidirectional: Send and receive messages at any time
//   - Stateful: Maintains conversation context across messages
//   - Interactive: Send follow-ups based on responses
//   - Control flow: Support for interrupts and session management
//
// When to use Client:
//   - Building chat interfaces or conversational UIs
//   - Interactive debugging or exploration sessions
//   - Multi-turn conversations with context
//   - When you need to react to Claude's responses
//   - Real-time applications with user input
//   - When you need interrupt capabilities
//
// When to use Query instead:
//   - Simple one-off questions
//   - Batch processing of prompts
//   - Fire-and-forget automation scripts
//   - When all inputs are known upfront
//   - Stateless operations
type Client struct {
	options   *Options
	transport transport.Transport
	query     *protocol.Query

	// Message channel for receiving parsed messages
	messageCh chan Message
	errorCh   chan error

	mu        sync.Mutex
	connected bool
	sessionID string
}

// NewClient creates a new Claude SDK client.
func NewClient(opts ...Option) *Client {
	return &Client{
		options:   NewOptions(opts...),
		messageCh: make(chan Message, 100),
		errorCh:   make(chan error, 1),
		sessionID: "default",
	}
}

// Connect connects to Claude Code.
//
// If prompt is provided, it will be used as the initial message or stream.
// If prompt is empty, the connection is established without sending an initial message.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	// Validate canUseTool settings
	if c.options.CanUseTool != nil && c.options.PermissionPromptToolName != "" {
		return NewClaudeSDKError("can_use_tool callback cannot be used with permission_prompt_tool_name")
	}

	// Auto-set permission_prompt_tool_name if canUseTool is provided
	if c.options.CanUseTool != nil {
		c.options.PermissionPromptToolName = "stdio"
	}

	// Convert options to transport options
	transportOpts := toTransportOptions(c.options)

	// Create transport - streaming mode for Client
	t, err := transport.NewSubprocessTransport("", true, transportOpts)
	if err != nil {
		return err
	}
	c.transport = t

	// Connect transport
	if err := c.transport.Connect(ctx); err != nil {
		return err
	}

	// Extract SDK MCP servers (convert to internal types)
	var sdkMCPServers map[string]*types.MCPServer
	if servers, ok := c.options.MCPServers.(map[string]MCPServerConfig); ok {
		sdkMCPServers = toInternalMCPServers(servers)
	}

	// Create query handler
	c.query = protocol.NewQuery(protocol.QueryConfig{
		Transport:       c.transport,
		IsStreamingMode: true,
		CanUseTool:      toInternalCanUseTool(c.options.CanUseTool),
		Hooks:           toInternalHooks(c.options.Hooks),
		SDKMCPServers:   sdkMCPServers,
	})

	// Start reading messages
	c.query.Start(ctx)

	// Initialize
	if _, err := c.query.Initialize(ctx); err != nil {
		_ = c.query.Close()
		return err
	}

	c.connected = true

	// Start message processing in background
	go c.processMessages()

	return nil
}

// processMessages reads from the query and sends parsed messages to the channel.
func (c *Client) processMessages() {
	defer close(c.messageCh)
	defer close(c.errorCh)

	for data := range c.query.ReceiveMessages() {
		if data["type"] == "end" {
			return
		}
		if data["type"] == "error" {
			errMsg, _ := data["error"].(string)
			c.errorCh <- NewClaudeSDKError(errMsg)
			return
		}

		msg, err := ParseMessage(data)
		if err != nil {
			c.errorCh <- err
			continue
		}

		c.messageCh <- msg
	}
}

// Query sends a new query to Claude.
//
// The prompt can be a simple string message. For more complex messages,
// use QueryMessage.
func (c *Client) Query(ctx context.Context, prompt string) error {
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		return NewCLIConnectionError("Not connected. Call Connect() first.")
	}
	c.mu.Unlock()

	message := map[string]any{
		"type": "user",
		"message": map[string]any{
			"role":    "user",
			"content": prompt,
		},
		"parent_tool_use_id": nil,
		"session_id":         c.sessionID,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return c.transport.Write(ctx, string(data)+"\n")
}

// QueryMessage sends a structured message to Claude.
func (c *Client) QueryMessage(ctx context.Context, message map[string]any) error {
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		return NewCLIConnectionError("Not connected. Call Connect() first.")
	}
	c.mu.Unlock()

	// Ensure session_id is set
	if _, ok := message["session_id"]; !ok {
		message["session_id"] = c.sessionID
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return c.transport.Write(ctx, string(data)+"\n")
}

// Messages returns a channel for receiving messages from Claude.
func (c *Client) Messages() <-chan Message {
	return c.messageCh
}

// Errors returns a channel for receiving errors.
func (c *Client) Errors() <-chan error {
	return c.errorCh
}

// ReceiveResponse receives messages until and including a ResultMessage.
//
// This is a convenience method for single-response workflows. It returns
// a channel that yields all messages and closes after a ResultMessage.
func (c *Client) ReceiveResponse(ctx context.Context) <-chan Message {
	ch := make(chan Message, 100)

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-c.messageCh:
				if !ok {
					return
				}

				ch <- msg

				if _, isResult := msg.(*ResultMessage); isResult {
					return
				}
			}
		}
	}()

	return ch
}

// Interrupt sends an interrupt signal to Claude.
func (c *Client) Interrupt(ctx context.Context) error {
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		return NewCLIConnectionError("Not connected. Call Connect() first.")
	}
	c.mu.Unlock()

	return c.query.Interrupt(ctx)
}

// SetPermissionMode changes the permission mode during conversation.
//
// Valid modes:
//   - PermissionModeDefault: CLI prompts for dangerous tools
//   - PermissionModeAcceptEdits: Auto-accept file edits
//   - PermissionModeBypassPermissions: Allow all tools (use with caution)
func (c *Client) SetPermissionMode(ctx context.Context, mode PermissionMode) error {
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		return NewCLIConnectionError("Not connected. Call Connect() first.")
	}
	c.mu.Unlock()

	return c.query.SetPermissionMode(ctx, types.PermissionMode(mode))
}

// SetModel changes the AI model during conversation.
//
// Examples: "claude-sonnet-4-5", "claude-opus-4-1-20250805"
func (c *Client) SetModel(ctx context.Context, model string) error {
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		return NewCLIConnectionError("Not connected. Call Connect() first.")
	}
	c.mu.Unlock()

	return c.query.SetModel(ctx, model)
}

// RewindFiles rewinds tracked files to their state at a specific user message.
//
// Requires:
//   - EnableFileCheckpointing to be true in options
//   - ExtraArgs with "replay-user-messages" to receive UserMessage objects with UUIDs
func (c *Client) RewindFiles(ctx context.Context, userMessageID string) error {
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		return NewCLIConnectionError("Not connected. Call Connect() first.")
	}
	c.mu.Unlock()

	return c.query.RewindFiles(ctx, userMessageID)
}

// GetMCPStatus gets current MCP server connection status.
//
// Returns a map with "mcpServers" key containing a list of server status objects.
func (c *Client) GetMCPStatus(ctx context.Context) (map[string]any, error) {
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		return nil, NewCLIConnectionError("Not connected. Call Connect() first.")
	}
	c.mu.Unlock()

	return c.query.GetMCPStatus(ctx)
}

// GetServerInfo returns server initialization info.
//
// Returns information from the Claude Code server including available commands
// and output styles, or nil if not available.
func (c *Client) GetServerInfo() map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.query == nil {
		return nil
	}
	return c.query.InitResult()
}

// Close disconnects from Claude Code.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	c.connected = false

	if c.query != nil {
		_ = c.query.Close()
		c.query = nil
	}

	c.transport = nil
	return nil
}

// SetSessionID sets the session ID for subsequent queries.
func (c *Client) SetSessionID(sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sessionID = sessionID
}
