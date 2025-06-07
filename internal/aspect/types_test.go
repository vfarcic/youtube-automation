package aspect

import (
	"encoding/json"
	"testing"
)

func TestAspectMetadataJSONSerialization(t *testing.T) {
	metadata := AspectMetadata{
		Aspects: []Aspect{
			{
				Key:         "test-aspect",
				Title:       "Test Aspect",
				Description: "Test description",
				Endpoint:    "/api/test",
				Icon:        "test-icon",
				Order:       1,
				Fields: []Field{
					{
						Name:        "TestField",
						Type:        FieldTypeString,
						Required:    true,
						Order:       1,
						Description: "Test field description",
						Options:     FieldOptions{},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal AspectMetadata: %v", err)
	}

	var deserializedMetadata AspectMetadata
	err = json.Unmarshal(jsonData, &deserializedMetadata)
	if err != nil {
		t.Fatalf("Failed to unmarshal AspectMetadata: %v", err)
	}

	if len(deserializedMetadata.Aspects) != 1 {
		t.Errorf("Expected 1 aspect, got %d", len(deserializedMetadata.Aspects))
	}

	aspect := deserializedMetadata.Aspects[0]
	if aspect.Key != "test-aspect" {
		t.Errorf("Expected key 'test-aspect', got '%s'", aspect.Key)
	}

	if len(aspect.Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(aspect.Fields))
	}
}

func TestAspectOverviewJSONSerialization(t *testing.T) {
	overview := AspectOverview{
		Aspects: []AspectSummary{
			{
				Key:         "test-aspect",
				Title:       "Test Aspect",
				Description: "Test description",
				Endpoint:    "/api/test",
				Icon:        "test-icon",
				Order:       1,
				FieldCount:  5,
			},
		},
	}

	jsonData, err := json.Marshal(overview)
	if err != nil {
		t.Fatalf("Failed to marshal AspectOverview: %v", err)
	}

	var deserializedOverview AspectOverview
	err = json.Unmarshal(jsonData, &deserializedOverview)
	if err != nil {
		t.Fatalf("Failed to unmarshal AspectOverview: %v", err)
	}

	if len(deserializedOverview.Aspects) != 1 {
		t.Errorf("Expected 1 aspect, got %d", len(deserializedOverview.Aspects))
	}

	aspect := deserializedOverview.Aspects[0]
	if aspect.FieldCount != 5 {
		t.Errorf("Expected field count 5, got %d", aspect.FieldCount)
	}
}

func TestAspectFieldsJSONSerialization(t *testing.T) {
	aspectFields := AspectFields{
		AspectKey:   "test-aspect",
		AspectTitle: "Test Aspect",
		Fields: []Field{
			{
				Name:        "TestField",
				Type:        FieldTypeString,
				Required:    true,
				Order:       1,
				Description: "Test field description",
				Options:     FieldOptions{Values: []string{"option1", "option2"}},
			},
		},
	}

	jsonData, err := json.Marshal(aspectFields)
	if err != nil {
		t.Fatalf("Failed to marshal AspectFields: %v", err)
	}

	var deserializedFields AspectFields
	err = json.Unmarshal(jsonData, &deserializedFields)
	if err != nil {
		t.Fatalf("Failed to unmarshal AspectFields: %v", err)
	}

	if deserializedFields.AspectKey != "test-aspect" {
		t.Errorf("Expected aspect key 'test-aspect', got '%s'", deserializedFields.AspectKey)
	}

	if len(deserializedFields.Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(deserializedFields.Fields))
	}

	field := deserializedFields.Fields[0]
	if len(field.Options.Values) != 2 {
		t.Errorf("Expected 2 option values, got %d", len(field.Options.Values))
	}
}

func TestFieldJSONSerialization(t *testing.T) {
	field := Field{
		Name:        "TestField",
		Type:        FieldTypeBoolean,
		Required:    false,
		Order:       2,
		Description: "Boolean test field",
		Options:     FieldOptions{},
	}

	jsonData, err := json.Marshal(field)
	if err != nil {
		t.Fatalf("Failed to marshal Field: %v", err)
	}

	var deserializedField Field
	err = json.Unmarshal(jsonData, &deserializedField)
	if err != nil {
		t.Fatalf("Failed to unmarshal Field: %v", err)
	}

	if deserializedField.Type != FieldTypeBoolean {
		t.Errorf("Expected type '%s', got '%s'", FieldTypeBoolean, deserializedField.Type)
	}

	if len(deserializedField.Options.Values) != 0 {
		t.Errorf("Expected empty options, got %v", deserializedField.Options.Values)
	}
}

func TestFieldTypeValidation(t *testing.T) {
	validTypes := []string{
		FieldTypeString,
		FieldTypeText,
		FieldTypeBoolean,
		FieldTypeDate,
		FieldTypeNumber,
		FieldTypeSelect,
	}

	for _, fieldType := range validTypes {
		field := Field{
			Name:        "TestField",
			Type:        fieldType,
			Required:    true,
			Order:       1,
			Description: "Test field",
			Options:     FieldOptions{},
		}

		jsonData, err := json.Marshal(field)
		if err != nil {
			t.Errorf("Failed to marshal field with type '%s': %v", fieldType, err)
		}

		var deserializedField Field
		err = json.Unmarshal(jsonData, &deserializedField)
		if err != nil {
			t.Errorf("Failed to unmarshal field with type '%s': %v", fieldType, err)
		}

		if deserializedField.Type != fieldType {
			t.Errorf("Expected type '%s', got '%s'", fieldType, deserializedField.Type)
		}
	}
}

func TestFieldOptionsWithSelectType(t *testing.T) {
	field := Field{
		Name:        "SelectField",
		Type:        FieldTypeSelect,
		Required:    false,
		Order:       1,
		Description: "Select field with options",
		Options:     FieldOptions{Values: []string{"option1", "option2", "option3"}},
	}

	jsonData, err := json.Marshal(field)
	if err != nil {
		t.Fatalf("Failed to marshal field with select options: %v", err)
	}

	var deserializedField Field
	err = json.Unmarshal(jsonData, &deserializedField)
	if err != nil {
		t.Fatalf("Failed to unmarshal field with select options: %v", err)
	}

	if len(deserializedField.Options.Values) != 3 {
		t.Errorf("Expected 3 option values, got %d", len(deserializedField.Options.Values))
	}

	expectedValues := []string{"option1", "option2", "option3"}
	for i, expectedValue := range expectedValues {
		if i >= len(deserializedField.Options.Values) || deserializedField.Options.Values[i] != expectedValue {
			t.Errorf("Expected option value '%s' at index %d, got '%s'", expectedValue, i, deserializedField.Options.Values[i])
		}
	}
}

func TestFieldOptionsWithEmptyValues(t *testing.T) {
	field := Field{
		Name:        "StringField",
		Type:        FieldTypeString,
		Required:    true,
		Order:       1,
		Description: "String field without options",
		Options:     FieldOptions{},
	}

	jsonData, err := json.Marshal(field)
	if err != nil {
		t.Fatalf("Failed to marshal field with empty options: %v", err)
	}

	var deserializedField Field
	err = json.Unmarshal(jsonData, &deserializedField)
	if err != nil {
		t.Fatalf("Failed to unmarshal field with empty options: %v", err)
	}

	if len(deserializedField.Options.Values) != 0 {
		t.Errorf("Expected empty options values, got %v", deserializedField.Options.Values)
	}
}

func TestConstantValues(t *testing.T) {
	expectedFieldTypes := map[string]string{
		"FieldTypeString":  "string",
		"FieldTypeText":    "text",
		"FieldTypeBoolean": "boolean",
		"FieldTypeDate":    "date",
		"FieldTypeNumber":  "number",
		"FieldTypeSelect":  "select",
	}

	actualFieldTypes := map[string]string{
		"FieldTypeString":  FieldTypeString,
		"FieldTypeText":    FieldTypeText,
		"FieldTypeBoolean": FieldTypeBoolean,
		"FieldTypeDate":    FieldTypeDate,
		"FieldTypeNumber":  FieldTypeNumber,
		"FieldTypeSelect":  FieldTypeSelect,
	}

	for constantName, expectedValue := range expectedFieldTypes {
		if actualValue, exists := actualFieldTypes[constantName]; !exists {
			t.Errorf("Constant %s is not defined", constantName)
		} else if actualValue != expectedValue {
			t.Errorf("Expected constant %s to be '%s', got '%s'", constantName, expectedValue, actualValue)
		}
	}
}
