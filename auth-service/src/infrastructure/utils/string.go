package utils

import (
	"strings"
	"unicode"
)

// IsEmpty checks if a string is empty or contains only whitespace
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty checks if a string is not empty and contains non-whitespace characters
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// Truncate truncates a string to the specified length
func Truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// Capitalize capitalizes the first letter of a string
func Capitalize(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// ToSnakeCase converts a string to snake_case
func ToSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// ToCamelCase converts a string to camelCase
func ToCamelCase(s string) string {
	words := strings.Split(s, "_")
	if len(words) == 0 {
		return s
	}

	result := strings.ToLower(words[0])
	for _, word := range words[1:] {
		if word != "" {
			result += Capitalize(strings.ToLower(word))
		}
	}
	return result
}
