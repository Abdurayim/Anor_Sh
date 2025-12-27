package repository

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
)

// StudentRepository handles student data operations
type StudentRepository struct {
	db *sql.DB
}

// NewStudentRepository creates a new student repository
func NewStudentRepository(db *sql.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

// Create creates a new student
func (r *StudentRepository) Create(student *models.CreateStudentRequest) (int64, error) {
	query := `
		INSERT INTO students (first_name, last_name, class_id, added_by_admin_id, added_by_teacher_id)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query,
		student.FirstName,
		student.LastName,
		student.ClassID,
		student.AddedByAdminID,
		student.AddedByTeacherID,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetByID retrieves a student by ID
func (r *StudentRepository) GetByID(id int) (*models.Student, error) {
	query := `
		SELECT id, first_name, last_name, class_id, is_active,
		       added_by_admin_id, added_by_teacher_id, created_at, updated_at
		FROM students
		WHERE id = ?
	`
	student := &models.Student{}
	err := r.db.QueryRow(query, id).Scan(
		&student.ID,
		&student.FirstName,
		&student.LastName,
		&student.ClassID,
		&student.IsActive,
		&student.AddedByAdminID,
		&student.AddedByTeacherID,
		&student.CreatedAt,
		&student.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return student, nil
}

// GetByIDWithClass retrieves a student with class information
func (r *StudentRepository) GetByIDWithClass(id int) (*models.StudentWithClass, error) {
	query := `
		SELECT id, first_name, last_name, class_id, class_name, is_active, created_at
		FROM v_students_with_class
		WHERE id = ?
	`
	student := &models.StudentWithClass{}
	err := r.db.QueryRow(query, id).Scan(
		&student.ID,
		&student.FirstName,
		&student.LastName,
		&student.ClassID,
		&student.ClassName,
		&student.IsActive,
		&student.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return student, nil
}

// GetByClassID retrieves all students in a class
func (r *StudentRepository) GetByClassID(classID int) ([]*models.StudentWithClass, error) {
	query := `
		SELECT id, first_name, last_name, class_id, class_name, is_active, created_at
		FROM v_students_with_class
		WHERE class_id = ? AND is_active = 1
		ORDER BY last_name, first_name
	`
	rows, err := r.db.Query(query, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []*models.StudentWithClass
	for rows.Next() {
		student := &models.StudentWithClass{}
		err := rows.Scan(
			&student.ID,
			&student.FirstName,
			&student.LastName,
			&student.ClassID,
			&student.ClassName,
			&student.IsActive,
			&student.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		students = append(students, student)
	}

	return students, nil
}

// SearchByName searches for students by first or last name in a specific class
func (r *StudentRepository) SearchByName(classID int, searchTerm string) ([]*models.StudentWithClass, error) {
	query := `
		SELECT id, first_name, last_name, class_id, class_name, is_active, created_at
		FROM v_students_with_class
		WHERE class_id = ? AND is_active = 1
		  AND (first_name LIKE ? OR last_name LIKE ?)
		ORDER BY last_name, first_name
	`
	searchPattern := "%" + searchTerm + "%"
	rows, err := r.db.Query(query, classID, searchPattern, searchPattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []*models.StudentWithClass
	for rows.Next() {
		student := &models.StudentWithClass{}
		err := rows.Scan(
			&student.ID,
			&student.FirstName,
			&student.LastName,
			&student.ClassID,
			&student.ClassName,
			&student.IsActive,
			&student.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		students = append(students, student)
	}

	return students, nil
}

// Update updates student information
func (r *StudentRepository) Update(id int, req *models.UpdateStudentRequest) error {
	query := `
		UPDATE students
		SET first_name = COALESCE(?, first_name),
		    last_name = COALESCE(?, last_name),
		    class_id = COALESCE(?, class_id),
		    is_active = COALESCE(?, is_active),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.Exec(query, req.FirstName, req.LastName, req.ClassID, req.IsActive, id)
	return err
}

// Delete deletes a student (soft delete by setting is_active = false)
func (r *StudentRepository) Delete(id int) error {
	query := "UPDATE students SET is_active = 0 WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// HardDelete permanently deletes a student
func (r *StudentRepository) HardDelete(id int) error {
	query := "DELETE FROM students WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// GetAll retrieves all students with pagination
func (r *StudentRepository) GetAll(limit, offset int) ([]*models.StudentWithClass, error) {
	query := `
		SELECT id, first_name, last_name, class_id, class_name, is_active, created_at
		FROM v_students_with_class
		WHERE is_active = 1
		ORDER BY class_name, last_name, first_name
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []*models.StudentWithClass
	for rows.Next() {
		student := &models.StudentWithClass{}
		err := rows.Scan(
			&student.ID,
			&student.FirstName,
			&student.LastName,
			&student.ClassID,
			&student.ClassName,
			&student.IsActive,
			&student.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		students = append(students, student)
	}

	return students, nil
}

// Count returns total number of active students
func (r *StudentRepository) Count() (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM students WHERE is_active = 1"
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

// CountByClass returns number of students in a class
func (r *StudentRepository) CountByClass(classID int) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM students WHERE class_id = ? AND is_active = 1"
	err := r.db.QueryRow(query, classID).Scan(&count)
	return count, err
}

// LinkToParent links a student to a parent
func (r *StudentRepository) LinkToParent(parentID, studentID int) error {
	query := "INSERT OR IGNORE INTO parent_students (parent_id, student_id) VALUES (?, ?)"
	_, err := r.db.Exec(query, parentID, studentID)
	if err != nil {
		return err
	}

	return nil
}

// UnlinkFromParent removes the link between student and parent
func (r *StudentRepository) UnlinkFromParent(parentID, studentID int) error {
	query := "DELETE FROM parent_students WHERE parent_id = ? AND student_id = ?"
	_, err := r.db.Exec(query, parentID, studentID)
	return err
}

// GetStudentParents retrieves all parents linked to a student
func (r *StudentRepository) GetStudentParents(studentID int) ([]*models.User, error) {
	query := `
		SELECT u.id, u.telegram_id, u.telegram_username, u.phone_number,
		       u.language, u.registered_at
		FROM users u
		INNER JOIN parent_students ps ON u.id = ps.parent_id
		WHERE ps.student_id = ?
		ORDER BY u.registered_at
	`
	rows, err := r.db.Query(query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parents []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.TelegramID,
			&user.TelegramUsername,
			&user.PhoneNumber,
			&user.Language,
			&user.RegisteredAt,
		)
		if err != nil {
			return nil, err
		}
		parents = append(parents, user)
	}

	return parents, nil
}

// GetParentStudents retrieves all students linked to a parent
func (r *StudentRepository) GetParentStudents(parentID int) ([]*models.ParentChild, error) {
	query := `
		SELECT id, parent_id, telegram_id, phone_number, student_id,
		       student_first_name, student_last_name, class_id, class_name, linked_at
		FROM v_parent_children
		WHERE parent_id = ?
		ORDER BY class_name, student_last_name, student_first_name
	`
	rows, err := r.db.Query(query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var children []*models.ParentChild
	for rows.Next() {
		child := &models.ParentChild{}
		err := rows.Scan(
			&child.ID,
			&child.ParentID,
			&child.TelegramID,
			&child.PhoneNumber,
			&child.StudentID,
			&child.StudentFirstName,
			&child.StudentLastName,
			&child.ClassID,
			&child.ClassName,
			&child.LinkedAt,
		)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}

	return children, nil
}

// CountParentStudents returns the number of students linked to a parent
func (r *StudentRepository) CountParentStudents(parentID int) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM parent_students WHERE parent_id = ?"
	err := r.db.QueryRow(query, parentID).Scan(&count)
	return count, err
}

// IsStudentLinkedToParent checks if a student is already linked to a parent
func (r *StudentRepository) IsStudentLinkedToParent(parentID, studentID int) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM parent_students WHERE parent_id = ? AND student_id = ?)"
	var exists bool
	err := r.db.QueryRow(query, parentID, studentID).Scan(&exists)
	return exists, err
}

// GetStudentsByIDs retrieves multiple students by their IDs
func (r *StudentRepository) GetStudentsByIDs(studentIDs []int) ([]*models.StudentWithClass, error) {
	if len(studentIDs) == 0 {
		return []*models.StudentWithClass{}, nil
	}

	// Build query with placeholders
	query := fmt.Sprintf(`
		SELECT id, first_name, last_name, class_id, class_name, is_active, created_at
		FROM v_students_with_class
		WHERE id IN (?%s)
		ORDER BY class_name, last_name, first_name
	`, buildPlaceholders(len(studentIDs)-1))

	// Convert studentIDs to interface slice
	args := make([]interface{}, len(studentIDs))
	for i, id := range studentIDs {
		args[i] = id
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []*models.StudentWithClass
	for rows.Next() {
		student := &models.StudentWithClass{}
		err := rows.Scan(
			&student.ID,
			&student.FirstName,
			&student.LastName,
			&student.ClassID,
			&student.ClassName,
			&student.IsActive,
			&student.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		students = append(students, student)
	}

	return students, nil
}

// Helper function to build SQL placeholders for IN clause
func buildPlaceholders(count int) string {
	if count <= 0 {
		return ""
	}
	placeholders := ""
	for i := 0; i < count; i++ {
		placeholders += ", ?"
	}
	return placeholders
}
