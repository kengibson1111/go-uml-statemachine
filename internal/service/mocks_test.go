package service

import (
	"github.com/kengibson1111/go-uml-statemachine/internal/models"
)

// mockRepository is a mock implementation of the Repository interface for testing
type mockRepository struct {
	readStateMachineFunc   func(name, version string, location models.Location) (*models.StateMachine, error)
	listStateMachinesFunc  func(location models.Location) ([]models.StateMachine, error)
	existsFunc             func(name, version string, location models.Location) (bool, error)
	writeStateMachineFunc  func(sm *models.StateMachine) error
	moveStateMachineFunc   func(name, version string, from, to models.Location) error
	deleteStateMachineFunc func(name, version string, location models.Location) error
	createDirectoryFunc    func(path string) error
	directoryExistsFunc    func(path string) (bool, error)
}

func (m *mockRepository) ReadStateMachine(name, version string, location models.Location) (*models.StateMachine, error) {
	if m.readStateMachineFunc != nil {
		return m.readStateMachineFunc(name, version, location)
	}
	return nil, nil
}

func (m *mockRepository) ListStateMachines(location models.Location) ([]models.StateMachine, error) {
	if m.listStateMachinesFunc != nil {
		return m.listStateMachinesFunc(location)
	}
	return nil, nil
}

func (m *mockRepository) Exists(name, version string, location models.Location) (bool, error) {
	if m.existsFunc != nil {
		return m.existsFunc(name, version, location)
	}
	return false, nil
}

func (m *mockRepository) WriteStateMachine(sm *models.StateMachine) error {
	if m.writeStateMachineFunc != nil {
		return m.writeStateMachineFunc(sm)
	}
	return nil
}

func (m *mockRepository) MoveStateMachine(name, version string, from, to models.Location) error {
	if m.moveStateMachineFunc != nil {
		return m.moveStateMachineFunc(name, version, from, to)
	}
	return nil
}

func (m *mockRepository) DeleteStateMachine(name, version string, location models.Location) error {
	if m.deleteStateMachineFunc != nil {
		return m.deleteStateMachineFunc(name, version, location)
	}
	return nil
}

func (m *mockRepository) CreateDirectory(path string) error {
	if m.createDirectoryFunc != nil {
		return m.createDirectoryFunc(path)
	}
	return nil
}

func (m *mockRepository) DirectoryExists(path string) (bool, error) {
	if m.directoryExistsFunc != nil {
		return m.directoryExistsFunc(path)
	}
	return false, nil
}

// mockValidator is a mock implementation of the Validator interface for testing
type mockValidator struct {
	validateFunc           func(sm *models.StateMachine, strictness models.ValidationStrictness) (*models.ValidationResult, error)
	validateReferencesFunc func(sm *models.StateMachine) (*models.ValidationResult, error)
}

func (m *mockValidator) Validate(sm *models.StateMachine, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
	if m.validateFunc != nil {
		return m.validateFunc(sm, strictness)
	}
	return &models.ValidationResult{IsValid: true}, nil
}

func (m *mockValidator) ValidateReferences(sm *models.StateMachine) (*models.ValidationResult, error) {
	if m.validateReferencesFunc != nil {
		return m.validateReferencesFunc(sm)
	}
	return &models.ValidationResult{IsValid: true}, nil
}
