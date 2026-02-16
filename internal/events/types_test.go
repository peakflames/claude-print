package events

import (
	"encoding/json"
	"testing"
)

func TestUserEventUnmarshal_ToolDenied(t *testing.T) {
	// This is the actual JSON from a tool denial
	jsonData := `{
		"type": "user",
		"message": {
			"role": "user",
			"content": [{
				"type": "tool_result",
				"tool_use_id": "toolu_01ABCD",
				"content": "Error: Permission to use Read has been denied by user configuration",
				"is_error": true
			}]
		},
		"tool_use_result": "Error: Permission to use Read has been denied by user configuration"
	}`

	var event UserEvent
	err := json.Unmarshal([]byte(jsonData), &event)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Check tool_use_result was parsed as string
	if event.ToolUseResult == nil {
		t.Fatal("ToolUseResult should not be nil")
	}
	if !event.ToolUseResult.IsStringValue {
		t.Error("ToolUseResult should be marked as string value")
	}
	if event.ToolUseResult.RawValue != "Error: Permission to use Read has been denied by user configuration" {
		t.Errorf("Unexpected RawValue: %s", event.ToolUseResult.RawValue)
	}

	// Check content block was parsed
	if len(event.Message.Content) != 1 {
		t.Fatalf("Expected 1 content block, got %d", len(event.Message.Content))
	}
	block := event.Message.Content[0]
	if block.Type != "tool_result" {
		t.Errorf("Expected tool_result type, got %s", block.Type)
	}
	if block.ContentString != "Error: Permission to use Read has been denied by user configuration" {
		t.Errorf("Unexpected ContentString: %s", block.ContentString)
	}
	if !block.IsError {
		t.Error("Block should be marked as error")
	}
}

func TestUserEventUnmarshal_ToolSuccess(t *testing.T) {
	// This is the JSON structure for a successful tool result
	jsonData := `{
		"type": "user",
		"message": {
			"role": "user",
			"content": [{
				"type": "tool_result",
				"tool_use_id": "toolu_01XYZ",
				"content": "File contents here...",
				"is_error": false
			}]
		},
		"tool_use_result": {"type": "text", "file": {"filePath": "test.txt"}}
	}`

	var event UserEvent
	err := json.Unmarshal([]byte(jsonData), &event)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Check tool_use_result was parsed as object
	if event.ToolUseResult == nil {
		t.Fatal("ToolUseResult should not be nil")
	}
	if event.ToolUseResult.IsStringValue {
		t.Error("ToolUseResult should not be marked as string value")
	}
	if event.ToolUseResult.Type != "text" {
		t.Errorf("Expected type 'text', got %s", event.ToolUseResult.Type)
	}
}

func TestUserEventUnmarshal_ContentArray(t *testing.T) {
	// This is the JSON structure when Task agent returns content as array
	jsonData := `{
		"type": "user",
		"message": {
			"role": "user",
			"content": [{
				"type": "tool_result",
				"tool_use_id": "toolu_01TASK",
				"content": [{"type": "text", "text": "Agent result here"}],
				"is_error": false
			}]
		}
	}`

	var event UserEvent
	err := json.Unmarshal([]byte(jsonData), &event)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Check content was parsed as array
	if len(event.Message.Content) != 1 {
		t.Fatalf("Expected 1 content block, got %d", len(event.Message.Content))
	}
	block := event.Message.Content[0]
	if len(block.ContentBlocks) != 1 {
		t.Fatalf("Expected 1 nested content block, got %d", len(block.ContentBlocks))
	}
	if block.ContentString != "Agent result here" {
		t.Errorf("ContentString should be set to first text block: %s", block.ContentString)
	}
}
