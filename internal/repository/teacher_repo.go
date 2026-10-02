package repository

import (
	"database/sql"

	"parent-bot/internal/models"
)

// TeacherRepository handles teacher data operations. Each teacher belongs to one branch.
type TeacherRepository struct {
	db *sql.DB
}

// NewTeacherRepository creates a new teacher repository
func NewTeacherRepository(db *sql.DB) *TeacherRepository {
	return &TeacherRepository{db: db}
}

const teacherColumns = `t.id, t.phone_number, t.telegram_id, t.first_name, t.last_name, t.language,
	t.is_active, COALESCE(t.branch_id, 0), t.added_by_admin_id, t.created_at`

func scanTeacher(row interface{ Scan(...any) error }) (*models.Teacher, error) {
	teacher := &models.Teacher{}
	err := row.Scan(
		&teacher.ID,
		&teacher.PhoneNumber,
		&teacher.TelegramID,
		&teacher.FirstName,
		&teacher.LastName,
		&teacher.Language,
		&teacher.IsActive,
		&teacher.BranchID,
		&teacher.AddedByAdminID,
		&teacher.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return teacher, nil
}

func (r *TeacherRepository) getMany(query string, args ...any) ([]*models.Teacher, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teachers []*models.Teacher
	for rows.Next() {
		teacher, err := scanTeacher(rows)
		if err != nil {
			return nil, err
		}
		teachers = append(teachers, teacher)
	}
	return teachers, rows.Err()
}

// Create creates a new teacher in a branch
func (r *TeacherRepository) Create(firstName, lastName, phoneNumber, language string, addedByAdminID, branchID int) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO teachers (phone_number, first_name, last_name, language, added_by_admin_id, branch_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`, phoneNumber, firstName, lastName, language, nullableID(addedByAdminID), branchID)
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

// GetByID retrieves a teacher by ID (returns sql.ErrNoRows if missing)
func (r *TeacherRepository) GetByID(id int) (*models.Teacher, error) {
	return scanTeacher(r.db.QueryRow("SELECT "+teacherColumns+" FROM teachers t WHERE t.id = ?", id))
}

// GetByPhoneNumber retrieves a teacher by phone number
func (r *TeacherRepository) GetByPhoneNumber(phoneNumber string) (*models.Teacher, error) {
	teacher, err := scanTeacher(r.db.QueryRow("SELECT "+teacherColumns+" FROM teachers t WHERE t.phone_number = ?", phoneNumber))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return teacher, err
}

// GetByPhone is an alias for GetByPhoneNumber
func (r *TeacherRepository) GetByPhone(phoneNumber string) (*models.Teacher, error) {
	return r.GetByPhoneNumber(phoneNumber)
}

// GetByTelegramID retrieves an active teacher by Telegram ID
func (r *TeacherRepository) GetByTelegramID(telegramID int64) (*models.Teacher, error) {
	teacher, err := scanTeacher(r.db.QueryRow(
		"SELECT "+teacherColumns+" FROM teachers t WHERE t.telegram_id = ? AND t.is_active = 1", telegramID))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return teacher, err
}

// LinkTelegramID links a Telegram ID to a teacher account
func (r *TeacherRepository) LinkTelegramID(phoneNumber string, telegramID int64, language string) error {
	query := `
		UPDATE teachers
		SET telegram_id = ?, language = COALESCE(NULLIF(?, ''), language), updated_at = CURRENT_TIMESTAMP
		WHERE phone_number = ?
	`
	_, err := r.db.Exec(query, telegramID, language, phoneNumber)
	return err
}

// Update updates teacher information; empty fields are left unchanged
func (r *TeacherRepository) Update(id int, req *models.UpdateTeacherRequest) error {
	query := `
		UPDATE teachers
		SET first_name = COALESCE(NULLIF(?, ''), first_name),
		    last_name = COALESCE(NULLIF(?, ''), last_name),
		    language = COALESCE(NULLIF(?, ''), language),
		    is_active = COALESCE(?, is_active),
		    updated_at = CURRENT_TIMESTAMP
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

// GetAll retrieves teachers of a branch with pagination (branchID 0 = all branches)
func (r *TeacherRepository) GetAll(branchID, limit, offset int) ([]*models.Teacher, error) {
	return r.getMany(`
		SELECT `+teacherColumns+`
		FROM teachers t
		WHERE (? = 0 OR t.branch_id = ?)
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?
	`, branchID, branchID, limit, offset)
}

// GetActiveTeachers retrieves active teachers of a branch (branchID 0 = all branches)
func (r *TeacherRepository) GetActiveTeachers(branchID int) ([]*models.Teacher, error) {
	return r.getMany(`
		SELECT `+teacherColumns+`
		FROM teachers t
		WHERE t.is_active = 1 AND (? = 0 OR t.branch_id = ?)
		ORDER BY t.last_name, t.first_name
	`, branchID, branchID)
}

// Count returns number of teachers in a branch (branchID 0 = all branches)
func (r *TeacherRepository) Count(branchID int) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM teachers WHERE (? = 0 OR branch_id = ?)", branchID, branchID).Scan(&count)
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
		SELECT c.id, c.branch_id, c.class_name, c.is_active, c.created_at, c.updated_at
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
		err := rows.Scan(&class.ID, &class.BranchID, &class.ClassName, &class.IsActive, &class.CreatedAt, &class.UpdatedAt)
		if err != nil {
			return nil, err
		}
		classes = append(classes, class)
	}

	return classes, rows.Err()
}

// GetClassTeachers retrieves all active teachers assigned to a class
func (r *TeacherRepository) GetClassTeachers(classID int) ([]*models.Teacher, error) {
	return r.getMany(`
		SELECT `+teacherColumns+`
		FROM teachers t
		INNER JOIN teacher_classes tc ON t.id = tc.teacher_id
		WHERE tc.class_id = ? AND t.is_active = 1
		ORDER BY t.last_name, t.first_name
	`, classID)
}

// IsTeacherAssignedToClass checks if a teacher is assigned to a class
func (r *TeacherRepository) IsTeacherAssignedToClass(teacherID, classID int) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM teacher_classes WHERE teacher_id = ? AND class_id = ?)"
	var exists bool
	err := r.db.QueryRow(query, teacherID, classID).Scan(&exists)
	return exists, err
}
