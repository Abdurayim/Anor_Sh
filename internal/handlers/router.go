package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
	"parent-bot/internal/utils"
)

// RouteByState routes messages based on user's current state
func RouteByState(botService *services.BotService, message *tgbotapi.Message, state string, stateData *models.StateData) error {
	switch state {
	case models.StateAwaitingLanguage:
		// Waiting for language selection (handled by callback)
		return nil

	case models.StateAwaitingPhone:
		return HandlePhoneNumber(botService, message, stateData)

	case models.StateAwaitingChildName:
		return HandleChildName(botService, message, stateData)

	case models.StateAwaitingChildClass:
		return HandleChildClass(botService, message, stateData)

	case models.StateAwaitingComplaint:
		return HandleComplaintText(botService, message, stateData)

	case models.StateConfirmingComplaint:
		// Waiting for confirmation (handled by callback)
		return nil

	case models.StateAwaitingAdminPhone:
		return HandleAdminLinkPhone(botService, message)

	case models.StateAwaitingClassName:
		return HandleClassNameInput(botService, message)

	case models.StateRegistered:
		// User is registered, get user data
		user, err := botService.UserService.GetUserByTelegramID(message.From.ID)
		if err != nil {
			return err
		}
		return HandleRegisteredUserMessage(botService, message, user)

	default:
		// Unknown state, restart
		return HandleStart(botService, message)
	}
}

// HandleRegisteredUserMessage handles messages from registered users
func HandleRegisteredUserMessage(botService *services.BotService, message *tgbotapi.Message, user *models.User) error {
	if user == nil {
		return HandleStart(botService, message)
	}

	lang := i18n.GetLanguage(user.Language)
	chatID := message.Chat.ID

	// Check if message is a button press
	buttonText := message.Text

	// Admin panel button (check both languages)
	if buttonText == "👨‍💼 Ma'muriyat paneli" || buttonText == "👨‍💼 Панель администратора" {
		return HandleAdminCommand(botService, message)
	}

	// Submit complaint button (check both languages)
	if buttonText == "✍️ Shikoyat yuborish" || buttonText == "✍️ Подать жалобу" {
		return HandleComplaintCommand(botService, message)
	}

	// My complaints button (check both languages)
	if buttonText == "📋 Mening shikoyatlarim" || buttonText == "📋 Мои жалобы" {
		return HandleMyComplaintsCommand(botService, message)
	}

	// Settings button (check both languages)
	if buttonText == "⚙️ Sozlamalar" || buttonText == "⚙️ Настройки" {
		return HandleSettingsCommand(botService, message)
	}

	// Default: show main menu
	text := i18n.Get(i18n.MsgMainMenu, lang)

	// Check if user is admin to show appropriate keyboard
	isAdmin, _ := botService.IsAdmin(user.PhoneNumber, user.TelegramID)
	keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)

	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleCallbackQuery handles inline button clicks
func HandleCallbackQuery(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	data := callback.Data

	// Language selection
	if data == "lang_uz" || data == "lang_ru" {
		return HandleLanguageSelection(botService, callback)
	}

	// Class selection (starts with "class_")
	if len(data) > 6 && data[:6] == "class_" {
		return HandleClassSelection(botService, callback)
	}

	// Complaint confirmation
	if data == "confirm_complaint" {
		return HandleComplaintConfirmation(botService, callback)
	}

	// Complaint cancellation
	if data == "cancel_complaint" {
		return HandleComplaintCancellation(botService, callback)
	}

	// Admin callbacks
	if data == "admin_users" {
		return HandleAdminUsersCallback(botService, callback)
	}

	if data == "admin_complaints" {
		return HandleAdminComplaintsCallback(botService, callback)
	}

	if data == "admin_stats" {
		return HandleAdminStatsCallback(botService, callback)
	}

	// Admin manage classes callback
	if data == "admin_manage_classes" {
		return HandleAdminManageClassesCallback(botService, callback)
	}

	// Admin create class callback
	if data == "admin_create_class" {
		return HandleAdminCreateClassCallback(botService, callback)
	}

	// Class action callbacks (activate, deactivate, delete)
	if len(data) > 12 && data[:12] == "class_delete" {
		return HandleClassDeleteCallback(botService, callback)
	}

	if len(data) > 13 && data[:13] == "class_toggle_" {
		return HandleClassToggleCallback(botService, callback)
	}

	// Admin back button
	if data == "admin_back" {
		return HandleAdminBackCallback(botService, callback)
	}

	// Unknown callback
	return botService.TelegramService.AnswerCallbackQuery(callback.ID, "Unknown action")
}
