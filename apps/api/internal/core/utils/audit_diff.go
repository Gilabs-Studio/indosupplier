package utils

import (
	"fmt"
	"reflect"
	"strings"
)

// FieldChange records the old and new value of a single field.
type FieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
}

// DiffStructs compares two structs of the same type and returns a list of changed fields.
// Only exported fields are compared. Fields tagged with `audit:"-"` are skipped.
// Fields tagged with `audit:"<name>"` use the tag value as the field name.
func DiffStructs(old, new interface{}) []FieldChange {
	if old == nil || new == nil {
		return nil
	}

	oldVal := reflect.ValueOf(old)
	newVal := reflect.ValueOf(new)

	// Handle pointers
	if oldVal.Kind() == reflect.Ptr {
		if oldVal.IsNil() || newVal.IsNil() {
			return nil
		}
		oldVal = oldVal.Elem()
		newVal = newVal.Elem()
	}

	if oldVal.Type() != newVal.Type() {
		return nil
	}

	var changes []FieldChange
	t := oldVal.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Check audit tag
		auditTag := field.Tag.Get("audit")
		if auditTag == "-" {
			continue
		}

		// Skip complex types (slices, maps, structs) unless they implement Stringer
		if field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Map {
			continue
		}

		fieldName := auditTag
		if fieldName == "" {
			// Use JSON tag if available, otherwise snake_case the field name
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" && jsonTag != "-" {
				parts := strings.Split(jsonTag, ",")
				fieldName = parts[0]
			} else {
				fieldName = toSnakeCase(field.Name)
			}
		}

		oldFieldVal := oldVal.Field(i)
		newFieldVal := newVal.Field(i)

		// Handle pointer types
		oldInterface := resolveValue(oldFieldVal)
		newInterface := resolveValue(newFieldVal)

		if !reflect.DeepEqual(oldInterface, newInterface) {
			changes = append(changes, FieldChange{
				Field:    fieldName,
				OldValue: formatValue(oldInterface),
				NewValue: formatValue(newInterface),
			})
		}
	}

	return changes
}

func resolveValue(v reflect.Value) interface{} {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		return v.Elem().Interface()
	}
	return v.Interface()
}

func formatValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	// Convert to string representation for consistent storage
	return fmt.Sprintf("%v", v)
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
