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

	// Set state to awaiting complaint
	stateData := &models.StateData{
		Language: user.Language,
	}
	err = botService.StateManager.Set(telegramID, models.StateAwaitingComplaint, stateData)
	if err != nil {
		return err
	}

	// Send request message
	text := i18n.Get(i18n.MsgRequestComplaint, lang)
	return botService.TelegramService.SendMessage(chatID, text, nil)
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

	// Get current selected student
	var student *models.StudentWithClass
	if user.CurrentSelectedStudentID != nil {
		student, err = botService.StudentService.GetStudentByIDWithClass(*user.CurrentSelectedStudentID)
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

	// Get user complaints
	complaints, err := botService.ComplaintService.GetUserComplaints(user.ID, 10, 0)
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(complaints) == 0 {
		text := "Sizda hali shikoyatlar yo'q / У вас пока нет жалоб"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Format complaints list
	text := "📋 Sizning shikoyatlaringiz / Ваши жалобы:\n\n"
	for i, c := range complaints {
		status := "⏳"
		if c.Status == models.StatusReviewed {
			status = "✅"
		}

		preview := utils.TruncateText(c.ComplaintText, 50)
		text += fmt.Sprintf("%d. %s %s\n   📅 %s\n\n",
			i+1,
			status,
			preview,
			utils.FormatDateTime(c.CreatedAt),
		)
	}

	return botService.TelegramService.SendMessage(chatID, text, nil)
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

	// Show current selected child if any
	if user.CurrentSelectedStudentID != nil {
		student, err := botService.StudentService.GetStudentByIDWithClass(*user.CurrentSelectedStudentID)
		if err == nil && student != nil {
			studentFullName := fmt.Sprintf("%s %s", student.LastName, student.FirstName)
			text += fmt.Sprintf("👤 Joriy farzand / Текущий ребенок: %s\n", studentFullName)
			text += fmt.Sprintf("🎓 Sinf / Класс: %s\n", student.ClassName)
		}
	}

	// Get all children
	children, err := botService.StudentService.GetParentStudents(user.ID)
	if err == nil && len(children) > 0 {
		text += fmt.Sprintf("\n👨‍👩‍👧‍👦 Barcha farzandlar / Все дети: %d\n", len(children))
	}

	text += fmt.Sprintf("\n📱 Telefon / Телефон: %s\n", utils.FormatPhoneNumber(user.PhoneNumber))
	text += fmt.Sprintf("🌍 Til / Язык: %s\n", user.Language)

	return botService.TelegramService.SendMessage(chatID, text, nil)
}
