package golog

import (
	"fmt"
	"io"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	sug   *zap.SugaredLogger
	level Level
}

func mapLevel(l Level) zapcore.Level {
	switch l {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// NewZapLogger creates a JSON logger backed by Zap that writes to out and respects golog.Level.
func NewZapLogger(out io.Writer, logLevel Level, prefix string) (MyLogger, error) {
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
	encoder := zapcore.NewJSONEncoder(encCfg)

	ws := zapcore.AddSync(out) // honors any io.Writer: file, stdout, stderr, io.Discard
	core := zapcore.NewCore(encoder, ws, mapLevel(logLevel))

	// add optional prefix as a constant field
	var opts []zap.Option
	opts = append(opts, zap.AddCaller(), zap.AddCallerSkip(1))
	if prefix != "" {
		opts = append(opts, zap.Fields(zap.String("prefix", prefix)))
	}

	l := zap.New(core, opts...)
	return &ZapLogger{sug: l.Sugar(), level: logLevel}, nil
}

func (z *ZapLogger) Debug(msg string, v ...any) {
	if z.level <= DebugLevel {
		z.sug.Debugf(msg, v...)
	}
}
func (z *ZapLogger) Info(msg string, v ...any) {
	if z.level <= InfoLevel {
		z.sug.Infof(msg, v...)
	}
}
func (z *ZapLogger) Warn(msg string, v ...any) {
	if z.level <= WarnLevel {
		z.sug.Warnf(msg, v...)
	}
}
func (z *ZapLogger) Error(msg string, v ...any) {
	if z.level <= ErrorLevel {
		z.sug.Errorf(msg, v...)
	}
}
func (z *ZapLogger) Fatal(msg string, v ...any) { z.sug.Fatalf(msg, v...) }

// GetDefaultLogger cannot return a *log.Logger for zap; mirror SimpleLogger's signature by returning an error.
func (z *ZapLogger) GetDefaultLogger() (*log.Logger, error) {
	// Not applicable for zap; stdlib logger not available here
	return nil, fmt.Errorf("not supported for zap JSON logger")
}
