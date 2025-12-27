package repository

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
)

type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// Create creates a new admin
func (r *AdminRepository) Create(phoneNumber, name string) (*models.Admin, error) {
	query := `
		INSERT INTO admins (phone_number, name)
		VALUES (?, ?)
		RETURNING id, phone_number, telegram_id, name, added_at
	`

	var admin models.Admin
	err := r.db.QueryRow(query, phoneNumber, name).Scan(
		&admin.ID,
		&admin.PhoneNumber,
		&admin.TelegramID,
		&admin.Name,
		&admin.AddedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	return &admin, nil
}

// GetByPhoneNumber gets admin by phone number (indexed, fast query)
func (r *AdminRepository) GetByPhoneNumber(phoneNumber string) (*models.Admin, error) {
	query := `
		SELECT id, phone_number, telegram_id, name, added_at
		FROM admins
		WHERE phone_number = ?
	`

	var admin models.Admin
	err := r.db.QueryRow(query, phoneNumber).Scan(
		&admin.ID,
		&admin.PhoneNumber,
		&admin.TelegramID,
		&admin.Name,
		&admin.AddedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}

	return &admin, nil
}

// GetByTelegramID gets admin by telegram ID (indexed, fast query)
func (r *AdminRepository) GetByTelegramID(telegramID int64) (*models.Admin, error) {
	query := `
		SELECT id, phone_number, telegram_id, name, added_at
		FROM admins
		WHERE telegram_id = ?
	`

	var admin models.Admin
	err := r.db.QueryRow(query, telegramID).Scan(
		&admin.ID,
		&admin.PhoneNumber,
		&admin.TelegramID,
		&admin.Name,
		&admin.AddedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}

	return &admin, nil
}

// GetAll gets all admins
func (r *AdminRepository) GetAll() ([]*models.Admin, error) {
	query := `
		SELECT id, phone_number, telegram_id, name, added_at
		FROM admins
		ORDER BY added_at ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get admins: %w", err)
	}
	defer rows.Close()

	var admins []*models.Admin
	for rows.Next() {
		var admin models.Admin
		err := rows.Scan(
			&admin.ID,
			&admin.PhoneNumber,
			&admin.TelegramID,
			&admin.Name,
			&admin.AddedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan admin: %w", err)
		}
		admins = append(admins, &admin)
	}

	return admins, nil
}

// UpdateTelegramID updates admin telegram ID
func (r *AdminRepository) UpdateTelegramID(phoneNumber string, telegramID int64) error {
	query := `UPDATE admins SET telegram_id = ? WHERE phone_number = ?`
	_, err := r.db.Exec(query, telegramID, phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to update admin telegram ID: %w", err)
	}
	return nil
}

// IsAdmin checks if phone number or telegram ID is an admin
// Note: Empty phone numbers are ignored to avoid false matches
func (r *AdminRepository) IsAdmin(phoneNumber string, telegramID int64) (bool, error) {
	// Build dynamic query based on what parameters are provided
	var query string
	var args []any

	if phoneNumber != "" && telegramID != 0 {
		// Both phone and telegram_id provided
		query = `
			SELECT EXISTS(
				SELECT 1 FROM admins
				WHERE phone_number = ? OR telegram_id = ?
			)
		`
		args = []any{phoneNumber, telegramID}
	} else if phoneNumber != "" {
		// Only phone number provided
		query = `
			SELECT EXISTS(
				SELECT 1 FROM admins
				WHERE phone_number = ?
			)
		`
		args = []any{phoneNumber}
	} else if telegramID != 0 {
		// Only telegram_id provided
		query = `
			SELECT EXISTS(
				SELECT 1 FROM admins
				WHERE telegram_id = ?
			)
		`
		args = []any{telegramID}
	} else {
		// Neither provided
		return false, nil
	}

	var exists bool
	err := r.db.QueryRow(query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check admin: %w", err)
	}

	return exists, nil
}

// Count counts total admins
func (r *AdminRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM admins").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count admins: %w", err)
	}
	return count, nil
}

// Delete deletes an admin by phone number
func (r *AdminRepository) Delete(phoneNumber string) error {
	query := `DELETE FROM admins WHERE phone_number = ?`
	_, err := r.db.Exec(query, phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to delete admin: %w", err)
	}
	return nil
}
