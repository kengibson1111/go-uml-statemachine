package main

import (
	"fmt"
	"log"

	"github.com/kengibson1111/go-uml-statemachine/internal/models"
	"github.com/kengibson1111/go-uml-statemachine/internal/validation"
)

func main() {
	fmt.Println("=== Reference Validation and Resolution Demo ===")

	// Create a validator
	validator := validation.NewPlantUMLValidator()

	// Test 1: State machine with valid product reference
	fmt.Println("\n1. Testing valid product reference:")
	sm1 := &models.StateMachine{
		Name:    "main-service",
		Version: "1.0.0",
		Content: `@startuml
!include products/auth-service-1.2.0/auth-service-1.2.0.puml
[*] --> Idle
Idle --> Authenticating
Authenticating --> Authenticated
Authenticated --> [*]
@enduml`,
	}

	result1, err := validator.ValidateReferences(sm1)
	if err != nil {
		log.Fatalf("Error validating references: %v", err)
	}

	fmt.Printf("Valid: %t\n", result1.IsValid)
	fmt.Printf("References found: %d\n", len(sm1.References))
	if len(sm1.References) > 0 {
		ref := sm1.References[0]
		fmt.Printf("  - Name: %s, Version: %s, Type: %s, Path: %s\n",
			ref.Name, ref.Version, ref.Type.String(), ref.Path)
	}
	fmt.Printf("Errors: %d, Warnings: %d\n", len(result1.Errors), len(result1.Warnings))

	// Test 2: State machine with valid nested reference
	fmt.Println("\n2. Testing valid nested reference:")
	sm2 := &models.StateMachine{
		Name:    "workflow",
		Version: "2.0.0",
		Content: `@startuml
!include nested/sub-workflow/sub-workflow.puml
[*] --> Start
Start --> Processing
Processing --> End
End --> [*]
@enduml`,
	}

	result2, err := validator.ValidateReferences(sm2)
	if err != nil {
		log.Fatalf("Error validating references: %v", err)
	}

	fmt.Printf("Valid: %t\n", result2.IsValid)
	fmt.Printf("References found: %d\n", len(sm2.References))
	if len(sm2.References) > 0 {
		ref := sm2.References[0]
		fmt.Printf("  - Name: %s, Version: %s, Type: %s, Path: %s\n",
			ref.Name, ref.Version, ref.Type.String(), ref.Path)
	}
	fmt.Printf("Errors: %d, Warnings: %d\n", len(result2.Errors), len(result2.Warnings))

	// Test 3: State machine with invalid version
	fmt.Println("\n3. Testing invalid product version:")
	sm3 := &models.StateMachine{
		Name:    "test-service",
		Version: "1.0.0",
		Content: `@startuml
!include products/bad-service-invalid.version/bad-service-invalid.version.puml
[*] --> Idle
@enduml`,
	}

	result3, err := validator.ValidateReferences(sm3)
	if err != nil {
		log.Fatalf("Error validating references: %v", err)
	}

	fmt.Printf("Valid: %t\n", result3.IsValid)
	fmt.Printf("References found: %d\n", len(sm3.References))
	fmt.Printf("Errors: %d, Warnings: %d\n", len(result3.Errors), len(result3.Warnings))
	if len(result3.Errors) > 0 {
		fmt.Printf("  Error: %s - %s\n", result3.Errors[0].Code, result3.Errors[0].Message)
	}

	// Test 4: State machine with self-reference
	fmt.Println("\n4. Testing self-reference:")
	sm4 := &models.StateMachine{
		Name:    "self-ref",
		Version: "1.0.0",
		Content: `@startuml
!include products/self-ref-1.0.0/self-ref-1.0.0.puml
[*] --> Idle
@enduml`,
	}

	result4, err := validator.ValidateReferences(sm4)
	if err != nil {
		log.Fatalf("Error validating references: %v", err)
	}

	fmt.Printf("Valid: %t\n", result4.IsValid)
	fmt.Printf("References found: %d\n", len(sm4.References))
	fmt.Printf("Errors: %d, Warnings: %d\n", len(result4.Errors), len(result4.Warnings))
	if len(result4.Errors) > 0 {
		fmt.Printf("  Error: %s - %s\n", result4.Errors[0].Code, result4.Errors[0].Message)
	}

	// Test 5: Multiple references
	fmt.Println("\n5. Testing multiple references:")
	sm5 := &models.StateMachine{
		Name:    "complex-service",
		Version: "1.0.0",
		Content: `@startuml
!include products/auth-service-1.2.0/auth-service-1.2.0.puml
!include products/payment-service-2.1.0/payment-service-2.1.0.puml
!include nested/validation/validation.puml
!include nested/logging/logging.puml
[*] --> Idle
@enduml`,
	}

	result5, err := validator.ValidateReferences(sm5)
	if err != nil {
		log.Fatalf("Error validating references: %v", err)
	}

	fmt.Printf("Valid: %t\n", result5.IsValid)
	fmt.Printf("References found: %d\n", len(sm5.References))
	for i, ref := range sm5.References {
		fmt.Printf("  %d. Name: %s, Version: %s, Type: %s\n",
			i+1, ref.Name, ref.Version, ref.Type.String())
	}
	fmt.Printf("Errors: %d, Warnings: %d\n", len(result5.Errors), len(result5.Warnings))

	fmt.Println("\n=== Demo Complete ===")
}
