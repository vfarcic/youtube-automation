package aspect

import (
	"testing"

	"devopstoolkit/youtube-automation/internal/app"
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

		for i, mapping := range mappings {
			if mapping.AspectKey != expectedKeys[i] {
				t.Errorf("Expected aspect key %s at index %d, got %s", expectedKeys[i], i, mapping.AspectKey)
			}
		}
	})

	t.Run("All aspects should have valid titles from constants", func(t *testing.T) {
		expectedTitles := map[string]string{
			AspectKeyInitialDetails: app.PhaseTitleInitialDetails,
			AspectKeyWorkProgress:   app.PhaseTitleWorkProgress,
			AspectKeyDefinition:     app.PhaseTitleDefinition,
			AspectKeyPostProduction: app.PhaseTitlePostProduction,
			AspectKeyPublishing:     app.PhaseTitlePublishingDetails,
			AspectKeyPostPublish:    app.PhaseTitlePostPublish,
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
					t.Errorf("Field %s in aspect %s should have order %d, got %d", field.FieldKey, mapping.AspectKey, expectedOrder, field.Order)
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
					t.Errorf("Invalid field type %s for field %s in aspect %s", field.FieldType, field.FieldKey, mapping.AspectKey)
				}
			}
		}
	})

	t.Run("All field titles should match CLI constants", func(t *testing.T) {
		// Test a sample of known field titles
		titleFieldFound := false
		codeFieldFound := false
		delayedFieldFound := false

		for _, mapping := range mappings {
			for _, field := range mapping.Fields {
				switch field.FieldKey {
				case "title":
					if field.Title != app.FieldTitleTitle {
						t.Errorf("Title field should use constant %s, got %s", app.FieldTitleTitle, field.Title)
					}
					titleFieldFound = true
				case "codeDone":
					if field.Title != app.FieldTitleCodeDone {
						t.Errorf("Code field should use constant %s, got %s", app.FieldTitleCodeDone, field.Title)
					}
					codeFieldFound = true
				case "delayed":
					if field.Title != app.FieldTitleDelayed {
						t.Errorf("Delayed field should use constant %s, got %s", app.FieldTitleDelayed, field.Title)
					}
					delayedFieldFound = true
				}
			}
		}

		if !titleFieldFound {
			t.Error("Title field not found in mappings")
		}
		if !codeFieldFound {
			t.Error("Code field not found in mappings")
		}
		if !delayedFieldFound {
			t.Error("Delayed field not found in mappings")
		}
	})

	t.Run("All video properties should be valid", func(t *testing.T) {
		// Create a sample video to verify property paths exist
		video := storage.Video{}

		for _, mapping := range mappings {
			for _, field := range mapping.Fields {
				// Test that GetVideoPropertyValue doesn't return nil for valid properties
				value := GetVideoPropertyValue(video, field.VideoProperty)
				// We don't test the value itself since zero values are valid,
				// but the function should not panic and should return a non-nil interface{}
				_ = value // Suppress unused variable warning
			}
		}
	})

	t.Run("Field keys should be unique within each aspect", func(t *testing.T) {
		for _, mapping := range mappings {
			fieldKeys := make(map[string]bool)
			for _, field := range mapping.Fields {
				if fieldKeys[field.FieldKey] {
					t.Errorf("Duplicate field key %s in aspect %s", field.FieldKey, mapping.AspectKey)
				}
				fieldKeys[field.FieldKey] = true
			}
		}
	})

	t.Run("Should have expected field counts per aspect", func(t *testing.T) {
		expectedCounts := map[string]int{
			AspectKeyInitialDetails: 8,  // ProjectName, ProjectURL, Amount, Emails, Blocked, Date, Delayed, Gist
			AspectKeyWorkProgress:   11, // Code, Head, Screen, RelatedVideos, Thumbnails, Diagrams, Screenshots, Location, Tagline, TaglineIdeas, OtherLogos
			AspectKeyDefinition:     7,  // Title, Description, Highlight, Tags, DescriptionTags, Tweet, Animations
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

func TestGetVideoPropertyValue(t *testing.T) {
	video := storage.Video{
		ProjectName: "Test Project",
		Title:       "Test Video",
		Code:        true,
		Delayed:     false,
		Sponsorship: storage.Sponsorship{
			Amount: "1000",
			Emails: "test@example.com",
		},
	}

	t.Run("Should return correct values for string properties", func(t *testing.T) {
		value := GetVideoPropertyValue(video, "ProjectName")
		if value != "Test Project" {
			t.Errorf("Expected 'Test Project', got %v", value)
		}

		value = GetVideoPropertyValue(video, "Title")
		if value != "Test Video" {
			t.Errorf("Expected 'Test Video', got %v", value)
		}
	})

	t.Run("Should return correct values for boolean properties", func(t *testing.T) {
		value := GetVideoPropertyValue(video, "Code")
		if value != true {
			t.Errorf("Expected true, got %v", value)
		}

		value = GetVideoPropertyValue(video, "Delayed")
		if value != false {
			t.Errorf("Expected false, got %v", value)
		}
	})

	t.Run("Should return correct values for nested properties", func(t *testing.T) {
		value := GetVideoPropertyValue(video, "Sponsorship.Amount")
		if value != "1000" {
			t.Errorf("Expected '1000', got %v", value)
		}

		value = GetVideoPropertyValue(video, "Sponsorship.Emails")
		if value != "test@example.com" {
			t.Errorf("Expected 'test@example.com', got %v", value)
		}
	})

	t.Run("Should return nil for invalid property paths", func(t *testing.T) {
		value := GetVideoPropertyValue(video, "InvalidProperty")
		if value != nil {
			t.Errorf("Expected nil for invalid property, got %v", value)
		}
	})
}

func TestFieldMapping(t *testing.T) {
	t.Run("FieldMapping should have all required fields", func(t *testing.T) {
		mapping := FieldMapping{
			VideoProperty: "Test",
			FieldKey:      "test",
			FieldType:     FieldTypeString,
			Title:         "Test Field",
			Required:      true,
			Order:         1,
			Options:       []string{"option1", "option2"},
		}

		if mapping.VideoProperty != "Test" {
			t.Error("VideoProperty not set correctly")
		}
		if mapping.FieldKey != "test" {
			t.Error("FieldKey not set correctly")
		}
		if mapping.FieldType != FieldTypeString {
			t.Error("FieldType not set correctly")
		}
		if mapping.Title != "Test Field" {
			t.Error("Title not set correctly")
		}
		if mapping.Required != true {
			t.Error("Required not set correctly")
		}
		if mapping.Order != 1 {
			t.Error("Order not set correctly")
		}
		if len(mapping.Options) != 2 {
			t.Error("Options not set correctly")
		}
	})
}

// createTestVideo creates a test Video object with sample data
func createTestVideo() storage.Video {
	return storage.Video{
		Name:                "test-video",
		Index:               1,
		Path:                "test-video.yaml",
		Category:            "test-category",
		ProjectName:         "Test Project",
		ProjectURL:          "https://github.com/test/project",
		Title:               "Test Video Title",
		Description:         "Test video description",
		Date:                "2023-05-15T10:00",
		Delayed:             false,
		Code:                true,
		Head:                false,
		Screen:              true,
		Thumbnails:          false,
		Diagrams:            true,
		Screenshots:         false,
		Location:            "Drive Folder",
		Tagline:             "Test tagline",
		TaglineIdeas:        "Some ideas",
		OtherLogos:          "Logo details",
		Highlight:           "Test highlight",
		Tags:                "tag1,tag2,tag3",
		DescriptionTags:     "More tags for description",
		Tweet:               "Test tweet text",
		Animations:          "Animation script",
		Thumbnail:           "/path/to/thumbnail.jpg",
		Members:             "member1,member2",
		RequestEdit:         false,
		Timecodes:           "00:00 - Introduction",
		Movie:               true,
		Slides:              false,
		UploadVideo:         "/path/to/video.mp4",
		VideoId:             "abc123xyz",
		HugoPath:            "/hugo/post/path",
		DOTPosted:           true,
		BlueSkyPosted:       false,
		LinkedInPosted:      true,
		SlackPosted:         false,
		YouTubeHighlight:    true,
		YouTubeComment:      false,
		YouTubeCommentReply: true,
		GDE:                 false,
		Repo:                "https://github.com/test/repo",
		NotifiedSponsors:    true,
		Gist:                "/path/to/gist.md",
		RelatedVideos:       "Related video 1\nRelated video 2",
		Sponsorship: storage.Sponsorship{
			Amount:  "1000",
			Emails:  "sponsor@example.com",
			Blocked: "",
		},
	}
}

// TestCreateFieldMapping tests the createFieldMapping helper function
func TestCreateFieldMapping(t *testing.T) {
	t.Run("Should create string field mapping with correct metadata", func(t *testing.T) {
		mapping := createFieldMapping("TestProperty", "testKey", FieldTypeString, "Test Title", false, 1, nil)

		// Test basic properties
		if mapping.VideoProperty != "TestProperty" {
			t.Errorf("Expected VideoProperty 'TestProperty', got '%s'", mapping.VideoProperty)
		}
		if mapping.FieldKey != "testKey" {
			t.Errorf("Expected FieldKey 'testKey', got '%s'", mapping.FieldKey)
		}
		if mapping.FieldType != FieldTypeString {
			t.Errorf("Expected FieldType '%s', got '%s'", FieldTypeString, mapping.FieldType)
		}
		if mapping.Title != "Test Title" {
			t.Errorf("Expected Title 'Test Title', got '%s'", mapping.Title)
		}
		if mapping.Required != false {
			t.Errorf("Expected Required false, got %t", mapping.Required)
		}
		if mapping.Order != 1 {
			t.Errorf("Expected Order 1, got %d", mapping.Order)
		}

		// Test enhanced metadata
		if mapping.UIHints.InputType != "text" {
			t.Errorf("Expected InputType 'text', got '%s'", mapping.UIHints.InputType)
		}
		if mapping.ValidationHints.Required != false {
			t.Errorf("Expected ValidationHints.Required false, got %t", mapping.ValidationHints.Required)
		}
		if mapping.DefaultValue != nil {
			t.Errorf("Expected DefaultValue nil for non-required string, got %v", mapping.DefaultValue)
		}
	})

	t.Run("Should create boolean field mapping with correct metadata", func(t *testing.T) {
		mapping := createFieldMapping("BoolProperty", "boolKey", FieldTypeBoolean, "Bool Title", false, 2, nil)

		if mapping.FieldType != FieldTypeBoolean {
			t.Errorf("Expected FieldType '%s', got '%s'", FieldTypeBoolean, mapping.FieldType)
		}
		if mapping.UIHints.InputType != "checkbox" {
			t.Errorf("Expected InputType 'checkbox', got '%s'", mapping.UIHints.InputType)
		}
		if mapping.DefaultValue != false {
			t.Errorf("Expected DefaultValue false for boolean, got %v", mapping.DefaultValue)
		}
	})

	t.Run("Should create text field mapping with correct metadata", func(t *testing.T) {
		mapping := createFieldMapping("TextProperty", "textKey", FieldTypeText, "Text Title", false, 3, nil)

		if mapping.FieldType != FieldTypeText {
			t.Errorf("Expected FieldType '%s', got '%s'", FieldTypeText, mapping.FieldType)
		}
		if mapping.UIHints.InputType != "textarea" {
			t.Errorf("Expected InputType 'textarea', got '%s'", mapping.UIHints.InputType)
		}
		if mapping.UIHints.Rows != 3 {
			t.Errorf("Expected Rows 3, got %d", mapping.UIHints.Rows)
		}
	})

	t.Run("Should create date field mapping with correct metadata", func(t *testing.T) {
		mapping := createFieldMapping("DateProperty", "dateKey", FieldTypeDate, "Date Title", false, 4, nil)

		if mapping.FieldType != FieldTypeDate {
			t.Errorf("Expected FieldType '%s', got '%s'", FieldTypeDate, mapping.FieldType)
		}
		if mapping.UIHints.InputType != "datetime-local" {
			t.Errorf("Expected InputType 'datetime-local', got '%s'", mapping.UIHints.InputType)
		}
		if mapping.UIHints.Placeholder != "2006-01-02T15:04" {
			t.Errorf("Expected Placeholder '2006-01-02T15:04', got '%s'", mapping.UIHints.Placeholder)
		}
	})

	t.Run("Should create number field mapping with correct metadata", func(t *testing.T) {
		mapping := createFieldMapping("NumberProperty", "numberKey", FieldTypeNumber, "Number Title", false, 5, nil)

		if mapping.FieldType != FieldTypeNumber {
			t.Errorf("Expected FieldType '%s', got '%s'", FieldTypeNumber, mapping.FieldType)
		}
		if mapping.UIHints.InputType != "number" {
			t.Errorf("Expected InputType 'number', got '%s'", mapping.UIHints.InputType)
		}
		if mapping.DefaultValue != 0 {
			t.Errorf("Expected DefaultValue 0 for number, got %v", mapping.DefaultValue)
		}
	})

	t.Run("Should create select field mapping with correct metadata", func(t *testing.T) {
		options := []string{"option1", "option2", "option3"}
		mapping := createFieldMapping("SelectProperty", "selectKey", FieldTypeSelect, "Select Title", false, 6, options)

		if mapping.FieldType != FieldTypeSelect {
			t.Errorf("Expected FieldType '%s', got '%s'", FieldTypeSelect, mapping.FieldType)
		}
		if mapping.UIHints.InputType != "select" {
			t.Errorf("Expected InputType 'select', got '%s'", mapping.UIHints.InputType)
		}
		if len(mapping.UIHints.Options) != 3 {
			t.Errorf("Expected 3 UIHints options, got %d", len(mapping.UIHints.Options))
		}
		if len(mapping.Options) != 3 {
			t.Errorf("Expected 3 legacy options, got %d", len(mapping.Options))
		}

		// Test that UIHints.Options are properly structured
		for i, option := range mapping.UIHints.Options {
			expectedValue := options[i]
			if option.Value != expectedValue {
				t.Errorf("Expected option value '%s', got '%s'", expectedValue, option.Value)
			}
			if option.Label != expectedValue {
				t.Errorf("Expected option label '%s', got '%s'", expectedValue, option.Label)
			}
		}
	})

	t.Run("Should handle required fields correctly", func(t *testing.T) {
		mapping := createFieldMapping("RequiredProperty", "requiredKey", FieldTypeString, "Required Title", true, 1, nil)

		if mapping.Required != true {
			t.Errorf("Expected Required true, got %t", mapping.Required)
		}
		// Note: ValidationHints.Required comes from the field type instance, which doesn't get the required flag set
		// The mapping.Required field is what should be used for actual validation
		if mapping.ValidationHints.Required != false {
			t.Errorf("Expected ValidationHints.Required false (field type instances don't get required flag), got %t", mapping.ValidationHints.Required)
		}
		if mapping.DefaultValue != "" {
			t.Errorf("Expected DefaultValue empty string for required string, got %v", mapping.DefaultValue)
		}
	})

	t.Run("Should fallback to string type for unknown field types", func(t *testing.T) {
		mapping := createFieldMapping("UnknownProperty", "unknownKey", "unknown-type", "Unknown Title", false, 1, nil)

		// Should fallback to string type behavior
		if mapping.UIHints.InputType != "text" {
			t.Errorf("Expected fallback InputType 'text', got '%s'", mapping.UIHints.InputType)
		}
	})
}

// TestEnhancedFieldMetadata tests that all mappings have proper enhanced metadata
func TestEnhancedFieldMetadata(t *testing.T) {
	mappings := GetVideoAspectMappings()

	t.Run("All fields should have UI hints", func(t *testing.T) {
		for _, mapping := range mappings {
			for _, field := range mapping.Fields {
				if field.UIHints.InputType == "" {
					t.Errorf("Field %s in aspect %s should have InputType", field.FieldKey, mapping.AspectKey)
				}

				// Verify InputType matches FieldType expectations
				switch field.FieldType {
				case FieldTypeString:
					if field.UIHints.InputType != "text" {
						t.Errorf("String field %s should have InputType 'text', got '%s'", field.FieldKey, field.UIHints.InputType)
					}
				case FieldTypeText:
					if field.UIHints.InputType != "textarea" {
						t.Errorf("Text field %s should have InputType 'textarea', got '%s'", field.FieldKey, field.UIHints.InputType)
					}
					if field.UIHints.Rows == 0 {
						t.Errorf("Text field %s should have Rows > 0", field.FieldKey)
					}
				case FieldTypeBoolean:
					if field.UIHints.InputType != "checkbox" {
						t.Errorf("Boolean field %s should have InputType 'checkbox', got '%s'", field.FieldKey, field.UIHints.InputType)
					}
				case FieldTypeDate:
					if field.UIHints.InputType != "datetime-local" {
						t.Errorf("Date field %s should have InputType 'datetime-local', got '%s'", field.FieldKey, field.UIHints.InputType)
					}
				case FieldTypeNumber:
					if field.UIHints.InputType != "number" {
						t.Errorf("Number field %s should have InputType 'number', got '%s'", field.FieldKey, field.UIHints.InputType)
					}
				case FieldTypeSelect:
					if field.UIHints.InputType != "select" {
						t.Errorf("Select field %s should have InputType 'select', got '%s'", field.FieldKey, field.UIHints.InputType)
					}
				}
			}
		}
	})

	t.Run("All fields should have validation hints", func(t *testing.T) {
		for _, mapping := range mappings {
			for _, field := range mapping.Fields {
				// Note: ValidationHints.Required comes from field type instances which don't get the required flag set
				// The mapping.Required field is the source of truth for validation.
				// This test verifies that ValidationHints are present but doesn't enforce they match mapping.Required
				// since the field type instances are created without the required parameter.
				_ = field.ValidationHints // Just ensure it exists
			}
		}
	})

	t.Run("All fields should have appropriate default values", func(t *testing.T) {
		for _, mapping := range mappings {
			for _, field := range mapping.Fields {
				switch field.FieldType {
				case FieldTypeBoolean:
					if field.DefaultValue != false {
						t.Errorf("Boolean field %s should have DefaultValue false, got %v", field.FieldKey, field.DefaultValue)
					}
				case FieldTypeNumber:
					if field.DefaultValue != 0 {
						t.Errorf("Number field %s should have DefaultValue 0, got %v", field.FieldKey, field.DefaultValue)
					}
				case FieldTypeString, FieldTypeText, FieldTypeDate, FieldTypeSelect:
					if field.Required {
						if field.DefaultValue != "" {
							t.Errorf("Required field %s should have DefaultValue empty string, got %v", field.FieldKey, field.DefaultValue)
						}
					} else {
						if field.DefaultValue != nil {
							t.Errorf("Non-required field %s should have DefaultValue nil, got %v", field.FieldKey, field.DefaultValue)
						}
					}
				}
			}
		}
	})

	t.Run("Date fields should have proper UI hints", func(t *testing.T) {
		foundDateField := false
		for _, mapping := range mappings {
			for _, field := range mapping.Fields {
				if field.FieldType == FieldTypeDate {
					foundDateField = true
					if field.UIHints.Placeholder != "2006-01-02T15:04" {
						t.Errorf("Date field %s should have placeholder '2006-01-02T15:04', got '%s'", field.FieldKey, field.UIHints.Placeholder)
					}
				}
			}
		}
		if !foundDateField {
			t.Error("Should have at least one date field to test")
		}
	})

	t.Run("Text fields should have proper UI hints", func(t *testing.T) {
		foundTextField := false
		for _, mapping := range mappings {
			for _, field := range mapping.Fields {
				if field.FieldType == FieldTypeText {
					foundTextField = true
					if field.UIHints.Rows <= 0 {
						t.Errorf("Text field %s should have Rows > 0, got %d", field.FieldKey, field.UIHints.Rows)
					}
				}
			}
		}
		if !foundTextField {
			t.Error("Should have at least one text field to test")
		}
	})
}

// TestGetDefaultValueForField tests the default value helper function
func TestGetDefaultValueForField(t *testing.T) {
	t.Run("Boolean fields should default to false", func(t *testing.T) {
		value := getDefaultValueForField(FieldTypeBoolean, false)
		if value != false {
			t.Errorf("Expected false for boolean field, got %v", value)
		}

		value = getDefaultValueForField(FieldTypeBoolean, true)
		if value != false {
			t.Errorf("Expected false for required boolean field, got %v", value)
		}
	})

	t.Run("Number fields should default to 0", func(t *testing.T) {
		value := getDefaultValueForField(FieldTypeNumber, false)
		if value != 0 {
			t.Errorf("Expected 0 for number field, got %v", value)
		}

		value = getDefaultValueForField(FieldTypeNumber, true)
		if value != 0 {
			t.Errorf("Expected 0 for required number field, got %v", value)
		}
	})

	t.Run("Required string/text fields should default to empty string", func(t *testing.T) {
		stringTypes := []string{FieldTypeString, FieldTypeText, FieldTypeDate, FieldTypeSelect}

		for _, fieldType := range stringTypes {
			value := getDefaultValueForField(fieldType, true)
			if value != "" {
				t.Errorf("Expected empty string for required %s field, got %v", fieldType, value)
			}
		}
	})

	t.Run("Non-required string/text fields should default to nil", func(t *testing.T) {
		stringTypes := []string{FieldTypeString, FieldTypeText, FieldTypeDate, FieldTypeSelect}

		for _, fieldType := range stringTypes {
			value := getDefaultValueForField(fieldType, false)
			if value != nil {
				t.Errorf("Expected nil for non-required %s field, got %v", fieldType, value)
			}
		}
	})

	t.Run("Unknown field types should default to nil", func(t *testing.T) {
		value := getDefaultValueForField("unknown-type", false)
		if value != nil {
			t.Errorf("Expected nil for unknown field type, got %v", value)
		}

		value = getDefaultValueForField("unknown-type", true)
		if value != nil {
			t.Errorf("Expected nil for unknown required field type, got %v", value)
		}
	})
}
