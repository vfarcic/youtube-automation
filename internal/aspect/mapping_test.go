package aspect

import (
	"reflect"
	"sort"
	"testing"

	"devopstoolkit/youtube-automation/internal/storage"
)

func TestGetVideoAspectMappings(t *testing.T) {
	mappings := GetVideoAspectMappings()

	t.Run("Should return exactly 6 aspect mappings", func(t *testing.T) {
		expectedCount := 6
		if len(mappings) != expectedCount {
			t.Errorf("Expected %d mappings, got %d", expectedCount, len(mappings))
		}
	})

	t.Run("Should have correct aspect keys in order", func(t *testing.T) {
		expectedKeys := []string{
			AspectKeyInitialDetails,
			AspectKeyWorkProgress,
			AspectKeyDefinition,
			AspectKeyPostProduction,
			AspectKeyPublishing,
			AspectKeyPostPublish,
		}

		// Sort mappings by order to ensure consistent testing
		sortedMappings := make([]AspectMapping, len(mappings))
		copy(sortedMappings, mappings)
		sort.Slice(sortedMappings, func(i, j int) bool {
			return sortedMappings[i].Order < sortedMappings[j].Order
		})

		for i, mapping := range sortedMappings {
			if mapping.AspectKey != expectedKeys[i] {
				t.Errorf("Expected aspect key %s at index %d, got %s", expectedKeys[i], i, mapping.AspectKey)
			}
		}
	})

	t.Run("All aspects should have valid titles", func(t *testing.T) {
		expectedTitles := map[string]string{
			AspectKeyInitialDetails: "Initial Details",
			AspectKeyWorkProgress:   "Work Progress",
			AspectKeyDefinition:     "Definition",
			AspectKeyPostProduction: "Post Production",
			AspectKeyPublishing:     "Publishing",
			AspectKeyPostPublish:    "Post Publish",
		}

		for _, mapping := range mappings {
			expectedTitle := expectedTitles[mapping.AspectKey]
			if mapping.Title != expectedTitle {
				t.Errorf("Expected title %s for aspect %s, got %s", expectedTitle, mapping.AspectKey, mapping.Title)
			}
		}
	})

	t.Run("All aspects should have non-empty descriptions", func(t *testing.T) {
		for _, mapping := range mappings {
			if mapping.Description == "" {
				t.Errorf("Aspect %s should have a non-empty description", mapping.AspectKey)
			}
		}
	})

	t.Run("All aspects should have at least one field", func(t *testing.T) {
		for _, mapping := range mappings {
			if len(mapping.Fields) == 0 {
				t.Errorf("Aspect %s should have at least one field", mapping.AspectKey)
			}
		}
	})

	t.Run("Fields should have correct order values starting from 1", func(t *testing.T) {
		for _, mapping := range mappings {
			for i, field := range mapping.Fields {
				expectedOrder := i + 1
				if field.Order != expectedOrder {
					t.Errorf("Field %s in aspect %s should have order %d, got %d", field.FieldName, mapping.AspectKey, expectedOrder, field.Order)
				}
			}
		}
	})

	t.Run("All field types should be valid", func(t *testing.T) {
		validTypes := map[string]bool{
			FieldTypeString:  true,
			FieldTypeText:    true,
			FieldTypeBoolean: true,
			FieldTypeDate:    true,
			FieldTypeNumber:  true,
			FieldTypeSelect:  true,
		}

		for _, mapping := range mappings {
			for _, field := range mapping.Fields {
				if !validTypes[field.FieldType] {
					t.Errorf("Invalid field type %s for field %s in aspect %s", field.FieldType, field.FieldName, mapping.AspectKey)
				}
			}
		}
	})

	t.Run("Field names should match JSON struct tags", func(t *testing.T) {
		// Test specific known fields to ensure they match JSON tags
		expectedFieldNames := map[string]string{
			"Project Name":        "projectName",
			"Project URL":         "projectURL",
			"Sponsorship Amount":  "sponsorship.amount",
			"Sponsorship Emails":  "sponsorship.emails",
			"Sponsorship Blocked": "sponsorship.blocked",
			"Date":                "date",
			"Delayed":             "delayed",
			"Gist":                "gist",
		}

		// Find initial details mapping by key, not by index
		var initialDetailsMapping *AspectMapping
		for _, mapping := range mappings {
			if mapping.AspectKey == AspectKeyInitialDetails {
				initialDetailsMapping = &mapping
				break
			}
		}

		if initialDetailsMapping == nil {
			t.Fatalf("Could not find initial details mapping")
		}

		for _, field := range initialDetailsMapping.Fields {
			if expectedFieldName, exists := expectedFieldNames[field.Name]; exists {
				if field.FieldName != expectedFieldName {
					t.Errorf("Field %s should have fieldName %s, got %s", field.Name, expectedFieldName, field.FieldName)
				}
			}
		}
	})

	t.Run("Field names should be unique within each aspect", func(t *testing.T) {
		for _, mapping := range mappings {
			fieldNames := make(map[string]bool)
			for _, field := range mapping.Fields {
				if fieldNames[field.FieldName] {
					t.Errorf("Duplicate field name %s in aspect %s", field.FieldName, mapping.AspectKey)
				}
				fieldNames[field.FieldName] = true
			}
		}
	})

	t.Run("Should have expected field counts per aspect", func(t *testing.T) {
		expectedCounts := map[string]int{
			AspectKeyInitialDetails: 8,  // ProjectName, ProjectURL, Amount, Emails, Blocked, Date, Delayed, Gist
			AspectKeyWorkProgress:   11, // Code, Head, Screen, RelatedVideos, Thumbnails, Diagrams, Screenshots, Location, Tagline, TaglineIdeas, OtherLogos
			AspectKeyDefinition:     8,  // Title, Description, Highlight, Tags, DescriptionTags, Tweet, Animations, RequestThumbnail
			AspectKeyPostProduction: 6,  // Thumbnail, Members, RequestEdit, Timecodes, Movie, Slides
			AspectKeyPublishing:     3,  // UploadVideo, VideoId, HugoPath
			AspectKeyPostPublish:    10, // DOT, BlueSky, LinkedIn, Slack, YouTube Highlight/Comment/Reply, GDE, Repo, NotifySponsors
		}

		for _, mapping := range mappings {
			expectedCount := expectedCounts[mapping.AspectKey]
			actualCount := len(mapping.Fields)
			if actualCount != expectedCount {
				t.Errorf("Expected %d fields for aspect %s, got %d", expectedCount, mapping.AspectKey, actualCount)
			}
		}
	})
}

func TestGetFieldValueByJSONPath(t *testing.T) {
	video := storage.Video{
		ProjectName: "Test Project",
		ProjectURL:  "https://github.com/test/project",
		Date:        "2024-01-01T10:00",
		Delayed:     false,
		Gist:        "test.md",
		Sponsorship: storage.Sponsorship{
			Amount:  "1000",
			Emails:  "test@example.com",
			Blocked: "No",
		},
	}

	t.Run("Should return correct values for simple string properties", func(t *testing.T) {
		value := GetFieldValueByJSONPath(video, "projectName")
		if value != "Test Project" {
			t.Errorf("Expected 'Test Project', got %v", value)
		}
	})

	t.Run("Should return correct values for boolean properties", func(t *testing.T) {
		value := GetFieldValueByJSONPath(video, "delayed")
		if value != false {
			t.Errorf("Expected false, got %v", value)
		}
	})

	t.Run("Should return correct values for nested properties", func(t *testing.T) {
		value := GetFieldValueByJSONPath(video, "sponsorship.amount")
		if value != "1000" {
			t.Errorf("Expected '1000', got %v", value)
		}

		value = GetFieldValueByJSONPath(video, "sponsorship.emails")
		if value != "test@example.com" {
			t.Errorf("Expected 'test@example.com', got %v", value)
		}
	})

	t.Run("Should return nil for non-existent properties", func(t *testing.T) {
		value := GetFieldValueByJSONPath(video, "nonexistent")
		if value != nil {
			t.Errorf("Expected nil for non-existent property, got %v", value)
		}

		value = GetFieldValueByJSONPath(video, "sponsorship.nonexistent")
		if value != nil {
			t.Errorf("Expected nil for non-existent nested property, got %v", value)
		}
	})
}

func TestGenerateDisplayName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ProjectName", "Project Name"},
		{"ProjectURL", "Project URL"},
		{"VideoId", "Video ID"},
		{"DOTPosted", "DOT Posted"},
		{"GDE", "GDE"},
		{"YouTubeHighlight", "YouTube Highlight"},
		{"BlueSkyPosted", "Blue Sky Posted"},
		{"HNPosted", "HN Posted"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := generateDisplayName(test.input)
			if result != test.expected {
				t.Errorf("generateDisplayName(%s) = %s, expected %s", test.input, result, test.expected)
			}
		})
	}
}

func TestDetermineFieldType(t *testing.T) {
	video := storage.Video{}
	videoType := reflect.TypeOf(video)

	tests := []struct {
		fieldName    string
		expectedType string
	}{
		{"ProjectName", FieldTypeString},
		{"Delayed", FieldTypeBoolean},
		{"Code", FieldTypeBoolean},
		{"Head", FieldTypeBoolean},
		{"Date", FieldTypeDate},
		{"Description", FieldTypeText},
		{"Tags", FieldTypeText},
		{"Timecodes", FieldTypeText},
	}

	for _, test := range tests {
		t.Run(test.fieldName, func(t *testing.T) {
			field, found := videoType.FieldByName(test.fieldName)
			if !found {
				t.Fatalf("Field %s not found in Video struct", test.fieldName)
			}

			result := determineFieldType(field.Type, test.fieldName)
			if result != test.expectedType {
				t.Errorf("determineFieldType for %s = %s, expected %s", test.fieldName, result, test.expectedType)
			}
		})
	}
}

func TestGenerateFieldMapping(t *testing.T) {
	videoType := reflect.TypeOf(storage.Video{})

	t.Run("Should generate correct mapping for simple field", func(t *testing.T) {
		mapping := generateFieldMapping(videoType, "ProjectName", 1)
		if mapping == nil {
			t.Fatal("Expected non-nil mapping")
		}

		if mapping.Name != "Project Name" {
			t.Errorf("Expected name 'Project Name', got %s", mapping.Name)
		}
		if mapping.FieldName != "projectName" {
			t.Errorf("Expected fieldName 'projectName', got %s", mapping.FieldName)
		}
		if mapping.FieldType != FieldTypeString {
			t.Errorf("Expected fieldType 'string', got %s", mapping.FieldType)
		}
		if mapping.Order != 1 {
			t.Errorf("Expected order 1, got %d", mapping.Order)
		}
	})

	t.Run("Should generate correct mapping for nested field", func(t *testing.T) {
		mapping := generateFieldMapping(videoType, "Sponsorship.Amount", 1)
		if mapping == nil {
			t.Fatal("Expected non-nil mapping")
		}

		if mapping.Name != "Sponsorship Amount" {
			t.Errorf("Expected name 'Sponsorship Amount', got %s", mapping.Name)
		}
		if mapping.FieldName != "sponsorship.amount" {
			t.Errorf("Expected fieldName 'sponsorship.amount', got %s", mapping.FieldName)
		}
		if mapping.FieldType != FieldTypeString {
			t.Errorf("Expected fieldType 'string', got %s", mapping.FieldType)
		}
	})

	t.Run("Should return nil for non-existent field", func(t *testing.T) {
		mapping := generateFieldMapping(videoType, "NonExistentField", 1)
		if mapping != nil {
			t.Error("Expected nil mapping for non-existent field")
		}
	})

	t.Run("Should return nil for non-existent nested field", func(t *testing.T) {
		mapping := generateFieldMapping(videoType, "Sponsorship.NonExistent", 1)
		if mapping != nil {
			t.Error("Expected nil mapping for non-existent nested field")
		}
	})
}
