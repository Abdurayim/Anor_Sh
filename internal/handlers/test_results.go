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

// HandleTeacherEnterGradesCommand allows teacher to enter test results
func HandleTeacherEnterGradesCommand(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	chatID := message.Chat.ID

	text := "📝 <b>Baho kiritish / Ввести оценку</b>\n\n" +
		"Quyidagi formatda ma'lumotlarni yuboring:\n" +
		"Отправьте данные в следующем формате:\n\n" +
		"<code>O'quvchi ID / ID ученика\n" +
		"Fan nomi / Название предмета\n" +
		"Baho / Оценка\n" +
		"Sana (YYYY-MM-DD)</code>\n\n" +
		"<b>Misol / Пример:</b>\n" +
		"<code>15\n" +
		"Matematika\n" +
		"5\n" +
		"2025-12-02</code>\n\n" +
		"💡 O'quvchi ID larini ko'rish uchun: /list_students\n" +
		"💡 Для просмотра ID учеников: /list_students"

	// Set state
	stateData := &models.StateData{}
	err := botService.StateManager.Set(message.From.ID, "awaiting_test_result_info", stateData)
	if err != nil {
		return err
	}

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleTestResultInfo processes test result input from teacher/admin
func HandleTestResultInfo(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Parse input
	lines := strings.Split(strings.TrimSpace(message.Text), "\n")
	if len(lines) < 4 {
		text := "❌ Noto'g'ri format. Iltimos, barcha ma'lumotlarni kiriting.\n\n" +
			"❌ Неверный формат. Пожалуйста, введите все данные."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	studentIDStr := strings.TrimSpace(lines[0])
	subjectName := strings.TrimSpace(lines[1])
	score := strings.TrimSpace(lines[2])
	testDate := strings.TrimSpace(lines[3])

	// Parse student ID
	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		text := "❌ O'quvchi ID noto'g'ri / Неверный ID ученика"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Verify student exists
	student, err := botService.StudentRepo.GetByID(studentID)
	if err != nil || student == nil {
		text := "❌ O'quvchi topilmadi / Ученик не найден"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Validate date format
	_, err = time.Parse("2006-01-02", testDate)
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

	// Create test result
	req := &models.CreateTestResultRequest{
		StudentID:   studentID,
		SubjectName: subjectName,
		Score:       score,
		TestDate:    testDate,
		TeacherID:   teacherID,
		AdminID:     adminID,
	}

	resultID, err := botService.TestResultService.CreateTestResult(req)
	if err != nil {
		log.Printf("Failed to create test result: %v", err)
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

	// Success message
	text := fmt.Sprintf(
		"✅ <b>Baho muvaffaqiyatli kiritildi!</b>\n\n"+
			"ID: <code>%d</code>\n"+
			"O'quvchi: <b>%s %s</b>\n"+
			"Sinf: <b>%s</b>\n"+
			"Fan: <b>%s</b>\n"+
			"Baho: <b>%s</b>\n"+
			"Sana: <b>%s</b>\n\n"+
			"✅ <b>Оценка успешно введена!</b>\n\n"+
			"ID: <code>%d</code>\n"+
			"Ученик: <b>%s %s</b>\n"+
			"Класс: <b>%s</b>\n"+
			"Предмет: <b>%s</b>\n"+
			"Оценка: <b>%s</b>\n"+
			"Дата: <b>%s</b>",
		resultID, student.FirstName, student.LastName, className, subjectName, score, testDate,
		resultID, student.FirstName, student.LastName, className, subjectName, score, testDate,
	)

	// Send to parent if they have telegram
	go notifyParentAboutGrade(botService, student.ID, subjectName, score, testDate)

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleViewTestResultsCommand allows parents to view test results
func HandleViewTestResultsCommand(botService *services.BotService, message *tgbotapi.Message, user *models.User) error {
	chatID := message.Chat.ID

	// Check if user has selected a child
	if user.CurrentSelectedStudentID == nil {
		text := "❌ Avval farzandingizni tanlang: /my_children\n\n❌ Сначала выберите ребенка: /my_children"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get test results
	results, err := botService.TestResultService.GetTestResultsByStudentID(*user.CurrentSelectedStudentID, 20, 0)
	if err != nil {
		log.Printf("Failed to get test results: %v", err)
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(results) == 0 {
		text := "📝 Hozircha baholar yo'q.\n\n📝 Пока нет оценок."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Format results by subject
	text := fmt.Sprintf("📝 <b>Baholar / Оценки</b>\n\n"+
		"O'quvchi / Ученик: <b>%s %s</b>\n"+
		"Sinf / Класс: <b>%s</b>\n\n",
		results[0].FirstName, results[0].LastName, results[0].ClassName)

	// Group by subject
	subjectMap := make(map[string][]*models.TestResultDetailed)
	for _, r := range results {
		subjectMap[r.SubjectName] = append(subjectMap[r.SubjectName], r)
	}

	for subject, scores := range subjectMap {
		text += fmt.Sprintf("📚 <b>%s</b>\n", subject)
		for _, s := range scores {
			dateStr := s.TestDate.Format("02.01.2006")
			text += fmt.Sprintf("   • %s - <b>%s</b>\n", dateStr, s.Score)
		}
		text += "\n"
	}

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleTeacherViewClassGradesCommand allows teacher to view class test results
func HandleTeacherViewClassGradesCommand(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
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
			fmt.Sprintf("view_grades_class_%d", class.ID),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{button})
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	text := "📚 <b>Sinf baholarini ko'rish / Просмотр оценок класса</b>\n\n" +
		"Sinfni tanlang / Выберите класс:"

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err = botService.Bot.Send(msg)
	return err
}

// HandleViewClassGradesCallback handles class selection for viewing grades
func HandleViewClassGradesCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, classID int) error {
	chatID := callback.Message.Chat.ID

	// Get test results for class
	results, err := botService.TestResultService.GetTestResultsByClassID(classID, 50, 0)
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(results) == 0 {
		text := "📝 Bu sinfda hali baholar yo'q.\n\n📝 В этом классе еще нет оценок."
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Format results
	text := fmt.Sprintf("📝 <b>Sinf baholari / Оценки класса</b>\n\n"+
		"Sinf / Класс: <b>%s</b>\n\n", results[0].ClassName)

	// Group by student
	studentMap := make(map[int][]*models.TestResultDetailed)
	for _, r := range results {
		studentMap[r.StudentID] = append(studentMap[r.StudentID], r)
	}

	for _, scores := range studentMap {
		if len(scores) > 0 {
			text += fmt.Sprintf("👤 <b>%s %s</b>\n", scores[0].FirstName, scores[0].LastName)
			for _, s := range scores {
				dateStr := s.TestDate.Format("02.01")
				text += fmt.Sprintf("   • %s: <b>%s</b> (%s)\n", s.SubjectName, s.Score, dateStr)
			}
			text += "\n"
		}
	}

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// notifyParentAboutGrade sends notification to parent about new grade
func notifyParentAboutGrade(botService *services.BotService, studentID int, subject, score, date string) {
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
				"🔔 <b>Yangi baho!</b>\n\n"+
					"O'quvchi: <b>%s %s</b>\n"+
					"Fan: <b>%s</b>\n"+
					"Baho: <b>%s</b>\n"+
					"Sana: <b>%s</b>",
				student.FirstName, student.LastName, subject, score, date)
		} else {
			text = fmt.Sprintf(
				"🔔 <b>Новая оценка!</b>\n\n"+
					"Ученик: <b>%s %s</b>\n"+
					"Предмет: <b>%s</b>\n"+
					"Оценка: <b>%s</b>\n"+
					"Дата: <b>%s</b>",
				student.FirstName, student.LastName, subject, score, date)
		}

		_ = botService.TelegramService.SendMessage(parent.TelegramID, text, nil)
	}
}
