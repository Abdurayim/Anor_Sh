package services

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
	"parent-bot/internal/repository"
	"parent-bot/internal/validator"
)

// TeacherService handles teacher business logic
type TeacherService struct {
	repo      *repository.TeacherRepository
	classRepo *repository.ClassRepository
}

// NewTeacherService creates a new teacher service
func NewTeacherService(db *sql.DB) *TeacherService {
	return &TeacherService{
		repo:      repository.NewTeacherRepository(db),
		classRepo: repository.NewClassRepository(db),
	}
}

// CreateTeacher creates a new teacher
func (s *TeacherService) CreateTeacher(req *models.CreateTeacherRequest) (int64, error) {
	// Validate and normalize phone number
	normalizedPhone, err := validator.ValidateUzbekPhone(req.PhoneNumber)
	if err != nil {
		return 0, fmt.Errorf("invalid phone number: %w", err)
	}
	req.PhoneNumber = normalizedPhone

	// Check if teacher already exists
	existing, _ := s.repo.GetByPhoneNumber(normalizedPhone)
	if existing != nil {
		return 0, fmt.Errorf("teacher with phone %s already exists", normalizedPhone)
	}

	// Create teacher
	return s.repo.Create(req.FirstName, req.LastName, req.PhoneNumber, req.Language, req.AddedByAdminID)
}

// GetTeacherByID retrieves a teacher by ID
func (s *TeacherService) GetTeacherByID(id int) (*models.Teacher, error) {
	return s.repo.GetByID(id)
}

// GetTeacherByPhoneNumber retrieves a teacher by phone number
func (s *TeacherService) GetTeacherByPhoneNumber(phoneNumber string) (*models.Teacher, error) {
	// Normalize phone first
	normalizedPhone, err := validator.ValidateUzbekPhone(phoneNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid phone number: %w", err)
	}

	return s.repo.GetByPhoneNumber(normalizedPhone)
}

// GetTeacherByTelegramID retrieves a teacher by Telegram ID
func (s *TeacherService) GetTeacherByTelegramID(telegramID int64) (*models.Teacher, error) {
	return s.repo.GetByTelegramID(telegramID)
}

// LinkTelegramID links a Telegram account to a teacher
func (s *TeacherService) LinkTelegramID(phoneNumber string, telegramID int64, language string) error {
	// Normalize phone first
	normalizedPhone, err := validator.ValidateUzbekPhone(phoneNumber)
	if err != nil {
		return fmt.Errorf("invalid phone number: %w", err)
	}

	// Check if teacher exists
	teacher, err := s.repo.GetByPhoneNumber(normalizedPhone)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("teacher not found with phone number %s", normalizedPhone)
		}
		return err
	}

	// Check if already linked to different telegram ID
	if teacher.TelegramID != nil && *teacher.TelegramID != telegramID {
		return fmt.Errorf("this phone number is already linked to another Telegram account")
	}

	return s.repo.LinkTelegramID(normalizedPhone, telegramID, language)
}

// UpdateTeacher updates teacher information
func (s *TeacherService) UpdateTeacher(id int, req *models.UpdateTeacherRequest) error {
	// Check if teacher exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("teacher not found")
		}
		return err
	}

	return s.repo.Update(id, req)
}

// DeleteTeacher deletes a teacher
func (s *TeacherService) DeleteTeacher(id int) error {
	return s.repo.Delete(id)
}

// DeactivateTeacher deactivates a teacher (soft delete)
func (s *TeacherService) DeactivateTeacher(id int) error {
	isActive := false
	return s.repo.Update(id, &models.UpdateTeacherRequest{
		IsActive: &isActive,
	})
}

// GetAllTeachers retrieves all teachers with pagination
func (s *TeacherService) GetAllTeachers(limit, offset int) ([]*models.Teacher, error) {
	return s.repo.GetAll(limit, offset)
}

// GetActiveTeachers retrieves all active teachers
func (s *TeacherService) GetActiveTeachers() ([]*models.Teacher, error) {
	return s.repo.GetActiveTeachers()
}

// CountTeachers returns total number of teachers
func (s *TeacherService) CountTeachers() (int, error) {
	return s.repo.Count()
}

// AssignToClass assigns a teacher to a class
func (s *TeacherService) AssignToClass(teacherID, classID int) error {
	// Verify teacher exists
	_, err := s.repo.GetByID(teacherID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("teacher not found")
		}
		return err
	}

	// Verify class exists
	_, err = s.classRepo.GetByID(classID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("class not found")
		}
		return err
	}

	return s.repo.AssignToClass(teacherID, classID)
}

// RemoveFromClass removes a teacher from a class
func (s *TeacherService) RemoveFromClass(teacherID, classID int) error {
	return s.repo.RemoveFromClass(teacherID, classID)
}

// GetTeacherClasses retrieves all classes assigned to a teacher
func (s *TeacherService) GetTeacherClasses(teacherID int) ([]*models.Class, error) {
	return s.repo.GetTeacherClasses(teacherID)
}

// GetClassTeachers retrieves all teachers assigned to a class
func (s *TeacherService) GetClassTeachers(classID int) ([]*models.Teacher, error) {
	return s.repo.GetClassTeachers(classID)
}

// IsTeacherAssignedToClass checks if a teacher is assigned to a class
func (s *TeacherService) IsTeacherAssignedToClass(teacherID, classID int) (bool, error) {
	return s.repo.IsTeacherAssignedToClass(teacherID, classID)
}

// IsTeacher checks if a phone number or telegram ID belongs to a teacher
func (s *TeacherService) IsTeacher(phoneNumber string, telegramID int64) (bool, *models.Teacher, error) {
	// Normalize phone if provided
	if phoneNumber != "" {
		normalized, err := validator.ValidateUzbekPhone(phoneNumber)
		if err == nil {
			phoneNumber = normalized
		}
	}

	return s.repo.IsTeacher(phoneNumber, telegramID)
}

// GetTeacherFullName returns the full name of a teacher
func (s *TeacherService) GetTeacherFullName(teacher *models.Teacher) string {
	if teacher == nil {
		return ""
	}
	return teacher.LastName + " " + teacher.FirstName
}
