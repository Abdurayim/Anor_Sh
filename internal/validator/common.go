package validator

import (
	"fmt"
	"strings"
)

// Common validation functions

// IsEmpty checks if string is empty or contains only whitespace
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// ValidateLanguage validates language code
func ValidateLanguage(lang string) (string, error) {
	lang = strings.ToLower(strings.TrimSpace(lang))

	if lang != "uz" && lang != "ru" {
		return "", fmt.Errorf("faqat uz yoki ru tili qo'llab-quvvatlanadi / поддерживаются только uz или ru языки")
	}

	return lang, nil
}

// ContainsDangerousPatterns checks for dangerous patterns
func ContainsDangerousPatterns(text string) bool {
	dangerous := []string{
		"<script", "javascript:", "onerror=", "onclick=",
		"eval(", "exec(", "../", "..\\",
	}

	lowerText := strings.ToLower(text)
	for _, pattern := range dangerous {
		if strings.Contains(lowerText, pattern) {
			return true
		}
	}

	return false
}

// SanitizeFilename sanitizes filename to prevent path traversal
func SanitizeFilename(filename string) string {
	// Remove path separators
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, "..", "_")

	// Remove dangerous characters
	dangerous := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range dangerous {
		filename = strings.ReplaceAll(filename, char, "_")
	}

	return strings.TrimSpace(filename)
}
