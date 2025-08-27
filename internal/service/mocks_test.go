package service

import (
	smmodels "github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/models"
)

// mockRepository is a mock implementation of the Repository interface for testing
type mockRepository struct {
	readStateMachineFunc   func(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error)
	listDiagramsFunc       func(diagramType smmodels.DiagramType, location models.Location) ([]models.StateMachineDiagram, error)
	existsFunc             func(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error)
	writeStateMachineFunc  func(diag *models.StateMachineDiagram) error
	moveStateMachineFunc   func(diagramType smmodels.DiagramType, name, version string, from, to models.Location) error
	deleteStateMachineFunc func(diagramType smmodels.DiagramType, name, version string, location models.Location) error
	createDirectoryFunc    func(path string) error
	directoryExistsFunc    func(path string) (bool, error)
}

func (m *mockRepository) ReadDiagram(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
	if m.readStateMachineFunc != nil {
		return m.readStateMachineFunc(diagramType, name, version, location)
	}
	return nil, nil
}

func (m *mockRepository) ListDiagrams(diagramType smmodels.DiagramType, location models.Location) ([]models.StateMachineDiagram, error) {
	if m.listDiagramsFunc != nil {
		return m.listDiagramsFunc(diagramType, location)
	}
	return nil, nil
}

func (m *mockRepository) Exists(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
	if m.existsFunc != nil {
		return m.existsFunc(diagramType, name, version, location)
	}
	return false, nil
}

func (m *mockRepository) WriteDiagram(diag *models.StateMachineDiagram) error {
	if m.writeStateMachineFunc != nil {
		return m.writeStateMachineFunc(diag)
	}
	return nil
}

func (m *mockRepository) MoveDiagram(diagramType smmodels.DiagramType, name, version string, from, to models.Location) error {
	if m.moveStateMachineFunc != nil {
		return m.moveStateMachineFunc(diagramType, name, version, from, to)
	}
	return nil
}

func (m *mockRepository) DeleteDiagram(diagramType smmodels.DiagramType, name, version string, location models.Location) error {
	if m.deleteStateMachineFunc != nil {
		return m.deleteStateMachineFunc(diagramType, name, version, location)
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
	validateFunc           func(diag *models.StateMachineDiagram, strictness models.ValidationStrictness) (*models.ValidationResult, error)
	validateReferencesFunc func(diag *models.StateMachineDiagram) (*models.ValidationResult, error)
}

func (m *mockValidator) Validate(diag *models.StateMachineDiagram, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
	if m.validateFunc != nil {
		return m.validateFunc(diag, strictness)
	}
	return &models.ValidationResult{IsValid: true}, nil
}

func (m *mockValidator) ValidateReferences(diag *models.StateMachineDiagram) (*models.ValidationResult, error) {
	if m.validateReferencesFunc != nil {
		return m.validateReferencesFunc(diag)
	}
	return &models.ValidationResult{IsValid: true}, nil
}
