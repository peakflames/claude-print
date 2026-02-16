#!/usr/bin/env python3
# /// script
# requires-python = ">=3.12"
# dependencies = ["psutil"]
# ///
"""Build automation for claude-print."""

import argparse
import os
import platform
import shutil
import subprocess
import sys
import time
from datetime import datetime
from pathlib import Path

# Fix Unicode output on Windows
if sys.platform == "win32":
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")

import psutil

BINARY_NAME = "claude-print"
DIST_DIR = Path("dist")
MAIN_PATH = "./cmd/claude-print"
PID_FILE = Path(".claude-print.pid")
LOG_FILE = Path(".claude-print.log")

IS_WINDOWS = platform.system() == "Windows"

PLATFORMS = [
    ("windows", "amd64", ".exe"),
    ("darwin", "amd64", ""),
    ("darwin", "arm64", ""),
    ("linux", "amd64", ""),
]


class Colors:
    CYAN = "\033[36m"
    GREEN = "\033[32m"
    RED = "\033[31m"
    RESET = "\033[0m"


def print_colored(msg: str, color: str = "") -> None:
    """Print with optional color."""
    if sys.stdout.isatty():
        print(f"{color}{msg}{Colors.RESET}")
    else:
        print(msg)


def run_cmd(cmd: list[str], env: dict | None = None) -> None:
    """Run command with proper environment handling."""
    full_env = {**os.environ, **(env or {})}
    result = subprocess.run(cmd, env=full_env)
    if result.returncode != 0:
        sys.exit(result.returncode)


def is_process_running(pid: int) -> bool:
    """Check if process is running."""
    try:
        process = psutil.Process(pid)
        return process.is_running() and process.status() != psutil.STATUS_ZOMBIE
    except (psutil.NoSuchProcess, psutil.AccessDenied):
        return False


def read_pid_file() -> int | None:
    """Read PID from file if it exists and is valid."""
    if not PID_FILE.exists():
        return None
    try:
        pid = int(PID_FILE.read_text().strip())
        if is_process_running(pid):
            return pid
        return None
    except (ValueError, OSError):
        return None


def write_pid_file(pid: int) -> None:
    """Write PID to file."""
    PID_FILE.write_text(str(pid))


def remove_pid_file() -> None:
    """Remove PID file."""
    PID_FILE.unlink(missing_ok=True)


def kill_process(pid: int) -> bool:
    """Kill process by PID."""
    try:
        process = psutil.Process(pid)
        process.terminate()
        try:
            process.wait(timeout=10)
            return True
        except psutil.TimeoutExpired:
            process.kill()
            process.wait(timeout=5)
            return True
    except (psutil.NoSuchProcess, psutil.AccessDenied):
        return False


def cmd_help() -> None:
    print("claude-print - Build commands:")
    print()
    commands = [
        ("build", "Build for current platform"),
        ("build-all", "Build for all platforms"),
        ("install", "Install to GOPATH/bin"),
        ("test", "Run tests"),
        ("fmt", "Format Go code"),
        ("vet", "Run go vet"),
        ("clean", "Remove build artifacts"),
        ("run", "Build and run with example prompt (foreground)"),
        ("start", "Build and start in background with prompt"),
        ("stop", "Stop background process"),
        ("status", "Check if running"),
        ("log", "Display recent logs"),
    ]
    for name, desc in commands:
        print_colored(f"  {name:15} {desc}", Colors.CYAN)
    print()
    print("Examples:")
    print("  uv run make.py build")
    print("  uv run make.py start 'What is 2+2?'")
    print("  uv run make.py log")
    print("  uv run make.py stop")


def get_binary_path() -> str:
    """Get platform-appropriate binary path."""
    return f"{BINARY_NAME}.exe" if IS_WINDOWS else BINARY_NAME


def cmd_build() -> None:
    binary = get_binary_path()
    print_colored(f"Building {binary}...", Colors.GREEN)
    run_cmd(["go", "build", "-o", binary, MAIN_PATH])
    print_colored(f"Build complete: ./{binary}", Colors.GREEN)


def cmd_build_all() -> None:
    print_colored(f"Building {BINARY_NAME} for all platforms...", Colors.GREEN)
    shutil.rmtree(DIST_DIR, ignore_errors=True)
    DIST_DIR.mkdir(exist_ok=True)

    for goos, goarch, ext in PLATFORMS:
        output = DIST_DIR / f"{BINARY_NAME}-{goos}-{goarch}{ext}"
        print(f"  Building {goos}/{goarch}...")
        run_cmd(
            ["go", "build", "-o", str(output), MAIN_PATH],
            env={"CGO_ENABLED": "0", "GOOS": goos, "GOARCH": goarch},
        )

    print_colored(f"Build complete! Binaries in {DIST_DIR}/", Colors.GREEN)


def cmd_install() -> None:
    print_colored(f"Installing {BINARY_NAME} to GOPATH/bin...", Colors.GREEN)
    run_cmd(["go", "install", MAIN_PATH])
    print_colored("Install complete", Colors.GREEN)


def cmd_test() -> None:
    print_colored("Running tests...", Colors.GREEN)
    run_cmd(["go", "test", "-v", "./..."])


def cmd_fmt() -> None:
    print_colored("Formatting code...", Colors.GREEN)
    run_cmd(["go", "fmt", "./..."])


def cmd_vet() -> None:
    print_colored("Running go vet...", Colors.GREEN)
    run_cmd(["go", "vet", "./..."])


def cmd_clean() -> None:
    print_colored("Cleaning build artifacts...", Colors.GREEN)
    run_cmd(["go", "clean"])
    # Clean both possible binary names
    Path(BINARY_NAME).unlink(missing_ok=True)
    Path(f"{BINARY_NAME}.exe").unlink(missing_ok=True)
    shutil.rmtree(DIST_DIR, ignore_errors=True)
    # Clean process management files
    PID_FILE.unlink(missing_ok=True)
    LOG_FILE.unlink(missing_ok=True)
    print_colored("Clean complete", Colors.GREEN)


def cmd_run() -> None:
    cmd_build()
    binary = get_binary_path()
    print_colored(f"Running {binary}...", Colors.GREEN)
    run_cmd([f"./{binary}", "What is 2+2?"])


def cmd_start(prompt: str, extra_args: list[str] | None = None) -> None:
    """Build and start claude-print in background."""
    # Check if already running
    existing_pid = read_pid_file()
    if existing_pid:
        print_colored(f"claude-print is already running (PID: {existing_pid})", Colors.CYAN)
        print_colored("Use 'uv run make.py stop' to stop it first", Colors.CYAN)
        return

    # Build first
    cmd_build()

    print_colored("Starting claude-print in background...", Colors.GREEN)

    # Prepare environment without CLAUDECODE
    env = os.environ.copy()
    env.pop("CLAUDECODE", None)

    # Prepare log file with timestamp
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    with LOG_FILE.open("w") as f:
        f.write(f"=== claude-print started at {timestamp} ===\n\n")

    # Determine binary path
    binary_path = f"./{get_binary_path()}"

    # Build command with extra args
    cmd = [binary_path]
    if extra_args:
        cmd.extend(extra_args)
    cmd.append(prompt)

    # Open log file for process output
    log_f = LOG_FILE.open("a")

    # Start process
    if IS_WINDOWS:
        # Windows: use CREATE_NEW_PROCESS_GROUP to prevent Ctrl+C propagation
        # Do NOT use DETACHED_PROCESS as it creates a visible console window
        process = subprocess.Popen(
            cmd,
            env=env,
            stdout=log_f,
            stderr=subprocess.STDOUT,
            creationflags=subprocess.CREATE_NEW_PROCESS_GROUP,
        )
    else:
        # Unix: use start_new_session to detach
        process = subprocess.Popen(
            cmd,
            env=env,
            stdout=log_f,
            stderr=subprocess.STDOUT,
            start_new_session=True,
        )

    # Save PID
    write_pid_file(process.pid)

    print_colored(f"claude-print started (PID: {process.pid})", Colors.GREEN)
    print_colored(f"Log output: {LOG_FILE}", Colors.CYAN)

    # Wait a few seconds and check if still running
    time.sleep(2)

    if is_process_running(process.pid):
        print_colored("✓ claude-print is running", Colors.GREEN)
    else:
        print_colored("✗ claude-print failed to start. Check log file:", Colors.RED)
        print_colored("  uv run make.py log", Colors.CYAN)


def cmd_stop() -> None:
    """Stop background claude-print process."""
    pid = read_pid_file()
    if not pid:
        print_colored("No running claude-print process found", Colors.CYAN)
        return

    print_colored(f"Stopping claude-print (PID: {pid})...", Colors.GREEN)

    if kill_process(pid):
        remove_pid_file()
        print_colored("✓ claude-print stopped", Colors.GREEN)
    else:
        print_colored("Process already stopped", Colors.CYAN)
        remove_pid_file()


def cmd_status() -> None:
    """Check status of claude-print process."""
    pid = read_pid_file()
    if pid:
        print_colored(f"✓ claude-print is running (PID: {pid})", Colors.GREEN)
        print_colored(f"  Logs: {LOG_FILE}", Colors.CYAN)
    else:
        print_colored("claude-print is not running", Colors.CYAN)


def cmd_log() -> None:
    """Display recent logs from claude-print."""
    if not LOG_FILE.exists():
        print_colored("No log file found", Colors.CYAN)
        print_colored("Start claude-print first: uv run make.py start 'prompt'", Colors.CYAN)
        return

    print_colored(f"=== Logs from {LOG_FILE} ===", Colors.CYAN)
    print()
    print(LOG_FILE.read_text(encoding="utf-8", errors="replace"))


def main() -> None:
    # Manual argument parsing to avoid argparse eating flags meant for claude-print
    argv = sys.argv[1:]

    if not argv or argv[0] in ("-h", "--help"):
        cmd_help()
        return

    command = argv[0]
    extra_args = argv[1:]

    # For start command, separate extra args from the prompt (last arg is prompt)
    prompt = None
    start_extra_args = []
    if command == "start" and extra_args:
        # Last argument is the prompt, everything else is extra args
        prompt = extra_args[-1] if extra_args else None
        start_extra_args = extra_args[:-1] if len(extra_args) > 1 else []

    commands = {
        "help": cmd_help,
        "build": cmd_build,
        "build-all": cmd_build_all,
        "install": cmd_install,
        "test": cmd_test,
        "fmt": cmd_fmt,
        "vet": cmd_vet,
        "clean": cmd_clean,
        "run": cmd_run,
        "start": lambda: cmd_start(prompt, start_extra_args),
        "stop": cmd_stop,
        "status": cmd_status,
        "log": cmd_log,
    }

    if command not in commands:
        print_colored(f"Error: Unknown command '{command}'", Colors.RED)
        print_colored("Run 'uv run make.py help' for available commands", Colors.CYAN)
        sys.exit(1)

    # Validate start requires prompt
    if command == "start" and not prompt:
        print_colored("Error: start command requires a prompt", Colors.RED)
        print_colored('Usage: uv run make.py start [flags] "your prompt here"', Colors.CYAN)
        sys.exit(1)

    commands[command]()


if __name__ == "__main__":
    main()
