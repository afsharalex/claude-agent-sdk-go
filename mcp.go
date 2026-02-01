package claude

// Tool is a convenience function for creating an MCPTool.
// This provides a pattern similar to the Python SDK's @tool decorator.
//
// Example:
//
//	addTool := claude.Tool("add", "Add two numbers", map[string]any{
//		"type": "object",
//		"properties": map[string]any{
//			"a": map[string]any{"type": "number"},
//			"b": map[string]any{"type": "number"},
//		},
//		"required": []string{"a", "b"},
//	}, func(ctx context.Context, args map[string]any) (claude.MCPToolResult, error) {
//		a := args["a"].(float64)
//		b := args["b"].(float64)
//		return claude.MCPToolResult{
//			Content: []claude.MCPContent{{
//				Type: "text",
//				Text: fmt.Sprintf("Result: %g", a+b),
//			}},
//		}, nil
//	})
func Tool(name, description string, inputSchema map[string]any, handler MCPToolHandler) MCPTool {
	return MCPTool{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
		Handler:     handler,
	}
}

// SimpleInputSchema creates a simple JSON schema from type descriptions.
// This is a convenience function for common schema patterns.
//
// Example:
//
//	schema := claude.SimpleInputSchema(map[string]string{
//		"name": "string",
//		"age":  "number",
//		"active": "boolean",
//	})
func SimpleInputSchema(fields map[string]string) map[string]any {
	properties := make(map[string]any)
	required := make([]string, 0, len(fields))

	for name, typ := range fields {
		var jsonType string
		switch typ {
		case "string":
			jsonType = "string"
		case "number", "float", "float64":
			jsonType = "number"
		case "integer", "int":
			jsonType = "integer"
		case "boolean", "bool":
			jsonType = "boolean"
		default:
			jsonType = "string"
		}
		properties[name] = map[string]any{"type": jsonType}
		required = append(required, name)
	}

	return map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

// TextResult creates an MCPToolResult with a single text content.
func TextResult(text string) MCPToolResult {
	return MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: text,
		}},
	}
}

// ErrorResult creates an MCPToolResult indicating an error.
func ErrorResult(message string) MCPToolResult {
	return MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: message,
		}},
		IsError: true,
	}
}

// ImageResult creates an MCPToolResult with an image content.
func ImageResult(data, mimeType string) MCPToolResult {
	return MCPToolResult{
		Content: []MCPContent{{
			Type:     "image",
			Data:     data,
			MimeType: mimeType,
		}},
	}
}
