package models

import (
	"path/filepath"
	"strings"
	"testing"
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
			location: LocationInProgress,
			expected: filepath.Join(RootDirectoryName, "in-progress"),
		},
		{
			name:     "products location",
			location: LocationProducts,
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

func TestPathManager_GetStateMachineDirectoryPath(t *testing.T) {
	pm := NewPathManager("")

	tests := []struct {
		name     string
		smName   string
		version  string
		location Location
		expected string
	}{
		{
			name:     "in-progress state-machine diagram",
			smName:   "user-auth",
			version:  "1.0.0",
			location: LocationInProgress,
			expected: filepath.Join(RootDirectoryName, "in-progress", "user-auth-1.0.0"),
		},
		{
			name:     "products state-machine diagram",
			smName:   "payment-flow",
			version:  "2.1.0",
			location: LocationProducts,
			expected: filepath.Join(RootDirectoryName, "products", "payment-flow-2.1.0"),
		},
		{
			name:     "nested state-machine diagram",
			smName:   "child-sm",
			version:  "1.0.0",
			location: LocationNested,
			expected: filepath.Join(RootDirectoryName, "child-sm"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pm.GetStateMachineDirectoryPath(tt.smName, tt.version, tt.location)
			if result != tt.expected {
				t.Errorf("GetStateMachineDirectoryPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPathManager_GetStateMachineFilePath(t *testing.T) {
	pm := NewPathManager("")

	tests := []struct {
		name     string
		smName   string
		version  string
		location Location
		expected string
	}{
		{
			name:     "in-progress state-machine diagram file",
			smName:   "user-auth",
			version:  "1.0.0",
			location: LocationInProgress,
			expected: filepath.Join(RootDirectoryName, "in-progress", "user-auth-1.0.0", "user-auth-1.0.0.puml"),
		},
		{
			name:     "products state-machine diagram file",
			smName:   "payment-flow",
			version:  "2.1.0",
			location: LocationProducts,
			expected: filepath.Join(RootDirectoryName, "products", "payment-flow-2.1.0", "payment-flow-2.1.0.puml"),
		},
		{
			name:     "nested state-machine diagram file",
			smName:   "child-sm",
			version:  "1.0.0",
			location: LocationNested,
			expected: filepath.Join(RootDirectoryName, "child-sm", "child-sm.puml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pm.GetStateMachineFilePath(tt.smName, tt.version, tt.location)
			if result != tt.expected {
				t.Errorf("GetStateMachineFilePath() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPathManager_GetNestedPaths(t *testing.T) {
	pm := NewPathManager("")

	t.Run("nested directory path", func(t *testing.T) {
		expected := filepath.Join(RootDirectoryName, "in-progress", "parent-1.0.0", "nested")
		result := pm.GetNestedDirectoryPath("parent", "1.0.0", LocationInProgress)
		if result != expected {
			t.Errorf("GetNestedDirectoryPath() = %v, want %v", result, expected)
		}
	})

	t.Run("nested state-machine diagram directory path", func(t *testing.T) {
		expected := filepath.Join(RootDirectoryName, "in-progress", "parent-1.0.0", "nested", "child")
		result := pm.GetNestedStateMachineDirectoryPath("parent", "1.0.0", LocationInProgress, "child")
		if result != expected {
			t.Errorf("GetNestedStateMachineDirectoryPath() = %v, want %v", result, expected)
		}
	})

	t.Run("nested state-machine diagram file path", func(t *testing.T) {
		expected := filepath.Join(RootDirectoryName, "in-progress", "parent-1.0.0", "nested", "child", "child.puml")
		result := pm.GetNestedStateMachineFilePath("parent", "1.0.0", LocationInProgress, "child")
		if result != expected {
			t.Errorf("GetNestedStateMachineFilePath() = %v, want %v", result, expected)
		}
	})
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
				if smErr, ok := err.(*StateMachineError); ok {
					if smErr.Type != tt.errorType {
						t.Errorf("ValidatePath() error type = %v, want %v", smErr.Type, tt.errorType)
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
		smName    string
		wantError bool
		errorType ErrorType
	}{
		{
			name:      "valid name",
			smName:    "user-auth",
			wantError: false,
		},
		{
			name:      "valid name with numbers",
			smName:    "user-auth-v2",
			wantError: false,
		},
		{
			name:      "valid name with underscores",
			smName:    "user_auth_system",
			wantError: false,
		},
		{
			name:      "empty name",
			smName:    "",
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:      "name starting with hyphen",
			smName:    "-invalid",
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:      "name with spaces",
			smName:    "user auth",
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:      "name with special characters",
			smName:    "user@auth",
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:      "reserved name - nested",
			smName:    "nested",
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:      "reserved name - CON (Windows)",
			smName:    "CON",
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:      "name too long",
			smName:    strings.Repeat("a", 101),
			wantError: true,
			errorType: ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pm.ValidateName(tt.smName)
			if tt.wantError {
				if err == nil {
					t.Errorf("ValidateName() expected error but got none")
					return
				}
				if smErr, ok := err.(*StateMachineError); ok {
					if smErr.Type != tt.errorType {
						t.Errorf("ValidateName() error type = %v, want %v", smErr.Type, tt.errorType)
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
				Name:     "user-auth",
				Version:  "1.0.0",
				IsNested: false,
			},
			wantError: false,
		},
		{
			name:    "valid nested directory",
			dirName: "child-sm",
			want: &PathInfo{
				Name:     "child-sm",
				Version:  "",
				IsNested: true,
			},
			wantError: false,
		},
		{
			name:    "complex name with hyphens",
			dirName: "user-auth-system-1.2.3",
			want: &PathInfo{
				Name:     "user-auth-system",
				Version:  "1.2.3",
				IsNested: false,
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
			name:    "invalid version treated as nested",
			dirName: "user-auth-invalid-version",
			want: &PathInfo{
				Name:     "user-auth-invalid-version",
				Version:  "",
				IsNested: true,
			},
			wantError: false,
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
				if smErr, ok := err.(*StateMachineError); ok {
					if smErr.Type != tt.errorType {
						t.Errorf("ParseDirectoryName() error type = %v, want %v", smErr.Type, tt.errorType)
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
			if got.IsNested != tt.want.IsNested {
				t.Errorf("ParseDirectoryName() isNested = %v, want %v", got.IsNested, tt.want.IsNested)
			}
		})
	}
}

func TestPathManager_ParseFileName(t *testing.T) {
	pm := NewPathManager("")

	tests := []struct {
		name      string
		fileName  string
		want      *PathInfo
		wantError bool
		errorType ErrorType
	}{
		{
			name:     "valid versioned file",
			fileName: "user-auth-1.0.0.puml",
			want: &PathInfo{
				Name:     "user-auth",
				Version:  "1.0.0",
				IsNested: false,
			},
			wantError: false,
		},
		{
			name:     "valid nested file",
			fileName: "child-sm.puml",
			want: &PathInfo{
				Name:     "child-sm",
				Version:  "",
				IsNested: true,
			},
			wantError: false,
		},
		{
			name:     "complex name with hyphens",
			fileName: "user-auth-system-1.2.3.puml",
			want: &PathInfo{
				Name:     "user-auth-system",
				Version:  "1.2.3",
				IsNested: false,
			},
			wantError: false,
		},
		{
			name:      "empty file name",
			fileName:  "",
			want:      nil,
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:      "wrong extension",
			fileName:  "user-auth-1.0.0.txt",
			want:      nil,
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:     "invalid version treated as nested",
			fileName: "user-auth-invalid-version.puml",
			want: &PathInfo{
				Name:     "user-auth-invalid-version",
				Version:  "",
				IsNested: true,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pm.ParseFileName(tt.fileName)
			if tt.wantError {
				if err == nil {
					t.Errorf("ParseFileName() expected error but got none")
					return
				}
				if smErr, ok := err.(*StateMachineError); ok {
					if smErr.Type != tt.errorType {
						t.Errorf("ParseFileName() error type = %v, want %v", smErr.Type, tt.errorType)
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
			if got.IsNested != tt.want.IsNested {
				t.Errorf("ParseFileName() isNested = %v, want %v", got.IsNested, tt.want.IsNested)
			}
		})
	}
}

func TestPathManager_ParseFullPath(t *testing.T) {
	pm := NewPathManager("test-root")

	tests := []struct {
		name      string
		fullPath  string
		want      *PathInfo
		wantError bool
		errorType ErrorType
	}{
		{
			name:     "valid in-progress path",
			fullPath: filepath.Join("test-root", "in-progress", "user-auth-1.0.0"),
			want: &PathInfo{
				Name:     "user-auth",
				Version:  "1.0.0",
				Location: LocationInProgress,
				IsNested: false,
			},
			wantError: false,
		},
		{
			name:     "valid products path",
			fullPath: filepath.Join("test-root", "products", "payment-flow-2.1.0"),
			want: &PathInfo{
				Name:     "payment-flow",
				Version:  "2.1.0",
				Location: LocationProducts,
				IsNested: false,
			},
			wantError: false,
		},
		{
			name:     "valid nested path",
			fullPath: filepath.Join("test-root", "in-progress", "parent-1.0.0", "nested", "child"),
			want: &PathInfo{
				Name:     "child",
				Version:  "",
				Location: LocationNested,
				IsNested: true,
				Parent: &PathInfo{
					Name:     "parent",
					Version:  "1.0.0",
					Location: LocationInProgress,
					IsNested: false,
				},
			},
			wantError: false,
		},
		{
			name:      "invalid location",
			fullPath:  filepath.Join("test-root", "invalid-location", "user-auth-1.0.0"),
			want:      nil,
			wantError: true,
			errorType: ErrorTypeValidation,
		},
		{
			name:      "path too short",
			fullPath:  "test-root",
			want:      nil,
			wantError: true,
			errorType: ErrorTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pm.ParseFullPath(tt.fullPath)
			if tt.wantError {
				if err == nil {
					t.Errorf("ParseFullPath() expected error but got none")
					return
				}
				if smErr, ok := err.(*StateMachineError); ok {
					if smErr.Type != tt.errorType {
						t.Errorf("ParseFullPath() error type = %v, want %v", smErr.Type, tt.errorType)
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
			if got.IsNested != tt.want.IsNested {
				t.Errorf("ParseFullPath() isNested = %v, want %v", got.IsNested, tt.want.IsNested)
			}

			// Check parent information for nested paths
			if tt.want.Parent != nil {
				if got.Parent == nil {
					t.Errorf("ParseFullPath() expected parent but got none")
					return
				}
				if got.Parent.Name != tt.want.Parent.Name {
					t.Errorf("ParseFullPath() parent name = %v, want %v", got.Parent.Name, tt.want.Parent.Name)
				}
				if got.Parent.Version != tt.want.Parent.Version {
					t.Errorf("ParseFullPath() parent version = %v, want %v", got.Parent.Version, tt.want.Parent.Version)
				}
				if got.Parent.Location != tt.want.Parent.Location {
					t.Errorf("ParseFullPath() parent location = %v, want %v", got.Parent.Location, tt.want.Parent.Location)
				}
			}
		})
	}
}
