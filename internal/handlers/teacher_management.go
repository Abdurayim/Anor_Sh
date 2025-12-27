package handlers

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
	"parent-bot/internal/utils"
	"parent-bot/internal/validator"
)

// HandleAddTeacherCommand initiates adding a new teacher (admin only)
func HandleAddTeacherCommand(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Check if user is admin
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	isAdmin := false
	if user != nil {
		isAdmin, _ = botService.IsAdmin(user.PhoneNumber, user.TelegramID)
	}

	if !isAdmin {
		text := "❌ Bu buyruq faqat administratorlar uchun / Эта команда только для администраторов"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	text := "👨‍🏫 <b>Yangi o'qituvchi qo'shish / Добавить нового учителя</b>\n\n" +
		"Iltimos, o'qituvchining to'liq ismini kiriting:\n" +
		"Пожалуйста, введите полное имя учителя:\n\n" +
		"<b>Misol / Пример:</b> Shahlo Rahimova"

	// Set state
	stateData := &models.StateData{}
	err = botService.StateManager.Set(telegramID, "awaiting_teacher_full_name", stateData)
	if err != nil {
		return err
	}

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleTeacherFullName processes teacher full name input
func HandleTeacherFullName(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	fullName := strings.TrimSpace(message.Text)

	// Split full name into first and last name
	parts := strings.Fields(fullName)
	if len(parts) < 2 {
		text := "❌ Iltimos, ism va familiyani kiriting.\n\n" +
			"❌ Пожалуйста, введите имя и фамилию.\n\n" +
			"<b>Misol / Пример:</b> Shahlo Rahimova"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	firstName := parts[0]
	lastName := strings.Join(parts[1:], " ")

	// Validate names
	if len(firstName) < 2 || len(firstName) > 100 {
		text := "❌ Ism juda qisqa yoki juda uzun (2-100 ta belgi).\n\n" +
			"❌ Имя слишком короткое или слишком длинное (2-100 символов)."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}
	if len(lastName) < 2 || len(lastName) > 100 {
		text := "❌ Familiya juda qisqa yoki juda uzun (2-100 ta belgi).\n\n" +
			"❌ Фамилия слишком короткая или слишком длинная (2-100 символов)."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Save names and ask for phone
	stateData.TeacherFirstName = firstName
	stateData.TeacherLastName = lastName
	err := botService.StateManager.Set(telegramID, "awaiting_teacher_phone", stateData)
	if err != nil {
		return err
	}

	text := "✅ To'liq ism qabul qilindi: <b>" + firstName + " " + lastName + "</b>\n\n" +
		"Endi telefon raqamini kiriting:\n" +
		"Теперь введите номер телефона:\n\n" +
		"<b>Format / Формат:</b> +998XXXXXXXXX\n" +
		"<b>Misol / Пример:</b> +998901234567"

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleTeacherPhone processes teacher phone number input
func HandleTeacherPhone(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	if stateData.TeacherFirstName == "" || stateData.TeacherLastName == "" {
		text := "❌ Sessiya tugagan. Iltimos, qaytadan boshlang.\n\n❌ Сессия истекла. Пожалуйста, начните заново."
		_ = botService.StateManager.Clear(telegramID)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	firstName := stateData.TeacherFirstName
	lastName := stateData.TeacherLastName
	phoneNumber := strings.TrimSpace(message.Text)

	// Validate and normalize phone number
	validPhone, err := validator.ValidateUzbekPhone(phoneNumber)
	if err != nil {
		text := "❌ Telefon raqami noto'g'ri formatda. Format: +998XXXXXXXXX\n\n" +
			"❌ Неверный формат номера телефона. Формат: +998XXXXXXXXX"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Check if phone already exists
	existingTeacher, err := botService.TeacherRepo.GetByPhone(validPhone)
	if err != nil {
		log.Printf("Error checking existing teacher: %v", err)
	}
	if existingTeacher != nil {
		text := fmt.Sprintf("❌ Bu telefon raqami allaqachon ro'yxatdan o'tgan.\n\n"+
			"❌ Этот номер телефона уже зарегистрирован.\n\n"+
			"O'qituvchi / Учитель: %s %s", existingTeacher.FirstName, existingTeacher.LastName)
		_ = botService.StateManager.Clear(telegramID)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get admin info
	admin, err := botService.AdminRepo.GetByTelegramID(telegramID)
	if err != nil || admin == nil {
		text := "❌ Admin ma'lumotlari topilmadi / Данные администратора не найдены"
		_ = botService.StateManager.Clear(telegramID)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Create teacher with default language "uz"
	language := "uz"
	teacherID, err := botService.TeacherRepo.Create(firstName, lastName, validPhone, language, admin.ID)
	if err != nil {
		log.Printf("Failed to create teacher: %v", err)
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		_ = botService.StateManager.Clear(telegramID)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Success message with created teacher info
	text := fmt.Sprintf(
		"✅ <b>O'qituvchi muvaffaqiyatli qo'shildi!</b>\n\n"+
			"👨‍🏫 <b>Ma'lumotlar / Данные:</b>\n"+
			"━━━━━━━━━━━━━━━━━━━━\n"+
			"📌 ID: <code>%d</code>\n"+
			"👤 Ism-Familiya: <b>%s %s</b>\n"+
			"📱 Telefon: <code>%s</code>\n"+
			"🌐 Til: <b>%s</b>\n"+
			"━━━━━━━━━━━━━━━━━━━━\n\n"+
			"📝 O'qituvchi botni ishga tushirganda telefon raqamini ulashishi kerak.\n\n"+
			"✅ <b>Учитель успешно добавлен!</b>\n\n"+
			"📝 Учитель должен поделиться номером телефона при запуске бота.",
		teacherID, firstName, lastName, validPhone, language,
	)

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleListTeachersCommand lists all teachers (admin only)
func HandleListTeachersCommand(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Check if user is admin
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	isAdmin := false
	if user != nil {
		isAdmin, _ = botService.IsAdmin(user.PhoneNumber, user.TelegramID)
	}

	if !isAdmin {
		text := "❌ Bu buyruq faqat administratorlar uchun / Эта команда только для администраторов"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get all teachers
	teachers, err := botService.TeacherRepo.GetAll(100, 0)
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(teachers) == 0 {
		text := "📝 Hozircha o'qituvchilar yo'q.\n\n📝 Пока нет учителей."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Format teacher list
	text := fmt.Sprintf("👨‍🏫 <b>O'qituvchilar ro'yxati / Список учителей</b>\n\n"+
		"Jami: <b>%d</b> ta\n\n", len(teachers))

	for i, t := range teachers {
		status := "✅"
		registered := "✓ Ro'yxatdan o'tgan / Зарегистрирован"
		if !t.IsActive {
			status = "❌"
			registered = "✗ Faol emas / Неактивен"
		} else if t.TelegramID == nil {
			registered = "⏳ Kutilmoqda / Ожидается"
		}

		text += fmt.Sprintf("%d. %s <b>%s %s</b>\n"+
			"   Tel: <code>%s</code> | ID: <code>%d</code>\n"+
			"   %s\n\n",
			i+1, status, t.FirstName, t.LastName,
			t.PhoneNumber, t.ID,
			registered)
	}

	text += "\n💡 O'qituvchi botni ishga tushirib telefon raqamini ulashganda avtomatik faollashadi"

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleTeacherRegistration handles teacher registration when they share phone
func HandleTeacherRegistration(botService *services.BotService, message *tgbotapi.Message, phoneNumber string) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Normalize phone number
	validPhone, err := validator.ValidateUzbekPhone(phoneNumber)
	if err != nil {
		text := "❌ Telefon raqami noto'g'ri formatda.\n\n❌ Неверный формат номера телефона."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Find teacher by phone
	teacher, err := botService.TeacherRepo.GetByPhone(validPhone)
	if err != nil {
		log.Printf("Error finding teacher: %v", err)
		text := "❌ Xatolik yuz berdi / Произошла ошибка"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if teacher == nil {
		// Not a teacher, try parent registration
		return HandlePhoneNumber(botService, message, &models.StateData{Language: "uz"})
	}

	// Check if already registered
	if teacher.TelegramID != nil && *teacher.TelegramID == telegramID {
		lang := i18n.GetLanguage(teacher.Language)
		text := "✅ Siz allaqachon ro'yxatdan o'tgansiz!\n\n✅ Вы уже зарегистрированы!"
		keyboard := utils.MakeTeacherMainMenuKeyboard(lang)
		return botService.TelegramService.SendMessage(chatID, text, keyboard)
	}

	// Register teacher
	err = botService.TeacherRepo.UpdateTelegramID(teacher.ID, telegramID, message.From.UserName)
	if err != nil {
		log.Printf("Failed to register teacher: %v", err)
		text := "❌ Ro'yxatdan o'tishda xatolik / Ошибка при регистрации"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Success message
	lang := i18n.GetLanguage(teacher.Language)
	text := fmt.Sprintf(
		"✅ <b>Ro'yxatdan o'tish muvaffaqiyatli yakunlandi!</b>\n\n"+
			"Xush kelibsiz, <b>%s %s</b>!\n"+
			"Siz o'qituvchi sifatida tizimga kiritildingiz.\n\n"+
			"✅ <b>Регистрация успешно завершена!</b>\n\n"+
			"Добро пожаловать, <b>%s %s</b>!\n"+
			"Вы вошли в систему как учитель.",
		teacher.FirstName, teacher.LastName,
		teacher.FirstName, teacher.LastName,
	)

	keyboard := utils.MakeTeacherMainMenuKeyboard(lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleTeacherMessage routes messages from registered teachers
// IMPORTANT: This function handles ALL teacher states internally and NEVER falls through to parent handlers
func HandleTeacherMessage(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID
	lang := i18n.GetLanguage(teacher.Language)

	// Check if it's a command (already handled by HandleCommand in webhook)
	if message.IsCommand() {
		return nil
	}

	// IMPORTANT: Check for main menu button presses FIRST - these should override any active state
	buttonText := message.Text
	if isTeacherMenuButton(buttonText) {
		// Clear any existing state when pressing menu buttons
		_ = botService.StateManager.Clear(telegramID)
		return HandleTeacherMainMenu(botService, message, teacher)
	}

	// Check if teacher has an active state
	state, err := botService.StateManager.Get(telegramID)
	if err != nil {
		log.Printf("[TEACHER] Error getting state for %d: %v", telegramID, err)
		// Clear any bad state and show teacher menu
		_ = botService.StateManager.Clear(telegramID)
		return showTeacherMainMenu(botService, chatID, lang)
	}

	// If teacher has a state, route by teacher-specific state handler
	if state != nil && state.State != "" {
		stateData, err := botService.StateManager.GetData(telegramID)
		if err != nil {
			log.Printf("[TEACHER] Error getting state data for %d: %v", telegramID, err)
			_ = botService.StateManager.Clear(telegramID)
			return showTeacherMainMenu(botService, chatID, lang)
		}

		// Route ONLY teacher states - never fall through to parent states
		return routeTeacherState(botService, message, teacher, state.State, stateData)
	}

	// No state - route to teacher menu handler
	return HandleTeacherMainMenu(botService, message, teacher)
}

// isTeacherMenuButton checks if the button text is one of the teacher main menu buttons
func isTeacherMenuButton(buttonText string) bool {
	// Check all teacher menu buttons in both languages
	teacherButtons := []string{
		i18n.Get(i18n.BtnAddStudent, i18n.LanguageUzbek),
		i18n.Get(i18n.BtnAddStudent, i18n.LanguageRussian),
		i18n.Get(i18n.BtnViewClassStudents, i18n.LanguageUzbek),
		i18n.Get(i18n.BtnViewClassStudents, i18n.LanguageRussian),
		i18n.Get(i18n.BtnMarkAttendance, i18n.LanguageUzbek),
		i18n.Get(i18n.BtnMarkAttendance, i18n.LanguageRussian),
		i18n.Get(i18n.BtnAddTestResult, i18n.LanguageUzbek),
		i18n.Get(i18n.BtnAddTestResult, i18n.LanguageRussian),
		i18n.Get(i18n.BtnPostAnnouncement, i18n.LanguageUzbek),
		i18n.Get(i18n.BtnPostAnnouncement, i18n.LanguageRussian),
	}

	for _, btn := range teacherButtons {
		if buttonText == btn {
			return true
		}
	}
	return false
}

// routeTeacherState handles teacher-specific states only
// If state doesn't match any teacher state, clears it and processes button press
func routeTeacherState(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher, state string, stateData *models.StateData) error {
	telegramID := message.From.ID

	switch state {
	case "teacher_selecting_announcement_classes":
		// Waiting for callback selection - ignore text messages
		return nil

	case "taking_attendance":
		// Waiting for callback selection (toggle/finish buttons) - ignore text messages
		return nil

	case "teacher_awaiting_announcement_content":
		return HandleTeacherAnnouncementContent(botService, message)

	case "teacher_awaiting_announcement_file":
		return HandleTeacherAnnouncementFile(botService, message, stateData)

	case "teacher_editing_announcement_content":
		return HandleTeacherEditedAnnouncementContent(botService, message)

	case "teacher_awaiting_student_name":
		return HandleTeacherStudentNameInput(botService, message, stateData)

	case "teacher_awaiting_test_results_text":
		return HandleTeacherTestResultsTextInput(botService, message, stateData)

	default:
		// Unknown or stale state (like 'registered' from parent flow) - clear it and PROCESS the button
		log.Printf("[TEACHER] Unknown state '%s' for teacher %d, clearing and processing button press", state, telegramID)
		_ = botService.StateManager.Clear(telegramID)
		// Process the button press instead of just showing menu
		return HandleTeacherMainMenu(botService, message, teacher)
	}
}

// showTeacherMainMenu displays the teacher main menu with keyboard
func showTeacherMainMenu(botService *services.BotService, chatID int64, lang i18n.Language) error {
	text := "👨‍🏫 Bosh menyu / Главное меню"
	keyboard := utils.MakeTeacherMainMenuKeyboard(lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleTeacherMainMenu handles teacher menu button presses
func HandleTeacherMainMenu(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	buttonText := message.Text
	chatID := message.Chat.ID
	lang := i18n.GetLanguage(teacher.Language)

	// Add student
	if buttonText == i18n.Get(i18n.BtnAddStudent, i18n.LanguageUzbek) ||
	   buttonText == i18n.Get(i18n.BtnAddStudent, i18n.LanguageRussian) {
		return HandleTeacherManageStudentsCommand(botService, message, teacher)
	}

	// View class students
	if buttonText == i18n.Get(i18n.BtnViewClassStudents, i18n.LanguageUzbek) ||
	   buttonText == i18n.Get(i18n.BtnViewClassStudents, i18n.LanguageRussian) {
		return HandleTeacherManageStudentsCommand(botService, message, teacher)
	}

	// Mark attendance
	if buttonText == i18n.Get(i18n.BtnMarkAttendance, i18n.LanguageUzbek) ||
	   buttonText == i18n.Get(i18n.BtnMarkAttendance, i18n.LanguageRussian) {
		return HandleTeacherTakeAttendanceCommand(botService, message, teacher)
	}

	// Add test result
	if buttonText == i18n.Get(i18n.BtnAddTestResult, i18n.LanguageUzbek) ||
	   buttonText == i18n.Get(i18n.BtnAddTestResult, i18n.LanguageRussian) {
		return HandleTeacherEnterGradesCommand(botService, message, teacher)
	}

	// Post announcement
	if buttonText == i18n.Get(i18n.BtnPostAnnouncement, i18n.LanguageUzbek) ||
	   buttonText == i18n.Get(i18n.BtnPostAnnouncement, i18n.LanguageRussian) {
		return HandleTeacherPostAnnouncementCommand(botService, message, teacher)
	}

	// Default: show main menu
	text := "👨‍🏫 Bosh menyu / Главное меню"
	keyboard := utils.MakeTeacherMainMenuKeyboard(lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// IsTeacher checks if user is a registered teacher
func IsTeacher(botService *services.BotService, telegramID int64) (*models.Teacher, bool) {
	teacher, err := botService.TeacherRepo.GetByTelegramID(telegramID)
	if err != nil || teacher == nil {
		return nil, false
	}
	return teacher, teacher.IsActive
}

// HandleTeacherManageStudentsCommand allows teacher to manage students
func HandleTeacherManageStudentsCommand(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID
	lang := i18n.GetLanguage(teacher.Language)

	// Get all classes (teachers can access all classes)
	classes, err := botService.ClassRepo.GetAll()
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(classes) == 0 {
		text := "📚 Hozircha sinflar yo'q. Admin sinf qo'shishi kerak.\n\n" +
			"📚 Пока нет классов. Администратор должен добавить классы."
		keyboard := utils.MakeTeacherMainMenuKeyboard(lang)
		return botService.TelegramService.SendMessage(chatID, text, keyboard)
	}

	// Create inline keyboard for class selection
	var buttons [][]tgbotapi.InlineKeyboardButton

	text := "👥 <b>O'quvchilarni boshqarish / Управление учениками</b>\n\n" +
		"Sinfni tanlang:\n" +
		"Выберите класс:"

	for _, class := range classes {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("📚 %s", class.ClassName),
			fmt.Sprintf("teacher_manage_class_%d", class.ID),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	// Add back button
	backButton := tgbotapi.NewInlineKeyboardButtonData(
		"⬅️ Orqaga / Назад",
		"teacher_back_to_main",
	)
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{backButton})

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	// Clear any existing state
	_ = botService.StateManager.Clear(telegramID)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err = botService.Bot.Send(msg)
	return err
}

// HandleTeacherManageClassCallback handles when teacher selects a class to manage
func HandleTeacherManageClassCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, classID int) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get teacher
	teacher, err := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if err != nil || teacher == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	lang := i18n.GetLanguage(teacher.Language)

	// Get class info
	class, err := botService.ClassRepo.GetByID(classID)
	if err != nil || class == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Sinf topilmadi")
		return nil
	}

	// Get students in this class
	students, err := botService.StudentRepo.GetByClassID(classID)
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Format text
	text := fmt.Sprintf("📚 <b>%s sinfi / Класс %s</b>\n\n", class.ClassName, class.ClassName)
	text += fmt.Sprintf("O'quvchilar soni / Количество учеников: <b>%d</b>\n\n", len(students))

	if len(students) == 0 {
		text += "📝 Bu sinfda hali o'quvchilar yo'q.\n\n📝 В этом классе пока нет учеников."
	} else {
		text += "<b>O'quvchilarni tanlang / Выберите ученика:</b>\n"
		text += "🗑 O'chirish uchun ismni bosing / Нажмите на имя для удаления"
	}

	// Create action buttons
	var buttons [][]tgbotapi.InlineKeyboardButton

	// Add each student as a button with delete icon
	for _, student := range students {
		studentBtn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("🗑 %s %s", student.LastName, student.FirstName),
			fmt.Sprintf("teacher_delete_student_%d_%d", classID, student.ID),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{studentBtn})
	}

	// Add student button
	addButton := tgbotapi.NewInlineKeyboardButtonData(
		"➕ O'quvchi qo'shish / Добавить ученика",
		fmt.Sprintf("teacher_add_student_%d", classID),
	)
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{addButton})

	// Back button
	backButton := tgbotapi.NewInlineKeyboardButtonData(
		"⬅️ Orqaga / Назад",
		"teacher_manage_students_back",
	)
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{backButton})

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Delete previous message and send new one
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
	_, _ = botService.Bot.Request(deleteMsg)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err = botService.Bot.Send(msg)
	_ = lang // for future use
	return err
}

// HandleTeacherAddStudentCallback handles initiating adding a student
func HandleTeacherAddStudentCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, classID int) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get teacher
	teacher, err := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if err != nil || teacher == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	// Get class info
	class, err := botService.ClassRepo.GetByID(classID)
	if err != nil || class == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Sinf topilmadi")
		return nil
	}

	// Set state for adding student
	stateData := &models.StateData{
		ClassID: &classID,
	}
	err = botService.StateManager.Set(telegramID, "teacher_awaiting_student_name", stateData)
	if err != nil {
		log.Printf("Failed to set state: %v", err)
	}

	text := fmt.Sprintf(
		"👤 <b>O'quvchi qo'shish / Добавить ученика</b>\n\n"+
			"Sinf / Класс: <b>%s</b>\n\n"+
			"Iltimos, o'quvchining ism-familiyasini yuboring:\n"+
			"Пожалуйста, отправьте имя-фамилию ученика:\n\n"+
			"<b>Format / Формат:</b>\n"+
			"<code>Ism Familiya</code>\n\n"+
			"<b>Misol / Пример:</b>\n"+
			"<code>Aziz Karimov</code>",
		class.ClassName,
	)

	// Create cancel button
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"❌ Bekor qilish / Отмена",
				fmt.Sprintf("teacher_manage_class_%d", classID),
			),
		),
	)

	// Delete previous message
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
	_, _ = botService.Bot.Request(deleteMsg)

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err = botService.Bot.Send(msg)
	return err
}

// HandleTeacherStudentNameInput handles student name input from teacher
func HandleTeacherStudentNameInput(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Get teacher
	teacher, err := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if err != nil || teacher == nil {
		text := "❌ Xatolik / Ошибка"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	lang := i18n.GetLanguage(teacher.Language)

	// Validate state data
	if stateData.ClassID == nil {
		text := "❌ Sessiya tugagan. Iltimos, qaytadan boshlang.\n\n" +
			"❌ Сессия истекла. Пожалуйста, начните заново."
		_ = botService.StateManager.Clear(telegramID)
		keyboard := utils.MakeTeacherMainMenuKeyboard(lang)
		return botService.TelegramService.SendMessage(chatID, text, &keyboard)
	}

	classID := *stateData.ClassID

	// Parse name into first and last name
	fullName := strings.TrimSpace(message.Text)
	nameParts := strings.Fields(fullName)

	if len(nameParts) < 2 {
		text := "❌ Iltimos, ism va familiyani kiriting.\n\n" +
			"❌ Пожалуйста, введите имя и фамилию.\n\n" +
			"Format: <code>Ism Familiya</code>"
		// Create cancel button
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"❌ Bekor qilish / Отмена",
					fmt.Sprintf("teacher_manage_class_%d", classID),
				),
			),
		)
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = keyboard
		_, err := botService.Bot.Send(msg)
		return err
	}

	firstName := nameParts[0]
	lastName := strings.Join(nameParts[1:], " ")

	// Create student
	studentReq := &models.CreateStudentRequest{
		FirstName:        firstName,
		LastName:         lastName,
		ClassID:          classID,
		AddedByTeacherID: &teacher.ID,
	}

	log.Printf("[TEACHER ADD STUDENT] Creating student: %s %s in class %d by teacher %d", firstName, lastName, classID, teacher.ID)

	studentID, err := botService.StudentRepo.Create(studentReq)
	if err != nil {
		log.Printf("[TEACHER ADD STUDENT ERROR] Failed to create student: %v", err)
		text := "❌ O'quvchi qo'shishda xatolik / Ошибка при добавлении ученика"
		// Create cancel button
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"◀️ Sinfga qaytish / Назад к классу",
					fmt.Sprintf("teacher_manage_class_%d", classID),
				),
			),
		)
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = keyboard
		_, sendErr := botService.Bot.Send(msg)
		return sendErr
	}

	log.Printf("[TEACHER ADD STUDENT] Student created successfully with ID: %d", studentID)

	// Clear state
	_ = botService.StateManager.Clear(telegramID)
	log.Printf("[TEACHER ADD STUDENT] State cleared for telegram ID: %d", telegramID)

	// Get class info
	class, _ := botService.ClassRepo.GetByID(classID)
	className := fmt.Sprintf("%d", classID)
	if class != nil {
		className = class.ClassName
	}

	// Success message
	text := fmt.Sprintf(
		"✅ <b>O'quvchi muvaffaqiyatli qo'shildi!</b>\n\n"+
			"📌 ID: <code>%d</code>\n"+
			"👤 Ism-Familiya: <b>%s %s</b>\n"+
			"📚 Sinf: <b>%s</b>\n\n"+
			"✅ <b>Ученик успешно добавлен!</b>\n\n"+
			"📌 ID: <code>%d</code>\n"+
			"👤 Имя-Фамилия: <b>%s %s</b>\n"+
			"📚 Класс: <b>%s</b>",
		studentID, firstName, lastName, className,
		studentID, firstName, lastName, className,
	)

	// Create keyboard with "Add More" and "Back" buttons
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"➕ Yana o'quvchi qo'shish / Добавить ещё",
				fmt.Sprintf("teacher_add_student_%d", classID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"◀️ Sinfga qaytish / Назад к классу",
				fmt.Sprintf("teacher_manage_class_%d", classID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"🏠 Bosh menyu / Главное меню",
				"teacher_back_to_main",
			),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	log.Printf("[TEACHER ADD STUDENT] Sending success message with student info and buttons")
	_, err = botService.Bot.Send(msg)
	if err != nil {
		log.Printf("[TEACHER ADD STUDENT ERROR] Failed to send success message: %v", err)
	} else {
		log.Printf("[TEACHER ADD STUDENT] Success message sent successfully")
	}
	return err
}

// HandleTeacherManageStudentsBackCallback handles back button in manage students
func HandleTeacherManageStudentsBackCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get teacher
	teacher, err := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if err != nil || teacher == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	lang := i18n.GetLanguage(teacher.Language)

	// Get all classes (teachers can access all classes)
	classes, err := botService.ClassRepo.GetAll()
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Create inline keyboard for class selection
	var buttons [][]tgbotapi.InlineKeyboardButton

	text := "👥 <b>O'quvchilarni boshqarish / Управление учениками</b>\n\n" +
		"Sinfni tanlang:\n" +
		"Выберите класс:"

	for _, class := range classes {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("📚 %s", class.ClassName),
			fmt.Sprintf("teacher_manage_class_%d", class.ID),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	// Add back button
	backButton := tgbotapi.NewInlineKeyboardButtonData(
		"⬅️ Orqaga / Назад",
		"teacher_back_to_main",
	)
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{backButton})

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	// Delete previous message and send new one
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
	_, _ = botService.Bot.Request(deleteMsg)

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err = botService.Bot.Send(msg)
	_ = lang // for future use
	return err
}

// HandleTeacherBackToMainCallback handles going back to teacher main menu
func HandleTeacherBackToMainCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get teacher
	teacher, err := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if err != nil || teacher == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	lang := i18n.GetLanguage(teacher.Language)

	// Delete previous message
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
	_, _ = botService.Bot.Request(deleteMsg)

	// Send main menu
	text := "👨‍🏫 Bosh menyu / Главное меню"
	keyboard := utils.MakeTeacherMainMenuKeyboard(lang)

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleTeacherEnterGradesCommand is now in test_results.go
// HandleTeacherTakeAttendanceCommand is now in attendance.go

// HandleTeacherPostAnnouncementCommand allows teacher to post announcements
func HandleTeacherPostAnnouncementCommand(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID
	lang := i18n.GetLanguage(teacher.Language)

	// Get all classes (teachers can access all classes)
	classes, err := botService.ClassRepo.GetAll()
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка bazы danних"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(classes) == 0 {
		text := "📚 Hozircha sinflar yo'q. Admin sinf qo'shishi kerak.\n\n" +
			"📚 Пока нет классов. Администратор должен добавить классы."
		keyboard := utils.MakeTeacherMainMenuKeyboard(lang)
		return botService.TelegramService.SendMessage(chatID, text, keyboard)
	}

	// Create inline keyboard for class selection (multiple selection)
	var buttons [][]tgbotapi.InlineKeyboardButton

	text := "📢 <b>E'lon qo'shish / Добавить объявление</b>\n\n" +
		"Qaysi sinf(lar) uchun e'lon qo'shmoqchisiz?\n" +
		"Для какого класса(ов) хотите добавить объявление?\n\n" +
		"Bir nechta sinfni tanlashingiz mumkin:\n" +
		"Вы можете выбрать несколько классов:"

	// Add toggle buttons for each class
	for _, class := range classes {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("☐ %s", class.ClassName),
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

	// Initialize state for multi-class selection
	stateData := &models.StateData{
		SelectedClasses: []int{}, // Empty initially
	}
	err = botService.StateManager.Set(telegramID, "teacher_selecting_announcement_classes", stateData)
	if err != nil {
		log.Printf("Failed to set state: %v", err)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err = botService.Bot.Send(msg)
	return err
}

// HandleTeacherMyAnnouncementsCommand shows teacher's announcements
func HandleTeacherMyAnnouncementsCommand(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	chatID := message.Chat.ID
	lang := i18n.GetLanguage(teacher.Language)

	// Get teacher's announcements
	announcements, err := botService.AnnouncementRepo.GetByTeacherID(teacher.ID, 100, 0)
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(announcements) == 0 {
		text := "📊 Sizda hali e'lonlar yo'q.\n\n📊 У вас еще нет объявлений."
		keyboard := utils.MakeTeacherMainMenuKeyboard(lang)
		return botService.TelegramService.SendMessage(chatID, text, keyboard)
	}

	// Send each announcement with edit/delete buttons
	for i, announcement := range announcements {
		// Format announcement text
		text := fmt.Sprintf("📢 E'lon / Объявление #%d\n\n", i+1)

		if announcement.Title != nil && *announcement.Title != "" {
			text += fmt.Sprintf("<b>%s</b>\n\n", *announcement.Title)
		}

		text += announcement.Content
		text += fmt.Sprintf("\n\n📅 %s", announcement.CreatedAt.Format("02.01.2006 15:04"))

		statusEmoji := "✅"
		statusText := "Faol / Активно"
		if !announcement.IsActive {
			statusEmoji = "❌"
			statusText = "Nofaol / Неактивно"
		}
		text += fmt.Sprintf("\n%s %s", statusEmoji, statusText)

		// Create inline keyboard with edit and delete buttons
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"✏️ Tahrirlash / Редактировать",
					fmt.Sprintf("teacher_announcement_edit_%d", announcement.ID),
				),
				tgbotapi.NewInlineKeyboardButtonData(
					"🗑 O'chirish / Удалить",
					fmt.Sprintf("teacher_announcement_delete_%d", announcement.ID),
				),
			),
		)

		// Send announcement with image if available
		if announcement.TelegramFileID != nil && *announcement.TelegramFileID != "" {
			photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(*announcement.TelegramFileID))
			photo.Caption = text
			photo.ParseMode = "HTML"
			photo.ReplyMarkup = keyboard
			_, err = botService.Bot.Send(photo)
			if err != nil {
				log.Printf("Failed to send photo: %v", err)
				// Fallback to text only
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ParseMode = "HTML"
				msg.ReplyMarkup = keyboard
				_, _ = botService.Bot.Send(msg)
			}
		} else {
			// Send text only
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = "HTML"
			msg.ReplyMarkup = keyboard
			_, err = botService.Bot.Send(msg)
			if err != nil {
				log.Printf("Failed to send announcement: %v", err)
			}
		}
	}

	return nil
}

// HandleTeacherSettingsCommand shows teacher settings
func HandleTeacherSettingsCommand(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	chatID := message.Chat.ID
	text := fmt.Sprintf(
		"⚙️ <b>Sozlamalar / Настройки</b>\n\n"+
			"Ism: <b>%s %s</b>\n"+
			"Telefon: <code>%s</code>\n"+
			"Til / Язык: <b>%s</b>\n\n"+
			"Имя: <b>%s %s</b>\n"+
			"Телефон: <code>%s</code>\n"+
			"Язык: <b>%s</b>",
		teacher.FirstName, teacher.LastName, teacher.PhoneNumber, teacher.Language,
		teacher.FirstName, teacher.LastName, teacher.PhoneNumber, teacher.Language,
	)

	lang := i18n.GetLanguage(teacher.Language)
	keyboard := utils.MakeTeacherMainMenuKeyboard(lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleTeacherDeleteStudentCallback handles deleting a student (teacher)
func HandleTeacherDeleteStudentCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, classID, studentID int) error {
	telegramID := callback.From.ID

	// Get teacher
	teacher, err := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if err != nil || teacher == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	// Get student info before deleting
	student, err := botService.StudentRepo.GetByID(studentID)
	if err != nil || student == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ O'quvchi topilmadi / Ученик не найден")
		return nil
	}

	// Delete the student
	err = botService.StudentRepo.Delete(studentID)
	if err != nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik yuz berdi / Произошла ошибка")
		return nil
	}

	// Success feedback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, fmt.Sprintf("✅ %s %s o'chirildi", student.LastName, student.FirstName))

	// Refresh the class view
	return HandleTeacherManageClassCallback(botService, callback, classID)
}
