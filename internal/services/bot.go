package services

import (
	"database/sql"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/config"
	"parent-bot/internal/repository"
	"parent-bot/internal/state"
)

// BotService is the main bot service
type BotService struct {
	Bot                 *tgbotapi.BotAPI
	Config              *config.Config
	UserRepo            *repository.UserRepository
	ComplaintRepo       *repository.ComplaintRepository
	ProposalRepo        *repository.ProposalRepository
	TimetableRepo       *repository.TimetableRepository
	AnnouncementRepo    *repository.AnnouncementRepository
	AdminRepo           *repository.AdminRepository
	ClassRepo           *repository.ClassRepository
	StateManager        *state.Manager
	TelegramService     *TelegramService
	UserService         *UserService
	ComplaintService    *ComplaintService
	ProposalService     *ProposalService
	TimetableService    *TimetableService
	AnnouncementService *AnnouncementService
	DocumentService     *DocumentService
}

// NewBotService creates a new bot service
func NewBotService(cfg *config.Config, db *sql.DB) (*BotService, error) {
	// Create bot instance
	bot, err := tgbotapi.NewBotAPI(cfg.Bot.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	complaintRepo := repository.NewComplaintRepository(db)
	proposalRepo := repository.NewProposalRepository(db)
	timetableRepo := repository.NewTimetableRepository(db)
	announcementRepo := repository.NewAnnouncementRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	classRepo := repository.NewClassRepository(db)

	// Initialize state manager
	stateManager := state.NewManager(db)

	// Initialize services
	telegramService := NewTelegramService(bot)
	userService := NewUserService(userRepo)
	complaintService := NewComplaintService(complaintRepo, userRepo)
	proposalService := NewProposalService(proposalRepo, userRepo)
	timetableService := NewTimetableService(timetableRepo, classRepo)
	announcementService := NewAnnouncementService(announcementRepo)
	documentService := NewDocumentService("./temp_docs") // temp directory for generated documents

	return &BotService{
		Bot:                 bot,
		Config:              cfg,
		UserRepo:            userRepo,
		ComplaintRepo:       complaintRepo,
		ProposalRepo:        proposalRepo,
		TimetableRepo:       timetableRepo,
		AnnouncementRepo:    announcementRepo,
		AdminRepo:           adminRepo,
		ClassRepo:           classRepo,
		StateManager:        stateManager,
		TelegramService:     telegramService,
		UserService:         userService,
		ComplaintService:    complaintService,
		ProposalService:     proposalService,
		TimetableService:    timetableService,
		AnnouncementService: announcementService,
		DocumentService:     documentService,
	}, nil
}

// SetWebhook sets up webhook
func (s *BotService) SetWebhook(webhookURL string) error {
	wh, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	_, err = s.Bot.Request(wh)
	if err != nil {
		return fmt.Errorf("failed to set webhook: %w", err)
	}

	return nil
}

// RemoveWebhook removes webhook
func (s *BotService) RemoveWebhook() error {
	_, err := s.Bot.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		return fmt.Errorf("failed to remove webhook: %w", err)
	}

	return nil
}

// InitializeAdmins initializes admins from config
func (s *BotService) InitializeAdmins() error {
	for _, phone := range s.Config.Admin.PhoneNumbers {
		// Check if admin already exists
		admin, err := s.AdminRepo.GetByPhoneNumber(phone)
		if err != nil {
			return fmt.Errorf("failed to check admin: %w", err)
		}

		if admin == nil {
			// Create admin
			_, err = s.AdminRepo.Create(phone, "Admin")
			if err != nil {
				fmt.Printf("Warning: failed to create admin %s: %v\n", phone, err)
			}
		}
	}

	return nil
}

// GetAdminTelegramIDs gets all admin telegram IDs
func (s *BotService) GetAdminTelegramIDs() ([]int64, error) {
	admins, err := s.AdminRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var ids []int64
	for _, admin := range admins {
		if admin.TelegramID != nil {
			ids = append(ids, *admin.TelegramID)
		}
	}

	return ids, nil
}

// IsAdmin checks if user is admin by checking:
// 1. Database admins table (by phone or telegram_id)
// 2. Config admin phones (if user is registered)
// 3. Attempts to link telegram_id if admin phone matches
func (s *BotService) IsAdmin(phoneNumber string, telegramID int64) (bool, error) {
	// First check database
	isAdminInDB, err := s.AdminRepo.IsAdmin(phoneNumber, telegramID)
	if err != nil {
		return false, err
	}

	if isAdminInDB {
		return true, nil
	}

	// If not found in DB, check if user's phone matches config admin phones
	if phoneNumber != "" {
		for _, adminPhone := range s.Config.Admin.PhoneNumbers {
			if phoneNumber == adminPhone {
				// Found admin by phone from config, link telegram_id
				_ = s.AdminRepo.UpdateTelegramID(phoneNumber, telegramID)
				return true, nil
			}
		}
	}

	// If phone is empty, try to get it from user record
	if phoneNumber == "" && telegramID != 0 {
		user, err := s.UserService.GetUserByTelegramID(telegramID)
		if err == nil && user != nil {
			// Check if user's phone is an admin phone
			for _, adminPhone := range s.Config.Admin.PhoneNumbers {
				if user.PhoneNumber == adminPhone {
					// Found admin by phone from config, link telegram_id
					_ = s.AdminRepo.UpdateTelegramID(user.PhoneNumber, telegramID)
					return true, nil
				}
			}
		}
	}

	return false, nil
}
