package handlers

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
	"parent-bot/internal/utils"
)

// HandleMyChildrenCommand shows parent's children with action buttons
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

	childCount := len(children)
	canAddMore := childCount < 4

	if childCount == 0 {
		// No children - show prompt to add first child
		text := i18n.Get(i18n.MsgNoChildrenLinked, lang)

		// Get active classes
		classes, err := botService.ClassRepo.GetActive()
		if err != nil || len(classes) == 0 {
			text += "\n\n" + i18n.Get(i18n.MsgWaitForStudentAdd, lang)
			return botService.TelegramService.SendMessage(chatID, text, nil)
		}

		// Set state for adding child
		err = botService.StateManager.Set(telegramID, models.StateAddingChild, &models.StateData{})
		if err != nil {
			return err
		}

		keyboard := makeClassSelectionKeyboardForMyKids(classes, lang)
		return botService.TelegramService.SendMessage(chatID, text, keyboard)
	}

	// Create text with list of all children
	text := fmt.Sprintf(i18n.Get(i18n.MsgMyKidsMenu, lang), childCount)
	text += "\n\n"

	for i, child := range children {
		text += fmt.Sprintf("%d. <b>%s %s</b>\n   📚 %s\n\n",
			i+1, child.StudentLastName, child.StudentFirstName, child.ClassName)
	}

	// Create inline keyboard with action buttons for each child
	var buttons [][]tgbotapi.InlineKeyboardButton

	for _, child := range children {
		// Row for this child with name
		childButtonText := fmt.Sprintf("👤 %s %s (%s)",
			child.StudentLastName, child.StudentFirstName, child.ClassName)

		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(childButtonText, fmt.Sprintf("view_child_%d", child.StudentID)),
		})
	}

	// Add "Add Another Child" button if under limit
	if canAddMore {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnAddAnotherChild, lang),
				"add_another_child",
			),
		))
	}

	// Add back button
	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			i18n.Get(i18n.BtnBack, lang),
			"back_to_main",
		),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err = botService.Bot.Send(msg)
	return err
}

// HandleChildInfoCallback handles showing child info when parent clicks on child button
func HandleChildInfoCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID

	// Extract student ID from callback data (format: "child_info_123")
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
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, i18n.Get(i18n.ErrDatabaseError, lang))
		return nil
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

	// Answer callback with child name
	answerText := fmt.Sprintf("👤 %s %s", selectedChild.StudentFirstName, selectedChild.StudentLastName)
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, answerText)

	return nil
}

// HandleSelectClassCallback handles class selection during registration
func HandleSelectClassCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Extract class name from callback data (format: "select_class_9A")
	className := strings.TrimPrefix(callback.Data, "select_class_")

	// Get user and state data
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil || user == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	lang := i18n.GetLanguage(user.Language)

	// Get class by name
	class, err := botService.ClassRepo.GetByName(className)
	if err != nil || class == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Sinf topilmadi / Класс не найден")
		return nil
	}

	// Get state data
	stateData, err := botService.StateManager.GetData(telegramID)
	if err != nil {
		stateData = &models.StateData{}
	}

	// Store selected class in state
	stateData.ClassID = &class.ID

	// Set state to selecting child
	err = botService.StateManager.Set(telegramID, models.StateSelectingChild, stateData)
	if err != nil {
		return err
	}

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, fmt.Sprintf("✅ %s", className))

	// Get students in this class
	students, err := botService.StudentRepo.GetByClassID(class.ID)
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(students) == 0 {
		// No students in this class yet
		text := i18n.Get(i18n.MsgWaitForStudentAdd, lang)

		// Back to class selection button
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					i18n.Get(i18n.BtnBack, lang),
					"back_to_class_selection",
				),
			),
		)

		return botService.TelegramService.SendMessage(chatID, text, keyboard)
	}

	// Show student selection
	text := fmt.Sprintf("📚 %s sinfi\n\n%s", className, i18n.Get(i18n.MsgSelectStudent, lang))

	keyboard := makeStudentSelectionKeyboard(students, lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleSelectStudentCallback handles student selection during registration
func HandleSelectStudentCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Extract student ID from callback data (format: "select_student_123")
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
	if err != nil || user == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	lang := i18n.GetLanguage(user.Language)

	// Get student
	student, err := botService.StudentRepo.GetByIDWithClass(studentID)
	if err != nil || student == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, i18n.Get(i18n.MsgChildNotFound, lang))
		return nil
	}

	// NOTE: We allow multiple parents per student (mother + father)
	// The UNIQUE(parent_id, student_id) constraint in database prevents
	// the SAME parent from linking to the SAME student multiple times

	// Link student to parent
	err = botService.StudentRepo.LinkToParent(user.ID, studentID)
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Clear state and show success
	err = botService.StateManager.Clear(telegramID)
	if err != nil {
		return err
	}

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅")

	// Success message
	text := fmt.Sprintf(i18n.Get(i18n.MsgChildLinked, lang), student.LastName, student.FirstName, student.ClassName)

	// Show parent menu
	keyboard := utils.MakeMainMenuKeyboard(lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleSkipChildSelectionCallback handles skipping child selection during registration
func HandleSkipChildSelectionCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil || user == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	lang := i18n.GetLanguage(user.Language)

	// Clear state
	err = botService.StateManager.Clear(telegramID)
	if err != nil {
		return err
	}

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅")

	// Show registration complete message
	text := "✅ Ro'yxatdan o'tish yakunlandi!\n\n" +
		"Farzandingizni keyinroq 'Mening farzandlarim' bo'limidan qo'shishingiz mumkin.\n\n" +
		"✅ Регистрация завершена!\n\n" +
		"Вы сможете добавить ребенка позже в разделе 'Мои дети'."

	keyboard := utils.MakeMainMenuKeyboard(lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleBackToClassSelectionCallback handles going back to class selection
func HandleBackToClassSelectionCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil || user == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	lang := i18n.GetLanguage(user.Language)

	// Get active classes
	classes, err := botService.ClassRepo.GetActive()
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Set state
	stateData, _ := botService.StateManager.GetData(telegramID)
	if stateData == nil {
		stateData = &models.StateData{}
	}
	err = botService.StateManager.Set(telegramID, models.StateSelectingClass, stateData)
	if err != nil {
		return err
	}

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Show class selection
	text := i18n.Get(i18n.MsgAddChildPrompt, lang)
	keyboard := makeClassSelectionKeyboardWithBack(classes, lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// Helper functions to avoid import cycles
func makeStudentSelectionKeyboard(students []*models.StudentWithClass, lang i18n.Language) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, student := range students {
		button := tgbotapi.NewInlineKeyboardButtonData(
			student.LastName+" "+student.FirstName,
			fmt.Sprintf("select_student_%d", student.ID),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}

	// Add back button
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			i18n.Get(i18n.BtnBack, lang),
			"back_to_class_selection",
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func makeClassSelectionKeyboardWithBack(classes []*models.Class, lang i18n.Language) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Create buttons in rows of 3
	var row []tgbotapi.InlineKeyboardButton
	for i, class := range classes {
		button := tgbotapi.NewInlineKeyboardButtonData(
			class.ClassName,
			"select_class_"+class.ClassName,
		)
		row = append(row, button)

		// Add row every 3 buttons or at the end
		if (i+1)%3 == 0 || i == len(classes)-1 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}

	// Add skip button for registration (can add child later)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			i18n.Get(i18n.BtnSkip, lang),
			"skip_child_selection",
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func makeClassSelectionKeyboardForMyKids(classes []*models.Class, lang i18n.Language) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Create buttons in rows of 3
	var row []tgbotapi.InlineKeyboardButton
	for i, class := range classes {
		button := tgbotapi.NewInlineKeyboardButtonData(
			class.ClassName,
			"mykids_class_"+class.ClassName,
		)
		row = append(row, button)

		// Add row every 3 buttons or at the end
		if (i+1)%3 == 0 || i == len(classes)-1 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}

	// Add back button
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			i18n.Get(i18n.BtnBack, lang),
			"back_to_my_kids",
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// HandleAddAnotherChildCallback handles the "Add Another Child" button
func HandleAddAnotherChildCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil || user == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	lang := i18n.GetLanguage(user.Language)

	// Check if parent can add more children
	children, err := botService.StudentRepo.GetParentStudents(user.ID)
	if err != nil {
		return err
	}

	if len(children) >= 4 {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, i18n.Get(i18n.MsgMaxChildrenReached, lang))
		return nil
	}

	// Get active classes
	classes, err := botService.ClassRepo.GetActive()
	if err != nil || len(classes) == 0 {
		text := i18n.Get(i18n.MsgWaitForStudentAdd, lang)
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Set state for adding child
	err = botService.StateManager.Set(telegramID, models.StateAddingChild, &models.StateData{})
	if err != nil {
		return err
	}

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Show class selection
	text := i18n.Get(i18n.MsgAddChildPrompt, lang)
	keyboard := makeClassSelectionKeyboardForMyKids(classes, lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleMyKidsClassCallback handles class selection when adding a child from My Kids
func HandleMyKidsClassCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Extract class name from callback data (format: "mykids_class_9A")
	className := strings.TrimPrefix(callback.Data, "mykids_class_")

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil || user == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	lang := i18n.GetLanguage(user.Language)

	// Get class by name
	class, err := botService.ClassRepo.GetByName(className)
	if err != nil || class == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Sinf topilmadi / Класс не найден")
		return nil
	}

	// Get state data
	stateData, err := botService.StateManager.GetData(telegramID)
	if err != nil {
		stateData = &models.StateData{}
	}

	// Store selected class in state
	stateData.ClassID = &class.ID

	// Set state to selecting child from class
	err = botService.StateManager.Set(telegramID, models.StateSelectingChildFromClass, stateData)
	if err != nil {
		return err
	}

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, fmt.Sprintf("✅ %s", className))

	// Get students in this class
	students, err := botService.StudentRepo.GetByClassID(class.ID)
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(students) == 0 {
		// No students in this class yet
		text := i18n.Get(i18n.MsgWaitForStudentAdd, lang)

		// Back button
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					i18n.Get(i18n.BtnBack, lang),
					"back_to_mykids_class_selection",
				),
			),
		)

		return botService.TelegramService.SendMessage(chatID, text, keyboard)
	}

	// Filter out students that are already linked to this parent
	children, _ := botService.StudentRepo.GetParentStudents(user.ID)
	linkedStudentIDs := make(map[int]bool)
	for _, child := range children {
		linkedStudentIDs[child.StudentID] = true
	}

	var availableStudents []*models.StudentWithClass
	for _, student := range students {
		if !linkedStudentIDs[student.ID] {
			availableStudents = append(availableStudents, student)
		}
	}

	if len(availableStudents) == 0 {
		text := i18n.Get(i18n.MsgChildAlreadyLinked, lang) + "\n\n" + i18n.Get(i18n.MsgWaitForStudentAdd, lang)

		// Back button
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					i18n.Get(i18n.BtnBack, lang),
					"back_to_mykids_class_selection",
				),
			),
		)

		return botService.TelegramService.SendMessage(chatID, text, keyboard)
	}

	// Show student selection
	text := fmt.Sprintf("📚 %s\n\n%s", className, i18n.Get(i18n.MsgSelectStudent, lang))
	keyboard := makeStudentSelectionKeyboardForMyKids(availableStudents, lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

func makeStudentSelectionKeyboardForMyKids(students []*models.StudentWithClass, lang i18n.Language) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, student := range students {
		button := tgbotapi.NewInlineKeyboardButtonData(
			student.LastName+" "+student.FirstName,
			fmt.Sprintf("mykids_student_%d", student.ID),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}

	// Add back button
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			i18n.Get(i18n.BtnBack, lang),
			"back_to_mykids_class_selection",
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// HandleMyKidsStudentCallback handles student selection when adding a child from My Kids
func HandleMyKidsStudentCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Extract student ID from callback data (format: "mykids_student_123")
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
	if err != nil || user == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	lang := i18n.GetLanguage(user.Language)

	// Check if parent can add more children
	children, err := botService.StudentRepo.GetParentStudents(user.ID)
	if err != nil {
		return err
	}

	if len(children) >= 4 {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, i18n.Get(i18n.MsgMaxChildrenReached, lang))
		return nil
	}

	// Get student
	student, err := botService.StudentRepo.GetByIDWithClass(studentID)
	if err != nil || student == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, i18n.Get(i18n.MsgChildNotFound, lang))
		return nil
	}

	// Check if student is already linked to THIS parent
	for _, child := range children {
		if child.StudentID == studentID {
			_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, i18n.Get(i18n.MsgChildAlreadyLinked, lang))
			return nil
		}
	}

	// NOTE: We allow multiple parents per student (mother + father)
	// The UNIQUE(parent_id, student_id) constraint in database prevents
	// the SAME parent from linking to the SAME student multiple times

	// Link student to parent
	err = botService.StudentRepo.LinkToParent(user.ID, studentID)
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Clear state
	err = botService.StateManager.Clear(telegramID)
	if err != nil {
		return err
	}

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅")

	// Success message
	text := fmt.Sprintf(i18n.Get(i18n.MsgChildLinked, lang), student.LastName, student.FirstName, student.ClassName)

	// Show My Kids menu again (with back to main button)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnMyChildren, lang),
				"show_my_kids",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnBack, lang),
				"back_to_main",
			),
		),
	)

	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleBackToMyKidsCallback handles going back to My Kids menu
func HandleBackToMyKidsCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil || user == nil {
		return nil
	}

	lang := i18n.GetLanguage(user.Language)

	// Get parent's children
	children, err := botService.StudentRepo.GetParentStudents(user.ID)
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	childCount := len(children)
	canAddMore := childCount < 4

	if childCount == 0 {
		text := i18n.Get(i18n.MsgNoChildrenLinked, lang)

		// Get active classes
		classes, err := botService.ClassRepo.GetActive()
		if err != nil || len(classes) == 0 {
			text += "\n\n" + i18n.Get(i18n.MsgWaitForStudentAdd, lang)
			return botService.TelegramService.SendMessage(chatID, text, nil)
		}

		err = botService.StateManager.Set(telegramID, models.StateAddingChild, &models.StateData{})
		if err != nil {
			return err
		}

		keyboard := makeClassSelectionKeyboardForMyKids(classes, lang)
		return botService.TelegramService.SendMessage(chatID, text, keyboard)
	}

	// Create text with list of all children
	text := fmt.Sprintf(i18n.Get(i18n.MsgMyKidsMenu, lang), childCount)
	text += "\n\n"

	for i, child := range children {
		text += fmt.Sprintf("%d. <b>%s %s</b>\n   📚 %s\n\n",
			i+1, child.StudentLastName, child.StudentFirstName, child.ClassName)
	}

	var buttons [][]tgbotapi.InlineKeyboardButton

	for _, child := range children {
		childButtonText := fmt.Sprintf("👤 %s %s (%s)",
			child.StudentLastName, child.StudentFirstName, child.ClassName)

		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(childButtonText, fmt.Sprintf("view_child_%d", child.StudentID)),
		})
	}

	if canAddMore {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnAddAnotherChild, lang),
				"add_another_child",
			),
		))
	}

	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			i18n.Get(i18n.BtnBack, lang),
			"back_to_main",
		),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err = botService.Bot.Send(msg)
	return err
}

// HandleBackToMyKidsClassSelectionCallback handles going back to class selection in My Kids
func HandleBackToMyKidsClassSelectionCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil || user == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	lang := i18n.GetLanguage(user.Language)

	// Get active classes
	classes, err := botService.ClassRepo.GetActive()
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Set state
	err = botService.StateManager.Set(telegramID, models.StateAddingChild, &models.StateData{})
	if err != nil {
		return err
	}

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Show class selection
	text := i18n.Get(i18n.MsgAddChildPrompt, lang)
	keyboard := makeClassSelectionKeyboardForMyKids(classes, lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleBackToMainCallback handles going back to main menu
func HandleBackToMainCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Get user
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil || user == nil {
		return nil
	}

	lang := i18n.GetLanguage(user.Language)

	// Show main menu
	text := i18n.Get(i18n.MsgMainMenu, lang)
	keyboard := utils.MakeMainMenuKeyboard(lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleViewChildCallback handles viewing a specific child's info
func HandleViewChildCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Extract student ID from callback data (format: "view_child_123")
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
	if err != nil || user == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	lang := i18n.GetLanguage(user.Language)

	// Verify student belongs to this parent
	children, err := botService.StudentRepo.GetParentStudents(user.ID)
	if err != nil {
		return err
	}

	var selectedChild *models.ParentChild
	for _, child := range children {
		if child.StudentID == studentID {
			selectedChild = child
			break
		}
	}

	if selectedChild == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Bu farzand sizga tegishli emas")
		return nil
	}

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Show child info with action buttons
	text := fmt.Sprintf(i18n.Get(i18n.MsgChildInfo, lang),
		selectedChild.StudentLastName, selectedChild.StudentFirstName, selectedChild.ClassName)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnViewChildAttendance, lang),
				fmt.Sprintf("view_child_attendance_%d", studentID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnViewChildTestResults, lang),
				fmt.Sprintf("view_child_grades_%d", studentID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnBack, lang),
				"back_to_my_kids",
			),
		),
	)

	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleShowMyKidsCallback handles showing my kids menu (for after adding child)
func HandleShowMyKidsCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	return HandleBackToMyKidsCallback(botService, callback)
}
