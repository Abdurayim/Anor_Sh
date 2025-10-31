package models

import "time"

// Admin represents an admin user
type Admin struct {
	ID          int       `json:"id" db:"id"`
	PhoneNumber string    `json:"phone_number" db:"phone_number"`
	TelegramID  *int64    `json:"telegram_id,omitempty" db:"telegram_id"`
	Name        string    `json:"name" db:"name"`
	AddedAt     time.Time `json:"added_at" db:"added_at"`
}

// IsAdmin checks if a phone number or telegram ID is an admin
type AdminCheck struct {
	PhoneNumber string
	TelegramID  int64
}
