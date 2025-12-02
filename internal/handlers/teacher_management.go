package handlers

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
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
		"Iltimos, quyidagi formatda ma'lumotlarni yuboring:\n" +
		"Пожалуйста, отправьте данные в следующем формате:\n\n" +
		"<code>Ism Familiya / Имя Фамилия\n" +
		"Telefon raqami / Номер телефона\n" +
		"Til (uz/ru) / Язык (uz/ru)</code>\n\n" +
		"<b>Misol / Пример:</b>\n" +
		"<code>Shahlo Rahimova\n" +
		"+998901234567\n" +
		"uz</code>"

	// Set state
	stateData := &models.StateData{}
	err = botService.StateManager.Set(telegramID, "awaiting_teacher_info", stateData)
	if err != nil {
		return err
	}

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleTeacherInfo processes teacher information input from admin
func HandleTeacherInfo(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Parse input
	lines := strings.Split(strings.TrimSpace(message.Text), "\n")
	if len(lines) < 3 {
		text := "❌ Noto'g'ri format. Iltimos, ism-familiya, telefon va tilni alohida qatorlarda yuboring.\n\n" +
			"❌ Неверный формат. Пожалуйста, отправьте имя-фамилию, телефон и язык на отдельных строках."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	fullName := strings.TrimSpace(lines[0])
	phoneNumber := strings.TrimSpace(lines[1])
	language := strings.ToLower(strings.TrimSpace(lines[2]))

	// Parse name into first and last name
	nameParts := strings.Fields(fullName)
	if len(nameParts) < 2 {
		text := "❌ Iltimos, ism va familiyani kiriting.\n\n" +
			"❌ Пожалуйста, введите имя и фамилию."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	firstName := nameParts[0]
	lastName := strings.Join(nameParts[1:], " ")

	// Validate and normalize phone number
	validPhone, err := validator.ValidateUzbekPhone(phoneNumber)
	if err != nil {
		text := "❌ Telefon raqami noto'g'ri formatda. Format: +998XXXXXXXXX\n\n" +
			"❌ Неверный формат номера телефона. Формат: +998XXXXXXXXX"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Validate language
	if language != "uz" && language != "ru" {
		text := "❌ Til noto'g'ri. Faqat 'uz' yoki 'ru' ruxsat etilgan.\n\n" +
			"❌ Неверный язык. Разрешены только 'uz' или 'ru'."
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
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get admin info
	admin, err := botService.AdminRepo.GetByTelegramID(telegramID)
	if err != nil || admin == nil {
		text := "❌ Admin ma'lumotlari topilmadi / Данные администратора не найдены"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Create teacher
	teacherID, err := botService.TeacherRepo.Create(firstName, lastName, validPhone, language, admin.ID)
	if err != nil {
		log.Printf("Failed to create teacher: %v", err)
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Success message
	text := fmt.Sprintf(
		"✅ <b>O'qituvchi muvaffaqiyatli qo'shildi!</b>\n\n"+
			"ID: <code>%d</code>\n"+
			"Ism: <b>%s %s</b>\n"+
			"Telefon: <code>%s</code>\n"+
			"Til: <b>%s</b>\n\n"+
			"O'qituvchi botni ishga tushirganda telefon raqamini ulashishi kerak.\n\n"+
			"✅ <b>Учитель успешно добавлен!</b>\n\n"+
			"ID: <code>%d</code>\n"+
			"Имя: <b>%s %s</b>\n"+
			"Телефон: <code>%s</code>\n"+
			"Язык: <b>%s</b>\n\n"+
			"Учитель должен поделиться номером телефона при запуске бота.",
		teacherID, firstName, lastName, validPhone, language,
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
		keyboard := MakeTeacherMainKeyboard(lang)
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

	keyboard := MakeTeacherMainKeyboard(lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleTeacherMessage routes messages from registered teachers
func HandleTeacherMessage(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	// Check if it's a command (already handled by HandleCommand in webhook)
	if message.IsCommand() {
		return nil
	}

	// Route to teacher menu handler
	return HandleTeacherMainMenu(botService, message, teacher)
}

// MakeTeacherMainKeyboard creates main menu keyboard for teachers
func MakeTeacherMainKeyboard(lang i18n.Language) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton

	if lang == i18n.LanguageUzbek {
		rows = [][]tgbotapi.KeyboardButton{
			{
				tgbotapi.NewKeyboardButton("📚 Sinflarim"),
				tgbotapi.NewKeyboardButton("👥 O'quvchilarni boshqarish"),
			},
			{
				tgbotapi.NewKeyboardButton("📝 Baholarni kiritish"),
				tgbotapi.NewKeyboardButton("📋 Yo'qlama olish"),
			},
			{
				tgbotapi.NewKeyboardButton("📢 E'lon qo'shish"),
				tgbotapi.NewKeyboardButton("📊 Mening e'lonlarim"),
			},
			{
				tgbotapi.NewKeyboardButton("⚙️ Sozlamalar"),
			},
		}
	} else {
		rows = [][]tgbotapi.KeyboardButton{
			{
				tgbotapi.NewKeyboardButton("📚 Мои классы"),
				tgbotapi.NewKeyboardButton("👥 Управление учениками"),
			},
			{
				tgbotapi.NewKeyboardButton("📝 Ввести оценки"),
				tgbotapi.NewKeyboardButton("📋 Отметить посещаемость"),
			},
			{
				tgbotapi.NewKeyboardButton("📢 Добавить объявление"),
				tgbotapi.NewKeyboardButton("📊 Мои объявления"),
			},
			{
				tgbotapi.NewKeyboardButton("⚙️ Настройки"),
			},
		}
	}

	return tgbotapi.NewReplyKeyboard(rows...)
}

// HandleTeacherMainMenu handles teacher menu button presses
func HandleTeacherMainMenu(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	buttonText := message.Text
	chatID := message.Chat.ID
	lang := i18n.GetLanguage(teacher.Language)

	// My classes
	if buttonText == "📚 Sinflarim" || buttonText == "📚 Мои классы" {
		return HandleTeacherMyClassesCommand(botService, message, teacher)
	}

	// Manage students
	if buttonText == "👥 O'quvchilarni boshqarish" || buttonText == "👥 Управление учениками" {
		return HandleTeacherManageStudentsCommand(botService, message, teacher)
	}

	// Enter grades
	if buttonText == "📝 Baholarni kiritish" || buttonText == "📝 Ввести оценки" {
		return HandleTeacherEnterGradesCommand(botService, message, teacher)
	}

	// Take attendance
	if buttonText == "📋 Yo'qlama olish" || buttonText == "📋 Отметить посещаемость" {
		return HandleTeacherTakeAttendanceCommand(botService, message, teacher)
	}

	// Post announcement
	if buttonText == "📢 E'lon qo'shish" || buttonText == "📢 Добавить объявление" {
		return HandleTeacherPostAnnouncementCommand(botService, message, teacher)
	}

	// My announcements
	if buttonText == "📊 Mening e'lonlarim" || buttonText == "📊 Мои объявления" {
		return HandleTeacherMyAnnouncementsCommand(botService, message, teacher)
	}

	// Settings
	if buttonText == "⚙️ Sozlamalar" || buttonText == "⚙️ Настройки" {
		return HandleTeacherSettingsCommand(botService, message, teacher)
	}

	// Default: show main menu
	text := "👨‍🏫 Bosh menyu / Главное меню"
	keyboard := MakeTeacherMainKeyboard(lang)
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

// Placeholder handlers for teacher menu items (to be implemented)

// HandleTeacherMyClassesCommand shows teacher's assigned classes
func HandleTeacherMyClassesCommand(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	chatID := message.Chat.ID
	lang := i18n.GetLanguage(teacher.Language)

	// Get teacher's classes
	classes, err := botService.TeacherRepo.GetTeacherClasses(teacher.ID)
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(classes) == 0 {
		text := "📚 Sizga hali sinflar biriktirilmagan. Admin bilan bog'laning.\n\n" +
			"📚 Вам еще не назначены классы. Свяжитесь с администратором."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Format classes list
	text := fmt.Sprintf("📚 <b>Mening sinflarim / Мои классы</b>\n\n"+
		"Jami: <b>%d</b> ta\n\n", len(classes))

	for i, class := range classes {
		status := "✅"
		if !class.IsActive {
			status = "❌"
		}
		text += fmt.Sprintf("%d. %s <b>%s</b>\n", i+1, status, class.ClassName)
	}

	keyboard := MakeTeacherMainKeyboard(lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleTeacherManageStudentsCommand allows teacher to manage students
func HandleTeacherManageStudentsCommand(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	chatID := message.Chat.ID
	text := "👥 O'quvchilarni boshqarish / Управление учениками\n\n" +
		"Bu funksiya tez orada qo'shiladi / Эта функция скоро будет добавлена"

	lang := i18n.GetLanguage(teacher.Language)
	keyboard := MakeTeacherMainKeyboard(lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleTeacherEnterGradesCommand is now in test_results.go
// HandleTeacherTakeAttendanceCommand is now in attendance.go

// HandleTeacherPostAnnouncementCommand allows teacher to post announcements
func HandleTeacherPostAnnouncementCommand(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	chatID := message.Chat.ID
	text := "📢 E'lon qo'shish / Добавить объявление\n\n" +
		"Bu funksiya tez orada qo'shiladi / Эта функция скоро будет добавлена"

	lang := i18n.GetLanguage(teacher.Language)
	keyboard := MakeTeacherMainKeyboard(lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleTeacherMyAnnouncementsCommand shows teacher's announcements
func HandleTeacherMyAnnouncementsCommand(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	chatID := message.Chat.ID
	text := "📊 Mening e'lonlarim / Мои объявления\n\n" +
		"Bu funksiya tez orada qo'shiladi / Эта функция скоро будет добавлена"

	lang := i18n.GetLanguage(teacher.Language)
	keyboard := MakeTeacherMainKeyboard(lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
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
	keyboard := MakeTeacherMainKeyboard(lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}
