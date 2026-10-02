package handlers

import (
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
	"parent-bot/internal/utils"
)

// HandleViewAnnouncementsCommand shows all active announcements for parents
func HandleViewAnnouncementsCommand(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Determine language and check admin status
	language := string(i18n.LanguageUzbek)
	phoneNumber := ""
	if user != nil {
		language = user.Language
		phoneNumber = user.PhoneNumber
	}
	lang := i18n.GetLanguage(language)

	// Check if user is admin (works even if user is nil)
	isAdmin, _ := botService.IsAdmin(phoneNumber, telegramID)

	// If not admin and not registered, return error
	if user == nil && !isAdmin {
		text := i18n.Get(i18n.ErrNotRegistered, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Admins see their branch; parents see branch-wide and their children's class announcements
	var announcements []*models.Announcement
	if isAdmin {
		announcements, err = botService.AnnouncementService.GetActiveAnnouncements(adminBranchID(botService, telegramID), 10, 0)
	} else {
		announcements, err = botService.AnnouncementService.GetActiveAnnouncementsForParent(user.ID, user.BranchID, 10, 0)
	}
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(announcements) == 0 {
		text := i18n.Get(i18n.MsgNoAnnouncements, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Send each announcement
	for i, announcement := range announcements {
		// Format announcement text
		text := fmt.Sprintf("📢 E'lon / Объявление #%d\n\n", i+1)

		if announcement.Title != nil && *announcement.Title != "" {
			text += fmt.Sprintf("<b>%s</b>\n\n", utils.EscapeHTML(*announcement.Title))
		}

		text += utils.EscapeHTML(announcement.Content)
		text += fmt.Sprintf("\n\n📅 %s", utils.FormatDateTime(announcement.CreatedAt))

		// Create inline keyboard for admin with edit and delete buttons
		var inlineKeyboard *tgbotapi.InlineKeyboardMarkup
		if isAdmin {
			inlineKeyboard = &tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
					{
						tgbotapi.NewInlineKeyboardButtonData(
							i18n.Get(i18n.BtnEdit, lang),
							fmt.Sprintf("announcement_edit_%d", announcement.ID),
						),
						tgbotapi.NewInlineKeyboardButtonData(
							i18n.Get(i18n.BtnDelete, lang),
							fmt.Sprintf("announcement_delete_%d", announcement.ID),
						),
					},
				},
			}
		}

		// Send announcement with image if available
		if announcement.TelegramFileID != nil && *announcement.TelegramFileID != "" {
			fileID := *announcement.TelegramFileID
			log.Printf("Sending announcement #%d with media (FileID: %s, Type: %v)", announcement.ID, fileID, announcement.FileType)

			// Check if it's a document or photo based on FileID prefix
			// Document FileIDs start with "BQAC", Photo FileIDs start with "AgAC"
			isDocument := len(fileID) > 4 && fileID[:4] == "BQAC"

			var sendErr error
			if isDocument {
				// Send as document
				log.Printf("Detected document type, sending as document")
				doc := tgbotapi.NewDocument(chatID, tgbotapi.FileID(fileID))
				doc.Caption = text
				doc.ParseMode = "HTML"
				if inlineKeyboard != nil {
					doc.ReplyMarkup = *inlineKeyboard
				}
				_, sendErr = botService.Bot.Send(doc)
			} else {
				// Send as photo
				log.Printf("Detected photo type, sending as photo")
				photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(fileID))
				photo.Caption = text
				photo.ParseMode = "HTML"
				if inlineKeyboard != nil {
					photo.ReplyMarkup = *inlineKeyboard
				}
				_, sendErr = botService.Bot.Send(photo)
			}

			if sendErr != nil {
				log.Printf("ERROR: Failed to send media for announcement #%d: %v (FileID: %s)", announcement.ID, sendErr, fileID)
				// Fallback to text only
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ParseMode = "HTML"
				if inlineKeyboard != nil {
					msg.ReplyMarkup = *inlineKeyboard
				}
				_, textErr := botService.Bot.Send(msg)
				if textErr != nil {
					log.Printf("ERROR: Failed to send text fallback: %v", textErr)
				}
			} else {
				log.Printf("Successfully sent announcement #%d with media", announcement.ID)
			}
		} else {
			log.Printf("Sending announcement #%d without image (text only)", announcement.ID)
			// Send text only
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = "HTML"
			if inlineKeyboard != nil {
				msg.ReplyMarkup = *inlineKeyboard
			}
			_, sendErr := botService.Bot.Send(msg)
			if sendErr != nil {
				log.Printf("ERROR: Failed to send text message: %v", sendErr)
			}
		}
	}

	// Send a final message with the main menu keyboard to ensure it stays visible
	mainMenuKeyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)
	finalMsg := tgbotapi.NewMessage(chatID, "👆 E'lonlar yuqorida / Объявления выше")
	finalMsg.ReplyMarkup = mainMenuKeyboard
	_, _ = botService.Bot.Send(finalMsg)

	return nil
}

// HandlePostAnnouncementCommand initiates announcement posting (admin only)
func HandlePostAnnouncementCommand(botService *services.BotService, message *tgbotapi.Message) error {
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
	language := string(i18n.LanguageUzbek)
	if user != nil {
		language = user.Language
	}
	lang := i18n.GetLanguage(language)

	// Set state to awaiting announcement content
	stateData := &models.StateData{
		Language: language,
	}
	err = botService.StateManager.Set(telegramID, models.StateAwaitingAnnouncementContent, stateData)
	if err != nil {
		return err
	}

	// Send request for content
	text := i18n.Get(i18n.MsgRequestAnnouncementContent, lang)
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleAnnouncementContent handles announcement content input
func HandleAnnouncementContent(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID
	lang := i18n.GetLanguage(stateData.Language)

	// Get user to check admin status for keyboard
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	var isAdmin bool
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
		isAdmin, _ = botService.IsAdmin(phoneNumber, telegramID)
	} else {
		isAdmin, _ = botService.IsAdmin("", telegramID)
	}

	// Check if message contains media instead of text
	if message.Text == "" {
		var errorMsg string
		if len(message.Photo) > 0 {
			errorMsg = "❌ Avval matn yuboring, keyin rasm yuborishingiz mumkin!\n\n❌ Сначала отправьте текст, затем вы сможете отправить изображение!"
		} else if message.Animation != nil {
			errorMsg = "❌ Iltimos, GIF emas, e'lon matnini yuboring!\n\n❌ Пожалуйста, отправьте текст объявления, а не GIF!"
		} else if message.Video != nil {
			errorMsg = "❌ Iltimos, video emas, e'lon matnini yuboring!\n\n❌ Пожалуйста, отправьте текст объявления, а не видео!"
		} else if message.Document != nil {
			errorMsg = "❌ Iltimos, fayl emas, e'lon matnini yuboring!\n\n❌ Пожалуйста, отправьте текст объявления, а не файл!"
		} else if message.Sticker != nil {
			errorMsg = "❌ Iltimos, stiker emas, e'lon matnini yuboring!\n\n❌ Пожалуйста, отправьте текст объявления, а не стикер!"
		} else if message.Voice != nil {
			errorMsg = "❌ Iltimos, ovozli xabar emas, e'lon matnini yuboring!\n\n❌ Пожалуйста, отправьте текст объявления, а не голосовое сообщение!"
		} else {
			errorMsg = "❌ Iltimos, e'lon matnini yuboring!\n\n❌ Пожалуйста, отправьте текст объявления!"
		}

		// Keep the main menu keyboard visible
		keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)
		return botService.TelegramService.SendMessage(chatID, errorMsg, &keyboard)
	}

	// Validate content (at least 10 characters)
	if len(message.Text) < 10 {
		text := "❌ E'lon matni juda qisqa! Kamida 10 ta belgi kiriting.\n\n❌ Текст объявления слишком короткий! Введите минимум 10 символов."
		// Keep the main menu keyboard visible on validation errors too
		keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)
		return botService.TelegramService.SendMessage(chatID, text, &keyboard)
	}

	// Save content in state
	stateData.AnnouncementText = message.Text

	// Move to file upload state
	err = botService.StateManager.Set(telegramID, models.StateAwaitingAnnouncementFile, stateData)
	if err != nil {
		return err
	}

	// Ask for optional image
	text := i18n.Get(i18n.MsgRequestAnnouncementFile, lang)

	// Create keyboard with skip button
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnSkip, lang),
				"announcement_skip_file",
			),
		),
	)

	return botService.TelegramService.SendMessage(chatID, text, &keyboard)
}

// HandleAnnouncementFile handles announcement file upload
func HandleAnnouncementFile(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID
	lang := i18n.GetLanguage(stateData.Language)

	// Get user to check admin status for keyboard
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	var isAdmin bool
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
		isAdmin, _ = botService.IsAdmin(phoneNumber, telegramID)
	} else {
		isAdmin, _ = botService.IsAdmin("", telegramID)
	}

	var fileID, filename *string
	fileType := "image"

	// Check if photo was sent (compressed)
	if len(message.Photo) > 0 {
		photo := message.Photo[len(message.Photo)-1] // Get largest photo
		fileID = &photo.FileID
		fname := fmt.Sprintf("announcement_%d.jpg", telegramID)
		filename = &fname
	} else if message.Document != nil {
		// Check if document is an image (including HEIC for iPhone)
		mimeType := message.Document.MimeType
		if mimeType == "image/jpeg" || mimeType == "image/jpg" || mimeType == "image/png" ||
			mimeType == "image/gif" || mimeType == "image/heic" || mimeType == "image/heif" {
			fileID = &message.Document.FileID
			fname := message.Document.FileName
			if fname == "" {
				fname = fmt.Sprintf("announcement_%d.jpg", telegramID)
			}
			filename = &fname
		} else {
			text := i18n.Get(i18n.ErrInvalidFile, lang) + "\n\nIltimos, rasm formatini yuboring (JPG, PNG, GIF, HEIC). / Пожалуйста, отправьте изображение в формате JPG, PNG, GIF или HEIC."
			// Keep the main menu keyboard visible on errors
			keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)
			return botService.TelegramService.SendMessage(chatID, text, &keyboard)
		}
	} else if message.Text != "" {
		// User sent text instead of image - show a helpful error
		text := "❌ Iltimos, rasm yuboring yoki 'O'tkazib yuborish' tugmasini bosing.\n\n❌ Пожалуйста, отправьте изображение или нажмите кнопку 'Пропустить'."
		// Keep the main menu keyboard visible
		keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)
		return botService.TelegramService.SendMessage(chatID, text, &keyboard)
	} else {
		text := i18n.Get(i18n.ErrInvalidFile, lang) + "\n\nIltimos, rasm yuboring. / Пожалуйста, отправьте изображение."
		// Keep the main menu keyboard visible on errors
		keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)
		return botService.TelegramService.SendMessage(chatID, text, &keyboard)
	}

	// Save announcement with file
	return saveAnnouncement(botService, telegramID, chatID, stateData, fileID, filename, &fileType)
}

// HandleAnnouncementSkipFile handles skipping file upload
func HandleAnnouncementSkipFile(botService *services.BotService, callback *tgbotapi.CallbackQuery, stateData *models.StateData) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Answer callback query
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅")

	// Save announcement without file
	return saveAnnouncement(botService, telegramID, chatID, stateData, nil, nil, nil)
}

// saveAnnouncement saves the announcement to database
func saveAnnouncement(botService *services.BotService, telegramID int64, chatID int64, stateData *models.StateData, fileID, filename, fileType *string) error {
	lang := i18n.GetLanguage(stateData.Language)

	// Get admin record
	admin := botService.GetAdmin(telegramID)

	if admin == nil {
		_ = botService.StateManager.Clear(telegramID)
		return botService.TelegramService.SendMessage(chatID, msgNotAllowed, nil)
	}
	adminID := &admin.ID

	// Create announcement record
	announcementReq := &models.CreateAnnouncementRequest{
		Title:           nil, // We're not asking for title in the current flow
		Content:         stateData.AnnouncementText,
		TelegramFileID:  fileID,
		Filename:        filename,
		FileType:        fileType,
		PostedByAdminID: adminID,
		BranchID:        admin.BranchID,
	}

	// Log file ID for debugging
	if fileID != nil {
		log.Printf("Creating announcement with FileID: %s", *fileID)
	} else {
		log.Printf("Creating announcement without image (FileID is nil)")
	}

	announcement, err := botService.AnnouncementService.CreateAnnouncement(announcementReq)
	if err != nil {
		log.Printf("Failed to save announcement: %v", err)
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Verify the announcement was created with file ID
	if announcement.TelegramFileID != nil {
		log.Printf("Announcement #%d created successfully with FileID: %s", announcement.ID, *announcement.TelegramFileID)
	} else {
		log.Printf("Announcement #%d created successfully without image", announcement.ID)
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Send success message
	text := i18n.Get(i18n.MsgAnnouncementPosted, lang)
	_ = botService.TelegramService.SendMessage(chatID, text, nil)

	// Notify all users about new announcement
	go notifyUsersAboutAnnouncement(botService, announcement)

	return nil
}

// notifyUsersAboutAnnouncement sends a new announcement to the parents it is meant for:
// parents of the targeted classes, or every parent of the announcement's branch.
func notifyUsersAboutAnnouncement(botService *services.BotService, announcement *models.Announcement) {
	var users []*models.User
	classIDs, err := botService.AnnouncementService.GetAnnouncementClassIDs(announcement.ID)
	if err != nil {
		log.Printf("Failed to get announcement classes: %v", err)
		return
	}
	if len(classIDs) > 0 {
		users, err = botService.UserService.GetParentsByClassIDs(classIDs)
	} else {
		const pageSize = 500
		for offset := 0; ; offset += pageSize {
			page, pageErr := botService.UserService.GetAllUsers(announcement.BranchID, pageSize, offset)
			if pageErr != nil {
				err = pageErr
				break
			}
			users = append(users, page...)
			if len(page) < pageSize {
				break
			}
		}
	}
	if err != nil {
		log.Printf("Failed to get announcement recipients: %v", err)
		return
	}

	text := "📢 YANGI E'LON / НОВОЕ ОБЪЯВЛЕНИЕ\n\n"
	if announcement.Title != nil && *announcement.Title != "" {
		text += fmt.Sprintf("<b>%s</b>\n\n", utils.EscapeHTML(*announcement.Title))
	}
	text += utils.EscapeHTML(announcement.Content)
	text += fmt.Sprintf("\n\n📅 %s", utils.FormatDateTime(announcement.CreatedAt))

	successCount, failCount := 0, 0
	for _, user := range users {
		if sendAnnouncement(botService, user.TelegramID, announcement, text) {
			successCount++
		} else {
			failCount++
		}
		// Stay well below Telegram's broadcast limit (~30 messages/second)
		time.Sleep(50 * time.Millisecond)
	}

	log.Printf("Announcement #%d delivered: %d ok, %d failed of %d parents", announcement.ID, successCount, failCount, len(users))
}

// sendAnnouncement sends announcement text (already HTML-escaped) with its attachment.
// A caption is limited to 1024 characters, so long texts are sent as a separate message.
func sendAnnouncement(botService *services.BotService, chatID int64, announcement *models.Announcement, text string) bool {
	if announcement.TelegramFileID != nil && *announcement.TelegramFileID != "" {
		fileID := *announcement.TelegramFileID
		caption := text
		if len([]rune(caption)) > 1000 {
			caption = ""
		}

		var chattable tgbotapi.Chattable
		// Document file IDs start with "BQAC", photo file IDs with "AgAC"
		if strings.HasPrefix(fileID, "BQAC") {
			doc := tgbotapi.NewDocument(chatID, tgbotapi.FileID(fileID))
			doc.Caption, doc.ParseMode = caption, "HTML"
			chattable = doc
		} else {
			photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(fileID))
			photo.Caption, photo.ParseMode = caption, "HTML"
			chattable = photo
		}
		if _, err := botService.Bot.Send(chattable); err != nil {
			log.Printf("Failed to send announcement media to %d: %v", chatID, err)
		} else if caption != "" {
			return true
		}
	}

	if err := botService.TelegramService.SendMessage(chatID, text, nil); err != nil {
		log.Printf("Failed to send announcement to %d: %v", chatID, err)
		return false
	}
	return true
}

// HandleAnnouncementDeleteCallback handles announcement deletion request
func HandleAnnouncementDeleteCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, announcementID int) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Determine language
	language := string(i18n.LanguageUzbek)
	if user != nil {
		language = user.Language
	}
	lang := i18n.GetLanguage(language)

	// Check if user is admin
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil || !isAdmin {
		text := "❌ Bu buyruq faqat ma'murlar uchun / Эта команда только для администраторов"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	// Delete the announcement
	err = botService.AnnouncementService.DeleteAnnouncement(announcementID)
	if err != nil {
		log.Printf("Failed to delete announcement: %v", err)
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Answer callback query with success
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅ O'chirildi / Удалено")

	// Send confirmation message
	text := "✅ E'lon muvaffaqiyatli o'chirildi! / Объявление успешно удалено!"
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleAnnouncementEditCallback handles announcement edit request
func HandleAnnouncementEditCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, announcementID int) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Determine language
	language := string(i18n.LanguageUzbek)
	if user != nil {
		language = user.Language
	}
	lang := i18n.GetLanguage(language)

	// Check if user is admin
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
	}

	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil || !isAdmin {
		text := "❌ Bu buyruq faqat ma'murlar uchun / Эта команда только для администраторов"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	// Get the announcement
	announcement, err := botService.AnnouncementService.GetAnnouncementByID(announcementID)
	if err != nil {
		log.Printf("Failed to get announcement: %v", err)
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if announcement == nil {
		text := "❌ E'lon topilmadi / Объявление не найдено"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	// Answer callback query
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Set state to awaiting edited announcement content
	stateData := &models.StateData{
		Language:       language,
		AnnouncementID: announcementID,
	}
	err = botService.StateManager.Set(telegramID, models.StateAwaitingEditedAnnouncementContent, stateData)
	if err != nil {
		return err
	}

	// Show current announcement and ask for new content
	text := "✏️ E'lonni tahrirlash / Редактировать объявление\n\n"
	text += "📄 Joriy matn / Текущий текст:\n\n"
	text += utils.EscapeHTML(announcement.Content)
	text += "\n\n━━━━━━━━━━━━━━━\n\n"
	text += "📝 Yangi matnni kiriting / Введите новый текст:"

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleEditedAnnouncementContent handles the edited announcement content
func HandleEditedAnnouncementContent(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID
	lang := i18n.GetLanguage(stateData.Language)

	// Get user to check admin status for keyboard
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	var isAdmin bool
	phoneNumber := ""
	if user != nil {
		phoneNumber = user.PhoneNumber
		isAdmin, _ = botService.IsAdmin(phoneNumber, telegramID)
	} else {
		isAdmin, _ = botService.IsAdmin("", telegramID)
	}

	// Check if message contains media instead of text
	if message.Text == "" {
		var errorMsg string
		if len(message.Photo) > 0 {
			errorMsg = "❌ Iltimos, rasm emas, yangi matnni yuboring!\n\n❌ Пожалуйста, отправьте новый текст, а не изображение!"
		} else if message.Animation != nil {
			errorMsg = "❌ Iltimos, GIF emas, yangi matnni yuboring!\n\n❌ Пожалуйста, отправьте новый текст, а не GIF!"
		} else if message.Video != nil {
			errorMsg = "❌ Iltimos, video emas, yangi matnni yuboring!\n\n❌ Пожалуйста, отправьте новый текст, а не видео!"
		} else if message.Document != nil {
			errorMsg = "❌ Iltimos, fayl emas, yangi matnni yuboring!\n\n❌ Пожалуйста, отправьте новый текст, а не файл!"
		} else if message.Sticker != nil {
			errorMsg = "❌ Iltimos, stiker emas, yangi matnni yuboring!\n\n❌ Пожалуйста, отправьте новый текст, а не стикер!"
		} else if message.Voice != nil {
			errorMsg = "❌ Iltimos, ovozli xabar emas, yangi matnni yuboring!\n\n❌ Пожалуйста, отправьте новый текст, а не голосовое сообщение!"
		} else {
			errorMsg = "❌ Iltimos, yangi matnni yuboring!\n\n❌ Пожалуйста, отправьте новый текст!"
		}

		// Keep the main menu keyboard visible
		keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)
		return botService.TelegramService.SendMessage(chatID, errorMsg, &keyboard)
	}

	// Validate content (at least 10 characters)
	if len(message.Text) < 10 {
		text := "❌ E'lon matni juda qisqa! Kamida 10 ta belgi kiriting.\n\n❌ Текст объявления слишком короткий! Введите минимум 10 символов."
		// Keep the main menu keyboard visible on validation errors too
		keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)
		return botService.TelegramService.SendMessage(chatID, text, &keyboard)
	}

	// Get the announcement to check if it exists
	announcement, err := botService.AnnouncementService.GetAnnouncementByID(stateData.AnnouncementID)
	if err != nil || announcement == nil {
		log.Printf("Failed to get announcement: %v", err)
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)
		return botService.TelegramService.SendMessage(chatID, text, &keyboard)
	}

	// Update announcement
	updateReq := &models.CreateAnnouncementRequest{
		Title:           announcement.Title,
		Content:         message.Text,
		TelegramFileID:  announcement.TelegramFileID,
		Filename:        announcement.Filename,
		FileType:        announcement.FileType,
		PostedByAdminID: announcement.PostedByAdminID,
	}

	_, err = botService.AnnouncementService.UpdateAnnouncement(stateData.AnnouncementID, updateReq)
	if err != nil {
		log.Printf("Failed to update announcement: %v", err)
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)
		return botService.TelegramService.SendMessage(chatID, text, &keyboard)
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Send success message with keyboard
	text := "✅ E'lon muvaffaqiyatli tahrirlandi! / Объявление успешно отредактировано!"
	keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)
	return botService.TelegramService.SendMessage(chatID, text, &keyboard)
}

// HandleAdminViewAnnouncementsCallback shows all announcements to admin with edit/delete buttons
func HandleAdminViewAnnouncementsCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Determine language
	language := string(i18n.LanguageUzbek)
	phoneNumber := ""
	if user != nil {
		language = user.Language
		phoneNumber = user.PhoneNumber
	}
	lang := i18n.GetLanguage(language)

	// Check if user is admin
	isAdmin, err := botService.IsAdmin(phoneNumber, telegramID)
	if err != nil || !isAdmin {
		text := "❌ Bu buyruq faqat ma'murlar uchun / Эта команда только для администраторов"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	// Answer callback query
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Get all announcements (not just active)
	announcements, err := botService.AnnouncementService.GetAllAnnouncements(adminBranchID(botService, callback.From.ID), 20, 0)
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(announcements) == 0 {
		text := i18n.Get(i18n.MsgNoAnnouncements, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Send each announcement with edit/delete buttons
	for i, announcement := range announcements {
		// Format announcement text
		statusEmoji := "✅"
		if !announcement.IsActive {
			statusEmoji = "❌"
		}

		text := fmt.Sprintf("%s E'lon / Объявление #%d (ID: %d)\n\n", statusEmoji, i+1, announcement.ID)

		if announcement.Title != nil && *announcement.Title != "" {
			text += fmt.Sprintf("<b>%s</b>\n\n", utils.EscapeHTML(*announcement.Title))
		}

		text += utils.EscapeHTML(announcement.Content)
		text += fmt.Sprintf("\n\n📅 %s", utils.FormatDateTime(announcement.CreatedAt))

		if !announcement.IsActive {
			text += "\n\n⚠️ Nofaol / Неактивно"
		}

		// Create keyboard with edit and delete buttons
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					i18n.Get(i18n.BtnEdit, lang),
					fmt.Sprintf("announcement_edit_%d", announcement.ID),
				),
				tgbotapi.NewInlineKeyboardButtonData(
					i18n.Get(i18n.BtnDelete, lang),
					fmt.Sprintf("announcement_delete_%d", announcement.ID),
				),
			),
		)

		// Send announcement with image if available
		if announcement.TelegramFileID != nil && *announcement.TelegramFileID != "" {
			fileID := *announcement.TelegramFileID
			log.Printf("Admin view: Sending announcement #%d with media (FileID: %s, Type: %v)", announcement.ID, fileID, announcement.FileType)

			// Check if it's a document or photo based on FileID prefix
			isDocument := len(fileID) > 4 && fileID[:4] == "BQAC"

			var sendErr error
			if isDocument {
				// Send as document
				log.Printf("Admin view: Detected document type, sending as document")
				doc := tgbotapi.NewDocument(chatID, tgbotapi.FileID(fileID))
				doc.Caption = text
				doc.ParseMode = "HTML"
				doc.ReplyMarkup = keyboard
				_, sendErr = botService.Bot.Send(doc)
			} else {
				// Send as photo
				log.Printf("Admin view: Detected photo type, sending as photo")
				photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(fileID))
				photo.Caption = text
				photo.ParseMode = "HTML"
				photo.ReplyMarkup = keyboard
				_, sendErr = botService.Bot.Send(photo)
			}

			if sendErr != nil {
				log.Printf("ERROR: Admin view failed to send media for announcement #%d: %v (FileID: %s)", announcement.ID, sendErr, fileID)
				// Fallback to text only
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ParseMode = "HTML"
				msg.ReplyMarkup = keyboard
				_, textErr := botService.Bot.Send(msg)
				if textErr != nil {
					log.Printf("ERROR: Failed to send text fallback: %v", textErr)
				}
			} else {
				log.Printf("Successfully sent admin announcement #%d with media", announcement.ID)
			}
		} else {
			log.Printf("Admin view: Sending announcement #%d without image (text only)", announcement.ID)
			// Send text only
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = "HTML"
			msg.ReplyMarkup = keyboard
			_, sendErr := botService.Bot.Send(msg)
			if sendErr != nil {
				log.Printf("ERROR: Failed to send text message: %v", sendErr)
			}
		}
	}

	return nil
}
