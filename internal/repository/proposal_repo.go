package repository

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
)

// ProposalRepository stores parent proposals. Branch filtering uses the parent's branch.
type ProposalRepository struct {
	db *sql.DB
}

func NewProposalRepository(db *sql.DB) *ProposalRepository {
	return &ProposalRepository{db: db}
}

const proposalColumns = "id, user_id, student_id, proposal_text, COALESCE(telegram_file_id, ''), COALESCE(filename, ''), created_at, status"

func scanProposal(row interface{ Scan(...any) error }, extra ...any) (*models.Proposal, error) {
	var x models.Proposal
	var studentID sql.NullInt64
	dest := append([]any{&x.ID, &x.UserID, &studentID, &x.ProposalText, &x.TelegramFileID, &x.Filename, &x.CreatedAt, &x.Status}, extra...)
	if err := row.Scan(dest...); err != nil {
		return nil, err
	}
	if studentID.Valid {
		id := int(studentID.Int64)
		x.StudentID = &id
	}
	return &x, nil
}

func (r *ProposalRepository) getMany(query string, args ...any) ([]*models.Proposal, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposals: %w", err)
	}
	defer rows.Close()

	var items []*models.Proposal
	for rows.Next() {
		x, err := scanProposal(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan proposal: %w", err)
		}
		items = append(items, x)
	}
	return items, rows.Err()
}

// Create creates a new proposal
func (r *ProposalRepository) Create(req *models.CreateProposalRequest) (*models.Proposal, error) {
	x, err := scanProposal(r.db.QueryRow(`
		INSERT INTO proposals (user_id, student_id, proposal_text, telegram_file_id, filename)
		VALUES (?, ?, ?, ?, ?)
		RETURNING `+proposalColumns,
		req.UserID, req.StudentID, req.ProposalText, req.TelegramFileID, req.Filename,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal: %w", err)
	}
	return x, nil
}

// SetFile stores the Telegram file of the generated document
func (r *ProposalRepository) SetFile(id int, telegramFileID, filename string) error {
	_, err := r.db.Exec("UPDATE proposals SET telegram_file_id = ?, filename = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		telegramFileID, filename, id)
	if err != nil {
		return fmt.Errorf("failed to update proposal file: %w", err)
	}
	return nil
}

// GetByID gets proposal by ID
func (r *ProposalRepository) GetByID(id int) (*models.Proposal, error) {
	x, err := scanProposal(r.db.QueryRow("SELECT "+proposalColumns+" FROM proposals WHERE id = ?", id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}
	return x, nil
}

// GetByUserID gets proposals of a parent
func (r *ProposalRepository) GetByUserID(userID int, limit, offset int) ([]*models.Proposal, error) {
	return r.getMany("SELECT "+proposalColumns+" FROM proposals WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?",
		userID, limit, offset)
}

// GetAll gets proposals of a branch (branchID 0 = all branches)
func (r *ProposalRepository) GetAll(branchID, limit, offset int) ([]*models.Proposal, error) {
	return r.getMany(`
		SELECT `+proposalColumns+` FROM proposals
		WHERE (? = 0 OR user_id IN (SELECT id FROM users WHERE branch_id = ?))
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, branchID, branchID, limit, offset)
}

// GetAllWithUser gets proposals of a branch with parent and child info (branchID 0 = all branches)
func (r *ProposalRepository) GetAllWithUser(branchID, limit, offset int) ([]*models.ProposalWithUser, error) {
	rows, err := r.db.Query(`
		SELECT `+proposalColumns+`,
		       telegram_id, COALESCE(telegram_username, ''), phone_number, language,
		       COALESCE(student_first_name || ' ' || student_last_name, ''), COALESCE(class_name, ''),
		       COALESCE(branch_id, 0)
		FROM v_proposals_with_user
		WHERE (? = 0 OR branch_id = ?)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, branchID, branchID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposals with user: %w", err)
	}
	defer rows.Close()

	var items []*models.ProposalWithUser
	for rows.Next() {
		var w models.ProposalWithUser
		x, err := scanProposal(rows, &w.UserTelegramID, &w.TelegramUsername, &w.PhoneNumber, &w.Language,
			&w.StudentName, &w.ClassName, &w.BranchID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan proposal with user: %w", err)
		}
		w.Proposal = *x
		items = append(items, &w)
	}
	return items, rows.Err()
}

// UpdateStatus updates proposal status
func (r *ProposalRepository) UpdateStatus(id int, status string) error {
	_, err := r.db.Exec("UPDATE proposals SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", status, id)
	if err != nil {
		return fmt.Errorf("failed to update proposal status: %w", err)
	}
	return nil
}

// Count counts proposals of a branch, optionally only with the given status
// (branchID 0 = all branches, status "" = any status)
func (r *ProposalRepository) Count(branchID int, status string) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM proposals
		WHERE (? = 0 OR user_id IN (SELECT id FROM users WHERE branch_id = ?))
		  AND (? = '' OR status = ?)
	`, branchID, branchID, status, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count proposals: %w", err)
	}
	return count, nil
}

// CountByUserID counts proposals of a parent
func (r *ProposalRepository) CountByUserID(userID int) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM proposals WHERE user_id = ?", userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user proposals: %w", err)
	}
	return count, nil
}
