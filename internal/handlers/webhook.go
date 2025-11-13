package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"parent-bot/internal/services"
)

// HandleUpdate is the main update handler that routes all Telegram updates
func HandleUpdate(botService *services.BotService, update tgbotapi.Update) {
	// Handle callback queries (inline button clicks)
	if update.CallbackQuery != nil {
		if err := HandleCallbackQuery(botService, update.CallbackQuery); err != nil {
			log.Printf("Error handling callback query: %v", err)
		}
		return
	}

	// Handle messages
	if update.Message != nil {
		if err := HandleMessage(botService, update.Message); err != nil {
			log.Printf("Error handling message: %v", err)
		}
		return
	}

	// Handle edited messages (optional)
	if update.EditedMessage != nil {
		log.Printf("Received edited message from %d", update.EditedMessage.From.ID)
		return
	}
}

// HandleMessage routes messages based on type and user state
func HandleMessage(botService *services.BotService, message *tgbotapi.Message) error {
	// Ignore messages from bots
	if message.From.IsBot {
		return nil
	}

	telegramID := message.From.ID

	// Handle commands first
	if message.IsCommand() {
		return HandleCommand(botService, message)
	}

	// Check for critical button presses that should override state (like admin panel)
	// This must be checked BEFORE state routing
	buttonText := message.Text
	if buttonText == "👨‍💼 Ma'muriyat paneli" || buttonText == "👨‍💼 Панель администратора" {
		// Clear any existing state and go to admin panel
		_ = botService.StateManager.Clear(telegramID)
		return HandleAdminCommand(botService, message)
	}

	// Check if user is registered
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// If not registered and not in registration flow, start registration
	if user == nil {
		state, err := botService.StateManager.GetState(telegramID)
		if err != nil {
			return err
		}

		// If no state, start registration
		if state == "" {
			return HandleStart(botService, message)
		}
	}

	// Get user state
	state, err := botService.StateManager.Get(telegramID)
	if err != nil {
		return err
	}

	// Route based on state
	if state != nil {
		stateData, err := botService.StateManager.GetData(telegramID)
		if err != nil {
			return err
		}

		return RouteByState(botService, message, state.State, stateData)
	}

	// Default: handle as registered user message
	return HandleRegisteredUserMessage(botService, message, user)
}

// HandleCommand handles bot commands
func HandleCommand(botService *services.BotService, message *tgbotapi.Message) error {
	switch message.Command() {
	case "start":
		return HandleStart(botService, message)
	case "help":
		return HandleHelp(botService, message)
	case "cancel":
		return HandleCancelCommand(botService, message)
	case "complaint":
		return HandleComplaintCommand(botService, message)
	case "proposal":
		return HandleProposalCommand(botService, message)
	case "my_proposals":
		return HandleMyProposalsCommand(botService, message)
	case "timetable":
		return HandleViewTimetableCommand(botService, message)
	case "announcements":
		return HandleViewAnnouncementsCommand(botService, message)
	case "admin":
		return HandleAdminCommand(botService, message)
	case "admin_link":
		return HandleAdminLinkCommand(botService, message)
	case "manage_classes":
		return HandleManageClassesCommand(botService, message)
	case "add_class":
		return HandleAddClassCommand(botService, message)
	case "delete_class":
		return HandleDeleteClassCommand(botService, message)
	case "toggle_class":
		return HandleToggleClassCommand(botService, message)
	case "upload_timetable":
		return HandleUploadTimetableCommand(botService, message)
	case "post_announcement":
		return HandlePostAnnouncementCommand(botService, message)
	default:
		// Unknown command
		return HandleStart(botService, message)
	}
}
