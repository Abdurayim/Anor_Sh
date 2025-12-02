package models

import "time"

// Student represents a student entity
type Student struct {
	ID               int        `json:"id" db:"id"`
	FirstName        string     `json:"first_name" db:"first_name"`
	LastName         string     `json:"last_name" db:"last_name"`
	ClassID          int        `json:"class_id" db:"class_id"`
	IsActive         bool       `json:"is_active" db:"is_active"`
	AddedByAdminID   *int       `json:"added_by_admin_id" db:"added_by_admin_id"`
	AddedByTeacherID *int       `json:"added_by_teacher_id" db:"added_by_teacher_id"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// StudentWithClass represents a student with class information (from view)
type StudentWithClass struct {
	ID        int       `json:"id" db:"id"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	ClassID   int       `json:"class_id" db:"class_id"`
	ClassName string    `json:"class_name" db:"class_name"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CreateStudentRequest is the request to create a new student
type CreateStudentRequest struct {
	FirstName        string `json:"first_name" validate:"required,min=2,max=100"`
	LastName         string `json:"last_name" validate:"required,min=2,max=100"`
	ClassID          int    `json:"class_id" validate:"required"`
	AddedByAdminID   *int   `json:"added_by_admin_id,omitempty"`
	AddedByTeacherID *int   `json:"added_by_teacher_id,omitempty"`
}

// UpdateStudentRequest is the request to update student data
type UpdateStudentRequest struct {
	FirstName string `json:"first_name,omitempty" validate:"omitempty,min=2,max=100"`
	LastName  string `json:"last_name,omitempty" validate:"omitempty,min=2,max=100"`
	ClassID   *int   `json:"class_id,omitempty"`
	IsActive  *bool  `json:"is_active,omitempty"`
}
