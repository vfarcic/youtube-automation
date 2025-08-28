package aspect

import (
	"devopstoolkit/youtube-automation/internal/storage"
	"reflect"
	"regexp"
	"strings"
)

// FieldMapping defines metadata for a form field, generated from struct reflection
type FieldMapping struct {
	Name            string          `json:"name"`              // Generated from struct field name
	FieldName       string          `json:"fieldName"`         // JSON field name from struct tags
	FieldType       string          `json:"fieldType"`         // Determined from Go type
	Title           string          `json:"title"`             // Same as Name for now
	Required        bool            `json:"required"`          // Could be determined from validation tags
	Order           int             `json:"order"`             // Based on struct field order
	Options         []string        `json:"options,omitempty"` // For select fields (if needed)
	UIHints         UIHints         `json:"uiHints"`           // Generated from field type
	ValidationHints ValidationHints `json:"validationHints"`   // Generated from field type
	DefaultValue    interface{}     `json:"defaultValue"`      // Based on Go zero values
}

// AspectMapping defines the complete mapping configuration for an editing aspect
type AspectMapping struct {
	AspectKey   string         `json:"aspectKey"`   // The aspect identifier
	Title       string         `json:"title"`       // Display title for the aspect
	Description string         `json:"description"` // Description of the aspect
	Fields      []FieldMapping `json:"fields"`      // Generated field mappings
	Order       int            `json:"order"`       // Workflow order (1-6)
}

// GetVideoAspectMappings generates aspect mappings directly from the storage.Video struct
func GetVideoAspectMappings() []AspectMapping {
	videoType := reflect.TypeOf(storage.Video{})

	// Define which fields belong to which aspects (this is the only configuration needed)
	aspectFieldGroups := map[string][]string{
		AspectKeyInitialDetails: {
			"ProjectName", "ProjectURL", "Sponsorship.Amount", "Sponsorship.Emails", "Sponsorship.Blocked", "Date", "Delayed", "Gist",
		},
		AspectKeyWorkProgress: {
			"Code", "Head", "Screen", "RelatedVideos", "Thumbnails", "Diagrams",
			"Screenshots", "Location", "Tagline", "TaglineIdeas", "OtherLogos",
		},
		AspectKeyDefinition: {
			"Title", "Description", "Tags", "DescriptionTags", "Tweet", "Animations", "RequestThumbnail",
		},
		AspectKeyPostProduction: {
			"Thumbnail", "Members", "RequestEdit", "Timecodes", "Movie", "Slides",
		},
		AspectKeyPublishing: {
			"UploadVideo", "VideoId", "HugoPath",
		},
		AspectKeyPostPublish: {
			"DOTPosted", "BlueSkyPosted", "LinkedInPosted", "SlackPosted",
			"YouTubeHighlight", "YouTubeComment", "YouTubeCommentReply",
			"GDE", "Repo", "NotifiedSponsors",
		},
	}

	aspectTitles := map[string]string{
		AspectKeyInitialDetails: "Initial Details",
		AspectKeyWorkProgress:   "Work Progress",
		AspectKeyDefinition:     "Definition",
		AspectKeyPostProduction: "Post Production",
		AspectKeyPublishing:     "Publishing",
		AspectKeyPostPublish:    "Post Publish",
	}

	aspectDescriptions := map[string]string{
		AspectKeyInitialDetails: "Initial video details and project information",
		AspectKeyWorkProgress:   "Work progress and content creation status",
		AspectKeyDefinition:     "Video content definition and metadata",
		AspectKeyPostProduction: "Post-production editing and review tasks",
		AspectKeyPublishing:     "Publishing settings and video upload",
		AspectKeyPostPublish:    "Post-publication tasks and social media",
	}

	// Define the correct workflow order for aspects (used by both CLI and API)
	aspectOrder := map[string]int{
		AspectKeyInitialDetails: 1,
		AspectKeyWorkProgress:   2,
		AspectKeyDefinition:     3,
		AspectKeyPostProduction: 4,
		AspectKeyPublishing:     5,
		AspectKeyPostPublish:    6,
	}

	var aspects []AspectMapping

	for aspectKey, fieldNames := range aspectFieldGroups {
		var fields []FieldMapping

		for order, fieldName := range fieldNames {
			field := generateFieldMapping(videoType, fieldName, order+1)
			if field != nil {
				fields = append(fields, *field)
			}
		}

		aspects = append(aspects, AspectMapping{
			AspectKey:   aspectKey,
			Title:       aspectTitles[aspectKey],
			Description: aspectDescriptions[aspectKey],
			Fields:      fields,
			Order:       aspectOrder[aspectKey], // Set the correct workflow order
		})
	}

	return aspects
}

// generateFieldMapping creates a FieldMapping from struct field reflection
func generateFieldMapping(structType reflect.Type, fieldPath string, order int) *FieldMapping {
	// Handle nested field paths like "Sponsorship.Amount"
	parts := strings.Split(fieldPath, ".")

	if len(parts) == 2 {
		// Nested field like "Sponsorship.Amount"
		parentFieldName := parts[0]
		childFieldName := parts[1]

		parentField, found := structType.FieldByName(parentFieldName)
		if !found {
			return nil
		}

		childField, found := parentField.Type.FieldByName(childFieldName)
		if !found {
			return nil
		}

		// Get JSON field names
		parentJsonTag := strings.Split(parentField.Tag.Get("json"), ",")[0]
		childJsonTag := strings.Split(childField.Tag.Get("json"), ",")[0]
		jsonFieldName := parentJsonTag + "." + childJsonTag

		// Generate display name
		displayName := generateDisplayName(parentFieldName) + " " + generateDisplayName(childFieldName)

		// Determine field type from Go type and field name
		fieldType := determineFieldType(childField.Type, childFieldName)

		// Create field type instance for UI hints
		fieldTypeInstance := createFieldTypeInstance(fieldType)

		return &FieldMapping{
			Name:            displayName,
			FieldName:       jsonFieldName,
			FieldType:       fieldType,
			Title:           displayName,
			Required:        false,
			Order:           order,
			Options:         nil,
			UIHints:         fieldTypeInstance.GetUIHints(),
			ValidationHints: fieldTypeInstance.GetValidationHints(),
			DefaultValue:    getDefaultValueForType(childField.Type),
		}
	}

	// Regular field
	field, found := structType.FieldByName(fieldPath)
	if !found {
		return nil
	}

	// Get JSON field name from tag
	jsonTag := field.Tag.Get("json")
	jsonFieldName := strings.Split(jsonTag, ",")[0]
	if jsonFieldName == "" {
		jsonFieldName = strings.ToLower(fieldPath)
	}

	// Generate display name from struct field name
	displayName := generateDisplayName(fieldPath)

	// Determine field type from Go type and field name
	fieldType := determineFieldType(field.Type, fieldPath)

	// Create field type instance for UI hints
	fieldTypeInstance := createFieldTypeInstance(fieldType)

	return &FieldMapping{
		Name:            displayName,
		FieldName:       jsonFieldName,
		FieldType:       fieldType,
		Title:           displayName,
		Required:        false, // Could be enhanced with validation tags
		Order:           order,
		Options:         nil,
		UIHints:         fieldTypeInstance.GetUIHints(),
		ValidationHints: fieldTypeInstance.GetValidationHints(),
		DefaultValue:    getDefaultValueForType(field.Type),
	}
}

// generateDisplayName converts struct field names to display names
func generateDisplayName(fieldName string) string {
	// Handle special cases first (before splitting)
	specialCases := map[string]string{
		"VideoId":          "Video ID",
		"DOTPosted":        "DOT Posted",
		"HNPosted":         "HN Posted",
		"ProjectURL":       "Project URL",
		"YouTubeHighlight": "YouTube Highlight",
		"GDE":              "GDE",
	}

	if special, exists := specialCases[fieldName]; exists {
		return special
	}

	// Convert camelCase/PascalCase to "Title Case"
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	result := re.ReplaceAllString(fieldName, `$1 $2`)

	// Handle common acronyms that might have been split
	result = strings.ReplaceAll(result, " U R L", " URL")
	result = strings.ReplaceAll(result, " I D", " ID")

	return result
}

// determineFieldType maps Go types to field types, considering semantic meaning
func determineFieldType(goType reflect.Type, fieldName string) string {
	// First check for semantic field types based on field names
	switch {
	case fieldName == "Date" || fieldName == "date":
		return FieldTypeDate
	case isMultiLineTextField(fieldName):
		return FieldTypeText
	}

	// Then check Go types
	switch goType.Kind() {
	case reflect.Bool:
		return FieldTypeBoolean
	case reflect.String:
		return FieldTypeString
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return FieldTypeNumber
	default:
		return FieldTypeString
	}
}

// isMultiLineTextField determines if a field should be treated as multi-line text
func isMultiLineTextField(fieldName string) bool {
	multiLineFields := map[string]bool{
		"Description":     true,
		"Tags":            true,
		"DescriptionTags": true,
		"Timecodes":       true,
		"RelatedVideos":   true,
		"Tweet":           true,
		"TaglineIdeas":    true,
		"Members":         true,
		"Animations":      true,
	}
	return multiLineFields[fieldName]
}

// createFieldTypeInstance creates appropriate field type instances
func createFieldTypeInstance(fieldType string) FieldType {
	switch fieldType {
	case FieldTypeString:
		return &StringFieldType{}
	case FieldTypeText:
		return &TextFieldType{}
	case FieldTypeBoolean:
		return &BooleanFieldType{}
	case FieldTypeDate:
		return &DateFieldType{}
	case FieldTypeNumber:
		return &NumberFieldType{}
	default:
		return &StringFieldType{}
	}
}

// getDefaultValueForType returns appropriate default values based on Go type
func getDefaultValueForType(goType reflect.Type) interface{} {
	switch goType.Kind() {
	case reflect.Bool:
		return false
	case reflect.String:
		return ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return 0
	default:
		return nil
	}
}

// GetFieldValueByJSONPath extracts a field value from any struct using JSON field path
func GetFieldValueByJSONPath(data interface{}, jsonPath string) interface{} {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Handle nested paths like "sponsorship.amount"
	parts := strings.Split(jsonPath, ".")

	for _, part := range parts {
		if v.Kind() != reflect.Struct {
			return nil
		}

		// Find field by JSON tag
		field := findFieldByJSONTag(v.Type(), part)
		if field == nil {
			return nil
		}

		v = v.FieldByName(field.Name)
		if !v.IsValid() {
			return nil
		}
	}

	return v.Interface()
}

// findFieldByJSONTag finds a struct field by its JSON tag
func findFieldByJSONTag(t reflect.Type, jsonTag string) *reflect.StructField {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")

		// Handle tags like "amount,omitempty"
		if tagName := strings.Split(tag, ",")[0]; tagName == jsonTag {
			return &field
		}
	}
	return nil
}
