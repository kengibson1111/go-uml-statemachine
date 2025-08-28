package main

import (
	"fmt"
	"os"

	smmodels "github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/logging"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/repository"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/service"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/validation"
)

func main() {
	fmt.Println("=== Go UML State-Machine Diagram - Error Handling and Logging Demo ===")
	fmt.Println()

	// Configure logging with debug level
	config := models.DefaultConfig()
	config.EnableDebugLogging = true

	// Create logger for the demo
	loggerConfig := &logging.LoggerConfig{
		Level:        logging.LogLevelDebug,
		Prefix:       "[ErrorHandlingDemo]",
		EnableCaller: true,
	}

	logger, err := logging.NewLogger(loggerConfig)
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}

	// Set as global logger
	logging.SetGlobalLogger(logger)

	logger.Info("Starting error handling and logging demonstration")

	// Create components
	repo := repository.NewFileSystemRepository(config)
	validator := validation.NewPlantUMLValidator()
	svc := service.NewService(repo, validator, config)

	// Demonstrate various error scenarios
	demonstrateValidationErrors(svc, logger)
	demonstrateFileSystemErrors(svc, logger)
	demonstrateErrorWrapping(svc, logger)
	demonstrateErrorSeverities(svc, logger)

	logger.Info("Error handling and logging demonstration completed")
}

func demonstrateValidationErrors(svc models.DiagramService, logger *logging.Logger) {
	logger.Info("=== Demonstrating Validation Errors ===")

	// Test empty name
	logger.Info("Testing empty name validation...")
	_, err := svc.CreateFile(smmodels.DiagramTypePUML, "", "1.0.0", "content", models.LocationInProgress)
	if err != nil {
		if diagErr, ok := err.(*models.StateMachineError); ok {
			logger.WithFields(map[string]any{
				"errorType":   diagErr.Type.String(),
				"severity":    diagErr.Severity.String(),
				"operation":   diagErr.Operation,
				"component":   diagErr.Component,
				"recoverable": diagErr.Recoverable,
			}).Error("Validation error caught as expected")

			fmt.Println("Detailed Error Information:")
			fmt.Println(diagErr.DetailedError())
		}
	}

	// Test empty version
	logger.Info("Testing empty version validation...")
	_, err = svc.CreateFile(smmodels.DiagramTypePUML, "test", "", "content", models.LocationInProgress)
	if err != nil {
		if diagErr, ok := err.(*models.StateMachineError); ok {
			logger.WithField("errorType", diagErr.Type.String()).Error("Version validation error caught as expected")
		}
	}

	// Test empty content
	logger.Info("Testing empty content validation...")
	_, err = svc.CreateFile(smmodels.DiagramTypePUML, "test", "1.0.0", "", models.LocationInProgress)
	if err != nil {
		if diagErr, ok := err.(*models.StateMachineError); ok {
			logger.WithField("errorType", diagErr.Type.String()).Error("Content validation error caught as expected")
		}
	}
}

func demonstrateFileSystemErrors(svc models.DiagramService, logger *logging.Logger) {
	logger.Info("=== Demonstrating File System Errors ===")

	// Try to read a non-existent state-machine diagram
	logger.Info("Testing file not found error...")
	_, err := svc.ReadFile(smmodels.DiagramTypePUML, "nonexistent", "1.0.0", models.LocationInProgress)
	if err != nil {
		if diagErr, ok := err.(*models.StateMachineError); ok {
			logger.WithFields(map[string]any{
				"errorType": diagErr.Type.String(),
				"severity":  diagErr.Severity.String(),
				"context":   diagErr.Context,
			}).Error("File not found error caught as expected")
		}
	}
}

func demonstrateErrorWrapping(svc models.DiagramService, logger *logging.Logger) {
	logger.Info("=== Demonstrating Error Wrapping ===")

	// Create a state-machine diagram first
	logger.Info("Creating a state-machine diagram for conflict demonstration...")
	validContent := `@startuml
[*] --> Idle
Idle --> Active : start
Active --> [*] : stop
@enduml`

	diag, err := svc.CreateFile(smmodels.DiagramTypePUML, "demo-diag", "1.0.0", validContent, models.LocationInProgress)
	if err != nil {
		logger.WithError(err).Error("Failed to create demo state-machine diagram")
		return
	}

	logger.WithField("name", diag.Name).Info("Demo state-machine diagram created successfully")

	// Try to create the same state-machine diagram again (should cause conflict)
	logger.Info("Attempting to create duplicate state-machine diagram...")
	_, err = svc.CreateFile(smmodels.DiagramTypePUML, "demo-diag", "1.0.0", validContent, models.LocationInProgress)
	if err != nil {
		if diagErr, ok := err.(*models.StateMachineError); ok {
			logger.WithFields(map[string]any{
				"errorType":   diagErr.Type.String(),
				"severity":    diagErr.Severity.String(),
				"recoverable": diagErr.Recoverable,
			}).Error("Directory conflict error caught as expected")

			// Demonstrate error unwrapping
			if diagErr.Cause != nil {
				logger.WithField("cause", diagErr.Cause.Error()).Info("Error has wrapped cause")
			}
		}
	}

	// Clean up
	logger.Info("Cleaning up demo state-machine diagram...")
	if err := svc.DeleteFile(smmodels.DiagramTypePUML, "demo-diag", "1.0.0", models.LocationInProgress); err != nil {
		logger.WithError(err).Warn("Failed to clean up demo state-machine diagram")
	}
}

func demonstrateErrorSeverities(svc models.DiagramService, logger *logging.Logger) {
	logger.Info("=== Demonstrating Error Severities ===")

	// Create errors with different severities
	errors := []error{
		models.NewStateMachineError(models.ErrorTypeValidation, "Low severity validation error", nil).
			WithSeverity(models.ErrorSeverityLow),
		models.NewStateMachineError(models.ErrorTypeFileSystem, "Medium severity file system error", nil).
			WithSeverity(models.ErrorSeverityMedium),
		models.NewStateMachineError(models.ErrorTypeDirectoryConflict, "High severity conflict error", nil).
			WithSeverity(models.ErrorSeverityHigh),
		models.NewCriticalError(models.ErrorTypeCorruption, "Critical data corruption error", nil),
	}

	for i, err := range errors {
		if diagErr, ok := err.(*models.StateMachineError); ok {
			logger.WithFields(map[string]any{
				"errorIndex":  i + 1,
				"errorType":   diagErr.Type.String(),
				"severity":    diagErr.Severity.String(),
				"recoverable": diagErr.Recoverable,
			}).Info("Demonstrating error severity")

			// Log at appropriate level based on severity
			switch diagErr.Severity {
			case models.ErrorSeverityLow:
				logger.WithError(err).Debug("Low severity error")
			case models.ErrorSeverityMedium:
				logger.WithError(err).Info("Medium severity error")
			case models.ErrorSeverityHigh:
				logger.WithError(err).Warn("High severity error")
			case models.ErrorSeverityCritical:
				logger.WithError(err).Error("Critical severity error")
			}
		}
	}

	// Demonstrate error utility functions
	logger.Info("=== Demonstrating Error Utility Functions ===")

	testErr := models.NewStateMachineError(models.ErrorTypeValidation, "test error", nil).
		WithSeverity(models.ErrorSeverityHigh).
		WithRecoverable(false)

	logger.WithFields(map[string]any{
		"isRecoverable": models.IsRecoverable(testErr),
		"errorType":     models.GetErrorType(testErr).String(),
		"severity":      models.GetErrorSeverity(testErr).String(),
	}).Info("Error utility functions demonstration")
}
