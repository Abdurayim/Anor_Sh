package handlers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
	"parent-bot/internal/utils"
	"parent-bot/internal/validator"
)

// HandleLanguageSelection handles language selection callback
func HandleLanguageSelection(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Parse language
	var lang i18n.Language
	if callback.Data == "lang_uz" {
		lang = i18n.LanguageUzbek
	} else if callback.Data == "lang_ru" {
		lang = i18n.LanguageRussian
	} else {
		return nil
	}

	// Save language in state
	data := &models.StateData{Language: string(lang)}
	err := botService.StateManager.Set(telegramID, models.StateAwaitingPhone, data)
	if err != nil {
		return err
	}

	// Answer callback
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Send phone request message
	text := i18n.Get(i18n.MsgLanguageSelected, lang) + "\n\n" +
		i18n.Get(i18n.MsgRequestPhone, lang)

	keyboard := utils.MakePhoneKeyboard(lang)

	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// verifiedContactPhone returns the phone number only if the user shared their *own*
// contact via the request_contact button. Typed numbers or forwarded contacts of other
// people are rejected, otherwise anyone could claim an admin's or teacher's phone.
func verifiedContactPhone(message *tgbotapi.Message) (string, bool) {
	if message.Contact == nil || message.Contact.UserID != message.From.ID {
		return "", false
	}
	phone, err := validator.ValidateUzbekPhone(message.Contact.PhoneNumber)
	if err != nil {
		return "", false
	}
	return phone, true
}

// HandlePhoneNumber handles the shared contact: admins and teachers are recognized by it,
// parents continue with branch selection.
func HandlePhoneNumber(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID
	lang := i18n.GetLanguage(stateData.Language)

	// The phone can be shared with the contact button (verified: Telegram confirms it is the
	// user's own number) or typed as +998XXXXXXXXX (unverified).
	validPhone, verified := verifiedContactPhone(message)
	if !verified {
		if message.Contact != nil {
			// Someone else's contact was forwarded
			text := i18n.Get(i18n.ErrInvalidPhone, lang) + "\n\n" + i18n.Get(i18n.MsgRequestPhone, lang)
			return botService.TelegramService.SendMessage(chatID, text, utils.MakePhoneKeyboard(lang))
		}
		typed, err := validator.ValidateUzbekPhone(message.Text)
		if err != nil {
			text := i18n.Get(i18n.ErrInvalidPhone, lang)
			return botService.TelegramService.SendMessage(chatID, text, utils.MakePhoneKeyboard(lang))
		}
		validPhone = typed

		// Staff roles are granted only for a verified number, otherwise anyone could type
		// an admin's or teacher's phone and get their rights.
		if isStaffPhone(botService, validPhone) {
			text := "🔐 Bu raqam maktab xodimiga tegishli. Tasdiqlash uchun pastdagi <b>📱 tugma</b> orqali raqamingizni yuboring.\n\n" +
				"🔐 Этот номер принадлежит сотруднику школы. Для подтверждения отправьте номер <b>кнопкой 📱</b> ниже."
			return botService.TelegramService.SendMessage(chatID, text, utils.MakePhoneKeyboard(lang))
		}
	}

	if verified {
		// Admin: link this Telegram account to the admin of the configured branch
		if admin, err := botService.LinkAdminByContact(validPhone, telegramID); err == nil && admin != nil {
			_ = botService.StateManager.Clear(telegramID)
			if admin.IsSuperAdmin() {
				text := "👑 Siz <b>super admin</b> sifatida tanildingiz.\n👑 Вы распознаны как <b>супер-админ</b>."
				return botService.TelegramService.SendMessage(chatID, text, makeSuperAdminReplyKeyboard())
			}
			text := fmt.Sprintf(
				"✅ Siz <b>%s</b> admini sifatida tanildingiz.\n"+
					"✅ Вы распознаны как администратор: <b>%s</b>.",
				botService.BranchName(admin.BranchID, "uz"), botService.BranchName(admin.BranchID, "ru"),
			)
			return botService.TelegramService.SendMessage(chatID, text, utils.MakeMainMenuKeyboardWithAdmin(lang))
		}

		// Teacher: link Telegram account to the teacher added by an admin
		teacher, _ := botService.TeacherService.GetTeacherByPhoneNumber(validPhone)
		if teacher != nil && teacher.IsActive {
			_ = botService.StateManager.Clear(telegramID)
			_ = botService.TeacherService.LinkTelegramID(validPhone, telegramID, stateData.Language)

			text := i18n.Get(i18n.MsgTeacherRegistered, lang)
			keyboard := utils.MakeTeacherMainMenuKeyboard(lang)
			return botService.TelegramService.SendMessage(chatID, text, keyboard)
		}
	}

	// Check if phone number already registered
	existingUser, _ := botService.UserService.GetUserByPhoneNumber(validPhone)
	if existingUser != nil {
		text := i18n.Get(i18n.ErrAlreadyRegistered, lang)
		return botService.TelegramService.SendMessage(chatID, text, nil)
	}

	// Parent: choose branch first; the account is created after the branch is picked
	stateData.PhoneNumber = validPhone
	if err := botService.StateManager.Set(telegramID, models.StateSelectingBranch, stateData); err != nil {
		return err
	}
	return sendBranchSelection(botService, chatID, lang)
}

// isStaffPhone reports whether the phone belongs to an active admin or teacher.
func isStaffPhone(botService *services.BotService, phone string) bool {
	if admin, err := botService.AdminRepo.GetByPhoneNumber(phone); err == nil && admin != nil {
		return true
	}
	teacher, _ := botService.TeacherService.GetTeacherByPhoneNumber(phone)
	return teacher != nil && teacher.IsActive
}

// sendBranchSelection asks a parent to choose the school branch.
func sendBranchSelection(botService *services.BotService, chatID int64, lang i18n.Language) error {
	branches, err := botService.BranchRepo.GetActive()
	if err != nil || len(branches) == 0 {
		return botService.TelegramService.SendMessage(chatID, i18n.Get(i18n.ErrDatabaseError, lang), nil)
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, b := range branches {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏫 "+b.Name(string(lang)), fmt.Sprintf("branch_select_%d", b.ID)),
		))
	}

	// Remove the phone keyboard and show branch buttons
	_ = botService.TelegramService.SendMessage(chatID, "✅", utils.RemoveKeyboard())
	text := "🏫 Farzandingiz qaysi filialda o'qiydi?\n🏫 В каком филиале учится ваш ребенок?"
	return botService.TelegramService.SendMessage(chatID, text, tgbotapi.NewInlineKeyboardMarkup(rows...))
}

// ensureParentBranch asks a registered parent without a branch to choose one. It returns
// true when the message was handled (the parent has to pick a branch first).
func ensureParentBranch(botService *services.BotService, message *tgbotapi.Message) (bool, error) {
	user := parentOf(botService, message.From.ID)
	if user == nil || user.BranchID > 0 || botService.GetAdmin(message.From.ID) != nil {
		return false, nil
	}
	if teacher, _ := botService.TeacherService.GetTeacherByTelegramID(message.From.ID); teacher != nil {
		return false, nil
	}

	lang := i18n.GetLanguage(user.Language)
	_ = botService.StateManager.Set(message.From.ID, models.StateSelectingBranch, &models.StateData{Language: user.Language})
	return true, sendBranchSelection(botService, message.Chat.ID, lang)
}

// setExistingParentBranch stores the branch of an already registered parent and continues
// with child linking if the parent has no children yet.
func setExistingParentBranch(botService *services.BotService, callback *tgbotapi.CallbackQuery, user *models.User, branch *models.Branch) error {
	chatID := callback.Message.Chat.ID
	lang := i18n.GetLanguage(user.Language)

	if user.BranchID > 0 {
		// Branch is chosen once; changing it would expose another branch's classes
		return botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅")
	}
	if err := botService.UserRepo.SetBranch(user.ID, branch.ID); err != nil {
		return botService.TelegramService.SendMessage(chatID, i18n.Get(i18n.ErrDatabaseError, lang), nil)
	}
	_ = botService.StateManager.Clear(callback.From.ID)
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅ "+branch.Name(string(lang)))

	children, _ := botService.StudentRepo.GetParentStudents(user.ID)
	classes, _ := botService.ClassRepo.GetActive(branch.ID)
	if len(children) > 0 || len(classes) == 0 {
		text := "🏫 " + branch.Name(string(lang)) + "\n\n" + i18n.Get(i18n.MsgMainMenu, lang)
		return botService.TelegramService.SendMessage(chatID, text, utils.MakeMainMenuKeyboard(lang))
	}

	_ = botService.StateManager.Set(callback.From.ID, models.StateSelectingClass, &models.StateData{Language: user.Language, BranchID: branch.ID})
	text := fmt.Sprintf("🏫 %s\n\n%s", branch.Name(string(lang)), i18n.Get(i18n.MsgAddChildPrompt, lang))
	return botService.TelegramService.SendMessage(chatID, text, utils.MakeClassSelectionKeyboardWithBack(classes, lang))
}

// HandleBranchSelection creates the parent account in the chosen branch and continues
// with linking the first child.
func HandleBranchSelection(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	branchID, ok := scan1(callback.Data, "branch_select_%d")
	branch, err := botService.BranchRepo.GetByID(branchID)
	if !ok || err != nil || branch == nil || !branch.IsActive {
		return botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Filial topilmadi / Филиал не найден")
	}

	state, _ := botService.StateManager.Get(telegramID)
	stateData, _ := botService.StateManager.GetData(telegramID)

	// Existing parent without a branch (registered before branches existed)
	if existing := parentOf(botService, telegramID); existing != nil {
		return setExistingParentBranch(botService, callback, existing, branch)
	}

	if state == nil || state.State != models.StateSelectingBranch || stateData == nil || stateData.PhoneNumber == "" {
		return botService.TelegramService.AnswerCallbackQuery(callback.ID, "⏳ /start")
	}
	lang := i18n.GetLanguage(stateData.Language)
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅ "+branch.Name(string(lang)))

	user, err := botService.UserService.CreateUser(&models.CreateUserRequest{
		TelegramID:       telegramID,
		TelegramUsername: callback.From.UserName,
		PhoneNumber:      stateData.PhoneNumber,
		Language:         stateData.Language,
		BranchID:         branch.ID,
	})
	if err != nil {
		return botService.TelegramService.SendMessage(chatID, i18n.Get(i18n.ErrDatabaseError, lang), nil)
	}

	classes, err := botService.ClassRepo.GetActive(user.BranchID)
	if err != nil || len(classes) == 0 {
		_ = botService.StateManager.Clear(telegramID)
		text := fmt.Sprintf(
			"✅ Ro'yxatdan o'tish muvaffaqiyatli yakunlandi! (%s)\n\n"+
				"Hozircha sinflar mavjud emas. Keyinroq farzandingizni qo'shishingiz mumkin.\n\n"+
				"✅ Регистрация успешно завершена! (%s)\n\n"+
				"Пока классов нет. Вы сможете добавить ребенка позже.",
			branch.NameUz, branch.NameRu,
		)
		return botService.TelegramService.SendMessage(chatID, text, utils.MakeMainMenuKeyboard(lang))
	}

	stateData.BranchID = branch.ID
	if err := botService.StateManager.Set(telegramID, models.StateSelectingClass, stateData); err != nil {
		return err
	}

	text := fmt.Sprintf("🏫 %s\n\n%s", branch.Name(string(lang)), i18n.Get(i18n.MsgAddChildPrompt, lang))
	keyboard := utils.MakeClassSelectionKeyboardWithBack(classes, lang)
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleChildName - DEPRECATED: No longer used in new architecture
// Students are now managed separately and linked to parents by admin/teachers
func HandleChildName(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Redirect to registration completion
	text := "Registration flow has been updated. Please use /start to begin registration."
	_ = botService.StateManager.Clear(telegramID)
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleClassSelection handles class selection from inline keyboard
func HandleClassSelection(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	telegramID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Answer callback query - DEPRECATED FLOW
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	// Redirect to new registration flow
	text := "Registration flow has been updated. Please use /start to begin registration."
	_ = botService.StateManager.Clear(telegramID)
	return botService.TelegramService.SendMessage(chatID, text, nil)
}

// HandleChildClass - DEPRECATED: No longer used in new architecture
// This is kept for backward compatibility but now we prefer inline buttons
func HandleChildClass(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	// Redirect to new registration flow
	text := "Registration flow has been updated. Please use /start to begin registration."
	_ = botService.StateManager.Clear(telegramID)
	return botService.TelegramService.SendMessage(chatID, text, nil)
}
