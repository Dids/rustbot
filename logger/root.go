package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Dids/rustbot/eventhandler"
)

// TODO: Implement from scratch, use this for reference:
//       https://www.ardanlabs.com/blog/2013/11/using-log-package-in-go.html

// Calls something exactly once (this initializes our Logger instance)
var callOnce sync.Once

// Logger is a flexible and configurable logging solution
type Logger struct {
	// Options to configure and customize the behavior
	Options Options

	// EventHandler responsible for emitting/receiving messages between services
	EventHandler *eventhandler.EventHandler

	// Internal log level specific loggers
	traceLog   *log.Logger
	infoLog    *log.Logger
	warningLog *log.Logger
	errorLog   *log.Logger

	// Internal reference to the log file
	logFile *os.File
}

// Options are used for customizing the Logger instance
type Options struct {
	// Level sets the minimum log level to include when writing log messages (levels below are dropped)
	Level LogLevel
	// File name for the log file
	File string
}

// Define the default options (constant)
const (
	// DefaultLogLevel is the default minimum log level to log
	DefaultLogLevel LogLevel = Info
	// DefaultLogFile is the default filename of the log file
	DefaultLogFile string = "rustbot.log"
)

// DefaultOptions returns an instance of Options with the default values applied
func DefaultOptions() Options {
	options := Options{
		Level: DefaultLogLevel,
		File:  DefaultLogFile,
	}
	return options
}

// LogLevel sets the importance of a single log message
type LogLevel int

// Define the log levels themselves (constant)
const (
	// Trace log level, logs everything (Trace, Info, Warning, Error)
	Trace LogLevel = 0
	// Info log level, logs info and above (Info, Warning, Error)
	Info LogLevel = 1
	// Warning log level, logs warning and above (Warning, Error)
	Warning LogLevel = 2
	// Error log level, only errors
	Error LogLevel = 3
)

// The shared Logger instance
var instance *Logger

// GetLogger returns the shared Logger instance
func GetLogger() *Logger {
	return NewLogger(Options{}, nil)
}

// NewLogger creates and returns the shared Logger instance
func NewLogger(options Options, handler *eventhandler.EventHandler) *Logger {
	// Apply default options if missing
	if options == (Options{}) {
		options = DefaultOptions()
	}

	// Initialize the Logger exactly once
	callOnce.Do(func() {
		// Create a new instance of Logger
		instance = &Logger{
			Options:      options,
			EventHandler: handler,
		}

		// Setup file logging
		file, err := os.OpenFile(options.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic("Failed to open log file: " + err.Error())
		}
		instance.logFile = file

		// Initialize the log level specific loggers
		instance.traceLog = log.New(io.MultiWriter(file, os.Stdout), "TRACE: ", log.Ldate|log.Ltime|log.Llongfile)
		instance.infoLog = log.New(io.MultiWriter(file, os.Stdout), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
		instance.warningLog = log.New(io.MultiWriter(file, os.Stdout), "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
		instance.errorLog = log.New(io.MultiWriter(file, os.Stderr), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	})

	return instance
}

// Trace log level message
func (logger *Logger) Trace(message ...interface{}) {
	// Only log if log level is below or equal to Trace
	if logger.Options.Level <= Trace {
		// Relay message to our event handler
		logger.logEvent(eventhandler.TraceLogType, message...)

		// Log the message
		logger.traceLog.Println(message...)
	}
}

// Info log level message
func (logger *Logger) Info(message ...interface{}) {
	// Only log if log level is below or equal to Info
	if logger.Options.Level <= Info {
		// Relay message to our event handler
		logger.logEvent(eventhandler.InfoLogType, message...)

		// Log the message
		logger.infoLog.Println(message...)
	}
}

// Warning log level message
func (logger *Logger) Warning(message ...interface{}) {
	// Only log if log level is below or equal to Warning
	if logger.Options.Level <= Warning {
		// Relay message to our event handler
		logger.logEvent(eventhandler.WarningLogType, message...)

		// Log the message
		logger.warningLog.Println(message...)
	}
}

// Error log level message
func (logger *Logger) Error(message ...interface{}) {
	// Only log if log level is below or equal to Error
	if logger.Options.Level <= Error {
		// Relay message to our event handler
		logger.logEvent(eventhandler.ErrorLogType, message...)

		// Log the message
		logger.errorLog.Println(message...)
	}
}

// Panic prints the message in the error log and exits/panics
func (logger *Logger) Panic(message ...interface{}) {
	// Relay message to our event handler
	logger.logEvent(eventhandler.PanicLogType, message...)

	// Delay execution, so our event handler has time to relay the panic message
	time.Sleep(5 * time.Second)

	// Log the message and exit
	logger.errorLog.Panic(message...)
}

func (logger *Logger) logEvent(messageType eventhandler.MessageType, message ...interface{}) {
	// Create a single string from each message element
	parsedMessage := ""
	for _, value := range message {
		if len(parsedMessage) > 0 {
			parsedMessage = fmt.Sprintf("%s %s", parsedMessage, value)
		} else {
			parsedMessage = fmt.Sprintf("%s", value)
		}
	}

	// Emit the constructed message string through the event handler
	if logger.EventHandler != nil {
		logger.EventHandler.Emit(eventhandler.Message{Event: "receive_logger_message", Message: parsedMessage, Type: messageType})
	}
}
