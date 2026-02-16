# Claude CLI Streaming JSON Event Processing

This document explains the key pattern for correctly processing Claude CLI's streaming JSON output when using `--include-partial-messages`.

## CLI Flags Used

```go
args := []string{
    "-p", prompt,
    "--dangerously-skip-permissions",
    "--include-partial-messages",
    "--verbose",
    "--output-format=stream-json",
}
```

The `--include-partial-messages` flag enables incremental streaming of tool inputs, but requires special handling.

## Event Flow Overview

### With `--include-partial-messages`

```
1. stream_event (content_block_start)  → tool_use with EMPTY input: {}
2. stream_event (input_json_delta)     → partial JSON fragments (multiple)
3. "type": "assistant"                 → COMPLETE tool_use with full input
4. stream_event (content_block_stop)
5. "type": "user"                      → tool_result with metadata
```

### Without `--include-partial-messages`

```
1. "type": "assistant"                 → complete tool_use with full input
2. "type": "user"                      → tool_result
```

## The Problem

When using `--include-partial-messages`, the `content_block_start` event arrives with an **empty input object**:

```json
{
  "type": "stream_event",
  "event": {
    "type": "content_block_start",
    "content_block": {
      "type": "tool_use",
      "name": "Read",
      "id": "toolu_bdrk_018J3HSjfv2v8wtYgm1iZbSw",
      "input": {}  // EMPTY!
    }
  }
}
```

The full input is streamed incrementally via `input_json_delta` events:

```json
{"type": "input_json_delta", "partial_json": "{\"file_path"}
{"type": "input_json_delta", "partial_json": "\":"}
{"type": "input_json_delta", "partial_json": " \""}
{"type": "input_json_delta", "partial_json": "C:\\"}
// ... more fragments
```

## The Solution

### 1. Track tool calls on `content_block_start`, but don't display

```go
case "tool_use":
    // Track the tool ID but DON'T display yet - input is empty
    d.State.PendingTools[block.ID] = &PendingToolCall{
        ID:    block.ID,
        Name:  block.Name,
        Input: block.Input,
    }
```

### 2. Display tool calls from the `"type": "assistant"` event

The complete tool call with full input arrives in a top-level `assistant` event:

```json
{
  "type": "assistant",
  "message": {
    "content": [{
      "type": "tool_use",
      "name": "Read",
      "id": "toolu_bdrk_018J3HSjfv2v8wtYgm1iZbSw",
      "input": {
        "file_path": "C:\\Users\\...\\README.md"  // COMPLETE!
      }
    }]
  }
}
```

Handle this event to display the tool call:

```go
func (d *Display) handleAssistantEvent(e events.AssistantEvent) {
    for _, block := range e.Message.Content {
        if block.Type == "tool_use" {
            d.showToolUse(block.Name, block.ID, block.Input)
        }
        // Skip text - already streamed via content_block_delta
    }
}
```

### 3. Handle tool results from `"type": "user"` events

Tool results arrive in user events with optional metadata:

```json
{
  "type": "user",
  "message": {
    "content": [{
      "type": "tool_result",
      "tool_use_id": "toolu_bdrk_018J3HSjfv2v8wtYgm1iZbSw",
      "content": "file contents here..."
    }]
  },
  "tool_use_result": {
    "type": "text",
    "file": {
      "filePath": "...",
      "numLines": 228
    }
  }
}
```

The `tool_use_result.file.numLines` provides metadata for display formatting.

## Text Streaming

Text content is streamed via `content_block_delta` events with `text_delta`:

```json
{"type": "content_block_delta", "delta": {"type": "text_delta", "text": "Hello"}}
{"type": "content_block_delta", "delta": {"type": "text_delta", "text": " world"}}
```

**Important**: The `assistant` event also contains the complete text, but you should NOT display it again - it would duplicate the streamed output.

```go
func (d *Display) handleAssistantEvent(e events.AssistantEvent) {
    for _, block := range e.Message.Content {
        if block.Type == "tool_use" {
            d.showToolUse(block.Name, block.ID, block.Input)
        }
        // Text already streamed via content_block_delta - skip here
    }
}
```

## Expected Output

With correct processing:

```
> User: Read the README.md

● Read(C:\Users\...\README.md)
  ⎿  Read 228 lines

● Here is the content of README.md...

Session complete: 2 turns, 7.9s total (7.7s API), $0.91
```

## Reference Files

Sample JSON captures for analysis:
- `docs/sample-stream-with-partial.json` - With `--include-partial-messages`
- `docs/sample-stream-without-partial.json` - Without `--include-partial-messages`

## Event Types Summary

| Event Type | Source | When to Display |
|------------|--------|-----------------|
| `stream_event/content_block_start` (tool_use) | Streaming | Track only, don't display |
| `stream_event/content_block_delta` (text) | Streaming | Display immediately |
| `"type": "assistant"` (tool_use) | Batched | Display with full input |
| `"type": "assistant"` (text) | Batched | Skip (already streamed) |
| `"type": "user"` (tool_result) | Batched | Display result summary |
| `"type": "result"` | Final | Display session stats |
