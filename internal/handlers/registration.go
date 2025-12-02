package handlers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
	"parent-bot/internal/utils"
	"parent-bot/internal/validator"
)

// HandleLanguageSelection handles language selection callback
func HandleLanguageSelection(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Parse language
	var lang i18n.Language
	if callback.Data == "lang_uz" {
		lang = i18n.LanguageUzbek
	} else if callback.Data == "lang_ru" {
		lang = i18n.LanguageRussian
	} else {
		return nil
	}

	// Save language in state
	data := &models.StateData{Language: string(lang)}
	err := botService.StateManager.Set(telegramID, models.StateAwaitingPhone, data)
	if err != nil {
		return err
	}

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Send phone request message
	text := i18n.Get(i18n.MsgLanguageSelected, lang) + "\n\n" +
		i18n.Get(i18n.MsgRequestPhone, lang)

	keyboard := utils.MakePhoneKeyboard(lang)

	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandlePhoneNumber handles phone number input and completes registration
func HandlePhoneNumber(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID
	lang := i18n.GetLanguage(stateData.Language)

	// Extract phone number
	var phoneNumber string
	if message.Contact != nil {
		phoneNumber = message.Contact.PhoneNumber
	} else {
		phoneNumber = message.Text
	}

	// Validate phone number
	validPhone, err := validator.ValidateUzbekPhone(phoneNumber)
	if err != nil {
		text := i18n.Get(i18n.ErrInvalidPhone, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Create user (parent) - students will be linked separately
	userReq := &models.CreateUserRequest{
		TelegramID:       telegramID,
		TelegramUsername: message.From.UserName,
		PhoneNumber:      validPhone,
		Language:         stateData.Language,
	}

	user, err := botService.UserService.CreateUser(userReq)
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Clear state
	err = botService.StateManager.Clear(telegramID)
	if err != nil {
		return err
	}

	// Link admin telegram ID if this user is an admin
	_ = botService.AdminRepo.UpdateTelegramID(user.PhoneNumber, user.TelegramID)

	// Send registration complete message
	text := fmt.Sprintf(
		"✅ Ro'yxatdan o'tish muvaffaqiyatli yakunlandi!\n"+
			"Telefon: %s\n\n"+
			"Farzandlaringiz ma'lumotlarini bog'lash uchun ma'muriyatga murojaat qiling.\n\n"+
			"✅ Регистрация успешно завершена!\n"+
			"Телефон: %s\n\n"+
			"Для привязки данных ваших детей обратитесь к администрации.",
		validPhone, validPhone,
	)

	// Check if user is admin to show appropriate keyboard
	isAdmin, _ := botService.IsAdmin(user.PhoneNumber, user.TelegramID)
	keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)

	return botService.TelegramService.SendMessage(
		chatID,
		text,
		keyboard,
	)
}

// HandleChildName - DEPRECATED: No longer used in new architecture
// Students are now managed separately and linked to parents by admin/teachers
func HandleChildName(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Redirect to registration completion
	text := "Registration flow has been updated. Please use /start to begin registration."
	_ = botService.StateManager.Clear(telegramID)
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleClassSelection handles class selection from inline keyboard
func HandleClassSelection(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Extract class name from callback data
	className := callback.Data[6:] // Remove "class_" prefix

	// Verify class exists and is active
	exists, err := botService.ClassRepo.Exists(className)
	if err != nil {
		return err
	}

	if !exists {
		text := "❌ Bu sinf mavjud emas / Этого класса не существует"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	// Answer callback query - DEPRECATED FLOW
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Redirect to new registration flow
	text := "Registration flow has been updated. Please use /start to begin registration."
	_ = botService.StateManager.Clear(telegramID)
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleChildClass - DEPRECATED: No longer used in new architecture
// This is kept for backward compatibility but now we prefer inline buttons
func HandleChildClass(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Redirect to new registration flow
	text := "Registration flow has been updated. Please use /start to begin registration."
	_ = botService.StateManager.Clear(telegramID)
	return botService.TelegramService.SendMessage(chatID, text, nil)
}
