package models

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version represents a semantic version
type Version struct {
	Major int
	Minor int
	Patch int
	Pre   string // pre-release identifier
}

// String returns the string representation of the version
func (v Version) String() string {
	version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Pre != "" {
		version += "-" + v.Pre
	}
	return version
}

// Compare compares two versions
// Returns -1 if v < other, 0 if v == other, 1 if v > other
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	// Handle pre-release versions
	if v.Pre == "" && other.Pre != "" {
		return 1 // release version is greater than pre-release
	}
	if v.Pre != "" && other.Pre == "" {
		return -1 // pre-release version is less than release
	}
	if v.Pre != "" && other.Pre != "" {
		return strings.Compare(v.Pre, other.Pre)
	}

	return 0
}

// ParseVersion parses a version string into a Version struct
func ParseVersion(s string) (Version, error) {
	// Regular expression for semantic versioning
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-([a-zA-Z0-9\-\.]+))?$`)
	matches := re.FindStringSubmatch(s)

	if len(matches) < 4 {
		return Version{}, fmt.Errorf("invalid version format: %s", s)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %s", matches[1])
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %s", matches[2])
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %s", matches[3])
	}

	pre := ""
	if len(matches) > 4 && matches[4] != "" {
		pre = matches[4]
	}

	return Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Pre:   pre,
	}, nil
}
