package repository

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
)

type ClassRepository struct {
	db *sql.DB
}

func NewClassRepository(db *sql.DB) *ClassRepository {
	return &ClassRepository{db: db}
}

// Create creates a new class
func (r *ClassRepository) Create(className string) (*models.Class, error) {
	query := `
		INSERT INTO classes (class_name, is_active)
		VALUES ($1, 1)
		RETURNING id, class_name, is_active, created_at
	`

	var class models.Class
	err := r.db.QueryRow(query, className).Scan(
		&class.ID,
		&class.ClassName,
		&class.IsActive,
		&class.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create class: %w", err)
	}

	return &class, nil
}

// GetAll gets all classes
func (r *ClassRepository) GetAll() ([]*models.Class, error) {
	query := `
		SELECT id, class_name, is_active, created_at
		FROM classes
		ORDER BY class_name ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes: %w", err)
	}
	defer rows.Close()

	var classes []*models.Class
	for rows.Next() {
		var class models.Class
		err := rows.Scan(
			&class.ID,
			&class.ClassName,
			&class.IsActive,
			&class.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan class: %w", err)
		}
		classes = append(classes, &class)
	}

	return classes, nil
}

// GetActive gets all active classes
func (r *ClassRepository) GetActive() ([]*models.Class, error) {
	query := `
		SELECT id, class_name, is_active, created_at
		FROM classes
		WHERE is_active = 1
		ORDER BY class_name ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active classes: %w", err)
	}
	defer rows.Close()

	var classes []*models.Class
	for rows.Next() {
		var class models.Class
		err := rows.Scan(
			&class.ID,
			&class.ClassName,
			&class.IsActive,
			&class.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan class: %w", err)
		}
		classes = append(classes, &class)
	}

	return classes, nil
}

// GetByName gets class by name
func (r *ClassRepository) GetByName(className string) (*models.Class, error) {
	query := `
		SELECT id, class_name, is_active, created_at
		FROM classes
		WHERE class_name = $1
	`

	var class models.Class
	err := r.db.QueryRow(query, className).Scan(
		&class.ID,
		&class.ClassName,
		&class.IsActive,
		&class.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get class: %w", err)
	}

	return &class, nil
}

// Delete deletes a class by name
func (r *ClassRepository) Delete(className string) error {
	query := `DELETE FROM classes WHERE class_name = $1`
	result, err := r.db.Exec(query, className)
	if err != nil {
		return fmt.Errorf("failed to delete class: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("class not found")
	}

	return nil
}

// ToggleActive toggles class active status
func (r *ClassRepository) ToggleActive(className string) error {
	query := `
		UPDATE classes
		SET is_active = CASE WHEN is_active = 1 THEN 0 ELSE 1 END
		WHERE class_name = $1
	`
	result, err := r.db.Exec(query, className)
	if err != nil {
		return fmt.Errorf("failed to toggle class status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("class not found")
	}

	return nil
}

// Count counts total classes
func (r *ClassRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM classes").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count classes: %w", err)
	}
	return count, nil
}

// Exists checks if a class name exists and is active
func (r *ClassRepository) Exists(className string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM classes WHERE class_name = $1 AND is_active = 1)`
	var exists bool
	err := r.db.QueryRow(query, className).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check class existence: %w", err)
	}
	return exists, nil
}
