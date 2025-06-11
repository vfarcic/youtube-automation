package aspect

import "errors"

// Completion criteria constants for field completion logic
const (
	// CompletionCriteriaFilledOnly indicates field is complete when not empty and not "-"
	CompletionCriteriaFilledOnly = "filled_only"

	// CompletionCriteriaEmptyOrFilled indicates field is complete when empty OR filled
	CompletionCriteriaEmptyOrFilled = "empty_or_filled"

	// CompletionCriteriaFilledRequired indicates field must be filled (required fields)
	CompletionCriteriaFilledRequired = "filled_required"

	// CompletionCriteriaTrueOnly indicates boolean field is complete when true
	CompletionCriteriaTrueOnly = "true_only"

	// CompletionCriteriaFalseOnly indicates boolean field is complete when false
	CompletionCriteriaFalseOnly = "false_only"

	// CompletionCriteriaConditional indicates field has special conditional logic
	CompletionCriteriaConditional = "conditional"

	// CompletionCriteriaNoFixme indicates field is complete when content doesn't contain "FIXME:"
	CompletionCriteriaNoFixme = "no_fixme"
)

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
	Name               string          `json:"name"`
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
