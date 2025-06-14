package aspect

import (
	"devopstoolkit/youtube-automation/internal/storage"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompletionService_GetFieldCompletionCriteria(t *testing.T) {
	service := NewCompletionService()

	testCases := []struct {
		aspectKey    string
		fieldKey     string
		expectedRule string
		description  string
	}{
		// Initial Details
		{"initial-details", "projectName", "filled_only", "Project name should use filled_only"},
		{"initial-details", "sponsorshipEmails", "conditional_sponsorship", "Sponsorship emails should use conditional_sponsorship logic"},
		{"initial-details", "sponsorshipBlockedReason", "empty_or_filled", "Sponsorship blocked should use empty_or_filled"},
		{"initial-details", "delayed", "false_only", "Delayed should use false_only"},

		// Work Progress
		{"work-progress", "code", "true_only", "Code done should use true_only"},
		{"work-progress", "relatedVideos", "filled_only", "Related videos should use filled_only"},

		// Definition
		{"definition", "title", "filled_only", "Title should use filled_only"},
		{"definition", "description", "filled_only", "Description should use filled_only"},

		// Post-Production
		{"post-production", "requestEdit", "true_only", "Request edit should use true_only"},
		{"post-production", "timecodes", "no_fixme", "Timecodes should use no_fixme"},

		// Publishing
		{"publishing", "videoFilePath", "filled_only", "Video file path should use filled_only"},

		// Post-Publish
		{"post-publish", "dotPosted", "true_only", "DOT posted should use true_only"},
		{"post-publish", "notifiedSponsors", "conditional_sponsors", "Notify sponsors should use conditional_sponsors logic"},

		// Unknown field should return default
		{"unknown-aspect", "unknown-field", "filled_only", "Unknown fields should default to filled_only"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := service.GetFieldCompletionCriteria(tc.aspectKey, tc.fieldKey)
			if result != tc.expectedRule {
				t.Errorf("Expected %s, got %s for %s.%s", tc.expectedRule, result, tc.aspectKey, tc.fieldKey)
			}
		})
	}
}

func TestCompletionService_IsFieldComplete_FilledOnly(t *testing.T) {
	service := NewCompletionService()
	video := storage.Video{}

	testCases := []struct {
		value       interface{}
		expected    bool
		description string
	}{
		{"", false, "Empty string should not be complete"},
		{"-", false, "Dash should not be complete"},
		{"  ", false, "Whitespace only should not be complete"},
		{"test", true, "Non-empty string should be complete"},
		{"valid content", true, "Valid content should be complete"},
		{true, true, "True boolean should be complete"},
		{false, false, "False boolean should not be complete"},
		{nil, false, "Nil value should not be complete"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := service.IsFieldComplete("test", "test", tc.value, video)
			if result != tc.expected {
				t.Errorf("Expected %v for %v (%T), got %v", tc.expected, tc.value, tc.value, result)
			}
		})
	}
}

func TestCompletionService_IsFieldComplete_ConditionalSponsorshipEmails(t *testing.T) {
	service := NewCompletionService()

	testCases := []struct {
		sponsorshipAmount string
		emailValue        string
		expected          bool
		description       string
	}{
		{"", "", true, "No sponsorship amount - emails should be complete"},
		{"N/A", "", true, "N/A sponsorship amount - emails should be complete"},
		{"-", "", true, "Dash sponsorship amount - emails should be complete"},
		{"1000", "", false, "Has sponsorship amount but no emails - should not be complete"},
		{"1000", "sponsor@example.com", true, "Has sponsorship amount and emails - should be complete"},
		{"500", "test@test.com", true, "Has sponsorship and valid emails - should be complete"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			video := storage.Video{
				Sponsorship: storage.Sponsorship{
					Amount: tc.sponsorshipAmount,
				},
			}

			result := service.IsFieldComplete("initial-details", "sponsorshipEmails", tc.emailValue, video)
			if result != tc.expected {
				t.Errorf("Expected %v for sponsorship amount '%s' and email '%s', got %v",
					tc.expected, tc.sponsorshipAmount, tc.emailValue, result)
			}
		})
	}
}

func TestCompletionService_IsFieldComplete_ConditionalNotifySponsors(t *testing.T) {
	service := NewCompletionService()

	testCases := []struct {
		sponsorshipAmount string
		notifyValue       bool
		expected          bool
		description       string
	}{
		{"", false, true, "No sponsorship amount - notification not needed"},
		{"N/A", false, true, "N/A sponsorship amount - notification not needed"},
		{"-", false, true, "Dash sponsorship amount - notification not needed"},
		{"1000", false, false, "Has sponsorship amount but not notified - should not be complete"},
		{"1000", true, true, "Has sponsorship amount and notified - should be complete"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			video := storage.Video{
				Sponsorship: storage.Sponsorship{
					Amount: tc.sponsorshipAmount,
				},
			}

			result := service.IsFieldComplete("post-publish", "notifiedSponsors", tc.notifyValue, video)
			if result != tc.expected {
				t.Errorf("Expected %v for sponsorship amount '%s' and notify value %v, got %v",
					tc.expected, tc.sponsorshipAmount, tc.notifyValue, result)
			}
		})
	}
}

func TestCompletionService_IsFieldComplete_TrueOnly(t *testing.T) {
	service := NewCompletionService()
	video := storage.Video{}

	testCases := []struct {
		value       interface{}
		expected    bool
		description string
	}{
		{true, true, "True boolean should be complete"},
		{false, false, "False boolean should not be complete"},
		{"true", false, "String 'true' should not be complete for boolean field"},
		{1, false, "Integer 1 should not be complete for boolean field"},
		{nil, false, "Nil should not be complete for boolean field"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// Use a field that should have true_only criteria
			result := service.IsFieldComplete("work-progress", "code", tc.value, video)
			if result != tc.expected {
				t.Errorf("Expected %v for %v (%T), got %v", tc.expected, tc.value, tc.value, result)
			}
		})
	}
}

func TestCompletionService_IsFieldComplete_FalseOnly(t *testing.T) {
	service := NewCompletionService()
	video := storage.Video{}

	testCases := []struct {
		value       interface{}
		expected    bool
		description string
	}{
		{false, true, "False boolean should be complete"},
		{true, false, "True boolean should not be complete"},
		{"false", false, "String 'false' should not be complete for boolean field"},
		{0, false, "Integer 0 should not be complete for boolean field"},
		{nil, false, "Nil should not be complete for boolean field"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// Use delayed field which should have false_only criteria
			result := service.IsFieldComplete("initial-details", "delayed", tc.value, video)
			if result != tc.expected {
				t.Errorf("Expected %v for %v (%T), got %v", tc.expected, tc.value, tc.value, result)
			}
		})
	}
}

func TestCompletionService_IsFieldComplete_NoFixme(t *testing.T) {
	service := NewCompletionService()
	video := storage.Video{}

	testCases := []struct {
		value       interface{}
		expected    bool
		description string
	}{
		{"00:00 Intro, 05:00 Main", true, "Valid timecodes should be complete"},
		{"", false, "Empty timecodes should not be complete"},
		{"FIXME: Add timecodes", false, "Timecodes with FIXME should not be complete"},
		{"00:00 Start, FIXME: Add more", false, "Timecodes containing FIXME should not be complete"},
		{"Some content without fixme", true, "Content without FIXME should be complete"},
		{nil, false, "Nil should not be complete"},
		{123, false, "Non-string should not be complete"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// Use timecodes field which should have no_fixme criteria
			result := service.IsFieldComplete("post-production", "timecodes", tc.value, video)
			if result != tc.expected {
				t.Errorf("Expected %v for %v (%T), got %v", tc.expected, tc.value, tc.value, result)
			}
		})
	}
}

func TestCompletionService_IsFieldComplete_EmptyOrFilled(t *testing.T) {
	service := NewCompletionService()
	video := storage.Video{}

	testCases := []struct {
		value       interface{}
		expected    bool
		description string
	}{
		{"", true, "Empty string should be complete (inverse logic)"},
		{"  ", true, "Whitespace should be complete (inverse logic)"},
		{"some content", false, "Non-empty content should not be complete (inverse logic)"},
		{false, true, "False boolean should be complete (inverse logic)"},
		{true, false, "True boolean should not be complete (inverse logic)"},
		{nil, true, "Nil should be complete (inverse logic)"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// Use sponsorshipBlocked field which should have empty_or_filled criteria
			result := service.IsFieldComplete("initial-details", "sponsorshipBlockedReason", tc.value, video)
			if result != tc.expected {
				t.Errorf("Expected %v for %v (%T), got %v", tc.expected, tc.value, tc.value, result)
			}
		})
	}
}

// TestCompletionService_CacheKeySeparatorPreventsCollisions tests that nested field keys
// use proper separators to prevent cache key collisions
func TestCompletionService_CacheKeySeparatorPreventsCollisions(t *testing.T) {
	service := NewCompletionService()

	// Test that sponsorship fields have proper separators in their cache keys
	// This should demonstrate the current bug where "sponsorshipamount" could collide
	// if there were fields like "sponsor" + "shipamount"

	// Check that the cache contains properly separated keys
	// The cache is private, so we'll test indirectly through the behavior

	// First, let's verify that sponsorship fields work correctly
	// (this test will pass even with the current bug, but sets up for the fix)
	criteria := service.GetFieldCompletionCriteria("initial-details", "sponsorshipAmount")
	assert.Equal(t, "filled_only", criteria, "sponsorshipAmount should have filled_only criteria")

	// Test the current behavior and then verify it should be improved
	mappedKey := service.mapFieldKeyForCompletion("sponsorshipAmount")

	// FIXED: Now we expect proper separators to prevent collisions
	expectedKey := "sponsorship.amount" // This should be the fixed behavior
	assert.Equal(t, expectedKey, mappedKey, "Fixed implementation should use separators to prevent cache key collisions")

	// Test multiple fields to show the fixed collision prevention
	testCases := []struct {
		field       string
		expectedKey string // What it should be after fix
	}{
		{"sponsorshipAmount", "sponsorship.amount"},
		{"sponsorshipEmails", "sponsorship.emails"},
		{"sponsorshipBlockedReason", "sponsorship.blocked"},
	}

	for _, tc := range testCases {
		t.Run(tc.field, func(t *testing.T) {
			mapped := service.mapFieldKeyForCompletion(tc.field)
			// Now assert the FIXED behavior
			assert.Equal(t, tc.expectedKey, mapped,
				"Fixed implementation for %s should be %s with proper separator",
				tc.field, tc.expectedKey)
		})
	}
}

// TestCompletionService_FieldMappingConsistency tests that field mappings are consistent
// and don't create ambiguous cache keys
func TestCompletionService_FieldMappingConsistency(t *testing.T) {
	service := NewCompletionService()

	// Test multiple sponsorship fields to ensure they map consistently WITH SEPARATORS
	testCases := []struct {
		fieldKey    string
		expectedKey string
	}{
		{"sponsorshipAmount", "sponsorship.amount"},         // Fixed behavior with separator
		{"sponsorshipEmails", "sponsorship.emails"},         // Fixed behavior with separator
		{"sponsorshipBlockedReason", "sponsorship.blocked"}, // Fixed behavior with separator
	}

	for _, tc := range testCases {
		t.Run(tc.fieldKey, func(t *testing.T) {
			mappedKey := service.mapFieldKeyForCompletion(tc.fieldKey)
			assert.Equal(t, tc.expectedKey, mappedKey,
				"Field %s should map to %s with proper separator", tc.fieldKey, tc.expectedKey)
		})
	}
}
