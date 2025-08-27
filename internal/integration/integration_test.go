package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	smmodels "github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/repository"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/service"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/validation"
)

// TestFixture represents test data for integration tests
type TestFixture struct {
	Name     string
	Version  string
	Content  string
	Location models.Location
}

// Integration test fixtures with sample PlantUML content
var testFixtures = []TestFixture{
	{
		Name:     "user-auth",
		Version:  "1.0.0",
		Content:  "@startuml\n[*] --> Idle\nIdle --> Authenticating : login\nAuthenticating --> Authenticated : success\nAuthenticating --> Failed : failure\nFailed --> Idle : retry\nAuthenticated --> [*] : logout\n@enduml",
		Location: models.LocationInProgress,
	},
	{
		Name:     "payment-flow",
		Version:  "2.1.0",
		Content:  "@startuml\n[*] --> Pending\nPending --> Processing : process\nProcessing --> Completed : success\nProcessing --> Failed : error\nFailed --> Pending : retry\nCompleted --> [*]\n@enduml",
		Location: models.LocationInProgress,
	},
	{
		Name:     "order-management",
		Version:  "1.5.2",
		Content:  "@startuml\n[*] --> Created\nCreated --> Confirmed : confirm\nConfirmed --> Shipped : ship\nShipped --> Delivered : deliver\nDelivered --> [*]\nCreated --> Cancelled : cancel\nConfirmed --> Cancelled : cancel\n@enduml",
		Location: models.LocationInProgress,
	},
	{
		Name:     "notification-system",
		Version:  "3.0.0-beta",
		Content:  "@startuml\n[*] --> Queued\nQueued --> Sending : send\nSending --> Sent : success\nSending --> Failed : error\nFailed --> Queued : retry\nSent --> [*]\n@enduml",
		Location: models.LocationInProgress,
	},
}

// Test fixtures with references for testing reference resolution
var referencedFixtures = []TestFixture{
	{
		Name:     "main-workflow",
		Version:  "1.0.0",
		Content:  "@startuml\n!include products/user-auth-1.0.0/user-auth-1.0.0.puml\n[*] --> Start\nStart --> UserAuth\nUserAuth --> [*]\n@enduml",
		Location: models.LocationInProgress,
	},
}

// setupTestEnvironment creates a temporary directory for testing
func setupTestEnvironment(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "go-uml-statemachine-parsers-integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: Failed to cleanup temp directory %s: %v", tempDir, err)
		}
	}

	return tempDir, cleanup
}

// createTestService creates a service instance for testing
func createTestService(rootDir string) models.DiagramService {
	config := &models.Config{
		RootDirectory: rootDir,
		MaxFileSize:   1024 * 1024, // 1MB
	}

	repo := repository.NewFileSystemRepository(config)
	validator := validation.NewPlantUMLValidatorWithRepository(repo)
	return service.NewService(repo, validator, config)
}

// TestCompleteWorkflowFromCreationToPromotion tests the complete workflow
func TestCompleteWorkflowFromCreationToPromotion(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	svc := createTestService(tempDir)

	// Test fixture
	fixture := testFixtures[0] // user-auth

	t.Run("Create state-machine diagram in in-progress", func(t *testing.T) {
		diag, err := svc.CreateFile(smmodels.DiagramTypePUML, fixture.Name, fixture.Version, fixture.Content, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to create state-machine diagram: %v", err)
		}

		if diag.Name != fixture.Name {
			t.Errorf("Expected name %s, got %s", fixture.Name, diag.Name)
		}
		if diag.Version != fixture.Version {
			t.Errorf("Expected version %s, got %s", fixture.Version, diag.Version)
		}
		if diag.Location != models.LocationInProgress {
			t.Errorf("Expected location %s, got %s", models.LocationInProgress.String(), diag.Location.String())
		}
	})

	t.Run("Read state-machine diagram from in-progress", func(t *testing.T) {
		diagram, err := svc.ReadFile(smmodels.DiagramTypePUML, fixture.Name, fixture.Version, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to read state-machine diagram: %v", err)
		}

		if diagram.Content != fixture.Content {
			t.Errorf("Content mismatch. Expected:\n%s\nGot:\n%s", fixture.Content, diagram.Content)
		}
	})

	t.Run("Update state-machine diagram content", func(t *testing.T) {
		// Read the current state-machine diagram
		diag, err := svc.ReadFile(smmodels.DiagramTypePUML, fixture.Name, fixture.Version, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to read state-machine diagram for update: %v", err)
		}

		// Update content
		updatedContent := fixture.Content + "\n' Updated content"
		diag.Content = updatedContent

		// Update the state-machine diagram
		err = svc.UpdateInProgressFile(diag)
		if err != nil {
			t.Fatalf("Failed to update state-machine diagram: %v", err)
		}

		// Verify the update
		updatedDiag, err := svc.ReadFile(smmodels.DiagramTypePUML, fixture.Name, fixture.Version, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to read updated state-machine diagram: %v", err)
		}

		if updatedDiag.Content != updatedContent {
			t.Errorf("Update failed. Expected:\n%s\nGot:\n%s", updatedContent, updatedDiag.Content)
		}
	})

	t.Run("Validate state-machine diagram", func(t *testing.T) {
		result, err := svc.Validate(smmodels.DiagramTypePUML, fixture.Name, fixture.Version, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to validate state-machine diagram: %v", err)
		}

		if !result.IsValid {
			t.Errorf("State-machine diagram should be valid. Errors: %v", result.Errors)
		}
	})

	t.Run("List state-machine diagrams in in-progress", func(t *testing.T) {
		diagrams, err := svc.ListAll(smmodels.DiagramTypePUML, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to list state-machine diagrams: %v", err)
		}

		found := false
		for _, diag := range diagrams {
			if diag.Name == fixture.Name && diag.Version == fixture.Version {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Created state-machine diagram not found in list")
		}
	})

	t.Run("Promote state-machine diagram to products", func(t *testing.T) {
		err := svc.Promote(smmodels.DiagramTypePUML, fixture.Name, fixture.Version)
		if err != nil {
			t.Fatalf("Failed to promote state-machine diagram: %v", err)
		}

		// Verify it exists in products
		diagram, err := svc.ReadFile(smmodels.DiagramTypePUML, fixture.Name, fixture.Version, models.LocationProducts)
		if err != nil {
			t.Fatalf("Failed to read promoted state-machine diagram: %v", err)
		}

		if diagram.Location != models.LocationProducts {
			t.Errorf("Expected location %s, got %s", models.LocationProducts.String(), diagram.Location.String())
		}

		// Verify it no longer exists in in-progress
		_, err = svc.ReadFile(smmodels.DiagramTypePUML, fixture.Name, fixture.Version, models.LocationInProgress)
		if err == nil {
			t.Errorf("State-machine diagram should no longer exist in in-progress after promotion")
		}
	})

	t.Run("List state-machine diagrams in products", func(t *testing.T) {
		diagrams, err := svc.ListAll(smmodels.DiagramTypePUML, models.LocationProducts)
		if err != nil {
			t.Fatalf("Failed to list state-machine diagrams in products: %v", err)
		}

		found := false
		for _, diag := range diagrams {
			if diag.Name == fixture.Name && diag.Version == fixture.Version {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Promoted state-machine diagram not found in products list")
		}
	})

	t.Run("Delete state-machine diagram from products", func(t *testing.T) {
		err := svc.DeleteFile(smmodels.DiagramTypePUML, fixture.Name, fixture.Version, models.LocationProducts)
		if err != nil {
			t.Fatalf("Failed to delete state-machine diagram: %v", err)
		}

		// Verify it no longer exists
		_, err = svc.ReadFile(smmodels.DiagramTypePUML, fixture.Name, fixture.Version, models.LocationProducts)
		if err == nil {
			t.Errorf("State-machine diagram should not exist after deletion")
		}
	})
}

// TestErrorScenariosAndEdgeCases tests various error conditions
func TestErrorScenariosAndEdgeCases(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	svc := createTestService(tempDir)
	fixture := testFixtures[1] // payment-flow

	t.Run("Create duplicate state-machine diagram", func(t *testing.T) {
		// Create first state-machine diagram
		_, err := svc.CreateFile(smmodels.DiagramTypePUML, fixture.Name, fixture.Version, fixture.Content, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to create first state-machine diagram: %v", err)
		}

		// Try to create duplicate
		_, err = svc.CreateFile(smmodels.DiagramTypePUML, fixture.Name, fixture.Version, fixture.Content, models.LocationInProgress)
		if err == nil {
			t.Errorf("Should not be able to create duplicate state-machine diagram")
		}

		// Verify error type
		if diagErr, ok := err.(*models.StateMachineError); ok {
			if diagErr.Type != models.ErrorTypeDirectoryConflict {
				t.Errorf("Expected ErrorTypeDirectoryConflict, got %v", diagErr.Type)
			}
		} else {
			t.Errorf("Expected StateMachineError, got %T", err)
		}
	})

	t.Run("Read non-existent state-machine diagram", func(t *testing.T) {
		_, err := svc.ReadFile(smmodels.DiagramTypePUML, "non-existent", "1.0.0", models.LocationInProgress)
		if err == nil {
			t.Errorf("Should not be able to read non-existent state-machine diagram")
		}

		// Verify error type
		if diagErr, ok := err.(*models.StateMachineError); ok {
			if diagErr.Type != models.ErrorTypeFileNotFound {
				t.Errorf("Expected ErrorTypeFileNotFound, got %v", diagErr.Type)
			}
		}
	})

	t.Run("Update non-existent state-machine diagram", func(t *testing.T) {
		diag := &models.StateMachineDiagram{
			Name:     "non-existent",
			Version:  "1.0.0",
			Content:  "@startuml\n[*] --> Test\n@enduml",
			Location: models.LocationInProgress,
		}

		err := svc.UpdateInProgressFile(diag)
		if err == nil {
			t.Errorf("Should not be able to update non-existent state-machine diagram")
		}

		// Verify error type
		if diagErr, ok := err.(*models.StateMachineError); ok {
			if diagErr.Type != models.ErrorTypeFileNotFound {
				t.Errorf("Expected ErrorTypeFileNotFound, got %v", diagErr.Type)
			}
		}
	})

	t.Run("Promote non-existent state-machine diagram", func(t *testing.T) {
		err := svc.Promote(smmodels.DiagramTypePUML, "non-existent", "1.0.0")
		if err == nil {
			t.Errorf("Should not be able to promote non-existent state-machine diagram")
		}

		// Verify error type
		if diagErr, ok := err.(*models.StateMachineError); ok {
			if diagErr.Type != models.ErrorTypeFileNotFound {
				t.Errorf("Expected ErrorTypeFileNotFound, got %v", diagErr.Type)
			}
		}
	})

	t.Run("Promote with validation errors", func(t *testing.T) {
		// Create state-machine diagram with invalid PlantUML content
		invalidContent := "@startuml\n' Missing @enduml tag"
		_, err := svc.CreateFile(smmodels.DiagramTypePUML, "invalid-diag", "1.0.0", invalidContent, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to create invalid state-machine diagram: %v", err)
		}

		// Try to promote - should fail due to validation errors
		err = svc.Promote(smmodels.DiagramTypePUML, "invalid-diag", "1.0.0")
		if err == nil {
			t.Errorf("Should not be able to promote state-machine diagram with validation errors")
		}

		// Verify error type
		if diagErr, ok := err.(*models.StateMachineError); ok {
			if diagErr.Type != models.ErrorTypeValidation {
				t.Errorf("Expected ErrorTypeValidation, got %v", diagErr.Type)
			}
		}
	})

	t.Run("Create in-progress when products exists", func(t *testing.T) {
		// First create and promote a state-machine diagram
		_, err := svc.CreateFile(smmodels.DiagramTypePUML, "conflict-test", "1.0.0", fixture.Content, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to create state-machine diagram: %v", err)
		}

		err = svc.Promote(smmodels.DiagramTypePUML, "conflict-test", "1.0.0")
		if err != nil {
			t.Fatalf("Failed to promote state-machine diagram: %v", err)
		}

		// Now try to create in-progress with same name/version
		_, err = svc.CreateFile(smmodels.DiagramTypePUML, "conflict-test", "1.0.0", fixture.Content, models.LocationInProgress)
		if err == nil {
			t.Errorf("Should not be able to create in-progress when products exists")
		}

		// Verify error type
		if diagErr, ok := err.(*models.StateMachineError); ok {
			if diagErr.Type != models.ErrorTypeDirectoryConflict {
				t.Errorf("Expected ErrorTypeDirectoryConflict, got %v", diagErr.Type)
			}
		}
	})

	t.Run("Invalid input validation", func(t *testing.T) {
		// Test empty name
		_, err := svc.CreateFile(smmodels.DiagramTypePUML, "", "1.0.0", fixture.Content, models.LocationInProgress)
		if err == nil {
			t.Errorf("Should not accept empty name")
		}

		// Test empty version
		_, err = svc.CreateFile(smmodels.DiagramTypePUML, "test", "", fixture.Content, models.LocationInProgress)
		if err == nil {
			t.Errorf("Should not accept empty version")
		}

		// Test empty content
		_, err = svc.CreateFile(smmodels.DiagramTypePUML, "test", "1.0.0", "", models.LocationInProgress)
		if err == nil {
			t.Errorf("Should not accept empty content")
		}

		// Test nil state-machine diagram for update
		err = svc.UpdateInProgressFile(nil)
		if err == nil {
			t.Errorf("Should not accept nil state-machine diagram for update")
		}
	})
}

// TestConcurrentOperationsAndThreadSafety tests concurrent access
func TestConcurrentOperationsAndThreadSafety(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	svc := createTestService(tempDir)

	t.Run("Concurrent creates with different names", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, len(testFixtures))
		successes := make(chan string, len(testFixtures))

		// Create multiple state-machine diagrams concurrently
		for i, fixture := range testFixtures {
			wg.Add(1)
			go func(idx int, f TestFixture) {
				defer wg.Done()

				// Use index to make names unique
				uniqueName := fmt.Sprintf("%s-%d", f.Name, idx)
				_, err := svc.CreateFile(smmodels.DiagramTypePUML, uniqueName, f.Version, f.Content, models.LocationInProgress)
				if err != nil {
					errors <- fmt.Errorf("failed to create %s: %w", uniqueName, err)
				} else {
					successes <- uniqueName
				}
			}(i, fixture)
		}

		wg.Wait()
		close(errors)
		close(successes)

		// Check for errors
		var errorList []error
		for err := range errors {
			errorList = append(errorList, err)
			t.Errorf("Concurrent create error: %v", err)
		}

		// Count successes
		var successList []string
		for name := range successes {
			successList = append(successList, name)
		}

		// Verify all state-machine diagrams were created
		diagrams, err := svc.ListAll(smmodels.DiagramTypePUML, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to list state-machine diagrams: %v", err)
		}

		t.Logf("Created %d state-machine diagrams successfully: %v", len(successList), successList)
		t.Logf("Found %d state-machine diagrams in listing", len(diagrams))
		t.Logf("Had %d errors: %v", len(errorList), errorList)

		if len(diagrams) != len(testFixtures) {
			t.Errorf("Expected %d state-machine diagrams, got %d", len(testFixtures), len(diagrams))
			for _, diag := range diagrams {
				t.Logf("Found state-machine diagram: %s-%s", diag.Name, diag.Version)
			}
		}
	})

	t.Run("Concurrent reads of same state-machine diagram", func(t *testing.T) {
		// First create a state-machine diagram
		fixture := testFixtures[0]
		_, err := svc.CreateFile(smmodels.DiagramTypePUML, "concurrent-read-test", fixture.Version, fixture.Content, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to create state-machine diagram for concurrent read test: %v", err)
		}

		var wg sync.WaitGroup
		errors := make(chan error, 10)
		results := make(chan *models.StateMachineDiagram, 10)

		// Read the same state-machine diagram concurrently
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				diag, err := svc.ReadFile(smmodels.DiagramTypePUML, "concurrent-read-test", fixture.Version, models.LocationInProgress)
				if err != nil {
					errors <- err
				} else {
					results <- diag
				}
			}()
		}

		wg.Wait()
		close(errors)
		close(results)

		// Check for errors
		for err := range errors {
			t.Errorf("Concurrent read error: %v", err)
		}

		// Verify all reads returned the same content
		var firstContent string
		resultCount := 0
		for diag := range results {
			if firstContent == "" {
				firstContent = diag.Content
			} else if diag.Content != firstContent {
				t.Errorf("Concurrent reads returned different content")
			}
			resultCount++
		}

		if resultCount != 10 {
			t.Errorf("Expected 10 successful reads, got %d", resultCount)
		}
	})

	t.Run("Concurrent promote attempts", func(t *testing.T) {
		// Create a state-machine diagram
		fixture := testFixtures[2]
		_, err := svc.CreateFile(smmodels.DiagramTypePUML, "concurrent-promote-test", fixture.Version, fixture.Content, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to create state-machine diagram for concurrent promote test: %v", err)
		}

		var wg sync.WaitGroup
		errors := make(chan error, 5)
		successes := make(chan bool, 5)

		// Try to promote the same state-machine diagram concurrently
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := svc.Promote(smmodels.DiagramTypePUML, "concurrent-promote-test", fixture.Version)
				if err != nil {
					errors <- err
				} else {
					successes <- true
				}
			}()
		}

		wg.Wait()
		close(errors)
		close(successes)

		// Only one promotion should succeed
		successCount := len(successes)
		errorCount := len(errors)

		if successCount != 1 {
			t.Errorf("Expected exactly 1 successful promotion, got %d", successCount)
		}

		if errorCount != 4 {
			t.Errorf("Expected exactly 4 failed promotions, got %d", errorCount)
		}

		// Verify the state-machine diagram is in products
		_, err = svc.ReadFile(smmodels.DiagramTypePUML, "concurrent-promote-test", fixture.Version, models.LocationProducts)
		if err != nil {
			t.Errorf("State-machine diagram should be in products after promotion: %v", err)
		}

		// Verify it's not in in-progress
		_, err = svc.ReadFile(smmodels.DiagramTypePUML, "concurrent-promote-test", fixture.Version, models.LocationInProgress)
		if err == nil {
			t.Errorf("State-machine diagram should not be in in-progress after promotion")
		}
	})

	t.Run("Mixed concurrent operations", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, 20)

		// Mix of different operations running concurrently
		for i := 0; i < 5; i++ {
			// Create operations
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				fixture := testFixtures[idx%len(testFixtures)]
				uniqueName := fmt.Sprintf("mixed-test-%d", idx)
				_, err := svc.CreateFile(smmodels.DiagramTypePUML, uniqueName, fixture.Version, fixture.Content, models.LocationInProgress)
				if err != nil {
					errors <- fmt.Errorf("create error: %w", err)
				}
			}(i)

			// List operations
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := svc.ListAll(smmodels.DiagramTypePUML, models.LocationInProgress)
				if err != nil {
					errors <- fmt.Errorf("list error: %w", err)
				}
			}()

			// Validation operations (on existing state-machine diagrams)
			if i > 0 { // Only after we've created some
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					time.Sleep(100 * time.Millisecond) // Small delay to ensure creation
					testName := fmt.Sprintf("mixed-test-%d", idx-1)
					fixture := testFixtures[(idx-1)%len(testFixtures)]
					_, err := svc.Validate(smmodels.DiagramTypePUML, testName, fixture.Version, models.LocationInProgress)
					if err != nil {
						errors <- fmt.Errorf("validate error: %w", err)
					}
				}(i)
			}
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Errorf("Mixed concurrent operation error: %v", err)
		}
	})
}

// TestReferenceResolutionWorkflow tests reference resolution functionality
func TestReferenceResolutionWorkflow(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	svc := createTestService(tempDir)

	t.Run("Setup referenced state-machine diagrams", func(t *testing.T) {
		// Create and promote the user-auth state-machine diagram that will be referenced
		userAuthFixture := testFixtures[0] // user-auth
		_, err := svc.CreateFile(smmodels.DiagramTypePUML, userAuthFixture.Name, userAuthFixture.Version, userAuthFixture.Content, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to create user-auth state-machine diagram: %v", err)
		}

		err = svc.Promote(smmodels.DiagramTypePUML, userAuthFixture.Name, userAuthFixture.Version)
		if err != nil {
			t.Fatalf("Failed to promote user-auth state-machine diagram: %v", err)
		}

		// Create the main workflow that references user-auth
		mainWorkflowFixture := referencedFixtures[0] // main-workflow
		_, err = svc.CreateFile(smmodels.DiagramTypePUML, mainWorkflowFixture.Name, mainWorkflowFixture.Version, mainWorkflowFixture.Content, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to create main-workflow state-machine diagram: %v", err)
		}
	})

	t.Run("Test reference resolution", func(t *testing.T) {
		// Read the main workflow
		diagram, err := svc.ReadFile(smmodels.DiagramTypePUML, "main-workflow", "1.0.0", models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to read main-workflow: %v", err)
		}

		// Try to resolve references - this will parse them first and then resolve
		err = svc.ResolveReferences(diagram)

		// Verify references were parsed (even if resolution failed)
		if len(diagram.References) == 0 {
			t.Errorf("Expected references to be parsed, but found none")
		} else {
			t.Logf("Found %d references", len(diagram.References))
			for _, ref := range diagram.References {
				t.Logf("Reference: %s-%s (type: %s)", ref.Name, ref.Version, ref.Type.String())
			}
		}

		// Check for the product reference
		foundProductRef := false
		for _, ref := range diagram.References {
			if ref.Type == models.ReferenceTypeProduct && ref.Name == "user-auth" && ref.Version == "1.0.0" {
				foundProductRef = true
			}
		}

		if !foundProductRef {
			t.Errorf("Expected to find product reference to user-auth-1.0.0")
		}

		// The resolution should succeed since we only have a valid product reference
		if err != nil {
			t.Errorf("Reference resolution should succeed, but got error: %v", err)
		} else {
			t.Logf("Reference resolution succeeded as expected")
		}
	})

	t.Run("Test validation with references", func(t *testing.T) {
		// Validate the main workflow - should pass since we only have valid product references
		result, err := svc.Validate(smmodels.DiagramTypePUML, "main-workflow", "1.0.0", models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to validate main-workflow: %v", err)
		}

		// The validation should pass since all references are valid
		if !result.IsValid {
			t.Errorf("Validation should pass with valid product references. Errors: %v", result.Errors)
		}
	})
}

// TestValidationStrictnessLevels tests different validation strictness levels
func TestValidationStrictnessLevels(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	svc := createTestService(tempDir)

	// Create a state-machine diagram with minor issues (warnings but not critical errors)
	contentWithWarnings := "@startuml\n[*] --> idle_state\nidle_state --> active-state : activate\nactive-state --> [*]\n@enduml"

	t.Run("Create state-machine diagram with warnings", func(t *testing.T) {
		_, err := svc.CreateFile(smmodels.DiagramTypePUML, "strictness-test", "1.0.0", contentWithWarnings, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to create state-machine diagram: %v", err)
		}
	})

	t.Run("Validate in-progress strictness", func(t *testing.T) {
		result, err := svc.Validate(smmodels.DiagramTypePUML, "strictness-test", "1.0.0", models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to validate with in-progress strictness: %v", err)
		}

		// In-progress validation should show both errors and warnings
		// The state-machine diagram should still be valid if there are no critical structural errors
		if !result.IsValid && len(result.Errors) > 0 {
			// Check if errors are critical
			hasCriticalErrors := false
			for _, err := range result.Errors {
				if err.Code == "MISSING_START" || err.Code == "MISSING_END" || err.Code == "NO_STATES" {
					hasCriticalErrors = true
					break
				}
			}
			if hasCriticalErrors {
				t.Errorf("State-machine diagram has critical structural errors: %v", result.Errors)
			}
		}
	})

	t.Run("Promote and validate products strictness", func(t *testing.T) {
		// Promote the state-machine diagram
		err := svc.Promote(smmodels.DiagramTypePUML, "strictness-test", "1.0.0")
		if err != nil {
			t.Fatalf("Failed to promote state-machine diagram: %v", err)
		}

		// Validate with products strictness
		result, err := svc.Validate(smmodels.DiagramTypePUML, "strictness-test", "1.0.0", models.LocationProducts)
		if err != nil {
			t.Fatalf("Failed to validate with products strictness: %v", err)
		}

		// Products validation should be more lenient - non-critical errors become warnings
		if !result.IsValid {
			t.Errorf("State-machine diagram should be valid in products with lenient strictness. Errors: %v", result.Errors)
		}
	})
}

// TestFileSystemEdgeCases tests file system related edge cases
func TestFileSystemEdgeCases(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	svc := createTestService(tempDir)

	t.Run("Large content handling", func(t *testing.T) {
		// Create content that's close to but under the limit
		largeContent := "@startuml\n"
		for i := 0; i < 1000; i++ {
			largeContent += fmt.Sprintf("State%d --> State%d : transition%d\n", i, i+1, i)
		}
		largeContent += "@enduml"

		_, err := svc.CreateFile(smmodels.DiagramTypePUML, "large-content-test", "1.0.0", largeContent, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to create state-machine diagram with large content: %v", err)
		}

		// Verify we can read it back
		diagram, err := svc.ReadFile(smmodels.DiagramTypePUML, "large-content-test", "1.0.0", models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to read large content state-machine diagram: %v", err)
		}

		if diagram.Content != largeContent {
			t.Errorf("Large content was not preserved correctly")
		}
	})

	t.Run("Special characters in content", func(t *testing.T) {
		specialContent := "@startuml\n[*] --> \"State with spaces\"\n\"State with spaces\" --> [*] : \"transition with spaces\"\n' Comment with special chars: !@#$%^&*()\n@enduml"

		_, err := svc.CreateFile(smmodels.DiagramTypePUML, "special-chars-test", "1.0.0", specialContent, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to create state-machine diagram with special characters: %v", err)
		}

		// Verify we can read it back correctly
		diagram, err := svc.ReadFile(smmodels.DiagramTypePUML, "special-chars-test", "1.0.0", models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to read special characters state-machine diagram: %v", err)
		}

		if diagram.Content != specialContent {
			t.Errorf("Special characters were not preserved correctly")
		}
	})

	t.Run("Directory structure verification", func(t *testing.T) {
		fixture := testFixtures[0]
		_, err := svc.CreateFile(smmodels.DiagramTypePUML, "dir-structure-test", fixture.Version, fixture.Content, models.LocationInProgress)
		if err != nil {
			t.Fatalf("Failed to create state-machine diagram: %v", err)
		}

		// Verify the flattened directory structure was created correctly
		expectedDir := filepath.Join(tempDir, "in-progress", "puml")
		if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
			t.Errorf("Expected directory %s was not created", expectedDir)
		}

		expectedFile := filepath.Join(expectedDir, "dir-structure-test-"+fixture.Version+".puml")
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not created", expectedFile)
		}

		// Promote and verify products directory structure
		err = svc.Promote(smmodels.DiagramTypePUML, "dir-structure-test", fixture.Version)
		if err != nil {
			t.Fatalf("Failed to promote state-machine diagram: %v", err)
		}

		expectedProductsDir := filepath.Join(tempDir, "products", "puml")
		if _, err := os.Stat(expectedProductsDir); os.IsNotExist(err) {
			t.Errorf("Expected products directory %s was not created", expectedProductsDir)
		}

		expectedProductsFile := filepath.Join(expectedProductsDir, "dir-structure-test-"+fixture.Version+".puml")
		if _, err := os.Stat(expectedProductsFile); os.IsNotExist(err) {
			t.Errorf("Expected products file %s was not created", expectedProductsFile)
		}

		// Verify in-progress file was removed
		if _, err := os.Stat(expectedFile); !os.IsNotExist(err) {
			t.Errorf("In-progress file %s should have been removed after promotion", expectedFile)
		}
	})
}

// TestVersionHandling tests version-related functionality
func TestVersionHandling(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	svc := createTestService(tempDir)

	versionTests := []struct {
		name    string
		version string
		valid   bool
	}{
		{"semantic-version", "1.0.0", true},
		{"semantic-with-patch", "2.1.5", true},
		{"pre-release", "1.0.0-beta", true},
		{"pre-release-with-number", "1.0.0-beta.1", true},
		{"pre-release-with-text", "1.0.0-alpha", true},
		{"complex-pre-release", "2.0.0-rc.1.2", true},
	}

	for _, tt := range versionTests {
		t.Run(fmt.Sprintf("Version %s", tt.name), func(t *testing.T) {
			testName := fmt.Sprintf("version-test-%s", tt.name)

			_, err := svc.CreateFile(smmodels.DiagramTypePUML, testName, tt.version, testFixtures[0].Content, models.LocationInProgress)

			if tt.valid && err != nil {
				t.Errorf("Expected version %s to be valid, but got error: %v", tt.version, err)
			} else if !tt.valid && err == nil {
				t.Errorf("Expected version %s to be invalid, but creation succeeded", tt.version)
			}

			if tt.valid && err == nil {
				// Verify we can read it back
				diag, err := svc.ReadFile(smmodels.DiagramTypePUML, testName, tt.version, models.LocationInProgress)
				if err != nil {
					t.Errorf("Failed to read back state-machine diagram with version %s: %v", tt.version, err)
				} else if diag.Version != tt.version {
					t.Errorf("Version mismatch: expected %s, got %s", tt.version, diag.Version)
				}
			}
		})
	}
}
