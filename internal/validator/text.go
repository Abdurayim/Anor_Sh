package validator

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

// SanitizeInput sanitizes user input to prevent injection attacks
// Removes dangerous characters: +, @, _, %, $, <, >, and SQL injection patterns
func SanitizeInput(text string) string {
	// Trim whitespace
	text = strings.TrimSpace(text)

	// Remove SQL injection patterns
	sqlPatterns := []string{
		"--", "/*", "*/", "xp_", "sp_", "exec", "execute",
		"union", "select", "insert", "update", "delete", "drop", "create",
		";", "||", "&&",
	}

	lowerText := strings.ToLower(text)
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerText, pattern) {
			// Replace with empty string
			text = strings.ReplaceAll(text, pattern, "")
			text = strings.ReplaceAll(text, strings.ToUpper(pattern), "")
			text = strings.ReplaceAll(text, strings.Title(pattern), "")
		}
	}

	// Escape HTML to prevent XSS
	text = html.EscapeString(text)

	return strings.TrimSpace(text)
}

// ValidateName validates user name (child name)
// Allows: letters, spaces, hyphens, apostrophes
// Disallows: numbers, special characters (+, @, _, %, $, etc.)
func ValidateName(name string) (string, error) {
	// Trim whitespace
	name = strings.TrimSpace(name)

	// Check length
	if len(name) < 2 {
		return "", fmt.Errorf("ism juda qisqa (kamida 2 ta belgi) / имя слишком короткое (минимум 2 символа)")
	}

	if len(name) > 100 {
		return "", fmt.Errorf("ism juda uzun (maksimal 100 ta belgi) / имя слишком длинное (максимум 100 символов)")
	}

	// Check for dangerous characters
	dangerousChars := []string{"+", "@", "_", "%", "$", "<", ">", "!", "#", "^", "&", "*", "(", ")", "=", "{", "}", "[", "]", "|", "\\", "/", "?"}
	for _, char := range dangerousChars {
		if strings.Contains(name, char) {
			return "", fmt.Errorf("ismda ruxsat etilmagan belgilar bor / имя содержит недопустимые символы: %s", char)
		}
	}

	// Check for numbers
	if regexp.MustCompile(`[0-9]`).MatchString(name) {
		return "", fmt.Errorf("ismda raqamlar bo'lmasligi kerak / имя не должно содержать цифры")
	}

	// Only allow letters (including Cyrillic and Latin), spaces, hyphens, apostrophes
	validPattern := regexp.MustCompile(`^[a-zA-ZА-Яа-яЎўҚқҲҳЁё\s'-]+$`)
	if !validPattern.MatchString(name) {
		return "", fmt.Errorf("ismda faqat harflar bo'lishi mumkin / имя может содержать только буквы")
	}

	return name, nil
}

// ValidateClass validates class format
// Expected formats: 1A, 2B, 11V, etc. (1-11 followed by A-Z or А-Я)
func ValidateClass(class string) (string, error) {
	// Trim and uppercase
	class = strings.TrimSpace(strings.ToUpper(class))

	if class == "" {
		return "", fmt.Errorf("sinf ko'rsatilishi kerak / необходимо указать класс")
	}

	// Check for dangerous characters
	dangerousChars := []string{"+", "@", "_", "%", "$", "<", ">", "!", "#", "^", "&", "*", "(", ")", "=", "{", "}", "[", "]", "|", "\\", "/", "?", ";"}
	for _, char := range dangerousChars {
		if strings.Contains(class, char) {
			return "", fmt.Errorf("sinfda ruxsat etilmagan belgilar bor / класс содержит недопустимые символы")
		}
	}

	// Validate format: 1-11 followed by A-Z or А-Я
	validPattern := regexp.MustCompile(`^([1-9]|1[01])[A-ZА-Я]$`)
	if !validPattern.MatchString(class) {
		return "", fmt.Errorf("noto'g'ri sinf formati. Namuna: 9A, 11B / неверный формат класса. Образец: 9A, 11B")
	}

	return class, nil
}

// ValidateComplaintText validates complaint text
func ValidateComplaintText(text string) (string, error) {
	// Trim whitespace
	text = strings.TrimSpace(text)

	// Check length
	if utf8.RuneCountInString(text) < 10 {
		return "", fmt.Errorf("shikoyat matni juda qisqa (kamida 10 ta belgi) / текст жалобы слишком короткий (минимум 10 символов)")
	}

	if utf8.RuneCountInString(text) > 5000 {
		return "", fmt.Errorf("shikoyat matni juda uzun (maksimal 5000 ta belgi) / текст жалобы слишком длинный (максимум 5000 символов)")
	}

	// Sanitize input
	text = SanitizeInput(text)

	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("shikoyat matni bo'sh bo'lishi mumkin emas / текст жалобы не может быть пустым")
	}

	return text, nil
}

// RemoveExcessWhitespace removes excess whitespace from text
func RemoveExcessWhitespace(text string) string {
	// Replace multiple spaces with single space
	space := regexp.MustCompile(`\s+`)
	text = space.ReplaceAllString(text, " ")

	// Replace multiple newlines with double newline
	newline := regexp.MustCompile(`\n{3,}`)
	text = newline.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}
