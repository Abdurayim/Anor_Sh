package models

import "time"

// User represents a registered parent user
type User struct {
	ID                      int       `json:"id" db:"id"`
	TelegramID              int64     `json:"telegram_id" db:"telegram_id"`
	TelegramUsername        string    `json:"telegram_username" db:"telegram_username"`
	PhoneNumber             string    `json:"phone_number" db:"phone_number"`
	Language                string    `json:"language" db:"language"`
	CurrentSelectedStudentID *int      `json:"current_selected_student_id" db:"current_selected_student_id"`
	RegisteredAt            time.Time `json:"registered_at" db:"registered_at"`
}

// CreateUserRequest is the request to create a new user
type CreateUserRequest struct {
	TelegramID       int64  `json:"telegram_id" validate:"required"`
	TelegramUsername string `json:"telegram_username"`
	PhoneNumber      string `json:"phone_number" validate:"required"`
	Language         string `json:"language" validate:"required,oneof=uz ru"`
}

// UpdateUserRequest is the request to update user data
type UpdateUserRequest struct {
	Language                string `json:"language,omitempty" validate:"omitempty,oneof=uz ru"`
	CurrentSelectedStudentID *int   `json:"current_selected_student_id,omitempty"`
}

// ParentStudent represents the junction table linking parents to students
type ParentStudent struct {
	ID        int       `json:"id" db:"id"`
	ParentID  int       `json:"parent_id" db:"parent_id"`
	StudentID int       `json:"student_id" db:"student_id"`
	LinkedAt  time.Time `json:"linked_at" db:"linked_at"`
}

// ParentChild represents a parent-student relationship with full details (from view)
type ParentChild struct {
	ID                int       `json:"id" db:"id"`
	ParentID          int       `json:"parent_id" db:"parent_id"`
	TelegramID        int64     `json:"telegram_id" db:"telegram_id"`
	PhoneNumber       string    `json:"phone_number" db:"phone_number"`
	StudentID         int       `json:"student_id" db:"student_id"`
	StudentFirstName  string    `json:"student_first_name" db:"student_first_name"`
	StudentLastName   string    `json:"student_last_name" db:"student_last_name"`
	ClassID           int       `json:"class_id" db:"class_id"`
	ClassName         string    `json:"class_name" db:"class_name"`
	LinkedAt          time.Time `json:"linked_at" db:"linked_at"`
}
