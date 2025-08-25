package main

import (
	"fmt"
	"log"

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

	sm, err := svc.Create(diagram.FileTypePUML, "api-test", "1.0.0", content, diagram.LocationInProgress)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✓ Created state-machine diagram: %s-%s\n", sm.Name, sm.Version)

	// Clean up
	err = svc.Delete(diagram.FileTypePUML, "api-test", "1.0.0", diagram.LocationInProgress)
	if err != nil {
		log.Printf("Warning: Could not clean up: %v", err)
	}

	// Suppress unused variable warnings
	_ = svc2
	_ = svc3

	fmt.Println("✓ Public API test completed successfully!")
}
