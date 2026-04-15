package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/peakflames/claude-print/internal/events"
)

// newTestDisplay creates a Display wired to a bytes.Buffer for both display
// and JSON output, making it easy to assert on both streams.
func newTestDisplay(jsonBuf *bytes.Buffer) *Display {
	formatter := NewFormatter(false, false, &bytes.Buffer{})
	d := NewDisplay(formatter, VerbosityNormal)
	d.JSONWriter = jsonBuf
	return d
}

// decodeLines parses newline-delimited JSON objects from buf into a slice of
// map[string]interface{} values.
func decodeLines(t *testing.T, buf *bytes.Buffer) []map[string]interface{} {
	t.Helper()
	var out []map[string]interface{}
	dec := json.NewDecoder(buf)
	for dec.More() {
		var m map[string]interface{}
		if err := dec.Decode(&m); err != nil {
			t.Fatalf("decode JSON: %v", err)
		}
		out = append(out, m)
	}
	return out
}

func TestJSONWriter_Nil_NoOutput(t *testing.T) {
	formatter := NewFormatter(false, false, &bytes.Buffer{})
	d := NewDisplay(formatter, VerbosityNormal)
	// JSONWriter is nil — no JSON should be emitted

	e := events.StreamEvent{}
	e.Event.Type = "content_block_delta"
	delta := &events.Delta{Text: "hello"}
	e.Event.Delta = delta

	d.HandleEvent(e)
	// Test passes as long as there is no panic (JSONWriter nil is a no-op)
}

func TestJSONWriter_TextDelta(t *testing.T) {
	buf := &bytes.Buffer{}
	d := newTestDisplay(buf)

	e := events.StreamEvent{}
	e.Event.Type = "content_block_delta"
	e.Event.Delta = &events.Delta{Text: "hello"}
	d.HandleEvent(e)

	lines := decodeLines(t, buf)
	if len(lines) != 1 {
		t.Fatalf("expected 1 JSON line, got %d", len(lines))
	}
	if lines[0]["type"] != "text" {
		t.Errorf("expected type=text, got %v", lines[0]["type"])
	}
	if lines[0]["content"] != "hello" {
		t.Errorf("expected content=hello, got %v", lines[0]["content"])
	}
}

func TestJSONWriter_ToolCall(t *testing.T) {
	buf := &bytes.Buffer{}
	d := newTestDisplay(buf)

	e := events.AssistantEvent{}
	e.Type = "assistant"
	e.Message.Content = []events.ContentBlock{
		{
			Type:  "tool_use",
			Name:  "Read",
			ID:    "tool_123",
			Input: map[string]interface{}{"file_path": "/foo/bar.go"},
		},
	}
	d.HandleEvent(e)

	lines := decodeLines(t, buf)
	if len(lines) != 1 {
		t.Fatalf("expected 1 JSON line, got %d", len(lines))
	}
	if lines[0]["type"] != "tool_call" {
		t.Errorf("expected type=tool_call, got %v", lines[0]["type"])
	}
	if lines[0]["tool"] != "Read" {
		t.Errorf("expected tool=Read, got %v", lines[0]["tool"])
	}
	inputMap, ok := lines[0]["input"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected input to be a map, got %T", lines[0]["input"])
	}
	if inputMap["file_path"] != "/foo/bar.go" {
		t.Errorf("expected file_path=/foo/bar.go, got %v", inputMap["file_path"])
	}
}

func TestJSONWriter_ToolResult(t *testing.T) {
	buf := &bytes.Buffer{}
	d := newTestDisplay(buf)

	// Pre-populate PendingTools so the result lookup finds the tool name.
	d.State.PendingTools["tool_456"] = &PendingToolCall{
		ID:    "tool_456",
		Name:  "Read",
		Input: map[string]interface{}{"file_path": "/foo/bar.go"},
	}

	e := events.UserEvent{}
	e.Type = "user"
	e.Message.Role = "user"
	e.Message.Content = []events.ContentBlock{
		{
			Type:          "tool_result",
			ToolUseID:     "tool_456",
			ContentString: "line1\nline2\nline3\n",
		},
	}
	d.HandleEvent(e)

	lines := decodeLines(t, buf)
	if len(lines) != 1 {
		t.Fatalf("expected 1 JSON line, got %d", len(lines))
	}
	if lines[0]["type"] != "tool_result" {
		t.Errorf("expected type=tool_result, got %v", lines[0]["type"])
	}
	if lines[0]["tool"] != "Read" {
		t.Errorf("expected tool=Read, got %v", lines[0]["tool"])
	}
	summary, _ := lines[0]["summary"].(string)
	if !strings.HasPrefix(summary, "Read ") {
		t.Errorf("expected summary to start with 'Read ', got %q", summary)
	}
}

func TestJSONWriter_Result(t *testing.T) {
	buf := &bytes.Buffer{}
	d := newTestDisplay(buf)

	e := events.ResultEvent{}
	e.Type = "result"
	e.TotalCostUSD = 0.0042
	e.DurationMS = 3210
	e.NumTurns = 2
	e.IsError = false
	d.HandleEvent(e)

	lines := decodeLines(t, buf)
	if len(lines) != 1 {
		t.Fatalf("expected 1 JSON line, got %d", len(lines))
	}
	if lines[0]["type"] != "result" {
		t.Errorf("expected type=result, got %v", lines[0]["type"])
	}
	if lines[0]["is_error"] != false {
		t.Errorf("expected is_error=false, got %v", lines[0]["is_error"])
	}
	if lines[0]["turns"] != float64(2) {
		t.Errorf("expected turns=2, got %v", lines[0]["turns"])
	}
}

func TestJSONWriter_DisplayUnchanged(t *testing.T) {
	displayBuf := &bytes.Buffer{}
	jsonBuf := &bytes.Buffer{}

	formatter := NewFormatter(false, false, displayBuf)
	d := NewDisplay(formatter, VerbosityNormal)
	d.JSONWriter = jsonBuf

	// Send a text delta — should go to JSON buf only (not display buf, since
	// PlainNoNewline writes to formatter.Writer == displayBuf).
	e := events.StreamEvent{}
	e.Event.Type = "content_block_delta"
	e.Event.Delta = &events.Delta{Text: "streaming text"}
	d.HandleEvent(e)

	// JSON buf should have one line.
	if jsonBuf.Len() == 0 {
		t.Error("expected JSON output, got nothing")
	}

	// Display buf receives the streamed text from handleContentBlockDelta.
	if !strings.Contains(displayBuf.String(), "streaming text") {
		t.Errorf("expected display buf to contain streamed text, got %q", displayBuf.String())
	}
}
