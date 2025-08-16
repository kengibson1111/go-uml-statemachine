# Go UML State Machine

A Go module for managing UML state machine diagrams in PlantUML format. This library provides functionality to read, write, validate, and organize PlantUML files within a structured directory hierarchy with support for versioning, validation levels, and cross-references between state machines.

## Features

- **Structured File Organization**: Automatic directory management with separate locations for in-progress and production-ready state machines
- **Version Management**: Semantic versioning support with automatic version parsing and comparison
- **PlantUML Validation**: Configurable validation with different strictness levels based on deployment status
- **Reference Resolution**: Support for cross-references between state machines with automatic dependency resolution
- **Promotion Workflow**: Safe promotion of state machines from in-progress to production with validation checks
- **Thread-Safe Operations**: Concurrent access protection with mutex locks
- **Comprehensive Error Handling**: Detailed error messages with context information
- **Configurable Logging**: Debug and info level logging with component-specific prefixes

## Installation

```cmd
go get github.com/kengibson1111/go-uml-statemachine
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/kengibson1111/go-uml-statemachine/internal/models"
    "github.com/kengibson1111/go-uml-statemachine/internal/repository"
    "github.com/kengibson1111/go-uml-statemachine/internal/service"
    "github.com/kengibson1111/go-uml-statemachine/internal/validation"
)

func main() {
    // Create configuration
    config := models.DefaultConfig()
    
    // Create dependencies
    repo := repository.NewFileSystemRepository(config)
    validator := validation.NewPlantUMLValidatorWithRepository(repo)
    svc := service.NewService(repo, validator, config)
    
    // Create a new state machine
    content := `@startuml
    [*] --> Idle
    Idle --> Active : start()
    Active --> Idle : stop()
    @enduml`
    
    sm, err := svc.Create("my-machine", "1.0.0", content, models.LocationInProgress)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Created state machine: %s-%s\n", sm.Name, sm.Version)
}
```

## Directory Structure

The module organizes state machines in a standardized directory structure:

**Windows:**
```
.go-uml-statemachine\
├── in-progress\
│   └── {name}-{version}\
│       ├── {name}-{version}.puml
│       └── nested\
│           └── {nested-name}\
│               └── {nested-name}.puml
└── products\
    └── {name}-{version}\
        ├── {name}-{version}.puml
        └── nested\
            └── {nested-name}\
                └── {nested-name}.puml
```

**Linux/macOS:**
```
.go-uml-statemachine/
├── in-progress/
│   └── {name}-{version}/
│       ├── {name}-{version}.puml
│       └── nested/
│           └── {nested-name}/
│               └── {nested-name}.puml
└── products/
    └── {name}-{version}/
        ├── {name}-{version}.puml
        └── nested/
            └── {nested-name}/
                └── {nested-name}.puml
```

## Configuration

### Default Configuration

```go
config := models.DefaultConfig()
// Uses:
// - RootDirectory: ".go-uml-statemachine"
// - ValidationLevel: StrictnessInProgress
// - BackupEnabled: false
// - MaxFileSize: 1MB
// - EnableDebugLogging: false
```

### Environment Variables

The module supports configuration through environment variables:

- `GO_UML_ROOT_DIRECTORY`: Root directory for state machines
- `GO_UML_VALIDATION_LEVEL`: Validation level (`in-progress` or `products`)
- `GO_UML_BACKUP_ENABLED`: Enable backups (`true` or `false`)
- `GO_UML_MAX_FILE_SIZE`: Maximum file size in bytes
- `GO_UML_DEBUG_LOGGING`: Enable debug logging (`true` or `false`)

```go
// Load configuration from environment
config := models.LoadConfigFromEnv()

// Or merge with existing config
config := models.DefaultConfig()
config.MergeWithEnv()
```

## Core Operations

### Creating State Machines

```go
// Create in in-progress location
sm, err := svc.Create("user-auth", "1.0.0", plantUMLContent, models.LocationInProgress)

// Create in products location (for direct production deployment)
sm, err := svc.Create("user-auth", "1.0.0", plantUMLContent, models.LocationProducts)
```

### Reading State Machines

```go
sm, err := svc.Read("user-auth", "1.0.0", models.LocationInProgress)
if err != nil {
    log.Printf("Error reading state machine: %v", err)
}
```

### Updating State Machines

```go
sm.Content = updatedPlantUMLContent
err := svc.Update(sm)
```

### Deleting State Machines

```go
err := svc.Delete("user-auth", "1.0.0", models.LocationInProgress)
```

### Listing State Machines

```go
// List all in-progress state machines
inProgressSMs, err := svc.ListAll(models.LocationInProgress)

// List all production state machines
productSMs, err := svc.ListAll(models.LocationProducts)
```

## Validation

The module supports two validation strictness levels:

### In-Progress Validation
- Returns both errors and warnings
- Prevents promotion if errors exist
- Used for development and testing

```go
result, err := svc.Validate("user-auth", "1.0.0", models.LocationInProgress)
if result.HasErrors() {
    fmt.Println("Validation failed with errors:")
    for _, err := range result.Errors {
        fmt.Printf("  - %s: %s\n", err.Code, err.Message)
    }
}
```

### Products Validation
- Returns warnings only (no blocking errors)
- Used for production state machines
- More lenient to allow operational flexibility

## Promotion Workflow

Move state machines from in-progress to products with validation:

```go
// Promotes only if validation passes
err := svc.Promote("user-auth", "1.0.0")
if err != nil {
    log.Printf("Promotion failed: %v", err)
}
```

The promotion process:
1. Validates the state machine exists in in-progress
2. Checks for conflicts in products directory
3. Validates the state machine content
4. Performs atomic move operation
5. Includes rollback capability on failure

## References and Dependencies

State machines can reference other state machines:

### Product References
Reference state machines in the products directory:

**Windows:**
```plantuml
@startuml
!include products\base-auth-1.0.0\base-auth-1.0.0.puml

[*] --> base-auth
base-auth --> TwoFactor : success
@enduml
```

**Linux/macOS:**
```plantuml
@startuml
!include products/base-auth-1.0.0/base-auth-1.0.0.puml

[*] --> base-auth
base-auth --> TwoFactor : success
@enduml
```

### Nested References
Reference state machines within the same parent directory:

**Windows:**
```plantuml
@startuml
!include nested\sub-process\sub-process.puml

[*] --> sub-process
@enduml
```

**Linux/macOS:**
```plantuml
@startuml
!include nested/sub-process/sub-process.puml

[*] --> sub-process
@enduml
```

### Reference Resolution

```go
err := svc.ResolveReferences(sm)
if err != nil {
    log.Printf("Reference resolution failed: %v", err)
}

// Check resolved references
for _, ref := range sm.References {
    fmt.Printf("Reference: %s (type: %s, path: %s)\n", 
        ref.Name, ref.Type.String(), ref.Path)
}
```

## Error Handling

The module provides comprehensive error handling with context:

```go
sm, err := svc.Read("non-existent", "1.0.0", models.LocationInProgress)
if err != nil {
    // Error includes context about the operation, component, and parameters
    fmt.Printf("Error: %v\n", err)
    
    // Check error type
    if smErr, ok := err.(*models.StateMachineError); ok {
        switch smErr.Type {
        case models.ErrorTypeFileNotFound:
            fmt.Println("File not found")
        case models.ErrorTypeValidation:
            fmt.Println("Validation error")
        case models.ErrorTypeDirectoryConflict:
            fmt.Println("Directory conflict")
        }
    }
}
```

## Examples

### Basic Usage
See `examples/basic_usage/main.go` (Linux/macOS) or `examples\basic_usage\main.go` (Windows) for a complete basic example.

### Advanced Usage
See `examples/advanced_usage/main.go` (Linux/macOS) or `examples\advanced_usage\main.go` (Windows) for advanced features including:
- Reference resolution
- Custom configuration
- Error handling scenarios
- Environment configuration
- Cleanup operations

## Running Examples

The module includes several examples demonstrating different use cases:

### Basic Usage Example
Demonstrates fundamental operations like creating, reading, validating, and promoting state machines.

**Windows Command Prompt:**
```cmd
cd examples\basic_usage
go run main.go
```

**Windows PowerShell:**
```powershell
Set-Location examples\basic_usage
go run main.go
```

**Linux/macOS:**
```bash
cd examples/basic_usage
go run main.go
```

### Advanced Usage Example
Shows complex scenarios including references, custom configuration, and error handling.

**Windows Command Prompt:**
```cmd
cd examples\advanced_usage
go run main.go
```

**Windows PowerShell:**
```powershell
Set-Location examples\advanced_usage
go run main.go
```

**Linux/macOS:**
```bash
cd examples/advanced_usage
go run main.go
```

### Common Use Cases Example
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

### API Test Example
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

### Comprehensive Demo
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

## Testing

Run the test suite:

**Windows:**
```cmd
go test .\...
```

**Linux/macOS:**
```bash
go test ./...
```

Run tests with verbose output:

**Windows:**
```cmd
go test -v .\...
```

**Linux/macOS:**
```bash
go test -v ./...
```

Run integration tests:

**Windows:**
```cmd
go test .\integration\...
```

**Linux/macOS:**
```bash
go test ./integration/...
```

## API Documentation

The module provides the following main interfaces:

### StateMachineService Interface

```go
type StateMachineService interface {
    // CRUD operations
    Create(name, version string, content string, location Location) (*StateMachine, error)
    Read(name, version string, location Location) (*StateMachine, error)
    Update(sm *StateMachine) error
    Delete(name, version string, location Location) error

    // Business operations
    Promote(name, version string) error
    Validate(name, version string, location Location) (*ValidationResult, error)
    ListAll(location Location) ([]StateMachine, error)

    // Reference operations
    ResolveReferences(sm *StateMachine) error
}
```

### Core Types

```go
// StateMachine represents a UML state machine
type StateMachine struct {
    Name       string
    Version    string
    Content    string
    References []Reference
    Location   Location
    Metadata   Metadata
}

// Location indicates storage location
type Location int
const (
    LocationInProgress Location = iota
    LocationProducts
    LocationNested
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
2. **Content Validation**: Always validate state machines before promotion
3. **Reference Management**: Resolve references after creating state machines with dependencies
4. **Error Handling**: Check and handle all error conditions appropriately
5. **Configuration**: Use environment variables for deployment-specific settings
6. **Testing**: Test state machine content with both strictness levels

## Troubleshooting

### Common Issues

**Directory Permission Errors**
- Ensure the application has write permissions to the root directory
- Check that the directory path is valid and accessible

**Validation Failures**
- Verify PlantUML syntax is correct
- Check that all referenced state machines exist
- Ensure proper @startuml/@enduml tags

**Reference Resolution Errors**
- Verify referenced state machines exist in the expected locations
- Check that product references use correct version numbers
- Ensure nested references are in the correct directory structure

**Promotion Failures**
- Validate the state machine passes all validation checks
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