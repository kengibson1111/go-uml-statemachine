package service

import (
	"errors"
	"testing"

	smmodels "github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/models"
)

func TestService_PromoteToProductsFile_Success(t *testing.T) {
	repo := &mockRepository{}
	validator := &mockValidator{}

	// Track the state changes through the promotion process
	var diagramState string = "in-progress"

	// Setup successful promotion scenario
	repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
		if location == models.LocationInProgress {
			return diagramState == "in-progress", nil
		}
		if location == models.LocationProducts {
			return diagramState == "products", nil
		}
		return false, nil
	}

	repo.readStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
		return &models.StateMachineDiagram{
			Name:     name,
			Version:  version,
			Content:  "@startuml\n[*] --> Idle\n@enduml",
			Location: location,
		}, nil
	}

	validator.validateFunc = func(diag *models.StateMachineDiagram, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
		return &models.ValidationResult{
			IsValid:  true,
			Errors:   []models.ValidationError{},
			Warnings: []models.ValidationWarning{},
		}, nil
	}

	repo.moveStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, from, to models.Location) error {
		if from == models.LocationInProgress && to == models.LocationProducts {
			diagramState = "products" // Simulate successful move
		}
		return nil
	}

	svc := NewService(repo, validator, nil)

	err := svc.PromoteToProductsFile(smmodels.DiagramTypePUML, "test-diag", "1.0.0")

	if err != nil {
		t.Errorf("PromoteToProductsFile() unexpected error: %v", err)
	}
}

func TestService_PromoteToProductsFile_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		inputVer    string
		wantErrType models.ErrorType
	}{
		{
			name:        "empty name validation",
			inputName:   "",
			inputVer:    "1.0.0",
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name:        "empty version validation",
			inputName:   "test-diag",
			inputVer:    "",
			wantErrType: models.ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			validator := &mockValidator{}
			svc := NewService(repo, validator, nil)

			err := svc.PromoteToProductsFile(smmodels.DiagramTypePUML, tt.inputName, tt.inputVer)

			if err == nil {
				t.Errorf("PromoteToProductsFile() expected error but got none")
				return
			}

			var diagErr *models.StateMachineError
			if !errors.As(err, &diagErr) {
				t.Errorf("PromoteToProductsFile() expected StateMachineError but got %T", err)
				return
			}

			if diagErr.Type != tt.wantErrType {
				t.Errorf("PromoteToProductsFile() expected error type %v but got %v", tt.wantErrType, diagErr.Type)
			}
		})
	}
}

func TestService_PromoteToProductsFile_FileNotFound(t *testing.T) {
	repo := &mockRepository{}
	validator := &mockValidator{}

	// Diagram doesn't exist in in-progress
	repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
		return false, nil
	}

	svc := NewService(repo, validator, nil)

	err := svc.PromoteToProductsFile(smmodels.DiagramTypePUML, "test-diag", "1.0.0")

	if err == nil {
		t.Error("PromoteToProductsFile() expected error but got none")
		return
	}

	var diagErr *models.StateMachineError
	if !errors.As(err, &diagErr) {
		t.Errorf("PromoteToProductsFile() expected StateMachineError but got %T", err)
		return
	}

	if diagErr.Type != models.ErrorTypeFileNotFound {
		t.Errorf("PromoteToProductsFile() expected error type %v but got %v", models.ErrorTypeFileNotFound, diagErr.Type)
	}
}

func TestService_PromoteToProductsFile_DirectoryConflict(t *testing.T) {
	repo := &mockRepository{}
	validator := &mockValidator{}

	// Diagram exists in both locations - conflict
	repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
		return true, nil // exists in both locations
	}

	svc := NewService(repo, validator, nil)

	err := svc.PromoteToProductsFile(smmodels.DiagramTypePUML, "test-diag", "1.0.0")

	if err == nil {
		t.Error("PromoteToProductsFile() expected error but got none")
		return
	}

	var diagErr *models.StateMachineError
	if !errors.As(err, &diagErr) {
		t.Errorf("PromoteToProductsFile() expected StateMachineError but got %T", err)
		return
	}

	if diagErr.Type != models.ErrorTypeDirectoryConflict {
		t.Errorf("PromoteToProductsFile() expected error type %v but got %v", models.ErrorTypeDirectoryConflict, diagErr.Type)
	}
}

func TestService_PromoteToProductsFile_ValidationFails(t *testing.T) {
	repo := &mockRepository{}
	validator := &mockValidator{}

	// Setup diagram exists in in-progress but validation fails
	repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
		if location == models.LocationInProgress {
			return true, nil
		}
		return false, nil
	}

	repo.readStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
		return &models.StateMachineDiagram{
			Name:     name,
			Version:  version,
			Content:  "invalid plantuml",
			Location: location,
		}, nil
	}

	// Validation fails with errors
	validator.validateFunc = func(diag *models.StateMachineDiagram, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
		return &models.ValidationResult{
			IsValid: false,
			Errors: []models.ValidationError{
				{Code: "SYNTAX_ERROR", Message: "Invalid PlantUML syntax"},
			},
			Warnings: []models.ValidationWarning{},
		}, nil
	}

	svc := NewService(repo, validator, nil)

	err := svc.PromoteToProductsFile(smmodels.DiagramTypePUML, "test-diag", "1.0.0")

	if err == nil {
		t.Error("PromoteToProductsFile() expected error but got none")
		return
	}

	var diagErr *models.StateMachineError
	if !errors.As(err, &diagErr) {
		t.Errorf("PromoteToProductsFile() expected StateMachineError but got %T", err)
		return
	}

	if diagErr.Type != models.ErrorTypeValidation {
		t.Errorf("PromoteToProductsFile() expected error type %v but got %v", models.ErrorTypeValidation, diagErr.Type)
	}
}

func TestService_PromoteToProductsFile_MoveOperationFails(t *testing.T) {
	repo := &mockRepository{}
	validator := &mockValidator{}

	// Setup successful validation but move fails
	repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
		if location == models.LocationInProgress {
			return true, nil
		}
		return false, nil
	}

	repo.readStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
		return &models.StateMachineDiagram{
			Name:     name,
			Version:  version,
			Content:  "@startuml\n[*] --> Idle\n@enduml",
			Location: location,
		}, nil
	}

	validator.validateFunc = func(diag *models.StateMachineDiagram, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
		return &models.ValidationResult{
			IsValid:  true,
			Errors:   []models.ValidationError{},
			Warnings: []models.ValidationWarning{},
		}, nil
	}

	// Move operation fails
	repo.moveStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, from, to models.Location) error {
		return errors.New("filesystem error during move")
	}

	svc := NewService(repo, validator, nil)

	err := svc.PromoteToProductsFile(smmodels.DiagramTypePUML, "test-diag", "1.0.0")

	if err == nil {
		t.Error("PromoteToProductsFile() expected error but got none")
		return
	}

	var diagErr *models.StateMachineError
	if !errors.As(err, &diagErr) {
		t.Errorf("PromoteToProductsFile() expected StateMachineError but got %T", err)
		return
	}

	if diagErr.Type != models.ErrorTypeFileSystem {
		t.Errorf("PromoteToProductsFile() expected error type %v but got %v", models.ErrorTypeFileSystem, diagErr.Type)
	}
}

func TestService_PromoteToProductsFile_AtomicOperationWithRollback(t *testing.T) {
	repo := &mockRepository{}
	validator := &mockValidator{}

	// Setup for rollback scenario - verification fails
	callCount := 0
	repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
		callCount++
		if callCount <= 2 {
			// Initial checks
			if location == models.LocationInProgress {
				return true, nil
			}
			return false, nil
		}
		// After move - simulate verification failure
		if location == models.LocationProducts {
			return false, nil // NOT found in products (verification failure)
		}
		return false, nil
	}

	repo.readStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
		return &models.StateMachineDiagram{
			Name:     name,
			Version:  version,
			Content:  "@startuml\n[*] --> Idle\n@enduml",
			Location: location,
		}, nil
	}

	validator.validateFunc = func(diag *models.StateMachineDiagram, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
		return &models.ValidationResult{
			IsValid:  true,
			Errors:   []models.ValidationError{},
			Warnings: []models.ValidationWarning{},
		}, nil
	}

	// Move operations (both promotion and rollback)
	moveCallCount := 0
	repo.moveStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, from, to models.Location) error {
		moveCallCount++
		return nil // Both moves succeed
	}

	svc := NewService(repo, validator, nil)

	err := svc.PromoteToProductsFile(smmodels.DiagramTypePUML, "test-diag", "1.0.0")

	if err == nil {
		t.Error("PromoteToProductsFile() expected error but got none")
		return
	}

	var diagErr *models.StateMachineError
	if !errors.As(err, &diagErr) {
		t.Errorf("PromoteToProductsFile() expected StateMachineError but got %T", err)
		return
	}

	if diagErr.Type != models.ErrorTypeFileSystem {
		t.Errorf("PromoteToProductsFile() expected error type %v but got %v", models.ErrorTypeFileSystem, diagErr.Type)
	}

	// Verify rollback was attempted (should have 2 move calls: promotion + rollback)
	if moveCallCount != 2 {
		t.Errorf("Expected 2 move operations (promotion + rollback), got %d", moveCallCount)
	}
}

// Integration test that exercises the full promotion workflow
func TestService_PromoteToProductsFile_IntegrationWorkflow(t *testing.T) {
	repo := &mockRepository{}
	validator := &mockValidator{}

	// Track the state changes through the promotion process
	var diagramState string = "in-progress"

	repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
		if location == models.LocationInProgress {
			return diagramState == "in-progress", nil
		}
		if location == models.LocationProducts {
			return diagramState == "products", nil
		}
		return false, nil
	}

	repo.readStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
		return &models.StateMachineDiagram{
			Name:        name,
			Version:     version,
			Content:     "@startuml\n[*] --> Idle\nIdle --> Active\nActive --> [*]\n@enduml",
			Location:    location,
			DiagramType: diagramType,
		}, nil
	}

	validator.validateFunc = func(diag *models.StateMachineDiagram, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
		// Ensure we're validating with the correct strictness for in-progress
		if strictness != models.StrictnessInProgress {
			t.Errorf("Expected validation strictness %v, got %v", models.StrictnessInProgress, strictness)
		}

		return &models.ValidationResult{
			IsValid: true,
			Errors:  []models.ValidationError{},
			Warnings: []models.ValidationWarning{
				{Code: "STYLE_WARNING", Message: "Consider adding more descriptive state names"},
			},
		}, nil
	}

	repo.moveStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, from, to models.Location) error {
		if from == models.LocationInProgress && to == models.LocationProducts {
			diagramState = "products" // Simulate successful move
			return nil
		}
		return errors.New("unexpected move operation")
	}

	svc := NewService(repo, validator, nil)

	// Execute the promotion
	err := svc.PromoteToProductsFile(smmodels.DiagramTypePUML, "user-workflow", "2.1.0")

	if err != nil {
		t.Errorf("PromoteToProductsFile() unexpected error: %v", err)
	}

	// Verify final state
	if diagramState != "products" {
		t.Errorf("Expected diagram to be in products state, got %s", diagramState)
	}
}
