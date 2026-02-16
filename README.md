# claude-print

A cross-platform CLI tool that wraps Claude CLI to provide real-time progress feedback during headless execution by parsing streaming JSON output.

## Features

- Real-time progress updates showing what Claude is doing
- Colored output with emoji indicators
- Three verbosity levels: quiet, normal, and verbose
- Automatic TTY detection for script-friendly piped output
- Cross-platform support (Windows, macOS, Linux)
- Graceful shutdown on Ctrl+C

## Project Structure

claude-print follows the standard Go project layout:

```
claude-print/
├── cmd/claude-print/    # Application entry point
├── internal/            # Private application packages
│   ├── cli/            # Command-line flags
│   ├── config/         # Configuration management
│   ├── detect/         # Claude CLI auto-detection
│   ├── events/         # Event parsing
│   ├── output/         # Display formatting
│   └── runner/         # Process execution
├── make.py            # Cross-platform build automation (psutil, uv-managed)
└── CONTRIBUTING.md    # Development guide
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed package descriptions.

## Installation

### Download Pre-built Binary

Download the appropriate binary for your platform from the releases page:

| Platform | Binary |
|----------|--------|
| Windows (64-bit) | `claude-print-windows-amd64.exe` |
| macOS (Intel) | `claude-print-darwin-amd64` |
| macOS (Apple Silicon) | `claude-print-darwin-arm64` |
| Linux (64-bit) | `claude-print-linux-amd64` |

### Add to PATH

#### Windows

1. Move `claude-print-windows-amd64.exe` to a directory (e.g., `C:\Tools`)
2. Rename to `claude-print.exe` for convenience
3. Add the directory to your PATH environment variable

#### macOS/Linux

```bash
# Move to a directory in your PATH
sudo mv claude-print-darwin-arm64 /usr/local/bin/claude-print

# Make it executable
sudo chmod +x /usr/local/bin/claude-print
```

### Build from Source

Requires Go 1.21+ and Python 3.12+ (or [uv](https://github.com/astral-sh/uv)).

```bash
# Clone the repository
git clone https://github.com/peakflames/claude-print.git
cd claude-print

# Using uv (recommended)
uv run make.py build

# Using Python directly
python make.py build

# Or build directly with Go
go build -o claude-print ./cmd/claude-print

# Build for all platforms
uv run make.py build-all

# Process management (headless execution)
uv run make.py start "your prompt"  # Build and start in background
uv run make.py status               # Check if running
uv run make.py log                  # View output logs
uv run make.py stop                 # Stop background process
```

For detailed development instructions, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Usage

```
claude-print [PROXY-FLAGS] <prompt> [CLAUDE-FLAGS]
```

**IMPORTANT:** The prompt must come BEFORE any Claude CLI flags that take values (like `--permission-mode plan` or `--max-turns 5`). This ensures those flags correctly receive their arguments.

### Running from Source

After building, run directly from the project directory:

```bash
# Windows (Git Bash, PowerShell, or CMD)
./claude-print.exe "What is the capital of France?"

# macOS/Linux
./claude-print "What is the capital of France?"
```

Or use the build script for common workflows:
```bash
uv run make.py run                      # Build + run test prompt (foreground)
uv run make.py start "your prompt"      # Build + run in background
```

### Running from PATH

If you've installed to GOPATH/bin or moved the binary to your PATH:

```bash
claude-print 'your prompt here'
```

### Examples

```bash
# Simple prompt
claude-print 'What is the capital of France?'

# Multi-word prompts
claude-print 'Explain how binary search works'

# With proxy flags (can come before prompt)
claude-print --verbose 'List files in this directory'
claude-print --quiet 'Generate a UUID'

# With Claude CLI flags (prompt MUST come first)
claude-print 'Design a feature' --permission-mode plan
claude-print 'Fix the bug' --allowedTools 'Read,Edit,Bash'
claude-print 'Quick task' --max-turns 5
claude-print 'Refactor everything' --dangerously-skip-permissions

# Continuing a session (no prompt required)
claude-print --continue
```

### Example Scripts

The `examples/` directory contains real-world usage patterns:

- **[plan_and_build](examples/plan_and_build/)** — Two-phase headless workflow that uses Opus to create a detailed plan (with restricted permissions), then Sonnet executes it. Demonstrates autonomous, non-interactive operation with permission modes.

## Command-line Flags

### Proxy Flags (consumed by claude-print)

| Flag | Description |
|------|-------------|
| `-v`, `--version` | Print version and exit |
| `-h`, `--help` | Show help |
| `--verbose` | Enable detailed output (also passed to Claude) |
| `--quiet` | Enable minimal output (only errors and final result) |
| `--no-color` | Disable colored output |
| `--config` | Path to config file (default: `~/.claude-print-config.json`) |
| `--debug-log` | Log raw JSON stream to directory |

### Claude CLI Flags (passed through)

All other flags are passed directly to Claude CLI. Common examples:

| Flag | Description |
|------|-------------|
| `--permission-mode <mode>` | Set permission mode (`plan`, `default`, etc.) |
| `--allowedTools <tools>` | Restrict allowed tools |
| `--dangerously-skip-permissions` | Skip all permission checks |
| `--continue` | Continue previous session |
| `--resume <id>` | Resume specific session |
| `--max-turns <n>` | Limit conversation turns |

**Note:** These flags must come AFTER the prompt (e.g., `claude-print "my prompt" --permission-mode plan`).

## Configuration

claude-print stores its configuration in `~/.claude-print-config.json`. The file is created automatically on first run.

### Example Configuration

```json
{
  "claudePath": "/usr/local/bin/claude",
  "defaultVerbosity": "normal",
  "colorEnabled": true,
  "emojiEnabled": true
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `claudePath` | string | (auto-detected) | Path to Claude CLI executable |
| `defaultVerbosity` | string | `"normal"` | Default verbosity level: `"quiet"`, `"normal"`, or `"verbose"` |
| `colorEnabled` | boolean | `true` | Enable colored output |
| `emojiEnabled` | boolean | `true` | Enable emoji indicators |

## Output Modes

### Normal Mode (default)

Shows real-time progress with tool indicators:

```
Starting Claude...
Reading file main.go
Running command: go build
Writing file output.txt
Hello! I've completed the task.

Session complete: 3 turns, 5.2s total (4.1s API), $0.02
```

### Verbose Mode (`--verbose`)

Shows detailed information including tool parameters, results, and token usage:

```
Starting Claude...
[Session] ID: abc123, Model: claude-sonnet-4-20250514
Reading file main.go
  path: /home/user/project/main.go
  result: (178 lines)
Running command: go build
  command: go build -o app .
  result: exit code 0
...
[Tokens] Input: 1,234 | Output: 567 | Cache Read: 890
```

### Quiet Mode (`--quiet`)

Minimal output for scripts - only shows errors and final result:

```
Starting...
Hello! I've completed the task.
Done
```

## Example Output

<!-- TODO: Add screenshot showing claude-print in action -->

*Screenshot placeholder: Example showing claude-print running with colored output and progress indicators*

## Requirements

- Claude CLI must be installed and accessible in your PATH
- Get Claude CLI from: https://docs.anthropic.com/en/docs/claude-cli

## License

MIT
