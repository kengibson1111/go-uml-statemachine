package main

import (
	"fmt"
	"os"

	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/repository"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/service"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/validation"
)

func main() {
	fmt.Println("Go UML State-Machine Diagram - Configuration Demo")
	fmt.Println("=========================================")

	// Demo 1: Default configuration
	fmt.Println("\n1. Default Configuration:")
	defaultConfig := models.DefaultConfig()
	printConfig(defaultConfig)

	// Demo 2: Configuration from environment variables
	fmt.Println("\n2. Setting environment variables...")
	os.Setenv("GO_UML_ROOT_DIRECTORY", "C:\\my-diagrams")
	os.Setenv("GO_UML_VALIDATION_LEVEL", "products")
	os.Setenv("GO_UML_BACKUP_ENABLED", "true")
	os.Setenv("GO_UML_MAX_FILE_SIZE", "2097152")

	fmt.Println("\n3. Configuration loaded from environment:")
	envConfig := models.LoadConfigFromEnv()
	printConfig(envConfig)

	// Demo 3: Merging configuration with environment overrides
	fmt.Println("\n4. Base configuration with environment overrides:")
	baseConfig := &models.Config{
		RootDirectory:   "D:\\base-directory",
		ValidationLevel: models.StrictnessInProgress,
		BackupEnabled:   false,
		MaxFileSize:     512000,
	}
	fmt.Println("Base config:")
	printConfig(baseConfig)

	// Only override some values via environment
	os.Setenv("GO_UML_ROOT_DIRECTORY", "E:\\override-directory")
	os.Setenv("GO_UML_BACKUP_ENABLED", "true")
	os.Unsetenv("GO_UML_VALIDATION_LEVEL") // Remove this env var
	os.Unsetenv("GO_UML_MAX_FILE_SIZE")    // Remove this env var

	mergedConfig := *baseConfig
	mergedConfig.MergeWithEnv()
	fmt.Println("\nMerged config (with partial env overrides):")
	printConfig(&mergedConfig)

	// Demo 4: Service factory functions
	fmt.Println("\n5. Service Factory Functions:")

	// Create dependencies
	tempConfig := &models.Config{
		RootDirectory:   "C:\\temp\\demo",
		ValidationLevel: models.StrictnessInProgress,
		BackupEnabled:   false,
		MaxFileSize:     1024 * 1024,
	}
	repo := repository.NewFileSystemRepository(tempConfig)
	validator := validation.NewPlantUMLValidator()

	// Service with default config
	fmt.Println("Creating service with default config...")
	svc1 := service.NewServiceWithDefaults(repo, validator)
	if svc1 != nil {
		fmt.Println("✓ Service created successfully with default config")
	}

	// Service from environment
	fmt.Println("Creating service from environment config...")
	svc2 := service.NewServiceFromEnv(repo, validator)
	if svc2 != nil {
		fmt.Println("✓ Service created successfully from environment config")
	}

	// Service with environment overrides
	fmt.Println("Creating service with environment overrides...")
	customConfig := &models.Config{
		RootDirectory:   "F:\\custom-path",
		ValidationLevel: models.StrictnessProducts,
		BackupEnabled:   false,
		MaxFileSize:     1024000,
	}
	svc3 := service.NewServiceWithEnvOverrides(repo, validator, customConfig)
	if svc3 != nil {
		fmt.Println("✓ Service created successfully with environment overrides")
	}

	// Clean up environment variables
	fmt.Println("\n6. Cleaning up environment variables...")
	os.Unsetenv("GO_UML_ROOT_DIRECTORY")
	os.Unsetenv("GO_UML_VALIDATION_LEVEL")
	os.Unsetenv("GO_UML_BACKUP_ENABLED")
	os.Unsetenv("GO_UML_MAX_FILE_SIZE")

	fmt.Println("\nDemo completed successfully!")
}

func printConfig(config *models.Config) {
	fmt.Printf("  Root Directory: %s\n", config.RootDirectory)
	fmt.Printf("  Validation Level: %s\n", config.ValidationLevel.String())
	fmt.Printf("  Backup Enabled: %t\n", config.BackupEnabled)
	fmt.Printf("  Max File Size: %d bytes\n", config.MaxFileSize)
}
