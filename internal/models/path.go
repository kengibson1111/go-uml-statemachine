package models

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	smmodels "github.com/kengibson1111/go-uml-statemachine-models/models"
)

const (
	// RootDirectoryName is the top-level directory for state-machine diagrams
	RootDirectoryName = ".go-uml-statemachine-parsers"

	// PlantUMLExtension is the file extension for PlantUML files
	PlantUMLExtension = ".puml"
)

// PathManager provides utilities for managing directory paths and file names
type PathManager struct {
	rootDir string
}

// NewPathManager creates a new PathManager with the specified root directory
func NewPathManager(rootDir string) *PathManager {
	if rootDir == "" {
		rootDir = RootDirectoryName
	}
	return &PathManager{rootDir: rootDir}
}

// GetRootPath returns the root directory path
func (pm *PathManager) GetRootPath() string {
	return pm.rootDir
}

// GetLocationPath returns the path for a specific location (in-progress or products)
func (pm *PathManager) GetLocationPath(location Location) string {
	switch location {
	case LocationInProgress:
		return filepath.Join(pm.rootDir, "in-progress")
	case LocationProducts:
		return filepath.Join(pm.rootDir, "products")
	default:
		return pm.rootDir
	}
}

// GetLocationWithDiagramTypePath returns the path for a specific location and diagram type
func (pm *PathManager) GetLocationWithDiagramTypePath(location Location, diagramType smmodels.DiagramType) string {
	locationPath := pm.GetLocationPath(location)
	return filepath.Join(locationPath, diagramType.String())
}

// GetDiagramFilePathWithDiagramType returns the full file path for a state-machine diagram with diagram type
func (pm *PathManager) GetDiagramFilePathWithDiagramType(name, version string, location Location, diagramType smmodels.DiagramType) string {
	dirPath := pm.GetLocationWithDiagramTypePath(location, diagramType)
	return filepath.Join(dirPath, fmt.Sprintf("%s-%s%s", name, version, PlantUMLExtension))
}

// PathInfo contains parsed information from a path
type PathInfo struct {
	Name     string
	Version  string
	Location Location
}

// ValidatePath validates a path to prevent directory traversal attacks
func (pm *PathManager) ValidatePath(path string) error {
	// Check for directory traversal attempts before cleaning
	if strings.Contains(path, "..") {
		return NewStateMachineError(ErrorTypeValidation, "path contains directory traversal", nil).
			WithContext("path", path)
	}

	// Clean the path to resolve any . components
	cleanPath := filepath.Clean(path)

	// Ensure the path is within the root directory
	if !strings.HasPrefix(cleanPath, pm.rootDir) && !filepath.IsAbs(cleanPath) {
		// For relative paths, make them absolute relative to root
		cleanPath = filepath.Join(pm.rootDir, cleanPath)
	}

	// Additional validation for absolute paths
	if filepath.IsAbs(path) {
		rootAbs, err := filepath.Abs(pm.rootDir)
		if err != nil {
			return NewStateMachineError(ErrorTypeFileSystem, "failed to resolve root directory", err).
				WithContext("rootDir", pm.rootDir)
		}

		pathAbs, err := filepath.Abs(cleanPath)
		if err != nil {
			return NewStateMachineError(ErrorTypeFileSystem, "failed to resolve path", err).
				WithContext("path", path)
		}

		// Check if the absolute path is within the root directory
		relPath, err := filepath.Rel(rootAbs, pathAbs)
		if err != nil || strings.HasPrefix(relPath, "..") {
			return NewStateMachineError(ErrorTypeValidation, "path is outside root directory", nil).
				WithContext("path", path).
				WithContext("rootDir", pm.rootDir)
		}
	}

	return nil
}

// ValidateName validates a state-machine diagram name
func (pm *PathManager) ValidateName(name string) error {
	if name == "" {
		return NewStateMachineError(ErrorTypeValidation, "name cannot be empty", nil)
	}

	// Check for invalid characters in names
	validNameRegex := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
	if !validNameRegex.MatchString(name) {
		return NewStateMachineError(ErrorTypeValidation, "name contains invalid characters", nil).
			WithContext("name", name).
			WithContext("validFormat", "must start with alphanumeric and contain only alphanumeric, underscore, or hyphen")
	}

	// Check for reserved names
	reservedNames := []string{"nested", "in-progress", "products", ".", "..", "CON", "PRN", "AUX", "NUL"}
	for _, reserved := range reservedNames {
		if strings.EqualFold(name, reserved) {
			return NewStateMachineError(ErrorTypeValidation, "name is reserved", nil).
				WithContext("name", name)
		}
	}

	// Check length limits
	if len(name) > 100 {
		return NewStateMachineError(ErrorTypeValidation, "name is too long", nil).
			WithContext("name", name).
			WithContext("maxLength", 100)
	}

	return nil
}

// ParseDirectoryName parses a directory name to extract name and version
func (pm *PathManager) ParseDirectoryName(dirName string) (*PathInfo, error) {
	if dirName == "" {
		return nil, NewStateMachineError(ErrorTypeValidation, "directory name cannot be empty", nil)
	}

	// All directories must have version format (name-version)
	if !strings.Contains(dirName, "-") {
		return nil, NewStateMachineError(ErrorTypeValidation, "invalid directory name format", nil).
			WithContext("dirName", dirName).
			WithContext("expectedFormat", "name-version")
	}

	// Try to find a valid version by working backwards through possible splits
	parts := strings.Split(dirName, "-")
	if len(parts) < 2 {
		return nil, NewStateMachineError(ErrorTypeValidation, "invalid directory name format", nil).
			WithContext("dirName", dirName).
			WithContext("expectedFormat", "name-version")
	}

	// Try different split points to handle pre-release versions
	for i := len(parts) - 1; i >= 1; i-- {
		name := strings.Join(parts[:i], "-")
		version := strings.Join(parts[i:], "-")

		// Try to parse as version
		if _, err := ParseVersion(version); err == nil {
			// Valid version found, validate the name
			if err := pm.ValidateName(name); err != nil {
				return nil, err
			}

			return &PathInfo{
				Name:    name,
				Version: version,
			}, nil
		}
	}

	// No valid version found
	return nil, NewStateMachineError(ErrorTypeValidation, "invalid directory name format", nil).
		WithContext("dirName", dirName).
		WithContext("expectedFormat", "name-version")
}

// ParseFileName parses a PlantUML file name to extract name and version
func (pm *PathManager) ParseFileName(fileName string) (*PathInfo, error) {
	if fileName == "" {
		return nil, NewStateMachineError(ErrorTypeValidation, "file name cannot be empty", nil)
	}

	// Remove the .puml extension
	if !strings.HasSuffix(fileName, PlantUMLExtension) {
		return nil, NewStateMachineError(ErrorTypeValidation, "file must have .puml extension", nil).
			WithContext("fileName", fileName).
			WithContext("expectedExtension", PlantUMLExtension)
	}

	baseName := strings.TrimSuffix(fileName, PlantUMLExtension)

	// All files must have version format (name-version)
	if !strings.Contains(baseName, "-") {
		return nil, NewStateMachineError(ErrorTypeValidation, "invalid file name format", nil).
			WithContext("fileName", fileName).
			WithContext("expectedFormat", "name-version.puml")
	}

	// Try to find a valid version by working backwards through possible splits
	parts := strings.Split(baseName, "-")
	if len(parts) < 2 {
		return nil, NewStateMachineError(ErrorTypeValidation, "invalid file name format", nil).
			WithContext("fileName", fileName).
			WithContext("expectedFormat", "name-version.puml")
	}

	// Try different split points to handle pre-release versions
	for i := len(parts) - 1; i >= 1; i-- {
		name := strings.Join(parts[:i], "-")
		version := strings.Join(parts[i:], "-")

		// Try to parse as version
		if _, err := ParseVersion(version); err == nil {
			// Valid version found, validate the name
			if err := pm.ValidateName(name); err != nil {
				return nil, err
			}

			return &PathInfo{
				Name:    name,
				Version: version,
			}, nil
		}
	}

	// No valid version found
	return nil, NewStateMachineError(ErrorTypeValidation, "invalid file name format", nil).
		WithContext("fileName", fileName).
		WithContext("expectedFormat", "name-version.puml")
}

// ParseFullPath parses a full path to extract location and path information
func (pm *PathManager) ParseFullPath(fullPath string) (*PathInfo, error) {
	// Clean and validate the path
	cleanPath := filepath.Clean(fullPath)
	if err := pm.ValidatePath(cleanPath); err != nil {
		return nil, err
	}

	// Make path relative to root directory
	relPath, err := filepath.Rel(pm.rootDir, cleanPath)
	if err != nil {
		return nil, NewStateMachineError(ErrorTypeFileSystem, "failed to make path relative", err).
			WithContext("path", fullPath).
			WithContext("rootDir", pm.rootDir)
	}

	// Split the path into components
	pathParts := strings.Split(filepath.ToSlash(relPath), "/")
	if len(pathParts) < 3 {
		return nil, NewStateMachineError(ErrorTypeValidation, "path is too short", nil).
			WithContext("path", fullPath)
	}

	// Determine location
	var location Location
	switch pathParts[0] {
	case "in-progress":
		location = LocationInProgress
	case "products":
		location = LocationProducts
	default:
		return nil, NewStateMachineError(ErrorTypeValidation, "invalid location in path", nil).
			WithContext("path", fullPath).
			WithContext("location", pathParts[0])
	}

	// Verify diagram type directory (pathParts[1] should be "puml")
	if len(pathParts) < 3 {
		return nil, NewStateMachineError(ErrorTypeValidation, "missing diagram file", nil).
			WithContext("path", fullPath)
	}

	// Parse the file name to get name and version
	pathInfo, err := pm.ParseFileName(pathParts[2])
	if err != nil {
		return nil, err
	}

	pathInfo.Location = location

	return pathInfo, nil
}
