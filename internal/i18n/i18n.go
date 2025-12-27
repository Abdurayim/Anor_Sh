package i18n

// Language represents a language code
type Language string

const (
	LanguageUzbek   Language = "uz"
	LanguageRussian Language = "ru"
)

// Message keys
const (
	// Commands
	MsgStart                  = "start"
	MsgHelp                   = "help"
	MsgRegister               = "register"
	MsgSubmitComplaint        = "submit_complaint"
	MsgSubmitProposal         = "submit_proposal"
	MsgMyComplaints           = "my_complaints"
	MsgMyProposals            = "my_proposals"
	MsgViewTimetable          = "view_timetable"
	MsgViewAnnouncements      = "view_announcements"
	MsgSettings               = "settings"

	// Registration flow
	MsgWelcome                = "welcome"
	MsgChooseLanguage         = "choose_language"
	MsgLanguageSelected       = "language_selected"
	MsgRequestPhone           = "request_phone"
	MsgPhoneReceived          = "phone_received"
	MsgRequestChildName       = "request_child_name"
	MsgChildNameReceived      = "child_name_received"
	MsgRequestChildClass      = "request_child_class"
	MsgRegistrationComplete   = "registration_complete"

	// Complaint flow
	MsgMainMenu               = "main_menu"
	MsgRequestComplaint       = "request_complaint"
	MsgComplaintReceived      = "complaint_received"
	MsgConfirmComplaint       = "confirm_complaint"
	MsgComplaintSubmitted     = "complaint_submitted"
	MsgComplaintCancelled     = "complaint_cancelled"

	// Proposal flow
	MsgRequestProposal        = "request_proposal"
	MsgProposalReceived       = "proposal_received"
	MsgConfirmProposal        = "confirm_proposal"
	MsgProposalSubmitted      = "proposal_submitted"
	MsgProposalCancelled      = "proposal_cancelled"

	// Timetable messages
	MsgTimetableNotFound      = "timetable_not_found"
	MsgTimetableUploaded      = "timetable_uploaded"
	MsgSelectClassForTimetable = "select_class_for_timetable"
	MsgUploadTimetableFile    = "upload_timetable_file"

	// Announcement messages
	MsgNoAnnouncements        = "no_announcements"
	MsgAnnouncementPosted     = "announcement_posted"
	MsgRequestAnnouncementTitle = "request_announcement_title"
	MsgRequestAnnouncementContent = "request_announcement_content"
	MsgRequestAnnouncementFile = "request_announcement_file"
	MsgAnnouncementSkipFile   = "announcement_skip_file"

	// Admin messages
	MsgAdminPanel             = "admin_panel"
	MsgUserList               = "user_list"
	MsgComplaintList          = "complaint_list"
	MsgProposalList           = "proposal_list"
	MsgAnnouncementsList      = "announcements_list"
	MsgStats                  = "stats"
	MsgNewComplaint           = "new_complaint"
	MsgNewProposal            = "new_proposal"

	// Teacher flow
	MsgTeacherWelcome         = "teacher_welcome"
	MsgTeacherRegistered      = "teacher_registered"
	MsgTeacherMainMenu        = "teacher_main_menu"
	MsgTeacherPanel           = "teacher_panel"

	// Student management
	MsgSelectClass            = "select_class"
	MsgRequestStudentFirstName = "request_student_first_name"
	MsgRequestStudentLastName = "request_student_last_name"
	MsgStudentAdded           = "student_added"
	MsgStudentList            = "student_list"
	MsgNoStudentsInClass      = "no_students_in_class"
	MsgSelectStudent          = "select_student"
	MsgStudentSelected        = "student_selected"

	// Parent multi-child flow
	MsgSelectYourChild        = "select_your_child"
	MsgChildLinked            = "child_linked"
	MsgMaxChildrenReached     = "max_children_reached"
	MsgMyChildren             = "my_children"
	MsgCurrentChild           = "current_child"
	MsgSwitchChild            = "switch_child"
	MsgChildSwitched          = "child_switched"
	MsgMyKidsMenu             = "my_kids_menu"
	MsgNoChildrenLinked       = "no_children_linked"
	MsgChildInfo              = "child_info"
	MsgAddChildPrompt         = "add_child_prompt"
	MsgChildNotFound          = "child_not_found"
	MsgChildAlreadyLinked     = "child_already_linked"
	MsgWaitForStudentAdd      = "wait_for_student_add"

	// Test results
	MsgRequestSubjectName     = "request_subject_name"
	MsgRequestTestScore       = "request_test_score"
	MsgRequestTestDate        = "request_test_date"
	MsgTestResultAdded        = "test_result_added"
	MsgTestResultUpdated      = "test_result_updated"
	MsgMyTestResults          = "my_test_results"
	MsgNoTestResults          = "no_test_results"
	MsgClassTestResults       = "class_test_results"

	// Attendance
	MsgMarkAttendance         = "mark_attendance"
	MsgSelectAbsentStudents   = "select_absent_students"
	MsgAttendanceMarked       = "attendance_marked"
	MsgMyAttendance           = "my_attendance"
	MsgNoAttendanceRecords    = "no_attendance_records"
	MsgAttendanceTaken        = "attendance_taken"
	MsgAttendanceNotTaken     = "attendance_not_taken"
	MsgAttendancePresent      = "attendance_present"
	MsgAttendanceAbsent       = "attendance_absent"

	// Announcement multi-class
	MsgSelectTargetClasses    = "select_target_classes"
	MsgClassesSelected        = "classes_selected"
	MsgAnnouncementBroadcast  = "announcement_broadcast"

	// Class deletion
	MsgClassDeleted           = "class_deleted"
	MsgClassDeletedReselect   = "class_deleted_reselect"
	MsgPleaseSelectNewClass   = "please_select_new_class"

	// Buttons
	BtnUzbek                  = "btn_uzbek"
	BtnRussian                = "btn_russian"
	BtnSharePhone             = "btn_share_phone"
	BtnSubmitComplaint        = "btn_submit_complaint"
	BtnSubmitProposal         = "btn_submit_proposal"
	BtnMyComplaints           = "btn_my_complaints"
	BtnMyProposals            = "btn_my_proposals"
	BtnViewTimetable          = "btn_view_timetable"
	BtnViewAnnouncements      = "btn_view_announcements"
	BtnSettings               = "btn_settings"
	BtnConfirm                = "btn_confirm"
	BtnCancel                 = "btn_cancel"
	BtnBack                   = "btn_back"
	BtnSkip                   = "btn_skip"

	// Admin buttons
	BtnAdminPanel             = "btn_admin_panel"
	BtnCreateClass            = "btn_create_class"
	BtnManageClasses          = "btn_manage_classes"
	BtnDeleteClass            = "btn_delete_class"
	BtnUploadTimetable        = "btn_upload_timetable"
	BtnViewTimetables         = "btn_view_timetables"
	BtnPostAnnouncement       = "btn_post_announcement"
	BtnViewUsers              = "btn_view_users"
	BtnViewComplaints         = "btn_view_complaints"
	BtnViewProposals          = "btn_view_proposals"
	BtnViewAllAnnouncements   = "btn_view_all_announcements"
	BtnViewStats              = "btn_view_stats"
	BtnExport                 = "btn_export"
	BtnEdit                   = "btn_edit"
	BtnDelete                 = "btn_delete"
	BtnAddTeacher             = "btn_add_teacher"
	BtnManageTeachers         = "btn_manage_teachers"
	BtnAddStudent             = "btn_add_student"
	BtnManageStudents         = "btn_manage_students"
	BtnExportTestResults      = "btn_export_test_results"
	BtnExportAttendance       = "btn_export_attendance"

	// Teacher buttons
	BtnTeacherPanel           = "btn_teacher_panel"
	BtnMyClasses              = "btn_my_classes"
	BtnAddTestResult          = "btn_add_test_result"
	BtnMarkAttendance         = "btn_mark_attendance"
	BtnViewClassStudents      = "btn_view_class_students"

	// Parent buttons
	BtnMyTestResults          = "btn_my_test_results"
	BtnMyAttendance           = "btn_my_attendance"
	BtnMyChildren             = "btn_my_children"
	BtnAddAnotherChild        = "btn_add_another_child"
	BtnSwitchChild            = "btn_switch_child"
	BtnFinishAttendance       = "btn_finish_attendance"
	BtnViewChildAttendance    = "btn_view_child_attendance"
	BtnViewChildTestResults   = "btn_view_child_test_results"

	// Errors
	ErrInvalidPhone           = "err_invalid_phone"
	ErrInvalidName            = "err_invalid_name"
	ErrInvalidClass           = "err_invalid_class"
	ErrInvalidComplaint       = "err_invalid_complaint"
	ErrInvalidProposal        = "err_invalid_proposal"
	ErrInvalidFile            = "err_invalid_file"
	ErrAlreadyRegistered      = "err_already_registered"
	ErrNotRegistered          = "err_not_registered"
	ErrNotAdmin               = "err_not_admin"
	ErrDatabaseError          = "err_database_error"
	ErrUnknownCommand         = "err_unknown_command"
	ErrTextOnly               = "err_text_only"
	ErrWrongInputType         = "err_wrong_input_type"

	// Info
	InfoProcessing            = "info_processing"
	InfoPleaseWait            = "info_please_wait"
)

// Get returns the translation for a given key and language
func Get(key string, lang Language) string {
	if lang == LanguageRussian {
		if msg, ok := russian[key]; ok {
			return msg
		}
	}

	// Default to Uzbek
	if msg, ok := uzbek[key]; ok {
		return msg
	}

	return key
}

// GetLanguage returns Language from string
func GetLanguage(lang string) Language {
	if lang == "ru" {
		return LanguageRussian
	}
	return LanguageUzbek
}

// GetLanguageString returns string from Language
func GetLanguageString(lang Language) string {
	return string(lang)
}
