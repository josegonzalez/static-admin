package util

// StringFromPointer safely dereferences a string pointer
func StringFromPointer(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
