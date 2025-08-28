// Package diagram provides functionality for managing UML state-machine diagram diagrams in PlantUML format.
//
// This package offers a comprehensive solution for organizing, validating, and managing PlantUML state-machine diagram files
// within a structured directory hierarchy. It supports versioning, validation levels based on deployment status,
// and cross-references between state-machine diagrams.
//
// # Key Features
//
//   - Structured file organization with separate locations for in-progress and production state-machine diagrams
//   - Semantic versioning support with automatic version parsing and comparison
//   - PlantUML validation with configurable strictness levels
//   - Reference resolution for cross-dependencies between state-machine diagrams
//   - Safe promotion workflow from development to production
//   - Thread-safe operations with comprehensive error handling
//
// # Basic Usage
//
//	// Create a new state-machine diagram service
//	svc, err := diagram.NewService()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create a state-machine diagram
//	content := `@startuml
//	[*] --> Idle
//	Idle --> Active : start()
//	Active --> Idle : stop()
//	@enduml`
//
//	diag, err := svc.CreateFile("my-machine", "1.0.0", content, diagram.LocationFileInProgress)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Validate and promote to production
//	result, err := svc.ValidateFile("my-machine", "1.0.0", diagram.LocationFileInProgress)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if result.IsValid && !result.HasErrors() {
//	    err = svc.Promote("my-machine", "1.0.0")
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	}
//
// # Directory Structure
//
// The package organizes state-machine diagrams in a standardized directory structure:
//
//	.go-uml-statemachine-parsers/
//	├── in-progress/
//	│   └── {name}-{version}/
//	│       └── {name}-{version}.puml
//	└── products/
//	    └── {name}-{version}/
//	        └── {name}-{version}.puml
//
// # Configuration
//
// The package supports configuration through environment variables:
//
//   - GO_UML_ROOT_DIRECTORY: Root directory for state-machine diagrams (default: ".go-uml-statemachine-parsers")
//   - GO_UML_VALIDATION_LEVEL: Validation level ("in-progress" or "products")
//   - GO_UML_BACKUP_ENABLED: Enable backups ("true" or "false")
//   - GO_UML_MAX_FILE_SIZE: Maximum file size in bytes
//   - GO_UML_DEBUG_LOGGING: Enable debug logging ("true" or "false")
package diagram

import (
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/repository"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/service"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/validation"
)

// Re-export key types for public API

// Location indicates where the state-machine diagram is stored.
type Location = models.Location

// Location constants for state-machine diagram storage locations.
const (
	// LocationFileInProgress indicates the state-machine diagram is in development/testing phase.
	LocationFileInProgress = models.LocationFileInProgress
	// LocationFileProducts indicates the state-machine diagram is production-ready.
	LocationFileProducts = models.LocationFileProducts
)

// ReferenceType indicates the type of reference between state-machine diagrams.
type ReferenceType = models.ReferenceType

// Reference type constants.
const (
	// ReferenceTypeProduct indicates a reference to a state-machine diagram in the products directory.
	ReferenceTypeProduct = models.ReferenceTypeProduct
)

// ValidationStrictness defines the level of validation strictness.
type ValidationStrictness = models.ValidationStrictness

// Validation strictness constants.
const (
	// StrictnessInProgress applies strict validation with both errors and warnings.
	// Used for in-progress state-machine diagrams to ensure quality before promotion.
	StrictnessInProgress = models.StrictnessInProgress
	// StrictnessProducts applies lenient validation with warnings only.
	// Used for production state-machine diagrams to allow operational flexibility.
	StrictnessProducts = models.StrictnessProducts
)

// StateMachineDiagram represents a UML state-machine diagram with its metadata and content.
type StateMachineDiagram = models.StateMachineDiagram

// Reference represents a reference to another state-machine diagram.
type Reference = models.Reference

// Metadata contains additional information about a state-machine diagram.
type Metadata = models.Metadata

// ValidationResult contains the outcome of state-machine diagram validation.
type ValidationResult = models.ValidationResult

// ValidationError represents a validation error that prevents promotion.
type ValidationError = models.ValidationError

// ValidationWarning represents a validation warning that doesn't prevent promotion.
type ValidationWarning = models.ValidationWarning

// Config represents the configuration for the state-machine diagram system.
type Config = models.Config

// DiagramService defines the interface for state-machine diagram operations.
//
// This interface provides all the functionality needed to manage state-machine diagrams
// including CRUD operations, validation, promotion, and reference resolution.
type DiagramService = models.DiagramService

// NewService creates a new DiagramService with default configuration.
//
// This is the recommended way to create a service instance for most use cases.
// It uses the default configuration and creates all necessary dependencies.
//
// Returns an error if the service cannot be initialized (e.g., due to file system issues).
//
// Example:
//
//	svc, err := diagram.NewService()
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewService() (DiagramService, error) {
	config := models.DefaultConfig()
	repo := repository.NewFileSystemRepository(config)
	validator := validation.NewPlantUMLValidatorWithRepository(repo)
	return service.NewService(repo, validator, config), nil
}

// NewServiceWithConfig creates a new DiagramService with the provided configuration.
//
// Use this function when you need custom configuration settings such as a different
// root directory, validation level, or logging preferences.
//
// Parameters:
//   - config: Configuration settings for the service. If nil, default configuration is used.
//
// Returns an error if the service cannot be initialized with the provided configuration.
//
// Example:
//
//	config := &diagram.Config{
//	    RootDirectory:      "custom-directory",
//	    EnableDebugLogging: true,
//	    MaxFileSize:        2 * 1024 * 1024, // 2MB
//	}
//	svc, err := diagram.NewServiceWithConfig(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewServiceWithConfig(config *Config) (DiagramService, error) {
	if config == nil {
		config = models.DefaultConfig()
	}
	repo := repository.NewFileSystemRepository(config)
	validator := validation.NewPlantUMLValidatorWithRepository(repo)
	return service.NewService(repo, validator, config), nil
}

// NewServiceFromEnv creates a new DiagramService with configuration loaded from environment variables.
//
// This function reads configuration from environment variables, making it ideal for
// deployment scenarios where configuration is managed externally.
//
// Environment variables:
//   - GO_UML_ROOT_DIRECTORY: Root directory for state-machine diagrams
//   - GO_UML_VALIDATION_LEVEL: Validation level ("in-progress" or "products")
//   - GO_UML_BACKUP_ENABLED: Enable backups ("true" or "false")
//   - GO_UML_MAX_FILE_SIZE: Maximum file size in bytes
//   - GO_UML_DEBUG_LOGGING: Enable debug logging ("true" or "false")
//
// Returns an error if the service cannot be initialized.
//
// Example:
//
//	// Set environment variables
//	os.Setenv("GO_UML_ROOT_DIRECTORY", "my-diagrams")
//	os.Setenv("GO_UML_DEBUG_LOGGING", "true")
//
//	svc, err := diagram.NewServiceFromEnv()
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewServiceFromEnv() (DiagramService, error) {
	config := models.LoadConfigFromEnv()
	repo := repository.NewFileSystemRepository(config)
	validator := validation.NewPlantUMLValidatorWithRepository(repo)
	return service.NewService(repo, validator, config), nil
}

// DefaultConfig returns a configuration with default values.
//
// Default values:
//   - RootDirectory: ".go-uml-statemachine-parsers"
//   - ValidationLevel: StrictnessInProgress
//   - BackupEnabled: false
//   - MaxFileSize: 1MB
//   - EnableDebugLogging: false
//
// Example:
//
//	config := diagram.DefaultConfig()
//	config.EnableDebugLogging = true
//	svc, err := diagram.NewServiceWithConfig(config)
func DefaultConfig() *Config {
	return models.DefaultConfig()
}

// LoadConfigFromEnv loads configuration from environment variables.
//
// This function creates a new configuration by reading from environment variables.
// If an environment variable is not set, the corresponding default value is used.
//
// See NewServiceFromEnv for the list of supported environment variables.
//
// Example:
//
//	config := diagram.LoadConfigFromEnv()
//	fmt.Printf("Root directory: %s\n", config.RootDirectory)
func LoadConfigFromEnv() *Config {
	return models.LoadConfigFromEnv()
}
