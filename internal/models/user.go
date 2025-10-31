package models

import "time"

// User represents a registered user
type User struct {
	ID               int       `json:"id" db:"id"`
	TelegramID       int64     `json:"telegram_id" db:"telegram_id"`
	TelegramUsername string    `json:"telegram_username" db:"telegram_username"`
	PhoneNumber      string    `json:"phone_number" db:"phone_number"`
	ChildName        string    `json:"child_name" db:"child_name"`
	ChildClass       string    `json:"child_class" db:"child_class"`
	Language         string    `json:"language" db:"language"`
	RegisteredAt     time.Time `json:"registered_at" db:"registered_at"`
}

// CreateUserRequest is the request to create a new user
type CreateUserRequest struct {
	TelegramID       int64  `json:"telegram_id" validate:"required"`
	TelegramUsername string `json:"telegram_username"`
	PhoneNumber      string `json:"phone_number" validate:"required"`
	ChildName        string `json:"child_name" validate:"required,min=2,max=255"`
	ChildClass       string `json:"child_class" validate:"required"`
	Language         string `json:"language" validate:"required,oneof=uz ru"`
}

// UpdateUserRequest is the request to update user data
type UpdateUserRequest struct {
	ChildName  string `json:"child_name,omitempty" validate:"omitempty,min=2,max=255"`
	ChildClass string `json:"child_class,omitempty"`
	Language   string `json:"language,omitempty" validate:"omitempty,oneof=uz ru"`
}
