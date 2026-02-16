# claude-print

CLI wrapper for Claude CLI that provides real-time progress feedback during headless execution.

## Build Commands

```bash
uv run make.py build      # Build for current platform
uv run make.py build-all  # Build for all platforms (dist/)
uv run make.py test       # Run tests
uv run make.py fmt        # Format code
uv run make.py vet        # Run go vet
uv run make.py clean      # Remove build artifacts
```

## Process Management Commands

```bash
uv run make.py start "prompt"  # Build and start in background
uv run make.py stop            # Stop background process
uv run make.py status          # Check if running
uv run make.py log             # Display recent logs
uv run make.py run             # Build and run in foreground (test)
```

## Project Layout

- `cmd/claude-print/` - Entry point (main.go)
- `internal/` - Private packages:
  - `cli/` - Flag parsing
  - `config/` - Configuration (~/.claude-print-config.json)
  - `detect/` - Claude CLI auto-detection
  - `events/` - JSON event stream parsing
  - `output/` - Display formatting, colors, errors
  - `runner/` - Process execution, signal handling

## Conventions

- Go 1.21+, standard project layout
- Platform-specific code uses build tags (`//go:build`)
- Signal handling: `signal_unix.go` / `signal_windows.go`
- Use `filepath.Join()` for paths
- Cross-platform binaries: `dist/claude-print-{os}-{arch}`

## Pre-Commit Protocol

**ALWAYS run before committing:**

```bash
uv run make.py fmt && uv run make.py vet && git commit -am "message"
```

- `fmt` - Formats all Go code to standard style
- `vet` - Static analysis to catch common bugs
- Run `uv run make.py test` if you've changed core logic

## Example Usage

```bash
./claude-print "What is 2+2?"           # Basic test
./claude-print --verbose "test"         # Verbose output
./claude-print --quiet "test"           # Minimal output
```

## Development Patterns

### Build Script Dependencies

`make.py` uses uv's inline script metadata for dependencies:
```python
# /// script
# requires-python = ">=3.12"
# dependencies = ["psutil"]
# ///
```
This allows `uv run make.py` to auto-install dependencies without requirements.txt.

### Windows Console Behavior

When spawning background processes on Windows:
- **DO NOT use `DETACHED_PROCESS`** - it creates a visible console window
- **USE `CREATE_NEW_PROCESS_GROUP`** only - prevents Ctrl+C propagation without popup
- Use `psutil` for cross-platform process management instead of ctypes/kernel32

### Unicode/Emoji Output on Windows

Configure encoding at script startup to avoid UnicodeEncodeError:
```python
if sys.platform == "win32":
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")
```

### Cross-Platform Process Management

Use `psutil` library for robust process detection and termination:
- Works identically on Windows, macOS, Linux
- Handles zombie processes, access denied, process not found
- Graceful termination with timeout fallback to force kill

### Claude CLI Streaming JSON Pattern

With `--include-partial-messages`, tool inputs stream incrementally:
- `content_block_start` has **empty** `input: {}`
- Full input arrives in subsequent `"type": "assistant"` event
- **Display tool calls from `assistant` event, not `content_block_start`**

Text content streams via `content_block_delta`:
- Don't re-display text from `assistant` events (causes duplicates)
- Only display `tool_use` blocks from `assistant` events

Reference files:
- `docs/sample-stream-with-partial.json` - Sample JSON output (this is what we use)
- `docs/streaming-event-processing.md` - Detailed event flow documentation
