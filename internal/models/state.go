package models

import (
	"encoding/json"
	"time"
)

// UserState represents the conversation state of a user
type UserState struct {
	TelegramID int64           `json:"telegram_id" db:"telegram_id"`
	State      string          `json:"state" db:"state"`
	Data       json.RawMessage `json:"data" db:"data"`
	UpdatedAt  time.Time       `json:"updated_at" db:"updated_at"`
}

// StateData is a helper struct for storing state data
type StateData struct {
	PhoneNumber       string `json:"phone_number,omitempty"`
	// ChildName and ChildClass are deprecated - students are now managed separately
	// Kept for backward compatibility with old state data
	ChildName         string `json:"child_name,omitempty"`
	ChildClass        string `json:"child_class,omitempty"`
	Language          string `json:"language,omitempty"`
	ComplaintText     string `json:"complaint_text,omitempty"`
	ProposalText      string `json:"proposal_text,omitempty"`
	ClassID           *int   `json:"class_id,omitempty"`
	// Selected student for complaints/proposals
	SelectedStudentID *int   `json:"selected_student_id,omitempty"`
	AnnouncementText  string `json:"announcement_text,omitempty"`
	AnnouncementTitle string `json:"announcement_title,omitempty"`
	AnnouncementID    int    `json:"announcement_id,omitempty"`
	// Attendance tracking
	AbsentList        []int  `json:"absent_list,omitempty"`
	PresentList       []int  `json:"present_list,omitempty"`
	Date              string `json:"date,omitempty"`
	// Multi-class selection (for announcements, etc.)
	SelectedClasses   []int  `json:"selected_classes,omitempty"`
	// Teacher management
	TeacherPhone      string `json:"teacher_phone,omitempty"`
	TeacherFirstName  string `json:"teacher_first_name,omitempty"`
	TeacherLastName   string `json:"teacher_last_name,omitempty"`
	// Student management
	StudentFirstName  string `json:"student_first_name,omitempty"`
	StudentLastName   string `json:"student_last_name,omitempty"`
	// Test results
	SubjectName       string `json:"subject_name,omitempty"`
	Score             string `json:"score,omitempty"`
	TestDate          string `json:"test_date,omitempty"`
	StartDate         string `json:"start_date,omitempty"`
	EndDate           string `json:"end_date,omitempty"`
	// Pagination
	Page              int    `json:"page,omitempty"`
}

// State constants
const (
	StateStart               = "start"
	StateAwaitingLanguage    = "awaiting_language"
	StateAwaitingPhone       = "awaiting_phone"
	// DEPRECATED: Child name/class are no longer collected during registration
	StateAwaitingChildName   = "awaiting_child_name"
	StateAwaitingChildClass  = "awaiting_child_class"
	StateRegistered          = "registered"

	// Parent registration - new flow: register + first child selection
	StateSelectingClass          = "selecting_class"
	StateSelectingChild          = "selecting_child"

	// Complaint/Proposal flow
	StateAwaitingComplaint   = "awaiting_complaint"
	StateConfirmingComplaint = "confirming_complaint"
	StateAwaitingProposal    = "awaiting_proposal"
	StateConfirmingProposal  = "confirming_proposal"

	// Admin states
	StateAwaitingAdminPhone  = "awaiting_admin_phone"
	StateAwaitingClassName   = "awaiting_class_name"

	// Timetable states
	StateSelectClassForTimetable = "select_class_for_timetable"
	StateAwaitingTimetableFile   = "awaiting_timetable_file"

	// Announcement states
	StateAwaitingAnnouncementTitle         = "awaiting_announcement_title"
	StateAwaitingAnnouncementContent       = "awaiting_announcement_content"
	StateAwaitingAnnouncementFile          = "awaiting_announcement_file"
	StateAwaitingEditedAnnouncementContent = "awaiting_edited_announcement_content"
	StateSelectingAnnouncementClasses      = "selecting_announcement_classes"

	// Teacher management states
	StateAwaitingTeacherPhone     = "awaiting_teacher_phone"
	StateAwaitingTeacherName      = "awaiting_teacher_name"
	StateSelectingTeacherClasses  = "selecting_teacher_classes"

	// Student management states
	StateSelectingStudentClass = "selecting_student_class"
	StateAwaitingStudentName   = "awaiting_student_name"

	// Test results states
	StateSelectingTestClass     = "selecting_test_class"
	StateSelectingTestStudent   = "selecting_test_student"
	StateAwaitingSubjectName    = "awaiting_subject_name"
	StateAwaitingScore          = "awaiting_score"
	StateAwaitingTestDate       = "awaiting_test_date"
	StateSelectingExportRange   = "selecting_export_range"

	// Attendance states
	StateSelectingAttendanceClass = "selecting_attendance_class"
	StateMarkingAttendance        = "marking_attendance"
	StateConfirmingAttendance     = "confirming_attendance"

	// My Kids states
	StateMyKidsMenu           = "my_kids_menu"
	StateAddingChild          = "adding_child"
	StateSelectingChildClass  = "selecting_child_class"
	StateSelectingChildFromClass = "selecting_child_from_class"
)
