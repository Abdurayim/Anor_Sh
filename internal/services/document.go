package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"parent-bot/internal/models"
	"parent-bot/internal/utils"
	"parent-bot/pkg/docx"
)

// DocumentService handles document generation and management
type DocumentService struct {
	tempDir string
}

// NewDocumentService creates a new document service
func NewDocumentService(tempDir string) *DocumentService {
	return &DocumentService{tempDir: tempDir}
}

// GenerateComplaintDocument generates a DOCX document for a complaint
// Returns the file path and filename
func (s *DocumentService) GenerateComplaintDocument(user *models.User, student *models.StudentWithClass, complaintText string) (filePath, filename string, err error) {
	// Generate filename
	childFullName := fmt.Sprintf("%s %s", student.LastName, student.FirstName)
	filename = utils.GenerateComplaintFilename(childFullName, student.ClassName)

	// Create full path
	filePath = filepath.Join(s.tempDir, filename)

	// Prepare document data
	data := &docx.ComplaintData{
		ChildName:     childFullName,
		ChildClass:    student.ClassName,
		PhoneNumber:   user.PhoneNumber,
		ComplaintText: complaintText,
		ParentName:    childFullName, // Using child name as reference
		Date:          time.Now(),
	}

	// Validate data
	if err := docx.ValidateData(data); err != nil {
		return "", "", fmt.Errorf("invalid complaint data: %w", err)
	}

	// Generate document
	if err := docx.Generate(data, filePath); err != nil {
		return "", "", fmt.Errorf("failed to generate document: %w", err)
	}

	// Verify file was created and has content
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to verify generated file: %w", err)
	}

	if fileInfo.Size() == 0 {
		return "", "", fmt.Errorf("generated file is empty")
	}

	if fileInfo.Size() < 1000 {
		return "", "", fmt.Errorf("generated file is too small (%d bytes), might be corrupted", fileInfo.Size())
	}

	fmt.Printf("[DEBUG] Document verified: %s, size: %d bytes\n", filename, fileInfo.Size())

	return filePath, filename, nil
}

// GenerateProposalDocument generates a DOCX document for a proposal
// Returns the file path and filename
func (s *DocumentService) GenerateProposalDocument(user *models.User, student *models.StudentWithClass, proposalText string) (filePath, filename string, err error) {
	// Generate filename (similar to complaint, but with "proposal" prefix)
	childFullName := fmt.Sprintf("%s %s", student.LastName, student.FirstName)
	filename = utils.GenerateProposalFilename(childFullName, student.ClassName)

	// Create full path
	filePath = filepath.Join(s.tempDir, filename)

	// Prepare document data (reusing ComplaintData structure with different title)
	data := &docx.ComplaintData{
		ChildName:     childFullName,
		ChildClass:    student.ClassName,
		PhoneNumber:   user.PhoneNumber,
		ComplaintText: proposalText, // Using same field for proposal text
		ParentName:    childFullName,
		Date:          time.Now(),
	}

	// Validate data
	if err := docx.ValidateData(data); err != nil {
		return "", "", fmt.Errorf("invalid proposal data: %w", err)
	}

	// Generate document with proposal flag
	if err := docx.GenerateProposal(data, filePath); err != nil {
		return "", "", fmt.Errorf("failed to generate document: %w", err)
	}

	// Verify file was created and has content
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to verify generated file: %w", err)
	}

	if fileInfo.Size() == 0 {
		return "", "", fmt.Errorf("generated file is empty")
	}

	if fileInfo.Size() < 1000 {
		return "", "", fmt.Errorf("generated file is too small (%d bytes), might be corrupted", fileInfo.Size())
	}

	fmt.Printf("[DEBUG] Document verified: %s, size: %d bytes\n", filename, fileInfo.Size())

	return filePath, filename, nil
}

// DeleteTempFile deletes a temporary file
func (s *DocumentService) DeleteTempFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete temp file: %w", err)
	}

	return nil
}

// CleanTempDirectory cleans old temporary files
func (s *DocumentService) CleanTempDirectory(maxAge time.Duration) error {
	files, err := os.ReadDir(s.tempDir)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %w", err)
	}

	now := time.Now()
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// Delete files older than maxAge
		if now.Sub(info.ModTime()) > maxAge {
			filePath := filepath.Join(s.tempDir, file.Name())
			_ = os.Remove(filePath) // Ignore errors
		}
	}

	return nil
}

// GetTempDir returns the temporary directory path
func (s *DocumentService) GetTempDir() string {
	return s.tempDir
}
