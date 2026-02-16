package cli

import (
	"fmt"
	"os"
	"strings"
)

// protectedFlags are flags that claude-print uses internally and cannot be
// passed through to Claude CLI. The map value explains why each is blocked.
var protectedFlags = map[string]string{
	"-p":                         "prompt is passed as the first positional argument",
	"--print":                    "prompt is passed as the first positional argument",
	"--output-format":            "claude-print requires stream-json format",
	"--include-partial-messages": "claude-print requires partial messages",
}

// Flags holds the parsed command-line options.
type Flags struct {
	// Proxy-specific flags
	Version    bool
	Verbose    bool
	Quiet      bool
	NoColor    bool
	NoEmoji    bool
	ConfigPath string
	DebugLog   string // --debug-log <dir> (log raw JSON to directory)
	ShowHelp   bool

	// Positional and passthrough
	Prompt          string   // First positional argument (the prompt for Claude)
	PassthroughArgs []string // All other args passed to Claude unchanged
}

// ParseFlags parses command-line arguments and returns the parsed Flags.
// Returns an error if a protected flag is used.
func ParseFlags() (Flags, error) {
	f := Flags{}
	args := os.Args[1:]

	// Track which args to pass through
	var passthrough []string
	skipNext := false

	for i := 0; i < len(args); i++ {
		if skipNext {
			skipNext = false
			continue
		}

		arg := args[i]

		// Check for protected flags (with or without values)
		baseFlagName := extractFlagName(arg)
		if reason, blocked := isProtectedFlag(baseFlagName); blocked {
			return Flags{}, fmt.Errorf("cannot use %s: %s", baseFlagName, reason)
		}

		// Handle proxy-specific flags
		switch arg {
		case "-v", "--version":
			f.Version = true
		case "-h", "--help":
			f.ShowHelp = true
		case "--verbose":
			// Verbose is consumed by proxy AND passed through to Claude
			f.Verbose = true
			passthrough = append(passthrough, arg)
		case "--quiet":
			f.Quiet = true
		case "--no-color":
			f.NoColor = true
		case "--no-emoji":
			f.NoEmoji = true
		case "--config":
			if i+1 < len(args) {
				f.ConfigPath = args[i+1]
				skipNext = true
			}
		case "--debug-log":
			if i+1 < len(args) {
				f.DebugLog = args[i+1]
				skipNext = true
			}
		default:
			// Handle --config=value and --debug-log=value forms
			if strings.HasPrefix(arg, "--config=") {
				f.ConfigPath = strings.TrimPrefix(arg, "--config=")
			} else if strings.HasPrefix(arg, "--debug-log=") {
				f.DebugLog = strings.TrimPrefix(arg, "--debug-log=")
			} else if strings.HasPrefix(arg, "-") {
				// Any other flag is passed through to Claude
				passthrough = append(passthrough, arg)

				// If it looks like a flag that expects a value (--flag value form),
				// we need to check if the next arg is its value
				// This handles --continue (no value), --resume <id> (has value), etc.
				// For simplicity, we pass both and let Claude parse them
				// Flags with = already contain their value
			} else if f.Prompt == "" {
				// First non-flag arg is the prompt
				f.Prompt = arg
			} else {
				// Additional positional args are passed through
				passthrough = append(passthrough, arg)
			}
		}
	}

	f.PassthroughArgs = passthrough
	return f, nil
}

// extractFlagName extracts the flag name from an argument, handling --flag=value forms.
func extractFlagName(arg string) string {
	if idx := strings.Index(arg, "="); idx != -1 {
		return arg[:idx]
	}
	return arg
}

// isProtectedFlag checks if a flag is protected and returns the reason if so.
func isProtectedFlag(flag string) (string, bool) {
	reason, blocked := protectedFlags[flag]
	return reason, blocked
}

// ContainsSessionFlag checks if passthrough args contain --continue or --resume.
// This is used to determine if a prompt is required.
func ContainsSessionFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--continue" || arg == "--resume" || strings.HasPrefix(arg, "--resume=") {
			return true
		}
	}
	return false
}
