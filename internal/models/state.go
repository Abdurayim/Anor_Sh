package models

import (
	"encoding/json"
	"time"
)

// UserState represents the conversation state of a user
type UserState struct {
	TelegramID int64           `json:"telegram_id" db:"telegram_id"`
	State      string          `json:"state" db:"state"`
	Data       json.RawMessage `json:"data" db:"data"`
	UpdatedAt  time.Time       `json:"updated_at" db:"updated_at"`
}

// StateData is a helper struct for storing state data
type StateData struct {
	PhoneNumber       string `json:"phone_number,omitempty"`
	// ChildName and ChildClass are deprecated - students are now managed separately
	// Kept for backward compatibility with old state data
	ChildName         string `json:"child_name,omitempty"`
	ChildClass        string `json:"child_class,omitempty"`
	Language          string `json:"language,omitempty"`
	ComplaintText     string `json:"complaint_text,omitempty"`
	ProposalText      string `json:"proposal_text,omitempty"`
	ClassID           int    `json:"class_id,omitempty"`
	AnnouncementText  string `json:"announcement_text,omitempty"`
	AnnouncementTitle string `json:"announcement_title,omitempty"`
	AnnouncementID    int    `json:"announcement_id,omitempty"`
}

// State constants
const (
	StateStart               = "start"
	StateAwaitingLanguage    = "awaiting_language"
	StateAwaitingPhone       = "awaiting_phone"
	// DEPRECATED: Child name/class are no longer collected during registration
	StateAwaitingChildName   = "awaiting_child_name"
	StateAwaitingChildClass  = "awaiting_child_class"
	StateRegistered          = "registered"
	StateAwaitingComplaint   = "awaiting_complaint"
	StateConfirmingComplaint = "confirming_complaint"
	StateAwaitingProposal    = "awaiting_proposal"
	StateConfirmingProposal  = "confirming_proposal"
	StateAwaitingAdminPhone  = "awaiting_admin_phone"
	StateAwaitingClassName   = "awaiting_class_name"
	StateSelectClassForTimetable = "select_class_for_timetable"
	StateAwaitingTimetableFile   = "awaiting_timetable_file"
	StateAwaitingAnnouncementTitle = "awaiting_announcement_title"
	StateAwaitingAnnouncementContent = "awaiting_announcement_content"
	StateAwaitingAnnouncementFile    = "awaiting_announcement_file"
	StateAwaitingEditedAnnouncementContent = "awaiting_edited_announcement_content"
)
