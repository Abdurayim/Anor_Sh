package handlers

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
	"parent-bot/internal/utils"
)

// Staff roles a phone number can be waiting to be linked to.
const (
	staffRoleAdmin   = "admin"
	staffRoleTeacher = "teacher"
)

const btnCancelStaffContact = "❌ Bekor qilish / Отмена"

// teacherMenuKeyboard returns the main menu of a teacher or tutor in their language.
func teacherMenuKeyboard(teacher *models.Teacher) tgbotapi.ReplyKeyboardMarkup {
	lang := i18n.GetLanguage(teacher.Language)
	if teacher.IsTutor() {
		return utils.MakeTutorMainMenuKeyboard(lang)
	}
	return utils.MakeTeacherMainMenuKeyboard(lang)
}

// teacherRoleLabel returns the bilingual name of a teacher role.
func teacherRoleLabel(role string) string {
	if role == models.TeacherRoleTutor {
		return "⭐ Tyutor / Тьютор"
	}
	return "👨‍🏫 O'qituvchi / Учитель"
}

// menuKeyboardFor returns the main menu of whatever role this Telegram account has now.
func menuKeyboardFor(botService *services.BotService, telegramID int64) interface{} {
	if botService.GetSuperAdmin(telegramID) != nil {
		return makeSuperAdminReplyKeyboard()
	}
	if teacher, _ := botService.TeacherService.GetTeacherByTelegramID(telegramID); teacher != nil {
		return teacherMenuKeyboard(teacher)
	}
	user := parentOf(botService, telegramID)
	lang := i18n.LanguageUzbek
	if user != nil {
		lang = i18n.GetLanguage(user.Language)
	}
	if botService.GetAdmin(telegramID) != nil {
		return utils.MakeMainMenuKeyboardWithAdmin(lang)
	}
	if user != nil {
		return utils.MakeMainMenuKeyboard(lang)
	}
	return utils.RemoveKeyboard()
}

// ============================================================================
// Linking an already registered account to a staff role
// ============================================================================

// pendingStaffRole reports whether the phone belongs to an admin or teacher record that no
// Telegram account is linked to yet. This is the case when someone registered as a parent
// first and was made a teacher or admin afterwards.
func pendingStaffRole(botService *services.BotService, phone string) string {
	if phone == "" {
		return ""
	}
	if admin, err := botService.AdminRepo.GetByPhoneNumber(phone); err == nil && admin != nil {
		if admin.TelegramID == nil {
			return staffRoleAdmin
		}
		return ""
	}
	teacher, _ := botService.TeacherRepo.GetByPhoneNumber(phone)
	if teacher != nil && teacher.IsActive && teacher.TelegramID == nil {
		return staffRoleTeacher
	}
	return ""
}

// promptStaffContact asks the user to confirm a staff role with the contact button. The
// phone stored for a parent may have been typed, so it is never trusted on its own.
func promptStaffContact(botService *services.BotService, chatID, telegramID int64, role string) error {
	if err := botService.StateManager.Set(telegramID, models.StateAwaitingStaffContact, &models.StateData{}); err != nil {
		return err
	}

	var text string
	switch role {
	case staffRoleAdmin:
		text = "👨‍💼 Sizning raqamingiz <b>administrator</b> sifatida qo'shilgan.\n" +
			"👨‍💼 Ваш номер добавлен как <b>администратор</b>.\n\n"
	case staffRoleTeacher:
		text = "👨‍🏫 Sizning raqamingiz <b>o'qituvchi</b> sifatida qo'shilgan.\n" +
			"👨‍🏫 Ваш номер добавлен как <b>учитель</b>.\n\n"
	default:
		text = "🔄 <b>Rolni yangilash / Обновление роли</b>\n\n" +
			"Agar raqamingiz administrator yoki o'qituvchi sifatida qo'shilgan bo'lsa, tegishli menyu ochiladi.\n" +
			"Если ваш номер добавлен как администратор или учитель, откроется нужное меню.\n\n"
	}
	text += "Tasdiqlash uchun pastdagi <b>📱 tugma</b> orqali raqamingizni yuboring.\n" +
		"Для подтверждения отправьте свой номер <b>кнопкой 📱</b> ниже."

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("📱 Raqamni yuborish / Отправить номер"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(btnCancelStaffContact),
		),
	)
	keyboard.ResizeKeyboard = true
	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleRoleCommand handles /role: anyone can re-check their role by sharing their contact.
func HandleRoleCommand(botService *services.BotService, message *tgbotapi.Message) error {
	return promptStaffContact(botService, message.Chat.ID, message.From.ID, "")
}

// HandleStaffContact links the Telegram account to the admin or teacher record of the
// shared (verified) phone number. Anything else cancels and brings the user's menu back.
func HandleStaffContact(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID
	_ = botService.StateManager.Clear(telegramID)

	phone, verified := verifiedContactPhone(message)
	if !verified {
		text := "❌ Bekor qilindi / Отменено"
		if message.Contact != nil {
			text = "❌ Iltimos, pastdagi tugma orqali <b>o'z</b> raqamingizni ulashing. Qaytadan: /role\n" +
				"❌ Пожалуйста, поделитесь <b>своим</b> номером кнопкой. Повторить: /role"
		}
		return botService.TelegramService.SendMessage(chatID, text, menuKeyboardFor(botService, telegramID))
	}

	if admin, err := botService.LinkAdminByContact(phone, telegramID); err == nil && admin != nil {
		if admin.IsSuperAdmin() {
			text := "👑 Siz <b>super admin</b> sifatida tanildingiz.\n👑 Вы распознаны как <b>супер-админ</b>."
			return botService.TelegramService.SendMessage(chatID, text, makeSuperAdminReplyKeyboard())
		}
		text := fmt.Sprintf(
			"✅ Siz <b>%s</b> admini sifatida tanildingiz.\n"+
				"✅ Вы распознаны как администратор: <b>%s</b>.",
			botService.BranchName(admin.BranchID, "uz"), botService.BranchName(admin.BranchID, "ru"),
		)
		return botService.TelegramService.SendMessage(chatID, text, menuKeyboardFor(botService, telegramID))
	}

	teacher, _ := botService.TeacherRepo.GetByPhoneNumber(phone)
	if teacher != nil && teacher.IsActive {
		// A parent keeps the language they already use in the bot
		language := ""
		if user := parentOf(botService, telegramID); user != nil && teacher.TelegramID == nil {
			language = user.Language
		}
		if err := botService.TeacherService.LinkTelegramID(phone, telegramID, language); err != nil {
			log.Printf("Failed to link teacher %d to telegram %d: %v", teacher.ID, telegramID, err)
			text := "❌ Bu raqam boshqa Telegram akkauntiga bog'langan. Administratorga murojaat qiling.\n" +
				"❌ Этот номер привязан к другому аккаунту Telegram. Обратитесь к администратору."
			return botService.TelegramService.SendMessage(chatID, text, menuKeyboardFor(botService, telegramID))
		}
		linked, _ := botService.TeacherService.GetTeacherByTelegramID(telegramID)
		if linked == nil {
			return botService.TelegramService.SendMessage(chatID, msgNotAllowed, menuKeyboardFor(botService, telegramID))
		}
		text := i18n.Get(i18n.MsgTeacherRegistered, i18n.GetLanguage(linked.Language))
		if linked.IsTutor() {
			text += "\n\n" + teacherRoleLabel(linked.Role)
		}
		return botService.TelegramService.SendMessage(chatID, text, teacherMenuKeyboard(linked))
	}

	text := fmt.Sprintf("ℹ️ <code>%s</code> raqami administrator yoki o'qituvchi sifatida qo'shilmagan.\n"+
		"ℹ️ Номер <code>%s</code> не добавлен как администратор или учитель.", phone, phone)
	return botService.TelegramService.SendMessage(chatID, text, menuKeyboardFor(botService, telegramID))
}

// notifyPendingStaff tells an already registered user that their phone was just given a
// staff role. It returns true when such a user exists; they confirm the role on /start.
func notifyPendingStaff(botService *services.BotService, phone, role string) bool {
	user, err := botService.UserRepo.GetByPhoneNumber(phone)
	if err != nil || user == nil {
		return false
	}

	text := "👨‍🏫 Sizga <b>o'qituvchi</b> huquqi berildi.\n👨‍🏫 Вам выданы права <b>учителя</b>.\n\n"
	if role == staffRoleAdmin {
		text = "👨‍💼 Sizga <b>administrator</b> huquqi berildi.\n👨‍💼 Вам выданы права <b>администратора</b>.\n\n"
	}
	text += "Faollashtirish uchun /start bosing va raqamingizni tasdiqlang.\n" +
		"Чтобы активировать, нажмите /start и подтвердите свой номер."
	if err := botService.TelegramService.SendMessage(user.TelegramID, text, nil); err != nil {
		log.Printf("Failed to notify %d about the %s role: %v", user.TelegramID, role, err)
	}
	return true
}

// ============================================================================
// Teacher language
// ============================================================================

// HandleTeacherLanguageCommand shows the language choice to a teacher.
func HandleTeacherLanguageCommand(botService *services.BotService, message *tgbotapi.Message) error {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(i18n.Get(i18n.BtnUzbek, i18n.LanguageUzbek), "teacher_lang_uz"),
			tgbotapi.NewInlineKeyboardButtonData(i18n.Get(i18n.BtnRussian, i18n.LanguageRussian), "teacher_lang_ru"),
		),
	)
	return botService.TelegramService.SendMessage(message.Chat.ID, "🌐 Tilni tanlang / Выберите язык:", &keyboard)
}

// HandleTeacherLanguageCallback saves the chosen language and redraws the teacher menu in it.
func HandleTeacherLanguageCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	teacher, _ := botService.TeacherService.GetTeacherByTelegramID(callback.From.ID)
	if teacher == nil {
		return botService.TelegramService.AnswerCallbackQuery(callback.ID, msgNotAllowed)
	}

	language := string(i18n.LanguageUzbek)
	text := "✅ Til o'zgartirildi: O'zbek"
	if callback.Data == "teacher_lang_ru" {
		language = string(i18n.LanguageRussian)
		text = "✅ Язык изменён: Русский"
	}

	if err := botService.TeacherRepo.Update(teacher.ID, &models.UpdateTeacherRequest{Language: language}); err != nil {
		log.Printf("Failed to change language of teacher %d: %v", teacher.ID, err)
		return botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
	}
	teacher.Language = language

	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "✅")
	if callback.Message != nil {
		_ = botService.TelegramService.DeleteMessage(callback.Message.Chat.ID, callback.Message.MessageID)
		return botService.TelegramService.SendMessage(callback.Message.Chat.ID, text, teacherMenuKeyboard(teacher))
	}
	return nil
}

// ============================================================================
// Admin: teacher card and tutor role
// ============================================================================

// HandleAdminTeacherCardCallback shows one teacher with the role and delete actions.
func HandleAdminTeacherCardCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, teacherID int) error {
	teacher, err := botService.TeacherRepo.GetByID(teacherID)
	if err != nil || teacher == nil {
		return botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ O'qituvchi topilmadi / Учитель не найден")
	}
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	status := "✅ Botga ulangan / Подключён к боту"
	if !teacher.IsActive {
		status = "❌ Faol emas / Неактивен"
	} else if teacher.TelegramID == nil {
		status = "⏳ Botga hali kirmagan / Ещё не вошёл в бот"
	}

	text := fmt.Sprintf("👤 <b>%s %s</b>\n📱 <code>%s</code>\n🎓 %s\n%s",
		utils.EscapeHTML(teacher.FirstName), utils.EscapeHTML(teacher.LastName),
		teacher.PhoneNumber, teacherRoleLabel(teacher.Role), status)

	roleButton := tgbotapi.NewInlineKeyboardButtonData(
		"⭐ Tyutor qilish / Сделать тьютором", fmt.Sprintf("admin_teacher_tutor_%d", teacher.ID))
	if teacher.IsTutor() {
		roleButton = tgbotapi.NewInlineKeyboardButtonData(
			"👨‍🏫 O'qituvchi qilish / Сделать учителем", fmt.Sprintf("admin_teacher_plain_%d", teacher.ID))
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(roleButton),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗑 O'chirish / Удалить", fmt.Sprintf("admin_delete_teacher_%d", teacher.ID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Orqaga / Назад", "admin_manage_teachers"),
		),
	)
	return botService.TelegramService.EditMessage(callback.Message.Chat.ID, callback.Message.MessageID, text, &keyboard)
}

// HandleAdminTeacherRoleCallback promotes a teacher to tutor or back, and gives the teacher
// their new menu straight away.
func HandleAdminTeacherRoleCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery, teacherID int, role string) error {
	teacher, err := botService.TeacherRepo.GetByID(teacherID)
	if err != nil || teacher == nil {
		return botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ O'qituvchi topilmadi / Учитель не найден")
	}

	if teacher.Role != role {
		if err := botService.TeacherRepo.SetRole(teacherID, role); err != nil {
			log.Printf("Failed to set role of teacher %d: %v", teacherID, err)
			return botService.TelegramService.AnswerCallbackQuery(callback.ID, "❌ Xatolik / Ошибка")
		}
		teacher.Role = role

		if teacher.IsActive && teacher.TelegramID != nil {
			text := "ℹ️ Sizning rolingiz o'zgardi / Ваша роль изменена:\n<b>" + teacherRoleLabel(role) + "</b>"
			if err := botService.TelegramService.SendMessage(*teacher.TelegramID, text, teacherMenuKeyboard(teacher)); err != nil {
				log.Printf("Failed to notify teacher %d about the new role: %v", teacherID, err)
			}
		}
	}

	return HandleAdminTeacherCardCallback(botService, callback, teacherID)
}

// ============================================================================
// Timetables: branch admins and tutors
// ============================================================================

// timetableActor is whoever may upload and delete timetables: a branch admin or a tutor.
type timetableActor struct {
	branchID     int
	lang         i18n.Language
	adminID      *int
	teacherID    *int
	backCallback string
}

// timetableActorOf returns the timetable rights of this Telegram account, or nil.
func timetableActorOf(botService *services.BotService, telegramID int64) *timetableActor {
	if admin := botService.GetAdmin(telegramID); admin != nil && admin.BranchID > 0 {
		lang := i18n.LanguageUzbek
		if user := parentOf(botService, telegramID); user != nil {
			lang = i18n.GetLanguage(user.Language)
		}
		return &timetableActor{branchID: admin.BranchID, lang: lang, adminID: &admin.ID, backCallback: "admin_back"}
	}
	teacher, _ := botService.TeacherService.GetTeacherByTelegramID(telegramID)
	if teacher != nil && teacher.IsTutor() && teacher.BranchID > 0 {
		return &timetableActor{
			branchID:     teacher.BranchID,
			lang:         i18n.GetLanguage(teacher.Language),
			teacherID:    &teacher.ID,
			backCallback: "teacher_back_to_main",
		}
	}
	return nil
}
