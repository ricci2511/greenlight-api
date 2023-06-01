package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// Represents the log entry severity level.
type Level int8

const (
	LevelInfo  Level = iota // 0
	LevelError              // 1
	LevelFatal              // 2
	LevelOff                // 3
)

// Returns human readable string based on the log severity level.
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

type Logger struct {
	out      io.Writer  // Output destination for the log entry
	minLevel Level      // Minimum level to log
	mu       sync.Mutex // Ensures atomic log writes
}

// Returns new Logger instance with the given output destination and minimum log level.
func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{out: out, minLevel: minLevel}
}

// Internal jsonlog method to write a log entry to the output destination.
func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {
	// Ignore log entries with a severity level below the minimum set.
	if level < l.minLevel {
		return 0, nil
	}

	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    message,
		Properties: properties,
	}

	// Only include stack trace info for error and fatal log entries.
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	var line []byte

	// Marshal log entry to JSON. If it fails, fallback to a simple string message.
	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": failed to marshal log message: " + err.Error())
	}

	// Prevent multiple log entries from being written over each other.
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.out.Write(append(line, '\n'))
}

// Helper to write a log entry with the INFO severity level.
func (l *Logger) PrintInfo(message string, data map[string]string) {
	l.print(LevelInfo, message, data)
}

// Helper to write a log entry with the ERROR severity level.
func (l *Logger) PrintError(err error, data map[string]string) {
	l.print(LevelError, err.Error(), data)
}

// Helper to write a log entry with the FATAL severity level.
func (l *Logger) PrintFatal(err error, data map[string]string) {
	l.print(LevelFatal, err.Error(), data)
	os.Exit(1) // FATAL errors should terminate the application
}

// Writes a log entry with the ERROR severity level without any additional properties.
//
// This is used by http.Server.ErrorLog to write HTTP server errors to our jsonlog.
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.print(LevelError, string(message), nil)
}
