package output

import (
	"fmt"
	"io"
	"os"
)

// ANSI escape codes for colors
const (
	colorReset  = "\033[0m"
	colorBlue   = "\033[34m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
)

// Formatter handles colored and emoji-enhanced output.
// When ColorEnabled is false, output is plain text without ANSI codes.
// When EmojiEnabled is false, emoji prefixes are omitted.
type Formatter struct {
	ColorEnabled bool
	EmojiEnabled bool
	Writer       io.Writer
}

// NewFormatter creates a new Formatter with the specified settings.
// If writer is nil, it defaults to os.Stdout.
func NewFormatter(colorEnabled, emojiEnabled bool, writer io.Writer) *Formatter {
	if writer == nil {
		writer = os.Stdout
	}
	return &Formatter{
		ColorEnabled: colorEnabled,
		EmojiEnabled: emojiEnabled,
		Writer:       writer,
	}
}

// colorize wraps text with ANSI color codes if colors are enabled.
func (f *Formatter) colorize(text, color string) string {
	if !f.ColorEnabled {
		return text
	}
	return color + text + colorReset
}

// Info outputs an informational message in blue.
func (f *Formatter) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	colored := f.colorize(msg, colorBlue)
	fmt.Fprintln(f.Writer, colored)
}

// Success outputs a success message in green.
func (f *Formatter) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	colored := f.colorize(msg, colorGreen)
	fmt.Fprintln(f.Writer, colored)
}

// Error outputs an error message in red.
func (f *Formatter) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	colored := f.colorize(msg, colorRed)
	fmt.Fprintln(f.Writer, colored)
}

// Warning outputs a warning message in yellow.
func (f *Formatter) Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	colored := f.colorize(msg, colorYellow)
	fmt.Fprintln(f.Writer, colored)
}

// Plain outputs text without any color formatting.
func (f *Formatter) Plain(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(f.Writer, msg)
}

// PlainNoNewline outputs text without a trailing newline.
func (f *Formatter) PlainNoNewline(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprint(f.Writer, msg)
}

// ToolCall outputs a tool call with only the bullet colored green and rest plain.
// Format: "● ToolName(params)" where only ● is green.
func (f *Formatter) ToolCall(bullet, text string) {
	if f.ColorEnabled {
		fmt.Fprintf(f.Writer, "%s%s%s %s\n", colorGreen, bullet, colorReset, text)
	} else {
		fmt.Fprintf(f.Writer, "%s %s\n", bullet, text)
	}
}

// InfoWithEmoji outputs an informational message with an optional emoji prefix.
func (f *Formatter) InfoWithEmoji(emoji, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if f.EmojiEnabled && emoji != "" {
		msg = emoji + " " + msg
	}
	colored := f.colorize(msg, colorBlue)
	fmt.Fprintln(f.Writer, colored)
}

// SuccessWithEmoji outputs a success message with an optional emoji prefix.
func (f *Formatter) SuccessWithEmoji(emoji, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if f.EmojiEnabled && emoji != "" {
		msg = emoji + " " + msg
	}
	colored := f.colorize(msg, colorGreen)
	fmt.Fprintln(f.Writer, colored)
}

// ErrorWithEmoji outputs an error message with an optional emoji prefix.
func (f *Formatter) ErrorWithEmoji(emoji, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if f.EmojiEnabled && emoji != "" {
		msg = emoji + " " + msg
	}
	colored := f.colorize(msg, colorRed)
	fmt.Fprintln(f.Writer, colored)
}

// WarningWithEmoji outputs a warning message with an optional emoji prefix.
func (f *Formatter) WarningWithEmoji(emoji, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if f.EmojiEnabled && emoji != "" {
		msg = emoji + " " + msg
	}
	colored := f.colorize(msg, colorYellow)
	fmt.Fprintln(f.Writer, colored)
}
