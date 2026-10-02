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
	return &ComplaintService{repo: repo, userRepo: userRepo}
}

// CreateComplaint creates a new complaint
func (s *ComplaintService) CreateComplaint(req *models.CreateComplaintRequest) (*models.Complaint, error) {
	return s.repo.Create(req)
}

// SetComplaintFile stores the Telegram file of the generated document
func (s *ComplaintService) SetComplaintFile(id int, telegramFileID, filename string) error {
	return s.repo.SetFile(id, telegramFileID, filename)
}

// GetComplaintByID gets complaint by ID
func (s *ComplaintService) GetComplaintByID(id int) (*models.Complaint, error) {
	return s.repo.GetByID(id)
}

// GetUserComplaints gets complaints of a parent
func (s *ComplaintService) GetUserComplaints(userID int, limit, offset int) ([]*models.Complaint, error) {
	return s.repo.GetByUserID(userID, limit, offset)
}

// GetAllComplaints gets complaints of a branch (branchID 0 = all branches)
func (s *ComplaintService) GetAllComplaints(branchID, limit, offset int) ([]*models.Complaint, error) {
	return s.repo.GetAll(branchID, limit, offset)
}

// GetAllComplaintsWithUser gets complaints of a branch with parent info (branchID 0 = all branches)
func (s *ComplaintService) GetAllComplaintsWithUser(branchID, limit, offset int) ([]*models.ComplaintWithUser, error) {
	return s.repo.GetAllWithUser(branchID, limit, offset)
}

// UpdateComplaintStatus updates complaint status
func (s *ComplaintService) UpdateComplaintStatus(id int, status string) error {
	valid := map[string]bool{models.StatusPending: true, models.StatusReviewed: true, models.StatusResolved: true}
	if !valid[status] {
		return fmt.Errorf("invalid status: %s", status)
	}
	return s.repo.UpdateStatus(id, status)
}

// CountComplaints counts complaints of a branch (branchID 0 = all branches)
func (s *ComplaintService) CountComplaints(branchID int) (int, error) {
	return s.repo.Count(branchID, "")
}

// CountComplaintsByStatus counts complaints of a branch with the given status
func (s *ComplaintService) CountComplaintsByStatus(branchID int, status string) (int, error) {
	return s.repo.Count(branchID, status)
}

// CountUserComplaints counts complaints of a parent
func (s *ComplaintService) CountUserComplaints(userID int) (int, error) {
	return s.repo.CountByUserID(userID)
}
