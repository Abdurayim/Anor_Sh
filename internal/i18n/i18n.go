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
	MsgMyComplaints           = "my_complaints"
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

	// Admin messages
	MsgAdminPanel             = "admin_panel"
	MsgUserList               = "user_list"
	MsgComplaintList          = "complaint_list"
	MsgStats                  = "stats"
	MsgNewComplaint           = "new_complaint"

	// Buttons
	BtnUzbek                  = "btn_uzbek"
	BtnRussian                = "btn_russian"
	BtnSharePhone             = "btn_share_phone"
	BtnSubmitComplaint        = "btn_submit_complaint"
	BtnMyComplaints           = "btn_my_complaints"
	BtnSettings               = "btn_settings"
	BtnConfirm                = "btn_confirm"
	BtnCancel                 = "btn_cancel"
	BtnBack                   = "btn_back"

	// Admin buttons
	BtnAdminPanel             = "btn_admin_panel"
	BtnCreateClass            = "btn_create_class"
	BtnManageClasses          = "btn_manage_classes"
	BtnViewUsers              = "btn_view_users"
	BtnViewComplaints         = "btn_view_complaints"
	BtnViewStats              = "btn_view_stats"
	BtnExport                 = "btn_export"

	// Errors
	ErrInvalidPhone           = "err_invalid_phone"
	ErrInvalidName            = "err_invalid_name"
	ErrInvalidClass           = "err_invalid_class"
	ErrInvalidComplaint       = "err_invalid_complaint"
	ErrAlreadyRegistered      = "err_already_registered"
	ErrNotRegistered          = "err_not_registered"
	ErrDatabaseError          = "err_database_error"
	ErrUnknownCommand         = "err_unknown_command"

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
