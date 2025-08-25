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
	smError := &StateMachineError{
		Type:    ErrorTypeFileSystem,
		Message: "wrapper message",
		Cause:   originalError,
	}

	unwrapped := smError.Unwrap()
	if unwrapped != originalError {
		t.Errorf("StateMachineError.Unwrap() = %v, want %v", unwrapped, originalError)
	}
}

func TestStateMachineError_UnwrapNil(t *testing.T) {
	smError := &StateMachineError{
		Type:    ErrorTypeValidation,
		Message: "validation failed",
		Cause:   nil,
	}

	unwrapped := smError.Unwrap()
	if unwrapped != nil {
		t.Errorf("StateMachineError.Unwrap() = %v, want nil", unwrapped)
	}
}

func TestNewStateMachineError(t *testing.T) {
	originalError := errors.New("original error")
	smError := NewStateMachineError(ErrorTypeDirectoryConflict, "directory already exists", originalError)

	if smError.Type != ErrorTypeDirectoryConflict {
		t.Errorf("NewStateMachineError().Type = %v, want %v", smError.Type, ErrorTypeDirectoryConflict)
	}
	if smError.Message != "directory already exists" {
		t.Errorf("NewStateMachineError().Message = %v, want %v", smError.Message, "directory already exists")
	}
	if smError.Cause != originalError {
		t.Errorf("NewStateMachineError().Cause = %v, want %v", smError.Cause, originalError)
	}
	if smError.Context == nil {
		t.Error("NewStateMachineError().Context should not be nil")
	}
	if len(smError.Context) != 0 {
		t.Errorf("NewStateMachineError().Context length = %v, want 0", len(smError.Context))
	}
}

func TestNewStateMachineError_NilCause(t *testing.T) {
	smError := NewStateMachineError(ErrorTypeVersionParsing, "invalid version format", nil)

	if smError.Type != ErrorTypeVersionParsing {
		t.Errorf("NewStateMachineError().Type = %v, want %v", smError.Type, ErrorTypeVersionParsing)
	}
	if smError.Message != "invalid version format" {
		t.Errorf("NewStateMachineError().Message = %v, want %v", smError.Message, "invalid version format")
	}
	if smError.Cause != nil {
		t.Errorf("NewStateMachineError().Cause = %v, want nil", smError.Cause)
	}
	if smError.Context == nil {
		t.Error("NewStateMachineError().Context should not be nil")
	}
}

func TestStateMachineError_WithContext(t *testing.T) {
	smError := NewStateMachineError(ErrorTypeReferenceResolution, "reference not found", nil)

	// Add context information
	smError.WithContext("reference_name", "user-auth")
	smError.WithContext("reference_version", "1.0.0")
	smError.WithContext("location", "products")

	if len(smError.Context) != 3 {
		t.Errorf("StateMachineError.Context length = %v, want 3", len(smError.Context))
	}

	if smError.Context["reference_name"] != "user-auth" {
		t.Errorf("StateMachineError.Context[reference_name] = %v, want user-auth", smError.Context["reference_name"])
	}
	if smError.Context["reference_version"] != "1.0.0" {
		t.Errorf("StateMachineError.Context[reference_version] = %v, want 1.0.0", smError.Context["reference_version"])
	}
	if smError.Context["location"] != "products" {
		t.Errorf("StateMachineError.Context[location] = %v, want products", smError.Context["location"])
	}
}

func TestStateMachineError_WithContext_Chaining(t *testing.T) {
	smError := NewStateMachineError(ErrorTypeFileNotFound, "file not found", nil).
		WithContext("filename", "test.puml").
		WithContext("directory", "/tmp/test")

	if len(smError.Context) != 2 {
		t.Errorf("StateMachineError.Context length = %v, want 2", len(smError.Context))
	}

	if smError.Context["filename"] != "test.puml" {
		t.Errorf("StateMachineError.Context[filename] = %v, want test.puml", smError.Context["filename"])
	}
	if smError.Context["directory"] != "/tmp/test" {
		t.Errorf("StateMachineError.Context[directory] = %v, want /tmp/test", smError.Context["directory"])
	}
}

func TestStateMachineError_WithContext_OverwriteValue(t *testing.T) {
	smError := NewStateMachineError(ErrorTypeValidation, "validation failed", nil)

	// Add initial context
	smError.WithContext("severity", "warning")

	// Overwrite the same key
	smError.WithContext("severity", "error")

	if smError.Context["severity"] != "error" {
		t.Errorf("StateMachineError.Context[severity] = %v, want error", smError.Context["severity"])
	}
}

func TestStateMachineError_IntegrationWithErrorsIs(t *testing.T) {
	originalError := errors.New("original error")
	smError := NewStateMachineError(ErrorTypeFileSystem, "wrapper message", originalError)

	// Test that errors.Is works correctly
	if !errors.Is(smError, originalError) {
		t.Error("errors.Is should return true for wrapped error")
	}

	// Test with different error
	differentError := errors.New("different error")
	if errors.Is(smError, differentError) {
		t.Error("errors.Is should return false for different error")
	}
}

func TestStateMachineError_IntegrationWithErrorsAs(t *testing.T) {
	smError := NewStateMachineError(ErrorTypeValidation, "validation failed", nil)

	var target *StateMachineError
	if !errors.As(smError, &target) {
		t.Error("errors.As should return true for StateMachineError")
	}

	if target.Type != ErrorTypeValidation {
		t.Errorf("errors.As target.Type = %v, want %v", target.Type, ErrorTypeValidation)
	}
}
