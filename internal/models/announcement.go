package models

import "time"

// Announcement represents a school announcement
type Announcement struct {
	ID              int       `json:"id" db:"id"`
	Title           *string   `json:"title" db:"title"`
	Content         string    `json:"content" db:"content"`
	TelegramFileID  *string   `json:"telegram_file_id" db:"telegram_file_id"` // optional image
	Filename        *string   `json:"filename" db:"filename"`
	FileType        *string   `json:"file_type" db:"file_type"` // image
	PostedByAdminID *int      `json:"posted_by_admin_id" db:"posted_by_admin_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	IsActive        bool      `json:"is_active" db:"is_active"`
}

// CreateAnnouncementRequest is the request to create a new announcement
type CreateAnnouncementRequest struct {
	Title           *string `json:"title"`
	Content         string  `json:"content" validate:"required,min=10,max=10000"`
	TelegramFileID  *string `json:"telegram_file_id"`
	Filename        *string `json:"filename"`
	FileType        *string `json:"file_type"`
	PostedByAdminID *int    `json:"posted_by_admin_id"`
}
