package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"parent-bot/internal/models"
)

// ErrAdminExists is returned when the phone already belongs to an active admin.
var ErrAdminExists = errors.New("admin already exists")

// AdminRepository manages admins. A branch admin (role "admin") belongs to one branch;
// the super admin (role "super_admin") has no branch. Only active admins count.
type AdminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

const adminColumns = "id, phone_number, telegram_id, COALESCE(name, ''), COALESCE(branch_id, 0), role, source, added_at"

func scanAdmin(row interface{ Scan(...any) error }) (*models.Admin, error) {
	var admin models.Admin
	err := row.Scan(&admin.ID, &admin.PhoneNumber, &admin.TelegramID, &admin.Name, &admin.BranchID,
		&admin.Role, &admin.Source, &admin.AddedAt)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) getOne(where string, args ...any) (*models.Admin, error) {
	admin, err := scanAdmin(r.db.QueryRow("SELECT "+adminColumns+" FROM admins WHERE is_active = 1 AND "+where, args...))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}
	return admin, nil
}

// UpsertFromEnv makes a phone configured in .env an active admin with the given role
// (branchID 0 for the super admin).
func (r *AdminRepository) UpsertFromEnv(phoneNumber, name, role string, branchID int) error {
	_, err := r.db.Exec(`
		INSERT INTO admins (phone_number, name, role, branch_id, source, is_active)
		VALUES (?, ?, ?, ?, 'env', 1)
		ON CONFLICT(phone_number) DO UPDATE SET
			role = excluded.role,
			branch_id = excluded.branch_id,
			source = 'env',
			is_active = 1,
			updated_at = CURRENT_TIMESTAMP
	`, phoneNumber, name, role, nullableID(branchID))
	if err != nil {
		return fmt.Errorf("failed to upsert admin: %w", err)
	}
	return nil
}

// CreateBranchAdmin adds a branch admin from the bot. A previously removed admin with the
// same phone is reactivated; an active admin returns ErrAdminExists.
func (r *AdminRepository) CreateBranchAdmin(phoneNumber, name string, branchID, addedByAdminID int) error {
	result, err := r.db.Exec(`
		INSERT INTO admins (phone_number, name, role, branch_id, source, added_by_admin_id, is_active)
		VALUES (?, ?, 'admin', ?, 'bot', ?, 1)
		ON CONFLICT(phone_number) DO UPDATE SET
			name = excluded.name,
			role = 'admin',
			branch_id = excluded.branch_id,
			source = 'bot',
			added_by_admin_id = excluded.added_by_admin_id,
			is_active = 1,
			updated_at = CURRENT_TIMESTAMP
		WHERE admins.is_active = 0
	`, phoneNumber, name, branchID, nullableID(addedByAdminID))
	if err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return ErrAdminExists
	}
	return nil
}

// DeactivateEnvAdminsExcept deactivates .env admins whose phone is no longer configured
// (an admin removed from .env loses access on the next start). Admins added in the bot stay.
func (r *AdminRepository) DeactivateEnvAdminsExcept(phones []string) error {
	query := "UPDATE admins SET is_active = 0, updated_at = CURRENT_TIMESTAMP WHERE is_active = 1 AND source = 'env'"
	args := make([]any, len(phones))
	if len(phones) > 0 {
		query += fmt.Sprintf(" AND phone_number NOT IN (?%s)", buildPlaceholders(len(phones)-1))
		for i, p := range phones {
			args[i] = p
		}
	}
	if _, err := r.db.Exec(query, args...); err != nil {
		return fmt.Errorf("failed to deactivate admins: %w", err)
	}
	return nil
}

// Deactivate removes admin rights (the row is kept for history)
func (r *AdminRepository) Deactivate(id int) error {
	_, err := r.db.Exec("UPDATE admins SET is_active = 0, telegram_id = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to deactivate admin: %w", err)
	}
	return nil
}

// GetByID gets an active admin by ID
func (r *AdminRepository) GetByID(id int) (*models.Admin, error) {
	return r.getOne("id = ?", id)
}

// GetByPhoneNumber gets an active admin (any role) by phone number
func (r *AdminRepository) GetByPhoneNumber(phoneNumber string) (*models.Admin, error) {
	return r.getOne("phone_number = ?", phoneNumber)
}

// GetByTelegramID gets an active admin (any role) by telegram ID
func (r *AdminRepository) GetByTelegramID(telegramID int64) (*models.Admin, error) {
	return r.getOne("telegram_id = ?", telegramID)
}

// GetAll gets active branch admins of a branch (branchID 0 = all branches); the super admin is not included
func (r *AdminRepository) GetAll(branchID int) ([]*models.Admin, error) {
	rows, err := r.db.Query(`
		SELECT `+adminColumns+`
		FROM admins
		WHERE is_active = 1 AND role = 'admin' AND (? = 0 OR branch_id = ?)
		ORDER BY branch_id, added_at ASC
	`, branchID, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get admins: %w", err)
	}
	defer rows.Close()

	var admins []*models.Admin
	for rows.Next() {
		admin, err := scanAdmin(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan admin: %w", err)
		}
		admins = append(admins, admin)
	}
	return admins, rows.Err()
}

// UpdateTelegramID links a telegram account to an admin phone. Any other admin row
// holding the same telegram ID is unlinked first (telegram_id is UNIQUE).
func (r *AdminRepository) UpdateTelegramID(phoneNumber string, telegramID int64) error {
	if _, err := r.db.Exec("UPDATE admins SET telegram_id = NULL WHERE telegram_id = ? AND phone_number <> ?", telegramID, phoneNumber); err != nil {
		return fmt.Errorf("failed to update admin telegram ID: %w", err)
	}
	if _, err := r.db.Exec("UPDATE admins SET telegram_id = ? WHERE phone_number = ?", telegramID, phoneNumber); err != nil {
		return fmt.Errorf("failed to update admin telegram ID: %w", err)
	}
	return nil
}

// Count counts active branch admins of a branch (branchID 0 = all branches)
func (r *AdminRepository) Count(branchID int) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM admins WHERE is_active = 1 AND role = 'admin' AND (? = 0 OR branch_id = ?)",
		branchID, branchID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count admins: %w", err)
	}
	return count, nil
}
