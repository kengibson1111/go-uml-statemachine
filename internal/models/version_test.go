package models

import (
	"fmt"
	"testing"
)

func TestVersion_String(t *testing.T) {
	tests := []struct {
		name     string
		version  Version
		expected string
	}{
		{
			name:     "basic version",
			version:  Version{Major: 1, Minor: 2, Patch: 3},
			expected: "1.2.3",
		},
		{
			name:     "version with pre-release",
			version:  Version{Major: 1, Minor: 2, Patch: 3, Pre: "alpha"},
			expected: "1.2.3-alpha",
		},
		{
			name:     "version with complex pre-release",
			version:  Version{Major: 2, Minor: 0, Patch: 0, Pre: "beta.1"},
			expected: "2.0.0-beta.1",
		},
		{
			name:     "zero version",
			version:  Version{Major: 0, Minor: 0, Patch: 0},
			expected: "0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.String()
			if result != tt.expected {
				t.Errorf("Version.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestVersion_Compare(t *testing.T) {
	tests := []struct {
		name     string
		v1       Version
		v2       Version
		expected int
	}{
		{
			name:     "equal versions",
			v1:       Version{Major: 1, Minor: 2, Patch: 3},
			v2:       Version{Major: 1, Minor: 2, Patch: 3},
			expected: 0,
		},
		{
			name:     "v1 major greater",
			v1:       Version{Major: 2, Minor: 0, Patch: 0},
			v2:       Version{Major: 1, Minor: 9, Patch: 9},
			expected: 1,
		},
		{
			name:     "v1 major less",
			v1:       Version{Major: 1, Minor: 9, Patch: 9},
			v2:       Version{Major: 2, Minor: 0, Patch: 0},
			expected: -1,
		},
		{
			name:     "v1 minor greater",
			v1:       Version{Major: 1, Minor: 3, Patch: 0},
			v2:       Version{Major: 1, Minor: 2, Patch: 9},
			expected: 1,
		},
		{
			name:     "v1 minor less",
			v1:       Version{Major: 1, Minor: 2, Patch: 9},
			v2:       Version{Major: 1, Minor: 3, Patch: 0},
			expected: -1,
		},
		{
			name:     "v1 patch greater",
			v1:       Version{Major: 1, Minor: 2, Patch: 4},
			v2:       Version{Major: 1, Minor: 2, Patch: 3},
			expected: 1,
		},
		{
			name:     "v1 patch less",
			v1:       Version{Major: 1, Minor: 2, Patch: 3},
			v2:       Version{Major: 1, Minor: 2, Patch: 4},
			expected: -1,
		},
		{
			name:     "release vs pre-release (release greater)",
			v1:       Version{Major: 1, Minor: 2, Patch: 3},
			v2:       Version{Major: 1, Minor: 2, Patch: 3, Pre: "alpha"},
			expected: 1,
		},
		{
			name:     "pre-release vs release (pre-release less)",
			v1:       Version{Major: 1, Minor: 2, Patch: 3, Pre: "alpha"},
			v2:       Version{Major: 1, Minor: 2, Patch: 3},
			expected: -1,
		},
		{
			name:     "pre-release comparison (alpha < beta)",
			v1:       Version{Major: 1, Minor: 2, Patch: 3, Pre: "alpha"},
			v2:       Version{Major: 1, Minor: 2, Patch: 3, Pre: "beta"},
			expected: -1,
		},
		{
			name:     "pre-release comparison (beta > alpha)",
			v1:       Version{Major: 1, Minor: 2, Patch: 3, Pre: "beta"},
			v2:       Version{Major: 1, Minor: 2, Patch: 3, Pre: "alpha"},
			expected: 1,
		},
		{
			name:     "equal pre-release versions",
			v1:       Version{Major: 1, Minor: 2, Patch: 3, Pre: "alpha"},
			v2:       Version{Major: 1, Minor: 2, Patch: 3, Pre: "alpha"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.Compare(tt.v2)
			if result != tt.expected {
				t.Errorf("Version.Compare() = %v, want %v", result, tt.expected)
			}
		})
	}
}
func TestParseVersion(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Version
		expectError bool
	}{
		{
			name:     "basic version",
			input:    "1.2.3",
			expected: Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:     "version with pre-release",
			input:    "1.2.3-alpha",
			expected: Version{Major: 1, Minor: 2, Patch: 3, Pre: "alpha"},
		},
		{
			name:     "version with complex pre-release",
			input:    "2.0.0-beta.1",
			expected: Version{Major: 2, Minor: 0, Patch: 0, Pre: "beta.1"},
		},
		{
			name:     "zero version",
			input:    "0.0.0",
			expected: Version{Major: 0, Minor: 0, Patch: 0},
		},
		{
			name:     "large version numbers",
			input:    "10.20.30",
			expected: Version{Major: 10, Minor: 20, Patch: 30},
		},
		{
			name:     "version with numeric pre-release",
			input:    "1.0.0-1",
			expected: Version{Major: 1, Minor: 0, Patch: 0, Pre: "1"},
		},
		{
			name:     "version with hyphenated pre-release",
			input:    "1.0.0-alpha-beta",
			expected: Version{Major: 1, Minor: 0, Patch: 0, Pre: "alpha-beta"},
		},
		{
			name:        "invalid format - missing patch",
			input:       "1.2",
			expectError: true,
		},
		{
			name:        "invalid format - too many parts",
			input:       "1.2.3.4",
			expectError: true,
		},
		{
			name:        "invalid format - non-numeric major",
			input:       "a.2.3",
			expectError: true,
		},
		{
			name:        "invalid format - non-numeric minor",
			input:       "1.b.3",
			expectError: true,
		},
		{
			name:        "invalid format - non-numeric patch",
			input:       "1.2.c",
			expectError: true,
		},
		{
			name:        "invalid format - negative numbers",
			input:       "-1.2.3",
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "invalid format - leading v",
			input:       "v1.2.3",
			expectError: true,
		},
		{
			name:        "invalid format - spaces",
			input:       "1.2.3 ",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseVersion(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseVersion() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseVersion() unexpected error: %v", err)
				return
			}

			if result.Major != tt.expected.Major {
				t.Errorf("ParseVersion() Major = %v, want %v", result.Major, tt.expected.Major)
			}
			if result.Minor != tt.expected.Minor {
				t.Errorf("ParseVersion() Minor = %v, want %v", result.Minor, tt.expected.Minor)
			}
			if result.Patch != tt.expected.Patch {
				t.Errorf("ParseVersion() Patch = %v, want %v", result.Patch, tt.expected.Patch)
			}
			if result.Pre != tt.expected.Pre {
				t.Errorf("ParseVersion() Pre = %v, want %v", result.Pre, tt.expected.Pre)
			}
		})
	}
}

// TestVersion_RoundTrip tests that parsing and string conversion are consistent
func TestVersion_RoundTrip(t *testing.T) {
	testVersions := []string{
		"1.2.3",
		"0.0.0",
		"10.20.30",
		"1.2.3-alpha",
		"2.0.0-beta.1",
		"1.0.0-alpha-beta",
		"3.1.4-rc.2",
	}

	for _, versionStr := range testVersions {
		t.Run(versionStr, func(t *testing.T) {
			parsed, err := ParseVersion(versionStr)
			if err != nil {
				t.Errorf("ParseVersion() error: %v", err)
				return
			}

			result := parsed.String()
			if result != versionStr {
				t.Errorf("Round trip failed: %s -> %s", versionStr, result)
			}
		})
	}
}

// TestVersion_CompareConsistency tests that comparison is consistent and transitive
func TestVersion_CompareConsistency(t *testing.T) {
	versions := []Version{
		{Major: 1, Minor: 0, Patch: 0},
		{Major: 1, Minor: 0, Patch: 1},
		{Major: 1, Minor: 1, Patch: 0},
		{Major: 2, Minor: 0, Patch: 0},
		{Major: 1, Minor: 0, Patch: 0, Pre: "alpha"},
		{Major: 1, Minor: 0, Patch: 0, Pre: "beta"},
	}

	// Test reflexivity: v.Compare(v) == 0
	for i, v := range versions {
		t.Run(fmt.Sprintf("reflexivity_%d", i), func(t *testing.T) {
			if v.Compare(v) != 0 {
				t.Errorf("Version %s should be equal to itself", v.String())
			}
		})
	}

	// Test antisymmetry: if v1.Compare(v2) == x, then v2.Compare(v1) == -x
	for i, v1 := range versions {
		for j, v2 := range versions {
			if i != j {
				t.Run(fmt.Sprintf("antisymmetry_%d_%d", i, j), func(t *testing.T) {
					result1 := v1.Compare(v2)
					result2 := v2.Compare(v1)
					if result1 != -result2 {
						t.Errorf("Antisymmetry failed: %s.Compare(%s) = %d, %s.Compare(%s) = %d",
							v1.String(), v2.String(), result1, v2.String(), v1.String(), result2)
					}
				})
			}
		}
	}
}
