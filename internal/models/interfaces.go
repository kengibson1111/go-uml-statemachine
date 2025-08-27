package models

import smmodels "github.com/kengibson1111/go-uml-statemachine-models/models"

// Repository interface defines the contract for data persistence operations
type Repository interface {
	// Read operations
	ReadDiagram(diagramType smmodels.DiagramType, name, version string, location Location) (*StateMachineDiagram, error)
	ListStateMachines(diagramType smmodels.DiagramType, location Location) ([]StateMachineDiagram, error)
	Exists(diagramType smmodels.DiagramType, name, version string, location Location) (bool, error)

	// Write operations
	WriteDiagram(diag *StateMachineDiagram) error
	MoveDiagram(diagramType smmodels.DiagramType, name, version string, from, to Location) error
	DeleteDiagram(diagramType smmodels.DiagramType, name, version string, location Location) error

	// Directory operations
	CreateDirectory(path string) error
	DirectoryExists(path string) (bool, error)
}

// Validator interface defines the contract for state-machine diagram validation
type Validator interface {
	Validate(diagram *StateMachineDiagram, strictness ValidationStrictness) (*ValidationResult, error)
	ValidateReferences(diagram *StateMachineDiagram) (*ValidationResult, error)
}

// DiagramService interface defines the contract for business operations
type DiagramService interface {
	// CRUD operations
	CreateFile(diagramType smmodels.DiagramType, name, version string, content string, location Location) (*StateMachineDiagram, error)
	ReadFile(diagramType smmodels.DiagramType, name, version string, location Location) (*StateMachineDiagram, error)
	UpdateInProgressFile(diag *StateMachineDiagram) error
	DeleteFile(diagramType smmodels.DiagramType, name, version string, location Location) error

	// Business operations
	PromoteToProductsFile(diagramType smmodels.DiagramType, name, version string) error // Move from in-progress to products
	ValidateFile(diagramType smmodels.DiagramType, name, version string, location Location) (*ValidationResult, error)
	ListAllFiles(diagramType smmodels.DiagramType, location Location) ([]StateMachineDiagram, error)

	// Reference operations
	ResolveFileReferences(diagram *StateMachineDiagram) error
}
