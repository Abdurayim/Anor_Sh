package i18n

var uzbek = map[string]string{
	// Commands
	MsgStart:           "/start - Botni ishga tushirish",
	MsgHelp:            "/help - Yordam",
	MsgRegister:        "/register - Ro'yxatdan o'tish",
	MsgSubmitComplaint: "/complaint - Shikoyat yuborish",
	MsgMyComplaints:    "/my_complaints - Mening shikoyatlarim",
	MsgSettings:        "/settings - Sozlamalar",

	// Registration flow
	MsgWelcome: "🙌 Assalomu aleykum!\n\nMaktab ota-onalari shikoyatlari botiga xush kelibsiz!\n\nBu bot orqali siz maktab bilan bog'liq shikoyatlaringizni rasmiy ravishda yubora olasiz.",

	MsgChooseLanguage: "Iltimos, tilni tanlang:\n\nПожалуйста, выберите язык:",

	MsgLanguageSelected: "✅ Til tanlandi: O'zbek\n\nDavom etish uchun ro'yxatdan o'ting.",

	MsgRequestPhone: "📱 Iltimos, telefon raqamingizni yuboring.\n\nTelefon raqam +998 bilan boshlanishi kerak.\n\nMisol: +998901234567\n\nYoki quyidagi tugma orqali raqamingizni yuboring 👇",

	MsgPhoneReceived: "✅ Telefon raqam qabul qilindi: %s",

	MsgRequestChildName: "👶 Iltimos, farzandingizning ismini kiriting.\n\nMisol: Akmal Rahimov",

	MsgChildNameReceived: "✅ Farzand ismi qabul qilindi: %s",

	MsgRequestChildClass: "🎓 Iltimos, farzandingiz o'qiyotgan sinfni kiriting.\n\nMisol: 9A, 11B\n\nSinf raqami (1-11) va harfi (A-Z) ko'rsatilishi kerak.",

	MsgRegistrationComplete: "✅ Ro'yxatdan o'tish muvaffaqiyatli yakunlandi!\n\n" +
		"👤 Farzand: %s\n" +
		"🎓 Sinf: %s\n" +
		"📱 Telefon: %s\n\n" +
		"Endi siz shikoyat yuborishingiz mumkin.",

	// Complaint flow
	MsgMainMenu: "📋 Asosiy menyu\n\nTanlang:",

	MsgRequestComplaint: "✍️ Iltimos, shikoyatingizni yozib yuboring.\n\n" +
		"Shikoyat matni kamida 10 ta belgidan iborat bo'lishi kerak.\n\n" +
		"Aniq va tushunarli yozing.",

	MsgComplaintReceived: "✅ Shikoyatingiz qabul qilindi.\n\nTasdiqlaysizmi?",

	MsgConfirmComplaint: "📄 Sizning shikoyatingiz:\n\n%s\n\nYuborilsinmi?",

	MsgComplaintSubmitted: "✅ Shikoyatingiz muvaffaqiyatli yuborildi!\n\n" +
		"Ma'muriyat tez orada ko'rib chiqadi.\n\n" +
		"Shikoyat hujjat sifatida saqlandi.",

	MsgComplaintCancelled: "❌ Shikoyat bekor qilindi.",

	// Admin messages
	MsgAdminPanel:      "👨‍💼 Ma'muriyat paneli",
	MsgUserList:        "👥 Ro'yxatdan o'tgan foydalanuvchilar ro'yxati",
	MsgComplaintList:   "📋 Shikoyatlar ro'yxati",
	MsgStats:           "📊 Statistika",
	MsgNewComplaint:    "🔔 Yangi shikoyat keldi!",

	// Buttons
	BtnUzbek:           "🇺🇿 O'zbek",
	BtnRussian:         "🇷🇺 Русский",
	BtnSharePhone:      "📱 Telefon raqamni yuborish",
	BtnSubmitComplaint: "✍️ Shikoyat yuborish",
	BtnMyComplaints:    "📋 Mening shikoyatlarim",
	BtnSettings:        "⚙️ Sozlamalar",
	BtnConfirm:         "✅ Tasdiqlash",
	BtnCancel:          "❌ Bekor qilish",
	BtnBack:            "◀️ Orqaga",

	// Admin buttons
	BtnAdminPanel:      "👨‍💼 Ma'muriyat paneli",
	BtnCreateClass:     "➕ Sinf yaratish",
	BtnManageClasses:   "📚 Sinflarni boshqarish",
	BtnViewUsers:       "👥 Foydalanuvchilar",
	BtnViewComplaints:  "📋 Shikoyatlar",
	BtnViewStats:       "📊 Statistika",
	BtnExport:          "📥 Eksport",

	// Errors
	ErrInvalidPhone:      "❌ Noto'g'ri telefon raqam formati!\n\nTelefon raqam +998 bilan boshlanishi va 9 ta raqamdan iborat bo'lishi kerak.\n\nMisol: +998901234567",
	ErrInvalidName:       "❌ Noto'g'ri ism formati!\n\nIsm faqat harflardan iborat bo'lishi kerak.",
	ErrInvalidClass:      "❌ Noto'g'ri sinf formati!\n\nSinf raqami (1-11) va harfi (A-Z) ko'rsatilishi kerak.\n\nMisol: 9A, 11B",
	ErrInvalidComplaint:  "❌ Shikoyat matni juda qisqa!\n\nKamida 10 ta belgi kiriting.",
	ErrAlreadyRegistered: "❌ Siz allaqachon ro'yxatdan o'tgansiz!",
	ErrNotRegistered:     "❌ Siz ro'yxatdan o'tmagansiz!\n\nIltimos, avval /start buyrug'ini bosing.",
	ErrDatabaseError:     "❌ Xatolik yuz berdi. Iltimos, keyinroq urinib ko'ring.",
	ErrUnknownCommand:    "❌ Noma'lum buyruq. /help ni bosing.",

	// Info
	InfoProcessing:  "⏳ Ishlov berilmoqda...",
	InfoPleaseWait:  "⏳ Iltimos, kuting...",
}
