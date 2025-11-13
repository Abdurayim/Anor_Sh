package repository

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
)

type TimetableRepository struct {
	db *sql.DB
}

func NewTimetableRepository(db *sql.DB) *TimetableRepository {
	return &TimetableRepository{db: db}
}

// Create creates a new timetable or updates existing one for the class
func (r *TimetableRepository) Create(req *models.CreateTimetableRequest) (*models.Timetable, error) {
	query := `
		INSERT INTO timetables (class_id, telegram_file_id, filename, file_type, mime_type, uploaded_by_admin_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, class_id, telegram_file_id, filename, file_type, mime_type, uploaded_by_admin_id, created_at, updated_at
	`

	var timetable models.Timetable
	err := r.db.QueryRow(
		query,
		req.ClassID,
		req.TelegramFileID,
		req.Filename,
		req.FileType,
		req.MimeType,
		req.UploadedByAdminID,
	).Scan(
		&timetable.ID,
		&timetable.ClassID,
		&timetable.TelegramFileID,
		&timetable.Filename,
		&timetable.FileType,
		&timetable.MimeType,
		&timetable.UploadedByAdminID,
		&timetable.CreatedAt,
		&timetable.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create timetable: %w", err)
	}

	return &timetable, nil
}

// GetByID gets timetable by ID
func (r *TimetableRepository) GetByID(id int) (*models.Timetable, error) {
	query := `
		SELECT id, class_id, telegram_file_id, filename, file_type, mime_type, uploaded_by_admin_id, created_at, updated_at
		FROM timetables
		WHERE id = $1
	`

	var timetable models.Timetable
	err := r.db.QueryRow(query, id).Scan(
		&timetable.ID,
		&timetable.ClassID,
		&timetable.TelegramFileID,
		&timetable.Filename,
		&timetable.FileType,
		&timetable.MimeType,
		&timetable.UploadedByAdminID,
		&timetable.CreatedAt,
		&timetable.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get timetable: %w", err)
	}

	return &timetable, nil
}

// GetByClassID gets the latest timetable for a specific class
func (r *TimetableRepository) GetByClassID(classID int) (*models.Timetable, error) {
	query := `
		SELECT id, class_id, telegram_file_id, filename, file_type, mime_type, uploaded_by_admin_id, created_at, updated_at
		FROM timetables
		WHERE class_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var timetable models.Timetable
	err := r.db.QueryRow(query, classID).Scan(
		&timetable.ID,
		&timetable.ClassID,
		&timetable.TelegramFileID,
		&timetable.Filename,
		&timetable.FileType,
		&timetable.MimeType,
		&timetable.UploadedByAdminID,
		&timetable.CreatedAt,
		&timetable.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get timetable by class: %w", err)
	}

	return &timetable, nil
}

// GetByClassName gets the latest timetable for a class by name
func (r *TimetableRepository) GetByClassName(className string) (*models.Timetable, error) {
	query := `
		SELECT t.id, t.class_id, t.telegram_file_id, t.filename, t.file_type, t.mime_type,
		       t.uploaded_by_admin_id, t.created_at, t.updated_at
		FROM timetables t
		INNER JOIN classes c ON t.class_id = c.id
		WHERE c.class_name = $1
		ORDER BY t.created_at DESC
		LIMIT 1
	`

	var timetable models.Timetable
	err := r.db.QueryRow(query, className).Scan(
		&timetable.ID,
		&timetable.ClassID,
		&timetable.TelegramFileID,
		&timetable.Filename,
		&timetable.FileType,
		&timetable.MimeType,
		&timetable.UploadedByAdminID,
		&timetable.CreatedAt,
		&timetable.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get timetable by class name: %w", err)
	}

	return &timetable, nil
}

// GetAll gets all timetables with pagination (for admin)
func (r *TimetableRepository) GetAll(limit, offset int) ([]*models.Timetable, error) {
	query := `
		SELECT id, class_id, telegram_file_id, filename, file_type, mime_type, uploaded_by_admin_id, created_at, updated_at
		FROM timetables
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get timetables: %w", err)
	}
	defer rows.Close()

	var timetables []*models.Timetable
	for rows.Next() {
		var timetable models.Timetable
		err := rows.Scan(
			&timetable.ID,
			&timetable.ClassID,
			&timetable.TelegramFileID,
			&timetable.Filename,
			&timetable.FileType,
			&timetable.MimeType,
			&timetable.UploadedByAdminID,
			&timetable.CreatedAt,
			&timetable.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan timetable: %w", err)
		}
		timetables = append(timetables, &timetable)
	}

	return timetables, nil
}

// Update updates an existing timetable
func (r *TimetableRepository) Update(id int, req *models.CreateTimetableRequest) (*models.Timetable, error) {
	query := `
		UPDATE timetables
		SET telegram_file_id = $1, filename = $2, file_type = $3, mime_type = $4, uploaded_by_admin_id = $5
		WHERE id = $6
		RETURNING id, class_id, telegram_file_id, filename, file_type, mime_type, uploaded_by_admin_id, created_at, updated_at
	`

	var timetable models.Timetable
	err := r.db.QueryRow(
		query,
		req.TelegramFileID,
		req.Filename,
		req.FileType,
		req.MimeType,
		req.UploadedByAdminID,
		id,
	).Scan(
		&timetable.ID,
		&timetable.ClassID,
		&timetable.TelegramFileID,
		&timetable.Filename,
		&timetable.FileType,
		&timetable.MimeType,
		&timetable.UploadedByAdminID,
		&timetable.CreatedAt,
		&timetable.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update timetable: %w", err)
	}

	return &timetable, nil
}

// Delete deletes a timetable
func (r *TimetableRepository) Delete(id int) error {
	query := `DELETE FROM timetables WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete timetable: %w", err)
	}
	return nil
}

// DeleteByClassID deletes all timetables for a specific class
func (r *TimetableRepository) DeleteByClassID(classID int) error {
	query := `DELETE FROM timetables WHERE class_id = $1`
	_, err := r.db.Exec(query, classID)
	if err != nil {
		return fmt.Errorf("failed to delete timetables for class: %w", err)
	}
	return nil
}

// Count counts total timetables
func (r *TimetableRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM timetables").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count timetables: %w", err)
	}
	return count, nil
}

// Exists checks if a timetable exists for a class
func (r *TimetableRepository) Exists(classID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM timetables WHERE class_id = $1)`
	err := r.db.QueryRow(query, classID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check timetable existence: %w", err)
	}
	return exists, nil
}
