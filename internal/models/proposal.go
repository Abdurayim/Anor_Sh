package models

import "time"

// Proposal represents a user proposal (similar to complaint)
type Proposal struct {
	ID             int       `json:"id" db:"id"`
	UserID         int       `json:"user_id" db:"user_id"`
	ProposalText   string    `json:"proposal_text" db:"proposal_text"`
	TelegramFileID string    `json:"telegram_file_id" db:"telegram_file_id"`
	Filename       string    `json:"filename" db:"filename"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	Status         string    `json:"status" db:"status"`
}

// ProposalWithUser represents a proposal with user information (from view)
type ProposalWithUser struct {
	ID               int       `json:"id" db:"id"`
	UserID           int       `json:"user_id" db:"user_id"`
	ProposalText     string    `json:"proposal_text" db:"proposal_text"`
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

// CreateProposalRequest is the request to create a new proposal
type CreateProposalRequest struct {
	UserID         int    `json:"user_id" validate:"required"`
	ProposalText   string `json:"proposal_text" validate:"required,min=10,max=5000"`
	TelegramFileID string `json:"telegram_file_id" validate:"required"`
	Filename       string `json:"filename" validate:"required"`
}
