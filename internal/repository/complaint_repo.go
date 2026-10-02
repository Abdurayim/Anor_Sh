package repository

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
)

// ComplaintRepository stores parent complaints. Branch filtering uses the parent's branch.
type ComplaintRepository struct {
	db *sql.DB
}

func NewComplaintRepository(db *sql.DB) *ComplaintRepository {
	return &ComplaintRepository{db: db}
}

const complaintColumns = "id, user_id, student_id, complaint_text, COALESCE(telegram_file_id, ''), COALESCE(filename, ''), created_at, status"

func scanComplaint(row interface{ Scan(...any) error }, extra ...any) (*models.Complaint, error) {
	var x models.Complaint
	var studentID sql.NullInt64
	dest := append([]any{&x.ID, &x.UserID, &studentID, &x.ComplaintText, &x.TelegramFileID, &x.Filename, &x.CreatedAt, &x.Status}, extra...)
	if err := row.Scan(dest...); err != nil {
		return nil, err
	}
	if studentID.Valid {
		id := int(studentID.Int64)
		x.StudentID = &id
	}
	return &x, nil
}

func (r *ComplaintRepository) getMany(query string, args ...any) ([]*models.Complaint, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get complaints: %w", err)
	}
	defer rows.Close()

	var items []*models.Complaint
	for rows.Next() {
		x, err := scanComplaint(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan complaint: %w", err)
		}
		items = append(items, x)
	}
	return items, rows.Err()
}

// Create creates a new complaint
func (r *ComplaintRepository) Create(req *models.CreateComplaintRequest) (*models.Complaint, error) {
	x, err := scanComplaint(r.db.QueryRow(`
		INSERT INTO complaints (user_id, student_id, complaint_text, telegram_file_id, filename)
		VALUES (?, ?, ?, ?, ?)
		RETURNING `+complaintColumns,
		req.UserID, req.StudentID, req.ComplaintText, req.TelegramFileID, req.Filename,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create complaint: %w", err)
	}
	return x, nil
}

// SetFile stores the Telegram file of the generated document
func (r *ComplaintRepository) SetFile(id int, telegramFileID, filename string) error {
	_, err := r.db.Exec("UPDATE complaints SET telegram_file_id = ?, filename = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		telegramFileID, filename, id)
	if err != nil {
		return fmt.Errorf("failed to update complaint file: %w", err)
	}
	return nil
}

// GetByID gets complaint by ID
func (r *ComplaintRepository) GetByID(id int) (*models.Complaint, error) {
	x, err := scanComplaint(r.db.QueryRow("SELECT "+complaintColumns+" FROM complaints WHERE id = ?", id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get complaint: %w", err)
	}
	return x, nil
}

// GetByUserID gets complaints of a parent
func (r *ComplaintRepository) GetByUserID(userID int, limit, offset int) ([]*models.Complaint, error) {
	return r.getMany("SELECT "+complaintColumns+" FROM complaints WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?",
		userID, limit, offset)
}

// GetAll gets complaints of a branch (branchID 0 = all branches)
func (r *ComplaintRepository) GetAll(branchID, limit, offset int) ([]*models.Complaint, error) {
	return r.getMany(`
		SELECT `+complaintColumns+` FROM complaints
		WHERE (? = 0 OR user_id IN (SELECT id FROM users WHERE branch_id = ?))
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, branchID, branchID, limit, offset)
}

// GetAllWithUser gets complaints of a branch with parent and child info (branchID 0 = all branches)
func (r *ComplaintRepository) GetAllWithUser(branchID, limit, offset int) ([]*models.ComplaintWithUser, error) {
	rows, err := r.db.Query(`
		SELECT `+complaintColumns+`,
		       telegram_id, COALESCE(telegram_username, ''), phone_number, language,
		       COALESCE(student_first_name || ' ' || student_last_name, ''), COALESCE(class_name, ''),
		       COALESCE(branch_id, 0)
		FROM v_complaints_with_user
		WHERE (? = 0 OR branch_id = ?)
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, branchID, branchID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get complaints with user: %w", err)
	}
	defer rows.Close()

	var items []*models.ComplaintWithUser
	for rows.Next() {
		var w models.ComplaintWithUser
		x, err := scanComplaint(rows, &w.UserTelegramID, &w.TelegramUsername, &w.PhoneNumber, &w.Language,
			&w.StudentName, &w.ClassName, &w.BranchID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan complaint with user: %w", err)
		}
		w.Complaint = *x
		items = append(items, &w)
	}
	return items, rows.Err()
}

// UpdateStatus updates complaint status
func (r *ComplaintRepository) UpdateStatus(id int, status string) error {
	_, err := r.db.Exec("UPDATE complaints SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", status, id)
	if err != nil {
		return fmt.Errorf("failed to update complaint status: %w", err)
	}
	return nil
}

// Count counts complaints of a branch, optionally only with the given status
// (branchID 0 = all branches, status "" = any status)
func (r *ComplaintRepository) Count(branchID int, status string) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM complaints
		WHERE (? = 0 OR user_id IN (SELECT id FROM users WHERE branch_id = ?))
		  AND (? = '' OR status = ?)
	`, branchID, branchID, status, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count complaints: %w", err)
	}
	return count, nil
}

// CountByUserID counts complaints of a parent
func (r *ComplaintRepository) CountByUserID(userID int) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM complaints WHERE user_id = ?", userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user complaints: %w", err)
	}
	return count, nil
}
