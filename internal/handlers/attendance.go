package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
	"parent-bot/internal/utils"
)

// HandleTeacherTakeAttendanceCommand allows teacher to mark attendance by class
func HandleTeacherTakeAttendanceCommand(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	chatID := message.Chat.ID

	// Get all classes (teachers can access all classes)
	classes, err := botService.ClassRepo.GetAll(teacher.BranchID)
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(classes) == 0 {
		text := "📚 Hozircha sinflar yo'q. Admin sinf qo'shishi kerak.\n\n" +
			"📚 Пока нет классов. Администратор должен добавить классы."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Create inline keyboard for class selection
	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, class := range classes {
		button := tgbotapi.NewInlineKeyboardButtonData(
			class.ClassName,
			fmt.Sprintf("attendance_select_class_%d", class.ID),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	text := "📋 <b>Yo'qlama olish / Отметить посещаемость</b>\n\n" +
		"Qaysi sinf uchun yo'qlama olmoqchisiz?\n" +
		"Для какого класса хотите отметить посещаемость?"

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err = botService.Bot.Send(msg)
	return err
}

// HandleAttendanceClassSelection handles when teacher selects a class for attendance
func HandleAttendanceClassSelection(botService *services.BotService, callback *tgbotapi.CallbackQuery, classID int) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get teacher
	teacher, err := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if err != nil || teacher == nil {
		// Could also be admin
		admin := botService.GetAdmin(telegramID)
		if admin == nil {
			text := "❌ Ruxsat yo'q / Нет разрешения"
			_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
			return botService.TelegramService.SendMessage(chatID, text, nil)
		}
	}

	// Teachers can access all classes - no verification needed

	// Get students in this class
	students, err := botService.StudentRepo.GetByClassID(classID)
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(students) == 0 {
		text := "📝 Bu sinfda o'quvchilar yo'q.\n\n📝 В этом классе нет учеников."
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get class name
	class, _ := botService.ClassRepo.GetByID(classID)
	className := fmt.Sprintf("%d", classID)
	if class != nil {
		className = class.ClassName
	}

	// Get today's date in Uzbekistan time (Asia/Tashkent UTC+5)
	location, _ := time.LoadLocation("Asia/Tashkent")
	today := time.Now().In(location)
	todayStr := today.Format("2006-01-02")

	// Check if attendance already exists for today
	existingRecords, _ := botService.AttendanceService.GetAttendanceByClassIDAndDate(classID, todayStr)
	existingAbsentMap := make(map[int]bool) // studentID -> isAbsent
	for _, record := range existingRecords {
		if record.Status == "absent" {
			existingAbsentMap[record.StudentID] = true
		}
	}

	// Create inline keyboard with students - SIMPLIFIED: only "-" button for absent
	var buttons [][]tgbotapi.InlineKeyboardButton

	text := fmt.Sprintf("📋 <b>Yo'qlama / Посещаемость</b>\n\n"+
		"📚 Sinf / Класс: <b>%s</b>\n"+
		"📅 Sana / Дата: <b>%s</b>\n\n"+
		"<b>❌ tugmasini bosing - kelmaganlar uchun</b>\n"+
		"Qolganlar avtomatik ✅ deb belgilanadi.\n\n"+
		"<b>Нажмите ❌ для отсутствующих</b>\n"+
		"Остальные автоматически будут ✅\n\n",
		className, today.Format("02.01.2006"))

	for i, student := range students {
		// Check if already marked absent
		isAbsent := existingAbsentMap[student.ID]

		var buttonText string
		if isAbsent {
			buttonText = fmt.Sprintf("❌ %d. %s %s", i+1, student.FirstName, student.LastName)
		} else {
			buttonText = fmt.Sprintf("➖ %d. %s %s", i+1, student.FirstName, student.LastName)
		}

		button := tgbotapi.NewInlineKeyboardButtonData(
			buttonText,
			fmt.Sprintf("attendance_toggle_%d_%d", classID, student.ID),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	// Add finish button
	finishButton := tgbotapi.NewInlineKeyboardButtonData(
		"✅ Tugatish / Завершить",
		fmt.Sprintf("attendance_finish_%d", classID),
	)
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{finishButton})

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	// Store class ID in state for this session - initialize with existing absent students
	initialAbsentList := []int{}
	for studentID := range existingAbsentMap {
		initialAbsentList = append(initialAbsentList, studentID)
	}

	stateData := &models.StateData{
		ClassID:    &classID,
		AbsentList: initialAbsentList,
		Date:       todayStr,
	}
	err = botService.StateManager.Set(telegramID, "taking_attendance", stateData)
	if err != nil {
		log.Printf("Failed to set state: %v", err)
	}

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err = botService.Bot.Send(msg)
	return err
}

// HandleAttendanceInfo processes attendance input from teacher/admin
func HandleAttendanceInfo(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Parse input
	lines := strings.Split(strings.TrimSpace(message.Text), "\n")
	if len(lines) < 3 {
		text := "❌ Noto'g'ri format. Iltimos, barcha ma'lumotlarni kiriting.\n\n" +
			"❌ Неверный формат. Пожалуйста, введите все данные."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	studentIDStr := strings.TrimSpace(lines[0])
	statusStr := strings.TrimSpace(lines[1])
	dateStr := strings.TrimSpace(lines[2])

	// Parse student ID
	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		text := "❌ O'quvchi ID noto'g'ri / Неверный ID ученика"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Parse status
	var status string
	if statusStr == "+" {
		status = "present"
	} else if statusStr == "-" {
		status = "absent"
	} else {
		text := "❌ Status noto'g'ri. Faqat + yoki - ruxsat etilgan.\n\n" +
			"❌ Неверный статус. Разрешены только + или -."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Verify student exists
	student, err := botService.StudentRepo.GetByID(studentID)
	if err != nil || student == nil {
		text := "❌ O'quvchi topilmadi / Ученик не найден"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Validate date format
	_, err = time.Parse("2006-01-02", dateStr)
	if err != nil {
		text := "❌ Sana formati noto'g'ri. Format: YYYY-MM-DD\n\n" +
			"❌ Неверный формат даты. Формат: YYYY-MM-DD"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Check if user is teacher or admin
	teacher, _ := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	var adminID *int
	var teacherID *int

	if teacher != nil {
		teacherID = &teacher.ID
	} else {
		// Check if admin
		admin := botService.GetAdmin(telegramID)
		if admin == nil {
			text := "❌ Ruxsat yo'q / Нет разрешения"
			return botService.TelegramService.SendMessage(chatID, text, nil)
		}
		adminID = &admin.ID
	}

	// Create attendance record
	req := &models.CreateAttendanceRequest{
		StudentID:         studentID,
		Date:              dateStr,
		Status:            status,
		MarkedByTeacherID: teacherID,
		MarkedByAdminID:   adminID,
	}

	attendanceID, err := botService.AttendanceService.CreateAttendance(req)
	if err != nil {
		log.Printf("Failed to create attendance: %v", err)
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			text := "❌ Bu sana uchun allaqachon yo'qlama olingan / Посещаемость уже отмечена для этой даты"
			return botService.TelegramService.SendMessage(chatID, text, nil)
		}
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Get class name
	class, _ := botService.ClassRepo.GetByID(student.ClassID)
	className := fmt.Sprintf("%d", student.ClassID)
	if class != nil {
		className = class.ClassName
	}

	statusEmoji := "✅"
	statusText := "Keldi / Пришел"
	if status == "absent" {
		statusEmoji = "❌"
		statusText = "Kelmadi / Не пришел"
	}

	// Success message
	text := fmt.Sprintf(
		"%s <b>Yo'qlama qabul qilindi!</b>\n\n"+
			"ID: <code>%d</code>\n"+
			"O'quvchi: <b>%s %s</b>\n"+
			"Sinf: <b>%s</b>\n"+
			"Status: <b>%s</b>\n"+
			"Sana: <b>%s</b>\n\n"+
			"%s <b>Посещаемость отмечена!</b>\n\n"+
			"ID: <code>%d</code>\n"+
			"Ученик: <b>%s %s</b>\n"+
			"Класс: <b>%s</b>\n"+
			"Статус: <b>%s</b>\n"+
			"Дата: <b>%s</b>",
		statusEmoji, attendanceID, student.FirstName, student.LastName, className, statusText, dateStr,
		statusEmoji, attendanceID, student.FirstName, student.LastName, className, statusText, dateStr,
	)

	// Send notification to parent if absent
	if status == "absent" {
		go notifyParentAboutAbsence(botService, student.ID, dateStr)
	}

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleViewChildAttendanceCallback handles viewing a specific child's attendance
func HandleViewChildAttendanceCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Extract student ID from callback data (format: "view_child_attendance_123")
	parts := strings.Split(callback.Data, "_")
	if len(parts) != 4 {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	studentID, err := strconv.Atoi(parts[3])
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

	// Verify student belongs to this parent
	children, err := botService.StudentRepo.GetParentStudents(user.ID)
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	studentBelongsToParent := false
	for _, child := range children {
		if child.StudentID == studentID {
			studentBelongsToParent = true
			break
		}
	}

	if !studentBelongsToParent {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Bu farzand sizga tegishli emas")
		return nil
	}

	// Get attendance records (last 30 days)
	records, err := botService.AttendanceService.GetAttendanceByStudentID(studentID, 30, 0)
	if err != nil {
		log.Printf("Failed to get attendance: %v", err)
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(records) == 0 {
		text := "📋 Hozircha yo'qlama ma'lumotlari yo'q.\n\n📋 Пока нет данных о посещаемости."
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Format attendance
	text := fmt.Sprintf("📋 <b>Yo'qlama / Посещаемость</b>\n\n"+
		"👤 O'quvchi / Ученик: <b>%s %s</b>\n"+
		"📚 Sinf / Класс: <b>%s</b>\n\n",
		records[0].FirstName, records[0].LastName, records[0].ClassName)

	presentCount := 0
	absentCount := 0

	text += "<b>Oxirgi 30 kun / Последние 30 дней:</b>\n\n"
	for _, r := range records {
		dateStr := r.Date.Format("02.01.2006")
		if r.Status == "present" {
			text += fmt.Sprintf("<b>+</b> %s - Keldi / Пришел\n", dateStr)
			presentCount++
		} else {
			text += fmt.Sprintf("<b>-</b> %s - Kelmadi / Не пришел\n", dateStr)
			absentCount++
		}
	}

	text += fmt.Sprintf("\n📊 <b>Statistika / Статистика:</b>\n"+
		"<b>+</b> Keldi / Пришел: <b>%d</b>\n"+
		"<b>-</b> Kelmadi / Не пришел: <b>%d</b>",
		presentCount, absentCount)

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleTeacherViewClassAttendanceCommand allows teacher to view class attendance
func HandleTeacherViewClassAttendanceCommand(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	chatID := message.Chat.ID

	// Get all classes (teachers can access all classes)
	classes, err := botService.ClassRepo.GetAll(teacher.BranchID)
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(classes) == 0 {
		text := "📚 Hozircha sinflar yo'q. Admin sinf qo'shishi kerak.\n\n" +
			"📚 Пока нет классов. Администратор должен добавить классы."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Create inline keyboard for class selection
	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, class := range classes {
		button := tgbotapi.NewInlineKeyboardButtonData(
			class.ClassName,
			fmt.Sprintf("view_attendance_class_%d", class.ID),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	text := "📋 <b>Sinf yo'qlamasini ko'rish / Просмотр посещаемости класса</b>\n\n" +
		"Sinfni tanlang / Выберите класс:"

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err = botService.Bot.Send(msg)
	return err
}

// HandleViewClassAttendanceCallback handles class selection for viewing attendance
func HandleViewClassAttendanceCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, classID int) error {
	chatID := callback.Message.Chat.ID

	// Get today's date
	today := time.Now().Format("2006-01-02")

	// Get attendance for class today
	records, err := botService.AttendanceService.GetAttendanceByClassIDAndDate(classID, today)
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get class info
	class, _ := botService.ClassRepo.GetByID(classID)
	className := fmt.Sprintf("%d", classID)
	if class != nil {
		className = class.ClassName
	}

	if len(records) == 0 {
		text := fmt.Sprintf("📋 <b>%s</b> sinfi uchun bugungi yo'qlama olinmagan.\n\n"+
			"📋 Посещаемость для класса <b>%s</b> сегодня не отмечена.", className, className)
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Format results
	text := fmt.Sprintf("📋 <b>Yo'qlama / Посещаемость</b>\n\n"+
		"Sinf / Класс: <b>%s</b>\n"+
		"Sana / Дата: <b>%s</b>\n\n", className, time.Now().Format("02.01.2006"))

	presentCount := 0
	absentCount := 0

	for _, r := range records {
		if r.Status == "present" {
			text += fmt.Sprintf("✅ <b>%s %s</b>\n", r.FirstName, r.LastName)
			presentCount++
		} else {
			text += fmt.Sprintf("❌ <b>%s %s</b>\n", r.FirstName, r.LastName)
			absentCount++
		}
	}

	text += fmt.Sprintf("\n📊 Keldi: <b>%d</b> | Kelmadi: <b>%d</b>", presentCount, absentCount)

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleAttendanceToggle handles toggling a student's attendance status
func HandleAttendanceToggle(botService *services.BotService, callback *tgbotapi.CallbackQuery, classID, studentID int) error {
	telegramID := callback.From.ID

	// Get state data
	stateData, err := botService.StateManager.GetData(telegramID)
	if err != nil || stateData == nil {
		stateData = &models.StateData{
			ClassID:    &classID,
			AbsentList: []int{},
		}
	}

	// Toggle student in absent list
	found := false
	newAbsentList := []int{}
	for _, id := range stateData.AbsentList {
		if id == studentID {
			found = true
			// Remove from list (toggle off - student is present)
		} else {
			newAbsentList = append(newAbsentList, id)
		}
	}

	if !found {
		// Add to absent list (toggle on - student is absent)
		newAbsentList = append(stateData.AbsentList, studentID)
	}

	stateData.AbsentList = newAbsentList

	// Update state
	err = botService.StateManager.Set(telegramID, "taking_attendance", stateData)
	if err != nil {
		log.Printf("Failed to update state: %v", err)
	}

	// Re-render the attendance selection screen with updated state
	chatID := callback.Message.Chat.ID

	// Get students in this class
	students, err := botService.StudentRepo.GetByClassID(classID)
	if err != nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik")
		return nil
	}

	// Get class name
	class, _ := botService.ClassRepo.GetByID(classID)
	className := fmt.Sprintf("%d", classID)
	if class != nil {
		className = class.ClassName
	}

	// Get today's date in Uzbekistan time
	location, _ := time.LoadLocation("Asia/Tashkent")
	today := time.Now().In(location)

	// Create absent map for quick lookup
	absentMap := make(map[int]bool)
	for _, id := range stateData.AbsentList {
		absentMap[id] = true
	}

	// Create inline keyboard with students - SIMPLIFIED UI
	var buttons [][]tgbotapi.InlineKeyboardButton

	text := fmt.Sprintf("📋 <b>Yo'qlama / Посещаемость</b>\n\n"+
		"📚 Sinf / Класс: <b>%s</b>\n"+
		"📅 Sana / Дата: <b>%s</b>\n\n"+
		"<b>❌ tugmasini bosing - kelmaganlar uchun</b>\n"+
		"Qolganlar avtomatik ✅ deb belgilanadi.\n\n"+
		"<b>Нажмите ❌ для отсутствующих</b>\n"+
		"Остальные автоматически будут ✅\n\n",
		className, today.Format("02.01.2006"))

	for i, student := range students {
		var buttonText string
		if absentMap[student.ID] {
			buttonText = fmt.Sprintf("❌ %d. %s %s", i+1, student.FirstName, student.LastName)
		} else {
			buttonText = fmt.Sprintf("➖ %d. %s %s", i+1, student.FirstName, student.LastName)
		}

		button := tgbotapi.NewInlineKeyboardButtonData(
			buttonText,
			fmt.Sprintf("attendance_toggle_%d_%d", classID, student.ID),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	// Add finish button
	finishButton := tgbotapi.NewInlineKeyboardButtonData(
		"✅ Tugatish / Завершить",
		fmt.Sprintf("attendance_finish_%d", classID),
	)
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{finishButton})

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	// Update message
	editMsg := tgbotapi.NewEditMessageText(chatID, callback.Message.MessageID, text)
	editMsg.ParseMode = "HTML"
	editMsg.ReplyMarkup = &keyboard

	_, err = botService.Bot.Send(editMsg)
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
	return err
}

// HandleAttendanceFinish handles finishing attendance for a class
func HandleAttendanceFinish(botService *services.BotService, callback *tgbotapi.CallbackQuery, classID int) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get state data
	stateData, err := botService.StateManager.GetData(telegramID)
	if err != nil || stateData == nil {
		text := "❌ Xatolik: Sessiya tugagan. Iltimos, qaytadan boshlang.\n\n" +
			"❌ Ошибка: Сессия истекла. Пожалуйста, начните заново."
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get teacher or admin
	teacher, _ := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	admin := botService.GetAdmin(telegramID)

	var teacherID *int
	var adminID *int
	var markedByName string

	if teacher != nil {
		teacherID = &teacher.ID
		markedByName = teacher.FirstName + " " + teacher.LastName
	} else if admin != nil {
		adminID = &admin.ID
		markedByName = "Admin"
	} else {
		text := "❌ Ruxsat yo'q / Нет разрешения"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get all students in class
	students, err := botService.StudentRepo.GetByClassID(classID)
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get today's date in Uzbekistan time
	location, _ := time.LoadLocation("Asia/Tashkent")
	today := time.Now().In(location)
	todayStr := today.Format("2006-01-02")

	// Create absent map
	absentMap := make(map[int]bool)
	for _, id := range stateData.AbsentList {
		absentMap[id] = true
	}

	// Create attendance records for all students
	successCount := 0
	errorCount := 0
	updatedCount := 0
	absentStudentNames := []string{}

	for _, student := range students {
		status := "present"
		if absentMap[student.ID] {
			status = "absent"
			absentStudentNames = append(absentStudentNames, student.FirstName+" "+student.LastName)
		}

		req := &models.CreateAttendanceRequest{
			StudentID:         student.ID,
			Date:              todayStr,
			Status:            status,
			MarkedByTeacherID: teacherID,
			MarkedByAdminID:   adminID,
		}

		_, err := botService.AttendanceService.CreateAttendance(req)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint") {
				// Already exists for today, skip
				log.Printf("Attendance already exists for student %d on %s, skipping", student.ID, todayStr)
				updatedCount++
			} else {
				errorCount++
				log.Printf("Failed to create attendance for student %d: %v", student.ID, err)
			}
		} else {
			successCount++
			// Send notification if absent
			if status == "absent" {
				go notifyParentAboutAbsence(botService, student.ID, todayStr)
			}
		}
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Get class name
	class, _ := botService.ClassRepo.GetByID(classID)
	className := fmt.Sprintf("%d", classID)
	if class != nil {
		className = class.ClassName
	}

	presentCount := len(students) - len(stateData.AbsentList)
	absentCount := len(stateData.AbsentList)

	// Success message
	text := fmt.Sprintf(
		"✅ <b>Yo'qlama saqlandi!</b>\n\n"+
			"📚 Sinf: <b>%s</b>\n"+
			"📅 Sana: <b>%s</b>\n\n"+
			"✅ Keldi: <b>%d</b>\n"+
			"❌ Kelmadi: <b>%d</b>\n\n"+
			"✅ <b>Посещаемость сохранена!</b>",
		className, today.Format("02.01.2006"),
		presentCount, absentCount,
	)

	// Delete the selection message
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, callback.Message.MessageID)
	_, _ = botService.Bot.Request(deleteMsg)

	// Return appropriate keyboard based on who finished attendance
	var keyboard interface{}
	if teacher != nil {
		lang := i18n.GetLanguage(teacher.Language)
		keyboard = utils.MakeTeacherMainMenuKeyboard(lang)
	} else if admin != nil {
		// Get admin's user record for language
		user, _ := botService.UserService.GetUserByTelegramID(telegramID)
		lang := i18n.LanguageUzbek
		if user != nil {
			lang = i18n.GetLanguage(user.Language)
		}
		keyboard = utils.MakeMainMenuKeyboardWithAdmin(lang)
	}

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅ Saqlandi!")

	// Send to teacher/admin first
	err = botService.TelegramService.SendMessage(chatID, text, keyboard)

	// Send notification to ALL admins about attendance
	go notifyAdminsAboutAttendance(botService, classID, className, todayStr, markedByName, presentCount, absentCount, absentStudentNames)

	return err
}

// notifyAdminsAboutAttendance notifies the admins of the class's branch about completed attendance
func notifyAdminsAboutAttendance(botService *services.BotService, classID int, className, date, markedBy string, presentCount, absentCount int, absentStudentNames []string) {
	class, err := botService.ClassRepo.GetByID(classID)
	if err != nil || class == nil {
		log.Printf("Failed to get class %d for attendance notification: %v", classID, err)
		return
	}
	admins, err := botService.AdminRepo.GetAll(class.BranchID)
	if err != nil {
		log.Printf("Failed to get admins for attendance notification: %v", err)
		return
	}

	// Build notification text
	text := fmt.Sprintf(
		"📋 <b>Yo'qlama olingan / Посещаемость отмечена</b>\n\n"+
			"📚 Sinf / Класс: <b>%s</b>\n"+
			"📅 Sana / Дата: <b>%s</b>\n"+
			"👤 Kim oldi / Отметил: <b>%s</b>\n\n"+
			"✅ Keldi / Пришли: <b>%d</b>\n"+
			"❌ Kelmadi / Отсутствуют: <b>%d</b>",
		className, date, markedBy, presentCount, absentCount,
	)

	// Add absent student names if any
	if len(absentStudentNames) > 0 {
		text += "\n\n<b>Kelmaganlar / Отсутствующие:</b>\n"
		for i, name := range absentStudentNames {
			text += fmt.Sprintf("%d. %s\n", i+1, name)
		}
	}

	// Send to all admins
	for _, admin := range admins {
		if admin.TelegramID == nil || *admin.TelegramID == 0 {
			continue
		}

		_ = botService.TelegramService.SendMessage(*admin.TelegramID, text, nil)
	}
}

// notifyParentAboutAbsence sends notification to parent about absence
func notifyParentAboutAbsence(botService *services.BotService, studentID int, date string) {
	// Get parents linked to this student
	parents, err := botService.StudentRepo.GetStudentParents(studentID)
	if err != nil {
		log.Printf("Failed to get parents for notification: %v", err)
		return
	}

	// Get student info
	student, err := botService.StudentRepo.GetByID(studentID)
	if err != nil {
		return
	}

	for _, parent := range parents {
		if parent.TelegramID == 0 {
			continue
		}

		var text string
		if parent.Language == "uz" {
			text = fmt.Sprintf(
				"⚠️ <b>Yo'qlama haqida xabar!</b>\n\n"+
					"O'quvchi: <b>%s %s</b>\n"+
					"Sana: <b>%s</b>\n\n"+
					"❌ Darsga kelmadi / Не пришел на занятие",
				student.FirstName, student.LastName, date)
		} else {
			text = fmt.Sprintf(
				"⚠️ <b>Уведомление о посещаемости!</b>\n\n"+
					"Ученик: <b>%s %s</b>\n"+
					"Дата: <b>%s</b>\n\n"+
					"❌ Не пришел на занятие",
				student.FirstName, student.LastName, date)
		}

		_ = botService.TelegramService.SendMessage(parent.TelegramID, text, nil)
	}
}
