package models

import (
	"testing"
)

func TestValidationStrictness_String(t *testing.T) {
	tests := []struct {
		name       string
		strictness ValidationStrictness
		expected   string
	}{
		{
			name:       "in-progress strictness",
			strictness: StrictnessInProgress,
			expected:   "in-progress",
		},
		{
			name:       "products strictness",
			strictness: StrictnessProducts,
			expected:   "products",
		},
		{
			name:       "unknown strictness",
			strictness: ValidationStrictness(999),
			expected:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.strictness.String()
			if result != tt.expected {
				t.Errorf("ValidationStrictness.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidationError_Creation(t *testing.T) {
	context := make(map[string]any)
	context["rule"] = "missing_start_tag"
	context["expected"] = "@startuml"

	validationError := ValidationError{
		Code:     "MISSING_START_TAG",
		Message:  "PlantUML diagram must start with @startuml",
		Line:     1,
		Column:   1,
		Severity: "error",
		Context:  context,
	}

	if validationError.Code != "MISSING_START_TAG" {
		t.Errorf("ValidationError.Code = %v, want %v", validationError.Code, "MISSING_START_TAG")
	}
	if validationError.Message != "PlantUML diagram must start with @startuml" {
		t.Errorf("ValidationError.Message = %v, want %v", validationError.Message, "PlantUML diagram must start with @startuml")
	}
	if validationError.Line != 1 {
		t.Errorf("ValidationError.Line = %v, want %v", validationError.Line, 1)
	}
	if validationError.Column != 1 {
		t.Errorf("ValidationError.Column = %v, want %v", validationError.Column, 1)
	}
	if validationError.Severity != "error" {
		t.Errorf("ValidationError.Severity = %v, want %v", validationError.Severity, "error")
	}
	if validationError.Context["rule"] != "missing_start_tag" {
		t.Errorf("ValidationError.Context[rule] = %v, want %v", validationError.Context["rule"], "missing_start_tag")
	}
}

func TestValidationWarning_Creation(t *testing.T) {
	context := make(map[string]any)
	context["suggestion"] = "Consider adding state descriptions"

	validationWarning := ValidationWarning{
		Code:    "MISSING_DESCRIPTION",
		Message: "State lacks description",
		Line:    5,
		Column:  10,
		Context: context,
	}

	if validationWarning.Code != "MISSING_DESCRIPTION" {
		t.Errorf("ValidationWarning.Code = %v, want %v", validationWarning.Code, "MISSING_DESCRIPTION")
	}
	if validationWarning.Message != "State lacks description" {
		t.Errorf("ValidationWarning.Message = %v, want %v", validationWarning.Message, "State lacks description")
	}
	if validationWarning.Line != 5 {
		t.Errorf("ValidationWarning.Line = %v, want %v", validationWarning.Line, 5)
	}
	if validationWarning.Column != 10 {
		t.Errorf("ValidationWarning.Column = %v, want %v", validationWarning.Column, 10)
	}
	if validationWarning.Context["suggestion"] != "Consider adding state descriptions" {
		t.Errorf("ValidationWarning.Context[suggestion] = %v, want %v", validationWarning.Context["suggestion"], "Consider adding state descriptions")
	}
}

func TestValidationResult_HasErrors(t *testing.T) {
	tests := []struct {
		name     string
		result   ValidationResult
		expected bool
	}{
		{
			name: "no errors",
			result: ValidationResult{
				Errors:   []ValidationError{},
				Warnings: []ValidationWarning{},
				IsValid:  true,
			},
			expected: false,
		},
		{
			name: "has errors",
			result: ValidationResult{
				Errors: []ValidationError{
					{Code: "ERROR1", Message: "Error 1"},
				},
				Warnings: []ValidationWarning{},
				IsValid:  false,
			},
			expected: true,
		},
		{
			name: "multiple errors",
			result: ValidationResult{
				Errors: []ValidationError{
					{Code: "ERROR1", Message: "Error 1"},
					{Code: "ERROR2", Message: "Error 2"},
				},
				Warnings: []ValidationWarning{},
				IsValid:  false,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.result.HasErrors()
			if result != tt.expected {
				t.Errorf("ValidationResult.HasErrors() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidationResult_HasWarnings(t *testing.T) {
	tests := []struct {
		name     string
		result   ValidationResult
		expected bool
	}{
		{
			name: "no warnings",
			result: ValidationResult{
				Errors:   []ValidationError{},
				Warnings: []ValidationWarning{},
				IsValid:  true,
			},
			expected: false,
		},
		{
			name: "has warnings",
			result: ValidationResult{
				Errors: []ValidationError{},
				Warnings: []ValidationWarning{
					{Code: "WARNING1", Message: "Warning 1"},
				},
				IsValid: true,
			},
			expected: true,
		},
		{
			name: "multiple warnings",
			result: ValidationResult{
				Errors: []ValidationError{},
				Warnings: []ValidationWarning{
					{Code: "WARNING1", Message: "Warning 1"},
					{Code: "WARNING2", Message: "Warning 2"},
				},
				IsValid: true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.result.HasWarnings()
			if result != tt.expected {
				t.Errorf("ValidationResult.HasWarnings() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidationResult_AddError(t *testing.T) {
	result := &ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		IsValid:  true,
	}

	result.AddError("SYNTAX_ERROR", "Invalid syntax at line 3", 3, 15)

	if len(result.Errors) != 1 {
		t.Errorf("ValidationResult.Errors length = %v, want 1", len(result.Errors))
	}

	error := result.Errors[0]
	if error.Code != "SYNTAX_ERROR" {
		t.Errorf("ValidationError.Code = %v, want %v", error.Code, "SYNTAX_ERROR")
	}
	if error.Message != "Invalid syntax at line 3" {
		t.Errorf("ValidationError.Message = %v, want %v", error.Message, "Invalid syntax at line 3")
	}
	if error.Line != 3 {
		t.Errorf("ValidationError.Line = %v, want %v", error.Line, 3)
	}
	if error.Column != 15 {
		t.Errorf("ValidationError.Column = %v, want %v", error.Column, 15)
	}
	if error.Severity != "error" {
		t.Errorf("ValidationError.Severity = %v, want %v", error.Severity, "error")
	}
	if error.Context == nil {
		t.Error("ValidationError.Context should not be nil")
	}
	if result.IsValid {
		t.Error("ValidationResult.IsValid should be false after adding error")
	}
}

func TestValidationResult_AddWarning(t *testing.T) {
	result := &ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		IsValid:  true,
	}

	result.AddWarning("STYLE_WARNING", "Consider using more descriptive state names", 5, 8)

	if len(result.Warnings) != 1 {
		t.Errorf("ValidationResult.Warnings length = %v, want 1", len(result.Warnings))
	}

	warning := result.Warnings[0]
	if warning.Code != "STYLE_WARNING" {
		t.Errorf("ValidationWarning.Code = %v, want %v", warning.Code, "STYLE_WARNING")
	}
	if warning.Message != "Consider using more descriptive state names" {
		t.Errorf("ValidationWarning.Message = %v, want %v", warning.Message, "Consider using more descriptive state names")
	}
	if warning.Line != 5 {
		t.Errorf("ValidationWarning.Line = %v, want %v", warning.Line, 5)
	}
	if warning.Column != 8 {
		t.Errorf("ValidationWarning.Column = %v, want %v", warning.Column, 8)
	}
	if warning.Context == nil {
		t.Error("ValidationWarning.Context should not be nil")
	}
	if !result.IsValid {
		t.Error("ValidationResult.IsValid should remain true after adding warning")
	}
}

func TestValidationResult_MultipleErrorsAndWarnings(t *testing.T) {
	result := &ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		IsValid:  true,
	}

	// Add multiple errors
	result.AddError("ERROR1", "First error", 1, 1)
	result.AddError("ERROR2", "Second error", 2, 5)

	// Add multiple warnings
	result.AddWarning("WARNING1", "First warning", 3, 10)
	result.AddWarning("WARNING2", "Second warning", 4, 15)

	if len(result.Errors) != 2 {
		t.Errorf("ValidationResult.Errors length = %v, want 2", len(result.Errors))
	}
	if len(result.Warnings) != 2 {
		t.Errorf("ValidationResult.Warnings length = %v, want 2", len(result.Warnings))
	}
	if result.IsValid {
		t.Error("ValidationResult.IsValid should be false after adding errors")
	}
	if !result.HasErrors() {
		t.Error("ValidationResult.HasErrors() should return true")
	}
	if !result.HasWarnings() {
		t.Error("ValidationResult.HasWarnings() should return true")
	}
}

func TestValidationResult_InitialState(t *testing.T) {
	result := ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		IsValid:  true,
	}

	if result.HasErrors() {
		t.Error("ValidationResult.HasErrors() should return false for empty errors")
	}
	if result.HasWarnings() {
		t.Error("ValidationResult.HasWarnings() should return false for empty warnings")
	}
	if !result.IsValid {
		t.Error("ValidationResult.IsValid should be true initially")
	}
}

func TestValidationResult_NilSlices(t *testing.T) {
	result := ValidationResult{
		Errors:   nil,
		Warnings: nil,
		IsValid:  true,
	}

	if result.HasErrors() {
		t.Error("ValidationResult.HasErrors() should return false for nil errors slice")
	}
	if result.HasWarnings() {
		t.Error("ValidationResult.HasWarnings() should return false for nil warnings slice")
	}
}

func TestValidationResult_AddErrorSetsIsValidFalse(t *testing.T) {
	result := &ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		IsValid:  true,
	}

	// Verify initial state
	if !result.IsValid {
		t.Error("ValidationResult.IsValid should be true initially")
	}

	// Add error and verify IsValid becomes false
	result.AddError("TEST_ERROR", "Test error", 1, 1)

	if result.IsValid {
		t.Error("ValidationResult.IsValid should be false after adding error")
	}
}

func TestValidationResult_AddWarningKeepsIsValidTrue(t *testing.T) {
	result := &ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		IsValid:  true,
	}

	// Add warning and verify IsValid remains true
	result.AddWarning("TEST_WARNING", "Test warning", 1, 1)

	if !result.IsValid {
		t.Error("ValidationResult.IsValid should remain true after adding warning")
	}
}
