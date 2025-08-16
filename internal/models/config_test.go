package models

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.RootDirectory != ".go-uml-statemachine" {
		t.Errorf("Expected RootDirectory to be '.go-uml-statemachine', got %s", config.RootDirectory)
	}

	if config.ValidationLevel != StrictnessInProgress {
		t.Errorf("Expected ValidationLevel to be StrictnessInProgress, got %v", config.ValidationLevel)
	}

	if config.BackupEnabled != false {
		t.Errorf("Expected BackupEnabled to be false, got %v", config.BackupEnabled)
	}

	if config.MaxFileSize != 1024*1024 {
		t.Errorf("Expected MaxFileSize to be 1048576, got %d", config.MaxFileSize)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Save original environment variables
	originalRootDir := os.Getenv("GO_UML_ROOT_DIRECTORY")
	originalValidationLevel := os.Getenv("GO_UML_VALIDATION_LEVEL")
	originalBackupEnabled := os.Getenv("GO_UML_BACKUP_ENABLED")
	originalMaxFileSize := os.Getenv("GO_UML_MAX_FILE_SIZE")

	// Clean up after test
	defer func() {
		os.Setenv("GO_UML_ROOT_DIRECTORY", originalRootDir)
		os.Setenv("GO_UML_VALIDATION_LEVEL", originalValidationLevel)
		os.Setenv("GO_UML_BACKUP_ENABLED", originalBackupEnabled)
		os.Setenv("GO_UML_MAX_FILE_SIZE", originalMaxFileSize)
	}()

	tests := []struct {
		name                    string
		rootDirectory           string
		validationLevel         string
		backupEnabled           string
		maxFileSize             string
		expectedRootDirectory   string
		expectedValidationLevel ValidationStrictness
		expectedBackupEnabled   bool
		expectedMaxFileSize     int64
	}{
		{
			name:                    "default values when no env vars set",
			rootDirectory:           "",
			validationLevel:         "",
			backupEnabled:           "",
			maxFileSize:             "",
			expectedRootDirectory:   ".go-uml-statemachine",
			expectedValidationLevel: StrictnessInProgress,
			expectedBackupEnabled:   false,
			expectedMaxFileSize:     1024 * 1024,
		},
		{
			name:                    "custom root directory",
			rootDirectory:           "C:\\custom\\path",
			validationLevel:         "",
			backupEnabled:           "",
			maxFileSize:             "",
			expectedRootDirectory:   "C:\\custom\\path",
			expectedValidationLevel: StrictnessInProgress,
			expectedBackupEnabled:   false,
			expectedMaxFileSize:     1024 * 1024,
		},
		{
			name:                    "validation level in-progress",
			rootDirectory:           "",
			validationLevel:         "in-progress",
			backupEnabled:           "",
			maxFileSize:             "",
			expectedRootDirectory:   ".go-uml-statemachine",
			expectedValidationLevel: StrictnessInProgress,
			expectedBackupEnabled:   false,
			expectedMaxFileSize:     1024 * 1024,
		},
		{
			name:                    "validation level products",
			rootDirectory:           "",
			validationLevel:         "products",
			backupEnabled:           "",
			maxFileSize:             "",
			expectedRootDirectory:   ".go-uml-statemachine",
			expectedValidationLevel: StrictnessProducts,
			expectedBackupEnabled:   false,
			expectedMaxFileSize:     1024 * 1024,
		},
		{
			name:                    "validation level case insensitive",
			rootDirectory:           "",
			validationLevel:         "PRODUCTS",
			backupEnabled:           "",
			maxFileSize:             "",
			expectedRootDirectory:   ".go-uml-statemachine",
			expectedValidationLevel: StrictnessProducts,
			expectedBackupEnabled:   false,
			expectedMaxFileSize:     1024 * 1024,
		},
		{
			name:                    "backup enabled true",
			rootDirectory:           "",
			validationLevel:         "",
			backupEnabled:           "true",
			maxFileSize:             "",
			expectedRootDirectory:   ".go-uml-statemachine",
			expectedValidationLevel: StrictnessInProgress,
			expectedBackupEnabled:   true,
			expectedMaxFileSize:     1024 * 1024,
		},
		{
			name:                    "backup enabled false",
			rootDirectory:           "",
			validationLevel:         "",
			backupEnabled:           "false",
			maxFileSize:             "",
			expectedRootDirectory:   ".go-uml-statemachine",
			expectedValidationLevel: StrictnessInProgress,
			expectedBackupEnabled:   false,
			expectedMaxFileSize:     1024 * 1024,
		},
		{
			name:                    "custom max file size",
			rootDirectory:           "",
			validationLevel:         "",
			backupEnabled:           "",
			maxFileSize:             "2097152",
			expectedRootDirectory:   ".go-uml-statemachine",
			expectedValidationLevel: StrictnessInProgress,
			expectedBackupEnabled:   false,
			expectedMaxFileSize:     2097152,
		},
		{
			name:                    "all custom values",
			rootDirectory:           "D:\\state-machines",
			validationLevel:         "products",
			backupEnabled:           "true",
			maxFileSize:             "5242880",
			expectedRootDirectory:   "D:\\state-machines",
			expectedValidationLevel: StrictnessProducts,
			expectedBackupEnabled:   true,
			expectedMaxFileSize:     5242880,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("GO_UML_ROOT_DIRECTORY", tt.rootDirectory)
			os.Setenv("GO_UML_VALIDATION_LEVEL", tt.validationLevel)
			os.Setenv("GO_UML_BACKUP_ENABLED", tt.backupEnabled)
			os.Setenv("GO_UML_MAX_FILE_SIZE", tt.maxFileSize)

			config := LoadConfigFromEnv()

			if config.RootDirectory != tt.expectedRootDirectory {
				t.Errorf("Expected RootDirectory to be %s, got %s", tt.expectedRootDirectory, config.RootDirectory)
			}

			if config.ValidationLevel != tt.expectedValidationLevel {
				t.Errorf("Expected ValidationLevel to be %v, got %v", tt.expectedValidationLevel, config.ValidationLevel)
			}

			if config.BackupEnabled != tt.expectedBackupEnabled {
				t.Errorf("Expected BackupEnabled to be %v, got %v", tt.expectedBackupEnabled, config.BackupEnabled)
			}

			if config.MaxFileSize != tt.expectedMaxFileSize {
				t.Errorf("Expected MaxFileSize to be %d, got %d", tt.expectedMaxFileSize, config.MaxFileSize)
			}
		})
	}
}

func TestLoadConfigFromEnvInvalidValues(t *testing.T) {
	// Save original environment variables
	originalBackupEnabled := os.Getenv("GO_UML_BACKUP_ENABLED")
	originalMaxFileSize := os.Getenv("GO_UML_MAX_FILE_SIZE")

	// Clean up after test
	defer func() {
		os.Setenv("GO_UML_BACKUP_ENABLED", originalBackupEnabled)
		os.Setenv("GO_UML_MAX_FILE_SIZE", originalMaxFileSize)
	}()

	tests := []struct {
		name                  string
		backupEnabled         string
		maxFileSize           string
		expectedBackupEnabled bool
		expectedMaxFileSize   int64
	}{
		{
			name:                  "invalid backup enabled value",
			backupEnabled:         "invalid",
			maxFileSize:           "",
			expectedBackupEnabled: false, // Should use default
			expectedMaxFileSize:   1024 * 1024,
		},
		{
			name:                  "invalid max file size value",
			backupEnabled:         "",
			maxFileSize:           "invalid",
			expectedBackupEnabled: false,
			expectedMaxFileSize:   1024 * 1024, // Should use default
		},
		{
			name:                  "negative max file size",
			backupEnabled:         "",
			maxFileSize:           "-1000",
			expectedBackupEnabled: false,
			expectedMaxFileSize:   1024 * 1024, // Should use default
		},
		{
			name:                  "zero max file size",
			backupEnabled:         "",
			maxFileSize:           "0",
			expectedBackupEnabled: false,
			expectedMaxFileSize:   1024 * 1024, // Should use default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("GO_UML_BACKUP_ENABLED", tt.backupEnabled)
			os.Setenv("GO_UML_MAX_FILE_SIZE", tt.maxFileSize)

			config := LoadConfigFromEnv()

			if config.BackupEnabled != tt.expectedBackupEnabled {
				t.Errorf("Expected BackupEnabled to be %v, got %v", tt.expectedBackupEnabled, config.BackupEnabled)
			}

			if config.MaxFileSize != tt.expectedMaxFileSize {
				t.Errorf("Expected MaxFileSize to be %d, got %d", tt.expectedMaxFileSize, config.MaxFileSize)
			}
		})
	}
}

func TestMergeWithEnv(t *testing.T) {
	// Save original environment variables
	originalRootDir := os.Getenv("GO_UML_ROOT_DIRECTORY")
	originalValidationLevel := os.Getenv("GO_UML_VALIDATION_LEVEL")
	originalBackupEnabled := os.Getenv("GO_UML_BACKUP_ENABLED")
	originalMaxFileSize := os.Getenv("GO_UML_MAX_FILE_SIZE")

	// Clean up after test
	defer func() {
		os.Setenv("GO_UML_ROOT_DIRECTORY", originalRootDir)
		os.Setenv("GO_UML_VALIDATION_LEVEL", originalValidationLevel)
		os.Setenv("GO_UML_BACKUP_ENABLED", originalBackupEnabled)
		os.Setenv("GO_UML_MAX_FILE_SIZE", originalMaxFileSize)
	}()

	// Create a base config with custom values
	baseConfig := &Config{
		RootDirectory:   "C:\\base\\path",
		ValidationLevel: StrictnessProducts,
		BackupEnabled:   true,
		MaxFileSize:     2048 * 1024,
	}

	tests := []struct {
		name                    string
		envRootDirectory        string
		envValidationLevel      string
		envBackupEnabled        string
		envMaxFileSize          string
		expectedRootDirectory   string
		expectedValidationLevel ValidationStrictness
		expectedBackupEnabled   bool
		expectedMaxFileSize     int64
	}{
		{
			name:                    "no env vars - keep base config",
			envRootDirectory:        "",
			envValidationLevel:      "",
			envBackupEnabled:        "",
			envMaxFileSize:          "",
			expectedRootDirectory:   "C:\\base\\path",
			expectedValidationLevel: StrictnessProducts,
			expectedBackupEnabled:   true,
			expectedMaxFileSize:     2048 * 1024,
		},
		{
			name:                    "override root directory only",
			envRootDirectory:        "D:\\env\\path",
			envValidationLevel:      "",
			envBackupEnabled:        "",
			envMaxFileSize:          "",
			expectedRootDirectory:   "D:\\env\\path",
			expectedValidationLevel: StrictnessProducts,
			expectedBackupEnabled:   true,
			expectedMaxFileSize:     2048 * 1024,
		},
		{
			name:                    "override all values",
			envRootDirectory:        "E:\\override\\path",
			envValidationLevel:      "in-progress",
			envBackupEnabled:        "false",
			envMaxFileSize:          "512000",
			expectedRootDirectory:   "E:\\override\\path",
			expectedValidationLevel: StrictnessInProgress,
			expectedBackupEnabled:   false,
			expectedMaxFileSize:     512000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("GO_UML_ROOT_DIRECTORY", tt.envRootDirectory)
			os.Setenv("GO_UML_VALIDATION_LEVEL", tt.envValidationLevel)
			os.Setenv("GO_UML_BACKUP_ENABLED", tt.envBackupEnabled)
			os.Setenv("GO_UML_MAX_FILE_SIZE", tt.envMaxFileSize)

			// Create a copy of base config and merge with env
			config := *baseConfig
			mergedConfig := config.MergeWithEnv()

			if mergedConfig.RootDirectory != tt.expectedRootDirectory {
				t.Errorf("Expected RootDirectory to be %s, got %s", tt.expectedRootDirectory, mergedConfig.RootDirectory)
			}

			if mergedConfig.ValidationLevel != tt.expectedValidationLevel {
				t.Errorf("Expected ValidationLevel to be %v, got %v", tt.expectedValidationLevel, mergedConfig.ValidationLevel)
			}

			if mergedConfig.BackupEnabled != tt.expectedBackupEnabled {
				t.Errorf("Expected BackupEnabled to be %v, got %v", tt.expectedBackupEnabled, mergedConfig.BackupEnabled)
			}

			if mergedConfig.MaxFileSize != tt.expectedMaxFileSize {
				t.Errorf("Expected MaxFileSize to be %d, got %d", tt.expectedMaxFileSize, mergedConfig.MaxFileSize)
			}
		})
	}
}
