package claude

import (
	"testing"
)

func TestContentBlock_Interface(t *testing.T) {
	// Verify all content block types implement ContentBlock interface
	var _ ContentBlock = TextBlock{}
	var _ ContentBlock = ThinkingBlock{}
	var _ ContentBlock = ToolUseBlock{}
	var _ ContentBlock = ToolResultBlock{}
}

func TestMessage_Interface(t *testing.T) {
	// Verify all message types implement Message interface
	var _ Message = &UserMessage{}
	var _ Message = &AssistantMessage{}
	var _ Message = &SystemMessage{}
	var _ Message = &ResultMessage{}
	var _ Message = &StreamEvent{}
}

func TestTextBlock(t *testing.T) {
	block := TextBlock{Text: "Hello, world!"}

	if block.Text != "Hello, world!" {
		t.Errorf("Expected text 'Hello, world!', got '%s'", block.Text)
	}
}

func TestThinkingBlock(t *testing.T) {
	block := ThinkingBlock{
		Thinking:  "Let me think about this...",
		Signature: "sig-abc-123",
	}

	if block.Thinking != "Let me think about this..." {
		t.Errorf("Expected thinking text to match")
	}
	if block.Signature != "sig-abc-123" {
		t.Errorf("Expected signature 'sig-abc-123', got '%s'", block.Signature)
	}
}

func TestToolUseBlock(t *testing.T) {
	block := ToolUseBlock{
		ID:    "tool-123",
		Name:  "read_file",
		Input: map[string]any{"path": "/tmp/test.txt"},
	}

	if block.ID != "tool-123" {
		t.Errorf("Expected ID 'tool-123', got '%s'", block.ID)
	}
	if block.Name != "read_file" {
		t.Errorf("Expected Name 'read_file', got '%s'", block.Name)
	}
	if block.Input["path"] != "/tmp/test.txt" {
		t.Errorf("Expected Input path '/tmp/test.txt', got '%v'", block.Input["path"])
	}
}

func TestToolResultBlock(t *testing.T) {
	isErr := true
	block := ToolResultBlock{
		ToolUseID: "tool-456",
		Content:   "File contents here",
		IsError:   &isErr,
	}

	if block.ToolUseID != "tool-456" {
		t.Errorf("Expected ToolUseID 'tool-456', got '%s'", block.ToolUseID)
	}
	if block.Content != "File contents here" {
		t.Errorf("Expected Content to match")
	}
	if block.IsError == nil || !*block.IsError {
		t.Error("Expected IsError to be true")
	}
}

func TestToolResultBlock_NilIsError(t *testing.T) {
	block := ToolResultBlock{
		ToolUseID: "tool-789",
		Content:   "result",
	}

	if block.IsError != nil {
		t.Error("Expected IsError to be nil")
	}
}

func TestUserMessage_GetContentString(t *testing.T) {
	msg := &UserMessage{Content: "Hello, Claude!"}

	if msg.GetContentString() != "Hello, Claude!" {
		t.Errorf("Expected 'Hello, Claude!', got '%s'", msg.GetContentString())
	}
}

func TestUserMessage_GetContentString_NotString(t *testing.T) {
	msg := &UserMessage{Content: []ContentBlock{TextBlock{Text: "Hello"}}}

	if msg.GetContentString() != "" {
		t.Errorf("Expected empty string for non-string content, got '%s'", msg.GetContentString())
	}
}

func TestUserMessage_GetContentBlocks(t *testing.T) {
	blocks := []ContentBlock{
		TextBlock{Text: "Hello"},
		ToolResultBlock{ToolUseID: "t1"},
	}
	msg := &UserMessage{Content: blocks}

	result := msg.GetContentBlocks()
	if len(result) != 2 {
		t.Fatalf("Expected 2 blocks, got %d", len(result))
	}

	_, ok := result[0].(TextBlock)
	if !ok {
		t.Errorf("Expected first block to be TextBlock")
	}
}

func TestUserMessage_GetContentBlocks_NotBlocks(t *testing.T) {
	msg := &UserMessage{Content: "string content"}

	result := msg.GetContentBlocks()
	if result != nil {
		t.Errorf("Expected nil for string content, got %v", result)
	}
}

func TestUserMessage_Fields(t *testing.T) {
	msg := &UserMessage{
		Content:         "Test",
		UUID:            "uuid-123",
		ParentToolUseID: "parent-456",
		ToolUseResult:   map[string]any{"success": true},
	}

	if msg.UUID != "uuid-123" {
		t.Errorf("Expected UUID 'uuid-123', got '%s'", msg.UUID)
	}
	if msg.ParentToolUseID != "parent-456" {
		t.Errorf("Expected ParentToolUseID 'parent-456', got '%s'", msg.ParentToolUseID)
	}
	if msg.ToolUseResult["success"] != true {
		t.Error("Expected ToolUseResult['success'] to be true")
	}
}

func TestAssistantMessage_Fields(t *testing.T) {
	msg := &AssistantMessage{
		Content: []ContentBlock{
			TextBlock{Text: "Hello!"},
		},
		Model:           "claude-3-opus",
		ParentToolUseID: "parent-123",
		Error:           AssistantMessageErrorRateLimit,
	}

	if len(msg.Content) != 1 {
		t.Errorf("Expected 1 content block, got %d", len(msg.Content))
	}
	if msg.Model != "claude-3-opus" {
		t.Errorf("Expected Model 'claude-3-opus', got '%s'", msg.Model)
	}
	if msg.ParentToolUseID != "parent-123" {
		t.Errorf("Expected ParentToolUseID 'parent-123', got '%s'", msg.ParentToolUseID)
	}
	if msg.Error != AssistantMessageErrorRateLimit {
		t.Errorf("Expected Error 'rate_limit', got '%s'", msg.Error)
	}
}

func TestAssistantMessageError_Constants(t *testing.T) {
	tests := []struct {
		err      AssistantMessageError
		expected string
	}{
		{AssistantMessageErrorAuthenticationFailed, "authentication_failed"},
		{AssistantMessageErrorBillingError, "billing_error"},
		{AssistantMessageErrorRateLimit, "rate_limit"},
		{AssistantMessageErrorInvalidRequest, "invalid_request"},
		{AssistantMessageErrorServerError, "server_error"},
		{AssistantMessageErrorUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.err) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.err))
			}
		})
	}
}

func TestSystemMessage_Fields(t *testing.T) {
	msg := &SystemMessage{
		Subtype: "init",
		Data: map[string]any{
			"session_id": "session-123",
			"tools":      []string{"read_file"},
		},
	}

	if msg.Subtype != "init" {
		t.Errorf("Expected Subtype 'init', got '%s'", msg.Subtype)
	}
	if msg.Data == nil {
		t.Error("Expected Data to be non-nil")
	}
	if msg.Data["session_id"] != "session-123" {
		t.Errorf("Expected session_id 'session-123', got '%v'", msg.Data["session_id"])
	}
}

func TestResultMessage_Fields(t *testing.T) {
	cost := 0.0045
	msg := &ResultMessage{
		Subtype:       "success",
		DurationMs:    1500,
		DurationAPIMs: 1200,
		IsError:       false,
		NumTurns:      3,
		SessionID:     "session-789",
		TotalCostUSD:  &cost,
		Usage: map[string]any{
			"input_tokens":  100,
			"output_tokens": 250,
		},
		Result:           "completed",
		StructuredOutput: map[string]any{"status": "ok"},
	}

	if msg.Subtype != "success" {
		t.Errorf("Expected Subtype 'success', got '%s'", msg.Subtype)
	}
	if msg.DurationMs != 1500 {
		t.Errorf("Expected DurationMs 1500, got %d", msg.DurationMs)
	}
	if msg.DurationAPIMs != 1200 {
		t.Errorf("Expected DurationAPIMs 1200, got %d", msg.DurationAPIMs)
	}
	if msg.IsError {
		t.Error("Expected IsError to be false")
	}
	if msg.NumTurns != 3 {
		t.Errorf("Expected NumTurns 3, got %d", msg.NumTurns)
	}
	if msg.SessionID != "session-789" {
		t.Errorf("Expected SessionID 'session-789', got '%s'", msg.SessionID)
	}
	if msg.TotalCostUSD == nil || *msg.TotalCostUSD != 0.0045 {
		t.Errorf("Expected TotalCostUSD 0.0045, got %v", msg.TotalCostUSD)
	}
	if msg.Usage == nil {
		t.Error("Expected Usage to be non-nil")
	}
	if msg.Result != "completed" {
		t.Errorf("Expected Result 'completed', got '%s'", msg.Result)
	}
	if msg.StructuredOutput == nil {
		t.Error("Expected StructuredOutput to be non-nil")
	}
}

func TestResultMessage_NilOptionalFields(t *testing.T) {
	msg := &ResultMessage{
		Subtype: "success",
	}

	if msg.TotalCostUSD != nil {
		t.Error("Expected TotalCostUSD to be nil")
	}
	if msg.Usage != nil {
		t.Error("Expected Usage to be nil")
	}
	if msg.StructuredOutput != nil {
		t.Error("Expected StructuredOutput to be nil")
	}
}

func TestStreamEvent_Fields(t *testing.T) {
	msg := &StreamEvent{
		UUID:      "stream-001",
		SessionID: "session-stream",
		Event: map[string]any{
			"type": "content_block_delta",
			"delta": map[string]any{
				"text": "Hello",
			},
		},
		ParentToolUseID: "parent-123",
	}

	if msg.UUID != "stream-001" {
		t.Errorf("Expected UUID 'stream-001', got '%s'", msg.UUID)
	}
	if msg.SessionID != "session-stream" {
		t.Errorf("Expected SessionID 'session-stream', got '%s'", msg.SessionID)
	}
	if msg.Event == nil {
		t.Error("Expected Event to be non-nil")
	}
	if msg.Event["type"] != "content_block_delta" {
		t.Errorf("Expected Event type 'content_block_delta', got '%v'", msg.Event["type"])
	}
	if msg.ParentToolUseID != "parent-123" {
		t.Errorf("Expected ParentToolUseID 'parent-123', got '%s'", msg.ParentToolUseID)
	}
}

// Benchmark tests

func BenchmarkUserMessage_GetContentString(b *testing.B) {
	msg := &UserMessage{Content: "Hello, Claude!"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.GetContentString()
	}
}

func BenchmarkUserMessage_GetContentBlocks(b *testing.B) {
	blocks := []ContentBlock{
		TextBlock{Text: "Hello"},
		TextBlock{Text: "World"},
	}
	msg := &UserMessage{Content: blocks}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = msg.GetContentBlocks()
	}
}
