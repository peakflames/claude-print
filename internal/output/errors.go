package output

import (
	"fmt"
	"strings"

	"github.com/peakflames/claude-print/internal/events"
)

// Common error codes and their user-friendly messages.
var errorMessages = map[int]string{
	1:   "General error - Claude CLI exited with an error",
	2:   "Misuse of shell command or invalid arguments",
	126: "Command invoked cannot execute - permission problem or not executable",
	127: "Command not found - Claude CLI binary may have been moved or deleted",
	128: "Invalid exit argument",
	130: "Terminated by Ctrl+C (SIGINT)",
	137: "Process killed (SIGKILL) - possibly ran out of memory",
	143: "Process terminated (SIGTERM)",
}

// ErrorContext holds information about an error for display.
type ErrorContext struct {
	IsError   bool
	ExitCode  int
	Message   string
	Stderr    string
	ToolName  string
	ToolError string
}

// NewErrorContext creates a new ErrorContext with default values.
func NewErrorContext() *ErrorContext {
	return &ErrorContext{
		IsError:  false,
		ExitCode: 0,
	}
}

// DetectToolError checks if a content block represents a tool error.
// Returns the error context if an error is detected.
func DetectToolError(block *events.ContentBlock) *ErrorContext {
	if block == nil {
		return nil
	}

	if block.Type == "tool_result" && block.IsError {
		return &ErrorContext{
			IsError:   true,
			ToolName:  "",
			ToolError: block.ContentString,
			Message:   fmt.Sprintf("Tool error: %s", truncateErrorMessage(block.ContentString, 500)),
		}
	}

	return nil
}

// DetectResultError checks if a ResultEvent indicates an error.
func DetectResultError(result events.ResultEvent) *ErrorContext {
	if !result.IsError {
		return nil
	}

	return &ErrorContext{
		IsError: true,
		Message: result.Result,
	}
}

// DetectExitCodeError creates an error context for non-zero exit codes.
func DetectExitCodeError(exitCode int, stderr string) *ErrorContext {
	if exitCode == 0 {
		return nil
	}

	ctx := &ErrorContext{
		IsError:  true,
		ExitCode: exitCode,
		Stderr:   stderr,
	}

	// Map exit code to user-friendly message
	if msg, ok := errorMessages[exitCode]; ok {
		ctx.Message = msg
	} else if exitCode > 128 && exitCode < 256 {
		// Exit codes > 128 indicate signal termination (128 + signal number)
		signalNum := exitCode - 128
		ctx.Message = fmt.Sprintf("Process terminated by signal %d", signalNum)
	} else {
		ctx.Message = fmt.Sprintf("Claude CLI exited with code %d", exitCode)
	}

	return ctx
}

// FormatError formats an error context for display.
// Returns the formatted error string with 'ERROR:' prefix.
func FormatError(ctx *ErrorContext) string {
	if ctx == nil || !ctx.IsError {
		return ""
	}

	var parts []string

	// Add the main error message with ERROR: prefix
	parts = append(parts, fmt.Sprintf("ERROR: %s", ctx.Message))

	// Add exit code if present
	if ctx.ExitCode != 0 {
		parts = append(parts, fmt.Sprintf("Exit code: %d", ctx.ExitCode))
	}

	// Add stderr content if present (truncated)
	if ctx.Stderr != "" {
		stderr := truncateErrorMessage(ctx.Stderr, 500)
		parts = append(parts, fmt.Sprintf("Details: %s", stderr))
	}

	// Add tool error content if present
	if ctx.ToolError != "" {
		parts = append(parts, fmt.Sprintf("Tool output: %s", truncateErrorMessage(ctx.ToolError, 500)))
	}

	return strings.Join(parts, "\n")
}

// DisplayError displays an error using the provided formatter.
func DisplayError(f *Formatter, ctx *ErrorContext) {
	if ctx == nil || !ctx.IsError {
		return
	}

	// Display main error message with ERROR: prefix in red
	f.ErrorWithEmoji(EmojiError, "ERROR: %s", ctx.Message)

	// Display exit code if present
	if ctx.ExitCode != 0 {
		f.Error("Exit code: %d", ctx.ExitCode)
	}

	// Display stderr content if present (truncated)
	if ctx.Stderr != "" {
		stderr := truncateErrorMessage(ctx.Stderr, 500)
		if stderr != "" {
			f.Error("Details: %s", stderr)
		}
	}

	// Display tool error content if present
	if ctx.ToolError != "" {
		toolErr := truncateErrorMessage(ctx.ToolError, 500)
		if toolErr != "" {
			f.Error("Tool output: %s", toolErr)
		}
	}
}

// truncateErrorMessage truncates an error message to the specified max length.
func truncateErrorMessage(msg string, maxLen int) string {
	// Remove leading/trailing whitespace
	msg = strings.TrimSpace(msg)

	if len(msg) <= maxLen {
		return msg
	}

	// Truncate and add ellipsis
	return msg[:maxLen-3] + "..."
}

// IsToolResultError checks if a stream event contains a tool result error.
func IsToolResultError(e events.StreamEvent) bool {
	if e.Event.Type != "content_block_start" {
		return false
	}

	if e.Event.ContentBlock == nil {
		return false
	}

	return e.Event.ContentBlock.Type == "tool_result" && e.Event.ContentBlock.IsError
}

// GetToolResultError extracts error information from a tool result event.
func GetToolResultError(e events.StreamEvent) *ErrorContext {
	if !IsToolResultError(e) {
		return nil
	}

	return DetectToolError(e.Event.ContentBlock)
}

// MapCommonError maps common error patterns to user-friendly messages.
func MapCommonError(errorContent string) string {
	errorLower := strings.ToLower(errorContent)

	// Permission errors
	if strings.Contains(errorLower, "permission denied") {
		return "Permission denied - check file or directory permissions"
	}
	if strings.Contains(errorLower, "eacces") {
		return "Access denied - insufficient permissions for this operation"
	}

	// File errors
	if strings.Contains(errorLower, "no such file or directory") ||
		strings.Contains(errorLower, "enoent") {
		return "File or directory not found"
	}
	if strings.Contains(errorLower, "file exists") ||
		strings.Contains(errorLower, "eexist") {
		return "File already exists"
	}
	if strings.Contains(errorLower, "is a directory") ||
		strings.Contains(errorLower, "eisdir") {
		return "Expected a file but found a directory"
	}
	if strings.Contains(errorLower, "not a directory") ||
		strings.Contains(errorLower, "enotdir") {
		return "Expected a directory but found a file"
	}

	// Process errors
	if strings.Contains(errorLower, "command not found") {
		return "Command not found - ensure the required tool is installed and in PATH"
	}
	if strings.Contains(errorLower, "timeout") ||
		strings.Contains(errorLower, "timed out") {
		return "Operation timed out"
	}

	// Network errors
	if strings.Contains(errorLower, "connection refused") {
		return "Connection refused - service may not be running"
	}
	if strings.Contains(errorLower, "network unreachable") ||
		strings.Contains(errorLower, "no route to host") {
		return "Network unreachable - check your network connection"
	}

	// Memory errors
	if strings.Contains(errorLower, "out of memory") ||
		strings.Contains(errorLower, "cannot allocate memory") {
		return "Out of memory - try closing other applications"
	}

	// Return original error if no common pattern matches
	return errorContent
}
