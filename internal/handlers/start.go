package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
	"parent-bot/internal/utils"
)

// HandleStart handles /start command
func HandleStart(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Super admin: statistics of all branches and admin management
	if botService.GetSuperAdmin(telegramID) != nil {
		_ = botService.StateManager.Clear(telegramID)
		text := "👑 <b>Super admin</b>\n\nAssalomu aleykum! / Здравствуйте!\n\n" +
			"Panelni ochish uchun pastdagi tugmani bosing.\nНажмите кнопку ниже, чтобы открыть панель."
		return botService.TelegramService.SendMessage(chatID, text, makeSuperAdminReplyKeyboard())
	}

	// FIRST: Check if this person is a teacher
	teacher, _ := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if teacher != nil {
		// Teacher interface - show teacher menu
		lang := i18n.GetLanguage(teacher.Language)
		text := "👨‍🏫 <b>O'qituvchi paneli / Панель учителя</b>\n\n"
		text += "Assalomu aleykum! / Здравствуйте!\n\n"
		text += i18n.Get(i18n.MsgMainMenu, lang)

		keyboard := utils.MakeTeacherMainMenuKeyboard(lang)
		return botService.TelegramService.SendMessage(chatID, text, keyboard)
	}

	// SECOND: Check if this person is an admin
	user, _ := botService.UserService.GetUserByTelegramID(telegramID)
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, _ := botService.IsAdmin(phoneNumber, telegramID)

	// ADMIN INTERFACE - No registration needed, but they can also register as parent if they want
	if isAdmin {
		lang := i18n.LanguageUzbek
		if user != nil {
			lang = i18n.GetLanguage(user.Language)
		}

		text := "👨‍💼 <b>Admin Panel / Панель Администратора</b>\n\n"
		text += "Assalomu aleykum! / Здравствуйте!\n\n"
		text += "Siz admin sifatida tanildingiz.\n"
		text += "Вы распознаны как администратор.\n\n"
		text += "Admin paneliga o'tish uchun quyidagi tugmani bosing:\n"
		text += "Нажмите кнопку ниже для доступа к панели администратора:"

		// Show admin panel button
		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnAdminPanel, lang)),
			),
		)

		return botService.TelegramService.SendMessage(chatID, text, keyboard)
	}

	// PARENT INTERFACE - Registration required
	if user != nil {
		// Parent already registered, show parent menu
		lang := i18n.GetLanguage(user.Language)
		text := i18n.Get(i18n.MsgMainMenu, lang)
		keyboard := utils.MakeMainMenuKeyboard(lang)

		return botService.TelegramService.SendMessage(chatID, text, keyboard)
	}

	// New parent user, show welcome and registration
	text := i18n.Get(i18n.MsgWelcome, i18n.LanguageUzbek) + "\n\n" +
		i18n.Get(i18n.MsgChooseLanguage, i18n.LanguageUzbek)

	keyboard := utils.MakeLanguageKeyboard()

	// Set initial state
	err := botService.StateManager.Set(telegramID, models.StateAwaitingLanguage, &models.StateData{})
	if err != nil {
		return err
	}

	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleHelp handles /help command
func HandleHelp(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID

	// Get user language
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	lang := i18n.LanguageUzbek
	if user != nil {
		lang = i18n.GetLanguage(user.Language)
	}

	helpText := "📋 <b>Bot haqida / О боте</b>\n\n"

	if lang == i18n.LanguageUzbek {
		helpText += "Bu bot maktab ota-onalarining shikoyatlarini rasmiy ravishda qabul qilish uchun mo'ljallangan.\n\n"
		helpText += "<b>Buyruqlar:</b>\n"
		helpText += "/start - Botni ishga tushirish\n"
		helpText += "/help - Yordam\n"
		helpText += "/cancel - Amalni bekor qilish\n"
		helpText += "/complaint - Shikoyat yuborish\n\n"
		helpText += "<b>Qo'llab-quvvatlash:</b>\n"
		helpText += "Muammolar yuzaga kelsa, maktab ma'muriyatiga murojaat qiling."
	} else {
		helpText += "Этот бот предназначен для официального приема жалоб родителей школьников.\n\n"
		helpText += "<b>Команды:</b>\n"
		helpText += "/start - Запустить бота\n"
		helpText += "/help - Помощь\n"
		helpText += "/cancel - Отменить действие\n"
		helpText += "/complaint - Подать жалобу\n\n"
		helpText += "<b>Поддержка:</b>\n"
		helpText += "Если возникли проблемы, обратитесь к администрации школы."
	}

	return botService.TelegramService.SendMessage(message.Chat.ID, helpText, nil)
}

// HandleCancelCommand handles /cancel command - clears any active state
func HandleCancelCommand(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Clear any active state
	_ = botService.StateManager.Clear(telegramID)

	// Check if this is a teacher
	teacher, _ := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if teacher != nil {
		lang := i18n.GetLanguage(teacher.Language)
		var text string
		if lang == i18n.LanguageUzbek {
			text = "❌ Amal bekor qilindi."
		} else {
			text = "❌ Действие отменено."
		}
		keyboard := utils.MakeTeacherMainMenuKeyboard(lang)
		return botService.TelegramService.SendMessage(chatID, text, keyboard)
	}

	// Get user (parent/admin)
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Determine language
	lang := i18n.LanguageUzbek
	if user != nil {
		lang = i18n.GetLanguage(user.Language)
	}

	// Send cancellation message
	var text string
	if lang == i18n.LanguageUzbek {
		text = "❌ Amal bekor qilindi."
	} else {
		text = "❌ Действие отменено."
	}

	// Check if admin to show appropriate keyboard
	var keyboard tgbotapi.ReplyKeyboardMarkup
	if user != nil {
		isAdmin, _ := botService.IsAdmin(user.PhoneNumber, user.TelegramID)
		keyboard = utils.MakeMainMenuKeyboardForUser(lang, isAdmin)
	} else {
		// No keyboard for unregistered users
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}
