package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
)

// HandleAddStudentCommand initiates adding a new student (admin only)
func HandleAddStudentCommand(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Check if user is admin
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	var lang i18n.Language = i18n.LanguageUzbek
	if user != nil {
		lang = i18n.GetLanguage(user.Language)
	}

	isAdmin := false
	if user != nil {
		isAdmin, _ = botService.IsAdmin(user.PhoneNumber, user.TelegramID)
	}

	if !isAdmin {
		text := "❌ Bu buyruq faqat administratorlar uchun / Эта команда только для администраторов"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get active classes
	classes, err := botService.ClassRepo.GetActive()
	if err != nil {
		text := i18n.Get(i18n.ErrDatabaseError, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(classes) == 0 {
		text := "❌ Hozircha mavjud sinflar yo'q. Avval sinf qo'shing.\n\n" +
			"❌ Пока нет доступных классов. Сначала добавьте класс."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	text := "👤 <b>Yangi o'quvchi qo'shish / Добавить нового ученика</b>\n\n" +
		"Iltimos, quyidagi formatda ma'lumotlarni yuboring:\n" +
		"Пожалуйста, отправьте данные в следующем формате:\n\n" +
		"<code>Ism Familiya / Имя Фамилия\n" +
		"Sinf / Класс</code>\n\n" +
		"<b>Misol / Пример:</b>\n" +
		"<code>Aziz Karimov\n" +
		"5-A</code>"

	// Set state
	stateData := &models.StateData{}
	err = botService.StateManager.Set(telegramID, "awaiting_student_info", stateData)
	if err != nil {
		return err
	}

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleStudentInfo processes student information input from admin
func HandleStudentInfo(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Parse input
	lines := strings.Split(strings.TrimSpace(message.Text), "\n")
	if len(lines) < 2 {
		text := "❌ Noto'g'ri format. Iltimos, ism-familiya va sinfni alohida qatorlarda yuboring.\n\n" +
			"❌ Неверный формат. Пожалуйста, отправьте имя-фамилию и класс на отдельных строках."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	fullName := strings.TrimSpace(lines[0])
	className := strings.TrimSpace(lines[1])

	// Parse name into first and last name
	nameParts := strings.Fields(fullName)
	if len(nameParts) < 2 {
		text := "❌ Iltimos, ism va familiyani kiriting.\n\n" +
			"❌ Пожалуйста, введите имя и фамилию."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	firstName := nameParts[0]
	lastName := strings.Join(nameParts[1:], " ")

	// Verify class exists
	class, err := botService.ClassRepo.GetByName(className)
	if err != nil {
		return err
	}

	if class == nil {
		text := fmt.Sprintf("❌ '%s' sinfi topilmadi. Avval sinfni qo'shing.\n\n"+
			"❌ Класс '%s' не найден. Сначала добавьте класс.", className, className)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get admin info
	user, err := botService.UserService.GetUserByTelegramID(telegramID)
	if err != nil {
		return err
	}

	admin, err := botService.AdminRepo.GetByTelegramID(telegramID)
	if err != nil || admin == nil {
		text := "❌ Admin ma'lumotlari topilmadi / Данные администратора не найдены"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Create student
	studentReq := &models.CreateStudentRequest{
		FirstName:      firstName,
		LastName:       lastName,
		ClassID:        class.ID,
		AddedByAdminID: &admin.ID,
	}
	studentID, err := botService.StudentRepo.Create(studentReq)
	if err != nil {
		log.Printf("Failed to create student: %v", err)
		studentLang := i18n.LanguageUzbek
		if user != nil {
			studentLang = i18n.GetLanguage(user.Language)
		}
		text := i18n.Get(i18n.ErrDatabaseError, studentLang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Success message
	text := fmt.Sprintf(
		"✅ <b>O'quvchi muvaffaqiyatli qo'shildi!</b>\n\n"+
			"ID: <code>%d</code>\n"+
			"Ism: <b>%s %s</b>\n"+
			"Sinf: <b>%s</b>\n\n"+
			"✅ <b>Ученик успешно добавлен!</b>\n\n"+
			"ID: <code>%d</code>\n"+
			"Имя: <b>%s %s</b>\n"+
			"Класс: <b>%s</b>",
		studentID, firstName, lastName, className,
		studentID, firstName, lastName, className,
	)

	// Create keyboard with "Add More" button
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"➕ Yana o'quvchi qo'shish / Добавить ещё",
				fmt.Sprintf("admin_add_student_%d", class.ID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"◀️ Sinfga qaytish / Назад к классу",
				fmt.Sprintf("admin_view_class_%d", class.ID),
			),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err = botService.Bot.Send(msg)
	return err
}

// HandleAdminStudentNameInput handles student name input when admin adds student to a specific class
func HandleAdminStudentNameInput(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Check if classID is set
	if stateData.ClassID == nil {
		text := "❌ Xatolik: sinf ma'lumoti topilmadi / Ошибка: информация о классе не найдена"
		_ = botService.StateManager.Clear(telegramID)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	classID := *stateData.ClassID

	// Parse full name
	fullName := strings.TrimSpace(message.Text)
	nameParts := strings.Fields(fullName)
	if len(nameParts) < 2 {
		text := "❌ Iltimos, ism va familiyani kiriting.\n\n" +
			"❌ Пожалуйста, введите имя и фамилию.\n\n" +
			"<b>Misol / Пример:</b> Jasur Rahimov"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	firstName := nameParts[0]
	lastName := strings.Join(nameParts[1:], " ")

	// Get class info
	class, err := botService.ClassRepo.GetByID(classID)
	if err != nil || class == nil {
		text := "❌ Sinf topilmadi / Класс не найден"
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

	// Create student
	studentReq := &models.CreateStudentRequest{
		FirstName:      firstName,
		LastName:       lastName,
		ClassID:        classID,
		AddedByAdminID: &admin.ID,
	}
	studentID, err := botService.StudentRepo.Create(studentReq)
	if err != nil {
		log.Printf("Failed to create student: %v", err)
		text := "❌ Xatolik / Ошибка: " + err.Error()
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Success message
	text := fmt.Sprintf(
		"✅ <b>O'quvchi muvaffaqiyatly qo'shildi!</b>\n\n"+
			"📌 ID: <code>%d</code>\n"+
			"👤 Ism-Familiya: <b>%s %s</b>\n"+
			"📚 Sinf: <b>%s</b>\n\n"+
			"✅ <b>Ученик успешно добавлен!</b>\n\n"+
			"📌 ID: <code>%d</code>\n"+
			"👤 Имя-Фамилия: <b>%s %s</b>\n"+
			"📚 Класс: <b>%s</b>",
		studentID, firstName, lastName, class.ClassName,
		studentID, firstName, lastName, class.ClassName,
	)

	// Create keyboard with "Add More" and "Back" buttons
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"➕ Yana o'quvchi qo'shish / Добавить ещё",
				fmt.Sprintf("admin_add_student_%d", classID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"◀️ Orqaga / Назад",
				fmt.Sprintf("admin_view_class_%d", classID),
			),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard
	_, err = botService.Bot.Send(msg)
	return err
}

// HandleLinkStudentCommand initiates linking a student to a parent (admin only)
func HandleLinkStudentCommand(botService *services.BotService, message *tgbotapi.Message) error {
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

	text := "🔗 <b>O'quvchini ota-onaga bog'lash / Привязать ученика к родителю</b>\n\n" +
		"Iltimos, quyidagi formatda ma'lumotlarni yuboring:\n" +
		"Пожалуйста, отправьте данные в следующем формате:\n\n" +
		"<code>Ota-ona telefoni / Телефон родителя\n" +
		"O'quvchi ID / ID ученика</code>\n\n" +
		"<b>Misol / Пример:</b>\n" +
		"<code>+998901234567\n" +
		"15</code>\n\n" +
		"O'quvchi ID raqamini bilish uchun /list_students buyrug'ini ishlating\n" +
		"Для получения ID ученика используйте команду /list_students"

	// Set state
	stateData := &models.StateData{}
	err = botService.StateManager.Set(telegramID, "awaiting_link_info", stateData)
	if err != nil {
		return err
	}

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleLinkInfo processes linking information input from admin
func HandleLinkInfo(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Parse input
	lines := strings.Split(strings.TrimSpace(message.Text), "\n")
	if len(lines) < 2 {
		text := "❌ Noto'g'ri format. Iltimos, telefon raqami va o'quvchi ID sini yuboring.\n\n" +
			"❌ Неверный формат. Пожалуйста, отправьте номер телефона и ID ученика."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	phoneNumber := strings.TrimSpace(lines[0])
	studentIDStr := strings.TrimSpace(lines[1])

	// Validate phone number format
	if !strings.HasPrefix(phoneNumber, "+998") || len(phoneNumber) != 13 {
		text := "❌ Telefon raqami noto'g'ri formatda. Format: +998XXXXXXXXX\n\n" +
			"❌ Неверный формат номера телефона. Формат: +998XXXXXXXXX"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Parse student ID
	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		text := "❌ O'quvchi ID noto'g'ri / Неверный ID ученика"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Find parent by phone
	parent, err := botService.UserRepo.GetByPhone(phoneNumber)
	if err != nil {
		log.Printf("Error finding parent: %v", err)
		text := "❌ Xatolik yuz berdi / Произошла ошибка"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if parent == nil {
		text := fmt.Sprintf("❌ '%s' raqami bilan ro'yxatdan o'tgan ota-ona topilmadi.\n\n"+
			"❌ Родитель с номером '%s' не найден в системе.", phoneNumber, phoneNumber)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Verify student exists
	student, err := botService.StudentRepo.GetByID(studentID)
	if err != nil || student == nil {
		text := fmt.Sprintf("❌ ID %d bo'lgan o'quvchi topilmadi / Ученик с ID %d не найден", studentID, studentID)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Check if already linked
	existingLinks, err := botService.StudentRepo.GetParentStudents(parent.ID)
	if err != nil {
		log.Printf("Error checking existing links: %v", err)
		text := "❌ Xatolik yuz berdi / Произошла ошибка"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Check if student already linked to this parent
	for _, linked := range existingLinks {
		if linked.StudentID == studentID {
			text := "⚠️ Bu o'quvchi allaqachon ushbu ota-onaga bog'langan.\n\n" +
				"⚠️ Этот ученик уже привязан к этому родителю."
			return botService.TelegramService.SendMessage(chatID, text, nil)
		}
	}

	// Check max children limit (5)
	if len(existingLinks) >= 5 {
		text := "❌ Bir ota-ona maksimal 5 ta farzandni bog'lash mumkin.\n\n" +
			"❌ Родитель может привязать максимум 5 детей."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Create link
	err = botService.StudentRepo.LinkToParent(parent.ID, studentID)
	if err != nil {
		log.Printf("Failed to link student to parent: %v", err)
		// Check if it's a UNIQUE constraint violation (student already linked to another parent)
		if strings.Contains(err.Error(), "UNIQUE") {
			text := "❌ Bu o'quvchi allaqachon boshqa ota-onaga bog'langan!\n" +
				"Bir o'quvchi faqat BITTA ota-onaga tegishli bo'lishi mumkin.\n\n" +
				"❌ Этот ученик уже привязан к другому родителю!\n" +
				"Один ученик может принадлежать только ОДНОМУ родителю."
			return botService.TelegramService.SendMessage(chatID, text, nil)
		}
		text := "❌ Bog'lashda xatolik / Ошибка при привязке"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Deprecated: No longer setting current selected student ID
	// Multi-child system uses callback-based selection

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Get class info for display
	class, _ := botService.ClassRepo.GetByID(student.ClassID)
	className := "N/A"
	if class != nil {
		className = class.ClassName
	}

	// Success message
	text := fmt.Sprintf(
		"✅ <b>Muvaffaqiyatli bog'landi!</b>\n\n"+
			"Ota-ona: <code>%s</code>\n"+
			"O'quvchi: <b>%s %s</b> (ID: %d)\n"+
			"Sinf: <b>%s</b>\n\n"+
			"✅ <b>Успешно привязано!</b>\n\n"+
			"Родитель: <code>%s</code>\n"+
			"Ученик: <b>%s %s</b> (ID: %d)\n"+
			"Класс: <b>%s</b>",
		phoneNumber, student.FirstName, student.LastName, student.ID, className,
		phoneNumber, student.FirstName, student.LastName, student.ID, className,
	)

	// Notify parent if they're online
	if parent.TelegramID != 0 {
		parentMsg := fmt.Sprintf(
			"👨‍👩‍👧‍👦 <b>Yangi farzand bog'landi!</b>\n\n"+
				"Ism: <b>%s %s</b>\n"+
				"Sinf: <b>%s</b>\n\n"+
				"👨‍👩‍👧‍👦 <b>Привязан новый ребенок!</b>\n\n"+
				"Имя: <b>%s %s</b>\n"+
				"Класс: <b>%s</b>",
			student.FirstName, student.LastName, className,
			student.FirstName, student.LastName, className,
		)
		_ = botService.TelegramService.SendMessage(parent.TelegramID, parentMsg, nil)
	}

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleListStudentsCommand lists all students (admin only)
func HandleListStudentsCommand(botService *services.BotService, message *tgbotapi.Message) error {
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

	// Get all students
	students, err := botService.StudentRepo.GetAll(100, 0)
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(students) == 0 {
		text := "📝 Hozircha o'quvchilar yo'q.\n\n📝 Пока нет учеников."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Format student list
	text := fmt.Sprintf("👥 <b>O'quvchilar ro'yxati / Список учеников</b>\n\n"+
		"Jami: <b>%d</b> ta\n\n", len(students))

	for i, s := range students {
		status := "✅"
		if !s.IsActive {
			status = "❌"
		}
		text += fmt.Sprintf("%d. %s <b>%s %s</b> (ID: <code>%d</code>)\n   Sinf/Класс: <b>%s</b>\n\n",
			i+1, status, s.FirstName, s.LastName, s.ID, s.ClassName)
	}

	text += "\n💡 ID raqamidan foydalanib o'quvchini ota-onaga bog'lashingiz mumkin"

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleViewParentChildrenCommand shows parent's children links (admin only)
func HandleViewParentChildrenCommand(botService *services.BotService, message *tgbotapi.Message) error {
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

	text := "🔍 <b>Ota-ona farzandlarini ko'rish / Просмотр детей родителя</b>\n\n" +
		"Iltimos, ota-onaning telefon raqamini yuboring:\n" +
		"Пожалуйста, отправьте номер телефона родителя:\n\n" +
		"Format: <code>+998XXXXXXXXX</code>"

	// Set state
	stateData := &models.StateData{}
	err = botService.StateManager.Set(telegramID, "awaiting_parent_phone_for_view", stateData)
	if err != nil {
		return err
	}

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleParentPhoneForView processes parent phone to view their children
func HandleParentPhoneForView(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	phoneNumber := strings.TrimSpace(message.Text)

	// Validate phone number format
	if !strings.HasPrefix(phoneNumber, "+998") || len(phoneNumber) != 13 {
		text := "❌ Telefon raqami noto'g'ri formatda. Format: +998XXXXXXXXX\n\n" +
			"❌ Неверный формат номера телефона. Формат: +998XXXXXXXXX"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Find parent
	parent, err := botService.UserRepo.GetByPhone(phoneNumber)
	if err != nil {
		log.Printf("Error finding parent: %v", err)
		text := "❌ Xatolik yuz berdi / Произошла ошибка"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if parent == nil {
		text := fmt.Sprintf("❌ '%s' raqami bilan ro'yxatdan o'tgan ota-ona topilmadi.\n\n"+
			"❌ Родитель с номером '%s' не найден в системе.", phoneNumber, phoneNumber)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get parent's children
	children, err := botService.StudentRepo.GetParentStudents(parent.ID)
	if err != nil {
		log.Printf("Error getting students: %v", err)
		text := "❌ Xatolik yuz berdi / Произошла ошибка"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	if len(children) == 0 {
		text := fmt.Sprintf("📝 '%s' raqamli ota-onaga farzandlar bog'lanmagan.\n\n"+
			"📝 К родителю '%s' не привязаны дети.", phoneNumber, phoneNumber)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Format children list
	text := fmt.Sprintf(
		"👨‍👩‍👧‍👦 <b>Ota-ona farzandlari / Дети родителя</b>\n\n"+
			"Telefon / Телефон: <code>%s</code>\n"+
			"Farzandlar / Дети: <b>%d</b> ta\n\n",
		phoneNumber, len(children),
	)

	for i, child := range children {
		text += fmt.Sprintf("%d. <b>%s %s</b>\n   ID: <code>%d</code> | Sinf/Класс: <b>%s</b>\n\n",
			i+1, child.StudentFirstName, child.StudentLastName, child.StudentID, child.ClassName)
	}

	return botService.TelegramService.SendMessage(chatID, text, nil)
}
