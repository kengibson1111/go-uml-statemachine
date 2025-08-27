package service

import (
	"errors"
	"os"
	"testing"

	smmodels "github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/models"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name      string
		repo      models.Repository
		validator models.Validator
		config    *models.Config
		wantNil   bool
	}{
		{
			name:      "valid dependencies with config",
			repo:      &mockRepository{},
			validator: &mockValidator{},
			config: &models.Config{
				RootDirectory:   ".test-uml-statemachine",
				ValidationLevel: models.StrictnessProducts,
				BackupEnabled:   true,
				MaxFileSize:     2048,
			},
			wantNil: false,
		},
		{
			name:      "valid dependencies with nil config",
			repo:      &mockRepository{},
			validator: &mockValidator{},
			config:    nil,
			wantNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo, tt.validator, tt.config)

			if (svc == nil) != tt.wantNil {
				t.Errorf("NewService() = %v, wantNil %v", svc, tt.wantNil)
				return
			}

			if svc != nil {
				// Verify the service implements the interface
				_, ok := svc.(models.DiagramService)
				if !ok {
					t.Error("NewService() does not implement DiagramService interface")
				}

				// Verify internal state (access through type assertion for testing)
				s := svc.(*service)
				if s.repo != tt.repo {
					t.Error("NewService() did not set repository correctly")
				}
				if s.validator != tt.validator {
					t.Error("NewService() did not set validator correctly")
				}

				// Check config handling
				if tt.config == nil {
					// Should use default config
					if s.config == nil {
						t.Error("NewService() should use default config when nil is provided")
					}
					if s.config.RootDirectory != ".go-uml-statemachine-parsers" {
						t.Error("NewService() should use default root directory")
					}
				} else {
					if s.config != tt.config {
						t.Error("NewService() did not set config correctly")
					}
				}
			}
		})
	}
}

func TestNewServiceWithDefaults(t *testing.T) {
	repo := &mockRepository{}
	validator := &mockValidator{}

	svc := NewServiceWithDefaults(repo, validator)

	if svc == nil {
		t.Fatal("NewServiceWithDefaults() returned nil")
	}

	// Verify the service implements the interface
	_, ok := svc.(models.DiagramService)
	if !ok {
		t.Error("NewServiceWithDefaults() does not implement DiagramService interface")
	}

	// Verify internal state
	s := svc.(*service)
	if s.repo != repo {
		t.Error("NewServiceWithDefaults() did not set repository correctly")
	}
	if s.validator != validator {
		t.Error("NewServiceWithDefaults() did not set validator correctly")
	}

	// Should use default config
	if s.config == nil {
		t.Error("NewServiceWithDefaults() should set default config")
	}
	if s.config.RootDirectory != ".go-uml-statemachine-parsers" {
		t.Error("NewServiceWithDefaults() should use default root directory")
	}
	if s.config.ValidationLevel != models.StrictnessInProgress {
		t.Error("NewServiceWithDefaults() should use default validation level")
	}
}

func TestServiceThreadSafety(t *testing.T) {
	repo := &mockRepository{}
	validator := &mockValidator{}
	svc := NewService(repo, validator, nil)

	// Verify service has mutex field (this test mainly ensures the struct has the field)
	s := svc.(*service)

	// Test that we can acquire and release the mutex without panic
	s.mu.Lock()
	s.mu.Unlock()

	s.mu.RLock()
	s.mu.RUnlock()
}

// CRUD operation tests

func TestService_Create(t *testing.T) {
	tests := []struct {
		name         string
		inputName    string
		inputVer     string
		inputContent string
		inputLoc     models.Location
		setupMock    func(*mockRepository)
		wantErr      bool
		wantErrType  models.ErrorType
	}{
		{
			name:         "successful create in progress",
			inputName:    "test-diag",
			inputVer:     "1.0.0",
			inputContent: "@startuml\n[*] --> Idle\n@enduml",
			inputLoc:     models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					if location == models.LocationInProgress {
						return false, nil // doesn't exist in progress
					}
					return false, nil // doesn't exist in products either
				}
				repo.writeStateMachineFunc = func(diag *models.StateMachineDiagram) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:         "empty name validation",
			inputName:    "",
			inputVer:     "1.0.0",
			inputContent: "content",
			inputLoc:     models.LocationInProgress,
			setupMock:    func(repo *mockRepository) {},
			wantErr:      true,
			wantErrType:  models.ErrorTypeValidation,
		},
		{
			name:         "empty version validation",
			inputName:    "test-diag",
			inputVer:     "",
			inputContent: "content",
			inputLoc:     models.LocationInProgress,
			setupMock:    func(repo *mockRepository) {},
			wantErr:      true,
			wantErrType:  models.ErrorTypeValidation,
		},
		{
			name:         "empty content validation",
			inputName:    "test-diag",
			inputVer:     "1.0.0",
			inputContent: "",
			inputLoc:     models.LocationInProgress,
			setupMock:    func(repo *mockRepository) {},
			wantErr:      true,
			wantErrType:  models.ErrorTypeValidation,
		},
		{
			name:         "already exists conflict",
			inputName:    "test-diag",
			inputVer:     "1.0.0",
			inputContent: "content",
			inputLoc:     models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					return true, nil // already exists
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeDirectoryConflict,
		},
		{
			name:         "products directory conflict for in-progress",
			inputName:    "test-diag",
			inputVer:     "1.0.0",
			inputContent: "content",
			inputLoc:     models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					if location == models.LocationInProgress {
						return false, nil // doesn't exist in progress
					}
					return true, nil // exists in products - conflict!
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeDirectoryConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			validator := &mockValidator{}
			tt.setupMock(repo)

			svc := NewService(repo, validator, nil)

			result, err := svc.CreateFile(smmodels.DiagramTypePUML, tt.inputName, tt.inputVer, tt.inputContent, tt.inputLoc)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Create() expected error but got none")
					return
				}

				var diagErr *models.StateMachineError
				if !errors.As(err, &diagErr) {
					t.Errorf("Create() expected StateMachineError but got %T", err)
					return
				}

				if diagErr.Type != tt.wantErrType {
					t.Errorf("Create() expected error type %v but got %v", tt.wantErrType, diagErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("Create() unexpected error: %v", err)
					return
				}

				if result == nil {
					t.Error("Create() expected result but got nil")
					return
				}

				if result.Name != tt.inputName {
					t.Errorf("Create() name = %v, want %v", result.Name, tt.inputName)
				}
				if result.Version != tt.inputVer {
					t.Errorf("Create() version = %v, want %v", result.Version, tt.inputVer)
				}
				if result.Content != tt.inputContent {
					t.Errorf("Create() content = %v, want %v", result.Content, tt.inputContent)
				}
				if result.Location != tt.inputLoc {
					t.Errorf("Create() location = %v, want %v", result.Location, tt.inputLoc)
				}
			}
		})
	}
}

func TestService_ReadFile(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		inputVer    string
		inputLoc    models.Location
		setupMock   func(*mockRepository)
		wantErr     bool
		wantErrType models.ErrorType
		wantResult  *models.StateMachineDiagram
	}{
		{
			name:      "successful read",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			inputLoc:  models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.readStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
					return &models.StateMachineDiagram{
						Name:     name,
						Version:  version,
						Content:  "@startuml\n[*] --> Idle\n@enduml",
						Location: location,
					}, nil
				}
			},
			wantErr: false,
			wantResult: &models.StateMachineDiagram{
				Name:     "test-diag",
				Version:  "1.0.0",
				Content:  "@startuml\n[*] --> Idle\n@enduml",
				Location: models.LocationInProgress,
			},
		},
		{
			name:        "empty name validation",
			inputName:   "",
			inputVer:    "1.0.0",
			inputLoc:    models.LocationInProgress,
			setupMock:   func(repo *mockRepository) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name:        "empty version validation",
			inputName:   "test-diag",
			inputVer:    "",
			inputLoc:    models.LocationInProgress,
			setupMock:   func(repo *mockRepository) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name:      "repository error",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			inputLoc:  models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.readStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
					return nil, errors.New("file not found")
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			validator := &mockValidator{}
			tt.setupMock(repo)

			svc := NewService(repo, validator, nil)

			result, err := svc.ReadFile(smmodels.DiagramTypePUML, tt.inputName, tt.inputVer, tt.inputLoc)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ReadFile() expected error but got none")
					return
				}

				var diagErr *models.StateMachineError
				if !errors.As(err, &diagErr) {
					t.Errorf("ReadFile() expected StateMachineError but got %T", err)
					return
				}

				if diagErr.Type != tt.wantErrType {
					t.Errorf("ReadFile() expected error type %v but got %v", tt.wantErrType, diagErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("ReadFile() unexpected error: %v", err)
					return
				}

				if result == nil {
					t.Error("ReadFile() expected result but got nil")
					return
				}

				if result.Name != tt.wantResult.Name {
					t.Errorf("ReadFile() name = %v, want %v", result.Name, tt.wantResult.Name)
				}
				if result.Version != tt.wantResult.Version {
					t.Errorf("ReadFile() version = %v, want %v", result.Version, tt.wantResult.Version)
				}
				if result.Content != tt.wantResult.Content {
					t.Errorf("ReadFile() content = %v, want %v", result.Content, tt.wantResult.Content)
				}
				if result.Location != tt.wantResult.Location {
					t.Errorf("ReadFile() location = %v, want %v", result.Location, tt.wantResult.Location)
				}
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	tests := []struct {
		name        string
		input       *models.StateMachineDiagram
		setupMock   func(*mockRepository)
		wantErr     bool
		wantErrType models.ErrorType
	}{
		{
			name: "successful update",
			input: &models.StateMachineDiagram{
				Name:     "test-diag",
				Version:  "1.0.0",
				Content:  "@startuml\n[*] --> Updated\n@enduml",
				Location: models.LocationInProgress,
			},
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					return true, nil // exists
				}
				repo.writeStateMachineFunc = func(diag *models.StateMachineDiagram) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:        "nil state-machine diagram validation",
			input:       nil,
			setupMock:   func(repo *mockRepository) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name: "empty name validation",
			input: &models.StateMachineDiagram{
				Name:     "",
				Version:  "1.0.0",
				Content:  "content",
				Location: models.LocationInProgress,
			},
			setupMock:   func(repo *mockRepository) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name: "empty version validation",
			input: &models.StateMachineDiagram{
				Name:     "test-diag",
				Version:  "",
				Content:  "content",
				Location: models.LocationInProgress,
			},
			setupMock:   func(repo *mockRepository) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name: "empty content validation",
			input: &models.StateMachineDiagram{
				Name:     "test-diag",
				Version:  "1.0.0",
				Content:  "",
				Location: models.LocationInProgress,
			},
			setupMock:   func(repo *mockRepository) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name: "state-machine diagram does not exist",
			input: &models.StateMachineDiagram{
				Name:     "test-diag",
				Version:  "1.0.0",
				Content:  "content",
				Location: models.LocationInProgress,
			},
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					return false, nil // doesn't exist
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			validator := &mockValidator{}
			tt.setupMock(repo)

			svc := NewService(repo, validator, nil)

			err := svc.UpdateInProgressFile(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Update() expected error but got none")
					return
				}

				var diagErr *models.StateMachineError
				if !errors.As(err, &diagErr) {
					t.Errorf("Update() expected StateMachineError but got %T", err)
					return
				}

				if diagErr.Type != tt.wantErrType {
					t.Errorf("Update() expected error type %v but got %v", tt.wantErrType, diagErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("Update() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		inputVer    string
		inputLoc    models.Location
		setupMock   func(*mockRepository)
		wantErr     bool
		wantErrType models.ErrorType
	}{
		{
			name:      "successful delete",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			inputLoc:  models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					return true, nil // exists
				}
				repo.deleteStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:        "empty name validation",
			inputName:   "",
			inputVer:    "1.0.0",
			inputLoc:    models.LocationInProgress,
			setupMock:   func(repo *mockRepository) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name:        "empty version validation",
			inputName:   "test-diag",
			inputVer:    "",
			inputLoc:    models.LocationInProgress,
			setupMock:   func(repo *mockRepository) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name:      "state-machine diagram does not exist",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			inputLoc:  models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					return false, nil // doesn't exist
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeFileNotFound,
		},
		{
			name:      "repository delete error",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			inputLoc:  models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					return true, nil // exists
				}
				repo.deleteStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) error {
					return errors.New("delete failed")
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeFileSystem,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			validator := &mockValidator{}
			tt.setupMock(repo)

			svc := NewService(repo, validator, nil)

			err := svc.DeleteFile(smmodels.DiagramTypePUML, tt.inputName, tt.inputVer, tt.inputLoc)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Delete() expected error but got none")
					return
				}

				var diagErr *models.StateMachineError
				if !errors.As(err, &diagErr) {
					t.Errorf("Delete() expected StateMachineError but got %T", err)
					return
				}

				if diagErr.Type != tt.wantErrType {
					t.Errorf("Delete() expected error type %v but got %v", tt.wantErrType, diagErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("Delete() unexpected error: %v", err)
				}
			}
		})
	}
}
func TestService_Promote(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		inputVer    string
		setupMock   func(*mockRepository, *mockValidator)
		wantErr     bool
		wantErrType models.ErrorType
	}{
		{
			name:      "successful promotion",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				// State-machine diagram exists in in-progress
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					if location == models.LocationInProgress {
						return true, nil
					}
					if location == models.LocationProducts {
						return false, nil // doesn't exist in products initially
					}
					return false, nil
				}

				// Read state-machine diagram for validation
				repo.readStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
					return &models.StateMachineDiagram{
						Name:     name,
						Version:  version,
						Content:  "@startuml\n[*] --> Idle\n@enduml",
						Location: location,
					}, nil
				}

				// Validation passes
				validator.validateFunc = func(diag *models.StateMachineDiagram, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
					return &models.ValidationResult{
						IsValid:  true,
						Errors:   []models.ValidationError{},
						Warnings: []models.ValidationWarning{},
					}, nil
				}

				// Move operation succeeds
				repo.moveStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, from, to models.Location) error {
					return nil
				}

				// After move verification - track call count to simulate state change
				callCount := 0
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					callCount++
					if callCount <= 2 {
						// First two calls are the initial checks
						if location == models.LocationInProgress {
							return true, nil
						}
						return false, nil
					}
					// After move operation
					if location == models.LocationProducts {
						return true, nil // now exists in products
					}
					if location == models.LocationInProgress {
						return false, nil // no longer exists in in-progress
					}
					return false, nil
				}
			},
			wantErr: false,
		},
		{
			name:        "empty name validation",
			inputName:   "",
			inputVer:    "1.0.0",
			setupMock:   func(repo *mockRepository, validator *mockValidator) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name:        "empty version validation",
			inputName:   "test-diag",
			inputVer:    "",
			setupMock:   func(repo *mockRepository, validator *mockValidator) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name:      "state-machine diagram does not exist in in-progress",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					return false, nil // doesn't exist anywhere
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeFileNotFound,
		},
		{
			name:      "directory conflict in products",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					if location == models.LocationInProgress {
						return true, nil // exists in in-progress
					}
					if location == models.LocationProducts {
						return true, nil // already exists in products - conflict!
					}
					return false, nil
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeDirectoryConflict,
		},
		{
			name:      "validation fails with errors",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				// State-machine diagram exists in in-progress
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					if location == models.LocationInProgress {
						return true, nil
					}
					return false, nil
				}

				// Read state-machine diagram for validation
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
					result := &models.ValidationResult{
						IsValid: false,
						Errors: []models.ValidationError{
							{Code: "SYNTAX_ERROR", Message: "Invalid PlantUML syntax"},
						},
						Warnings: []models.ValidationWarning{},
					}
					return result, nil
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name:      "move operation fails",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				// State-machine diagram exists in in-progress
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					if location == models.LocationInProgress {
						return true, nil
					}
					return false, nil
				}

				// Read state-machine diagram for validation
				repo.readStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
					return &models.StateMachineDiagram{
						Name:     name,
						Version:  version,
						Content:  "@startuml\n[*] --> Idle\n@enduml",
						Location: location,
					}, nil
				}

				// Validation passes
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
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeFileSystem,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			validator := &mockValidator{}
			tt.setupMock(repo, validator)

			svc := NewService(repo, validator, nil)

			err := svc.Promote(smmodels.DiagramTypePUML, tt.inputName, tt.inputVer)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Promote() expected error but got none")
					return
				}

				var diagErr *models.StateMachineError
				if !errors.As(err, &diagErr) {
					t.Errorf("Promote() expected StateMachineError but got %T", err)
					return
				}

				if diagErr.Type != tt.wantErrType {
					t.Errorf("Promote() expected error type %v but got %v", tt.wantErrType, diagErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("Promote() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestService_PromoteWithRollback(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		inputVer    string
		setupMock   func(*mockRepository, *mockValidator)
		wantErr     bool
		wantErrType models.ErrorType
	}{
		{
			name:      "rollback on verification failure - not found in products",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				// Initial setup - state-machine diagram exists in in-progress
				initialCallCount := 0
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					initialCallCount++
					if initialCallCount <= 2 {
						// First two calls are the initial checks
						if location == models.LocationInProgress {
							return true, nil
						}
						return false, nil
					}
					// After move operation - simulate partial failure
					if location == models.LocationProducts {
						return false, nil // NOT found in products (verification failure)
					}
					if location == models.LocationInProgress {
						return false, nil // not in in-progress either
					}
					return false, nil
				}

				// Read state-machine diagram for validation
				repo.readStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
					return &models.StateMachineDiagram{
						Name:     name,
						Version:  version,
						Content:  "@startuml\n[*] --> Idle\n@enduml",
						Location: location,
					}, nil
				}

				// Validation passes
				validator.validateFunc = func(diag *models.StateMachineDiagram, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
					return &models.ValidationResult{
						IsValid:  true,
						Errors:   []models.ValidationError{},
						Warnings: []models.ValidationWarning{},
					}, nil
				}

				// Move operation appears to succeed initially
				moveCallCount := 0
				repo.moveStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, from, to models.Location) error {
					moveCallCount++
					if moveCallCount == 1 {
						// First move (promotion) succeeds
						return nil
					}
					// Second move (rollback) also succeeds
					return nil
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeFileSystem,
		},
		{
			name:      "rollback on verification failure - still in in-progress",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				// Initial setup - state-machine diagram exists in in-progress
				initialCallCount := 0
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					initialCallCount++
					if initialCallCount <= 2 {
						// First two calls are the initial checks
						if location == models.LocationInProgress {
							return true, nil
						}
						return false, nil
					}
					// After move operation - simulate partial failure
					if location == models.LocationProducts {
						return true, nil // found in products
					}
					if location == models.LocationInProgress {
						return true, nil // STILL in in-progress (verification failure)
					}
					return false, nil
				}

				// Read state-machine diagram for validation
				repo.readStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
					return &models.StateMachineDiagram{
						Name:     name,
						Version:  version,
						Content:  "@startuml\n[*] --> Idle\n@enduml",
						Location: location,
					}, nil
				}

				// Validation passes
				validator.validateFunc = func(diag *models.StateMachineDiagram, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
					return &models.ValidationResult{
						IsValid:  true,
						Errors:   []models.ValidationError{},
						Warnings: []models.ValidationWarning{},
					}, nil
				}

				// Move operation appears to succeed
				repo.moveStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, from, to models.Location) error {
					return nil
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeFileSystem,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			validator := &mockValidator{}
			tt.setupMock(repo, validator)

			svc := NewService(repo, validator, nil)

			err := svc.Promote(smmodels.DiagramTypePUML, tt.inputName, tt.inputVer)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Promote() expected error but got none")
					return
				}

				var diagErr *models.StateMachineError
				if !errors.As(err, &diagErr) {
					t.Errorf("Promote() expected StateMachineError but got %T", err)
					return
				}

				if diagErr.Type != tt.wantErrType {
					t.Errorf("Promote() expected error type %v but got %v", tt.wantErrType, diagErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("Promote() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestService_PromoteValidationScenarios(t *testing.T) {
	tests := []struct {
		name           string
		inputName      string
		inputVer       string
		validationFunc func(*models.StateMachineDiagram, models.ValidationStrictness) (*models.ValidationResult, error)
		wantErr        bool
		wantErrType    models.ErrorType
	}{
		{
			name:      "validation passes with warnings only",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			validationFunc: func(diag *models.StateMachineDiagram, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
				return &models.ValidationResult{
					IsValid: true,
					Errors:  []models.ValidationError{},
					Warnings: []models.ValidationWarning{
						{Code: "STYLE_WARNING", Message: "Consider using better naming"},
					},
				}, nil
			},
			wantErr: false,
		},
		{
			name:      "validation fails with errors and warnings",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			validationFunc: func(diag *models.StateMachineDiagram, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
				return &models.ValidationResult{
					IsValid: false,
					Errors: []models.ValidationError{
						{Code: "SYNTAX_ERROR", Message: "Invalid syntax"},
					},
					Warnings: []models.ValidationWarning{
						{Code: "STYLE_WARNING", Message: "Consider using better naming"},
					},
				}, nil
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name:      "validation error during validation process",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			validationFunc: func(diag *models.StateMachineDiagram, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
				return nil, errors.New("validation process failed")
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			validator := &mockValidator{}

			// Setup common mock behavior
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

			// Set the specific validation function for this test
			validator.validateFunc = tt.validationFunc

			// Setup successful move if validation passes
			if !tt.wantErr {
				callCount := 0
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					callCount++
					if callCount <= 2 {
						if location == models.LocationInProgress {
							return true, nil
						}
						return false, nil
					}
					// After move
					if location == models.LocationProducts {
						return true, nil
					}
					if location == models.LocationInProgress {
						return false, nil
					}
					return false, nil
				}

				repo.moveStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, from, to models.Location) error {
					return nil
				}
			}

			svc := NewService(repo, validator, nil)

			err := svc.Promote(smmodels.DiagramTypePUML, tt.inputName, tt.inputVer)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Promote() expected error but got none")
					return
				}

				var diagErr *models.StateMachineError
				if !errors.As(err, &diagErr) {
					t.Errorf("Promote() expected StateMachineError but got %T", err)
					return
				}

				if diagErr.Type != tt.wantErrType {
					t.Errorf("Promote() expected error type %v but got %v", tt.wantErrType, diagErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("Promote() unexpected error: %v", err)
				}
			}
		})
	}
}

// Tests for task 7.4 - Listing and validation operations

func TestService_Validate(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		inputVer    string
		inputLoc    models.Location
		setupMock   func(*mockRepository, *mockValidator)
		wantErr     bool
		wantErrType models.ErrorType
		wantResult  *models.ValidationResult
	}{
		{
			name:      "successful validation in-progress with strictness in-progress",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			inputLoc:  models.LocationInProgress,
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				repo.readStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
					return &models.StateMachineDiagram{
						Name:     name,
						Version:  version,
						Content:  "@startuml\n[*] --> Idle\n@enduml",
						Location: location,
					}, nil
				}
				validator.validateFunc = func(diag *models.StateMachineDiagram, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
					if strictness != models.StrictnessInProgress {
						t.Errorf("Expected StrictnessInProgress but got %v", strictness)
					}
					return &models.ValidationResult{
						IsValid:  true,
						Errors:   []models.ValidationError{},
						Warnings: []models.ValidationWarning{},
					}, nil
				}
			},
			wantErr: false,
			wantResult: &models.ValidationResult{
				IsValid:  true,
				Errors:   []models.ValidationError{},
				Warnings: []models.ValidationWarning{},
			},
		},
		{
			name:      "successful validation products with strictness products",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			inputLoc:  models.LocationProducts,
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				repo.readStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
					return &models.StateMachineDiagram{
						Name:     name,
						Version:  version,
						Content:  "@startuml\n[*] --> Idle\n@enduml",
						Location: location,
					}, nil
				}
				validator.validateFunc = func(diag *models.StateMachineDiagram, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
					if strictness != models.StrictnessProducts {
						t.Errorf("Expected StrictnessProducts but got %v", strictness)
					}
					return &models.ValidationResult{
						IsValid: true,
						Errors:  []models.ValidationError{},
						Warnings: []models.ValidationWarning{
							{Code: "STYLE_WARNING", Message: "Consider improving naming"},
						},
					}, nil
				}
			},
			wantErr: false,
			wantResult: &models.ValidationResult{
				IsValid: true,
				Errors:  []models.ValidationError{},
				Warnings: []models.ValidationWarning{
					{Code: "STYLE_WARNING", Message: "Consider improving naming"},
				},
			},
		},
		{
			name:        "empty name validation",
			inputName:   "",
			inputVer:    "1.0.0",
			inputLoc:    models.LocationInProgress,
			setupMock:   func(repo *mockRepository, validator *mockValidator) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name:        "empty version validation",
			inputName:   "test-diag",
			inputVer:    "",
			inputLoc:    models.LocationInProgress,
			setupMock:   func(repo *mockRepository, validator *mockValidator) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name:      "state-machine diagram not found",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			inputLoc:  models.LocationInProgress,
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				repo.readStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
					return nil, errors.New("file not found")
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeFileNotFound,
		},
		{
			name:      "validator error",
			inputName: "test-diag",
			inputVer:  "1.0.0",
			inputLoc:  models.LocationInProgress,
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				repo.readStateMachineFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
					return &models.StateMachineDiagram{
						Name:     name,
						Version:  version,
						Content:  "invalid content",
						Location: location,
					}, nil
				}
				validator.validateFunc = func(diag *models.StateMachineDiagram, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
					return nil, errors.New("validation engine error")
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			validator := &mockValidator{}
			tt.setupMock(repo, validator)

			svc := NewService(repo, validator, nil)

			result, err := svc.Validate(smmodels.DiagramTypePUML, tt.inputName, tt.inputVer, tt.inputLoc)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got none")
					return
				}

				var diagErr *models.StateMachineError
				if !errors.As(err, &diagErr) {
					t.Errorf("Validate() expected StateMachineError but got %T", err)
					return
				}

				if diagErr.Type != tt.wantErrType {
					t.Errorf("Validate() expected error type %v but got %v", tt.wantErrType, diagErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
					return
				}

				if result == nil {
					t.Error("Validate() expected result but got nil")
					return
				}

				if result.IsValid != tt.wantResult.IsValid {
					t.Errorf("Validate() IsValid = %v, want %v", result.IsValid, tt.wantResult.IsValid)
				}
				if len(result.Errors) != len(tt.wantResult.Errors) {
					t.Errorf("Validate() Errors count = %v, want %v", len(result.Errors), len(tt.wantResult.Errors))
				}
				if len(result.Warnings) != len(tt.wantResult.Warnings) {
					t.Errorf("Validate() Warnings count = %v, want %v", len(result.Warnings), len(tt.wantResult.Warnings))
				}
			}
		})
	}
}

func TestService_ListAll(t *testing.T) {
	tests := []struct {
		name        string
		inputLoc    models.Location
		setupMock   func(*mockRepository)
		wantErr     bool
		wantErrType models.ErrorType
		wantCount   int
	}{
		{
			name:     "successful list in-progress",
			inputLoc: models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.listStateMachinesFunc = func(diagramType smmodels.DiagramType, location models.Location) ([]models.StateMachineDiagram, error) {
					return []models.StateMachineDiagram{
						{
							Name:     "diag1",
							Version:  "1.0.0",
							Content:  "@startuml\n[*] --> Idle\n@enduml",
							Location: location,
						},
						{
							Name:     "diag2",
							Version:  "2.0.0",
							Content:  "@startuml\n[*] --> Active\n@enduml",
							Location: location,
						},
					}, nil
				}
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:     "successful list products",
			inputLoc: models.LocationProducts,
			setupMock: func(repo *mockRepository) {
				repo.listStateMachinesFunc = func(diagramType smmodels.DiagramType, location models.Location) ([]models.StateMachineDiagram, error) {
					return []models.StateMachineDiagram{
						{
							Name:     "prod-diag",
							Version:  "1.0.0",
							Content:  "@startuml\n[*] --> Ready\n@enduml",
							Location: location,
						},
					}, nil
				}
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:     "empty list",
			inputLoc: models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.listStateMachinesFunc = func(diagramType smmodels.DiagramType, location models.Location) ([]models.StateMachineDiagram, error) {
					return []models.StateMachineDiagram{}, nil
				}
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:     "repository error",
			inputLoc: models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.listStateMachinesFunc = func(diagramType smmodels.DiagramType, location models.Location) ([]models.StateMachineDiagram, error) {
					return nil, errors.New("directory read error")
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeFileSystem,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			validator := &mockValidator{}
			tt.setupMock(repo)

			svc := NewService(repo, validator, nil)

			result, err := svc.ListAll(smmodels.DiagramTypePUML, tt.inputLoc)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ListAll() expected error but got none")
					return
				}

				var diagErr *models.StateMachineError
				if !errors.As(err, &diagErr) {
					t.Errorf("ListAll() expected StateMachineError but got %T", err)
					return
				}

				if diagErr.Type != tt.wantErrType {
					t.Errorf("ListAll() expected error type %v but got %v", tt.wantErrType, diagErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("ListAll() unexpected error: %v", err)
					return
				}

				if result == nil {
					t.Error("ListAll() expected result but got nil")
					return
				}

				if len(result) != tt.wantCount {
					t.Errorf("ListAll() count = %v, want %v", len(result), tt.wantCount)
				}

				// Verify all returned state-machine diagrams have the correct location
				for _, diag := range result {
					if diag.Location != tt.inputLoc {
						t.Errorf("ListAll() state-machine diagram location = %v, want %v", diag.Location, tt.inputLoc)
					}
				}
			}
		})
	}
}

func TestService_ResolveReferences(t *testing.T) {
	tests := []struct {
		name        string
		input       *models.StateMachineDiagram
		setupMock   func(*mockRepository)
		wantErr     bool
		wantErrType models.ErrorType
	}{
		{
			name: "successful resolve no references",
			input: &models.StateMachineDiagram{
				Name:       "test-diag",
				Version:    "1.0.0",
				Content:    "@startuml\n[*] --> Idle\n@enduml",
				Location:   models.LocationInProgress,
				References: []models.Reference{},
			},
			setupMock: func(repo *mockRepository) {},
			wantErr:   false,
		},
		{
			name: "successful resolve product reference",
			input: &models.StateMachineDiagram{
				Name:     "test-diag",
				Version:  "1.0.0",
				Content:  "@startuml\n[*] --> Idle\n@enduml",
				Location: models.LocationInProgress,
				References: []models.Reference{
					{
						Name:    "auth-diag",
						Version: "2.0.0",
						Type:    models.ReferenceTypeProduct,
					},
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					if name == "auth-diag" && version == "2.0.0" && location == models.LocationProducts {
						return true, nil
					}
					return false, nil
				}
			},
			wantErr: false,
		},

		{
			name: "successful resolve multiple references",
			input: &models.StateMachineDiagram{
				Name:     "test-diag",
				Version:  "1.0.0",
				Content:  "@startuml\n[*] --> Idle\n@enduml",
				Location: models.LocationInProgress,
				References: []models.Reference{
					{
						Name:    "auth-diag",
						Version: "2.0.0",
						Type:    models.ReferenceTypeProduct,
					},
					{
						Name:    "payment-diag",
						Version: "1.5.0",
						Type:    models.ReferenceTypeProduct,
					},
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					if (name == "auth-diag" && version == "2.0.0" && location == models.LocationProducts) ||
						(name == "payment-diag" && version == "1.5.0" && location == models.LocationProducts) {
						return true, nil
					}
					return false, nil
				}
			},
			wantErr: false,
		},
		{
			name:        "nil state-machine diagram validation",
			input:       nil,
			setupMock:   func(repo *mockRepository) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name: "product reference not found",
			input: &models.StateMachineDiagram{
				Name:     "test-diag",
				Version:  "1.0.0",
				Content:  "@startuml\n[*] --> Idle\n@enduml",
				Location: models.LocationInProgress,
				References: []models.Reference{
					{
						Name:    "missing-diag",
						Version: "1.0.0",
						Type:    models.ReferenceTypeProduct,
					},
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					return false, nil // not found
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeReferenceResolution,
		},

		{
			name: "repository error checking product reference",
			input: &models.StateMachineDiagram{
				Name:     "test-diag",
				Version:  "1.0.0",
				Content:  "@startuml\n[*] --> Idle\n@enduml",
				Location: models.LocationInProgress,
				References: []models.Reference{
					{
						Name:    "auth-diag",
						Version: "2.0.0",
						Type:    models.ReferenceTypeProduct,
					},
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					return false, models.NewStateMachineError(models.ErrorTypeFileSystem, "filesystem error", nil)
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeFileSystem,
		},
		{
			name: "verify only product references are processed",
			input: &models.StateMachineDiagram{
				Name:     "test-diag",
				Version:  "1.0.0",
				Content:  "@startuml\n[*] --> Idle\n@enduml",
				Location: models.LocationInProgress,
				References: []models.Reference{
					{
						Name:    "product-diag",
						Version: "1.0.0",
						Type:    models.ReferenceTypeProduct,
					},
				},
			},
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
					// Only product references should be checked
					if location == models.LocationProducts {
						return true, nil
					}
					return false, nil
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			validator := &mockValidator{}
			tt.setupMock(repo)

			svc := NewService(repo, validator, nil)

			err := svc.ResolveReferences(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ResolveReferences() expected error but got none")
					return
				}

				var diagErr *models.StateMachineError
				if !errors.As(err, &diagErr) {
					t.Errorf("ResolveReferences() expected StateMachineError but got %T", err)
					return
				}

				if diagErr.Type != tt.wantErrType {
					t.Errorf("ResolveReferences() expected error type %v but got %v", tt.wantErrType, diagErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("ResolveReferences() unexpected error: %v", err)
					return
				}

				// Verify that references have been resolved with correct paths
				if tt.input != nil && len(tt.input.References) > 0 {
					for _, ref := range tt.input.References {
						if ref.Path == "" {
							t.Errorf("ResolveReferences() reference path not set for %s", ref.Name)
						}

						// Verify path format - should only be product references now
						if ref.Type != models.ReferenceTypeProduct {
							t.Errorf("ResolveReferences() should only handle product references, got %v", ref.Type)
						}

						expectedPath := "products\\puml\\" + ref.Name + "-" + ref.Version + "\\" + ref.Name + "-" + ref.Version + ".puml"
						if ref.Path != expectedPath {
							t.Errorf("ResolveReferences() product reference path = %v, want %v", ref.Path, expectedPath)
						}
					}
				}
			}
		})
	}
}

func TestService_ResolveReferencesPathBuilding(t *testing.T) {
	tests := []struct {
		name         string
		diagram      *models.StateMachineDiagram
		reference    models.Reference
		expectedPath string
	}{
		{
			name: "product reference path building",
			diagram: &models.StateMachineDiagram{
				Name:     "test-diag",
				Version:  "1.0.0",
				Location: models.LocationInProgress,
			},
			reference: models.Reference{
				Name:    "auth-service",
				Version: "2.1.0",
				Type:    models.ReferenceTypeProduct,
			},
			expectedPath: "products\\puml\\auth-service-2.1.0\\auth-service-2.1.0.puml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{}
			validator := &mockValidator{}

			// Setup mocks to return success for existence checks
			repo.existsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
				return true, nil
			}
			repo.directoryExistsFunc = func(path string) (bool, error) {
				return true, nil
			}

			svc := NewService(repo, validator, nil)

			// Create a state-machine diagram with the test reference
			diag := &models.StateMachineDiagram{
				Name:       tt.diagram.Name,
				Version:    tt.diagram.Version,
				Location:   tt.diagram.Location,
				References: []models.Reference{tt.reference},
			}

			err := svc.ResolveReferences(diag)

			if err != nil {
				t.Errorf("ResolveReferences() unexpected error: %v", err)
				return
			}

			if len(diag.References) != 1 {
				t.Errorf("ResolveReferences() expected 1 reference but got %d", len(diag.References))
				return
			}

			actualPath := diag.References[0].Path
			if actualPath != tt.expectedPath {
				t.Errorf("ResolveReferences() path = %v, want %v", actualPath, tt.expectedPath)
			}
		})
	}
}
func TestNewServiceFromEnv(t *testing.T) {
	// Save original environment variables
	originalRootDir := os.Getenv("GO_UML_ROOT_DIRECTORY")
	originalValidationLevel := os.Getenv("GO_UML_VALIDATION_LEVEL")
	originalBackupEnabled := os.Getenv("GO_UML_BACKUP_ENABLED")
	originalMaxFileSize := os.Getenv("GO_UML_MAX_FILE_SIZE")

	// Clean up after test
	defer func() {
		os.Setenv("GO_UML_ROOT_DIRECTORY", originalRootDir)
		os.Setenv("GO_UML_VALIDATION_LEVEL", originalValidationLevel)
		os.Setenv("GO_UML_BACKUP_ENABLED", originalBackupEnabled)
		os.Setenv("GO_UML_MAX_FILE_SIZE", originalMaxFileSize)
	}()

	tests := []struct {
		name           string
		envRootDir     string
		envValidation  string
		envBackup      string
		envMaxFileSize string
	}{
		{
			name:           "default environment",
			envRootDir:     "",
			envValidation:  "",
			envBackup:      "",
			envMaxFileSize: "",
		},
		{
			name:           "custom environment",
			envRootDir:     "C:\\custom\\env\\path",
			envValidation:  "products",
			envBackup:      "true",
			envMaxFileSize: "2097152",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("GO_UML_ROOT_DIRECTORY", tt.envRootDir)
			os.Setenv("GO_UML_VALIDATION_LEVEL", tt.envValidation)
			os.Setenv("GO_UML_BACKUP_ENABLED", tt.envBackup)
			os.Setenv("GO_UML_MAX_FILE_SIZE", tt.envMaxFileSize)

			repo := &mockRepository{}
			validator := &mockValidator{}

			svc := NewServiceFromEnv(repo, validator)

			if svc == nil {
				t.Error("NewServiceFromEnv() returned nil")
				return
			}

			// Verify the service implements the interface
			_, ok := svc.(models.DiagramService)
			if !ok {
				t.Error("NewServiceFromEnv() did not return a DiagramService")
			}
		})
	}
}

func TestNewServiceWithEnvOverrides(t *testing.T) {
	// Save original environment variables
	originalRootDir := os.Getenv("GO_UML_ROOT_DIRECTORY")
	originalValidationLevel := os.Getenv("GO_UML_VALIDATION_LEVEL")
	originalBackupEnabled := os.Getenv("GO_UML_BACKUP_ENABLED")
	originalMaxFileSize := os.Getenv("GO_UML_MAX_FILE_SIZE")

	// Clean up after test
	defer func() {
		os.Setenv("GO_UML_ROOT_DIRECTORY", originalRootDir)
		os.Setenv("GO_UML_VALIDATION_LEVEL", originalValidationLevel)
		os.Setenv("GO_UML_BACKUP_ENABLED", originalBackupEnabled)
		os.Setenv("GO_UML_MAX_FILE_SIZE", originalMaxFileSize)
	}()

	baseConfig := &models.Config{
		RootDirectory:   "C:\\base\\path",
		ValidationLevel: models.StrictnessInProgress,
		BackupEnabled:   false,
		MaxFileSize:     1024,
	}

	tests := []struct {
		name           string
		baseConfig     *models.Config
		envRootDir     string
		envValidation  string
		envBackup      string
		envMaxFileSize string
	}{
		{
			name:           "nil base config with env overrides",
			baseConfig:     nil,
			envRootDir:     "D:\\env\\path",
			envValidation:  "products",
			envBackup:      "true",
			envMaxFileSize: "4096",
		},
		{
			name:           "base config with partial env overrides",
			baseConfig:     baseConfig,
			envRootDir:     "E:\\override\\path",
			envValidation:  "",
			envBackup:      "true",
			envMaxFileSize: "",
		},
		{
			name:           "base config with no env overrides",
			baseConfig:     baseConfig,
			envRootDir:     "",
			envValidation:  "",
			envBackup:      "",
			envMaxFileSize: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("GO_UML_ROOT_DIRECTORY", tt.envRootDir)
			os.Setenv("GO_UML_VALIDATION_LEVEL", tt.envValidation)
			os.Setenv("GO_UML_BACKUP_ENABLED", tt.envBackup)
			os.Setenv("GO_UML_MAX_FILE_SIZE", tt.envMaxFileSize)

			repo := &mockRepository{}
			validator := &mockValidator{}

			svc := NewServiceWithEnvOverrides(repo, validator, tt.baseConfig)

			if svc == nil {
				t.Error("NewServiceWithEnvOverrides() returned nil")
				return
			}

			// Verify the service implements the interface
			_, ok := svc.(models.DiagramService)
			if !ok {
				t.Error("NewServiceWithEnvOverrides() did not return a DiagramService")
			}
		})
	}
}

func TestServiceFactoryFunctions(t *testing.T) {
	repo := &mockRepository{}
	validator := &mockValidator{}

	t.Run("NewServiceWithDefaults", func(t *testing.T) {
		svc := NewServiceWithDefaults(repo, validator)
		if svc == nil {
			t.Error("NewServiceWithDefaults() returned nil")
		}
	})

	t.Run("NewService with nil config", func(t *testing.T) {
		svc := NewService(repo, validator, nil)
		if svc == nil {
			t.Error("NewService() with nil config returned nil")
		}
	})
}
