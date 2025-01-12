package utils

import "time"

// Helper function to safely convert a pointer to a string
func SafeString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.String()
}
