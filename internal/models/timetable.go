package models

import "time"

// Timetable represents a class timetable file
type Timetable struct {
	ID                int       `json:"id" db:"id"`
	ClassID           int       `json:"class_id" db:"class_id"`
	TelegramFileID    string    `json:"telegram_file_id" db:"telegram_file_id"`
	Filename          string    `json:"filename" db:"filename"`
	FileType          string    `json:"file_type" db:"file_type"`         // image, document
	MimeType          string    `json:"mime_type" db:"mime_type"`         // image/jpeg, application/pdf, etc.
	UploadedByAdminID *int      `json:"uploaded_by_admin_id" db:"uploaded_by_admin_id"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// TimetableWithClass represents a timetable with class information
type TimetableWithClass struct {
	ID                int       `json:"id" db:"id"`
	ClassID           int       `json:"class_id" db:"class_id"`
	ClassName         string    `json:"class_name" db:"class_name"`
	TelegramFileID    string    `json:"telegram_file_id" db:"telegram_file_id"`
	Filename          string    `json:"filename" db:"filename"`
	FileType          string    `json:"file_type" db:"file_type"`
	MimeType          string    `json:"mime_type" db:"mime_type"`
	UploadedByAdminID *int      `json:"uploaded_by_admin_id" db:"uploaded_by_admin_id"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// CreateTimetableRequest is the request to create a new timetable
type CreateTimetableRequest struct {
	ClassID           int    `json:"class_id" validate:"required"`
	TelegramFileID    string `json:"telegram_file_id" validate:"required"`
	Filename          string `json:"filename" validate:"required"`
	FileType          string `json:"file_type" validate:"required"`
	MimeType          string `json:"mime_type"`
	UploadedByAdminID *int   `json:"uploaded_by_admin_id"`
}
