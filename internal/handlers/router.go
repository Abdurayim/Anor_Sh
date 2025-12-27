package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
	"parent-bot/internal/utils"
)

// RouteByState routes messages based on user's current state
func RouteByState(botService *services.BotService, message *tgbotapi.Message, state string, stateData *models.StateData) error {
	switch state {
	case models.StateAwaitingLanguage:
		// Waiting for language selection (handled by callback)
		return nil

	case models.StateAwaitingPhone:
		return HandlePhoneNumber(botService, message, stateData)

	case models.StateAwaitingChildName:
		return HandleChildName(botService, message, stateData)

	case models.StateAwaitingChildClass:
		return HandleChildClass(botService, message, stateData)

	case models.StateAwaitingComplaint:
		return HandleComplaintText(botService, message, stateData)

	case models.StateConfirmingComplaint:
		// Waiting for confirmation (handled by callback)
		return nil

	case models.StateAwaitingAdminPhone:
		return HandleAdminLinkPhone(botService, message)

	case models.StateAwaitingClassName:
		return HandleClassNameInput(botService, message)

	case models.StateAwaitingProposal:
		return HandleProposalText(botService, message, stateData)

	case models.StateConfirmingProposal:
		// Waiting for confirmation (handled by callback)
		return nil

	case models.StateAwaitingTimetableFile:
		return HandleTimetableFileUpload(botService, message, stateData)

	case models.StateAwaitingAnnouncementContent:
		return HandleAnnouncementContent(botService, message, stateData)

	case models.StateAwaitingAnnouncementFile:
		return HandleAnnouncementFile(botService, message, stateData)

	case models.StateAwaitingEditedAnnouncementContent:
		return HandleEditedAnnouncementContent(botService, message, stateData)

	case "awaiting_student_info":
		return HandleStudentInfo(botService, message, stateData)

	case "awaiting_admin_student_name":
		return HandleAdminStudentNameInput(botService, message, stateData)

	case "awaiting_link_info":
		return HandleLinkInfo(botService, message, stateData)

	case "awaiting_parent_phone_for_view":
		return HandleParentPhoneForView(botService, message, stateData)

	case "awaiting_teacher_full_name":
		return HandleTeacherFullName(botService, message, stateData)

	case "awaiting_teacher_phone":
		return HandleTeacherPhone(botService, message, stateData)

	case "awaiting_test_result_info":
		return HandleTestResultInfo(botService, message, stateData)

	case "awaiting_attendance_info":
		return HandleAttendanceInfo(botService, message, stateData)

	case "teacher_selecting_announcement_classes":
		// Waiting for callback selection
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

	case "admin_awaiting_export_custom_dates":
		return HandleAdminExportCustomDatesInput(botService, message, stateData)

	case "selecting_child_for_complaint":
		// Waiting for callback selection
		return nil

	case "selecting_child_for_proposal":
		// Waiting for callback selection
		return nil

	case models.StateRegistered:
		// User is registered, get user data
		user, err := botService.UserService.GetUserByTelegramID(message.From.ID)
		if err != nil {
			return err
		}
		return HandleRegisteredUserMessage(botService, message, user)

	default:
		// Unknown state, restart
		return HandleStart(botService, message)
	}
}

// HandleRegisteredUserMessage handles messages from registered users
func HandleRegisteredUserMessage(botService *services.BotService, message *tgbotapi.Message, user *models.User) error {
	// Check if message is a button press
	buttonText := message.Text

	// Admin panel button (check both languages) - check this BEFORE checking if user is nil
	// because admin might not be registered as a parent
	if buttonText == "👨‍💼 Ma'muriyat paneli" || buttonText == "👨‍💼 Панель администратора" {
		return HandleAdminCommand(botService, message)
	}

	if user == nil {
		return HandleStart(botService, message)
	}

	lang := i18n.GetLanguage(user.Language)
	chatID := message.Chat.ID

	// Submit complaint button (check both languages)
	if buttonText == i18n.Get(i18n.BtnSubmitComplaint, i18n.LanguageUzbek) ||
	   buttonText == i18n.Get(i18n.BtnSubmitComplaint, i18n.LanguageRussian) {
		return HandleComplaintCommand(botService, message)
	}

	// My complaints button (check both languages)
	if buttonText == i18n.Get(i18n.BtnMyComplaints, i18n.LanguageUzbek) ||
	   buttonText == i18n.Get(i18n.BtnMyComplaints, i18n.LanguageRussian) {
		return HandleMyComplaintsCommand(botService, message)
	}

	// Submit proposal button (check both languages)
	if buttonText == i18n.Get(i18n.BtnSubmitProposal, i18n.LanguageUzbek) ||
	   buttonText == i18n.Get(i18n.BtnSubmitProposal, i18n.LanguageRussian) {
		return HandleProposalCommand(botService, message)
	}

	// View timetable button (check both languages)
	if buttonText == i18n.Get(i18n.BtnViewTimetable, i18n.LanguageUzbek) ||
	   buttonText == i18n.Get(i18n.BtnViewTimetable, i18n.LanguageRussian) {
		return HandleViewTimetableCommand(botService, message)
	}

	// View announcements button (check both languages)
	if buttonText == i18n.Get(i18n.BtnViewAnnouncements, i18n.LanguageUzbek) ||
	   buttonText == i18n.Get(i18n.BtnViewAnnouncements, i18n.LanguageRussian) {
		return HandleViewAnnouncementsCommand(botService, message)
	}

	// Settings button (check both languages)
	if buttonText == i18n.Get(i18n.BtnSettings, i18n.LanguageUzbek) ||
	   buttonText == i18n.Get(i18n.BtnSettings, i18n.LanguageRussian) {
		return HandleSettingsCommand(botService, message)
	}

	// My children button (check both languages)
	if buttonText == i18n.Get(i18n.BtnMyChildren, i18n.LanguageUzbek) ||
	   buttonText == i18n.Get(i18n.BtnMyChildren, i18n.LanguageRussian) {
		return HandleMyChildrenCommand(botService, message)
	}

	// Test results button (check both languages) - redirect to My Children for child selection
	if buttonText == i18n.Get(i18n.BtnMyTestResults, i18n.LanguageUzbek) ||
	   buttonText == i18n.Get(i18n.BtnMyTestResults, i18n.LanguageRussian) {
		return HandleMyChildrenCommand(botService, message)
	}

	// Attendance button (check both languages) - redirect to My Children for child selection
	if buttonText == i18n.Get(i18n.BtnMyAttendance, i18n.LanguageUzbek) ||
	   buttonText == i18n.Get(i18n.BtnMyAttendance, i18n.LanguageRussian) {
		return HandleMyChildrenCommand(botService, message)
	}

	// Default: show main menu
	text := i18n.Get(i18n.MsgMainMenu, lang)

	// Check if user is admin to show appropriate keyboard
	isAdmin, _ := botService.IsAdmin(user.PhoneNumber, user.TelegramID)
	keyboard := utils.MakeMainMenuKeyboardForUser(lang, isAdmin)

	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleCallbackQuery handles inline button clicks
func HandleCallbackQuery(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	data := callback.Data

	// Language selection
	if data == "lang_uz" || data == "lang_ru" {
		return HandleLanguageSelection(botService, callback)
	}

	// Parent registration: class selection
	if len(data) > 13 && data[:13] == "select_class_" {
		return HandleSelectClassCallback(botService, callback)
	}

	// Parent registration: student selection
	if len(data) > 15 && data[:15] == "select_student_" {
		return HandleSelectStudentCallback(botService, callback)
	}

	// Parent registration: skip child selection
	if data == "skip_child_selection" {
		return HandleSkipChildSelectionCallback(botService, callback)
	}

	// Parent registration: back to class selection
	if data == "back_to_class_selection" {
		return HandleBackToClassSelectionCallback(botService, callback)
	}

	// Parent My Kids: add another child
	if data == "add_another_child" {
		return HandleAddAnotherChildCallback(botService, callback)
	}

	// Parent My Kids: class selection
	if len(data) > 13 && data[:13] == "mykids_class_" {
		return HandleMyKidsClassCallback(botService, callback)
	}

	// Parent My Kids: student selection
	if len(data) > 15 && data[:15] == "mykids_student_" {
		return HandleMyKidsStudentCallback(botService, callback)
	}

	// Parent My Kids: back to my kids menu
	if data == "back_to_my_kids" {
		return HandleBackToMyKidsCallback(botService, callback)
	}

	// Parent My Kids: back to class selection
	if data == "back_to_mykids_class_selection" {
		return HandleBackToMyKidsClassSelectionCallback(botService, callback)
	}

	// Parent: back to main menu
	if data == "back_to_main" {
		return HandleBackToMainCallback(botService, callback)
	}

	// Parent: view child attendance (MUST be before generic view_child_ check)
	if strings.HasPrefix(data, "view_child_attendance_") {
		return HandleViewChildAttendanceCallback(botService, callback)
	}

	// Parent: view child test results/grades (MUST be before generic view_child_ check)
	if strings.HasPrefix(data, "view_child_grades_") {
		return HandleViewChildGradesCallback(botService, callback)
	}

	// Parent: view child info (generic - must be AFTER specific view_child_ checks)
	if strings.HasPrefix(data, "view_child_") {
		return HandleViewChildCallback(botService, callback)
	}

	// Parent: show my kids (after adding child)
	if data == "show_my_kids" {
		return HandleShowMyKidsCallback(botService, callback)
	}

	// Child selection (for parents with multiple children) - DEPRECATED, use view_child_
	if len(data) > 13 && data[:13] == "select_child_" {
		return HandleChildInfoCallback(botService, callback)
	}

	// Class delete callback (MUST be checked BEFORE generic "class_" check)
	if len(data) > 13 && data[:13] == "class_delete_" {
		fmt.Printf("[ROUTER DEBUG] Routing to HandleClassDeleteCallback, data: %s\n", data)
		return HandleClassDeleteCallback(botService, callback)
	}

	// Class info callback (MUST be checked BEFORE generic "class_" check)
	if len(data) > 11 && data[:11] == "class_info_" {
		_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")
		return nil
	}

	// Class selection for registration (generic "class_" - must be AFTER specific class_ checks)
	if len(data) > 6 && data[:6] == "class_" {
		return HandleClassSelection(botService, callback)
	}

	// Complaint child selection callback
	if strings.HasPrefix(data, "complaint_select_child_") {
		var studentID int
		fmt.Sscanf(data, "complaint_select_child_%d", &studentID)
		return HandleComplaintSelectChildCallback(botService, callback, studentID)
	}

	// Complaint confirmation
	if data == "confirm_complaint" {
		return HandleComplaintConfirmation(botService, callback)
	}

	// Complaint cancellation
	if data == "cancel_complaint" {
		return HandleComplaintCancellation(botService, callback)
	}

	// Proposal child selection callback
	if strings.HasPrefix(data, "proposal_select_child_") {
		var studentID int
		fmt.Sscanf(data, "proposal_select_child_%d", &studentID)
		return HandleProposalSelectChildCallback(botService, callback, studentID)
	}

	// Proposal confirmation
	if data == "confirm_proposal" {
		return HandleProposalConfirmation(botService, callback)
	}

	// Proposal cancellation
	if data == "cancel_proposal" {
		return HandleProposalCancellation(botService, callback)
	}

	// Timetable child selection (for parents viewing child's timetable)
	if strings.HasPrefix(data, "timetable_child_") {
		var studentID int
		fmt.Sscanf(data, "timetable_child_%d", &studentID)
		return HandleTimetableChildSelection(botService, callback, studentID)
	}

	// Timetable class selection (for admin uploading)
	if strings.HasPrefix(data, "timetable_select_") {
		var classID int
		fmt.Sscanf(data, "timetable_select_%d", &classID)
		return HandleTimetableClassSelection(botService, callback, classID)
	}

	// Announcement skip file
	if data == "announcement_skip_file" {
		stateData, err := botService.StateManager.GetData(callback.From.ID)
		if err != nil {
			return err
		}
		return HandleAnnouncementSkipFile(botService, callback, stateData)
	}

	// Announcement edit callback
	if len(data) > 18 && data[:18] == "announcement_edit_" {
		var announcementID int
		fmt.Sscanf(data, "announcement_edit_%d", &announcementID)
		return HandleAnnouncementEditCallback(botService, callback, announcementID)
	}

	// Announcement delete callback
	if len(data) > 20 && data[:20] == "announcement_delete_" {
		var announcementID int
		fmt.Sscanf(data, "announcement_delete_%d", &announcementID)
		return HandleAnnouncementDeleteCallback(botService, callback, announcementID)
	}

	// Admin callbacks
	if data == "admin_users" {
		return HandleAdminUsersCallback(botService, callback)
	}

	if data == "admin_complaints" {
		return HandleAdminComplaintsCallback(botService, callback)
	}

	if data == "admin_stats" {
		return HandleAdminStatsCallback(botService, callback)
	}

	// Admin manage classes callback
	if data == "admin_manage_classes" {
		return HandleAdminManageClassesCallback(botService, callback)
	}

	// Admin view class callback
	if len(data) > 17 && data[:17] == "admin_view_class_" {
		var classID int
		fmt.Sscanf(data, "admin_view_class_%d", &classID)
		return HandleAdminViewClassCallback(botService, callback, classID)
	}

	// Admin add student callback
	if len(data) > 18 && data[:18] == "admin_add_student_" {
		var classID int
		fmt.Sscanf(data, "admin_add_student_%d", &classID)
		return HandleAdminAddStudentCallback(botService, callback, classID)
	}

	// Admin delete student callback
	if len(data) > 21 && data[:21] == "admin_delete_student_" {
		var classID, studentID int
		fmt.Sscanf(data, "admin_delete_student_%d_%d", &classID, &studentID)
		return HandleAdminDeleteStudentCallback(botService, callback, classID, studentID)
	}

	// Admin create class callback
	if data == "admin_create_class" {
		return HandleAdminCreateClassCallback(botService, callback)
	}

	// Admin upload timetable callback
	if data == "admin_upload_timetable" {
		return HandleAdminUploadTimetableCallback(botService, callback)
	}

	// Admin post announcement callback
	if data == "admin_post_announcement" {
		return HandleAdminPostAnnouncementCallback(botService, callback)
	}

	// Admin view proposals callback
	if data == "admin_proposals" {
		return HandleAdminProposalsCallback(botService, callback)
	}

	// Admin view timetables callback
	if data == "admin_view_timetables" {
		return HandleAdminViewTimetablesCallback(botService, callback)
	}

	// Admin view announcements callback
	if data == "admin_view_announcements" {
		return HandleAdminViewAnnouncementsCallback(botService, callback)
	}

	// Admin manage teachers callback
	if data == "admin_manage_teachers" {
		return HandleAdminManageTeachersCallback(botService, callback)
	}

	// Admin add teacher callback
	if data == "admin_add_teacher" {
		return HandleAdminAddTeacherCallback(botService, callback)
	}

	// Admin delete teacher callback
	if strings.HasPrefix(data, "admin_delete_teacher_") {
		var teacherID int
		fmt.Sscanf(data, "admin_delete_teacher_%d", &teacherID)
		return HandleAdminDeleteTeacherCallback(botService, callback, teacherID)
	}

	// Admin export attendance callback
	if data == "admin_export_attendance" {
		return HandleAdminExportAttendanceCallback(botService, callback)
	}

	// Admin export test results callback
	if data == "admin_export_test_results" {
		return HandleAdminExportTestResultsCallback(botService, callback)
	}

	// Timetable delete callback
	if len(data) > 17 && data[:17] == "timetable_delete_" {
		return HandleTimetableDeleteCallback(botService, callback)
	}

	// Admin back button
	if data == "admin_back" {
		return HandleAdminBackCallback(botService, callback)
	}

	// Teacher manage class callback
	if len(data) > 21 && data[:21] == "teacher_manage_class_" {
		var classID int
		fmt.Sscanf(data, "teacher_manage_class_%d", &classID)
		return HandleTeacherManageClassCallback(botService, callback, classID)
	}

	// Teacher add student callback
	if len(data) > 20 && data[:20] == "teacher_add_student_" {
		var classID int
		fmt.Sscanf(data, "teacher_add_student_%d", &classID)
		return HandleTeacherAddStudentCallback(botService, callback, classID)
	}

	// Teacher delete student callback
	if len(data) > 23 && data[:23] == "teacher_delete_student_" {
		var classID, studentID int
		fmt.Sscanf(data, "teacher_delete_student_%d_%d", &classID, &studentID)
		return HandleTeacherDeleteStudentCallback(botService, callback, classID, studentID)
	}

	// Teacher manage students back callback
	if data == "teacher_manage_students_back" {
		return HandleTeacherManageStudentsBackCallback(botService, callback)
	}

	// Teacher back to main callback
	if data == "teacher_back_to_main" {
		return HandleTeacherBackToMainCallback(botService, callback)
	}

	// Teacher announcement toggle class callback
	if len(data) > 34 && data[:34] == "teacher_announcement_toggle_class_" {
		var classID int
		fmt.Sscanf(data, "teacher_announcement_toggle_class_%d", &classID)
		return HandleTeacherAnnouncementToggleClass(botService, callback, classID)
	}

	// Teacher announcement continue callback
	if data == "teacher_announcement_continue" {
		return HandleTeacherAnnouncementContinue(botService, callback)
	}

	// Teacher announcement cancel callback
	if data == "teacher_announcement_cancel" {
		return HandleTeacherAnnouncementCancel(botService, callback)
	}

	// Teacher announcement skip file callback
	if data == "teacher_announcement_skip_file" {
		return HandleTeacherAnnouncementSkipFile(botService, callback)
	}

	// Teacher announcement edit callback
	if len(data) > 26 && data[:26] == "teacher_announcement_edit_" {
		var announcementID int
		fmt.Sscanf(data, "teacher_announcement_edit_%d", &announcementID)
		return HandleTeacherAnnouncementEdit(botService, callback, announcementID)
	}

	// Teacher announcement delete callback
	if len(data) > 28 && data[:28] == "teacher_announcement_delete_" {
		var announcementID int
		fmt.Sscanf(data, "teacher_announcement_delete_%d", &announcementID)
		return HandleTeacherAnnouncementDelete(botService, callback, announcementID)
	}

	// Attendance select class callback (for taking attendance)
	if strings.HasPrefix(data, "attendance_select_class_") {
		var classID int
		fmt.Sscanf(data, "attendance_select_class_%d", &classID)
		return HandleAttendanceClassSelection(botService, callback, classID)
	}

	// Attendance toggle student callback
	if strings.HasPrefix(data, "attendance_toggle_") {
		var classID, studentID int
		fmt.Sscanf(data, "attendance_toggle_%d_%d", &classID, &studentID)
		return HandleAttendanceToggle(botService, callback, classID, studentID)
	}

	// Attendance finish callback
	if strings.HasPrefix(data, "attendance_finish_") {
		var classID int
		fmt.Sscanf(data, "attendance_finish_%d", &classID)
		return HandleAttendanceFinish(botService, callback, classID)
	}

	// Export grades select class callback (for date range)
	if len(data) > 21 && data[:21] == "export_grades_select_" {
		var classID int
		fmt.Sscanf(data, "export_grades_select_%d", &classID)
		return HandleAdminExportGradesSelectClassCallback(botService, callback, classID)
	}

	// Export grades custom date callback
	if len(data) > 21 && data[:21] == "export_grades_custom_" {
		var classID int
		fmt.Sscanf(data, "export_grades_custom_%d", &classID)
		return HandleAdminExportGradesCustomDateCallback(botService, callback, classID)
	}

	// Export grades callback (all grades - no date filter)
	if len(data) > 14 && data[:14] == "export_grades_" {
		var classID int
		fmt.Sscanf(data, "export_grades_%d", &classID)
		return HandleAdminExportGradesCallback(botService, callback, classID)
	}

	// View class grades callback
	if len(data) > 17 && data[:17] == "view_grades_class_" {
		var classID int
		fmt.Sscanf(data, "view_grades_class_%d", &classID)
		return HandleViewClassGradesCallback(botService, callback, classID)
	}

	// View class attendance callback
	if len(data) > 22 && data[:22] == "view_attendance_class_" {
		var classID int
		fmt.Sscanf(data, "view_attendance_class_%d", &classID)
		return HandleViewClassAttendanceCallback(botService, callback, classID)
	}

	// Test result select class callback
	if len(data) > 25 && data[:25] == "test_result_select_class_" {
		var classID int
		fmt.Sscanf(data, "test_result_select_class_%d", &classID)
		return HandleTestResultClassSelectionCallback(botService, callback, classID)
	}

	// Test result add student callback
	if len(data) > 24 && data[:24] == "test_result_add_student_" {
		var studentID int
		fmt.Sscanf(data, "test_result_add_student_%d", &studentID)
		return HandleTestResultAddStudentCallback(botService, callback, studentID)
	}

	// Test result back to classes callback
	if data == "test_result_back_to_classes" {
		return HandleTestResultBackToClassesCallback(botService, callback)
	}

	// Unknown callback
	return botService.TelegramService.AnswerCallbackQuery(callback.ID, "Unknown action")
}
