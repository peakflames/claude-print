//go:build !windows

package runner

import (
	"os"
	"syscall"
)

// interruptProcess sends SIGINT to the process for graceful interrupt.
func interruptProcess(process *os.Process) error {
	return process.Signal(syscall.SIGINT)
}

// terminateProcess sends SIGTERM to the process for graceful termination.
func terminateProcess(process *os.Process) error {
	return process.Signal(syscall.SIGTERM)
}
