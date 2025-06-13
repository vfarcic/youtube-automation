package aspect

import (
	"fmt"
	"testing"

	"devopstoolkit/youtube-automation/internal/constants"
)

func TestNewService(t *testing.T) {
	service := NewService()
	if service == nil {
		t.Fatal("NewService() returned nil")
	}
}

func TestGetAspects(t *testing.T) {
	service := NewService()
	metadata := service.GetAspects()

	// Test that we have the expected number of aspects (6 phases)
	expectedAspectCount := 6
	if len(metadata.Aspects) != expectedAspectCount {
		t.Errorf("Expected %d aspects, got %d", expectedAspectCount, len(metadata.Aspects))
	}

	// Test that all aspects have the required fields
	for i, aspect := range metadata.Aspects {
		t.Run(aspect.Key, func(t *testing.T) {
			// Required aspect fields
			if aspect.Key == "" {
				t.Error("Aspect Key is empty")
			}
			if aspect.Title == "" {
				t.Error("Aspect Title is empty")
			}
			if aspect.Description == "" {
				t.Error("Aspect Description is empty")
			}
			if aspect.Endpoint == "" {
				t.Error("Aspect Endpoint is empty")
			}
			if aspect.Icon == "" {
				t.Error("Aspect Icon is empty")
			}
			if aspect.Order == 0 {
				t.Error("Aspect Order is zero (should be 1-based)")
			}

			// Order should match array index + 1
			expectedOrder := i + 1
			if aspect.Order != expectedOrder {
				t.Errorf("Expected Order %d, got %d", expectedOrder, aspect.Order)
			}

			// Test that aspect has fields
			if len(aspect.Fields) == 0 {
				t.Error("Aspect has no fields")
			}

			// Test each field
			for _, field := range aspect.Fields {
				if field.Name == "" {
					t.Error("Field Name is empty")
				}
				if field.Type == "" {
					t.Error("Field Type is empty")
				}
				if field.Description == "" {
					t.Error("Field Description is empty")
				}

				// Validate field types
				validTypes := []string{
					FieldTypeString, FieldTypeText, FieldTypeBoolean,
					FieldTypeDate, FieldTypeNumber, FieldTypeSelect,
				}
				isValidType := false
				for _, validType := range validTypes {
					if field.Type == validType {
						isValidType = true
						break
					}
				}
				if !isValidType {
					t.Errorf("Invalid field type: %s", field.Type)
				}
			}
		})
	}
}

func TestAspectWorkflowOrder(t *testing.T) {
	service := NewService()
	metadata := service.GetAspects()

	expectedWorkflow := []struct {
		order    int
		key      string
		title    string
		endpoint string
	}{
		{1, AspectKeyInitialDetails, constants.PhaseTitleInitialDetails, "/api/videos/{videoName}/initial-details"},
		{2, AspectKeyWorkProgress, constants.PhaseTitleWorkProgress, "/api/videos/{videoName}/work-progress"},
		{3, AspectKeyDefinition, constants.PhaseTitleDefinition, "/api/videos/{videoName}/definition"},
		{4, AspectKeyPostProduction, constants.PhaseTitlePostProduction, "/api/videos/{videoName}/post-production"},
		{5, AspectKeyPublishing, constants.PhaseTitlePublishingDetails, "/api/videos/{videoName}/publishing"},
		{6, AspectKeyPostPublish, constants.PhaseTitlePostPublish, "/api/videos/{videoName}/post-publish"},
	}

	for i, expected := range expectedWorkflow {
		aspect := metadata.Aspects[i]

		if aspect.Order != expected.order {
			t.Errorf("Aspect %d: expected order %d, got %d", i, expected.order, aspect.Order)
		}
		if aspect.Key != expected.key {
			t.Errorf("Aspect %d: expected key %s, got %s", i, expected.key, aspect.Key)
		}
		if aspect.Title != expected.title {
			t.Errorf("Aspect %d: expected title %s, got %s", i, expected.title, aspect.Title)
		}
		if aspect.Endpoint != expected.endpoint {
			t.Errorf("Aspect %d: expected endpoint %s, got %s", i, expected.endpoint, aspect.Endpoint)
		}
	}
}

func TestFieldTitleConsistency(t *testing.T) {
	service := NewService()
	metadata := service.GetAspects()

	// Test that Work Progress fields use constants consistently
	workProgressAspect := metadata.Aspects[1] // Index 1 = Work Progress

	expectedWorkProgressFields := map[string]bool{
		constants.FieldTitleCodeDone:            true,
		constants.FieldTitleTalkingHeadDone:     true,
		constants.FieldTitleScreenRecordingDone: true,
		constants.FieldTitleRelatedVideos:       true,
		constants.FieldTitleThumbnailsDone:      true,
		constants.FieldTitleDiagramsDone:        true,
		constants.FieldTitleScreenshotsDone:     true,
		constants.FieldTitleFilesLocation:       true,
		constants.FieldTitleTagline:             true,
		constants.FieldTitleTaglineIdeas:        true,
		constants.FieldTitleOtherLogos:          true,
	}

	for _, field := range workProgressAspect.Fields {
		if expectedWorkProgressFields[field.Name] {
			// This field should use constants - mark as found
			delete(expectedWorkProgressFields, field.Name)
		}
	}

	// Check if any expected fields were missing
	if len(expectedWorkProgressFields) > 0 {
		for missingField := range expectedWorkProgressFields {
			t.Errorf("Work Progress aspect missing expected field: %s", missingField)
		}
	}
}

func TestPostProductionFieldConsistency(t *testing.T) {
	service := NewService()
	metadata := service.GetAspects()

	// Test that Post-Production fields use constants
	postProdAspect := metadata.Aspects[3] // Index 3 = Post-Production

	expectedFields := []string{
		constants.FieldTitleThumbnailPath,
		constants.FieldTitleMembers,
		constants.FieldTitleRequestEdit,
		constants.FieldTitleTimecodes,
		constants.FieldTitleMovieDone,
		constants.FieldTitleSlidesDone,
	}

	actualFieldNames := make([]string, len(postProdAspect.Fields))
	for i, field := range postProdAspect.Fields {
		actualFieldNames[i] = field.Name
	}

	for _, expectedField := range expectedFields {
		found := false
		for _, actualField := range actualFieldNames {
			if actualField == expectedField {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Post-Production aspect missing expected field: %s", expectedField)
		}
	}
}

func TestPostPublishFieldConsistency(t *testing.T) {
	service := NewService()
	metadata := service.GetAspects()

	// Test that Post-Publish fields use constants
	postPublishAspect := metadata.Aspects[5] // Index 5 = Post-Publish

	expectedFields := []string{
		constants.FieldTitleDOTPosted,
		constants.FieldTitleBlueSkyPosted,
		constants.FieldTitleLinkedInPosted,
		constants.FieldTitleSlackPosted,
		constants.FieldTitleYouTubeHighlight,
		constants.FieldTitleYouTubeComment,
		constants.FieldTitleYouTubeCommentReply,
		constants.FieldTitleGDEPosted,
		constants.FieldTitleCodeRepository,
		constants.FieldTitleNotifySponsors,
	}

	actualFieldNames := make([]string, len(postPublishAspect.Fields))
	for i, field := range postPublishAspect.Fields {
		actualFieldNames[i] = field.Name
	}

	for _, expectedField := range expectedFields {
		found := false
		for _, actualField := range actualFieldNames {
			if actualField == expectedField {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Post-Publish aspect missing expected field: %s", expectedField)
		}
	}
}

func TestRequiredFields(t *testing.T) {
	service := NewService()
	metadata := service.GetAspects()

	// Test required fields for Initial Details (based on actual mapping)
	// According to the mapping, no fields are actually required in initial details

	// Test required fields for Definition (based on actual mapping)
	definitionAspect := metadata.Aspects[2]
	requiredFieldsDefinition := []string{
		constants.FieldTitleTitle, // Required in mapping
	}

	for _, requiredField := range requiredFieldsDefinition {
		found := false
		for _, field := range definitionAspect.Fields {
			if field.Name == requiredField && field.Required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required field %s not found or not marked as required in aspect %s", requiredField, definitionAspect.Key)
		}
	}
}

// Tests for new two-endpoint functionality

func TestGetAspectsOverview(t *testing.T) {
	service := NewService()
	overview := service.GetAspectsOverview()

	t.Run("Should return correct number of aspects", func(t *testing.T) {
		expectedCount := 6
		if len(overview.Aspects) != expectedCount {
			t.Errorf("Expected %d aspects, got %d", expectedCount, len(overview.Aspects))
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

		for i, expectedKey := range expectedKeys {
			if i >= len(overview.Aspects) {
				t.Fatalf("Missing aspect at index %d", i)
			}
			if overview.Aspects[i].Key != expectedKey {
				t.Errorf("Expected aspect key '%s' at index %d, got '%s'", expectedKey, i, overview.Aspects[i].Key)
			}
		}
	})

	t.Run("Should have correct field counts for each aspect", func(t *testing.T) {
		// Test expected field counts based on actual mapping
		expectedFieldCounts := map[string]int{
			AspectKeyInitialDetails: 8,  // actual count from mapping
			AspectKeyWorkProgress:   11, // actual count from mapping
			AspectKeyDefinition:     7,  // actual count from mapping
			AspectKeyPostProduction: 6,  // actual count from mapping
			AspectKeyPublishing:     3,  // actual count from mapping
			AspectKeyPostPublish:    10, // actual count from mapping
		}

		for _, aspect := range overview.Aspects {
			expectedCount, exists := expectedFieldCounts[aspect.Key]
			if !exists {
				t.Errorf("Unexpected aspect key: %s", aspect.Key)
				continue
			}
			if aspect.FieldCount != expectedCount {
				t.Errorf("Expected %d fields for aspect '%s', got %d", expectedCount, aspect.Key, aspect.FieldCount)
			}
			// TDD: Verify CompletedFieldCount is present and zero for now
			if aspect.CompletedFieldCount != 0 {
				t.Errorf("Expected CompletedFieldCount to be 0 for aspect '%s', got %d", aspect.Key, aspect.CompletedFieldCount)
			}
		}
	})

	t.Run("Should have proper order values", func(t *testing.T) {
		for i, aspect := range overview.Aspects {
			expectedOrder := i + 1
			if aspect.Order != expectedOrder {
				t.Errorf("Expected order %d for aspect '%s', got %d", expectedOrder, aspect.Key, aspect.Order)
			}
		}
	})

	t.Run("Should have non-empty titles and descriptions", func(t *testing.T) {
		for _, aspect := range overview.Aspects {
			if aspect.Title == "" {
				t.Errorf("Aspect '%s' has empty title", aspect.Key)
			}
			if aspect.Description == "" {
				t.Errorf("Aspect '%s' has empty description", aspect.Key)
			}
		}
	})

	t.Run("Should have valid endpoint patterns", func(t *testing.T) {
		for _, aspect := range overview.Aspects {
			if aspect.Endpoint == "" {
				t.Errorf("Aspect '%s' has empty endpoint", aspect.Key)
			}
			// Endpoints should contain the videoName placeholder
			if !contains(aspect.Endpoint, "{videoName}") {
				t.Errorf("Aspect '%s' endpoint '%s' should contain {videoName} placeholder", aspect.Key, aspect.Endpoint)
			}
		}
	})
}

func TestGetAspectFields(t *testing.T) {
	service := NewService()

	t.Run("Should return fields for a valid aspect key", func(t *testing.T) {
		aspectFields, err := service.GetAspectFields(AspectKeyInitialDetails)
		if err != nil {
			t.Fatalf("Expected no error for valid aspect, got: %v", err)
		}

		if aspectFields.AspectKey != AspectKeyInitialDetails {
			t.Errorf("Expected aspect key '%s', got '%s'", AspectKeyInitialDetails, aspectFields.AspectKey)
		}

		if aspectFields.AspectTitle == "" {
			t.Error("Expected non-empty aspect title")
		}

		if len(aspectFields.Fields) == 0 {
			t.Error("Expected at least one field")
		}
	})

	t.Run("Should return error for non-existent aspect", func(t *testing.T) {
		_, err := service.GetAspectFields("non-existent-aspect")
		if err == nil {
			t.Error("Expected error for non-existent aspect")
		}

		if err != ErrAspectNotFound {
			t.Errorf("Expected ErrAspectNotFound, got: %v", err)
		}
	})

	t.Run("Should validate field structure for all aspects", func(t *testing.T) {
		aspectKeys := []string{
			AspectKeyInitialDetails,
			AspectKeyWorkProgress,
			AspectKeyDefinition,
			AspectKeyPostProduction,
			AspectKeyPublishing,
			AspectKeyPostPublish,
		}

		for _, aspectKey := range aspectKeys {
			aspectFields, err := service.GetAspectFields(aspectKey)
			if err != nil {
				t.Errorf("Error getting fields for aspect '%s': %v", aspectKey, err)
				continue
			}

			t.Run(fmt.Sprintf("Fields for %s should have proper order", aspectKey), func(t *testing.T) {
				for i, field := range aspectFields.Fields {
					expectedOrder := i + 1
					if field.Order != expectedOrder {
						t.Errorf("Expected field order %d for field '%s' in aspect '%s', got %d", expectedOrder, field.Name, aspectKey, field.Order)
					}
				}
			})

			t.Run(fmt.Sprintf("Fields for %s should have valid types", aspectKey), func(t *testing.T) {
				validTypes := map[string]bool{
					FieldTypeString:  true,
					FieldTypeText:    true,
					FieldTypeBoolean: true,
					FieldTypeDate:    true,
					FieldTypeNumber:  true,
					FieldTypeSelect:  true,
				}

				for _, field := range aspectFields.Fields {
					if !validTypes[field.Type] {
						t.Errorf("Invalid field type '%s' for field '%s' in aspect '%s'", field.Type, field.Name, aspectKey)
					}
				}
			})

			t.Run(fmt.Sprintf("Fields for %s should have non-empty names", aspectKey), func(t *testing.T) {
				for _, field := range aspectFields.Fields {
					if field.Name == "" {
						t.Errorf("Field in aspect '%s' has empty name", aspectKey)
					}
				}
			})
		}
	})

	t.Run("Should return specific fields for work-progress aspect", func(t *testing.T) {
		result, err := service.GetAspectFields(AspectKeyWorkProgress)
		if err != nil {
			t.Fatalf("GetAspectFields failed: %v", err)
		}

		expectedFieldCount := 11 // updated to match actual mapping
		if len(result.Fields) != expectedFieldCount {
			t.Errorf("Expected %d fields for work-progress, got %d", expectedFieldCount, len(result.Fields))
		}

		// Test the actual field names from mapping (using constants)
		expectedFieldNames := []string{
			constants.FieldTitleCodeDone,
			constants.FieldTitleTalkingHeadDone,
			constants.FieldTitleScreenRecordingDone,
			constants.FieldTitleRelatedVideos,
			constants.FieldTitleThumbnailsDone,
			constants.FieldTitleDiagramsDone,
			constants.FieldTitleScreenshotsDone,
			constants.FieldTitleFilesLocation,
			constants.FieldTitleTagline,
			constants.FieldTitleTaglineIdeas,
			constants.FieldTitleOtherLogos,
		}

		for i, expectedName := range expectedFieldNames {
			if i >= len(result.Fields) {
				t.Errorf("Missing field at index %d: expected %s", i, expectedName)
				continue
			}
			if result.Fields[i].Name != expectedName {
				t.Errorf("Expected field name '%s' at index %d, got '%s'", expectedName, i, result.Fields[i].Name)
			}
		}
	})
}

func TestGetAspectFieldsMatchesGetAspects(t *testing.T) {
	service := NewService()

	// Get both full aspects and overview
	fullAspects := service.GetAspects()
	overview := service.GetAspectsOverview()

	t.Run("Field counts should match between overview and individual calls", func(t *testing.T) {
		for i, overviewAspect := range overview.Aspects {
			aspectFields, err := service.GetAspectFields(overviewAspect.Key)
			if err != nil {
				t.Errorf("Error getting fields for aspect '%s': %v", overviewAspect.Key, err)
				continue
			}

			if len(aspectFields.Fields) != overviewAspect.FieldCount {
				t.Errorf("Field count mismatch for aspect '%s': overview says %d, GetAspectFields returns %d",
					overviewAspect.Key, overviewAspect.FieldCount, len(aspectFields.Fields))
			}

			// Also verify against full aspects
			if i < len(fullAspects.Aspects) {
				fullAspect := fullAspects.Aspects[i]
				if len(aspectFields.Fields) != len(fullAspect.Fields) {
					t.Errorf("Field count mismatch for aspect '%s': GetAspects has %d fields, GetAspectFields returns %d",
						overviewAspect.Key, len(fullAspect.Fields), len(aspectFields.Fields))
				}
			}
		}
	})

	t.Run("Field content should match between GetAspects and GetAspectFields", func(t *testing.T) {
		for _, fullAspect := range fullAspects.Aspects {
			aspectFields, err := service.GetAspectFields(fullAspect.Key)
			if err != nil {
				t.Errorf("Error getting fields for aspect '%s': %v", fullAspect.Key, err)
				continue
			}

			for j, fullField := range fullAspect.Fields {
				if j >= len(aspectFields.Fields) {
					t.Errorf("Missing field at index %d for aspect '%s'", j, fullAspect.Key)
					continue
				}

				individualField := aspectFields.Fields[j]
				if fullField.Name != individualField.Name {
					t.Errorf("Field name mismatch at index %d for aspect '%s': full='%s', individual='%s'",
						j, fullAspect.Key, fullField.Name, individualField.Name)
				}
				if fullField.Type != individualField.Type {
					t.Errorf("Field type mismatch for field '%s' in aspect '%s': full='%s', individual='%s'",
						fullField.Name, fullAspect.Key, fullField.Type, individualField.Type)
				}
				if fullField.Required != individualField.Required {
					t.Errorf("Field required mismatch for field '%s' in aspect '%s': full=%v, individual=%v",
						fullField.Name, fullAspect.Key, fullField.Required, individualField.Required)
				}
			}
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr ||
		len(s) > len(substr) && s[:len(substr)] == substr ||
		(len(s) > len(substr) && func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

// TestServiceIntegration tests the complete pipeline from mapping to service
func TestServiceIntegration(t *testing.T) {
	service := NewService()

	t.Run("GetAspectsOverview should return correct field counts", func(t *testing.T) {
		overview := service.GetAspectsOverview()

		if len(overview.Aspects) == 0 {
			t.Fatal("Expected aspects in overview, got none")
		}

		// Check that field counts match the actual mappings
		for _, aspect := range overview.Aspects {
			aspectFields, err := service.GetAspectFields(aspect.Key)
			if err != nil {
				t.Errorf("Failed to get fields for aspect %s: %v", aspect.Key, err)
				continue
			}

			if aspect.FieldCount != len(aspectFields.Fields) {
				t.Errorf("Aspect %s field count mismatch: overview says %d, actual is %d",
					aspect.Key, aspect.FieldCount, len(aspectFields.Fields))
			}
		}
	})

	t.Run("GetAspectFields should include enhanced metadata", func(t *testing.T) {
		// Test initial-details aspect which has various field types
		aspectFields, err := service.GetAspectFields("initial-details")
		if err != nil {
			t.Fatalf("Failed to get initial-details fields: %v", err)
		}

		if len(aspectFields.Fields) == 0 {
			t.Fatal("Expected fields in initial-details aspect, got none")
		}

		// Check that all fields have enhanced metadata
		for _, field := range aspectFields.Fields {
			// Every field should have UIHints
			if field.UIHints.InputType == "" {
				t.Errorf("Field %s missing UIHints.InputType", field.Name)
			}

			// Check specific field types have appropriate UI hints
			switch field.Type {
			case FieldTypeString:
				if field.UIHints.InputType != "text" {
					t.Errorf("String field %s should have InputType 'text', got '%s'",
						field.Name, field.UIHints.InputType)
				}
			case FieldTypeDate:
				if field.UIHints.InputType != "datetime-local" {
					t.Errorf("Date field %s should have InputType 'datetime-local', got '%s'",
						field.Name, field.UIHints.InputType)
				}
				if field.UIHints.Placeholder != "2006-01-02T15:04" {
					t.Errorf("Date field %s should have placeholder '2006-01-02T15:04', got '%s'",
						field.Name, field.UIHints.Placeholder)
				}
			case FieldTypeBoolean:
				if field.UIHints.InputType != "checkbox" {
					t.Errorf("Boolean field %s should have InputType 'checkbox', got '%s'",
						field.Name, field.UIHints.InputType)
				}
			case FieldTypeText:
				if field.UIHints.InputType != "textarea" {
					t.Errorf("Text field %s should have InputType 'textarea', got '%s'",
						field.Name, field.UIHints.InputType)
				}
				if field.UIHints.Rows != 3 {
					t.Errorf("Text field %s should have Rows 3, got %d",
						field.Name, field.UIHints.Rows)
				}
			}

			// ValidationHints should be present (even if just defaults)
			// Note: the ValidationHints.Required comes from field type instances, not mapping.Required
			_ = field.ValidationHints // Just verify it exists
		}
	})

	t.Run("Enhanced metadata should be consistent across all aspects", func(t *testing.T) {
		overview := service.GetAspectsOverview()

		for _, aspectSummary := range overview.Aspects {
			aspectFields, err := service.GetAspectFields(aspectSummary.Key)
			if err != nil {
				t.Errorf("Failed to get fields for aspect %s: %v", aspectSummary.Key, err)
				continue
			}

			for _, field := range aspectFields.Fields {
				// Verify each field has the basic enhanced metadata structure
				if field.UIHints.InputType == "" {
					t.Errorf("Aspect %s field %s missing UIHints.InputType",
						aspectSummary.Key, field.Name)
				}

				// Verify field type consistency
				switch field.Type {
				case FieldTypeString, FieldTypeText, FieldTypeBoolean, FieldTypeDate, FieldTypeNumber, FieldTypeSelect:
					// These are valid types
				default:
					t.Errorf("Aspect %s field %s has unknown field type: %s",
						aspectSummary.Key, field.Name, field.Type)
				}
			}
		}
	})

	t.Run("Select fields should have options and proper UI hints", func(t *testing.T) {
		// Look for any select fields across all aspects
		overview := service.GetAspectsOverview()
		selectFieldFound := false

		for _, aspectSummary := range overview.Aspects {
			aspectFields, err := service.GetAspectFields(aspectSummary.Key)
			if err != nil {
				continue
			}

			for _, field := range aspectFields.Fields {
				if field.Type == FieldTypeSelect {
					selectFieldFound = true

					if field.UIHints.InputType != "select" {
						t.Errorf("Select field %s should have InputType 'select', got '%s'",
							field.Name, field.UIHints.InputType)
					}

					// Note: Options for select fields are in field.Options.Values, not field.DefaultValue
					if len(field.Options.Values) == 0 {
						t.Errorf("Select field %s should have options in field.Options.Values", field.Name)
					}
				}
			}
		}

		// This test may not fail if there are no select fields in the current mappings,
		// which is fine - it's future-proofing
		if !selectFieldFound {
			t.Log("No select fields found in current mappings - test is informational")
		}
	})

	t.Run("Field order should be preserved", func(t *testing.T) {
		aspectFields, err := service.GetAspectFields("initial-details")
		if err != nil {
			t.Fatalf("Failed to get initial-details fields: %v", err)
		}

		// Fields should be ordered by their Order property
		for i := 1; i < len(aspectFields.Fields); i++ {
			if aspectFields.Fields[i].Order <= aspectFields.Fields[i-1].Order {
				t.Errorf("Fields not properly ordered: field at index %d has order %d, previous has order %d",
					i, aspectFields.Fields[i].Order, aspectFields.Fields[i-1].Order)
			}
		}
	})
}

// TestFieldTypeToUIHintsMapping verifies that createFieldMapping properly converts field types to UI hints
func TestFieldTypeToUIHintsMapping(t *testing.T) {
	tests := []struct {
		fieldType    string
		expectedUI   string
		expectedRows int
	}{
		{FieldTypeString, "text", 0},
		{FieldTypeText, "textarea", 3},
		{FieldTypeBoolean, "checkbox", 0},
		{FieldTypeDate, "datetime-local", 0},
		{FieldTypeNumber, "number", 0},
		{FieldTypeSelect, "select", 0},
	}

	for _, test := range tests {
		t.Run("Field type "+test.fieldType, func(t *testing.T) {
			mapping := createFieldMapping("TestProp", "testKey", test.fieldType, "Test Title", false, 1, nil)

			if mapping.UIHints.InputType != test.expectedUI {
				t.Errorf("Expected InputType '%s' for %s, got '%s'",
					test.expectedUI, test.fieldType, mapping.UIHints.InputType)
			}

			if test.expectedRows > 0 && mapping.UIHints.Rows != test.expectedRows {
				t.Errorf("Expected Rows %d for %s, got %d",
					test.expectedRows, test.fieldType, mapping.UIHints.Rows)
			}
		})
	}
}

func TestService_GetAspectFields_IncludesCompletionCriteria(t *testing.T) {
	service := NewService()

	// Test initial-details aspect
	aspectFields, err := service.GetAspectFields("initial-details")
	if err != nil {
		t.Fatalf("Failed to get aspect fields: %v", err)
	}

	if aspectFields == nil {
		t.Fatal("Expected aspect fields, got nil")
	}

	// Verify some specific fields have the expected completion criteria
	expectedCriteria := map[string]string{
		"Project Name":                         CompletionCriteriaFilledOnly,
		"Sponsorship Amount":                   CompletionCriteriaFilledOnly,
		"Sponsorship Emails (comma separated)": CompletionCriteriaConditional,
		"Sponsorship Blocked Reason":           CompletionCriteriaEmptyOrFilled,
		"Delayed":                              CompletionCriteriaFalseOnly,
	}

	fieldMap := make(map[string]Field)
	for _, field := range aspectFields.Fields {
		fieldMap[field.Name] = field
	}

	for fieldName, expectedCriteria := range expectedCriteria {
		field, exists := fieldMap[fieldName]
		if !exists {
			t.Errorf("Expected field %s not found", fieldName)
			continue
		}

		if field.CompletionCriteria == "" {
			t.Errorf("Field %s missing completion criteria", fieldName)
			continue
		}

		if field.CompletionCriteria != expectedCriteria {
			t.Errorf("Field %s: expected completion criteria %s, got %s",
				fieldName, expectedCriteria, field.CompletionCriteria)
		}
	}
}

func TestService_GetAspects_IncludesCompletionCriteria(t *testing.T) {
	service := NewService()

	metadata := service.GetAspects()

	if len(metadata.Aspects) == 0 {
		t.Fatal("Expected aspects, got none")
	}

	// Find initial-details aspect
	var initialDetailsAspect *Aspect
	for i := range metadata.Aspects {
		if metadata.Aspects[i].Key == "initial-details" {
			initialDetailsAspect = &metadata.Aspects[i]
			break
		}
	}

	if initialDetailsAspect == nil {
		t.Fatal("Expected to find initial-details aspect")
	}

	// Verify fields have completion criteria
	foundFieldWithCriteria := false
	for _, field := range initialDetailsAspect.Fields {
		if field.CompletionCriteria != "" {
			foundFieldWithCriteria = true
			break
		}
	}

	if !foundFieldWithCriteria {
		t.Error("Expected at least one field to have completion criteria")
	}

	// Check specific field
	for _, field := range initialDetailsAspect.Fields {
		if field.Name == "Project Name" {
			if field.CompletionCriteria != CompletionCriteriaFilledOnly {
				t.Errorf("Expected Project Name to have %s criteria, got %s",
					CompletionCriteriaFilledOnly, field.CompletionCriteria)
			}
			break
		}
	}
}

func TestService_GetAspectFields_AllAspects(t *testing.T) {
	service := NewService()

	aspectKeys := []string{
		"initial-details",
		"work-progress",
		"definition",
		"post-production",
		"publishing",
		"post-publish",
	}

	for _, aspectKey := range aspectKeys {
		t.Run(aspectKey, func(t *testing.T) {
			aspectFields, err := service.GetAspectFields(aspectKey)
			if err != nil {
				t.Fatalf("Failed to get aspect fields for %s: %v", aspectKey, err)
			}

			if len(aspectFields.Fields) == 0 {
				t.Errorf("Expected fields for aspect %s, got none", aspectKey)
				return
			}

			// Verify all fields have completion criteria
			for _, field := range aspectFields.Fields {
				if field.CompletionCriteria == "" {
					t.Errorf("Field %s in aspect %s missing completion criteria", field.Name, aspectKey)
				}

				// Verify completion criteria is a valid value
				validCriteria := []string{
					CompletionCriteriaFilledOnly,
					CompletionCriteriaEmptyOrFilled,
					CompletionCriteriaFilledRequired,
					CompletionCriteriaTrueOnly,
					CompletionCriteriaFalseOnly,
					CompletionCriteriaConditional,
					CompletionCriteriaNoFixme,
				}

				found := false
				for _, valid := range validCriteria {
					if field.CompletionCriteria == valid {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("Field %s in aspect %s has invalid completion criteria: %s",
						field.Name, aspectKey, field.CompletionCriteria)
				}
			}
		})
	}
}

func TestService_GetAspectFields_UnknownAspect(t *testing.T) {
	service := NewService()

	aspectFields, err := service.GetAspectFields("unknown-aspect")
	if err == nil {
		t.Error("Expected error for unknown aspect, got nil")
	}

	if aspectFields != nil {
		t.Error("Expected nil aspect fields for unknown aspect, got non-nil")
	}

	if err != ErrAspectNotFound {
		t.Errorf("Expected ErrAspectNotFound, got %v", err)
	}
}
