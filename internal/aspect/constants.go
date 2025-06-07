package aspect

// FieldType constants for supported field types
const (
	FieldTypeString  = "string"  // Short text input
	FieldTypeText    = "text"    // Multi-line text area
	FieldTypeBoolean = "boolean" // Checkbox/toggle
	FieldTypeDate    = "date"    // Date picker
	FieldTypeNumber  = "number"  // Numeric input
	FieldTypeSelect  = "select"  // Dropdown selection
)

// AspectKey constants for predefined aspects
// These match the exact phase strings used in internal/service/video_service.go
// to ensure consistency between CLI and API
const (
	AspectKeyInitialDetails = "initial-details" // Matches video service phase
	AspectKeyWorkProgress   = "work-progress"   // Matches video service phase
	AspectKeyDefinition     = "definition"      // Matches video service phase
	AspectKeyPostProduction = "post-production" // Matches video service phase
	AspectKeyPublishing     = "publishing"      // Matches video service phase
	AspectKeyPostPublish    = "post-publish"    // Matches video service phase
)
