package repository

import (
	"fmt"
	"os"

	smmodels "github.com/kengibson1111/go-uml-statemachine-models/models"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/logging"
	"github.com/kengibson1111/go-uml-statemachine-parsers/internal/models"
)

// FileSystemRepository implements the Repository interface using the file system
type FileSystemRepository struct {
	pathManager *models.PathManager
	config      *models.Config
	logger      *logging.Logger
}

// NewFileSystemRepository creates a new FileSystemRepository
func NewFileSystemRepository(config *models.Config) *FileSystemRepository {
	if config == nil {
		config = models.DefaultConfig()
	}

	pathManager := models.NewPathManager(config.RootDirectory)

	// Create logger with repository-specific configuration
	loggerConfig := &logging.LoggerConfig{
		Level:        logging.LogLevelInfo,
		Prefix:       "[FileSystemRepository]",
		EnableCaller: true,
	}

	// Set log level based on config if available
	if config.EnableDebugLogging {
		loggerConfig.Level = logging.LogLevelDebug
	}

	logger, err := logging.NewLogger(loggerConfig)
	if err != nil {
		// Fallback to default logger if creation fails
		logger = logging.NewDefaultLogger()
		logger.Warn("Failed to create repository logger, using default")
	}

	repo := &FileSystemRepository{
		pathManager: pathManager,
		config:      config,
		logger:      logger,
	}

	repo.logger.WithField("rootDirectory", config.RootDirectory).Info("FileSystemRepository initialized")
	return repo
}

// ReadDiagram reads a state-machine diagram from the file system
func (r *FileSystemRepository) ReadDiagram(diagramType smmodels.DiagramType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
	// Create operation logger with context
	opLogger := r.logger.WithFields(map[string]interface{}{
		"operation":   "ReadStateMachine",
		"diagramType": diagramType.String(),
		"name":        name,
		"version":     version,
		"location":    location.String(),
	})

	opLogger.Debug("Starting state-machine diagram read operation")

	// Validate inputs
	if err := r.pathManager.ValidateName(name); err != nil {
		wrappedErr := models.WrapError(err, models.ErrorTypeValidation, "invalid state-machine diagram name").
			WithOperation("ReadStateMachine").
			WithComponent("repository").
			WithContext("name", name)
		opLogger.WithError(wrappedErr).Error("Name validation failed")
		return nil, wrappedErr
	}

	if version == "" {
		err := models.NewStateMachineError(models.ErrorTypeValidation, "version is required for all state-machine diagrams", nil).
			WithContext("name", name).
			WithContext("location", location.String()).
			WithOperation("ReadStateMachine").
			WithComponent("repository").
			WithSeverity(models.ErrorSeverityHigh)
		opLogger.WithError(err).Error("Version validation failed")
		return nil, err
	}

	opLogger.Debug("Input validation passed")

	// Get the file path
	filePath := r.pathManager.GetDiagramFilePathWithDiagramType(name, version, location, diagramType)
	opLogger.WithField("filePath", filePath).Debug("Resolved file path")

	// Check if file exists
	opLogger.Debug("Checking if file exists")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		notFoundErr := models.NewStateMachineError(models.ErrorTypeFileNotFound, "state-machine diagram file not found", err).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", location.String()).
			WithContext("filePath", filePath).
			WithOperation("ReadStateMachine").
			WithComponent("repository").
			WithSeverity(models.ErrorSeverityMedium)
		opLogger.WithError(notFoundErr).Warn("State-machine diagram file not found")
		return nil, notFoundErr
	} else if err != nil {
		statErr := models.WrapError(err, models.ErrorTypeFileSystem, "failed to check file existence").
			WithContext("filePath", filePath).
			WithOperation("ReadStateMachine").
			WithComponent("repository")
		opLogger.WithError(statErr).Error("Failed to check file existence")
		return nil, statErr
	}

	opLogger.Debug("File exists, reading content")

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		readErr := models.WrapError(err, models.ErrorTypeFileSystem, "failed to read state-machine diagram file").
			WithContext("filePath", filePath).
			WithOperation("ReadStateMachine").
			WithComponent("repository").
			WithSeverity(models.ErrorSeverityHigh)
		opLogger.WithError(readErr).Error("Failed to read file content")
		return nil, readErr
	}

	opLogger.WithField("contentSize", len(content)).Debug("File content read successfully")

	// Check file size
	if int64(len(content)) > r.config.MaxFileSize {
		sizeErr := models.NewStateMachineError(models.ErrorTypeFileSystem, "file size exceeds maximum allowed", nil).
			WithContext("fileSize", len(content)).
			WithContext("maxSize", r.config.MaxFileSize).
			WithContext("filePath", filePath).
			WithOperation("ReadStateMachine").
			WithComponent("repository").
			WithSeverity(models.ErrorSeverityHigh)
		opLogger.WithError(sizeErr).Error("File size exceeds maximum allowed")
		return nil, sizeErr
	}

	// Get file info for metadata
	opLogger.Debug("Getting file metadata")
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		infoErr := models.WrapError(err, models.ErrorTypeFileSystem, "failed to get file info").
			WithContext("filePath", filePath).
			WithOperation("ReadStateMachine").
			WithComponent("repository")
		opLogger.WithError(infoErr).Error("Failed to get file metadata")
		return nil, infoErr
	}

	// Create StateMachineDiagram object
	opLogger.Debug("Creating state-machine diagram object")
	diag := &models.StateMachineDiagram{
		Name:        name,
		Version:     version,
		Content:     string(content),
		Location:    location,
		DiagramType: diagramType,
		Metadata: models.Metadata{
			ModifiedAt: fileInfo.ModTime(),
			// CreatedAt and Author would need additional metadata storage
			CreatedAt: fileInfo.ModTime(), // Using ModTime as fallback
		},
	}

	// TODO: Parse references from content (will be implemented in validation layer)
	diag.References = []models.Reference{}

	opLogger.WithFields(map[string]interface{}{
		"contentLength": len(diag.Content),
		"modifiedAt":    diag.Metadata.ModifiedAt,
	}).Info("State-machine diagram read successfully")

	return diag, nil
}

// WriteDiagram writes a state-machine diagram to the file system
func (r *FileSystemRepository) WriteDiagram(diag *models.StateMachineDiagram) error {
	if diag == nil {
		return models.NewStateMachineError(models.ErrorTypeValidation, "state-machine diagram cannot be nil", nil)
	}

	// Validate inputs
	if err := r.pathManager.ValidateName(diag.Name); err != nil {
		return err
	}

	if diag.Version == "" {
		return models.NewStateMachineError(models.ErrorTypeValidation, "version is required for all state-machine diagrams", nil).
			WithContext("name", diag.Name).
			WithContext("location", diag.Location.String())
	}

	// Check content size
	if int64(len(diag.Content)) > r.config.MaxFileSize {
		return models.NewStateMachineError(models.ErrorTypeFileSystem, "content size exceeds maximum allowed", nil).
			WithContext("contentSize", len(diag.Content)).
			WithContext("maxSize", r.config.MaxFileSize)
	}

	// Get directory and file paths
	dirPath := r.pathManager.GetLocationWithDiagramTypePath(diag.Location, diag.DiagramType)
	filePath := r.pathManager.GetDiagramFilePathWithDiagramType(diag.Name, diag.Version, diag.Location, diag.DiagramType)

	// Create directory if it doesn't exist
	if err := r.CreateDirectory(dirPath); err != nil {
		return fmt.Errorf("failed to create directory for state-machine diagram: %w", err)
	}

	// Write the file
	if err := os.WriteFile(filePath, []byte(diag.Content), 0644); err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem, "failed to write state-machine diagram file", err).
			WithContext("filePath", filePath)
	}

	return nil
}

// Exists checks if a state-machine diagram exists
func (r *FileSystemRepository) Exists(diagramType smmodels.DiagramType, name, version string, location models.Location) (bool, error) {
	// Validate inputs
	if err := r.pathManager.ValidateName(name); err != nil {
		return false, err
	}

	if version == "" {
		return false, models.NewStateMachineError(models.ErrorTypeValidation, "version is required for all state-machine diagrams", nil).
			WithContext("name", name).
			WithContext("location", location.String())
	}

	// Get the file path
	filePath := r.pathManager.GetDiagramFilePathWithDiagramType(name, version, location, diagramType)

	// Check if file exists
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	// Some other error occurred
	return false, models.NewStateMachineError(models.ErrorTypeFileSystem, "failed to check file existence", err).
		WithContext("filePath", filePath)
}

// CreateDirectory creates a directory with proper permissions
func (r *FileSystemRepository) CreateDirectory(path string) error {
	// Validate the path
	if err := r.pathManager.ValidatePath(path); err != nil {
		return err
	}

	// Check if directory already exists
	if _, err := os.Stat(path); err == nil {
		return nil // Directory already exists
	}

	// Create directory with proper permissions (0755 = rwxr-xr-x)
	if err := os.MkdirAll(path, 0755); err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem, "failed to create directory", err).
			WithContext("path", path)
	}

	return nil
}

// DirectoryExists checks if a directory exists
func (r *FileSystemRepository) DirectoryExists(path string) (bool, error) {
	// Validate the path
	if err := r.pathManager.ValidatePath(path); err != nil {
		return false, err
	}

	// Check if directory exists
	info, err := os.Stat(path)
	if err == nil {
		return info.IsDir(), nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	// Some other error occurred
	return false, models.NewStateMachineError(models.ErrorTypeFileSystem, "failed to check directory existence", err).
		WithContext("path", path)
}

// MoveDiagram moves a state-machine diagram from one location to another
func (r *FileSystemRepository) MoveDiagram(diagramType smmodels.DiagramType, name, version string, from, to models.Location) error {
	// Validate inputs
	if err := r.pathManager.ValidateName(name); err != nil {
		return err
	}

	if version == "" {
		return models.NewStateMachineError(models.ErrorTypeValidation, "version is required for move operation", nil).
			WithContext("name", name)
	}

	if from == to {
		return models.NewStateMachineError(models.ErrorTypeValidation, "source and destination locations cannot be the same", nil).
			WithContext("from", from.String()).
			WithContext("to", to.String())
	}

	// Get source and destination file paths
	sourceFilePath := r.pathManager.GetDiagramFilePathWithDiagramType(name, version, from, diagramType)
	destFilePath := r.pathManager.GetDiagramFilePathWithDiagramType(name, version, to, diagramType)

	// Check if source file exists
	if _, err := os.Stat(sourceFilePath); os.IsNotExist(err) {
		return models.NewStateMachineError(models.ErrorTypeFileNotFound, "source state-machine diagram file not found", nil).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", from.String()).
			WithContext("sourceFilePath", sourceFilePath)
	} else if err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem, "failed to check source file existence", err).
			WithContext("sourceFilePath", sourceFilePath)
	}

	// Check if destination file already exists
	if _, err := os.Stat(destFilePath); err == nil {
		return models.NewStateMachineError(models.ErrorTypeFileConflict, "destination file already exists", nil).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", to.String()).
			WithContext("destFilePath", destFilePath)
	}

	// Create destination directory if needed
	destDir := r.pathManager.GetLocationWithDiagramTypePath(to, diagramType)
	if err := r.CreateDirectory(destDir); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Move the file
	if err := os.Rename(sourceFilePath, destFilePath); err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem, "failed to move state-machine diagram file", err).
			WithContext("sourceFilePath", sourceFilePath).
			WithContext("destFilePath", destFilePath)
	}

	return nil
}

// DeleteDiagram deletes a state-machine diagram file
func (r *FileSystemRepository) DeleteDiagram(diagramType smmodels.DiagramType, name, version string, location models.Location) error {
	// Validate inputs
	if err := r.pathManager.ValidateName(name); err != nil {
		return err
	}

	if version == "" {
		return models.NewStateMachineError(models.ErrorTypeValidation, "version is required for all state-machine diagrams", nil).
			WithContext("name", name).
			WithContext("location", location.String())
	}

	// Get the file path
	filePath := r.pathManager.GetDiagramFilePathWithDiagramType(name, version, location, diagramType)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return models.NewStateMachineError(models.ErrorTypeFileNotFound, "state-machine diagram file not found", nil).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", location.String()).
			WithContext("filePath", filePath)
	} else if err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem, "failed to check file existence", err).
			WithContext("filePath", filePath)
	}

	// Remove the file
	if err := os.Remove(filePath); err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem, "failed to delete state-machine diagram file", err).
			WithContext("filePath", filePath)
	}

	return nil
}

// ListStateMachines lists all state-machine diagrams in a location
func (r *FileSystemRepository) ListStateMachines(diagramType smmodels.DiagramType, location models.Location) ([]models.StateMachineDiagram, error) {
	// Get the location path
	locationPath := r.pathManager.GetLocationWithDiagramTypePath(location, diagramType)

	// Check if location directory exists
	exists, err := r.DirectoryExists(locationPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check location directory: %w", err)
	}
	if !exists {
		// Return empty list if location doesn't exist
		return []models.StateMachineDiagram{}, nil
	}

	// Read directory entries
	entries, err := os.ReadDir(locationPath)
	if err != nil {
		return nil, models.NewStateMachineError(models.ErrorTypeFileSystem, "failed to read location directory", err).
			WithContext("locationPath", locationPath)
	}

	var diagrams []models.StateMachineDiagram

	// Process each directory entry
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories, we only want files
		}

		// Parse file name to get name and version
		pathInfo, err := r.pathManager.ParseFileName(entry.Name())
		if err != nil {
			// Skip files that don't match our naming convention
			continue
		}

		// Try to read the state-machine diagram
		diag, err := r.ReadDiagram(diagramType, pathInfo.Name, pathInfo.Version, location)
		if err != nil {
			// Skip state-machine diagrams that can't be read, but continue processing others
			continue
		}

		diagrams = append(diagrams, *diag)
	}

	return diagrams, nil
}
