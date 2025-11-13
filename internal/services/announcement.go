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

// GetActiveAnnouncements gets all active announcements with pagination
func (s *AnnouncementService) GetActiveAnnouncements(limit, offset int) ([]*models.Announcement, error) {
	announcements, err := s.repo.GetActive(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get active announcements: %w", err)
	}

	return announcements, nil
}

// GetAllAnnouncements gets all announcements with pagination (for admin)
func (s *AnnouncementService) GetAllAnnouncements(limit, offset int) ([]*models.Announcement, error) {
	announcements, err := s.repo.GetAll(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get announcements: %w", err)
	}

	return announcements, nil
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

// CountAnnouncements counts total announcements
func (s *AnnouncementService) CountAnnouncements() (int, error) {
	count, err := s.repo.Count()
	if err != nil {
		return 0, fmt.Errorf("failed to count announcements: %w", err)
	}

	return count, nil
}

// CountActiveAnnouncements counts active announcements
func (s *AnnouncementService) CountActiveAnnouncements() (int, error) {
	count, err := s.repo.CountActive()
	if err != nil {
		return 0, fmt.Errorf("failed to count active announcements: %w", err)
	}

	return count, nil
}
