package golog

import (
	"fmt"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	sug   *zap.SugaredLogger
	level Level
}

// NewZapLogger creates a JSON logger backed by Uber Zap that conforms to MyLogger.
// It maps golog.Level to zapcore.Level and configures JSON output with RFC3339 timestamps and caller info.
func NewZapLogger(logLevel Level, prefix string) (MyLogger, error) {
	// Map golog.Level to zapcore.Level
	var zl zapcore.Level
	switch logLevel {
	case DebugLevel:
		zl = zapcore.DebugLevel
	case InfoLevel:
		zl = zapcore.InfoLevel
	case WarnLevel:
		zl = zapcore.WarnLevel
	case ErrorLevel:
		zl = zapcore.ErrorLevel
	case FatalLevel:
		zl = zapcore.FatalLevel
	default:
		zl = zapcore.InfoLevel
	}

	encCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zl),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    encCfg,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields:    map[string]any{},
	}

	if prefix != "" {
		cfg.InitialFields["prefix"] = prefix
	}

	l, err := cfg.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {
		return nil, fmt.Errorf("build zap logger: %w", err)
	}

	return &ZapLogger{sug: l.Sugar(), level: logLevel}, nil
}

func (z *ZapLogger) Debug(msg string, v ...interface{}) {
	if z.level <= DebugLevel {
		z.sug.Debugf(msg, v...)
	}
}

func (z *ZapLogger) Info(msg string, v ...interface{}) {
	if z.level <= InfoLevel {
		z.sug.Infof(msg, v...)
	}
}

func (z *ZapLogger) Warn(msg string, v ...interface{}) {
	if z.level <= WarnLevel {
		z.sug.Warnf(msg, v...)
	}
}

func (z *ZapLogger) Error(msg string, v ...interface{}) {
	if z.level <= ErrorLevel {
		z.sug.Errorf(msg, v...)
	}
}

func (z *ZapLogger) Fatal(msg string, v ...interface{}) {
	// zap.SugaredLogger.Fatalf calls os.Exit(1) after logging
	z.sug.Fatalf(msg, v...)
}

func (z *ZapLogger) GetDefaultLogger() (*log.Logger, error) {
	// Not applicable for zap; stdlib logger not available here
	return nil, fmt.Errorf("not supported for zap JSON logger")
}
