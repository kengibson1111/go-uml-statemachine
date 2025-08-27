# Go UML State Machine Parsers API Documentation

This document provides comprehensive API documentation for the Go UML State Machine Parsers module.

## Package Overview

The `diagram` package provides functionality for managing UML state-machine diagrams in PlantUML format with structured file organization, versioning, validation, and cross-references.

```go
import "github.com/kengibson1111/go-uml-statemachine-models/models"
import "github.com/kengibson1111/go-uml-statemachine-parsers/diagram"
```

## Core Types

### diagram

Represents a UML state-machine diagram with its metadata and content.

```go
type diagram struct {
    Name        string             // State-machine diagram name
    Version     string             // Semantic version (e.g., "1.0.0")
    Content     string             // PlantUML content
    References  []Reference        // References to other state-machines diagrams
    Location    Location           // Storage location
    DiagramType models.DiagramType // Type of diagram (e.g., PUML)
    Metadata    Metadata           // Additional metadata
}
```

### Location

Indicates where the state-machine diagram is stored.

```go
type Location int

const (
    LocationInProgress Location = iota // Development/testing phase
    LocationProducts                   // Production-ready
)
```

### Reference

Represents a reference to another state-machine diagram.

```go
type Reference struct {
    Name    string        // Referenced state-machine diagram name
    Version string        // Version (required for product references)
    Type    ReferenceType // Type of reference
    Path    string        // Resolved file path
}
```

### ReferenceType

Indicates the type of reference between state-machine diagrams.

```go
type ReferenceType int

const (
    ReferenceTypeProduct ReferenceType = iota // Reference to products directory
)
```

### ValidationResult

Contains the outcome of state-machine diagram validation.

```go
type ValidationResult struct {
    Errors   []ValidationError   // Blocking validation errors
    Warnings []ValidationWarning // Non-blocking warnings
    IsValid  bool               // Overall validation status
}

// HasErrors returns true if there are any validation errors
func (vr *ValidationResult) HasErrors() bool
```

### ValidationStrictness

Defines the level of validation strictness.

```go
type ValidationStrictness int

const (
    StrictnessInProgress ValidationStrictness = iota // Strict validation (errors + warnings)
    StrictnessProducts                               // Lenient validation (warnings only)
)
```

### Config

Represents the configuration for the state-machine diagram system.

```go
type Config struct {
    RootDirectory      string               // Root directory (default: ".go-uml-statemachine-parsers")
    ValidationLevel    ValidationStrictness // Default validation level
    BackupEnabled      bool                 // Whether to create backups
    MaxFileSize        int64                // Maximum file size in bytes
    EnableDebugLogging bool                 // Whether to enable debug logging
}
```

## Service Interface

### DiagramService

The main interface for state-machine diagram operations.

```go
type DiagramService interface {
    // CRUD operations
    CreateFile(diagramType models.DiagramType, name, version string, content string, location Location) (*diagram, error)
    Read(diagramType models.DiagramType, name, version string, location Location) (*diagram, error)
    UpdateInProgressFile(diag *StateMachineDiagram) error
    Delete(diagramType models.DiagramType, name, version string, location Location) error

    // Business operations
    Promote(diagramType models.DiagramType, name, version string) error
    ValidateFile(diagramType models.DiagramType, name, version string, location Location) (*ValidationResult, error)
    ListAll(diagramType models.DiagramType, location Location) ([]diagram, error)

    // Reference operations
    ResolveReferences(diagram *StateMachineDiagram) error
}
```

## Factory Functions

### NewService

Creates a new DiagramService with default configuration.

```go
func NewService() (DiagramService, error)
```

**Example:**
```go
svc, err := diagram.NewService()
if err != nil {
    log.Fatal(err)
}
```

### NewServiceWithConfig

Creates a new DiagramService with custom configuration.

```go
func NewServiceWithConfig(config *Config) (DiagramService, error)
```

**Parameters:**
- `config`: Configuration settings. If nil, default configuration is used.

**Example:**
```go
config := &diagram.Config{
    RootDirectory:      "custom-directory",
    EnableDebugLogging: true,
    MaxFileSize:        2 * 1024 * 1024, // 2MB
}
svc, err := diagram.NewServiceWithConfig(config)
if err != nil {
    log.Fatal(err)
}
```

### NewServiceFromEnv

Creates a new DiagramService with configuration from environment variables.

```go
func NewServiceFromEnv() (DiagramService, error)
```

**Environment Variables:**
- `GO_UML_ROOT_DIRECTORY`: Root directory for state-machine diagrams
- `GO_UML_VALIDATION_LEVEL`: Validation level ("in-progress" or "products")
- `GO_UML_BACKUP_ENABLED`: Enable backups ("true" or "false")
- `GO_UML_MAX_FILE_SIZE`: Maximum file size in bytes
- `GO_UML_DEBUG_LOGGING`: Enable debug logging ("true" or "false")

**Example:**
```go
os.Setenv("GO_UML_ROOT_DIRECTORY", "my-diagrams")
os.Setenv("GO_UML_DEBUG_LOGGING", "true")

svc, err := diagram.NewServiceFromEnv()
if err != nil {
    log.Fatal(err)
}
```

## Configuration Functions

### DefaultConfig

Returns a configuration with default values.

```go
func DefaultConfig() *Config
```

**Default Values:**
- RootDirectory: ".go-uml-statemachine-parsers"
- ValidationLevel: StrictnessInProgress
- BackupEnabled: false
- MaxFileSize: 1MB
- EnableDebugLogging: false

### LoadConfigFromEnv

Loads configuration from environment variables.

```go
func LoadConfigFromEnv() *Config
```

## Service Operations

### CRUD Operations

#### Create

Creates a new state-machine diagram with the specified parameters.

```go
CreateFile(diagramType models.DiagramType, name, version string, content string, location Location) (*diagram, error)
```

**Parameters:**
- `diagramType`: Type of file (e.g., models.DiagramTypePUML)
- `name`: State-machine diagram name (must be non-empty)
- `version`: Semantic version (must be non-empty)
- `content`: PlantUML content (must be non-empty)
- `location`: Storage location

**Returns:**
- `*diagram`: Created state-machine diagram
- `error`: Error if creation fails

**Errors:**
- Validation error if parameters are empty
- Directory conflict if state-machine diagram already exists
- File system error if write operation fails

**Example:**
```go
content := `@startuml
[*] --> Idle
Idle --> Active : start()
Active --> Idle : stop()
@enduml`

diag, err := svc.CreateFile(models.DiagramTypePUML, "my-machine", "1.0.0", content, diagram.LocationInProgress)
if err != nil {
    log.Fatal(err)
}
```

#### Read

Retrieves a state-machine diagram by name, version, and location.

```go
Read(diagramType models.DiagramType, name, version string, location Location) (*diagram, error)
```

**Parameters:**
- `diagramType`: Type of file (e.g., models.DiagramTypePUML)
- `name`: State-machine diagram name
- `version`: State-machine diagram version
- `location`: Storage location

**Returns:**
- `*diagram`: Retrieved state-machine diagram
- `error`: Error if read fails

**Errors:**
- Validation error if parameters are empty
- File not found error if state-machine diagram doesn't exist

**Example:**
```go
diag, err := svc.Read(models.DiagramTypePUML, "my-machine", "1.0.0", diagram.LocationInProgress)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Content: %s\n", diagram.Content)
```

#### Update

Modifies an existing state-machine diagram. Updates are only allowed for diagrams in the in-progress location.

```go
UpdateInProgressFile(diag *StateMachineDiagram) error
```

**Parameters:**
- `diag`: State-machine diagram to update (must not be nil)

**Returns:**
- `error`: Error if update fails

**Errors:**
- Validation error if state-machine diagram is nil or has empty fields
- Validation error if location is not LocationInProgress
- File not found error if state-machine diagram doesn't exist
- File system error if write operation fails

**Important:** Updates are restricted to diagrams in the in-progress location only. Production diagrams cannot be modified directly and must be updated through the promotion workflow.

**Example:**
```go
diag.Content = updatedPlantUMLContent
err := svc.UpdateInProgressFile(diag)
if err != nil {
    log.Fatal(err)
}
```

#### Delete

Removes a state-machine diagram by name, version, and location.

```go
Delete(diagramType models.DiagramType, name, version string, location Location) error
```

**Parameters:**
- `diagramType`: Type of file (e.g., models.DiagramTypePUML)
- `name`: State-machine diagram name
- `version`: State-machine diagram version
- `location`: Storage location

**Returns:**
- `error`: Error if deletion fails

**Errors:**
- Validation error if parameters are empty
- File not found error if state-machine diagram doesn't exist
- File system error if delete operation fails

**Example:**
```go
err := svc.DeleteFile(models.DiagramTypePUML, "my-machine", "1.0.0", diagram.LocationInProgress)
if err != nil {
    log.Fatal(err)
}
```

### Business Operations

#### Promote

Moves a state-machine diagram from in-progress to products with validation.

```go
Promote(diagramType models.DiagramType, name, version string) error
```

**Parameters:**
- `diagramType`: Type of file (e.g., models.DiagramTypePUML)
- `name`: State-machine diagram name
- `version`: State-machine diagram version

**Returns:**
- `error`: Error if promotion fails

**Process:**
1. Validates state-machine diagram exists in in-progress
2. Checks for conflicts in products directory
3. Validates state-machine diagram content
4. Performs atomic move operation
5. Includes rollback capability on failure

**Errors:**
- Validation error if parameters are empty
- File not found error if state-machine diagram doesn't exist in in-progress
- Directory conflict error if same name exists in products
- Validation error if state-machine diagram has validation errors

**Example:**
```go
err := svc.Promote(models.DiagramTypePUML, "my-machine", "1.0.0")
if err != nil {
    log.Fatal(err)
}
```

#### ValidateFile

Validates a state-machine diagram with the specified strictness level.

```go
ValidateFile(diagramType models.DiagramType, name, version string, location Location) (*ValidationResult, error)
```

**Parameters:**
- `diagramType`: Type of file (e.g., models.DiagramTypePUML)
- `name`: State-machine diagram name
- `version`: State-machine diagram version
- `location`: Storage location

**Returns:**
- `*ValidationResult`: Validation results
- `error`: Error if validation process fails

**Strictness Levels:**
- `LocationInProgress`: Uses `StrictnessInProgress` (errors and warnings)
- `LocationProducts`: Uses `StrictnessProducts` (warnings only)

**Example:**
```go
result, err := svc.ValidateFile(models.DiagramTypePUML, "my-machine", "1.0.0", diagram.LocationInProgress)
if err != nil {
    log.Fatal(err)
}

if result.HasErrors() {
    fmt.Println("Validation failed:")
    for _, err := range result.Errors {
        fmt.Printf("  - %s: %s\n", err.Code, err.Message)
    }
}
```

#### ListAll

Lists all state-machine diagrams in the specified location.

```go
ListAll(diagramType models.DiagramType, location Location) ([]diagram, error)
```

**Parameters:**
- `diagramType`: Type of file (e.g., models.DiagramTypePUML)
- `location`: Storage location to list

**Returns:**
- `[]diagram`: List of state-machine diagrams
- `error`: Error if listing fails

**Example:**
```go
diagrams, err := svc.ListAll(models.DiagramTypePUML, diagram.LocationInProgress)
if err != nil {
    log.Fatal(err)
}

for _, diag := range diagrams {
    fmt.Printf("- %s-%s\n", diag.Name, diag.Version)
}
```

### Reference Operations

#### ResolveReferences

Resolves all references in a state-machine diagram.

```go
ResolveReferences(diag *StateMachineDiagram) error
```

**Parameters:**
- `diag`: State-machine diagram with references to resolve

**Returns:**
- `error`: Error if reference resolution fails

**Process:**
1. Parses references from PlantUML content if not already done
2. Resolves each reference by checking existence
3. Sets resolved paths for valid references

**Reference Types:**
- **Product References**: References to state-machine diagrams in products directory

**Example:**
```go
err := svc.ResolveReferences(diagram)
if err != nil {
    log.Fatal(err)
}

for _, ref := range diagram.References {
    fmt.Printf("Reference: %s (type: %s, path: %s)\n", 
        ref.Name, ref.Type.String(), ref.Path)
}
```

## Error Handling

The module provides comprehensive error handling with context information.

### Error Types

```go
type ErrorType int

const (
    ErrorTypeFileNotFound ErrorType = iota
    ErrorTypeValidation
    ErrorTypeDirectoryConflict
    ErrorTypeReferenceResolution
    ErrorTypeFileSystem
    ErrorTypeVersionParsing
)
```

### Error Context

Errors include context information such as:
- Operation being performed
- Component that generated the error
- Severity level
- Additional context parameters

### Example Error Handling

```go
diag, err := svc.Read("non-existent", "1.0.0", diagram.LocationInProgress)
if err != nil {
    // Check if it's a specific error type
    if diagErr, ok := err.(*models.StateMachineError); ok {
        switch diagErr.Type {
        case models.ErrorTypeFileNotFound:
            fmt.Println("State-machine diagram not found")
        case models.ErrorTypeValidation:
            fmt.Println("Validation error")
        case models.ErrorTypeDirectoryConflict:
            fmt.Println("Directory conflict")
        default:
            fmt.Printf("Other error: %v\n", err)
        }
    }
}
```

## Best Practices

### 1. Version Management
- Use semantic versioning (e.g., "1.0.0", "1.2.3")
- Increment versions appropriately for changes
- Consider backward compatibility when updating

### 2. Content Validation
- Always validate state-machine diagrams before promotion
- Handle validation errors appropriately
- Use appropriate strictness levels for different environments

### 3. Update Restrictions
- Updates are only allowed for diagrams in the in-progress location
- Production diagrams are immutable and cannot be modified directly
- To update production diagrams, modify the in-progress version and promote it

### 4. Reference Management
- Resolve references after creating state-machine diagrams with dependencies
- Ensure referenced state-machine diagrams exist before creating references
- Use product references for stable dependencies

### 5. Error Handling
- Check and handle all error conditions
- Use appropriate error types for different scenarios
- Provide meaningful error messages to users

### 6. Configuration
- Use environment variables for deployment-specific settings
- Set appropriate file size limits
- Enable debug logging for troubleshooting

### 7. Resource Management
- Clean up test data in examples and tests
- Handle concurrent access appropriately
- Monitor file system usage

## Thread Safety

The service implementation is thread-safe and uses mutex locks to protect concurrent operations. Multiple goroutines can safely use the same service instance.

## Performance Considerations

- State-machine diagram content is loaded on-demand
- File system operations are optimized for common use cases
- Validation is performed efficiently with configurable strictness
- Reference resolution is cached where appropriate

## Limitations

- Maximum file size is configurable (default: 1MB)
- Directory depth is limited by file system constraints
- PlantUML syntax validation is basic (focused on structure)
- Reference resolution requires referenced state-machine diagrams to exist
