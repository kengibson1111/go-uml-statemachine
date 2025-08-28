package models

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/kengibson1111/go-uml-statemachine-models/models"
)

func TestNewPathManager(t *testing.T) {
	tests := []struct {
		name     string
		rootDir  string
		expected string
	}{
		{
			name:     "default root directory",
			rootDir:  "",
			expected: RootDirectoryName,
		},
		{
			name:     "custom root directory",
			rootDir:  "custom-root",
			expected: "custom-root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPathManager(tt.rootDir)
			if pm.GetRootPath() != tt.expected {
				t.Errorf("NewPathManager() root path = %v, want %v", pm.GetRootPath(), tt.expected)
			}
		})
	}
}

func TestPathManager_GetLocationPath(t *testing.T) {
	pm := NewPathManager("")

	tests := []struct {
		name     string
		location Location
		expected string
	}{
		{
			name:     "in-progress location",
			location: LocationFileInProgress,
			expected: filepath.Join(RootDirectoryName, "in-progress"),
		},
		{
			name:     "products location",
			location: LocationFileProducts,
			expected: filepath.Join(RootDirectoryName, "products"),
		},
		{
			name:     "unknown location",
			location: Location(999),
			expected: RootDirectoryName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pm.GetLocationPath(tt.location)
			if result != tt.expected {
				t.Errorf("GetLocationPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPathManager_ValidatePath(t *testing.T) {
	pm := NewPathManager("test-root")

	tests := []struct {
		name      string
		path      string
		wantError bool
		errorType ErrorType
	}{
		{
			name:      "valid relative path",
			path:      "in-progress/user-auth-1.0.0",
			wantError: false,
		},
		{
			name:      "directory traversal attempt",
			path:      "../../../etc/passwd",
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:      "path with dot dot",
			path:      "in-progress/../products/user-auth-1.0.0",
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:      "clean path with resolved dots",
			path:      "in-progress/./user-auth-1.0.0",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.ValidatePath(tt.path)
			if tt.wantError {
				if err == nil {
					t.Errorf("ValidatePath() expected error but got none")
					return
				}
				if diagErr, ok := err.(*StateMachineError); ok {
					if diagErr.Type != tt.errorType {
						t.Errorf("ValidatePath() error type = %v, want %v", diagErr.Type, tt.errorType)
					}
				} else {
					t.Errorf("ValidatePath() expected StateMachineError but got %T", err)
				}
			} else if err != nil {
				t.Errorf("ValidatePath() unexpected error = %v", err)
			}
		})
	}
}

func TestPathManager_ValidateName(t *testing.T) {
	pm := NewPathManager("")

	tests := []struct {
		name      string
		diagName  string
		wantError bool
		errorType ErrorType
	}{
		{
			name:      "valid name",
			diagName:  "user-auth",
			wantError: false,
		},
		{
			name:      "valid name with numbers",
			diagName:  "user-auth-v2",
			wantError: false,
		},
		{
			name:      "valid name with underscores",
			diagName:  "user_auth_system",
			wantError: false,
		},
		{
			name:      "empty name",
			diagName:  "",
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:      "name starting with hyphen",
			diagName:  "-invalid",
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:      "name with spaces",
			diagName:  "user auth",
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:      "name with special characters",
			diagName:  "user@auth",
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:      "reserved name - nested",
			diagName:  "nested",
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:      "reserved name - CON (Windows)",
			diagName:  "CON",
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:      "name too long",
			diagName:  strings.Repeat("a", 101),
			wantError: true,
			errorType: ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.ValidateName(tt.diagName)
			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateName() expected error but got none")
					return
				}
				if diagErr, ok := err.(*StateMachineError); ok {
					if diagErr.Type != tt.errorType {
						t.Errorf("ValidateName() error type = %v, want %v", diagErr.Type, tt.errorType)
					}
				} else {
					t.Errorf("ValidateName() expected StateMachineError but got %T", err)
				}
			} else if err != nil {
				t.Errorf("ValidateName() unexpected error = %v", err)
			}
		})
	}
}

func TestPathManager_ParseDirectoryName(t *testing.T) {
	pm := NewPathManager("")

	tests := []struct {
		name      string
		dirName   string
		want      *PathInfo
		wantError bool
		errorType ErrorType
	}{
		{
			name:    "valid versioned directory",
			dirName: "user-auth-1.0.0",
			want: &PathInfo{
				Name:    "user-auth",
				Version: "1.0.0",
			},
			wantError: false,
		},

		{
			name:    "complex name with hyphens",
			dirName: "user-auth-system-1.2.3",
			want: &PathInfo{
				Name:    "user-auth-system",
				Version: "1.2.3",
			},
			wantError: false,
		},
		{
			name:      "empty directory name",
			dirName:   "",
			want:      nil,
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:      "invalid version format",
			dirName:   "user-auth-invalid-version",
			want:      nil,
			wantError: true,
			errorType: ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pm.ParseDirectoryName(tt.dirName)
			if tt.wantError {
				if err == nil {
					t.Errorf("ParseDirectoryName() expected error but got none")
					return
				}
				if diagErr, ok := err.(*StateMachineError); ok {
					if diagErr.Type != tt.errorType {
						t.Errorf("ParseDirectoryName() error type = %v, want %v", diagErr.Type, tt.errorType)
					}
				} else {
					t.Errorf("ParseDirectoryName() expected StateMachineError but got %T", err)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseDirectoryName() unexpected error = %v", err)
				return
			}

			if got.Name != tt.want.Name {
				t.Errorf("ParseDirectoryName() name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Version != tt.want.Version {
				t.Errorf("ParseDirectoryName() version = %v, want %v", got.Version, tt.want.Version)
			}

		})
	}
}

func TestPathManager_ParseFileName(t *testing.T) {
	pm := NewPathManager("")

	tests := []struct {
		name        string
		diagramType models.DiagramType
		fileName    string
		want        *PathInfo
		wantError   bool
		errorType   ErrorType
	}{
		{
			name:        "valid versioned file with PUML type",
			diagramType: models.DiagramTypePUML,
			fileName:    "user-auth-1.0.0.puml",
			want: &PathInfo{
				Name:    "user-auth",
				Version: "1.0.0",
			},
			wantError: false,
		},
		{
			name:        "complex name with hyphens and PUML type",
			diagramType: models.DiagramTypePUML,
			fileName:    "user-auth-system-1.2.3.puml",
			want: &PathInfo{
				Name:    "user-auth-system",
				Version: "1.2.3",
			},
			wantError: false,
		},
		{
			name:        "empty file name with PUML type",
			diagramType: models.DiagramTypePUML,
			fileName:    "",
			want:        nil,
			wantError:   true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "wrong extension with PUML type",
			diagramType: models.DiagramTypePUML,
			fileName:    "user-auth-1.0.0.txt",
			want:        nil,
			wantError:   true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "invalid version format with PUML type",
			diagramType: models.DiagramTypePUML,
			fileName:    "user-auth-invalid-version.puml",
			want:        nil,
			wantError:   true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "unsupported diagram type - invalid type (1)",
			diagramType: models.DiagramType(1),
			fileName:    "user-auth-1.0.0.puml",
			want:        nil,
			wantError:   true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "unsupported diagram type - invalid type (99)",
			diagramType: models.DiagramType(99),
			fileName:    "user-auth-1.0.0.puml",
			want:        nil,
			wantError:   true,
			errorType:   ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pm.ParseFileName(tt.diagramType, tt.fileName)
			if tt.wantError {
				if err == nil {
					t.Errorf("ParseFileName() expected error but got none")
					return
				}
				if diagErr, ok := err.(*StateMachineError); ok {
					if diagErr.Type != tt.errorType {
						t.Errorf("ParseFileName() error type = %v, want %v", diagErr.Type, tt.errorType)
					}
				} else {
					t.Errorf("ParseFileName() expected StateMachineError but got %T", err)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseFileName() unexpected error = %v", err)
				return
			}

			if got.Name != tt.want.Name {
				t.Errorf("ParseFileName() name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Version != tt.want.Version {
				t.Errorf("ParseFileName() version = %v, want %v", got.Version, tt.want.Version)
			}
		})
	}
}

func TestPathManager_ParseFullPath(t *testing.T) {
	pm := NewPathManager("test-root")

	tests := []struct {
		name        string
		diagramType models.DiagramType
		fullPath    string
		want        *PathInfo
		wantError   bool
		errorType   ErrorType
	}{
		{
			name:        "valid in-progress path with PUML type",
			diagramType: models.DiagramTypePUML,
			fullPath:    filepath.Join("test-root", "in-progress", "puml", "user-auth-1.0.0.puml"),
			want: &PathInfo{
				Name:     "user-auth",
				Version:  "1.0.0",
				Location: LocationFileInProgress,
			},
			wantError: false,
		},
		{
			name:        "valid products path with PUML type",
			diagramType: models.DiagramTypePUML,
			fullPath:    filepath.Join("test-root", "products", "puml", "payment-flow-2.1.0.puml"),
			want: &PathInfo{
				Name:     "payment-flow",
				Version:  "2.1.0",
				Location: LocationFileProducts,
			},
			wantError: false,
		},
		{
			name:        "invalid location with PUML type",
			diagramType: models.DiagramTypePUML,
			fullPath:    filepath.Join("test-root", "invalid-location", "puml", "user-auth-1.0.0.puml"),
			want:        nil,
			wantError:   true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "path too short with PUML type",
			diagramType: models.DiagramTypePUML,
			fullPath:    "test-root",
			want:        nil,
			wantError:   true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "unsupported diagram type - invalid type (1)",
			diagramType: models.DiagramType(1),
			fullPath:    filepath.Join("test-root", "in-progress", "puml", "user-auth-1.0.0.puml"),
			want:        nil,
			wantError:   true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "unsupported diagram type - invalid type (99)",
			diagramType: models.DiagramType(99),
			fullPath:    filepath.Join("test-root", "products", "puml", "payment-flow-2.1.0.puml"),
			want:        nil,
			wantError:   true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "invalid diagram type in path - unknown directory",
			diagramType: models.DiagramTypePUML,
			fullPath:    filepath.Join("test-root", "in-progress", "unknown", "user-auth-1.0.0.puml"),
			want:        nil,
			wantError:   true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "invalid diagram type in path - empty directory",
			diagramType: models.DiagramTypePUML,
			fullPath:    filepath.Join("test-root", "products", "", "payment-flow-2.1.0.puml"),
			want:        nil,
			wantError:   true,
			errorType:   ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pm.ParseFullPath(tt.diagramType, tt.fullPath)
			if tt.wantError {
				if err == nil {
					t.Errorf("ParseFullPath() expected error but got none")
					return
				}
				if diagErr, ok := err.(*StateMachineError); ok {
					if diagErr.Type != tt.errorType {
						t.Errorf("ParseFullPath() error type = %v, want %v", diagErr.Type, tt.errorType)
					}
				} else {
					t.Errorf("ParseFullPath() expected StateMachineError but got %T", err)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseFullPath() unexpected error = %v", err)
				return
			}

			if got.Name != tt.want.Name {
				t.Errorf("ParseFullPath() name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Version != tt.want.Version {
				t.Errorf("ParseFullPath() version = %v, want %v", got.Version, tt.want.Version)
			}
			if got.Location != tt.want.Location {
				t.Errorf("ParseFullPath() location = %v, want %v", got.Location, tt.want.Location)
			}
		})
	}
}

func TestBuildFileName(t *testing.T) {
	tests := []struct {
		name        string
		diagramType models.DiagramType
		fileName    string
		version     string
		want        string
		wantError   bool
		errorType   ErrorType
	}{
		{
			name:        "valid PUML file name",
			diagramType: models.DiagramTypePUML,
			fileName:    "user-auth",
			version:     "1.0.0",
			want:        "user-auth-1.0.0.puml",
			wantError:   false,
		},
		{
			name:        "valid PUML file name with complex name",
			diagramType: models.DiagramTypePUML,
			fileName:    "user-auth-system",
			version:     "2.1.3",
			want:        "user-auth-system-2.1.3.puml",
			wantError:   false,
		},
		{
			name:        "valid PUML file name with pre-release version",
			diagramType: models.DiagramTypePUML,
			fileName:    "payment-flow",
			version:     "1.0.0-alpha.1",
			want:        "payment-flow-1.0.0-alpha.1.puml",
			wantError:   false,
		},
		{
			name:        "unsupported diagram type - invalid type (1)",
			diagramType: models.DiagramType(1),
			fileName:    "user-auth",
			version:     "1.0.0",
			want:        "",
			wantError:   true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "unsupported diagram type - invalid type (99)",
			diagramType: models.DiagramType(99),
			fileName:    "user-auth",
			version:     "1.0.0",
			want:        "",
			wantError:   true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "empty name with PUML type",
			diagramType: models.DiagramTypePUML,
			fileName:    "",
			version:     "1.0.0",
			want:        "",
			wantError:   true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "empty version with PUML type",
			diagramType: models.DiagramTypePUML,
			fileName:    "user-auth",
			version:     "",
			want:        "",
			wantError:   true,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "both empty name and version with PUML type",
			diagramType: models.DiagramTypePUML,
			fileName:    "",
			version:     "",
			want:        "",
			wantError:   true,
			errorType:   ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildFileName(tt.diagramType, tt.fileName, tt.version)
			if tt.wantError {
				if err == nil {
					t.Errorf("BuildFileName() expected error but got none")
					return
				}
				if diagErr, ok := err.(*StateMachineError); ok {
					if diagErr.Type != tt.errorType {
						t.Errorf("BuildFileName() error type = %v, want %v", diagErr.Type, tt.errorType)
					}
				} else {
					t.Errorf("BuildFileName() expected StateMachineError but got %T", err)
				}
				return
			}

			if err != nil {
				t.Errorf("BuildFileName() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("BuildFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}
