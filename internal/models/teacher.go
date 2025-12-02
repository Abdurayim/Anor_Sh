package models

import "time"

// Teacher represents a teacher account
type Teacher struct {
	ID            int        `json:"id" db:"id"`
	PhoneNumber   string     `json:"phone_number" db:"phone_number"`
	TelegramID    *int64     `json:"telegram_id" db:"telegram_id"`
	FirstName     string     `json:"first_name" db:"first_name"`
	LastName      string     `json:"last_name" db:"last_name"`
	Language      string     `json:"language" db:"language"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	AddedByAdminID *int      `json:"added_by_admin_id" db:"added_by_admin_id"`
	RegisteredAt  *time.Time `json:"registered_at" db:"registered_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// CreateTeacherRequest is the request to create a new teacher
type CreateTeacherRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	FirstName   string `json:"first_name" validate:"required,min=2,max=100"`
	LastName    string `json:"last_name" validate:"required,min=2,max=100"`
	Language    string `json:"language" validate:"required,oneof=uz ru"`
	AddedByAdminID int `json:"added_by_admin_id" validate:"required"`
}

// UpdateTeacherRequest is the request to update teacher data
type UpdateTeacherRequest struct {
	FirstName string `json:"first_name,omitempty" validate:"omitempty,min=2,max=100"`
	LastName  string `json:"last_name,omitempty" validate:"omitempty,min=2,max=100"`
	Language  string `json:"language,omitempty" validate:"omitempty,oneof=uz ru"`
	IsActive  *bool  `json:"is_active,omitempty"`
}

// TeacherClass represents the junction table linking teachers to classes
type TeacherClass struct {
	ID         int       `json:"id" db:"id"`
	TeacherID  int       `json:"teacher_id" db:"teacher_id"`
	ClassID    int       `json:"class_id" db:"class_id"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
}
