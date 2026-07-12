package coinglass

// IntPtr returns a pointer to the provided int value.
// It is useful for populating optional pointer fields in parameter structs.
func IntPtr(i int) *int {
	return &i
}

// StringPtr returns a pointer to the provided string value.
func StringPtr(s string) *string {
	return &s
}

// BoolPtr returns a pointer to the provided bool value.
func BoolPtr(b bool) *bool {
	return &b
}

// Int64Ptr returns a pointer to the provided int64 value.
func Int64Ptr(i int64) *int64 {
	return &i
}

// Float64Ptr returns a pointer to the provided float64 value.
func Float64Ptr(f float64) *float64 {
	return &f
}
