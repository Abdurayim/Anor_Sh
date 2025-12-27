package repository

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
)

type ComplaintRepository struct {
	db *sql.DB
}

func NewComplaintRepository(db *sql.DB) *ComplaintRepository {
	return &ComplaintRepository{db: db}
}

// Create creates a new complaint
func (r *ComplaintRepository) Create(req *models.CreateComplaintRequest) (*models.Complaint, error) {
	query := `
		INSERT INTO complaints (user_id, complaint_text, telegram_file_id, filename)
		VALUES (?, ?, ?, ?)
		RETURNING id, user_id, complaint_text, telegram_file_id, filename, created_at, status
	`

	var complaint models.Complaint
	err := r.db.QueryRow(
		query,
		req.UserID,
		req.ComplaintText,
		req.TelegramFileID,
		req.Filename,
	).Scan(
		&complaint.ID,
		&complaint.UserID,
		&complaint.ComplaintText,
		&complaint.TelegramFileID,
		&complaint.Filename,
		&complaint.CreatedAt,
		&complaint.Status,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create complaint: %w", err)
	}

	return &complaint, nil
}

// GetByID gets complaint by ID
func (r *ComplaintRepository) GetByID(id int) (*models.Complaint, error) {
	query := `
		SELECT id, user_id, complaint_text, telegram_file_id, filename, created_at, status
		FROM complaints
		WHERE id = ?
	`

	var complaint models.Complaint
	err := r.db.QueryRow(query, id).Scan(
		&complaint.ID,
		&complaint.UserID,
		&complaint.ComplaintText,
		&complaint.TelegramFileID,
		&complaint.Filename,
		&complaint.CreatedAt,
		&complaint.Status,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get complaint: %w", err)
	}

	return &complaint, nil
}

// GetByUserID gets complaints by user ID (indexed, fast query)
func (r *ComplaintRepository) GetByUserID(userID int, limit, offset int) ([]*models.Complaint, error) {
	query := `
		SELECT id, user_id, complaint_text, telegram_file_id, filename, created_at, status
		FROM complaints
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get complaints: %w", err)
	}
	defer rows.Close()

	var complaints []*models.Complaint
	for rows.Next() {
		var complaint models.Complaint
		err := rows.Scan(
			&complaint.ID,
			&complaint.UserID,
			&complaint.ComplaintText,
			&complaint.TelegramFileID,
			&complaint.Filename,
			&complaint.CreatedAt,
			&complaint.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan complaint: %w", err)
		}
		complaints = append(complaints, &complaint)
	}

	return complaints, nil
}

// GetAll gets all complaints with pagination (for admin)
func (r *ComplaintRepository) GetAll(limit, offset int) ([]*models.Complaint, error) {
	query := `
		SELECT id, user_id, complaint_text, telegram_file_id, filename, created_at, status
		FROM complaints
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get complaints: %w", err)
	}
	defer rows.Close()

	var complaints []*models.Complaint
	for rows.Next() {
		var complaint models.Complaint
		err := rows.Scan(
			&complaint.ID,
			&complaint.UserID,
			&complaint.ComplaintText,
			&complaint.TelegramFileID,
			&complaint.Filename,
			&complaint.CreatedAt,
			&complaint.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan complaint: %w", err)
		}
		complaints = append(complaints, &complaint)
	}

	return complaints, nil
}

// GetAllWithUser gets all complaints with user info using view (optimized for admin)
func (r *ComplaintRepository) GetAllWithUser(limit, offset int) ([]*models.ComplaintWithUser, error) {
	query := `
		SELECT id, user_id, complaint_text, telegram_file_id, filename, created_at, status,
		       telegram_id as user_telegram_id, telegram_username, phone_number, language
		FROM v_complaints_with_user
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get complaints with user: %w", err)
	}
	defer rows.Close()

	var complaints []*models.ComplaintWithUser
	for rows.Next() {
		var complaint models.ComplaintWithUser
		err := rows.Scan(
			&complaint.ID,
			&complaint.UserID,
			&complaint.ComplaintText,
			&complaint.TelegramFileID,
			&complaint.Filename,
			&complaint.CreatedAt,
			&complaint.Status,
			&complaint.UserTelegramID,
			&complaint.TelegramUsername,
			&complaint.PhoneNumber,
			&complaint.Language,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan complaint with user: %w", err)
		}
		complaints = append(complaints, &complaint)
	}

	return complaints, nil
}

// GetByStatus gets complaints by status (indexed, fast query)
func (r *ComplaintRepository) GetByStatus(status string, limit, offset int) ([]*models.Complaint, error) {
	query := `
		SELECT id, user_id, complaint_text, telegram_file_id, filename, created_at, status
		FROM complaints
		WHERE status = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get complaints by status: %w", err)
	}
	defer rows.Close()

	var complaints []*models.Complaint
	for rows.Next() {
		var complaint models.Complaint
		err := rows.Scan(
			&complaint.ID,
			&complaint.UserID,
			&complaint.ComplaintText,
			&complaint.TelegramFileID,
			&complaint.Filename,
			&complaint.CreatedAt,
			&complaint.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan complaint: %w", err)
		}
		complaints = append(complaints, &complaint)
	}

	return complaints, nil
}

// UpdateStatus updates complaint status
func (r *ComplaintRepository) UpdateStatus(id int, status string) error {
	query := `UPDATE complaints SET status = ? WHERE id = ?`
	_, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update complaint status: %w", err)
	}
	return nil
}

// Count counts total complaints
func (r *ComplaintRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM complaints").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count complaints: %w", err)
	}
	return count, nil
}

// CountByStatus counts complaints by status
func (r *ComplaintRepository) CountByStatus(status string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM complaints WHERE status = ?`
	err := r.db.QueryRow(query, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count complaints by status: %w", err)
	}
	return count, nil
}

// CountByUserID counts complaints by user ID
func (r *ComplaintRepository) CountByUserID(userID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM complaints WHERE user_id = ?`
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user complaints: %w", err)
	}
	return count, nil
}
