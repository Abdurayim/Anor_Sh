package models

import "time"

// TestResult represents a student's test score
type TestResult struct {
	ID        int       `json:"id" db:"id"`
	StudentID int       `json:"student_id" db:"student_id"`
	SubjectName string  `json:"subject_name" db:"subject_name"`
	Score     string    `json:"score" db:"score"`
	TestDate  time.Time `json:"test_date" db:"test_date"`
	TeacherID *int      `json:"teacher_id" db:"teacher_id"`
	AdminID   *int      `json:"admin_id" db:"admin_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TestResultDetailed represents test result with student and class info (from view)
type TestResultDetailed struct {
	ID          int       `json:"id" db:"id"`
	StudentID   int       `json:"student_id" db:"student_id"`
	FirstName   string    `json:"first_name" db:"first_name"`
	LastName    string    `json:"last_name" db:"last_name"`
	ClassID     int       `json:"class_id" db:"class_id"`
	ClassName   string    `json:"class_name" db:"class_name"`
	SubjectName string    `json:"subject_name" db:"subject_name"`
	Score       string    `json:"score" db:"score"`
	TestDate    time.Time `json:"test_date" db:"test_date"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateTestResultRequest is the request to create a test result
type CreateTestResultRequest struct {
	StudentID   int    `json:"student_id" validate:"required"`
	SubjectName string `json:"subject_name" validate:"required,min=2,max=100"`
	Score       string `json:"score" validate:"required,min=1,max=50"`
	TestDate    string `json:"test_date" validate:"required"` // Format: YYYY-MM-DD
	TeacherID   *int   `json:"teacher_id,omitempty"`
	AdminID     *int   `json:"admin_id,omitempty"`
}

// UpdateTestResultRequest is the request to update a test result
type UpdateTestResultRequest struct {
	SubjectName string `json:"subject_name,omitempty" validate:"omitempty,min=2,max=100"`
	Score       string `json:"score,omitempty" validate:"omitempty,min=1,max=50"`
	TestDate    string `json:"test_date,omitempty"` // Format: YYYY-MM-DD
}
