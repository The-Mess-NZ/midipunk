package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

/*
LogLevel represents the different logging levels:
  - NONE: No logging
  - ERROR: Only error messages
  - INFO: Informational messages and above
  - DEBUG: Debug messages and above
  - ALL: All messages including high-frequency MIDI timing messages
*/
type LogLevel int

const (
	NONE LogLevel = iota
	ERROR
	INFO
	DEBUG
	ALL
)

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	switch l {
	case NONE:
		return "none"
	case ERROR:
		return "error"
	case INFO:
		return "info"
	case DEBUG:
		return "debug"
	case ALL:
		return "all"
	default:
		return "unknown"
	}
}

// LogLevelFromString converts a string to LogLevel
func LogLevelFromString(s string) LogLevel {
	switch strings.ToLower(s) {
	case "none":
		return NONE
	case "error":
		return ERROR
	case "info":
		return INFO
	case "debug":
		return DEBUG
	case "all":
		return ALL
	default:
		return NONE
	}
}

// LogMessage represents a log message
type LogMessage struct {
	Level     LogLevel
	Timestamp time.Time
	Message   string
	IsTiming  bool // Flag to identify MIDI timing messages
}

// Logger represents the logging interface
type Logger struct {
	level      LogLevel
	msgChannel chan LogMessage
	done       chan bool
}

// Global logger instance
var globalLogger *Logger

// Init initializes the global logger
func Init(level LogLevel) {
	globalLogger = &Logger{
		level:      level,
		msgChannel: make(chan LogMessage, 1000), // Large buffer for high-frequency MIDI messages
		done:       make(chan bool),
	}

	go globalLogger.processMessages()
}

// SetLevel updates the logging level
func SetLevel(level LogLevel) {
	if globalLogger != nil {
		globalLogger.level = level
	}
}

// GetLevel retrieves the current logging level set from the config file or command line.
func GetLevel() LogLevel {
	if globalLogger != nil {
		return globalLogger.level
	}
	return NONE
}

// GetChannel returns the message channel for sending log messages
func GetChannel() chan<- LogMessage {
	if globalLogger == nil {
		return nil
	}
	return globalLogger.msgChannel
}

// processMessages handles incoming log messages
func (l *Logger) processMessages() {
	for {
		select {
		case msg := <-l.msgChannel:
			l.handleMessage(msg)
		case <-l.done:
			return
		}
	}
}

// handleMessage processes a single log message
func (l *Logger) handleMessage(msg LogMessage) {
	// Check if we should log this message based on level
	if l.level == NONE {
		return
	}

	// Filter out timing messages unless level is ALL
	if msg.IsTiming && l.level != ALL {
		return
	}

	// Check if message level is high enough to be logged
	if msg.Level > l.level {
		return
	}

	// Format and output the message
	prefix := fmt.Sprintf("[%s] %s: ", msg.Timestamp.Format("15:04:05.000"), strings.ToUpper(msg.Level.String()))
	log.Printf("%s%s", prefix, msg.Message)
}

// Shutdown stops the logger
func Shutdown() {
	if globalLogger != nil {
		close(globalLogger.done)
	}
}

// Helper functions for different log levels

// Error logs an error message
func Error(format string, args ...interface{}) {
	if globalLogger == nil {
		return
	}

	select {
	case globalLogger.msgChannel <- LogMessage{
		Level:     ERROR,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf(format, args...),
		IsTiming:  false,
	}:
	default:
		// Drop message if channel is full
	}
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	if globalLogger == nil {
		return
	}

	select {
	case globalLogger.msgChannel <- LogMessage{
		Level:     INFO,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf(format, args...),
		IsTiming:  false,
	}:
	default:
		// Drop message if channel is full
	}
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	if globalLogger == nil {
		return
	}

	select {
	case globalLogger.msgChannel <- LogMessage{
		Level:     DEBUG,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf(format, args...),
		IsTiming:  false,
	}:
	default:
		// Drop message if channel is full
	}
}

// DebugMIDI logs a MIDI message (non-timing)
func DebugMIDI(format string, args ...interface{}) {
	if globalLogger == nil {
		return
	}

	select {
	case globalLogger.msgChannel <- LogMessage{
		Level:     DEBUG,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf(format, args...),
		IsTiming:  false,
	}:
	default:
		// Drop message if channel is full
	}
}

// DebugMIDITiming logs a MIDI timing message (only shown at ALL level)
func DebugMIDITiming(format string, args ...interface{}) {
	if globalLogger == nil {
		return
	}

	select {
	case globalLogger.msgChannel <- LogMessage{
		Level:     ALL,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf(format, args...),
		IsTiming:  true,
	}:
	default:
		// Drop message if channel is full
	}
}

func init() {
	// Set up log package to not include timestamp since we handle it ourselves
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
}
