package handlers

import (
	"fmt"
	"log"
	"time"

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

	// Check if there are no users
	if len(users) == 0 {
		text += "Hozircha foydalanuvchilar yo'q.\nПока нет пользователей."
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	for i, user := range users {
		// Get children count for this parent
		children, _ := botService.StudentService.GetParentStudents(user.ID)
		childrenCount := len(children)

		text += fmt.Sprintf("%d. 📱 %s\n", i+1, user.PhoneNumber)
		if user.TelegramUsername != "" {
			text += fmt.Sprintf("   @%s\n", user.TelegramUsername)
		}
		text += fmt.Sprintf("   👶 Farzandlar / Дети: %d\n", childrenCount)

		// Show children if any
		for j, child := range children {
			if j < 2 { // Show max 2 children in list
				text += fmt.Sprintf("      • %s %s (%s)\n", child.StudentLastName, child.StudentFirstName, child.ClassName)
			}
		}
		if childrenCount > 2 {
			text += fmt.Sprintf("      ...va yana %d ta / ...и ещё %d\n", childrenCount-2, childrenCount-2)
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

	// Check if there are no complaints
	if len(complaints) == 0 {
		text += "Hozircha shikoyatlar yo'q.\nПока нет жалоб."
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

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

		text += fmt.Sprintf("%d. %s #%d\n", i+1, statusEmoji, c.ID)
		text += fmt.Sprintf("   📱 %s", c.PhoneNumber)
		if c.TelegramUsername != "" {
			text += fmt.Sprintf(" (@%s)", c.TelegramUsername)
		}
		text += "\n"
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

	// Add each class with view and delete buttons
	for _, class := range classes {
		var row []tgbotapi.InlineKeyboardButton

		// Class name button (clickable to view details)
		nameBtn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("📚 %s", class.ClassName),
			fmt.Sprintf("admin_view_class_%d", class.ID),
		)
		row = append(row, nameBtn)

		// Delete button (separate)
		deleteBtn := tgbotapi.NewInlineKeyboardButtonData(
			"🗑",
			fmt.Sprintf("class_delete_%d", class.ID),
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

// HandleAdminViewClassCallback handles viewing class details with student management
func HandleAdminViewClassCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, classID int) error {
	chatID := callback.Message.Chat.ID
	telegramID := callback.From.ID

	// Check if user is admin
	user, _ := botService.UserService.GetUserByTelegramID(telegramID)
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, _ := botService.IsAdmin(phoneNumber, telegramID)
	if !isAdmin {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Bu buyruq faqat ma'murlar uchun")
		return nil
	}

	lang := i18n.LanguageUzbek
	if user != nil {
		lang = i18n.GetLanguage(user.Language)
	}

	// Get class info
	class, err := botService.ClassRepo.GetByID(classID)
	if err != nil {
		text := "❌ Sinf topilmadi / Класс не найден"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get students in this class
	students, err := botService.StudentService.GetStudentsByClassID(classID)
	if err != nil {
		text := "❌ Xatolik / Ошибка: " + err.Error()
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Format message
	text := fmt.Sprintf("📚 <b>Sinf: %s</b>\n\n", class.ClassName)
	text += fmt.Sprintf("👨‍🎓 O'quvchilar / Студенты: %d\n\n", len(students))

	if len(students) == 0 {
		text += "Hozircha o'quvchilar yo'q.\nПока нет студентов."
	} else {
		text += "<b>O'quvchilarni tanlang / Выберите ученика:</b>\n"
		text += "🗑 O'chirish uchun ismni bosing / Нажмите на имя для удаления"
	}

	// Create keyboard
	var rows [][]tgbotapi.InlineKeyboardButton

	// Add each student as a button with delete icon
	for _, student := range students {
		studentBtn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("🗑 %s %s", student.FirstName, student.LastName),
			fmt.Sprintf("admin_delete_student_%d_%d", classID, student.ID),
		)
		rows = append(rows, []tgbotapi.InlineKeyboardButton{studentBtn})
	}

	// Add student button
	addStudentBtn := tgbotapi.NewInlineKeyboardButtonData(
		"➕ O'quvchi qo'shish / Добавить студента",
		fmt.Sprintf("admin_add_student_%d", classID),
	)
	rows = append(rows, []tgbotapi.InlineKeyboardButton{addStudentBtn})

	// Back button
	backBtn := tgbotapi.NewInlineKeyboardButtonData(
		i18n.Get(i18n.BtnBack, lang),
		"admin_manage_classes",
	)
	rows = append(rows, []tgbotapi.InlineKeyboardButton{backBtn})

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleAdminAddStudentCallback handles adding a student to a class (admin)
func HandleAdminAddStudentCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, classID int) error {
	chatID := callback.Message.Chat.ID
	telegramID := callback.From.ID

	// Check if user is admin
	user, _ := botService.UserService.GetUserByTelegramID(telegramID)
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, _ := botService.IsAdmin(phoneNumber, telegramID)
	if !isAdmin {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Bu buyruq faqat ma'murlar uchun")
		return nil
	}

	// Get class info
	class, err := botService.ClassRepo.GetByID(classID)
	if err != nil {
		text := "❌ Sinf topilmadi / Класс не найден"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Set state for student name input (class already selected)
	stateData := &models.StateData{
		ClassID: &classID,
	}
	err = botService.StateManager.Set(telegramID, "awaiting_admin_student_name", stateData)
	if err != nil {
		return err
	}

	text := fmt.Sprintf("👨‍🎓 <b>%s sinfiga o'quvchi qo'shish</b>\n", class.ClassName) +
		fmt.Sprintf("<b>Добавить студента в класс %s</b>\n\n", class.ClassName) +
		"O'quvchining to'liq ismini kiriting:\n" +
		"Введите полное имя студента:\n\n" +
		"<b>Format / Формат:</b> Ism Familiya\n" +
		"<b>Misol / Пример:</b> Jasur Rahimov"

	// Create cancel button
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"❌ Bekor qilish / Отмена",
				fmt.Sprintf("admin_view_class_%d", classID),
			),
		),
	)

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err = botService.Bot.Send(msg)
	return err
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

	// Debug log
	fmt.Printf("[DEBUG] HandleClassDeleteCallback called. CallbackData: %s, UserID: %d\n", callback.Data, telegramID)

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

	// Extract class ID from callback data
	var classID int
	n, err := fmt.Sscanf(callback.Data, "class_delete_%d", &classID)
	fmt.Printf("[DEBUG] Parsed class ID: %d, n=%d, err=%v\n", classID, n, err)

	if err != nil || n != 1 || classID == 0 {
		text := "❌ Noto'g'ri ma'lumot / Неверные данные"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return fmt.Errorf("failed to parse class ID from callback data: %s", callback.Data)
	}

	// Delete class
	fmt.Printf("[DEBUG] Attempting to delete class with ID: %d\n", classID)
	err = botService.ClassRepo.DeleteByID(classID)
	if err != nil {
		fmt.Printf("[DEBUG] Delete failed: %v\n", err)
	} else {
		fmt.Printf("[DEBUG] Delete successful for class ID: %d\n", classID)
	}
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

	// Check if there are no proposals
	if len(proposals) == 0 {
		text += "Hozircha takliflar yo'q.\nПока нет предложений."
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

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
		userPhone := "N/A"
		if user != nil {
			userPhone = user.PhoneNumber
		}

		text += fmt.Sprintf("%d. %s #%d\n", i+1, statusEmoji, p.ID)
		text += fmt.Sprintf("   📱 %s\n", userPhone)
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

// HandleAdminManageTeachersCallback handles admin manage teachers callback
func HandleAdminManageTeachersCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	chatID := callback.Message.Chat.ID

	// Get all teachers
	teachers, err := botService.TeacherRepo.GetAll(100, 0)
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Format teacher list
	text := "👥 <b>O'qituvchilarni boshqarish / Управление учителями</b>\n\n"

	if len(teachers) == 0 {
		text += "📝 Hozircha o'qituvchilar yo'q.\n\n📝 Пока нет учителей.\n\n"
		text += "Yangi o'qituvchi qo'shish uchun quyidagi tugmani bosing:"
	} else {
		text += fmt.Sprintf("Jami: <b>%d</b> ta\n\n", len(teachers))
		text += "🗑 O'chirish uchun o'qituvchini tanlang:\n"
		text += "🗑 Выберите учителя для удаления:\n\n"
	}

	// Create keyboard
	var rows [][]tgbotapi.InlineKeyboardButton

	// Add delete button for each teacher
	for _, t := range teachers {
		status := "✅"
		if !t.IsActive {
			status = "❌"
		} else if t.TelegramID == nil {
			status = "⏳"
		}

		btnText := fmt.Sprintf("🗑 %s %s %s (%s)", status, t.FirstName, t.LastName, t.PhoneNumber)
		deleteBtn := tgbotapi.NewInlineKeyboardButtonData(
			btnText,
			fmt.Sprintf("admin_delete_teacher_%d", t.ID),
		)
		rows = append(rows, []tgbotapi.InlineKeyboardButton{deleteBtn})
	}

	// Add teacher button
	addBtn := tgbotapi.NewInlineKeyboardButtonData(
		"➕ O'qituvchi qo'shish / Добавить учителя",
		"admin_add_teacher",
	)
	rows = append(rows, []tgbotapi.InlineKeyboardButton{addBtn})

	// Add back button
	backBtn := tgbotapi.NewInlineKeyboardButtonData(
		"⬅️ Orqaga / Назад",
		"admin_back",
	)
	rows = append(rows, []tgbotapi.InlineKeyboardButton{backBtn})

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	return botService.TelegramService.SendMessage(chatID, text, &keyboard)
}

// HandleAdminDeleteTeacherCallback handles admin delete teacher callback
func HandleAdminDeleteTeacherCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, teacherID int) error {
	// Get teacher info before deleting
	teacher, err := botService.TeacherRepo.GetByID(teacherID)
	if err != nil || teacher == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ O'qituvchi topilmadi")
		return nil
	}

	// Delete the teacher
	err = botService.TeacherRepo.Delete(teacherID)
	if err != nil {
		log.Printf("Failed to delete teacher %d: %v", teacherID, err)
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ O'chirishda xatolik")
		return nil
	}

	// Success notification
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, fmt.Sprintf("✅ %s %s o'chirildi", teacher.FirstName, teacher.LastName))

	// Refresh the teacher list
	return HandleAdminManageTeachersCallback(botService, callback)
}

// HandleAdminAddTeacherCallback handles admin add teacher callback
func HandleAdminAddTeacherCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	// Convert callback to message for the add teacher command
	message := &tgbotapi.Message{
		From: callback.From,
		Chat: &tgbotapi.Chat{
			ID: callback.Message.Chat.ID,
		},
	}

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
	return HandleAddTeacherCommand(botService, message)
}

// HandleAdminExportAttendanceCallback handles admin export attendance callback
func HandleAdminExportAttendanceCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	chatID := callback.Message.Chat.ID

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Get today's attendance for all classes
	classes, err := botService.ClassRepo.GetAll()
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(classes) == 0 {
		text := "❌ Sinflar topilmadi / Классы не найдены"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Use Uzbekistan timezone to match attendance taking
	location, _ := time.LoadLocation("Asia/Tashkent")
	today := time.Now().In(location)

	text := fmt.Sprintf("📋 <b>Bugungi davomat / Сегодняшняя посещаемость</b>\n📅 <b>%s</b>\n\n", today.Format("02.01.2006"))

	totalPresent := 0
	totalAbsent := 0

	for _, class := range classes {
		// Get today's attendance for this class
		attendance, err := botService.AttendanceRepo.GetTodayAttendanceByClass(class.ID)
		if err != nil {
			continue
		}

		present := 0
		absent := 0
		var absentStudents []string

		for _, a := range attendance {
			if a.Status == "present" {
				present++
			} else {
				absent++
				absentStudents = append(absentStudents, fmt.Sprintf("%s %s", a.FirstName, a.LastName))
			}
		}

		totalPresent += present
		totalAbsent += absent

		if present > 0 || absent > 0 {
			text += fmt.Sprintf("📚 <b>%s</b>: ✅ %d | ❌ %d\n", class.ClassName, present, absent)
			if len(absentStudents) > 0 {
				text += "   <i>Kelmadi / Отсутствуют:</i>\n"
				for i, name := range absentStudents {
					text += fmt.Sprintf("   %d. %s\n", i+1, name)
				}
			}
			text += "\n"
		} else {
			text += fmt.Sprintf("📚 <b>%s</b>: <i>davomat olinmagan</i>\n\n", class.ClassName)
		}
	}

	// Add totals
	text += fmt.Sprintf("━━━━━━━━━━━━━━━━━━━━\n")
	text += fmt.Sprintf("📊 <b>Jami / Всего:</b> ✅ %d | ❌ %d", totalPresent, totalAbsent)

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleAdminExportTestResultsCallback handles admin export test results callback
func HandleAdminExportTestResultsCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	chatID := callback.Message.Chat.ID

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Get all classes for selection
	classes, err := botService.ClassRepo.GetAll()
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(classes) == 0 {
		text := "❌ Sinflar topilmadi / Классы не найдены"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	text := "📊 <b>Test natijalarini eksport qilish / Экспорт результатов тестов</b>\n\n" +
		"Sinfni tanlang / Выберите класс:"

	var rows [][]tgbotapi.InlineKeyboardButton

	for _, class := range classes {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("📚 %s", class.ClassName),
			fmt.Sprintf("export_grades_select_%d", class.ID),
		)
		rows = append(rows, []tgbotapi.InlineKeyboardButton{btn})
	}

	// Back button
	backBtn := tgbotapi.NewInlineKeyboardButtonData(
		"⬅️ Orqaga / Назад",
		"admin_back",
	)
	rows = append(rows, []tgbotapi.InlineKeyboardButton{backBtn})

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	return botService.TelegramService.SendMessage(chatID, text, &keyboard)
}

// HandleAdminExportGradesSelectClassCallback handles class selection for grade export
func HandleAdminExportGradesSelectClassCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, classID int) error {
	chatID := callback.Message.Chat.ID

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Get class info
	class, err := botService.ClassRepo.GetByID(classID)
	if err != nil || class == nil {
		text := "❌ Sinf topilmadi / Класс не найден"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	text := fmt.Sprintf("📊 <b>%s sinfi natijalarini eksport</b>\n\n"+
		"Vaqt oralig'ini tanlang / Выберите период:", class.ClassName)

	var rows [][]tgbotapi.InlineKeyboardButton

	// Date range options
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("📅 Oxirgi 7 kun / Последние 7 дней", fmt.Sprintf("export_grades_%d", classID)),
	})
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("📅 Barchasi / Все", fmt.Sprintf("export_grades_%d", classID)),
	})

	// Back button
	backBtn := tgbotapi.NewInlineKeyboardButtonData(
		"⬅️ Orqaga / Назад",
		"admin_export_test_results",
	)
	rows = append(rows, []tgbotapi.InlineKeyboardButton{backBtn})

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	return botService.TelegramService.SendMessage(chatID, text, &keyboard)
}

// HandleAdminExportGradesCustomDateCallback handles custom date input for grade export
func HandleAdminExportGradesCustomDateCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, classID int) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Set state for custom date input
	stateData := &models.StateData{
		ClassID: &classID,
	}
	_ = botService.StateManager.Set(telegramID, "admin_awaiting_export_custom_dates", stateData)

	text := "📅 <b>Vaqt oralig'ini kiriting / Введите период</b>\n\n" +
		"Format: <code>YYYY-MM-DD YYYY-MM-DD</code>\n\n" +
		"Misol / Пример: <code>2025-01-01 2025-12-31</code>"

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleAdminExportCustomDatesInput handles custom date input for exports
func HandleAdminExportCustomDatesInput(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	if stateData.ClassID == nil {
		text := "❌ Sessiya tugagan / Сессия истекла"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// For now, just export all grades
	return HandleAdminExportGradesCallback(botService, &tgbotapi.CallbackQuery{
		From:    message.From,
		Message: &tgbotapi.Message{Chat: message.Chat},
	}, *stateData.ClassID)
}

// HandleAdminExportGradesCallback handles exporting grades for a class
func HandleAdminExportGradesCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, classID int) error {
	chatID := callback.Message.Chat.ID

	if callback.ID != "" {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
	}

	// Get class info
	class, err := botService.ClassRepo.GetByID(classID)
	if err != nil || class == nil {
		text := "❌ Sinf topilmadi / Класс не найден"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get test results for the class
	results, err := botService.TestResultRepo.GetAllByClassID(classID)
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(results) == 0 {
		text := fmt.Sprintf("📊 <b>%s</b> sinfida test natijalari topilmadi.\n\n"+
			"📊 Результаты тестов для класса <b>%s</b> не найдены.", class.ClassName, class.ClassName)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Group results by student
	type studentResults struct {
		Name    string
		Results []struct {
			Subject string
			Score   string
			Date    string
		}
	}
	studentMap := make(map[int]*studentResults)
	var studentOrder []int // to preserve order

	for _, result := range results {
		if _, exists := studentMap[result.StudentID]; !exists {
			studentMap[result.StudentID] = &studentResults{
				Name: fmt.Sprintf("%s %s", result.LastName, result.FirstName),
			}
			studentOrder = append(studentOrder, result.StudentID)
		}
		studentMap[result.StudentID].Results = append(studentMap[result.StudentID].Results, struct {
			Subject string
			Score   string
			Date    string
		}{
			Subject: result.SubjectName,
			Score:   result.Score,
			Date:    result.TestDate.Format("02.01.2006"),
		})
	}

	// Format results as text grouped by student
	text := fmt.Sprintf("📊 <b>%s sinfi test natijalari</b>\n", class.ClassName)
	text += fmt.Sprintf("📊 <b>Результаты тестов класса %s</b>\n\n", class.ClassName)

	for i, studentID := range studentOrder {
		student := studentMap[studentID]
		text += fmt.Sprintf("%d. 👤 <b>%s</b>\n", i+1, student.Name)
		for _, r := range student.Results {
			text += fmt.Sprintf("   • %s: <b>%s</b> (%s)\n", r.Subject, r.Score, r.Date)
		}
		text += "\n"
	}

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleAdminDeleteStudentCallback handles deleting a student (admin)
func HandleAdminDeleteStudentCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, classID, studentID int) error {
	telegramID := callback.From.ID

	// Check if user is admin
	user, _ := botService.UserService.GetUserByTelegramID(telegramID)
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, _ := botService.IsAdmin(phoneNumber, telegramID)
	if !isAdmin {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Faqat ma'murlar uchun")
		return nil
	}

	// Get student info before deleting
	student, err := botService.StudentRepo.GetByID(studentID)
	if err != nil || student == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ O'quvchi topilmadi")
		return nil
	}

	// Delete the student
	err = botService.StudentRepo.Delete(studentID)
	if err != nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik yuz berdi")
		return nil
	}

	// Success feedback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, fmt.Sprintf("✅ %s %s o'chirildi", student.FirstName, student.LastName))

	// Refresh the class view
	return HandleAdminViewClassCallback(botService, callback, classID)
}
