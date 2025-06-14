package aspect

import (
	"reflect"
	"strings"

	"devopstoolkit/youtube-automation/internal/storage"
)

// CompletionService handles field completion logic using struct tag reflection
type CompletionService struct {
	fieldCompletionCache map[string]string // Cache for field completion criteria
}

// NewCompletionService creates a new completion service
func NewCompletionService() *CompletionService {
	service := &CompletionService{
		fieldCompletionCache: make(map[string]string),
	}
	service.initializeCompletionCache()
	return service
}

// initializeCompletionCache uses reflection to build a cache of field completion criteria from struct tags
func (s *CompletionService) initializeCompletionCache() {
	// Get completion criteria from Video struct
	videoType := reflect.TypeOf(storage.Video{})
	s.cacheStructCompletionCriteria(videoType, "")

	// Get completion criteria from nested Sponsorship struct
	sponsorshipType := reflect.TypeOf(storage.Sponsorship{})
	s.cacheStructCompletionCriteria(sponsorshipType, "sponsorship")
}

// cacheStructCompletionCriteria extracts completion criteria from struct tags and caches them
func (s *CompletionService) cacheStructCompletionCriteria(structType reflect.Type, prefix string) {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Get JSON field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse JSON tag to get field name (remove omitempty, etc.)
		jsonFieldName := strings.Split(jsonTag, ",")[0]
		if jsonFieldName == "" {
			continue
		}

		// Build full field name with prefix for nested structs
		var fullFieldName string
		if prefix != "" {
			fullFieldName = prefix + jsonFieldName // e.g., "sponsorshipamount"
		} else {
			fullFieldName = jsonFieldName
		}

		// Get completion criteria from struct tag
		completionTag := field.Tag.Get("completion")
		if completionTag != "" {
			s.fieldCompletionCache[fullFieldName] = completionTag
		} else {
			// Default to filled_only if no completion tag specified
			s.fieldCompletionCache[fullFieldName] = "filled_only"
		}
	}
}

// GetFieldCompletionCriteria returns the completion criteria for a specific field
// This now uses reflection to read criteria from struct tags instead of hard-coded mappings
func (s *CompletionService) GetFieldCompletionCriteria(aspectKey, fieldKey string) string {
	// Handle special field name mappings for nested fields
	mappedFieldKey := s.mapFieldKeyForCompletion(fieldKey)

	// Look up completion criteria from cache
	if criteria, exists := s.fieldCompletionCache[mappedFieldKey]; exists {
		return criteria
	}

	// Default fallback
	return "filled_only"
}

// mapFieldKeyForCompletion handles special field name mappings for nested and special fields
func (s *CompletionService) mapFieldKeyForCompletion(fieldKey string) string {
	// Map special field names to their struct tag equivalents
	mappings := map[string]string{
		"sponsorshipAmount":        "sponsorshipamount",
		"sponsorshipEmails":        "sponsorshipemails",
		"sponsorshipBlockedReason": "sponsorshipblocked",
		"notifySponsors":           "notifiedSponsors", // Handle legacy field name
		"notifiedSponsors":         "notifiedSponsors", // Direct mapping
	}

	if mapped, exists := mappings[fieldKey]; exists {
		return mapped
	}

	return fieldKey
}

// IsFieldComplete checks if a specific field is complete based on its completion criteria
// This provides a centralized way to check field completion that both API and CLI can use
func (s *CompletionService) IsFieldComplete(aspectKey, fieldKey string, fieldValue interface{}, video storage.Video) bool {
	criteria := s.GetFieldCompletionCriteria(aspectKey, fieldKey)

	switch criteria {
	case "filled_only":
		return s.isFilledOnly(fieldValue)
	case "empty_or_filled":
		return s.isEmptyOrFilled(fieldValue)
	case "filled_required":
		return s.isFilledRequired(fieldValue)
	case "true_only":
		return s.isTrueOnly(fieldValue)
	case "false_only":
		return s.isFalseOnly(fieldValue)
	case "no_fixme":
		return s.isNoFixme(fieldValue)
	case "conditional_sponsorship":
		return s.isConditionalSponsorshipComplete(fieldKey, fieldValue, video)
	case "conditional_sponsors":
		return s.isConditionalSponsorsComplete(fieldKey, fieldValue, video)
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

func (s *CompletionService) isConditionalSponsorshipComplete(fieldKey string, value interface{}, video storage.Video) bool {
	// Handle sponsorship emails field - complete if sponsorshipAmount is empty/N/A/- OR if emails has content
	amount := video.Sponsorship.Amount
	if len(amount) == 0 || amount == "N/A" || amount == "-" {
		return true // No sponsorship, so emails field is complete
	}
	// Has sponsorship, check if emails are filled
	return s.isFilledOnly(value)
}

func (s *CompletionService) isConditionalSponsorsComplete(fieldKey string, value interface{}, video storage.Video) bool {
	// Handle notifySponsors field - complete if no sponsorship OR if notified
	amount := video.Sponsorship.Amount
	if len(amount) == 0 || amount == "N/A" || amount == "-" {
		return true // No sponsorship, so notification not needed
	}
	// Has sponsorship, check if notified
	return s.isTrueOnly(value)
}
