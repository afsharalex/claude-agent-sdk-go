package claude

import (
	"context"
	"reflect"
	"sort"
	"testing"
)

func TestTool(t *testing.T) {
	handler := func(ctx context.Context, args map[string]any) (MCPToolResult, error) {
		return TextResult("result"), nil
	}

	inputSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
	}

	tool := Tool("test-tool", "A test tool", inputSchema, handler)

	if tool.Name != "test-tool" {
		t.Errorf("Expected Name='test-tool', got '%s'", tool.Name)
	}
	if tool.Description != "A test tool" {
		t.Errorf("Expected Description='A test tool', got '%s'", tool.Description)
	}
	if tool.InputSchema == nil {
		t.Error("Expected InputSchema to be non-nil")
	}
	if tool.Handler == nil {
		t.Error("Expected Handler to be non-nil")
	}

	// Verify handler works
	result, err := tool.Handler(context.Background(), nil)
	if err != nil {
		t.Errorf("Unexpected error from handler: %v", err)
	}
	if len(result.Content) != 1 || result.Content[0].Text != "result" {
		t.Errorf("Unexpected result from handler: %v", result)
	}
}

func TestSimpleInputSchema_StringType(t *testing.T) {
	schema := SimpleInputSchema(map[string]string{
		"name": "string",
	})

	if schema["type"] != "object" {
		t.Errorf("Expected type='object', got '%v'", schema["type"])
	}

	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("Expected properties to be map[string]any")
	}

	nameProp, ok := properties["name"].(map[string]any)
	if !ok {
		t.Fatal("Expected name property to be map[string]any")
	}

	if nameProp["type"] != "string" {
		t.Errorf("Expected name type='string', got '%v'", nameProp["type"])
	}

	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Expected required to be []string")
	}
	if len(required) != 1 || required[0] != "name" {
		t.Errorf("Expected required=['name'], got %v", required)
	}
}

func TestSimpleInputSchema_NumberTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"number", "number"},
		{"float", "number"},
		{"float64", "number"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			schema := SimpleInputSchema(map[string]string{
				"value": tt.input,
			})

			properties := schema["properties"].(map[string]any)
			valueProp := properties["value"].(map[string]any)

			if valueProp["type"] != tt.expected {
				t.Errorf("Expected type='%s' for input '%s', got '%v'", tt.expected, tt.input, valueProp["type"])
			}
		})
	}
}

func TestSimpleInputSchema_IntegerTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"integer", "integer"},
		{"int", "integer"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			schema := SimpleInputSchema(map[string]string{
				"count": tt.input,
			})

			properties := schema["properties"].(map[string]any)
			countProp := properties["count"].(map[string]any)

			if countProp["type"] != tt.expected {
				t.Errorf("Expected type='%s' for input '%s', got '%v'", tt.expected, tt.input, countProp["type"])
			}
		})
	}
}

func TestSimpleInputSchema_BooleanTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"boolean", "boolean"},
		{"bool", "boolean"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			schema := SimpleInputSchema(map[string]string{
				"active": tt.input,
			})

			properties := schema["properties"].(map[string]any)
			activeProp := properties["active"].(map[string]any)

			if activeProp["type"] != tt.expected {
				t.Errorf("Expected type='%s' for input '%s', got '%v'", tt.expected, tt.input, activeProp["type"])
			}
		})
	}
}

func TestSimpleInputSchema_UnknownTypeDefaultsToString(t *testing.T) {
	schema := SimpleInputSchema(map[string]string{
		"custom": "custom_type",
		"also":   "unknown",
	})

	properties := schema["properties"].(map[string]any)

	customProp := properties["custom"].(map[string]any)
	if customProp["type"] != "string" {
		t.Errorf("Expected unknown type to default to 'string', got '%v'", customProp["type"])
	}

	alsoProp := properties["also"].(map[string]any)
	if alsoProp["type"] != "string" {
		t.Errorf("Expected unknown type to default to 'string', got '%v'", alsoProp["type"])
	}
}

func TestSimpleInputSchema_MultipleFields(t *testing.T) {
	schema := SimpleInputSchema(map[string]string{
		"name":   "string",
		"age":    "integer",
		"active": "boolean",
		"score":  "number",
	})

	properties := schema["properties"].(map[string]any)
	if len(properties) != 4 {
		t.Errorf("Expected 4 properties, got %d", len(properties))
	}

	required := schema["required"].([]string)
	if len(required) != 4 {
		t.Errorf("Expected 4 required fields, got %d", len(required))
	}

	// Sort for consistent comparison
	sort.Strings(required)
	expected := []string{"active", "age", "name", "score"}
	if !reflect.DeepEqual(required, expected) {
		t.Errorf("Expected required=%v, got %v", expected, required)
	}
}

func TestSimpleInputSchema_EmptyFields(t *testing.T) {
	schema := SimpleInputSchema(map[string]string{})

	if schema["type"] != "object" {
		t.Errorf("Expected type='object', got '%v'", schema["type"])
	}

	properties := schema["properties"].(map[string]any)
	if len(properties) != 0 {
		t.Errorf("Expected 0 properties, got %d", len(properties))
	}

	required := schema["required"].([]string)
	if len(required) != 0 {
		t.Errorf("Expected 0 required fields, got %d", len(required))
	}
}

func TestTextResult(t *testing.T) {
	result := TextResult("Hello, world!")

	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(result.Content))
	}
	if result.Content[0].Type != "text" {
		t.Errorf("Expected content type='text', got '%s'", result.Content[0].Type)
	}
	if result.Content[0].Text != "Hello, world!" {
		t.Errorf("Expected text='Hello, world!', got '%s'", result.Content[0].Text)
	}
	if result.IsError {
		t.Error("Expected IsError to be false")
	}
}

func TestTextResult_EmptyText(t *testing.T) {
	result := TextResult("")

	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(result.Content))
	}
	if result.Content[0].Text != "" {
		t.Errorf("Expected empty text, got '%s'", result.Content[0].Text)
	}
}

func TestErrorResult(t *testing.T) {
	result := ErrorResult("Something went wrong")

	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(result.Content))
	}
	if result.Content[0].Type != "text" {
		t.Errorf("Expected content type='text', got '%s'", result.Content[0].Type)
	}
	if result.Content[0].Text != "Something went wrong" {
		t.Errorf("Expected text='Something went wrong', got '%s'", result.Content[0].Text)
	}
	if !result.IsError {
		t.Error("Expected IsError to be true")
	}
}

func TestImageResult(t *testing.T) {
	result := ImageResult("base64data==", "image/png")

	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(result.Content))
	}
	if result.Content[0].Type != "image" {
		t.Errorf("Expected content type='image', got '%s'", result.Content[0].Type)
	}
	if result.Content[0].Data != "base64data==" {
		t.Errorf("Expected data='base64data==', got '%s'", result.Content[0].Data)
	}
	if result.Content[0].MimeType != "image/png" {
		t.Errorf("Expected mimeType='image/png', got '%s'", result.Content[0].MimeType)
	}
	if result.IsError {
		t.Error("Expected IsError to be false")
	}
}

func TestImageResult_DifferentMimeTypes(t *testing.T) {
	tests := []struct {
		mimeType string
	}{
		{"image/png"},
		{"image/jpeg"},
		{"image/gif"},
		{"image/webp"},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			result := ImageResult("data", tt.mimeType)
			if result.Content[0].MimeType != tt.mimeType {
				t.Errorf("Expected mimeType='%s', got '%s'", tt.mimeType, result.Content[0].MimeType)
			}
		})
	}
}

// Benchmark tests

func BenchmarkSimpleInputSchema(b *testing.B) {
	fields := map[string]string{
		"name":   "string",
		"age":    "integer",
		"active": "boolean",
		"score":  "number",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SimpleInputSchema(fields)
	}
}

func BenchmarkTextResult(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = TextResult("benchmark result text")
	}
}

func BenchmarkTool(b *testing.B) {
	handler := func(ctx context.Context, args map[string]any) (MCPToolResult, error) {
		return TextResult("result"), nil
	}
	schema := SimpleInputSchema(map[string]string{"arg": "string"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Tool("bench", "benchmark tool", schema, handler)
	}
}
