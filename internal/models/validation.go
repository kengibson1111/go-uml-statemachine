package models

// ValidationStrictness defines the level of validation strictness
type ValidationStrictness int

const (
	StrictnessInProgress ValidationStrictness = iota // Errors and warnings
	StrictnessProducts                               // Warnings only
)

// String returns the string representation of ValidationStrictness
func (vs ValidationStrictness) String() string {
	switch vs {
	case StrictnessInProgress:
		return "in-progress"
	case StrictnessProducts:
		return "products"
	default:
		return "unknown"
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Code     string
	Message  string
	Line     int
	Column   int
	Severity string
	Context  map[string]interface{}
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Code    string
	Message string
	Line    int
	Column  int
	Context map[string]interface{}
}

// ValidationResult contains validation outcomes
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationWarning
	IsValid  bool
}

// HasErrors returns true if there are validation errors
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// HasWarnings returns true if there are validation warnings
func (vr *ValidationResult) HasWarnings() bool {
	return len(vr.Warnings) > 0
}

// AddError adds a validation error
func (vr *ValidationResult) AddError(code, message string, line, column int) {
	vr.Errors = append(vr.Errors, ValidationError{
		Code:     code,
		Message:  message,
		Line:     line,
		Column:   column,
		Severity: "error",
		Context:  make(map[string]interface{}),
	})
	vr.IsValid = false
}

// AddWarning adds a validation warning
func (vr *ValidationResult) AddWarning(code, message string, line, column int) {
	vr.Warnings = append(vr.Warnings, ValidationWarning{
		Code:    code,
		Message: message,
		Line:    line,
		Column:  column,
		Context: make(map[string]interface{}),
	})
}
