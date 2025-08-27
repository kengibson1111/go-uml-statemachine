package validation

import (
	"fmt"
	"strings"
	"testing"

	smmodels "github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/models"
)

func TestNewPlantUMLValidator(t *testing.T) {
	validator := NewPlantUMLValidator()
	if validator == nil {
		t.Error("NewPlantUMLValidator() should return a non-nil validator")
	}
}

func TestPlantUMLValidator_ValidBasicStateMachine(t *testing.T) {
	validator := NewPlantUMLValidator()

	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
[*] --> Idle
Idle --> Active
Active --> [*]
@enduml`,
	}

	result, err := validator.Validate(diag, models.StrictnessInProgress)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if !result.IsValid {
		t.Error("Expected valid result for basic state-machine diagram")
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(result.Errors))
	}
}

func TestPlantUMLValidator_MissingStartTag(t *testing.T) {
	validator := NewPlantUMLValidator()

	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: `Idle --> Active
@enduml`,
	}

	result, err := validator.Validate(diag, models.StrictnessInProgress)
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

	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
Idle --> Active`,
	}

	result, err := validator.Validate(diag, models.StrictnessInProgress)
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

	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
StateA --> StateB
@enduml`,
	}

	result, err := validator.Validate(diag, models.StrictnessInProgress)
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

	diag := &models.StateMachineDiagram{
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

	result, err := validator.Validate(diag, models.StrictnessInProgress)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if !result.IsValid {
		t.Error("Expected valid result for complex state-machine diagram")
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
	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
StateA --> StateB
@enduml`,
	}

	// Test in-progress strictness
	result1, err := validator.Validate(diag, models.StrictnessInProgress)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if !result1.IsValid {
		t.Error("Expected valid result in in-progress mode (warnings don't invalidate)")
	}

	// Test products strictness
	result2, err := validator.Validate(diag, models.StrictnessProducts)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if !result2.IsValid {
		t.Error("Expected valid result in products mode")
	}
}

func TestPlantUMLValidator_ValidateReferences_NoReferences(t *testing.T) {
	validator := NewPlantUMLValidator()

	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: "@startuml\n[*] --> Idle\n@enduml",
	}

	result, err := validator.ValidateReferences(diag)
	if err != nil {
		t.Fatalf("ValidateReferences() error = %v", err)
	}

	if !result.IsValid {
		t.Error("ValidateReferences() should return valid result when no references")
	}

	if len(result.Errors) != 0 {
		t.Error("ValidateReferences() should return no errors when no references")
	}

	if len(diag.References) != 0 {
		t.Error("StateMachineDiagram should have no references")
	}
}

func TestPlantUMLValidator_ValidateReferences_ProductReference(t *testing.T) {
	validator := NewPlantUMLValidator()

	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
!include products/auth-service-1.2.0/auth-service-1.2.0.puml
[*] --> Idle
@enduml`,
	}

	result, err := validator.ValidateReferences(diag)
	if err != nil {
		t.Fatalf("ValidateReferences() error = %v", err)
	}

	if !result.IsValid {
		t.Error("ValidateReferences() should return valid result for valid product reference")
	}

	if len(result.Errors) != 0 {
		t.Errorf("ValidateReferences() should return no errors, got %d", len(result.Errors))
	}

	if len(diag.References) != 1 {
		t.Errorf("StateMachineDiagram should have 1 reference, got %d", len(diag.References))
	}

	ref := diag.References[0]
	if ref.Name != "auth-service" {
		t.Errorf("Reference name should be 'auth-service', got '%s'", ref.Name)
	}
	if ref.Version != "1.2.0" {
		t.Errorf("Reference version should be '1.2.0', got '%s'", ref.Version)
	}
	if ref.Type != models.ReferenceTypeProduct {
		t.Errorf("Reference type should be Product, got %v", ref.Type)
	}
}

func TestPlantUMLValidator_ValidateReferences_InvalidProductVersion(t *testing.T) {
	validator := NewPlantUMLValidator()

	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
!include products/auth-service-invalid.version/auth-service-invalid.version.puml
[*] --> Idle
@enduml`,
	}

	result, err := validator.ValidateReferences(diag)
	if err != nil {
		t.Fatalf("ValidateReferences() error = %v", err)
	}

	if result.IsValid {
		t.Error("ValidateReferences() should return invalid result for invalid version")
	}

	if len(result.Errors) == 0 {
		t.Error("ValidateReferences() should return errors for invalid version")
	}

	// Check for specific error
	foundInvalidVersion := false
	for _, err := range result.Errors {
		if err.Code == "INVALID_REFERENCE_VERSION" {
			foundInvalidVersion = true
			break
		}
	}
	if !foundInvalidVersion {
		t.Error("Expected INVALID_REFERENCE_VERSION error")
	}
}

func TestPlantUMLValidator_ValidateReferences_SelfReference(t *testing.T) {
	validator := NewPlantUMLValidator()

	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
!include products/test-1.0.0/test-1.0.0.puml
[*] --> Idle
@enduml`,
	}

	result, err := validator.ValidateReferences(diag)
	if err != nil {
		t.Fatalf("ValidateReferences() error = %v", err)
	}

	if result.IsValid {
		t.Error("ValidateReferences() should return invalid result for self reference")
	}

	// Check for specific error
	foundSelfReference := false
	for _, err := range result.Errors {
		if err.Code == "SELF_REFERENCE" {
			foundSelfReference = true
			break
		}
	}
	if !foundSelfReference {
		t.Error("Expected SELF_REFERENCE error")
	}
}

func TestPlantUMLValidator_ValidateReferences_MultipleReferences(t *testing.T) {
	validator := NewPlantUMLValidator()

	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
!include products/auth-service-1.2.0/auth-service-1.2.0.puml
!include products/payment-service-2.1.0/payment-service-2.1.0.puml
[*] --> Idle
@enduml`,
	}

	result, err := validator.ValidateReferences(diag)
	if err != nil {
		t.Fatalf("ValidateReferences() error = %v", err)
	}

	if !result.IsValid {
		t.Error("ValidateReferences() should return valid result for multiple valid references")
	}

	if len(result.Errors) != 0 {
		t.Errorf("ValidateReferences() should return no errors, got %d", len(result.Errors))
	}

	if len(diag.References) != 2 {
		t.Errorf("StateMachineDiagram should have 2 references, got %d", len(diag.References))
	}

	// Check that we have only product references
	productCount := 0
	for _, ref := range diag.References {
		switch ref.Type {
		case models.ReferenceTypeProduct:
			productCount++
		}
	}

	if productCount != 2 {
		t.Errorf("Expected 2 product references, got %d", productCount)
	}
}

func TestPlantUMLValidator_ParseReferences_InvalidPatterns(t *testing.T) {
	validator := NewPlantUMLValidator()

	// Test content with invalid include patterns that should be ignored
	content := `@startuml
!include some/invalid/path.puml
!include products/invalid-name/file.puml
[*] --> Idle
@enduml`

	references, err := validator.parseReferences(content)
	if err != nil {
		t.Fatalf("parseReferences() error = %v", err)
	}

	// Should not parse any references from invalid patterns
	if len(references) != 0 {
		t.Errorf("Expected 0 references from invalid patterns, got %d", len(references))
	}
}

func TestPlantUMLValidator_IsValidVersion(t *testing.T) {
	validator := NewPlantUMLValidator()

	validVersions := []string{
		"1.0.0",
		"10.20.30",
		"1.0.0-alpha",
		"1.0.0-beta.1",
		"2.1.0-rc.1",
	}

	invalidVersions := []string{
		"1.0",
		"1",
		"1.0.0.0",
		"v1.0.0",
		"1.0.0-",
		"invalid.version",
		"",
	}

	for _, version := range validVersions {
		if !validator.isValidVersion(version) {
			t.Errorf("Version '%s' should be valid", version)
		}
	}

	for _, version := range invalidVersions {
		if validator.isValidVersion(version) {
			t.Errorf("Version '%s' should be invalid", version)
		}
	}
}

// MockRepository is a mock implementation of Repository for testing
type MockRepository struct {
	diagrams  map[string]*models.StateMachineDiagram
	existsMap map[string]bool
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		diagrams:  make(map[string]*models.StateMachineDiagram),
		existsMap: make(map[string]bool),
	}
}

func (m *MockRepository) AddStateMachine(diag *models.StateMachineDiagram) {
	key := fmt.Sprintf("%s-%s-%s-%s", diag.DiagramType, diag.Name, diag.Version, diag.Location.String())
	m.diagrams[key] = diag
	m.existsMap[key] = true
}

func (m *MockRepository) ReadDiagram(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
	key := fmt.Sprintf("%s-%s-%s-%s", diagramType, name, version, location.String())
	if diag, exists := m.diagrams[key]; exists {
		return diag, nil
	}
	return nil, fmt.Errorf("state-machine diagram not found: %s", key)
}

func (m *MockRepository) Exists(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
	key := fmt.Sprintf("%s-%s-%s-%s", diagramType, name, version, location.String())
	return m.existsMap[key], nil
}

// Implement other Repository methods as no-ops for testing
func (m *MockRepository) ListStateMachines(diagramType smmodels.DiagramType, location models.Location) ([]models.StateMachineDiagram, error) {
	return nil, nil
}
func (m *MockRepository) WriteDiagram(diag *models.StateMachineDiagram) error { return nil }
func (m *MockRepository) MoveDiagram(diagramType smmodels.DiagramType, name, version string, from, to models.Location) error {
	return nil
}
func (m *MockRepository) DeleteDiagram(diagramType smmodels.DiagramType, name, version string, location models.Location) error {
	return nil
}
func (m *MockRepository) CreateDirectory(path string) error         { return nil }
func (m *MockRepository) DirectoryExists(path string) (bool, error) { return false, nil }

func TestPlantUMLValidator_ResolveFileReferences_NoRepository(t *testing.T) {
	validator := NewPlantUMLValidator()

	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
!include products/auth-service-1.2.0/auth-service-1.2.0.puml
[*] --> Idle
@enduml`,
	}

	result, err := validator.ResolveFileReferences(diag)
	if err != nil {
		t.Fatalf("ResolveFileReferences() error = %v", err)
	}

	if !result.IsValid {
		t.Error("ResolveFileReferences() should return valid result when no repository")
	}

	// Should have a warning about no repository
	foundNoRepository := false
	for _, warn := range result.Warnings {
		if warn.Code == "NO_REPOSITORY" {
			foundNoRepository = true
			break
		}
	}
	if !foundNoRepository {
		t.Error("Expected NO_REPOSITORY warning")
	}
}

func TestPlantUMLValidator_ResolveFileReferences_ExistingReference(t *testing.T) {
	mockRepo := NewMockRepository()
	validator := NewPlantUMLValidatorWithRepository(mockRepo)

	// Add a referenced state-machine diagram to the mock repository
	referencedDiag := &models.StateMachineDiagram{
		Name:     "auth-service",
		Version:  "1.2.0",
		Location: models.LocationProducts,
		Content:  "@startuml\n[*] --> AuthIdle\n@enduml",
	}
	mockRepo.AddStateMachine(referencedDiag)

	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
!include products/auth-service-1.2.0/auth-service-1.2.0.puml
[*] --> Idle
@enduml`,
	}

	result, err := validator.ResolveFileReferences(diag)
	if err != nil {
		t.Fatalf("ResolveFileReferences() error = %v", err)
	}

	if !result.IsValid {
		t.Error("ResolveFileReferences() should return valid result for existing reference")
	}

	if len(result.Errors) != 0 {
		t.Errorf("ResolveFileReferences() should return no errors, got %d", len(result.Errors))
	}
}

func TestPlantUMLValidator_ResolveFileReferences_MissingReference(t *testing.T) {
	mockRepo := NewMockRepository()
	validator := NewPlantUMLValidatorWithRepository(mockRepo)

	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
!include products/missing-service-1.0.0/missing-service-1.0.0.puml
[*] --> Idle
@enduml`,
	}

	result, err := validator.ResolveFileReferences(diag)
	if err != nil {
		t.Fatalf("ResolveFileReferences() error = %v", err)
	}

	if result.IsValid {
		t.Error("ResolveFileReferences() should return invalid result for missing reference")
	}

	// Check for specific error
	foundMissingReference := false
	for _, err := range result.Errors {
		if err.Code == "PRODUCT_REFERENCE_NOT_FOUND" {
			foundMissingReference = true
			break
		}
	}
	if !foundMissingReference {
		t.Error("Expected PRODUCT_REFERENCE_NOT_FOUND error")
	}
}

func TestPlantUMLValidator_ResolveFileReferences_CircularReference(t *testing.T) {
	mockRepo := NewMockRepository()
	validator := NewPlantUMLValidatorWithRepository(mockRepo)

	// Create two state-machine diagrams that reference each other
	diag1 := &models.StateMachineDiagram{
		Name:     "service-a",
		Version:  "1.0.0",
		Location: models.LocationProducts,
		Content: `@startuml
!include products/service-b-1.0.0/service-b-1.0.0.puml
[*] --> StateA
@enduml`,
	}

	diag2 := &models.StateMachineDiagram{
		Name:     "service-b",
		Version:  "1.0.0",
		Location: models.LocationProducts,
		Content: `@startuml
!include products/service-a-1.0.0/service-a-1.0.0.puml
[*] --> StateB
@enduml`,
	}

	mockRepo.AddStateMachine(diag1)
	mockRepo.AddStateMachine(diag2)

	result, err := validator.ResolveFileReferences(diag1)
	if err != nil {
		t.Fatalf("ResolveFileReferences() error = %v", err)
	}

	if result.IsValid {
		t.Error("ResolveFileReferences() should return invalid result for circular reference")
	}

	// Check for circular reference error
	foundCircularReference := false
	for _, err := range result.Errors {
		if err.Code == "CIRCULAR_REFERENCE" || err.Code == "DIRECT_CIRCULAR_REFERENCE" {
			foundCircularReference = true
			break
		}
	}
	if !foundCircularReference {
		t.Error("Expected CIRCULAR_REFERENCE or DIRECT_CIRCULAR_REFERENCE error")
	}
}

// TestPlantUMLValidator_StrictnessInProgress_ErrorsAndWarnings tests that in-progress validation returns both errors and warnings
func TestPlantUMLValidator_StrictnessInProgress_ErrorsAndWarnings(t *testing.T) {
	validator := NewPlantUMLValidator()

	// Content with both errors and warnings
	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
StateA --> StateB
invalid_state_name! --> StateC
@enduml`,
	}

	result, err := validator.Validate(diag, models.StrictnessInProgress)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Should be valid because warnings don't invalidate, and invalid state names are warnings
	if !result.IsValid {
		t.Error("Expected valid result in in-progress mode (warnings don't invalidate)")
	}

	// Should have warnings for missing initial state
	if len(result.Warnings) == 0 {
		t.Error("Expected warnings in in-progress mode")
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
		t.Error("Expected NO_INITIAL_STATE warning in in-progress mode")
	}
}

// TestPlantUMLValidator_StrictnessProducts_WarningsOnly tests that products validation converts non-critical errors to warnings
func TestPlantUMLValidator_StrictnessProducts_WarningsOnly(t *testing.T) {
	validator := NewPlantUMLValidator()

	// Content with non-critical validation issues
	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
StateA --> StateB
invalid_state_name! --> StateC
@enduml`,
	}

	result, err := validator.Validate(diag, models.StrictnessProducts)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Should be valid in products mode (non-critical errors converted to warnings)
	if !result.IsValid {
		t.Error("Expected valid result in products mode")
	}

	// Should have warnings (including converted errors)
	if len(result.Warnings) == 0 {
		t.Error("Expected warnings in products mode")
	}

	// Should have no errors (all non-critical errors converted to warnings)
	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors in products mode, got %d", len(result.Errors))
	}
}

// TestPlantUMLValidator_StrictnessProducts_CriticalErrors tests that critical errors remain as errors in products mode
func TestPlantUMLValidator_StrictnessProducts_CriticalErrors(t *testing.T) {
	validator := NewPlantUMLValidator()

	// Content with critical structural errors
	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: `StateA --> StateB`, // Missing @startuml and @enduml tags
	}

	result, err := validator.Validate(diag, models.StrictnessProducts)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Should be invalid due to critical errors
	if result.IsValid {
		t.Error("Expected invalid result in products mode due to critical errors")
	}

	// Should have critical errors that weren't converted to warnings
	if len(result.Errors) == 0 {
		t.Error("Expected critical errors to remain as errors in products mode")
	}

	// Check for specific critical errors
	foundMissingStart := false
	foundMissingEnd := false
	for _, err := range result.Errors {
		if err.Code == "MISSING_START" {
			foundMissingStart = true
		}
		if err.Code == "MISSING_END" {
			foundMissingEnd = true
		}
	}
	if !foundMissingStart {
		t.Error("Expected MISSING_START critical error to remain in products mode")
	}
	if !foundMissingEnd {
		t.Error("Expected MISSING_END critical error to remain in products mode")
	}
}

// TestPlantUMLValidator_StrictnessComparison tests the difference between strictness levels
func TestPlantUMLValidator_StrictnessComparison(t *testing.T) {
	validator := NewPlantUMLValidator()

	// Content with non-critical issues that would be errors in in-progress but warnings in products
	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
StateA --> StateB
@enduml`,
	}

	// Test in-progress strictness
	resultInProgress, err := validator.Validate(diag, models.StrictnessInProgress)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Test products strictness
	resultProducts, err := validator.Validate(diag, models.StrictnessProducts)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Both should be valid (warnings don't invalidate)
	if !resultInProgress.IsValid {
		t.Error("Expected valid result in in-progress mode")
	}
	if !resultProducts.IsValid {
		t.Error("Expected valid result in products mode")
	}

	// Both should have warnings for missing initial state
	if len(resultInProgress.Warnings) == 0 {
		t.Error("Expected warnings in in-progress mode")
	}
	if len(resultProducts.Warnings) == 0 {
		t.Error("Expected warnings in products mode")
	}

	// Products mode might have more warnings due to error conversion
	// but in this case, there are no non-critical errors to convert
	if len(resultProducts.Warnings) < len(resultInProgress.Warnings) {
		t.Error("Products mode should have at least as many warnings as in-progress mode")
	}
}

// TestPlantUMLValidator_StrictnessWithReferences tests strictness levels with reference validation
func TestPlantUMLValidator_StrictnessWithReferences(t *testing.T) {
	validator := NewPlantUMLValidator()

	// Content with invalid reference version (non-critical error)
	diag := &models.StateMachineDiagram{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
!include products/auth-service-invalid.version/auth-service-invalid.version.puml
[*] --> Idle
@enduml`,
	}

	// Test in-progress strictness with reference validation
	resultInProgress, err := validator.ValidateReferences(diag)
	if err != nil {
		t.Fatalf("ValidateReferences() error = %v", err)
	}

	// Apply strictness filtering manually for reference validation
	validator.applyStrictnessFiltering(resultInProgress, models.StrictnessInProgress)

	// Test products strictness with reference validation
	resultProducts, err := validator.ValidateReferences(diag)
	if err != nil {
		t.Fatalf("ValidateReferences() error = %v", err)
	}

	// Apply strictness filtering manually for reference validation
	validator.applyStrictnessFiltering(resultProducts, models.StrictnessProducts)

	// In-progress should be invalid due to reference errors
	if resultInProgress.IsValid {
		t.Error("Expected invalid result in in-progress mode due to reference errors")
	}

	// Products should convert non-critical reference errors to warnings
	// INVALID_REFERENCE_VERSION is not critical, so it should be converted
	if !resultProducts.IsValid {
		// Check if all errors are critical
		allCritical := true
		for _, err := range resultProducts.Errors {
			if !validator.isCriticalError(err.Code) {
				allCritical = false
				break
			}
		}
		if !allCritical {
			t.Error("Expected valid result in products mode (non-critical errors should be converted to warnings)")
		}
	}
}

// TestPlantUMLValidator_StrictnessWithCircularReference tests strictness with critical reference errors
func TestPlantUMLValidator_StrictnessWithCircularReference(t *testing.T) {
	mockRepo := NewMockRepository()
	validator := NewPlantUMLValidatorWithRepository(mockRepo)

	// Create a state-machine diagram that references itself (critical error)
	diag := &models.StateMachineDiagram{
		Name:     "test",
		Version:  "1.0.0",
		Location: models.LocationProducts,
		Content: `@startuml
!include products/test-1.0.0/test-1.0.0.puml
[*] --> Idle
@enduml`,
	}

	mockRepo.AddStateMachine(diag)

	// Test both strictness levels
	resultInProgress, err := validator.ResolveFileReferences(diag)
	if err != nil {
		t.Fatalf("ResolveFileReferences() error = %v", err)
	}
	validator.applyStrictnessFiltering(resultInProgress, models.StrictnessInProgress)

	resultProducts, err := validator.ResolveFileReferences(diag)
	if err != nil {
		t.Fatalf("ResolveFileReferences() error = %v", err)
	}
	validator.applyStrictnessFiltering(resultProducts, models.StrictnessProducts)

	// Both should be invalid due to critical self-reference error
	if resultInProgress.IsValid {
		t.Error("Expected invalid result in in-progress mode due to self-reference")
	}
	if resultProducts.IsValid {
		t.Error("Expected invalid result in products mode due to critical self-reference error")
	}

	// Both should have the same critical error
	if len(resultInProgress.Errors) == 0 {
		t.Error("Expected errors in in-progress mode")
	}
	if len(resultProducts.Errors) == 0 {
		t.Error("Expected critical errors to remain in products mode")
	}
}

// TestPlantUMLValidator_StrictnessErrorConversion tests that error conversion includes proper messaging
func TestPlantUMLValidator_StrictnessErrorConversion(t *testing.T) {
	validator := NewPlantUMLValidator()

	// Create a validation result with non-critical errors
	result := &models.ValidationResult{
		Errors:   []models.ValidationError{},
		Warnings: []models.ValidationWarning{},
		IsValid:  true,
	}

	// Add a non-critical error
	result.AddError("INVALID_REFERENCE_VERSION", "Invalid version format", 1, 1)
	result.AddError("MISSING_START", "Missing @startuml tag", 1, 1) // Critical error

	// Apply products strictness
	validator.applyStrictnessFiltering(result, models.StrictnessProducts)

	// Should have one critical error remaining
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 critical error, got %d", len(result.Errors))
	}

	// Should have one converted warning
	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 converted warning, got %d", len(result.Warnings))
	}

	// Check that the warning has the conversion message
	if len(result.Warnings) > 0 {
		warning := result.Warnings[0]
		if warning.Code != "INVALID_REFERENCE_VERSION" {
			t.Errorf("Expected converted warning code 'INVALID_REFERENCE_VERSION', got '%s'", warning.Code)
		}
		if !strings.Contains(warning.Message, "(Converted from error)") {
			t.Error("Expected converted warning to have conversion message")
		}
	}

	// Check that the critical error remains
	if len(result.Errors) > 0 {
		err := result.Errors[0]
		if err.Code != "MISSING_START" {
			t.Errorf("Expected critical error code 'MISSING_START', got '%s'", err.Code)
		}
	}
}

// TestPlantUMLValidator_StrictnessUnknownLevel tests behavior with unknown strictness level
func TestPlantUMLValidator_StrictnessUnknownLevel(t *testing.T) {
	validator := NewPlantUMLValidator()

	// Create a validation result with errors
	result := &models.ValidationResult{
		Errors:   []models.ValidationError{},
		Warnings: []models.ValidationWarning{},
		IsValid:  true,
	}

	result.AddError("SOME_ERROR", "Some error message", 1, 1)

	// Apply unknown strictness level (should default to in-progress behavior)
	unknownStrictness := models.ValidationStrictness(999)
	validator.applyStrictnessFiltering(result, unknownStrictness)

	// Should behave like in-progress mode (keep errors as errors)
	if result.IsValid {
		t.Error("Expected invalid result for unknown strictness level (should default to in-progress)")
	}

	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error for unknown strictness level, got %d", len(result.Errors))
	}
}
