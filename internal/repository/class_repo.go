package repository

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
)

// ClassRepository manages classes. Class names are unique within a branch, so every
// name-based lookup takes the branch ID.
type ClassRepository struct {
	db *sql.DB
}

func NewClassRepository(db *sql.DB) *ClassRepository {
	return &ClassRepository{db: db}
}

const classColumns = "id, branch_id, class_name, is_active, created_at"

func scanClass(row interface{ Scan(...any) error }) (*models.Class, error) {
	var class models.Class
	if err := row.Scan(&class.ID, &class.BranchID, &class.ClassName, &class.IsActive, &class.CreatedAt); err != nil {
		return nil, err
	}
	return &class, nil
}

func (r *ClassRepository) getOne(query string, args ...any) (*models.Class, error) {
	class, err := scanClass(r.db.QueryRow(query, args...))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get class: %w", err)
	}
	return class, nil
}

func (r *ClassRepository) getMany(query string, args ...any) ([]*models.Class, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes: %w", err)
	}
	defer rows.Close()

	var classes []*models.Class
	for rows.Next() {
		class, err := scanClass(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan class: %w", err)
		}
		classes = append(classes, class)
	}
	return classes, rows.Err()
}

// Create creates a new class in a branch
func (r *ClassRepository) Create(branchID int, className string) (*models.Class, error) {
	result, err := r.db.Exec("INSERT INTO classes (branch_id, class_name, is_active) VALUES (?, ?, 1)", branchID, className)
	if err != nil {
		return nil, fmt.Errorf("failed to create class: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return r.GetByID(int(id))
}

// GetAll gets all classes of a branch (branchID 0 = all branches)
func (r *ClassRepository) GetAll(branchID int) ([]*models.Class, error) {
	return r.getMany(`
		SELECT `+classColumns+` FROM classes
		WHERE (? = 0 OR branch_id = ?)
		ORDER BY branch_id, class_name ASC
	`, branchID, branchID)
}

// GetActive gets active classes of a branch (branchID 0 = all branches)
func (r *ClassRepository) GetActive(branchID int) ([]*models.Class, error) {
	return r.getMany(`
		SELECT `+classColumns+` FROM classes
		WHERE is_active = 1 AND (? = 0 OR branch_id = ?)
		ORDER BY branch_id, class_name ASC
	`, branchID, branchID)
}

// GetByID gets class by ID
func (r *ClassRepository) GetByID(id int) (*models.Class, error) {
	return r.getOne("SELECT "+classColumns+" FROM classes WHERE id = ?", id)
}

// GetByName gets class by name within a branch
func (r *ClassRepository) GetByName(branchID int, className string) (*models.Class, error) {
	return r.getOne("SELECT "+classColumns+" FROM classes WHERE branch_id = ? AND class_name = ?", branchID, className)
}

// Delete deletes a class by name within a branch
func (r *ClassRepository) Delete(branchID int, className string) error {
	return r.execOne("DELETE FROM classes WHERE branch_id = ? AND class_name = ?", "delete class", branchID, className)
}

// DeleteByID deletes a class by ID
func (r *ClassRepository) DeleteByID(id int) error {
	return r.execOne("DELETE FROM classes WHERE id = ?", "delete class", id)
}

// ToggleActive toggles class active status
func (r *ClassRepository) ToggleActive(branchID int, className string) error {
	return r.execOne(`
		UPDATE classes
		SET is_active = CASE WHEN is_active = 1 THEN 0 ELSE 1 END, updated_at = CURRENT_TIMESTAMP
		WHERE branch_id = ? AND class_name = ?
	`, "toggle class status", branchID, className)
}

func (r *ClassRepository) execOne(query, action string, args ...any) error {
	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to %s: %w", action, err)
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

// Count counts classes of a branch (branchID 0 = all branches)
func (r *ClassRepository) Count(branchID int) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM classes WHERE (? = 0 OR branch_id = ?)", branchID, branchID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count classes: %w", err)
	}
	return count, nil
}

// Exists checks if an active class with this name exists in the branch
func (r *ClassRepository) Exists(branchID int, className string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM classes WHERE branch_id = ? AND class_name = ? AND is_active = 1)",
		branchID, className,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check class existence: %w", err)
	}
	return exists, nil
}

// BelongsToBranch reports whether the class exists and is in the given branch.
func (r *ClassRepository) BelongsToBranch(classID, branchID int) (bool, error) {
	var ok bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM classes WHERE id = ? AND branch_id = ?)", classID, branchID).Scan(&ok)
	if err != nil {
		return false, fmt.Errorf("failed to check class branch: %w", err)
	}
	return ok, nil
}
