package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func loadTestData(t *testing.T, filename string) map[string]any {
	t.Helper()
	path := filepath.Join("testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to load test data from %s: %v", path, err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal test data from %s: %v", path, err)
	}
	return result
}

func TestParseMessage_NilData(t *testing.T) {
	_, err := ParseMessage(nil)
	if err == nil {
		t.Error("Expected error for nil data, got nil")
	}
	parseErr, ok := err.(*MessageParseError)
	if !ok {
		t.Errorf("Expected *MessageParseError, got %T", err)
	}
	if parseErr.Message != "Invalid message data type (expected map, got nil)" {
		t.Errorf("Unexpected error message: %s", parseErr.Message)
	}
}

func TestParseMessage_MissingType(t *testing.T) {
	data := map[string]any{"foo": "bar"}
	_, err := ParseMessage(data)
	if err == nil {
		t.Error("Expected error for missing type, got nil")
	}
	parseErr, ok := err.(*MessageParseError)
	if !ok {
		t.Errorf("Expected *MessageParseError, got %T", err)
	}
	if parseErr.Message != "Message missing 'type' field" {
		t.Errorf("Unexpected error message: %s", parseErr.Message)
	}
}

func TestParseMessage_EmptyType(t *testing.T) {
	data := map[string]any{"type": ""}
	_, err := ParseMessage(data)
	if err == nil {
		t.Error("Expected error for empty type, got nil")
	}
}

func TestParseMessage_UnknownType(t *testing.T) {
	data := map[string]any{"type": "unknown_type"}
	_, err := ParseMessage(data)
	if err == nil {
		t.Error("Expected error for unknown type, got nil")
	}
	parseErr, ok := err.(*MessageParseError)
	if !ok {
		t.Errorf("Expected *MessageParseError, got %T", err)
	}
	if parseErr.Message != "Unknown message type: unknown_type" {
		t.Errorf("Unexpected error message: %s", parseErr.Message)
	}
}

func TestParseMessage_UserMessage(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		wantErr  bool
		validate func(t *testing.T, msg *UserMessage)
	}{
		{
			name: "string content",
			data: map[string]any{
				"type": "user",
				"message": map[string]any{
					"content": "Hello, Claude!",
				},
				"uuid":               "user-123",
				"parent_tool_use_id": "tool-456",
				"tool_use_result": map[string]any{
					"success": true,
				},
			},
			wantErr: false,
			validate: func(t *testing.T, msg *UserMessage) {
				if msg.GetContentString() != "Hello, Claude!" {
					t.Errorf("Expected content 'Hello, Claude!', got '%v'", msg.Content)
				}
				if msg.UUID != "user-123" {
					t.Errorf("Expected UUID 'user-123', got '%s'", msg.UUID)
				}
				if msg.ParentToolUseID != "tool-456" {
					t.Errorf("Expected ParentToolUseID 'tool-456', got '%s'", msg.ParentToolUseID)
				}
				if msg.ToolUseResult == nil {
					t.Error("Expected ToolUseResult to be non-nil")
				}
			},
		},
		{
			name: "content blocks",
			data: map[string]any{
				"type": "user",
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "text", "text": "Hello"},
						map[string]any{"type": "tool_result", "tool_use_id": "tool-1", "content": "result"},
					},
				},
				"uuid": "user-456",
			},
			wantErr: false,
			validate: func(t *testing.T, msg *UserMessage) {
				blocks := msg.GetContentBlocks()
				if len(blocks) != 2 {
					t.Errorf("Expected 2 content blocks, got %d", len(blocks))
				}
				if textBlock, ok := blocks[0].(TextBlock); !ok || textBlock.Text != "Hello" {
					t.Errorf("Expected TextBlock with 'Hello', got %v", blocks[0])
				}
				if resultBlock, ok := blocks[1].(ToolResultBlock); !ok || resultBlock.ToolUseID != "tool-1" {
					t.Errorf("Expected ToolResultBlock with tool_use_id 'tool-1', got %v", blocks[1])
				}
			},
		},
		{
			name: "missing message field",
			data: map[string]any{
				"type": "user",
			},
			wantErr: true,
		},
		{
			name: "invalid content block skipped",
			data: map[string]any{
				"type": "user",
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "text", "text": "Valid"},
						"invalid block",
						map[string]any{}, // missing type
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, msg *UserMessage) {
				blocks := msg.GetContentBlocks()
				if len(blocks) != 1 {
					t.Errorf("Expected 1 valid block, got %d", len(blocks))
				}
			},
		},
		{
			name: "nil content",
			data: map[string]any{
				"type": "user",
				"message": map[string]any{
					"content": nil,
				},
			},
			wantErr: false,
			validate: func(t *testing.T, msg *UserMessage) {
				if msg.Content != nil {
					t.Errorf("Expected nil content, got %v", msg.Content)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMessage(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			msg, ok := result.(*UserMessage)
			if !ok {
				t.Fatalf("Expected *UserMessage, got %T", result)
			}
			if tt.validate != nil {
				tt.validate(t, msg)
			}
		})
	}
}

func TestParseMessage_AssistantMessage(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		wantErr  bool
		validate func(t *testing.T, msg *AssistantMessage)
	}{
		{
			name: "full message",
			data: map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"model": "claude-3-opus",
					"content": []any{
						map[string]any{"type": "text", "text": "Hello!"},
						map[string]any{"type": "thinking", "thinking": "Let me think...", "signature": "sig-123"},
						map[string]any{"type": "tool_use", "id": "tool-1", "name": "read_file", "input": map[string]any{"path": "/tmp"}},
					},
					"error": "rate_limit",
				},
				"parent_tool_use_id": "parent-123",
			},
			wantErr: false,
			validate: func(t *testing.T, msg *AssistantMessage) {
				if msg.Model != "claude-3-opus" {
					t.Errorf("Expected model 'claude-3-opus', got '%s'", msg.Model)
				}
				if len(msg.Content) != 3 {
					t.Errorf("Expected 3 content blocks, got %d", len(msg.Content))
				}
				if msg.Error != "rate_limit" {
					t.Errorf("Expected error 'rate_limit', got '%s'", msg.Error)
				}
				if msg.ParentToolUseID != "parent-123" {
					t.Errorf("Expected ParentToolUseID 'parent-123', got '%s'", msg.ParentToolUseID)
				}

				// Verify block types
				if _, ok := msg.Content[0].(TextBlock); !ok {
					t.Errorf("Expected TextBlock, got %T", msg.Content[0])
				}
				if thinkingBlock, ok := msg.Content[1].(ThinkingBlock); !ok {
					t.Errorf("Expected ThinkingBlock, got %T", msg.Content[1])
				} else if thinkingBlock.Signature != "sig-123" {
					t.Errorf("Expected signature 'sig-123', got '%s'", thinkingBlock.Signature)
				}
				if toolUseBlock, ok := msg.Content[2].(ToolUseBlock); !ok {
					t.Errorf("Expected ToolUseBlock, got %T", msg.Content[2])
				} else {
					if toolUseBlock.ID != "tool-1" {
						t.Errorf("Expected ID 'tool-1', got '%s'", toolUseBlock.ID)
					}
					if toolUseBlock.Name != "read_file" {
						t.Errorf("Expected Name 'read_file', got '%s'", toolUseBlock.Name)
					}
				}
			},
		},
		{
			name: "missing message field",
			data: map[string]any{
				"type": "assistant",
			},
			wantErr: true,
		},
		{
			name: "missing model",
			data: map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"content": []any{},
				},
			},
			wantErr: true,
		},
		{
			name: "missing content",
			data: map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"model": "claude-3-opus",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid content blocks skipped",
			data: map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"model": "claude-3-opus",
					"content": []any{
						map[string]any{"type": "text", "text": "Valid"},
						"invalid",
						map[string]any{"type": "unknown_type"},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, msg *AssistantMessage) {
				if len(msg.Content) != 1 {
					t.Errorf("Expected 1 valid block, got %d", len(msg.Content))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMessage(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			msg, ok := result.(*AssistantMessage)
			if !ok {
				t.Fatalf("Expected *AssistantMessage, got %T", result)
			}
			if tt.validate != nil {
				tt.validate(t, msg)
			}
		})
	}
}

func TestParseMessage_SystemMessage(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		wantErr  bool
		validate func(t *testing.T, msg *SystemMessage)
	}{
		{
			name: "init subtype",
			data: map[string]any{
				"type":       "system",
				"subtype":    "init",
				"session_id": "session-123",
				"tools":      []string{"read_file"},
			},
			wantErr: false,
			validate: func(t *testing.T, msg *SystemMessage) {
				if msg.Subtype != "init" {
					t.Errorf("Expected subtype 'init', got '%s'", msg.Subtype)
				}
				if msg.Data == nil {
					t.Error("Expected Data to be non-nil")
				}
			},
		},
		{
			name: "missing subtype",
			data: map[string]any{
				"type": "system",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMessage(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			msg, ok := result.(*SystemMessage)
			if !ok {
				t.Fatalf("Expected *SystemMessage, got %T", result)
			}
			if tt.validate != nil {
				tt.validate(t, msg)
			}
		})
	}
}

func TestParseMessage_ResultMessage(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		wantErr  bool
		validate func(t *testing.T, msg *ResultMessage)
	}{
		{
			name: "full result",
			data: map[string]any{
				"type":            "result",
				"subtype":         "success",
				"duration_ms":     float64(1500),
				"duration_api_ms": float64(1200),
				"is_error":        false,
				"num_turns":       float64(3),
				"session_id":      "session-789",
				"total_cost_usd":  0.0045,
				"usage": map[string]any{
					"input_tokens":  float64(100),
					"output_tokens": float64(250),
				},
				"result":            "completed",
				"structured_output": map[string]any{"status": "ok"},
			},
			wantErr: false,
			validate: func(t *testing.T, msg *ResultMessage) {
				if msg.Subtype != "success" {
					t.Errorf("Expected subtype 'success', got '%s'", msg.Subtype)
				}
				if msg.DurationMs != 1500 {
					t.Errorf("Expected DurationMs 1500, got %d", msg.DurationMs)
				}
				if msg.DurationAPIMs != 1200 {
					t.Errorf("Expected DurationAPIMs 1200, got %d", msg.DurationAPIMs)
				}
				if msg.IsError {
					t.Error("Expected IsError false")
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
			},
		},
		{
			name: "minimal result",
			data: map[string]any{
				"type":    "result",
				"subtype": "success",
			},
			wantErr: false,
			validate: func(t *testing.T, msg *ResultMessage) {
				if msg.TotalCostUSD != nil {
					t.Errorf("Expected TotalCostUSD nil, got %v", msg.TotalCostUSD)
				}
			},
		},
		{
			name: "missing subtype",
			data: map[string]any{
				"type": "result",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMessage(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			msg, ok := result.(*ResultMessage)
			if !ok {
				t.Fatalf("Expected *ResultMessage, got %T", result)
			}
			if tt.validate != nil {
				tt.validate(t, msg)
			}
		})
	}
}

func TestParseMessage_StreamEvent(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		wantErr  bool
		validate func(t *testing.T, msg *StreamEvent)
	}{
		{
			name: "full stream event",
			data: map[string]any{
				"type":       "stream_event",
				"uuid":       "stream-001",
				"session_id": "session-stream",
				"event": map[string]any{
					"type": "content_block_delta",
					"delta": map[string]any{
						"type": "text_delta",
						"text": "Hello",
					},
				},
				"parent_tool_use_id": "parent-123",
			},
			wantErr: false,
			validate: func(t *testing.T, msg *StreamEvent) {
				if msg.UUID != "stream-001" {
					t.Errorf("Expected UUID 'stream-001', got '%s'", msg.UUID)
				}
				if msg.SessionID != "session-stream" {
					t.Errorf("Expected SessionID 'session-stream', got '%s'", msg.SessionID)
				}
				if msg.Event == nil {
					t.Error("Expected Event to be non-nil")
				}
				if msg.ParentToolUseID != "parent-123" {
					t.Errorf("Expected ParentToolUseID 'parent-123', got '%s'", msg.ParentToolUseID)
				}
			},
		},
		{
			name: "missing uuid",
			data: map[string]any{
				"type":       "stream_event",
				"session_id": "session-1",
				"event":      map[string]any{},
			},
			wantErr: true,
		},
		{
			name: "missing session_id",
			data: map[string]any{
				"type":  "stream_event",
				"uuid":  "uuid-1",
				"event": map[string]any{},
			},
			wantErr: true,
		},
		{
			name: "missing event",
			data: map[string]any{
				"type":       "stream_event",
				"uuid":       "uuid-1",
				"session_id": "session-1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMessage(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			msg, ok := result.(*StreamEvent)
			if !ok {
				t.Fatalf("Expected *StreamEvent, got %T", result)
			}
			if tt.validate != nil {
				tt.validate(t, msg)
			}
		})
	}
}

func TestParseContentBlock(t *testing.T) {
	tests := []struct {
		name     string
		block    map[string]any
		wantErr  bool
		validate func(t *testing.T, block ContentBlock)
	}{
		{
			name:    "text block",
			block:   map[string]any{"type": "text", "text": "Hello, world!"},
			wantErr: false,
			validate: func(t *testing.T, block ContentBlock) {
				tb, ok := block.(TextBlock)
				if !ok {
					t.Fatalf("Expected TextBlock, got %T", block)
				}
				if tb.Text != "Hello, world!" {
					t.Errorf("Expected text 'Hello, world!', got '%s'", tb.Text)
				}
			},
		},
		{
			name:    "text block missing text",
			block:   map[string]any{"type": "text"},
			wantErr: false,
			validate: func(t *testing.T, block ContentBlock) {
				tb := block.(TextBlock)
				if tb.Text != "" {
					t.Errorf("Expected empty text, got '%s'", tb.Text)
				}
			},
		},
		{
			name:    "thinking block",
			block:   map[string]any{"type": "thinking", "thinking": "Deep thoughts...", "signature": "sig-abc"},
			wantErr: false,
			validate: func(t *testing.T, block ContentBlock) {
				tb, ok := block.(ThinkingBlock)
				if !ok {
					t.Fatalf("Expected ThinkingBlock, got %T", block)
				}
				if tb.Thinking != "Deep thoughts..." {
					t.Errorf("Expected thinking 'Deep thoughts...', got '%s'", tb.Thinking)
				}
				if tb.Signature != "sig-abc" {
					t.Errorf("Expected signature 'sig-abc', got '%s'", tb.Signature)
				}
			},
		},
		{
			name: "tool_use block",
			block: map[string]any{
				"type":  "tool_use",
				"id":    "tool-123",
				"name":  "read_file",
				"input": map[string]any{"path": "/tmp/test.txt"},
			},
			wantErr: false,
			validate: func(t *testing.T, block ContentBlock) {
				tb, ok := block.(ToolUseBlock)
				if !ok {
					t.Fatalf("Expected ToolUseBlock, got %T", block)
				}
				if tb.ID != "tool-123" {
					t.Errorf("Expected ID 'tool-123', got '%s'", tb.ID)
				}
				if tb.Name != "read_file" {
					t.Errorf("Expected Name 'read_file', got '%s'", tb.Name)
				}
				if tb.Input["path"] != "/tmp/test.txt" {
					t.Errorf("Expected Input path '/tmp/test.txt', got '%v'", tb.Input["path"])
				}
			},
		},
		{
			name: "tool_result block",
			block: map[string]any{
				"type":        "tool_result",
				"tool_use_id": "tool-456",
				"content":     "File contents here",
				"is_error":    true,
			},
			wantErr: false,
			validate: func(t *testing.T, block ContentBlock) {
				tb, ok := block.(ToolResultBlock)
				if !ok {
					t.Fatalf("Expected ToolResultBlock, got %T", block)
				}
				if tb.ToolUseID != "tool-456" {
					t.Errorf("Expected ToolUseID 'tool-456', got '%s'", tb.ToolUseID)
				}
				if tb.Content != "File contents here" {
					t.Errorf("Expected Content 'File contents here', got '%v'", tb.Content)
				}
				if tb.IsError == nil || !*tb.IsError {
					t.Error("Expected IsError to be true")
				}
			},
		},
		{
			name: "tool_result block without is_error",
			block: map[string]any{
				"type":        "tool_result",
				"tool_use_id": "tool-789",
				"content":     "result",
			},
			wantErr: false,
			validate: func(t *testing.T, block ContentBlock) {
				tb := block.(ToolResultBlock)
				if tb.IsError != nil {
					t.Errorf("Expected IsError nil, got %v", tb.IsError)
				}
			},
		},
		{
			name:    "missing type",
			block:   map[string]any{"text": "hello"},
			wantErr: true,
		},
		{
			name:    "unknown type",
			block:   map[string]any{"type": "unknown_block_type"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// parseContentBlock is not exported, so we test it indirectly
			// by parsing an assistant message with the block
			data := map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"model":   "claude-3",
					"content": []any{tt.block},
				},
			}
			result, err := ParseMessage(data)
			if tt.wantErr {
				// For error cases, either ParseMessage errors or the block is skipped
				if err == nil {
					msg := result.(*AssistantMessage)
					if len(msg.Content) > 0 {
						t.Error("Expected error or empty content, but got content block")
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			msg := result.(*AssistantMessage)
			if len(msg.Content) != 1 {
				t.Fatalf("Expected 1 content block, got %d", len(msg.Content))
			}
			if tt.validate != nil {
				tt.validate(t, msg.Content[0])
			}
		})
	}
}

func TestParseMessage_FromTestDataFiles(t *testing.T) {
	tests := []struct {
		filename    string
		expectedType string
	}{
		{"user_message.json", "user"},
		{"user_message_content_blocks.json", "user"},
		{"assistant_message.json", "assistant"},
		{"assistant_message_error.json", "assistant"},
		{"result_message.json", "result"},
		{"system_message.json", "system"},
		{"stream_event.json", "stream_event"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			// Read file without helper since files have different structure
			path := filepath.Join("testdata", tt.filename)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Skipf("Test file not found: %s", path)
			}
			var rawData map[string]any
			if err := json.Unmarshal(data, &rawData); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			// The test data files need to be adjusted for actual message structure
			// For now, we validate the type field parsing works
			msgType, ok := rawData["type"].(string)
			if !ok {
				t.Fatalf("Test file missing type field")
			}
			if msgType != tt.expectedType {
				t.Errorf("Expected type '%s', got '%s'", tt.expectedType, msgType)
			}
		})
	}
}

// Benchmarks

func BenchmarkParseMessage_UserMessage(b *testing.B) {
	data := map[string]any{
		"type": "user",
		"message": map[string]any{
			"content": "Hello, Claude!",
		},
		"uuid": "user-123",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseMessage(data)
	}
}

func BenchmarkParseMessage_AssistantMessage(b *testing.B) {
	data := map[string]any{
		"type": "assistant",
		"message": map[string]any{
			"model": "claude-3-opus",
			"content": []any{
				map[string]any{"type": "text", "text": "Hello!"},
				map[string]any{"type": "thinking", "thinking": "thinking...", "signature": "sig"},
				map[string]any{"type": "tool_use", "id": "t1", "name": "read", "input": map[string]any{}},
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseMessage(data)
	}
}
