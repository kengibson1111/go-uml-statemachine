package models

import "time"

// Location indicates where the state machine is stored
type Location int

const (
	LocationInProgress Location = iota
	LocationProducts
	LocationNested
)

// String returns the string representation of Location
func (l Location) String() string {
	switch l {
	case LocationInProgress:
		return "in-progress"
	case LocationProducts:
		return "products"
	case LocationNested:
		return "nested"
	default:
		return "unknown"
	}
}

// ReferenceType indicates the type of reference
type ReferenceType int

const (
	ReferenceTypeProduct ReferenceType = iota
	ReferenceTypeNested
)

// String returns the string representation of ReferenceType
func (rt ReferenceType) String() string {
	switch rt {
	case ReferenceTypeProduct:
		return "product"
	case ReferenceTypeNested:
		return "nested"
	default:
		return "unknown"
	}
}

// StateMachine represents a UML state machine
type StateMachine struct {
	Name       string
	Version    string
	Content    string
	References []Reference
	Location   Location
	Metadata   Metadata
}

// Reference represents a reference to another state machine
type Reference struct {
	Name    string
	Version string // empty for nested references
	Type    ReferenceType
	Path    string
}

// Metadata contains additional information about the state machine
type Metadata struct {
	CreatedAt  time.Time
	ModifiedAt time.Time
	Author     string
	Tags       []string
}
