package models

import (
	"testing"
	"time"
)

func TestLocation_String(t *testing.T) {
	tests := []struct {
		name     string
		location Location
		expected string
	}{
		{
			name:     "in-progress location",
			location: LocationInProgress,
			expected: "in-progress",
		},
		{
			name:     "products location",
			location: LocationProducts,
			expected: "products",
		},
		{
			name:     "nested location",
			location: LocationNested,
			expected: "nested",
		},
		{
			name:     "unknown location",
			location: Location(999),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.location.String()
			if result != tt.expected {
				t.Errorf("Location.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReferenceType_String(t *testing.T) {
	tests := []struct {
		name     string
		refType  ReferenceType
		expected string
	}{
		{
			name:     "product reference type",
			refType:  ReferenceTypeProduct,
			expected: "product",
		},
		{
			name:     "nested reference type",
			refType:  ReferenceTypeNested,
			expected: "nested",
		},
		{
			name:     "unknown reference type",
			refType:  ReferenceType(999),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.refType.String()
			if result != tt.expected {
				t.Errorf("ReferenceType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStateMachine_Creation(t *testing.T) {
	now := time.Now()
	metadata := Metadata{
		CreatedAt:  now,
		ModifiedAt: now,
		Author:     "test-author",
		Tags:       []string{"tag1", "tag2"},
	}

	references := []Reference{
		{
			Name:    "ref1",
			Version: "1.0.0",
			Type:    ReferenceTypeProduct,
			Path:    "/path/to/ref1",
		},
		{
			Name: "nested-ref",
			Type: ReferenceTypeNested,
			Path: "/path/to/nested",
		},
	}

	diag := StateMachineDiagram{
		Name:       "test-machine",
		Version:    "1.0.0",
		Content:    "@startuml\n[*] --> State1\n@enduml",
		References: references,
		Location:   LocationInProgress,
		Metadata:   metadata,
	}

	// Test that all fields are properly set
	if diag.Name != "test-machine" {
		t.Errorf("StateMachine.Name = %v, want %v", diag.Name, "test-machine")
	}
	if diag.Version != "1.0.0" {
		t.Errorf("StateMachine.Version = %v, want %v", diag.Version, "1.0.0")
	}
	if diag.Content != "@startuml\n[*] --> State1\n@enduml" {
		t.Errorf("StateMachine.Content = %v, want %v", diag.Content, "@startuml\n[*] --> State1\n@enduml")
	}
	if len(diag.References) != 2 {
		t.Errorf("StateMachine.References length = %v, want %v", len(diag.References), 2)
	}
	if diag.Location != LocationInProgress {
		t.Errorf("StateMachine.Location = %v, want %v", diag.Location, LocationInProgress)
	}
	if diag.Metadata.Author != "test-author" {
		t.Errorf("StateMachine.Metadata.Author = %v, want %v", diag.Metadata.Author, "test-author")
	}
}

func TestReference_ProductReference(t *testing.T) {
	ref := Reference{
		Name:    "user-auth",
		Version: "2.1.0",
		Type:    ReferenceTypeProduct,
		Path:    "/products/user-auth-2.1.0",
	}

	if ref.Name != "user-auth" {
		t.Errorf("Reference.Name = %v, want %v", ref.Name, "user-auth")
	}
	if ref.Version != "2.1.0" {
		t.Errorf("Reference.Version = %v, want %v", ref.Version, "2.1.0")
	}
	if ref.Type != ReferenceTypeProduct {
		t.Errorf("Reference.Type = %v, want %v", ref.Type, ReferenceTypeProduct)
	}
	if ref.Path != "/products/user-auth-2.1.0" {
		t.Errorf("Reference.Path = %v, want %v", ref.Path, "/products/user-auth-2.1.0")
	}
}

func TestReference_NestedReference(t *testing.T) {
	ref := Reference{
		Name:    "validation-states",
		Version: "", // nested references don't have versions
		Type:    ReferenceTypeNested,
		Path:    "/nested/validation-states",
	}

	if ref.Name != "validation-states" {
		t.Errorf("Reference.Name = %v, want %v", ref.Name, "validation-states")
	}
	if ref.Version != "" {
		t.Errorf("Reference.Version = %v, want empty string", ref.Version)
	}
	if ref.Type != ReferenceTypeNested {
		t.Errorf("Reference.Type = %v, want %v", ref.Type, ReferenceTypeNested)
	}
	if ref.Path != "/nested/validation-states" {
		t.Errorf("Reference.Path = %v, want %v", ref.Path, "/nested/validation-states")
	}
}

func TestMetadata_Creation(t *testing.T) {
	createdTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	modifiedTime := time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC)
	tags := []string{"authentication", "security", "v1"}

	metadata := Metadata{
		CreatedAt:  createdTime,
		ModifiedAt: modifiedTime,
		Author:     "john.doe",
		Tags:       tags,
	}

	if !metadata.CreatedAt.Equal(createdTime) {
		t.Errorf("Metadata.CreatedAt = %v, want %v", metadata.CreatedAt, createdTime)
	}
	if !metadata.ModifiedAt.Equal(modifiedTime) {
		t.Errorf("Metadata.ModifiedAt = %v, want %v", metadata.ModifiedAt, modifiedTime)
	}
	if metadata.Author != "john.doe" {
		t.Errorf("Metadata.Author = %v, want %v", metadata.Author, "john.doe")
	}
	if len(metadata.Tags) != 3 {
		t.Errorf("Metadata.Tags length = %v, want %v", len(metadata.Tags), 3)
	}
	if metadata.Tags[0] != "authentication" {
		t.Errorf("Metadata.Tags[0] = %v, want %v", metadata.Tags[0], "authentication")
	}
}

func TestMetadata_EmptyTags(t *testing.T) {
	metadata := Metadata{
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
		Author:     "test-author",
		Tags:       []string{},
	}

	if len(metadata.Tags) != 0 {
		t.Errorf("Metadata.Tags length = %v, want %v", len(metadata.Tags), 0)
	}
}

func TestMetadata_NilTags(t *testing.T) {
	metadata := Metadata{
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
		Author:     "test-author",
		Tags:       nil,
	}

	if metadata.Tags != nil {
		t.Errorf("Metadata.Tags = %v, want nil", metadata.Tags)
	}
}
