package models

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// ErrorType represents the type of error
type ErrorType int

const (
	ErrorTypeFileNotFound ErrorType = iota
	ErrorTypeValidation
	ErrorTypeDirectoryConflict
	ErrorTypeReferenceResolution
	ErrorTypeFileSystem
	ErrorTypeVersionParsing
	ErrorTypeConfiguration
	ErrorTypePermission
	ErrorTypeTimeout
	ErrorTypeNetwork
	ErrorTypeCorruption
)

// String returns the string representation of ErrorType
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeFileNotFound:
		return "file_not_found"
	case ErrorTypeValidation:
		return "validation"
	case ErrorTypeDirectoryConflict:
		return "directory_conflict"
	case ErrorTypeReferenceResolution:
		return "reference_resolution"
	case ErrorTypeFileSystem:
		return "file_system"
	case ErrorTypeVersionParsing:
		return "version_parsing"
	case ErrorTypeConfiguration:
		return "configuration"
	case ErrorTypePermission:
		return "permission"
	case ErrorTypeTimeout:
		return "timeout"
	case ErrorTypeNetwork:
		return "network"
	case ErrorTypeCorruption:
		return "corruption"
	default:
		return "unknown"
	}
}

// ErrorSeverity represents the severity level of an error
type ErrorSeverity int

const (
	ErrorSeverityLow ErrorSeverity = iota
	ErrorSeverityMedium
	ErrorSeverityHigh
	ErrorSeverityCritical
)

// String returns the string representation of ErrorSeverity
func (es ErrorSeverity) String() string {
	switch es {
	case ErrorSeverityLow:
		return "low"
	case ErrorSeverityMedium:
		return "medium"
	case ErrorSeverityHigh:
		return "high"
	case ErrorSeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// StackFrame represents a single frame in the call stack
type StackFrame struct {
	Function string
	File     string
	Line     int
}

// String returns the string representation of StackFrame
func (sf StackFrame) String() string {
	return fmt.Sprintf("%s (%s:%d)", sf.Function, sf.File, sf.Line)
}

// StateMachineError represents a custom error with comprehensive context
type StateMachineError struct {
	Type        ErrorType
	Message     string
	Cause       error
	Context     map[string]interface{}
	Severity    ErrorSeverity
	Timestamp   time.Time
	StackTrace  []StackFrame
	Operation   string // The operation that was being performed when the error occurred
	Component   string // The component where the error occurred
	Recoverable bool   // Whether the error is recoverable
}

// Error implements the error interface
func (e *StateMachineError) Error() string {
	var parts []string

	// Add error type and message
	parts = append(parts, fmt.Sprintf("[%s] %s", e.Type.String(), e.Message))

	// Add operation if available
	if e.Operation != "" {
		parts = append(parts, fmt.Sprintf("operation=%s", e.Operation))
	}

	// Add component if available
	if e.Component != "" {
		parts = append(parts, fmt.Sprintf("component=%s", e.Component))
	}

	// Add severity
	parts = append(parts, fmt.Sprintf("severity=%s", e.Severity.String()))

	// Add cause if available
	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("cause=%v", e.Cause))
	}

	return strings.Join(parts, " | ")
}

// DetailedError returns a detailed error message with context and stack trace
func (e *StateMachineError) DetailedError() string {
	var builder strings.Builder

	// Basic error information
	builder.WriteString(fmt.Sprintf("Error Type: %s\n", e.Type.String()))
	builder.WriteString(fmt.Sprintf("Message: %s\n", e.Message))
	builder.WriteString(fmt.Sprintf("Severity: %s\n", e.Severity.String()))
	builder.WriteString(fmt.Sprintf("Timestamp: %s\n", e.Timestamp.Format("2006-01-02 15:04:05.000")))
	builder.WriteString(fmt.Sprintf("Recoverable: %t\n", e.Recoverable))

	if e.Operation != "" {
		builder.WriteString(fmt.Sprintf("Operation: %s\n", e.Operation))
	}

	if e.Component != "" {
		builder.WriteString(fmt.Sprintf("Component: %s\n", e.Component))
	}

	// Context information
	if len(e.Context) > 0 {
		builder.WriteString("Context:\n")
		for key, value := range e.Context {
			builder.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	// Cause chain
	if e.Cause != nil {
		builder.WriteString(fmt.Sprintf("Caused by: %v\n", e.Cause))

		// If the cause is also a StateMachineError, show its details
		if smErr, ok := e.Cause.(*StateMachineError); ok {
			builder.WriteString("Cause details:\n")
			causeDetails := smErr.DetailedError()
			// Indent the cause details
			for _, line := range strings.Split(causeDetails, "\n") {
				if line != "" {
					builder.WriteString(fmt.Sprintf("  %s\n", line))
				}
			}
		}
	}

	// Stack trace
	if len(e.StackTrace) > 0 {
		builder.WriteString("Stack Trace:\n")
		for i, frame := range e.StackTrace {
			builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, frame.String()))
		}
	}

	return builder.String()
}

// Unwrap returns the underlying error
func (e *StateMachineError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target error type
func (e *StateMachineError) Is(target error) bool {
	if targetErr, ok := target.(*StateMachineError); ok {
		return e.Type == targetErr.Type
	}
	return false
}

// NewStateMachineError creates a new StateMachineError with default values
func NewStateMachineError(errorType ErrorType, message string, cause error) *StateMachineError {
	return &StateMachineError{
		Type:        errorType,
		Message:     message,
		Cause:       cause,
		Context:     make(map[string]interface{}),
		Severity:    ErrorSeverityMedium, // Default severity
		Timestamp:   time.Now(),
		StackTrace:  captureStackTrace(2), // Skip this function and the caller
		Recoverable: true,                 // Default to recoverable
	}
}

// NewStateMachineErrorWithSeverity creates a new StateMachineError with specified severity
func NewStateMachineErrorWithSeverity(errorType ErrorType, message string, cause error, severity ErrorSeverity) *StateMachineError {
	err := NewStateMachineError(errorType, message, cause)
	err.Severity = severity
	return err
}

// NewCriticalError creates a new critical StateMachineError
func NewCriticalError(errorType ErrorType, message string, cause error) *StateMachineError {
	err := NewStateMachineError(errorType, message, cause)
	err.Severity = ErrorSeverityCritical
	err.Recoverable = false
	return err
}

// WithContext adds context information to the error
func (e *StateMachineError) WithContext(key string, value interface{}) *StateMachineError {
	e.Context[key] = value
	return e
}

// WithContextMap adds multiple context entries to the error
func (e *StateMachineError) WithContextMap(context map[string]interface{}) *StateMachineError {
	for key, value := range context {
		e.Context[key] = value
	}
	return e
}

// WithOperation sets the operation that was being performed
func (e *StateMachineError) WithOperation(operation string) *StateMachineError {
	e.Operation = operation
	return e
}

// WithComponent sets the component where the error occurred
func (e *StateMachineError) WithComponent(component string) *StateMachineError {
	e.Component = component
	return e
}

// WithSeverity sets the error severity
func (e *StateMachineError) WithSeverity(severity ErrorSeverity) *StateMachineError {
	e.Severity = severity
	return e
}

// WithRecoverable sets whether the error is recoverable
func (e *StateMachineError) WithRecoverable(recoverable bool) *StateMachineError {
	e.Recoverable = recoverable
	return e
}

// captureStackTrace captures the current call stack
func captureStackTrace(skip int) []StackFrame {
	var frames []StackFrame

	// Capture up to 10 stack frames
	for i := skip; i < skip+10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		funcName := runtime.FuncForPC(pc).Name()

		// Clean up the function name
		if lastSlash := strings.LastIndex(funcName, "/"); lastSlash >= 0 {
			funcName = funcName[lastSlash+1:]
		}

		// Clean up the file path
		if lastSlash := strings.LastIndex(file, "/"); lastSlash >= 0 {
			file = file[lastSlash+1:]
		}
		if lastSlash := strings.LastIndex(file, "\\"); lastSlash >= 0 {
			file = file[lastSlash+1:]
		}

		frames = append(frames, StackFrame{
			Function: funcName,
			File:     file,
			Line:     line,
		})
	}

	return frames
}

// WrapError wraps an existing error with additional context
func WrapError(err error, errorType ErrorType, message string) *StateMachineError {
	if err == nil {
		return nil
	}

	// If the error is already a StateMachineError, preserve its context
	if smErr, ok := err.(*StateMachineError); ok {
		return &StateMachineError{
			Type:        errorType,
			Message:     message,
			Cause:       smErr,
			Context:     make(map[string]interface{}),
			Severity:    smErr.Severity, // Inherit severity from wrapped error
			Timestamp:   time.Now(),
			StackTrace:  captureStackTrace(2),
			Recoverable: smErr.Recoverable, // Inherit recoverability
		}
	}

	return NewStateMachineError(errorType, message, err)
}

// IsRecoverable checks if an error is recoverable
func IsRecoverable(err error) bool {
	if smErr, ok := err.(*StateMachineError); ok {
		return smErr.Recoverable
	}
	// Default to recoverable for non-StateMachineError errors
	return true
}

// GetErrorSeverity returns the severity of an error
func GetErrorSeverity(err error) ErrorSeverity {
	if smErr, ok := err.(*StateMachineError); ok {
		return smErr.Severity
	}
	// Default to medium severity for non-StateMachineError errors
	return ErrorSeverityMedium
}

// GetErrorType returns the type of an error
func GetErrorType(err error) ErrorType {
	if smErr, ok := err.(*StateMachineError); ok {
		return smErr.Type
	}
	// Default to file system error for unknown errors
	return ErrorTypeFileSystem
}
