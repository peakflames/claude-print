package cli

import (
	"os"
	"testing"
)

// saveAndSetArgs replaces os.Args for the duration of the test and restores it
// on cleanup.
func saveAndSetArgs(t *testing.T, args []string) {
	t.Helper()
	orig := os.Args
	t.Cleanup(func() { os.Args = orig })
	os.Args = args
}

func TestParseFlags_PositionalPrompt(t *testing.T) {
	saveAndSetArgs(t, []string{"claude-print", "hello world"})

	flags, err := ParseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flags.Prompt != "hello world" {
		t.Errorf("expected Prompt %q, got %q", "hello world", flags.Prompt)
	}
}

func TestParseFlags_StreamJSON(t *testing.T) {
	saveAndSetArgs(t, []string{"claude-print", "--stream-json", "my prompt"})

	flags, err := ParseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !flags.StreamJSON {
		t.Error("expected StreamJSON to be true")
	}
	if flags.Prompt != "my prompt" {
		t.Errorf("expected Prompt %q, got %q", "my prompt", flags.Prompt)
	}
}

func TestParseFlags_StreamJSON_NoPrompt(t *testing.T) {
	saveAndSetArgs(t, []string{"claude-print", "--stream-json"})

	flags, err := ParseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !flags.StreamJSON {
		t.Error("expected StreamJSON to be true")
	}
}

func TestParseFlags_StdinPrompt(t *testing.T) {
	// Set up a pipe to simulate piped stdin (non-TTY).
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	// Write the prompt and close the write end.
	if _, err := w.WriteString("prompt from stdin\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	w.Close()

	// Swap os.Stdin for the read end of the pipe.
	origStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = origStdin
		r.Close()
	})

	saveAndSetArgs(t, []string{"claude-print"})

	flags, err := ParseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Trailing newline should be stripped.
	if flags.Prompt != "prompt from stdin" {
		t.Errorf("expected Prompt %q, got %q", "prompt from stdin", flags.Prompt)
	}
}

func TestParseFlags_ProtectedFlags(t *testing.T) {
	tests := []struct {
		flag string
	}{
		{"-p"},
		{"--print"},
		{"--output-format"},
		{"--include-partial-messages"},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			saveAndSetArgs(t, []string{"claude-print", tt.flag})
			_, err := ParseFlags()
			if err == nil {
				t.Errorf("expected error for protected flag %q, got nil", tt.flag)
			}
		})
	}
}

func TestParseFlags_PassthroughArgs(t *testing.T) {
	saveAndSetArgs(t, []string{"claude-print", "my prompt", "--continue", "--max-turns", "5"})

	flags, err := ParseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flags.Prompt != "my prompt" {
		t.Errorf("expected Prompt %q, got %q", "my prompt", flags.Prompt)
	}
	// --continue, --max-turns, and 5 should all be in passthrough
	if len(flags.PassthroughArgs) < 1 {
		t.Errorf("expected passthrough args, got none")
	}
}
