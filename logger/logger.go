package logger

import (
	"io"
	"log"
	"os"
)

// Logger is the global debug logger, initially disabled
var Logger *log.Logger

func init() {
	// Initialize debug logger to discard output by default
	Logger = log.New(io.Discard, "", log.LstdFlags|log.Lshortfile)
}

// EnableDebug enables debug logging output to stderr
func EnableDebug() {
	Logger.SetOutput(os.Stderr)
}

// DisableDebug disables debug logging by discarding output
func DisableDebug() {
	Logger.SetOutput(io.Discard)
}

// Printf logs a formatted debug message
func Printf(format string, args ...interface{}) {
	Logger.Printf(format, args...)
}