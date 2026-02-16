# Contributing to claude-print

Thank you for your interest in contributing to claude-print! This guide will help you get started with development.

## Prerequisites

Before you begin, make sure you have the following installed:

- **Go 1.21 or higher** - [Download](https://go.dev/dl/)
- **Python 3.12+** or **uv** - For build automation
  - Python: [Download](https://www.python.org/downloads/)
  - uv (recommended): [Install](https://github.com/astral-sh/uv) - can bootstrap Python automatically
- **Claude CLI** - Required for testing the proxy functionality
  - Install: `npm install -g @anthropic-ai/claude-cli`
  - Or via Homebrew: `brew install claude`
- **Git** - For version control

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/peakflames/claude-print.git
cd claude-print
```

### 2. Verify Your Environment

Check that you have the required tools:

```bash
go version          # Should be 1.21 or higher
claude --version    # Verify Claude CLI is installed
python --version    # Should be 3.12 or higher (or use uv)
python make.py help # View available build commands
```

### 3. Build the Project

Using the build script:
```bash
python make.py build
# or with uv:
uv run make.py build
```

Or using Go directly:
```bash
go build -o claude-print ./cmd/claude-print
```

### 4. Run the Binary

Test the built binary:
```bash
./claude-print "What is 2+2?"
```

Or use the build script to build and run in one step:
```bash
uv run make.py run                    # Foreground (interactive)
uv run make.py start "What is 2+2?"   # Background (headless)
uv run make.py log                    # View background output
uv run make.py stop                   # Stop background process
```

## Project Structure

```
claude-print/
├── cmd/
│   └── claude-print/        # Application entry point
│       └── main.go          # Main function and setup
├── internal/                # Private application packages
│   ├── cli/                 # Command-line flag parsing
│   ├── config/              # Configuration management
│   ├── detect/              # Claude CLI auto-detection
│   ├── events/              # Event parsing and types
│   ├── output/              # Display formatting and error handling
│   └── runner/              # Process execution and event streaming
├── docs/prd/               # Product Requirements Documents
├── go.mod                  # Go module definition
├── go.sum                  # Go dependency checksums
├── make.py                 # Cross-platform build automation (uses psutil via uv)
├── README.md               # Project overview
└── CONTRIBUTING.md         # This file
```

### Package Overview

- **cmd/claude-print**: Entry point that wires together all components
- **internal/cli**: Defines and parses command-line flags (`--verbose`, `--quiet`, etc.)
- **internal/config**: Loads/saves configuration from `~/.config/claude-print/config.json`
- **internal/detect**: Auto-detects Claude CLI installation path
- **internal/events**: Parses JSON event stream from Claude CLI
- **internal/output**: Handles display formatting, colors, emojis, and error messages
- **internal/runner**: Spawns Claude CLI process and streams events

## Development Workflow

### Making Changes

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** in the appropriate package under `internal/` or `cmd/`

3. **Format your code**:
   ```bash
   python make.py fmt
   # or: go fmt ./...
   ```

4. **Check for issues**:
   ```bash
   python make.py vet
   # or: go vet ./...
   ```

5. **Test your changes**:
   ```bash
   python make.py build
   ./claude-print "test prompt"
   ```

### Running Tests

```bash
python make.py test
# or: go test -v ./...
```

### Building for All Platforms

To create release binaries for Windows, macOS (Intel + ARM), and Linux:

```bash
python make.py build-all
# or with uv:
uv run make.py build-all
```

Binaries will be created in the `dist/` directory.

### Code Style Guidelines

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and single-purpose
- Prefer explicit error handling over panics

### Commit Messages

Use clear, descriptive commit messages:

```
feat: add support for custom config paths
fix: resolve path detection on Windows
docs: update installation instructions
refactor: simplify event parsing logic
```

Prefix conventions:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Test additions or changes
- `chore:` - Maintenance tasks

## Cross-Platform Considerations

claude-print runs on Windows, macOS, and Linux. When making changes:

### Path Handling
- Use forward slashes (`/`) in code; Go normalizes them on Windows
- Use `filepath.Join()` for path construction
- Test on Windows if modifying path-related code

### Signal Handling
- Separate implementations exist for Unix and Windows
- See `internal/runner/signal_unix.go` and `internal/runner/signal_windows.go`
- Use build tags (`//go:build`) for platform-specific code

### Testing Cross-Platform
Even if you're on one platform, consider how your changes might affect others:
- Does it assume Unix-style paths?
- Does it rely on Unix-specific signals or commands?
- Does it use shell features not available on all platforms?

### Python Build Script (make.py)

The build script uses patterns that work cross-platform:

**Dependency management**: Uses uv's inline script metadata instead of requirements.txt:
```python
# /// script
# requires-python = ">=3.12"
# dependencies = ["psutil"]
# ///
```

**Windows console behavior**: When spawning background processes:
- Use `CREATE_NEW_PROCESS_GROUP` to prevent Ctrl+C propagation
- Do NOT use `DETACHED_PROCESS` - it creates a visible console window popup

**Process management**: Use `psutil` library for cross-platform compatibility:
```python
import psutil
# Check if running
psutil.Process(pid).is_running()
# Graceful termination
process.terminate()
process.wait(timeout=10)
```

**Unicode output on Windows**: Configure encoding at startup:
```python
if sys.platform == "win32":
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
```

## Submitting Changes

### Pull Request Process

1. **Ensure your branch is up to date**:
   ```bash
   git fetch origin
   git rebase origin/main
   ```

2. **Push your branch**:
   ```bash
   git push origin feature/your-feature-name
   ```

3. **Create a Pull Request** on GitHub with:
   - Clear title describing the change
   - Description of what changed and why
   - Any related issue numbers (e.g., "Fixes #123")

4. **Respond to feedback** and make requested changes

5. **Squash commits** if requested before merging

### What We Look For

- Clear, focused changes that solve one problem
- Code that follows Go best practices
- Cross-platform compatibility
- Proper error handling
- Tests for new functionality (when applicable)

## Getting Help

- **Questions?** Open a GitHub Discussion
- **Bug reports?** Open a GitHub Issue
- **Security issues?** See SECURITY.md

## Development Tips

### Debugging Event Parsing

Use `--verbose` to see full event details:
```bash
./claude-print --verbose "your prompt"
```

### Viewing Raw JSON Events

Bypass claude-print and run Claude CLI directly:
```bash
claude --format stream "your prompt"
```

### Testing Different Verbosity Levels

```bash
# Normal mode (default)
./claude-print "test"

# Quiet mode (minimal output)
./claude-print --quiet "test"

# Verbose mode (detailed)
./claude-print --verbose "test"
```

### Configuration File Location

- **macOS/Linux**: `~/.config/claude-print/config.json`
- **Windows**: `%APPDATA%\claude-print\config.json`

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (see LICENSE file).
