package models

import "time"

// Attendance represents a student's attendance record
type Attendance struct {
	ID                int       `json:"id" db:"id"`
	StudentID         int       `json:"student_id" db:"student_id"`
	Date              time.Time `json:"date" db:"date"`
	Status            string    `json:"status" db:"status"` // 'present' or 'absent'
	MarkedByTeacherID *int      `json:"marked_by_teacher_id" db:"marked_by_teacher_id"`
	MarkedByAdminID   *int      `json:"marked_by_admin_id" db:"marked_by_admin_id"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// AttendanceDetailed represents attendance with student and class info (from view)
type AttendanceDetailed struct {
	ID        int       `json:"id" db:"id"`
	StudentID int       `json:"student_id" db:"student_id"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	ClassID   int       `json:"class_id" db:"class_id"`
	ClassName string    `json:"class_name" db:"class_name"`
	Date      time.Time `json:"date" db:"date"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CreateAttendanceRequest is the request to mark attendance
type CreateAttendanceRequest struct {
	StudentID         int    `json:"student_id" validate:"required"`
	Date              string `json:"date" validate:"required"` // Format: YYYY-MM-DD
	Status            string `json:"status" validate:"required,oneof=present absent"`
	MarkedByTeacherID *int   `json:"marked_by_teacher_id,omitempty"`
	MarkedByAdminID   *int   `json:"marked_by_admin_id,omitempty"`
}

// UpdateAttendanceRequest is the request to update attendance
type UpdateAttendanceRequest struct {
	Status string `json:"status" validate:"required,oneof=present absent"`
}

// BulkAttendanceRequest is the request to mark attendance for multiple students
type BulkAttendanceRequest struct {
	Date              string `json:"date" validate:"required"` // Format: YYYY-MM-DD
	ClassID           int    `json:"class_id" validate:"required"`
	AbsentStudentIDs  []int  `json:"absent_student_ids"`  // Students marked absent
	PresentStudentIDs []int  `json:"present_student_ids"` // Students explicitly marked present
	MarkedByTeacherID *int   `json:"marked_by_teacher_id,omitempty"`
	MarkedByAdminID   *int   `json:"marked_by_admin_id,omitempty"`
}
