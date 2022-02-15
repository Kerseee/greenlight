// Package jsonlog implements functions and types for logging in JSON.
package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// Level indicates the severity of the log entry.
type Level int8

// Severity levels. From low to high are LevelInfo, LevelError, LevelFatal, LevelOff.
const (
	LevelInfo Level = iota
	LevelError
	LevelFatal
	LevelOff
)

// String() return the severity level of l. It satisfies the fmt.Stringer interface.
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

// Logger holds an output destination and a minimum severity level
// that log entries will be written for, and a sync.Mutex for coordinating the writes.
type Logger struct {
	out      io.Writer // Output destination
	minLevel Level
	mu       sync.Mutex
}

// New returns a logger that writes to out
// when the severity level is equal or larger then minLevel.
func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

// print write the severity level, message and properties to l.out.
func (l *Logger) print(level Level, msg string, properties map[string]string) (int, error) {
	// Do not write anything if level is lower than l.minLevel.
	if level < l.minLevel {
		return 0, nil
	}

	// A struct holding informations of logging entry for JSON marshalling.
	aux := struct {
		Level      string            `json:"level"`                // severity level
		Time       string            `json:"time"`                 // logging time
		Message    string            `json:"message"`              // message of the log entry
		Properties map[string]string `json:"properties,omitempty"` // additional information
		Trace      string            `json:"trace,omitempty"`      // stack trace for debugging
	}{
		Level:      level.String(),
		Time:       time.Now().Local().Format(time.RFC3339Nano),
		Message:    msg,
		Properties: properties,
	}
	// Add debug stack to aux if severity level is higher.
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	// Encoding the logging entry.
	entry, err := json.Marshal(aux)
	if err != nil {
		entry = []byte(LevelError.String() + ": unable to marshal log message: " + err.Error())
	}

	// Write to l.out.
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.out.Write(append(entry, '\n'))
}

// PrintInfo write msg and properties to l.out with normal severity level.
func (l *Logger) PrintInfo(msg string, properties map[string]string) {
	l.print(LevelInfo, msg, properties)
}

// PrintError write msg and properties to l.out with error severity level.
func (l *Logger) PrintError(msg string, properties map[string]string) {
	l.print(LevelInfo, msg, properties)
}

// PrintFatal write msg and properties to l.out with fatal severity level.
// It terminates the application.
func (l *Logger) PrintFatal(msg string, properties map[string]string) {
	l.print(LevelInfo, msg, properties)
	os.Exit(1)
}

// Write writes the message msg to l.out with error severity level.
// It satisfies the io.Writer interface.
func (l *Logger) Write(msg []byte) (n int, err error) {
	return l.print(LevelError, string(msg), nil)
}
