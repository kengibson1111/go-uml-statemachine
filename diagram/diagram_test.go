package diagram

import (
	"os"
	"testing"
)

func TestNewService(t *testing.T) {
	svc, err := NewService()
	if err != nil {
		t.Fatalf("NewService() failed: %v", err)
	}
	if svc == nil {
		t.Fatal("NewService() returned nil service")
	}
}

func TestNewServiceWithConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name: "custom config",
			config: &Config{
				RootDirectory:      ".test-statemachine",
				EnableDebugLogging: true,
				MaxFileSize:        2 * 1024 * 1024,
				BackupEnabled:      true,
			},
		},
		{
			name:   "default config",
			config: DefaultConfig(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewServiceWithConfig(tt.config)
			if err != nil {
				t.Fatalf("NewServiceWithConfig() failed: %v", err)
			}
			if svc == nil {
				t.Fatal("NewServiceWithConfig() returned nil service")
			}
		})
	}
}

func TestNewServiceFromEnv(t *testing.T) {
	// Save original environment
	originalRootDir := os.Getenv("GO_UML_ROOT_DIRECTORY")
	originalDebugLogging := os.Getenv("GO_UML_DEBUG_LOGGING")
	originalMaxFileSize := os.Getenv("GO_UML_MAX_FILE_SIZE")

	// Clean up after test
	defer func() {
		if originalRootDir != "" {
			os.Setenv("GO_UML_ROOT_DIRECTORY", originalRootDir)
		} else {
			os.Unsetenv("GO_UML_ROOT_DIRECTORY")
		}
		if originalDebugLogging != "" {
			os.Setenv("GO_UML_DEBUG_LOGGING", originalDebugLogging)
		} else {
			os.Unsetenv("GO_UML_DEBUG_LOGGING")
		}
		if originalMaxFileSize != "" {
			os.Setenv("GO_UML_MAX_FILE_SIZE", originalMaxFileSize)
		} else {
			os.Unsetenv("GO_UML_MAX_FILE_SIZE")
		}
	}()

	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name:    "no environment variables",
			envVars: map[string]string{},
			wantErr: false,
		},
		{
			name: "custom environment variables",
			envVars: map[string]string{
				"GO_UML_ROOT_DIRECTORY": ".env-test-statemachine",
				"GO_UML_DEBUG_LOGGING":  "true",
				"GO_UML_MAX_FILE_SIZE":  "5242880",
			},
			wantErr: false,
		},
		{
			name: "invalid max file size",
			envVars: map[string]string{
				"GO_UML_MAX_FILE_SIZE": "invalid",
			},
			wantErr: false, // Should use default value, not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			svc, err := NewServiceFromEnv()
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewServiceFromEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && svc == nil {
				t.Fatal("NewServiceFromEnv() returned nil service")
			}

			// Clean up environment variables for this test
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Verify default values
	if config.RootDirectory != ".go-uml-statemachine-parsers" {
		t.Errorf("DefaultConfig().RootDirectory = %q, want %q", config.RootDirectory, ".go-uml-statemachine-parsers")
	}
	if config.ValidationLevel != StrictnessInProgress {
		t.Errorf("DefaultConfig().ValidationLevel = %v, want %v", config.ValidationLevel, StrictnessInProgress)
	}
	if config.BackupEnabled != false {
		t.Errorf("DefaultConfig().BackupEnabled = %v, want %v", config.BackupEnabled, false)
	}
	if config.MaxFileSize != 1024*1024 {
		t.Errorf("DefaultConfig().MaxFileSize = %d, want %d", config.MaxFileSize, 1024*1024)
	}
	if config.EnableDebugLogging != false {
		t.Errorf("DefaultConfig().EnableDebugLogging = %v, want %v", config.EnableDebugLogging, false)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Save original environment
	originalRootDir := os.Getenv("GO_UML_ROOT_DIRECTORY")
	originalValidationLevel := os.Getenv("GO_UML_VALIDATION_LEVEL")
	originalBackupEnabled := os.Getenv("GO_UML_BACKUP_ENABLED")
	originalMaxFileSize := os.Getenv("GO_UML_MAX_FILE_SIZE")
	originalDebugLogging := os.Getenv("GO_UML_DEBUG_LOGGING")

	// Clean up after test
	defer func() {
		restoreEnv := func(key, original string) {
			if original != "" {
				os.Setenv(key, original)
			} else {
				os.Unsetenv(key)
			}
		}
		restoreEnv("GO_UML_ROOT_DIRECTORY", originalRootDir)
		restoreEnv("GO_UML_VALIDATION_LEVEL", originalValidationLevel)
		restoreEnv("GO_UML_BACKUP_ENABLED", originalBackupEnabled)
		restoreEnv("GO_UML_MAX_FILE_SIZE", originalMaxFileSize)
		restoreEnv("GO_UML_DEBUG_LOGGING", originalDebugLogging)
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		expected *Config
	}{
		{
			name:    "no environment variables",
			envVars: map[string]string{},
			expected: &Config{
				RootDirectory:      ".go-uml-statemachine-parsers",
				ValidationLevel:    StrictnessInProgress,
				BackupEnabled:      false,
				MaxFileSize:        1024 * 1024,
				EnableDebugLogging: false,
			},
		},
		{
			name: "all environment variables set",
			envVars: map[string]string{
				"GO_UML_ROOT_DIRECTORY":   ".custom-statemachine",
				"GO_UML_VALIDATION_LEVEL": "products",
				"GO_UML_BACKUP_ENABLED":   "true",
				"GO_UML_MAX_FILE_SIZE":    "2097152",
				"GO_UML_DEBUG_LOGGING":    "true",
			},
			expected: &Config{
				RootDirectory:      ".custom-statemachine",
				ValidationLevel:    StrictnessProducts,
				BackupEnabled:      true,
				MaxFileSize:        2097152,
				EnableDebugLogging: true,
			},
		},
		{
			name: "validation level variations",
			envVars: map[string]string{
				"GO_UML_VALIDATION_LEVEL": "in-progress",
			},
			expected: &Config{
				RootDirectory:      ".go-uml-statemachine-parsers",
				ValidationLevel:    StrictnessInProgress,
				BackupEnabled:      false,
				MaxFileSize:        1024 * 1024,
				EnableDebugLogging: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all environment variables first
			os.Unsetenv("GO_UML_ROOT_DIRECTORY")
			os.Unsetenv("GO_UML_VALIDATION_LEVEL")
			os.Unsetenv("GO_UML_BACKUP_ENABLED")
			os.Unsetenv("GO_UML_MAX_FILE_SIZE")
			os.Unsetenv("GO_UML_DEBUG_LOGGING")

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			config := LoadConfigFromEnv()
			if config == nil {
				t.Fatal("LoadConfigFromEnv() returned nil")
			}

			// Verify configuration values
			if config.RootDirectory != tt.expected.RootDirectory {
				t.Errorf("RootDirectory = %q, want %q", config.RootDirectory, tt.expected.RootDirectory)
			}
			if config.ValidationLevel != tt.expected.ValidationLevel {
				t.Errorf("ValidationLevel = %v, want %v", config.ValidationLevel, tt.expected.ValidationLevel)
			}
			if config.BackupEnabled != tt.expected.BackupEnabled {
				t.Errorf("BackupEnabled = %v, want %v", config.BackupEnabled, tt.expected.BackupEnabled)
			}
			if config.MaxFileSize != tt.expected.MaxFileSize {
				t.Errorf("MaxFileSize = %d, want %d", config.MaxFileSize, tt.expected.MaxFileSize)
			}
			if config.EnableDebugLogging != tt.expected.EnableDebugLogging {
				t.Errorf("EnableDebugLogging = %v, want %v", config.EnableDebugLogging, tt.expected.EnableDebugLogging)
			}
		})
	}
}

func TestPublicAPIIntegration(t *testing.T) {
	// Test the complete workflow using the public API
	svc, err := NewService()
	if err != nil {
		t.Fatalf("NewService() failed: %v", err)
	}

	// Test content
	content := `@startuml
title Test State-Machine Diagram

[*] --> Idle
Idle --> Active : start()
Active --> Idle : stop()
Active --> Error : error()
Error --> Idle : reset()

@enduml`

	// Test Create
	sm, err := svc.Create(FileTypePUML, "test-integration", "1.0.0", content, LocationInProgress)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}
	if sm == nil {
		t.Fatal("Create() returned nil state-machine diagram")
	}
	if sm.Name != "test-integration" {
		t.Errorf("Created state-machine diagram name = %q, want %q", sm.Name, "test-integration")
	}
	if sm.Version != "1.0.0" {
		t.Errorf("Created state-machine diagram version = %q, want %q", sm.Version, "1.0.0")
	}

	// Test Read
	readSM, err := svc.Read(FileTypePUML, "test-integration", "1.0.0", LocationInProgress)
	if err != nil {
		t.Fatalf("Read() failed: %v", err)
	}
	if readSM.Content != content {
		t.Errorf("Read content mismatch")
	}

	// Test Validate
	result, err := svc.Validate(FileTypePUML, "test-integration", "1.0.0", LocationInProgress)
	if err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}
	if result == nil {
		t.Fatal("Validate() returned nil result")
	}

	// Test ListAll
	stateMachines, err := svc.ListAll(FileTypePUML, LocationInProgress)
	if err != nil {
		t.Fatalf("ListAll() failed: %v", err)
	}
	found := false
	for _, sm := range stateMachines {
		if sm.Name == "test-integration" && sm.Version == "1.0.0" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Created state-machine diagram not found in ListAll() results")
	}

	// Test Update
	updatedContent := content + "\n' Updated comment"
	readSM.Content = updatedContent
	err = svc.Update(readSM)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	// Verify update
	updatedSM, err := svc.Read(FileTypePUML, "test-integration", "1.0.0", LocationInProgress)
	if err != nil {
		t.Fatalf("Read() after update failed: %v", err)
	}
	if updatedSM.Content != updatedContent {
		t.Error("Update() did not persist changes")
	}

	// Test Promote (if validation passes)
	if result.IsValid && !result.HasErrors() {
		err = svc.Promote(FileTypePUML, "test-integration", "1.0.0")
		if err != nil {
			t.Fatalf("Promote() failed: %v", err)
		}

		// Verify promotion
		productSMs, err := svc.ListAll(FileTypePUML, LocationProducts)
		if err != nil {
			t.Fatalf("ListAll(LocationProducts) failed: %v", err)
		}
		found = false
		for _, sm := range productSMs {
			if sm.Name == "test-integration" && sm.Version == "1.0.0" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Promoted state-machine diagram not found in products")
		}

		// Clean up from products
		err = svc.Delete(FileTypePUML, "test-integration", "1.0.0", LocationProducts)
		if err != nil {
			t.Logf("Warning: Could not clean up from products: %v", err)
		}
	} else {
		// Clean up from in-progress
		err = svc.Delete(FileTypePUML, "test-integration", "1.0.0", LocationInProgress)
		if err != nil {
			t.Logf("Warning: Could not clean up from in-progress: %v", err)
		}
	}
}

func TestPublicAPIErrorHandling(t *testing.T) {
	svc, err := NewService()
	if err != nil {
		t.Fatalf("NewService() failed: %v", err)
	}

	// Test Create with invalid parameters
	_, err = svc.Create(FileTypePUML, "", "1.0.0", "content", LocationInProgress)
	if err == nil {
		t.Error("Create() with empty name should fail")
	}

	_, err = svc.Create(FileTypePUML, "test", "", "content", LocationInProgress)
	if err == nil {
		t.Error("Create() with empty version should fail")
	}

	_, err = svc.Create(FileTypePUML, "test", "1.0.0", "", LocationInProgress)
	if err == nil {
		t.Error("Create() with empty content should fail")
	}

	// Test Read non-existent
	_, err = svc.Read(FileTypePUML, "non-existent", "1.0.0", LocationInProgress)
	if err == nil {
		t.Error("Read() of non-existent state-machine diagram should fail")
	}

	// Test Delete non-existent
	err = svc.Delete(FileTypePUML, "non-existent", "1.0.0", LocationInProgress)
	if err == nil {
		t.Error("Delete() of non-existent state-machine diagram should fail")
	}

	// Test Promote non-existent
	err = svc.Promote(FileTypePUML, "non-existent", "1.0.0")
	if err == nil {
		t.Error("Promote() of non-existent state-machine diagram should fail")
	}

	// Test Validate non-existent
	_, err = svc.Validate(FileTypePUML, "non-existent", "1.0.0", LocationInProgress)
	if err == nil {
		t.Error("Validate() of non-existent state-machine diagram should fail")
	}
}

func TestConstants(t *testing.T) {
	// Test Location constants
	if LocationInProgress != 0 {
		t.Errorf("LocationInProgress = %d, want 0", LocationInProgress)
	}
	if LocationProducts != 1 {
		t.Errorf("LocationProducts = %d, want 1", LocationProducts)
	}
	if LocationNested != 2 {
		t.Errorf("LocationNested = %d, want 2", LocationNested)
	}

	// Test ReferenceType constants
	if ReferenceTypeProduct != 0 {
		t.Errorf("ReferenceTypeProduct = %d, want 0", ReferenceTypeProduct)
	}
	if ReferenceTypeNested != 1 {
		t.Errorf("ReferenceTypeNested = %d, want 1", ReferenceTypeNested)
	}

	// Test ValidationStrictness constants
	if StrictnessInProgress != 0 {
		t.Errorf("StrictnessInProgress = %d, want 0", StrictnessInProgress)
	}
	if StrictnessProducts != 1 {
		t.Errorf("StrictnessProducts = %d, want 1", StrictnessProducts)
	}
}
