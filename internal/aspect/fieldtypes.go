package aspect

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FieldType defines the interface for field type validation and UI metadata
type FieldType interface {
	// Validate validates a value according to the field type rules
	Validate(value interface{}) error

	// GetUIHints returns metadata for frontend rendering
	GetUIHints() UIHints

	// GetValidationHints returns validation rules for frontend
	GetValidationHints() ValidationHints

	// GetDefaultValue returns the default value for this field type
	GetDefaultValue() interface{}

	// GetTypeName returns the string name of this field type
	GetTypeName() string
}

// UIHints provides rendering guidance for frontends
type UIHints struct {
	InputType   string                 `json:"inputType"`            // "text", "textarea", "checkbox", "select", "date"
	Placeholder string                 `json:"placeholder"`          // Placeholder text
	HelpText    string                 `json:"helpText"`             // Help or description text
	Rows        int                    `json:"rows,omitempty"`       // For textarea
	CharLimit   int                    `json:"charLimit,omitempty"`  // Character limit
	Multiline   bool                   `json:"multiline"`            // Whether field supports multiple lines
	Options     []SelectOption         `json:"options,omitempty"`    // For select fields
	Attributes  map[string]interface{} `json:"attributes,omitempty"` // Additional HTML attributes
}

// ValidationHints provides validation rules for frontends
type ValidationHints struct {
	Required    bool   `json:"required"`
	MinLength   int    `json:"minLength,omitempty"`
	MaxLength   int    `json:"maxLength,omitempty"`
	Pattern     string `json:"pattern,omitempty"`     // Regex pattern
	PatternDesc string `json:"patternDesc,omitempty"` // Human description of pattern
	Min         *int   `json:"min,omitempty"`         // For numeric fields
	Max         *int   `json:"max,omitempty"`         // For numeric fields
}

// SelectOption represents an option for select fields
type SelectOption struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

// StringFieldType handles string input fields
type StringFieldType struct {
	MinLength   int
	MaxLength   int
	Pattern     *regexp.Regexp
	PatternDesc string
	Placeholder string
	HelpText    string
	Required    bool
}

func (s StringFieldType) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return errors.New("value must be a string")
	}

	if s.Required && strings.TrimSpace(str) == "" {
		return errors.New("field is required")
	}

	if s.MinLength > 0 && len(str) < s.MinLength {
		return errors.New("value is too short")
	}

	if s.MaxLength > 0 && len(str) > s.MaxLength {
		return errors.New("value is too long")
	}

	if s.Pattern != nil && !s.Pattern.MatchString(str) {
		return errors.New("value does not match required pattern")
	}

	return nil
}

func (s StringFieldType) GetUIHints() UIHints {
	return UIHints{
		InputType:   "text",
		Placeholder: s.Placeholder,
		HelpText:    s.HelpText,
		CharLimit:   s.MaxLength,
		Multiline:   false,
	}
}

func (s StringFieldType) GetValidationHints() ValidationHints {
	hints := ValidationHints{
		Required:  s.Required,
		MinLength: s.MinLength,
		MaxLength: s.MaxLength,
	}

	if s.Pattern != nil {
		hints.Pattern = s.Pattern.String()
		hints.PatternDesc = s.PatternDesc
	}

	return hints
}

func (s StringFieldType) GetDefaultValue() interface{} {
	return ""
}

func (s StringFieldType) GetTypeName() string {
	return "string"
}

// TextFieldType handles multi-line text fields
type TextFieldType struct {
	MinLength   int
	MaxLength   int
	Rows        int
	Placeholder string
	HelpText    string
	Required    bool
}

func (t TextFieldType) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return errors.New("value must be a string")
	}

	if t.Required && strings.TrimSpace(str) == "" {
		return errors.New("field is required")
	}

	if t.MinLength > 0 && len(str) < t.MinLength {
		return errors.New("value is too short")
	}

	if t.MaxLength > 0 && len(str) > t.MaxLength {
		return errors.New("value is too long")
	}

	return nil
}

func (t TextFieldType) GetUIHints() UIHints {
	rows := t.Rows
	if rows == 0 {
		rows = 3 // Default rows
	}

	return UIHints{
		InputType:   "textarea",
		Placeholder: t.Placeholder,
		HelpText:    t.HelpText,
		Rows:        rows,
		CharLimit:   t.MaxLength,
		Multiline:   true,
	}
}

func (t TextFieldType) GetValidationHints() ValidationHints {
	return ValidationHints{
		Required:  t.Required,
		MinLength: t.MinLength,
		MaxLength: t.MaxLength,
	}
}

func (t TextFieldType) GetDefaultValue() interface{} {
	return ""
}

func (t TextFieldType) GetTypeName() string {
	return "text"
}

// BooleanFieldType handles checkbox/confirm fields
type BooleanFieldType struct {
	HelpText string
	Required bool
}

func (b BooleanFieldType) Validate(value interface{}) error {
	_, ok := value.(bool)
	if !ok {
		return errors.New("value must be a boolean")
	}

	// For boolean fields, "required" typically means it must be true
	if b.Required {
		if boolVal, ok := value.(bool); ok && !boolVal {
			return errors.New("field must be confirmed")
		}
	}

	return nil
}

func (b BooleanFieldType) GetUIHints() UIHints {
	return UIHints{
		InputType: "checkbox",
		HelpText:  b.HelpText,
		Multiline: false,
	}
}

func (b BooleanFieldType) GetValidationHints() ValidationHints {
	return ValidationHints{
		Required: b.Required,
	}
}

func (b BooleanFieldType) GetDefaultValue() interface{} {
	return false
}

func (b BooleanFieldType) GetTypeName() string {
	return "boolean"
}

// DateFieldType handles date input fields
type DateFieldType struct {
	Format      string // Expected format, e.g., "2006-01-02T15:04"
	HelpText    string
	Required    bool
	Placeholder string
}

func (d DateFieldType) Validate(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return errors.New("value must be a string")
	}

	if d.Required && strings.TrimSpace(str) == "" {
		return errors.New("field is required")
	}

	if str != "" {
		format := d.Format
		if format == "" {
			format = "2006-01-02T15:04" // Default format for UTC datetime
		}

		_, err := time.Parse(format, str)
		if err != nil {
			return errors.New("invalid date format")
		}
	}

	return nil
}

func (d DateFieldType) GetUIHints() UIHints {
	placeholder := d.Placeholder

	if placeholder == "" {
		format := d.Format
		if format == "" {
			format = "2006-01-02T15:04" // Default format for UTC datetime
		}
		placeholder = "YYYY-MM-DDTHH:MM"
	}

	return UIHints{
		InputType:   "datetime",
		Placeholder: placeholder,
		HelpText:    d.HelpText,
		Multiline:   false,
	}
}

func (d DateFieldType) GetValidationHints() ValidationHints {
	return ValidationHints{
		Required: d.Required,
	}
}

func (d DateFieldType) GetDefaultValue() interface{} {
	return ""
}

func (d DateFieldType) GetTypeName() string {
	return "date"
}

// SelectFieldType handles dropdown selection fields
type SelectFieldType struct {
	Options     []SelectOption
	HelpText    string
	Required    bool
	Placeholder string
}

func (s SelectFieldType) Validate(value interface{}) error {
	if s.Required && value == nil {
		return errors.New("field is required")
	}

	if value == nil {
		return nil
	}

	// Check if value is one of the valid options
	for _, option := range s.Options {
		if option.Value == value {
			return nil
		}
	}

	return errors.New("invalid option selected")
}

func (s SelectFieldType) GetUIHints() UIHints {
	return UIHints{
		InputType:   "select",
		Placeholder: s.Placeholder,
		HelpText:    s.HelpText,
		Options:     s.Options,
		Multiline:   false,
	}
}

func (s SelectFieldType) GetValidationHints() ValidationHints {
	return ValidationHints{
		Required: s.Required,
	}
}

func (s SelectFieldType) GetDefaultValue() interface{} {
	if len(s.Options) > 0 {
		return s.Options[0].Value
	}
	return nil
}

func (s SelectFieldType) GetTypeName() string {
	return "select"
}

// NumberFieldType handles numeric input fields
type NumberFieldType struct {
	Min         *int
	Max         *int
	HelpText    string
	Required    bool
	Placeholder string
}

func (n NumberFieldType) Validate(value interface{}) error {
	// Handle string input (from forms)
	if str, ok := value.(string); ok {
		if n.Required && strings.TrimSpace(str) == "" {
			return errors.New("field is required")
		}

		if str == "" {
			return nil // Optional field
		}

		num, err := strconv.Atoi(str)
		if err != nil {
			return errors.New("value must be a number")
		}

		value = num
	}

	// Handle numeric input
	var num int
	switch v := value.(type) {
	case int:
		num = v
	case float64:
		num = int(v)
	default:
		return errors.New("value must be a number")
	}

	if n.Min != nil && num < *n.Min {
		return errors.New("value is too small")
	}

	if n.Max != nil && num > *n.Max {
		return errors.New("value is too large")
	}

	return nil
}

func (n NumberFieldType) GetUIHints() UIHints {
	return UIHints{
		InputType:   "number",
		Placeholder: n.Placeholder,
		HelpText:    n.HelpText,
		Multiline:   false,
	}
}

func (n NumberFieldType) GetValidationHints() ValidationHints {
	return ValidationHints{
		Required: n.Required,
		Min:      n.Min,
		Max:      n.Max,
	}
}

func (n NumberFieldType) GetDefaultValue() interface{} {
	if n.Min != nil {
		return *n.Min
	}
	return 0
}

func (n NumberFieldType) GetTypeName() string {
	return "number"
}
