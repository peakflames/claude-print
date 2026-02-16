package output

import (
	"os"
)

// IsTTY checks if the given file is a terminal (TTY).
// Returns true if the file is connected to an interactive terminal,
// false if it's being piped or redirected.
func IsTTY(f *os.File) bool {
	if f == nil {
		return false
	}
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	// Check if the file mode indicates a character device (terminal)
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// IsStdoutTTY checks if stdout is a terminal.
// This is useful for determining whether to enable colors by default.
func IsStdoutTTY() bool {
	return IsTTY(os.Stdout)
}

// ShouldEnableColor determines if colors should be enabled based on:
// 1. Explicit user flag (noColorFlag) - if true, colors are disabled
// 2. NO_COLOR environment variable - if set, colors are disabled (https://no-color.org/)
// 3. Config file setting (configColorEnabled) - user preference from config
// 4. TTY detection - if not a TTY, colors are disabled by default
//
// The priority is:
// - If noColorFlag is true, return false (user explicitly disabled)
// - If NO_COLOR env var is set, return false (respect convention)
// - If stdout is not a TTY, return false (pipe/redirect scenario)
// - Otherwise, return configColorEnabled (respect config file setting)
func ShouldEnableColor(noColorFlag bool, configColorEnabled bool) bool {
	// Explicit --no-color flag takes highest priority
	if noColorFlag {
		return false
	}

	// Respect NO_COLOR environment variable (https://no-color.org/)
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		return false
	}

	// If stdout is not a TTY (piped/redirected), disable colors
	if !IsStdoutTTY() {
		return false
	}

	// Respect config file setting
	return configColorEnabled
}
