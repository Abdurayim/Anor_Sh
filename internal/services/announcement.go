package services

import (
	"fmt"

	"parent-bot/internal/models"
	"parent-bot/internal/repository"
)

// AnnouncementService handles announcement-related business logic
type AnnouncementService struct {
	repo *repository.AnnouncementRepository
}

// NewAnnouncementService creates a new announcement service
func NewAnnouncementService(repo *repository.AnnouncementRepository) *AnnouncementService {
	return &AnnouncementService{
		repo: repo,
	}
}

// CreateAnnouncement creates a new announcement
func (s *AnnouncementService) CreateAnnouncement(req *models.CreateAnnouncementRequest) (*models.Announcement, error) {
	announcement, err := s.repo.Create(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create announcement: %w", err)
	}

	return announcement, nil
}

// GetAnnouncementByID gets announcement by ID
func (s *AnnouncementService) GetAnnouncementByID(id int) (*models.Announcement, error) {
	announcement, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get announcement: %w", err)
	}

	return announcement, nil
}

// GetActiveAnnouncements gets active announcements of a branch (branchID 0 = all branches)
func (s *AnnouncementService) GetActiveAnnouncements(branchID, limit, offset int) ([]*models.Announcement, error) {
	return s.repo.GetActive(branchID, limit, offset)
}

// GetActiveAnnouncementsForParent gets the announcements a parent should see
func (s *AnnouncementService) GetActiveAnnouncementsForParent(parentID, branchID, limit, offset int) ([]*models.Announcement, error) {
	return s.repo.GetActiveForParent(parentID, branchID, limit, offset)
}

// GetAllAnnouncements gets announcements of a branch (branchID 0 = all branches)
func (s *AnnouncementService) GetAllAnnouncements(branchID, limit, offset int) ([]*models.Announcement, error) {
	return s.repo.GetAll(branchID, limit, offset)
}

// GetAnnouncementClassIDs returns target classes of an announcement (empty = whole branch)
func (s *AnnouncementService) GetAnnouncementClassIDs(id int) ([]int, error) {
	return s.repo.GetClassIDs(id)
}

// UpdateAnnouncement updates an existing announcement
func (s *AnnouncementService) UpdateAnnouncement(id int, req *models.CreateAnnouncementRequest) (*models.Announcement, error) {
	announcement, err := s.repo.Update(id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update announcement: %w", err)
	}

	return announcement, nil
}

// ToggleAnnouncementActive toggles the active status of an announcement
func (s *AnnouncementService) ToggleAnnouncementActive(id int) error {
	err := s.repo.ToggleActive(id)
	if err != nil {
		return fmt.Errorf("failed to toggle announcement active status: %w", err)
	}

	return nil
}

// DeleteAnnouncement deletes an announcement
func (s *AnnouncementService) DeleteAnnouncement(id int) error {
	err := s.repo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete announcement: %w", err)
	}

	return nil
}

// CountAnnouncements counts announcements of a branch (branchID 0 = all branches)
func (s *AnnouncementService) CountAnnouncements(branchID int) (int, error) {
	return s.repo.Count(branchID)
}
