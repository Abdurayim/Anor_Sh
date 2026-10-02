package handlers

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
	"parent-bot/internal/utils"
	"parent-bot/internal/validator"
)

// The super admin sees statistics, complaints and proposals of all branches (read-only)
// and manages branch admins. All callbacks start with "sa_" and are checked in authorizeCallback.

const (
	btnSuperAdminPanel = "👑 Super admin paneli"

	stateSuperAdminAwaitingAdminPhone = "sa_awaiting_admin_phone"
)

// makeSuperAdminReplyKeyboard is the persistent keyboard of the super admin.
func makeSuperAdminReplyKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(btnSuperAdminPanel)))
	keyboard.ResizeKeyboard = true
	return keyboard
}

func makeSuperAdminPanelKeyboard(branches []*models.Branch) tgbotapi.InlineKeyboardMarkup {
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("📊 Umumiy dashboard / Общий дашборд", "sa_dashboard")),
	}
	var branchRow []tgbotapi.InlineKeyboardButton
	for _, b := range branches {
		branchRow = append(branchRow, tgbotapi.NewInlineKeyboardButtonData("🏫 "+b.NameUz, fmt.Sprintf("sa_branch_%d", b.ID)))
	}
	if len(branchRow) > 0 {
		rows = append(rows, branchRow)
	}
	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Shikoyatlar / Жалобы", "sa_complaints"),
			tgbotapi.NewInlineKeyboardButtonData("💡 Takliflar / Предложения", "sa_proposals"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👨‍💼 Adminlar / Админы", "sa_admins"),
			tgbotapi.NewInlineKeyboardButtonData("➕ Admin qo'shish / Добавить", "sa_add_admin"),
		),
	)
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func backToSuperAdminPanel() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Orqaga / Назад", "sa_panel"),
	))
}

// HandleSuperAdminPanel shows the super admin panel (reply button or /superadmin).
func HandleSuperAdminPanel(botService *services.BotService, message *tgbotapi.Message) error {
	if botService.GetSuperAdmin(message.From.ID) == nil {
		return botService.TelegramService.SendMessage(message.Chat.ID, msgNotAllowed, nil)
	}
	_ = botService.StateManager.Clear(message.From.ID)
	branches, _ := botService.BranchRepo.GetActive()
	text := "👑 <b>Super admin paneli / Панель супер-админа</b>\n\n" +
		"Barcha filiallar statistikasi va adminlarni boshqarish.\n" +
		"Статистика всех филиалов и управление администраторами."
	return botService.TelegramService.SendMessage(message.Chat.ID, text, makeSuperAdminPanelKeyboard(branches))
}

// HandleSuperAdminCallback routes all "sa_" callbacks.
func HandleSuperAdminCallback(botService *services.BotService, callback *tgbotapi.CallbackQuery) error {
	superAdmin := botService.GetSuperAdmin(callback.From.ID)
	if superAdmin == nil {
		return botService.TelegramService.AnswerCallbackQuery(callback.ID, msgNotAllowed)
	}
	_ = botService.TelegramService.AnswerCallbackQuery(callback.ID, "")

	chatID := callback.Message.Chat.ID
	messageID := callback.Message.MessageID
	edit := func(text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
		return botService.TelegramService.EditMessage(chatID, messageID, text, &keyboard)
	}
	data := callback.Data

	switch {
	case data == "sa_panel":
		branches, _ := botService.BranchRepo.GetActive()
		text := "👑 <b>Super admin paneli / Панель супер-админа</b>"
		return edit(text, makeSuperAdminPanelKeyboard(branches))

	case data == "sa_dashboard":
		return edit(superAdminDashboardText(botService), backToSuperAdminPanel())

	case strings.HasPrefix(data, "sa_branch_"):
		branchID, _ := scan1(data, "sa_branch_%d")
		return edit(superAdminBranchText(botService, branchID), backToSuperAdminPanel())

	case data == "sa_complaints":
		return edit(superAdminComplaintsText(botService), backToSuperAdminPanel())

	case data == "sa_proposals":
		return edit(superAdminProposalsText(botService), backToSuperAdminPanel())

	case data == "sa_admins":
		text, keyboard := superAdminAdminsView(botService)
		return edit(text, keyboard)

	case data == "sa_add_admin":
		branches, _ := botService.BranchRepo.GetActive()
		var rows [][]tgbotapi.InlineKeyboardButton
		for _, b := range branches {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏫 "+b.NameUz, fmt.Sprintf("sa_add_admin_branch_%d", b.ID)),
			))
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️ Orqaga / Назад", "sa_panel")))
		return edit("➕ Yangi admin qaysi filialga? / Для какого филиала новый админ?", tgbotapi.NewInlineKeyboardMarkup(rows...))

	case strings.HasPrefix(data, "sa_add_admin_branch_"):
		branchID, ok := scan1(data, "sa_add_admin_branch_%d")
		branch, err := botService.BranchRepo.GetByID(branchID)
		if !ok || err != nil || branch == nil {
			return nil
		}
		if err := botService.StateManager.Set(callback.From.ID, stateSuperAdminAwaitingAdminPhone, &models.StateData{BranchID: branch.ID}); err != nil {
			return err
		}
		text := fmt.Sprintf("🏫 <b>%s</b>\n\n"+
			"Yangi adminning telefon raqamini yuboring (masalan: +998901234567).\n"+
			"Отправьте номер телефона нового админа (например: +998901234567).\n\n"+
			"Bekor qilish / Отмена: /cancel", utils.EscapeHTML(branch.NameUz))
		return botService.TelegramService.SendMessage(chatID, text, nil)

	case strings.HasPrefix(data, "sa_remove_admin_confirm_"):
		adminID, _ := scan1(data, "sa_remove_admin_confirm_%d")
		if err := botService.RemoveBranchAdmin(superAdmin, adminID); err != nil {
			return botService.TelegramService.SendMessage(chatID, "❌ "+utils.EscapeHTML(err.Error()), nil)
		}
		text, keyboard := superAdminAdminsView(botService)
		return edit("✅ Admin o'chirildi / Админ удален\n\n"+text, keyboard)

	case strings.HasPrefix(data, "sa_remove_admin_"):
		adminID, _ := scan1(data, "sa_remove_admin_%d")
		admin, err := botService.AdminRepo.GetByID(adminID)
		if err != nil || admin == nil {
			return nil
		}
		text := fmt.Sprintf("⚠️ %s (%s) adminlikdan olib tashlansinmi?\n⚠️ Снять права администратора?",
			admin.PhoneNumber, utils.EscapeHTML(botService.BranchName(admin.BranchID, "uz")))
		keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗑 Ha / Да", fmt.Sprintf("sa_remove_admin_confirm_%d", adminID)),
			tgbotapi.NewInlineKeyboardButtonData("↩️ Bekor / Отмена", "sa_admins"),
		))
		return edit(text, keyboard)
	}
	return nil
}

// HandleSuperAdminAdminPhone receives the phone of a new branch admin.
func HandleSuperAdminAdminPhone(botService *services.BotService, message *tgbotapi.Message, stateData *models.StateData) error {
	chatID := message.Chat.ID
	superAdmin := botService.GetSuperAdmin(message.From.ID)
	if superAdmin == nil {
		_ = botService.StateManager.Clear(message.From.ID)
		return botService.TelegramService.SendMessage(chatID, msgNotAllowed, nil)
	}

	raw := message.Text
	if message.Contact != nil {
		raw = message.Contact.PhoneNumber
	}
	phone, err := validator.ValidateUzbekPhone(raw)
	if err != nil {
		return botService.TelegramService.SendMessage(chatID,
			"❌ Noto'g'ri raqam. Namuna: +998901234567\n❌ Неверный номер. Пример: +998901234567", nil)
	}

	branchName := botService.BranchName(stateData.BranchID, "uz")
	if err := botService.AddBranchAdmin(superAdmin, phone, "Admin "+branchName, stateData.BranchID); err != nil {
		_ = botService.StateManager.Clear(message.From.ID)
		return botService.TelegramService.SendMessage(chatID, "❌ "+utils.EscapeHTML(err.Error()), makeSuperAdminReplyKeyboard())
	}
	_ = botService.StateManager.Clear(message.From.ID)

	text := fmt.Sprintf("✅ <b>%s</b> — <b>%s</b> admini qo'shildi.\n\n"+
		"Admin botga /start yuborib, raqamini <b>📱 tugma</b> orqali ulashsin — shundan keyin admin paneli ochiladi.\n\n"+
		"✅ Админ филиала добавлен. Пусть он отправит боту /start и поделится номером <b>кнопкой 📱</b>.",
		phone, utils.EscapeHTML(branchName))
	return botService.TelegramService.SendMessage(chatID, text, makeSuperAdminReplyKeyboard())
}

func formatStats(st services.BranchStats) string {
	attendance := "olinmagan / не отмечена"
	if rate := st.AttendanceRate(); rate >= 0 {
		attendance = fmt.Sprintf("%.0f%% (✅ %d / ❌ %d)", rate, st.PresentToday, st.AbsentToday)
	}
	avg := "—"
	if st.Scores30d > 0 {
		avg = fmt.Sprintf("%.1f (%d ta baho)", st.AvgScore30d, st.Scores30d)
	}
	return fmt.Sprintf(
		"👨‍👩‍👧 Ota-onalar / Родители: <b>%d</b>\n"+
			"👦 O'quvchilar / Ученики: <b>%d</b>\n"+
			"👨‍🏫 O'qituvchilar / Учителя: <b>%d</b>\n"+
			"📚 Sinflar / Классы: <b>%d</b>\n"+
			"👨‍💼 Adminlar / Админы: <b>%d</b>\n"+
			"📋 Bugungi davomat / Посещаемость сегодня: <b>%s</b>\n"+
			"📈 O'rtacha baho (30 kun) / Средний балл: <b>%s</b>\n"+
			"📋 Shikoyatlar / Жалобы: <b>%d</b> (⏳ %d)\n"+
			"💡 Takliflar / Предложения: <b>%d</b> (⏳ %d)\n",
		st.Parents, st.Students, st.Teachers, st.Classes, st.Admins, attendance, avg,
		st.Complaints, st.PendingComplaints, st.Proposals, st.PendingProposals,
	)
}

func superAdminDashboardText(botService *services.BotService) string {
	text := "📊 <b>Umumiy dashboard / Общий дашборд</b>\n\n"
	text += "🌐 <b>Jami / Всего</b>\n" + formatStats(botService.GetBranchStats(0)) + "\n"

	branches, _ := botService.BranchRepo.GetActive()
	for _, b := range branches {
		st := botService.GetBranchStats(b.ID)
		text += fmt.Sprintf("🏫 <b>%s</b>: 👦 %d · 👨‍🏫 %d · 📚 %d", utils.EscapeHTML(b.NameUz), st.Students, st.Teachers, st.Classes)
		if rate := st.AttendanceRate(); rate >= 0 {
			text += fmt.Sprintf(" · ✅ %.0f%%", rate)
		}
		if st.PendingComplaints > 0 {
			text += fmt.Sprintf(" · ⏳ %d", st.PendingComplaints)
		}
		text += "\n"
	}
	return text
}

func superAdminBranchText(botService *services.BotService, branchID int) string {
	branch, err := botService.BranchRepo.GetByID(branchID)
	if err != nil || branch == nil {
		return "❌ Filial topilmadi / Филиал не найден"
	}
	text := fmt.Sprintf("🏫 <b>%s</b>\n\n", utils.EscapeHTML(branch.NameUz))
	text += formatStats(botService.GetBranchStats(branchID))

	classes, _ := botService.ClassRepo.GetAll(branchID)
	if len(classes) > 0 {
		text += "\n<b>Sinflar / Классы:</b>\n"
		for _, c := range classes {
			n, _ := botService.StudentRepo.CountByClass(c.ID)
			status := ""
			if !c.IsActive {
				status = " (nofaol)"
			}
			text += fmt.Sprintf("• %s — %d o'quvchi%s\n", utils.EscapeHTML(c.ClassName), n, status)
		}
	}

	admins, _ := botService.AdminRepo.GetAll(branchID)
	if len(admins) > 0 {
		text += "\n<b>Adminlar / Админы:</b>\n"
		for _, a := range admins {
			text += "• " + a.PhoneNumber + adminLinkMark(a) + "\n"
		}
	}
	return text
}

func superAdminComplaintsText(botService *services.BotService) string {
	items, err := botService.ComplaintService.GetAllComplaintsWithUser(0, 15, 0)
	if err != nil {
		return "❌ " + utils.EscapeHTML(err.Error())
	}
	text := "📋 <b>So'nggi shikoyatlar / Последние жалобы</b>\n\n"
	if len(items) == 0 {
		return text + "Hozircha yo'q / Пока нет"
	}
	for _, c := range items {
		text += fmt.Sprintf("%s #%d · 🏫 %s · %s\n", statusEmoji(c.Status), c.ID,
			utils.EscapeHTML(botService.BranchName(c.BranchID, "uz")), utils.FormatDateTime(c.CreatedAt))
		if c.StudentName != "" {
			text += fmt.Sprintf("   👦 %s (%s)\n", utils.EscapeHTML(c.StudentName), utils.EscapeHTML(c.ClassName))
		}
		text += "   💬 " + utils.EscapeHTML(utils.TruncateText(c.ComplaintText, 80)) + "\n\n"
	}
	return text
}

func superAdminProposalsText(botService *services.BotService) string {
	items, err := botService.ProposalService.GetAllProposalsWithUser(0, 15, 0)
	if err != nil {
		return "❌ " + utils.EscapeHTML(err.Error())
	}
	text := "💡 <b>So'nggi takliflar / Последние предложения</b>\n\n"
	if len(items) == 0 {
		return text + "Hozircha yo'q / Пока нет"
	}
	for _, p := range items {
		text += fmt.Sprintf("%s #%d · 🏫 %s · %s\n", statusEmoji(p.Status), p.ID,
			utils.EscapeHTML(botService.BranchName(p.BranchID, "uz")), utils.FormatDateTime(p.CreatedAt))
		if p.StudentName != "" {
			text += fmt.Sprintf("   👦 %s (%s)\n", utils.EscapeHTML(p.StudentName), utils.EscapeHTML(p.ClassName))
		}
		text += "   💬 " + utils.EscapeHTML(utils.TruncateText(p.ProposalText, 80)) + "\n\n"
	}
	return text
}

func superAdminAdminsView(botService *services.BotService) (string, tgbotapi.InlineKeyboardMarkup) {
	text := "👨‍💼 <b>Filial adminlari / Админы филиалов</b>\n\n"
	var rows [][]tgbotapi.InlineKeyboardButton

	branches, _ := botService.BranchRepo.GetActive()
	for _, b := range branches {
		text += fmt.Sprintf("🏫 <b>%s</b>\n", utils.EscapeHTML(b.NameUz))
		admins, _ := botService.AdminRepo.GetAll(b.ID)
		if len(admins) == 0 {
			text += "   — yo'q / нет\n"
		}
		for _, a := range admins {
			source := ".env"
			if a.Source == models.AdminSourceBot {
				source = "bot"
				rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("🗑 %s (%s)", a.PhoneNumber, b.NameUz), fmt.Sprintf("sa_remove_admin_%d", a.ID),
				)))
			}
			text += fmt.Sprintf("   • %s%s · %s\n", a.PhoneNumber, adminLinkMark(a), source)
		}
		text += "\n"
	}
	text += "✅ — Telegram bog'langan / привязан, ⏳ — hali kirmagan / еще не вошел"

	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("➕ Admin qo'shish / Добавить", "sa_add_admin")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️ Orqaga / Назад", "sa_panel")),
	)
	return text, tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func adminLinkMark(a *models.Admin) string {
	if a.TelegramID != nil {
		return " ✅"
	}
	return " ⏳"
}

func statusEmoji(status string) string {
	switch status {
	case models.StatusReviewed:
		return "✅"
	case models.StatusResolved, models.StatusImplemented:
		return "🏁"
	default:
		return "⏳"
	}
}
