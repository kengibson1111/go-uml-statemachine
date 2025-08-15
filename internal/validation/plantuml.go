package validation

import (
	"regexp"
	"strings"

	"github.com/kengibson1111/go-uml-statemachine/internal/models"
)

// PlantUMLValidator implements the Validator interface for PlantUML syntax validation
type PlantUMLValidator struct{}

// NewPlantUMLValidator creates a new PlantUML validator instance
func NewPlantUMLValidator() *PlantUMLValidator {
	return &PlantUMLValidator{}
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

// ValidateReferences validates references in a state machine (placeholder for now)
func (v *PlantUMLValidator) ValidateReferences(sm *models.StateMachine) (*models.ValidationResult, error) {
	// This will be implemented in task 6.2
	return &models.ValidationResult{
		Errors:   []models.ValidationError{},
		Warnings: []models.ValidationWarning{},
		IsValid:  true,
	}, nil
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
		// For products, convert errors to warnings (except critical structural errors)
		var criticalErrors []models.ValidationError
		for _, err := range result.Errors {
			if v.isCriticalError(err.Code) {
				criticalErrors = append(criticalErrors, err)
			} else {
				// Convert to warning
				result.AddWarning(err.Code, err.Message, err.Line, err.Column)
			}
		}
		result.Errors = criticalErrors
		result.IsValid = len(criticalErrors) == 0

	case models.StrictnessInProgress:
		// For in-progress, keep all errors and warnings
		result.IsValid = len(result.Errors) == 0
	}
}

// isCriticalError determines if an error is critical and should not be converted to warning
func (v *PlantUMLValidator) isCriticalError(code string) bool {
	criticalErrors := map[string]bool{
		"MISSING_START":   true,
		"MISSING_END":     true,
		"DUPLICATE_START": true,
		"DUPLICATE_END":   true,
		"INVALID_ORDER":   true,
		"NO_STATES":       true,
	}

	return criticalErrors[code]
}
