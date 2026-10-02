package handlers

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
)

// HandleTeacherAnnouncementToggleClass handles toggling class selection for announcement
func HandleTeacherAnnouncementToggleClass(botService *services.BotService, callback *tgbotapi.CallbackQuery, classID int) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get teacher
	teacher, err := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if err != nil || teacher == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik")
		return nil
	}

	// Get state data
	stateData, err := botService.StateManager.GetData(telegramID)
	if err != nil || stateData == nil {
		stateData = &models.StateData{
			SelectedClasses: []int{},
		}
	}

	// Toggle class in selected list
	found := false
	newSelectedClasses := []int{}
	for _, id := range stateData.SelectedClasses {
		if id == classID {
			found = true
			// Remove from list (toggle off)
		} else {
			newSelectedClasses = append(newSelectedClasses, id)
		}
	}

	if !found {
		// Add to list (toggle on)
		newSelectedClasses = append(stateData.SelectedClasses, classID)
	}

	stateData.SelectedClasses = newSelectedClasses

	// Update state
	err = botService.StateManager.Set(telegramID, "teacher_selecting_announcement_classes", stateData)
	if err != nil {
		log.Printf("Failed to update state: %v", err)
	}

	// Re-render the class selection screen with updated checkboxes
	// Teachers can see all classes
	classes, err := botService.ClassRepo.GetAll(staffBranchID(botService, callback.From.ID))
	if err != nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik")
		return nil
	}

	// Create selected map for quick lookup
	selectedMap := make(map[int]bool)
	for _, id := range stateData.SelectedClasses {
		selectedMap[id] = true
	}

	// Create inline keyboard
	var buttons [][]tgbotapi.InlineKeyboardButton

	text := "📢 <b>E'lon qo'shish / Добавить объявление</b>\n\n" +
		"Qaysi sinf(lar) uchun e'lon qo'shmoqchisiz?\n" +
		"Для какого класса(ов) хотите добавить объявление?\n\n" +
		fmt.Sprintf("Tanlangan: %d / Выбрано: %d", len(stateData.SelectedClasses), len(stateData.SelectedClasses))

	for _, class := range classes {
		checkbox := "☐"
		if selectedMap[class.ID] {
			checkbox = "☑"
		}

		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s %s", checkbox, class.ClassName),
			fmt.Sprintf("teacher_announcement_toggle_class_%d", class.ID),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	// Add "Continue" button
	continueButton := tgbotapi.NewInlineKeyboardButtonData(
		"✅ Davom etish / Продолжить",
		"teacher_announcement_continue",
	)
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{continueButton})

	// Add "Cancel" button
	cancelButton := tgbotapi.NewInlineKeyboardButtonData(
		"❌ Bekor qilish / Отмена",
		"teacher_announcement_cancel",
	)
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{cancelButton})

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	// Update message
	editMsg := tgbotapi.NewEditMessageText(chatID, callback.Message.MessageID, text)
	editMsg.ParseMode = "HTML"
	editMsg.ReplyMarkup = &keyboard

	_, err = botService.Bot.Send(editMsg)
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
	return err
}

// HandleTeacherAnnouncementContinue handles continuing to announcement content input
func HandleTeacherAnnouncementContinue(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get state data
	stateData, err := botService.StateManager.GetData(telegramID)
	if err != nil || stateData == nil || len(stateData.SelectedClasses) == 0 {
		text := "❌ Iltimos, kamida bitta sinf tanlang.\n\n❌ Пожалуйста, выберите хотя бы один класс."
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	// Delete the selection message
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
	_, _ = botService.Bot.Request(deleteMsg)

	// Ask for announcement content
	text := "📢 <b>E'lon matni / Текст объявления</b>\n\n" +
		"E'lon matnini yuboring (kamida 10 ta belgi):\n" +
		"Отправьте текст объявления (минимум 10 символов):\n\n" +
		"💡 Rasm qo'shish uchun keyingi qadamda imkoniyat bo'ladi.\n" +
		"💡 Возможность добавить изображение будет на следующем шаге."

	// Update state
	err = botService.StateManager.Set(telegramID, "teacher_awaiting_announcement_content", stateData)
	if err != nil {
		log.Printf("Failed to update state: %v", err)
	}

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleTeacherAnnouncementCancel handles canceling announcement creation
func HandleTeacherAnnouncementCancel(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Delete the selection message
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
	_, _ = botService.Bot.Request(deleteMsg)

	text := "❌ E'lon qo'shish bekor qilindi.\n\n❌ Создание объявления отменено."
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleTeacherAnnouncementContent handles announcement content input
func HandleTeacherAnnouncementContent(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	content := message.Text

	// Validate content
	if len(content) < 10 {
		text := "❌ E'lon matni juda qisqa (kamida 10 ta belgi).\n\n" +
			"❌ Текст объявления слишком короткий (минимум 10 символов)."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(content) > 4000 {
		text := "❌ E'lon matni juda uzun (maksimal 4000 ta belgi).\n\n" +
			"❌ Текст объявления слишком длинный (максимум 4000 символов)."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get state data
	stateData, err := botService.StateManager.GetData(telegramID)
	if err != nil || stateData == nil {
		text := "❌ Sessiya tugagan. Iltimos, qaytadan boshlang.\n\n" +
			"❌ Сессия истекла. Пожалуйста, начните заново."
		_ = botService.StateManager.Clear(telegramID)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Save content to state
	stateData.AnnouncementText = content

	// Move to file upload state (ask for optional image)
	err = botService.StateManager.Set(telegramID, "teacher_awaiting_announcement_file", stateData)
	if err != nil {
		log.Printf("Failed to update state: %v", err)
	}

	// Ask for optional image
	text := "📸 <b>Rasm qo'shish (ixtiyoriy) / Добавить изображение (необязательно)</b>\n\n" +
		"Rasm yubormoqchimisiz? Yuborishingiz mumkin (JPG, PNG, GIF, HEIC).\n" +
		"Хотите отправить изображение? Можете отправить (JPG, PNG, GIF, HEIC).\n\n" +
		"💡 Yoki 'O'tkazib yuborish' tugmasini bosing.\n" +
		"💡 Или нажмите кнопку 'Пропустить'."

	// Create keyboard with skip button
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"⏩ O'tkazib yuborish / Пропустить",
				"teacher_announcement_skip_file",
			),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard
	_, err = botService.Bot.Send(msg)
	return err
}

// HandleTeacherAnnouncementEdit handles editing an announcement
func HandleTeacherAnnouncementEdit(botService *services.BotService, callback *tgbotapi.CallbackQuery, announcementID int) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get teacher
	teacher, err := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if err != nil || teacher == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik")
		return nil
	}

	// Get announcement
	announcement, err := botService.AnnouncementRepo.GetByID(announcementID)
	if err != nil || announcement == nil {
		text := "❌ E'lon topilmadi / Объявление не найдено"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Verify teacher owns this announcement
	if announcement.PostedByTeacherID == nil || *announcement.PostedByTeacherID != teacher.ID {
		text := "❌ Siz bu e'lonni tahrirlay olmaysiz / Вы не можете редактировать это объявление"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	// Ask for new content
	text := fmt.Sprintf(
		"✏️ <b>E'lonni tahrirlash / Редактирование объявления</b>\n\n"+
			"Joriy matn / Текущий текст:\n"+
			"<code>%s</code>\n\n"+
			"Yangi matnni yuboring:\n"+
			"Отправьте новый текст:",
		announcement.Content,
	)

	// Set state
	stateData := &models.StateData{
		AnnouncementID: announcementID,
	}
	err = botService.StateManager.Set(telegramID, "teacher_editing_announcement_content", stateData)
	if err != nil {
		log.Printf("Failed to set state: %v", err)
	}

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleTeacherEditedAnnouncementContent handles the edited announcement content
func HandleTeacherEditedAnnouncementContent(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	content := message.Text

	// Validate content
	if len(content) < 10 {
		text := "❌ E'lon matni juda qisqa (kamida 10 ta belgi).\n\n" +
			"❌ Текст объявления слишком короткий (минимум 10 символов)."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get state data
	stateData, err := botService.StateManager.GetData(telegramID)
	if err != nil || stateData == nil {
		text := "❌ Sessiya tugagan. Iltimos, qaytadan boshlang.\n\n" +
			"❌ Сессия истекла. Пожалуйста, начните заново."
		_ = botService.StateManager.Clear(telegramID)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get existing announcement to preserve fields
	existingAnnouncement, err := botService.AnnouncementRepo.GetByID(stateData.AnnouncementID)
	if err != nil || existingAnnouncement == nil {
		log.Printf("Failed to get announcement: %v", err)
		text := "❌ E'lon topilmadi / Объявление не найдено"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Update announcement with new content
	req := &models.CreateAnnouncementRequest{
		Title:             existingAnnouncement.Title,
		Content:           content,
		TelegramFileID:    existingAnnouncement.TelegramFileID,
		Filename:          existingAnnouncement.Filename,
		FileType:          existingAnnouncement.FileType,
		PostedByTeacherID: existingAnnouncement.PostedByTeacherID,
	}

	_, err = botService.AnnouncementService.UpdateAnnouncement(stateData.AnnouncementID, req)
	if err != nil {
		log.Printf("Failed to update announcement: %v", err)
		text := "❌ E'lonni yangilashda xatolik / Ошибка при обновлении объявления"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Success message
	text := fmt.Sprintf(
		"✅ E'lon muvaffaqiyatli yangilandi!\n\n"+
			"ID: <code>%d</code>\n\n"+
			"✅ Объявление успешно обновлено!\n\n"+
			"ID: <code>%d</code>",
		stateData.AnnouncementID,
		stateData.AnnouncementID,
	)

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleTeacherAnnouncementDelete handles deleting an announcement
func HandleTeacherAnnouncementDelete(botService *services.BotService, callback *tgbotapi.CallbackQuery, announcementID int) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get teacher
	teacher, err := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if err != nil || teacher == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik")
		return nil
	}

	// Get announcement
	announcement, err := botService.AnnouncementRepo.GetByID(announcementID)
	if err != nil || announcement == nil {
		text := "❌ E'lon topilmadi / Объявление не найдено"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Verify teacher owns this announcement
	if announcement.PostedByTeacherID == nil || *announcement.PostedByTeacherID != teacher.ID {
		text := "❌ Siz bu e'lonni o'chira olmaysiz / Вы не можете удалить это объявление"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, text)
		return nil
	}

	// Delete announcement
	err = botService.AnnouncementRepo.Delete(announcementID)
	if err != nil {
		log.Printf("Failed to delete announcement: %v", err)
		text := "❌ E'lonni o'chirishda xatolik / Ошибка при удалении объявления"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Delete the message
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
	_, _ = botService.Bot.Request(deleteMsg)

	text := fmt.Sprintf(
		"✅ E'lon o'chirildi!\n\n"+
			"ID: <code>%d</code>\n\n"+
			"✅ Объявление удалено!\n\n"+
			"ID: <code>%d</code>",
		announcementID,
		announcementID,
	)

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅ O'chirildi!")
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleTeacherAnnouncementFile handles announcement file upload for teachers
func HandleTeacherAnnouncementFile(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	var fileID, filename *string
	fileType := "image"

	// Check if photo was sent (compressed)
	if len(message.Photo) > 0 {
		photo := message.Photo[len(message.Photo)-1] // Get largest photo
		fileID = &photo.FileID
		fname := fmt.Sprintf("teacher_announcement_%d.jpg", telegramID)
		filename = &fname
	} else if message.Document != nil {
		// Check if document is an image (including HEIC for iPhone)
		mimeType := message.Document.MimeType
		if mimeType == "image/jpeg" || mimeType == "image/jpg" || mimeType == "image/png" ||
		   mimeType == "image/gif" || mimeType == "image/heic" || mimeType == "image/heif" {
			fileID = &message.Document.FileID
			fname := message.Document.FileName
			if fname == "" {
				fname = fmt.Sprintf("teacher_announcement_%d.jpg", telegramID)
			}
			filename = &fname
		} else {
			text := "❌ Noto'g'ri fayl formati. Iltimos, rasm formatini yuboring (JPG, PNG, GIF, HEIC).\n\n" +
				"❌ Неверный формат файла. Пожалуйста, отправьте изображение в формате JPG, PNG, GIF или HEIC."
			return botService.TelegramService.SendMessage(chatID, text, nil)
		}
	} else if message.Text != "" {
		// User sent text instead of image - show a helpful error
		text := "❌ Iltimos, rasm yuboring yoki 'O'tkazib yuborish' tugmasini bosing.\n\n" +
			"❌ Пожалуйста, отправьте изображение или нажмите кнопку 'Пропустить'."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	} else {
		text := "❌ Iltimos, rasm yuboring.\n\n❌ Пожалуйста, отправьте изображение."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Save announcement with file
	return saveTeacherAnnouncement(botService, telegramID, chatID, stateData, fileID, filename, &fileType)
}

// HandleTeacherAnnouncementSkipFile handles skipping file upload for teacher announcements
func HandleTeacherAnnouncementSkipFile(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get state data
	stateData, err := botService.StateManager.GetData(telegramID)
	if err != nil || stateData == nil {
		text := "❌ Sessiya tugagan. Iltimos, qaytadan boshlang.\n\n" +
			"❌ Сессия истекла. Пожалуйста, начните заново."
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Answer callback query
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅")

	// Save announcement without file
	return saveTeacherAnnouncement(botService, telegramID, chatID, stateData, nil, nil, nil)
}

// saveTeacherAnnouncement saves the teacher's announcement to database
func saveTeacherAnnouncement(botService *services.BotService, telegramID int64, chatID int64, stateData *models.StateData, fileID, filename, fileType *string) error {
	// Get teacher
	teacher, err := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if err != nil || teacher == nil {
		text := "❌ O'qituvchi ma'lumotlari topilmadi / Данные учителя не найдены"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Create announcement
	req := &models.CreateAnnouncementRequest{
		Content:           stateData.AnnouncementText,
		TelegramFileID:    fileID,
		Filename:          filename,
		FileType:          fileType,
		PostedByTeacherID: &teacher.ID,
		ClassIDs:          stateData.SelectedClasses,
		BranchID:          teacher.BranchID,
	}

	if len(req.ClassIDs) == 0 {
		return botService.TelegramService.SendMessage(chatID, "❌ Sinf tanlanmagan / Класс не выбран", nil)
	}
	for _, classID := range req.ClassIDs {
		if !botService.ClassInBranch(classID, teacher.BranchID) {
			return botService.TelegramService.SendMessage(chatID, msgNotAllowed, nil)
		}
	}

	announcement, err := botService.AnnouncementService.CreateAnnouncement(req)
	if err != nil {
		log.Printf("Failed to create announcement: %v", err)
		text := "❌ E'lon yaratishda xatolik / Ошибка при создании объявления"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Deliver to parents of the selected classes
	go notifyUsersAboutAnnouncement(botService, announcement)

	// Get class names
	classNames := []string{}
	for _, classID := range stateData.SelectedClasses {
		class, err := botService.ClassRepo.GetByID(classID)
		if err == nil && class != nil {
			classNames = append(classNames, class.ClassName)
		}
	}

	imageInfo := ""
	if fileID != nil {
		imageInfo = "🖼 Rasm qo'shildi / Изображение добавлено\n"
	}

	// Success message
	text := fmt.Sprintf(
		"✅ <b>E'lon muvaffaqiyatli yaratildi!</b>\n\n"+
			"ID: <code>%d</code>\n"+
			"%s"+
			"Sinflar / Классы: <b>%s</b>\n\n"+
			"E'lon ota-onalarga ko'rsatilmoqda.\n\n"+
			"✅ <b>Объявление успешно создано!</b>\n\n"+
			"ID: <code>%d</code>\n"+
			"%s"+
			"Классы: <b>%s</b>\n\n"+
			"Объявление показывается родителям.",
		announcement.ID, imageInfo, fmt.Sprintf("%v", classNames),
		announcement.ID, imageInfo, fmt.Sprintf("%v", classNames),
	)

	return botService.TelegramService.SendMessage(chatID, text, nil)
}
