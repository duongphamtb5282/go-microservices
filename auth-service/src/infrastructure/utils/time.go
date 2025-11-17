package utils

import (
	"time"
)

// ParseDuration parses a duration string to time.Duration
// Returns a default value if parsing fails or string is empty
func ParseDuration(durationStr string) time.Duration {
	if durationStr == "" {
		return time.Hour // Default value
	}
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return time.Hour // Default value on error
	}
	return duration
}

// ParseDurationWithDefault parses a duration string with a custom default value
func ParseDurationWithDefault(durationStr string, defaultValue time.Duration) time.Duration {
	if durationStr == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return defaultValue
	}
	return duration
}

// FormatDuration formats a duration to a human-readable string
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return d.Round(time.Second).String()
	}
	if d < time.Hour {
		return d.Round(time.Minute).String()
	}
	if d < 24*time.Hour {
		return d.Round(time.Hour).String()
	}
	return d.Round(24 * time.Hour).String()
}

// IsValidDuration checks if a duration string is valid
func IsValidDuration(durationStr string) bool {
	if durationStr == "" {
		return false
	}
	_, err := time.ParseDuration(durationStr)
	return err == nil
}
