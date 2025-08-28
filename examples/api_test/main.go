package main

import (
	"fmt"
	"log"

	"github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/diagram"
)

func main() {
	fmt.Println("Testing public API...")

	// Test NewService
	svc, err := diagram.NewService()
	if err != nil {
		log.Fatal(err)
	}

	// Test NewServiceWithConfig
	config := diagram.DefaultConfig()
	config.EnableDebugLogging = false
	svc2, err := diagram.NewServiceWithConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// Test NewServiceFromEnv
	svc3, err := diagram.NewServiceFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	// Test basic operations
	content := `@startuml
[*] --> Test
Test --> [*]
@enduml`

	diag, err := svc.CreateFile(models.DiagramTypePUML, "api-test", "1.0.0", content, diagram.LocationFileInProgress)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✓ Created state-machine diagram: %s-%s\n", diag.Name, diag.Version)

	// Clean up
	err = svc.DeleteFile(models.DiagramTypePUML, "api-test", "1.0.0", diagram.LocationFileInProgress)
	if err != nil {
		log.Printf("Warning: Could not clean up: %v", err)
	}

	// Suppress unused variable warnings
	_ = svc2
	_ = svc3

	fmt.Println("✓ Public API test completed successfully!")
}
