package aspect

import (
	"devopstoolkit/youtube-automation/internal/storage"
	"strings"
)

// CompletionService provides field-level completion criteria logic
// This service extracts the completion logic from the video manager functions
// to enable both API and CLI to use the same field completion rules
type CompletionService struct{}

// NewCompletionService creates a new completion service
func NewCompletionService() *CompletionService {
	return &CompletionService{}
}

// GetFieldCompletionCriteria returns the completion criteria for a specific field
// This maps field keys to their completion criteria based on the existing logic in video manager
func (s *CompletionService) GetFieldCompletionCriteria(aspectKey, fieldKey string) string {
	// Map field keys to their completion criteria based on existing video manager logic
	// Updated to use the new reflection-based field names (JSON paths)
	fieldCompletionMap := map[string]map[string]string{
		"initial-details": {
			"projectName":         CompletionCriteriaFilledOnly,    // non-empty, not "-"
			"projectURL":          CompletionCriteriaFilledOnly,    // non-empty, not "-"
			"sponsorship.amount":  CompletionCriteriaFilledOnly,    // non-empty, not "-"
			"sponsorship.emails":  CompletionCriteriaConditional,   // special logic: complete if sponsorshipAmount is empty/N/A/- OR if emails has content
			"sponsorship.blocked": CompletionCriteriaEmptyOrFilled, // complete when empty (no blocking)
			"date":                CompletionCriteriaFilledOnly,    // non-empty, not "-"
			"delayed":             CompletionCriteriaFalseOnly,     // complete when false (not delayed)
			"gist":                CompletionCriteriaFilledOnly,    // non-empty, not "-"
		},
		"work-progress": {
			"codeDone":            CompletionCriteriaTrueOnly,   // true when complete
			"talkingHeadDone":     CompletionCriteriaTrueOnly,   // true when complete
			"screenRecordingDone": CompletionCriteriaTrueOnly,   // true when complete
			"relatedVideos":       CompletionCriteriaFilledOnly, // non-empty, not "-"
			"thumbnailsDone":      CompletionCriteriaTrueOnly,   // true when complete
			"diagramsDone":        CompletionCriteriaTrueOnly,   // true when complete
			"screenshotsDone":     CompletionCriteriaTrueOnly,   // true when complete
			"filesLocation":       CompletionCriteriaFilledOnly, // non-empty, not "-"
			"tagline":             CompletionCriteriaFilledOnly, // non-empty, not "-"
			"taglineIdeas":        CompletionCriteriaFilledOnly, // non-empty, not "-"
			"otherLogos":          CompletionCriteriaFilledOnly, // non-empty, not "-"
		},
		"definition": {
			"title":            CompletionCriteriaFilledOnly, // non-empty, not "-"
			"description":      CompletionCriteriaFilledOnly, // non-empty, not "-"
			"highlight":        CompletionCriteriaFilledOnly, // non-empty, not "-"
			"tags":             CompletionCriteriaFilledOnly, // non-empty, not "-"
			"descriptionTags":  CompletionCriteriaFilledOnly, // non-empty, not "-"
			"tweet":            CompletionCriteriaFilledOnly, // non-empty, not "-"
			"animationsScript": CompletionCriteriaFilledOnly, // non-empty, not "-"
		},
		"post-production": {
			"thumbnailPath": CompletionCriteriaFilledOnly, // non-empty, not "-"
			"members":       CompletionCriteriaFilledOnly, // non-empty, not "-"
			"requestEdit":   CompletionCriteriaTrueOnly,   // true when complete
			"timecodes":     CompletionCriteriaNoFixme,    // complete when content doesn't contain "FIXME:"
			"movieDone":     CompletionCriteriaTrueOnly,   // true when complete
			"slidesDone":    CompletionCriteriaTrueOnly,   // true when complete
		},
		"publishing": {
			"videoFilePath":  CompletionCriteriaFilledOnly, // non-empty, not "-"
			"youTubeVideoId": CompletionCriteriaFilledOnly, // non-empty, not "-"
			"hugoPostPath":   CompletionCriteriaFilledOnly, // non-empty, not "-"
		},
		"post-publish": {
			"dotPosted":           CompletionCriteriaTrueOnly,    // true when complete
			"blueSkyPosted":       CompletionCriteriaTrueOnly,    // true when complete
			"linkedInPosted":      CompletionCriteriaTrueOnly,    // true when complete
			"slackPosted":         CompletionCriteriaTrueOnly,    // true when complete
			"youTubeHighlight":    CompletionCriteriaTrueOnly,    // true when complete
			"youTubeComment":      CompletionCriteriaTrueOnly,    // true when complete
			"youTubeCommentReply": CompletionCriteriaTrueOnly,    // true when complete
			"gdePosted":           CompletionCriteriaTrueOnly,    // true when complete
			"codeRepository":      CompletionCriteriaFilledOnly,  // non-empty, not "-"
			"notifySponsors":      CompletionCriteriaConditional, // complete if no sponsorship OR if notified
		},
	}

	if aspectFields, exists := fieldCompletionMap[aspectKey]; exists {
		if criteria, exists := aspectFields[fieldKey]; exists {
			return criteria
		}
	}

	// Default completion criteria based on field type if not specifically mapped
	return CompletionCriteriaFilledOnly
}

// IsFieldComplete checks if a specific field is complete based on its completion criteria
// This provides a centralized way to check field completion that both API and CLI can use
func (s *CompletionService) IsFieldComplete(aspectKey, fieldKey string, fieldValue interface{}, video storage.Video) bool {
	criteria := s.GetFieldCompletionCriteria(aspectKey, fieldKey)

	switch criteria {
	case CompletionCriteriaFilledOnly:
		return s.isFilledOnly(fieldValue)
	case CompletionCriteriaEmptyOrFilled:
		return s.isEmptyOrFilled(fieldValue)
	case CompletionCriteriaFilledRequired:
		return s.isFilledRequired(fieldValue)
	case CompletionCriteriaTrueOnly:
		return s.isTrueOnly(fieldValue)
	case CompletionCriteriaFalseOnly:
		return s.isFalseOnly(fieldValue)
	case CompletionCriteriaNoFixme:
		return s.isNoFixme(fieldValue)
	case CompletionCriteriaConditional:
		return s.isConditionalComplete(aspectKey, fieldKey, fieldValue, video)
	default:
		return s.isFilledOnly(fieldValue) // Default behavior
	}
}

// Completion criteria implementation functions

func (s *CompletionService) isFilledOnly(value interface{}) bool {
	switch v := value.(type) {
	case string:
		return len(strings.TrimSpace(v)) > 0 && strings.TrimSpace(v) != "-"
	case bool:
		return v
	default:
		return false
	}
}

func (s *CompletionService) isEmptyOrFilled(value interface{}) bool {
	switch v := value.(type) {
	case string:
		return len(strings.TrimSpace(v)) == 0 // Complete when empty
	case bool:
		return !v // Complete when false
	default:
		return true
	}
}

func (s *CompletionService) isFilledRequired(value interface{}) bool {
	// Same as filled_only for now - this could be enhanced for stricter validation
	return s.isFilledOnly(value)
}

func (s *CompletionService) isTrueOnly(value interface{}) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	return false
}

func (s *CompletionService) isFalseOnly(value interface{}) bool {
	if b, ok := value.(bool); ok {
		return !b
	}
	return false
}

func (s *CompletionService) isNoFixme(value interface{}) bool {
	if str, ok := value.(string); ok {
		return len(strings.TrimSpace(str)) > 0 && !strings.Contains(str, "FIXME:")
	}
	return false
}

func (s *CompletionService) isConditionalComplete(aspectKey, fieldKey string, value interface{}, video storage.Video) bool {
	// Handle special conditional logic cases
	switch aspectKey {
	case "initial-details":
		if fieldKey == "sponsorship.emails" {
			// Complete if sponsorshipAmount is empty/N/A/- OR if emails has content
			amount := video.Sponsorship.Amount
			if len(amount) == 0 || amount == "N/A" || amount == "-" {
				return true // No sponsorship, so emails field is complete
			}
			// Has sponsorship, check if emails are filled
			return s.isFilledOnly(value)
		}
	case "post-publish":
		if fieldKey == "notifySponsors" {
			// Complete if no sponsorship OR if notified
			amount := video.Sponsorship.Amount
			if len(amount) == 0 || amount == "N/A" || amount == "-" {
				return true // No sponsorship, so notification not needed
			}
			// Has sponsorship, check if notified
			return s.isTrueOnly(value)
		}
	}

	// Default to filled_only for unknown conditional cases
	return s.isFilledOnly(value)
}
