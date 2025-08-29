package service

import (
	"errors"
	"testing"

	smmodels "github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/models"
)

// MockRepository for testing error scenarios
type MockErrorRepository struct {
	shouldFailExists    bool
	shouldFailRead      bool
	shouldFailWrite     bool
	shouldFailMove      bool
	shouldFailDelete    bool
	shouldFailList      bool
	shouldFailCreateDir bool
	shouldFailDirExists bool
	existsResult        bool
	dirExistsResult     bool
	readResult          *models.StateMachineDiagram
	listResult          []models.StateMachineDiagram

	// Function fields for custom behavior
	ExistsFunc func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error)
}

func (m *MockErrorRepository) Exists(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(diagramType, name, version, location)
	}
	if m.shouldFailExists {
		return false, errors.New("mock exists error")
	}
	return m.existsResult, nil
}

func (m *MockErrorRepository) ReadDiagram(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
	if m.shouldFailRead {
		return nil, errors.New("mock read error")
	}
	return m.readResult, nil
}

func (m *MockErrorRepository) WriteDiagram(diag *models.StateMachineDiagram) error {
	if m.shouldFailWrite {
		return errors.New("mock write error")
	}
	return nil
}

func (m *MockErrorRepository) MoveDiagram(diagramType smmodels.DiagramType, name, version string, from, to models.Location) error {
	if m.shouldFailMove {
		return errors.New("mock move error")
	}
	return nil
}

func (m *MockErrorRepository) DeleteDiagram(diagramType smmodels.DiagramType, name, version string, location models.Location) error {
	if m.shouldFailDelete {
		return errors.New("mock delete error")
	}
	return nil
}

func (m *MockErrorRepository) ListDiagrams(diagramType smmodels.DiagramType, location models.Location) ([]models.StateMachineDiagram, error) {
	if m.shouldFailList {
		return nil, errors.New("mock list error")
	}
	return m.listResult, nil
}

func (m *MockErrorRepository) CreateDirectory(path string) error {
	if m.shouldFailCreateDir {
		return errors.New("mock create directory error")
	}
	return nil
}

func (m *MockErrorRepository) DirectoryExists(path string) (bool, error) {
	if m.shouldFailDirExists {
		return false, errors.New("mock directory exists error")
	}
	return m.dirExistsResult, nil
}

// MockValidator for testing error scenarios
type MockErrorValidator struct {
	shouldFailValidate           bool
	shouldFailValidateReferences bool
	validationResult             *models.ValidationResult
	referencesResult             *models.ValidationResult
}

func (m *MockErrorValidator) Validate(diag *models.StateMachineDiagram, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
	if m.shouldFailValidate {
		return nil, errors.New("mock validation error")
	}
	return m.validationResult, nil
}

func (m *MockErrorValidator) ValidateReferences(diag *models.StateMachineDiagram) (*models.ValidationResult, error) {
	if m.shouldFailValidateReferences {
		return nil, errors.New("mock reference validation error")
	}
	return m.referencesResult, nil
}

func TestService_CreateFile_ValidationErrors(t *testing.T) {
	repo := &MockErrorRepository{}
	validator := &MockErrorValidator{}
	config := models.DefaultConfig()
	svc := NewService(repo, validator, config)

	tests := []struct {
		name         string
		inputName    string
		inputVersion string
		inputContent string
		location     models.Location
		expectedErr  string
	}{
		{
			name:         "empty name",
			inputName:    "",
			inputVersion: "1.0.0",
			inputContent: "content",
			location:     models.LocationFileInProgress,
			expectedErr:  "name cannot be empty",
		},
		{
			name:         "empty version",
			inputName:    "test",
			inputVersion: "",
			inputContent: "content",
			location:     models.LocationFileInProgress,
			expectedErr:  "version cannot be empty",
		},
		{
			name:         "empty content",
			inputName:    "test",
			inputVersion: "1.0.0",
			inputContent: "",
			location:     models.LocationFileInProgress,
			expectedErr:  "content cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateFile(smmodels.DiagramTypePUML, tt.inputName, tt.inputVersion, tt.inputContent, tt.location)

			if err == nil {
				t.Error("Expected error but got nil")
				return
			}

			diagErr, ok := err.(*models.StateMachineError)
			if !ok {
				t.Errorf("Expected StateMachineError, got %T", err)
				return
			}

			if diagErr.Type != models.ErrorTypeValidation {
				t.Errorf("Expected ErrorTypeValidation, got %v", diagErr.Type)
			}

			if diagErr.Operation != "CreateFile" {
				t.Errorf("Expected operation 'Create', got %v", diagErr.Operation)
			}

			if diagErr.Component != "service" {
				t.Errorf("Expected component 'service', got %v", diagErr.Component)
			}

			if diagErr.Severity != models.ErrorSeverityHigh {
				t.Errorf("Expected ErrorSeverityHigh, got %v", diagErr.Severity)
			}
		})
	}
}

func TestService_CreateFile_RepositoryErrors(t *testing.T) {
	tests := []struct {
		name            string
		setupRepo       func(*MockErrorRepository)
		expectedErrType models.ErrorType
		expectedMsg     string
	}{
		{
			name: "exists check fails",
			setupRepo: func(repo *MockErrorRepository) {
				repo.shouldFailExists = true
			},
			expectedErrType: models.ErrorTypeFileSystem,
			expectedMsg:     "failed to check if state-machine diagram exists",
		},
		{
			name: "state-machine diagram already exists",
			setupRepo: func(repo *MockErrorRepository) {
				repo.existsResult = true
			},
			expectedErrType: models.ErrorTypeDirectoryConflict,
			expectedMsg:     "state-machine diagram already exists",
		},
		{
			name: "write fails",
			setupRepo: func(repo *MockErrorRepository) {
				repo.existsResult = false
				repo.shouldFailWrite = true
			},
			expectedErrType: models.ErrorTypeFileSystem,
			expectedMsg:     "failed to write state-machine diagram",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockErrorRepository{}
			tt.setupRepo(repo)

			validator := &MockErrorValidator{}
			config := models.DefaultConfig()
			svc := NewService(repo, validator, config)

			_, err := svc.CreateFile(smmodels.DiagramTypePUML, "test", "1.0.0", "content", models.LocationFileInProgress)

			if err == nil {
				t.Error("Expected error but got nil")
				return
			}

			diagErr, ok := err.(*models.StateMachineError)
			if !ok {
				t.Errorf("Expected StateMachineError, got %T", err)
				return
			}

			if diagErr.Type != tt.expectedErrType {
				t.Errorf("Expected error type %v, got %v", tt.expectedErrType, diagErr.Type)
			}

			if diagErr.Operation != "CreateFile" {
				t.Errorf("Expected operation 'Create', got %v", diagErr.Operation)
			}

			if diagErr.Component != "service" {
				t.Errorf("Expected component 'service', got %v", diagErr.Component)
			}
		})
	}
}

func TestService_CreateFile_ProductConflictCheck(t *testing.T) {
	// Test case where checking products directory fails
	t.Run("products check fails", func(t *testing.T) {
		repo := &MockErrorRepository{}
		validator := &MockErrorValidator{}
		config := models.DefaultConfig()

		// Create a custom exists function that fails on the second call
		callCount := 0
		repo.ExistsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
			callCount++
			if callCount == 2 { // Second call (products check)
				return false, errors.New("mock products check error")
			}
			return false, nil // First call succeeds
		}

		svc := NewService(repo, validator, config)

		_, err := svc.CreateFile(smmodels.DiagramTypePUML, "test", "1.0.0", "content", models.LocationFileInProgress)

		if err == nil {
			t.Error("Expected error but got nil")
			return
		}

		diagErr, ok := err.(*models.StateMachineError)
		if !ok {
			t.Errorf("Expected StateMachineError, got %T", err)
			return
		}

		if diagErr.Type != models.ErrorTypeFileSystem {
			t.Errorf("Expected ErrorTypeFileSystem, got %v", diagErr.Type)
		}
	})

	// Test case where product already exists
	t.Run("product conflict exists", func(t *testing.T) {
		repo := &MockErrorRepository{}
		validator := &MockErrorValidator{}
		config := models.DefaultConfig()

		// Create a custom exists function that returns conflict on second call
		callCount := 0
		repo.ExistsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
			callCount++
			if callCount == 1 { // First call (in-progress check)
				return false, nil
			}
			return true, nil // Second call (products check) - conflict found
		}

		svc := NewService(repo, validator, config)

		_, err := svc.CreateFile(smmodels.DiagramTypePUML, "test", "1.0.0", "content", models.LocationFileInProgress)

		if err == nil {
			t.Error("Expected error but got nil")
			return
		}

		diagErr, ok := err.(*models.StateMachineError)
		if !ok {
			t.Errorf("Expected StateMachineError, got %T", err)
			return
		}

		if diagErr.Type != models.ErrorTypeDirectoryConflict {
			t.Errorf("Expected ErrorTypeDirectoryConflict, got %v", diagErr.Type)
		}
	})
}

func TestService_Read_ValidationErrors(t *testing.T) {
	repo := &MockErrorRepository{}
	validator := &MockErrorValidator{}
	config := models.DefaultConfig()
	svc := NewService(repo, validator, config)

	tests := []struct {
		name         string
		inputName    string
		inputVersion string
		location     models.Location
		expectedErr  string
	}{
		{
			name:         "empty name",
			inputName:    "",
			inputVersion: "1.0.0",
			location:     models.LocationFileInProgress,
			expectedErr:  "name cannot be empty",
		},
		{
			name:         "empty version",
			inputName:    "test",
			inputVersion: "",
			location:     models.LocationFileInProgress,
			expectedErr:  "version cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ReadFile(smmodels.DiagramTypePUML, tt.inputName, tt.inputVersion, tt.location)

			if err == nil {
				t.Error("Expected error but got nil")
				return
			}

			diagErr, ok := err.(*models.StateMachineError)
			if !ok {
				t.Errorf("Expected StateMachineError, got %T", err)
				return
			}

			if diagErr.Type != models.ErrorTypeValidation {
				t.Errorf("Expected ErrorTypeValidation, got %v", diagErr.Type)
			}

			if diagErr.Operation != "ReadFile" {
				t.Errorf("Expected operation 'ReadFile', got %v", diagErr.Operation)
			}
		})
	}
}

func TestService_Read_RepositoryError(t *testing.T) {
	repo := &MockErrorRepository{
		shouldFailRead: true,
	}
	validator := &MockErrorValidator{}
	config := models.DefaultConfig()
	svc := NewService(repo, validator, config)

	_, err := svc.ReadFile(smmodels.DiagramTypePUML, "test", "1.0.0", models.LocationFileInProgress)

	if err == nil {
		t.Error("Expected error but got nil")
		return
	}

	diagErr, ok := err.(*models.StateMachineError)
	if !ok {
		t.Errorf("Expected StateMachineError, got %T", err)
		return
	}

	if diagErr.Type != models.ErrorTypeFileNotFound {
		t.Errorf("Expected ErrorTypeFileNotFound, got %v", diagErr.Type)
	}

	if diagErr.Operation != "ReadFile" {
		t.Errorf("Expected operation 'ReadFile', got %v", diagErr.Operation)
	}

	if diagErr.Component != "service" {
		t.Errorf("Expected component 'service', got %v", diagErr.Component)
	}
}

func TestService_ErrorContextPropagation(t *testing.T) {
	repo := &MockErrorRepository{
		shouldFailExists: true,
	}
	validator := &MockErrorValidator{}
	config := models.DefaultConfig()
	svc := NewService(repo, validator, config)

	_, err := svc.CreateFile(smmodels.DiagramTypePUML, "test-name", "1.2.3", "content", models.LocationFileInProgress)

	if err == nil {
		t.Error("Expected error but got nil")
		return
	}

	diagErr, ok := err.(*models.StateMachineError)
	if !ok {
		t.Errorf("Expected StateMachineError, got %T", err)
		return
	}

	// Check that context was properly set
	if diagErr.Context["name"] != "test-name" {
		t.Errorf("Expected context name 'test-name', got %v", diagErr.Context["name"])
	}

	if diagErr.Context["version"] != "1.2.3" {
		t.Errorf("Expected context version '1.2.3', got %v", diagErr.Context["version"])
	}

	if diagErr.Context["location"] != "in-progress" {
		t.Errorf("Expected context location 'in-progress', got %v", diagErr.Context["location"])
	}
}

func TestService_ErrorWrapping(t *testing.T) {
	originalErr := errors.New("original repository error")
	repo := &MockErrorRepository{}
	repo.ExistsFunc = func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
		return false, originalErr
	}

	validator := &MockErrorValidator{}
	config := models.DefaultConfig()
	svc := NewService(repo, validator, config)

	_, err := svc.CreateFile(smmodels.DiagramTypePUML, "test", "1.0.0", "content", models.LocationFileInProgress)

	if err == nil {
		t.Error("Expected error but got nil")
		return
	}

	_, ok := err.(*models.StateMachineError)
	if !ok {
		t.Errorf("Expected StateMachineError, got %T", err)
		return
	}

	// Check that the original error is wrapped
	if !errors.Is(err, originalErr) {
		t.Error("Expected error to wrap the original error")
	}

	// Check that we can unwrap to get the original error
	unwrapped := errors.Unwrap(err)
	if unwrapped == nil {
		t.Error("Expected to be able to unwrap the error")
	}
}

func TestService_ErrorSeverityAssignment(t *testing.T) {
	tests := []struct {
		name             string
		setupRepo        func(*MockErrorRepository)
		expectedSeverity models.ErrorSeverity
	}{
		{
			name: "validation error - high severity",
			setupRepo: func(repo *MockErrorRepository) {
				// Will trigger validation error for empty name
			},
			expectedSeverity: models.ErrorSeverityHigh,
		},
		{
			name: "conflict error - medium severity",
			setupRepo: func(repo *MockErrorRepository) {
				repo.existsResult = true // State-machine diagram already exists
			},
			expectedSeverity: models.ErrorSeverityMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockErrorRepository{}
			tt.setupRepo(repo)

			validator := &MockErrorValidator{}
			config := models.DefaultConfig()
			svc := NewService(repo, validator, config)

			var err error
			if tt.name == "validation error - high severity" {
				_, err = svc.CreateFile(smmodels.DiagramTypePUML, "", "1.0.0", "content", models.LocationFileInProgress) // Empty name
			} else {
				_, err = svc.CreateFile(smmodels.DiagramTypePUML, "test", "1.0.0", "content", models.LocationFileInProgress)
			}

			if err == nil {
				t.Error("Expected error but got nil")
				return
			}

			diagErr, ok := err.(*models.StateMachineError)
			if !ok {
				t.Errorf("Expected StateMachineError, got %T", err)
				return
			}

			if diagErr.Severity != tt.expectedSeverity {
				t.Errorf("Expected severity %v, got %v", tt.expectedSeverity, diagErr.Severity)
			}
		})
	}
}
