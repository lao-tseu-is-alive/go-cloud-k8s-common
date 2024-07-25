package golog

import (
	"errors"
	"fmt"
	"github.com/mgutz/ansi"
	"log"
	"os"
)

type SimpleLogger struct {
	logger   *log.Logger
	maxLevel Level
}

func NewSimpleLogger(logLevel Level, prefix string) (MyLogger, error) {

	l := log.New(os.Stdout, prefix, log.Ldate|log.Ltime|log.Lshortfile)
	return &SimpleLogger{logger: l, maxLevel: logLevel}, nil
}

func (l *SimpleLogger) Debug(msg string, v ...any) {
	if l.maxLevel <= DebugLevel {
		color := ansi.ColorFunc("cyan")
		l.logger.Output(2, color(fmt.Sprintf("DEBUG: %s", fmt.Sprintf(msg, v...))))
	}
}

func (l *SimpleLogger) Info(msg string, v ...interface{}) {
	if l.maxLevel <= InfoLevel {
		color := ansi.ColorFunc("white+h")
		l.logger.Output(2, color(fmt.Sprintf(" 📣 INFO : %s", fmt.Sprintf(msg, v...))))
	}
}

func (l *SimpleLogger) Warn(msg string, v ...interface{}) {
	if l.maxLevel <= WarnLevel {
		color := ansi.ColorFunc("yellow+h")
		l.logger.Output(2, color(fmt.Sprintf(" 🚩 WARN : %s", fmt.Sprintf(msg, v...))))
	}
}

func (l *SimpleLogger) Error(msg string, v ...interface{}) {
	if l.maxLevel <= ErrorLevel {
		color := ansi.ColorFunc("red+b:white+h")
		l.logger.Output(2, color(fmt.Sprintf(" ⚠️ ⚡ ERROR: %s", fmt.Sprintf(msg, v...))))
	}
}

func (l *SimpleLogger) Fatal(msg string, v ...interface{}) {
	color := ansi.ColorFunc("yellow+h:red+h")
	l.logger.Output(2, color(fmt.Sprintf("💥 💥 FATAL: %s", fmt.Sprintf(msg, v...))))
	os.Exit(1)
}

func (l *SimpleLogger) GetDefaultLogger() (*log.Logger, error) {
	if l.logger != nil {
		return l.logger, nil
	} else {
		return nil, errors.New("sorry, no default logger initialised at this time")
	}
}
