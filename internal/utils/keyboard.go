package utils

import (
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

// MakeMainMenuKeyboard creates main menu keyboard
func MakeMainMenuKeyboard(lang i18n.Language) tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnSubmitComplaint, lang)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnMyComplaints, lang)),
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnSettings, lang)),
		),
	)
}

// MakeMainMenuKeyboardWithAdmin creates main menu keyboard with admin button
func MakeMainMenuKeyboardWithAdmin(lang i18n.Language) tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnSubmitComplaint, lang)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnMyComplaints, lang)),
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnSettings, lang)),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(i18n.Get(i18n.BtnAdminPanel, lang)),
		),
	)
}

// MakeMainMenuKeyboardForUser creates main menu keyboard based on user's admin status
func MakeMainMenuKeyboardForUser(lang i18n.Language, isAdmin bool) tgbotapi.ReplyKeyboardMarkup {
	if isAdmin {
		return MakeMainMenuKeyboardWithAdmin(lang)
	}
	return MakeMainMenuKeyboard(lang)
}

// MakeConfirmationKeyboard creates confirmation keyboard
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

// MakeAdminKeyboard creates admin panel keyboard
func MakeAdminKeyboard(lang i18n.Language) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnManageClasses, lang),
				"admin_manage_classes",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnViewUsers, lang),
				"admin_users",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				i18n.Get(i18n.BtnViewComplaints, lang),
				"admin_complaints",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
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
