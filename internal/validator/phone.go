package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateUzbekPhone validates Uzbek phone number format
// Expected format: +998XXXXXXXXX (exactly 13 characters)
func ValidateUzbekPhone(phone string) (string, error) {
	// Remove all whitespace
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.TrimSpace(cleaned)

	// Check if it starts with +998
	if !strings.HasPrefix(cleaned, "+998") {
		// Try to add +998 if user entered without it
		if strings.HasPrefix(cleaned, "998") {
			cleaned = "+" + cleaned
		} else if len(cleaned) == 9 {
			// User entered only 9 digits
			cleaned = "+998" + cleaned
		} else {
			return "", fmt.Errorf("telefon raqam +998 bilan boshlanishi kerak / номер должен начинаться с +998")
		}
	}

	// Validate format: +998XXXXXXXXX (exactly 13 characters)
	re := regexp.MustCompile(`^\+998[0-9]{9}$`)
	if !re.MatchString(cleaned) {
		return "", fmt.Errorf("noto'g'ri telefon format. Namuna: +998901234567 / неверный формат телефона. Образец: +998901234567")
	}

	// Additional validation: check operator codes
	// Valid Uzbek operator codes as of 2024:
	// Beeline: 90, 91, 20
	// Ucell: 93, 94, 50, 55
	// UMS/Humans: 66, 88
	// Mobiuz: 97, 98, 99
	// Uzmobile: 33, 95, 77, 71
	validOperators := []string{
		"90", "91", "20",        // Beeline
		"93", "94", "50", "55",  // Ucell
		"66", "88",              // UMS/Humans
		"97", "98", "99",        // Mobiuz
		"33", "95", "77", "71",  // Uzmobile
	}
	operatorCode := cleaned[4:6]

	isValid := false
	for _, op := range validOperators {
		if operatorCode == op {
			isValid = true
			break
		}
	}

	if !isValid {
		return "", fmt.Errorf("noto'g'ri operator kodi: %s. Iltimos, O'zbekiston operator raqamini kiriting / неверный код оператора: %s. Пожалуйста, введите номер узбекского оператора", operatorCode, operatorCode)
	}

	return cleaned, nil
}

// NormalizePhone normalizes phone number to standard format
func NormalizePhone(phone string) string {
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	return strings.TrimSpace(cleaned)
}
