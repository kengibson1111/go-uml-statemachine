package main

import (
	"fmt"
	"log"

	smmodels "github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/repository"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/service"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/validation"
)

func main() {
	fmt.Println("Go UML State-Machine Diagram - Basic Usage Example")
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

	// Example PlantUML content for a simple user authentication state-machine diagram
	authContent := `@startuml
title User Authentication State-Machine Diagram

[*] --> Idle : Start

Idle --> Authenticating : login()
Authenticating --> Authenticated : success
Authenticating --> Failed : failure
Failed --> Idle : retry()
Authenticated --> Idle : logout()

@enduml`

	// Example 1: Create a new state-machine diagram in in-progress
	fmt.Println("\n1. Creating a new state-machine diagram...")
	diag, err := svc.CreateFile(smmodels.DiagramTypePUML, "user-auth", "1.0.0", authContent, models.LocationFileInProgress)
	if err != nil {
		log.Printf("Error creating state-machine diagram: %v", err)
		return
	}
	fmt.Printf("✓ Created state-machine diagram: %s-%s\n", diag.Name, diag.Version)

	// Example 2: Read the state-machine diagram back
	fmt.Println("\n2. Reading the state-machine diagram...")
	readDiag, err := svc.ReadFile(smmodels.DiagramTypePUML, "user-auth", "1.0.0", models.LocationFileInProgress)
	if err != nil {
		log.Printf("Error reading state-machine diagram: %v", err)
		return
	}
	fmt.Printf("✓ Read state-machine diagram: %s-%s (content length: %d)\n",
		readDiag.Name, readDiag.Version, len(readDiag.Content))

	// Example 3: Validate the state-machine diagram
	fmt.Println("\n3. Validating the state-machine diagram...")
	validationResult, err := svc.ValidateFile(smmodels.DiagramTypePUML, "user-auth", "1.0.0", models.LocationFileInProgress)
	if err != nil {
		log.Printf("Error validating state-machine diagram: %v", err)
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

	// Example 4: List all state-machine diagrams in in-progress
	fmt.Println("\n4. Listing all in-progress state-machine diagrams...")
	diagrams, err := svc.ListAllFiles(smmodels.DiagramTypePUML, models.LocationFileInProgress)
	if err != nil {
		log.Printf("Error listing state-machine diagrams: %v", err)
		return
	}
	fmt.Printf("✓ Found %d state-machine diagram(s) in in-progress:\n", len(diagrams))
	for _, diag := range diagrams {
		fmt.Printf("  - %s-%s (created: %s)\n",
			diag.Name, diag.Version, diag.Metadata.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Example 5: Promote to products (if validation passes)
	if validationResult.IsValid && !validationResult.HasErrors() {
		fmt.Println("\n5. Promoting state-machine diagram to products...")
		err = svc.PromoteToProductsFile(smmodels.DiagramTypePUML, "user-auth", "1.0.0")
		if err != nil {
			log.Printf("Error promoting state-machine diagram: %v", err)
		} else {
			fmt.Println("✓ Successfully promoted to products")
		}
	} else {
		fmt.Println("\n5. Skipping promotion due to validation issues")
	}

	// Example 6: List products
	fmt.Println("\n6. Listing all product state-machine diagrams...")
	productDiags, err := svc.ListAllFiles(smmodels.DiagramTypePUML, models.LocationFileProducts)
	if err != nil {
		log.Printf("Error listing product state-machine diagrams: %v", err)
		return
	}
	fmt.Printf("✓ Found %d state-machine diagram(s) in products:\n", len(productDiags))
	for _, diag := range productDiags {
		fmt.Printf("  - %s-%s (created: %s)\n",
			diag.Name, diag.Version, diag.Metadata.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	fmt.Println("\n✓ Example completed successfully!")
}
