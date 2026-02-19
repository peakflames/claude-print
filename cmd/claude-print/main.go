package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/peakflames/claude-print/internal/cli"
	"github.com/peakflames/claude-print/internal/config"
	"github.com/peakflames/claude-print/internal/detect"
	"github.com/peakflames/claude-print/internal/output"
	"github.com/peakflames/claude-print/internal/runner"
)

var version = "0.2.0"

func printUsage(ver string) {
	fmt.Printf("claude-print %s\n", ver)
	fmt.Println()
	fmt.Println("CLI wrapper for Claude CLI that provides real-time progress feedback")
	fmt.Println("during headless execution.")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("    claude-print [PROXY-FLAGS] <prompt> [CLAUDE-FLAGS]")
	fmt.Println()
	fmt.Println("IMPORTANT: The prompt must come BEFORE any Claude flags that take values.")
	fmt.Println("           This ensures flags like --permission-mode correctly receive their arguments.")
	fmt.Println()
	fmt.Println("PROXY FLAGS (consumed by claude-print):")
	fmt.Println("    -v, --version    Print version and exit")
	fmt.Println("    -h, --help       Show this help")
	fmt.Println("        --verbose    Enable detailed output (also passed to Claude)")
	fmt.Println("        --quiet      Enable minimal output (results only)")
	fmt.Println("        --no-color   Disable colored output")
	fmt.Println("        --no-emoji   Disable emoji in output")
	fmt.Println("        --config     Path to config file (default: ~/.claude-print-config.json)")
	fmt.Println("        --debug-log  Log raw JSON stream to directory")
	fmt.Println()
	fmt.Println("All other flags are passed through to Claude CLI unchanged.")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("    claude-print \"What is 2+2?\"")
	fmt.Println("    claude-print --verbose \"Explain this code\"")
	fmt.Println("    claude-print --quiet \"Generate a haiku\"")
	fmt.Println()
	fmt.Println("    # Pass Claude CLI flags through (prompt MUST come before flags with values):")
	fmt.Println("    claude-print \"Design a feature\" --permission-mode plan")
	fmt.Println("    claude-print \"Fix the bug\" --allowedTools \"Read,Edit,Bash\"")
	fmt.Println("    claude-print \"Refactor everything\" --dangerously-skip-permissions")
	fmt.Println("    claude-print --continue")
	fmt.Println("    claude-print \"Quick task\" --max-turns 5")
	fmt.Println()
	fmt.Println("PROTECTED FLAGS (cannot be used - required by claude-print):")
	fmt.Println("    -p, --print                  Prompt is passed as positional argument")
	fmt.Println("    --output-format              Must be stream-json")
	fmt.Println("    --include-partial-messages   Always enabled")
	fmt.Println()
	fmt.Println("CONFIG FILE:")
	fmt.Println("    ~/.claude-print-config.json")
	fmt.Println()
	fmt.Println("    Available settings:")
	fmt.Println("      claudePath        Path to Claude CLI executable (auto-detected)")
	fmt.Println("      defaultVerbosity  Default output level: normal, verbose, quiet")
	fmt.Println("      colorEnabled      Enable colored output (default: true)")
	fmt.Println("      emojiEnabled      Enable emoji in output (default: true)")
	fmt.Println()
	fmt.Println("ENVIRONMENT:")
	fmt.Println("    NO_COLOR    Set to disable colored output")
	fmt.Println()
	fmt.Println("MORE INFO:")
	fmt.Println("    https://github.com/peakflames/claude-print")
}

func main() {
	os.Exit(run())
}

func run() int {
	// Parse command-line flags first
	flags, err := cli.ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Handle version flag immediately (before any other setup)
	if flags.Version {
		fmt.Printf("claude-print %s\n", version)
		return 0
	}

	// Handle help flag
	if flags.ShowHelp {
		printUsage(version)
		return 0
	}

	// Ensure we always end with a newline
	defer fmt.Println()

	// Load config (returns default if file doesn't exist)
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return 1
	}

	// Determine color and emoji settings
	colorEnabled := output.ShouldEnableColor(flags.NoColor, cfg.ColorEnabled)
	emojiEnabled := cfg.EmojiEnabled && !flags.NoEmoji

	// Create formatter and display
	formatter := output.NewFormatter(colorEnabled, emojiEnabled, nil)

	// Determine verbosity level
	verbosity := output.VerbosityNormal
	if flags.Verbose {
		verbosity = output.VerbosityVerbose
	} else if flags.Quiet {
		verbosity = output.VerbosityQuiet
	} else if cfg.DefaultVerbosity == "verbose" {
		verbosity = output.VerbosityVerbose
	} else if cfg.DefaultVerbosity == "quiet" {
		verbosity = output.VerbosityQuiet
	}

	display := output.NewDisplay(formatter, verbosity)

	// Auto-detect Claude path if not configured
	claudePath := cfg.ClaudePath
	if claudePath == "" {
		detectedPath, err := detect.DetectClaudePath()
		if err != nil {
			formatter.ErrorWithEmoji(output.EmojiError, "%v", err)
			return 1
		}
		claudePath = detectedPath

		// Save detected path to config for future use
		cfg.ClaudePath = claudePath
		if saveErr := config.SaveConfig(cfg); saveErr != nil {
			// Non-fatal: just warn if we can't save
			formatter.Warning("Could not save config: %v", saveErr)
		}
	}

	// Validate Claude path exists
	if err := config.ValidatePath(claudePath); err != nil {
		formatter.ErrorWithEmoji(output.EmojiError, "%v", err)
		return 1
	}

	// Check if we have a prompt (not required for --continue or --resume)
	hasSessionFlag := cli.ContainsSessionFlag(flags.PassthroughArgs)
	if flags.Prompt == "" && !hasSessionFlag {
		printUsage(version)
		return 0
	}

	// Pass prompt to display for rendering
	if flags.Prompt != "" {
		display.SetUserPrompt(flags.Prompt)
		// Show start indicator with user prompt
		display.ShowStart()
	} else if hasSessionFlag {
		display.SetUserPrompt("(continuing session)")
		display.ShowStart()
	}

	// Enable debug logging if requested
	if flags.DebugLog != "" {
		if err := runner.EnableDebugLogging(flags.DebugLog); err != nil {
			formatter.Warning("Could not enable debug logging: %v", err)
		} else {
			defer runner.CloseDebugLogging()
		}
	}

	// Build run options - simple pass-through architecture
	opts := runner.RunOptions{
		ClaudePath:      claudePath,
		Prompt:          flags.Prompt,
		PassthroughArgs: flags.PassthroughArgs,
	}

	// Spawn Claude CLI process
	process, err := runner.RunClaude(opts)
	if err != nil {
		formatter.ErrorWithEmoji(output.EmojiError, "Failed to start Claude: %v", err)
		return 1
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Channel to communicate when event streaming is done
	doneChan := make(chan struct{})

	// Stream events from the process
	eventChan := runner.StreamEventsFromProcess(process)

	// Handle events in real-time (in a goroutine to allow signal handling)
	go func() {
		for event := range eventChan {
			display.HandleEvent(event)
		}
		close(doneChan)
	}()

	// Wait for either completion or signal
	var receivedSignal os.Signal
	select {
	case <-doneChan:
		// Normal completion - event streaming finished
		signal.Stop(sigChan)
	case sig := <-sigChan:
		// Received interrupt signal
		receivedSignal = sig
		signal.Stop(sigChan)

		// Send termination signal to child process
		if sig == syscall.SIGINT {
			_ = process.Interrupt()
		} else {
			_ = process.Terminate()
		}

		// Wait for event channel to drain (child process cleanup)
		<-doneChan
	}

	// Wait for process to complete
	_ = process.Wait()

	// If we received a signal, return appropriate exit code
	if receivedSignal != nil {
		// 128 + signal number is the conventional exit code for signal termination
		// SIGINT = 2, so exit code = 130
		// SIGTERM = 15, so exit code = 143
		switch receivedSignal {
		case syscall.SIGINT:
			return 130
		case syscall.SIGTERM:
			return 143
		default:
			return 128
		}
	}

	// Check for process error
	exitCode := process.ExitCode()
	if exitCode != 0 {
		stderr := process.Stderr()

		// Detect and display error
		errCtx := output.DetectExitCodeError(exitCode, stderr)
		if errCtx != nil {
			output.DisplayError(formatter, errCtx)
		}
	}

	// Return Claude CLI exit code
	return exitCode
}
