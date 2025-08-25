package models

// Repository interface defines the contract for data persistence operations
type Repository interface {
	// Read operations
	ReadStateMachine(fileType FileType, name, version string, location Location) (*StateMachine, error)
	ListStateMachines(fileType FileType, location Location) ([]StateMachine, error)
	Exists(fileType FileType, name, version string, location Location) (bool, error)

	// Write operations
	WriteStateMachine(sm *StateMachine) error
	MoveStateMachine(fileType FileType, name, version string, from, to Location) error
	DeleteStateMachine(fileType FileType, name, version string, location Location) error

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
	Create(fileType FileType, name, version string, content string, location Location) (*StateMachine, error)
	Read(fileType FileType, name, version string, location Location) (*StateMachine, error)
	Update(sm *StateMachine) error
	Delete(fileType FileType, name, version string, location Location) error

	// Business operations
	Promote(fileType FileType, name, version string) error // Move from in-progress to products
	Validate(fileType FileType, name, version string, location Location) (*ValidationResult, error)
	ListAll(fileType FileType, location Location) ([]StateMachine, error)

	// Reference operations
	ResolveReferences(sm *StateMachine) error
}
