package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kengibson1111/go-uml-statemachine-parsers/statemachine"
)

func main() {
	fmt.Println("Go UML State Machine - Common Use Cases")
	fmt.Println("=======================================")

	// Use Case 1: Simple state machine creation and management
	fmt.Println("\n=== Use Case 1: Simple State Machine Management ===")
	simpleUsageExample()

	// Use Case 2: Environment-based configuration
	fmt.Println("\n=== Use Case 2: Environment Configuration ===")
	environmentConfigExample()

	// Use Case 3: State machine with validation workflow
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

// Use Case 1: Simple state machine creation and management
func simpleUsageExample() {
	// Create service with default configuration
	svc, err := statemachine.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	// Define a simple order processing state machine
	orderContent := `@startuml
title Order Processing State Machine

[*] --> Pending : create_order()
Pending --> Processing : payment_received()
Pending --> Cancelled : cancel_order()
Processing --> Shipped : ship_order()
Processing --> Cancelled : cancel_order()
Shipped --> Delivered : confirm_delivery()
Delivered --> [*]
Cancelled --> [*]

@enduml`

	// Create the state machine
	sm, err := svc.Create("order-processing", "1.0.0", orderContent, statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error creating state machine: %v", err)
		return
	}
	fmt.Printf("✓ Created state machine: %s-%s\n", sm.Name, sm.Version)

	// Read it back to verify
	readSM, err := svc.Read("order-processing", "1.0.0", statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error reading state machine: %v", err)
		return
	}
	fmt.Printf("✓ Verified state machine exists (content length: %d)\n", len(readSM.Content))

	// Clean up
	err = svc.Delete("order-processing", "1.0.0", statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Warning: Could not clean up: %v", err)
	}
}

// Use Case 2: Environment-based configuration
func environmentConfigExample() {
	// Set some environment variables for demonstration
	os.Setenv("GO_UML_ROOT_DIRECTORY", ".demo-state-machines")
	os.Setenv("GO_UML_DEBUG_LOGGING", "true")
	os.Setenv("GO_UML_MAX_FILE_SIZE", "2097152") // 2MB

	// Create service from environment
	svc, err := statemachine.NewServiceFromEnv()
	if err != nil {
		log.Printf("Error creating service from env: %v", err)
		return
	}

	// Show the loaded configuration
	config := statemachine.LoadConfigFromEnv()
	fmt.Printf("✓ Loaded configuration from environment:\n")
	fmt.Printf("  - Root Directory: %s\n", config.RootDirectory)
	fmt.Printf("  - Debug Logging: %t\n", config.EnableDebugLogging)
	fmt.Printf("  - Max File Size: %d bytes\n", config.MaxFileSize)

	// Create a simple state machine to test the configuration
	content := `@startuml
[*] --> Test
Test --> [*]
@enduml`

	_, err = svc.Create("env-test", "1.0.0", content, statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error creating test state machine: %v", err)
		return
	}
	fmt.Printf("✓ Created test state machine in custom directory\n")

	// Clean up
	err = svc.Delete("env-test", "1.0.0", statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Warning: Could not clean up: %v", err)
	}

	// Clean up environment variables
	os.Unsetenv("GO_UML_ROOT_DIRECTORY")
	os.Unsetenv("GO_UML_DEBUG_LOGGING")
	os.Unsetenv("GO_UML_MAX_FILE_SIZE")
}

// Use Case 3: State machine with validation workflow
func validationWorkflowExample() {
	svc, err := statemachine.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	// Create a user authentication state machine
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

	// Create the state machine
	sm, err := svc.Create("user-auth", "1.0.0", authContent, statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error creating state machine: %v", err)
		return
	}
	fmt.Printf("✓ Created state machine: %s-%s\n", sm.Name, sm.Version)

	// Validate the state machine
	result, err := svc.Validate("user-auth", "1.0.0", statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error validating state machine: %v", err)
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
		err = svc.Promote("user-auth", "1.0.0")
		if err != nil {
			log.Printf("Error promoting state machine: %v", err)
		} else {
			fmt.Printf("✓ Successfully promoted to products\n")

			// Clean up from products
			err = svc.Delete("user-auth", "1.0.0", statemachine.LocationProducts)
			if err != nil {
				log.Printf("Warning: Could not clean up from products: %v", err)
			}
		}
	} else {
		fmt.Printf("⚠ Skipping promotion due to validation issues\n")

		// Clean up from in-progress
		err = svc.Delete("user-auth", "1.0.0", statemachine.LocationInProgress)
		if err != nil {
			log.Printf("Warning: Could not clean up from in-progress: %v", err)
		}
	}
}

// Use Case 4: Reference management and dependencies
func referenceManagementExample() {
	svc, err := statemachine.NewService()
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

	baseSM, err := svc.Create("base-auth", "1.0.0", baseAuthContent, statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error creating base auth: %v", err)
		return
	}
	fmt.Printf("✓ Created base component: %s-%s\n", baseSM.Name, baseSM.Version)

	// Validate and promote the base component
	result, err := svc.Validate("base-auth", "1.0.0", statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error validating base auth: %v", err)
		return
	}

	if result.IsValid && !result.HasErrors() {
		err = svc.Promote("base-auth", "1.0.0")
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

	complexSM, err := svc.Create("complex-auth", "1.0.0", complexAuthContent, statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error creating complex auth: %v", err)
		return
	}
	fmt.Printf("✓ Created complex system: %s-%s\n", complexSM.Name, complexSM.Version)

	// Resolve references in the complex system
	err = svc.ResolveReferences(complexSM)
	if err != nil {
		log.Printf("Error resolving references: %v", err)
	} else {
		fmt.Printf("✓ Resolved %d references:\n", len(complexSM.References))
		for _, ref := range complexSM.References {
			fmt.Printf("  - %s (type: %s)\n", ref.Name, ref.Type.String())
		}
	}

	// Clean up
	svc.Delete("complex-auth", "1.0.0", statemachine.LocationInProgress)
	svc.Delete("base-auth", "1.0.0", statemachine.LocationProducts)
}

// Use Case 5: Error handling and recovery
func errorHandlingExample() {
	svc, err := statemachine.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	fmt.Printf("✓ Demonstrating various error scenarios:\n")

	// 1. Try to read non-existent state machine
	_, err = svc.Read("non-existent", "1.0.0", statemachine.LocationInProgress)
	if err != nil {
		fmt.Printf("  ✓ Expected error for non-existent read: File not found\n")
	}

	// 2. Try to create with invalid parameters
	_, err = svc.Create("", "1.0.0", "content", statemachine.LocationInProgress)
	if err != nil {
		fmt.Printf("  ✓ Expected error for empty name: Validation failed\n")
	}

	// 3. Create a state machine, then try to create duplicate
	content := `@startuml
[*] --> Test
@enduml`

	_, err = svc.Create("test-duplicate", "1.0.0", content, statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error creating test state machine: %v", err)
		return
	}

	_, err = svc.Create("test-duplicate", "1.0.0", content, statemachine.LocationInProgress)
	if err != nil {
		fmt.Printf("  ✓ Expected error for duplicate creation: Already exists\n")
	}

	// 4. Try to promote non-existent state machine
	err = svc.Promote("non-existent", "1.0.0")
	if err != nil {
		fmt.Printf("  ✓ Expected error for non-existent promotion: File not found\n")
	}

	// 5. Try to delete non-existent state machine
	err = svc.Delete("non-existent", "1.0.0", statemachine.LocationInProgress)
	if err != nil {
		fmt.Printf("  ✓ Expected error for non-existent deletion: File not found\n")
	}

	// Clean up
	err = svc.Delete("test-duplicate", "1.0.0", statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Warning: Could not clean up: %v", err)
	}
}

// Use Case 6: Batch operations
func batchOperationsExample() {
	svc, err := statemachine.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	// Create multiple state machines
	stateMachines := []struct {
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

	fmt.Printf("✓ Creating %d state machines:\n", len(stateMachines))

	// Create all state machines
	created := []string{}
	for _, sm := range stateMachines {
		_, err := svc.Create(sm.name, sm.version, sm.content, statemachine.LocationInProgress)
		if err != nil {
			log.Printf("Error creating %s: %v", sm.name, err)
			continue
		}
		created = append(created, sm.name+"-"+sm.version)
		fmt.Printf("  ✓ Created: %s-%s\n", sm.name, sm.version)
	}

	// List all in-progress state machines
	allSMs, err := svc.ListAll(statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error listing state machines: %v", err)
		return
	}

	fmt.Printf("✓ Found %d state machines in in-progress:\n", len(allSMs))
	for _, sm := range allSMs {
		fmt.Printf("  - %s-%s (created: %s)\n",
			sm.Name, sm.Version, sm.Metadata.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Validate all created state machines
	fmt.Printf("✓ Validating all state machines:\n")
	validCount := 0
	for _, sm := range stateMachines {
		result, err := svc.Validate(sm.name, sm.version, statemachine.LocationInProgress)
		if err != nil {
			log.Printf("Error validating %s: %v", sm.name, err)
			continue
		}

		if result.IsValid && !result.HasErrors() {
			validCount++
			fmt.Printf("  ✓ %s-%s: Valid\n", sm.name, sm.version)
		} else {
			fmt.Printf("  ⚠ %s-%s: Invalid (%d errors, %d warnings)\n",
				sm.name, sm.version, len(result.Errors), len(result.Warnings))
		}
	}

	fmt.Printf("✓ %d out of %d state machines are valid\n", validCount, len(stateMachines))

	// Clean up all created state machines
	fmt.Printf("✓ Cleaning up created state machines:\n")
	for _, sm := range stateMachines {
		err := svc.Delete(sm.name, sm.version, statemachine.LocationInProgress)
		if err != nil {
			log.Printf("Warning: Could not delete %s: %v", sm.name, err)
		} else {
			fmt.Printf("  ✓ Deleted: %s-%s\n", sm.name, sm.version)
		}
	}
}
