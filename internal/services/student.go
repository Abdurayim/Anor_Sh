package services

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
	"parent-bot/internal/repository"
)

// StudentService handles student business logic
type StudentService struct {
	repo      *repository.StudentRepository
	classRepo *repository.ClassRepository
	userRepo  *repository.UserRepository
}

// NewStudentService creates a new student service
func NewStudentService(db *sql.DB) *StudentService {
	return &StudentService{
		repo:      repository.NewStudentRepository(db),
		classRepo: repository.NewClassRepository(db),
		userRepo:  repository.NewUserRepository(db),
	}
}

// CreateStudent creates a new student
func (s *StudentService) CreateStudent(req *models.CreateStudentRequest) (int64, error) {
	// Verify class exists
	_, err := s.classRepo.GetByID(req.ClassID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("class not found")
		}
		return 0, err
	}

	// Ensure either admin or teacher is set
	if req.AddedByAdminID == nil && req.AddedByTeacherID == nil {
		return 0, fmt.Errorf("either admin or teacher must be specified")
	}

	return s.repo.Create(req)
}

// GetStudentByID retrieves a student by ID
func (s *StudentService) GetStudentByID(id int) (*models.Student, error) {
	return s.repo.GetByID(id)
}

// GetStudentByIDWithClass retrieves a student with class information
func (s *StudentService) GetStudentByIDWithClass(id int) (*models.StudentWithClass, error) {
	return s.repo.GetByIDWithClass(id)
}

// GetStudentsByClassID retrieves all students in a class
func (s *StudentService) GetStudentsByClassID(classID int) ([]*models.StudentWithClass, error) {
	return s.repo.GetByClassID(classID)
}

// SearchStudentsByName searches for students by name in a specific class
func (s *StudentService) SearchStudentsByName(classID int, searchTerm string) ([]*models.StudentWithClass, error) {
	return s.repo.SearchByName(classID, searchTerm)
}

// UpdateStudent updates student information
func (s *StudentService) UpdateStudent(id int, req *models.UpdateStudentRequest) error {
	// Check if student exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("student not found")
		}
		return err
	}

	// If updating class, verify new class exists
	if req.ClassID != nil {
		_, err := s.classRepo.GetByID(*req.ClassID)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("new class not found")
			}
			return err
		}
	}

	return s.repo.Update(id, req)
}

// DeleteStudent soft deletes a student
func (s *StudentService) DeleteStudent(id int) error {
	return s.repo.Delete(id)
}

// HardDeleteStudent permanently deletes a student
func (s *StudentService) HardDeleteStudent(id int) error {
	return s.repo.HardDelete(id)
}

// GetAllStudents retrieves all students with pagination
func (s *StudentService) GetAllStudents(limit, offset int) ([]*models.StudentWithClass, error) {
	return s.repo.GetAll(limit, offset)
}

// CountStudents returns total number of active students
func (s *StudentService) CountStudents() (int, error) {
	return s.repo.Count()
}

// CountStudentsByClass returns number of students in a class
func (s *StudentService) CountStudentsByClass(classID int) (int, error) {
	return s.repo.CountByClass(classID)
}

// LinkToParent links a student to a parent
func (s *StudentService) LinkToParent(parentID, studentID int) error {
	// Check if parent exists
	_, err := s.userRepo.GetByID(parentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("parent not found")
		}
		return err
	}

	// Check if student exists
	_, err = s.repo.GetByID(studentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("student not found")
		}
		return err
	}

	// Check if already linked
	isLinked, err := s.repo.IsStudentLinkedToParent(parentID, studentID)
	if err != nil {
		return err
	}
	if isLinked {
		return fmt.Errorf("student already linked to this parent")
	}

	// Check parent's current children count
	count, err := s.repo.CountParentStudents(parentID)
	if err != nil {
		return err
	}
	if count >= 4 {
		return fmt.Errorf("parent already has maximum 4 children")
	}

	return s.repo.LinkToParent(parentID, studentID)
}

// UnlinkFromParent removes the link between student and parent
func (s *StudentService) UnlinkFromParent(parentID, studentID int) error {
	return s.repo.UnlinkFromParent(parentID, studentID)
}

// GetStudentParents retrieves all parents linked to a student
func (s *StudentService) GetStudentParents(studentID int) ([]*models.User, error) {
	return s.repo.GetStudentParents(studentID)
}

// GetParentStudents retrieves all students linked to a parent
func (s *StudentService) GetParentStudents(parentID int) ([]*models.ParentChild, error) {
	return s.repo.GetParentStudents(parentID)
}

// CountParentStudents returns the number of students linked to a parent
func (s *StudentService) CountParentStudents(parentID int) (int, error) {
	return s.repo.CountParentStudents(parentID)
}

// IsStudentLinkedToParent checks if a student is already linked to a parent
func (s *StudentService) IsStudentLinkedToParent(parentID, studentID int) (bool, error) {
	return s.repo.IsStudentLinkedToParent(parentID, studentID)
}

// GetStudentFullName returns the full name of a student
func (s *StudentService) GetStudentFullName(student *models.Student) string {
	if student == nil {
		return ""
	}
	return student.LastName + " " + student.FirstName
}

// GetStudentWithClassFullName returns the full name of a student with class
func (s *StudentService) GetStudentWithClassFullName(student *models.StudentWithClass) string {
	if student == nil {
		return ""
	}
	return student.LastName + " " + student.FirstName
}

// GetStudentsByIDs retrieves multiple students by their IDs
func (s *StudentService) GetStudentsByIDs(studentIDs []int) ([]*models.StudentWithClass, error) {
	return s.repo.GetStudentsByIDs(studentIDs)
}

// CanParentAddChild checks if a parent can add another child
func (s *StudentService) CanParentAddChild(parentID int) (bool, error) {
	count, err := s.repo.CountParentStudents(parentID)
	if err != nil {
		return false, err
	}
	return count < 4, nil
}

// SetCurrentSelectedStudent sets the current selected student for a parent
func (s *StudentService) SetCurrentSelectedStudent(parentID, studentID int) error {
	// Verify student is linked to parent
	isLinked, err := s.repo.IsStudentLinkedToParent(parentID, studentID)
	if err != nil {
		return err
	}
	if !isLinked {
		return fmt.Errorf("student is not linked to this parent")
	}

	// Update user's current selected student
	return s.userRepo.Update(parentID, &models.UpdateUserRequest{
		CurrentSelectedStudentID: &studentID,
	})
}

// GetCurrentSelectedStudent retrieves the current selected student for a parent
func (s *StudentService) GetCurrentSelectedStudent(parentID int) (*models.StudentWithClass, error) {
	user, err := s.userRepo.GetByID(parentID)
	if err != nil {
		return nil, err
	}

	if user.CurrentSelectedStudentID == nil {
		return nil, fmt.Errorf("no student selected")
	}

	return s.repo.GetByIDWithClass(*user.CurrentSelectedStudentID)
}
