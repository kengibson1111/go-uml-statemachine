package validation

import (
	"testing"

	"github.com/kengibson1111/go-uml-statemachine/internal/models"
)

func TestNewPlantUMLValidator(t *testing.T) {
	validator := NewPlantUMLValidator()
	if validator == nil {
		t.Error("NewPlantUMLValidator() should return a non-nil validator")
	}
}

func TestPlantUMLValidator_ValidBasicStateMachine(t *testing.T) {
	validator := NewPlantUMLValidator()

	sm := &models.StateMachine{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
[*] --> Idle
Idle --> Active
Active --> [*]
@enduml`,
	}

	result, err := validator.Validate(sm, models.StrictnessInProgress)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if !result.IsValid {
		t.Error("Expected valid result for basic state machine")
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(result.Errors))
	}
}

func TestPlantUMLValidator_MissingStartTag(t *testing.T) {
	validator := NewPlantUMLValidator()

	sm := &models.StateMachine{
		Name:    "test",
		Version: "1.0.0",
		Content: `Idle --> Active
@enduml`,
	}

	result, err := validator.Validate(sm, models.StrictnessInProgress)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if result.IsValid {
		t.Error("Expected invalid result for missing start tag")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected errors for missing start tag")
	}

	// Check for specific error
	foundMissingStart := false
	for _, err := range result.Errors {
		if err.Code == "MISSING_START" {
			foundMissingStart = true
			break
		}
	}
	if !foundMissingStart {
		t.Error("Expected MISSING_START error")
	}
}

func TestPlantUMLValidator_MissingEndTag(t *testing.T) {
	validator := NewPlantUMLValidator()

	sm := &models.StateMachine{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
Idle --> Active`,
	}

	result, err := validator.Validate(sm, models.StrictnessInProgress)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if result.IsValid {
		t.Error("Expected invalid result for missing end tag")
	}

	// Check for specific error
	foundMissingEnd := false
	for _, err := range result.Errors {
		if err.Code == "MISSING_END" {
			foundMissingEnd = true
			break
		}
	}
	if !foundMissingEnd {
		t.Error("Expected MISSING_END error")
	}
}

func TestPlantUMLValidator_NoInitialStateWarning(t *testing.T) {
	validator := NewPlantUMLValidator()

	sm := &models.StateMachine{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
StateA --> StateB
@enduml`,
	}

	result, err := validator.Validate(sm, models.StrictnessInProgress)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if !result.IsValid {
		t.Error("Expected valid result (warnings don't make it invalid)")
	}

	// Check for specific warning
	foundNoInitialState := false
	for _, warn := range result.Warnings {
		if warn.Code == "NO_INITIAL_STATE" {
			foundNoInitialState = true
			break
		}
	}
	if !foundNoInitialState {
		t.Error("Expected NO_INITIAL_STATE warning")
	}
}

func TestPlantUMLValidator_ComplexValidStateMachine(t *testing.T) {
	validator := NewPlantUMLValidator()

	sm := &models.StateMachine{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
[*] --> Idle
Idle --> Processing : start
Processing --> Completed : success
Processing --> Failed : error
Completed --> [*]
Failed --> Idle : retry
@enduml`,
	}

	result, err := validator.Validate(sm, models.StrictnessInProgress)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if !result.IsValid {
		t.Error("Expected valid result for complex state machine")
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(result.Errors))
	}

	if len(result.Warnings) != 0 {
		t.Errorf("Expected no warnings, got %d", len(result.Warnings))
	}
}

func TestPlantUMLValidator_StrictnessLevels(t *testing.T) {
	validator := NewPlantUMLValidator()

	// Content with warnings but no critical errors
	sm := &models.StateMachine{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
StateA --> StateB
@enduml`,
	}

	// Test in-progress strictness
	result1, err := validator.Validate(sm, models.StrictnessInProgress)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if !result1.IsValid {
		t.Error("Expected valid result in in-progress mode (warnings don't invalidate)")
	}

	// Test products strictness
	result2, err := validator.Validate(sm, models.StrictnessProducts)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if !result2.IsValid {
		t.Error("Expected valid result in products mode")
	}
}

func TestPlantUMLValidator_ValidateReferences(t *testing.T) {
	validator := NewPlantUMLValidator()

	sm := &models.StateMachine{
		Name:    "test",
		Version: "1.0.0",
		Content: "@startuml\n[*] --> Idle\n@enduml",
	}

	result, err := validator.ValidateReferences(sm)
	if err != nil {
		t.Fatalf("ValidateReferences() error = %v", err)
	}

	if !result.IsValid {
		t.Error("ValidateReferences() should return valid result (placeholder)")
	}

	if len(result.Errors) != 0 {
		t.Error("ValidateReferences() should return no errors (placeholder)")
	}
}
