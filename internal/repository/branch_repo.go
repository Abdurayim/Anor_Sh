package repository

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
)

type BranchRepository struct {
	db *sql.DB
}

func NewBranchRepository(db *sql.DB) *BranchRepository {
	return &BranchRepository{db: db}
}

const branchColumns = "id, code, name_uz, name_ru, is_active"

func scanBranch(row interface{ Scan(...any) error }) (*models.Branch, error) {
	var b models.Branch
	if err := row.Scan(&b.ID, &b.Code, &b.NameUz, &b.NameRu, &b.IsActive); err != nil {
		return nil, err
	}
	return &b, nil
}

// GetActive returns active branches ordered by ID.
func (r *BranchRepository) GetActive() ([]*models.Branch, error) {
	rows, err := r.db.Query("SELECT " + branchColumns + " FROM branches WHERE is_active = 1 ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}
	defer rows.Close()

	var branches []*models.Branch
	for rows.Next() {
		b, err := scanBranch(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan branch: %w", err)
		}
		branches = append(branches, b)
	}
	return branches, rows.Err()
}

// GetByID returns a branch or nil if it does not exist.
func (r *BranchRepository) GetByID(id int) (*models.Branch, error) {
	b, err := scanBranch(r.db.QueryRow("SELECT "+branchColumns+" FROM branches WHERE id = ?", id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}
	return b, nil
}

// GetByCode returns a branch or nil if it does not exist.
func (r *BranchRepository) GetByCode(code string) (*models.Branch, error) {
	b, err := scanBranch(r.db.QueryRow("SELECT "+branchColumns+" FROM branches WHERE code = ?", code))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}
	return b, nil
}
