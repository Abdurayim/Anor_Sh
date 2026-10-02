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
	return &ProposalService{repo: repo, userRepo: userRepo}
}

// CreateProposal creates a new proposal
func (s *ProposalService) CreateProposal(req *models.CreateProposalRequest) (*models.Proposal, error) {
	return s.repo.Create(req)
}

// SetProposalFile stores the Telegram file of the generated document
func (s *ProposalService) SetProposalFile(id int, telegramFileID, filename string) error {
	return s.repo.SetFile(id, telegramFileID, filename)
}

// GetProposalByID gets proposal by ID
func (s *ProposalService) GetProposalByID(id int) (*models.Proposal, error) {
	return s.repo.GetByID(id)
}

// GetUserProposals gets proposals of a parent
func (s *ProposalService) GetUserProposals(userID int, limit, offset int) ([]*models.Proposal, error) {
	return s.repo.GetByUserID(userID, limit, offset)
}

// GetAllProposals gets proposals of a branch (branchID 0 = all branches)
func (s *ProposalService) GetAllProposals(branchID, limit, offset int) ([]*models.Proposal, error) {
	return s.repo.GetAll(branchID, limit, offset)
}

// GetAllProposalsWithUser gets proposals of a branch with parent info (branchID 0 = all branches)
func (s *ProposalService) GetAllProposalsWithUser(branchID, limit, offset int) ([]*models.ProposalWithUser, error) {
	return s.repo.GetAllWithUser(branchID, limit, offset)
}

// UpdateProposalStatus updates proposal status
func (s *ProposalService) UpdateProposalStatus(id int, status string) error {
	valid := map[string]bool{models.StatusPending: true, models.StatusReviewed: true, models.StatusImplemented: true}
	if !valid[status] {
		return fmt.Errorf("invalid status: %s", status)
	}
	return s.repo.UpdateStatus(id, status)
}

// CountProposals counts proposals of a branch (branchID 0 = all branches)
func (s *ProposalService) CountProposals(branchID int) (int, error) {
	return s.repo.Count(branchID, "")
}

// CountProposalsByStatus counts proposals of a branch with the given status
func (s *ProposalService) CountProposalsByStatus(branchID int, status string) (int, error) {
	return s.repo.Count(branchID, status)
}

// CountUserProposals counts proposals of a parent
func (s *ProposalService) CountUserProposals(userID int) (int, error) {
	return s.repo.CountByUserID(userID)
}
