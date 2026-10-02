package handlers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
	"parent-bot/internal/utils"
)

// HandleAdminLinkCommand handles /admin_link command for admins to link their telegram account
func HandleAdminLinkCommand(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	text := "🔑 <b>Admin Linking / Связка администратора</b>\n\n"
	text += "⚠️ <b>XAVFSIZLIK / БЕЗОПАСНОСТЬ:</b>\n"
	text += "Telefon raqamingizni xavfsiz tarzda ulashish uchun quyidagi tugmani bosing.\n"
	text += "Для безопасной передачи номера телефона нажмите кнопку ниже.\n\n"
	text += "📱 Tugma Telegram'da ro'yxatdan o'tgan telefon raqamingizni avtomatik yuboradi.\n"
	text += "📱 Кнопка автоматически отправит ваш номер телефона, зарегистрированный в Telegram.\n\n"
	text += "⚠️ MUHIM / ВАЖНО:\n"
	text += "Faqat .env faylida ko'rsatilgan admin raqamlari qabul qilinadi.\n"
	text += "Принимаются только номера администраторов, указанные в файле .env."

	// Set state to awaiting phone for admin link
	err := botService.StateManager.Set(telegramID, models.StateAwaitingAdminPhone, &models.StateData{})
	if err != nil {
		return err
	}

	// Create keyboard with phone sharing button
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("📱 Telefon raqamni ulashish / Поделиться номером"),
		),
	)

	return botService.TelegramService.SendMessage(chatID, text, keyboard)
}

// HandleAdminLinkPhone handles phone number for admin linking
func HandleAdminLinkPhone(botService *services.BotService, message *tgbotapi.Message) error {
	telegramID := message.From.ID
	chatID := message.Chat.ID

	validPhone, ok := verifiedContactPhone(message)
	if !ok {
		text := "❌ Iltimos, pastdagi tugma orqali <b>o'z</b> raqamingizni ulashing.\n" +
			"❌ Пожалуйста, поделитесь <b>своим</b> номером кнопкой ниже."
		_ = botService.StateManager.Clear(telegramID)
		return botService.TelegramService.SendMessage(chatID, text, utils.RemoveKeyboard())
	}

	admin, err := botService.LinkAdminByContact(validPhone, telegramID)
	if err != nil || admin == nil {
		text := "❌ Bu raqam admin sifatida ro'yxatga olinmagan / Этот номер не зарегистрирован как администратор\n\n"
		text += fmt.Sprintf("Sizning raqamingiz: %s\n", validPhone)
		_ = botService.StateManager.Clear(telegramID)
		return botService.TelegramService.SendMessage(chatID, text, utils.RemoveKeyboard())
	}

	// Clear state
	_ = botService.StateManager.Clear(telegramID)

	if admin.IsSuperAdmin() {
		text := "👑 Siz <b>super admin</b> sifatida bog'landingiz.\n👑 Вы привязаны как <b>супер-админ</b>."
		return botService.TelegramService.SendMessage(chatID, text, makeSuperAdminReplyKeyboard())
	}

	// Send success message with keyboard removed
	text := "✅ <b>Muvaffaqiyatli!</b> / <b>Успешно!</b>\n\n"
	text += "Sizning Telegram akkauntingiz admin sifatida bog'landi.\n"
	text += "Ваш Telegram аккаунт привязан как администратор.\n\n"
	text += fmt.Sprintf("📱 Telefon: %s\n", validPhone)
	text += fmt.Sprintf("🏫 %s\n\n", botService.BranchName(admin.BranchID, "uz"))
	text += "⚠️ <b>MUHIM / ВАЖНО:</b>\n"
	text += "Admin tugmasini ko'rish uchun <b>/start</b> buyrug'ini yuboring!\n"
	text += "Чтобы увидеть кнопку администратора, отправьте команду <b>/start</b>!"

	return botService.TelegramService.SendMessage(chatID, text, utils.RemoveKeyboard())

}
