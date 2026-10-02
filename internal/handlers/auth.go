package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
)

// noBranch never matches a real branch, so branch-filtered queries return nothing.
// (0 means "all branches" in repositories and must never come from a user.)
const noBranch = -1

const msgNotAllowed = "❌ Ruxsat yo'q / Нет доступа"

// adminBranchID returns the branch of the admin linked to this Telegram account,
// or noBranch if the account is not an admin.
func adminBranchID(botService *services.BotService, telegramID int64) int {
	if admin := botService.GetAdmin(telegramID); admin != nil && admin.BranchID > 0 {
		return admin.BranchID
	}
	return noBranch
}

// teacherBranchID returns the branch of the teacher linked to this Telegram account,
// or noBranch if the account is not an active teacher.
func teacherBranchID(botService *services.BotService, telegramID int64) int {
	teacher, _ := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if teacher != nil && teacher.BranchID > 0 {
		return teacher.BranchID
	}
	return noBranch
}

// staffBranchID returns the branch of an admin or teacher account (admin wins), or noBranch.
func staffBranchID(botService *services.BotService, telegramID int64) int {
	if b := adminBranchID(botService, telegramID); b != noBranch {
		return b
	}
	return teacherBranchID(botService, telegramID)
}

// parentOf returns the registered parent for this Telegram account, or nil.
func parentOf(botService *services.BotService, telegramID int64) *models.User {
	user, _ := botService.UserService.GetUserByTelegramID(telegramID)
	return user
}

// parentBranchID returns the branch of the parent, or noBranch.
func parentBranchID(botService *services.BotService, telegramID int64) int {
	if user := parentOf(botService, telegramID); user != nil && user.BranchID > 0 {
		return user.BranchID
	}
	return noBranch
}

// adminCommands are slash commands only admins may use.
var adminCommands = map[string]bool{
	"admin": true, "manage_classes": true, "add_class": true, "delete_class": true, "toggle_class": true,
	"upload_timetable": true, "post_announcement": true, "add_student": true, "link_student": true,
	"list_students": true, "view_parent_children": true, "add_teacher": true, "list_teachers": true,
}

// authorizeCommand checks role-restricted slash commands before they are dispatched.
func authorizeCommand(botService *services.BotService, message *tgbotapi.Message) bool {
	command := message.Command()
	if command == "superadmin" {
		return botService.GetSuperAdmin(message.From.ID) != nil
	}
	if adminCommands[command] {
		return botService.GetAdmin(message.From.ID) != nil
	}
	if command == "edit_grade" || command == "delete_grade" {
		return staffBranchID(botService, message.From.ID) != noBranch
	}
	return true
}

// scan1 parses "<prefix>%d" style callback data.
func scan1(data, format string) (int, bool) {
	var id int
	n, err := fmt.Sscanf(data, format, &id)
	return id, err == nil && n == 1
}

// scan2 parses "<prefix>%d_%d" style callback data.
func scan2(data, format string) (int, int, bool) {
	var a, b int
	n, err := fmt.Sscanf(data, format, &a, &b)
	return a, b, err == nil && n == 2
}

// authorizeCallback is the single permission gate for inline-button callbacks. Callback data
// can be replayed from old messages or forged by a modified client, so every role-specific
// action is checked here against the caller's role and branch before any handler runs.
func authorizeCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) bool {
	data := callback.Data
	tgID := callback.From.ID
	has := func(prefix string) bool { return strings.HasPrefix(data, prefix) }

	switch {
	// ---------------- super admin ----------------
	case has("sa_"):
		return botService.GetSuperAdmin(tgID) != nil

	// ---------------- admin-only ----------------
	case has("admin_"), has("class_delete_"), has("class_toggle_"), has("timetable_delete_"),
		has("timetable_select_"), has("announcement_edit_"), has("announcement_delete_"),
		data == "announcement_skip_file", has("export_grades_"):
		branchID := adminBranchID(botService, tgID)
		if branchID == noBranch {
			return false
		}
		return adminObjectInBranch(botService, data, branchID)

	// ---------------- teacher (or admin) ----------------
	case has("teacher_"), has("attendance_"), has("test_result_"), has("view_grades_class_"),
		has("view_attendance_class_"):
		branchID := staffBranchID(botService, tgID)
		if branchID == noBranch {
			return false
		}
		return staffObjectInBranch(botService, data, branchID)

	// ---------------- parent: own children only ----------------
	case has("view_child_attendance_"), has("view_child_grades_"), has("view_child_"),
		has("complaint_select_child_"), has("proposal_select_child_"), has("timetable_child_"):
		parent := parentOf(botService, tgID)
		if parent == nil {
			return false
		}
		idx := strings.LastIndex(data, "_")
		var studentID int
		if _, err := fmt.Sscanf(data[idx+1:], "%d", &studentID); err != nil {
			return false
		}
		linked, err := botService.StudentRepo.IsStudentLinkedToParent(parent.ID, studentID)
		return err == nil && linked

	// ---------------- parent: linking children of own branch ----------------
	case has("select_class_"), has("mykids_class_"):
		if id, ok := scan1(data[strings.LastIndex(data, "_")+1:], "%d"); ok {
			return botService.ClassInBranch(id, registrationBranchID(botService, tgID))
		}
		return false
	case has("select_student_"), has("mykids_student_"):
		if id, ok := scan1(data[strings.LastIndex(data, "_")+1:], "%d"); ok {
			return botService.StudentInBranch(id, registrationBranchID(botService, tgID))
		}
		return false
	}
	return true
}

// registrationBranchID is the parent's branch, or the branch picked during registration
// (stored in state) for a parent who is not created yet.
func registrationBranchID(botService *services.BotService, telegramID int64) int {
	if b := parentBranchID(botService, telegramID); b != noBranch {
		return b
	}
	if data, err := botService.StateManager.GetData(telegramID); err == nil && data != nil && data.BranchID > 0 {
		return data.BranchID
	}
	return noBranch
}

// adminObjectInBranch checks that the class/student/teacher/timetable/announcement
// referenced by an admin callback belongs to the admin's branch.
func adminObjectInBranch(botService *services.BotService, data string, branchID int) bool {
	if classID, _, ok := scan2(data, "admin_delete_student_%d_%d"); ok {
		return botService.ClassInBranch(classID, branchID)
	}
	for _, format := range []string{"admin_view_class_%d", "admin_add_student_%d", "class_delete_confirm_%d", "class_delete_%d",
		"class_toggle_%d", "export_grades_select_%d", "export_grades_custom_%d", "export_grades_%d", "timetable_select_%d"} {
		if classID, ok := scan1(data, format); ok {
			return botService.ClassInBranch(classID, branchID)
		}
	}
	if teacherID, ok := scan1(data, "admin_delete_teacher_confirm_%d"); ok {
		teacher, err := botService.TeacherRepo.GetByID(teacherID)
		return err == nil && teacher.BranchID == branchID
	}
	if teacherID, ok := scan1(data, "admin_delete_teacher_%d"); ok {
		teacher, err := botService.TeacherRepo.GetByID(teacherID)
		return err == nil && teacher.BranchID == branchID
	}
	if timetableID, ok := scan1(data, "timetable_delete_%d"); ok {
		timetable, err := botService.TimetableService.GetTimetableByID(timetableID)
		return err == nil && timetable != nil && botService.ClassInBranch(timetable.ClassID, branchID)
	}
	for _, format := range []string{"announcement_edit_%d", "announcement_delete_%d"} {
		if id, ok := scan1(data, format); ok {
			a, err := botService.AnnouncementService.GetAnnouncementByID(id)
			return err == nil && a != nil && a.BranchID == branchID
		}
	}
	return true
}

// staffObjectInBranch checks that the class/student referenced by a teacher callback
// belongs to the caller's branch.
func staffObjectInBranch(botService *services.BotService, data string, branchID int) bool {
	for _, format := range []string{"teacher_delete_student_%d_%d", "attendance_toggle_%d_%d"} {
		if classID, studentID, ok := scan2(data, format); ok {
			return botService.ClassInBranch(classID, branchID) && botService.StudentInBranch(studentID, branchID)
		}
	}
	for _, format := range []string{"teacher_manage_class_%d", "teacher_add_student_%d",
		"teacher_announcement_toggle_class_%d", "attendance_select_class_%d", "attendance_finish_%d",
		"view_grades_class_%d", "view_attendance_class_%d", "test_result_select_class_%d"} {
		if classID, ok := scan1(data, format); ok {
			return botService.ClassInBranch(classID, branchID)
		}
	}
	if studentID, ok := scan1(data, "test_result_add_student_%d"); ok {
		return botService.StudentInBranch(studentID, branchID)
	}
	for _, format := range []string{"teacher_announcement_edit_%d", "teacher_announcement_delete_%d"} {
		if id, ok := scan1(data, format); ok {
			a, err := botService.AnnouncementService.GetAnnouncementByID(id)
			return err == nil && a != nil && a.BranchID == branchID
		}
	}
	return true
}
