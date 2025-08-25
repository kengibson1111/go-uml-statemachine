package repository

import (
	"fmt"
	"os"

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

// ReadStateMachine reads a state-machine diagram from the file system
func (r *FileSystemRepository) ReadStateMachine(fileType models.FileType, name, version string, location models.Location) (*models.StateMachineDiagram, error) {
	// Create operation logger with context
	opLogger := r.logger.WithFields(map[string]interface{}{
		"operation": "ReadStateMachine",
		"fileType":  fileType.String(),
		"name":      name,
		"version":   version,
		"location":  location.String(),
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

	if location != models.LocationNested && version == "" {
		err := models.NewStateMachineError(models.ErrorTypeValidation, "version is required for non-nested state-machine diagrams", nil).
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
	filePath := r.pathManager.GetStateMachineFilePathWithFileType(name, version, location, fileType)
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

	// Create StateMachine object
	opLogger.Debug("Creating state-machine diagram object")
	diag := &models.StateMachineDiagram{
		Name:     name,
		Version:  version,
		Content:  string(content),
		Location: location,
		FileType: fileType,
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

// WriteStateMachine writes a state-machine diagram to the file system
func (r *FileSystemRepository) WriteStateMachine(diag *models.StateMachineDiagram) error {
	if diag == nil {
		return models.NewStateMachineError(models.ErrorTypeValidation, "state-machine diagram cannot be nil", nil)
	}

	// Validate inputs
	if err := r.pathManager.ValidateName(diag.Name); err != nil {
		return err
	}

	if diag.Location != models.LocationNested && diag.Version == "" {
		return models.NewStateMachineError(models.ErrorTypeValidation, "version is required for non-nested state-machine diagrams", nil).
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
	dirPath := r.pathManager.GetStateMachineDirectoryPathWithFileType(diag.Name, diag.Version, diag.Location, diag.FileType)
	filePath := r.pathManager.GetStateMachineFilePathWithFileType(diag.Name, diag.Version, diag.Location, diag.FileType)

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
func (r *FileSystemRepository) Exists(fileType models.FileType, name, version string, location models.Location) (bool, error) {
	// Validate inputs
	if err := r.pathManager.ValidateName(name); err != nil {
		return false, err
	}

	if location != models.LocationNested && version == "" {
		return false, models.NewStateMachineError(models.ErrorTypeValidation, "version is required for non-nested state-machine diagrams", nil).
			WithContext("name", name).
			WithContext("location", location.String())
	}

	// Get the file path
	filePath := r.pathManager.GetStateMachineFilePathWithFileType(name, version, location, fileType)

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

// MoveStateMachine moves a state-machine diagram from one location to another
func (r *FileSystemRepository) MoveStateMachine(fileType models.FileType, name, version string, from, to models.Location) error {
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

	// Get source and destination paths
	sourcePath := r.pathManager.GetStateMachineDirectoryPathWithFileType(name, version, from, fileType)
	destPath := r.pathManager.GetStateMachineDirectoryPathWithFileType(name, version, to, fileType)

	// Check if source exists
	sourceExists, err := r.DirectoryExists(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to check source directory: %w", err)
	}
	if !sourceExists {
		return models.NewStateMachineError(models.ErrorTypeFileNotFound, "source state-machine diagram directory not found", nil).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", from.String()).
			WithContext("sourcePath", sourcePath)
	}

	// Check if destination already exists
	destExists, err := r.DirectoryExists(destPath)
	if err != nil {
		return fmt.Errorf("failed to check destination directory: %w", err)
	}
	if destExists {
		return models.NewStateMachineError(models.ErrorTypeDirectoryConflict, "destination directory already exists", nil).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", to.String()).
			WithContext("destPath", destPath)
	}

	// Create destination parent directory if needed
	destParent := r.pathManager.GetLocationWithFileTypePath(to, fileType)
	if err := r.CreateDirectory(destParent); err != nil {
		return fmt.Errorf("failed to create destination parent directory: %w", err)
	}

	// Move the directory
	if err := os.Rename(sourcePath, destPath); err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem, "failed to move state-machine diagram directory", err).
			WithContext("sourcePath", sourcePath).
			WithContext("destPath", destPath)
	}

	return nil
}

// DeleteStateMachine deletes a state-machine diagram and its directory
func (r *FileSystemRepository) DeleteStateMachine(fileType models.FileType, name, version string, location models.Location) error {
	// Validate inputs
	if err := r.pathManager.ValidateName(name); err != nil {
		return err
	}

	if location != models.LocationNested && version == "" {
		return models.NewStateMachineError(models.ErrorTypeValidation, "version is required for non-nested state-machine diagrams", nil).
			WithContext("name", name).
			WithContext("location", location.String())
	}

	// Get the directory path
	dirPath := r.pathManager.GetStateMachineDirectoryPathWithFileType(name, version, location, fileType)

	// Check if directory exists
	exists, err := r.DirectoryExists(dirPath)
	if err != nil {
		return fmt.Errorf("failed to check directory existence: %w", err)
	}
	if !exists {
		return models.NewStateMachineError(models.ErrorTypeFileNotFound, "state-machine diagram directory not found", nil).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", location.String()).
			WithContext("dirPath", dirPath)
	}

	// Remove the entire directory
	if err := os.RemoveAll(dirPath); err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem, "failed to delete state-machine diagram directory", err).
			WithContext("dirPath", dirPath)
	}

	return nil
}

// ListStateMachines lists all state-machine diagrams in a location
func (r *FileSystemRepository) ListStateMachines(fileType models.FileType, location models.Location) ([]models.StateMachineDiagram, error) {
	// Get the location path
	locationPath := r.pathManager.GetLocationWithFileTypePath(location, fileType)

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
		if !entry.IsDir() {
			continue // Skip files, we only want directories
		}

		// Parse directory name to get name and version
		pathInfo, err := r.pathManager.ParseDirectoryName(entry.Name())
		if err != nil {
			// Skip directories that don't match our naming convention
			continue
		}

		// Skip nested directories (they should be handled separately)
		if pathInfo.IsNested {
			continue
		}

		// Try to read the state-machine diagram
		diag, err := r.ReadStateMachine(fileType, pathInfo.Name, pathInfo.Version, location)
		if err != nil {
			// Skip state-machine diagrams that can't be read, but continue processing others
			continue
		}

		diagrams = append(diagrams, *diag)
	}

	return diagrams, nil
}
