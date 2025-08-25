package main

import (
	"fmt"
	"log"

	"github.com/kengibson1111/go-uml-statemachine-parsers/statemachine"
)

func main() {
	fmt.Println("Testing public API...")

	// Test NewService
	svc, err := statemachine.NewService()
	if err != nil {
		log.Fatal(err)
	}

	// Test NewServiceWithConfig
	config := statemachine.DefaultConfig()
	config.EnableDebugLogging = false
	svc2, err := statemachine.NewServiceWithConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// Test NewServiceFromEnv
	svc3, err := statemachine.NewServiceFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	// Test basic operations
	content := `@startuml
[*] --> Test
Test --> [*]
@enduml`

	sm, err := svc.Create(statemachine.FileTypePUML, "api-test", "1.0.0", content, statemachine.LocationInProgress)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✓ Created state machine: %s-%s\n", sm.Name, sm.Version)

	// Clean up
	err = svc.Delete(statemachine.FileTypePUML, "api-test", "1.0.0", statemachine.LocationInProgress)
	if err != nil {
		log.Printf("Warning: Could not clean up: %v", err)
	}

	// Suppress unused variable warnings
	_ = svc2
	_ = svc3

	fmt.Println("✓ Public API test completed successfully!")
}
