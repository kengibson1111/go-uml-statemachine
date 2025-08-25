package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging capabilities
type Logger struct {
	level        LogLevel
	output       io.Writer
	prefix       string
	enableCaller bool
	fields       map[string]interface{}
}

// LoggerConfig holds configuration for creating a logger
type LoggerConfig struct {
	Level        LogLevel
	Output       io.Writer
	Prefix       string
	EnableCaller bool
	LogFile      string // Optional file path for logging
}

// DefaultLoggerConfig returns a default logger configuration
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:        LogLevelInfo,
		Output:       os.Stdout,
		Prefix:       "[go-uml-statemachine-parsers]",
		EnableCaller: true,
	}
}

// NewLogger creates a new logger with the specified configuration
func NewLogger(config *LoggerConfig) (*Logger, error) {
	if config == nil {
		config = DefaultLoggerConfig()
	}

	output := config.Output
	if config.LogFile != "" {
		// Create log directory if it doesn't exist
		logDir := filepath.Dir(config.LogFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file for writing (append mode)
		file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		output = file
	} else if output == nil {
		// Default to stdout if no output is specified
		output = os.Stdout
	}

	return &Logger{
		level:        config.Level,
		output:       output,
		prefix:       config.Prefix,
		enableCaller: config.EnableCaller,
		fields:       make(map[string]interface{}),
	}, nil
}

// NewDefaultLogger creates a logger with default configuration
func NewDefaultLogger() *Logger {
	logger, _ := NewLogger(DefaultLoggerConfig())
	return logger
}

// WithField adds a field to the logger context
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := &Logger{
		level:        l.level,
		output:       l.output,
		prefix:       l.prefix,
		enableCaller: l.enableCaller,
		fields:       make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new field
	newLogger.fields[key] = value
	return newLogger
}

// WithFields adds multiple fields to the logger context
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := &Logger{
		level:        l.level,
		output:       l.output,
		prefix:       l.prefix,
		enableCaller: l.enableCaller,
		fields:       make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// WithError adds an error to the logger context
func (l *Logger) WithError(err error) *Logger {
	return l.WithField("error", err.Error())
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// Debug logs a debug message
func (l *Logger) Debug(message string) {
	l.log(LogLevelDebug, message)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(LogLevelDebug, fmt.Sprintf(format, args...))
}

// Info logs an info message
func (l *Logger) Info(message string) {
	l.log(LogLevelInfo, message)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(LogLevelInfo, fmt.Sprintf(format, args...))
}

// Warn logs a warning message
func (l *Logger) Warn(message string) {
	l.log(LogLevelWarn, message)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(LogLevelWarn, fmt.Sprintf(format, args...))
}

// Error logs an error message
func (l *Logger) Error(message string) {
	l.log(LogLevelError, message)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(LogLevelError, fmt.Sprintf(format, args...))
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(message string) {
	l.log(LogLevelFatal, message)
	os.Exit(1)
}

// Fatalf logs a formatted fatal message and exits the program
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(LogLevelFatal, fmt.Sprintf(format, args...))
	os.Exit(1)
}

// log is the internal logging method
func (l *Logger) log(level LogLevel, message string) {
	if level < l.level {
		return
	}

	// Build log entry
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	levelStr := level.String()

	// Build fields string
	var fieldsStr strings.Builder
	if len(l.fields) > 0 {
		fieldsStr.WriteString(" [")
		first := true
		for k, v := range l.fields {
			if !first {
				fieldsStr.WriteString(", ")
			}
			fieldsStr.WriteString(fmt.Sprintf("%s=%v", k, v))
			first = false
		}
		fieldsStr.WriteString("]")
	}

	// Add caller information if enabled
	var callerStr string
	if l.enableCaller {
		if pc, file, line, ok := runtime.Caller(2); ok {
			funcName := runtime.FuncForPC(pc).Name()
			// Extract just the function name without package path
			if lastSlash := strings.LastIndex(funcName, "/"); lastSlash >= 0 {
				funcName = funcName[lastSlash+1:]
			}
			if lastDot := strings.LastIndex(funcName, "."); lastDot >= 0 {
				funcName = funcName[lastDot+1:]
			}

			// Extract just the filename without full path
			filename := filepath.Base(file)
			callerStr = fmt.Sprintf(" [%s:%d:%s]", filename, line, funcName)
		}
	}

	// Format final log message
	logMessage := fmt.Sprintf("%s %s %s%s%s: %s\n",
		timestamp, levelStr, l.prefix, fieldsStr.String(), callerStr, message)

	// Write to output
	fmt.Fprint(l.output, logMessage)
}

// Close closes the logger and any associated resources
func (l *Logger) Close() error {
	if closer, ok := l.output.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// Global logger instance
var globalLogger *Logger

func init() {
	globalLogger = NewDefaultLogger()
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	return globalLogger
}

// Global logging functions that use the global logger
func Debug(message string) {
	globalLogger.Debug(message)
}

func Debugf(format string, args ...interface{}) {
	globalLogger.Debugf(format, args...)
}

func Info(message string) {
	globalLogger.Info(message)
}

func Infof(format string, args ...interface{}) {
	globalLogger.Infof(format, args...)
}

func Warn(message string) {
	globalLogger.Warn(message)
}

func Warnf(format string, args ...interface{}) {
	globalLogger.Warnf(format, args...)
}

func Error(message string) {
	globalLogger.Error(message)
}

func Errorf(format string, args ...interface{}) {
	globalLogger.Errorf(format, args...)
}

func Fatal(message string) {
	globalLogger.Fatal(message)
}

func Fatalf(format string, args ...interface{}) {
	globalLogger.Fatalf(format, args...)
}

func WithField(key string, value interface{}) *Logger {
	return globalLogger.WithField(key, value)
}

func WithFields(fields map[string]interface{}) *Logger {
	return globalLogger.WithFields(fields)
}

func WithError(err error) *Logger {
	return globalLogger.WithError(err)
}
