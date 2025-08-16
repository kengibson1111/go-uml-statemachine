package service

import (
	"sync"
	"time"

	"github.com/kengibson1111/go-uml-statemachine/internal/models"
)

// service implements the StateMachineService interface
type service struct {
	repo      models.Repository
	validator models.Validator
	config    *models.Config
	mu        sync.RWMutex
}

// NewService creates a new StateMachineService with the provided dependencies
func NewService(repo models.Repository, validator models.Validator, config *models.Config) models.StateMachineService {
	if config == nil {
		config = models.DefaultConfig()
	}

	return &service{
		repo:      repo,
		validator: validator,
		config:    config,
	}
}

// NewServiceWithDefaults creates a new StateMachineService with default configuration
func NewServiceWithDefaults(repo models.Repository, validator models.Validator) models.StateMachineService {
	return NewService(repo, validator, models.DefaultConfig())
}

// NewServiceFromEnv creates a new StateMachineService with configuration loaded from environment variables
func NewServiceFromEnv(repo models.Repository, validator models.Validator) models.StateMachineService {
	config := models.LoadConfigFromEnv()
	return NewService(repo, validator, config)
}

// NewServiceWithEnvOverrides creates a new StateMachineService with the provided config merged with environment variables
// Environment variables take precedence over the provided config
func NewServiceWithEnvOverrides(repo models.Repository, validator models.Validator, baseConfig *models.Config) models.StateMachineService {
	if baseConfig == nil {
		baseConfig = models.DefaultConfig()
	}

	// Create a copy of the base config and merge with environment
	config := *baseConfig
	config.MergeWithEnv()

	return NewService(repo, validator, &config)
}

// Create creates a new state machine with the specified parameters
func (s *service) Create(name, version string, content string, location models.Location) (*models.StateMachine, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate input parameters
	if name == "" {
		return nil, models.NewStateMachineError(models.ErrorTypeValidation, "name cannot be empty", nil)
	}
	if version == "" {
		return nil, models.NewStateMachineError(models.ErrorTypeValidation, "version cannot be empty", nil)
	}
	if content == "" {
		return nil, models.NewStateMachineError(models.ErrorTypeValidation, "content cannot be empty", nil)
	}

	// Check if state machine already exists
	exists, err := s.repo.Exists(name, version, location)
	if err != nil {
		return nil, models.NewStateMachineError(models.ErrorTypeFileSystem,
			"failed to check if state machine exists", err).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", location.String())
	}
	if exists {
		return nil, models.NewStateMachineError(models.ErrorTypeDirectoryConflict,
			"state machine already exists", nil).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", location.String())
	}

	// For in-progress state machines, check if there's a conflicting directory in products
	if location == models.LocationInProgress {
		productExists, err := s.repo.Exists(name, version, models.LocationProducts)
		if err != nil {
			return nil, models.NewStateMachineError(models.ErrorTypeFileSystem,
				"failed to check products directory for conflicts", err).
				WithContext("name", name).
				WithContext("version", version)
		}
		if productExists {
			return nil, models.NewStateMachineError(models.ErrorTypeDirectoryConflict,
				"cannot create in-progress state machine: directory with same name exists in products", nil).
				WithContext("name", name).
				WithContext("version", version)
		}
	}

	// Create the state machine object
	sm := &models.StateMachine{
		Name:     name,
		Version:  version,
		Content:  content,
		Location: location,
		Metadata: models.Metadata{
			CreatedAt:  time.Now(),
			ModifiedAt: time.Now(),
		},
	}

	// Write the state machine to disk
	if err := s.repo.WriteStateMachine(sm); err != nil {
		return nil, models.NewStateMachineError(models.ErrorTypeFileSystem,
			"failed to write state machine", err).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", location.String())
	}

	return sm, nil
}

// Read retrieves a state machine by name, version, and location
func (s *service) Read(name, version string, location models.Location) (*models.StateMachine, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input parameters
	if name == "" {
		return nil, models.NewStateMachineError(models.ErrorTypeValidation, "name cannot be empty", nil)
	}
	if version == "" {
		return nil, models.NewStateMachineError(models.ErrorTypeValidation, "version cannot be empty", nil)
	}

	// Read the state machine from repository
	sm, err := s.repo.ReadStateMachine(name, version, location)
	if err != nil {
		return nil, models.NewStateMachineError(models.ErrorTypeFileNotFound,
			"failed to read state machine", err).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", location.String())
	}

	return sm, nil
}

// Update modifies an existing state machine
func (s *service) Update(sm *models.StateMachine) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate input parameters
	if sm == nil {
		return models.NewStateMachineError(models.ErrorTypeValidation, "state machine cannot be nil", nil)
	}
	if sm.Name == "" {
		return models.NewStateMachineError(models.ErrorTypeValidation, "name cannot be empty", nil)
	}
	if sm.Version == "" {
		return models.NewStateMachineError(models.ErrorTypeValidation, "version cannot be empty", nil)
	}
	if sm.Content == "" {
		return models.NewStateMachineError(models.ErrorTypeValidation, "content cannot be empty", nil)
	}

	// Check if state machine exists
	exists, err := s.repo.Exists(sm.Name, sm.Version, sm.Location)
	if err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem,
			"failed to check if state machine exists", err).
			WithContext("name", sm.Name).
			WithContext("version", sm.Version).
			WithContext("location", sm.Location.String())
	}
	if !exists {
		return models.NewStateMachineError(models.ErrorTypeFileNotFound,
			"state machine does not exist", nil).
			WithContext("name", sm.Name).
			WithContext("version", sm.Version).
			WithContext("location", sm.Location.String())
	}

	// Update the modified timestamp
	sm.Metadata.ModifiedAt = time.Now()

	// Write the updated state machine to disk
	if err := s.repo.WriteStateMachine(sm); err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem,
			"failed to update state machine", err).
			WithContext("name", sm.Name).
			WithContext("version", sm.Version).
			WithContext("location", sm.Location.String())
	}

	return nil
}

// Delete removes a state machine by name, version, and location
func (s *service) Delete(name, version string, location models.Location) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate input parameters
	if name == "" {
		return models.NewStateMachineError(models.ErrorTypeValidation, "name cannot be empty", nil)
	}
	if version == "" {
		return models.NewStateMachineError(models.ErrorTypeValidation, "version cannot be empty", nil)
	}

	// Check if state machine exists
	exists, err := s.repo.Exists(name, version, location)
	if err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem,
			"failed to check if state machine exists", err).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", location.String())
	}
	if !exists {
		return models.NewStateMachineError(models.ErrorTypeFileNotFound,
			"state machine does not exist", nil).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", location.String())
	}

	// Delete the state machine from repository
	if err := s.repo.DeleteStateMachine(name, version, location); err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem,
			"failed to delete state machine", err).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", location.String())
	}

	return nil
}

// Promote moves a state machine from in-progress to products
func (s *service) Promote(name, version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate input parameters
	if name == "" {
		return models.NewStateMachineError(models.ErrorTypeValidation, "name cannot be empty", nil)
	}
	if version == "" {
		return models.NewStateMachineError(models.ErrorTypeValidation, "version cannot be empty", nil)
	}

	// Step 1: Check if state machine exists in in-progress
	exists, err := s.repo.Exists(name, version, models.LocationInProgress)
	if err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem,
			"failed to check if state machine exists in in-progress", err).
			WithContext("name", name).
			WithContext("version", version)
	}
	if !exists {
		return models.NewStateMachineError(models.ErrorTypeFileNotFound,
			"state machine does not exist in in-progress", nil).
			WithContext("name", name).
			WithContext("version", version)
	}

	// Step 2: Check if there's already a directory with the same name in products
	productExists, err := s.repo.Exists(name, version, models.LocationProducts)
	if err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem,
			"failed to check products directory for conflicts", err).
			WithContext("name", name).
			WithContext("version", version)
	}
	if productExists {
		return models.NewStateMachineError(models.ErrorTypeDirectoryConflict,
			"cannot promote: directory with same name already exists in products", nil).
			WithContext("name", name).
			WithContext("version", version)
	}

	// Step 3: Read the state machine for validation
	sm, err := s.repo.ReadStateMachine(name, version, models.LocationInProgress)
	if err != nil {
		return models.NewStateMachineError(models.ErrorTypeFileSystem,
			"failed to read state machine for validation", err).
			WithContext("name", name).
			WithContext("version", version)
	}

	// Step 4: Validate the state machine with in-progress strictness (errors and warnings)
	validationResult, err := s.validator.Validate(sm, models.StrictnessInProgress)
	if err != nil {
		return models.NewStateMachineError(models.ErrorTypeValidation,
			"failed to validate state machine", err).
			WithContext("name", name).
			WithContext("version", version)
	}

	// Step 5: Check if validation passed (no errors allowed for promotion)
	if !validationResult.IsValid || validationResult.HasErrors() {
		return models.NewStateMachineError(models.ErrorTypeValidation,
			"state machine validation failed: cannot promote with validation errors", nil).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("errors", len(validationResult.Errors)).
			WithContext("warnings", len(validationResult.Warnings))
	}

	// Step 6: Perform atomic move operation with rollback capability
	err = s.performAtomicPromotion(name, version)
	if err != nil {
		return err
	}

	return nil
}

// performAtomicPromotion performs the actual move operation with rollback capability
func (s *service) performAtomicPromotion(name, version string) error {
	// Step 1: Attempt to move the state machine
	err := s.repo.MoveStateMachine(name, version, models.LocationInProgress, models.LocationProducts)
	if err != nil {
		// If move fails, no rollback needed since nothing was changed
		return models.NewStateMachineError(models.ErrorTypeFileSystem,
			"failed to move state machine to products", err).
			WithContext("name", name).
			WithContext("version", version)
	}

	// Step 2: Verify the move was successful by checking both locations
	// Check that it exists in products
	productExists, err := s.repo.Exists(name, version, models.LocationProducts)
	if err != nil {
		// Attempt rollback - move back to in-progress
		s.attemptRollback(name, version)
		return models.NewStateMachineError(models.ErrorTypeFileSystem,
			"failed to verify promotion: cannot check products directory", err).
			WithContext("name", name).
			WithContext("version", version)
	}
	if !productExists {
		// Attempt rollback - move back to in-progress
		s.attemptRollback(name, version)
		return models.NewStateMachineError(models.ErrorTypeFileSystem,
			"promotion verification failed: state machine not found in products after move", nil).
			WithContext("name", name).
			WithContext("version", version)
	}

	// Check that it no longer exists in in-progress
	inProgressExists, err := s.repo.Exists(name, version, models.LocationInProgress)
	if err != nil {
		// Attempt rollback - move back to in-progress
		s.attemptRollback(name, version)
		return models.NewStateMachineError(models.ErrorTypeFileSystem,
			"failed to verify promotion: cannot check in-progress directory", err).
			WithContext("name", name).
			WithContext("version", version)
	}
	if inProgressExists {
		// Attempt rollback - move back to in-progress (though it's already there)
		// This indicates a partial failure where the move didn't complete properly
		return models.NewStateMachineError(models.ErrorTypeFileSystem,
			"promotion verification failed: state machine still exists in in-progress after move", nil).
			WithContext("name", name).
			WithContext("version", version)
	}

	return nil
}

// attemptRollback attempts to rollback a failed promotion by moving the state machine back to in-progress
func (s *service) attemptRollback(name, version string) {
	// This is a best-effort rollback - we don't return errors from here
	// as we're already in an error state
	rollbackErr := s.repo.MoveStateMachine(name, version, models.LocationProducts, models.LocationInProgress)
	if rollbackErr != nil {
		// Log the rollback failure but don't return it - the original error is more important
		// In a real implementation, this would be logged to a proper logging system
		// For now, we'll just ignore the rollback error
	}
}

// Validate validates a state machine with the specified strictness level
func (s *service) Validate(name, version string, location models.Location) (*models.ValidationResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input parameters
	if name == "" {
		return nil, models.NewStateMachineError(models.ErrorTypeValidation, "name cannot be empty", nil)
	}
	if version == "" {
		return nil, models.NewStateMachineError(models.ErrorTypeValidation, "version cannot be empty", nil)
	}

	// Read the state machine from repository
	sm, err := s.repo.ReadStateMachine(name, version, location)
	if err != nil {
		return nil, models.NewStateMachineError(models.ErrorTypeFileNotFound,
			"failed to read state machine for validation", err).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", location.String())
	}

	// Determine validation strictness based on location
	strictness := models.StrictnessInProgress
	if location == models.LocationProducts {
		strictness = models.StrictnessProducts
	}

	// Validate the state machine using the validator
	validationResult, err := s.validator.Validate(sm, strictness)
	if err != nil {
		return nil, models.NewStateMachineError(models.ErrorTypeValidation,
			"validation failed", err).
			WithContext("name", name).
			WithContext("version", version).
			WithContext("location", location.String()).
			WithContext("strictness", strictness.String())
	}

	return validationResult, nil
}

// ListAll lists all state machines in the specified location
func (s *service) ListAll(location models.Location) ([]models.StateMachine, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Use the repository to list all state machines in the specified location
	stateMachines, err := s.repo.ListStateMachines(location)
	if err != nil {
		return nil, models.NewStateMachineError(models.ErrorTypeFileSystem,
			"failed to list state machines", err).
			WithContext("location", location.String())
	}

	return stateMachines, nil
}

// ResolveReferences resolves all references in a state machine
func (s *service) ResolveReferences(sm *models.StateMachine) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Validate input parameters
	if sm == nil {
		return models.NewStateMachineError(models.ErrorTypeValidation, "state machine cannot be nil", nil)
	}

	// First, parse references from the content if not already done
	if len(sm.References) == 0 {
		// Use the validator to parse references from content
		validationResult, err := s.validator.ValidateReferences(sm)
		if err != nil {
			return models.NewStateMachineError(models.ErrorTypeValidation,
				"failed to parse references from content", err)
		}

		// Check if reference parsing had critical errors
		if validationResult.HasErrors() {
			// Find the first critical error related to reference parsing
			for _, validationErr := range validationResult.Errors {
				if validationErr.Code == "REFERENCE_PARSE_ERROR" {
					return models.NewStateMachineError(models.ErrorTypeValidation,
						"failed to parse references: "+validationErr.Message, nil)
				}
			}
		}
	}

	// If there are still no references after parsing, nothing to resolve
	if len(sm.References) == 0 {
		return nil
	}

	// Resolve each reference
	for i := range sm.References {
		err := s.resolveReference(sm, &sm.References[i])
		if err != nil {
			return err // Return the original error from resolveReference
		}
	}

	return nil
}

// resolveReference resolves a single reference within a state machine
func (s *service) resolveReference(sm *models.StateMachine, ref *models.Reference) error {
	switch ref.Type {
	case models.ReferenceTypeProduct:
		// For product references, check if the referenced state machine exists in products
		exists, err := s.repo.Exists(ref.Name, ref.Version, models.LocationProducts)
		if err != nil {
			return models.NewStateMachineError(models.ErrorTypeFileSystem,
				"failed to check product reference existence", err).
				WithContext("reference_name", ref.Name).
				WithContext("reference_version", ref.Version)
		}
		if !exists {
			return models.NewStateMachineError(models.ErrorTypeReferenceResolution,
				"product reference not found", nil).
				WithContext("reference_name", ref.Name).
				WithContext("reference_version", ref.Version)
		}

		// Set the resolved path for the reference
		ref.Path = s.buildProductReferencePath(ref.Name, ref.Version)

	case models.ReferenceTypeNested:
		// For nested references, check if the referenced state machine exists as a nested item
		// within the same parent directory as the current state machine
		nestedPath := s.buildNestedReferencePath(sm, ref.Name)

		// Check if the nested reference exists by attempting to read it
		// Note: For nested references, we don't use version in the path
		exists, err := s.checkNestedReferenceExists(sm, ref.Name)
		if err != nil {
			return models.NewStateMachineError(models.ErrorTypeFileSystem,
				"failed to check nested reference existence", err).
				WithContext("reference_name", ref.Name).
				WithContext("parent_state_machine", sm.Name)
		}
		if !exists {
			return models.NewStateMachineError(models.ErrorTypeReferenceResolution,
				"nested reference not found", nil).
				WithContext("reference_name", ref.Name).
				WithContext("parent_state_machine", sm.Name)
		}

		// Set the resolved path for the reference
		ref.Path = nestedPath

	default:
		return models.NewStateMachineError(models.ErrorTypeReferenceResolution,
			"unknown reference type", nil).
			WithContext("reference_type", ref.Type.String())
	}

	return nil
}

// buildProductReferencePath builds the path for a product reference
func (s *service) buildProductReferencePath(name, version string) string {
	// Product references are in the format: products/{name}-{version}/{name}-{version}.puml
	return "products\\" + name + "-" + version + "\\" + name + "-" + version + ".puml"
}

// buildNestedReferencePath builds the path for a nested reference
func (s *service) buildNestedReferencePath(sm *models.StateMachine, refName string) string {
	// Nested references are in the format: {location}/{parent-name}-{parent-version}/nested/{ref-name}/{ref-name}.puml
	locationStr := sm.Location.String()
	return locationStr + "\\" + sm.Name + "-" + sm.Version + "\\nested\\" + refName + "\\" + refName + ".puml"
}

// checkNestedReferenceExists checks if a nested reference exists within the parent state machine's directory
func (s *service) checkNestedReferenceExists(sm *models.StateMachine, refName string) (bool, error) {
	// For nested references, we need to check if the nested directory and file exist
	// This is a simplified check - in a real implementation, we might need more sophisticated
	// directory traversal logic depending on how the repository is implemented

	// We'll use the repository's directory checking capabilities
	nestedDirPath := s.buildNestedDirectoryPath(sm, refName)

	// Check if the nested directory exists
	exists, err := s.repo.DirectoryExists(nestedDirPath)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// buildNestedDirectoryPath builds the directory path for a nested reference
func (s *service) buildNestedDirectoryPath(sm *models.StateMachine, refName string) string {
	// Nested directory path: {location}/{parent-name}-{parent-version}/nested/{ref-name}
	locationStr := sm.Location.String()
	return locationStr + "\\" + sm.Name + "-" + sm.Version + "\\nested\\" + refName
}
