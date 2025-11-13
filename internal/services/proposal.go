package services

import (
	"fmt"

	"parent-bot/internal/models"
	"parent-bot/internal/repository"
)

// ProposalService handles proposal-related business logic
type ProposalService struct {
	repo     *repository.ProposalRepository
	userRepo *repository.UserRepository
}

// NewProposalService creates a new proposal service
func NewProposalService(repo *repository.ProposalRepository, userRepo *repository.UserRepository) *ProposalService {
	return &ProposalService{
		repo:     repo,
		userRepo: userRepo,
	}
}

// CreateProposal creates a new proposal
func (s *ProposalService) CreateProposal(req *models.CreateProposalRequest) (*models.Proposal, error) {
	// Create proposal directly - user validation already done in handler
	proposal, err := s.repo.Create(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create proposal: %w", err)
	}

	return proposal, nil
}

// GetProposalByID gets proposal by ID
func (s *ProposalService) GetProposalByID(id int) (*models.Proposal, error) {
	proposal, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal: %w", err)
	}

	return proposal, nil
}

// GetUserProposals gets proposals by user ID
func (s *ProposalService) GetUserProposals(userID int, limit, offset int) ([]*models.Proposal, error) {
	proposals, err := s.repo.GetByUserID(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user proposals: %w", err)
	}

	return proposals, nil
}

// GetAllProposals gets all proposals with pagination
func (s *ProposalService) GetAllProposals(limit, offset int) ([]*models.Proposal, error) {
	proposals, err := s.repo.GetAll(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposals: %w", err)
	}

	return proposals, nil
}

// GetAllProposalsWithUser gets all proposals with user info
func (s *ProposalService) GetAllProposalsWithUser(limit, offset int) ([]*models.ProposalWithUser, error) {
	proposals, err := s.repo.GetAllWithUser(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposals with user: %w", err)
	}

	return proposals, nil
}

// GetProposalsByStatus gets proposals by status
func (s *ProposalService) GetProposalsByStatus(status string, limit, offset int) ([]*models.Proposal, error) {
	proposals, err := s.repo.GetByStatus(status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposals by status: %w", err)
	}

	return proposals, nil
}

// UpdateProposalStatus updates proposal status
func (s *ProposalService) UpdateProposalStatus(id int, status string) error {
	// Validate status (using same statuses as complaints)
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
		return fmt.Errorf("failed to update proposal status: %w", err)
	}

	return nil
}

// CountProposals counts total proposals
func (s *ProposalService) CountProposals() (int, error) {
	count, err := s.repo.Count()
	if err != nil {
		return 0, fmt.Errorf("failed to count proposals: %w", err)
	}

	return count, nil
}

// CountProposalsByStatus counts proposals by status
func (s *ProposalService) CountProposalsByStatus(status string) (int, error) {
	count, err := s.repo.CountByStatus(status)
	if err != nil {
		return 0, fmt.Errorf("failed to count proposals by status: %w", err)
	}

	return count, nil
}

// CountUserProposals counts proposals by user ID
func (s *ProposalService) CountUserProposals(userID int) (int, error) {
	count, err := s.repo.CountByUserID(userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count user proposals: %w", err)
	}

	return count, nil
}
