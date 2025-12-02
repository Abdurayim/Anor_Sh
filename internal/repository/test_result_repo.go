package repository

import (
	"database/sql"
	"time"

	"parent-bot/internal/models"
)

// TestResultRepository handles test result data operations
type TestResultRepository struct {
	db *sql.DB
}

// NewTestResultRepository creates a new test result repository
func NewTestResultRepository(db *sql.DB) *TestResultRepository {
	return &TestResultRepository{db: db}
}

// Create creates a new test result
func (r *TestResultRepository) Create(req *models.CreateTestResultRequest) (int64, error) {
	testDate, err := time.Parse("2006-01-02", req.TestDate)
	if err != nil {
		return 0, err
	}

	query := `
		INSERT INTO test_results (student_id, subject_name, score, test_date, teacher_id, admin_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query,
		req.StudentID,
		req.SubjectName,
		req.Score,
		testDate,
		req.TeacherID,
		req.AdminID,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// GetByID retrieves a test result by ID
func (r *TestResultRepository) GetByID(id int) (*models.TestResult, error) {
	query := `
		SELECT id, student_id, subject_name, score, test_date,
		       teacher_id, admin_id, created_at, updated_at
		FROM test_results
		WHERE id = ?
	`
	result := &models.TestResult{}
	err := r.db.QueryRow(query, id).Scan(
		&result.ID,
		&result.StudentID,
		&result.SubjectName,
		&result.Score,
		&result.TestDate,
		&result.TeacherID,
		&result.AdminID,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetByStudentID retrieves all test results for a student
func (r *TestResultRepository) GetByStudentID(studentID int, limit, offset int) ([]*models.TestResultDetailed, error) {
	query := `
		SELECT id, student_id, first_name, last_name, class_id, class_name,
		       subject_name, score, test_date, created_at, updated_at
		FROM v_test_results_detailed
		WHERE student_id = ?
		ORDER BY test_date DESC, created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(query, studentID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.TestResultDetailed
	for rows.Next() {
		result := &models.TestResultDetailed{}
		err := rows.Scan(
			&result.ID,
			&result.StudentID,
			&result.FirstName,
			&result.LastName,
			&result.ClassID,
			&result.ClassName,
			&result.SubjectName,
			&result.Score,
			&result.TestDate,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// GetByClassID retrieves all test results for a class
func (r *TestResultRepository) GetByClassID(classID int, limit, offset int) ([]*models.TestResultDetailed, error) {
	query := `
		SELECT id, student_id, first_name, last_name, class_id, class_name,
		       subject_name, score, test_date, created_at, updated_at
		FROM v_test_results_detailed
		WHERE class_id = ?
		ORDER BY class_name, last_name, first_name, test_date DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(query, classID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.TestResultDetailed
	for rows.Next() {
		result := &models.TestResultDetailed{}
		err := rows.Scan(
			&result.ID,
			&result.StudentID,
			&result.FirstName,
			&result.LastName,
			&result.ClassID,
			&result.ClassName,
			&result.SubjectName,
			&result.Score,
			&result.TestDate,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// GetAllByClassID retrieves all test results for a class without pagination (for export)
func (r *TestResultRepository) GetAllByClassID(classID int) ([]*models.TestResultDetailed, error) {
	query := `
		SELECT id, student_id, first_name, last_name, class_id, class_name,
		       subject_name, score, test_date, created_at, updated_at
		FROM v_test_results_detailed
		WHERE class_id = ?
		ORDER BY last_name, first_name, subject_name, test_date DESC
	`
	rows, err := r.db.Query(query, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.TestResultDetailed
	for rows.Next() {
		result := &models.TestResultDetailed{}
		err := rows.Scan(
			&result.ID,
			&result.StudentID,
			&result.FirstName,
			&result.LastName,
			&result.ClassID,
			&result.ClassName,
			&result.SubjectName,
			&result.Score,
			&result.TestDate,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// Update updates a test result
func (r *TestResultRepository) Update(id int, req *models.UpdateTestResultRequest) error {
	var testDate *time.Time
	if req.TestDate != "" {
		t, err := time.Parse("2006-01-02", req.TestDate)
		if err != nil {
			return err
		}
		testDate = &t
	}

	query := `
		UPDATE test_results
		SET subject_name = COALESCE(?, subject_name),
		    score = COALESCE(?, score),
		    test_date = COALESCE(?, test_date),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.Exec(query, req.SubjectName, req.Score, testDate, id)
	return err
}

// Delete deletes a test result (only admins can delete)
func (r *TestResultRepository) Delete(id int) error {
	query := "DELETE FROM test_results WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// Count returns total number of test results
func (r *TestResultRepository) Count() (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM test_results"
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

// CountByStudent returns number of test results for a student
func (r *TestResultRepository) CountByStudent(studentID int) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM test_results WHERE student_id = ?"
	err := r.db.QueryRow(query, studentID).Scan(&count)
	return count, err
}

// GetLatestByStudent retrieves the most recent test results for a student
func (r *TestResultRepository) GetLatestByStudent(studentID int, limit int) ([]*models.TestResultDetailed, error) {
	query := `
		SELECT id, student_id, first_name, last_name, class_id, class_name,
		       subject_name, score, test_date, created_at, updated_at
		FROM v_test_results_detailed
		WHERE student_id = ?
		ORDER BY test_date DESC, created_at DESC
		LIMIT ?
	`
	rows, err := r.db.Query(query, studentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.TestResultDetailed
	for rows.Next() {
		result := &models.TestResultDetailed{}
		err := rows.Scan(
			&result.ID,
			&result.StudentID,
			&result.FirstName,
			&result.LastName,
			&result.ClassID,
			&result.ClassName,
			&result.SubjectName,
			&result.Score,
			&result.TestDate,
			&result.CreatedAt,
			&result.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}
