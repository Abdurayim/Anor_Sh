package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
)

// HandleTeacherTakeAttendanceCommand allows teacher to mark attendance
func HandleTeacherTakeAttendanceCommand(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	chatID := message.Chat.ID

	text := "📋 <b>Yo'qlama olish / Отметить посещаемость</b>\n\n" +
		"Quyidagi formatda ma'lumotlarni yuboring:\n" +
		"Отправьте данные в следующем формате:\n\n" +
		"<code>O'quvchi ID / ID ученика\n" +
		"Status: + (keldi) yoki - (kelmadi)\n" +
		"Sana (YYYY-MM-DD)</code>\n\n" +
		"<b>Misol / Пример:</b>\n" +
		"<code>15\n" +
		"+\n" +
		"2025-12-02</code>\n\n" +
		"💡 + = keldi (present) / пришел\n" +
		"💡 - = kelmadi (absent) / не пришел\n" +
		"💡 O'quvchi ID larini ko'rish: /list_students"

	// Set state
	stateData := &models.StateData{}
	err := botService.StateManager.Set(message.From.ID, "awaiting_attendance_info", stateData)
	if err != nil {
		return err
	}

	return botService.TelegramService.SendMessage(chatID, text, nil)
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

		// Verify teacher is assigned to student's class
		isAssigned, err := botService.TeacherRepo.IsTeacherAssignedToClass(teacher.ID, student.ClassID)
		if err != nil || !isAssigned {
			text := "❌ Siz bu sinfga biriktirilmagansiz / Вы не назначены на этот класс"
			return botService.TelegramService.SendMessage(chatID, text, nil)
		}
	} else {
		// Check if admin
		admin, err := botService.AdminRepo.GetByTelegramID(telegramID)
		if err != nil || admin == nil {
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

// HandleViewAttendanceCommand allows parents to view attendance
func HandleViewAttendanceCommand(botService *services.BotService, message *tgbotapi.Message, user *models.User) error {
	chatID := message.Chat.ID

	// Check if user has selected a child
	if user.CurrentSelectedStudentID == nil {
		text := "❌ Avval farzandingizni tanlang: /my_children\n\n❌ Сначала выберите ребенка: /my_children"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get attendance records (last 30 days)
	records, err := botService.AttendanceService.GetAttendanceByStudentID(*user.CurrentSelectedStudentID, 30, 0)
	if err != nil {
		log.Printf("Failed to get attendance: %v", err)
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(records) == 0 {
		text := "📋 Hozircha yo'qlama ma'lumotlari yo'q.\n\n📋 Пока нет данных о посещаемости."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Format attendance
	text := fmt.Sprintf("📋 <b>Yo'qlama / Посещаемость</b>\n\n"+
		"O'quvchi / Ученик: <b>%s %s</b>\n"+
		"Sinf / Класс: <b>%s</b>\n\n",
		records[0].FirstName, records[0].LastName, records[0].ClassName)

	presentCount := 0
	absentCount := 0

	text += "<b>Oxirgi 30 kun / Последние 30 дней:</b>\n\n"
	for _, r := range records {
		dateStr := r.Date.Format("02.01.2006")
		if r.Status == "present" {
			text += fmt.Sprintf("✅ %s - <b>Keldi / Пришел</b>\n", dateStr)
			presentCount++
		} else {
			text += fmt.Sprintf("❌ %s - <b>Kelmadi / Не пришел</b>\n", dateStr)
			absentCount++
		}
	}

	text += fmt.Sprintf("\n📊 <b>Statistika / Статистика:</b>\n"+
		"Keldi / Пришел: <b>%d</b>\n"+
		"Kelmadi / Не пришел: <b>%d</b>",
		presentCount, absentCount)

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleTeacherViewClassAttendanceCommand allows teacher to view class attendance
func HandleTeacherViewClassAttendanceCommand(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	chatID := message.Chat.ID

	// Get teacher's classes
	classes, err := botService.TeacherRepo.GetTeacherClasses(teacher.ID)
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(classes) == 0 {
		text := "📚 Sizga hali sinflar biriktirilmagan.\n\n📚 Вам еще не назначены классы."
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
