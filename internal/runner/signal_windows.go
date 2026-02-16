//go:build windows

package runner

import (
	"os"
)

// interruptProcess on Windows terminates the process as Windows doesn't support SIGINT.
// Claude CLI on Windows should handle process termination gracefully.
func interruptProcess(process *os.Process) error {
	return process.Kill()
}

// terminateProcess on Windows terminates the process.
func terminateProcess(process *os.Process) error {
	return process.Kill()
}
