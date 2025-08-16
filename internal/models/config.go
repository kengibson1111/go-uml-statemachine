package models

import (
	"os"
	"strconv"
	"strings"
)

// Config represents the configuration for the state machine system
type Config struct {
	RootDirectory      string               // Default: ".go-uml-statemachine"
	ValidationLevel    ValidationStrictness // Default validation level
	BackupEnabled      bool                 // Whether to create backups
	MaxFileSize        int64                // Maximum file size in bytes
	EnableDebugLogging bool                 // Whether to enable debug logging
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		RootDirectory:      ".go-uml-statemachine",
		ValidationLevel:    StrictnessInProgress,
		BackupEnabled:      false,
		MaxFileSize:        1024 * 1024, // 1MB
		EnableDebugLogging: false,
	}
}

// LoadConfigFromEnv loads configuration from environment variables
// Environment variables:
// - GO_UML_ROOT_DIRECTORY: Root directory for state machines
// - GO_UML_VALIDATION_LEVEL: Validation level (in-progress, products)
// - GO_UML_BACKUP_ENABLED: Whether to enable backups (true/false)
// - GO_UML_MAX_FILE_SIZE: Maximum file size in bytes
// - GO_UML_DEBUG_LOGGING: Whether to enable debug logging (true/false)
func LoadConfigFromEnv() *Config {
	config := DefaultConfig()

	// Load root directory
	if rootDir := os.Getenv("GO_UML_ROOT_DIRECTORY"); rootDir != "" {
		config.RootDirectory = rootDir
	}

	// Load validation level
	if validationLevel := os.Getenv("GO_UML_VALIDATION_LEVEL"); validationLevel != "" {
		switch strings.ToLower(validationLevel) {
		case "in-progress", "inprogress":
			config.ValidationLevel = StrictnessInProgress
		case "products", "product":
			config.ValidationLevel = StrictnessProducts
		}
	}

	// Load backup enabled
	if backupEnabled := os.Getenv("GO_UML_BACKUP_ENABLED"); backupEnabled != "" {
		if enabled, err := strconv.ParseBool(backupEnabled); err == nil {
			config.BackupEnabled = enabled
		}
	}

	// Load max file size
	if maxFileSize := os.Getenv("GO_UML_MAX_FILE_SIZE"); maxFileSize != "" {
		if size, err := strconv.ParseInt(maxFileSize, 10, 64); err == nil && size > 0 {
			config.MaxFileSize = size
		}
	}

	// Load debug logging
	if debugLogging := os.Getenv("GO_UML_DEBUG_LOGGING"); debugLogging != "" {
		if enabled, err := strconv.ParseBool(debugLogging); err == nil {
			config.EnableDebugLogging = enabled
		}
	}

	return config
}

// MergeWithEnv merges the current config with environment variables
// Environment variables take precedence over existing values
func (c *Config) MergeWithEnv() *Config {
	envConfig := LoadConfigFromEnv()

	// Only override if environment variable was actually set
	if os.Getenv("GO_UML_ROOT_DIRECTORY") != "" {
		c.RootDirectory = envConfig.RootDirectory
	}
	if os.Getenv("GO_UML_VALIDATION_LEVEL") != "" {
		c.ValidationLevel = envConfig.ValidationLevel
	}
	if os.Getenv("GO_UML_BACKUP_ENABLED") != "" {
		c.BackupEnabled = envConfig.BackupEnabled
	}
	if os.Getenv("GO_UML_MAX_FILE_SIZE") != "" {
		c.MaxFileSize = envConfig.MaxFileSize
	}
	if os.Getenv("GO_UML_DEBUG_LOGGING") != "" {
		c.EnableDebugLogging = envConfig.EnableDebugLogging
	}

	return c
}
