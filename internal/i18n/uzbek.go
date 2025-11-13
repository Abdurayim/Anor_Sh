package i18n

var uzbek = map[string]string{
	// Commands
	MsgStart:              "/start - Botni ishga tushirish",
	MsgHelp:               "/help - Yordam",
	MsgRegister:           "/register - Ro'yxatdan o'tish",
	MsgSubmitComplaint:    "/complaint - Shikoyat yuborish",
	MsgSubmitProposal:     "/proposal - Taklif yuborish",
	MsgMyComplaints:       "/my_complaints - Mening shikoyatlarim",
	MsgMyProposals:        "/my_proposals - Mening takliflarim",
	MsgViewTimetable:      "/timetable - Dars jadvali",
	MsgViewAnnouncements:  "/announcements - E'lonlar",
	MsgSettings:           "/settings - Sozlamalar",

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

	// Proposal flow
	MsgRequestProposal: "💡 Iltimos, taklifingizni yozib yuboring.\n\n" +
		"Taklif matni kamida 10 ta belgidan iborat bo'lishi kerak.\n\n" +
		"Aniq va tushunarli yozing.",
	MsgProposalReceived:  "✅ Taklifingiz qabul qilindi.\n\nTasdiqlaysizmi?",
	MsgConfirmProposal:   "📄 Sizning taklifingiz:\n\n%s\n\nYuborilsinmi?",
	MsgProposalSubmitted: "✅ Taklifingiz muvaffaqiyatli yuborildi!\n\n" +
		"Ma'muriyat tez orada ko'rib chiqadi.\n\n" +
		"Taklif hujjat sifatida saqlandi.",
	MsgProposalCancelled: "❌ Taklif bekor qilindi.",

	// Timetable messages
	MsgTimetableNotFound:       "❌ Sizning sinfingiz uchun dars jadvali topilmadi.",
	MsgTimetableUploaded:       "✅ Dars jadvali muvaffaqiyatli yuklandi!",
	MsgSelectClassForTimetable: "📚 Dars jadvali yuklash uchun sinfni tanlang:",
	MsgUploadTimetableFile:     "📎 Iltimos, dars jadvali faylini yuboring (rasm, PDF, Word, Excel).",

	// Announcement messages
	MsgNoAnnouncements:              "📭 Hozircha e'lonlar yo'q.",
	MsgAnnouncementPosted:           "✅ E'lon muvaffaqiyatli e'lon qilindi!",
	MsgRequestAnnouncementTitle:     "📝 Iltimos, e'lon sarlavhasini kiriting (ixtiyoriy, o'tkazib yuborish mumkin):",
	MsgRequestAnnouncementContent:   "📝 Iltimos, e'lon matnini kiriting:",
	MsgRequestAnnouncementFile:      "📎 Iltimos, rasm yuboring (ixtiyoriy, o'tkazib yuborish mumkin):",
	MsgAnnouncementSkipFile:         "O'tkazib yuborish",

	// Admin messages
	MsgAdminPanel:         "👨‍💼 Ma'muriyat paneli",
	MsgUserList:           "👥 Ro'yxatdan o'tgan foydalanuvchilar ro'yxati",
	MsgComplaintList:      "📋 Shikoyatlar ro'yxati",
	MsgProposalList:       "💡 Takliflar ro'yxati",
	MsgAnnouncementsList:  "📢 E'lonlar ro'yxati",
	MsgStats:              "📊 Statistika",
	MsgNewComplaint:       "🔔 Yangi shikoyat keldi!",
	MsgNewProposal:        "🔔 Yangi taklif keldi!",

	// Buttons
	BtnUzbek:             "🇺🇿 O'zbek",
	BtnRussian:           "🇷🇺 Русский",
	BtnSharePhone:        "📱 Telefon raqamni yuborish",
	BtnSubmitComplaint:   "✍️ Shikoyat yuborish",
	BtnSubmitProposal:    "💡 Taklif yuborish",
	BtnMyComplaints:      "📋 Mening shikoyatlarim",
	BtnMyProposals:       "💡 Mening takliflarim",
	BtnViewTimetable:     "📅 Dars jadvali",
	BtnViewAnnouncements: "📢 E'lonlar",
	BtnSettings:          "⚙️ Sozlamalar",
	BtnConfirm:           "✅ Tasdiqlash",
	BtnCancel:            "❌ Bekor qilish",
	BtnBack:              "◀️ Orqaga",
	BtnSkip:              "⏭ O'tkazib yuborish",

	// Admin buttons
	BtnAdminPanel:           "👨‍💼 Ma'muriyat paneli",
	BtnCreateClass:          "➕ Sinf yaratish",
	BtnManageClasses:        "📚 Sinflarni boshqarish",
	BtnDeleteClass:          "🗑 Sinf o'chirish",
	BtnUploadTimetable:      "📅 Dars jadvali yuklash",
	BtnViewTimetables:       "📋 Dars jadvallarini ko'rish",
	BtnPostAnnouncement:     "📢 E'lon chiqarish",
	BtnViewUsers:            "👥 Foydalanuvchilar",
	BtnViewComplaints:       "📋 Shikoyatlar",
	BtnViewProposals:        "💡 Takliflar",
	BtnViewAllAnnouncements: "📢 Barcha e'lonlar",
	BtnViewStats:            "📊 Statistika",
	BtnExport:               "📥 Eksport",
	BtnEdit:                 "✏️ Tahrirlash",
	BtnDelete:               "🗑 O'chirish",

	// Errors
	ErrInvalidPhone:      "❌ Noto'g'ri telefon raqam formati!\n\nTelefon raqam +998 bilan boshlanishi va 9 ta raqamdan iborat bo'lishi kerak.\n\nMisol: +998901234567",
	ErrInvalidName:       "❌ Noto'g'ri ism formati!\n\nIsm faqat harflardan iborat bo'lishi kerak.",
	ErrInvalidClass:      "❌ Noto'g'ri sinf formati!\n\nSinf raqami (1-11) va harfi (A-Z) ko'rsatilishi kerak.\n\nMisol: 9A, 11B",
	ErrInvalidComplaint:  "❌ Shikoyat matni juda qisqa!\n\nKamida 10 ta belgi kiriting.",
	ErrInvalidProposal:   "❌ Taklif matni juda qisqa!\n\nKamida 10 ta belgi kiriting.",
	ErrInvalidFile:       "❌ Noto'g'ri fayl formati!",
	ErrAlreadyRegistered: "❌ Siz allaqachon ro'yxatdan o'tgansiz!",
	ErrNotRegistered:     "❌ Siz ro'yxatdan o'tmagansiz!\n\nIltimos, avval /start buyrug'ini bosing.",
	ErrNotAdmin:          "❌ Sizda ma'muriyat huquqlari yo'q!",
	ErrDatabaseError:     "❌ Xatolik yuz berdi. Iltimos, keyinroq urinib ko'ring.",
	ErrUnknownCommand:    "❌ Noma'lum buyruq. /help ni bosing.",
	ErrTextOnly:          "❌ Iltimos, faqat matn yuboring!\n\nRasm, video, GIF yoki boshqa fayllarni yuborish mumkin emas.",
	ErrWrongInputType:    "❌ Noto'g'ri ma'lumot turi!\n\nIltimos, faqat matn kiriting.",

	// Info
	InfoProcessing:  "⏳ Ishlov berilmoqda...",
	InfoPleaseWait:  "⏳ Iltimos, kuting...",
}
