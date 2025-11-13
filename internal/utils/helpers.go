package utils

import (
	"fmt"
	"strings"
	"time"

	"parent-bot/internal/validator"
)

// GenerateComplaintFilename generates a filename for complaint document
// Format: Shikoyat_ParentName_ClassName_Date.docx
func GenerateComplaintFilename(childName, childClass string) string {
	date := time.Now().Format("2006-01-02")

	// Sanitize name
	safeName := validator.SanitizeFilename(childName)
	safeName = strings.ReplaceAll(safeName, " ", "_")

	// Create filename
	filename := fmt.Sprintf("Shikoyat_%s_%s_sinf_%s.docx", safeName, childClass, date)

	return filename
}

// GenerateProposalFilename generates a filename for proposal document
// Format: Taklif_ParentName_ClassName_Date.docx
func GenerateProposalFilename(childName, childClass string) string {
	date := time.Now().Format("2006-01-02")

	// Sanitize name
	safeName := validator.SanitizeFilename(childName)
	safeName = strings.ReplaceAll(safeName, " ", "_")

	// Create filename
	filename := fmt.Sprintf("Taklif_%s_%s_sinf_%s.docx", safeName, childClass, date)

	return filename
}

// GenerateComplaintCaption generates caption for complaint document
func GenerateComplaintCaption(childName, childClass, phoneNumber string) string {
	return fmt.Sprintf(
		"YANGI SHIKOYAT / НОВАЯ ЖАЛОБА\n\n"+
			"Ota-ona / Родитель: %s\n"+
			"Sinf / Класс: %s\n"+
			"Telefon / Телефон: %s\n"+
			"Sana / Дата: %s",
		childName,
		childClass,
		phoneNumber,
		time.Now().Format("02.01.2006 15:04"),
	)
}

// GenerateProposalCaption generates caption for proposal document
func GenerateProposalCaption(childName, childClass, phoneNumber string) string {
	return fmt.Sprintf(
		"YANGI TAKLIF / НОВОЕ ПРЕДЛОЖЕНИЕ\n\n"+
			"Ota-ona / Родитель: %s\n"+
			"Sinf / Класс: %s\n"+
			"Telefon / Телефон: %s\n"+
			"Sana / Дата: %s",
		childName,
		childClass,
		phoneNumber,
		time.Now().Format("02.01.2006 15:04"),
	)
}

// TruncateText truncates text to specified length
func TruncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}

	return text[:maxLen] + "..."
}

// FormatPhoneNumber formats phone number for display
func FormatPhoneNumber(phone string) string {
	// +998 90 123 45 67
	if len(phone) != 13 {
		return phone
	}

	return fmt.Sprintf("%s %s %s %s %s",
		phone[:4],   // +998
		phone[4:6],  // 90
		phone[6:9],  // 123
		phone[9:11], // 45
		phone[11:],  // 67
	)
}

// EscapeMarkdown escapes special characters for Telegram Markdown
func EscapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)

	return replacer.Replace(text)
}

// FormatDateTime formats datetime for display
func FormatDateTime(t time.Time) string {
	return t.Format("02.01.2006 15:04")
}

// FormatDate formats date for display
func FormatDate(t time.Time) string {
	return t.Format("02.01.2006")
}

// SanitizeClassName sanitizes class name
func SanitizeClassName(className string) string {
	// Remove extra spaces and trim
	className = strings.TrimSpace(className)
	// Remove multiple spaces
	className = strings.Join(strings.Fields(className), " ")
	return className
}
