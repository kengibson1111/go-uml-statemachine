package models

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestStateMachineError_EnhancedError(t *testing.T) {
	tests := []struct {
		name     string
		err      *StateMachineError
		expected string
	}{
		{
			name: "basic error without cause",
			err: &StateMachineError{
				Type:      ErrorTypeValidation,
				Message:   "test error",
				Severity:  ErrorSeverityMedium,
				Operation: "TestOp",
				Component: "TestComponent",
			},
			expected: "[validation] test error | operation=TestOp | component=TestComponent | severity=medium",
		},
		{
			name: "error with cause",
			err: &StateMachineError{
				Type:      ErrorTypeFileSystem,
				Message:   "file operation failed",
				Cause:     errors.New("permission denied"),
				Severity:  ErrorSeverityHigh,
				Operation: "WriteFile",
			},
			expected: "[file_system] file operation failed | operation=WriteFile | severity=high | cause=permission denied",
		},
		{
			name: "minimal error",
			err: &StateMachineError{
				Type:     ErrorTypeFileNotFound,
				Message:  "not found",
				Severity: ErrorSeverityLow,
			},
			expected: "[file_not_found] not found | severity=low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStateMachineError_DetailedError(t *testing.T) {
	err := &StateMachineError{
		Type:      ErrorTypeValidation,
		Message:   "validation failed",
		Severity:  ErrorSeverityHigh,
		Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		Operation: "Validate",
		Component: "Validator",
		Context: map[string]any{
			"name":    "test-diag",
			"version": "1.0.0",
		},
		Recoverable: true,
	}

	detailed := err.DetailedError()

	// Check that detailed error contains expected information
	expectedContains := []string{
		"Error Type: validation",
		"Message: validation failed",
		"Severity: high",
		"Timestamp: 2023-01-01 12:00:00.000",
		"Recoverable: true",
		"Operation: Validate",
		"Component: Validator",
		"Context:",
		"name: test-diag",
		"version: 1.0.0",
	}

	for _, expected := range expectedContains {
		if !contains(detailed, expected) {
			t.Errorf("DetailedError() missing expected content: %s", expected)
		}
	}
}

func TestStateMachineError_WithContextMap(t *testing.T) {
	err := NewStateMachineError(ErrorTypeValidation, "test", nil)

	contextMap := map[string]any{
		"name":    "test-diag",
		"version": "1.0.0",
		"count":   5,
	}

	result := err.WithContextMap(contextMap)

	for key, expectedValue := range contextMap {
		if result.Context[key] != expectedValue {
			t.Errorf("WithContextMap() %s = %v, want %v", key, result.Context[key], expectedValue)
		}
	}
}

func TestStateMachineError_WithOperation(t *testing.T) {
	err := NewStateMachineError(ErrorTypeValidation, "test", nil)

	result := err.WithOperation("TestOperation")

	if result.Operation != "TestOperation" {
		t.Errorf("WithOperation() = %v, want %v", result.Operation, "TestOperation")
	}
}

func TestStateMachineError_WithComponent(t *testing.T) {
	err := NewStateMachineError(ErrorTypeValidation, "test", nil)

	result := err.WithComponent("TestComponent")

	if result.Component != "TestComponent" {
		t.Errorf("WithComponent() = %v, want %v", result.Component, "TestComponent")
	}
}

func TestStateMachineError_WithSeverity(t *testing.T) {
	err := NewStateMachineError(ErrorTypeValidation, "test", nil)

	result := err.WithSeverity(ErrorSeverityCritical)

	if result.Severity != ErrorSeverityCritical {
		t.Errorf("WithSeverity() = %v, want %v", result.Severity, ErrorSeverityCritical)
	}
}

func TestStateMachineError_WithRecoverable(t *testing.T) {
	err := NewStateMachineError(ErrorTypeValidation, "test", nil)

	result := err.WithRecoverable(false)

	if result.Recoverable != false {
		t.Errorf("WithRecoverable() = %v, want %v", result.Recoverable, false)
	}
}

func TestStateMachineError_Is(t *testing.T) {
	err1 := NewStateMachineError(ErrorTypeValidation, "test1", nil)
	err2 := NewStateMachineError(ErrorTypeValidation, "test2", nil)
	err3 := NewStateMachineError(ErrorTypeFileSystem, "test3", nil)
	regularErr := errors.New("regular error")

	// Same error type should match
	if !err1.Is(err2) {
		t.Error("Is() should return true for same error type")
	}

	// Different error type should not match
	if err1.Is(err3) {
		t.Error("Is() should return false for different error type")
	}

	// Regular error should not match
	if err1.Is(regularErr) {
		t.Error("Is() should return false for regular error")
	}
}

func TestNewStateMachineErrorWithSeverity(t *testing.T) {
	err := NewStateMachineErrorWithSeverity(ErrorTypeValidation, "test", nil, ErrorSeverityCritical)

	if err.Severity != ErrorSeverityCritical {
		t.Errorf("NewStateMachineErrorWithSeverity() severity = %v, want %v", err.Severity, ErrorSeverityCritical)
	}
}

func TestNewCriticalError(t *testing.T) {
	err := NewCriticalError(ErrorTypeFileSystem, "critical error", nil)

	if err.Severity != ErrorSeverityCritical {
		t.Errorf("NewCriticalError() severity = %v, want %v", err.Severity, ErrorSeverityCritical)
	}

	if err.Recoverable != false {
		t.Errorf("NewCriticalError() recoverable = %v, want %v", err.Recoverable, false)
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := WrapError(originalErr, ErrorTypeFileSystem, "wrapped message")

	if wrappedErr == nil {
		t.Fatal("WrapError() returned nil")
	}

	if wrappedErr.Type != ErrorTypeFileSystem {
		t.Errorf("WrapError() type = %v, want %v", wrappedErr.Type, ErrorTypeFileSystem)
	}

	if wrappedErr.Message != "wrapped message" {
		t.Errorf("WrapError() message = %v, want %v", wrappedErr.Message, "wrapped message")
	}

	if wrappedErr.Cause != originalErr {
		t.Errorf("WrapError() cause = %v, want %v", wrappedErr.Cause, originalErr)
	}
}

func TestWrapError_WithStateMachineError(t *testing.T) {
	originalErr := NewStateMachineError(ErrorTypeValidation, "original", nil).
		WithSeverity(ErrorSeverityHigh).
		WithRecoverable(false)

	wrappedErr := WrapError(originalErr, ErrorTypeFileSystem, "wrapped message")

	if wrappedErr.Severity != ErrorSeverityHigh {
		t.Errorf("WrapError() should inherit severity, got %v, want %v", wrappedErr.Severity, ErrorSeverityHigh)
	}

	if wrappedErr.Recoverable != false {
		t.Errorf("WrapError() should inherit recoverability, got %v, want %v", wrappedErr.Recoverable, false)
	}
}

func TestWrapError_WithNilError(t *testing.T) {
	result := WrapError(nil, ErrorTypeFileSystem, "message")

	if result != nil {
		t.Errorf("WrapError() with nil error should return nil, got %v", result)
	}
}

func TestIsRecoverable(t *testing.T) {
	recoverableErr := NewStateMachineError(ErrorTypeValidation, "test", nil).WithRecoverable(true)
	nonRecoverableErr := NewStateMachineError(ErrorTypeValidation, "test", nil).WithRecoverable(false)
	regularErr := errors.New("regular error")

	if !IsRecoverable(recoverableErr) {
		t.Error("IsRecoverable() should return true for recoverable StateMachineError")
	}

	if IsRecoverable(nonRecoverableErr) {
		t.Error("IsRecoverable() should return false for non-recoverable StateMachineError")
	}

	if !IsRecoverable(regularErr) {
		t.Error("IsRecoverable() should return true for regular errors (default)")
	}
}

func TestGetErrorSeverity(t *testing.T) {
	highSeverityErr := NewStateMachineError(ErrorTypeValidation, "test", nil).WithSeverity(ErrorSeverityHigh)
	regularErr := errors.New("regular error")

	if GetErrorSeverity(highSeverityErr) != ErrorSeverityHigh {
		t.Errorf("GetErrorSeverity() = %v, want %v", GetErrorSeverity(highSeverityErr), ErrorSeverityHigh)
	}

	if GetErrorSeverity(regularErr) != ErrorSeverityMedium {
		t.Errorf("GetErrorSeverity() should return medium for regular errors, got %v", GetErrorSeverity(regularErr))
	}
}

func TestGetErrorType(t *testing.T) {
	validationErr := NewStateMachineError(ErrorTypeValidation, "test", nil)
	regularErr := errors.New("regular error")

	if GetErrorType(validationErr) != ErrorTypeValidation {
		t.Errorf("GetErrorType() = %v, want %v", GetErrorType(validationErr), ErrorTypeValidation)
	}

	if GetErrorType(regularErr) != ErrorTypeFileSystem {
		t.Errorf("GetErrorType() should return file_system for regular errors, got %v", GetErrorType(regularErr))
	}
}

func TestEnhancedErrorType_String(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		expected  string
	}{
		{ErrorTypeConfiguration, "configuration"},
		{ErrorTypePermission, "permission"},
		{ErrorTypeTimeout, "timeout"},
		{ErrorTypeNetwork, "network"},
		{ErrorTypeCorruption, "corruption"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.errorType.String()
			if result != tt.expected {
				t.Errorf("ErrorType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestErrorSeverity_String(t *testing.T) {
	tests := []struct {
		severity ErrorSeverity
		expected string
	}{
		{ErrorSeverityLow, "low"},
		{ErrorSeverityMedium, "medium"},
		{ErrorSeverityHigh, "high"},
		{ErrorSeverityCritical, "critical"},
		{ErrorSeverity(999), "unknown"}, // Unknown severity
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.severity.String()
			if result != tt.expected {
				t.Errorf("ErrorSeverity.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStackFrame_String(t *testing.T) {
	frame := StackFrame{
		Function: "TestFunction",
		File:     "test.go",
		Line:     42,
	}

	expected := "TestFunction (test.go:42)"
	result := frame.String()

	if result != expected {
		t.Errorf("StackFrame.String() = %v, want %v", result, expected)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
