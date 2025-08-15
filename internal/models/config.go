package models

// Config represents the configuration for the state machine system
type Config struct {
	RootDirectory   string               // Default: ".go-uml-statemachine"
	ValidationLevel ValidationStrictness // Default validation level
	BackupEnabled   bool                 // Whether to create backups
	MaxFileSize     int64                // Maximum file size in bytes
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		RootDirectory:   ".go-uml-statemachine",
		ValidationLevel: StrictnessInProgress,
		BackupEnabled:   false,
		MaxFileSize:     1024 * 1024, // 1MB
	}
}
