package models

import "time"

// Announcement represents a school announcement
type Announcement struct {
	ID                 int       `json:"id" db:"id"`
	Title              *string   `json:"title" db:"title"`
	Content            string    `json:"content" db:"content"`
	TelegramFileID     *string   `json:"telegram_file_id" db:"telegram_file_id"` // optional image
	Filename           *string   `json:"filename" db:"filename"`
	FileType           *string   `json:"file_type" db:"file_type"` // image, document
	PostedByAdminID    *int      `json:"posted_by_admin_id" db:"posted_by_admin_id"`
	PostedByTeacherID  *int      `json:"posted_by_teacher_id" db:"posted_by_teacher_id"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	IsActive           bool      `json:"is_active" db:"is_active"`
}

// AnnouncementClass represents the junction table linking announcements to classes
type AnnouncementClass struct {
	ID             int `json:"id" db:"id"`
	AnnouncementID int `json:"announcement_id" db:"announcement_id"`
	ClassID        int `json:"class_id" db:"class_id"`
}

// CreateAnnouncementRequest is the request to create a new announcement
type CreateAnnouncementRequest struct {
	Title             *string `json:"title"`
	Content           string  `json:"content" validate:"required,min=10,max=10000"`
	TelegramFileID    *string `json:"telegram_file_id"`
	Filename          *string `json:"filename"`
	FileType          *string `json:"file_type"`
	PostedByAdminID   *int    `json:"posted_by_admin_id"`
	PostedByTeacherID *int    `json:"posted_by_teacher_id"`
	ClassIDs          []int   `json:"class_ids" validate:"required,min=1"` // Target classes
}

// UpdateAnnouncementRequest is the request to update an announcement
type UpdateAnnouncementRequest struct {
	Title    *string `json:"title"`
	Content  string  `json:"content,omitempty" validate:"omitempty,min=10,max=10000"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// AnnouncementWithClasses represents an announcement with its target classes
type AnnouncementWithClasses struct {
	Announcement
	ClassIDs   []int    `json:"class_ids"`
	ClassNames []string `json:"class_names"`
}
