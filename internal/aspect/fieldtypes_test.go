package aspect

import (
	"testing"
)

// TestStringFieldType tests all methods of StringFieldType
func TestStringFieldType(t *testing.T) {
	fieldType := &StringFieldType{}

	t.Run("Validate should accept valid strings", func(t *testing.T) {
		// Test valid string
		if err := fieldType.Validate("valid string"); err != nil {
			t.Errorf("Expected no error for valid string, got: %v", err)
		}

		// Test empty string (should be valid for StringFieldType)
		if err := fieldType.Validate(""); err != nil {
			t.Errorf("Expected no error for empty string, got: %v", err)
		}
	})

	t.Run("Validate should reject non-string types", func(t *testing.T) {
		// Test integer
		if err := fieldType.Validate(123); err == nil {
			t.Error("Expected error for integer value, got nil")
		}

		// Test boolean
		if err := fieldType.Validate(true); err == nil {
			t.Error("Expected error for boolean value, got nil")
		}

		// Test nil
		if err := fieldType.Validate(nil); err == nil {
			t.Error("Expected error for nil value, got nil")
		}
	})

	t.Run("GetDefaultValue should return empty string", func(t *testing.T) {
		defaultValue := fieldType.GetDefaultValue()
		if defaultValue != "" {
			t.Errorf("Expected empty string, got: %v", defaultValue)
		}
	})

	t.Run("GetTypeName should return 'string'", func(t *testing.T) {
		typeName := fieldType.GetTypeName()
		if typeName != "string" {
			t.Errorf("Expected 'string', got: %s", typeName)
		}
	})

	t.Run("GetValidationHints should return default values", func(t *testing.T) {
		hints := fieldType.GetValidationHints()

		// For a default StringFieldType, Required should be false
		if hints.Required {
			t.Error("Expected Required to be false for default StringFieldType")
		}

		// Check other default values
		if hints.MinLength != 0 {
			t.Errorf("Expected MinLength 0, got %d", hints.MinLength)
		}
		if hints.MaxLength != 0 {
			t.Errorf("Expected MaxLength 0, got %d", hints.MaxLength)
		}
	})

	t.Run("GetValidationHints should respect field configuration", func(t *testing.T) {
		requiredFieldType := &StringFieldType{Required: true, MinLength: 5, MaxLength: 100}
		hints := requiredFieldType.GetValidationHints()

		if !hints.Required {
			t.Error("Expected Required to be true when field is configured as required")
		}
		if hints.MinLength != 5 {
			t.Errorf("Expected MinLength 5, got %d", hints.MinLength)
		}
		if hints.MaxLength != 100 {
			t.Errorf("Expected MaxLength 100, got %d", hints.MaxLength)
		}
	})
}

// TestTextFieldType tests all methods of TextFieldType
func TestTextFieldType(t *testing.T) {
	fieldType := &TextFieldType{}

	t.Run("Validate should accept valid strings", func(t *testing.T) {
		if err := fieldType.Validate("multiline\ntext\nhere"); err != nil {
			t.Errorf("Expected no error for multiline text, got: %v", err)
		}

		if err := fieldType.Validate(""); err != nil {
			t.Errorf("Expected no error for empty text, got: %v", err)
		}
	})

	t.Run("Validate should reject non-string types", func(t *testing.T) {
		if err := fieldType.Validate([]string{"array"}); err == nil {
			t.Error("Expected error for array value, got nil")
		}
	})

	t.Run("GetDefaultValue should return empty string", func(t *testing.T) {
		defaultValue := fieldType.GetDefaultValue()
		if defaultValue != "" {
			t.Errorf("Expected empty string, got: %v", defaultValue)
		}
	})

	t.Run("GetTypeName should return 'text'", func(t *testing.T) {
		typeName := fieldType.GetTypeName()
		if typeName != "text" {
			t.Errorf("Expected 'text', got: %s", typeName)
		}
	})
}

// TestBooleanFieldType tests all methods of BooleanFieldType
func TestBooleanFieldType(t *testing.T) {
	fieldType := &BooleanFieldType{}

	t.Run("Validate should accept boolean values", func(t *testing.T) {
		if err := fieldType.Validate(true); err != nil {
			t.Errorf("Expected no error for true, got: %v", err)
		}

		if err := fieldType.Validate(false); err != nil {
			t.Errorf("Expected no error for false, got: %v", err)
		}
	})

	t.Run("Validate should reject non-boolean types", func(t *testing.T) {
		if err := fieldType.Validate("true"); err == nil {
			t.Error("Expected error for string 'true', got nil")
		}

		if err := fieldType.Validate(1); err == nil {
			t.Error("Expected error for integer 1, got nil")
		}
	})

	t.Run("GetDefaultValue should return false", func(t *testing.T) {
		defaultValue := fieldType.GetDefaultValue()
		if defaultValue != false {
			t.Errorf("Expected false, got: %v", defaultValue)
		}
	})

	t.Run("GetTypeName should return 'boolean'", func(t *testing.T) {
		typeName := fieldType.GetTypeName()
		if typeName != "boolean" {
			t.Errorf("Expected 'boolean', got: %s", typeName)
		}
	})
}

// TestDateFieldType tests all methods of DateFieldType
func TestDateFieldType(t *testing.T) {
	fieldType := &DateFieldType{}

	t.Run("Validate should accept valid date strings", func(t *testing.T) {
		validDates := []string{
			"2023-12-25T15:04",
			"2023-01-01T00:00",
			"2024-02-29T12:30", // leap year
		}

		for _, date := range validDates {
			if err := fieldType.Validate(date); err != nil {
				t.Errorf("Expected no error for valid date %s, got: %v", date, err)
			}
		}
	})

	t.Run("Validate should accept empty string when not required", func(t *testing.T) {
		if err := fieldType.Validate(""); err != nil {
			t.Errorf("Expected no error for empty string when not required, got: %v", err)
		}
	})

	t.Run("Validate should reject invalid date strings", func(t *testing.T) {
		invalidDates := []string{
			"2023-13-01T15:04", // invalid month
			"2023-02-30T15:04", // invalid day
			"not-a-date",
			"2023/12/25", // wrong format
			"2023-12-25", // missing time part
		}

		for _, date := range invalidDates {
			if err := fieldType.Validate(date); err == nil {
				t.Errorf("Expected error for invalid date %s, got nil", date)
			}
		}
	})

	t.Run("Validate should reject non-date types", func(t *testing.T) {
		if err := fieldType.Validate(123); err == nil {
			t.Error("Expected error for integer, got nil")
		}

		if err := fieldType.Validate(true); err == nil {
			t.Error("Expected error for boolean, got nil")
		}
	})

	t.Run("GetDefaultValue should return empty string", func(t *testing.T) {
		defaultValue := fieldType.GetDefaultValue()
		if defaultValue != "" {
			t.Errorf("Expected empty string, got: %v", defaultValue)
		}
	})

	t.Run("GetTypeName should return 'date'", func(t *testing.T) {
		typeName := fieldType.GetTypeName()
		if typeName != "date" {
			t.Errorf("Expected 'date', got: %s", typeName)
		}
	})
}

// TestSelectFieldType tests all methods of SelectFieldType
func TestSelectFieldType(t *testing.T) {
	options := []SelectOption{
		{Value: "option1", Label: "Option 1"},
		{Value: "option2", Label: "Option 2"},
		{Value: "option3", Label: "Option 3"},
	}
	fieldType := &SelectFieldType{Options: options}

	t.Run("Validate should accept valid option values", func(t *testing.T) {
		if err := fieldType.Validate("option1"); err != nil {
			t.Errorf("Expected no error for valid option, got: %v", err)
		}

		if err := fieldType.Validate("option2"); err != nil {
			t.Errorf("Expected no error for valid option, got: %v", err)
		}
	})

	t.Run("Validate should reject invalid option values", func(t *testing.T) {
		if err := fieldType.Validate("invalid-option"); err == nil {
			t.Error("Expected error for invalid option, got nil")
		}

		if err := fieldType.Validate(""); err == nil {
			t.Error("Expected error for empty string, got nil")
		}
	})

	t.Run("Validate should reject non-string types", func(t *testing.T) {
		if err := fieldType.Validate(123); err == nil {
			t.Error("Expected error for integer, got nil")
		}
	})

	t.Run("GetDefaultValue should return first option value", func(t *testing.T) {
		defaultValue := fieldType.GetDefaultValue()
		if defaultValue != "option1" {
			t.Errorf("Expected 'option1', got: %v", defaultValue)
		}
	})

	t.Run("GetDefaultValue should return nil for empty options", func(t *testing.T) {
		emptyFieldType := &SelectFieldType{Options: []SelectOption{}}
		defaultValue := emptyFieldType.GetDefaultValue()
		if defaultValue != nil {
			t.Errorf("Expected nil for empty options, got: %v", defaultValue)
		}
	})

	t.Run("GetTypeName should return 'select'", func(t *testing.T) {
		typeName := fieldType.GetTypeName()
		if typeName != "select" {
			t.Errorf("Expected 'select', got: %s", typeName)
		}
	})
}

// TestNumberFieldType tests all methods of NumberFieldType
func TestNumberFieldType(t *testing.T) {
	fieldType := &NumberFieldType{}

	t.Run("Validate should accept numeric types", func(t *testing.T) {
		validNumbers := []interface{}{
			123,   // int
			456,   // int as float64 conversion
			"123", // string number
			"456", // string number
		}

		for _, num := range validNumbers {
			if err := fieldType.Validate(num); err != nil {
				t.Errorf("Expected no error for valid number %v, got: %v", num, err)
			}
		}
	})

	t.Run("Validate should accept empty string when not required", func(t *testing.T) {
		if err := fieldType.Validate(""); err != nil {
			t.Errorf("Expected no error for empty string when not required, got: %v", err)
		}
	})

	t.Run("Validate should reject non-numeric types", func(t *testing.T) {
		invalidNumbers := []interface{}{
			"not-a-number",
			"123abc",
			"456.78", // float strings are not accepted by strconv.Atoi
			true,
			[]int{1, 2, 3},
		}

		for _, num := range invalidNumbers {
			if err := fieldType.Validate(num); err == nil {
				t.Errorf("Expected error for invalid number %v, got nil", num)
			}
		}
	})

	t.Run("GetDefaultValue should return 0", func(t *testing.T) {
		defaultValue := fieldType.GetDefaultValue()
		if defaultValue != 0 {
			t.Errorf("Expected 0, got: %v", defaultValue)
		}
	})

	t.Run("GetTypeName should return 'number'", func(t *testing.T) {
		typeName := fieldType.GetTypeName()
		if typeName != "number" {
			t.Errorf("Expected 'number', got: %s", typeName)
		}
	})
}

// TestEdgeCases tests various edge cases for all field types
func TestFieldTypeEdgeCases(t *testing.T) {
	t.Run("All field types should handle nil validation gracefully", func(t *testing.T) {
		fieldTypes := []FieldType{
			&StringFieldType{},
			&TextFieldType{},
			&BooleanFieldType{},
			&DateFieldType{},
			&SelectFieldType{Required: true, Options: []SelectOption{{Value: "test", Label: "Test"}}}, // Required SelectFieldType should reject nil
			&NumberFieldType{},
		}

		for i, ft := range fieldTypes {
			err := ft.Validate(nil)
			if err == nil {
				t.Errorf("Field type %d (%T) should return error for nil value", i, ft)
			}
		}
	})

	t.Run("All field types should return consistent type names", func(t *testing.T) {
		expectedNames := []string{"string", "text", "boolean", "date", "select", "number"}
		fieldTypes := []FieldType{
			&StringFieldType{},
			&TextFieldType{},
			&BooleanFieldType{},
			&DateFieldType{},
			&SelectFieldType{},
			&NumberFieldType{},
		}

		for i, ft := range fieldTypes {
			typeName := ft.GetTypeName()
			if typeName != expectedNames[i] {
				t.Errorf("Expected type name %s, got %s", expectedNames[i], typeName)
			}
		}
	})
}
