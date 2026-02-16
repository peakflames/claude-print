package events

import (
	"encoding/json"
	"fmt"
)

// Event is an interface that all event types implement.
type Event interface {
	EventType() string
}

// EventType returns the type of the BaseEvent.
func (e BaseEvent) EventType() string {
	return e.Type
}

// EventType returns the type of the SystemEvent.
func (e SystemEvent) EventType() string {
	return e.Type
}

// EventType returns the type of the StreamEvent.
func (e StreamEvent) EventType() string {
	return e.Type
}

// EventType returns the type of the ResultEvent.
func (e ResultEvent) EventType() string {
	return e.Type
}

// EventType returns the type of the AssistantMessageEvent.
func (e AssistantMessageEvent) EventType() string {
	return e.Type
}

// EventType returns the type of the UserMessageEvent.
func (e UserMessageEvent) EventType() string {
	return e.Type
}

// EventType returns the type of the AssistantEvent.
func (e AssistantEvent) EventType() string {
	return e.Type
}

// EventType returns the type of the UserEvent.
func (e UserEvent) EventType() string {
	return e.Type
}

// ParseEvent parses a JSON string and returns a typed event.
// It determines the event type from the "type" field and returns
// the appropriate struct. Returns an error for malformed JSON.
func ParseEvent(jsonStr string) (Event, error) {
	if jsonStr == "" {
		return nil, fmt.Errorf("empty JSON string")
	}

	// First, parse just the type field to determine the event type
	var base BaseEvent
	if err := json.Unmarshal([]byte(jsonStr), &base); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if base.Type == "" {
		return nil, fmt.Errorf("missing 'type' field in JSON")
	}

	// Parse into the appropriate struct based on type
	switch base.Type {
	case "system.init", "hook_started", "hook_response":
		var event SystemEvent
		if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
			return nil, fmt.Errorf("failed to parse system event: %w", err)
		}
		return event, nil

	case "stream_event":
		var event StreamEvent
		if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
			return nil, fmt.Errorf("failed to parse stream event: %w", err)
		}
		return event, nil

	case "result":
		var event ResultEvent
		if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
			return nil, fmt.Errorf("failed to parse result event: %w", err)
		}
		return event, nil

	case "assistant_message":
		var event AssistantMessageEvent
		if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
			return nil, fmt.Errorf("failed to parse assistant message event: %w", err)
		}
		return event, nil

	case "user_message":
		var event UserMessageEvent
		if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
			return nil, fmt.Errorf("failed to parse user message event: %w", err)
		}
		return event, nil

	case "assistant":
		var event AssistantEvent
		if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
			return nil, fmt.Errorf("failed to parse assistant event: %w", err)
		}
		return event, nil

	case "user":
		var event UserEvent
		if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
			return nil, fmt.Errorf("failed to parse user event: %w", err)
		}
		return event, nil

	default:
		// For unknown types, return the base event to allow graceful handling
		return base, nil
	}
}

// GetStreamEventType returns the nested event type for a StreamEvent.
// For example, a StreamEvent may contain a "message_start", "content_block_delta", etc.
func GetStreamEventType(event StreamEvent) string {
	return event.Event.Type
}

// IsContentBlockDelta checks if a StreamEvent is a content block delta event.
func IsContentBlockDelta(event StreamEvent) bool {
	return event.Event.Type == "content_block_delta"
}

// IsContentBlockStart checks if a StreamEvent is a content block start event.
func IsContentBlockStart(event StreamEvent) bool {
	return event.Event.Type == "content_block_start"
}

// IsContentBlockStop checks if a StreamEvent is a content block stop event.
func IsContentBlockStop(event StreamEvent) bool {
	return event.Event.Type == "content_block_stop"
}

// IsMessageStart checks if a StreamEvent is a message start event.
func IsMessageStart(event StreamEvent) bool {
	return event.Event.Type == "message_start"
}

// IsMessageDelta checks if a StreamEvent is a message delta event.
func IsMessageDelta(event StreamEvent) bool {
	return event.Event.Type == "message_delta"
}

// IsMessageStop checks if a StreamEvent is a message stop event.
func IsMessageStop(event StreamEvent) bool {
	return event.Event.Type == "message_stop"
}
