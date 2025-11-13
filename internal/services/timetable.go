package services

import (
	"fmt"

	"parent-bot/internal/models"
	"parent-bot/internal/repository"
)

// TimetableService handles timetable-related business logic
type TimetableService struct {
	repo      *repository.TimetableRepository
	classRepo *repository.ClassRepository
}

// NewTimetableService creates a new timetable service
func NewTimetableService(repo *repository.TimetableRepository, classRepo *repository.ClassRepository) *TimetableService {
	return &TimetableService{
		repo:      repo,
		classRepo: classRepo,
	}
}

// CreateTimetable creates a new timetable
func (s *TimetableService) CreateTimetable(req *models.CreateTimetableRequest) (*models.Timetable, error) {
	// Validate that class exists
	class, err := s.classRepo.GetByID(req.ClassID)
	if err != nil {
		return nil, fmt.Errorf("failed to get class: %w", err)
	}
	if class == nil {
		return nil, fmt.Errorf("class not found")
	}

	timetable, err := s.repo.Create(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create timetable: %w", err)
	}

	return timetable, nil
}

// GetTimetableByID gets timetable by ID
func (s *TimetableService) GetTimetableByID(id int) (*models.Timetable, error) {
	timetable, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get timetable: %w", err)
	}

	return timetable, nil
}

// GetTimetableByClassID gets the latest timetable for a class
func (s *TimetableService) GetTimetableByClassID(classID int) (*models.Timetable, error) {
	timetable, err := s.repo.GetByClassID(classID)
	if err != nil {
		return nil, fmt.Errorf("failed to get timetable by class: %w", err)
	}

	return timetable, nil
}

// GetTimetableByClassName gets the latest timetable for a class by name
func (s *TimetableService) GetTimetableByClassName(className string) (*models.Timetable, error) {
	timetable, err := s.repo.GetByClassName(className)
	if err != nil {
		return nil, fmt.Errorf("failed to get timetable by class name: %w", err)
	}

	return timetable, nil
}

// GetAllTimetables gets all timetables with pagination
func (s *TimetableService) GetAllTimetables(limit, offset int) ([]*models.Timetable, error) {
	timetables, err := s.repo.GetAll(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get timetables: %w", err)
	}

	return timetables, nil
}

// UpdateTimetable updates an existing timetable
func (s *TimetableService) UpdateTimetable(id int, req *models.CreateTimetableRequest) (*models.Timetable, error) {
	timetable, err := s.repo.Update(id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update timetable: %w", err)
	}

	return timetable, nil
}

// DeleteTimetable deletes a timetable
func (s *TimetableService) DeleteTimetable(id int) error {
	err := s.repo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete timetable: %w", err)
	}

	return nil
}

// DeleteTimetablesByClassID deletes all timetables for a class
func (s *TimetableService) DeleteTimetablesByClassID(classID int) error {
	err := s.repo.DeleteByClassID(classID)
	if err != nil {
		return fmt.Errorf("failed to delete timetables for class: %w", err)
	}

	return nil
}

// CountTimetables counts total timetables
func (s *TimetableService) CountTimetables() (int, error) {
	count, err := s.repo.Count()
	if err != nil {
		return 0, fmt.Errorf("failed to count timetables: %w", err)
	}

	return count, nil
}

// TimetableExists checks if a timetable exists for a class
func (s *TimetableService) TimetableExists(classID int) (bool, error) {
	exists, err := s.repo.Exists(classID)
	if err != nil {
		return false, fmt.Errorf("failed to check timetable existence: %w", err)
	}

	return exists, nil
}
