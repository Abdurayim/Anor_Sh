package handlers

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
)

// HandleMyChildrenCommand shows parent's children and allows selection
func HandleMyChildrenCommand(botService *services.BotService, message *tgbotapi.Message) error {
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
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(children) == 0 {
		text := "📝 Sizga hali farzandlar bog'lanmagan. Iltimos, ma'muriyatga murojaat qiling.\n\n" +
			"📝 К вам еще не привязаны дети. Пожалуйста, обратитесь к администрации."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Create inline keyboard for child selection
	var buttons [][]tgbotapi.InlineKeyboardButton

	for _, child := range children {
		marker := ""
		if user.CurrentSelectedStudentID != nil && *user.CurrentSelectedStudentID == child.StudentID {
			marker = " ✅"
		}

		buttonText := fmt.Sprintf("%s %s%s - %s",
			child.StudentFirstName, child.StudentLastName, marker, child.ClassName)

		button := tgbotapi.NewInlineKeyboardButtonData(
			buttonText,
			fmt.Sprintf("select_child_%d", child.StudentID),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	text := "👨‍👩‍👧‍👦 <b>Farzandlaringiz / Ваши дети</b>\n\n" +
		"Qaysi farzandingiz haqida ma'lumot olishni xohlaysiz? Tugmani bosing:\n" +
		"О каком ребенке вы хотите получить информацию? Нажмите кнопку:\n\n" +
		"✅ - hozirda tanlangan / текущий выбранный"

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err = botService.Bot.Send(msg)
	return err
}

// HandleChildSelectionCallback handles child selection from inline buttons
func HandleChildSelectionCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Extract student ID from callback data (format: "select_child_123")
	parts := strings.Split(callback.Data, "_")
	if len(parts) != 3 {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	studentID, err := strconv.Atoi(parts[2])
	if err != nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	if user == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Foydalanuvchi topilmadi")
		return nil
	}

	lang := i18n.GetLanguage(user.Language)

	// Verify student belongs to this parent
	children, err := botService.StudentRepo.GetParentStudents(user.ID)
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	studentBelongsToParent := false
	var selectedChild *models.ParentChild
	for _, child := range children {
		if child.StudentID == studentID {
			studentBelongsToParent = true
			selectedChild = child
			break
		}
	}

	if !studentBelongsToParent {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Bu farzand sizga tegishli emas")
		return nil
	}

	// Update current selected student
	err = botService.UserRepo.Update(user.ID, &models.UpdateUserRequest{
		CurrentSelectedStudentID: &studentID,
	})
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Answer callback
	answerText := fmt.Sprintf("✅ %s %s tanlandi", selectedChild.StudentFirstName, selectedChild.StudentLastName)
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, answerText)

	// Send confirmation message
	text := fmt.Sprintf(
		"✅ <b>Farzand tanlandi / Ребенок выбран</b>\n\n"+
			"Ism / Имя: <b>%s %s</b>\n"+
			"Sinf / Класс: <b>%s</b>\n\n"+
			"Endi barcha ma'lumotlar (dars jadvali, baholar, yo'qlamalar) ushbu farzand uchun ko'rsatiladi.\n\n"+
			"Теперь вся информация (расписание, оценки, посещаемость) будет показана для этого ребенка.",
		selectedChild.StudentFirstName, selectedChild.StudentLastName, selectedChild.ClassName,
	)

	// Delete the selection message
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
	_, _ = botService.Bot.Request(deleteMsg)

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleSwitchChildButton handles the "Switch Child" button press (for parents with multiple children)
func HandleSwitchChildButton(botService *services.BotService, message *tgbotapi.Message) error {
	return HandleMyChildrenCommand(botService, message)
}

// GetCurrentChildInfo returns formatted info about the currently selected child
func GetCurrentChildInfo(botService *services.BotService, user *models.User) string {
	if user.CurrentSelectedStudentID == nil {
		return "❓ Farzand tanlanmagan / Ребенок не выбран"
	}

	student, err := botService.StudentRepo.GetByIDWithClass(*user.CurrentSelectedStudentID)
	if err != nil || student == nil {
		return "❓ Farzand topilmadi / Ребенок не найден"
	}

	return fmt.Sprintf("👤 %s %s (%s)", student.FirstName, student.LastName, student.ClassName)
}
