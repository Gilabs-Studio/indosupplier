package utils

// Float64Value returns the value of the float64 pointer or 0 if nil.
func Float64Value(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

// BoolValue returns the value of the bool pointer or false if nil.
func BoolValue(v *bool) bool {
	if v == nil {
		return false
	}
	return *v
}

// StringValue returns the value of the string pointer or "" if nil.
func StringValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// Float64Ptr returns a pointer to the float64 value.
func Float64Ptr(v float64) *float64 {
	return &v
}

// BoolPtr returns a pointer to the bool value.
func BoolPtr(v bool) *bool {
	return &v
}

// StringPtr returns a pointer to the string value.
func StringPtr(v string) *string {
	return &v
}
