package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/peakflames/claude-print/internal/events"
)

// Verbosity levels for display output.
type Verbosity int

const (
	// VerbosityQuiet shows minimal output (start/end indicators only).
	VerbosityQuiet Verbosity = iota
	// VerbosityNormal shows real-time progress updates.
	VerbosityNormal
	// VerbosityVerbose shows detailed tool call information.
	VerbosityVerbose
)

// Visual indicators for Claude Code style output
const (
	Bullet     = "\u25cf"     // ● solid circle
	TreeBranch = "  \u23bf  " // ⎿ indented tree branch for results
	UserPrefix = "> User: "
)

// Legacy emojis kept for error handling compatibility
const (
	EmojiError   = "\u274c"       // ❌
	EmojiWarning = "\u26a0\ufe0f" // ⚠️
	EmojiDone    = "\u2705"       // ✅
)

// PendingToolCall tracks a tool invocation awaiting its result
type PendingToolCall struct {
	ID    string
	Name  string
	Input map[string]interface{}
}

// DisplayState tracks state across events
type DisplayState struct {
	UserPrompt              string
	PendingTools            map[string]*PendingToolCall
	LastOutputWasText       bool // Track if we need newline before tool output
	InTextBlock             bool // Track if we're currently in a text block
	LastMessageWasToolUse   bool // Track if last message was tool use (suppress extra newline)
	ToolResultJustDisplayed bool // Track if we just showed a tool result
}

// Display handles event display with configurable verbosity and formatting.
type Display struct {
	Formatter *Formatter
	Verbosity Verbosity
	Writer    io.Writer
	State     *DisplayState
}

// NewDisplay creates a new Display with the specified settings.
func NewDisplay(formatter *Formatter, verbosity Verbosity) *Display {
	var writer io.Writer = os.Stdout
	if formatter != nil && formatter.Writer != nil {
		writer = formatter.Writer
	}
	return &Display{
		Formatter: formatter,
		Verbosity: verbosity,
		Writer:    writer,
		State: &DisplayState{
			PendingTools: make(map[string]*PendingToolCall),
		},
	}
}

// SetUserPrompt sets the user prompt for display in the header
func (d *Display) SetUserPrompt(prompt string) {
	d.State.UserPrompt = prompt
}

// HandleEvent processes an event and outputs appropriate display text
// based on the current verbosity level.
func (d *Display) HandleEvent(event events.Event) {
	switch d.Verbosity {
	case VerbosityQuiet:
		d.handleQuietEvent(event)
	case VerbosityNormal:
		d.handleNormalEvent(event)
	case VerbosityVerbose:
		d.handleVerboseEvent(event)
	}
}

// handleNormalEvent handles events in normal verbosity mode.
// Shows tool use summaries, streams text, and displays message separators.
func (d *Display) handleNormalEvent(event events.Event) {
	switch e := event.(type) {
	case events.StreamEvent:
		d.handleStreamEvent(e)
	case events.AssistantMessageEvent:
		d.handleAssistantMessage(e)
	case events.AssistantEvent:
		// Top-level "assistant" event with complete tool_use inputs
		d.handleAssistantEvent(e)
	case events.UserEvent:
		d.handleUserEvent(e)
	case events.ResultEvent:
		d.showResultSummary(e, false)
	case events.SystemEvent:
		// System events are not shown in normal mode
	}
}

// handleQuietEvent handles events in quiet mode with minimal output.
// Shows only start/end indicators, errors, and preserves final text result.
func (d *Display) handleQuietEvent(event events.Event) {
	switch e := event.(type) {
	case events.StreamEvent:
		d.handleQuietStreamEvent(e)
	case events.ResultEvent:
		d.showQuietCompletion(e)
	case events.AssistantMessageEvent:
		// In quiet mode, only show errors from assistant messages
		for _, block := range e.Message.Content {
			if block.Type == "tool_result" && block.IsError {
				d.Formatter.Error("%s%s", TreeBranch, block.Content)
			}
		}
	case events.AssistantEvent:
		// In quiet mode, ignore assistant events (tool calls)
	case events.UserEvent:
		// Show errors in quiet mode
		for _, block := range e.Message.Content {
			if block.Type == "tool_result" && block.IsError {
				d.Formatter.Error("%s%s", TreeBranch, block.Content)
			}
		}
	case events.SystemEvent:
		// System events are suppressed in quiet mode
	}
}

// handleQuietStreamEvent processes stream events in quiet mode.
// Only displays errors, suppresses all other intermediate progress.
func (d *Display) handleQuietStreamEvent(e events.StreamEvent) {
	switch e.Event.Type {
	case "content_block_start":
		// Only show errors in quiet mode
		if e.Event.ContentBlock != nil && e.Event.ContentBlock.Type == "tool_result" && e.Event.ContentBlock.IsError {
			d.Formatter.Error("%s%s", TreeBranch, e.Event.ContentBlock.Content)
		}
	case "content_block_delta":
		// Stream final text output (important to preserve Claude's response)
		if e.Event.Delta != nil && e.Event.Delta.Text != "" {
			d.Formatter.PlainNoNewline("%s", e.Event.Delta.Text)
		}
	case "message_stop":
		// Add newline after streaming text if there was any
		fmt.Fprintln(d.Writer)
	}
}

// showQuietCompletion displays minimal completion message in quiet mode.
// Shows session summary with cost and duration even in quiet mode.
func (d *Display) showQuietCompletion(e events.ResultEvent) {
	// Show error if the result indicates an error
	if e.IsError {
		d.Formatter.Error("Session ended with error")
		if e.Result != "" {
			d.Formatter.Error("%s", e.Result)
		}
		return
	}

	// Format duration values
	totalDuration := formatDuration(e.DurationMS)
	apiDuration := formatDuration(e.DurationAPIMS)
	cost := formatCost(e.TotalCostUSD)

	// Calculate total tokens from model usage
	totalIn, totalOut := calculateTotalTokens(e)

	// Display summary line with token counts
	d.Formatter.Success("Session complete: %d turns, %s total (%s API), %d in / %d out, %s",
		e.NumTurns, totalDuration, apiDuration, totalIn, totalOut, cost)

	// Show condensed per-model usage
	d.showModelUsageSummary(e)
}

// handleVerboseEvent handles events in verbose mode with detailed output.
// Shows full tool parameters, results, token usage, and session metadata.
func (d *Display) handleVerboseEvent(event events.Event) {
	switch e := event.(type) {
	case events.StreamEvent:
		d.handleVerboseStreamEvent(e)
	case events.AssistantMessageEvent:
		d.handleVerboseAssistantMessage(e)
	case events.AssistantEvent:
		// Top-level "assistant" event with complete tool_use inputs
		d.handleVerboseAssistantEvent(e)
	case events.UserEvent:
		d.handleUserEvent(e)
	case events.ResultEvent:
		d.showResultSummary(e, true)
	case events.SystemEvent:
		d.handleVerboseSystemEvent(e)
	}
}

// handleVerboseStreamEvent processes stream events with detailed output.
func (d *Display) handleVerboseStreamEvent(e events.StreamEvent) {
	switch e.Event.Type {
	case "message_start":
		d.showVerboseMessageStart(e)
	case "message_stop":
		d.showMessageStop()
	case "content_block_start":
		d.handleVerboseContentBlockStart(e)
	case "content_block_delta":
		d.handleVerboseContentBlockDelta(e)
	case "message_delta":
		d.handleMessageDelta(e)
	}
}

// handleVerboseContentBlockStart processes content block start with full details.
// NOTE: With --include-partial-messages, tool_use input is empty here.
// The full input comes in the subsequent "assistant" event, so we skip tool_use display.
func (d *Display) handleVerboseContentBlockStart(e events.StreamEvent) {
	if e.Event.ContentBlock == nil {
		return
	}

	block := e.Event.ContentBlock
	switch block.Type {
	case "tool_use":
		// Track tool but don't display - input is empty here
		// Full display happens in handleVerboseAssistantEvent
		d.State.PendingTools[block.ID] = &PendingToolCall{
			ID:    block.ID,
			Name:  block.Name,
			Input: block.Input,
		}
	case "tool_result":
		d.showVerboseToolResult(block)
	}
}

// handleVerboseContentBlockDelta processes content deltas with full details.
func (d *Display) handleVerboseContentBlockDelta(e events.StreamEvent) {
	if e.Event.Delta == nil {
		return
	}

	// Stream text output in real-time
	if e.Event.Delta.Text != "" {
		d.Formatter.PlainNoNewline("%s", e.Event.Delta.Text)
	}
}

// handleMessageDelta processes message_delta events for token usage.
func (d *Display) handleMessageDelta(e events.StreamEvent) {
	if e.Event.Usage != nil {
		d.showTokenUsage(e.Event.Usage)
	}
}

// handleVerboseAssistantMessage processes assistant messages with full details.
func (d *Display) handleVerboseAssistantMessage(e events.AssistantMessageEvent) {
	for _, block := range e.Message.Content {
		switch block.Type {
		case "tool_use":
			d.showVerboseToolUse(block.Name, block.ID, block.Input)
		case "tool_result":
			d.showVerboseToolResult(&block)
		}
	}
}

// handleVerboseAssistantEvent processes top-level "assistant" events with full details.
// NOTE: Text content is NOT displayed here because it was already streamed
// via content_block_delta events. Only tool_use needs display here.
func (d *Display) handleVerboseAssistantEvent(e events.AssistantEvent) {
	for _, block := range e.Message.Content {
		if block.Type == "tool_use" {
			d.showVerboseToolUse(block.Name, block.ID, block.Input)
		}
		// Text content is already streamed via content_block_delta, so skip here
	}
}

// handleVerboseSystemEvent displays system event metadata.
func (d *Display) handleVerboseSystemEvent(e events.SystemEvent) {
	switch e.Type {
	case "system.init":
		d.showSessionMetadata(e)
	case "hook_started":
		d.Formatter.Info("%s Hook started: %s (%s)", Bullet, e.HookName, e.HookType)
	case "hook_response":
		d.Formatter.Info("%s Hook response: %s", Bullet, e.Response)
	}
}

// showSessionMetadata displays session initialization metadata.
func (d *Display) showSessionMetadata(e events.SystemEvent) {
	d.Formatter.Info("=== Session Metadata ===")
	if e.SessionID != "" {
		d.Formatter.Plain("  Session ID: %s", e.SessionID)
	}
	if e.Model != "" {
		d.Formatter.Plain("  Model: %s", e.Model)
	}
	if e.Cwd != "" {
		d.Formatter.Plain("  Working Directory: %s", e.Cwd)
	}
	if len(e.Tools) > 0 {
		d.Formatter.Plain("  Available Tools: %d", len(e.Tools))
		for _, tool := range e.Tools {
			d.Formatter.Plain("    - %s", tool.Name)
		}
	}
	if len(e.McpServers) > 0 {
		d.Formatter.Plain("  MCP Servers: %d", len(e.McpServers))
		for _, server := range e.McpServers {
			d.Formatter.Plain("    - %s (%s)", server.Name, server.Status)
		}
	}
	d.Formatter.Plain("========================")
}

// showVerboseToolUse displays a tool use event with full parameters.
func (d *Display) showVerboseToolUse(toolName string, toolID string, input map[string]interface{}) {
	// Track pending tool for result matching
	d.State.PendingTools[toolID] = &PendingToolCall{
		ID:    toolID,
		Name:  toolName,
		Input: input,
	}

	d.Formatter.Info("%s %s", Bullet, toolName)
	d.Formatter.Plain("  Parameters:")
	for key, value := range input {
		d.formatParameterValue(key, value, "    ")
	}
}

// formatParameterValue formats a parameter value with appropriate truncation.
func (d *Display) formatParameterValue(key string, value interface{}, indent string) {
	switch v := value.(type) {
	case string:
		// Truncate very long strings (e.g., file contents)
		if len(v) > 200 {
			lines := strings.Split(v, "\n")
			if len(lines) > 5 {
				d.Formatter.Plain("%s%s: (%d lines, showing first 5)", indent, key, len(lines))
				for i := 0; i < 5 && i < len(lines); i++ {
					line := lines[i]
					if len(line) > 80 {
						line = line[:77] + "..."
					}
					d.Formatter.Plain("%s  %s", indent, line)
				}
			} else {
				d.Formatter.Plain("%s%s: %s...", indent, key, v[:197])
			}
		} else {
			d.Formatter.Plain("%s%s: %s", indent, key, v)
		}
	case bool:
		d.Formatter.Plain("%s%s: %v", indent, key, v)
	case float64:
		d.Formatter.Plain("%s%s: %v", indent, key, v)
	case nil:
		d.Formatter.Plain("%s%s: null", indent, key)
	default:
		d.Formatter.Plain("%s%s: %v", indent, key, v)
	}
}

// showVerboseToolResult displays a tool result with full output.
func (d *Display) showVerboseToolResult(block *events.ContentBlock) {
	if block.IsError {
		d.Formatter.Error("%sTool Result (ERROR):", TreeBranch)
	} else {
		d.Formatter.Success("%sTool Result:", TreeBranch)
	}

	content := block.ContentString
	if content != "" {
		lines := strings.Split(content, "\n")
		if len(lines) > 20 {
			// Show first 10 and last 5 lines for long output
			d.Formatter.Plain("  (Showing %d of %d lines)", 15, len(lines))
			for i := 0; i < 10; i++ {
				d.Formatter.Plain("  %s", truncateLine(lines[i], 120))
			}
			d.Formatter.Plain("  ...")
			for i := len(lines) - 5; i < len(lines); i++ {
				d.Formatter.Plain("  %s", truncateLine(lines[i], 120))
			}
		} else {
			for _, line := range lines {
				d.Formatter.Plain("  %s", truncateLine(line, 120))
			}
		}
	}
}

// truncateLine truncates a line to the specified max length.
func truncateLine(line string, maxLen int) string {
	if len(line) > maxLen {
		return line[:maxLen-3] + "..."
	}
	return line
}

// showTokenUsage displays token usage from message_delta events.
func (d *Display) showTokenUsage(usage *events.Usage) {
	if usage.InputTokens > 0 || usage.OutputTokens > 0 {
		d.Formatter.Info("  Tokens - Input: %d, Output: %d", usage.InputTokens, usage.OutputTokens)
		if usage.CacheReadInputTokens > 0 {
			d.Formatter.Plain("    Cache read: %d tokens", usage.CacheReadInputTokens)
		}
		if usage.CacheCreationInputTokens > 0 {
			d.Formatter.Plain("    Cache creation: %d tokens", usage.CacheCreationInputTokens)
		}
	}
}

// showVerboseMessageStart displays message start with model info if available.
func (d *Display) showVerboseMessageStart(e events.StreamEvent) {
	fmt.Fprintln(d.Writer) // Blank line before message
	if e.Event.Message != nil && e.Event.Message.Model != "" {
		d.Formatter.Info("  Model: %s", e.Event.Message.Model)
	}
}

// handleStreamEvent processes stream events containing message content.
func (d *Display) handleStreamEvent(e events.StreamEvent) {
	switch e.Event.Type {
	case "message_start":
		d.showMessageStart()
	case "message_stop":
		d.showMessageStop()
	case "content_block_start":
		d.handleContentBlockStart(e)
	case "content_block_delta":
		d.handleContentBlockDelta(e)
	case "content_block_stop":
		d.handleContentBlockStop(e)
	}
}

// handleContentBlockStart processes the start of a content block.
func (d *Display) handleContentBlockStart(e events.StreamEvent) {
	if e.Event.ContentBlock == nil {
		return
	}

	block := e.Event.ContentBlock
	switch block.Type {
	case "tool_use":
		// NOTE: With --include-partial-messages, the input is empty here.
		// The full input comes in the subsequent "assistant" event.
		// We track the tool ID but don't display yet.
		d.State.PendingTools[block.ID] = &PendingToolCall{
			ID:    block.ID,
			Name:  block.Name,
			Input: block.Input,
		}
	case "text":
		// Add newline before text if we have pending tool results displayed
		fmt.Fprintln(d.Writer)
		// Start text with bullet
		d.State.InTextBlock = true
		d.Formatter.PlainNoNewline("%s ", Bullet)
	case "tool_result":
		if block.IsError {
			d.Formatter.Error("%sError: %s", TreeBranch, block.Content)
		}
	}
}

// handleContentBlockDelta processes incremental content updates.
func (d *Display) handleContentBlockDelta(e events.StreamEvent) {
	if e.Event.Delta == nil {
		return
	}

	// Stream text output in real-time
	if e.Event.Delta.Text != "" {
		d.Formatter.PlainNoNewline("%s", e.Event.Delta.Text)
	}
}

// handleContentBlockStop processes the end of a content block.
func (d *Display) handleContentBlockStop(_ events.StreamEvent) {
	if d.State.InTextBlock {
		d.State.InTextBlock = false
		fmt.Fprintln(d.Writer) // Newline after text block
	}
}

// handleAssistantMessage processes complete assistant messages.
// This is called for "assistant_message" events.
func (d *Display) handleAssistantMessage(e events.AssistantMessageEvent) {
	for _, block := range e.Message.Content {
		switch block.Type {
		case "tool_use":
			// This is where we get the COMPLETE tool call with full input
			d.showToolUse(block.Name, block.ID, block.Input)
		case "tool_result":
			if block.IsError {
				d.Formatter.Error("%sError: %s", TreeBranch, block.Content)
			}
		}
	}
}

// handleAssistantEvent processes top-level "assistant" events.
// This contains the complete tool_use with full input parameters.
// NOTE: Text content is NOT displayed here because it was already streamed
// via content_block_delta events. Only tool_use needs display here.
func (d *Display) handleAssistantEvent(e events.AssistantEvent) {
	for _, block := range e.Message.Content {
		if block.Type == "tool_use" {
			// This is where we get the COMPLETE tool call with full input
			d.showToolUse(block.Name, block.ID, block.Input)
		}
		// Text content is already streamed via content_block_delta, so skip here
	}
}

// handleUserEvent handles user events containing tool results
func (d *Display) handleUserEvent(e events.UserEvent) {
	for _, block := range e.Message.Content {
		if block.Type == "tool_result" {
			// Check if this was a denied tool (error with permission message)
			if block.IsError && d.isToolDenied(block.ContentString) {
				d.showToolDenied(block.ToolUseID, block.ContentString)
			} else {
				d.showToolResult(block.ToolUseID, e.ToolUseResult, block.ContentString)
			}
		}
	}
}

// isToolDenied checks if the content indicates a permission denial
func (d *Display) isToolDenied(content string) bool {
	return strings.Contains(content, "Permission to use") && strings.Contains(content, "has been denied")
}

// showToolDenied displays a tool denial with appropriate formatting
func (d *Display) showToolDenied(toolID string, content string) {
	pending := d.State.PendingTools[toolID]
	if pending == nil {
		return
	}
	delete(d.State.PendingTools, toolID)

	// Format: ⎿ Tool denied (not in allowed-tools)
	d.Formatter.Warning("%sTool denied (not in allowed-tools)", TreeBranch)
	d.State.LastMessageWasToolUse = false
	d.State.ToolResultJustDisplayed = true
}

// showToolUse displays a tool use event with Claude Code style.
// Format: ● ToolName(param) where only ● is green
func (d *Display) showToolUse(toolName string, toolID string, input map[string]interface{}) {
	// Track pending tool for result matching
	d.State.PendingTools[toolID] = &PendingToolCall{
		ID:    toolID,
		Name:  toolName,
		Input: input,
	}

	// Format: ● ToolName(param) - only bullet is colored green
	paramStr := d.formatToolParams(toolName, input)
	var text string
	if paramStr != "" {
		text = fmt.Sprintf("%s(%s)", toolName, paramStr)
	} else {
		text = toolName
	}
	d.Formatter.ToolCall(Bullet, text)
	d.State.LastMessageWasToolUse = true
}

// formatToolParams formats tool parameters for compact display
func (d *Display) formatToolParams(toolName string, input map[string]interface{}) string {
	switch strings.ToLower(toolName) {
	case "read":
		if path, ok := input["file_path"].(string); ok {
			return path
		}
	case "glob":
		if pattern, ok := input["pattern"].(string); ok {
			return fmt.Sprintf("pattern: \"%s\"", pattern)
		}
	case "grep":
		if pattern, ok := input["pattern"].(string); ok {
			return fmt.Sprintf("pattern: \"%s\"", pattern)
		}
	case "write", "edit":
		if path, ok := input["file_path"].(string); ok {
			return path
		}
	case "bash":
		if cmd, ok := input["command"].(string); ok {
			if len(cmd) > 40 {
				cmd = cmd[:37] + "..."
			}
			return fmt.Sprintf("command: \"%s\"", cmd)
		}
	case "task":
		if desc, ok := input["description"].(string); ok {
			return desc
		}
	}
	return ""
}

// showToolResult displays a tool result with tree branch
func (d *Display) showToolResult(toolID string, result *events.ToolUseResult, content string) {
	pending := d.State.PendingTools[toolID]
	if pending == nil {
		return
	}
	delete(d.State.PendingTools, toolID)

	// Format result based on tool type
	resultStr := d.formatToolResult(pending.Name, result, content)
	d.Formatter.Plain("%s%s", TreeBranch, resultStr)

	// Reset tool use state, mark that we just displayed a result
	d.State.LastMessageWasToolUse = false
	d.State.ToolResultJustDisplayed = true
}

// formatToolResult formats tool result for display
func (d *Display) formatToolResult(toolName string, result *events.ToolUseResult, content string) string {
	switch strings.ToLower(toolName) {
	case "read":
		if result != nil && result.File != nil && result.File.NumLines > 0 {
			return fmt.Sprintf("Read %d lines", result.File.NumLines)
		}
		// Fallback: count lines in content
		lines := strings.Count(content, "\n") + 1
		if content == "" {
			lines = 0
		}
		return fmt.Sprintf("Read %d lines", lines)
	case "glob":
		// Count files found (lines in output)
		count := 0
		if content != "" {
			count = strings.Count(content, "\n")
			if !strings.HasSuffix(content, "\n") {
				count++
			}
		}
		return fmt.Sprintf("Found %d files", count)
	case "grep":
		count := 0
		if content != "" {
			count = strings.Count(content, "\n")
			if !strings.HasSuffix(content, "\n") {
				count++
			}
		}
		return fmt.Sprintf("%d matches", count)
	case "bash":
		// Show first line of output or "Done"
		if content == "" {
			return "Done"
		}
		lines := strings.SplitN(content, "\n", 2)
		first := lines[0]
		if len(first) > 60 {
			return first[:57] + "..."
		}
		if first == "" {
			return "Done"
		}
		return first
	case "write":
		return "Wrote file"
	case "edit":
		return "Edited file"
	default:
		return "Done"
	}
}

// showMessageStart displays visual indicator at message start.
func (d *Display) showMessageStart() {
	// No separator - bullet structure provides hierarchy
}

// showMessageStop ensures newline after streaming text.
// Suppresses extra newline after tool use to keep result immediately below.
func (d *Display) showMessageStop() {
	// Skip newline if we just displayed a tool use (result should appear immediately below)
	if d.State.LastMessageWasToolUse {
		return
	}
	if !d.State.InTextBlock {
		fmt.Fprintln(d.Writer)
	}
}

// showResultSummary displays the session result summary with cost and duration.
// Format: 'Session complete: N turns, X.Xs total (Y.Ys API), XXXX in / YYY out, $Z.ZZ'
// Shows per-model usage in both normal and verbose modes.
func (d *Display) showResultSummary(e events.ResultEvent, verbose bool) {
	// Check for errors first
	if e.IsError {
		d.Formatter.Error("Session ended with error")
		if e.Result != "" {
			d.Formatter.Error("%s", e.Result)
		}
		return
	}

	// Format duration values
	totalDuration := formatDuration(e.DurationMS)
	apiDuration := formatDuration(e.DurationAPIMS)

	// Format cost as currency
	cost := formatCost(e.TotalCostUSD)

	// Calculate total tokens from model usage
	totalIn, totalOut := calculateTotalTokens(e)

	// Display summary line with token counts
	d.Formatter.Success("Session complete: %d turns, %s total (%s API), %d in / %d out, %s",
		e.NumTurns, totalDuration, apiDuration, totalIn, totalOut, cost)

	// Always show per-model usage summary
	d.showModelUsageSummary(e)

	// In verbose mode, show additional detailed statistics
	if verbose {
		d.showVerboseResultDetails(e)
	}
}

// showModelUsageSummary displays per-model token counts and costs.
// Format: '  - model-name: 12345 in / 678 out (85%) $0.42'
func (d *Display) showModelUsageSummary(e events.ResultEvent) {
	if len(e.ModelUsage) == 0 {
		return
	}

	for model, usage := range e.ModelUsage {
		pct := calculateModelPercentage(usage.CostUSD, e.TotalCostUSD)
		cost := formatCost(usage.CostUSD)
		d.Formatter.Plain("  - %s: %d in / %d out (%.0f%%) %s",
			model, usage.InputTokens, usage.OutputTokens, pct, cost)
	}
}

// calculateModelPercentage calculates this model's share of total cost.
func calculateModelPercentage(modelCost, totalCost float64) float64 {
	if totalCost <= 0 {
		return 0
	}
	return (modelCost / totalCost) * 100
}

// calculateTotalTokens sums input and output tokens across all models.
func calculateTotalTokens(e events.ResultEvent) (totalIn, totalOut int) {
	for _, usage := range e.ModelUsage {
		totalIn += usage.InputTokens
		totalOut += usage.OutputTokens
	}
	return
}

// showVerboseResultDetails displays detailed session statistics in verbose mode.
func (d *Display) showVerboseResultDetails(e events.ResultEvent) {
	d.Formatter.Plain("")
	d.Formatter.Info("=== Session Statistics ===")

	// Show aggregated usage if available
	if e.Usage != nil {
		d.Formatter.Plain("  Total Tokens:")
		d.Formatter.Plain("    Input: %d", e.Usage.InputTokens)
		d.Formatter.Plain("    Output: %d", e.Usage.OutputTokens)
		if e.Usage.CacheReadInputTokens > 0 {
			d.Formatter.Plain("    Cache read: %d", e.Usage.CacheReadInputTokens)
		}
		if e.Usage.CacheCreationInputTokens > 0 {
			d.Formatter.Plain("    Cache creation: %d", e.Usage.CacheCreationInputTokens)
		}
		if e.Usage.TotalTokens > 0 {
			d.Formatter.Plain("    Total: %d", e.Usage.TotalTokens)
		}
	}

	// Show per-model usage if available
	if len(e.ModelUsage) > 0 {
		d.Formatter.Plain("")
		d.Formatter.Plain("  Per-Model Usage:")
		for model, usage := range e.ModelUsage {
			d.Formatter.Plain("    %s:", model)
			d.Formatter.Plain("      Input: %d, Output: %d", usage.InputTokens, usage.OutputTokens)
			if usage.CacheReadInputTokens > 0 {
				d.Formatter.Plain("      Cache read: %d", usage.CacheReadInputTokens)
			}
			if usage.CacheCreationInputTokens > 0 {
				d.Formatter.Plain("      Cache creation: %d", usage.CacheCreationInputTokens)
			}
		}
	}

	// Show tool usage if available
	if e.TotalToolUse > 0 {
		d.Formatter.Plain("")
		d.Formatter.Plain("  Tool Usage: %d total", e.TotalToolUse)
		if e.TotalToolErrors > 0 {
			d.Formatter.Warning("    Errors: %d", e.TotalToolErrors)
		}
		if e.TotalToolCancels > 0 {
			d.Formatter.Plain("    Cancels: %d", e.TotalToolCancels)
		}
	}

	d.Formatter.Plain("===========================")
}

// formatDuration converts milliseconds to a human-readable format (e.g., "1.5s", "2m30s").
func formatDuration(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}

	seconds := float64(ms) / 1000.0
	if seconds < 60 {
		return fmt.Sprintf("%.1fs", seconds)
	}

	minutes := int(seconds) / 60
	remainingSecs := seconds - float64(minutes*60)
	if remainingSecs > 0 {
		return fmt.Sprintf("%dm%.1fs", minutes, remainingSecs)
	}
	return fmt.Sprintf("%dm", minutes)
}

// formatCost formats a USD cost value as currency (e.g., "$0.12", "$1.50").
func formatCost(costUSD float64) string {
	if costUSD < 0.01 {
		return fmt.Sprintf("$%.4f", costUSD)
	}
	return fmt.Sprintf("$%.2f", costUSD)
}

// ShowStart displays the start indicator with user prompt.
func (d *Display) ShowStart() {
	if d.Verbosity == VerbosityQuiet {
		return
	}
	// Newline before prompt (matches Claude Code style)
	fmt.Fprintln(d.Writer)
	// Simple header format: "> User: prompt" - plain text, no color
	d.Formatter.Plain("%s%s", UserPrefix, d.State.UserPrompt)
	fmt.Fprintln(d.Writer) // Blank line after prompt
}

// ShowAllowedTools displays the allowed tools banner.
func (d *Display) ShowAllowedTools(tools string, dangerous bool) {
	if d.Verbosity == VerbosityQuiet {
		return
	}
	fmt.Fprintln(d.Writer) // Blank line before banner
	if dangerous {
		d.Formatter.Warning("AllowedTools: ALL (dangerous mode)")
	} else {
		d.Formatter.Info("AllowedTools: %s", tools)
	}
}

// ShowPermissionMode displays the permission mode banner.
func (d *Display) ShowPermissionMode(mode string) {
	if d.Verbosity == VerbosityQuiet {
		return
	}
	if mode == "" {
		return // Don't show if not specified
	}
	d.Formatter.Info("Permission Mode: %s", mode)
}
