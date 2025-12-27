package repository

import (
	"database/sql"
	"time"

	"parent-bot/internal/models"
)

// AttendanceRepository handles attendance data operations
type AttendanceRepository struct {
	db *sql.DB
}

// NewAttendanceRepository creates a new attendance repository
func NewAttendanceRepository(db *sql.DB) *AttendanceRepository {
	return &AttendanceRepository{db: db}
}

// Create creates or updates an attendance record
func (r *AttendanceRepository) Create(req *models.CreateAttendanceRequest) (int64, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return 0, err
	}

	// Use INSERT OR REPLACE to handle duplicate date+student
	query := `
		INSERT INTO attendance (student_id, date, status, marked_by_teacher_id, marked_by_admin_id)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(student_id, date)
		DO UPDATE SET
			status = excluded.status,
			marked_by_teacher_id = excluded.marked_by_teacher_id,
			marked_by_admin_id = excluded.marked_by_admin_id,
			updated_at = CURRENT_TIMESTAMP
	`
	result, err := r.db.Exec(query,
		req.StudentID,
		date,
		req.Status,
		req.MarkedByTeacherID,
		req.MarkedByAdminID,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// BulkCreate creates or updates multiple attendance records
func (r *AttendanceRepository) BulkCreate(req *models.BulkAttendanceRequest) error {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return err
	}

	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert/update absent students
	for _, studentID := range req.AbsentStudentIDs {
		query := `
			INSERT INTO attendance (student_id, date, status, marked_by_teacher_id, marked_by_admin_id)
			VALUES (?, ?, 'absent', ?, ?)
			ON CONFLICT(student_id, date)
			DO UPDATE SET
				status = 'absent',
				marked_by_teacher_id = excluded.marked_by_teacher_id,
				marked_by_admin_id = excluded.marked_by_admin_id,
				updated_at = CURRENT_TIMESTAMP
		`
		_, err = tx.Exec(query, studentID, date, req.MarkedByTeacherID, req.MarkedByAdminID)
		if err != nil {
			return err
		}
	}

	// Insert/update explicitly present students
	for _, studentID := range req.PresentStudentIDs {
		query := `
			INSERT INTO attendance (student_id, date, status, marked_by_teacher_id, marked_by_admin_id)
			VALUES (?, ?, 'present', ?, ?)
			ON CONFLICT(student_id, date)
			DO UPDATE SET
				status = 'present',
				marked_by_teacher_id = excluded.marked_by_teacher_id,
				marked_by_admin_id = excluded.marked_by_admin_id,
				updated_at = CURRENT_TIMESTAMP
		`
		_, err = tx.Exec(query, studentID, date, req.MarkedByTeacherID, req.MarkedByAdminID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// MarkAllPresentExcept marks all students in a class as present except the specified ones
func (r *AttendanceRepository) MarkAllPresentExcept(classID int, date string, absentStudentIDs []int, markedByTeacherID, markedByAdminID *int) error {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return err
	}

	// Get all students in the class
	studentsQuery := `
		SELECT id FROM students WHERE class_id = ? AND is_active = 1
	`
	rows, err := r.db.Query(studentsQuery, classID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var allStudentIDs []int
	for rows.Next() {
		var studentID int
		if err := rows.Scan(&studentID); err != nil {
			return err
		}
		allStudentIDs = append(allStudentIDs, studentID)
	}

	// Create a map of absent students for quick lookup
	absentMap := make(map[int]bool)
	for _, id := range absentStudentIDs {
		absentMap[id] = true
	}

	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO attendance (student_id, date, status, marked_by_teacher_id, marked_by_admin_id)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(student_id, date)
		DO UPDATE SET
			status = excluded.status,
			marked_by_teacher_id = excluded.marked_by_teacher_id,
			marked_by_admin_id = excluded.marked_by_admin_id,
			updated_at = CURRENT_TIMESTAMP
	`

	// Mark attendance for all students
	for _, studentID := range allStudentIDs {
		status := "present"
		if absentMap[studentID] {
			status = "absent"
		}
		_, err = tx.Exec(query, studentID, parsedDate, status, markedByTeacherID, markedByAdminID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetByID retrieves an attendance record by ID
func (r *AttendanceRepository) GetByID(id int) (*models.Attendance, error) {
	query := `
		SELECT id, student_id, date, status, marked_by_teacher_id, marked_by_admin_id,
		       created_at, updated_at
		FROM attendance
		WHERE id = ?
	`
	record := &models.Attendance{}
	err := r.db.QueryRow(query, id).Scan(
		&record.ID,
		&record.StudentID,
		&record.Date,
		&record.Status,
		&record.MarkedByTeacherID,
		&record.MarkedByAdminID,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// GetByStudentID retrieves attendance records for a student
func (r *AttendanceRepository) GetByStudentID(studentID int, limit, offset int) ([]*models.AttendanceDetailed, error) {
	query := `
		SELECT id, student_id, first_name, last_name, class_id, class_name,
		       date, status, created_at
		FROM v_attendance_detailed
		WHERE student_id = ?
		ORDER BY date DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(query, studentID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.AttendanceDetailed
	for rows.Next() {
		record := &models.AttendanceDetailed{}
		err := rows.Scan(
			&record.ID,
			&record.StudentID,
			&record.FirstName,
			&record.LastName,
			&record.ClassID,
			&record.ClassName,
			&record.Date,
			&record.Status,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// GetByStudentIDAndDateRange retrieves attendance for a student within a date range
func (r *AttendanceRepository) GetByStudentIDAndDateRange(studentID int, startDate, endDate string) ([]*models.AttendanceDetailed, error) {
	query := `
		SELECT id, student_id, first_name, last_name, class_id, class_name,
		       date, status, created_at
		FROM v_attendance_detailed
		WHERE student_id = ? AND date(date) BETWEEN date(?) AND date(?)
		ORDER BY date DESC
	`
	rows, err := r.db.Query(query, studentID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.AttendanceDetailed
	for rows.Next() {
		record := &models.AttendanceDetailed{}
		err := rows.Scan(
			&record.ID,
			&record.StudentID,
			&record.FirstName,
			&record.LastName,
			&record.ClassID,
			&record.ClassName,
			&record.Date,
			&record.Status,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// GetByClassIDAndDate retrieves attendance for all students in a class on a specific date
func (r *AttendanceRepository) GetByClassIDAndDate(classID int, date string) ([]*models.AttendanceDetailed, error) {
	query := `
		SELECT id, student_id, first_name, last_name, class_id, class_name,
		       date, status, created_at
		FROM v_attendance_detailed
		WHERE class_id = ? AND date(date) = date(?)
		ORDER BY last_name, first_name
	`
	rows, err := r.db.Query(query, classID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.AttendanceDetailed
	for rows.Next() {
		record := &models.AttendanceDetailed{}
		err := rows.Scan(
			&record.ID,
			&record.StudentID,
			&record.FirstName,
			&record.LastName,
			&record.ClassID,
			&record.ClassName,
			&record.Date,
			&record.Status,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// GetTodayAttendanceByClass retrieves today's attendance for a specific class
func (r *AttendanceRepository) GetTodayAttendanceByClass(classID int) ([]*models.AttendanceDetailed, error) {
	// Use Uzbekistan timezone (Asia/Tashkent UTC+5) to match attendance taking
	location, _ := time.LoadLocation("Asia/Tashkent")
	today := time.Now().In(location).Format("2006-01-02")
	return r.GetByClassIDAndDate(classID, today)
}

// GetTodayAttendanceAllClasses retrieves today's attendance for all classes
func (r *AttendanceRepository) GetTodayAttendanceAllClasses() ([]*models.AttendanceDetailed, error) {
	// Use Uzbekistan timezone (Asia/Tashkent UTC+5) to match attendance taking
	location, _ := time.LoadLocation("Asia/Tashkent")
	today := time.Now().In(location).Format("2006-01-02")
	query := `
		SELECT id, student_id, first_name, last_name, class_id, class_name,
		       date, status, created_at
		FROM v_attendance_detailed
		WHERE date(date) = date(?)
		ORDER BY class_name, last_name, first_name
	`
	rows, err := r.db.Query(query, today)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*models.AttendanceDetailed
	for rows.Next() {
		record := &models.AttendanceDetailed{}
		err := rows.Scan(
			&record.ID,
			&record.StudentID,
			&record.FirstName,
			&record.LastName,
			&record.ClassID,
			&record.ClassName,
			&record.Date,
			&record.Status,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// Update updates an attendance record
func (r *AttendanceRepository) Update(id int, req *models.UpdateAttendanceRequest) error {
	query := `
		UPDATE attendance
		SET status = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.Exec(query, req.Status, id)
	return err
}

// Delete deletes an attendance record (restricted)
func (r *AttendanceRepository) Delete(id int) error {
	query := "DELETE FROM attendance WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// Count returns total number of attendance records
func (r *AttendanceRepository) Count() (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM attendance"
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

// IsAttendanceTaken checks if attendance has been taken for a class on a specific date
func (r *AttendanceRepository) IsAttendanceTaken(classID int, date string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM attendance a
			INNER JOIN students s ON a.student_id = s.id
			WHERE s.class_id = ? AND date(a.date) = date(?)
		)
	`
	var exists bool
	err := r.db.QueryRow(query, classID, date).Scan(&exists)
	return exists, err
}

// GetLast30DaysByStudent retrieves last 30 days attendance for a student
func (r *AttendanceRepository) GetLast30DaysByStudent(studentID int) ([]*models.AttendanceDetailed, error) {
	// Use Uzbekistan timezone (Asia/Tashkent UTC+5) to match attendance taking
	location, _ := time.LoadLocation("Asia/Tashkent")
	now := time.Now().In(location)
	thirtyDaysAgo := now.AddDate(0, 0, -30).Format("2006-01-02")
	today := now.Format("2006-01-02")
	return r.GetByStudentIDAndDateRange(studentID, thirtyDaysAgo, today)
}
