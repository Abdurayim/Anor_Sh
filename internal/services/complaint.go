package services

import (
	"fmt"

	"parent-bot/internal/models"
	"parent-bot/internal/repository"
)

// ComplaintService handles complaint-related business logic
type ComplaintService struct {
	repo     *repository.ComplaintRepository
	userRepo *repository.UserRepository
}

// NewComplaintService creates a new complaint service
func NewComplaintService(repo *repository.ComplaintRepository, userRepo *repository.UserRepository) *ComplaintService {
	return &ComplaintService{
		repo:     repo,
		userRepo: userRepo,
	}
}

// CreateComplaint creates a new complaint
func (s *ComplaintService) CreateComplaint(req *models.CreateComplaintRequest) (*models.Complaint, error) {
	// Create complaint directly - user validation already done in handler
	complaint, err := s.repo.Create(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create complaint: %w", err)
	}

	return complaint, nil
}

// GetComplaintByID gets complaint by ID
func (s *ComplaintService) GetComplaintByID(id int) (*models.Complaint, error) {
	complaint, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get complaint: %w", err)
	}

	return complaint, nil
}

// GetUserComplaints gets complaints by user ID
func (s *ComplaintService) GetUserComplaints(userID int, limit, offset int) ([]*models.Complaint, error) {
	complaints, err := s.repo.GetByUserID(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user complaints: %w", err)
	}

	return complaints, nil
}

// GetAllComplaints gets all complaints with pagination
func (s *ComplaintService) GetAllComplaints(limit, offset int) ([]*models.Complaint, error) {
	complaints, err := s.repo.GetAll(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get complaints: %w", err)
	}

	return complaints, nil
}

// GetAllComplaintsWithUser gets all complaints with user info
func (s *ComplaintService) GetAllComplaintsWithUser(limit, offset int) ([]*models.ComplaintWithUser, error) {
	complaints, err := s.repo.GetAllWithUser(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get complaints with user: %w", err)
	}

	return complaints, nil
}

// GetComplaintsByStatus gets complaints by status
func (s *ComplaintService) GetComplaintsByStatus(status string, limit, offset int) ([]*models.Complaint, error) {
	complaints, err := s.repo.GetByStatus(status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get complaints by status: %w", err)
	}

	return complaints, nil
}

// UpdateComplaintStatus updates complaint status
func (s *ComplaintService) UpdateComplaintStatus(id int, status string) error {
	// Validate status
	validStatuses := map[string]bool{
		models.StatusPending:  true,
		models.StatusReviewed: true,
		models.StatusArchived: true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	err := s.repo.UpdateStatus(id, status)
	if err != nil {
		return fmt.Errorf("failed to update complaint status: %w", err)
	}

	return nil
}

// CountComplaints counts total complaints
func (s *ComplaintService) CountComplaints() (int, error) {
	count, err := s.repo.Count()
	if err != nil {
		return 0, fmt.Errorf("failed to count complaints: %w", err)
	}

	return count, nil
}

// CountComplaintsByStatus counts complaints by status
func (s *ComplaintService) CountComplaintsByStatus(status string) (int, error) {
	count, err := s.repo.CountByStatus(status)
	if err != nil {
		return 0, fmt.Errorf("failed to count complaints by status: %w", err)
	}

	return count, nil
}

// CountUserComplaints counts complaints by user ID
func (s *ComplaintService) CountUserComplaints(userID int) (int, error) {
	count, err := s.repo.CountByUserID(userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count user complaints: %w", err)
	}

	return count, nil
}
