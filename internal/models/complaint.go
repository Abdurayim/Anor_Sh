package models

import "time"

// Complaint represents a parent complaint
type Complaint struct {
	ID             int       `json:"id" db:"id"`
	UserID         int       `json:"user_id" db:"user_id"`
	StudentID      *int      `json:"student_id,omitempty" db:"student_id"`
	ComplaintText  string    `json:"complaint_text" db:"complaint_text"`
	TelegramFileID string    `json:"telegram_file_id" db:"telegram_file_id"`
	Filename       string    `json:"filename" db:"filename"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	Status         string    `json:"status" db:"status"`
}

// ComplaintWithUser is a complaint with parent and child info (from v_complaints_with_user)
type ComplaintWithUser struct {
	Complaint
	UserTelegramID   int64  `json:"user_telegram_id" db:"telegram_id"`
	TelegramUsername string `json:"telegram_username" db:"telegram_username"`
	PhoneNumber      string `json:"phone_number" db:"phone_number"`
	Language         string `json:"language" db:"language"`
	StudentName      string `json:"student_name" db:"student_name"`
	ClassName        string `json:"class_name" db:"class_name"`
	BranchID         int    `json:"branch_id" db:"branch_id"`
}

// CreateComplaintRequest is the request to create a new complaint
type CreateComplaintRequest struct {
	UserID         int    `json:"user_id" validate:"required"`
	StudentID      *int   `json:"student_id"`
	ComplaintText  string `json:"complaint_text" validate:"required,min=10,max=5000"`
	TelegramFileID string `json:"telegram_file_id"`
	Filename       string `json:"filename"`
}

// Status constants. Complaints end as "resolved", proposals as "implemented"
// (matching the CHECK constraints in the schema).
const (
	StatusPending     = "pending"
	StatusReviewed    = "reviewed"
	StatusResolved    = "resolved"
	StatusImplemented = "implemented"
)
