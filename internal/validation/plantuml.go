package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kengibson1111/go-uml-statemachine/internal/logging"
	"github.com/kengibson1111/go-uml-statemachine/internal/models"
)

// PlantUMLValidator implements the Validator interface for PlantUML syntax validation
type PlantUMLValidator struct {
	repository models.Repository // Optional repository for reference resolution
	logger     *logging.Logger
}

// NewPlantUMLValidator creates a new PlantUML validator instance
func NewPlantUMLValidator() *PlantUMLValidator {
	logger := logging.NewDefaultLogger().WithField("component", "PlantUMLValidator")
	return &PlantUMLValidator{
		logger: logger,
	}
}

// NewPlantUMLValidatorWithRepository creates a new PlantUML validator instance with repository for reference resolution
func NewPlantUMLValidatorWithRepository(repo models.Repository) *PlantUMLValidator {
	logger := logging.NewDefaultLogger().WithField("component", "PlantUMLValidator")
	return &PlantUMLValidator{
		repository: repo,
		logger:     logger,
	}
}

// Validate validates a state machine according to the specified strictness level
func (v *PlantUMLValidator) Validate(sm *models.StateMachine, strictness models.ValidationStrictness) (*models.ValidationResult, error) {
	result := &models.ValidationResult{
		Errors:   []models.ValidationError{},
		Warnings: []models.ValidationWarning{},
		IsValid:  true,
	}

	// Validate PlantUML structure
	v.validatePlantUMLStructure(sm.Content, result)

	// Validate state machine syntax
	v.validateStateMachineSyntax(sm.Content, result)

	// Apply strictness filtering
	v.applyStrictnessFiltering(result, strictness)

	return result, nil
}

// ValidateReferences validates references in a state machine
func (v *PlantUMLValidator) ValidateReferences(sm *models.StateMachine) (*models.ValidationResult, error) {
	result := &models.ValidationResult{
		Errors:   []models.ValidationError{},
		Warnings: []models.ValidationWarning{},
		IsValid:  true,
	}

	// Parse references from PlantUML content
	references, err := v.parseReferences(sm.Content)
	if err != nil {
		result.AddError("REFERENCE_PARSE_ERROR", "Failed to parse references from content", 1, 1)
		return result, nil
	}

	// Update the state machine with parsed references
	sm.References = references

	// Validate each reference
	for _, ref := range references {
		v.validateReference(ref, sm, result)
	}

	return result, nil
}

// parseReferences extracts references from PlantUML content
func (v *PlantUMLValidator) parseReferences(content string) ([]models.Reference, error) {
	var references []models.Reference
	lines := strings.Split(content, "\n")

	// Regular expressions for different reference patterns
	// Product references: !include products/{name}-{version}/{name}-{version}.puml
	productRefRegex := regexp.MustCompile(`!include\s+products/([a-zA-Z_][a-zA-Z0-9_-]*)-([a-zA-Z0-9_.-]+)/([a-zA-Z_][a-zA-Z0-9_-]*)-([a-zA-Z0-9_.-]+)\.puml`)

	// Nested references: !include nested/{name}/{name}.puml
	nestedRefRegex := regexp.MustCompile(`!include\s+nested/([a-zA-Z_][a-zA-Z0-9_-]*)/([a-zA-Z_][a-zA-Z0-9_-]*)\.puml`)

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check for product references
		if matches := productRefRegex.FindStringSubmatch(trimmedLine); matches != nil {
			if len(matches) >= 5 {
				dirName := matches[1]
				dirVersion := matches[2]
				fileName := matches[3]
				fileVersion := matches[4]

				// Validate that directory name matches file name and versions match
				if dirName == fileName && dirVersion == fileVersion {
					reference := models.Reference{
						Name:    dirName,
						Version: dirVersion,
						Type:    models.ReferenceTypeProduct,
						Path:    fmt.Sprintf("products/%s-%s/%s-%s.puml", dirName, dirVersion, fileName, fileVersion),
					}
					references = append(references, reference)
				}
			}
		}

		// Check for nested references
		if matches := nestedRefRegex.FindStringSubmatch(trimmedLine); matches != nil {
			if len(matches) >= 3 {
				dirName := matches[1]
				fileName := matches[2]

				// Validate that directory name matches file name
				if dirName == fileName {
					reference := models.Reference{
						Name:    dirName,
						Version: "", // Nested references don't have versions
						Type:    models.ReferenceTypeNested,
						Path:    fmt.Sprintf("nested/%s/%s.puml", dirName, fileName),
					}
					references = append(references, reference)
				}
			}
		}

		// Check for invalid reference patterns and warn
		if strings.Contains(trimmedLine, "!include") && !productRefRegex.MatchString(trimmedLine) && !nestedRefRegex.MatchString(trimmedLine) {
			// This might be an invalid reference pattern
			if strings.Contains(trimmedLine, ".puml") {
				// Log this as a potential issue but don't fail validation
				continue
			}
		}
	}

	return references, nil
}

// validateReference validates a single reference
func (v *PlantUMLValidator) validateReference(ref models.Reference, sm *models.StateMachine, result *models.ValidationResult) {
	// Validate reference name
	if !v.isValidStateName(ref.Name) {
		result.AddError("INVALID_REFERENCE_NAME",
			fmt.Sprintf("Reference name '%s' is invalid", ref.Name), 1, 1)
		return
	}

	// Validate reference type and structure
	switch ref.Type {
	case models.ReferenceTypeProduct:
		v.validateProductReference(ref, sm, result)
	case models.ReferenceTypeNested:
		v.validateNestedReference(ref, sm, result)
	default:
		result.AddError("UNKNOWN_REFERENCE_TYPE",
			fmt.Sprintf("Unknown reference type for '%s'", ref.Name), 1, 1)
	}
}

// validateProductReference validates a product reference
func (v *PlantUMLValidator) validateProductReference(ref models.Reference, sm *models.StateMachine, result *models.ValidationResult) {
	// Product references must have a version
	if ref.Version == "" {
		result.AddError("MISSING_REFERENCE_VERSION",
			fmt.Sprintf("Product reference '%s' must have a version", ref.Name), 1, 1)
		return
	}

	// Validate version format (basic semantic versioning check)
	if !v.isValidVersion(ref.Version) {
		result.AddError("INVALID_REFERENCE_VERSION",
			fmt.Sprintf("Product reference '%s' has invalid version '%s'", ref.Name, ref.Version), 1, 1)
		return
	}

	// Check for self-reference
	if ref.Name == sm.Name && ref.Version == sm.Version {
		result.AddError("SELF_REFERENCE",
			"State machine cannot reference itself", 1, 1)
		return
	}

	// Validate path format
	expectedPath := fmt.Sprintf("products/%s-%s/%s-%s.puml", ref.Name, ref.Version, ref.Name, ref.Version)
	if ref.Path != expectedPath {
		result.AddWarning("INCORRECT_REFERENCE_PATH",
			fmt.Sprintf("Reference path '%s' should be '%s'", ref.Path, expectedPath), 1, 1)
	}
}

// validateNestedReference validates a nested reference
func (v *PlantUMLValidator) validateNestedReference(ref models.Reference, sm *models.StateMachine, result *models.ValidationResult) {
	// Nested references should not have a version
	if ref.Version != "" {
		result.AddWarning("UNEXPECTED_NESTED_VERSION",
			fmt.Sprintf("Nested reference '%s' should not have a version", ref.Name), 1, 1)
	}

	// Check for self-reference (nested can't reference the parent)
	if ref.Name == sm.Name {
		result.AddError("NESTED_SELF_REFERENCE",
			fmt.Sprintf("Nested state machine cannot reference parent '%s'", ref.Name), 1, 1)
		return
	}

	// Validate path format
	expectedPath := fmt.Sprintf("nested/%s/%s.puml", ref.Name, ref.Name)
	if ref.Path != expectedPath {
		result.AddWarning("INCORRECT_NESTED_PATH",
			fmt.Sprintf("Nested reference path '%s' should be '%s'", ref.Path, expectedPath), 1, 1)
	}
}

// isValidVersion checks if a version string follows semantic versioning
func (v *PlantUMLValidator) isValidVersion(version string) bool {
	// Basic semantic versioning pattern: major.minor.patch[-prerelease]
	versionRegex := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[a-zA-Z0-9_.-]+)?$`)
	return versionRegex.MatchString(version)
}

// ResolveReferences resolves and validates reference accessibility
func (v *PlantUMLValidator) ResolveReferences(sm *models.StateMachine) (*models.ValidationResult, error) {
	result := &models.ValidationResult{
		Errors:   []models.ValidationError{},
		Warnings: []models.ValidationWarning{},
		IsValid:  true,
	}

	// If no repository is available, we can't resolve references
	if v.repository == nil {
		result.AddWarning("NO_REPOSITORY", "Cannot resolve references without repository", 1, 1)
		return result, nil
	}

	// First parse references if not already done
	if len(sm.References) == 0 {
		references, err := v.parseReferences(sm.Content)
		if err != nil {
			result.AddError("REFERENCE_PARSE_ERROR", "Failed to parse references from content", 1, 1)
			return result, nil
		}
		sm.References = references
	}

	// Resolve each reference
	for _, ref := range sm.References {
		v.resolveReference(ref, sm, result)
	}

	return result, nil
}

// resolveReference resolves a single reference and checks its accessibility
func (v *PlantUMLValidator) resolveReference(ref models.Reference, sm *models.StateMachine, result *models.ValidationResult) {
	var targetLocation models.Location
	var checkVersion string

	// Determine target location and version based on reference type
	switch ref.Type {
	case models.ReferenceTypeProduct:
		targetLocation = models.LocationProducts
		checkVersion = ref.Version
	case models.ReferenceTypeNested:
		// For nested references, we need to check within the same parent directory
		// This is more complex as nested references are relative to the current state machine
		targetLocation = models.LocationNested
		checkVersion = "" // Nested references don't have versions
	default:
		result.AddError("UNKNOWN_REFERENCE_TYPE",
			fmt.Sprintf("Cannot resolve unknown reference type for '%s'", ref.Name), 1, 1)
		return
	}

	// Check if the referenced state machine exists
	exists, err := v.repository.Exists(ref.Name, checkVersion, targetLocation)
	if err != nil {
		result.AddWarning("REFERENCE_CHECK_ERROR",
			fmt.Sprintf("Failed to check existence of reference '%s': %v", ref.Name, err), 1, 1)
		return
	}

	if !exists {
		// For nested references, the error is more critical since they should be in the same directory structure
		if ref.Type == models.ReferenceTypeNested {
			result.AddError("NESTED_REFERENCE_NOT_FOUND",
				fmt.Sprintf("Nested reference '%s' not found", ref.Name), 1, 1)
		} else {
			result.AddError("PRODUCT_REFERENCE_NOT_FOUND",
				fmt.Sprintf("Product reference '%s-%s' not found", ref.Name, ref.Version), 1, 1)
		}
		return
	}

	// Try to read the referenced state machine to ensure it's accessible
	referencedSM, err := v.repository.ReadStateMachine(ref.Name, checkVersion, targetLocation)
	if err != nil {
		result.AddWarning("REFERENCE_READ_ERROR",
			fmt.Sprintf("Referenced state machine '%s' exists but cannot be read: %v", ref.Name, err), 1, 1)
		return
	}

	// Additional validation: check for circular references
	v.checkCircularReference(ref, referencedSM, sm, result, make(map[string]bool))
}

// checkCircularReference detects circular references between state machines
func (v *PlantUMLValidator) checkCircularReference(ref models.Reference, referencedSM *models.StateMachine, originalSM *models.StateMachine, result *models.ValidationResult, visited map[string]bool) {
	// Create a unique key for the referenced state machine
	refKey := fmt.Sprintf("%s-%s-%s", referencedSM.Name, referencedSM.Version, referencedSM.Location.String())
	originalKey := fmt.Sprintf("%s-%s-%s", originalSM.Name, originalSM.Version, originalSM.Location.String())

	// Check if we've already visited this reference (circular reference detected)
	if visited[refKey] {
		result.AddError("CIRCULAR_REFERENCE",
			fmt.Sprintf("Circular reference detected: '%s' references '%s'", originalSM.Name, ref.Name), 1, 1)
		return
	}

	// Check if the referenced state machine references back to the original
	if refKey == originalKey {
		result.AddError("DIRECT_CIRCULAR_REFERENCE",
			fmt.Sprintf("Direct circular reference: '%s' references itself", ref.Name), 1, 1)
		return
	}

	// Mark this reference as visited
	visited[refKey] = true

	// Parse references from the referenced state machine if not already done
	if len(referencedSM.References) == 0 {
		references, err := v.parseReferences(referencedSM.Content)
		if err != nil {
			// Can't check further, but don't fail validation for this
			return
		}
		referencedSM.References = references
	}

	// Check each reference in the referenced state machine
	for _, nestedRef := range referencedSM.References {
		// Only check if we can resolve the nested reference
		var targetLocation models.Location
		var checkVersion string

		switch nestedRef.Type {
		case models.ReferenceTypeProduct:
			targetLocation = models.LocationProducts
			checkVersion = nestedRef.Version
		case models.ReferenceTypeNested:
			targetLocation = models.LocationNested
			checkVersion = ""
		default:
			continue // Skip unknown reference types
		}

		// Try to read the nested referenced state machine
		nestedReferencedSM, err := v.repository.ReadStateMachine(nestedRef.Name, checkVersion, targetLocation)
		if err != nil {
			continue // Skip if we can't read it
		}

		// Recursively check for circular references
		v.checkCircularReference(nestedRef, nestedReferencedSM, originalSM, result, visited)
	}

	// Remove from visited when we're done with this branch
	delete(visited, refKey)
}

// validatePlantUMLStructure validates the basic PlantUML start/end tags
func (v *PlantUMLValidator) validatePlantUMLStructure(content string, result *models.ValidationResult) {
	lines := strings.Split(content, "\n")

	var startFound, endFound bool
	var startLine, endLine int

	// Check for @startuml and @enduml tags
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "@startuml") {
			if startFound {
				result.AddError("DUPLICATE_START", "Multiple @startuml tags found", i+1, 1)
			} else {
				startFound = true
				startLine = i + 1
			}
		}

		if strings.HasPrefix(trimmedLine, "@enduml") {
			if endFound {
				result.AddError("DUPLICATE_END", "Multiple @enduml tags found", i+1, 1)
			} else {
				endFound = true
				endLine = i + 1
			}
		}
	}

	// Validate structure requirements
	if !startFound {
		result.AddError("MISSING_START", "Missing @startuml tag", 1, 1)
	}

	if !endFound {
		result.AddError("MISSING_END", "Missing @enduml tag", len(lines), 1)
	}

	if startFound && endFound && startLine >= endLine {
		result.AddError("INVALID_ORDER", "@startuml must come before @enduml", startLine, 1)
	}
}

// validateStateMachineSyntax validates state machine specific syntax
func (v *PlantUMLValidator) validateStateMachineSyntax(content string, result *models.ValidationResult) {
	lines := strings.Split(content, "\n")

	// Regular expressions for state machine syntax
	transitionRegex := regexp.MustCompile(`^(.+)\s*-->\s*(.+)$`)
	initialStateRegex := regexp.MustCompile(`^\[\*\]\s*-->\s*(.+)$`)
	finalStateRegex := regexp.MustCompile(`^(.+)\s*-->\s*\[\*\]$`)

	var inPlantUML bool
	var hasInitialState bool
	states := make(map[string]bool)

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		lineNum := i + 1

		// Skip empty lines and comments
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "'") {
			continue
		}

		// Track PlantUML boundaries
		if strings.HasPrefix(trimmedLine, "@startuml") {
			inPlantUML = true
			continue
		}
		if strings.HasPrefix(trimmedLine, "@enduml") {
			inPlantUML = false
			continue
		}

		// Only validate content within PlantUML tags
		if !inPlantUML {
			continue
		}

		// Check for initial state transition
		if initialStateRegex.MatchString(trimmedLine) {
			hasInitialState = true
			matches := initialStateRegex.FindStringSubmatch(trimmedLine)
			if len(matches) > 1 {
				stateName := v.extractStateName(strings.TrimSpace(matches[1]))
				states[stateName] = true

				// Validate state name (only the core state name, not labels)
				if !v.isValidStateName(stateName) {
					result.AddWarning("INVALID_STATE_NAME", "State name should follow naming conventions", lineNum, 1)
				}
			}
			continue
		}

		// Check for final state transition
		if finalStateRegex.MatchString(trimmedLine) {
			matches := finalStateRegex.FindStringSubmatch(trimmedLine)
			if len(matches) > 1 {
				stateName := v.extractStateName(strings.TrimSpace(matches[1]))
				states[stateName] = true

				// Validate state name (only the core state name, not labels)
				if !v.isValidStateName(stateName) {
					result.AddWarning("INVALID_STATE_NAME", "State name should follow naming conventions", lineNum, 1)
				}
			}
			continue
		}

		// Check for regular transitions
		if transitionRegex.MatchString(trimmedLine) {
			matches := transitionRegex.FindStringSubmatch(trimmedLine)
			if len(matches) > 2 {
				fromState := v.extractStateName(strings.TrimSpace(matches[1]))
				toState := v.extractStateName(strings.TrimSpace(matches[2]))

				states[fromState] = true
				states[toState] = true

				// Validate state names (only the core state names, not labels)
				if !v.isValidStateName(fromState) {
					result.AddWarning("INVALID_STATE_NAME", "State name should follow naming conventions", lineNum, 1)
				}
				if !v.isValidStateName(toState) {
					result.AddWarning("INVALID_STATE_NAME", "State name should follow naming conventions", lineNum, 1)
				}
			}
			continue
		}

		// Check for standalone state definitions
		coreStateName := v.extractStateName(trimmedLine)
		if v.isValidStateName(coreStateName) {
			states[coreStateName] = true
			continue
		}

		// Check for known PlantUML constructs that should not trigger unknown syntax warning
		if v.isKnownPlantUMLConstruct(trimmedLine) {
			continue
		}

		// If we reach here, the line might contain invalid syntax
		result.AddWarning("UNKNOWN_SYNTAX", "Line contains unrecognized PlantUML syntax", lineNum, 1)
	}

	// Validate state machine requirements - only check if we found PlantUML tags
	var foundStartTag bool
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "@startuml") {
			foundStartTag = true
			break
		}
	}

	if foundStartTag && !hasInitialState {
		result.AddWarning("NO_INITIAL_STATE", "State machine should have an initial state transition", 1, 1)
	}

	if foundStartTag && len(states) == 0 {
		result.AddError("NO_STATES", "State machine must contain at least one state", 1, 1)
	}
}

// extractStateName extracts the core state name from a potentially labeled state
func (v *PlantUMLValidator) extractStateName(state string) string {
	// Handle special states
	if state == "[*]" {
		return state
	}

	// Extract state name before any colon (label separator)
	if colonIndex := strings.Index(state, ":"); colonIndex != -1 {
		return strings.TrimSpace(state[:colonIndex])
	}

	return state
}

// isValidStateName checks if a state name follows naming conventions
func (v *PlantUMLValidator) isValidStateName(stateName string) bool {
	if stateName == "[*]" {
		return true
	}

	// Allow alphanumeric characters, underscores, and hyphens
	stateRegex := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)
	return stateRegex.MatchString(stateName)
}

// isKnownPlantUMLConstruct checks if a line contains known PlantUML constructs
func (v *PlantUMLValidator) isKnownPlantUMLConstruct(line string) bool {
	knownConstructs := []string{
		"note",
		"title",
		"skinparam",
		"!define",
		"!include",
		"scale",
		"left to right direction",
		"top to bottom direction",
	}

	lowerLine := strings.ToLower(line)
	for _, construct := range knownConstructs {
		if strings.Contains(lowerLine, construct) {
			return true
		}
	}

	// Check for state definitions with descriptions (contains colon)
	if strings.Contains(line, ":") {
		return true
	}

	return false
}

// applyStrictnessFiltering applies validation strictness rules
func (v *PlantUMLValidator) applyStrictnessFiltering(result *models.ValidationResult, strictness models.ValidationStrictness) {
	switch strictness {
	case models.StrictnessProducts:
		// For products, convert non-critical errors to warnings
		// This allows products to have minor issues but still be valid
		var criticalErrors []models.ValidationError
		var convertedWarnings []models.ValidationWarning

		for _, err := range result.Errors {
			if v.isCriticalError(err.Code) {
				// Keep critical errors as errors
				criticalErrors = append(criticalErrors, err)
			} else {
				// Convert non-critical errors to warnings
				warning := models.ValidationWarning{
					Code:    err.Code,
					Message: fmt.Sprintf("(Converted from error) %s", err.Message),
					Line:    err.Line,
					Column:  err.Column,
					Context: err.Context,
				}
				convertedWarnings = append(convertedWarnings, warning)
			}
		}

		// Update result with filtered errors and additional warnings
		result.Errors = criticalErrors
		result.Warnings = append(result.Warnings, convertedWarnings...)
		result.IsValid = len(criticalErrors) == 0

	case models.StrictnessInProgress:
		// For in-progress, keep all errors and warnings as-is
		// This provides the strictest validation for development
		result.IsValid = len(result.Errors) == 0

	default:
		// Default to in-progress behavior for unknown strictness levels
		result.IsValid = len(result.Errors) == 0
	}
}

// isCriticalError determines if an error is critical and should not be converted to warning
// Critical errors are structural issues that make the PlantUML invalid regardless of deployment stage
func (v *PlantUMLValidator) isCriticalError(code string) bool {
	criticalErrors := map[string]bool{
		// PlantUML structural errors - these make the diagram unparseable
		"MISSING_START":   true,
		"MISSING_END":     true,
		"DUPLICATE_START": true,
		"DUPLICATE_END":   true,
		"INVALID_ORDER":   true,
		"NO_STATES":       true,

		// Reference errors that break functionality
		"SELF_REFERENCE":            true,
		"DIRECT_CIRCULAR_REFERENCE": true,
		"CIRCULAR_REFERENCE":        true,
		"REFERENCE_PARSE_ERROR":     true,

		// Critical reference validation errors
		"NESTED_SELF_REFERENCE":  true,
		"UNKNOWN_REFERENCE_TYPE": true,
	}

	return criticalErrors[code]
}
