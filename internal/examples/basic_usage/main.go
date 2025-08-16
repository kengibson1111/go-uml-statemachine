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
	fmt.Println("Go UML State Machine - Basic Usage Example")
	fmt.Println("==========================================")

	// Create configuration
	config := models.DefaultConfig()
	config.EnableDebugLogging = true
	fmt.Printf("Using root directory: %s\n", config.RootDirectory)

	// Create repository
	repo := repository.NewFileSystemRepository(config)

	// Create validator
	validator := validation.NewPlantUMLValidatorWithRepository(repo)

	// Create service
	svc := service.NewService(repo, validator, config)

	// Example PlantUML content for a simple user authentication state machine
	authContent := `@startuml
title User Authentication State Machine

[*] --> Idle : Start

Idle --> Authenticating : login()
Authenticating --> Authenticated : success
Authenticating --> Failed : failure
Failed --> Idle : retry()
Authenticated --> Idle : logout()

@enduml`

	// Example 1: Create a new state machine in in-progress
	fmt.Println("\n1. Creating a new state machine...")
	sm, err := svc.Create("user-auth", "1.0.0", authContent, models.LocationInProgress)
	if err != nil {
		log.Printf("Error creating state machine: %v", err)
		return
	}
	fmt.Printf("✓ Created state machine: %s-%s\n", sm.Name, sm.Version)

	// Example 2: Read the state machine back
	fmt.Println("\n2. Reading the state machine...")
	readSM, err := svc.Read("user-auth", "1.0.0", models.LocationInProgress)
	if err != nil {
		log.Printf("Error reading state machine: %v", err)
		return
	}
	fmt.Printf("✓ Read state machine: %s-%s (content length: %d)\n",
		readSM.Name, readSM.Version, len(readSM.Content))

	// Example 3: Validate the state machine
	fmt.Println("\n3. Validating the state machine...")
	validationResult, err := svc.Validate("user-auth", "1.0.0", models.LocationInProgress)
	if err != nil {
		log.Printf("Error validating state machine: %v", err)
		return
	}
	fmt.Printf("✓ Validation result: Valid=%t, Errors=%d, Warnings=%d\n",
		validationResult.IsValid, len(validationResult.Errors), len(validationResult.Warnings))

	// Show validation details if there are issues
	if len(validationResult.Errors) > 0 {
		fmt.Println("  Errors:")
		for _, err := range validationResult.Errors {
			fmt.Printf("    - %s: %s\n", err.Code, err.Message)
		}
	}
	if len(validationResult.Warnings) > 0 {
		fmt.Println("  Warnings:")
		for _, warning := range validationResult.Warnings {
			fmt.Printf("    - %s: %s\n", warning.Code, warning.Message)
		}
	}

	// Example 4: List all state machines in in-progress
	fmt.Println("\n4. Listing all in-progress state machines...")
	stateMachines, err := svc.ListAll(models.LocationInProgress)
	if err != nil {
		log.Printf("Error listing state machines: %v", err)
		return
	}
	fmt.Printf("✓ Found %d state machine(s) in in-progress:\n", len(stateMachines))
	for _, sm := range stateMachines {
		fmt.Printf("  - %s-%s (created: %s)\n",
			sm.Name, sm.Version, sm.Metadata.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Example 5: Promote to products (if validation passes)
	if validationResult.IsValid && !validationResult.HasErrors() {
		fmt.Println("\n5. Promoting state machine to products...")
		err = svc.Promote("user-auth", "1.0.0")
		if err != nil {
			log.Printf("Error promoting state machine: %v", err)
		} else {
			fmt.Println("✓ Successfully promoted to products")
		}
	} else {
		fmt.Println("\n5. Skipping promotion due to validation issues")
	}

	// Example 6: List products
	fmt.Println("\n6. Listing all product state machines...")
	productSMs, err := svc.ListAll(models.LocationProducts)
	if err != nil {
		log.Printf("Error listing product state machines: %v", err)
		return
	}
	fmt.Printf("✓ Found %d state machine(s) in products:\n", len(productSMs))
	for _, sm := range productSMs {
		fmt.Printf("  - %s-%s (created: %s)\n",
			sm.Name, sm.Version, sm.Metadata.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	fmt.Println("\n✓ Example completed successfully!")
}
