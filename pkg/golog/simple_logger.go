package golog

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

// Define ANSI color codes
const (
	reset                   = "\033[0m"
	black                   = "\033[30m"
	red                     = "\033[31m"
	green                   = "\033[32m"
	yellow                  = "\033[33m"
	blue                    = "\033[34m"
	purple                  = "\033[35m"
	cyan                    = "\033[36m"
	white                   = "\033[37m"
	blackHighIntensity      = "\033[1;90m"
	redHighIntensity        = "\033[1;91m"
	greenHighIntensity      = "\033[1;92m"
	yellowHighIntensity     = "\033[1;93m"
	blueHighIntensity       = "\033[1;94m"
	purpleHighIntensity     = "\033[1;95m"
	cyanHighIntensity       = "\033[1;96m"
	whiteHighIntensity      = "\033[1;97m"
	redBold                 = "\033[1;31m"
	redBackGroundWhiteText  = "\033[1;97;41m" // Red background with white text
	redBackGroundYellowText = "\033[1;93;41m" // Red background with white text
)

type SimpleLogger struct {
	logger   *log.Logger
	maxLevel Level
}

func NewSimpleLogger(out io.Writer, logLevel Level, prefix string) (MyLogger, error) {
	l := log.New(out, prefix, log.Ldate|log.Ltime|log.Lshortfile)
	return &SimpleLogger{logger: l, maxLevel: logLevel}, nil
}

func (l *SimpleLogger) Debug(msg string, v ...any) {
	if l.maxLevel <= DebugLevel {
		l.logger.Output(2, fmt.Sprintf("%sDEBUG: %s%s", cyan, fmt.Sprintf(msg, v...), reset))
	}
}

func (l *SimpleLogger) Info(msg string, v ...any) {
	if l.maxLevel <= InfoLevel {
		l.logger.Output(2, fmt.Sprintf("%s 📣 INFO : %s%s", whiteHighIntensity, fmt.Sprintf(msg, v...), reset))
	}
}

func (l *SimpleLogger) Warn(msg string, v ...any) {
	if l.maxLevel <= WarnLevel {
		l.logger.Output(2, fmt.Sprintf("%s 🚩 WARN : %s%s", yellowHighIntensity, fmt.Sprintf(msg, v...), reset))
	}
}

func (l *SimpleLogger) Error(msg string, v ...any) {
	if l.maxLevel <= ErrorLevel {
		l.logger.Output(2, fmt.Sprintf("%s ⚠️ ⚡ ERROR: %s%s", redBackGroundWhiteText, fmt.Sprintf(msg, v...), reset))
	}
}

func (l *SimpleLogger) Fatal(msg string, v ...any) {
	l.logger.Output(2, fmt.Sprintf("%s💥 💥 FATAL: %s%s", redBackGroundYellowText, fmt.Sprintf(msg, v...), reset))
	os.Exit(1)
}

func (l *SimpleLogger) GetDefaultLogger() (*log.Logger, error) {
	if l.logger != nil {
		return l.logger, nil
	} else {
		return nil, errors.New("sorry, no default logger initialized at this time")
	}
}
