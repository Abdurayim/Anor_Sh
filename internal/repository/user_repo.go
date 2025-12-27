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

// Create creates a new user (parent)
func (r *UserRepository) Create(req *models.CreateUserRequest) (*models.User, error) {
	query := `
		INSERT INTO users (telegram_id, telegram_username, phone_number, language)
		VALUES (?, ?, ?, ?)
	`

	result, err := r.db.Exec(
		query,
		req.TelegramID,
		req.TelegramUsername,
		req.PhoneNumber,
		req.Language,
	)
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
	query := `
		SELECT id, telegram_id, telegram_username, phone_number, language, registered_at
		FROM users
		WHERE telegram_id = ?
	`

	var user models.User
	err := r.db.QueryRow(query, telegramID).Scan(
		&user.ID,
		&user.TelegramID,
		&user.TelegramUsername,
		&user.PhoneNumber,
		&user.Language,
		&user.RegisteredAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByPhoneNumber gets user by phone number (indexed, fast query)
func (r *UserRepository) GetByPhoneNumber(phoneNumber string) (*models.User, error) {
	query := `
		SELECT id, telegram_id, telegram_username, phone_number, language, registered_at
		FROM users
		WHERE phone_number = ?
	`

	var user models.User
	err := r.db.QueryRow(query, phoneNumber).Scan(
		&user.ID,
		&user.TelegramID,
		&user.TelegramUsername,
		&user.PhoneNumber,
		&user.Language,
		&user.RegisteredAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByPhone is an alias for GetByPhoneNumber
func (r *UserRepository) GetByPhone(phoneNumber string) (*models.User, error) {
	return r.GetByPhoneNumber(phoneNumber)
}

// GetByID gets user by ID
func (r *UserRepository) GetByID(id int) (*models.User, error) {
	query := `
		SELECT id, telegram_id, telegram_username, phone_number, language, registered_at
		FROM users
		WHERE id = ?
	`

	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.TelegramID,
		&user.TelegramUsername,
		&user.PhoneNumber,
		&user.Language,
		&user.RegisteredAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetAll gets all users with pagination
func (r *UserRepository) GetAll(limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, telegram_id, telegram_username, phone_number, language,
 registered_at
		FROM users
		ORDER BY registered_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.TelegramID,
			&user.TelegramUsername,
			&user.PhoneNumber,
			&user.Language,
				&user.RegisteredAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// GetParentsByClassID gets all parents who have children in a specific class
func (r *UserRepository) GetParentsByClassID(classID int) ([]*models.User, error) {
	query := `
		SELECT DISTINCT u.id, u.telegram_id, u.telegram_username, u.phone_number,
		       u.language, u.registered_at
		FROM users u
		INNER JOIN parent_students ps ON u.id = ps.parent_id
		INNER JOIN students s ON ps.student_id = s.id
		WHERE s.class_id = ? AND s.is_active = 1
		ORDER BY u.registered_at DESC
	`

	rows, err := r.db.Query(query, classID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parents by class: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.TelegramID,
			&user.TelegramUsername,
			&user.PhoneNumber,
			&user.Language,
			&user.RegisteredAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// GetParentsByClassIDs gets all parents who have children in any of the specified classes
func (r *UserRepository) GetParentsByClassIDs(classIDs []int) ([]*models.User, error) {
	if len(classIDs) == 0 {
		return []*models.User{}, nil
	}

	// Build query with placeholders
	query := fmt.Sprintf(`
		SELECT DISTINCT u.id, u.telegram_id, u.telegram_username, u.phone_number,
		       u.language, u.registered_at
		FROM users u
		INNER JOIN parent_students ps ON u.id = ps.parent_id
		INNER JOIN students s ON ps.student_id = s.id
		WHERE s.class_id IN (?%s) AND s.is_active = 1
		ORDER BY u.registered_at DESC
	`, buildPlaceholders(len(classIDs)-1))

	// Convert classIDs to interface slice
	args := make([]interface{}, len(classIDs))
	for i, id := range classIDs {
		args[i] = id
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get parents by classes: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.TelegramID,
			&user.TelegramUsername,
			&user.PhoneNumber,
			&user.Language,
			&user.RegisteredAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	return users, nil
}

// Update updates user data
func (r *UserRepository) Update(userID int, req *models.UpdateUserRequest) error {
	query := `
		UPDATE users
		SET language = COALESCE(?, language)
		WHERE id = ?
	`

	_, err := r.db.Exec(query, req.Language, userID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Count counts total users
func (r *UserRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
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

