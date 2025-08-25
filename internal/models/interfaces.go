package models

// Repository interface defines the contract for data persistence operations
type Repository interface {
	// Read operations
	ReadStateMachine(fileType FileType, name, version string, location Location) (*StateMachineDiagram, error)
	ListStateMachines(fileType FileType, location Location) ([]StateMachineDiagram, error)
	Exists(fileType FileType, name, version string, location Location) (bool, error)

	// Write operations
	WriteStateMachine(diag *StateMachineDiagram) error
	MoveStateMachine(fileType FileType, name, version string, from, to Location) error
	DeleteStateMachine(fileType FileType, name, version string, location Location) error

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
	Create(fileType FileType, name, version string, content string, location Location) (*StateMachineDiagram, error)
	Read(fileType FileType, name, version string, location Location) (*StateMachineDiagram, error)
	Update(diag *StateMachineDiagram) error
	Delete(fileType FileType, name, version string, location Location) error

	// Business operations
	Promote(fileType FileType, name, version string) error // Move from in-progress to products
	Validate(fileType FileType, name, version string, location Location) (*ValidationResult, error)
	ListAll(fileType FileType, location Location) ([]StateMachineDiagram, error)

	// Reference operations
	ResolveReferences(diagram *StateMachineDiagram) error
}
