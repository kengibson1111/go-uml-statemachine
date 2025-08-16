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
