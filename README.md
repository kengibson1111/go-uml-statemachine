# Go UML State Machine Parsers

A Go module for managing UML state-machine diagrams. This library provides functionality to read, write, validate, and organize files within a structured directory hierarchy with support for versioning, validation levels, and cross-references between state-machine diagrams.

## Features

- **Structured File Organization**: Automatic directory management with separate locations for in-progress and production-ready state-machine diagrams
- **Version Management**: Semantic versioning support with automatic version parsing and comparison
- **PlantUML Validation**: Configurable validation with different strictness levels based on deployment status
- **Reference Resolution**: Support for cross-references between state-machine diagrams with automatic dependency resolution
- **Promotion Workflow**: Safe promotion of state-machine diagrams from in-progress to production with validation checks
- **Thread-Safe Operations**: Concurrent access protection with mutex locks
- **Comprehensive Error Handling**: Detailed error messages with context information
- **Configurable Logging**: Debug and info level logging with component-specific prefixes

## Installation

```cmd
go get github.com/kengibson1111/go-uml-statemachine-parsers
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/kengibson1111/go-uml-statemachine-models/models"
    "github.com/kengibson1111/go-uml-statemachine-parsers/diagram"
)

func main() {
    // Create a new state-machine diagram service
    svc, err := diagram.NewService()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create a new state-machine diagram
    content := `@startuml
    [*] --> Idle
    Idle --> Active : start()
    Active --> Idle : stop()
    @enduml`
    
    diag, err := svc.CreateFile(models.DiagramTypePUML, "my-machine", "1.0.0", content, diagram.LocationFileInProgress)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Created state-machine diagram: %s-%s\n", diag.Name, diag.Version)
}
```

## Directory Structure

The module organizes state-machine diagrams in a standardized directory structure:

**Windows:**

```text
.go-uml-statemachine-parsers\
├── in-progress\
│   └── puml\
│       └── {name}-{version}.puml
└── products\
    └── puml\
        └── {name}-{version}.puml
```

**Linux/macOS:**

```text
.go-uml-statemachine-parsers/
├── in-progress/
│   └── puml/
│       └── {name}-{version}.puml
└── products/
    └── puml/
        └── {name}-{version}.puml
```

## Configuration

### Default Configuration

```go
config := diagram.DefaultConfig()
// Uses:
// - RootDirectory: ".go-uml-statemachine-parsers"
// - ValidationLevel: StrictnessInProgress
// - BackupEnabled: false
// - MaxFileSize: 1MB
// - EnableDebugLogging: false
```

### Environment Variables

The module supports configuration through environment variables:

- `GO_UML_ROOT_DIRECTORY`: Root directory for state-machine diagrams
- `GO_UML_VALIDATION_LEVEL`: Validation level (`in-progress` or `products`)
- `GO_UML_BACKUP_ENABLED`: Enable backups (`true` or `false`)
- `GO_UML_MAX_FILE_SIZE`: Maximum file size in bytes
- `GO_UML_DEBUG_LOGGING`: Enable debug logging (`true` or `false`)

```go
// Load configuration from environment
config := diagram.LoadConfigFromEnv()

// Or merge with existing config
config := diagram.DefaultConfig()
config.MergeWithEnv()
```

## Core Operations

### Creating State-Machine Diagrams

```go
// Create in in-progress location
diag, err := svc.CreateFile(models.DiagramTypePUML, "user-auth", "1.0.0", plantUMLContent, diagram.LocationFileInProgress)

// Create in products location (for direct production deployment)
diag, err := svc.CreateFile(models.DiagramTypePUML, "user-auth", "1.0.0", plantUMLContent, diagram.LocationProducts)
```

### Reading State-Machine Diagrams

```go
diag, err := svc.Read(models.DiagramTypePUML, "user-auth", "1.0.0", diagram.LocationFileInProgress)
if err != nil {
    log.Printf("Error reading state-machine diagram: %v", err)
}
```

### Updating State-Machine Diagrams

Updates are only allowed for diagrams in the in-progress location. Production diagrams cannot be modified directly.

```go
// Only works for diagrams in LocationFileInProgress
diag.Content = updatedPlantUMLContent
err := svc.UpdateInProgressFile(diag)
if err != nil {
    log.Printf("Update failed: %v", err)
}
```

### Deleting State-Machine Diagrams

```go
err := svc.DeleteFile(models.DiagramTypePUML, "user-auth", "1.0.0", diagram.LocationFileInProgress)
```

### Listing State-Machine Diagrams

```go
// List all in-progress state-machine diagrams
inProgressDiags, err := svc.ListAllFiles(models.DiagramTypePUML, diagram.LocationFileInProgress)

// List all production state-machine diagrams
productDiags, err := svc.ListAllFiles(models.DiagramTypePUML, diagram.LocationProducts)
```

## Validation

The module supports two validation strictness levels:

### In-Progress Validation

- Returns both errors and warnings
- Prevents promotion if errors exist
- Used for development and testing

```go
result, err := svc.Validate(models.DiagramTypePUML, "user-auth", "1.0.0", diagram.LocationFileInProgress)
if result.HasErrors() {
    fmt.Println("Validation failed with errors:")
    for _, err := range result.Errors {
        fmt.Printf("  - %s: %s\n", err.Code, err.Message)
    }
}
```

### Products Validation

- Returns warnings only (no blocking errors)
- Used for production state-machine diagrams
- More lenient to allow operational flexibility

## Promotion Workflow

Move state-machine diagrams from in-progress to products with validation:

```go
// Promotes only if validation passes
err := svc.Promote(models.DiagramTypePUML, "user-auth", "1.0.0")
if err != nil {
    log.Printf("Promotion failed: %v", err)
}
```

The promotion process:

1. Validates the state-machine diagram exists in in-progress
2. Checks for conflicts in products directory
3. Validates the state-machine diagram content
4. Performs atomic move operation
5. Includes rollback capability on failure

## References and Dependencies

State-machine diagrams can reference other state-machine diagrams:

### Product References

Reference state-machine diagrams in the products directory:

**Windows:**

```plantuml
@startuml
!include products\puml\base-auth-1.0.0\base-auth-1.0.0.puml

[*] --> base-auth
base-auth --> TwoFactor : success
@enduml
```

**Linux/macOS:**

```plantuml
@startuml
!include products/puml/base-auth-1.0.0/base-auth-1.0.0.puml

[*] --> base-auth
base-auth --> TwoFactor : success
@enduml
```

### Reference Resolution

```go
err := svc.ResolveFileReferences(diagram)
if err != nil {
    log.Printf("Reference resolution failed: %v", err)
}

// Check resolved references
for _, ref := range diagram.References {
    fmt.Printf("Reference: %s (type: %s, path: %s)\n", 
        ref.Name, ref.Type.String(), ref.Path)
}
```

## Error Handling

The module provides comprehensive error handling with context:

```go
diag, err := svc.Read(models.DiagramTypePUML, "non-existent", "1.0.0", diagram.LocationFileInProgress)
if err != nil {
    // Error includes context about the operation, component, and parameters
    fmt.Printf("Error: %v\n", err)
    return
}

// Use the state-machine diagram
fmt.Printf("Content: %s\n", diagram.Content)
```

## Examples

The module includes examples organized into two categories:

### Public API Examples (examples/)

These examples use the public `diagram` package and demonstrate how external users would interact with the module. **These are the recommended examples for learning how to use the module.**

### Internal Examples (internal/examples/)

These examples use internal packages directly and are primarily for development and testing purposes. They provide deeper insight into the module's internal architecture but are not recommended for typical usage.

## Running Examples

### Public API Examples

These examples demonstrate the recommended way to use the module through its public API:

#### API Test Example

Simple test to verify the public API works correctly.

**Windows Command Prompt:**

```cmd
cd examples\api_test
go run main.go
```

**Windows PowerShell:**

```powershell
Set-Location examples\api_test
go run main.go
```

**Linux/macOS:**

```bash
cd examples/api_test
go run main.go
```

#### Common Use Cases Example

Comprehensive example covering six different use cases including batch operations and environment configuration.

**Windows Command Prompt:**

```cmd
cd examples\common_use_cases
go run main.go
```

**Windows PowerShell:**

```powershell
Set-Location examples\common_use_cases
go run main.go
```

**Linux/macOS:**

```bash
cd examples/common_use_cases
go run main.go
```

#### Comprehensive Demo

Complete demonstration of all major features including service creation, CRUD operations, validation workflow, reference management, batch operations, error handling, and environment configuration.

**Windows Command Prompt:**

```cmd
cd examples\comprehensive_demo
go run main.go
```

**Windows PowerShell:**

```powershell
Set-Location examples\comprehensive_demo
go run main.go
```

**Linux/macOS:**

```bash
cd examples/comprehensive_demo
go run main.go
```

### Internal Examples

These examples use internal packages directly and are primarily for development and testing purposes:

#### Basic Usage (Internal)

Demonstrates fundamental operations using internal packages directly.

**Windows Command Prompt:**

```cmd
cd internal\examples\basic_usage
go run main.go
```

**Windows PowerShell:**

```powershell
Set-Location internal\examples\basic_usage
go run main.go
```

**Linux/macOS:**

```bash
cd internal/examples/basic_usage
go run main.go
```

#### Advanced Usage (Internal)

Shows complex scenarios including references, custom configuration, and error handling using internal packages.

**Windows Command Prompt:**

```cmd
cd internal\examples\advanced_usage
go run main.go
```

**Windows PowerShell:**

```powershell
Set-Location internal\examples\advanced_usage
go run main.go
```

**Linux/macOS:**

```bash
cd internal/examples/advanced_usage
go run main.go
```

#### Configuration Demo (Internal)

Demonstrates configuration management and environment variable usage.

**Windows Command Prompt:**

```cmd
cd internal\examples\config_demo
go run main.go
```

**Windows PowerShell:**

```powershell
Set-Location internal\examples\config_demo
go run main.go
```

**Linux/macOS:**

```bash
cd internal/examples/config_demo
go run main.go
```

#### Error Handling Demo (Internal)

Shows comprehensive error handling patterns and recovery scenarios.

**Windows Command Prompt:**

```cmd
cd internal\examples\error_handling_demo
go run main.go
```

**Windows PowerShell:**

```powershell
Set-Location internal\examples\error_handling_demo
go run main.go
```

**Linux/macOS:**

```bash
cd internal/examples/error_handling_demo
go run main.go
```

#### Reference Validation Demo (Internal)

Demonstrates reference resolution and validation between state-machine diagrams.

**Windows Command Prompt:**

```cmd
cd internal\examples\reference_validation_demo
go run main.go
```

**Windows PowerShell:**

```powershell
Set-Location internal\examples\reference_validation_demo
go run main.go
```

**Linux/macOS:**

```bash
cd internal/examples/reference_validation_demo
go run main.go
```

## Testing

The module includes comprehensive tests for both internal components and the public API.

### Run All Tests

**Windows:**

```cmd
go test .\internal\... .\diagram\...
```

**Linux/macOS:**

```bash
go test ./internal/... ./diagram/...
```

**Alternative (includes informational messages for example directories):**

**Windows:**

```cmd
go test .\...
```

**Linux/macOS:**

```bash
go test ./...
```

### Run Tests with Verbose Output

**Windows:**

```cmd
go test -v .\internal\... .\diagram\...
```

**Linux/macOS:**

```bash
go test -v ./internal/... ./diagram/...
```

### Run Public API Tests Only

**Windows:**

```cmd
go test -v .\diagram\
```

**Linux/macOS:**

```bash
go test -v ./diagram/
```

### Run Integration Tests

**Windows:**

```cmd
go test .\internal\integration\...
```

**Linux/macOS:**

```bash
go test ./internal/integration/...
```

### Test Coverage

The test suite includes:

- **Public API Tests** (`diagram/diagram_test.go`) - Tests all public functions and integration scenarios
- **Unit Tests** - Individual component tests in each internal package (`internal/*/`)
- **Integration Tests** (`internal/integration/`) - End-to-end workflow tests using internal packages
- **Error Handling Tests** - Comprehensive error scenario coverage

## API Documentation

The module provides the following main interfaces:

### DiagramService Interface

```go
import "github.com/kengibson1111/go-uml-statemachine-models/models"
import "github.com/kengibson1111/go-uml-statemachine-parsers/diagram"

type DiagramService interface {
    // CRUD operations
    Create(diagramType models.DiagramType, name, version string, content string, location Location) (*diagram, error)
    Read(diagramType models.DiagramType, name, version string, location Location) (*diagram, error)
    UpdateInProgressFile(diag *StateMachineDiagram) error
    Delete(diagramType models.DiagramType, name, version string, location Location) error

    // Business operations
    Promote(diagramType models.DiagramType, name, version string) error
    Validate(diagramType models.DiagramType, name, version string, location Location) (*ValidationResult, error)
    ListAllFiles(diagramType models.DiagramType, location Location) ([]diagram, error)

    // Reference operations
    ResolveFileReferences(diagram *StateMachineDiagram) error
}
```

### Core Types

```go
import "github.com/kengibson1111/go-uml-statemachine-models/models"
import "github.com/kengibson1111/go-uml-statemachine-parsers/diagram"

// diagram represents a UML state-machine diagram
type diagram struct {
    Name        string
    Version     string
    Content     string
    References  []Reference
    Location    Location
    DiagramType models.DiagramType
    Metadata    Metadata
}

// Location indicates storage location
type Location int
const (
    LocationFileInProgress Location = iota
    LocationProducts
)

// ValidationResult contains validation outcomes
type ValidationResult struct {
    Errors   []ValidationError
    Warnings []ValidationWarning
    IsValid  bool
}
```

## Best Practices

1. **Version Management**: Use semantic versioning (e.g., "1.0.0", "1.2.3")
2. **Content Validation**: Always validate state-machine diagrams before promotion
3. **Update Workflow**: Remember that updates are only allowed for in-progress diagrams; production diagrams are immutable
4. **Reference Management**: Resolve references after creating state-machine diagrams with dependencies
5. **Error Handling**: Check and handle all error conditions appropriately
6. **Configuration**: Use environment variables for deployment-specific settings
7. **Testing**: Test state-machine diagram content with both strictness levels

## Troubleshooting

### Common Issues

- **Directory Permission Errors**
  - Ensure the application has write permissions to the root directory
  - Check that the directory path is valid and accessible

- **Validation Failures**
  - Verify PlantUML syntax is correct
  - Check that all referenced state-machine diagrams exist
  - Ensure proper @startuml/@enduml tags

- **Reference Resolution Errors**
  - Verify referenced state-machine diagrams exist in the expected locations
  - Check that product references use correct version numbers


- **Update Failures**
  - Ensure the diagram is in the in-progress location (updates are not allowed for production diagrams)
  - Verify the diagram exists before attempting to update
  - Check file system permissions for write operations

- **Promotion Failures**
  - Validate the state-machine diagram passes all validation checks
  - Ensure no conflicting directories exist in products
  - Check file system permissions for move operations

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
