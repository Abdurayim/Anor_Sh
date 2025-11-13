package handlers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
	"parent-bot/internal/utils"
)

// HandleAdminCommand handles /admin command
func HandleAdminCommand(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Get user to extract phone number
	// Note: If user hasn't registered yet, IsAdmin will check by telegram_id
	// and also look up the user internally if needed
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Extract phone number for admin check
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	// Check if user is admin (checks DB and config admin phones)
	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil {
		return err
	}

	if !isAdmin {
		text := "❌ Bu buyruq faqat ma'murlar uchun / Эта команда только для администраторов"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	lang := i18n.LanguageUzbek
	if user != nil {
		lang = i18n.GetLanguage(user.Language)
	}

	// Show admin panel
	text := i18n.Get(i18n.MsgAdminPanel, lang)
	keyboard := utils.MakeAdminKeyboard(lang)

	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleAdminUsersCallback handles admin users list callback
func HandleAdminUsersCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	chatID := callback.Message.Chat.ID

	// Get users
	users, err := botService.UserService.GetAllUsers(20, 0)
	if err != nil {
		text := "Xatolik / Ошибка: " + err.Error()
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Count total users
	totalCount, _ := botService.UserService.CountUsers()

	// Format user list
	text := fmt.Sprintf("👥 Ro'yxatdan o'tgan foydalanuvchilar / Зарегистрированные пользователи\n\n")
	text += fmt.Sprintf("Jami / Всего: %d\n\n", totalCount)

	for i, user := range users {
		text += fmt.Sprintf("%d. %s - %s sinf\n", i+1, user.ChildName, user.ChildClass)
		text += fmt.Sprintf("   📱 %s\n", user.PhoneNumber)
		if user.TelegramUsername != "" {
			text += fmt.Sprintf("   @%s\n", user.TelegramUsername)
		}
		text += fmt.Sprintf("   📅 %s\n\n", utils.FormatDate(user.RegisteredAt))
	}

	if len(users) < totalCount {
		text += fmt.Sprintf("...va yana %d ta / ...и ещё %d", totalCount-len(users), totalCount-len(users))
	}

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleAdminComplaintsCallback handles admin complaints list callback
func HandleAdminComplaintsCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	chatID := callback.Message.Chat.ID

	// Get complaints with user info
	complaints, err := botService.ComplaintService.GetAllComplaintsWithUser(10, 0)
	if err != nil {
		text := "Xatolik / Ошибка: " + err.Error()
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Count total complaints
	totalCount, _ := botService.ComplaintService.CountComplaints()

	// Format complaints list
	text := fmt.Sprintf("📋 Shikoyatlar / Жалобы\n\n")
	text += fmt.Sprintf("Jami / Всего: %d\n\n", totalCount)

	for i, c := range complaints {
		statusEmoji := "⏳"
		statusText := "Kutilmoqda / Ожидание"

		if c.Status == models.StatusReviewed {
			statusEmoji = "✅"
			statusText = "Ko'rib chiqildi / Рассмотрено"
		} else if c.Status == models.StatusArchived {
			statusEmoji = "📦"
			statusText = "Arxivlangan / Архивировано"
		}

		text += fmt.Sprintf("%d. %s #%d - %s %s\n", i+1, statusEmoji, c.ID, c.ChildName, c.ChildClass)
		text += fmt.Sprintf("   📱 %s\n", c.PhoneNumber)
		preview := utils.TruncateText(c.ComplaintText, 60)
		text += fmt.Sprintf("   💬 %s\n", preview)
		text += fmt.Sprintf("   📅 %s\n", utils.FormatDateTime(c.CreatedAt))
		text += fmt.Sprintf("   📊 %s\n\n", statusText)
	}

	if len(complaints) < totalCount {
		text += fmt.Sprintf("...va yana %d ta / ...и ещё %d", totalCount-len(complaints), totalCount-len(complaints))
	}

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleAdminStatsCallback handles admin statistics callback
func HandleAdminStatsCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	chatID := callback.Message.Chat.ID

	// Get statistics
	totalUsers, _ := botService.UserService.CountUsers()
	totalComplaints, _ := botService.ComplaintService.CountComplaints()
	pendingComplaints, _ := botService.ComplaintService.CountComplaintsByStatus(models.StatusPending)
	reviewedComplaints, _ := botService.ComplaintService.CountComplaintsByStatus(models.StatusReviewed)

	// Format statistics
	text := "📊 Statistika / Статистика\n\n"
	text += fmt.Sprintf("👥 Foydalanuvchilar / Пользователи: %d\n\n", totalUsers)
	text += fmt.Sprintf("📋 Jami shikoyatlar / Всего жалоб: %d\n", totalComplaints)
	text += fmt.Sprintf("⏳ Kutilmoqda / Ожидание: %d\n", pendingComplaints)
	text += fmt.Sprintf("✅ Ko'rib chiqildi / Рассмотрено: %d\n", reviewedComplaints)

	if totalComplaints > 0 {
		percentage := float64(reviewedComplaints) / float64(totalComplaints) * 100
		text += fmt.Sprintf("\n📈 Ko'rilganlik / Процент рассмотрения: %.1f%%\n", percentage)
	}

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleManageClassesCommand handles /manage_classes command
func HandleManageClassesCommand(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Check if user is admin
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil {
		return err
	}

	if !isAdmin {
		text := "❌ Bu buyruq faqat ma'murlar uchun / Эта команда только для администраторов"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	lang := i18n.LanguageUzbek
	if user != nil {
		lang = i18n.GetLanguage(user.Language)
	}

	// Get all classes
	classes, err := botService.ClassRepo.GetAll()
	if err != nil {
		text := "Xatolik / Ошибка: " + err.Error()
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Format classes list
	text := "📚 Sinflarni boshqarish / Управление классами\n\n"

	if len(classes) == 0 {
		text += "Hozircha sinflar yo'q / Пока нет классов\n\n"
	} else {
		text += "Mavjud sinflar / Существующие классы:\n\n"
		for i, class := range classes {
			status := "✅"
			if !class.IsActive {
				status = "❌"
			}
			text += fmt.Sprintf("%d. %s %s\n", i+1, status, class.ClassName)
		}
		text += "\n"
	}

	text += "Buyruqlar / Команды:\n"
	text += "/add_class <sinf nomi> - Sinf qo'shish\n"
	text += "   Misol: /add_class 9A\n\n"
	text += "/delete_class <sinf nomi> - Sinfni o'chirish\n"
	text += "   Misol: /delete_class 9A\n\n"
	text += "/toggle_class <sinf nomi> - Sinfni faollashtirish/o'chirish\n"
	text += "   Misol: /toggle_class 9A"

	_ = lang // Will be used in future for localized messages

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleAddClassCommand handles /add_class command
func HandleAddClassCommand(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Check if user is admin
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil {
		return err
	}

	if !isAdmin {
		text := "❌ Bu buyruq faqat ma'murlar uchun / Эта команда только для администраторов"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Parse class name from command
	className := message.CommandArguments()
	if className == "" {
		text := "❌ Sinf nomini kiriting / Введите название класса\n\nMisol / Пример: /add_class 9A"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Validate class name (allow alphanumeric + dash)
	className = utils.SanitizeClassName(className)

	// Create class
	class, err := botService.ClassRepo.Create(className)
	if err != nil {
		text := "❌ Xatolik / Ошибка: " + err.Error()
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	text := fmt.Sprintf("✅ Sinf qo'shildi / Класс добавлен: %s", class.ClassName)
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleDeleteClassCommand handles /delete_class command
func HandleDeleteClassCommand(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Check if user is admin
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil {
		return err
	}

	if !isAdmin {
		text := "❌ Bu buyruq faqat ma'murlar uchun / Эта команда только для администраторов"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Parse class name from command
	className := message.CommandArguments()
	if className == "" {
		text := "❌ Sinf nomini kiriting / Введите название класса\n\nMisol / Пример: /delete_class 9A"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Delete class
	err = botService.ClassRepo.Delete(className)
	if err != nil {
		text := "❌ Xatolik / Ошибка: " + err.Error()
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	text := fmt.Sprintf("✅ Sinf o'chirildi / Класс удален: %s", className)
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleToggleClassCommand handles /toggle_class command
func HandleToggleClassCommand(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Check if user is admin
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil {
		return err
	}

	if !isAdmin {
		text := "❌ Bu buyruq faqat ma'murlar uchun / Эта команда только для администраторов"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Parse class name from command
	className := message.CommandArguments()
	if className == "" {
		text := "❌ Sinf nomini kiriting / Введите название класса\n\nMisol / Пример: /toggle_class 9A"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Toggle class
	err = botService.ClassRepo.ToggleActive(className)
	if err != nil {
		text := "❌ Xatolik / Ошибка: " + err.Error()
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	text := fmt.Sprintf("✅ Sinf holati o'zgartirildi / Статус класса изменен: %s", className)
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleAdminManageClassesCallback handles admin manage classes callback
func HandleAdminManageClassesCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Check if user is admin
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil {
		return err
	}

	if !isAdmin {
		text := "❌ Bu buyruq faqat ma'murlar uchun / Эта команда только для администраторов"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	lang := i18n.LanguageUzbek
	if user != nil {
		lang = i18n.GetLanguage(user.Language)
	}

	// Get all classes
	classes, err := botService.ClassRepo.GetAll()
	if err != nil {
		text := "❌ Xatolik / Ошибка: " + err.Error()
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Format class management message
	text := "📚 <b>Sinflarni boshqarish / Управление классами</b>\n\n"

	if len(classes) == 0 {
		text += "Hozircha sinflar yo'q / Пока нет классов\n\n"
		text += "Yangi sinf yaratish uchun quyidagi tugmani bosing:"
	} else {
		text += "Jami / Всего: " + fmt.Sprintf("%d", len(classes)) + "\n\n"

		text += "<b>Sinflar ro'yxati / Список классов:</b>\n\n"

		for i, class := range classes {
			text += fmt.Sprintf("%d. <b>%s</b>\n", i+1, class.ClassName)
		}

		text += "\n👇 O'chirish uchun sinfni tanlang / Выберите класс для удаления:"
	}

	// Create keyboard with class management options
	keyboard := makeClassManagementKeyboard(classes, lang)

	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// makeClassManagementKeyboard creates keyboard for class management
func makeClassManagementKeyboard(classes []*models.Class, lang i18n.Language) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Add each class with separate name and delete buttons
	for _, class := range classes {
		var row []tgbotapi.InlineKeyboardButton

		// Class name button (non-clickable, just for display)
		nameBtn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("📚 %s", class.ClassName),
			fmt.Sprintf("class_info_%s", class.ClassName),
		)
		row = append(row, nameBtn)

		// Delete button (separate)
		deleteBtn := tgbotapi.NewInlineKeyboardButtonData(
			"🗑 O'chirish / Удалить",
			fmt.Sprintf("class_delete_%s", class.ClassName),
		)
		row = append(row, deleteBtn)

		rows = append(rows, row)
	}

	// Add create class button
	createBtn := tgbotapi.NewInlineKeyboardButtonData(
		i18n.Get(i18n.BtnCreateClass, lang),
		"admin_create_class",
	)
	rows = append(rows, []tgbotapi.InlineKeyboardButton{createBtn})

	// Add back button
	backBtn := tgbotapi.NewInlineKeyboardButtonData(
		i18n.Get(i18n.BtnBack, lang),
		"admin_back",
	)
	rows = append(rows, []tgbotapi.InlineKeyboardButton{backBtn})

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// HandleClassToggleCallback handles toggling class active status
func HandleClassToggleCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Check if user is admin
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil {
		return err
	}

	if !isAdmin {
		text := "❌ Faqat ma'murlar uchun / Только для администраторов"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	// Extract class name from callback data
	className := callback.Data[13:] // Remove "class_toggle_" prefix

	// Toggle class status
	err = botService.ClassRepo.ToggleActive(className)
	if err != nil {
		text := "❌ Xatolik / Ошибка"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	// Answer callback with success
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅ Holat o'zgartirildi / Статус изменен")

	// Refresh the class management view
	return HandleAdminManageClassesCallback(botService, callback)
}

// HandleClassDeleteCallback handles deleting a class
func HandleClassDeleteCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Check if user is admin
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil {
		return err
	}

	if !isAdmin {
		text := "❌ Faqat ma'murlar uchun / Только для администраторов"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	// Extract class name from callback data
	className := callback.Data[13:] // Remove "class_delete_" prefix

	// Delete class
	err = botService.ClassRepo.Delete(className)
	if err != nil {
		text := fmt.Sprintf("❌ Xatolik / Ошибка: %v", err)
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return err
	}

	// Answer callback with success
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅ Sinf o'chirildi / Класс удален")

	// Refresh the class management view by editing the current message
	lang := i18n.LanguageUzbek
	if user != nil {
		lang = i18n.GetLanguage(user.Language)
	}

	// Get updated list of classes
	classes, err := botService.ClassRepo.GetAll()
	if err != nil {
		return err
	}

	// Format updated message
	text := "📚 <b>Sinflarni boshqarish / Управление классами</b>\n\n"

	if len(classes) == 0 {
		text += "Hozircha sinflar yo'q / Пока нет классов\n\n"
		text += "Yangi sinf yaratish uchun quyidagi tugmani bosing:"
	} else {
		text += "Jami / Всего: " + fmt.Sprintf("%d", len(classes)) + "\n\n"

		text += "<b>Sinflar ro'yxati / Список классов:</b>\n\n"

		for i, class := range classes {
			text += fmt.Sprintf("%d. <b>%s</b>\n", i+1, class.ClassName)
		}

		text += "\n👇 O'chirish uchun sinfni tanlang / Выберите класс для удаления:"
	}

	// Create updated keyboard
	keyboard := makeClassManagementKeyboard(classes, lang)

	// Edit the message
	return botService.TelegramService.EditMessage(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		&keyboard,
	)
}

// HandleAdminCreateClassCallback handles admin create class callback
func HandleAdminCreateClassCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Check if user is admin
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil {
		return err
	}

	if !isAdmin {
		text := "❌ Bu buyruq faqat ma'murlar uchun / Эта команда только для администраторов"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	lang := i18n.LanguageUzbek
	if user != nil {
		lang = i18n.GetLanguage(user.Language)
	}

	// Set state to awaiting class name
	err = botService.StateManager.Set(telegramID, models.StateAwaitingClassName, &models.StateData{
		Language: string(lang),
	})
	if err != nil {
		return err
	}

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Send prompt for class name
	text := "➕ <b>Yangi sinf yaratish / Создать новый класс</b>\n\n"
	text += "Sinf nomini kiriting (masalan: 9A, 10B, 11V)\n"
	text += "Введите название класса (например: 9A, 10B, 11V)\n\n"
	text += "Yoki /cancel bekor qilish uchun"

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleClassNameInput handles class name input from admin
func HandleClassNameInput(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Check if user cancelled
	if message.Text == "/cancel" {
		_ = botService.StateManager.Clear(telegramID)
		text := "❌ Bekor qilindi / Отменено"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Check if user is admin
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil {
		return err
	}

	if !isAdmin {
		text := "❌ Bu buyruq faqat ma'murlar uchun / Эта команда только для администраторов"
		_ = botService.StateManager.Clear(telegramID)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Validate and sanitize class name
	className := utils.SanitizeClassName(message.Text)

	if className == "" {
		text := "❌ Noto'g'ri sinf nomi / Неверное название класса\n\nQaytadan urinib ko'ring:"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Check if class already exists
	exists, err := botService.ClassRepo.GetByName(className)
	if err != nil {
		text := "❌ Xatolik / Ошибка: " + err.Error()
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if exists != nil {
		text := fmt.Sprintf("❌ Bu sinf allaqachon mavjud / Этот класс уже существует: %s\n\nBoshqa nom kiriting:", className)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Create the class
	class, err := botService.ClassRepo.Create(className)
	if err != nil {
		text := "❌ Xatolik / Ошибка: " + err.Error()
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	lang := i18n.LanguageUzbek
	if user != nil {
		lang = i18n.GetLanguage(user.Language)
	}

	// Send success message with back button
	text := fmt.Sprintf("✅ <b>Sinf muvaffaqiyatli yaratildi! / Класс успешно создан!</b>\n\n")
	text += fmt.Sprintf("📚 Sinf nomi / Название класса: <b>%s</b>\n\n", class.ClassName)
	text += "Endi bu sinf barcha foydalanuvchilar uchun mavjud.\n"
	text += "Теперь этот класс доступен для всех пользователей."

	// Create keyboard with back button
	backBtn := tgbotapi.NewInlineKeyboardButtonData(
		i18n.Get(i18n.BtnBack, lang),
		"admin_back",
	)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{backBtn},
	)

	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleAdminBackCallback handles going back to admin panel
func HandleAdminBackCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Check if user is admin
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil {
		return err
	}

	if !isAdmin {
		text := "❌ Bu buyruq faqat ma'murlar uchun / Эта команда только для администраторов"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	lang := i18n.LanguageUzbek
	if user != nil {
		lang = i18n.GetLanguage(user.Language)
	}

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Show admin panel
	text := i18n.Get(i18n.MsgAdminPanel, lang)
	keyboard := utils.MakeAdminKeyboard(lang)

	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleAdminUploadTimetableCallback handles admin upload timetable callback
func HandleAdminUploadTimetableCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Convert callback to message for HandleUploadTimetableCommand
	message := &tgbotapi.Message{
		From: callback.From,
		Chat: &tgbotapi.Chat{
			ID: callback.Message.Chat.ID,
		},
	}

	return HandleUploadTimetableCommand(botService, message)
}

// HandleAdminPostAnnouncementCallback handles admin post announcement callback
func HandleAdminPostAnnouncementCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Convert callback to message for HandlePostAnnouncementCommand
	message := &tgbotapi.Message{
		From: callback.From,
		Chat: &tgbotapi.Chat{
			ID: callback.Message.Chat.ID,
		},
	}

	return HandlePostAnnouncementCommand(botService, message)
}

// HandleAdminProposalsCallback handles admin proposals list callback
func HandleAdminProposalsCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	chatID := callback.Message.Chat.ID

	// Get proposals with user info
	proposals, err := botService.ProposalService.GetAllProposals(10, 0)
	if err != nil {
		text := "Xatolik / Ошибка: " + err.Error()
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Count total proposals
	totalCount, _ := botService.ProposalService.CountProposals()

	// Format proposals list
	text := fmt.Sprintf("💡 Takliflar / Предложения\n\n")
	text += fmt.Sprintf("Jami / Всего: %d\n\n", totalCount)

	for i, p := range proposals {
		statusEmoji := "⏳"
		statusText := "Kutilmoqda / Ожидание"

		if p.Status == models.StatusReviewed {
			statusEmoji = "✅"
			statusText = "Ko'rib chiqildi / Рассмотрено"
		} else if p.Status == models.StatusArchived {
			statusEmoji = "📦"
			statusText = "Arxivlangan / Архивировано"
		}

		// Get user info
		user, _ := botService.UserService.GetUserByID(p.UserID)
		userName := "N/A"
		userClass := "N/A"
		if user != nil {
			userName = user.ChildName
			userClass = user.ChildClass
		}

		text += fmt.Sprintf("%d. %s #%d - %s %s\n", i+1, statusEmoji, p.ID, userName, userClass)
		preview := utils.TruncateText(p.ProposalText, 60)
		text += fmt.Sprintf("   💡 %s\n", preview)
		text += fmt.Sprintf("   📅 %s\n", utils.FormatDateTime(p.CreatedAt))
		text += fmt.Sprintf("   📊 %s\n\n", statusText)
	}

	if len(proposals) < totalCount {
		text += fmt.Sprintf("...va yana %d ta / ...и ещё %d", totalCount-len(proposals), totalCount-len(proposals))
	}

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleAdminViewTimetablesCallback handles admin view timetables callback
func HandleAdminViewTimetablesCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Check if user is admin
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil {
		return err
	}

	if !isAdmin {
		text := "❌ Bu buyruq faqat ma'murlar uchun / Эта команда только для администраторов"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	lang := i18n.LanguageUzbek
	if user != nil {
		lang = i18n.GetLanguage(user.Language)
	}

	// Get all timetables
	timetables, err := botService.TimetableRepo.GetAll(50, 0)
	if err != nil {
		text := "❌ Xatolik / Ошибка: " + err.Error()
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get all classes for mapping
	classes, err := botService.ClassRepo.GetAll()
	if err != nil {
		text := "❌ Xatolik / Ошибка: " + err.Error()
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Create class ID to name map
	classMap := make(map[int]string)
	for _, class := range classes {
		classMap[class.ID] = class.ClassName
	}

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Format timetables list
	text := "📅 <b>Dars jadvallari / Расписания уроков</b>\n\n"

	if len(timetables) == 0 {
		text += "Hozircha dars jadvallari yo'q / Пока нет расписаний\n\n"
		text += "Yangi dars jadvali yuklash uchun quyidagi tugmani bosing:"
	} else {
		text += fmt.Sprintf("Jami / Всего: %d\n\n", len(timetables))

		for i, timetable := range timetables {
			className := classMap[timetable.ClassID]
			if className == "" {
				className = fmt.Sprintf("ID:%d", timetable.ClassID)
			}

			text += fmt.Sprintf("%d. 📚 <b>%s</b>\n", i+1, className)
			text += fmt.Sprintf("   📄 %s\n", timetable.Filename)
			text += fmt.Sprintf("   📅 %s\n", utils.FormatDateTime(timetable.CreatedAt))
			text += fmt.Sprintf("   🆔 ID: %d\n\n", timetable.ID)
		}

		text += "\n👇 Dars jadvalini o'chirish uchun ID raqamini tanlang:"
	}

	// Create keyboard with timetable management options
	keyboard := makeTimetableManagementKeyboard(timetables, classMap, lang)

	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// makeTimetableManagementKeyboard creates keyboard for timetable management
func makeTimetableManagementKeyboard(timetables []*models.Timetable, classMap map[int]string, lang i18n.Language) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Add timetable buttons with delete options (2 per row)
	for i := 0; i < len(timetables); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// First timetable in row
		timetable := timetables[i]
		className := classMap[timetable.ClassID]
		if className == "" {
			className = fmt.Sprintf("ID:%d", timetable.ClassID)
		}
		deleteBtn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("🗑 %s (ID:%d)", className, timetable.ID),
			fmt.Sprintf("timetable_delete_%d", timetable.ID),
		)
		row = append(row, deleteBtn)

		// Second timetable in row (if exists)
		if i+1 < len(timetables) {
			timetable := timetables[i+1]
			className := classMap[timetable.ClassID]
			if className == "" {
				className = fmt.Sprintf("ID:%d", timetable.ClassID)
			}
			deleteBtn := tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("🗑 %s (ID:%d)", className, timetable.ID),
				fmt.Sprintf("timetable_delete_%d", timetable.ID),
			)
			row = append(row, deleteBtn)
		}

		rows = append(rows, row)
	}

	// Add back button
	backBtn := tgbotapi.NewInlineKeyboardButtonData(
		i18n.Get(i18n.BtnBack, lang),
		"admin_back",
	)
	rows = append(rows, []tgbotapi.InlineKeyboardButton{backBtn})

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// HandleTimetableDeleteCallback handles deleting a timetable
func HandleTimetableDeleteCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Check if user is admin
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil {
		return err
	}

	if !isAdmin {
		text := "❌ Faqat ma'murlar uchun / Только для администраторов"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	// Extract timetable ID from callback data
	var timetableID int
	fmt.Sscanf(callback.Data, "timetable_delete_%d", &timetableID)

	// Delete timetable
	err = botService.TimetableRepo.Delete(timetableID)
	if err != nil {
		text := "❌ Xatolik / Ошибка"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	// Answer callback with success
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅ Dars jadvali o'chirildi / Расписание удалено")

	// Refresh the timetable management view
	return HandleAdminViewTimetablesCallback(botService, callback)
}
