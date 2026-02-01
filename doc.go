// Package claude provides a Go SDK for interacting with Claude Code.
//
// This SDK enables programmatic access to Claude Code from Go applications,
// supporting both simple one-shot queries and complex interactive sessions
// with custom tools and hooks.
//
// # Two Primary APIs
//
// The SDK provides two main ways to interact with Claude Code:
//
// Query is a function for one-shot queries, ideal for automation and scripting:
//
//	messages, errors := claude.Query(ctx, "What is 2 + 2?")
//	for msg := range messages {
//	    if m, ok := msg.(*claude.AssistantMessage); ok {
//	        for _, block := range m.Content {
//	            if text, ok := block.(claude.TextBlock); ok {
//	                fmt.Println(text.Text)
//	            }
//	        }
//	    }
//	}
//
// Client provides bidirectional, interactive conversations with streaming:
//
//	client := claude.NewClient(claude.WithCwd("/path/to/project"))
//	client.Connect(ctx)
//	defer client.Close()
//
//	client.Query(ctx, "Help me understand this codebase")
//	for msg := range client.Messages() {
//	    // Handle messages...
//	}
//
// # Configuration
//
// Both APIs accept functional options for configuration:
//
//	claude.Query(ctx, "prompt",
//	    claude.WithModel("claude-sonnet-4-5"),
//	    claude.WithMaxTurns(5),
//	    claude.WithSystemPrompt("You are helpful"),
//	)
//
// See the [Options] type and With* functions for all available options.
//
// # Custom Tools
//
// The Client API supports custom tools via in-process MCP servers:
//
//	tools := []claude.MCPTool{
//	    claude.Tool("add", "Add numbers", schema, handler),
//	}
//	server := claude.CreateSDKMCPServer("calc", "1.0.0", tools)
//	client := claude.NewClient(
//	    claude.WithMCPServers(map[string]claude.MCPServerConfig{
//	        "calc": server,
//	    }),
//	)
//
// # Hooks
//
// Hooks intercept and modify Claude's behavior at specific points:
//
//	hooks := map[claude.HookEvent][]claude.HookMatcher{
//	    claude.HookEventPreToolUse: {{
//	        Matcher: "Bash",
//	        Hooks:   []claude.HookCallback{myHook},
//	    }},
//	}
//	client := claude.NewClient(claude.WithHooks(hooks))
//
// # Message Types
//
// Messages implement the [Message] interface:
//   - [AssistantMessage]: Claude's responses with content blocks
//   - [UserMessage]: User input and tool results
//   - [SystemMessage]: System notifications
//   - [ResultMessage]: Query completion with cost/usage info
//
// Content blocks implement the [ContentBlock] interface:
//   - [TextBlock]: Plain text content
//   - [ToolUseBlock]: Tool invocation requests
//   - [ToolResultBlock]: Tool execution results
//   - [ThinkingBlock]: Extended thinking content
//
// # Error Handling
//
// The SDK defines specific error types for different failure modes:
//   - [ClaudeSDKError]: Base error type
//   - [CLINotFoundError]: Claude Code CLI not installed
//   - [CLIConnectionError]: Connection to CLI failed
//   - [ProcessError]: CLI process exited with error
//   - [JSONDecodeError]: Failed to parse CLI output
//
// # Requirements
//
// This SDK requires the Claude Code CLI to be installed. Install it with:
//
//	curl -fsSL https://claude.ai/install.sh | bash
package claude
