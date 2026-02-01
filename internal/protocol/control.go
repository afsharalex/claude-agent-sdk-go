// Package protocol provides control protocol handling for Claude SDK.
package protocol

// SDKControlRequest represents a control request from/to CLI.
type SDKControlRequest struct {
	Type      string         `json:"type"` // "control_request"
	RequestID string         `json:"request_id"`
	Request   map[string]any `json:"request"`
}

// SDKControlResponse represents a control response.
type SDKControlResponse struct {
	Type     string              `json:"type"` // "control_response"
	Response ControlResponseData `json:"response"`
}

// ControlResponseData is the response payload.
type ControlResponseData struct {
	Subtype   string         `json:"subtype"` // "success" or "error"
	RequestID string         `json:"request_id"`
	Response  map[string]any `json:"response,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// Request subtypes
const (
	RequestSubtypeInitialize        = "initialize"
	RequestSubtypeInterrupt         = "interrupt"
	RequestSubtypeCanUseTool        = "can_use_tool"
	RequestSubtypeSetPermissionMode = "set_permission_mode"
	RequestSubtypeSetModel          = "set_model"
	RequestSubtypeHookCallback      = "hook_callback"
	RequestSubtypeMCPMessage        = "mcp_message"
	RequestSubtypeMCPStatus         = "mcp_status"
	RequestSubtypeRewindFiles       = "rewind_files"
)

// PermissionRequest is the data for a can_use_tool control request.
type PermissionRequest struct {
	ToolName              string         `json:"tool_name"`
	Input                 map[string]any `json:"input"`
	PermissionSuggestions []any          `json:"permission_suggestions,omitempty"`
	BlockedPath           string         `json:"blocked_path,omitempty"`
}

// HookCallbackRequest is the data for a hook_callback control request.
type HookCallbackRequest struct {
	CallbackID string `json:"callback_id"`
	Input      any    `json:"input"`
	ToolUseID  string `json:"tool_use_id,omitempty"`
}

// MCPMessageRequest is the data for an mcp_message control request.
type MCPMessageRequest struct {
	ServerName string         `json:"server_name"`
	Message    map[string]any `json:"message"`
}
