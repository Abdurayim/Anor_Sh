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

// Create creates a new user
func (r *UserRepository) Create(req *models.CreateUserRequest) (*models.User, error) {
	query := `
		INSERT INTO users (telegram_id, telegram_username, phone_number, child_name, child_class, language)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, telegram_id, telegram_username, phone_number, child_name, child_class, language, registered_at
	`

	var user models.User
	err := r.db.QueryRow(
		query,
		req.TelegramID,
		req.TelegramUsername,
		req.PhoneNumber,
		req.ChildName,
		req.ChildClass,
		req.Language,
	).Scan(
		&user.ID,
		&user.TelegramID,
		&user.TelegramUsername,
		&user.PhoneNumber,
		&user.ChildName,
		&user.ChildClass,
		&user.Language,
		&user.RegisteredAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// GetByTelegramID gets user by telegram ID (indexed, fast query)
func (r *UserRepository) GetByTelegramID(telegramID int64) (*models.User, error) {
	query := `
		SELECT id, telegram_id, telegram_username, phone_number, child_name, child_class, language, registered_at
		FROM users
		WHERE telegram_id = $1
	`

	var user models.User
	err := r.db.QueryRow(query, telegramID).Scan(
		&user.ID,
		&user.TelegramID,
		&user.TelegramUsername,
		&user.PhoneNumber,
		&user.ChildName,
		&user.ChildClass,
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
		SELECT id, telegram_id, telegram_username, phone_number, child_name, child_class, language, registered_at
		FROM users
		WHERE phone_number = $1
	`

	var user models.User
	err := r.db.QueryRow(query, phoneNumber).Scan(
		&user.ID,
		&user.TelegramID,
		&user.TelegramUsername,
		&user.PhoneNumber,
		&user.ChildName,
		&user.ChildClass,
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

// GetByID gets user by ID
func (r *UserRepository) GetByID(id int) (*models.User, error) {
	query := `
		SELECT id, telegram_id, telegram_username, phone_number, child_name, child_class, language, registered_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.TelegramID,
		&user.TelegramUsername,
		&user.PhoneNumber,
		&user.ChildName,
		&user.ChildClass,
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
		SELECT id, telegram_id, telegram_username, phone_number, child_name, child_class, language, registered_at
		FROM users
		ORDER BY registered_at DESC
		LIMIT $1 OFFSET $2
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
			&user.ChildName,
			&user.ChildClass,
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

// GetByClass gets users by class (indexed, fast query)
func (r *UserRepository) GetByClass(class string) ([]*models.User, error) {
	query := `
		SELECT id, telegram_id, telegram_username, phone_number, child_name, child_class, language, registered_at
		FROM users
		WHERE child_class = $1
		ORDER BY registered_at DESC
	`

	rows, err := r.db.Query(query, class)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by class: %w", err)
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
			&user.ChildName,
			&user.ChildClass,
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
func (r *UserRepository) Update(telegramID int64, req *models.UpdateUserRequest) error {
	query := `
		UPDATE users
		SET child_name = COALESCE(NULLIF($1, ''), child_name),
		    child_class = COALESCE(NULLIF($2, ''), child_class),
		    language = COALESCE(NULLIF($3, ''), language)
		WHERE telegram_id = $4
	`

	_, err := r.db.Exec(query, req.ChildName, req.ChildClass, req.Language, telegramID)
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
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE telegram_id = $1)`
	err := r.db.QueryRow(query, telegramID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}
