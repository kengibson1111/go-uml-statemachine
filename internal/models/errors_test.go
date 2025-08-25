package models

import (
	"errors"
	"testing"
)

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		expected  string
	}{
		{
			name:      "file not found error",
			errorType: ErrorTypeFileNotFound,
			expected:  "file_not_found",
		},
		{
			name:      "validation error",
			errorType: ErrorTypeValidation,
			expected:  "validation",
		},
		{
			name:      "directory conflict error",
			errorType: ErrorTypeDirectoryConflict,
			expected:  "directory_conflict",
		},
		{
			name:      "reference resolution error",
			errorType: ErrorTypeReferenceResolution,
			expected:  "reference_resolution",
		},
		{
			name:      "file system error",
			errorType: ErrorTypeFileSystem,
			expected:  "file_system",
		},
		{
			name:      "version parsing error",
			errorType: ErrorTypeVersionParsing,
			expected:  "version_parsing",
		},
		{
			name:      "unknown error type",
			errorType: ErrorType(999),
			expected:  "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errorType.String()
			if result != tt.expected {
				t.Errorf("ErrorType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStateMachineError_Error(t *testing.T) {
	tests := []struct {
		name     string
		error    *StateMachineError
		expected string
	}{
		{
			name: "error without cause",
			error: &StateMachineError{
				Type:     ErrorTypeFileNotFound,
				Message:  "state-machine diagram not found",
				Cause:    nil,
				Severity: ErrorSeverityMedium, // Default severity
			},
			expected: "[file_not_found] state-machine diagram not found | severity=medium",
		},
		{
			name: "error with cause",
			error: &StateMachineError{
				Type:     ErrorTypeFileSystem,
				Message:  "failed to read file",
				Cause:    errors.New("permission denied"),
				Severity: ErrorSeverityMedium, // Default severity
			},
			expected: "[file_system] failed to read file | severity=medium | cause=permission denied",
		},
		{
			name: "validation error with cause",
			error: &StateMachineError{
				Type:     ErrorTypeValidation,
				Message:  "invalid PlantUML syntax",
				Cause:    errors.New("missing @enduml tag"),
				Severity: ErrorSeverityMedium, // Default severity
			},
			expected: "[validation] invalid PlantUML syntax | severity=medium | cause=missing @enduml tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.error.Error()
			if result != tt.expected {
				t.Errorf("StateMachineError.Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStateMachineError_Unwrap(t *testing.T) {
	originalError := errors.New("original error")
	diagError := &StateMachineError{
		Type:    ErrorTypeFileSystem,
		Message: "wrapper message",
		Cause:   originalError,
	}

	unwrapped := diagError.Unwrap()
	if unwrapped != originalError {
		t.Errorf("StateMachineError.Unwrap() = %v, want %v", unwrapped, originalError)
	}
}

func TestStateMachineError_UnwrapNil(t *testing.T) {
	diagError := &StateMachineError{
		Type:    ErrorTypeValidation,
		Message: "validation failed",
		Cause:   nil,
	}

	unwrapped := diagError.Unwrap()
	if unwrapped != nil {
		t.Errorf("StateMachineError.Unwrap() = %v, want nil", unwrapped)
	}
}

func TestNewStateMachineError(t *testing.T) {
	originalError := errors.New("original error")
	diagError := NewStateMachineError(ErrorTypeDirectoryConflict, "directory already exists", originalError)

	if diagError.Type != ErrorTypeDirectoryConflict {
		t.Errorf("NewStateMachineError().Type = %v, want %v", diagError.Type, ErrorTypeDirectoryConflict)
	}
	if diagError.Message != "directory already exists" {
		t.Errorf("NewStateMachineError().Message = %v, want %v", diagError.Message, "directory already exists")
	}
	if diagError.Cause != originalError {
		t.Errorf("NewStateMachineError().Cause = %v, want %v", diagError.Cause, originalError)
	}
	if diagError.Context == nil {
		t.Error("NewStateMachineError().Context should not be nil")
	}
	if len(diagError.Context) != 0 {
		t.Errorf("NewStateMachineError().Context length = %v, want 0", len(diagError.Context))
	}
}

func TestNewStateMachineError_NilCause(t *testing.T) {
	diagError := NewStateMachineError(ErrorTypeVersionParsing, "invalid version format", nil)

	if diagError.Type != ErrorTypeVersionParsing {
		t.Errorf("NewStateMachineError().Type = %v, want %v", diagError.Type, ErrorTypeVersionParsing)
	}
	if diagError.Message != "invalid version format" {
		t.Errorf("NewStateMachineError().Message = %v, want %v", diagError.Message, "invalid version format")
	}
	if diagError.Cause != nil {
		t.Errorf("NewStateMachineError().Cause = %v, want nil", diagError.Cause)
	}
	if diagError.Context == nil {
		t.Error("NewStateMachineError().Context should not be nil")
	}
}

func TestStateMachineError_WithContext(t *testing.T) {
	diagError := NewStateMachineError(ErrorTypeReferenceResolution, "reference not found", nil)

	// Add context information
	diagError.WithContext("reference_name", "user-auth")
	diagError.WithContext("reference_version", "1.0.0")
	diagError.WithContext("location", "products")

	if len(diagError.Context) != 3 {
		t.Errorf("StateMachineError.Context length = %v, want 3", len(diagError.Context))
	}

	if diagError.Context["reference_name"] != "user-auth" {
		t.Errorf("StateMachineError.Context[reference_name] = %v, want user-auth", diagError.Context["reference_name"])
	}
	if diagError.Context["reference_version"] != "1.0.0" {
		t.Errorf("StateMachineError.Context[reference_version] = %v, want 1.0.0", diagError.Context["reference_version"])
	}
	if diagError.Context["location"] != "products" {
		t.Errorf("StateMachineError.Context[location] = %v, want products", diagError.Context["location"])
	}
}

func TestStateMachineError_WithContext_Chaining(t *testing.T) {
	diagError := NewStateMachineError(ErrorTypeFileNotFound, "file not found", nil).
		WithContext("filename", "test.puml").
		WithContext("directory", "/tmp/test")

	if len(diagError.Context) != 2 {
		t.Errorf("StateMachineError.Context length = %v, want 2", len(diagError.Context))
	}

	if diagError.Context["filename"] != "test.puml" {
		t.Errorf("StateMachineError.Context[filename] = %v, want test.puml", diagError.Context["filename"])
	}
	if diagError.Context["directory"] != "/tmp/test" {
		t.Errorf("StateMachineError.Context[directory] = %v, want /tmp/test", diagError.Context["directory"])
	}
}

func TestStateMachineError_WithContext_OverwriteValue(t *testing.T) {
	diagError := NewStateMachineError(ErrorTypeValidation, "validation failed", nil)

	// Add initial context
	diagError.WithContext("severity", "warning")

	// Overwrite the same key
	diagError.WithContext("severity", "error")

	if diagError.Context["severity"] != "error" {
		t.Errorf("StateMachineError.Context[severity] = %v, want error", diagError.Context["severity"])
	}
}

func TestStateMachineError_IntegrationWithErrorsIs(t *testing.T) {
	originalError := errors.New("original error")
	diagError := NewStateMachineError(ErrorTypeFileSystem, "wrapper message", originalError)

	// Test that errors.Is works correctly
	if !errors.Is(diagError, originalError) {
		t.Error("errors.Is should return true for wrapped error")
	}

	// Test with different error
	differentError := errors.New("different error")
	if errors.Is(diagError, differentError) {
		t.Error("errors.Is should return false for different error")
	}
}

func TestStateMachineError_IntegrationWithErrorsAs(t *testing.T) {
	diagError := NewStateMachineError(ErrorTypeValidation, "validation failed", nil)

	var target *StateMachineError
	if !errors.As(diagError, &target) {
		t.Error("errors.As should return true for StateMachineError")
	}

	if target.Type != ErrorTypeValidation {
		t.Errorf("errors.As target.Type = %v, want %v", target.Type, ErrorTypeValidation)
	}
}
