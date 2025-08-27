package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/diagram"
)

func main() {
	fmt.Println("Go UML State-Machine Diagram - Common Use Cases")
	fmt.Println("=======================================")

	// Use Case 1: Simple state-machine diagram creation and management
	fmt.Println("\n=== Use Case 1: Simple State-Machine Diagram Management ===")
	simpleUsageExample()

	// Use Case 2: Environment-based configuration
	fmt.Println("\n=== Use Case 2: Environment Configuration ===")
	environmentConfigExample()

	// Use Case 3: State-machine diagram with validation workflow
	fmt.Println("\n=== Use Case 3: Validation Workflow ===")
	validationWorkflowExample()

	// Use Case 4: Reference management and dependencies
	fmt.Println("\n=== Use Case 4: Reference Management ===")
	referenceManagementExample()

	// Use Case 5: Error handling and recovery
	fmt.Println("\n=== Use Case 5: Error Handling ===")
	errorHandlingExample()

	// Use Case 6: Batch operations
	fmt.Println("\n=== Use Case 6: Batch Operations ===")
	batchOperationsExample()

	fmt.Println("\n✓ All use cases completed successfully!")
}

// Use Case 1: Simple state-machine diagram creation and management
func simpleUsageExample() {
	// Create service with default configuration
	svc, err := diagram.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	// Define a simple order processing state-machine diagram
	orderContent := `@startuml
title Order Processing State-Machine Diagram

[*] --> Pending : create_order()
Pending --> Processing : payment_received()
Pending --> Cancelled : cancel_order()
Processing --> Shipped : ship_order()
Processing --> Cancelled : cancel_order()
Shipped --> Delivered : confirm_delivery()
Delivered --> [*]
Cancelled --> [*]

@enduml`

	// Create the state-machine diagram
	diag, err := svc.CreateFile(models.DiagramTypePUML, "order-processing", "1.0.0", orderContent, diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error creating state-machine diagram: %v", err)
		return
	}
	fmt.Printf("✓ Created state-machine diagram: %s-%s\n", diag.Name, diag.Version)

	// Read it back to verify
	readDiag, err := svc.Read(models.DiagramTypePUML, "order-processing", "1.0.0", diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error reading state-machine diagram: %v", err)
		return
	}
	fmt.Printf("✓ Verified state-machine diagram exists (content length: %d)\n", len(readDiag.Content))

	// Clean up
	err = svc.Delete(models.DiagramTypePUML, "order-processing", "1.0.0", diagram.LocationInProgress)
	if err != nil {
		log.Printf("Warning: Could not clean up: %v", err)
	}
}

// Use Case 2: Environment-based configuration
func environmentConfigExample() {
	// Set some environment variables for demonstration
	os.Setenv("GO_UML_ROOT_DIRECTORY", ".demo-diagrams")
	os.Setenv("GO_UML_DEBUG_LOGGING", "true")
	os.Setenv("GO_UML_MAX_FILE_SIZE", "2097152") // 2MB

	// Create service from environment
	svc, err := diagram.NewServiceFromEnv()
	if err != nil {
		log.Printf("Error creating service from env: %v", err)
		return
	}

	// Show the loaded configuration
	config := diagram.LoadConfigFromEnv()
	fmt.Printf("✓ Loaded configuration from environment:\n")
	fmt.Printf("  - Root Directory: %s\n", config.RootDirectory)
	fmt.Printf("  - Debug Logging: %t\n", config.EnableDebugLogging)
	fmt.Printf("  - Max File Size: %d bytes\n", config.MaxFileSize)

	// Create a simple state-machine diagram to test the configuration
	content := `@startuml
[*] --> Test
Test --> [*]
@enduml`

	_, err = svc.CreateFile(models.DiagramTypePUML, "env-test", "1.0.0", content, diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error creating test state-machine diagram: %v", err)
		return
	}
	fmt.Printf("✓ Created test state-machine diagram in custom directory\n")

	// Clean up
	err = svc.Delete(models.DiagramTypePUML, "env-test", "1.0.0", diagram.LocationInProgress)
	if err != nil {
		log.Printf("Warning: Could not clean up: %v", err)
	}

	// Clean up environment variables
	os.Unsetenv("GO_UML_ROOT_DIRECTORY")
	os.Unsetenv("GO_UML_DEBUG_LOGGING")
	os.Unsetenv("GO_UML_MAX_FILE_SIZE")
}

// Use Case 3: State-machine diagram with validation workflow
func validationWorkflowExample() {
	svc, err := diagram.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	// Create a user authentication state-machine diagram
	authContent := `@startuml
title User Authentication

[*] --> Idle
Idle --> Authenticating : login()
Authenticating --> Authenticated : success
Authenticating --> Failed : failure
Failed --> Idle : retry()
Failed --> Locked : max_attempts_reached()
Authenticated --> Idle : logout()
Locked --> Idle : unlock_timeout()

@enduml`

	// Create the state-machine diagram
	diag, err := svc.CreateFile(models.DiagramTypePUML, "user-auth", "1.0.0", authContent, diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error creating state-machine diagram: %v", err)
		return
	}
	fmt.Printf("✓ Created state-machine diagram: %s-%s\n", diag.Name, diag.Version)

	// Validate the state-machine diagram
	result, err := svc.Validate(models.DiagramTypePUML, "user-auth", "1.0.0", diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error validating state-machine diagram: %v", err)
		return
	}

	fmt.Printf("✓ Validation completed:\n")
	fmt.Printf("  - Valid: %t\n", result.IsValid)
	fmt.Printf("  - Errors: %d\n", len(result.Errors))
	fmt.Printf("  - Warnings: %d\n", len(result.Warnings))

	// Show validation details
	if len(result.Errors) > 0 {
		fmt.Println("  Errors:")
		for _, err := range result.Errors {
			fmt.Printf("    - %s: %s\n", err.Code, err.Message)
		}
	}
	if len(result.Warnings) > 0 {
		fmt.Println("  Warnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("    - %s: %s\n", warning.Code, warning.Message)
		}
	}

	// Attempt promotion if validation passes
	if result.IsValid && !result.HasErrors() {
		err = svc.Promote(models.DiagramTypePUML, "user-auth", "1.0.0")
		if err != nil {
			log.Printf("Error promoting state-machine diagram: %v", err)
		} else {
			fmt.Printf("✓ Successfully promoted to products\n")

			// Clean up from products
			err = svc.Delete(models.DiagramTypePUML, "user-auth", "1.0.0", diagram.LocationProducts)
			if err != nil {
				log.Printf("Warning: Could not clean up from products: %v", err)
			}
		}
	} else {
		fmt.Printf("⚠ Skipping promotion due to validation issues\n")

		// Clean up from in-progress
		err = svc.Delete(models.DiagramTypePUML, "user-auth", "1.0.0", diagram.LocationInProgress)
		if err != nil {
			log.Printf("Warning: Could not clean up from in-progress: %v", err)
		}
	}
}

// Use Case 4: Reference management and dependencies
func referenceManagementExample() {
	svc, err := diagram.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	// First, create a base authentication component
	baseAuthContent := `@startuml
title Base Authentication Component

[*] --> CheckCredentials
CheckCredentials --> Authenticated : valid_credentials
CheckCredentials --> Failed : invalid_credentials
Failed --> [*]
Authenticated --> [*]

@enduml`

	baseDiag, err := svc.CreateFile(models.DiagramTypePUML, "base-auth", "1.0.0", baseAuthContent, diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error creating base auth: %v", err)
		return
	}
	fmt.Printf("✓ Created base component: %s-%s\n", baseDiag.Name, baseDiag.Version)

	// Validate and promote the base component
	result, err := svc.Validate(models.DiagramTypePUML, "base-auth", "1.0.0", diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error validating base auth: %v", err)
		return
	}

	if result.IsValid && !result.HasErrors() {
		err = svc.Promote(models.DiagramTypePUML, "base-auth", "1.0.0")
		if err != nil {
			log.Printf("Error promoting base auth: %v", err)
			return
		}
		fmt.Printf("✓ Base component promoted to products\n")
	}

	// Now create a complex authentication system that references the base
	complexAuthContent := `@startuml
title Complex Authentication System

' Reference to base authentication component
!include products\base-auth-1.0.0\base-auth-1.0.0.puml

[*] --> CheckSession
CheckSession --> SessionValid : has_valid_session
CheckSession --> RequireAuth : no_session
SessionValid --> Authenticated
RequireAuth --> base-auth : delegate_to_base
base-auth --> TwoFactorAuth : base_auth_success
TwoFactorAuth --> Authenticated : 2fa_success
TwoFactorAuth --> Failed : 2fa_failed
Failed --> RequireAuth : retry
Authenticated --> [*]

@enduml`

	complexDiag, err := svc.CreateFile(models.DiagramTypePUML, "complex-auth", "1.0.0", complexAuthContent, diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error creating complex auth: %v", err)
		return
	}
	fmt.Printf("✓ Created complex system: %s-%s\n", complexDiag.Name, complexDiag.Version)

	// Resolve references in the complex system
	err = svc.ResolveReferences(complexDiag)
	if err != nil {
		log.Printf("Error resolving references: %v", err)
	} else {
		fmt.Printf("✓ Resolved %d references:\n", len(complexDiag.References))
		for _, ref := range complexDiag.References {
			fmt.Printf("  - %s (type: %s)\n", ref.Name, ref.Type.String())
		}
	}

	// Clean up
	svc.Delete(models.DiagramTypePUML, "complex-auth", "1.0.0", diagram.LocationInProgress)
	svc.Delete(models.DiagramTypePUML, "base-auth", "1.0.0", diagram.LocationProducts)
}

// Use Case 5: Error handling and recovery
func errorHandlingExample() {
	svc, err := diagram.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	fmt.Printf("✓ Demonstrating various error scenarios:\n")

	// 1. Try to read non-existent state-machine diagram
	_, err = svc.Read(models.DiagramTypePUML, "non-existent", "1.0.0", diagram.LocationInProgress)
	if err != nil {
		fmt.Printf("  ✓ Expected error for non-existent read: File not found\n")
	}

	// 2. Try to create with invalid parameters
	_, err = svc.CreateFile(models.DiagramTypePUML, "", "1.0.0", "content", diagram.LocationInProgress)
	if err != nil {
		fmt.Printf("  ✓ Expected error for empty name: Validation failed\n")
	}

	// 3. Create a state-machine diagram, then try to create duplicate
	content := `@startuml
[*] --> Test
@enduml`

	_, err = svc.CreateFile(models.DiagramTypePUML, "test-duplicate", "1.0.0", content, diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error creating test state-machine diagram: %v", err)
		return
	}

	_, err = svc.CreateFile(models.DiagramTypePUML, "test-duplicate", "1.0.0", content, diagram.LocationInProgress)
	if err != nil {
		fmt.Printf("  ✓ Expected error for duplicate creation: Already exists\n")
	}

	// 4. Try to promote non-existent state-machine diagram
	err = svc.Promote(models.DiagramTypePUML, "non-existent", "1.0.0")
	if err != nil {
		fmt.Printf("  ✓ Expected error for non-existent promotion: File not found\n")
	}

	// 5. Try to delete non-existent state-machine diagram
	err = svc.Delete(models.DiagramTypePUML, "non-existent", "1.0.0", diagram.LocationInProgress)
	if err != nil {
		fmt.Printf("  ✓ Expected error for non-existent deletion: File not found\n")
	}

	// Clean up
	err = svc.Delete(models.DiagramTypePUML, "test-duplicate", "1.0.0", diagram.LocationInProgress)
	if err != nil {
		log.Printf("Warning: Could not clean up: %v", err)
	}
}

// Use Case 6: Batch operations
func batchOperationsExample() {
	svc, err := diagram.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	// Create multiple state-machine diagrams
	diagrams := []struct {
		name    string
		version string
		content string
	}{
		{
			name:    "payment-processing",
			version: "1.0.0",
			content: `@startuml
[*] --> Pending
Pending --> Processing : process()
Processing --> Completed : success
Processing --> Failed : error
@enduml`,
		},
		{
			name:    "inventory-management",
			version: "1.0.0",
			content: `@startuml
[*] --> Available
Available --> Reserved : reserve()
Reserved --> Sold : sell()
Reserved --> Available : release()
@enduml`,
		},
		{
			name:    "shipping-tracking",
			version: "1.0.0",
			content: `@startuml
[*] --> Preparing
Preparing --> Shipped : ship()
Shipped --> InTransit : pickup()
InTransit --> Delivered : deliver()
@enduml`,
		},
	}

	fmt.Printf("✓ Creating %d state-machine diagrams:\n", len(diagrams))

	// Create all state-machine diagrams
	created := []string{}
	for _, diag := range diagrams {
		_, err := svc.CreateFile(models.DiagramTypePUML, diag.name, diag.version, diag.content, diagram.LocationInProgress)
		if err != nil {
			log.Printf("Error creating %s: %v", diag.name, err)
			continue
		}
		created = append(created, diag.name+"-"+diag.version)
		fmt.Printf("  ✓ Created: %s-%s\n", diag.name, diag.version)
	}

	// List all in-progress state-machine diagrams
	allDiagrams, err := svc.ListAll(models.DiagramTypePUML, diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error listing state-machine diagrams: %v", err)
		return
	}

	fmt.Printf("✓ Found %d state-machine diagrams in in-progress:\n", len(allDiagrams))
	for _, diag := range allDiagrams {
		fmt.Printf("  - %s-%s (created: %s)\n",
			diag.Name, diag.Version, diag.Metadata.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Validate all created state-machine diagrams
	fmt.Printf("✓ Validating all state-machine diagrams:\n")
	validCount := 0
	for _, diag := range diagrams {
		result, err := svc.Validate(models.DiagramTypePUML, diag.name, diag.version, diagram.LocationInProgress)
		if err != nil {
			log.Printf("Error validating %s: %v", diag.name, err)
			continue
		}

		if result.IsValid && !result.HasErrors() {
			validCount++
			fmt.Printf("  ✓ %s-%s: Valid\n", diag.name, diag.version)
		} else {
			fmt.Printf("  ⚠ %s-%s: Invalid (%d errors, %d warnings)\n",
				diag.name, diag.version, len(result.Errors), len(result.Warnings))
		}
	}

	fmt.Printf("✓ %d out of %d state-machine diagrams are valid\n", validCount, len(diagrams))

	// Clean up all created state-machine diagrams
	fmt.Printf("✓ Cleaning up created state-machine diagrams:\n")
	for _, diag := range diagrams {
		err := svc.Delete(models.DiagramTypePUML, diag.name, diag.version, diagram.LocationInProgress)
		if err != nil {
			log.Printf("Warning: Could not delete %s: %v", diag.name, err)
		} else {
			fmt.Printf("  ✓ Deleted: %s-%s\n", diag.name, diag.version)
		}
	}
}
