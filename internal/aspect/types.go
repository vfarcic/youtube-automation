package aspect

import "errors"

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
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
	Icon        string `json:"icon"`
	Order       int    `json:"order"`
	FieldCount  int    `json:"fieldCount"`
}

// Field represents a single editable field within an aspect
type Field struct {
	Name            string          `json:"name"`
	Type            string          `json:"type"`
	Required        bool            `json:"required"`
	Order           int             `json:"order"`
	Description     string          `json:"description"`
	Options         FieldOptions    `json:"options,omitempty"`
	UIHints         UIHints         `json:"uiHints,omitempty"`
	ValidationHints ValidationHints `json:"validationHints,omitempty"`
	DefaultValue    interface{}     `json:"defaultValue,omitempty"`
}

// FieldOptions provides additional configuration for select-type fields
type FieldOptions struct {
	Values []string `json:"values,omitempty"`
}

// Errors
var (
	ErrAspectNotFound = errors.New("aspect not found")
)
