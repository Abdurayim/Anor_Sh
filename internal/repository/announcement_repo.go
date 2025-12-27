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
		INSERT INTO announcements (title, content, telegram_file_id, filename, file_type, admin_id)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id, title, content, telegram_file_id, filename, file_type, admin_id, created_at, is_active
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
		SELECT id, title, content, telegram_file_id, filename, file_type, admin_id, teacher_id, created_at, is_active
		FROM announcements
		WHERE id = ?
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
		&announcement.PostedByTeacherID,
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
		SELECT id, title, content, telegram_file_id, filename, file_type, admin_id, teacher_id, created_at, is_active
		FROM announcements
		WHERE is_active = 1
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
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
			&announcement.PostedByTeacherID,
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
		SELECT id, title, content, telegram_file_id, filename, file_type, admin_id, teacher_id, created_at, is_active
		FROM announcements
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
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
			&announcement.PostedByTeacherID,
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
		SET title = ?, content = ?, telegram_file_id = ?, filename = ?, file_type = ?
		WHERE id = ?
		RETURNING id, title, content, telegram_file_id, filename, file_type, admin_id, created_at, is_active
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
	query := `UPDATE announcements SET is_active = CASE WHEN is_active = 1 THEN 0 ELSE 1 END WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to toggle announcement active status: %w", err)
	}
	return nil
}

// Delete deletes an announcement
func (r *AnnouncementRepository) Delete(id int) error {
	query := `DELETE FROM announcements WHERE id = ?`
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

// GetByTeacherID gets all announcements posted by a specific teacher
func (r *AnnouncementRepository) GetByTeacherID(teacherID int, limit, offset int) ([]*models.AnnouncementWithClasses, error) {
	query := `
		SELECT
			a.id, a.title, a.content, a.telegram_file_id, a.filename, a.file_type,
			a.admin_id, a.teacher_id, a.created_at, a.is_active
		FROM announcements a
		WHERE a.teacher_id = ?
		ORDER BY a.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, teacherID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get teacher announcements: %w", err)
	}
	defer rows.Close()

	var announcements []*models.AnnouncementWithClasses
	for rows.Next() {
		var a models.AnnouncementWithClasses
		err := rows.Scan(
			&a.ID,
			&a.Title,
			&a.Content,
			&a.TelegramFileID,
			&a.Filename,
			&a.FileType,
			&a.PostedByAdminID,
			&a.PostedByTeacherID,
			&a.CreatedAt,
			&a.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan announcement: %w", err)
		}

		// Get associated classes
		classQuery := `
			SELECT ac.class_id, c.class_name
			FROM announcement_classes ac
			JOIN classes c ON ac.class_id = c.id
			WHERE ac.announcement_id = ?
			ORDER BY c.class_name
		`
		classRows, err := r.db.Query(classQuery, a.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get announcement classes: %w", err)
		}

		var classIDs []int
		var classNames []string
		for classRows.Next() {
			var classID int
			var className string
			if err := classRows.Scan(&classID, &className); err != nil {
				classRows.Close()
				return nil, fmt.Errorf("failed to scan class: %w", err)
			}
			classIDs = append(classIDs, classID)
			classNames = append(classNames, className)
		}
		classRows.Close()

		a.ClassIDs = classIDs
		a.ClassNames = classNames

		announcements = append(announcements, &a)
	}

	return announcements, nil
}
