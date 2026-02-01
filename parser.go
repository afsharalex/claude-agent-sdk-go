package claude

import "fmt"

// ParseMessage parses a message from CLI output into typed Message objects.
func ParseMessage(data map[string]any) (Message, error) {
	if data == nil {
		return nil, NewMessageParseError("Invalid message data type (expected map, got nil)", nil)
	}

	msgType, ok := data["type"].(string)
	if !ok || msgType == "" {
		return nil, NewMessageParseError("Message missing 'type' field", data)
	}

	switch msgType {
	case "user":
		return parseUserMessage(data)
	case "assistant":
		return parseAssistantMessage(data)
	case "system":
		return parseSystemMessage(data)
	case "result":
		return parseResultMessage(data)
	case "stream_event":
		return parseStreamEvent(data)
	default:
		return nil, NewMessageParseError(fmt.Sprintf("Unknown message type: %s", msgType), data)
	}
}

func parseUserMessage(data map[string]any) (*UserMessage, error) {
	message, ok := data["message"].(map[string]any)
	if !ok {
		return nil, NewMessageParseError("Missing required field in user message: message", data)
	}

	msg := &UserMessage{}

	// Parse UUID
	if uuid, ok := data["uuid"].(string); ok {
		msg.UUID = uuid
	}

	// Parse parent_tool_use_id
	if parentID, ok := data["parent_tool_use_id"].(string); ok {
		msg.ParentToolUseID = parentID
	}

	// Parse tool_use_result
	if result, ok := data["tool_use_result"].(map[string]any); ok {
		msg.ToolUseResult = result
	}

	// Parse content - can be string or list of content blocks
	content := message["content"]
	if contentStr, ok := content.(string); ok {
		msg.Content = contentStr
	} else if contentList, ok := content.([]any); ok {
		blocks := make([]ContentBlock, 0, len(contentList))
		for _, item := range contentList {
			block, ok := item.(map[string]any)
			if !ok {
				continue
			}
			parsed, err := parseContentBlock(block)
			if err != nil {
				continue
			}
			blocks = append(blocks, parsed)
		}
		msg.Content = blocks
	} else {
		msg.Content = content
	}

	return msg, nil
}

func parseAssistantMessage(data map[string]any) (*AssistantMessage, error) {
	message, ok := data["message"].(map[string]any)
	if !ok {
		return nil, NewMessageParseError("Missing required field in assistant message: message", data)
	}

	msg := &AssistantMessage{}

	// Parse model
	if model, ok := message["model"].(string); ok {
		msg.Model = model
	} else {
		return nil, NewMessageParseError("Missing required field in assistant message: model", data)
	}

	// Parse parent_tool_use_id
	if parentID, ok := data["parent_tool_use_id"].(string); ok {
		msg.ParentToolUseID = parentID
	}

	// Parse error
	if errStr, ok := message["error"].(string); ok {
		msg.Error = AssistantMessageError(errStr)
	}

	// Parse content blocks
	contentList, ok := message["content"].([]any)
	if !ok {
		return nil, NewMessageParseError("Missing required field in assistant message: content", data)
	}

	msg.Content = make([]ContentBlock, 0, len(contentList))
	for _, item := range contentList {
		block, ok := item.(map[string]any)
		if !ok {
			continue
		}
		parsed, err := parseContentBlock(block)
		if err != nil {
			continue
		}
		msg.Content = append(msg.Content, parsed)
	}

	return msg, nil
}

func parseContentBlock(block map[string]any) (ContentBlock, error) {
	blockType, ok := block["type"].(string)
	if !ok {
		return nil, fmt.Errorf("content block missing type")
	}

	switch blockType {
	case "text":
		text, _ := block["text"].(string)
		return TextBlock{Text: text}, nil

	case "thinking":
		thinking, _ := block["thinking"].(string)
		signature, _ := block["signature"].(string)
		return ThinkingBlock{
			Thinking:  thinking,
			Signature: signature,
		}, nil

	case "tool_use":
		id, _ := block["id"].(string)
		name, _ := block["name"].(string)
		input, _ := block["input"].(map[string]any)
		return ToolUseBlock{
			ID:    id,
			Name:  name,
			Input: input,
		}, nil

	case "tool_result":
		toolUseID, _ := block["tool_use_id"].(string)
		content := block["content"]
		var isError *bool
		if e, ok := block["is_error"].(bool); ok {
			isError = &e
		}
		return ToolResultBlock{
			ToolUseID: toolUseID,
			Content:   content,
			IsError:   isError,
		}, nil

	default:
		return nil, fmt.Errorf("unknown content block type: %s", blockType)
	}
}

func parseSystemMessage(data map[string]any) (*SystemMessage, error) {
	subtype, ok := data["subtype"].(string)
	if !ok {
		return nil, NewMessageParseError("Missing required field in system message: subtype", data)
	}

	return &SystemMessage{
		Subtype: subtype,
		Data:    data,
	}, nil
}

func parseResultMessage(data map[string]any) (*ResultMessage, error) {
	msg := &ResultMessage{}

	// Required fields
	if subtype, ok := data["subtype"].(string); ok {
		msg.Subtype = subtype
	} else {
		return nil, NewMessageParseError("Missing required field in result message: subtype", data)
	}

	if durationMs, ok := data["duration_ms"].(float64); ok {
		msg.DurationMs = int(durationMs)
	}

	if durationAPIMs, ok := data["duration_api_ms"].(float64); ok {
		msg.DurationAPIMs = int(durationAPIMs)
	}

	if isError, ok := data["is_error"].(bool); ok {
		msg.IsError = isError
	}

	if numTurns, ok := data["num_turns"].(float64); ok {
		msg.NumTurns = int(numTurns)
	}

	if sessionID, ok := data["session_id"].(string); ok {
		msg.SessionID = sessionID
	}

	// Optional fields
	if totalCostUSD, ok := data["total_cost_usd"].(float64); ok {
		msg.TotalCostUSD = &totalCostUSD
	}

	if usage, ok := data["usage"].(map[string]any); ok {
		msg.Usage = usage
	}

	if result, ok := data["result"].(string); ok {
		msg.Result = result
	}

	if structuredOutput := data["structured_output"]; structuredOutput != nil {
		msg.StructuredOutput = structuredOutput
	}

	return msg, nil
}

func parseStreamEvent(data map[string]any) (*StreamEvent, error) {
	msg := &StreamEvent{}

	// Required fields
	if uuid, ok := data["uuid"].(string); ok {
		msg.UUID = uuid
	} else {
		return nil, NewMessageParseError("Missing required field in stream_event message: uuid", data)
	}

	if sessionID, ok := data["session_id"].(string); ok {
		msg.SessionID = sessionID
	} else {
		return nil, NewMessageParseError("Missing required field in stream_event message: session_id", data)
	}

	if event, ok := data["event"].(map[string]any); ok {
		msg.Event = event
	} else {
		return nil, NewMessageParseError("Missing required field in stream_event message: event", data)
	}

	// Optional fields
	if parentID, ok := data["parent_tool_use_id"].(string); ok {
		msg.ParentToolUseID = parentID
	}

	return msg, nil
}
