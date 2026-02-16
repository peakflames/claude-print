# PRD: claude-print - Real-time CLI Progress Wrapper

## Introduction

`claude-print` is a cross-platform command-line tool written in Go that wraps the official Claude CLI to provide real-time progress feedback during headless execution. When running Claude CLI in headless mode with `-p`, users receive no output until the entire operation completes, which can be frustrating for long-running tasks. claude-print solves this by invoking Claude with streaming JSON output enabled, parsing the JSON stream in real-time, and presenting distilled, user-friendly progress information to the console.

## Goals

- Provide real-time progress visibility when running Claude CLI in headless mode
- Distribute as a single executable binary for Windows, macOS (Intel & Apple Silicon), and Linux
- Pass through all user-provided arguments to the underlying Claude CLI seamlessly
- Parse JSON streaming output to display meaningful progress indicators
- Auto-detect Claude CLI installation path and cache it for future use
- Support multiple output verbosity levels (quiet, normal, verbose)
- Format errors in a user-friendly way while preserving error context

## User Stories

### US-001: Initial setup and Claude CLI path detection
**Description:** As a user running claude-print for the first time, I want the tool to automatically find my Claude CLI installation so I don't have to manually configure it.

**Acceptance Criteria:**
- [ ] On first run, check if `~/.claude-print-config.json` exists
- [ ] If config doesn't exist, detect OS (Windows/Mac/Linux)
- [ ] On Linux/Mac: run `which claude` to find the executable path
- [ ] On Windows: run `where claude` to find the executable path
- [ ] Save detected path to `~/.claude-print-config.json` in user's home directory
- [ ] Handle case where Claude CLI is not found: display clear error message with installation instructions
- [ ] Config file includes: `claude_path`, `default_verbosity`, `color_enabled`, `emoji_enabled`
- [ ] Typecheck passes
- [ ] Cross-compile successfully for all target platforms

### US-002: Config file validation and error handling
**Description:** As a user, if my Claude CLI path becomes invalid (moved, uninstalled), I want clear instructions on how to fix the configuration.

**Acceptance Criteria:**
- [ ] On startup, verify that the `claude_path` from config file exists and is executable
- [ ] If path is invalid, show error: "Claude CLI not found at [path]. Please update ~/.claude-print-config.json or delete it to auto-detect."
- [ ] Provide example config JSON in error message
- [ ] Do not proceed with execution if Claude CLI cannot be found
- [ ] Typecheck passes

### US-003: Pass-through argument handling
**Description:** As a user, I want to pass any arguments I would normally use with `claude` directly to `claude-print` so I can use it as a drop-in replacement.

**Acceptance Criteria:**
- [ ] Accept all command-line arguments from user (e.g., `claude-print "write hello world"`)
- [ ] Internally construct invocation: `claude -p "$user_prompt" --dangerously-skip-permissions --include-partial-messages --output-format=stream-json`
- [ ] Preserve all user-provided flags/arguments in the constructed command
- [ ] Set environment variables from claude-print's process to be inherited by the spawned Claude CLI process
- [ ] Typecheck passes
- [ ] Test with various argument combinations (quoted strings, multiple flags, etc.)

### US-004: Spawn Claude CLI process and capture streaming output
**Description:** As a developer, I need to spawn the Claude CLI as a separate process and read its streaming JSON output line-by-line.

**Acceptance Criteria:**
- [ ] Use Go's `os/exec` package to spawn Claude CLI process
- [ ] Capture stdout as a pipe
- [ ] Environment variables are inherited (default behavior)
- [ ] Read stdout line-by-line (each line is a JSON object)
- [ ] Handle process errors (exit codes, stderr)
- [ ] Typecheck passes

### US-005: Parse streaming JSON events
**Description:** As a developer, I need to parse each JSON line from Claude's stream to extract relevant information for display.

**Acceptance Criteria:**
- [ ] Use Go's `encoding/json` to decode each line
- [ ] Support all event types from stream: `loop_start_marker`, `system`, `stream_event`, `assistant`, `user`
- [ ] Extract nested event types within `stream_event`: `message_start`, `content_block_start`, `content_block_delta`, `message_delta`, `message_stop`, `tool_use`
- [ ] Handle malformed JSON gracefully (log error, continue processing)
- [ ] Track message flow: text deltas, tool calls, tool results
- [ ] Typecheck passes

### US-006: Display real-time progress (normal verbosity)
**Description:** As a user running claude-print, I want to see real-time updates showing what Claude is currently doing so I know the process is working.

**Acceptance Criteria:**
- [ ] Mimic the visual format shown in the reference screenshot
- [ ] Show high-level activity: "Planning...", "Reading file X", "Running command Y", "Writing file Z"
- [ ] Display tool names when tool_use events are detected (e.g., "Bash", "Write", "Read")
- [ ] Show text output from Claude as it streams (content_block_delta events)
- [ ] Update display in real-time without excessive flicker
- [ ] Clear visual separation between phases (message_start, message_stop)
- [ ] Show completion message when Claude finishes
- [ ] Typecheck passes

### US-007: Display detailed progress (verbose mode)
**Description:** As a user debugging an issue, I want to see all tool calls, their parameters, and their results so I can understand exactly what Claude is doing.

**Acceptance Criteria:**
- [ ] Add `--verbose` flag to enable detailed output
- [ ] Show full tool call parameters (command, description, file paths, etc.)
- [ ] Display tool call results (stdout, stderr, exit codes)
- [ ] Show token usage and cost information from `message_delta` events
- [ ] Display timing information (timestamps from events)
- [ ] Print session metadata (session_id, model, permission mode)
- [ ] Typecheck passes

### US-008: Display minimal progress (quiet mode)
**Description:** As a user running claude-print in a script, I want a quiet mode that shows only essential information so I can parse the output programmatically.

**Acceptance Criteria:**
- [ ] Add `--quiet` flag to enable minimal output
- [ ] Show only: start indicator, completion indicator, and errors
- [ ] Suppress all intermediate progress updates
- [ ] Preserve final text output from Claude
- [ ] Return appropriate exit codes (0 for success, non-zero for errors)
- [ ] Typecheck passes

### US-009: Error handling and formatting
**Description:** As a user, when Claude CLI encounters an error, I want to see a clearly formatted error message that helps me understand what went wrong.

**Acceptance Criteria:**
- [ ] Detect error events in JSON stream (`is_error: true` in tool results)
- [ ] Detect non-zero exit codes from Claude CLI process
- [ ] Format errors with clear headers: "ERROR: [description]"
- [ ] Preserve error context (exit codes, stderr output)
- [ ] Show user-facing error messages for common errors (permissions, file not found, etc.)
- [ ] Return same exit code as Claude CLI process
- [ ] Typecheck passes

### US-010: Cross-platform build and distribution
**Description:** As a developer, I want to build claude-print for all target platforms from a single command so users can download the right binary for their OS.

**Acceptance Criteria:**
- [ ] Create build script that cross-compiles for: Windows (amd64), macOS Intel (amd64), macOS Silicon (arm64), Linux (amd64)
- [ ] Output binaries: `claude-print-windows-amd64.exe`, `claude-print-darwin-amd64`, `claude-print-darwin-arm64`, `claude-print-linux-amd64`
- [ ] Binaries are statically linked (no external dependencies)
- [ ] Verify binary sizes are reasonable (<15MB each)
- [ ] Test each binary on target platform (manual or CI)
- [ ] Document installation instructions in README

### US-011: Configuration preferences (colors, emojis)
**Description:** As a user, I want to configure output preferences like colors and emoji support so the output matches my terminal capabilities.

**Acceptance Criteria:**
- [ ] Config file supports `color_enabled` (boolean, default: true)
- [ ] Config file supports `emoji_enabled` (boolean, default: true)
- [ ] When `color_enabled` is false, output plain text without ANSI color codes
- [ ] When `emoji_enabled` is false, don't use emoji characters in output
- [ ] Allow override via command-line flags: `--no-color`, `--no-emoji`
- [ ] Detect if stdout is not a TTY (piped) and default to no colors
- [ ] Typecheck passes

### US-012: Display final result summary
**Description:** As a user, when Claude finishes execution, I want to see a summary of the session including cost, duration, and token usage so I can track resource consumption.

**Acceptance Criteria:**
- [ ] Detect `type: "result"` event at end of stream
- [ ] Parse and display total cost in USD (`total_cost_usd`)
- [ ] Display total duration (`duration_ms`) and API time (`duration_api_ms`)
- [ ] Show number of conversation turns (`num_turns`)
- [ ] Display token usage breakdown by model (from `modelUsage` field)
- [ ] Format output clearly: "Session complete: 4 turns, 29.7s total (27.1s API), $0.16"
- [ ] In verbose mode, show detailed per-model token counts (input, output, cache reads)
- [ ] Show permission denials if any occurred
- [ ] Typecheck passes

## Functional Requirements

- FR-1: On first run, auto-detect Claude CLI executable path using `which claude` (Unix) or `where claude` (Windows)
- FR-2: Store configuration in `~/.claude-print-config.json` including: `claude_path`, `default_verbosity`, `color_enabled`, `emoji_enabled`
- FR-3: Validate Claude CLI path on startup and provide clear error message if invalid
- FR-4: Accept all command-line arguments as pass-through to Claude CLI
- FR-5: Invoke Claude CLI with flags: `-p $prompt --dangerously-skip-permissions --include-partial-messages --output-format=stream-json`
- FR-6: Parse streaming JSON output line-by-line using Go's `encoding/json`
- FR-7: Display real-time progress updates based on JSON events (tool calls, text deltas, etc.)
- FR-8: Support three verbosity modes: quiet (`--quiet`), normal (default), verbose (`--verbose`)
- FR-9: Format and display errors from Claude CLI in a user-friendly way
- FR-10: Return the same exit code as the Claude CLI process
- FR-11: Cross-compile to single binaries for Windows, macOS (Intel/ARM), and Linux
- FR-12: Respect output preferences (colors, emojis) from config and command-line flags
- FR-13: Display final result summary with cost, duration, and token usage when `type: "result"` event is received
- FR-14: Handle all stream event types: system (init, hooks), stream_event (message lifecycle, content blocks), assistant, user, and result

## Non-Goals (Out of Scope)

- Interactive prompting or modifying Claude CLI behavior beyond what's specified
- Logging or recording of Claude conversations to disk (users can use shell redirection if needed)
- Retry logic or error recovery for failed Claude operations
- Advanced terminal UI features (progress bars, split panes, etc.) - keep it simple
- Parsing or understanding the semantic meaning of Claude's responses
- Support for non-streaming Claude CLI modes
- Configuration via environment variables (only config file and flags)
- Automatic updates or version checking

## Design Considerations

### Output Format
The tool should mimic the visual style shown in the reference screenshot:
- Clear phase indicators (e.g., "🔍 Reading file: src/main.go")
- Tool call descriptions (e.g., "⚡ Running: ls -la ./tmp/")
- Streaming text output from Claude in real-time
- Visual separators between major phases
- Color coding: info (blue), success (green), error (red), warning (yellow)
- Emoji indicators for different activity types (when enabled)

### JSON Stream Event Types
Based on example stream-json output, handle these event types:

**Top-level event types:**
- `system`: System events with subtypes (init, hook_started, hook_response)
  - `system.init`: Session metadata (cwd, tools, model, permission mode, etc.)
  - `system.hook_started`: Hook execution started
  - `system.hook_response`: Hook execution completed
- `stream_event`: Streaming events from Claude API (contains nested `event` object)
- `assistant`: Complete assistant message summary
- `user`: User message (often tool results)
- `result`: **Final event** with complete session summary including:
  - Total cost (`total_cost_usd`)
  - Token usage breakdown (`usage`, `modelUsage`)
  - Duration metrics (`duration_ms`, `duration_api_ms`)
  - Number of turns (`num_turns`)
  - Final result text
  - Stop reason
  - Permission denials list

**Nested event types within `stream_event`:**
- `message_start`: New message begins
- `content_block_start`: New content block (text or tool_use)
- `content_block_delta`: Incremental updates
  - Contains `text_delta` for streaming text
  - Contains `input_json_delta` for tool parameters
- `content_block_stop`: Content block complete
- `message_delta`: Token usage, cost, stop reason updates
- `message_stop`: Message complete

**Content block types:**
- `text`: Text content from Claude
- `tool_use`: Tool call with name, id, and input parameters
- `tool_result`: Result from tool execution

### Configuration Schema
```json
{
  "claude_path": "/usr/local/bin/claude",
  "default_verbosity": "normal",
  "color_enabled": true,
  "emoji_enabled": true
}
```

## Technical Considerations

### Go Packages
- `os/exec`: Process spawning and management
- `encoding/json`: JSON parsing (use `json.Decoder` for streaming)
- `bufio`: Line-by-line reading from stdout
- `os`: File system operations, environment variables
- `path/filepath`: Cross-platform path handling
- Standard library ANSI color codes (or simple third-party lib like fatih/color)

### Platform Detection
- Use `runtime.GOOS` to detect platform
- Windows: use `where` command
- Unix (Linux/macOS): use `which` command
- Home directory: use `os.UserHomeDir()`

### Process Management
- Inherit environment variables (default in `os/exec`)
- Capture stdout via pipe, let stderr pass through to terminal
- Handle signals (SIGINT, SIGTERM) to gracefully terminate child process
- Use `cmd.Wait()` to get final exit code

### Performance
- JSON parsing should be fast enough for real-time display (Go's decoder is efficient)
- Avoid buffering entire output in memory
- Update display without excessive writes (buffer output if needed)

## Success Metrics

- Users can run `claude-print "task"` and see real-time progress instead of waiting for completion
- Binary size under 15MB for all platforms
- First-run experience is smooth (auto-detects Claude CLI path)
- Output format is clear and matches reference screenshot style
- Errors are presented in a user-friendly, actionable way
- Cross-platform builds work without platform-specific code changes (except path detection)

## Open Questions

- Should we support config file in multiple locations (e.g., project-specific `.claude-print.json`)?
  - **Decision:** Start with global config only, can add project-specific later if needed
- Should we cache/buffer output to reduce flicker, or prioritize real-time updates?
  - **Decision:** Prioritize real-time, add buffering only if flicker becomes problematic
- Should we provide a `--config` flag to override config file path?
  - **Decision:** Yes, add `--config` flag for advanced users
- How should we handle very long tool call parameters in verbose mode?
  - **Decision:** Truncate with "..." and offer `--full-verbose` flag if needed
