package logging

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelDebug, "DEBUG"},
		{LogLevelInfo, "INFO"},
		{LogLevelWarn, "WARN"},
		{LogLevelError, "ERROR"},
		{LogLevelFatal, "FATAL"},
		{LogLevel(999), "UNKNOWN"}, // Unknown level
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDefaultLoggerConfig(t *testing.T) {
	config := DefaultLoggerConfig()

	if config.Level != LogLevelInfo {
		t.Errorf("DefaultLoggerConfig() Level = %v, want %v", config.Level, LogLevelInfo)
	}

	if config.Output != os.Stdout {
		t.Errorf("DefaultLoggerConfig() Output = %v, want %v", config.Output, os.Stdout)
	}

	if config.Prefix != "[go-uml-statemachine-parsers]" {
		t.Errorf("DefaultLoggerConfig() Prefix = %v, want %v", config.Prefix, "[go-uml-statemachine-parsers]")
	}

	if !config.EnableCaller {
		t.Error("DefaultLoggerConfig() EnableCaller should be true")
	}
}

func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	config := &LoggerConfig{
		Level:        LogLevelDebug,
		Output:       &buf,
		Prefix:       "[test]",
		EnableCaller: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	if logger.level != LogLevelDebug {
		t.Errorf("NewLogger() level = %v, want %v", logger.level, LogLevelDebug)
	}

	if logger.prefix != "[test]" {
		t.Errorf("NewLogger() prefix = %v, want %v", logger.prefix, "[test]")
	}

	if logger.enableCaller {
		t.Error("NewLogger() enableCaller should be false")
	}
}

func TestNewLogger_WithNilConfig(t *testing.T) {
	logger, err := NewLogger(nil)
	if err != nil {
		t.Fatalf("NewLogger() with nil config error = %v", err)
	}

	// Should use default config
	if logger.level != LogLevelInfo {
		t.Errorf("NewLogger() with nil config level = %v, want %v", logger.level, LogLevelInfo)
	}
}

func TestNewLogger_WithLogFile(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := &LoggerConfig{
		Level:   LogLevelInfo,
		LogFile: logFile,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger() with log file error = %v", err)
	}

	// Test that we can write to the log file
	logger.Info("test message")

	// Close the logger to flush any buffers
	if err := logger.Close(); err != nil {
		t.Errorf("Logger.Close() error = %v", err)
	}

	// Check that the log file was created and contains the message
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "test message") {
		t.Error("Log file should contain the test message")
	}
}

func TestLogger_WithField(t *testing.T) {
	var buf bytes.Buffer
	config := &LoggerConfig{
		Level:        LogLevelInfo,
		Output:       &buf,
		Prefix:       "[test]",
		EnableCaller: false,
	}

	logger, _ := NewLogger(config)
	fieldLogger := logger.WithField("key", "value")

	fieldLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "key=value") {
		t.Error("Log output should contain the field")
	}
}

func TestLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	config := &LoggerConfig{
		Level:        LogLevelInfo,
		Output:       &buf,
		Prefix:       "[test]",
		EnableCaller: false,
	}

	logger, _ := NewLogger(config)
	fieldsLogger := logger.WithFields(map[string]any{
		"key1": "value1",
		"key2": 42,
	})

	fieldsLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "key1=value1") {
		t.Error("Log output should contain key1=value1")
	}
	if !strings.Contains(output, "key2=42") {
		t.Error("Log output should contain key2=42")
	}
}

func TestLogger_WithError(t *testing.T) {
	var buf bytes.Buffer
	config := &LoggerConfig{
		Level:        LogLevelInfo,
		Output:       &buf,
		Prefix:       "[test]",
		EnableCaller: false,
	}

	logger, _ := NewLogger(config)
	errorLogger := logger.WithError(os.ErrNotExist)

	errorLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "error=file does not exist") {
		t.Error("Log output should contain the error")
	}
}

func TestLogger_SetLevel(t *testing.T) {
	var buf bytes.Buffer
	config := &LoggerConfig{
		Level:        LogLevelInfo,
		Output:       &buf,
		EnableCaller: false,
	}

	logger, _ := NewLogger(config)

	// Debug message should not appear with Info level
	logger.Debug("debug message")
	if strings.Contains(buf.String(), "debug message") {
		t.Error("Debug message should not appear with Info level")
	}

	// Change level to Debug
	logger.SetLevel(LogLevelDebug)
	buf.Reset()

	// Now debug message should appear
	logger.Debug("debug message")
	if !strings.Contains(buf.String(), "debug message") {
		t.Error("Debug message should appear with Debug level")
	}
}

func TestLogger_LogLevels(t *testing.T) {
	var buf bytes.Buffer
	config := &LoggerConfig{
		Level:        LogLevelDebug,
		Output:       &buf,
		Prefix:       "[test]",
		EnableCaller: false,
	}

	logger, _ := NewLogger(config)

	// Test all log levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()

	// Check that all messages appear
	if !strings.Contains(output, "DEBUG") || !strings.Contains(output, "debug message") {
		t.Error("Debug message should appear")
	}
	if !strings.Contains(output, "INFO") || !strings.Contains(output, "info message") {
		t.Error("Info message should appear")
	}
	if !strings.Contains(output, "WARN") || !strings.Contains(output, "warn message") {
		t.Error("Warn message should appear")
	}
	if !strings.Contains(output, "ERROR") || !strings.Contains(output, "error message") {
		t.Error("Error message should appear")
	}
}

func TestLogger_FormattedLogs(t *testing.T) {
	var buf bytes.Buffer
	config := &LoggerConfig{
		Level:        LogLevelDebug,
		Output:       &buf,
		Prefix:       "[test]",
		EnableCaller: false,
	}

	logger, _ := NewLogger(config)

	// Test formatted log methods
	logger.Debugf("debug %s %d", "message", 1)
	logger.Infof("info %s %d", "message", 2)
	logger.Warnf("warn %s %d", "message", 3)
	logger.Errorf("error %s %d", "message", 4)

	output := buf.String()

	// Check that formatted messages appear correctly
	if !strings.Contains(output, "debug message 1") {
		t.Error("Formatted debug message should appear")
	}
	if !strings.Contains(output, "info message 2") {
		t.Error("Formatted info message should appear")
	}
	if !strings.Contains(output, "warn message 3") {
		t.Error("Formatted warn message should appear")
	}
	if !strings.Contains(output, "error message 4") {
		t.Error("Formatted error message should appear")
	}
}

func TestLogger_CallerInfo(t *testing.T) {
	var buf bytes.Buffer
	config := &LoggerConfig{
		Level:        LogLevelInfo,
		Output:       &buf,
		Prefix:       "[test]",
		EnableCaller: true,
	}

	logger, _ := NewLogger(config)
	logger.Info("test message")

	output := buf.String()

	// Should contain caller information
	if !strings.Contains(output, "logger_test.go") {
		t.Error("Log output should contain caller file information")
	}
	if !strings.Contains(output, "TestLogger_CallerInfo") {
		t.Error("Log output should contain caller function information")
	}
}

func TestLogger_NoCallerInfo(t *testing.T) {
	var buf bytes.Buffer
	config := &LoggerConfig{
		Level:        LogLevelInfo,
		Output:       &buf,
		Prefix:       "[test]",
		EnableCaller: false,
	}

	logger, _ := NewLogger(config)
	logger.Info("test message")

	output := buf.String()

	// Should not contain caller information
	if strings.Contains(output, "logger_test.go") {
		t.Error("Log output should not contain caller file information when disabled")
	}
}

func TestLogger_LogFiltering(t *testing.T) {
	var buf bytes.Buffer
	config := &LoggerConfig{
		Level:        LogLevelWarn, // Only warn and above
		Output:       &buf,
		EnableCaller: false,
	}

	logger, _ := NewLogger(config)

	// These should not appear
	logger.Debug("debug message")
	logger.Info("info message")

	// These should appear
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()

	// Check filtering
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should be filtered out")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should be filtered out")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should appear")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message should appear")
	}
}

func TestGlobalLogger(t *testing.T) {
	// Test that global logger functions work
	var buf bytes.Buffer

	// Create a custom logger and set it as global
	config := &LoggerConfig{
		Level:        LogLevelDebug,
		Output:       &buf,
		Prefix:       "[global-test]",
		EnableCaller: false,
	}

	logger, _ := NewLogger(config)
	SetGlobalLogger(logger)

	// Test global functions
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	output := buf.String()

	if !strings.Contains(output, "debug message") {
		t.Error("Global Debug() should work")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Global Info() should work")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Global Warn() should work")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Global Error() should work")
	}
}

func TestGlobalLoggerFormatted(t *testing.T) {
	var buf bytes.Buffer

	config := &LoggerConfig{
		Level:        LogLevelDebug,
		Output:       &buf,
		Prefix:       "[global-test]",
		EnableCaller: false,
	}

	logger, _ := NewLogger(config)
	SetGlobalLogger(logger)

	// Test global formatted functions
	Debugf("debug %d", 1)
	Infof("info %d", 2)
	Warnf("warn %d", 3)
	Errorf("error %d", 4)

	output := buf.String()

	if !strings.Contains(output, "debug 1") {
		t.Error("Global Debugf() should work")
	}
	if !strings.Contains(output, "info 2") {
		t.Error("Global Infof() should work")
	}
	if !strings.Contains(output, "warn 3") {
		t.Error("Global Warnf() should work")
	}
	if !strings.Contains(output, "error 4") {
		t.Error("Global Errorf() should work")
	}
}

func TestGlobalLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer

	config := &LoggerConfig{
		Level:        LogLevelInfo,
		Output:       &buf,
		Prefix:       "[global-test]",
		EnableCaller: false,
	}

	logger, _ := NewLogger(config)
	SetGlobalLogger(logger)

	// Test global field functions
	WithField("key", "value").Info("test message")
	WithFields(map[string]any{
		"key1": "value1",
		"key2": 42,
	}).Info("test message 2")
	WithError(os.ErrNotExist).Info("test message 3")

	output := buf.String()

	if !strings.Contains(output, "key=value") {
		t.Error("Global WithField() should work")
	}
	if !strings.Contains(output, "key1=value1") || !strings.Contains(output, "key2=42") {
		t.Error("Global WithFields() should work")
	}
	if !strings.Contains(output, "error=file does not exist") {
		t.Error("Global WithError() should work")
	}
}

func TestGetGlobalLogger(t *testing.T) {
	// Test that we can get the global logger
	logger := GetGlobalLogger()
	if logger == nil {
		t.Error("GetGlobalLogger() should not return nil")
	}

	// Test that setting and getting works
	var buf bytes.Buffer
	config := &LoggerConfig{
		Level:  LogLevelInfo,
		Output: &buf,
		Prefix: "[test-global]",
	}

	newLogger, _ := NewLogger(config)
	SetGlobalLogger(newLogger)

	retrieved := GetGlobalLogger()
	if retrieved != newLogger {
		t.Error("GetGlobalLogger() should return the logger that was set")
	}
}

func TestNewDefaultLogger(t *testing.T) {
	logger := NewDefaultLogger()
	if logger == nil {
		t.Error("NewDefaultLogger() should not return nil")
	}

	// Should have default configuration
	if logger.level != LogLevelInfo {
		t.Errorf("NewDefaultLogger() level = %v, want %v", logger.level, LogLevelInfo)
	}
}
