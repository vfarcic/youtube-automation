package aspect

import "errors"

// Note: Completion criteria are now defined directly in struct tags in storage/yaml.go
// This eliminates the need for constants and makes the system more maintainable

// AspectMetadata represents the complete metadata structure for all aspects
type AspectMetadata struct {
	Aspects []Aspect `json:"aspects"`
}

// AspectOverview represents a lightweight overview of all aspects without fields
type AspectOverview struct {
	Aspects []AspectSummary `json:"aspects"`
}

// AspectFields represents detailed field information for a specific aspect
type AspectFields struct {
	AspectKey   string  `json:"aspectKey"`
	AspectTitle string  `json:"aspectTitle"`
	Fields      []Field `json:"fields"`
}

// Aspect represents a single editing aspect with complete metadata
type Aspect struct {
	Key         string  `json:"key"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Endpoint    string  `json:"endpoint"`
	Icon        string  `json:"icon"`
	Order       int     `json:"order"`
	Fields      []Field `json:"fields"`
}

// AspectSummary represents a lightweight aspect overview without fields
type AspectSummary struct {
	Key                 string `json:"key"`
	Title               string `json:"title"`
	Description         string `json:"description"`
	Endpoint            string `json:"endpoint"`
	Icon                string `json:"icon"`
	Order               int    `json:"order"`
	FieldCount          int    `json:"fieldCount"`
	CompletedFieldCount int    `json:"completedFieldCount"`
}

// Field represents a single editable field within an aspect
type Field struct {
	Name               string          `json:"name"`      // Display name for UI
	FieldName          string          `json:"fieldName"` // Actual camelCase property name in video data API
	Type               string          `json:"type"`
	Required           bool            `json:"required"`
	Order              int             `json:"order"`
	Description        string          `json:"description"`
	Options            FieldOptions    `json:"options,omitempty"`
	UIHints            UIHints         `json:"uiHints,omitempty"`
	ValidationHints    ValidationHints `json:"validationHints,omitempty"`
	DefaultValue       interface{}     `json:"defaultValue,omitempty"`
	CompletionCriteria string          `json:"completionCriteria"`
}

// FieldOptions provides additional configuration for select-type fields
type FieldOptions struct {
	Values []string `json:"values,omitempty"`
}

// Errors
var (
	ErrAspectNotFound = errors.New("aspect not found")
)
