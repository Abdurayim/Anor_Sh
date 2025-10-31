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
	PhoneNumber string `json:"phone_number,omitempty"`
	ChildName   string `json:"child_name,omitempty"`
	ChildClass  string `json:"child_class,omitempty"`
	Language    string `json:"language,omitempty"`
	ComplaintText string `json:"complaint_text,omitempty"`
}

// State constants
const (
	StateStart               = "start"
	StateAwaitingLanguage    = "awaiting_language"
	StateAwaitingPhone       = "awaiting_phone"
	StateAwaitingChildName   = "awaiting_child_name"
	StateAwaitingChildClass  = "awaiting_child_class"
	StateRegistered          = "registered"
	StateAwaitingComplaint   = "awaiting_complaint"
	StateConfirmingComplaint = "confirming_complaint"
	StateAwaitingAdminPhone  = "awaiting_admin_phone"
	StateAwaitingClassName   = "awaiting_class_name"
)
