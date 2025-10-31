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

// HandlePhoneNumber handles phone number input
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

	// Update state with phone number
	stateData.PhoneNumber = validPhone
	err = botService.StateManager.Set(telegramID, models.StateAwaitingChildName, stateData)
	if err != nil {
		return err
	}

	// Send confirmation and request child name
	text := i18n.Get(i18n.MsgPhoneReceived, lang)
	text = fmt.Sprintf(text, validPhone)
	_ = botService.TelegramService.SendMessage(chatID, text, utils.RemoveKeyboard())

	text = i18n.Get(i18n.MsgRequestChildName, lang)
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleChildName handles child name input
func HandleChildName(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID
	lang := i18n.GetLanguage(stateData.Language)

	// Validate name
	childName, err := validator.ValidateName(message.Text)
	if err != nil {
		text := i18n.Get(i18n.ErrInvalidName, lang) + "\n\n" + err.Error()
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Update state with child name
	stateData.ChildName = childName
	err = botService.StateManager.Set(telegramID, models.StateAwaitingChildClass, stateData)
	if err != nil {
		return err
	}

	// Send confirmation
	text := i18n.Get(i18n.MsgChildNameReceived, lang)
	text = fmt.Sprintf(text, childName)
	_ = botService.TelegramService.SendMessage(chatID, text, nil)

	// Get active classes
	classes, err := botService.ClassRepo.GetActive()
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(classes) == 0 {
		text := "❌ Hozircha mavjud sinflar yo'q. Iltimos, ma'muriyatga murojaat qiling.\n\n" +
			"❌ Пока нет доступных классов. Пожалуйста, обратитесь к администрации."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Send class selection keyboard with improved formatting
	text = "🎓 <b>Sinf tanlash / Выбор класса</b>\n\n"
	text += "Farzandingiz qaysi sinfda o'qiydi? Quyidagi tugmalardan tanlang:\n"
	text += "В каком классе учится ваш ребенок? Выберите из кнопок ниже:\n\n"
	text += fmt.Sprintf("📚 Mavjud sinflar / Доступные классы: <b>%d ta</b>\n\n", len(classes))
	text += "👇 Sinfni tanlash uchun tugmani bosing:"

	keyboard := utils.MakeClassSelectionKeyboard(classes, lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleClassSelection handles class selection from inline keyboard
func HandleClassSelection(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Extract class name from callback data
	className := callback.Data[6:] // Remove "class_" prefix

	// Get state data
	stateData, err := botService.StateManager.GetData(telegramID)
	if err != nil {
		return err
	}

	lang := i18n.GetLanguage(stateData.Language)

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

	// Answer callback query
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Create user
	userReq := &models.CreateUserRequest{
		TelegramID:       telegramID,
		TelegramUsername: callback.From.UserName,
		PhoneNumber:      stateData.PhoneNumber,
		ChildName:        stateData.ChildName,
		ChildClass:       className,
		Language:         stateData.Language,
	}

	user, err := botService.UserService.CreateUser(userReq)
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Update state to registered
	err = botService.StateManager.Clear(telegramID)
	if err != nil {
		return err
	}

	// Send registration complete message
	text := i18n.Get(i18n.MsgRegistrationComplete, lang)
	text = fmt.Sprintf(text, user.ChildName, user.ChildClass, user.PhoneNumber)
	text = utils.EscapeMarkdown(text)

	// Link admin telegram ID if this user is an admin
	// This links the admin's telegram_id to their admin record for faster future checks
	_ = botService.AdminRepo.UpdateTelegramID(user.PhoneNumber, user.TelegramID)

	// Check if user is admin to show appropriate keyboard
	isAdmin, _ := botService.IsAdmin(user.PhoneNumber, user.TelegramID)
	keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)

	return botService.TelegramService.SendMessage(
		chatID,
		text,
		keyboard,
	)
}

// HandleChildClass handles child class input and completes registration
// This is kept for backward compatibility but now we prefer inline buttons
func HandleChildClass(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID
	lang := i18n.GetLanguage(stateData.Language)

	// Validate class - check if it exists in database
	className := utils.SanitizeClassName(message.Text)

	exists, err := botService.ClassRepo.Exists(className)
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if !exists {
		text := "❌ Bu sinf ro'yxatda yo'q. Iltimos, tugmalardan tanlang yoki ma'muriyatga murojaat qiling.\n\n" +
			"❌ Этого класса нет в списке. Пожалуйста, выберите из кнопок или обратитесь к администрации."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	childClass := className

	// Create user
	userReq := &models.CreateUserRequest{
		TelegramID:       telegramID,
		TelegramUsername: message.From.UserName,
		PhoneNumber:      stateData.PhoneNumber,
		ChildName:        stateData.ChildName,
		ChildClass:       childClass,
		Language:         stateData.Language,
	}

	user, err := botService.UserService.CreateUser(userReq)
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Update state to registered
	err = botService.StateManager.Clear(telegramID)
	if err != nil {
		return err
	}

	// Send registration complete message
	text := i18n.Get(i18n.MsgRegistrationComplete, lang)
	text = fmt.Sprintf(text, user.ChildName, user.ChildClass, user.PhoneNumber)
	text = utils.EscapeMarkdown(text)

	// Link admin telegram ID if this user is an admin
	// This links the admin's telegram_id to their admin record for faster future checks
	_ = botService.AdminRepo.UpdateTelegramID(user.PhoneNumber, user.TelegramID)

	// Check if user is admin to show appropriate keyboard
	isAdmin, _ := botService.IsAdmin(user.PhoneNumber, user.TelegramID)
	keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)

	return botService.TelegramService.SendMessage(
		chatID,
		text,
		keyboard,
	)
}
