package models

import "time"

// Proposal represents a parent proposal
type Proposal struct {
	ID             int       `json:"id" db:"id"`
	UserID         int       `json:"user_id" db:"user_id"`
	StudentID      *int      `json:"student_id,omitempty" db:"student_id"`
	ProposalText   string    `json:"proposal_text" db:"proposal_text"`
	TelegramFileID string    `json:"telegram_file_id" db:"telegram_file_id"`
	Filename       string    `json:"filename" db:"filename"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	Status         string    `json:"status" db:"status"`
}

// ProposalWithUser is a proposal with parent and child info (from v_proposals_with_user)
type ProposalWithUser struct {
	Proposal
	UserTelegramID   int64  `json:"user_telegram_id" db:"telegram_id"`
	TelegramUsername string `json:"telegram_username" db:"telegram_username"`
	PhoneNumber      string `json:"phone_number" db:"phone_number"`
	Language         string `json:"language" db:"language"`
	StudentName      string `json:"student_name" db:"student_name"`
	ClassName        string `json:"class_name" db:"class_name"`
	BranchID         int    `json:"branch_id" db:"branch_id"`
}

// CreateProposalRequest is the request to create a new proposal
type CreateProposalRequest struct {
	UserID         int    `json:"user_id" validate:"required"`
	StudentID      *int   `json:"student_id"`
	ProposalText   string `json:"proposal_text" validate:"required,min=10,max=5000"`
	TelegramFileID string `json:"telegram_file_id"`
	Filename       string `json:"filename"`
}
