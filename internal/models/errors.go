package models

import "fmt"

// ErrorType represents the type of error
type ErrorType int

const (
	ErrorTypeFileNotFound ErrorType = iota
	ErrorTypeValidation
	ErrorTypeDirectoryConflict
	ErrorTypeReferenceResolution
	ErrorTypeFileSystem
	ErrorTypeVersionParsing
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
	default:
		return "unknown"
	}
}

// StateMachineError represents a custom error with context
type StateMachineError struct {
	Type    ErrorType
	Message string
	Cause   error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *StateMachineError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type.String(), e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type.String(), e.Message)
}

// Unwrap returns the underlying error
func (e *StateMachineError) Unwrap() error {
	return e.Cause
}

// NewStateMachineError creates a new StateMachineError
func NewStateMachineError(errorType ErrorType, message string, cause error) *StateMachineError {
	return &StateMachineError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context information to the error
func (e *StateMachineError) WithContext(key string, value interface{}) *StateMachineError {
	e.Context[key] = value
	return e
}
