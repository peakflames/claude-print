package runner

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/peakflames/claude-print/internal/events"
)

// debugLogFile is the file handle for debug JSON logging (nil if not enabled)
var debugLogFile *os.File

// EnableDebugLogging creates a timestamped log file in the specified directory
// and logs all raw JSON lines to it. Call CloseDebugLogging when done.
func EnableDebugLogging(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	timestamp := time.Now().Format("2006-01-02_150405")
	filename := filepath.Join(dir, "stream-"+timestamp+".jsonl")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	debugLogFile = f
	log.Printf("Debug logging to: %s", filename)
	return nil
}

// CloseDebugLogging closes the debug log file if it's open
func CloseDebugLogging() {
	if debugLogFile != nil {
		debugLogFile.Close()
		debugLogFile = nil
	}
}

// StreamEvents reads lines from the given reader and emits parsed events
// through a channel. Each line is expected to be a JSON event from Claude's
// streaming output. Malformed JSON lines are logged and skipped.
// The channel is closed when EOF is reached or an error occurs.
func StreamEvents(reader io.Reader) <-chan events.Event {
	eventChan := make(chan events.Event)

	go func() {
		defer close(eventChan)

		scanner := bufio.NewScanner(reader)
		// Increase buffer size for potentially large JSON lines
		const maxTokenSize = 1024 * 1024 // 1MB
		scanner.Buffer(make([]byte, maxTokenSize), maxTokenSize)

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			// Write raw JSON to debug log if enabled
			if debugLogFile != nil {
				debugLogFile.WriteString(line + "\n")
				debugLogFile.Sync()
			}

			event, err := events.ParseEvent(line)
			if err != nil {
				log.Printf("Warning: skipping malformed JSON line: %v", err)
				// Log the actual line content when we have a parse error
				if debugLogFile != nil {
					debugLogFile.WriteString("# PARSE ERROR: " + err.Error() + "\n")
				}
				continue
			}

			eventChan <- event
		}

		// Handle scanner errors (if any)
		if err := scanner.Err(); err != nil {
			log.Printf("Warning: error reading stream: %v", err)
		}
	}()

	return eventChan
}

// StreamEventsFromProcess is a convenience function that streams events
// from a ClaudeProcess's stdout.
func StreamEventsFromProcess(process *ClaudeProcess) <-chan events.Event {
	return StreamEvents(process.Stdout)
}
