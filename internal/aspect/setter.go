package aspect

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// SetFieldValueByJSONPath sets a field value on a struct using a dot-separated JSON tag path.
// It is the write counterpart to GetFieldValueByJSONPath.
// data must be a pointer to a struct. jsonPath uses JSON tag names (e.g. "sponsorship.amount").
func SetFieldValueByJSONPath(data interface{}, jsonPath string, value interface{}) error {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("data must be a pointer to a struct")
	}
	v = v.Elem()

	parts := strings.Split(jsonPath, ".")

	// Traverse to the parent of the target field
	for _, part := range parts[:len(parts)-1] {
		if v.Kind() != reflect.Struct {
			return fmt.Errorf("invalid path %q: intermediate value is not a struct", jsonPath)
		}
		field := findFieldByJSONTag(v.Type(), part)
		if field == nil {
			return fmt.Errorf("invalid path %q: field %q not found", jsonPath, part)
		}
		v = v.FieldByName(field.Name)
		if !v.IsValid() {
			return fmt.Errorf("invalid path %q: field %q not valid", jsonPath, part)
		}
	}

	// Find and set the target field
	lastPart := parts[len(parts)-1]
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("invalid path %q: parent is not a struct", jsonPath)
	}
	field := findFieldByJSONTag(v.Type(), lastPart)
	if field == nil {
		return fmt.Errorf("invalid path %q: field %q not found", jsonPath, lastPart)
	}

	target := v.FieldByName(field.Name)
	if !target.IsValid() || !target.CanSet() {
		return fmt.Errorf("invalid path %q: field is not settable", jsonPath)
	}

	return setFieldValue(target, value, jsonPath)
}

// setFieldValue assigns value to target, performing type coercion as needed.
func setFieldValue(target reflect.Value, value interface{}, path string) error {
	if value == nil {
		target.Set(reflect.Zero(target.Type()))
		return nil
	}

	targetType := target.Type()
	val := reflect.ValueOf(value)

	// Direct assignment if types match
	if val.Type().AssignableTo(targetType) {
		target.Set(val)
		return nil
	}

	// Type coercion for common JSON-decoded types
	switch targetType.Kind() {
	case reflect.String:
		if val.Kind() == reflect.String {
			target.SetString(val.String())
			return nil
		}
		return fmt.Errorf("cannot assign %T to string field at %q", value, path)

	case reflect.Bool:
		if val.Kind() == reflect.Bool {
			target.SetBool(val.Bool())
			return nil
		}
		return fmt.Errorf("cannot assign %T to bool field at %q", value, path)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// JSON numbers are decoded as float64
		if val.Kind() == reflect.Float64 {
			target.SetInt(int64(val.Float()))
			return nil
		}
		if val.CanInt() {
			target.SetInt(val.Int())
			return nil
		}
		return fmt.Errorf("cannot assign %T to int field at %q", value, path)

	case reflect.Float32, reflect.Float64:
		if val.Kind() == reflect.Float64 || val.Kind() == reflect.Float32 {
			target.SetFloat(val.Float())
			return nil
		}
		if val.CanInt() {
			target.SetFloat(float64(val.Int()))
			return nil
		}
		return fmt.Errorf("cannot assign %T to float field at %q", value, path)

	case reflect.Slice:
		// For slices, use JSON round-trip to handle []interface{} → typed slice
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("cannot convert value for slice field at %q: %w", path, err)
		}
		newSlice := reflect.New(targetType)
		if err := json.Unmarshal(jsonBytes, newSlice.Interface()); err != nil {
			return fmt.Errorf("cannot unmarshal into slice field at %q: %w", path, err)
		}
		target.Set(newSlice.Elem())
		return nil

	case reflect.Map:
		// For maps, use JSON round-trip to handle map[string]interface{} → typed map
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("cannot convert value for map field at %q: %w", path, err)
		}
		newMap := reflect.New(targetType)
		if err := json.Unmarshal(jsonBytes, newMap.Interface()); err != nil {
			return fmt.Errorf("cannot unmarshal into map field at %q: %w", path, err)
		}
		target.Set(newMap.Elem())
		return nil

	case reflect.Struct:
		// For structs, use JSON round-trip to handle map[string]interface{} → typed struct
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("cannot convert value for struct field at %q: %w", path, err)
		}
		newStruct := reflect.New(targetType)
		if err := json.Unmarshal(jsonBytes, newStruct.Interface()); err != nil {
			return fmt.Errorf("cannot unmarshal into struct field at %q: %w", path, err)
		}
		target.Set(newStruct.Elem())
		return nil

	default:
		return fmt.Errorf("unsupported target type %s for field at %q", targetType.Kind(), path)
	}
}
