package protocol

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/afsharalex/claude-agent-sdk-go/internal/transport"
	"github.com/afsharalex/claude-agent-sdk-go/internal/types"
)

// Query handles bidirectional control protocol on top of Transport.
type Query struct {
	transport       transport.Transport
	isStreamingMode bool
	canUseTool      types.CanUseToolFunc
	hooks           map[types.HookEvent][]types.HookMatcher
	sdkMCPServers   map[string]*types.MCPServer

	pendingResponses sync.Map
	hookCallbacks    map[string]types.HookCallback
	nextCallbackID   atomic.Int64
	requestCounter   atomic.Int64

	messageChan     chan map[string]any
	initialized     bool
	closed          atomic.Bool
	initResult      map[string]any
	firstResultOnce sync.Once
	firstResultCh   chan struct{}

	initTimeout        time.Duration
	streamCloseTimeout time.Duration

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// QueryConfig configures Query behavior.
type QueryConfig struct {
	Transport          transport.Transport
	IsStreamingMode    bool
	CanUseTool         types.CanUseToolFunc
	Hooks              map[types.HookEvent][]types.HookMatcher
	SDKMCPServers      map[string]*types.MCPServer
	InitializeTimeout  time.Duration
	StreamCloseTimeout time.Duration
}

// NewQuery creates a new Query with the given configuration.
func NewQuery(cfg QueryConfig) *Query {
	ctx, cancel := context.WithCancel(context.Background())

	initTimeout := cfg.InitializeTimeout
	if initTimeout == 0 {
		initTimeout = 60 * time.Second
	}

	streamCloseTimeout := cfg.StreamCloseTimeout
	if streamCloseTimeout == 0 {
		streamCloseTimeout = 60 * time.Second
	}

	return &Query{
		transport:          cfg.Transport,
		isStreamingMode:    cfg.IsStreamingMode,
		canUseTool:         cfg.CanUseTool,
		hooks:              cfg.Hooks,
		sdkMCPServers:      cfg.SDKMCPServers,
		hookCallbacks:      make(map[string]types.HookCallback),
		messageChan:        make(chan map[string]any, 100),
		firstResultCh:      make(chan struct{}),
		initTimeout:        initTimeout,
		streamCloseTimeout: streamCloseTimeout,
		ctx:                ctx,
		cancel:             cancel,
	}
}

// Initialize initializes the control protocol if in streaming mode.
func (q *Query) Initialize(ctx context.Context) (map[string]any, error) {
	if !q.isStreamingMode {
		return nil, nil
	}

	hooksConfig := make(map[string]any)
	for event, matchers := range q.hooks {
		if len(matchers) > 0 {
			matcherConfigs := make([]map[string]any, 0, len(matchers))
			for _, matcher := range matchers {
				callbackIDs := make([]string, 0, len(matcher.Hooks))
				for _, callback := range matcher.Hooks {
					callbackID := fmt.Sprintf("hook_%d", q.nextCallbackID.Add(1))
					q.hookCallbacks[callbackID] = callback
					callbackIDs = append(callbackIDs, callbackID)
				}
				matcherConfig := map[string]any{
					"matcher":         matcher.Matcher,
					"hookCallbackIds": callbackIDs,
				}
				if matcher.Timeout > 0 {
					matcherConfig["timeout"] = matcher.Timeout
				}
				matcherConfigs = append(matcherConfigs, matcherConfig)
			}
			hooksConfig[string(event)] = matcherConfigs
		}
	}

	request := map[string]any{
		"subtype": RequestSubtypeInitialize,
	}
	if len(hooksConfig) > 0 {
		request["hooks"] = hooksConfig
	}

	response, err := q.sendControlRequest(ctx, request, q.initTimeout)
	if err != nil {
		return nil, err
	}

	q.initialized = true
	q.initResult = response
	return response, nil
}

// Start starts reading messages from transport.
func (q *Query) Start(ctx context.Context) {
	q.wg.Add(1)
	go q.readMessages(ctx)
}

func (q *Query) readMessages(ctx context.Context) {
	defer q.wg.Done()
	defer close(q.messageChan)

	resultCh := q.transport.ReadMessages(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-q.ctx.Done():
			return
		case result, ok := <-resultCh:
			if !ok {
				return
			}

			if result.Error != nil {
				q.pendingResponses.Range(func(key, value any) bool {
					ch := value.(chan map[string]any)
					close(ch)
					return true
				})
				q.messageChan <- map[string]any{"type": "error", "error": result.Error.Error()}
				return
			}

			message := result.Data
			msgType, _ := message["type"].(string)

			switch msgType {
			case "control_response":
				q.handleControlResponse(message)
			case "control_request":
				go q.handleControlRequest(ctx, message)
			case "control_cancel_request":
				continue
			default:
				if msgType == "result" {
					q.firstResultOnce.Do(func() {
						close(q.firstResultCh)
					})
				}

				select {
				case q.messageChan <- message:
				case <-ctx.Done():
					return
				case <-q.ctx.Done():
					return
				}
			}
		}
	}
}

func (q *Query) handleControlResponse(message map[string]any) {
	response, ok := message["response"].(map[string]any)
	if !ok {
		return
	}

	requestID, ok := response["request_id"].(string)
	if !ok {
		return
	}

	if ch, ok := q.pendingResponses.LoadAndDelete(requestID); ok {
		resultCh := ch.(chan map[string]any)
		resultCh <- response
		close(resultCh)
	}
}

func (q *Query) handleControlRequest(ctx context.Context, request map[string]any) {
	requestID, _ := request["request_id"].(string)
	requestData, ok := request["request"].(map[string]any)
	if !ok {
		q.sendErrorResponse(ctx, requestID, "Invalid request format")
		return
	}

	subtype, _ := requestData["subtype"].(string)
	var responseData map[string]any
	var err error

	switch subtype {
	case RequestSubtypeCanUseTool:
		responseData, err = q.handleCanUseTool(ctx, requestData)
	case RequestSubtypeHookCallback:
		responseData, err = q.handleHookCallback(ctx, requestData)
	case RequestSubtypeMCPMessage:
		responseData, err = q.handleMCPMessage(ctx, requestData)
	default:
		err = fmt.Errorf("unsupported control request subtype: %s", subtype)
	}

	if err != nil {
		q.sendErrorResponse(ctx, requestID, err.Error())
		return
	}

	q.sendSuccessResponse(ctx, requestID, responseData)
}

func (q *Query) handleCanUseTool(ctx context.Context, request map[string]any) (map[string]any, error) {
	if q.canUseTool == nil {
		return nil, fmt.Errorf("canUseTool callback is not provided")
	}

	toolName, _ := request["tool_name"].(string)
	input, _ := request["input"].(map[string]any)
	originalInput := input

	permCtx := types.ToolPermissionContext{
		Signal: nil,
	}

	result, err := q.canUseTool(ctx, toolName, input, permCtx)
	if err != nil {
		return nil, err
	}

	if result.IsAllow() {
		allow := result.(types.PermissionResultAllow)
		response := map[string]any{
			"behavior": "allow",
		}
		if allow.UpdatedInput != nil {
			response["updatedInput"] = allow.UpdatedInput
		} else {
			response["updatedInput"] = originalInput
		}
		if allow.UpdatedPermissions != nil {
			permissions := make([]map[string]any, len(allow.UpdatedPermissions))
			for i, p := range allow.UpdatedPermissions {
				permissions[i] = p.ToMap()
			}
			response["updatedPermissions"] = permissions
		}
		return response, nil
	}

	deny := result.(types.PermissionResultDeny)
	response := map[string]any{
		"behavior": "deny",
		"message":  deny.Message,
	}
	if deny.Interrupt {
		response["interrupt"] = true
	}
	return response, nil
}

func (q *Query) handleHookCallback(ctx context.Context, request map[string]any) (map[string]any, error) {
	callbackID, _ := request["callback_id"].(string)
	callback, ok := q.hookCallbacks[callbackID]
	if !ok {
		return nil, fmt.Errorf("no hook callback found for ID: %s", callbackID)
	}

	inputData := request["input"]
	toolUseID, _ := request["tool_use_id"].(string)

	hookInput, err := parseHookInput(inputData)
	if err != nil {
		return nil, err
	}

	hookCtx := types.HookContext{
		Signal: nil,
	}

	output, err := callback(ctx, hookInput, toolUseID, hookCtx)
	if err != nil {
		return nil, err
	}

	return output.ToMap(), nil
}

func parseHookInput(data any) (types.HookInput, error) {
	m, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid hook input format")
	}

	eventName, _ := m["hook_event_name"].(string)
	base := types.BaseHookInput{
		SessionID:      getString(m, "session_id"),
		TranscriptPath: getString(m, "transcript_path"),
		Cwd:            getString(m, "cwd"),
		PermissionMode: getString(m, "permission_mode"),
	}

	switch eventName {
	case "PreToolUse":
		return types.PreToolUseHookInput{
			BaseHookInput: base,
			ToolName:      getString(m, "tool_name"),
			ToolInput:     getMap(m, "tool_input"),
		}, nil

	case "PostToolUse":
		return types.PostToolUseHookInput{
			BaseHookInput: base,
			ToolName:      getString(m, "tool_name"),
			ToolInput:     getMap(m, "tool_input"),
			ToolResponse:  m["tool_response"],
		}, nil

	case "PostToolUseFailure":
		return types.PostToolUseFailureHookInput{
			BaseHookInput: base,
			ToolName:      getString(m, "tool_name"),
			ToolInput:     getMap(m, "tool_input"),
			ToolUseID:     getString(m, "tool_use_id"),
			Error:         getString(m, "error"),
			IsInterrupt:   getBool(m, "is_interrupt"),
		}, nil

	case "UserPromptSubmit":
		return types.UserPromptSubmitHookInput{
			BaseHookInput: base,
			Prompt:        getString(m, "prompt"),
		}, nil

	case "Stop":
		return types.StopHookInput{
			BaseHookInput:  base,
			StopHookActive: getBool(m, "stop_hook_active"),
		}, nil

	case "SubagentStop":
		return types.SubagentStopHookInput{
			BaseHookInput:  base,
			StopHookActive: getBool(m, "stop_hook_active"),
		}, nil

	case "PreCompact":
		var customInstructions *string
		if ci, ok := m["custom_instructions"].(string); ok {
			customInstructions = &ci
		}
		return types.PreCompactHookInput{
			BaseHookInput:      base,
			Trigger:            types.PreCompactTrigger(getString(m, "trigger")),
			CustomInstructions: customInstructions,
		}, nil

	default:
		return nil, fmt.Errorf("unknown hook event name: %s", eventName)
	}
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getMap(m map[string]any, key string) map[string]any {
	if v, ok := m[key].(map[string]any); ok {
		return v
	}
	return nil
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func (q *Query) handleMCPMessage(ctx context.Context, request map[string]any) (map[string]any, error) {
	serverName, _ := request["server_name"].(string)
	message, _ := request["message"].(map[string]any)

	if serverName == "" || message == nil {
		return nil, fmt.Errorf("missing server_name or message for MCP request")
	}

	server, ok := q.sdkMCPServers[serverName]
	if !ok {
		return map[string]any{
			"mcp_response": map[string]any{
				"jsonrpc": "2.0",
				"id":      message["id"],
				"error": map[string]any{
					"code":    -32601,
					"message": fmt.Sprintf("Server '%s' not found", serverName),
				},
			},
		}, nil
	}

	method, _ := message["method"].(string)
	params, _ := message["params"].(map[string]any)

	var result map[string]any

	switch method {
	case "initialize":
		result = map[string]any{
			"jsonrpc": "2.0",
			"id":      message["id"],
			"result": map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
				"serverInfo": map[string]any{
					"name":    server.Name,
					"version": server.Version,
				},
			},
		}

	case "tools/list":
		toolsList := make([]map[string]any, 0, len(server.Tools))
		for _, tool := range server.Tools {
			toolsList = append(toolsList, map[string]any{
				"name":        tool.Name,
				"description": tool.Description,
				"inputSchema": tool.InputSchema,
			})
		}
		result = map[string]any{
			"jsonrpc": "2.0",
			"id":      message["id"],
			"result": map[string]any{
				"tools": toolsList,
			},
		}

	case "tools/call":
		name, _ := params["name"].(string)
		arguments, _ := params["arguments"].(map[string]any)

		var tool *types.MCPTool
		for i := range server.Tools {
			if server.Tools[i].Name == name {
				tool = &server.Tools[i]
				break
			}
		}

		if tool == nil {
			result = map[string]any{
				"jsonrpc": "2.0",
				"id":      message["id"],
				"error": map[string]any{
					"code":    -32601,
					"message": fmt.Sprintf("Tool '%s' not found", name),
				},
			}
		} else {
			toolResult, err := tool.Handler(ctx, arguments)
			if err != nil {
				result = map[string]any{
					"jsonrpc": "2.0",
					"id":      message["id"],
					"error": map[string]any{
						"code":    -32603,
						"message": err.Error(),
					},
				}
			} else {
				content := make([]map[string]any, 0, len(toolResult.Content))
				for _, c := range toolResult.Content {
					item := map[string]any{"type": c.Type}
					switch c.Type {
					case "text":
						item["text"] = c.Text
					case "image":
						item["data"] = c.Data
						item["mimeType"] = c.MimeType
					}
					content = append(content, item)
				}
				responseData := map[string]any{"content": content}
				if toolResult.IsError {
					responseData["is_error"] = true
				}
				result = map[string]any{
					"jsonrpc": "2.0",
					"id":      message["id"],
					"result":  responseData,
				}
			}
		}

	case "notifications/initialized":
		result = map[string]any{"jsonrpc": "2.0", "result": map[string]any{}}

	default:
		result = map[string]any{
			"jsonrpc": "2.0",
			"id":      message["id"],
			"error": map[string]any{
				"code":    -32601,
				"message": fmt.Sprintf("Method '%s' not found", method),
			},
		}
	}

	return map[string]any{"mcp_response": result}, nil
}

func (q *Query) sendControlRequest(ctx context.Context, request map[string]any, timeout time.Duration) (map[string]any, error) {
	if !q.isStreamingMode {
		return nil, fmt.Errorf("control requests require streaming mode")
	}

	randBytes := make([]byte, 4)
	_, _ = rand.Read(randBytes)
	requestID := fmt.Sprintf("req_%d_%s", q.requestCounter.Add(1), hex.EncodeToString(randBytes))

	responseCh := make(chan map[string]any, 1)
	q.pendingResponses.Store(requestID, responseCh)

	controlRequest := map[string]any{
		"type":       "control_request",
		"request_id": requestID,
		"request":    request,
	}

	data, err := json.Marshal(controlRequest)
	if err != nil {
		q.pendingResponses.Delete(requestID)
		return nil, err
	}

	if err := q.transport.Write(ctx, string(data)+"\n"); err != nil {
		q.pendingResponses.Delete(requestID)
		return nil, err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case <-timeoutCtx.Done():
		q.pendingResponses.Delete(requestID)
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("control request timeout: %s", request["subtype"])
		}
		return nil, timeoutCtx.Err()

	case response, ok := <-responseCh:
		if !ok {
			return nil, fmt.Errorf("control request channel closed unexpectedly")
		}

		if response["subtype"] == "error" {
			errMsg, _ := response["error"].(string)
			return nil, fmt.Errorf("%s", errMsg)
		}

		if respData, ok := response["response"].(map[string]any); ok {
			return respData, nil
		}
		return map[string]any{}, nil
	}
}

func (q *Query) sendSuccessResponse(ctx context.Context, requestID string, response map[string]any) {
	resp := SDKControlResponse{
		Type: "control_response",
		Response: ControlResponseData{
			Subtype:   "success",
			RequestID: requestID,
			Response:  response,
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return
	}

	_ = q.transport.Write(ctx, string(data)+"\n")
}

func (q *Query) sendErrorResponse(ctx context.Context, requestID string, errMsg string) {
	resp := SDKControlResponse{
		Type: "control_response",
		Response: ControlResponseData{
			Subtype:   "error",
			RequestID: requestID,
			Error:     errMsg,
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return
	}

	_ = q.transport.Write(ctx, string(data)+"\n")
}

// GetMCPStatus gets current MCP server connection status.
func (q *Query) GetMCPStatus(ctx context.Context) (map[string]any, error) {
	return q.sendControlRequest(ctx, map[string]any{"subtype": RequestSubtypeMCPStatus}, 60*time.Second)
}

// Interrupt sends an interrupt control request.
func (q *Query) Interrupt(ctx context.Context) error {
	_, err := q.sendControlRequest(ctx, map[string]any{"subtype": RequestSubtypeInterrupt}, 60*time.Second)
	return err
}

// SetPermissionMode changes the permission mode.
func (q *Query) SetPermissionMode(ctx context.Context, mode types.PermissionMode) error {
	_, err := q.sendControlRequest(ctx, map[string]any{
		"subtype": RequestSubtypeSetPermissionMode,
		"mode":    string(mode),
	}, 60*time.Second)
	return err
}

// SetModel changes the AI model.
func (q *Query) SetModel(ctx context.Context, model string) error {
	req := map[string]any{
		"subtype": RequestSubtypeSetModel,
	}
	if model != "" {
		req["model"] = model
	}
	_, err := q.sendControlRequest(ctx, req, 60*time.Second)
	return err
}

// RewindFiles rewinds tracked files to their state at a specific user message.
func (q *Query) RewindFiles(ctx context.Context, userMessageID string) error {
	_, err := q.sendControlRequest(ctx, map[string]any{
		"subtype":         RequestSubtypeRewindFiles,
		"user_message_id": userMessageID,
	}, 60*time.Second)
	return err
}

// StreamInput streams input messages to transport.
func (q *Query) StreamInput(ctx context.Context, messages <-chan map[string]any) {
	for {
		select {
		case <-ctx.Done():
			_ = q.transport.EndInput()
			return
		case <-q.ctx.Done():
			_ = q.transport.EndInput()
			return
		case msg, ok := <-messages:
			if !ok {
				if len(q.sdkMCPServers) > 0 || len(q.hooks) > 0 {
					select {
					case <-q.firstResultCh:
					case <-time.After(q.streamCloseTimeout):
					case <-ctx.Done():
					}
				}
				_ = q.transport.EndInput()
				return
			}

			data, err := json.Marshal(msg)
			if err != nil {
				continue
			}
			_ = q.transport.Write(ctx, string(data)+"\n")
		}
	}
}

// ReceiveMessages returns a channel for receiving SDK messages.
func (q *Query) ReceiveMessages() <-chan map[string]any {
	return q.messageChan
}

// InitResult returns the initialization result.
func (q *Query) InitResult() map[string]any {
	return q.initResult
}

// Close closes the query and transport.
func (q *Query) Close() error {
	if q.closed.Swap(true) {
		return nil
	}
	q.cancel()
	q.wg.Wait()
	return q.transport.Close()
}
