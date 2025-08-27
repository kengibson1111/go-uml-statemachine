package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/diagram"
)

func main() {
	fmt.Println("Go UML State-Machine Diagram - Comprehensive Demo")
	fmt.Println("=========================================")
	fmt.Println("This demo showcases all major features of the Go UML State-Machine Diagram module.")
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
	svc1, err := diagram.NewService()
	if err != nil {
		log.Printf("Error creating default service: %v", err)
		return
	}
	fmt.Println("✓ Created service with default configuration")

	// Custom configuration
	config := &diagram.Config{
		RootDirectory:      ".demo-diagrams",
		EnableDebugLogging: false,           // Keep logs clean for demo
		MaxFileSize:        2 * 1024 * 1024, // 2MB
		BackupEnabled:      true,
	}
	svc2, err := diagram.NewServiceWithConfig(config)
	if err != nil {
		log.Printf("Error creating custom service: %v", err)
		return
	}
	fmt.Println("✓ Created service with custom configuration")

	// Environment-based service
	svc3, err := diagram.NewServiceFromEnv()
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
	svc, err := diagram.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	// Create
	fmt.Println("Creating a new state-machine diagram...")
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

	diag, err := svc.CreateFile(models.DiagramTypePUML, "ecommerce-order", "1.0.0", content, diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error creating state-machine diagram: %v", err)
		return
	}
	fmt.Printf("✓ Created: %s-%s\n", diag.Name, diag.Version)

	// Read
	fmt.Println("Reading the state-machine diagram...")
	readDiag, err := svc.ReadFile(models.DiagramTypePUML, "ecommerce-order", "1.0.0", diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error reading state-machine diagram: %v", err)
		return
	}
	fmt.Printf("✓ Read: %s-%s (content length: %d)\n", readDiag.Name, readDiag.Version, len(readDiag.Content))

	// Update
	fmt.Println("Updating the state-machine diagram...")
	updatedContent := content + `
' Added refund process
PaymentFailed --> RefundProcessing : request_refund()
RefundProcessing --> RefundCompleted : refund_success()
RefundCompleted --> [*]`

	readDiag.Content = updatedContent
	err = svc.UpdateInProgressFile(readDiag)
	if err != nil {
		log.Printf("Error updating state-machine diagram: %v", err)
		return
	}
	fmt.Println("✓ Updated state-machine diagram with refund process")

	// Clean up for next demo
	defer func() {
		svc.DeleteFile(models.DiagramTypePUML, "ecommerce-order", "1.0.0", diagram.LocationInProgress)
	}()
}

func demonstrateValidationWorkflow() {
	svc, err := diagram.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	// Create a state-machine diagram for validation demo
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

	diag, err := svc.CreateFile(models.DiagramTypePUML, "user-session", "1.0.0", content, diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error creating state-machine diagram: %v", err)
		return
	}
	fmt.Printf("✓ Created state-machine diagram: %s-%s\n", diag.Name, diag.Version)

	// Validate
	fmt.Println("Validating state-machine diagram...")
	result, err := svc.Validate(models.DiagramTypePUML, "user-session", "1.0.0", diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error validating state-machine diagram: %v", err)
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
		err = svc.Promote(models.DiagramTypePUML, "user-session", "1.0.0")
		if err != nil {
			log.Printf("Error promoting state-machine diagram: %v", err)
		} else {
			fmt.Println("✓ Successfully promoted to products")
		}
	} else {
		fmt.Println("⚠ Skipping promotion due to validation issues")
	}

	// Clean up
	defer func() {
		svc.DeleteFile(models.DiagramTypePUML, "user-session", "1.0.0", diagram.LocationProducts)
		svc.DeleteFile(models.DiagramTypePUML, "user-session", "1.0.0", diagram.LocationInProgress)
	}()
}

func demonstrateReferenceManagement() {
	svc, err := diagram.NewService()
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

	baseDiag, err := svc.CreateFile(models.DiagramTypePUML, "auth-component", "1.0.0", baseContent, diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error creating base component: %v", err)
		return
	}
	fmt.Printf("✓ Created base component: %s-%s\n", baseDiag.Name, baseDiag.Version)

	// Promote base component to products
	result, err := svc.Validate(models.DiagramTypePUML, "auth-component", "1.0.0", diagram.LocationInProgress)
	if err == nil && result.IsValid && !result.HasErrors() {
		err = svc.Promote(models.DiagramTypePUML, "auth-component", "1.0.0")
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

	complexDiag, err := svc.CreateFile(models.DiagramTypePUML, "mfa-system", "1.0.0", complexContent, diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error creating complex system: %v", err)
		return
	}
	fmt.Printf("✓ Created complex system: %s-%s\n", complexDiag.Name, complexDiag.Version)

	// Resolve references
	fmt.Println("Resolving references...")
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
	defer func() {
		svc.DeleteFile(models.DiagramTypePUML, "mfa-system", "1.0.0", diagram.LocationInProgress)
		svc.DeleteFile(models.DiagramTypePUML, "auth-component", "1.0.0", diagram.LocationProducts)
	}()
}

func demonstrateBatchOperations() {
	svc, err := diagram.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	// Define multiple state-machine diagrams for different business processes
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

	// Create all state-machine diagrams
	fmt.Printf("Creating %d business process state-machine diagrams...\n", len(processes))
	for _, process := range processes {
		_, err := svc.CreateFile(models.DiagramTypePUML, process.name, process.version, process.content, diagram.LocationInProgress)
		if err != nil {
			log.Printf("Error creating %s: %v", process.name, err)
			continue
		}
		fmt.Printf("  ✓ Created: %s-%s\n", process.name, process.version)
	}

	// List all in-progress state-machine diagrams
	fmt.Println("Listing all in-progress state-machine diagrams...")
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

	// Validate all and count valid ones
	fmt.Println("Batch validation of all state-machine diagrams...")
	validCount := 0
	for _, process := range processes {
		result, err := svc.Validate(models.DiagramTypePUML, process.name, process.version, diagram.LocationInProgress)
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

	fmt.Printf("✓ Batch validation completed: %d/%d state-machine diagrams are valid\n", validCount, len(processes))

	// Clean up all created state-machine diagrams
	fmt.Println("Cleaning up batch-created state-machine diagrams...")
	for _, process := range processes {
		err := svc.DeleteFile(models.DiagramTypePUML, process.name, process.version, diagram.LocationInProgress)
		if err != nil {
			log.Printf("Warning: Could not delete %s: %v", process.name, err)
		} else {
			fmt.Printf("  ✓ Deleted: %s-%s\n", process.name, process.version)
		}
	}
}

func demonstrateErrorHandling() {
	svc, err := diagram.NewService()
	if err != nil {
		log.Printf("Error creating service: %v", err)
		return
	}

	fmt.Println("Demonstrating various error scenarios and recovery...")

	// 1. Invalid parameters
	fmt.Println("1. Testing invalid parameters...")
	_, err = svc.CreateFile(models.DiagramTypePUML, "", "1.0.0", "content", diagram.LocationInProgress)
	if err != nil {
		fmt.Println("  ✓ Correctly handled empty name error")
	}

	_, err = svc.CreateFile(models.DiagramTypePUML, "test", "", "content", diagram.LocationInProgress)
	if err != nil {
		fmt.Println("  ✓ Correctly handled empty version error")
	}

	_, err = svc.CreateFile(models.DiagramTypePUML, "test", "1.0.0", "", diagram.LocationInProgress)
	if err != nil {
		fmt.Println("  ✓ Correctly handled empty content error")
	}

	// 2. Non-existent operations
	fmt.Println("2. Testing non-existent resource operations...")
	_, err = svc.ReadFile(models.DiagramTypePUML, "non-existent", "1.0.0", diagram.LocationInProgress)
	if err != nil {
		fmt.Println("  ✓ Correctly handled non-existent read error")
	}

	err = svc.DeleteFile(models.DiagramTypePUML, "non-existent", "1.0.0", diagram.LocationInProgress)
	if err != nil {
		fmt.Println("  ✓ Correctly handled non-existent delete error")
	}

	err = svc.Promote(models.DiagramTypePUML, "non-existent", "1.0.0")
	if err != nil {
		fmt.Println("  ✓ Correctly handled non-existent promote error")
	}

	// 3. Duplicate creation
	fmt.Println("3. Testing duplicate creation...")
	content := `@startuml
[*] --> Test
Test --> [*]
@enduml`

	_, err = svc.CreateFile(models.DiagramTypePUML, "duplicate-test", "1.0.0", content, diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error creating first instance: %v", err)
		return
	}

	_, err = svc.CreateFile(models.DiagramTypePUML, "duplicate-test", "1.0.0", content, diagram.LocationInProgress)
	if err != nil {
		fmt.Println("  ✓ Correctly handled duplicate creation error")
	}

	// 4. Update non-existent
	fmt.Println("4. Testing update of non-existent state-machine diagram...")
	nonExistentDiag := &diagram.StateMachineDiagram{
		Name:        "non-existent",
		Version:     "1.0.0",
		Content:     content,
		Location:    diagram.LocationInProgress,
		DiagramType: models.DiagramTypePUML,
	}
	err = svc.UpdateInProgressFile(nonExistentDiag)
	if err != nil {
		fmt.Println("  ✓ Correctly handled update of non-existent state-machine diagram")
	}

	// Clean up
	svc.DeleteFile(models.DiagramTypePUML, "duplicate-test", "1.0.0", diagram.LocationInProgress)

	fmt.Println("✓ Error handling demonstration completed")
}

func demonstrateEnvironmentConfiguration() {
	fmt.Println("Demonstrating environment-based configuration...")

	// Set some environment variables
	originalRootDir := os.Getenv("GO_UML_ROOT_DIRECTORY")
	originalDebugLogging := os.Getenv("GO_UML_DEBUG_LOGGING")

	os.Setenv("GO_UML_ROOT_DIRECTORY", ".env-demo-diagrams")
	os.Setenv("GO_UML_DEBUG_LOGGING", "false")   // Keep demo clean
	os.Setenv("GO_UML_MAX_FILE_SIZE", "5242880") // 5MB

	// Load configuration from environment
	config := diagram.LoadConfigFromEnv()
	fmt.Printf("✓ Environment configuration loaded:\n")
	fmt.Printf("  - Root Directory: %s\n", config.RootDirectory)
	fmt.Printf("  - Debug Logging: %t\n", config.EnableDebugLogging)
	fmt.Printf("  - Max File Size: %d bytes (%.1f MB)\n", config.MaxFileSize, float64(config.MaxFileSize)/(1024*1024))

	// Create service from environment
	svc, err := diagram.NewServiceFromEnv()
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

	diag, err := svc.CreateFile(models.DiagramTypePUML, "env-config-test", "1.0.0", content, diagram.LocationInProgress)
	if err != nil {
		log.Printf("Error creating test state-machine diagram: %v", err)
	} else {
		fmt.Printf("✓ Created state-machine diagram using environment configuration: %s-%s\n", diag.Name, diag.Version)

		// Clean up
		svc.DeleteFile(models.DiagramTypePUML, "env-config-test", "1.0.0", diagram.LocationInProgress)
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
