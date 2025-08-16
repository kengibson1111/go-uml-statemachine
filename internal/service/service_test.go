package service

import (
	"errors"
	"testing"

	"github.com/kengibson1111/go-uml-statemachine/internal/models"
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
				_, ok := svc.(models.StateMachineService)
				if !ok {
					t.Error("NewService() does not implement StateMachineService interface")
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
					if s.config.RootDirectory != ".go-uml-statemachine" {
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
	_, ok := svc.(models.StateMachineService)
	if !ok {
		t.Error("NewServiceWithDefaults() does not implement StateMachineService interface")
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
	if s.config.RootDirectory != ".go-uml-statemachine" {
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
			inputName:    "test-sm",
			inputVer:     "1.0.0",
			inputContent: "@startuml\n[*] --> Idle\n@enduml",
			inputLoc:     models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
					if location == models.LocationInProgress {
						return false, nil // doesn't exist in progress
					}
					return false, nil // doesn't exist in products either
				}
				repo.writeStateMachineFunc = func(sm *models.StateMachine) error {
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
			inputName:    "test-sm",
			inputVer:     "",
			inputContent: "content",
			inputLoc:     models.LocationInProgress,
			setupMock:    func(repo *mockRepository) {},
			wantErr:      true,
			wantErrType:  models.ErrorTypeValidation,
		},
		{
			name:         "empty content validation",
			inputName:    "test-sm",
			inputVer:     "1.0.0",
			inputContent: "",
			inputLoc:     models.LocationInProgress,
			setupMock:    func(repo *mockRepository) {},
			wantErr:      true,
			wantErrType:  models.ErrorTypeValidation,
		},
		{
			name:         "already exists conflict",
			inputName:    "test-sm",
			inputVer:     "1.0.0",
			inputContent: "content",
			inputLoc:     models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
					return true, nil // already exists
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeDirectoryConflict,
		},
		{
			name:         "products directory conflict for in-progress",
			inputName:    "test-sm",
			inputVer:     "1.0.0",
			inputContent: "content",
			inputLoc:     models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
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

			result, err := svc.Create(tt.inputName, tt.inputVer, tt.inputContent, tt.inputLoc)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Create() expected error but got none")
					return
				}

				var smErr *models.StateMachineError
				if !errors.As(err, &smErr) {
					t.Errorf("Create() expected StateMachineError but got %T", err)
					return
				}

				if smErr.Type != tt.wantErrType {
					t.Errorf("Create() expected error type %v but got %v", tt.wantErrType, smErr.Type)
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

func TestService_Read(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		inputVer    string
		inputLoc    models.Location
		setupMock   func(*mockRepository)
		wantErr     bool
		wantErrType models.ErrorType
		wantResult  *models.StateMachine
	}{
		{
			name:      "successful read",
			inputName: "test-sm",
			inputVer:  "1.0.0",
			inputLoc:  models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.readStateMachineFunc = func(name, version string, location models.Location) (*models.StateMachine, error) {
					return &models.StateMachine{
						Name:     name,
						Version:  version,
						Content:  "@startuml\n[*] --> Idle\n@enduml",
						Location: location,
					}, nil
				}
			},
			wantErr: false,
			wantResult: &models.StateMachine{
				Name:     "test-sm",
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
			inputName:   "test-sm",
			inputVer:    "",
			inputLoc:    models.LocationInProgress,
			setupMock:   func(repo *mockRepository) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name:      "repository error",
			inputName: "test-sm",
			inputVer:  "1.0.0",
			inputLoc:  models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.readStateMachineFunc = func(name, version string, location models.Location) (*models.StateMachine, error) {
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

			result, err := svc.Read(tt.inputName, tt.inputVer, tt.inputLoc)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Read() expected error but got none")
					return
				}

				var smErr *models.StateMachineError
				if !errors.As(err, &smErr) {
					t.Errorf("Read() expected StateMachineError but got %T", err)
					return
				}

				if smErr.Type != tt.wantErrType {
					t.Errorf("Read() expected error type %v but got %v", tt.wantErrType, smErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("Read() unexpected error: %v", err)
					return
				}

				if result == nil {
					t.Error("Read() expected result but got nil")
					return
				}

				if result.Name != tt.wantResult.Name {
					t.Errorf("Read() name = %v, want %v", result.Name, tt.wantResult.Name)
				}
				if result.Version != tt.wantResult.Version {
					t.Errorf("Read() version = %v, want %v", result.Version, tt.wantResult.Version)
				}
				if result.Content != tt.wantResult.Content {
					t.Errorf("Read() content = %v, want %v", result.Content, tt.wantResult.Content)
				}
				if result.Location != tt.wantResult.Location {
					t.Errorf("Read() location = %v, want %v", result.Location, tt.wantResult.Location)
				}
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	tests := []struct {
		name        string
		input       *models.StateMachine
		setupMock   func(*mockRepository)
		wantErr     bool
		wantErrType models.ErrorType
	}{
		{
			name: "successful update",
			input: &models.StateMachine{
				Name:     "test-sm",
				Version:  "1.0.0",
				Content:  "@startuml\n[*] --> Updated\n@enduml",
				Location: models.LocationInProgress,
			},
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
					return true, nil // exists
				}
				repo.writeStateMachineFunc = func(sm *models.StateMachine) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:        "nil state machine validation",
			input:       nil,
			setupMock:   func(repo *mockRepository) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name: "empty name validation",
			input: &models.StateMachine{
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
			input: &models.StateMachine{
				Name:     "test-sm",
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
			input: &models.StateMachine{
				Name:     "test-sm",
				Version:  "1.0.0",
				Content:  "",
				Location: models.LocationInProgress,
			},
			setupMock:   func(repo *mockRepository) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name: "state machine does not exist",
			input: &models.StateMachine{
				Name:     "test-sm",
				Version:  "1.0.0",
				Content:  "content",
				Location: models.LocationInProgress,
			},
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
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

			err := svc.Update(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Update() expected error but got none")
					return
				}

				var smErr *models.StateMachineError
				if !errors.As(err, &smErr) {
					t.Errorf("Update() expected StateMachineError but got %T", err)
					return
				}

				if smErr.Type != tt.wantErrType {
					t.Errorf("Update() expected error type %v but got %v", tt.wantErrType, smErr.Type)
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
			inputName: "test-sm",
			inputVer:  "1.0.0",
			inputLoc:  models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
					return true, nil // exists
				}
				repo.deleteStateMachineFunc = func(name, version string, location models.Location) error {
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
			inputName:   "test-sm",
			inputVer:    "",
			inputLoc:    models.LocationInProgress,
			setupMock:   func(repo *mockRepository) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name:      "state machine does not exist",
			inputName: "test-sm",
			inputVer:  "1.0.0",
			inputLoc:  models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
					return false, nil // doesn't exist
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeFileNotFound,
		},
		{
			name:      "repository delete error",
			inputName: "test-sm",
			inputVer:  "1.0.0",
			inputLoc:  models.LocationInProgress,
			setupMock: func(repo *mockRepository) {
				repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
					return true, nil // exists
				}
				repo.deleteStateMachineFunc = func(name, version string, location models.Location) error {
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

			err := svc.Delete(tt.inputName, tt.inputVer, tt.inputLoc)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Delete() expected error but got none")
					return
				}

				var smErr *models.StateMachineError
				if !errors.As(err, &smErr) {
					t.Errorf("Delete() expected StateMachineError but got %T", err)
					return
				}

				if smErr.Type != tt.wantErrType {
					t.Errorf("Delete() expected error type %v but got %v", tt.wantErrType, smErr.Type)
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
			inputName: "test-sm",
			inputVer:  "1.0.0",
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				// State machine exists in in-progress
				repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
					if location == models.LocationInProgress {
						return true, nil
					}
					if location == models.LocationProducts {
						return false, nil // doesn't exist in products initially
					}
					return false, nil
				}

				// Read state machine for validation
				repo.readStateMachineFunc = func(name, version string, location models.Location) (*models.StateMachine, error) {
					return &models.StateMachine{
						Name:     name,
						Version:  version,
						Content:  "@startuml\n[*] --> Idle\n@enduml",
						Location: location,
					}, nil
				}

				// Validation passes
				validator.validateFunc = func(sm *models.StateMachine, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
					return &models.ValidationResult{
						IsValid:  true,
						Errors:   []models.ValidationError{},
						Warnings: []models.ValidationWarning{},
					}, nil
				}

				// Move operation succeeds
				repo.moveStateMachineFunc = func(name, version string, from, to models.Location) error {
					return nil
				}

				// After move verification - track call count to simulate state change
				callCount := 0
				repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
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
			inputName:   "test-sm",
			inputVer:    "",
			setupMock:   func(repo *mockRepository, validator *mockValidator) {},
			wantErr:     true,
			wantErrType: models.ErrorTypeValidation,
		},
		{
			name:      "state machine does not exist in in-progress",
			inputName: "test-sm",
			inputVer:  "1.0.0",
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
					return false, nil // doesn't exist anywhere
				}
			},
			wantErr:     true,
			wantErrType: models.ErrorTypeFileNotFound,
		},
		{
			name:      "directory conflict in products",
			inputName: "test-sm",
			inputVer:  "1.0.0",
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
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
			inputName: "test-sm",
			inputVer:  "1.0.0",
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				// State machine exists in in-progress
				repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
					if location == models.LocationInProgress {
						return true, nil
					}
					return false, nil
				}

				// Read state machine for validation
				repo.readStateMachineFunc = func(name, version string, location models.Location) (*models.StateMachine, error) {
					return &models.StateMachine{
						Name:     name,
						Version:  version,
						Content:  "invalid plantuml",
						Location: location,
					}, nil
				}

				// Validation fails with errors
				validator.validateFunc = func(sm *models.StateMachine, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
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
			inputName: "test-sm",
			inputVer:  "1.0.0",
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				// State machine exists in in-progress
				repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
					if location == models.LocationInProgress {
						return true, nil
					}
					return false, nil
				}

				// Read state machine for validation
				repo.readStateMachineFunc = func(name, version string, location models.Location) (*models.StateMachine, error) {
					return &models.StateMachine{
						Name:     name,
						Version:  version,
						Content:  "@startuml\n[*] --> Idle\n@enduml",
						Location: location,
					}, nil
				}

				// Validation passes
				validator.validateFunc = func(sm *models.StateMachine, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
					return &models.ValidationResult{
						IsValid:  true,
						Errors:   []models.ValidationError{},
						Warnings: []models.ValidationWarning{},
					}, nil
				}

				// Move operation fails
				repo.moveStateMachineFunc = func(name, version string, from, to models.Location) error {
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

			err := svc.Promote(tt.inputName, tt.inputVer)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Promote() expected error but got none")
					return
				}

				var smErr *models.StateMachineError
				if !errors.As(err, &smErr) {
					t.Errorf("Promote() expected StateMachineError but got %T", err)
					return
				}

				if smErr.Type != tt.wantErrType {
					t.Errorf("Promote() expected error type %v but got %v", tt.wantErrType, smErr.Type)
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
			inputName: "test-sm",
			inputVer:  "1.0.0",
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				// Initial setup - state machine exists in in-progress
				initialCallCount := 0
				repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
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

				// Read state machine for validation
				repo.readStateMachineFunc = func(name, version string, location models.Location) (*models.StateMachine, error) {
					return &models.StateMachine{
						Name:     name,
						Version:  version,
						Content:  "@startuml\n[*] --> Idle\n@enduml",
						Location: location,
					}, nil
				}

				// Validation passes
				validator.validateFunc = func(sm *models.StateMachine, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
					return &models.ValidationResult{
						IsValid:  true,
						Errors:   []models.ValidationError{},
						Warnings: []models.ValidationWarning{},
					}, nil
				}

				// Move operation appears to succeed initially
				moveCallCount := 0
				repo.moveStateMachineFunc = func(name, version string, from, to models.Location) error {
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
			inputName: "test-sm",
			inputVer:  "1.0.0",
			setupMock: func(repo *mockRepository, validator *mockValidator) {
				// Initial setup - state machine exists in in-progress
				initialCallCount := 0
				repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
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

				// Read state machine for validation
				repo.readStateMachineFunc = func(name, version string, location models.Location) (*models.StateMachine, error) {
					return &models.StateMachine{
						Name:     name,
						Version:  version,
						Content:  "@startuml\n[*] --> Idle\n@enduml",
						Location: location,
					}, nil
				}

				// Validation passes
				validator.validateFunc = func(sm *models.StateMachine, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
					return &models.ValidationResult{
						IsValid:  true,
						Errors:   []models.ValidationError{},
						Warnings: []models.ValidationWarning{},
					}, nil
				}

				// Move operation appears to succeed
				repo.moveStateMachineFunc = func(name, version string, from, to models.Location) error {
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

			err := svc.Promote(tt.inputName, tt.inputVer)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Promote() expected error but got none")
					return
				}

				var smErr *models.StateMachineError
				if !errors.As(err, &smErr) {
					t.Errorf("Promote() expected StateMachineError but got %T", err)
					return
				}

				if smErr.Type != tt.wantErrType {
					t.Errorf("Promote() expected error type %v but got %v", tt.wantErrType, smErr.Type)
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
		validationFunc func(*models.StateMachine, models.ValidationStrictness) (*models.ValidationResult, error)
		wantErr        bool
		wantErrType    models.ErrorType
	}{
		{
			name:      "validation passes with warnings only",
			inputName: "test-sm",
			inputVer:  "1.0.0",
			validationFunc: func(sm *models.StateMachine, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
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
			inputName: "test-sm",
			inputVer:  "1.0.0",
			validationFunc: func(sm *models.StateMachine, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
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
			inputName: "test-sm",
			inputVer:  "1.0.0",
			validationFunc: func(sm *models.StateMachine, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
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
			repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
				if location == models.LocationInProgress {
					return true, nil
				}
				return false, nil
			}

			repo.readStateMachineFunc = func(name, version string, location models.Location) (*models.StateMachine, error) {
				return &models.StateMachine{
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
				repo.existsFunc = func(name, version string, location models.Location) (bool, error) {
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

				repo.moveStateMachineFunc = func(name, version string, from, to models.Location) error {
					return nil
				}
			}

			svc := NewService(repo, validator, nil)

			err := svc.Promote(tt.inputName, tt.inputVer)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Promote() expected error but got none")
					return
				}

				var smErr *models.StateMachineError
				if !errors.As(err, &smErr) {
					t.Errorf("Promote() expected StateMachineError but got %T", err)
					return
				}

				if smErr.Type != tt.wantErrType {
					t.Errorf("Promote() expected error type %v but got %v", tt.wantErrType, smErr.Type)
				}
			} else {
				if err != nil {
					t.Errorf("Promote() unexpected error: %v", err)
				}
			}
		})
	}
}
