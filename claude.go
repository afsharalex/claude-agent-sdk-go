// Package claude provides a Go SDK for interacting with Claude Code.
//
// This package provides two main APIs:
//
//   - Query: A simple function for one-shot or unidirectional streaming interactions
//   - Client: A full-featured client for bidirectional, interactive conversations
//
// # Simple Query Example
//
//	ctx := context.Background()
//	messages, errs := claude.Query(ctx, "What is the capital of France?")
//
//	for {
//		select {
//		case msg, ok := <-messages:
//			if !ok {
//				return
//			}
//			switch m := msg.(type) {
//			case *claude.AssistantMessage:
//				for _, block := range m.Content {
//					if text, ok := block.(claude.TextBlock); ok {
//						fmt.Println(text.Text)
//					}
//				}
//			case *claude.ResultMessage:
//				fmt.Printf("Cost: $%.4f\n", *m.TotalCostUSD)
//			}
//		case err := <-errs:
//			log.Fatal(err)
//		}
//	}
//
// # Interactive Client Example
//
//	client := claude.NewClient(
//		claude.WithCwd("/path/to/project"),
//	)
//
//	if err := client.Connect(ctx); err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	if err := client.Query(ctx, "Help me understand this codebase"); err != nil {
//		log.Fatal(err)
//	}
//
//	for msg := range client.Messages() {
//		// Handle messages...
//	}
package claude

import (
	"context"

	"github.com/afsharalex/claude-agent-sdk-go/internal/protocol"
	"github.com/afsharalex/claude-agent-sdk-go/internal/transport"
	"github.com/afsharalex/claude-agent-sdk-go/internal/types"
)

// Version is the SDK version.
const Version = "0.1.0"

// toTransportOptions converts Options to internal transport.Options.
func toTransportOptions(o *Options) *transport.Options {
	var betas []string
	for _, b := range o.Betas {
		betas = append(betas, string(b))
	}

	var settingSources []string
	for _, s := range o.SettingSources {
		settingSources = append(settingSources, string(s))
	}

	var sandbox *transport.SandboxSettings
	if o.Sandbox != nil {
		sandbox = &transport.SandboxSettings{
			Enabled:                   o.Sandbox.Enabled,
			AutoAllowBashIfSandboxed:  o.Sandbox.AutoAllowBashIfSandboxed,
			ExcludedCommands:          o.Sandbox.ExcludedCommands,
			AllowUnsandboxedCommands:  o.Sandbox.AllowUnsandboxedCommands,
			EnableWeakerNestedSandbox: o.Sandbox.EnableWeakerNestedSandbox,
		}
		if o.Sandbox.Network != nil {
			sandbox.Network = &transport.SandboxNetworkConfig{
				AllowUnixSockets:    o.Sandbox.Network.AllowUnixSockets,
				AllowAllUnixSockets: o.Sandbox.Network.AllowAllUnixSockets,
				AllowLocalBinding:   o.Sandbox.Network.AllowLocalBinding,
				HTTPProxyPort:       o.Sandbox.Network.HTTPProxyPort,
				SOCKSProxyPort:      o.Sandbox.Network.SOCKSProxyPort,
			}
		}
		if o.Sandbox.IgnoreViolations != nil {
			sandbox.IgnoreViolations = &transport.SandboxIgnoreViolations{
				File:    o.Sandbox.IgnoreViolations.File,
				Network: o.Sandbox.IgnoreViolations.Network,
			}
		}
	}

	agents := make(map[string]transport.AgentDefinition)
	for k, v := range o.Agents {
		agents[k] = transport.AgentDefinition{
			Description: v.Description,
			Prompt:      v.Prompt,
			Tools:       v.Tools,
			Model:       v.Model,
		}
	}

	var plugins []transport.PluginConfig
	for _, p := range o.Plugins {
		plugins = append(plugins, transport.PluginConfig{
			Type: string(p.Type),
			Path: p.Path,
		})
	}

	// Convert MCP servers to map[string]any for transport
	var mcpServers any
	if o.MCPServers != nil {
		if servers, ok := o.MCPServers.(map[string]MCPServerConfig); ok {
			serversMap := make(map[string]any)
			for name, config := range servers {
				switch c := config.(type) {
				case MCPStdioServerConfig:
					serversMap[name] = map[string]any{
						"type":    c.GetType(),
						"command": c.Command,
						"args":    c.Args,
						"env":     c.Env,
					}
				case MCPSSEServerConfig:
					serversMap[name] = map[string]any{
						"type":    "sse",
						"url":     c.URL,
						"headers": c.Headers,
					}
				case MCPHTTPServerConfig:
					serversMap[name] = map[string]any{
						"type":    "http",
						"url":     c.URL,
						"headers": c.Headers,
					}
				case MCPSDKServerConfig:
					serversMap[name] = map[string]any{
						"type": "sdk",
						"name": c.Name,
					}
				}
			}
			mcpServers = serversMap
		} else if path, ok := o.MCPServers.(string); ok {
			mcpServers = path
		}
	}

	// Combine SystemPrompt and AppendSystemPrompt
	var systemPrompt any = o.SystemPrompt
	if o.AppendSystemPrompt != "" {
		if systemPrompt == nil {
			systemPrompt = o.AppendSystemPrompt
		} else if sp, ok := systemPrompt.(string); ok {
			systemPrompt = sp + "\n" + o.AppendSystemPrompt
		}
		// Note: If SystemPrompt is a SystemPromptPreset, AppendSystemPrompt is ignored
		// as presets have their own append mechanism
	}

	return &transport.Options{
		Tools:                    o.Tools,
		AllowedTools:             o.AllowedTools,
		SystemPrompt:             systemPrompt,
		MCPServers:               mcpServers,
		PermissionMode:           string(o.PermissionMode),
		ContinueConversation:     o.ContinueConversation,
		Resume:                   o.Resume,
		MaxTurns:                 o.MaxTurns,
		MaxBudgetUSD:             o.MaxBudgetUSD,
		DisallowedTools:          o.DisallowedTools,
		Model:                    o.Model,
		FallbackModel:            o.FallbackModel,
		Betas:                    betas,
		PermissionPromptToolName: o.PermissionPromptToolName,
		Cwd:                      o.Cwd,
		CLIPath:                  o.CLIPath,
		Settings:                 o.Settings,
		AddDirs:                  o.AddDirs,
		Env:                      o.Env,
		ExtraArgs:                o.ExtraArgs,
		MaxBufferSize:            o.MaxBufferSize,
		DebugStderr:              o.DebugStderr,
		Stderr:                   o.Stderr,
		User:                     o.User,
		IncludePartialMessages:   o.IncludePartialMessages,
		ForkSession:              o.ForkSession,
		Agents:                   agents,
		SettingSources:           settingSources,
		Sandbox:                  sandbox,
		Plugins:                  plugins,
		MaxThinkingTokens:        o.MaxThinkingTokens,
		OutputFormat:             o.OutputFormat,
		EnableFileCheckpointing:  o.EnableFileCheckpointing,
	}
}

// toInternalHooks converts public hook types to internal types.
func toInternalHooks(hooks map[HookEvent][]HookMatcher) map[types.HookEvent][]types.HookMatcher {
	if hooks == nil {
		return nil
	}

	result := make(map[types.HookEvent][]types.HookMatcher)
	for event, matchers := range hooks {
		var internalMatchers []types.HookMatcher
		for _, m := range matchers {
			var internalHooks []types.HookCallback
			for _, h := range m.Hooks {
				// Capture h in closure
				hookFunc := h
				internalHooks = append(internalHooks, func(ctx context.Context, input types.HookInput, toolUseID string, hookCtx types.HookContext) (types.HookOutput, error) {
					// Convert internal input to public input
					publicInput := toPublicHookInput(input)
					publicCtx := HookContext{Signal: hookCtx.Signal}

					output, err := hookFunc(ctx, publicInput, toolUseID, publicCtx)
					if err != nil {
						return types.HookOutput{}, err
					}

					return toInternalHookOutput(output), nil
				})
			}
			internalMatchers = append(internalMatchers, types.HookMatcher{
				Matcher: m.Matcher,
				Hooks:   internalHooks,
				Timeout: m.Timeout,
			})
		}
		result[types.HookEvent(event)] = internalMatchers
	}
	return result
}

// toPublicHookInput converts internal hook input to public type.
func toPublicHookInput(input types.HookInput) HookInput {
	switch v := input.(type) {
	case types.PreToolUseHookInput:
		return PreToolUseHookInput{
			BaseHookInput: BaseHookInput{
				SessionID:      v.SessionID,
				TranscriptPath: v.TranscriptPath,
				Cwd:            v.Cwd,
				PermissionMode: v.PermissionMode,
			},
			ToolName:  v.ToolName,
			ToolInput: v.ToolInput,
		}
	case types.PostToolUseHookInput:
		return PostToolUseHookInput{
			BaseHookInput: BaseHookInput{
				SessionID:      v.SessionID,
				TranscriptPath: v.TranscriptPath,
				Cwd:            v.Cwd,
				PermissionMode: v.PermissionMode,
			},
			ToolName:     v.ToolName,
			ToolInput:    v.ToolInput,
			ToolResponse: v.ToolResponse,
		}
	case types.UserPromptSubmitHookInput:
		return UserPromptSubmitHookInput{
			BaseHookInput: BaseHookInput{
				SessionID:      v.SessionID,
				TranscriptPath: v.TranscriptPath,
				Cwd:            v.Cwd,
				PermissionMode: v.PermissionMode,
			},
			Prompt: v.Prompt,
		}
	case types.StopHookInput:
		return StopHookInput{
			BaseHookInput: BaseHookInput{
				SessionID:      v.SessionID,
				TranscriptPath: v.TranscriptPath,
				Cwd:            v.Cwd,
				PermissionMode: v.PermissionMode,
			},
			StopHookActive: v.StopHookActive,
		}
	default:
		return nil
	}
}

// toInternalHookOutput converts public hook output to internal type.
func toInternalHookOutput(output HookOutput) types.HookOutput {
	result := types.HookOutput{
		Async:          output.Async,
		AsyncTimeout:   output.AsyncTimeout,
		Continue:       output.Continue,
		SuppressOutput: output.SuppressOutput,
		StopReason:     output.StopReason,
		Decision:       types.HookDecision(output.Decision),
		SystemMessage:  output.SystemMessage,
		Reason:         output.Reason,
	}

	if output.HookSpecificOutput != nil {
		switch v := output.HookSpecificOutput.(type) {
		case PreToolUseHookSpecificOutput:
			result.HookEventName = types.HookEvent(v.HookEventName)
			result.PermissionDecision = types.HookPermissionDecision(v.PermissionDecision)
			result.UpdatedInput = v.UpdatedInput
		case PostToolUseHookSpecificOutput:
			result.HookEventName = types.HookEvent(v.HookEventName)
			result.AdditionalContext = v.AdditionalContext
		}
	}

	return result
}

// toInternalCanUseTool converts public callback to internal type.
func toInternalCanUseTool(fn CanUseToolFunc) types.CanUseToolFunc {
	if fn == nil {
		return nil
	}
	return func(ctx context.Context, toolName string, input map[string]any, permCtx types.ToolPermissionContext) (types.PermissionResult, error) {
		publicCtx := ToolPermissionContext{
			Signal: permCtx.Signal,
		}

		result, err := fn(ctx, toolName, input, publicCtx)
		if err != nil {
			return nil, err
		}

		switch r := result.(type) {
		case PermissionResultAllow:
			return types.PermissionResultAllow{
				UpdatedInput: r.UpdatedInput,
			}, nil
		case PermissionResultDeny:
			return types.PermissionResultDeny{
				Message:   r.Message,
				Interrupt: r.Interrupt,
			}, nil
		default:
			return nil, NewClaudeSDKError("invalid permission result type")
		}
	}
}

// toInternalMCPServers converts public MCP servers to internal type.
func toInternalMCPServers(servers map[string]MCPServerConfig) map[string]*types.MCPServer {
	if servers == nil {
		return nil
	}

	result := make(map[string]*types.MCPServer)
	for name, config := range servers {
		if sdkConfig, ok := config.(MCPSDKServerConfig); ok && sdkConfig.Server != nil {
			var tools []types.MCPTool
			for _, t := range sdkConfig.Server.Tools() {
				tools = append(tools, types.MCPTool{
					Name:        t.Name,
					Description: t.Description,
					InputSchema: t.InputSchema,
					Handler: func(ctx context.Context, args map[string]any) (types.MCPToolResult, error) {
						result, err := t.Handler(ctx, args)
						if err != nil {
							return types.MCPToolResult{}, err
						}
						var content []types.MCPContent
						for _, c := range result.Content {
							content = append(content, types.MCPContent{
								Type:     c.Type,
								Text:     c.Text,
								Data:     c.Data,
								MimeType: c.MimeType,
							})
						}
						return types.MCPToolResult{
							Content: content,
							IsError: result.IsError,
						}, nil
					},
				})
			}
			result[name] = &types.MCPServer{
				Name:    sdkConfig.Server.Name(),
				Version: sdkConfig.Server.Version(),
				Tools:   tools,
			}
		}
	}
	return result
}

// Query performs a one-shot query to Claude Code.
func Query(ctx context.Context, prompt string, opts ...Option) (<-chan Message, <-chan error) {
	messages := make(chan Message, 100)
	errors := make(chan error, 1)

	go func() {
		defer close(messages)
		defer close(errors)

		options := NewOptions(opts...)
		transportOpts := toTransportOptions(options)

		t, err := transport.NewSubprocessTransport(prompt, false, transportOpts)
		if err != nil {
			errors <- err
			return
		}

		if err := t.Connect(ctx); err != nil {
			errors <- err
			return
		}
		defer func() { _ = t.Close() }()

		q := protocol.NewQuery(protocol.QueryConfig{
			Transport:       t,
			IsStreamingMode: false,
		})
		defer func() { _ = q.Close() }()

		q.Start(ctx)

		for data := range q.ReceiveMessages() {
			if data["type"] == "end" {
				return
			}
			if data["type"] == "error" {
				errMsg, _ := data["error"].(string)
				errors <- NewClaudeSDKError(errMsg)
				return
			}

			msg, err := ParseMessage(data)
			if err != nil {
				errors <- err
				return
			}

			select {
			case messages <- msg:
			case <-ctx.Done():
				errors <- ctx.Err()
				return
			}
		}
	}()

	return messages, errors
}

// QueryStreaming performs a streaming query to Claude Code.
func QueryStreaming(ctx context.Context, inputCh <-chan map[string]any, opts ...Option) (<-chan Message, <-chan error) {
	messages := make(chan Message, 100)
	errors := make(chan error, 1)

	go func() {
		defer close(messages)
		defer close(errors)

		options := NewOptions(opts...)

		if options.CanUseTool != nil && options.PermissionPromptToolName != "" {
			errors <- NewClaudeSDKError("can_use_tool callback cannot be used with permission_prompt_tool_name")
			return
		}

		if options.CanUseTool != nil {
			options.PermissionPromptToolName = "stdio"
		}

		transportOpts := toTransportOptions(options)

		t, err := transport.NewSubprocessTransport("", true, transportOpts)
		if err != nil {
			errors <- err
			return
		}

		if err := t.Connect(ctx); err != nil {
			errors <- err
			return
		}
		defer func() { _ = t.Close() }()

		var sdkMCPServers map[string]*types.MCPServer
		if servers, ok := options.MCPServers.(map[string]MCPServerConfig); ok {
			sdkMCPServers = toInternalMCPServers(servers)
		}

		q := protocol.NewQuery(protocol.QueryConfig{
			Transport:       t,
			IsStreamingMode: true,
			CanUseTool:      toInternalCanUseTool(options.CanUseTool),
			Hooks:           toInternalHooks(options.Hooks),
			SDKMCPServers:   sdkMCPServers,
		})
		defer func() { _ = q.Close() }()

		q.Start(ctx)

		if _, err := q.Initialize(ctx); err != nil {
			errors <- err
			return
		}

		go q.StreamInput(ctx, inputCh)

		for data := range q.ReceiveMessages() {
			if data["type"] == "end" {
				return
			}
			if data["type"] == "error" {
				errMsg, _ := data["error"].(string)
				errors <- NewClaudeSDKError(errMsg)
				return
			}

			msg, err := ParseMessage(data)
			if err != nil {
				errors <- err
				return
			}

			select {
			case messages <- msg:
			case <-ctx.Done():
				errors <- ctx.Err()
				return
			}
		}
	}()

	return messages, errors
}

// WithClient creates a client, executes the function, and ensures cleanup.
// This is the recommended pattern for short-lived client operations.
//
// Example:
//
//	err := claude.WithClient(ctx, func(c *claude.Client) error {
//	    c.Query(ctx, "Hello")
//	    for msg := range c.ReceiveResponse(ctx) {
//	        // handle messages
//	    }
//	    return nil
//	}, claude.WithCwd("/path"))
func WithClient(ctx context.Context, fn func(*Client) error, opts ...Option) error {
	client := NewClient(opts...)

	if err := client.Connect(ctx); err != nil {
		return err
	}
	defer client.Close()

	return fn(client)
}

// QueryWithSession performs a one-shot query with a specific session ID.
// This is useful for maintaining conversation context across multiple
// independent queries without using the full Client interface.
//
// If sessionID is empty, a new session is created (default behavior).
func QueryWithSession(ctx context.Context, sessionID, prompt string, opts ...Option) (<-chan Message, <-chan error) {
	if sessionID != "" {
		opts = append(opts, WithResume(sessionID))
	}
	return Query(ctx, prompt, opts...)
}
