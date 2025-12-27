package handlers

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
)

// HandleViewTimetableCommand shows timetable for user's class
func HandleViewTimetableCommand(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	if user == nil {
		lang := i18n.LanguageUzbek
		text := i18n.Get(i18n.ErrNotRegistered, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	lang := i18n.GetLanguage(user.Language)

	// Get parent's children
	children, err := botService.StudentRepo.GetParentStudents(user.ID)
	if err != nil {
		return err
	}

	if len(children) == 0 {
		text := "⚠️ Sizda hali bog'langan farzand yo'q.\n\n⚠️ У вас еще нет привязанных детей."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// If multiple children, show selection
	if len(children) > 1 {
		text := "📅 <b>Dars jadvali / Расписание уроков</b>\n\n" +
			"Qaysi farzandingiz jadvalini ko'rmoqchisiz?\n" +
			"Расписание какого ребенка хотите посмотреть?"

		var buttons [][]tgbotapi.InlineKeyboardButton
		for _, child := range children {
			buttonText := fmt.Sprintf("%s %s (%s)", child.StudentLastName, child.StudentFirstName, child.ClassName)
			button := tgbotapi.NewInlineKeyboardButtonData(
				buttonText,
				fmt.Sprintf("timetable_child_%d", child.StudentID),
			)
			buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
		}

		keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = keyboard

		_, err = botService.Bot.Send(msg)
		return err
	}

	// Single child - show timetable directly
	return showTimetableForStudent(botService, chatID, children[0].StudentID, lang)
}

// showTimetableForStudent displays timetable for a specific student
func showTimetableForStudent(botService *services.BotService, chatID int64, studentID int, lang i18n.Language) error {
	student, err := botService.StudentRepo.GetByID(studentID)
	if err != nil || student == nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get class information
	class, err := botService.ClassRepo.GetByID(student.ClassID)
	if err != nil || class == nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get timetable for student's class
	timetable, err := botService.TimetableService.GetTimetableByClassName(class.ClassName)
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if timetable == nil {
		text := i18n.Get(i18n.MsgTimetableNotFound, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Send timetable file
	caption := fmt.Sprintf("📅 Dars jadvali / Расписание уроков\nSinf / Класс: %s\nO'quvchi / Ученик: %s %s",
		class.ClassName, student.LastName, student.FirstName)
	return botService.TelegramService.SendDocumentByFileID(chatID, timetable.TelegramFileID, caption)
}

// HandleTimetableChildSelection handles child selection for timetable view
func HandleTimetableChildSelection(botService *services.BotService, callback *tgbotapi.CallbackQuery, studentID int) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil || user == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	lang := i18n.GetLanguage(user.Language)

	// Verify student belongs to this parent
	children, err := botService.StudentRepo.GetParentStudents(user.ID)
	if err != nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	studentBelongsToParent := false
	for _, child := range children {
		if child.StudentID == studentID {
			studentBelongsToParent = true
			break
		}
	}

	if !studentBelongsToParent {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Bu farzand sizga tegishli emas")
		return nil
	}

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	return showTimetableForStudent(botService, chatID, studentID, lang)
}

// HandleUploadTimetableCommand initiates timetable upload (admin only)
func HandleUploadTimetableCommand(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Get user (may be nil for admin-only accounts)
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Check if user is admin (works with or without user registration)
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil || !isAdmin {
		text := "❌ Bu buyruq faqat ma'murlar uchun / Эта команда только для администраторов"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Determine language
	lang := i18n.LanguageUzbek
	if user != nil {
		lang = i18n.GetLanguage(user.Language)
	}

	// Get all classes
	classes, err := botService.ClassRepo.GetAll()
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(classes) == 0 {
		text := "Sinflar topilmadi. Avval sinf yarating. / Классы не найдены. Сначала создайте класс."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Create keyboard with classes
	keyboard := tgbotapi.NewInlineKeyboardMarkup()
	for _, class := range classes {
		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("📚 %s", class.ClassName),
				fmt.Sprintf("timetable_select_%d", class.ID),
			),
		)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}

	text := i18n.Get(i18n.MsgSelectClassForTimetable, lang)
	return botService.TelegramService.SendMessage(chatID, text, &keyboard)
}

// HandleTimetableClassSelection handles class selection for timetable upload
func HandleTimetableClassSelection(botService *services.BotService, callback *tgbotapi.CallbackQuery, classID int) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get user (may be nil for admin-only accounts)
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Check if user is admin (works with or without user registration)
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil || !isAdmin {
		return botService.TelegramService.AnswerCallbackQuery(callback.ID, "Access denied")
	}

	// Determine language
	lang := i18n.LanguageUzbek
	langStr := "uz"
	if user != nil {
		lang = i18n.GetLanguage(user.Language)
		langStr = user.Language
	}

	// Save class ID in state
	stateData := &models.StateData{
		Language: langStr,
		ClassID:  &classID,
	}
	err = botService.StateManager.Set(telegramID, models.StateAwaitingTimetableFile, stateData)
	if err != nil {
		return err
	}

	// Answer callback query
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅")

	// Send request for file
	text := i18n.Get(i18n.MsgUploadTimetableFile, lang)
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleTimetableFileUpload handles timetable file upload
func HandleTimetableFileUpload(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID
	lang := i18n.GetLanguage(stateData.Language)

	var fileID, filename, fileType, mimeType string

	// Check what type of file was sent
	if message.Document != nil {
		// Document (PDF, Word, Excel, etc.)
		fileID = message.Document.FileID
		filename = message.Document.FileName
		mimeType = message.Document.MimeType
		fileType = "document"
	} else if len(message.Photo) > 0 {
		// Photo/Image
		photo := message.Photo[len(message.Photo)-1] // Get largest photo
		fileID = photo.FileID
		filename = fmt.Sprintf("timetable_%d.jpg", stateData.ClassID)
		mimeType = "image/jpeg"
		fileType = "image"
	} else {
		text := i18n.Get(i18n.ErrInvalidFile, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get admin record
	admin, err := botService.AdminRepo.GetByTelegramID(telegramID)
	if err != nil {
		log.Printf("Failed to get admin: %v", err)
	}

	var adminID *int
	if admin != nil {
		adminID = &admin.ID
	}

	// Create timetable record
	timetableReq := &models.CreateTimetableRequest{
		ClassID:           *stateData.ClassID,
		TelegramFileID:    fileID,
		Filename:          filename,
		FileType:          fileType,
		MimeType:          mimeType,
		UploadedByAdminID: adminID,
	}

	_, err = botService.TimetableService.CreateTimetable(timetableReq)
	if err != nil {
		log.Printf("Failed to save timetable: %v", err)
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Send success message
	text := i18n.Get(i18n.MsgTimetableUploaded, lang)
	return botService.TelegramService.SendMessage(chatID, text, nil)
}
