package repository

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/models"
)

// TestHelper provides utilities for testing
type TestHelper struct {
	tempDir string
	repo    *FileSystemRepository
}

// NewTestHelper creates a new test helper with a temporary directory
func NewTestHelper(t *testing.T) *TestHelper {
	tempDir, err := os.MkdirTemp("", "go-uml-statemachine-parsers-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	config := &models.Config{
		RootDirectory: tempDir,
		MaxFileSize:   1024 * 1024, // 1MB
	}

	repo := NewFileSystemRepository(config)

	return &TestHelper{
		tempDir: tempDir,
		repo:    repo,
	}
}

// Cleanup removes the temporary directory
func (th *TestHelper) Cleanup() {
	os.RemoveAll(th.tempDir)
}

// CreateTestStateMachine creates a test state machine
func (th *TestHelper) CreateTestStateMachine(name, version string, location models.Location) *models.StateMachine {
	content := `@startuml
[*] --> Idle
Idle --> Active : start
Active --> Idle : stop
@enduml`

	return &models.StateMachine{
		Name:     name,
		Version:  version,
		Content:  content,
		Location: location,
		Metadata: models.Metadata{
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
		},
	}
}

func TestNewFileSystemRepository(t *testing.T) {
	tests := []struct {
		name   string
		config *models.Config
	}{
		{
			name:   "with config",
			config: &models.Config{RootDirectory: "/test"},
		},
		{
			name:   "with nil config",
			config: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewFileSystemRepository(tt.config)
			if repo == nil {
				t.Error("Expected repository to be created")
			}
			if repo.pathManager == nil {
				t.Error("Expected path manager to be initialized")
			}
			if repo.config == nil {
				t.Error("Expected config to be initialized")
			}
		})
	}
}

func TestFileSystemRepository_WriteStateMachine(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	tests := []struct {
		name         string
		stateMachine *models.StateMachine
		expectError  bool
		errorType    models.ErrorType
	}{
		{
			name:         "valid in-progress state machine",
			stateMachine: th.CreateTestStateMachine("test-sm", "1.0.0", models.LocationInProgress),
			expectError:  false,
		},
		{
			name:         "valid products state machine",
			stateMachine: th.CreateTestStateMachine("test-sm", "1.0.0", models.LocationProducts),
			expectError:  false,
		},
		{
			name:         "nil state machine",
			stateMachine: nil,
			expectError:  true,
			errorType:    models.ErrorTypeValidation,
		},
		{
			name: "empty name",
			stateMachine: &models.StateMachine{
				Name:     "",
				Version:  "1.0.0",
				Content:  "test content",
				Location: models.LocationInProgress,
			},
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
		{
			name: "missing version for non-nested",
			stateMachine: &models.StateMachine{
				Name:     "test-sm",
				Version:  "",
				Content:  "test content",
				Location: models.LocationInProgress,
			},
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := th.repo.WriteStateMachine(tt.stateMachine)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}

				if smErr, ok := err.(*models.StateMachineError); ok {
					if smErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, smErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				// Verify file was created
				filePath := th.repo.pathManager.GetStateMachineFilePath(
					tt.stateMachine.Name,
					tt.stateMachine.Version,
					tt.stateMachine.Location,
				)

				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Error("Expected file to be created")
				}

				// Verify content
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Errorf("Failed to read created file: %v", err)
				}

				if string(content) != tt.stateMachine.Content {
					t.Error("File content doesn't match expected content")
				}
			}
		})
	}
}

func TestFileSystemRepository_ReadStateMachine(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	// Create a test state machine first
	testSM := th.CreateTestStateMachine("test-read", "1.0.0", models.LocationInProgress)
	err := th.repo.WriteStateMachine(testSM)
	if err != nil {
		t.Fatalf("Failed to create test state machine: %v", err)
	}

	tests := []struct {
		name        string
		smName      string
		version     string
		location    models.Location
		expectError bool
		errorType   models.ErrorType
	}{
		{
			name:        "existing state machine",
			smName:      "test-read",
			version:     "1.0.0",
			location:    models.LocationInProgress,
			expectError: false,
		},
		{
			name:        "non-existent state machine",
			smName:      "non-existent",
			version:     "1.0.0",
			location:    models.LocationInProgress,
			expectError: true,
			errorType:   models.ErrorTypeFileNotFound,
		},
		{
			name:        "empty name",
			smName:      "",
			version:     "1.0.0",
			location:    models.LocationInProgress,
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
		{
			name:        "missing version for non-nested",
			smName:      "test-read",
			version:     "",
			location:    models.LocationInProgress,
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm, err := th.repo.ReadStateMachine(tt.smName, tt.version, tt.location)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}

				if smErr, ok := err.(*models.StateMachineError); ok {
					if smErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, smErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				if sm == nil {
					t.Error("Expected state machine to be returned")
					return
				}

				if sm.Name != tt.smName {
					t.Errorf("Expected name %s, got %s", tt.smName, sm.Name)
				}

				if sm.Version != tt.version {
					t.Errorf("Expected version %s, got %s", tt.version, sm.Version)
				}

				if sm.Location != tt.location {
					t.Errorf("Expected location %v, got %v", tt.location, sm.Location)
				}

				if sm.Content != testSM.Content {
					t.Error("Content doesn't match expected content")
				}
			}
		})
	}
}

func TestFileSystemRepository_Exists(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	// Create a test state machine first
	testSM := th.CreateTestStateMachine("test-exists", "1.0.0", models.LocationInProgress)
	err := th.repo.WriteStateMachine(testSM)
	if err != nil {
		t.Fatalf("Failed to create test state machine: %v", err)
	}

	tests := []struct {
		name         string
		smName       string
		version      string
		location     models.Location
		expectExists bool
		expectError  bool
		errorType    models.ErrorType
	}{
		{
			name:         "existing state machine",
			smName:       "test-exists",
			version:      "1.0.0",
			location:     models.LocationInProgress,
			expectExists: true,
			expectError:  false,
		},
		{
			name:         "non-existent state machine",
			smName:       "non-existent",
			version:      "1.0.0",
			location:     models.LocationInProgress,
			expectExists: false,
			expectError:  false,
		},
		{
			name:        "empty name",
			smName:      "",
			version:     "1.0.0",
			location:    models.LocationInProgress,
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := th.repo.Exists(tt.smName, tt.version, tt.location)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}

				if smErr, ok := err.(*models.StateMachineError); ok {
					if smErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, smErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				if exists != tt.expectExists {
					t.Errorf("Expected exists %v, got %v", tt.expectExists, exists)
				}
			}
		})
	}
}
func TestFileSystemRepository_CreateDirectory(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	tests := []struct {
		name        string
		path        string
		expectError bool
		errorType   models.ErrorType
	}{
		{
			name:        "valid path",
			path:        filepath.Join(th.tempDir, "test-dir"),
			expectError: false,
		},
		{
			name:        "nested path",
			path:        filepath.Join(th.tempDir, "test", "nested", "dir"),
			expectError: false,
		},
		{
			name:        "existing directory",
			path:        th.tempDir, // Already exists
			expectError: false,
		},
		{
			name:        "path with traversal",
			path:        filepath.Join(th.tempDir, "..", "malicious"),
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := th.repo.CreateDirectory(tt.path)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}

				if smErr, ok := err.(*models.StateMachineError); ok {
					if smErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, smErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				// Verify directory was created
				if info, err := os.Stat(tt.path); err != nil || !info.IsDir() {
					t.Error("Expected directory to be created")
				}
			}
		})
	}
}

func TestFileSystemRepository_DirectoryExists(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	// Create a test directory
	testDir := filepath.Join(th.tempDir, "test-exists-dir")
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	tests := []struct {
		name         string
		path         string
		expectExists bool
		expectError  bool
		errorType    models.ErrorType
	}{
		{
			name:         "existing directory",
			path:         testDir,
			expectExists: true,
			expectError:  false,
		},
		{
			name:         "non-existent directory",
			path:         filepath.Join(th.tempDir, "non-existent"),
			expectExists: false,
			expectError:  false,
		},
		{
			name:        "path with traversal",
			path:        filepath.Join(th.tempDir, "..", "malicious"),
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := th.repo.DirectoryExists(tt.path)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}

				if smErr, ok := err.(*models.StateMachineError); ok {
					if smErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, smErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				if exists != tt.expectExists {
					t.Errorf("Expected exists %v, got %v", tt.expectExists, exists)
				}
			}
		})
	}
}

func TestFileSystemRepository_MoveStateMachine(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	// Create a test state machine in in-progress
	testSM := th.CreateTestStateMachine("test-move", "1.0.0", models.LocationInProgress)
	err := th.repo.WriteStateMachine(testSM)
	if err != nil {
		t.Fatalf("Failed to create test state machine: %v", err)
	}

	tests := []struct {
		name        string
		smName      string
		version     string
		from        models.Location
		to          models.Location
		expectError bool
		errorType   models.ErrorType
	}{
		{
			name:        "valid move from in-progress to products",
			smName:      "test-move",
			version:     "1.0.0",
			from:        models.LocationInProgress,
			to:          models.LocationProducts,
			expectError: false,
		},
		{
			name:        "empty name",
			smName:      "",
			version:     "1.0.0",
			from:        models.LocationInProgress,
			to:          models.LocationProducts,
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
		{
			name:        "empty version",
			smName:      "test-move",
			version:     "",
			from:        models.LocationInProgress,
			to:          models.LocationProducts,
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
		{
			name:        "same source and destination",
			smName:      "test-move",
			version:     "1.0.0",
			from:        models.LocationInProgress,
			to:          models.LocationInProgress,
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
		{
			name:        "non-existent source",
			smName:      "non-existent",
			version:     "1.0.0",
			from:        models.LocationInProgress,
			to:          models.LocationProducts,
			expectError: true,
			errorType:   models.ErrorTypeFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := th.repo.MoveStateMachine(tt.smName, tt.version, tt.from, tt.to)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}

				if smErr, ok := err.(*models.StateMachineError); ok {
					if smErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, smErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				// Verify source no longer exists
				sourceExists, err := th.repo.Exists(tt.smName, tt.version, tt.from)
				if err != nil {
					t.Errorf("Error checking source existence: %v", err)
				}
				if sourceExists {
					t.Error("Expected source to be moved")
				}

				// Verify destination exists
				destExists, err := th.repo.Exists(tt.smName, tt.version, tt.to)
				if err != nil {
					t.Errorf("Error checking destination existence: %v", err)
				}
				if !destExists {
					t.Error("Expected destination to exist after move")
				}

				// Verify content is preserved
				sm, err := th.repo.ReadStateMachine(tt.smName, tt.version, tt.to)
				if err != nil {
					t.Errorf("Error reading moved state machine: %v", err)
				}
				if sm.Content != testSM.Content {
					t.Error("Content was not preserved during move")
				}
			}
		})
	}
}

func TestFileSystemRepository_DeleteStateMachine(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	tests := []struct {
		name        string
		setupSM     bool
		smName      string
		version     string
		location    models.Location
		expectError bool
		errorType   models.ErrorType
	}{
		{
			name:        "delete existing state machine",
			setupSM:     true,
			smName:      "test-delete",
			version:     "1.0.0",
			location:    models.LocationInProgress,
			expectError: false,
		},
		{
			name:        "delete non-existent state machine",
			setupSM:     false,
			smName:      "non-existent",
			version:     "1.0.0",
			location:    models.LocationInProgress,
			expectError: true,
			errorType:   models.ErrorTypeFileNotFound,
		},
		{
			name:        "empty name",
			setupSM:     false,
			smName:      "",
			version:     "1.0.0",
			location:    models.LocationInProgress,
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
		{
			name:        "missing version for non-nested",
			setupSM:     false,
			smName:      "test-delete",
			version:     "",
			location:    models.LocationInProgress,
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test state machine if needed
			if tt.setupSM {
				testSM := th.CreateTestStateMachine(tt.smName, tt.version, tt.location)
				err := th.repo.WriteStateMachine(testSM)
				if err != nil {
					t.Fatalf("Failed to create test state machine: %v", err)
				}
			}

			err := th.repo.DeleteStateMachine(tt.smName, tt.version, tt.location)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}

				if smErr, ok := err.(*models.StateMachineError); ok {
					if smErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, smErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				// Verify state machine no longer exists
				exists, err := th.repo.Exists(tt.smName, tt.version, tt.location)
				if err != nil {
					t.Errorf("Error checking existence after delete: %v", err)
				}
				if exists {
					t.Error("Expected state machine to be deleted")
				}
			}
		})
	}
}

func TestFileSystemRepository_ListStateMachines(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	// Create multiple test state machines
	testSMs := []*models.StateMachine{
		th.CreateTestStateMachine("sm1", "1.0.0", models.LocationInProgress),
		th.CreateTestStateMachine("sm2", "1.1.0", models.LocationInProgress),
		th.CreateTestStateMachine("sm3", "2.0.0", models.LocationProducts),
	}

	// Write the state machines
	for _, sm := range testSMs {
		err := th.repo.WriteStateMachine(sm)
		if err != nil {
			t.Fatalf("Failed to create test state machine %s: %v", sm.Name, err)
		}
	}

	tests := []struct {
		name          string
		location      models.Location
		expectedCount int
		expectError   bool
	}{
		{
			name:          "list in-progress state machines",
			location:      models.LocationInProgress,
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "list products state machines",
			location:      models.LocationProducts,
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "list from non-existent location",
			location:      models.LocationNested, // No nested SMs created
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sms, err := th.repo.ListStateMachines(tt.location)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				if len(sms) != tt.expectedCount {
					t.Errorf("Expected %d state machines, got %d", tt.expectedCount, len(sms))
				}

				// Verify all returned state machines have the correct location
				for _, sm := range sms {
					if sm.Location != tt.location {
						t.Errorf("Expected location %v, got %v for state machine %s", tt.location, sm.Location, sm.Name)
					}
				}
			}
		})
	}
}

func TestFileSystemRepository_Integration(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	// Test a complete workflow: create -> read -> move -> delete
	testSM := th.CreateTestStateMachine("integration-test", "1.0.0", models.LocationInProgress)

	// 1. Write state machine
	err := th.repo.WriteStateMachine(testSM)
	if err != nil {
		t.Fatalf("Failed to write state machine: %v", err)
	}

	// 2. Verify it exists
	exists, err := th.repo.Exists(testSM.Name, testSM.Version, testSM.Location)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Fatal("State machine should exist after writing")
	}

	// 3. Read it back
	readSM, err := th.repo.ReadStateMachine(testSM.Name, testSM.Version, testSM.Location)
	if err != nil {
		t.Fatalf("Failed to read state machine: %v", err)
	}
	if readSM.Content != testSM.Content {
		t.Error("Read content doesn't match written content")
	}

	// 4. List state machines
	sms, err := th.repo.ListStateMachines(models.LocationInProgress)
	if err != nil {
		t.Fatalf("Failed to list state machines: %v", err)
	}
	if len(sms) != 1 {
		t.Errorf("Expected 1 state machine, got %d", len(sms))
	}

	// 5. Move to products
	err = th.repo.MoveStateMachine(testSM.Name, testSM.Version, models.LocationInProgress, models.LocationProducts)
	if err != nil {
		t.Fatalf("Failed to move state machine: %v", err)
	}

	// 6. Verify it's in products now
	exists, err = th.repo.Exists(testSM.Name, testSM.Version, models.LocationProducts)
	if err != nil {
		t.Fatalf("Failed to check existence in products: %v", err)
	}
	if !exists {
		t.Error("State machine should exist in products after move")
	}

	// 7. Verify it's no longer in in-progress
	exists, err = th.repo.Exists(testSM.Name, testSM.Version, models.LocationInProgress)
	if err != nil {
		t.Fatalf("Failed to check existence in in-progress: %v", err)
	}
	if exists {
		t.Error("State machine should not exist in in-progress after move")
	}

	// 8. Delete from products
	err = th.repo.DeleteStateMachine(testSM.Name, testSM.Version, models.LocationProducts)
	if err != nil {
		t.Fatalf("Failed to delete state machine: %v", err)
	}

	// 9. Verify it's gone
	exists, err = th.repo.Exists(testSM.Name, testSM.Version, models.LocationProducts)
	if err != nil {
		t.Fatalf("Failed to check existence after delete: %v", err)
	}
	if exists {
		t.Error("State machine should not exist after delete")
	}
}
