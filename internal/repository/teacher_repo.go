package repository

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
)

// TeacherRepository handles teacher data operations
type TeacherRepository struct {
	db *sql.DB
}

// NewTeacherRepository creates a new teacher repository
func NewTeacherRepository(db *sql.DB) *TeacherRepository {
	return &TeacherRepository{db: db}
}

// Create creates a new teacher
func (r *TeacherRepository) Create(firstName, lastName, phoneNumber, language string, addedByAdminID int) (int64, error) {
	query := `
		INSERT INTO teachers (phone_number, first_name, last_name, language, added_by_admin_id)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, phoneNumber, firstName, lastName, language, addedByAdminID)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdateTelegramID updates teacher's telegram ID and marks as registered
func (r *TeacherRepository) UpdateTelegramID(teacherID int, telegramID int64, username string) error {
	query := `
		UPDATE teachers
		SET telegram_id = ?, telegram_username = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.Exec(query, telegramID, username, teacherID)
	return err
}

// GetByID retrieves a teacher by ID
func (r *TeacherRepository) GetByID(id int) (*models.Teacher, error) {
	query := `
		SELECT id, phone_number, telegram_id, first_name, last_name, language,
		       is_active, added_by_admin_id, created_at
		FROM teachers
		WHERE id = ?
	`
	teacher := &models.Teacher{}
	err := r.db.QueryRow(query, id).Scan(
		&teacher.ID,
		&teacher.PhoneNumber,
		&teacher.TelegramID,
		&teacher.FirstName,
		&teacher.LastName,
		&teacher.Language,
		&teacher.IsActive,
		&teacher.AddedByAdminID,
		&teacher.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return teacher, nil
}

// GetByPhoneNumber retrieves a teacher by phone number
func (r *TeacherRepository) GetByPhoneNumber(phoneNumber string) (*models.Teacher, error) {
	query := `
		SELECT id, phone_number, telegram_id, first_name, last_name, language,
		       is_active, added_by_admin_id, created_at
		FROM teachers
		WHERE phone_number = ?
	`
	teacher := &models.Teacher{}
	err := r.db.QueryRow(query, phoneNumber).Scan(
		&teacher.ID,
		&teacher.PhoneNumber,
		&teacher.TelegramID,
		&teacher.FirstName,
		&teacher.LastName,
		&teacher.Language,
		&teacher.IsActive,
		&teacher.AddedByAdminID,
		&teacher.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return teacher, nil
}

// GetByPhone is an alias for GetByPhoneNumber
func (r *TeacherRepository) GetByPhone(phoneNumber string) (*models.Teacher, error) {
	return r.GetByPhoneNumber(phoneNumber)
}

// GetByTelegramID retrieves a teacher by Telegram ID
func (r *TeacherRepository) GetByTelegramID(telegramID int64) (*models.Teacher, error) {
	query := `
		SELECT id, phone_number, telegram_id, first_name, last_name, language,
		       is_active, added_by_admin_id, created_at
		FROM teachers
		WHERE telegram_id = ?
	`
	teacher := &models.Teacher{}
	err := r.db.QueryRow(query, telegramID).Scan(
		&teacher.ID,
		&teacher.PhoneNumber,
		&teacher.TelegramID,
		&teacher.FirstName,
		&teacher.LastName,
		&teacher.Language,
		&teacher.IsActive,
		&teacher.AddedByAdminID,
		&teacher.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return teacher, nil
}

// LinkTelegramID links a Telegram ID to a teacher account
func (r *TeacherRepository) LinkTelegramID(phoneNumber string, telegramID int64, language string) error {
	query := `
		UPDATE teachers
		SET telegram_id = ?, language = ?, updated_at = CURRENT_TIMESTAMP
		WHERE phone_number = ?
	`
	_, err := r.db.Exec(query, telegramID, language, phoneNumber)
	return err
}

// Update updates teacher information
func (r *TeacherRepository) Update(id int, req *models.UpdateTeacherRequest) error {
	query := `
		UPDATE teachers
		SET first_name = COALESCE(?, first_name),
		    last_name = COALESCE(?, last_name),
		    language = COALESCE(?, language),
		    is_active = COALESCE(?, is_active)
		WHERE id = ?
	`
	_, err := r.db.Exec(query, req.FirstName, req.LastName, req.Language, req.IsActive, id)
	return err
}

// Delete deletes a teacher
func (r *TeacherRepository) Delete(id int) error {
	query := "DELETE FROM teachers WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// GetAll retrieves all teachers with pagination
func (r *TeacherRepository) GetAll(limit, offset int) ([]*models.Teacher, error) {
	query := `
		SELECT id, phone_number, telegram_id, first_name, last_name, language,
		       is_active, added_by_admin_id, created_at
		FROM teachers
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teachers []*models.Teacher
	for rows.Next() {
		teacher := &models.Teacher{}
		err := rows.Scan(
			&teacher.ID,
			&teacher.PhoneNumber,
			&teacher.TelegramID,
			&teacher.FirstName,
			&teacher.LastName,
			&teacher.Language,
			&teacher.IsActive,
			&teacher.AddedByAdminID,
			&teacher.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		teachers = append(teachers, teacher)
	}

	return teachers, nil
}

// GetActiveTeachers retrieves all active teachers
func (r *TeacherRepository) GetActiveTeachers() ([]*models.Teacher, error) {
	query := `
		SELECT id, phone_number, telegram_id, first_name, last_name, language,
		       is_active, added_by_admin_id, created_at
		FROM teachers
		WHERE is_active = 1
		ORDER BY last_name, first_name
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teachers []*models.Teacher
	for rows.Next() {
		teacher := &models.Teacher{}
		err := rows.Scan(
			&teacher.ID,
			&teacher.PhoneNumber,
			&teacher.TelegramID,
			&teacher.FirstName,
			&teacher.LastName,
			&teacher.Language,
			&teacher.IsActive,
			&teacher.AddedByAdminID,
			&teacher.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		teachers = append(teachers, teacher)
	}

	return teachers, nil
}

// Count returns total number of teachers
func (r *TeacherRepository) Count() (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM teachers"
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

// AssignToClass assigns a teacher to a class
func (r *TeacherRepository) AssignToClass(teacherID, classID int) error {
	query := "INSERT OR IGNORE INTO teacher_classes (teacher_id, class_id) VALUES (?, ?)"
	_, err := r.db.Exec(query, teacherID, classID)
	return err
}

// RemoveFromClass removes a teacher from a class
func (r *TeacherRepository) RemoveFromClass(teacherID, classID int) error {
	query := "DELETE FROM teacher_classes WHERE teacher_id = ? AND class_id = ?"
	_, err := r.db.Exec(query, teacherID, classID)
	return err
}

// GetTeacherClasses retrieves all classes assigned to a teacher
func (r *TeacherRepository) GetTeacherClasses(teacherID int) ([]*models.Class, error) {
	query := `
		SELECT c.id, c.class_name, c.is_active, c.created_at, c.updated_at
		FROM classes c
		INNER JOIN teacher_classes tc ON c.id = tc.class_id
		WHERE tc.teacher_id = ?
		ORDER BY c.class_name
	`
	rows, err := r.db.Query(query, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []*models.Class
	for rows.Next() {
		class := &models.Class{}
		err := rows.Scan(&class.ID, &class.ClassName, &class.IsActive, &class.CreatedAt, &class.UpdatedAt)
		if err != nil {
			return nil, err
		}
		classes = append(classes, class)
	}

	return classes, nil
}

// GetClassTeachers retrieves all teachers assigned to a class
func (r *TeacherRepository) GetClassTeachers(classID int) ([]*models.Teacher, error) {
	query := `
		SELECT t.id, t.phone_number, t.telegram_id, t.first_name, t.last_name,
		       t.language, t.is_active, t.added_by_admin_id, t.created_at
		FROM teachers t
		INNER JOIN teacher_classes tc ON t.id = tc.teacher_id
		WHERE tc.class_id = ? AND t.is_active = 1
		ORDER BY t.last_name, t.first_name
	`
	rows, err := r.db.Query(query, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teachers []*models.Teacher
	for rows.Next() {
		teacher := &models.Teacher{}
		err := rows.Scan(
			&teacher.ID,
			&teacher.PhoneNumber,
			&teacher.TelegramID,
			&teacher.FirstName,
			&teacher.LastName,
			&teacher.Language,
			&teacher.IsActive,
			&teacher.AddedByAdminID,
			&teacher.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		teachers = append(teachers, teacher)
	}

	return teachers, nil
}

// IsTeacherAssignedToClass checks if a teacher is assigned to a class
func (r *TeacherRepository) IsTeacherAssignedToClass(teacherID, classID int) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM teacher_classes WHERE teacher_id = ? AND class_id = ?)"
	var exists bool
	err := r.db.QueryRow(query, teacherID, classID).Scan(&exists)
	return exists, err
}

// IsTeacher checks if a phone number or telegram ID belongs to a teacher
func (r *TeacherRepository) IsTeacher(phoneNumber string, telegramID int64) (bool, *models.Teacher, error) {
	var teacher *models.Teacher
	var err error

	// Try by phone number first
	if phoneNumber != "" {
		teacher, err = r.GetByPhoneNumber(phoneNumber)
		if err == nil {
			return true, teacher, nil
		}
		if err != sql.ErrNoRows {
			return false, nil, fmt.Errorf("error checking teacher by phone: %w", err)
		}
	}

	// Try by telegram ID
	if telegramID != 0 {
		teacher, err = r.GetByTelegramID(telegramID)
		if err == nil {
			return true, teacher, nil
		}
		if err != sql.ErrNoRows {
			return false, nil, fmt.Errorf("error checking teacher by telegram ID: %w", err)
		}
	}

	return false, nil, nil
}
