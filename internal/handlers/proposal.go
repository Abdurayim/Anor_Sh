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

// HandleProposalCommand initiates proposal submission
func HandleProposalCommand(botService *services.BotService, message *tgbotapi.Message) error {
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

	// Set state to awaiting proposal
	stateData := &models.StateData{
		Language: user.Language,
	}
	err = botService.StateManager.Set(telegramID, models.StateAwaitingProposal, stateData)
	if err != nil {
		return err
	}

	// Send request message
	text := i18n.Get(i18n.MsgRequestProposal, lang)
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleProposalText handles proposal text input
func HandleProposalText(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
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
			errorMsg = "❌ Iltimos, taklif matnini yuboring!\n\n❌ Пожалуйста, отправьте текст предложения!"
		}

		// Keep the main menu keyboard visible
		keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)
		return botService.TelegramService.SendMessage(chatID, errorMsg, &keyboard)
	}

	// Validate proposal text (using same validator as complaint)
	proposalText, err := validator.ValidateComplaintText(message.Text)
	if err != nil {
		text := i18n.Get(i18n.ErrInvalidProposal, lang) + "\n\n" + err.Error()
		// Keep the main menu keyboard visible on validation errors too
		keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)
		return botService.TelegramService.SendMessage(chatID, text, &keyboard)
	}

	// Save proposal text in state
	stateData.ProposalText = proposalText
	err = botService.StateManager.Set(telegramID, models.StateConfirmingProposal, stateData)
	if err != nil {
		return err
	}

	// Show preview and confirmation
	text := fmt.Sprintf(i18n.Get(i18n.MsgConfirmProposal, lang), proposalText)
	keyboard := utils.MakeProposalConfirmationKeyboard(lang)

	return botService.TelegramService.SendMessage(chatID, text, &keyboard)
}

// HandleProposalConfirmation handles proposal confirmation
func HandleProposalConfirmation(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
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

	// Get proposal text from state
	stateData, err := botService.StateManager.GetData(telegramID)
	if err != nil {
		return err
	}

	if stateData.ProposalText == "" {
		return botService.TelegramService.AnswerCallbackQuery(callback.ID, "Proposal text not found")
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
	docPath, filename, err := botService.DocumentService.GenerateProposalDocument(user, student, stateData.ProposalText)
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

	// Save proposal to database with document info
	proposalReq := &models.CreateProposalRequest{
		UserID:         user.ID,
		ProposalText:   stateData.ProposalText,
		TelegramFileID: fileID,
		Filename:       filename,
	}

	proposal, err := botService.ProposalService.CreateProposal(proposalReq)
	if err != nil {
		log.Printf("Failed to save proposal: %v", err)
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Send success message
	text := i18n.Get(i18n.MsgProposalSubmitted, lang)
	keyboard := utils.MakeMainMenuKeyboard(lang)
	_ = botService.TelegramService.SendMessage(chatID, text, keyboard)

	// Notify admins with DOCX document
	go notifyAdminsWithProposalDocument(botService, user, proposal, fileID)

	return nil
}

// HandleProposalCancellation handles proposal cancellation
func HandleProposalCancellation(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
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
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, i18n.Get(i18n.MsgProposalCancelled, lang))

	// Send cancellation message
	text := i18n.Get(i18n.MsgProposalCancelled, lang)
	keyboard := utils.MakeMainMenuKeyboard(lang)

	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// notifyAdminsWithProposalDocument sends proposal as DOCX document to all admins
func notifyAdminsWithProposalDocument(botService *services.BotService, user *models.User, proposal *models.Proposal, fileID string) {
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
		"<b>YANGI TAKLIF / НОВОЕ ПРЕДЛОЖЕНИЕ</b>\n\n"+
			"ID: #%d\n"+
			"Telefon / Телефон: %s\n"+
			"Username: @%s\n"+
			"Sana / Дата: %s\n\n"+
			"Taklif hujjat sifatida yuqorida\n"+
			"Предложение в формате документа выше",
		proposal.ID,
		user.PhoneNumber,
		username,
		utils.FormatDateTime(proposal.CreatedAt),
	)

	// Send document to all admins
	err = botService.TelegramService.SendDocumentToAdmins(adminIDs, fileID, caption)
	if err != nil {
		log.Printf("Failed to send document to admins: %v", err)
	}
}

// HandleMyProposalsCommand shows user's proposal history
func HandleMyProposalsCommand(botService *services.BotService, message *tgbotapi.Message) error {
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

	// Get user proposals
	proposals, err := botService.ProposalService.GetUserProposals(user.ID, 10, 0)
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(proposals) == 0 {
		text := "Sizda hali takliflar yo'q / У вас пока нет предложений"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Format proposals list
	text := "💡 Sizning takliflaringiz / Ваши предложения:\n\n"
	for i, p := range proposals {
		status := "⏳"
		if p.Status == models.StatusReviewed {
			status = "✅"
		}

		preview := utils.TruncateText(p.ProposalText, 50)
		text += fmt.Sprintf("%d. %s %s\n   📅 %s\n\n",
			i+1,
			status,
			preview,
			utils.FormatDateTime(p.CreatedAt),
		)
	}

	return botService.TelegramService.SendMessage(chatID, text, nil)
}
