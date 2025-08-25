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

// CreateTestStateMachine creates a test state-machine diagram
func (th *TestHelper) CreateTestStateMachine(name, version string, location models.Location) *models.StateMachineDiagram {
	content := `@startuml
[*] --> Idle
Idle --> Active : start
Active --> Idle : stop
@enduml`

	return &models.StateMachineDiagram{
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
		name        string
		diagram     *models.StateMachineDiagram
		expectError bool
		errorType   models.ErrorType
	}{
		{
			name:        "valid in-progress state-machine diagram",
			diagram:     th.CreateTestStateMachine("test-diag", "1.0.0", models.LocationInProgress),
			expectError: false,
		},
		{
			name:        "valid products state-machine diagram",
			diagram:     th.CreateTestStateMachine("test-diag", "1.0.0", models.LocationProducts),
			expectError: false,
		},
		{
			name:        "nil state-machine diagram",
			diagram:     nil,
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
		{
			name: "empty name",
			diagram: &models.StateMachineDiagram{
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
			diagram: &models.StateMachineDiagram{
				Name:     "test-diag",
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
			err := th.repo.WriteStateMachine(tt.diagram)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}

				if diagErr, ok := err.(*models.StateMachineError); ok {
					if diagErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, diagErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				// Verify file was created
				filePath := th.repo.pathManager.GetStateMachineFilePathWithFileType(
					tt.diagram.Name,
					tt.diagram.Version,
					tt.diagram.Location,
					tt.diagram.FileType,
				)

				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Error("Expected file to be created")
				}

				// Verify content
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Errorf("Failed to read created file: %v", err)
				}

				if string(content) != tt.diagram.Content {
					t.Error("File content doesn't match expected content")
				}
			}
		})
	}
}

func TestFileSystemRepository_ReadStateMachine(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	// Create a test state-machine diagram first
	testSM := th.CreateTestStateMachine("test-read", "1.0.0", models.LocationInProgress)
	err := th.repo.WriteStateMachine(testSM)
	if err != nil {
		t.Fatalf("Failed to create test state-machine diagram: %v", err)
	}

	tests := []struct {
		name        string
		diagName    string
		version     string
		location    models.Location
		expectError bool
		errorType   models.ErrorType
	}{
		{
			name:        "existing state-machine diagram",
			diagName:    "test-read",
			version:     "1.0.0",
			location:    models.LocationInProgress,
			expectError: false,
		},
		{
			name:        "non-existent state-machine diagram",
			diagName:    "non-existent",
			version:     "1.0.0",
			location:    models.LocationInProgress,
			expectError: true,
			errorType:   models.ErrorTypeFileNotFound,
		},
		{
			name:        "empty name",
			diagName:    "",
			version:     "1.0.0",
			location:    models.LocationInProgress,
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
		{
			name:        "missing version for non-nested",
			diagName:    "test-read",
			version:     "",
			location:    models.LocationInProgress,
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag, err := th.repo.ReadStateMachine(models.FileTypePUML, tt.diagName, tt.version, tt.location)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}

				if diagErr, ok := err.(*models.StateMachineError); ok {
					if diagErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, diagErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				if diag == nil {
					t.Error("Expected state-machine diagram to be returned")
					return
				}

				if diag.Name != tt.diagName {
					t.Errorf("Expected name %s, got %s", tt.diagName, diag.Name)
				}

				if diag.Version != tt.version {
					t.Errorf("Expected version %s, got %s", tt.version, diag.Version)
				}

				if diag.Location != tt.location {
					t.Errorf("Expected location %v, got %v", tt.location, diag.Location)
				}

				if diag.Content != testSM.Content {
					t.Error("Content doesn't match expected content")
				}
			}
		})
	}
}

func TestFileSystemRepository_Exists(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	// Create a test state-machine diagram first
	testSM := th.CreateTestStateMachine("test-exists", "1.0.0", models.LocationInProgress)
	err := th.repo.WriteStateMachine(testSM)
	if err != nil {
		t.Fatalf("Failed to create test state-machine diagram: %v", err)
	}

	tests := []struct {
		name         string
		diagName     string
		version      string
		location     models.Location
		expectExists bool
		expectError  bool
		errorType    models.ErrorType
	}{
		{
			name:         "existing state-machine diagram",
			diagName:     "test-exists",
			version:      "1.0.0",
			location:     models.LocationInProgress,
			expectExists: true,
			expectError:  false,
		},
		{
			name:         "non-existent state-machine diagram",
			diagName:     "non-existent",
			version:      "1.0.0",
			location:     models.LocationInProgress,
			expectExists: false,
			expectError:  false,
		},
		{
			name:        "empty name",
			diagName:    "",
			version:     "1.0.0",
			location:    models.LocationInProgress,
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := th.repo.Exists(models.FileTypePUML, tt.diagName, tt.version, tt.location)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}

				if diagErr, ok := err.(*models.StateMachineError); ok {
					if diagErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, diagErr.Type)
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

				if diagErr, ok := err.(*models.StateMachineError); ok {
					if diagErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, diagErr.Type)
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

				if diagErr, ok := err.(*models.StateMachineError); ok {
					if diagErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, diagErr.Type)
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

	// Create a test state-machine diagram in in-progress
	testSM := th.CreateTestStateMachine("test-move", "1.0.0", models.LocationInProgress)
	err := th.repo.WriteStateMachine(testSM)
	if err != nil {
		t.Fatalf("Failed to create test state-machine diagram: %v", err)
	}

	tests := []struct {
		name        string
		diagName    string
		version     string
		from        models.Location
		to          models.Location
		expectError bool
		errorType   models.ErrorType
	}{
		{
			name:        "valid move from in-progress to products",
			diagName:    "test-move",
			version:     "1.0.0",
			from:        models.LocationInProgress,
			to:          models.LocationProducts,
			expectError: false,
		},
		{
			name:        "empty name",
			diagName:    "",
			version:     "1.0.0",
			from:        models.LocationInProgress,
			to:          models.LocationProducts,
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
		{
			name:        "empty version",
			diagName:    "test-move",
			version:     "",
			from:        models.LocationInProgress,
			to:          models.LocationProducts,
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
		{
			name:        "same source and destination",
			diagName:    "test-move",
			version:     "1.0.0",
			from:        models.LocationInProgress,
			to:          models.LocationInProgress,
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
		{
			name:        "non-existent source",
			diagName:    "non-existent",
			version:     "1.0.0",
			from:        models.LocationInProgress,
			to:          models.LocationProducts,
			expectError: true,
			errorType:   models.ErrorTypeFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := th.repo.MoveStateMachine(models.FileTypePUML, tt.diagName, tt.version, tt.from, tt.to)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}

				if diagErr, ok := err.(*models.StateMachineError); ok {
					if diagErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, diagErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				// Verify source no longer exists
				sourceExists, err := th.repo.Exists(models.FileTypePUML, tt.diagName, tt.version, tt.from)
				if err != nil {
					t.Errorf("Error checking source existence: %v", err)
				}
				if sourceExists {
					t.Error("Expected source to be moved")
				}

				// Verify destination exists
				destExists, err := th.repo.Exists(models.FileTypePUML, tt.diagName, tt.version, tt.to)
				if err != nil {
					t.Errorf("Error checking destination existence: %v", err)
				}
				if !destExists {
					t.Error("Expected destination to exist after move")
				}

				// Verify content is preserved
				diagram, err := th.repo.ReadStateMachine(models.FileTypePUML, tt.diagName, tt.version, tt.to)
				if err != nil {
					t.Errorf("Error reading moved state-machine diagram: %v", err)
				}
				if diagram.Content != testSM.Content {
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
		diagName    string
		version     string
		location    models.Location
		expectError bool
		errorType   models.ErrorType
	}{
		{
			name:        "delete existing state-machine diagram",
			setupSM:     true,
			diagName:    "test-delete",
			version:     "1.0.0",
			location:    models.LocationInProgress,
			expectError: false,
		},
		{
			name:        "delete non-existent state-machine diagram",
			setupSM:     false,
			diagName:    "non-existent",
			version:     "1.0.0",
			location:    models.LocationInProgress,
			expectError: true,
			errorType:   models.ErrorTypeFileNotFound,
		},
		{
			name:        "empty name",
			setupSM:     false,
			diagName:    "",
			version:     "1.0.0",
			location:    models.LocationInProgress,
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
		{
			name:        "missing version for non-nested",
			setupSM:     false,
			diagName:    "test-delete",
			version:     "",
			location:    models.LocationInProgress,
			expectError: true,
			errorType:   models.ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test state-machine diagram if needed
			if tt.setupSM {
				testSM := th.CreateTestStateMachine(tt.diagName, tt.version, tt.location)
				err := th.repo.WriteStateMachine(testSM)
				if err != nil {
					t.Fatalf("Failed to create test state-machine diagram: %v", err)
				}
			}

			err := th.repo.DeleteStateMachine(models.FileTypePUML, tt.diagName, tt.version, tt.location)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}

				if diagErr, ok := err.(*models.StateMachineError); ok {
					if diagErr.Type != tt.errorType {
						t.Errorf("Expected error type %v, got %v", tt.errorType, diagErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				// Verify state-machine diagram no longer exists
				exists, err := th.repo.Exists(models.FileTypePUML, tt.diagName, tt.version, tt.location)
				if err != nil {
					t.Errorf("Error checking existence after delete: %v", err)
				}
				if exists {
					t.Error("Expected state-machine diagram to be deleted")
				}
			}
		})
	}
}

func TestFileSystemRepository_ListStateMachines(t *testing.T) {
	th := NewTestHelper(t)
	defer th.Cleanup()

	// Create multiple test state-machine diagrams
	testDiagrams := []*models.StateMachineDiagram{
		th.CreateTestStateMachine("sm1", "1.0.0", models.LocationInProgress),
		th.CreateTestStateMachine("sm2", "1.1.0", models.LocationInProgress),
		th.CreateTestStateMachine("sm3", "2.0.0", models.LocationProducts),
	}

	// Write the state-machine diagrams
	for _, diag := range testDiagrams {
		err := th.repo.WriteStateMachine(diag)
		if err != nil {
			t.Fatalf("Failed to create test state-machine diagram %s: %v", diag.Name, err)
		}
	}

	tests := []struct {
		name          string
		location      models.Location
		expectedCount int
		expectError   bool
	}{
		{
			name:          "list in-progress state-machine diagrams",
			location:      models.LocationInProgress,
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "list products state-machine diagrams",
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
			diagrams, err := th.repo.ListStateMachines(models.FileTypePUML, tt.location)

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

				if len(diagrams) != tt.expectedCount {
					t.Errorf("Expected %d state-machine diagrams, got %d", tt.expectedCount, len(diagrams))
				}

				// Verify all returned state-machine diagrams have the correct location
				for _, diag := range diagrams {
					if diag.Location != tt.location {
						t.Errorf("Expected location %v, got %v for state-machine diagram %s", tt.location, diag.Location, diag.Name)
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

	// 1. Write state-machine diagram
	err := th.repo.WriteStateMachine(testSM)
	if err != nil {
		t.Fatalf("Failed to write state-machine diagram: %v", err)
	}

	// 2. Verify it exists
	exists, err := th.repo.Exists(models.FileTypePUML, testSM.Name, testSM.Version, testSM.Location)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Fatal("State-machine diagram should exist after writing")
	}

	// 3. Read it back
	readSM, err := th.repo.ReadStateMachine(models.FileTypePUML, testSM.Name, testSM.Version, testSM.Location)
	if err != nil {
		t.Fatalf("Failed to read state-machine diagram: %v", err)
	}
	if readSM.Content != testSM.Content {
		t.Error("Read content doesn't match written content")
	}

	// 4. List state-machine diagrams
	diagrams, err := th.repo.ListStateMachines(models.FileTypePUML, models.LocationInProgress)
	if err != nil {
		t.Fatalf("Failed to list state-machine diagrams: %v", err)
	}
	if len(diagrams) != 1 {
		t.Errorf("Expected 1 state-machine diagram, got %d", len(diagrams))
	}

	// 5. Move to products
	err = th.repo.MoveStateMachine(models.FileTypePUML, testSM.Name, testSM.Version, models.LocationInProgress, models.LocationProducts)
	if err != nil {
		t.Fatalf("Failed to move state-machine diagram: %v", err)
	}

	// 6. Verify it's in products now
	exists, err = th.repo.Exists(models.FileTypePUML, testSM.Name, testSM.Version, models.LocationProducts)
	if err != nil {
		t.Fatalf("Failed to check existence in products: %v", err)
	}
	if !exists {
		t.Error("State-machine diagram should exist in products after move")
	}

	// 7. Verify it's no longer in in-progress
	exists, err = th.repo.Exists(models.FileTypePUML, testSM.Name, testSM.Version, models.LocationInProgress)
	if err != nil {
		t.Fatalf("Failed to check existence in in-progress: %v", err)
	}
	if exists {
		t.Error("State-machine diagram should not exist in in-progress after move")
	}

	// 8. Delete from products
	err = th.repo.DeleteStateMachine(models.FileTypePUML, testSM.Name, testSM.Version, models.LocationProducts)
	if err != nil {
		t.Fatalf("Failed to delete state-machine diagram: %v", err)
	}

	// 9. Verify it's gone
	exists, err = th.repo.Exists(models.FileTypePUML, testSM.Name, testSM.Version, models.LocationProducts)
	if err != nil {
		t.Fatalf("Failed to check existence after delete: %v", err)
	}
	if exists {
		t.Error("State-machine diagram should not exist after delete")
	}
}
