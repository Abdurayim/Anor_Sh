package models

import "time"

// Class represents a school class
type Class struct {
	ID        int       `json:"id" db:"id"`
	ClassName string    `json:"class_name" db:"class_name"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateClassRequest is the request to create a new class
type CreateClassRequest struct {
	ClassName string `json:"class_name"`
}
