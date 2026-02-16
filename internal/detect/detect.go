package detect

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const claudeInstallURL = "https://docs.anthropic.com/en/docs/claude-code/getting-started"

// DetectClaudePath attempts to automatically find the Claude CLI executable.
// On Windows, it uses 'where claude' to find the executable.
// On Unix (Linux/macOS), it uses 'which claude' to find the executable.
// Returns the path to the Claude CLI or an error if not found.
func DetectClaudePath() (string, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("where", "claude")
	default:
		// Linux, macOS, and other Unix-like systems
		cmd = exec.Command("which", "claude")
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Claude CLI not found. Please install it from %s", claudeInstallURL)
	}

	// Trim whitespace and take the first line (in case of multiple matches on Windows)
	path := strings.TrimSpace(string(output))
	lines := strings.Split(path, "\n")
	if len(lines) > 0 {
		path = strings.TrimSpace(lines[0])
	}

	if path == "" {
		return "", fmt.Errorf("Claude CLI not found. Please install it from %s", claudeInstallURL)
	}

	return path, nil
}
