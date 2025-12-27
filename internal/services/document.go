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

// GenerateClassTestResultsDocument generates a DOCX document for class test results
func (s *DocumentService) GenerateClassTestResultsDocument(className string, results []*models.TestResultDetailed) (filePath, filename string, err error) {
	// Generate filename
	filename = fmt.Sprintf("Baholar_%s_%s.docx", className, time.Now().Format("2006-01-02"))

	// Create full path
	filePath = filepath.Join(s.tempDir, filename)

	// Prepare test result data
	testResults := make([]docx.TestResultData, 0)
	for _, result := range results {
		testResults = append(testResults, docx.TestResultData{
			StudentName: fmt.Sprintf("%s %s", result.LastName, result.FirstName),
			SubjectName: result.SubjectName,
			Score:       result.Score,
		})
	}

	data := &docx.ClassTestResultsData{
		ClassName:   className,
		Date:        time.Now(),
		TestResults: testResults,
	}

	// Generate document
	if err := docx.GenerateClassTestResults(data, filePath); err != nil {
		return "", "", fmt.Errorf("failed to generate test results document: %w", err)
	}

	// Verify file was created and has content
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to verify generated file: %w", err)
	}

	if fileInfo.Size() == 0 {
		return "", "", fmt.Errorf("generated file is empty")
	}

	fmt.Printf("[DEBUG] Test results document verified: %s, size: %d bytes\n", filename, fileInfo.Size())

	return filePath, filename, nil
}

// GenerateTodayAttendanceDocument generates a DOCX document for today's attendance across all classes
func (s *DocumentService) GenerateTodayAttendanceDocument(classesData []struct {
	ClassName string
	Records   []*models.AttendanceDetailed
}) (filePath, filename string, err error) {
	// Generate filename
	filename = fmt.Sprintf("Yoqlama_%s.docx", time.Now().Format("2006-01-02"))

	// Create full path
	filePath = filepath.Join(s.tempDir, filename)

	// Prepare attendance data
	allClassesData := make([]docx.ClassAttendanceData, 0)

	for _, classData := range classesData {
		records := make([]docx.AttendanceRecord, 0)
		presentCount := 0
		absentCount := 0

		for _, record := range classData.Records {
			records = append(records, docx.AttendanceRecord{
				StudentName: fmt.Sprintf("%s %s", record.LastName, record.FirstName),
				Status:      record.Status,
			})

			if record.Status == "present" {
				presentCount++
			} else {
				absentCount++
			}
		}

		allClassesData = append(allClassesData, docx.ClassAttendanceData{
			ClassName:         classData.ClassName,
			Date:              time.Now(),
			AttendanceRecords: records,
			PresentCount:      presentCount,
			AbsentCount:       absentCount,
		})
	}

	data := &docx.AllClassesAttendanceData{
		Date:        time.Now(),
		ClassesData: allClassesData,
	}

	// Generate document
	if err := docx.GenerateTodayAttendance(data, filePath); err != nil {
		return "", "", fmt.Errorf("failed to generate attendance document: %w", err)
	}

	// Verify file was created and has content
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to verify generated file: %w", err)
	}

	if fileInfo.Size() == 0 {
		return "", "", fmt.Errorf("generated file is empty")
	}

	fmt.Printf("[DEBUG] Attendance document verified: %s, size: %d bytes\n", filename, fileInfo.Size())

	return filePath, filename, nil
}
