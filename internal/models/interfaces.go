package models

// Repository interface defines the contract for data persistence operations
type Repository interface {
	// Read operations
	ReadStateMachine(name, version string, location Location) (*StateMachine, error)
	ListStateMachines(location Location) ([]StateMachine, error)
	Exists(name, version string, location Location) (bool, error)

	// Write operations
	WriteStateMachine(sm *StateMachine) error
	MoveStateMachine(name, version string, from, to Location) error
	DeleteStateMachine(name, version string, location Location) error

	// Directory operations
	CreateDirectory(path string) error
	DirectoryExists(path string) (bool, error)
}

// Validator interface defines the contract for state machine validation
type Validator interface {
	Validate(sm *StateMachine, strictness ValidationStrictness) (*ValidationResult, error)
	ValidateReferences(sm *StateMachine) (*ValidationResult, error)
}

// StateMachineService interface defines the contract for business operations
type StateMachineService interface {
	// CRUD operations
	Create(name, version string, content string, location Location) (*StateMachine, error)
	Read(name, version string, location Location) (*StateMachine, error)
	Update(sm *StateMachine) error
	Delete(name, version string, location Location) error

	// Business operations
	Promote(name, version string) error // Move from in-progress to products
	Validate(name, version string, location Location) (*ValidationResult, error)
	ListAll(location Location) ([]StateMachine, error)

	// Reference operations
	ResolveReferences(sm *StateMachine) error
}
