package models

import "time"

// Complaint represents a user complaint
type Complaint struct {
	ID             int       `json:"id" db:"id"`
	UserID         int       `json:"user_id" db:"user_id"`
	ComplaintText  string    `json:"complaint_text" db:"complaint_text"`
	TelegramFileID string    `json:"telegram_file_id" db:"telegram_file_id"`
	Filename       string    `json:"filename" db:"filename"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	Status         string    `json:"status" db:"status"`
}

// ComplaintWithUser represents a complaint with user information (from view)
type ComplaintWithUser struct {
	ID               int       `json:"id" db:"id"`
	UserID           int       `json:"user_id" db:"user_id"`
	ComplaintText    string    `json:"complaint_text" db:"complaint_text"`
	TelegramFileID   string    `json:"telegram_file_id" db:"telegram_file_id"`
	Filename         string    `json:"filename" db:"filename"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	Status           string    `json:"status" db:"status"`
	UserTelegramID   int64     `json:"user_telegram_id" db:"user_telegram_id"`
	TelegramUsername string    `json:"telegram_username" db:"telegram_username"`
	PhoneNumber      string    `json:"phone_number" db:"phone_number"`
	ChildName        string    `json:"child_name" db:"child_name"`
	ChildClass       string    `json:"child_class" db:"child_class"`
}

// CreateComplaintRequest is the request to create a new complaint
type CreateComplaintRequest struct {
	UserID         int    `json:"user_id" validate:"required"`
	ComplaintText  string `json:"complaint_text" validate:"required,min=10,max=5000"`
	TelegramFileID string `json:"telegram_file_id" validate:"required"`
	Filename       string `json:"filename" validate:"required"`
}

// ComplaintStatus constants
const (
	StatusPending  = "pending"
	StatusReviewed = "reviewed"
	StatusArchived = "archived"
)
