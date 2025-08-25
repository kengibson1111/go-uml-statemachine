package main

import (
	"fmt"
	"log"

	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/repository"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/service"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/validation"
)

func main() {
	fmt.Println("Go UML State Machine - Advanced Usage Example")
	fmt.Println("==============================================")

	// Create configuration with custom settings
	config := &models.Config{
		RootDirectory:      ".go-uml-statemachine-parsers",
		ValidationLevel:    models.StrictnessInProgress,
		BackupEnabled:      true,
		MaxFileSize:        2 * 1024 * 1024, // 2MB
		EnableDebugLogging: true,
	}

	// Create dependencies
	repo := repository.NewFileSystemRepository(config)
	validator := validation.NewPlantUMLValidatorWithRepository(repo)
	svc := service.NewService(repo, validator, config)

	// Example 1: Create a base authentication state machine
	fmt.Println("\n1. Creating base authentication state machine...")
	baseAuthContent := `@startuml
title Base Authentication

[*] --> Idle
Idle --> Authenticating : login()
Authenticating --> Authenticated : success
Authenticating --> Failed : failure
Failed --> Idle : retry()
Authenticated --> Idle : logout()

@enduml`

	baseSM, err := svc.Create(models.FileTypePUML, "base-auth", "1.0.0", baseAuthContent, models.LocationInProgress)
	if err != nil {
		log.Printf("Error creating base auth: %v", err)
		return
	}
	fmt.Printf("✓ Created: %s-%s\n", baseSM.Name, baseSM.Version)

	// Validate and promote base auth to products
	validationResult, err := svc.Validate(models.FileTypePUML, "base-auth", "1.0.0", models.LocationInProgress)
	if err != nil {
		log.Printf("Error validating base auth: %v", err)
		return
	}

	if validationResult.IsValid && !validationResult.HasErrors() {
		err = svc.Promote(models.FileTypePUML, "base-auth", "1.0.0")
		if err != nil {
			log.Printf("Error promoting base auth: %v", err)
			return
		}
		fmt.Println("✓ Base auth promoted to products")
	}

	// Example 2: Create an advanced auth system that references the base
	fmt.Println("\n2. Creating advanced authentication with references...")
	advancedAuthContent := `@startuml
title Advanced Authentication System

' Reference to base authentication
!include products/base-auth-1.0.0/base-auth-1.0.0.puml

[*] --> CheckSession
CheckSession --> SessionValid : valid_session
CheckSession --> RequireAuth : no_session
SessionValid --> Authenticated
RequireAuth --> base-auth : delegate_auth
base-auth --> TwoFactorAuth : auth_success
TwoFactorAuth --> Authenticated : 2fa_success
TwoFactorAuth --> Failed : 2fa_failed
Failed --> RequireAuth : retry

@enduml`

	advancedSM, err := svc.Create(models.FileTypePUML, "advanced-auth", "1.0.0", advancedAuthContent, models.LocationInProgress)
	if err != nil {
		log.Printf("Error creating advanced auth: %v", err)
		return
	}
	fmt.Printf("✓ Created: %s-%s\n", advancedSM.Name, advancedSM.Version)

	// Example 3: Resolve references in the advanced auth system
	fmt.Println("\n3. Resolving references...")
	err = svc.ResolveReferences(advancedSM)
	if err != nil {
		log.Printf("Error resolving references: %v", err)
	} else {
		fmt.Printf("✓ Resolved %d references\n", len(advancedSM.References))
		for _, ref := range advancedSM.References {
			fmt.Printf("  - %s (type: %s, path: %s)\n", ref.Name, ref.Type.String(), ref.Path)
		}
	}

	// Example 4: Update an existing state machine
	fmt.Println("\n4. Updating state machine...")
	updatedContent := `@startuml
title Advanced Authentication System v1.1

' Reference to base authentication
!include products/base-auth-1.0.0/base-auth-1.0.0.puml

[*] --> CheckSession
CheckSession --> SessionValid : valid_session
CheckSession --> RequireAuth : no_session
SessionValid --> Authenticated
RequireAuth --> base-auth : delegate_auth
base-auth --> TwoFactorAuth : auth_success
TwoFactorAuth --> Authenticated : 2fa_success
TwoFactorAuth --> Failed : 2fa_failed
Failed --> RequireAuth : retry

' New: Add session timeout handling
Authenticated --> SessionTimeout : timeout
SessionTimeout --> RequireAuth : session_expired

@enduml`

	advancedSM.Content = updatedContent
	err = svc.Update(advancedSM)
	if err != nil {
		log.Printf("Error updating state machine: %v", err)
	} else {
		fmt.Println("✓ State machine updated successfully")
	}

	// Example 5: Validate with different strictness levels
	fmt.Println("\n5. Testing validation strictness...")

	// Validate with in-progress strictness (errors and warnings)
	inProgressResult, err := svc.Validate(models.FileTypePUML, "advanced-auth", "1.0.0", models.LocationInProgress)
	if err != nil {
		log.Printf("Error validating with in-progress strictness: %v", err)
	} else {
		fmt.Printf("✓ In-progress validation: Valid=%t, Errors=%d, Warnings=%d\n",
			inProgressResult.IsValid, len(inProgressResult.Errors), len(inProgressResult.Warnings))
	}

	// Example 6: Demonstrate error handling
	fmt.Println("\n6. Demonstrating error handling...")

	// Try to create a duplicate
	_, err = svc.Create(models.FileTypePUML, "advanced-auth", "1.0.0", advancedAuthContent, models.LocationInProgress)
	if err != nil {
		fmt.Printf("✓ Expected error for duplicate creation: %v\n", err)
	}

	// Try to read non-existent state machine
	_, err = svc.Read(models.FileTypePUML, "non-existent", "1.0.0", models.LocationInProgress)
	if err != nil {
		fmt.Printf("✓ Expected error for non-existent read: %v\n", err)
	}

	// Try to promote without validation
	err = svc.Promote(models.FileTypePUML, "non-existent", "1.0.0")
	if err != nil {
		fmt.Printf("✓ Expected error for non-existent promotion: %v\n", err)
	}

	// Example 7: Configuration from environment
	fmt.Println("\n7. Demonstrating environment configuration...")
	envConfig := models.LoadConfigFromEnv()
	fmt.Printf("✓ Environment config loaded:\n")
	fmt.Printf("  - Root Directory: %s\n", envConfig.RootDirectory)
	fmt.Printf("  - Validation Level: %s\n", envConfig.ValidationLevel.String())
	fmt.Printf("  - Backup Enabled: %t\n", envConfig.BackupEnabled)
	fmt.Printf("  - Max File Size: %d bytes\n", envConfig.MaxFileSize)
	fmt.Printf("  - Debug Logging: %t\n", envConfig.EnableDebugLogging)

	// Example 8: Clean up (delete the test state machines)
	fmt.Println("\n8. Cleaning up test data...")

	// Delete from products first
	err = svc.Delete(models.FileTypePUML, "base-auth", "1.0.0", models.LocationProducts)
	if err != nil {
		log.Printf("Warning: Could not delete base-auth from products: %v", err)
	} else {
		fmt.Println("✓ Deleted base-auth from products")
	}

	// Delete from in-progress
	err = svc.Delete(models.FileTypePUML, "advanced-auth", "1.0.0", models.LocationInProgress)
	if err != nil {
		log.Printf("Warning: Could not delete advanced-auth from in-progress: %v", err)
	} else {
		fmt.Println("✓ Deleted advanced-auth from in-progress")
	}

	fmt.Println("\n✓ Advanced example completed!")
}
