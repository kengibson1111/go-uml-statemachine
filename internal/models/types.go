package models

import (
	"time"

	smmodels "github.com/kengibson1111/go-uml-statemachine-models/models"
)

// Location indicates where the state-machine diagram is stored
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

// StateMachineDiagram represents a UML state-machine diagram
type StateMachineDiagram struct {
	Name        string
	Version     string
	Content     string
	References  []Reference
	Location    Location
	DiagramType smmodels.DiagramType
	Metadata    Metadata
}

// Reference represents a reference to another state-machine diagram
type Reference struct {
	Name    string
	Version string // empty for nested references
	Type    ReferenceType
	Path    string
}

// Metadata contains additional information about the state-machine diagram
type Metadata struct {
	CreatedAt  time.Time
	ModifiedAt time.Time
	Author     string
	Tags       []string
}
