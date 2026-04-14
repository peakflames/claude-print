# claude-print

![Claude Print](claude-print-preview.gif)

A cross-platform CLI wrapper for Claude CLI that provides real-time progress feedback during headless execution.

## Features

- Real-time progress updates showing what Claude is doing
- Colored, formatted output
- Three verbosity levels: quiet, normal, and verbose
- Structured JSON output mode (`--stream-json`) for programmatic consumption
- Automatic TTY detection for script-friendly output
- Cross-platform support (Windows, macOS, Linux)
- Graceful shutdown on Ctrl+C

## Installation

### Quick Install

Downloads the latest release binary and places it in `~/.local/bin` alongside your Claude CLI installation.

**macOS (Apple Silicon):**
```bash
curl -fsSL https://github.com/peakflames/claude-print/releases/latest/download/claude-print-darwin-arm64 -o ~/.local/bin/claude-print && chmod +x ~/.local/bin/claude-print
```

**macOS (Intel):**
```bash
curl -fsSL https://github.com/peakflames/claude-print/releases/latest/download/claude-print-darwin-amd64 -o ~/.local/bin/claude-print && chmod +x ~/.local/bin/claude-print
```

**Linux:**
```bash
curl -fsSL https://github.com/peakflames/claude-print/releases/latest/download/claude-print-linux-amd64 -o ~/.local/bin/claude-print && chmod +x ~/.local/bin/claude-print
```

**Windows (PowerShell):**
```powershell
Invoke-WebRequest -Uri "https://github.com/peakflames/claude-print/releases/latest/download/claude-print-windows-amd64.exe" -OutFile "$env:USERPROFILE\.local\bin\claude-print.exe"
```

### Build from Source

Requires Go 1.21+ and optionally [uv](https://github.com/astral-sh/uv) for build automation.

```bash
git clone https://github.com/peakflames/claude-print.git
cd claude-print

# Build with Go directly
go build -o claude-print ./cmd/claude-print

# Or use the build script
uv run make.py build
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed development instructions.

## Usage

```
claude-print [OPTIONS] <prompt> [CLAUDE-FLAGS]
```

**Important:** The prompt must come BEFORE any Claude CLI flags that take values (like `--permission-mode plan`). This ensures those flags correctly receive their arguments.

### Basic Examples

```bash
# Simple prompts
claude-print "What is the capital of France?"
claude-print "Explain how binary search works"

# With verbosity control
claude-print --verbose "List files in this directory"
claude-print --quiet "Generate a UUID"

# With Claude CLI flags (prompt MUST come first)
claude-print "Design a feature" --permission-mode plan
claude-print "Fix the bug" --allowedTools "Read,Edit,Bash"
claude-print "Quick task" --max-turns 5

# Continue previous session
claude-print --continue

# Read prompt from stdin
echo "What is 2+2?" | claude-print
cat prompt.txt | claude-print --quiet

# Stream structured JSON events (display goes to stderr, JSON to stdout)
claude-print --stream-json "Summarize this" | jq .
echo "large prompt" | claude-print --stream-json | post-processing-tool
```

### Headless Automation

The `examples/` directory contains real-world automation patterns:

- **[plan_and_build](examples/plan_and_build/)** - Two-phase workflow: Opus creates a plan in restricted mode, then Sonnet executes it. Demonstrates autonomous, non-interactive operation.

## Command-line Options

### Proxy Flags (consumed by claude-print)

| Flag | Description |
|------|-------------|
| `-v`, `--version` | Print version and exit |
| `-h`, `--help` | Show help |
| `--verbose` | Enable detailed output |
| `--quiet` | Minimal output (errors and results only) |
| `--no-color` | Disable colored output |
| `--stream-json` | Write structured JSON events to stdout; display goes to stderr |
| `--config` | Path to config file (default: `~/.claude-print-config.json`) |
| `--debug-log` | Log raw JSON stream to directory |

### Claude CLI Flags (passed through)

All other flags are passed directly to Claude CLI:

| Flag | Description |
|------|-------------|
| `--permission-mode <mode>` | Set permission mode (`plan`, `default`, etc.) |
| `--allowedTools <tools>` | Restrict allowed tools |
| `--dangerously-skip-permissions` | Skip all permission checks |
| `--continue` | Continue previous session |
| `--resume <id>` | Resume specific session |
| `--max-turns <n>` | Limit conversation turns |

## Configuration

Configuration is stored in `~/.claude-print-config.json` and created automatically on first run.

**Example:**
```json
{
  "claudePath": "/usr/local/bin/claude",
  "defaultVerbosity": "normal",
  "colorEnabled": true
}
```

**Options:**

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `claudePath` | string | (auto-detected) | Path to Claude CLI executable |
| `defaultVerbosity` | string | `"normal"` | Default verbosity: `"quiet"`, `"normal"`, or `"verbose"` |
| `colorEnabled` | boolean | `true` | Enable colored output |

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

Shows detailed information including tool parameters and token usage:

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

Minimal output for scripts - only errors and final result:

```
Starting...
Hello! I've completed the task.
Done
```

### Stream JSON Mode (`--stream-json`)

Routes visual progress output to **stderr** and emits newline-delimited JSON
events to **stdout**. Ideal for piping into downstream tools while still seeing
human-readable progress.

```bash
# Visual progress on stderr, structured JSON on stdout
claude-print --stream-json "Refactor main.go" 2>progress.log | jq .

# Combine with stdin prompt for fully scriptable pipelines
echo "Summarize this file" | claude-print --stream-json 2>/dev/null | jq -r 'select(.type=="text") | .content'
```

**JSON event schema:**

| Event | Fields |
|-------|--------|
| Text chunk | `{"type":"text","content":"..."}` |
| Tool call | `{"type":"tool_call","tool":"Read","input":{...}}` |
| Tool result | `{"type":"tool_result","tool":"Read","summary":"Read 15 lines"}` |
| Session end | `{"type":"result","cost":0.002,"duration_ms":3210,"turns":2,"is_error":false}` |

Each line is a complete JSON object. Events are emitted in real-time as Claude
runs, so consumers can process them incrementally.

## Requirements

- Claude CLI must be installed and accessible in your PATH
- Get Claude CLI from: https://docs.anthropic.com/en/docs/claude-cli

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Project structure and architecture
- Build commands and workflows
- Testing and contribution guidelines

## License

MIT
