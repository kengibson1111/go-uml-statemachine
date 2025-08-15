package validation

import (
	"fmt"
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

func TestPlantUMLValidator_ValidateReferences_NoReferences(t *testing.T) {
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
		t.Error("ValidateReferences() should return valid result when no references")
	}

	if len(result.Errors) != 0 {
		t.Error("ValidateReferences() should return no errors when no references")
	}

	if len(sm.References) != 0 {
		t.Error("StateMachine should have no references")
	}
}

func TestPlantUMLValidator_ValidateReferences_ProductReference(t *testing.T) {
	validator := NewPlantUMLValidator()

	sm := &models.StateMachine{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
!include products/auth-service-1.2.0/auth-service-1.2.0.puml
[*] --> Idle
@enduml`,
	}

	result, err := validator.ValidateReferences(sm)
	if err != nil {
		t.Fatalf("ValidateReferences() error = %v", err)
	}

	if !result.IsValid {
		t.Error("ValidateReferences() should return valid result for valid product reference")
	}

	if len(result.Errors) != 0 {
		t.Errorf("ValidateReferences() should return no errors, got %d", len(result.Errors))
	}

	if len(sm.References) != 1 {
		t.Errorf("StateMachine should have 1 reference, got %d", len(sm.References))
	}

	ref := sm.References[0]
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

func TestPlantUMLValidator_ValidateReferences_NestedReference(t *testing.T) {
	validator := NewPlantUMLValidator()

	sm := &models.StateMachine{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
!include nested/sub-process/sub-process.puml
[*] --> Idle
@enduml`,
	}

	result, err := validator.ValidateReferences(sm)
	if err != nil {
		t.Fatalf("ValidateReferences() error = %v", err)
	}

	if !result.IsValid {
		t.Error("ValidateReferences() should return valid result for valid nested reference")
	}

	if len(result.Errors) != 0 {
		t.Errorf("ValidateReferences() should return no errors, got %d", len(result.Errors))
	}

	if len(sm.References) != 1 {
		t.Errorf("StateMachine should have 1 reference, got %d", len(sm.References))
	}

	ref := sm.References[0]
	if ref.Name != "sub-process" {
		t.Errorf("Reference name should be 'sub-process', got '%s'", ref.Name)
	}
	if ref.Version != "" {
		t.Errorf("Nested reference version should be empty, got '%s'", ref.Version)
	}
	if ref.Type != models.ReferenceTypeNested {
		t.Errorf("Reference type should be Nested, got %v", ref.Type)
	}
}

func TestPlantUMLValidator_ValidateReferences_InvalidProductVersion(t *testing.T) {
	validator := NewPlantUMLValidator()

	sm := &models.StateMachine{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
!include products/auth-service-invalid.version/auth-service-invalid.version.puml
[*] --> Idle
@enduml`,
	}

	result, err := validator.ValidateReferences(sm)
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

	sm := &models.StateMachine{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
!include products/test-1.0.0/test-1.0.0.puml
[*] --> Idle
@enduml`,
	}

	result, err := validator.ValidateReferences(sm)
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

	sm := &models.StateMachine{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
!include products/auth-service-1.2.0/auth-service-1.2.0.puml
!include nested/sub-process/sub-process.puml
!include products/payment-service-2.1.0/payment-service-2.1.0.puml
[*] --> Idle
@enduml`,
	}

	result, err := validator.ValidateReferences(sm)
	if err != nil {
		t.Fatalf("ValidateReferences() error = %v", err)
	}

	if !result.IsValid {
		t.Error("ValidateReferences() should return valid result for multiple valid references")
	}

	if len(result.Errors) != 0 {
		t.Errorf("ValidateReferences() should return no errors, got %d", len(result.Errors))
	}

	if len(sm.References) != 3 {
		t.Errorf("StateMachine should have 3 references, got %d", len(sm.References))
	}

	// Check that we have both product and nested references
	productCount := 0
	nestedCount := 0
	for _, ref := range sm.References {
		switch ref.Type {
		case models.ReferenceTypeProduct:
			productCount++
		case models.ReferenceTypeNested:
			nestedCount++
		}
	}

	if productCount != 2 {
		t.Errorf("Expected 2 product references, got %d", productCount)
	}
	if nestedCount != 1 {
		t.Errorf("Expected 1 nested reference, got %d", nestedCount)
	}
}

func TestPlantUMLValidator_ParseReferences_InvalidPatterns(t *testing.T) {
	validator := NewPlantUMLValidator()

	// Test content with invalid include patterns that should be ignored
	content := `@startuml
!include some/invalid/path.puml
!include products/invalid-name/file.puml
!include nested/invalid.puml
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
	stateMachines map[string]*models.StateMachine
	existsMap     map[string]bool
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		stateMachines: make(map[string]*models.StateMachine),
		existsMap:     make(map[string]bool),
	}
}

func (m *MockRepository) AddStateMachine(sm *models.StateMachine) {
	key := fmt.Sprintf("%s-%s-%s", sm.Name, sm.Version, sm.Location.String())
	m.stateMachines[key] = sm
	m.existsMap[key] = true
}

func (m *MockRepository) ReadStateMachine(name, version string, location models.Location) (*models.StateMachine, error) {
	key := fmt.Sprintf("%s-%s-%s", name, version, location.String())
	if sm, exists := m.stateMachines[key]; exists {
		return sm, nil
	}
	return nil, fmt.Errorf("state machine not found: %s", key)
}

func (m *MockRepository) Exists(name, version string, location models.Location) (bool, error) {
	key := fmt.Sprintf("%s-%s-%s", name, version, location.String())
	return m.existsMap[key], nil
}

// Implement other Repository methods as no-ops for testing
func (m *MockRepository) ListStateMachines(location models.Location) ([]models.StateMachine, error) {
	return nil, nil
}
func (m *MockRepository) WriteStateMachine(sm *models.StateMachine) error { return nil }
func (m *MockRepository) MoveStateMachine(name, version string, from, to models.Location) error {
	return nil
}
func (m *MockRepository) DeleteStateMachine(name, version string, location models.Location) error {
	return nil
}
func (m *MockRepository) CreateDirectory(path string) error         { return nil }
func (m *MockRepository) DirectoryExists(path string) (bool, error) { return false, nil }

func TestPlantUMLValidator_ResolveReferences_NoRepository(t *testing.T) {
	validator := NewPlantUMLValidator()

	sm := &models.StateMachine{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
!include products/auth-service-1.2.0/auth-service-1.2.0.puml
[*] --> Idle
@enduml`,
	}

	result, err := validator.ResolveReferences(sm)
	if err != nil {
		t.Fatalf("ResolveReferences() error = %v", err)
	}

	if !result.IsValid {
		t.Error("ResolveReferences() should return valid result when no repository")
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

func TestPlantUMLValidator_ResolveReferences_ExistingReference(t *testing.T) {
	mockRepo := NewMockRepository()
	validator := NewPlantUMLValidatorWithRepository(mockRepo)

	// Add a referenced state machine to the mock repository
	referencedSM := &models.StateMachine{
		Name:     "auth-service",
		Version:  "1.2.0",
		Location: models.LocationProducts,
		Content:  "@startuml\n[*] --> AuthIdle\n@enduml",
	}
	mockRepo.AddStateMachine(referencedSM)

	sm := &models.StateMachine{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
!include products/auth-service-1.2.0/auth-service-1.2.0.puml
[*] --> Idle
@enduml`,
	}

	result, err := validator.ResolveReferences(sm)
	if err != nil {
		t.Fatalf("ResolveReferences() error = %v", err)
	}

	if !result.IsValid {
		t.Error("ResolveReferences() should return valid result for existing reference")
	}

	if len(result.Errors) != 0 {
		t.Errorf("ResolveReferences() should return no errors, got %d", len(result.Errors))
	}
}

func TestPlantUMLValidator_ResolveReferences_MissingReference(t *testing.T) {
	mockRepo := NewMockRepository()
	validator := NewPlantUMLValidatorWithRepository(mockRepo)

	sm := &models.StateMachine{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
!include products/missing-service-1.0.0/missing-service-1.0.0.puml
[*] --> Idle
@enduml`,
	}

	result, err := validator.ResolveReferences(sm)
	if err != nil {
		t.Fatalf("ResolveReferences() error = %v", err)
	}

	if result.IsValid {
		t.Error("ResolveReferences() should return invalid result for missing reference")
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

func TestPlantUMLValidator_ResolveReferences_CircularReference(t *testing.T) {
	mockRepo := NewMockRepository()
	validator := NewPlantUMLValidatorWithRepository(mockRepo)

	// Create two state machines that reference each other
	sm1 := &models.StateMachine{
		Name:     "service-a",
		Version:  "1.0.0",
		Location: models.LocationProducts,
		Content: `@startuml
!include products/service-b-1.0.0/service-b-1.0.0.puml
[*] --> StateA
@enduml`,
	}

	sm2 := &models.StateMachine{
		Name:     "service-b",
		Version:  "1.0.0",
		Location: models.LocationProducts,
		Content: `@startuml
!include products/service-a-1.0.0/service-a-1.0.0.puml
[*] --> StateB
@enduml`,
	}

	mockRepo.AddStateMachine(sm1)
	mockRepo.AddStateMachine(sm2)

	result, err := validator.ResolveReferences(sm1)
	if err != nil {
		t.Fatalf("ResolveReferences() error = %v", err)
	}

	if result.IsValid {
		t.Error("ResolveReferences() should return invalid result for circular reference")
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

func TestPlantUMLValidator_ResolveReferences_NestedReference(t *testing.T) {
	mockRepo := NewMockRepository()
	validator := NewPlantUMLValidatorWithRepository(mockRepo)

	// Add a nested state machine to the mock repository
	nestedSM := &models.StateMachine{
		Name:     "sub-process",
		Version:  "",
		Location: models.LocationNested,
		Content:  "@startuml\n[*] --> SubState\n@enduml",
	}
	mockRepo.AddStateMachine(nestedSM)

	sm := &models.StateMachine{
		Name:    "test",
		Version: "1.0.0",
		Content: `@startuml
!include nested/sub-process/sub-process.puml
[*] --> Idle
@enduml`,
	}

	result, err := validator.ResolveReferences(sm)
	if err != nil {
		t.Fatalf("ResolveReferences() error = %v", err)
	}

	if !result.IsValid {
		t.Error("ResolveReferences() should return valid result for existing nested reference")
	}

	if len(result.Errors) != 0 {
		t.Errorf("ResolveReferences() should return no errors, got %d", len(result.Errors))
	}
}
