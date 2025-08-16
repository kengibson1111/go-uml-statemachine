package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kengibson1111/go-uml-statemachine/statemachine"
)

func main() {
	fmt.Println("Go UML State Machine - Comprehensive Demo")
	fmt.Println("=========================================")
	fmt.Println("This demo showcases all major features of the Go UML State Machine module.")
	fmt.Println()

	// Demo 1: Service Creation and Configuration
	fmt.Println("=== Demo 1: Service Creation and Configuration ===")
	demonstrateServiceCreation()

	// Demo 2: Basic CRUD Operations
	fmt.Println("\n=== Demo 2: Basic CRUD Operations ===")
	demonstrateCRUDOperations()

	// Demo 3: Validation and Promotion Workflow
	fmt.Println("\n=== Demo 3: Validation and Promotion Workflow ===")
	demonstrateValidationWorkflow()

	// Demo 4: Reference Management
	fmt.Println("\n=== Demo 4: Reference Management ===")
	demonstrateReferenceManagement()

	// Demo 5: Batch Operations and Listing
	fmt.Println("\n=== Demo 5: Batch Operations and Listing ===")
	demonstrateBatchOperations()

	// Demo 6: Error Handling and Edge Cases
	fmt.Println("\n=== Demo 6: Error Handling and Edge Cases ===")
	demonstrateErrorHandling()

	// Demo 7: Environment Configuration
	fmt.Println("\n=== Demo 7: Environment Configuration ===")
	demonstrateEnvironmentConfiguration()

	fmt.Println("\n✓ Comprehensive demo completed successfully!")
	fmt.Println("\nFor more detailed examples, check out:")
	fmt.Println("  - examples\\basic_usage\\main.go")
	fmt.Println("  - examples\\advanced_usage\\main.go")
	fmt.Println("  - examples\\common_use_cases\\main.go")
	fmt.Println("\nFor API documentation, see API.md")
}

func demonstrateServiceCreation() {
	fmt.Println("Creating services with different configurations...")

	// Default service
	svc1, err := statemachine.NewService()
	if err != nil {
		log.Printf("Error creating default service: %v", err)
		return
	}
	fmt.Println("✓ Created service with default configuration")

	// Custom configuration
	config := &statemachine.Config{
		RootDirectory:      ".demo-statemachines",
		EnableDebugLogging: false,           // Keep logs clean for demo
		MaxFileSize:        2 * 1024 * 1024, // 2MB
		BackupEnabled:      true,
	}
	svc2, err := statemachine.NewServiceWithConfig(config)
	if err != nil {
		log.Printf("Error creating custom service: %v", err)
		return
	}
	fmt.Println("✓ Created service with custom configuration")

	// Environment-based service
	svc3, err := statemachine.NewServiceFromEnv()
	if err != nil {
		log.Printf("Error creating env service: %v", err)
		return
	}
	fmt.Println("✓ Created service from environment configuration")

	// Suppress unused variable warnings
	_ = svc1
	_ = svc2
	_ = svc3
}

func demonstrateCRUDOperations() {
	svc, err := statemachine.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	// Create
	fmt.Println("Creating a new state machine...")
	content := `@startuml
title E-commerce Order Processing

[*] --> OrderReceived : place_order()
OrderReceived --> PaymentPending : validate_order()
PaymentPending --> PaymentProcessing : process_payment()
PaymentProcessing --> OrderConfirmed : payment_success()
PaymentProcessing --> PaymentFailed : payment_failure()
PaymentFailed --> OrderCancelled : cancel_order()
OrderConfirmed --> Fulfillment : prepare_shipment()
Fulfillment --> Shipped : ship_order()
Shipped --> Delivered : confirm_delivery()
Delivered --> [*]
OrderCancelled --> [*]

@enduml`

	sm, err := svc.Create("ecommerce-order", "1.0.0", content, statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error creating state machine: %v", err)
		return
	}
	fmt.Printf("✓ Created: %s-%s\n", sm.Name, sm.Version)

	// Read
	fmt.Println("Reading the state machine...")
	readSM, err := svc.Read("ecommerce-order", "1.0.0", statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error reading state machine: %v", err)
		return
	}
	fmt.Printf("✓ Read: %s-%s (content length: %d)\n", readSM.Name, readSM.Version, len(readSM.Content))

	// Update
	fmt.Println("Updating the state machine...")
	updatedContent := content + `
' Added refund process
PaymentFailed --> RefundProcessing : request_refund()
RefundProcessing --> RefundCompleted : refund_success()
RefundCompleted --> [*]`

	readSM.Content = updatedContent
	err = svc.Update(readSM)
	if err != nil {
		log.Printf("Error updating state machine: %v", err)
		return
	}
	fmt.Println("✓ Updated state machine with refund process")

	// Clean up for next demo
	defer func() {
		svc.Delete("ecommerce-order", "1.0.0", statemachine.LocationInProgress)
	}()
}

func demonstrateValidationWorkflow() {
	svc, err := statemachine.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	// Create a state machine for validation demo
	content := `@startuml
title User Session Management

[*] --> Anonymous
Anonymous --> Authenticating : login()
Authenticating --> Authenticated : success
Authenticating --> LoginFailed : failure
LoginFailed --> Anonymous : retry()
LoginFailed --> AccountLocked : max_attempts()
Authenticated --> SessionActive : establish_session()
SessionActive --> SessionExpired : timeout()
SessionActive --> Anonymous : logout()
SessionExpired --> Anonymous : cleanup()
AccountLocked --> Anonymous : unlock_timeout()

@enduml`

	sm, err := svc.Create("user-session", "1.0.0", content, statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error creating state machine: %v", err)
		return
	}
	fmt.Printf("✓ Created state machine: %s-%s\n", sm.Name, sm.Version)

	// Validate
	fmt.Println("Validating state machine...")
	result, err := svc.Validate("user-session", "1.0.0", statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error validating state machine: %v", err)
		return
	}

	fmt.Printf("✓ Validation completed:\n")
	fmt.Printf("  - Valid: %t\n", result.IsValid)
	fmt.Printf("  - Errors: %d\n", len(result.Errors))
	fmt.Printf("  - Warnings: %d\n", len(result.Warnings))

	// Show validation details if any
	if len(result.Errors) > 0 {
		fmt.Println("  Validation Errors:")
		for _, err := range result.Errors {
			fmt.Printf("    - %s: %s\n", err.Code, err.Message)
		}
	}
	if len(result.Warnings) > 0 {
		fmt.Println("  Validation Warnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("    - %s: %s\n", warning.Code, warning.Message)
		}
	}

	// Promote if validation passes
	if result.IsValid && !result.HasErrors() {
		fmt.Println("Promoting to products...")
		err = svc.Promote("user-session", "1.0.0")
		if err != nil {
			log.Printf("Error promoting state machine: %v", err)
		} else {
			fmt.Println("✓ Successfully promoted to products")
		}
	} else {
		fmt.Println("⚠ Skipping promotion due to validation issues")
	}

	// Clean up
	defer func() {
		svc.Delete("user-session", "1.0.0", statemachine.LocationProducts)
		svc.Delete("user-session", "1.0.0", statemachine.LocationInProgress)
	}()
}

func demonstrateReferenceManagement() {
	svc, err := statemachine.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	// Create a base component
	baseContent := `@startuml
title Authentication Component

[*] --> ValidateCredentials
ValidateCredentials --> Authenticated : valid
ValidateCredentials --> AuthenticationFailed : invalid
AuthenticationFailed --> [*]
Authenticated --> [*]

@enduml`

	baseSM, err := svc.Create("auth-component", "1.0.0", baseContent, statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error creating base component: %v", err)
		return
	}
	fmt.Printf("✓ Created base component: %s-%s\n", baseSM.Name, baseSM.Version)

	// Promote base component to products
	result, err := svc.Validate("auth-component", "1.0.0", statemachine.LocationInProgress)
	if err == nil && result.IsValid && !result.HasErrors() {
		err = svc.Promote("auth-component", "1.0.0")
		if err != nil {
			log.Printf("Error promoting base component: %v", err)
			return
		}
		fmt.Println("✓ Base component promoted to products")
	}

	// Create a complex system that references the base
	complexContent := `@startuml
title Multi-Factor Authentication System

' Reference to authentication component
!include products\auth-component-1.0.0\auth-component-1.0.0.puml

[*] --> CheckExistingSession
CheckExistingSession --> SessionValid : valid_session
CheckExistingSession --> RequireAuthentication : no_session
SessionValid --> Authenticated
RequireAuthentication --> auth-component : delegate_auth
auth-component --> TwoFactorRequired : primary_auth_success
TwoFactorRequired --> TwoFactorValidation : request_2fa()
TwoFactorValidation --> Authenticated : 2fa_success
TwoFactorValidation --> AuthenticationFailed : 2fa_failed
AuthenticationFailed --> RequireAuthentication : retry
Authenticated --> [*]

@enduml`

	complexSM, err := svc.Create("mfa-system", "1.0.0", complexContent, statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error creating complex system: %v", err)
		return
	}
	fmt.Printf("✓ Created complex system: %s-%s\n", complexSM.Name, complexSM.Version)

	// Resolve references
	fmt.Println("Resolving references...")
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
	defer func() {
		svc.Delete("mfa-system", "1.0.0", statemachine.LocationInProgress)
		svc.Delete("auth-component", "1.0.0", statemachine.LocationProducts)
	}()
}

func demonstrateBatchOperations() {
	svc, err := statemachine.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	// Define multiple state machines for different business processes
	processes := []struct {
		name    string
		version string
		content string
	}{
		{
			name:    "inventory-tracking",
			version: "1.0.0",
			content: `@startuml
[*] --> Available
Available --> Reserved : reserve()
Reserved --> Sold : purchase()
Reserved --> Available : release()
Sold --> [*]
@enduml`,
		},
		{
			name:    "customer-support",
			version: "1.0.0",
			content: `@startuml
[*] --> TicketCreated
TicketCreated --> InProgress : assign()
InProgress --> Resolved : resolve()
InProgress --> Escalated : escalate()
Escalated --> Resolved : resolve()
Resolved --> Closed : close()
Closed --> [*]
@enduml`,
		},
		{
			name:    "content-moderation",
			version: "1.0.0",
			content: `@startuml
[*] --> Submitted
Submitted --> UnderReview : review()
UnderReview --> Approved : approve()
UnderReview --> Rejected : reject()
Approved --> Published : publish()
Published --> [*]
Rejected --> [*]
@enduml`,
		},
	}

	// Create all state machines
	fmt.Printf("Creating %d business process state machines...\n", len(processes))
	for _, process := range processes {
		_, err := svc.Create(process.name, process.version, process.content, statemachine.LocationInProgress)
		if err != nil {
			log.Printf("Error creating %s: %v", process.name, err)
			continue
		}
		fmt.Printf("  ✓ Created: %s-%s\n", process.name, process.version)
	}

	// List all in-progress state machines
	fmt.Println("Listing all in-progress state machines...")
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

	// Validate all and count valid ones
	fmt.Println("Batch validation of all state machines...")
	validCount := 0
	for _, process := range processes {
		result, err := svc.Validate(process.name, process.version, statemachine.LocationInProgress)
		if err != nil {
			log.Printf("Error validating %s: %v", process.name, err)
			continue
		}

		if result.IsValid && !result.HasErrors() {
			validCount++
			fmt.Printf("  ✓ %s-%s: Valid\n", process.name, process.version)
		} else {
			fmt.Printf("  ⚠ %s-%s: Invalid (%d errors, %d warnings)\n",
				process.name, process.version, len(result.Errors), len(result.Warnings))
		}
	}

	fmt.Printf("✓ Batch validation completed: %d/%d state machines are valid\n", validCount, len(processes))

	// Clean up all created state machines
	fmt.Println("Cleaning up batch-created state machines...")
	for _, process := range processes {
		err := svc.Delete(process.name, process.version, statemachine.LocationInProgress)
		if err != nil {
			log.Printf("Warning: Could not delete %s: %v", process.name, err)
		} else {
			fmt.Printf("  ✓ Deleted: %s-%s\n", process.name, process.version)
		}
	}
}

func demonstrateErrorHandling() {
	svc, err := statemachine.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	fmt.Println("Demonstrating various error scenarios and recovery...")

	// 1. Invalid parameters
	fmt.Println("1. Testing invalid parameters...")
	_, err = svc.Create("", "1.0.0", "content", statemachine.LocationInProgress)
	if err != nil {
		fmt.Println("  ✓ Correctly handled empty name error")
	}

	_, err = svc.Create("test", "", "content", statemachine.LocationInProgress)
	if err != nil {
		fmt.Println("  ✓ Correctly handled empty version error")
	}

	_, err = svc.Create("test", "1.0.0", "", statemachine.LocationInProgress)
	if err != nil {
		fmt.Println("  ✓ Correctly handled empty content error")
	}

	// 2. Non-existent operations
	fmt.Println("2. Testing non-existent resource operations...")
	_, err = svc.Read("non-existent", "1.0.0", statemachine.LocationInProgress)
	if err != nil {
		fmt.Println("  ✓ Correctly handled non-existent read error")
	}

	err = svc.Delete("non-existent", "1.0.0", statemachine.LocationInProgress)
	if err != nil {
		fmt.Println("  ✓ Correctly handled non-existent delete error")
	}

	err = svc.Promote("non-existent", "1.0.0")
	if err != nil {
		fmt.Println("  ✓ Correctly handled non-existent promote error")
	}

	// 3. Duplicate creation
	fmt.Println("3. Testing duplicate creation...")
	content := `@startuml
[*] --> Test
Test --> [*]
@enduml`

	_, err = svc.Create("duplicate-test", "1.0.0", content, statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error creating first instance: %v", err)
		return
	}

	_, err = svc.Create("duplicate-test", "1.0.0", content, statemachine.LocationInProgress)
	if err != nil {
		fmt.Println("  ✓ Correctly handled duplicate creation error")
	}

	// 4. Update non-existent
	fmt.Println("4. Testing update of non-existent state machine...")
	nonExistentSM := &statemachine.StateMachine{
		Name:     "non-existent",
		Version:  "1.0.0",
		Content:  content,
		Location: statemachine.LocationInProgress,
	}
	err = svc.Update(nonExistentSM)
	if err != nil {
		fmt.Println("  ✓ Correctly handled update of non-existent state machine")
	}

	// Clean up
	svc.Delete("duplicate-test", "1.0.0", statemachine.LocationInProgress)

	fmt.Println("✓ Error handling demonstration completed")
}

func demonstrateEnvironmentConfiguration() {
	fmt.Println("Demonstrating environment-based configuration...")

	// Set some environment variables
	originalRootDir := os.Getenv("GO_UML_ROOT_DIRECTORY")
	originalDebugLogging := os.Getenv("GO_UML_DEBUG_LOGGING")

	os.Setenv("GO_UML_ROOT_DIRECTORY", ".env-demo-statemachines")
	os.Setenv("GO_UML_DEBUG_LOGGING", "false")   // Keep demo clean
	os.Setenv("GO_UML_MAX_FILE_SIZE", "5242880") // 5MB

	// Load configuration from environment
	config := statemachine.LoadConfigFromEnv()
	fmt.Printf("✓ Environment configuration loaded:\n")
	fmt.Printf("  - Root Directory: %s\n", config.RootDirectory)
	fmt.Printf("  - Debug Logging: %t\n", config.EnableDebugLogging)
	fmt.Printf("  - Max File Size: %d bytes (%.1f MB)\n", config.MaxFileSize, float64(config.MaxFileSize)/(1024*1024))

	// Create service from environment
	svc, err := statemachine.NewServiceFromEnv()
	if err != nil {
		log.Printf("Error creating service from environment: %v", err)
		return
	}

	// Test the service with environment configuration
	content := `@startuml
title Environment Test
[*] --> EnvConfigured
EnvConfigured --> [*]
@enduml`

	sm, err := svc.Create("env-config-test", "1.0.0", content, statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Error creating test state machine: %v", err)
	} else {
		fmt.Printf("✓ Created state machine using environment configuration: %s-%s\n", sm.Name, sm.Version)

		// Clean up
		svc.Delete("env-config-test", "1.0.0", statemachine.LocationInProgress)
	}

	// Restore original environment variables
	if originalRootDir != "" {
		os.Setenv("GO_UML_ROOT_DIRECTORY", originalRootDir)
	} else {
		os.Unsetenv("GO_UML_ROOT_DIRECTORY")
	}
	if originalDebugLogging != "" {
		os.Setenv("GO_UML_DEBUG_LOGGING", originalDebugLogging)
	} else {
		os.Unsetenv("GO_UML_DEBUG_LOGGING")
	}
	os.Unsetenv("GO_UML_MAX_FILE_SIZE")

	fmt.Println("✓ Environment configuration demonstration completed")
}
