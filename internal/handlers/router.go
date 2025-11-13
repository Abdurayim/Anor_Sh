package handlers

import (
	"fmt"

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

	case models.StateAwaitingProposal:
		return HandleProposalText(botService, message, stateData)

	case models.StateConfirmingProposal:
		// Waiting for confirmation (handled by callback)
		return nil

	case models.StateAwaitingTimetableFile:
		return HandleTimetableFileUpload(botService, message, stateData)

	case models.StateAwaitingAnnouncementContent:
		return HandleAnnouncementContent(botService, message, stateData)

	case models.StateAwaitingAnnouncementFile:
		return HandleAnnouncementFile(botService, message, stateData)

	case models.StateAwaitingEditedAnnouncementContent:
		return HandleEditedAnnouncementContent(botService, message, stateData)

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
	// Check if message is a button press
	buttonText := message.Text

	// Admin panel button (check both languages) - check this BEFORE checking if user is nil
	// because admin might not be registered as a parent
	if buttonText == "👨‍💼 Ma'muriyat paneli" || buttonText == "👨‍💼 Панель администратора" {
		return HandleAdminCommand(botService, message)
	}

	if user == nil {
		return HandleStart(botService, message)
	}

	lang := i18n.GetLanguage(user.Language)
	chatID := message.Chat.ID

	// Submit complaint button (check both languages)
	if buttonText == "✍️ Shikoyat yuborish" || buttonText == "✍️ Подать жалобу" {
		return HandleComplaintCommand(botService, message)
	}

	// My complaints button (check both languages)
	if buttonText == "📋 Mening shikoyatlarim" || buttonText == "📋 Мои жалобы" {
		return HandleMyComplaintsCommand(botService, message)
	}

	// Submit proposal button (check both languages)
	if buttonText == "💡 Taklif yuborish" || buttonText == "💡 Подать предложение" {
		return HandleProposalCommand(botService, message)
	}

	// View timetable button (check both languages)
	if buttonText == "📅 Dars jadvali" || buttonText == "📅 Расписание уроков" {
		return HandleViewTimetableCommand(botService, message)
	}

	// View announcements button (check both languages)
	if buttonText == "📢 E'lonlar" || buttonText == "📢 Объявления" {
		return HandleViewAnnouncementsCommand(botService, message)
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

	// Proposal confirmation
	if data == "confirm_proposal" {
		return HandleProposalConfirmation(botService, callback)
	}

	// Proposal cancellation
	if data == "cancel_proposal" {
		return HandleProposalCancellation(botService, callback)
	}

	// Timetable class selection
	if len(data) > 17 && data[:17] == "timetable_select_" {
		var classID int
		fmt.Sscanf(data, "timetable_select_%d", &classID)
		return HandleTimetableClassSelection(botService, callback, classID)
	}

	// Announcement skip file
	if data == "announcement_skip_file" {
		stateData, err := botService.StateManager.GetData(callback.From.ID)
		if err != nil {
			return err
		}
		return HandleAnnouncementSkipFile(botService, callback, stateData)
	}

	// Announcement edit callback
	if len(data) > 18 && data[:18] == "announcement_edit_" {
		var announcementID int
		fmt.Sscanf(data, "announcement_edit_%d", &announcementID)
		return HandleAnnouncementEditCallback(botService, callback, announcementID)
	}

	// Announcement delete callback
	if len(data) > 20 && data[:20] == "announcement_delete_" {
		var announcementID int
		fmt.Sscanf(data, "announcement_delete_%d", &announcementID)
		return HandleAnnouncementDeleteCallback(botService, callback, announcementID)
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

	// Admin upload timetable callback
	if data == "admin_upload_timetable" {
		return HandleAdminUploadTimetableCallback(botService, callback)
	}

	// Admin post announcement callback
	if data == "admin_post_announcement" {
		return HandleAdminPostAnnouncementCallback(botService, callback)
	}

	// Admin view proposals callback
	if data == "admin_proposals" {
		return HandleAdminProposalsCallback(botService, callback)
	}

	// Admin view timetables callback
	if data == "admin_view_timetables" {
		return HandleAdminViewTimetablesCallback(botService, callback)
	}

	// Admin view announcements callback
	if data == "admin_view_announcements" {
		return HandleAdminViewAnnouncementsCallback(botService, callback)
	}

	// Class delete callback
	if len(data) > 13 && data[:13] == "class_delete_" {
		return HandleClassDeleteCallback(botService, callback)
	}

	// Class info callback (just acknowledge, no action needed)
	if len(data) > 11 && data[:11] == "class_info_" {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return nil
	}

	// Timetable delete callback
	if len(data) > 17 && data[:17] == "timetable_delete_" {
		return HandleTimetableDeleteCallback(botService, callback)
	}

	// Admin back button
	if data == "admin_back" {
		return HandleAdminBackCallback(botService, callback)
	}

	// Unknown callback
	return botService.TelegramService.AnswerCallbackQuery(callback.ID, "Unknown action")
}
