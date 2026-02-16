package events

import "encoding/json"

// BaseEvent represents the base structure for all Claude streaming events.
// All events have a Type field that identifies the event type.
type BaseEvent struct {
	Type string `json:"type"`
}

// SystemEvent represents system-level events like system.init, hook_started, hook_response.
type SystemEvent struct {
	BaseEvent
	SessionID      string            `json:"session_id,omitempty"`
	Tools          []ToolInfo        `json:"tools,omitempty"`
	McpServers     []MCPServerInfo   `json:"mcp_servers,omitempty"`
	Model          string            `json:"model,omitempty"`
	Cwd            string            `json:"cwd,omitempty"`
	HookName       string            `json:"hook_name,omitempty"`
	HookType       string            `json:"hook_type,omitempty"`
	TriggeringTool string            `json:"triggering_tool,omitempty"`
	Response       string            `json:"response,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// ToolInfo represents information about an available tool.
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// MCPServerInfo represents information about an MCP server.
type MCPServerInfo struct {
	Name   string `json:"name"`
	Status string `json:"status,omitempty"`
}

// StreamEvent is a wrapper for streaming message events from Claude.
// It contains a nested Event object with the actual message content.
type StreamEvent struct {
	BaseEvent
	Event MessageEvent `json:"event,omitempty"`
}

// MessageEvent represents events within a stream (message_start, content_block_delta, etc.).
type MessageEvent struct {
	Type         string        `json:"type"`
	Index        int           `json:"index,omitempty"`
	Message      *Message      `json:"message,omitempty"`
	ContentBlock *ContentBlock `json:"content_block,omitempty"`
	Delta        *Delta        `json:"delta,omitempty"`
	Usage        *Usage        `json:"usage,omitempty"`
}

// Message represents a Claude message in the stream.
type Message struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content,omitempty"`
	Model        string         `json:"model,omitempty"`
	StopReason   string         `json:"stop_reason,omitempty"`
	StopSequence string         `json:"stop_sequence,omitempty"`
	Usage        *Usage         `json:"usage,omitempty"`
}

// ContentBlock represents a block of content (text, tool_use, tool_result).
type ContentBlock struct {
	Type string `json:"type"`
	// For text blocks
	Text string `json:"text,omitempty"`
	// For tool_use blocks
	ID    string                 `json:"id,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`
	// For tool_result blocks
	ToolUseID string `json:"tool_use_id,omitempty"`
	// Content can be either a string or an array of content blocks (from Task agents).
	// We use json.RawMessage to handle both and populate ContentString/ContentBlocks accordingly.
	Content       json.RawMessage `json:"content,omitempty"`
	ContentString string          `json:"-"` // Populated when content is a string
	ContentBlocks []ContentBlock  `json:"-"` // Populated when content is an array
	IsError       bool            `json:"is_error,omitempty"`
}

// TextBlock represents a text content block.
type TextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ToolUseBlock represents a tool use content block.
type ToolUseBlock struct {
	Type  string                 `json:"type"`
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// ToolResultBlock represents a tool result content block.
type ToolResultBlock struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error,omitempty"`
}

// Delta represents incremental updates in streaming events.
type Delta struct {
	Type       string `json:"type,omitempty"`
	Text       string `json:"text,omitempty"`
	StopReason string `json:"stop_reason,omitempty"`
}

// Usage represents token usage information.
type Usage struct {
	InputTokens              int `json:"input_tokens,omitempty"`
	OutputTokens             int `json:"output_tokens,omitempty"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
}

// ResultEvent represents the final result event at the end of a Claude session.
type ResultEvent struct {
	BaseEvent
	Subtype           string                 `json:"subtype,omitempty"`
	CostUSD           float64                `json:"cost_usd,omitempty"`
	TotalCostUSD      float64                `json:"total_cost_usd,omitempty"`
	DurationMS        int64                  `json:"duration_ms,omitempty"`
	DurationAPIMS     int64                  `json:"duration_api_ms,omitempty"`
	NumTurns          int                    `json:"num_turns,omitempty"`
	Result            string                 `json:"result,omitempty"`
	SessionID         string                 `json:"session_id,omitempty"`
	IsError           bool                   `json:"is_error,omitempty"`
	Usage             *AggregatedUsage       `json:"usage,omitempty"`
	ModelUsage        map[string]*ModelUsage `json:"modelUsage,omitempty"`
	ToolUseCount      map[string]int         `json:"tool_use_count,omitempty"`
	ToolErrorCount    map[string]int         `json:"tool_error_count,omitempty"`
	ToolCancelCount   map[string]int         `json:"tool_cancel_count,omitempty"`
	ToolMistakeCount  map[string]int         `json:"tool_mistake_count,omitempty"`
	TotalToolUse      int                    `json:"total_tool_use,omitempty"`
	TotalToolErrors   int                    `json:"total_tool_errors,omitempty"`
	TotalToolCancels  int                    `json:"total_tool_cancels,omitempty"`
	TotalToolMistakes int                    `json:"total_tool_mistakes,omitempty"`
}

// AggregatedUsage represents aggregated token usage across a session.
type AggregatedUsage struct {
	InputTokens              int `json:"input_tokens,omitempty"`
	OutputTokens             int `json:"output_tokens,omitempty"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
	TotalTokens              int `json:"total_tokens,omitempty"`
}

// ModelUsage represents token usage for a specific model.
type ModelUsage struct {
	InputTokens              int     `json:"inputTokens,omitempty"`
	OutputTokens             int     `json:"outputTokens,omitempty"`
	CacheCreationInputTokens int     `json:"cacheCreationInputTokens,omitempty"`
	CacheReadInputTokens     int     `json:"cacheReadInputTokens,omitempty"`
	CostUSD                  float64 `json:"costUSD,omitempty"`
	ContextWindow            int     `json:"contextWindow,omitempty"`
	MaxOutputTokens          int     `json:"maxOutputTokens,omitempty"`
	WebSearchRequests        int     `json:"webSearchRequests,omitempty"`
}

// AssistantMessageEvent represents an assistant message in the stream.
type AssistantMessageEvent struct {
	BaseEvent
	Message Message `json:"message"`
}

// UserMessageEvent represents a user message event.
type UserMessageEvent struct {
	BaseEvent
	Message UserMessage `json:"message"`
}

// UserMessage represents a user message.
type UserMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ToolUseResult contains structured metadata about a tool execution result.
// This can be either:
// - A string (when tool was denied or errored): "Error: Permission to use X has been denied..."
// - An object (when tool succeeded): {"type": "text", "file": {...}}
type ToolUseResult struct {
	// RawValue holds the string value when tool_use_result is a plain string
	RawValue string `json:"-"`
	// IsStringValue indicates whether this was parsed from a string (error/denied case)
	IsStringValue bool `json:"-"`
	// Structured fields for object case
	Type   string      `json:"type,omitempty"`
	File   *FileResult `json:"file,omitempty"`
	Status string      `json:"status,omitempty"` // For Task agent results: "completed"
	// Future: add GlobResult, GrepResult, BashResult as discovered
}

// FileResult contains metadata for Read tool results
type FileResult struct {
	FilePath   string `json:"filePath,omitempty"`
	Content    string `json:"content,omitempty"`
	NumLines   int    `json:"numLines,omitempty"`
	StartLine  int    `json:"startLine,omitempty"`
	TotalLines int    `json:"totalLines,omitempty"`
}

// UserEvent represents a user message event with tool results
type UserEvent struct {
	BaseEvent
	Message       UserMessageContentBlocks `json:"message"`
	ToolUseResult *ToolUseResult           `json:"tool_use_result,omitempty"`
}

// UserMessageContentBlocks handles the user message structure with content blocks
type UserMessageContentBlocks struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

// AssistantEvent represents top-level assistant messages
type AssistantEvent struct {
	BaseEvent
	Message Message `json:"message"`
}

// userEventRaw is used for initial unmarshaling before handling polymorphic fields
type userEventRaw struct {
	BaseEvent
	Message          json.RawMessage `json:"message"`
	ToolUseResultRaw json.RawMessage `json:"tool_use_result,omitempty"`
}

// UnmarshalJSON handles polymorphic tool_use_result field (string or object)
func (u *UserEvent) UnmarshalJSON(data []byte) error {
	var raw userEventRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	u.Type = raw.Type

	// Parse message (standard structure)
	if len(raw.Message) > 0 {
		if err := json.Unmarshal(raw.Message, &u.Message); err != nil {
			return err
		}
		// Post-process content blocks to handle polymorphic content field
		for i := range u.Message.Content {
			if err := u.Message.Content[i].parseContent(); err != nil {
				// Non-fatal: just log and continue
				continue
			}
		}
	}

	// Parse tool_use_result (polymorphic: string or object)
	if len(raw.ToolUseResultRaw) > 0 {
		u.ToolUseResult = &ToolUseResult{}

		// Try string first (error/denied case)
		var strVal string
		if err := json.Unmarshal(raw.ToolUseResultRaw, &strVal); err == nil {
			u.ToolUseResult.RawValue = strVal
			u.ToolUseResult.IsStringValue = true
			return nil
		}

		// Try object (success case with structured metadata)
		if err := json.Unmarshal(raw.ToolUseResultRaw, u.ToolUseResult); err != nil {
			return err
		}
	}

	return nil
}

// parseContent handles the polymorphic content field in tool_result blocks
func (cb *ContentBlock) parseContent() error {
	if len(cb.Content) == 0 {
		return nil
	}

	// Try string first (most common case)
	var strVal string
	if err := json.Unmarshal(cb.Content, &strVal); err == nil {
		cb.ContentString = strVal
		return nil
	}

	// Try array of content blocks (Task agent results)
	var blocks []ContentBlock
	if err := json.Unmarshal(cb.Content, &blocks); err == nil {
		cb.ContentBlocks = blocks
		// Also set ContentString to first text block for convenience
		for _, block := range blocks {
			if block.Type == "text" && block.Text != "" {
				cb.ContentString = block.Text
				break
			}
		}
		return nil
	}

	return nil
}
