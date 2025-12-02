package services

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
	"parent-bot/internal/repository"
)

// AttendanceService handles attendance business logic
type AttendanceService struct {
	repo        *repository.AttendanceRepository
	studentRepo *repository.StudentRepository
	classRepo   *repository.ClassRepository
}

// NewAttendanceService creates a new attendance service
func NewAttendanceService(db *sql.DB) *AttendanceService {
	return &AttendanceService{
		repo:        repository.NewAttendanceRepository(db),
		studentRepo: repository.NewStudentRepository(db),
		classRepo:   repository.NewClassRepository(db),
	}
}

// CreateAttendance creates or updates an attendance record
func (s *AttendanceService) CreateAttendance(req *models.CreateAttendanceRequest) (int64, error) {
	// Verify student exists
	_, err := s.studentRepo.GetByID(req.StudentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("student not found")
		}
		return 0, err
	}

	// Ensure either teacher or admin is set
	if req.MarkedByTeacherID == nil && req.MarkedByAdminID == nil {
		return 0, fmt.Errorf("either teacher or admin must be specified")
	}

	return s.repo.Create(req)
}

// BulkCreateAttendance creates or updates multiple attendance records
func (s *AttendanceService) BulkCreateAttendance(req *models.BulkAttendanceRequest) error {
	// Verify class exists
	_, err := s.classRepo.GetByID(req.ClassID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("class not found")
		}
		return err
	}

	// Ensure either teacher or admin is set
	if req.MarkedByTeacherID == nil && req.MarkedByAdminID == nil {
		return fmt.Errorf("either teacher or admin must be specified")
	}

	return s.repo.BulkCreate(req)
}

// MarkAllPresentExcept marks all students in a class as present except the specified ones
func (s *AttendanceService) MarkAllPresentExcept(classID int, date string, absentStudentIDs []int, markedByTeacherID, markedByAdminID *int) error {
	// Verify class exists
	_, err := s.classRepo.GetByID(classID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("class not found")
		}
		return err
	}

	// Ensure either teacher or admin is set
	if markedByTeacherID == nil && markedByAdminID == nil {
		return fmt.Errorf("either teacher or admin must be specified")
	}

	return s.repo.MarkAllPresentExcept(classID, date, absentStudentIDs, markedByTeacherID, markedByAdminID)
}

// GetAttendanceByID retrieves an attendance record by ID
func (s *AttendanceService) GetAttendanceByID(id int) (*models.Attendance, error) {
	return s.repo.GetByID(id)
}

// GetAttendanceByStudentID retrieves attendance records for a student
func (s *AttendanceService) GetAttendanceByStudentID(studentID int, limit, offset int) ([]*models.AttendanceDetailed, error) {
	return s.repo.GetByStudentID(studentID, limit, offset)
}

// GetAttendanceByStudentIDAndDateRange retrieves attendance for a student within a date range
func (s *AttendanceService) GetAttendanceByStudentIDAndDateRange(studentID int, startDate, endDate string) ([]*models.AttendanceDetailed, error) {
	return s.repo.GetByStudentIDAndDateRange(studentID, startDate, endDate)
}

// GetAttendanceByClassIDAndDate retrieves attendance for all students in a class on a specific date
func (s *AttendanceService) GetAttendanceByClassIDAndDate(classID int, date string) ([]*models.AttendanceDetailed, error) {
	return s.repo.GetByClassIDAndDate(classID, date)
}

// GetTodayAttendanceByClass retrieves today's attendance for a specific class
func (s *AttendanceService) GetTodayAttendanceByClass(classID int) ([]*models.AttendanceDetailed, error) {
	return s.repo.GetTodayAttendanceByClass(classID)
}

// GetTodayAttendanceAllClasses retrieves today's attendance for all classes
func (s *AttendanceService) GetTodayAttendanceAllClasses() ([]*models.AttendanceDetailed, error) {
	return s.repo.GetTodayAttendanceAllClasses()
}

// UpdateAttendance updates an attendance record
func (s *AttendanceService) UpdateAttendance(id int, req *models.UpdateAttendanceRequest) error {
	// Check if attendance exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("attendance record not found")
		}
		return err
	}

	return s.repo.Update(id, req)
}

// DeleteAttendance deletes an attendance record (restricted)
func (s *AttendanceService) DeleteAttendance(id int) error {
	return s.repo.Delete(id)
}

// CountAttendance returns total number of attendance records
func (s *AttendanceService) CountAttendance() (int, error) {
	return s.repo.Count()
}

// IsAttendanceTaken checks if attendance has been taken for a class on a specific date
func (s *AttendanceService) IsAttendanceTaken(classID int, date string) (bool, error) {
	return s.repo.IsAttendanceTaken(classID, date)
}

// GetLast30DaysByStudent retrieves last 30 days attendance for a student
func (s *AttendanceService) GetLast30DaysByStudent(studentID int) ([]*models.AttendanceDetailed, error) {
	return s.repo.GetLast30DaysByStudent(studentID)
}

// GetAttendanceForParent retrieves attendance for a parent's selected child (last 30 days)
func (s *AttendanceService) GetAttendanceForParent(parentID int, currentStudentID int) ([]*models.AttendanceDetailed, error) {
	// Verify student is linked to parent
	isLinked, err := s.studentRepo.IsStudentLinkedToParent(parentID, currentStudentID)
	if err != nil {
		return nil, err
	}
	if !isLinked {
		return nil, fmt.Errorf("student is not linked to this parent")
	}

	return s.repo.GetLast30DaysByStudent(currentStudentID)
}
