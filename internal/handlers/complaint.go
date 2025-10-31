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

	// Validate complaint text
	complaintText, err := validator.ValidateComplaintText(message.Text)
	if err != nil {
		text := i18n.Get(i18n.ErrInvalidComplaint, lang) + "\n\n" + err.Error()
		return botService.TelegramService.SendMessage(chatID, text, nil)
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

	// Generate DOCX document
	docPath, filename, err := botService.DocumentService.GenerateComplaintDocument(user, stateData.ComplaintText)
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
	go notifyAdminsWithDocument(botService, user, complaint, fileID)

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
func notifyAdminsWithDocument(botService *services.BotService, user *models.User, complaint *models.Complaint, fileID string) {
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
		user.ChildName,
		user.ChildClass,
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

	lang := i18n.GetLanguage(user.Language)
	_ = lang // Will be used in future for localized settings

	// Format user info
	text := "⚙️ Sozlamalar / Настройки\n\n"
	text += fmt.Sprintf("👤 Farzand / Ребенок: %s\n", user.ChildName)
	text += fmt.Sprintf("🎓 Sinf / Класс: %s\n", user.ChildClass)
	text += fmt.Sprintf("📱 Telefon / Телефон: %s\n", utils.FormatPhoneNumber(user.PhoneNumber))
	text += fmt.Sprintf("🌍 Til / Язык: %s\n", user.Language)

	return botService.TelegramService.SendMessage(chatID, text, nil)
}
