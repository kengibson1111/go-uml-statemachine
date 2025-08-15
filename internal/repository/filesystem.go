package repository

import (
	"fmt"
	"os"

	"github.com/kengibson1111/go-uml-statemachine/internal/models"
)

// FileSystemRepository implements the Repository interface using the file system
type FileSystemRepository struct {
	pathManager *models.PathManager
	config      *models.Config
}

// NewFileSystemRepository creates a new FileSystemRepository
func NewFileSystemRepository(config *models.Config) *FileSystemRepository {
	if config == nil {
		config = models.DefaultConfig()
	}

	pathManager := models.NewPathManager(config.RootDirectory)

	return &FileSystemRepository{
		pathManager: pathManager,
		config:      config,
	}
}

// ReadStateMachine reads a state machine from the file system
func (r *FileSystemRepository) ReadStateMachine(name, version string, location models.Location) (*models.StateMachine, error) {
	// Validate inputs
	if err := r.pathManager.ValidateName(name); err != nil {
		return nil, err
	}

	if location != models.LocationNested && version == "" {
		return nil, models.NewStateMachineError(models.ErrorTypeValidation, "version is required for non-nested state machines", nil).
			WithContext("name", name).
			WithContext("location", location.String())
	}

	// Get the file path
	filePath := r.pathManager.GetStateMachineFilePath(name, version, location)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, models.NewStateMachineError(models.ErrorTypeFileNotFound, "state machine file not found", err).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", location.String()).
			WithContext("filePath", filePath)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, models.NewStateMachineError(models.ErrorTypeFileSystem, "failed to read state machine file", err).
			WithContext("filePath", filePath)
	}

	// Check file size
	if int64(len(content)) > r.config.MaxFileSize {
		return nil, models.NewStateMachineError(models.ErrorTypeFileSystem, "file size exceeds maximum allowed", nil).
			WithContext("fileSize", len(content)).
			WithContext("maxSize", r.config.MaxFileSize).
			WithContext("filePath", filePath)
	}

	// Get file info for metadata
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, models.NewStateMachineError(models.ErrorTypeFileSystem, "failed to get file info", err).
			WithContext("filePath", filePath)
	}

	// Create StateMachine object
	stateMachine := &models.StateMachine{
		Name:     name,
		Version:  version,
		Content:  string(content),
		Location: location,
		Metadata: models.Metadata{
			ModifiedAt: fileInfo.ModTime(),
			// CreatedAt and Author would need additional metadata storage
			CreatedAt: fileInfo.ModTime(), // Using ModTime as fallback
		},
	}

	// TODO: Parse references from content (will be implemented in validation layer)
	stateMachine.References = []models.Reference{}

	return stateMachine, nil
}

// WriteStateMachine writes a state machine to the file system
func (r *FileSystemRepository) WriteStateMachine(sm *models.StateMachine) error {
	if sm == nil {
		return models.NewStateMachineError(models.ErrorTypeValidation, "state machine cannot be nil", nil)
	}

	// Validate inputs
	if err := r.pathManager.ValidateName(sm.Name); err != nil {
		return err
	}

	if sm.Location != models.LocationNested && sm.Version == "" {
		return models.NewStateMachineError(models.ErrorTypeValidation, "version is required for non-nested state machines", nil).
			WithContext("name", sm.Name).
			WithContext("location", sm.Location.String())
	}

	// Check content size
	if int64(len(sm.Content)) > r.config.MaxFileSize {
		return models.NewStateMachineError(models.ErrorTypeFileSystem, "content size exceeds maximum allowed", nil).
			WithContext("contentSize", len(sm.Content)).
			WithContext("maxSize", r.config.MaxFileSize)
	}

	// Get directory and file paths
	dirPath := r.pathManager.GetStateMachineDirectoryPath(sm.Name, sm.Version, sm.Location)
	filePath := r.pathManager.GetStateMachineFilePath(sm.Name, sm.Version, sm.Location)

	// Create directory if it doesn't exist
	if err := r.CreateDirectory(dirPath); err != nil {
		return fmt.Errorf("failed to create directory for state machine: %w", err)
	}

	// Write the file
	if err := os.WriteFile(filePath, []byte(sm.Content), 0644); err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem, "failed to write state machine file", err).
			WithContext("filePath", filePath)
	}

	return nil
}

// Exists checks if a state machine exists
func (r *FileSystemRepository) Exists(name, version string, location models.Location) (bool, error) {
	// Validate inputs
	if err := r.pathManager.ValidateName(name); err != nil {
		return false, err
	}

	if location != models.LocationNested && version == "" {
		return false, models.NewStateMachineError(models.ErrorTypeValidation, "version is required for non-nested state machines", nil).
			WithContext("name", name).
			WithContext("location", location.String())
	}

	// Get the file path
	filePath := r.pathManager.GetStateMachineFilePath(name, version, location)

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

// MoveStateMachine moves a state machine from one location to another
func (r *FileSystemRepository) MoveStateMachine(name, version string, from, to models.Location) error {
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
	sourcePath := r.pathManager.GetStateMachineDirectoryPath(name, version, from)
	destPath := r.pathManager.GetStateMachineDirectoryPath(name, version, to)

	// Check if source exists
	sourceExists, err := r.DirectoryExists(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to check source directory: %w", err)
	}
	if !sourceExists {
		return models.NewStateMachineError(models.ErrorTypeFileNotFound, "source state machine directory not found", nil).
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
	destParent := r.pathManager.GetLocationPath(to)
	if err := r.CreateDirectory(destParent); err != nil {
		return fmt.Errorf("failed to create destination parent directory: %w", err)
	}

	// Move the directory
	if err := os.Rename(sourcePath, destPath); err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem, "failed to move state machine directory", err).
			WithContext("sourcePath", sourcePath).
			WithContext("destPath", destPath)
	}

	return nil
}

// DeleteStateMachine deletes a state machine and its directory
func (r *FileSystemRepository) DeleteStateMachine(name, version string, location models.Location) error {
	// Validate inputs
	if err := r.pathManager.ValidateName(name); err != nil {
		return err
	}

	if location != models.LocationNested && version == "" {
		return models.NewStateMachineError(models.ErrorTypeValidation, "version is required for non-nested state machines", nil).
			WithContext("name", name).
			WithContext("location", location.String())
	}

	// Get the directory path
	dirPath := r.pathManager.GetStateMachineDirectoryPath(name, version, location)

	// Check if directory exists
	exists, err := r.DirectoryExists(dirPath)
	if err != nil {
		return fmt.Errorf("failed to check directory existence: %w", err)
	}
	if !exists {
		return models.NewStateMachineError(models.ErrorTypeFileNotFound, "state machine directory not found", nil).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", location.String()).
			WithContext("dirPath", dirPath)
	}

	// Remove the entire directory
	if err := os.RemoveAll(dirPath); err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem, "failed to delete state machine directory", err).
			WithContext("dirPath", dirPath)
	}

	return nil
}

// ListStateMachines lists all state machines in a location
func (r *FileSystemRepository) ListStateMachines(location models.Location) ([]models.StateMachine, error) {
	// Get the location path
	locationPath := r.pathManager.GetLocationPath(location)

	// Check if location directory exists
	exists, err := r.DirectoryExists(locationPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check location directory: %w", err)
	}
	if !exists {
		// Return empty list if location doesn't exist
		return []models.StateMachine{}, nil
	}

	// Read directory entries
	entries, err := os.ReadDir(locationPath)
	if err != nil {
		return nil, models.NewStateMachineError(models.ErrorTypeFileSystem, "failed to read location directory", err).
			WithContext("locationPath", locationPath)
	}

	var stateMachines []models.StateMachine

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

		// Try to read the state machine
		sm, err := r.ReadStateMachine(pathInfo.Name, pathInfo.Version, location)
		if err != nil {
			// Skip state machines that can't be read, but continue processing others
			continue
		}

		stateMachines = append(stateMachines, *sm)
	}

	return stateMachines, nil
}
