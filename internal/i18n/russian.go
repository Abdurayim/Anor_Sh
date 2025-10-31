package i18n

var russian = map[string]string{
	// Commands
	MsgStart:           "/start - Запустить бота",
	MsgHelp:            "/help - Помощь",
	MsgRegister:        "/register - Регистрация",
	MsgSubmitComplaint: "/complaint - Подать жалобу",
	MsgMyComplaints:    "/my_complaints - Мои жалобы",
	MsgSettings:        "/settings - Настройки",

	// Registration flow
	MsgWelcome: "🙌 Здравствуйте!\n\nДобро пожаловать в бот жалоб родителей школьников!\n\nЧерез этот бот вы можете официально подавать жалобы, связанные со школой.",

	MsgChooseLanguage: "Пожалуйста, выберите язык:\n\nIltimos, tilni tanlang:",

	MsgLanguageSelected: "✅ Язык выбран: Русский\n\nДля продолжения пройдите регистрацию.",

	MsgRequestPhone: "📱 Пожалуйста, отправьте ваш номер телефона.\n\nНомер должен начинаться с +998.\n\nПример: +998901234567\n\nИли отправьте через кнопку ниже 👇",

	MsgPhoneReceived: "✅ Номер телефона получен: %s",

	MsgRequestChildName: "👶 Пожалуйста, введите имя вашего ребенка.\n\nПример: Акмал Рахимов",

	MsgChildNameReceived: "✅ Имя ребенка получено: %s",

	MsgRequestChildClass: "🎓 Пожалуйста, введите класс, в котором учится ваш ребенок.\n\nПример: 9A, 11B\n\nНеобходимо указать номер класса (1-11) и букву (A-Z).",

	MsgRegistrationComplete: "✅ Регистрация успешно завершена!\n\n" +
		"👤 Ребенок: %s\n" +
		"🎓 Класс: %s\n" +
		"📱 Телефон: %s\n\n" +
		"Теперь вы можете подавать жалобы.",

	// Complaint flow
	MsgMainMenu: "📋 Главное меню\n\nВыберите:",

	MsgRequestComplaint: "✍️ Пожалуйста, напишите вашу жалобу.\n\n" +
		"Текст жалобы должен содержать минимум 10 символов.\n\n" +
		"Пишите четко и понятно.",

	MsgComplaintReceived: "✅ Ваша жалоба получена.\n\nПодтверждаете?",

	MsgConfirmComplaint: "📄 Ваша жалоба:\n\n%s\n\nОтправить?",

	MsgComplaintSubmitted: "✅ Ваша жалоба успешно отправлена!\n\n" +
		"Администрация скоро рассмотрит её.\n\n" +
		"Жалоба сохранена как документ.",

	MsgComplaintCancelled: "❌ Жалоба отменена.",

	// Admin messages
	MsgAdminPanel:      "👨‍💼 Панель администратора",
	MsgUserList:        "👥 Список зарегистрированных пользователей",
	MsgComplaintList:   "📋 Список жалоб",
	MsgStats:           "📊 Статистика",
	MsgNewComplaint:    "🔔 Получена новая жалоба!",

	// Buttons
	BtnUzbek:           "🇺🇿 O'zbek",
	BtnRussian:         "🇷🇺 Русский",
	BtnSharePhone:      "📱 Отправить номер телефона",
	BtnSubmitComplaint: "✍️ Подать жалобу",
	BtnMyComplaints:    "📋 Мои жалобы",
	BtnSettings:        "⚙️ Настройки",
	BtnConfirm:         "✅ Подтвердить",
	BtnCancel:          "❌ Отменить",
	BtnBack:            "◀️ Назад",

	// Admin buttons
	BtnAdminPanel:      "👨‍💼 Панель администратора",
	BtnCreateClass:     "➕ Создать класс",
	BtnManageClasses:   "📚 Управление классами",
	BtnViewUsers:       "👥 Пользователи",
	BtnViewComplaints:  "📋 Жалобы",
	BtnViewStats:       "📊 Статистика",
	BtnExport:          "📥 Экспорт",

	// Errors
	ErrInvalidPhone:      "❌ Неверный формат номера телефона!\n\nНомер должен начинаться с +998 и содержать 9 цифр.\n\nПример: +998901234567",
	ErrInvalidName:       "❌ Неверный формат имени!\n\nИмя должно содержать только буквы.",
	ErrInvalidClass:      "❌ Неверный формат класса!\n\nНеобходимо указать номер класса (1-11) и букву (A-Z).\n\nПример: 9A, 11B",
	ErrInvalidComplaint:  "❌ Текст жалобы слишком короткий!\n\nВведите минимум 10 символов.",
	ErrAlreadyRegistered: "❌ Вы уже зарегистрированы!",
	ErrNotRegistered:     "❌ Вы не зарегистрированы!\n\nПожалуйста, сначала нажмите /start.",
	ErrDatabaseError:     "❌ Произошла ошибка. Пожалуйста, попробуйте позже.",
	ErrUnknownCommand:    "❌ Неизвестная команда. Нажмите /help.",

	// Info
	InfoProcessing:  "⏳ Обрабатывается...",
	InfoPleaseWait:  "⏳ Пожалуйста, подождите...",
}
