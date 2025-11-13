package docx

import (
	"fmt"
	"os"
	"time"

	"github.com/fumiama/go-docx"
)

// ComplaintData holds data for complaint document
type ComplaintData struct {
	ChildName     string
	ChildClass    string
	PhoneNumber   string
	ComplaintText string
	ParentName    string
	Date          time.Time
}

// Generate generates a formatted DOCX document for a complaint
func Generate(data *ComplaintData, outputPath string) error {
	// Create new document with default theme and A4 page
	doc := docx.New().WithDefaultTheme().WithA4Page()

	// Add header/title
	para := doc.AddParagraph()
	para.AddText("SHIKOYAT / ЖАЛОБА").Size("32").Bold()
	para.Justification("center")

	// Add spacing
	doc.AddParagraph()

	// Add date
	para = doc.AddParagraph()
	para.AddText(fmt.Sprintf("Sana / Дата: %s", data.Date.Format("02.01.2006")))

	// Add spacing
	doc.AddParagraph()

	// Add parent/student information section
	para = doc.AddParagraph()
	para.AddText("OTA-ONA MA'LUMOTLARI / ИНФОРМАЦИЯ О РОДИТЕЛЕ:").Bold()

	doc.AddParagraph()

	// Child name
	para = doc.AddParagraph()
	para.AddText("Farzand ismi / Имя ребенка:")
	para = doc.AddParagraph()
	para.AddText(data.ChildName)

	doc.AddParagraph()

	// Class
	para = doc.AddParagraph()
	para.AddText("Sinf / Класс:")
	para = doc.AddParagraph()
	para.AddText(data.ChildClass)

	doc.AddParagraph()

	// Phone number
	para = doc.AddParagraph()
	para.AddText("Telefon raqam / Номер телефона:")
	para = doc.AddParagraph()
	para.AddText(data.PhoneNumber)

	// Add spacing
	doc.AddParagraph()
	doc.AddParagraph()

	// Add complaint text section
	para = doc.AddParagraph()
	para.AddText("SHIKOYAT MATNI / ТЕКСТ ЖАЛОБЫ:").Bold()

	doc.AddParagraph()

	// Add complaint text
	para = doc.AddParagraph()
	para.AddText(data.ComplaintText)

	// Add spacing
	doc.AddParagraph()
	doc.AddParagraph()

	// Add footer
	para = doc.AddParagraph()
	para.AddText("Hujjat avtomatik tarzda yaratilgan / Документ создан автоматически").Size("18")
	para.Justification("center")

	para = doc.AddParagraph()
	para.AddText(fmt.Sprintf("Yaratilgan / Создано: %s", time.Now().Format("02.01.2006 15:04"))).Size("18")
	para.Justification("center")

	// Save document
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	written, err := doc.WriteTo(f)
	if err != nil {
		return fmt.Errorf("failed to write document: %w", err)
	}

	// Ensure all data is written to disk before returning
	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	fmt.Printf("[DEBUG] DOCX generated: %s, size: %d bytes\n", outputPath, written)

	return nil
}

// GenerateProposal generates a formatted DOCX document for a proposal
func GenerateProposal(data *ComplaintData, outputPath string) error {
	// Create new document with default theme and A4 page
	doc := docx.New().WithDefaultTheme().WithA4Page()

	// Add header/title
	para := doc.AddParagraph()
	para.AddText("TAKLIF / ПРЕДЛОЖЕНИЕ").Size("32").Bold()
	para.Justification("center")

	// Add spacing
	doc.AddParagraph()

	// Add date
	para = doc.AddParagraph()
	para.AddText(fmt.Sprintf("Sana / Дата: %s", data.Date.Format("02.01.2006")))

	// Add spacing
	doc.AddParagraph()

	// Add parent/student information section
	para = doc.AddParagraph()
	para.AddText("OTA-ONA MA'LUMOTLARI / ИНФОРМАЦИЯ О РОДИТЕЛЕ:").Bold()

	doc.AddParagraph()

	// Child name
	para = doc.AddParagraph()
	para.AddText("Farzand ismi / Имя ребенка:")
	para = doc.AddParagraph()
	para.AddText(data.ChildName)

	doc.AddParagraph()

	// Class
	para = doc.AddParagraph()
	para.AddText("Sinf / Класс:")
	para = doc.AddParagraph()
	para.AddText(data.ChildClass)

	doc.AddParagraph()

	// Phone number
	para = doc.AddParagraph()
	para.AddText("Telefon raqam / Номер телефона:")
	para = doc.AddParagraph()
	para.AddText(data.PhoneNumber)

	// Add spacing
	doc.AddParagraph()
	doc.AddParagraph()

	// Add proposal text section
	para = doc.AddParagraph()
	para.AddText("TAKLIF MATNI / ТЕКСТ ПРЕДЛОЖЕНИЯ:").Bold()

	doc.AddParagraph()

	// Add proposal text
	para = doc.AddParagraph()
	para.AddText(data.ComplaintText) // Reusing ComplaintText field for proposal

	// Add spacing
	doc.AddParagraph()
	doc.AddParagraph()

	// Add footer
	para = doc.AddParagraph()
	para.AddText("Hujjat avtomatik tarzda yaratilgan / Документ создан автоматически").Size("18")
	para.Justification("center")

	para = doc.AddParagraph()
	para.AddText(fmt.Sprintf("Yaratilgan / Создано: %s", time.Now().Format("02.01.2006 15:04"))).Size("18")
	para.Justification("center")

	// Save document
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	written, err := doc.WriteTo(f)
	if err != nil {
		return fmt.Errorf("failed to write document: %w", err)
	}

	// Ensure all data is written to disk before returning
	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	fmt.Printf("[DEBUG] DOCX generated: %s, size: %d bytes\n", outputPath, written)

	return nil
}

// ValidateData validates complaint data before generating document
func ValidateData(data *ComplaintData) error {
	if data.ChildName == "" {
		return fmt.Errorf("child name is required")
	}

	if data.ChildClass == "" {
		return fmt.Errorf("child class is required")
	}

	if data.PhoneNumber == "" {
		return fmt.Errorf("phone number is required")
	}

	if data.ComplaintText == "" {
		return fmt.Errorf("complaint text is required")
	}

	if len(data.ComplaintText) < 10 {
		return fmt.Errorf("complaint text is too short")
	}

	return nil
}
