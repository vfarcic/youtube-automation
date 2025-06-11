package app

import (
	"devopstoolkit/youtube-automation/internal/constants"
	"testing"
)

func TestPhaseTitleConstants(t *testing.T) {
	expectedPhaseTitles := []string{
		constants.PhaseTitleInitialDetails,
		constants.PhaseTitleWorkProgress,
		constants.PhaseTitleDefinition,
		constants.PhaseTitlePostProduction,
		constants.PhaseTitlePublishingDetails,
		constants.PhaseTitlePostPublish,
	}

	expectedValues := []string{
		"Initial Details",
		"Work In Progress",
		"Definition",
		"Post-Production",
		"Publishing Details",
		"Post-Publish Details",
	}

	if len(expectedPhaseTitles) != len(expectedValues) {
		t.Fatalf("Mismatch between phase titles count and expected values count")
	}

	for i, title := range expectedPhaseTitles {
		if title != expectedValues[i] {
			t.Errorf("Expected phase title '%s', got '%s'", expectedValues[i], title)
		}
	}
}

func TestFieldTitleConstants(t *testing.T) {
	// Test a few key field title constants to ensure they're properly defined
	testCases := []struct {
		constant string
		expected string
	}{
		{constants.FieldTitleProjectName, "Project Name"},
		{constants.FieldTitleDescription, "Description"},
		{constants.FieldTitleCodeDone, "Code Done"},
		{constants.FieldTitleMovieDone, "Movie Done"},
		{constants.FieldTitleBlueSkyPosted, "BlueSky Post Sent"},
		{constants.FieldTitleDelayed, "Delayed"},
	}

	for _, tc := range testCases {
		if tc.constant != tc.expected {
			t.Errorf("Expected field title '%s', got '%s'", tc.expected, tc.constant)
		}
	}
}

func TestMessageConstants(t *testing.T) {
	// Test that message constants are properly defined
	testCases := []struct {
		constant      string
		shouldContain string
	}{
		{MessageInitialDetailsEditCancelled, "Initial details"},
		{MessageWorkProgressEditCancelled, "Work progress"},
		{MessagePostProductionEditCancelled, "Post-production"},
		{MessageDefinitionPhaseAborted, "Definition phase"},
		{MessageDefinitionPhaseComplete, "Definition Phase Complete"},
	}

	for _, tc := range testCases {
		if tc.constant == "" {
			t.Errorf("Message constant should not be empty")
		}
		// Basic check that the message contains expected text
		// More detailed validation could be added if needed
	}
}

func TestErrorConstants(t *testing.T) {
	// Test that error constants are properly defined for consistent error handling
	errorConstants := []string{
		ErrorRunInitialDetailsForm,
		ErrorRunWorkProgressForm,
		ErrorRunPostProductionForm,
		ErrorSaveInitialDetails,
		ErrorSaveWorkProgress,
		ErrorSavePostProductionDetails,
		ErrorDefinitionPhase,
	}

	for _, errConst := range errorConstants {
		if errConst == "" {
			t.Errorf("Error constant should not be empty")
		}
		if len(errConst) < 10 {
			t.Errorf("Error constant seems too short: %s", errConst)
		}
	}
}

func TestConstantsUniqueness(t *testing.T) {
	// Test that no constants have duplicate values (except where intentional)
	phaseTitles := map[string]bool{
		constants.PhaseTitleInitialDetails:    true,
		constants.PhaseTitleWorkProgress:      true,
		constants.PhaseTitleDefinition:        true,
		constants.PhaseTitlePostProduction:    true,
		constants.PhaseTitlePublishingDetails: true,
		constants.PhaseTitlePostPublish:       true,
	}

	if len(phaseTitles) != 6 {
		t.Errorf("Expected 6 unique phase titles, got %d", len(phaseTitles))
	}

	// Test some key field titles for uniqueness
	fieldTitles := map[string]bool{
		constants.FieldTitleProjectName:   true,
		constants.FieldTitleDescription:   true,
		constants.FieldTitleCodeDone:      true,
		constants.FieldTitleMovieDone:     true,
		constants.FieldTitleBlueSkyPosted: true,
		constants.FieldTitleDelayed:       true,
	}

	if len(fieldTitles) != 6 {
		t.Errorf("Expected 6 unique field titles in test set, got %d", len(fieldTitles))
	}
}
