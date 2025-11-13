package repository

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
)

type ProposalRepository struct {
	db *sql.DB
}

func NewProposalRepository(db *sql.DB) *ProposalRepository {
	return &ProposalRepository{db: db}
}

// Create creates a new proposal
func (r *ProposalRepository) Create(req *models.CreateProposalRequest) (*models.Proposal, error) {
	query := `
		INSERT INTO proposals (user_id, proposal_text, telegram_file_id, filename)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, proposal_text, telegram_file_id, filename, created_at, status
	`

	var proposal models.Proposal
	err := r.db.QueryRow(
		query,
		req.UserID,
		req.ProposalText,
		req.TelegramFileID,
		req.Filename,
	).Scan(
		&proposal.ID,
		&proposal.UserID,
		&proposal.ProposalText,
		&proposal.TelegramFileID,
		&proposal.Filename,
		&proposal.CreatedAt,
		&proposal.Status,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create proposal: %w", err)
	}

	return &proposal, nil
}

// GetByID gets proposal by ID
func (r *ProposalRepository) GetByID(id int) (*models.Proposal, error) {
	query := `
		SELECT id, user_id, proposal_text, telegram_file_id, filename, created_at, status
		FROM proposals
		WHERE id = $1
	`

	var proposal models.Proposal
	err := r.db.QueryRow(query, id).Scan(
		&proposal.ID,
		&proposal.UserID,
		&proposal.ProposalText,
		&proposal.TelegramFileID,
		&proposal.Filename,
		&proposal.CreatedAt,
		&proposal.Status,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	return &proposal, nil
}

// GetByUserID gets proposals by user ID (indexed, fast query)
func (r *ProposalRepository) GetByUserID(userID int, limit, offset int) ([]*models.Proposal, error) {
	query := `
		SELECT id, user_id, proposal_text, telegram_file_id, filename, created_at, status
		FROM proposals
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposals: %w", err)
	}
	defer rows.Close()

	var proposals []*models.Proposal
	for rows.Next() {
		var proposal models.Proposal
		err := rows.Scan(
			&proposal.ID,
			&proposal.UserID,
			&proposal.ProposalText,
			&proposal.TelegramFileID,
			&proposal.Filename,
			&proposal.CreatedAt,
			&proposal.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan proposal: %w", err)
		}
		proposals = append(proposals, &proposal)
	}

	return proposals, nil
}

// GetAll gets all proposals with pagination (for admin)
func (r *ProposalRepository) GetAll(limit, offset int) ([]*models.Proposal, error) {
	query := `
		SELECT id, user_id, proposal_text, telegram_file_id, filename, created_at, status
		FROM proposals
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposals: %w", err)
	}
	defer rows.Close()

	var proposals []*models.Proposal
	for rows.Next() {
		var proposal models.Proposal
		err := rows.Scan(
			&proposal.ID,
			&proposal.UserID,
			&proposal.ProposalText,
			&proposal.TelegramFileID,
			&proposal.Filename,
			&proposal.CreatedAt,
			&proposal.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan proposal: %w", err)
		}
		proposals = append(proposals, &proposal)
	}

	return proposals, nil
}

// GetAllWithUser gets all proposals with user info using view (optimized for admin)
func (r *ProposalRepository) GetAllWithUser(limit, offset int) ([]*models.ProposalWithUser, error) {
	query := `
		SELECT id, user_id, proposal_text, telegram_file_id, filename, created_at, status,
		       user_telegram_id, telegram_username, phone_number, child_name, child_class
		FROM v_proposals_with_user
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposals with user: %w", err)
	}
	defer rows.Close()

	var proposals []*models.ProposalWithUser
	for rows.Next() {
		var proposal models.ProposalWithUser
		err := rows.Scan(
			&proposal.ID,
			&proposal.UserID,
			&proposal.ProposalText,
			&proposal.TelegramFileID,
			&proposal.Filename,
			&proposal.CreatedAt,
			&proposal.Status,
			&proposal.UserTelegramID,
			&proposal.TelegramUsername,
			&proposal.PhoneNumber,
			&proposal.ChildName,
			&proposal.ChildClass,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan proposal with user: %w", err)
		}
		proposals = append(proposals, &proposal)
	}

	return proposals, nil
}

// GetByStatus gets proposals by status (indexed, fast query)
func (r *ProposalRepository) GetByStatus(status string, limit, offset int) ([]*models.Proposal, error) {
	query := `
		SELECT id, user_id, proposal_text, telegram_file_id, filename, created_at, status
		FROM proposals
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposals by status: %w", err)
	}
	defer rows.Close()

	var proposals []*models.Proposal
	for rows.Next() {
		var proposal models.Proposal
		err := rows.Scan(
			&proposal.ID,
			&proposal.UserID,
			&proposal.ProposalText,
			&proposal.TelegramFileID,
			&proposal.Filename,
			&proposal.CreatedAt,
			&proposal.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan proposal: %w", err)
		}
		proposals = append(proposals, &proposal)
	}

	return proposals, nil
}

// UpdateStatus updates proposal status
func (r *ProposalRepository) UpdateStatus(id int, status string) error {
	query := `UPDATE proposals SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update proposal status: %w", err)
	}
	return nil
}

// Count counts total proposals
func (r *ProposalRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM proposals").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count proposals: %w", err)
	}
	return count, nil
}

// CountByStatus counts proposals by status
func (r *ProposalRepository) CountByStatus(status string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM proposals WHERE status = $1`
	err := r.db.QueryRow(query, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count proposals by status: %w", err)
	}
	return count, nil
}

// CountByUserID counts proposals by user ID
func (r *ProposalRepository) CountByUserID(userID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM proposals WHERE user_id = $1`
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user proposals: %w", err)
	}
	return count, nil
}
