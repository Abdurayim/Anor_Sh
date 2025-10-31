package services

import (
	"fmt"

	"parent-bot/internal/models"
	"parent-bot/internal/repository"
)

// UserService handles user-related business logic
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(req *models.CreateUserRequest) (*models.User, error) {
	// Check if user already exists
	existing, err := s.repo.GetByTelegramID(req.TelegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existing != nil {
		return nil, fmt.Errorf("user already registered")
	}

	// Check if phone number is already used
	existingPhone, err := s.repo.GetByPhoneNumber(req.PhoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing phone: %w", err)
	}

	if existingPhone != nil {
		return nil, fmt.Errorf("phone number already registered")
	}

	// Create user
	user, err := s.repo.Create(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByTelegramID gets user by telegram ID
func (s *UserService) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	user, err := s.repo.GetByTelegramID(telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByPhoneNumber gets user by phone number
func (s *UserService) GetUserByPhoneNumber(phoneNumber string) (*models.User, error) {
	user, err := s.repo.GetByPhoneNumber(phoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(telegramID int64, req *models.UpdateUserRequest) error {
	err := s.repo.Update(telegramID, req)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// GetAllUsers gets all users with pagination
func (s *UserService) GetAllUsers(limit, offset int) ([]*models.User, error) {
	users, err := s.repo.GetAll(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	return users, nil
}

// GetUsersByClass gets users by class
func (s *UserService) GetUsersByClass(class string) ([]*models.User, error) {
	users, err := s.repo.GetByClass(class)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by class: %w", err)
	}

	return users, nil
}

// CountUsers counts total users
func (s *UserService) CountUsers() (int, error) {
	count, err := s.repo.Count()
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// IsUserRegistered checks if user is registered
func (s *UserService) IsUserRegistered(telegramID int64) (bool, error) {
	exists, err := s.repo.Exists(telegramID)
	if err != nil {
		return false, fmt.Errorf("failed to check user registration: %w", err)
	}

	return exists, nil
}
