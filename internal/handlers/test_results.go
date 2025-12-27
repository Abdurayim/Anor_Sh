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
	telegramID := message.From.ID

	// Clear any existing state
	_ = botService.StateManager.Clear(telegramID)

	// Get all classes (teachers can access all classes)
	classes, err := botService.ClassRepo.GetAll()
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
	text := "📊 <b>Baho kiritish / Ввести оценку</b>\n\n" +
		"Sinfni tanlang:\n" +
		"Выберите класс:"

	for _, class := range classes {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("📚 %s", class.ClassName),
			fmt.Sprintf("test_result_select_class_%d", class.ID),
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

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	_, err = botService.Bot.Send(msg)
	return err
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

// HandleViewChildGradesCallback handles viewing a specific child's grades
func HandleViewChildGradesCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Extract student ID from callback data (format: "view_child_grades_123")
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

	// Get test results
	results, err := botService.TestResultService.GetTestResultsByStudentID(studentID, 50, 0)
	if err != nil {
		log.Printf("Failed to get test results: %v", err)
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(results) == 0 {
		text := "📝 Hozircha baholar yo'q.\n\n📝 Пока нет оценок."
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Format results by subject
	text := fmt.Sprintf("📊 <b>Baholar / Оценки</b>\n\n"+
		"👤 O'quvchi / Ученик: <b>%s %s</b>\n"+
		"📚 Sinf / Класс: <b>%s</b>\n\n",
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

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleTeacherViewClassGradesCommand allows teacher to view class test results
func HandleTeacherViewClassGradesCommand(botService *services.BotService, message *tgbotapi.Message, teacher *models.Teacher) error {
	chatID := message.Chat.ID

	// Get all classes (teachers can access all classes)
	classes, err := botService.ClassRepo.GetAll()
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка bazы danных"
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

	// Check if user is teacher to show edit buttons
	teacher, _ := botService.TeacherService.GetTeacherByTelegramID(callback.From.ID)
	isTeacher := teacher != nil

	// Format results
	text := fmt.Sprintf("📝 <b>Sinf baholari / Оценки класса</b>\n\n"+
		"Sinf / Класс: <b>%s</b>\n\n", results[0].ClassName)

	// Group by student
	studentMap := make(map[int][]*models.TestResultDetailed)
	for _, r := range results {
		studentMap[r.StudentID] = append(studentMap[r.StudentID], r)
	}

	// For teachers, show results with edit/delete buttons
	if isTeacher {
		for _, scores := range studentMap {
			if len(scores) > 0 {
				text += fmt.Sprintf("👤 <b>%s %s</b>\n", scores[0].FirstName, scores[0].LastName)
				for _, s := range scores {
					dateStr := s.TestDate.Format("02.01")
					text += fmt.Sprintf("   • %s: <b>%s</b> (%s) - ID: %d\n", s.SubjectName, s.Score, dateStr, s.ID)
				}
				text += "\n"
			}
		}

		text += "\n💡 Bahoni o'zgartirish uchun: /edit_grade <ID> <yangi_baho>\n"
		text += "💡 Bahoni o'chirish uchun: /delete_grade <ID>\n\n"
		text += "💡 Для изменения оценки: /edit_grade <ID> <новая_оценка>\n"
		text += "💡 Для удаления оценки: /delete_grade <ID>"

		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// For non-teachers, show simple view
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

// HandleEditGradeCommand handles the /edit_grade command for teachers
func HandleEditGradeCommand(botService *services.BotService, message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	// Check if user is teacher
	teacher, err := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if err != nil || teacher == nil {
		text := "❌ Faqat o'qituvchilar baholarni o'zgartira oladi.\n\n" +
			"❌ Только учителя могут изменять оценки."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Parse command: /edit_grade <ID> <new_score>
	parts := strings.Fields(message.Text)
	if len(parts) < 3 {
		text := "❌ Noto'g'ri format.\n\n" +
			"Format: /edit_grade <ID> <yangi_baho>\n\n" +
			"Misol: /edit_grade 123 5\n\n" +
			"Формат: /edit_grade <ID> <новая_оценка>\n\n" +
			"Пример: /edit_grade 123 5"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Parse test result ID
	testResultID, err := strconv.Atoi(parts[1])
	if err != nil {
		text := "❌ Noto'g'ri ID format / Неверный формат ID"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	newScore := parts[2]

	// Get test result
	testResult, err := botService.TestResultService.GetTestResultByID(testResultID)
	if err != nil || testResult == nil {
		text := "❌ Baho topilmadi / Оценка не найдена"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Update the score
	updateReq := &models.UpdateTestResultRequest{
		Score: newScore,
	}
	err = botService.TestResultService.UpdateTestResult(testResultID, updateReq)
	if err != nil {
		log.Printf("Failed to update test result: %v", err)
		text := "❌ Bahoni yangilashda xatolik / Ошибка при обновлении оценки"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	text := fmt.Sprintf(
		"✅ Baho muvaffaqiyatli yangilandi!\n\n"+
			"ID: %d\n"+
			"Yangi baho / Новая оценка: <b>%s</b>\n\n"+
			"✅ Оценка успешно обновлена!",
		testResultID, newScore,
	)

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleDeleteGradeCommand handles the /delete_grade command for teachers
func HandleDeleteGradeCommand(botService *services.BotService, message *tgbotapi.Message) error {
	chatID := message.Chat.ID
	telegramID := message.From.ID

	// Check if user is teacher
	teacher, err := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if err != nil || teacher == nil {
		text := "❌ Faqat o'qituvchilar baholarni o'chira oladi.\n\n" +
			"❌ Только учителя могут удалять оценки."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Parse command: /delete_grade <ID>
	parts := strings.Fields(message.Text)
	if len(parts) < 2 {
		text := "❌ Noto'g'ri format.\n\n" +
			"Format: /delete_grade <ID>\n\n" +
			"Misol: /delete_grade 123\n\n" +
			"Формат: /delete_grade <ID>\n\n" +
			"Пример: /delete_grade 123"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Parse test result ID
	testResultID, err := strconv.Atoi(parts[1])
	if err != nil {
		text := "❌ Noto'g'ri ID format / Неверный формат ID"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get test result to show what's being deleted
	testResult, err := botService.TestResultService.GetTestResultByID(testResultID)
	if err != nil || testResult == nil {
		text := "❌ Baho topilmadi / Оценка не найдена"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Delete the test result
	err = botService.TestResultService.DeleteTestResult(testResultID)
	if err != nil {
		log.Printf("Failed to delete test result: %v", err)
		text := "❌ Bahoni o'chirishda xatolik / Ошибка при удалении оценки"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	text := fmt.Sprintf(
		"✅ Baho muvaffaqiyatli o'chirildi!\n\n"+
			"ID: %d\n\n"+
			"✅ Оценка успешно удалена!",
		testResultID,
	)

	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleTestResultClassSelectionCallback handles class selection for adding test results
func HandleTestResultClassSelectionCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, classID int) error {
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

	// Get students in this class
	students, err := botService.StudentRepo.GetByClassID(classID)
	if err != nil {
		text := "❌ Ma'lumotlar bazasida xatolik / Ошибка базы данных"
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	if len(students) == 0 {
		text := "📝 Bu sinfda hali o'quvchilar yo'q.\n\n📝 В этом классе пока нет учеников."
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Format text with students list
	text := fmt.Sprintf("📊 <b>Baho kiritish / Ввести оценку</b>\n\n"+
		"Sinf / Класс: <b>%s</b>\n"+
		"O'quvchilar soni / Количество учеников: <b>%d</b>\n\n"+
		"O'quvchini tanlang:\n"+
		"Выберите ученика:", class.ClassName, len(students))

	// Create buttons for each student
	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, student := range students {
		studentBtn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("➕ %s %s", student.FirstName, student.LastName),
			fmt.Sprintf("test_result_add_student_%d", student.ID),
		)
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{studentBtn})
	}

	// Add back button
	backButton := tgbotapi.NewInlineKeyboardButtonData(
		"⬅️ Orqaga / Назад",
		"test_result_back_to_classes",
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
	return err
}

// HandleTestResultAddStudentCallback handles student selection for adding test results
func HandleTestResultAddStudentCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, studentID int) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Get teacher
	teacher, err := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if err != nil || teacher == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	// Get student info
	student, err := botService.StudentRepo.GetByID(studentID)
	if err != nil || student == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ O'quvchi topilmadi")
		return nil
	}

	// Get class info
	class, err := botService.ClassRepo.GetByID(student.ClassID)
	if err != nil || class == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Sinf topilmadi")
		return nil
	}

	// Set state with student ID
	stateData := &models.StateData{
		SelectedStudentID: &studentID,
	}
	err = botService.StateManager.Set(telegramID, "teacher_awaiting_test_results_text", stateData)
	if err != nil {
		log.Printf("Failed to set state: %v", err)
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	// Send instructions
	text := fmt.Sprintf(
		"📊 <b>Baho kiritish / Ввести оценку</b>\n\n"+
			"O'quvchi / Ученик: <b>%s %s</b>\n"+
			"Sinf / Класс: <b>%s</b>\n\n"+
			"Iltimos, baholarni quyidagi formatda yuboring:\n"+
			"Пожалуйста, отправьте оценки в следующем формате:\n\n"+
			"<code>fan:baho, fan:baho, ...</code>\n\n"+
			"<b>Misol / Пример:</b>\n"+
			"<code>matematika:90%%, ona tili:85%%, ingliz tili:75%%</code>\n\n"+
			"yoki / или\n\n"+
			"<code>matematika:5, ona tili:4, ingliz tili:5</code>\n\n"+
			"💡 Siz bir nechta fanni bir vaqtning o'zida kirita olasiz\n"+
			"💡 Вы можете ввести несколько предметов одновременно",
		student.FirstName, student.LastName, class.ClassName,
	)

	// Create cancel button
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"❌ Bekor qilish / Отмена",
				fmt.Sprintf("test_result_select_class_%d", student.ClassID),
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

// HandleTeacherTestResultsTextInput handles free-form test results text input
func HandleTeacherTestResultsTextInput(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Get teacher
	teacher, err := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if err != nil || teacher == nil {
		text := "❌ Xatolik / Ошибка"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Validate state data
	if stateData.SelectedStudentID == nil {
		text := "❌ Sessiya tugagan. Iltimos, qaytadan boshlang.\n\n" +
			"❌ Сессия истекла. Пожалуйста, начните заново."
		_ = botService.StateManager.Clear(telegramID)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	studentID := *stateData.SelectedStudentID

	// Get student info
	student, err := botService.StudentRepo.GetByID(studentID)
	if err != nil || student == nil {
		text := "❌ O'quvchi topilmadi / Ученик не найден"
		_ = botService.StateManager.Clear(telegramID)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Parse the free-form text
	inputText := strings.TrimSpace(message.Text)
	testResults := parseTestResultsText(inputText)

	if len(testResults) == 0 {
		text := "❌ Noto'g'ri format. Iltimos, formatni tekshiring:\n\n" +
			"❌ Неверный формат. Пожалуйста, проверьте формат:\n\n" +
			"<code>fan:baho, fan:baho, ...</code>\n\n" +
			"<b>Misol / Пример:</b>\n" +
			"<code>matematika:90%, ona tili:85%</code>"
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get today's date for test results
	today := time.Now().Format("2006-01-02")

	// Create test results for each subject
	var createdResults []struct {
		Subject string
		Score   string
		ID      int64
	}

	for subject, score := range testResults {
		req := &models.CreateTestResultRequest{
			StudentID:   studentID,
			SubjectName: subject,
			Score:       score,
			TestDate:    today,
			TeacherID:   &teacher.ID,
			AdminID:     nil,
		}

		resultID, err := botService.TestResultService.CreateTestResult(req)
		if err != nil {
			log.Printf("Failed to create test result for %s: %v", subject, err)
			continue
		}

		createdResults = append(createdResults, struct {
			Subject string
			Score   string
			ID      int64
		}{Subject: subject, Score: score, ID: resultID})

		// Notify parent about each grade
		go notifyParentAboutGrade(botService, studentID, subject, score, today)
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	if len(createdResults) == 0 {
		text := "❌ Baholarni kiritishda xatolik yuz berdi.\n\n" +
			"❌ Произошла ошибка при вводе оценок."
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Get class info
	class, _ := botService.ClassRepo.GetByID(student.ClassID)
	className := fmt.Sprintf("%d", student.ClassID)
	if class != nil {
		className = class.ClassName
	}

	// Build success message
	text := fmt.Sprintf(
		"✅ <b>Baholar muvaffaqiyatli kiritildi!</b>\n\n"+
			"O'quvchi / Ученик: <b>%s %s</b>\n"+
			"Sinf / Класс: <b>%s</b>\n"+
			"Sana / Дата: <b>%s</b>\n\n"+
			"Kiritilgan baholar / Введенные оценки:\n",
		student.FirstName, student.LastName, className, today,
	)

	for _, result := range createdResults {
		text += fmt.Sprintf("  📚 <b>%s:</b> %s\n", result.Subject, result.Score)
	}

	text += fmt.Sprintf("\n<i>Jami / Всего: %d ta fan</i>\n\n"+
		"✅ <b>Оценки успешно введены!</b>", len(createdResults))

	// Create keyboard with options
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"➕ Yana baho qo'shish / Добавить ещё оценку",
				fmt.Sprintf("test_result_add_student_%d", studentID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"◀️ Sinfga qaytish / Назад к классу",
				fmt.Sprintf("test_result_select_class_%d", student.ClassID),
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

	_, err = botService.Bot.Send(msg)
	return err
}

// parseTestResultsText parses free-form test results text
// Examples:
//   - "matematika:90%, ona tili:85%, ingliz tili:75%"
//   - "matematika:5, ona tili:4, ingliz tili:5"
//   - "math:90%, literature:85%"
func parseTestResultsText(text string) map[string]string {
	results := make(map[string]string)

	// Split by comma or semicolon
	entries := strings.FieldsFunc(text, func(r rune) bool {
		return r == ',' || r == ';'
	})

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		// Split by colon
		parts := strings.SplitN(entry, ":", 2)
		if len(parts) != 2 {
			continue
		}

		subject := strings.TrimSpace(parts[0])
		score := strings.TrimSpace(parts[1])

		if subject == "" || score == "" {
			continue
		}

		results[subject] = score
	}

	return results
}

// HandleTestResultBackToClassesCallback handles back button to classes list
func HandleTestResultBackToClassesCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	// Get teacher
	teacher, err := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if err != nil || teacher == nil {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		return nil
	}

	// Create a fake message to reuse the command handler
	fakeMessage := &tgbotapi.Message{
		From: callback.From,
		Chat: &tgbotapi.Chat{
			ID: callback.Message.Chat.ID,
		},
	}

	// Delete previous message
	deleteMsg := tgbotapi.NewDeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
	_, _ = botService.Bot.Request(deleteMsg)

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	return HandleTeacherEnterGradesCommand(botService, fakeMessage, teacher)
}
