package utils

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
)

// MakeLanguageKeyboard creates language selection keyboard
func MakeLanguageKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnUzbek, i18n.LanguageUzbek),
				"lang_uz",
			),
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnRussian, i18n.LanguageRussian),
				"lang_ru",
			),
		),
	)
}

// MakePhoneKeyboard creates phone number request keyboard
func MakePhoneKeyboard(lang i18n.Language) tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact(i18n.Get(i18n.BtnSharePhone, lang)),
		),
	)
}

// MakeMainMenuKeyboard creates main menu keyboard for parents
func MakeMainMenuKeyboard(lang i18n.Language) tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		// Row 1: My Children (main action)
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnMyChildren, lang)),
		),
		// Row 2: Child Info (Attendance & Test Results)
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnMyAttendance, lang)),
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnMyTestResults, lang)),
		),
		// Row 3: School Info (Timetable & Announcements)
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnViewTimetable, lang)),
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnViewAnnouncements, lang)),
		),
		// Row 4: Complaints & Proposals
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnSubmitComplaint, lang)),
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnSubmitProposal, lang)),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}

// MakeMainMenuKeyboardWithAdmin creates admin-only main menu keyboard
func MakeMainMenuKeyboardWithAdmin(lang i18n.Language) tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnAdminPanel, lang)),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}

// MakeMainMenuKeyboardForUser creates main menu keyboard based on user's admin status
func MakeMainMenuKeyboardForUser(lang i18n.Language, isAdmin bool) tgbotapi.ReplyKeyboardMarkup {
	if isAdmin {
		return MakeMainMenuKeyboardWithAdmin(lang)
	}
	return MakeMainMenuKeyboard(lang)
}

// MakeConfirmationKeyboard creates confirmation keyboard for complaints
func MakeConfirmationKeyboard(lang i18n.Language) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnConfirm, lang),
				"confirm_complaint",
			),
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnCancel, lang),
				"cancel_complaint",
			),
		),
	)
}

// MakeProposalConfirmationKeyboard creates confirmation keyboard for proposals
func MakeProposalConfirmationKeyboard(lang i18n.Language) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnConfirm, lang),
				"confirm_proposal",
			),
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnCancel, lang),
				"cancel_proposal",
			),
		),
	)
}

// MakeAdminKeyboard creates admin panel keyboard
func MakeAdminKeyboard(lang i18n.Language) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		// Row 1: Class Management
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnManageClasses, lang),
				"admin_manage_classes",
			),
		),
		// Row 2: Teacher Management
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnManageTeachers, lang),
				"admin_manage_teachers",
			),
		),
		// Row 3: Announcement Management
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnPostAnnouncement, lang),
				"admin_post_announcement",
			),
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnViewAllAnnouncements, lang),
				"admin_view_announcements",
			),
		),
		// Row 4: Attendance Export
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnExportAttendance, lang),
				"admin_export_attendance",
			),
		),
		// Row 5: Test Results Export
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnExportTestResults, lang),
				"admin_export_test_results",
			),
		),
		// Row 6: Timetable Management
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnUploadTimetable, lang),
				"admin_upload_timetable",
			),
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnViewTimetables, lang),
				"admin_view_timetables",
			),
		),
		// Row 7: Complaints & Proposals
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnViewComplaints, lang),
				"admin_complaints",
			),
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnViewProposals, lang),
				"admin_proposals",
			),
		),
		// Row 8: Users & Stats
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnViewUsers, lang),
				"admin_users",
			),
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnViewStats, lang),
				"admin_stats",
			),
		),
	)
}

// RemoveKeyboard creates a keyboard removal markup
func RemoveKeyboard() tgbotapi.ReplyKeyboardRemove {
	return tgbotapi.NewRemoveKeyboard(true)
}

// MakeClassSelectionKeyboard creates class selection inline keyboard
func MakeClassSelectionKeyboard(classes []*models.Class, lang i18n.Language) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Create buttons in rows of 3
	var row []tgbotapi.InlineKeyboardButton
	for i, class := range classes {
		button := tgbotapi.NewInlineKeyboardButtonData(
			class.ClassName,
			"class_"+class.ClassName,
		)
		row = append(row, button)

		// Add row every 3 buttons or at the end
		if (i+1)%3 == 0 || i == len(classes)-1 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}

	_ = lang // Will be used in future for localized buttons

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// MakeClassSelectionKeyboardWithBack creates class selection keyboard with back button
func MakeClassSelectionKeyboardWithBack(classes []*models.Class, lang i18n.Language) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Create buttons in rows of 3
	var row []tgbotapi.InlineKeyboardButton
	for i, class := range classes {
		button := tgbotapi.NewInlineKeyboardButtonData(
			class.ClassName,
			"select_class_"+class.ClassName,
		)
		row = append(row, button)

		// Add row every 3 buttons or at the end
		if (i+1)%3 == 0 || i == len(classes)-1 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}

	// Add skip button for registration (can add child later)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			i18n.Get(i18n.BtnSkip, lang),
			"skip_child_selection",
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// MakeStudentSelectionKeyboard creates student selection keyboard for a class
func MakeStudentSelectionKeyboard(students []*models.StudentWithClass, lang i18n.Language) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, student := range students {
		button := tgbotapi.NewInlineKeyboardButtonData(
			student.LastName+" "+student.FirstName,
			fmt.Sprintf("select_student_%d", student.ID),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}

	// Add back button
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			i18n.Get(i18n.BtnBack, lang),
			"back_to_class_selection",
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// MakeTeacherMainMenuKeyboard creates teacher main menu keyboard
func MakeTeacherMainMenuKeyboard(lang i18n.Language) tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		// Row 1: Student Management
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnAddStudent, lang)),
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnViewClassStudents, lang)),
		),
		// Row 2: Attendance & Test Results
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnMarkAttendance, lang)),
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnAddTestResult, lang)),
		),
		// Row 3: Announcements
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnPostAnnouncement, lang)),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}

// MakeMyKidsKeyboard creates My Kids menu inline keyboard
func MakeMyKidsKeyboard(children []*models.ParentChild, canAddMore bool, lang i18n.Language) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Add buttons for each child
	for _, child := range children {
		childName := child.StudentLastName + " " + child.StudentFirstName + " (" + child.ClassName + ")"
		button := tgbotapi.NewInlineKeyboardButtonData(
			childName,
			fmt.Sprintf("view_child_%d", child.StudentID),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}

	// Add "Add Another Child" button if under limit
	if canAddMore {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnAddAnotherChild, lang),
				"add_another_child",
			),
		))
	}

	// Add back button
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(
			i18n.Get(i18n.BtnBack, lang),
			"back_to_main",
		),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// MakeChildActionsKeyboard creates action buttons for a specific child
func MakeChildActionsKeyboard(childID int, lang i18n.Language) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnViewChildAttendance, lang),
				fmt.Sprintf("view_child_attendance_%d", childID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnViewChildTestResults, lang),
				fmt.Sprintf("view_child_grades_%d", childID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnBack, lang),
				"back_to_my_kids",
			),
		),
	)
}
