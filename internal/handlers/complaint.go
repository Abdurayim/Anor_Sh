package handlers

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
	"parent-bot/internal/utils"
	"parent-bot/internal/validator"
)

// HandleComplaintCommand initiates complaint submission
func HandleComplaintCommand(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Check if user is registered
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

	// If only one child, use that child automatically
	if len(children) == 1 {
		stateData := &models.StateData{
			Language:          user.Language,
			SelectedStudentID: &children[0].StudentID,
		}
		err = botService.StateManager.Set(telegramID, models.StateAwaitingComplaint, stateData)
		if err != nil {
			return err
		}

		text := i18n.Get(i18n.MsgRequestComplaint, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Multiple children - show selection
	text := "👨‍👩‍👧‍👦 <b>Shikoyatni qaysi farzandingiz uchun yozmoqchisiz?</b>\n\n" +
		"👨‍👩‍👧‍👦 <b>На какого ребенка хотите написать жалобу?</b>"

	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, child := range children {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s %s (%s)", child.StudentFirstName, child.StudentLastName, child.ClassName),
			fmt.Sprintf("complaint_select_child_%d", child.StudentID),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	// Set state to selecting child for complaint
	stateData := &models.StateData{
		Language: user.Language,
	}
	err = botService.StateManager.Set(telegramID, "selecting_child_for_complaint", stateData)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard
	_, err = botService.Bot.Send(msg)
	return err
}

// HandleComplaintText handles complaint text input
func HandleComplaintText(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID
	lang := i18n.GetLanguage(stateData.Language)

	// Get user to check admin status for keyboard
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	var isAdmin bool
	if user != nil {
		isAdmin, _ = botService.IsAdmin(user.PhoneNumber, user.TelegramID)
	}

	// Check if message contains media instead of text
	if message.Text == "" {
		var errorMsg string
		if len(message.Photo) > 0 {
			errorMsg = "❌ Iltimos, rasm emas, matn yuboring!\n\n❌ Пожалуйста, отправьте текст, а не изображение!"
		} else if message.Animation != nil {
			errorMsg = "❌ Iltimos, GIF emas, matn yuboring!\n\n❌ Пожалуйста, отправьте текст, а не GIF!"
		} else if message.Video != nil {
			errorMsg = "❌ Iltimos, video emas, matn yuboring!\n\n❌ Пожалуйста, отправьте текст, а не видео!"
		} else if message.Document != nil {
			errorMsg = "❌ Iltimos, fayl emas, matn yuboring!\n\n❌ Пожалуйста, отправьте текст, а не файл!"
		} else if message.Sticker != nil {
			errorMsg = "❌ Iltimos, stiker emas, matn yuboring!\n\n❌ Пожалуйста, отправьте текст, а не стикер!"
		} else if message.Voice != nil {
			errorMsg = "❌ Iltimos, ovozli xabar emas, matn yuboring!\n\n❌ Пожалуйста, отправьте текст, а не голосовое сообщение!"
		} else {
			errorMsg = "❌ Iltimos, shikoyat matnini yuboring!\n\n❌ Пожалуйста, отправьте текст жалобы!"
		}

		// Keep the main menu keyboard visible
		keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)
		return botService.TelegramService.SendMessage(chatID, errorMsg, &keyboard)
	}

	// Validate complaint text
	complaintText, err := validator.ValidateComplaintText(message.Text)
	if err != nil {
		text := i18n.Get(i18n.ErrInvalidComplaint, lang) + "\n\n" + err.Error()
		// Keep the main menu keyboard visible on validation errors too
		keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)
		return botService.TelegramService.SendMessage(chatID, text, &keyboard)
	}

	// Save complaint text in state
	stateData.ComplaintText = complaintText
	err = botService.StateManager.Set(telegramID, models.StateConfirmingComplaint, stateData)
	if err != nil {
		return err
	}

	// Show preview and confirmation
	text := fmt.Sprintf(i18n.Get(i18n.MsgConfirmComplaint, lang), complaintText)
	keyboard := utils.MakeConfirmationKeyboard(lang)

	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleComplaintConfirmation handles complaint confirmation
func HandleComplaintConfirmation(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	if user == nil {
		return botService.TelegramService.AnswerCallbackQuery(callback.ID, "User not found")
	}

	lang := i18n.GetLanguage(user.Language)

	// Get complaint text from state
	stateData, err := botService.StateManager.GetData(telegramID)
	if err != nil {
		return err
	}

	if stateData.ComplaintText == "" {
		return botService.TelegramService.AnswerCallbackQuery(callback.ID, "Complaint text not found")
	}

	// Answer callback query
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅")

	// Get selected student from state data
	var student *models.StudentWithClass
	if stateData.SelectedStudentID != nil {
		student, err = botService.StudentService.GetStudentByIDWithClass(*stateData.SelectedStudentID)
		if err != nil || student == nil {
			log.Printf("Failed to get student: %v", err)
			text := "⚠️ Iltimos, avval farzandingizni tanlang / Пожалуйста, сначала выберите ребенка"
			return botService.TelegramService.SendMessage(chatID, text, nil)
		}
	} else {
		text := "⚠️ Iltimos, avval farzandingizni tanlang / Пожалуйста, сначала выберите ребенка"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Generate DOCX document
	docPath, filename, err := botService.DocumentService.GenerateComplaintDocument(user, student, stateData.ComplaintText)
	if err != nil {
		log.Printf("Failed to generate document: %v", err)
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Upload document to Telegram and get file_id
	fileID, err := botService.TelegramService.UploadDocument(chatID, docPath, filename)
	if err != nil {
		log.Printf("Failed to upload document: %v", err)
		// Clean up temp file
		_ = botService.DocumentService.DeleteTempFile(docPath)
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Clean up temp file after upload
	_ = botService.DocumentService.DeleteTempFile(docPath)

	// Save complaint to database with document info
	complaintReq := &models.CreateComplaintRequest{
		UserID:         user.ID,
		ComplaintText:  stateData.ComplaintText,
		TelegramFileID: fileID,
		Filename:       filename,
	}

	complaint, err := botService.ComplaintService.CreateComplaint(complaintReq)
	if err != nil {
		log.Printf("Failed to save complaint: %v", err)
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Send success message
	text := i18n.Get(i18n.MsgComplaintSubmitted, lang)
	keyboard := utils.MakeMainMenuKeyboard(lang)
	_ = botService.TelegramService.SendMessage(chatID, text, keyboard)

	// Notify admins with DOCX document
	go notifyAdminsWithDocument(botService, user, student, complaint, fileID)

	return nil
}

// HandleComplaintCancellation handles complaint cancellation
func HandleComplaintCancellation(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	if user == nil {
		return botService.TelegramService.AnswerCallbackQuery(callback.ID, "User not found")
	}

	lang := i18n.GetLanguage(user.Language)

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Answer callback query
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, i18n.Get(i18n.MsgComplaintCancelled, lang))

	// Send cancellation message
	text := i18n.Get(i18n.MsgComplaintCancelled, lang)
	keyboard := utils.MakeMainMenuKeyboard(lang)

	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// notifyAdminsWithDocument sends complaint as DOCX document to all admins
func notifyAdminsWithDocument(botService *services.BotService, user *models.User, student *models.StudentWithClass, complaint *models.Complaint, fileID string) {
	// Get admin telegram IDs
	adminIDs, err := botService.GetAdminTelegramIDs()
	if err != nil {
		log.Printf("Failed to get admin IDs: %v", err)
		return
	}

	if len(adminIDs) == 0 {
		log.Println("No admins configured")
		return
	}

	// Generate caption for the document
	username := user.TelegramUsername
	if username == "" {
		username = "yo'q / нет"
	}

	studentFullName := fmt.Sprintf("%s %s", student.LastName, student.FirstName)
	caption := fmt.Sprintf(
		"<b>YANGI SHIKOYAT / НОВАЯ ЖАЛОБА</b>\n\n"+
			"ID: #%d\n"+
			"Farzand / Ребенок: <b>%s</b>\n"+
			"Sinf / Класс: <b>%s</b>\n"+
			"Telefon / Телефон: %s\n"+
			"Username: @%s\n"+
			"Sana / Дата: %s\n\n"+
			"Shikoyat hujjat sifatida yuqorida\n"+
			"Жалоба в формате документа выше",
		complaint.ID,
		studentFullName,
		student.ClassName,
		user.PhoneNumber,
		username,
		utils.FormatDateTime(complaint.CreatedAt),
	)

	// Send document to all admins
	err = botService.TelegramService.SendDocumentToAdmins(adminIDs, fileID, caption)
	if err != nil {
		log.Printf("Failed to send document to admins: %v", err)
	}
}

// HandleMyComplaintsCommand shows user's complaint history
func HandleMyComplaintsCommand(botService *services.BotService, message *tgbotapi.Message) error {
	return handleComplaintsPage(botService, message.From.ID, message.Chat.ID, 0)
}

// handleComplaintsPage shows complaints with pagination
func handleComplaintsPage(botService *services.BotService, telegramID int64, chatID int64, offset int) error {
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

	// Get user complaints with pagination (10 per page)
	const pageSize = 10
	complaints, err := botService.ComplaintService.GetUserComplaints(user.ID, pageSize, offset)
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(complaints) == 0 && offset == 0 {
		text := "Sizda hali shikoyatlar yo'q / У вас пока нет жалоб"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Format complaints list
	currentPage := (offset / pageSize) + 1
	text := fmt.Sprintf("📋 Sizning shikoyatlaringiz / Ваши жалобы (sahifa %d):\n\n", currentPage)

	for i, c := range complaints {
		status := "⏳"
		if c.Status == models.StatusReviewed {
			status = "✅"
		}

		preview := utils.TruncateText(c.ComplaintText, 50)
		text += fmt.Sprintf("%d. %s %s\n   📅 %s\n\n",
			offset+i+1,
			status,
			preview,
			utils.FormatDateTime(c.CreatedAt),
		)
	}

	// Add pagination buttons if needed
	var buttons [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	// Previous button
	if offset > 0 {
		prevOffset := offset - pageSize
		if prevOffset < 0 {
			prevOffset = 0
		}
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(
			"◀️ Oldingi / Предыдущая",
			fmt.Sprintf("complaints_page_%d", prevOffset),
		))
	}

	// Next button (show if we got full page)
	if len(complaints) == pageSize {
		nextOffset := offset + pageSize
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(
			"Keyingi / Следующая ▶️",
			fmt.Sprintf("complaints_page_%d", nextOffset),
		))
	}

	if len(row) > 0 {
		buttons = append(buttons, row)
	}

	if len(buttons) > 0 {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = keyboard
		_, err = botService.Bot.Send(msg)
		return err
	}

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleComplaintsPageCallback handles pagination for complaints
func HandleComplaintsPageCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, offset int) error {
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Delete old message
	deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
	_, _ = botService.Bot.Request(deleteMsg)

	// Show new page
	return handleComplaintsPage(botService, callback.From.ID, callback.Message.Chat.ID, offset)
}

// HandleSettingsCommand shows settings menu
func HandleSettingsCommand(botService *services.BotService, message *tgbotapi.Message) error {
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

	// Format user info
	text := "⚙️ Sozlamalar / Настройки\n\n"

	// Get all children
	children, err := botService.StudentService.GetParentStudents(user.ID)
	if err == nil && len(children) > 0 {
		text += fmt.Sprintf("👨‍👩‍👧‍👦 Barcha farzandlar / Все дети: %d\n\n", len(children))
	}

	text += fmt.Sprintf("📱 Telefon / Телефон: %s\n", utils.FormatPhoneNumber(user.PhoneNumber))
	text += fmt.Sprintf("🌍 Til / Язык: %s\n", user.Language)

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleComplaintSelectChildCallback handles child selection for complaint
func HandleComplaintSelectChildCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, studentID int) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil || user == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik")
		return nil
	}

	lang := i18n.GetLanguage(user.Language)

	// Verify student belongs to parent
	children, err := botService.StudentRepo.GetParentStudents(user.ID)
	if err != nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik")
		return nil
	}

	found := false
	for _, child := range children {
		if child.StudentID == studentID {
			found = true
			break
		}
	}

	if !found {
		text := "❌ Bu farzand sizga tegishli emas / Этот ребенок вам не принадлежит"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	// Set state with selected student
	stateData := &models.StateData{
		Language:          user.Language,
		SelectedStudentID: &studentID,
	}
	err = botService.StateManager.Set(telegramID, models.StateAwaitingComplaint, stateData)
	if err != nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik")
		return nil
	}

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅")

	// Delete the selection message
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
	_, _ = botService.Bot.Request(deleteMsg)

	// Send request for complaint text
	text := i18n.Get(i18n.MsgRequestComplaint, lang)
	return botService.TelegramService.SendMessage(chatID, text, nil)
}
