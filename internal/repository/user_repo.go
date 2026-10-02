package repository

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

const userColumns = `u.id, u.telegram_id, COALESCE(u.telegram_username, ''), u.phone_number, u.language,
	COALESCE(u.branch_id, 0), u.registered_at`

func scanUser(row interface{ Scan(...any) error }) (*models.User, error) {
	var user models.User
	err := row.Scan(
		&user.ID,
		&user.TelegramID,
		&user.TelegramUsername,
		&user.PhoneNumber,
		&user.Language,
		&user.BranchID,
		&user.RegisteredAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) getOne(where string, arg any) (*models.User, error) {
	user, err := scanUser(r.db.QueryRow("SELECT "+userColumns+" FROM users u WHERE "+where, arg))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) getMany(query string, args ...any) ([]*models.User, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// Create creates a new user (parent)
func (r *UserRepository) Create(req *models.CreateUserRequest) (*models.User, error) {
	result, err := r.db.Exec(`
		INSERT INTO users (telegram_id, telegram_username, phone_number, language, branch_id)
		VALUES (?, ?, ?, ?, ?)
	`, req.TelegramID, req.TelegramUsername, req.PhoneNumber, req.Language, nullableID(req.BranchID))
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return r.GetByID(int(id))
}

// GetByTelegramID gets user by telegram ID (indexed, fast query)
func (r *UserRepository) GetByTelegramID(telegramID int64) (*models.User, error) {
	return r.getOne("u.telegram_id = ?", telegramID)
}

// GetByPhoneNumber gets user by phone number (indexed, fast query)
func (r *UserRepository) GetByPhoneNumber(phoneNumber string) (*models.User, error) {
	return r.getOne("u.phone_number = ?", phoneNumber)
}

// GetByPhone is an alias for GetByPhoneNumber
func (r *UserRepository) GetByPhone(phoneNumber string) (*models.User, error) {
	return r.GetByPhoneNumber(phoneNumber)
}

// GetByID gets user by ID
func (r *UserRepository) GetByID(id int) (*models.User, error) {
	return r.getOne("u.id = ?", id)
}

// GetAll gets users of a branch with pagination (branchID 0 = all branches)
func (r *UserRepository) GetAll(branchID, limit, offset int) ([]*models.User, error) {
	return r.getMany(`
		SELECT `+userColumns+`
		FROM users u
		WHERE (? = 0 OR u.branch_id = ?)
		ORDER BY u.registered_at DESC
		LIMIT ? OFFSET ?
	`, branchID, branchID, limit, offset)
}

// GetParentsByClassID gets all parents who have children in a specific class
func (r *UserRepository) GetParentsByClassID(classID int) ([]*models.User, error) {
	return r.GetParentsByClassIDs([]int{classID})
}

// GetParentsByClassIDs gets all parents who have children in any of the specified classes
func (r *UserRepository) GetParentsByClassIDs(classIDs []int) ([]*models.User, error) {
	if len(classIDs) == 0 {
		return []*models.User{}, nil
	}

	query := fmt.Sprintf(`
		SELECT `+userColumns+`
		FROM users u
		WHERE u.id IN (
			SELECT ps.parent_id
			FROM parent_students ps
			INNER JOIN students s ON ps.student_id = s.id
			WHERE s.class_id IN (?%s) AND s.is_active = 1
		)
		ORDER BY u.registered_at DESC
	`, buildPlaceholders(len(classIDs)-1))

	args := make([]any, len(classIDs))
	for i, id := range classIDs {
		args[i] = id
	}
	return r.getMany(query, args...)
}

// Update updates user data; empty fields are left unchanged
func (r *UserRepository) Update(userID int, req *models.UpdateUserRequest) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET language = COALESCE(NULLIF(?, ''), language)
		WHERE id = ?
	`, req.Language, userID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// SetBranch moves a parent to a branch
func (r *UserRepository) SetBranch(userID, branchID int) error {
	if _, err := r.db.Exec("UPDATE users SET branch_id = ? WHERE id = ?", branchID, userID); err != nil {
		return fmt.Errorf("failed to set user branch: %w", err)
	}
	return nil
}

// Count counts users of a branch (branchID 0 = all branches)
func (r *UserRepository) Count(branchID int) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE (? = 0 OR branch_id = ?)", branchID, branchID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}

// Exists checks if user exists by telegram ID
func (r *UserRepository) Exists(telegramID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE telegram_id = ?)`
	err := r.db.QueryRow(query, telegramID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}

// Delete deletes a user
func (r *UserRepository) Delete(id int) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// nullableID stores 0 as NULL.
func nullableID(id int) any {
	if id == 0 {
		return nil
	}
	return id
}
