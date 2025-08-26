package golog

import (
	"io"
	"log"
)

// MyLogger defines a minimal leveled logging interface with printf-style methods.
// Implementations should treat v as fmt.Sprintf arguments and honor the configured level.
type MyLogger interface {
	Debug(msg string, v ...any)
	Info(msg string, v ...any)
	Warn(msg string, v ...any)
	Error(msg string, v ...any)
	Fatal(msg string, v ...any)
	GetDefaultLogger() (*log.Logger, error)
}

// Level represents the logging verbosity threshold.
// Lower values are more verbose (Debug) and higher values are more severe (Fatal).
type Level int8

// Supported logging levels in increasing order of severity.
const (
	// DebugLevel enables verbose diagnostic logs useful during development.
	DebugLevel Level = iota // most verbose
	// InfoLevel enables general informational logs for normal operation.
	InfoLevel
	// WarnLevel enables logs for unexpected but non-fatal conditions.
	WarnLevel
	// ErrorLevel enables logs for failures that require attention.
	ErrorLevel
	// FatalLevel logs a critical error and typically terminates the process.
	FatalLevel // most severe
)

func NewLogger(loggerType string, out io.Writer, logLevel Level, prefix string) (MyLogger, error) {
	var (
		logger MyLogger
		err    error
	)

	switch loggerType {
	case "zap":
		logger, err = NewZapLogger(out, logLevel, prefix)
	default:
		logger, err = NewSimpleLogger(out, logLevel, prefix)
		if err != nil {
			return nil, err
		}
	}

	return logger, nil
}
