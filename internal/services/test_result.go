package services

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
	"parent-bot/internal/repository"
)

// TestResultService handles test result business logic
type TestResultService struct {
	repo        *repository.TestResultRepository
	studentRepo *repository.StudentRepository
}

// NewTestResultService creates a new test result service
func NewTestResultService(db *sql.DB) *TestResultService {
	return &TestResultService{
		repo:        repository.NewTestResultRepository(db),
		studentRepo: repository.NewStudentRepository(db),
	}
}

// CreateTestResult creates a new test result
func (s *TestResultService) CreateTestResult(req *models.CreateTestResultRequest) (int64, error) {
	// Verify student exists
	_, err := s.studentRepo.GetByID(req.StudentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("student not found")
		}
		return 0, err
	}

	// Ensure either teacher or admin is set
	if req.TeacherID == nil && req.AdminID == nil {
		return 0, fmt.Errorf("either teacher or admin must be specified")
	}

	return s.repo.Create(req)
}

// GetTestResultByID retrieves a test result by ID
func (s *TestResultService) GetTestResultByID(id int) (*models.TestResult, error) {
	return s.repo.GetByID(id)
}

// GetTestResultsByStudentID retrieves all test results for a student
func (s *TestResultService) GetTestResultsByStudentID(studentID int, limit, offset int) ([]*models.TestResultDetailed, error) {
	return s.repo.GetByStudentID(studentID, limit, offset)
}

// GetTestResultsByClassID retrieves all test results for a class
func (s *TestResultService) GetTestResultsByClassID(classID int, limit, offset int) ([]*models.TestResultDetailed, error) {
	return s.repo.GetByClassID(classID, limit, offset)
}

// GetAllTestResultsByClassID retrieves all test results for a class without pagination (for export)
func (s *TestResultService) GetAllTestResultsByClassID(classID int) ([]*models.TestResultDetailed, error) {
	return s.repo.GetAllByClassID(classID)
}

// UpdateTestResult updates a test result
func (s *TestResultService) UpdateTestResult(id int, req *models.UpdateTestResultRequest) error {
	// Check if test result exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("test result not found")
		}
		return err
	}

	return s.repo.Update(id, req)
}

// DeleteTestResult deletes a test result (admin only)
func (s *TestResultService) DeleteTestResult(id int) error {
	return s.repo.Delete(id)
}

// CountTestResults returns total number of test results
func (s *TestResultService) CountTestResults() (int, error) {
	return s.repo.Count()
}

// CountTestResultsByStudent returns number of test results for a student
func (s *TestResultService) CountTestResultsByStudent(studentID int) (int, error) {
	return s.repo.CountByStudent(studentID)
}

// GetLatestTestResultsByStudent retrieves the most recent test results for a student
func (s *TestResultService) GetLatestTestResultsByStudent(studentID int, limit int) ([]*models.TestResultDetailed, error) {
	return s.repo.GetLatestByStudent(studentID, limit)
}

// GetTestResultsForParent retrieves test results for a parent's selected child
func (s *TestResultService) GetTestResultsForParent(parentID int, currentStudentID int, limit, offset int) ([]*models.TestResultDetailed, error) {
	// Verify student is linked to parent
	isLinked, err := s.studentRepo.IsStudentLinkedToParent(parentID, currentStudentID)
	if err != nil {
		return nil, err
	}
	if !isLinked {
		return nil, fmt.Errorf("student is not linked to this parent")
	}

	return s.repo.GetByStudentID(currentStudentID, limit, offset)
}
