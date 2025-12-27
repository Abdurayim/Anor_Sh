package handlers

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
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
	buttonText := message.Text
	if buttonText == "👨‍💼 Ma'muriyat paneli" || buttonText == "👨‍💼 Панель администратора" {
		_ = botService.StateManager.Clear(telegramID)
		return HandleAdminCommand(botService, message)
	}

	// ============================================
	// TEACHER CHECK - Teachers get their own flow
	// ============================================
	teacher, _ := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if teacher != nil {
		// Found a teacher - ALWAYS use teacher flow, regardless of IsActive
		// HandleTeacherMessage will handle everything internally
		return HandleTeacherMessage(botService, message, teacher)
	}

	// ============================================
	// PARENT/USER FLOW - Only if NOT a teacher
	// ============================================

	// Check if user is registered as parent FIRST
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		log.Printf("[WEBHOOK] Error getting user for %d: %v", telegramID, err)
	}

	// IMPORTANT: Check for parent menu button presses FIRST - these should override any active state
	if user != nil && isParentMenuButton(buttonText) {
		// Clear any existing state when pressing menu buttons
		_ = botService.StateManager.Clear(telegramID)
		return HandleRegisteredUserMessage(botService, message, user)
	}

	// Check if user has an active state (registration/complaint/etc.)
	state, _ := botService.StateManager.Get(telegramID)
	if state != nil && state.State != "" {
		stateData, err := botService.StateManager.GetData(telegramID)
		if err != nil {
			log.Printf("[WEBHOOK] Error getting state data for %d: %v", telegramID, err)
			_ = botService.StateManager.Clear(telegramID)
		} else {
			return RouteByState(botService, message, state.State, stateData)
		}
	}

	if user != nil {
		return HandleRegisteredUserMessage(botService, message, user)
	}

	// New user - start registration
	return HandleStart(botService, message)
}

// isParentMenuButton checks if the button text is one of the parent main menu buttons
func isParentMenuButton(buttonText string) bool {
	// Check all parent menu buttons in both languages
	parentButtons := []string{
		i18n.Get(i18n.BtnMyChildren, i18n.LanguageUzbek),
		i18n.Get(i18n.BtnMyChildren, i18n.LanguageRussian),
		i18n.Get(i18n.BtnMyAttendance, i18n.LanguageUzbek),
		i18n.Get(i18n.BtnMyAttendance, i18n.LanguageRussian),
		i18n.Get(i18n.BtnMyTestResults, i18n.LanguageUzbek),
		i18n.Get(i18n.BtnMyTestResults, i18n.LanguageRussian),
		i18n.Get(i18n.BtnViewTimetable, i18n.LanguageUzbek),
		i18n.Get(i18n.BtnViewTimetable, i18n.LanguageRussian),
		i18n.Get(i18n.BtnViewAnnouncements, i18n.LanguageUzbek),
		i18n.Get(i18n.BtnViewAnnouncements, i18n.LanguageRussian),
		i18n.Get(i18n.BtnSubmitComplaint, i18n.LanguageUzbek),
		i18n.Get(i18n.BtnSubmitComplaint, i18n.LanguageRussian),
		i18n.Get(i18n.BtnSubmitProposal, i18n.LanguageUzbek),
		i18n.Get(i18n.BtnSubmitProposal, i18n.LanguageRussian),
	}

	for _, btn := range parentButtons {
		if buttonText == btn {
			return true
		}
	}
	return false
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
	case "add_student":
		return HandleAddStudentCommand(botService, message)
	case "link_student":
		return HandleLinkStudentCommand(botService, message)
	case "list_students":
		return HandleListStudentsCommand(botService, message)
	case "view_parent_children":
		return HandleViewParentChildrenCommand(botService, message)
	case "my_children":
		return HandleMyChildrenCommand(botService, message)
	case "add_teacher":
		return HandleAddTeacherCommand(botService, message)
	case "list_teachers":
		return HandleListTeachersCommand(botService, message)
	case "edit_grade":
		return HandleEditGradeCommand(botService, message)
	case "delete_grade":
		return HandleDeleteGradeCommand(botService, message)
	default:
		// Unknown command
		return HandleStart(botService, message)
	}
}
