package logger

import (
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	// Try creating a new Logger
	logger := GetLogger()

	// Try all log levels
	logger.Trace("This is a Trace level message")
	logger.Trace("This is an Info level message")
	logger.Trace("This is a Warning level message")
	logger.Trace("This is an Error level message")

	// Remove the log file
	os.Remove(DefaultLogFile)
}
