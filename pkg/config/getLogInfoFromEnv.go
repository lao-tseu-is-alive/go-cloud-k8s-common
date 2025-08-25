package config

import (
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

// GetLogWriterFromEnvOrPanic returns the name of the filename to use for LOG from the content of the env variable :
// LOG_FILE : string containing the filename to use for LOG, use DISCARD for no log, default is STDERR
func GetLogWriterFromEnvOrPanic(defaultLogName string) io.Writer {
	logFileName := defaultLogName
	val, exist := os.LookupEnv("LOG_FILE")
	if exist {
		logFileName = val
	}
	if utf8.RuneCountInString(logFileName) < 5 {
		panic(fmt.Sprintf("💥💥 error env LOG_FILE filename should contain at least %d characters (got %d).",
			5, utf8.RuneCountInString(val)))
	}
	switch logFileName {
	case "stdout":
		return os.Stdout
	case "stderr":
		return os.Stderr
	case "DISCARD":
		return io.Discard
	default:
		// Open the file with append, create, and write permissions.
		// The 0644 permission allows the owner to read/write and others to read.
		file, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// Return an error if the file cannot be opened (e.g., due to permissions).
			panic(fmt.Sprintf("💥💥 ERROR: LOG_FILE %q could not be open : %v", logFileName, err))
		}
		return file
	}
}
