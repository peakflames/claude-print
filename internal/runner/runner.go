package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/peakflames/claude-print/internal/cli"
)

// RunOptions configures how the Claude CLI process is invoked.
type RunOptions struct {
	ClaudePath      string
	Prompt          string
	PassthroughArgs []string // Args to pass through to Claude unchanged
}

// ClaudeProcess represents a running Claude CLI process.
type ClaudeProcess struct {
	Cmd    *exec.Cmd
	Stdout io.ReadCloser
	stderr *bytes.Buffer
}

// RunClaude spawns the Claude CLI process with the given options and returns
// the process handle and stdout reader for streaming output.
func RunClaude(opts RunOptions) (*ClaudeProcess, error) {
	if opts.ClaudePath == "" {
		return nil, fmt.Errorf("Claude CLI path is empty")
	}

	// Prompt is required unless continuing/resuming a session
	if opts.Prompt == "" && !cli.ContainsSessionFlag(opts.PassthroughArgs) {
		return nil, fmt.Errorf("prompt is empty")
	}

	args := buildArgs(opts)
	cmd := exec.Command(opts.ClaudePath, args...)

	// Inherit environment variables from parent process
	cmd.Env = os.Environ()

	// Capture stdout as a pipe for streaming
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Capture stderr to a buffer for error context
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start Claude CLI: %w", err)
	}

	return &ClaudeProcess{
		Cmd:    cmd,
		Stdout: stdout,
		stderr: &stderrBuf,
	}, nil
}

// Wait waits for the Claude CLI process to complete and returns any error.
func (p *ClaudeProcess) Wait() error {
	return p.Cmd.Wait()
}

// ExitCode returns the exit code of the process after it has completed.
// Returns -1 if the process hasn't exited yet or if unable to determine the exit code.
func (p *ClaudeProcess) ExitCode() int {
	if p.Cmd.ProcessState == nil {
		return -1
	}
	return p.Cmd.ProcessState.ExitCode()
}

// Kill terminates the Claude CLI process.
func (p *ClaudeProcess) Kill() error {
	if p.Cmd.Process == nil {
		return nil
	}
	return p.Cmd.Process.Kill()
}

// Interrupt sends an interrupt signal to the Claude CLI process for graceful shutdown.
// On Unix systems, this sends SIGINT. On Windows, this sends CTRL_BREAK_EVENT.
func (p *ClaudeProcess) Interrupt() error {
	if p.Cmd.Process == nil {
		return nil
	}
	return interruptProcess(p.Cmd.Process)
}

// Terminate sends a termination signal to the Claude CLI process.
// On Unix systems, this sends SIGTERM. On Windows, this terminates the process.
func (p *ClaudeProcess) Terminate() error {
	if p.Cmd.Process == nil {
		return nil
	}
	return terminateProcess(p.Cmd.Process)
}

// Stderr returns the stderr output captured from the Claude CLI process.
// This should be called after the process has completed.
func (p *ClaudeProcess) Stderr() string {
	if p.stderr == nil {
		return ""
	}
	return p.stderr.String()
}

// buildArgs constructs the Claude CLI arguments from RunOptions.
// Required flags for streaming JSON are prepended, then passthrough args, then prompt.
func buildArgs(opts RunOptions) []string {
	// Required flags for claude-print to work correctly
	args := []string{
		"--include-partial-messages",
		"--verbose",
		"--output-format=stream-json",
	}

	// Append all passthrough args from user
	args = append(args, opts.PassthroughArgs...)

	// Prompt comes last via -p flag
	if opts.Prompt != "" {
		args = append(args, "-p", opts.Prompt)
	}

	return args
}
