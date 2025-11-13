package repository

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
)

type AnnouncementRepository struct {
	db *sql.DB
}

func NewAnnouncementRepository(db *sql.DB) *AnnouncementRepository {
	return &AnnouncementRepository{db: db}
}

// Create creates a new announcement
func (r *AnnouncementRepository) Create(req *models.CreateAnnouncementRequest) (*models.Announcement, error) {
	query := `
		INSERT INTO announcements (title, content, telegram_file_id, filename, file_type, posted_by_admin_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, title, content, telegram_file_id, filename, file_type, posted_by_admin_id, created_at, is_active
	`

	var announcement models.Announcement
	err := r.db.QueryRow(
		query,
		req.Title,
		req.Content,
		req.TelegramFileID,
		req.Filename,
		req.FileType,
		req.PostedByAdminID,
	).Scan(
		&announcement.ID,
		&announcement.Title,
		&announcement.Content,
		&announcement.TelegramFileID,
		&announcement.Filename,
		&announcement.FileType,
		&announcement.PostedByAdminID,
		&announcement.CreatedAt,
		&announcement.IsActive,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create announcement: %w", err)
	}

	return &announcement, nil
}

// GetByID gets announcement by ID
func (r *AnnouncementRepository) GetByID(id int) (*models.Announcement, error) {
	query := `
		SELECT id, title, content, telegram_file_id, filename, file_type, posted_by_admin_id, created_at, is_active
		FROM announcements
		WHERE id = $1
	`

	var announcement models.Announcement
	err := r.db.QueryRow(query, id).Scan(
		&announcement.ID,
		&announcement.Title,
		&announcement.Content,
		&announcement.TelegramFileID,
		&announcement.Filename,
		&announcement.FileType,
		&announcement.PostedByAdminID,
		&announcement.CreatedAt,
		&announcement.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get announcement: %w", err)
	}

	return &announcement, nil
}

// GetActive gets all active announcements with pagination
func (r *AnnouncementRepository) GetActive(limit, offset int) ([]*models.Announcement, error) {
	query := `
		SELECT id, title, content, telegram_file_id, filename, file_type, posted_by_admin_id, created_at, is_active
		FROM announcements
		WHERE is_active = 1
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get active announcements: %w", err)
	}
	defer rows.Close()

	var announcements []*models.Announcement
	for rows.Next() {
		var announcement models.Announcement
		err := rows.Scan(
			&announcement.ID,
			&announcement.Title,
			&announcement.Content,
			&announcement.TelegramFileID,
			&announcement.Filename,
			&announcement.FileType,
			&announcement.PostedByAdminID,
			&announcement.CreatedAt,
			&announcement.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan announcement: %w", err)
		}
		announcements = append(announcements, &announcement)
	}

	return announcements, nil
}

// GetAll gets all announcements with pagination (for admin)
func (r *AnnouncementRepository) GetAll(limit, offset int) ([]*models.Announcement, error) {
	query := `
		SELECT id, title, content, telegram_file_id, filename, file_type, posted_by_admin_id, created_at, is_active
		FROM announcements
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get announcements: %w", err)
	}
	defer rows.Close()

	var announcements []*models.Announcement
	for rows.Next() {
		var announcement models.Announcement
		err := rows.Scan(
			&announcement.ID,
			&announcement.Title,
			&announcement.Content,
			&announcement.TelegramFileID,
			&announcement.Filename,
			&announcement.FileType,
			&announcement.PostedByAdminID,
			&announcement.CreatedAt,
			&announcement.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan announcement: %w", err)
		}
		announcements = append(announcements, &announcement)
	}

	return announcements, nil
}

// Update updates an existing announcement
func (r *AnnouncementRepository) Update(id int, req *models.CreateAnnouncementRequest) (*models.Announcement, error) {
	query := `
		UPDATE announcements
		SET title = $1, content = $2, telegram_file_id = $3, filename = $4, file_type = $5
		WHERE id = $6
		RETURNING id, title, content, telegram_file_id, filename, file_type, posted_by_admin_id, created_at, is_active
	`

	var announcement models.Announcement
	err := r.db.QueryRow(
		query,
		req.Title,
		req.Content,
		req.TelegramFileID,
		req.Filename,
		req.FileType,
		id,
	).Scan(
		&announcement.ID,
		&announcement.Title,
		&announcement.Content,
		&announcement.TelegramFileID,
		&announcement.Filename,
		&announcement.FileType,
		&announcement.PostedByAdminID,
		&announcement.CreatedAt,
		&announcement.IsActive,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update announcement: %w", err)
	}

	return &announcement, nil
}

// ToggleActive toggles the active status of an announcement
func (r *AnnouncementRepository) ToggleActive(id int) error {
	query := `UPDATE announcements SET is_active = CASE WHEN is_active = 1 THEN 0 ELSE 1 END WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to toggle announcement active status: %w", err)
	}
	return nil
}

// Delete deletes an announcement
func (r *AnnouncementRepository) Delete(id int) error {
	query := `DELETE FROM announcements WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete announcement: %w", err)
	}
	return nil
}

// Count counts total announcements
func (r *AnnouncementRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM announcements").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count announcements: %w", err)
	}
	return count, nil
}

// CountActive counts active announcements
func (r *AnnouncementRepository) CountActive() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM announcements WHERE is_active = 1`
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count active announcements: %w", err)
	}
	return count, nil
}
