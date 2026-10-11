package i18n

var russian = map[string]string{
	// Commands
	MsgStart:              "/start - Запустить бота",
	MsgHelp:               "/help - Помощь",
	MsgRegister:           "/register - Регистрация",
	MsgSubmitComplaint:    "/complaint - Подать жалобу",
	MsgSubmitProposal:     "/proposal - Подать предложение",
	MsgMyComplaints:       "/my_complaints - Мои жалобы",
	MsgMyProposals:        "/my_proposals - Мои предложения",
	MsgViewTimetable:      "/timetable - Расписание уроков",
	MsgViewAnnouncements:  "/announcements - Объявления",
	MsgSettings:           "/settings - Настройки",

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

	// Proposal flow
	MsgRequestProposal: "💡 Пожалуйста, напишите ваше предложение.\n\n" +
		"Текст предложения должен содержать минимум 10 символов.\n\n" +
		"Пишите четко и понятно.",
	MsgProposalReceived:  "✅ Ваше предложение получено.\n\nПодтверждаете?",
	MsgConfirmProposal:   "📄 Ваше предложение:\n\n%s\n\nОтправить?",
	MsgProposalSubmitted: "✅ Ваше предложение успешно отправлено!\n\n" +
		"Администрация скоро рассмотрит его.\n\n" +
		"Предложение сохранено как документ.",
	MsgProposalCancelled: "❌ Предложение отменено.",

	// Timetable messages
	MsgTimetableNotFound:       "❌ Расписание для вашего класса не найдено.",
	MsgTimetableUploaded:       "✅ Расписание успешно загружено!",
	MsgSelectClassForTimetable: "📚 Выберите класс для загрузки расписания:",
	MsgUploadTimetableFile:     "📎 Пожалуйста, отправьте файл расписания (изображение, PDF, Word, Excel).",

	// Announcement messages
	MsgNoAnnouncements:              "📭 Пока объявлений нет.",
	MsgAnnouncementPosted:           "✅ Объявление успешно опубликовано!",
	MsgRequestAnnouncementTitle:     "📝 Пожалуйста, введите заголовок объявления (необязательно, можно пропустить):",
	MsgRequestAnnouncementContent:   "📝 Пожалуйста, введите текст объявления:",
	MsgRequestAnnouncementFile:      "📎 Пожалуйста, отправьте изображение (необязательно, можно пропустить):",
	MsgAnnouncementSkipFile:         "Пропустить",

	// Admin messages
	MsgAdminPanel:         "👨‍💼 Панель администратора",
	MsgUserList:           "👥 Список зарегистрированных пользователей",
	MsgComplaintList:      "📋 Список жалоб",
	MsgProposalList:       "💡 Список предложений",
	MsgAnnouncementsList:  "📢 Список объявлений",
	MsgStats:              "📊 Статистика",
	MsgNewComplaint:       "🔔 Получена новая жалоба!",
	MsgNewProposal:        "🔔 Получено новое предложение!",

	// Teacher flow
	MsgTeacherWelcome:      "👨‍🏫 Здравствуйте!\n\nДобро пожаловать как учитель!\n\nДля подтверждения вашего номера телефона нажмите кнопку ниже.",
	MsgTeacherRegistered:   "✅ Вы успешно зарегистрированы!\n\nТеперь вы можете управлять своими классами, добавлять учеников, вводить результаты тестов и отмечать посещаемость.",
	MsgTeacherMainMenu:     "📋 Главное меню учителя\n\nВыберите:",
	MsgTeacherPanel:        "👨‍🏫 Панель учителя",

	// Student management
	MsgSelectClass:              "📚 Выберите класс:",
	MsgRequestStudentFirstName:  "👤 Пожалуйста, введите имя ученика:",
	MsgRequestStudentLastName:   "👤 Пожалуйста, введите фамилию ученика:",
	MsgStudentAdded:             "✅ Ученик успешно добавлен!\n\n👤 %s %s\n🎓 Класс: %s",
	MsgStudentList:              "👥 Ученики класса %s:",
	MsgNoStudentsInClass:        "❌ В этом классе пока нет учеников.",
	MsgSelectStudent:            "👤 Выберите ученика:",
	MsgStudentSelected:          "✅ Ученик выбран: %s",

	// Parent multi-child flow
	MsgSelectYourChild:      "👶 Пожалуйста, выберите своего ребенка:\n\nСначала выберите класс, затем ученика.",
	MsgChildLinked:          "✅ Ребенок успешно привязан!\n\n👤 %s %s\n🎓 Класс: %s",
	MsgMaxChildrenReached:   "❌ Вы можете добавить максимум 4 детей.",
	MsgMyChildren:           "👨‍👩‍👧‍👦 Мои дети:",
	MsgCurrentChild:         "✅ Текущий ребенок: %s %s (класс %s)",
	MsgSwitchChild:          "🔄 Выберите для переключения ребенка:",
	MsgChildSwitched:        "✅ Ребенок переключен: %s %s",
	MsgMyKidsMenu:           "👨‍👩‍👧‍👦 Мои дети\n\nВыберите ребенка или добавьте нового:",
	MsgNoChildrenLinked:     "❌ К вам еще не привязаны дети.\n\nНажмите кнопку ниже, чтобы добавить ребенка:",
	MsgChildInfo:            "👤 <b>%s %s</b>\n🎓 Класс: %s",
	MsgAddChildPrompt:       "👶 Чтобы добавить ребенка, сначала выберите класс:",
	MsgChildNotFound:        "❌ Ребенок не найден.",
	MsgChildAlreadyLinked:   "❌ Этот ребенок уже привязан к вам.",
	MsgWaitForStudentAdd:    "⏳ В этом классе пока нет учеников.\n\nКогда учитель или администратор добавит учеников, вы сможете выбрать своего ребенка.",

	// Test results
	MsgRequestSubjectName:  "📖 Введите название предмета:\n\nПример: Математика, Физика, Английский язык",
	MsgRequestTestScore:    "💯 Введите результат теста:\n\nПример: 85/100, 5, A",
	MsgRequestTestDate:     "📅 Введите дату теста:\n\nФормат: YYYY-MM-DD\nПример: 2025-12-01",
	MsgTestResultAdded:     "✅ Результат теста добавлен!\n\n👤 Ученик: %s\n📖 Предмет: %s\n💯 Результат: %s\n📅 Дата: %s",
	MsgTestResultUpdated:   "✅ Результат теста обновлен!",
	MsgMyTestResults:       "📊 Мои результаты тестов:",
	MsgNoTestResults:       "❌ Результатов тестов пока нет.",
	MsgClassTestResults:    "📊 Результаты тестов класса %s:",

	// Attendance
	MsgMarkAttendance:        "✅ Отметить посещаемость",
	MsgSelectAbsentStudents:  "❌ Выберите отсутствующих учеников:\n\nКогда закончите, нажмите 'Завершить'.\n\nЕсли никто не отсутствует, все будут отмечены как 'Присутствует'.",
	MsgAttendanceMarked:      "✅ Посещаемость отмечена!\n\n📅 Дата: %s\n🎓 Класс: %s\n✅ Присутствуют: %d\n❌ Отсутствуют: %d",
	MsgMyAttendance:          "📋 Моя посещаемость (последние 30 дней):",
	MsgNoAttendanceRecords:   "❌ Записей о посещаемости не найдено.",
	MsgAttendanceTaken:       "✅ Посещаемость отмечена",
	MsgAttendanceNotTaken:    "❌ Посещаемость для класса %s еще не отмечена",
	MsgAttendancePresent:     "✅ Присутствует",
	MsgAttendanceAbsent:      "❌ Отсутствует",

	// Announcement multi-class
	MsgSelectTargetClasses:   "🎯 Выберите классы для объявления:\n\nВы можете выбрать несколько классов.",
	MsgClassesSelected:       "✅ Выбрано классов: %d",
	MsgAnnouncementBroadcast: "📢 Объявление отправлено в %d классов!",

	// Class deletion
	MsgClassDeleted:          "🗑 Класс удален: %s",
	MsgClassDeletedReselect:  "⚠️ Класс вашего ребенка (%s) был удален!\n\nПожалуйста, выберите новый класс.",
	MsgPleaseSelectNewClass:  "📚 Пожалуйста, выберите новый класс для вашего ребенка:",

	// Buttons
	BtnUzbek:             "🇺🇿 O'zbek",
	BtnRussian:           "🇷🇺 Русский",
	BtnSharePhone:        "📱 Отправить номер телефона",
	BtnSubmitComplaint:   "✍️ Подать жалобу",
	BtnSubmitProposal:    "💡 Подать предложение",
	BtnMyComplaints:      "📋 Мои жалобы",
	BtnMyProposals:       "💡 Мои предложения",
	BtnViewTimetable:     "📅 Расписание уроков",
	BtnViewAnnouncements: "📢 Объявления",
	BtnSettings:          "⚙️ Настройки",
	BtnConfirm:           "✅ Подтвердить",
	BtnCancel:            "❌ Отменить",
	BtnBack:              "◀️ Назад",
	BtnSkip:              "⏭ Пропустить",

	// Admin buttons
	BtnAdminPanel:           "👨‍💼 Панель администратора",
	BtnCreateClass:          "➕ Создать класс",
	BtnManageClasses:        "📚 Управление классами",
	BtnDeleteClass:          "🗑 Удалить класс",
	BtnUploadTimetable:      "📅 Загрузить расписание",
	BtnViewTimetables:       "📋 Просмотр расписаний",
	BtnPostAnnouncement:     "📢 Опубликовать объявление",
	BtnViewUsers:            "👥 Пользователи",
	BtnViewComplaints:       "📋 Жалобы",
	BtnViewProposals:        "💡 Предложения",
	BtnViewAllAnnouncements: "📢 Все объявления",
	BtnViewStats:            "📊 Статистика",
	BtnExport:               "📥 Экспорт",
	BtnEdit:                 "✏️ Редактировать",
	BtnDelete:               "🗑 Удалить",
	BtnAddTeacher:           "👨‍🏫 Добавить учителя",
	BtnManageTeachers:       "👥 Управление учителями",
	BtnAddStudent:           "👤 Добавить ученика",
	BtnManageStudents:       "👥 Управление учениками",
	BtnExportTestResults:    "📊 Экспорт результатов",
	BtnExportAttendance:     "📋 Экспорт посещаемости",

	// Teacher buttons
	BtnTeacherPanel:      "👨‍🏫 Панель учителя",
	BtnMyClasses:         "📚 Мои классы",
	BtnAddTestResult:     "📊 Добавить результат теста",
	BtnMarkAttendance:    "✅ Отметить посещаемость",
	BtnViewClassStudents: "👥 Ученики класса",
	BtnChangeLanguage:    "🌐 Til / Язык",

	// Parent buttons
	BtnMyTestResults:        "📊 Мои результаты",
	BtnMyAttendance:         "📋 Моя посещаемость",
	BtnMyChildren:           "👨‍👩‍👧‍👦 Мои дети",
	BtnAddAnotherChild:      "➕ Добавить другого ребенка",
	BtnSwitchChild:          "🔄 Переключить ребенка",
	BtnFinishAttendance:     "✅ Завершить",
	BtnViewChildAttendance:  "📋 Посещаемость",
	BtnViewChildTestResults: "📊 Оценки",

	// Errors
	ErrInvalidPhone:      "❌ Неверный формат номера телефона!\n\nНомер должен начинаться с +998 и содержать 9 цифр.\n\nПример: +998901234567",
	ErrInvalidName:       "❌ Неверный формат имени!\n\nИмя должно содержать только буквы.",
	ErrInvalidClass:      "❌ Неверный формат класса!\n\nНеобходимо указать номер класса (1-11) и букву (A-Z).\n\nПример: 9A, 11B",
	ErrInvalidComplaint:  "❌ Текст жалобы слишком короткий!\n\nВведите минимум 10 символов.",
	ErrInvalidProposal:   "❌ Текст предложения слишком короткий!\n\nВведите минимум 10 символов.",
	ErrInvalidFile:       "❌ Неверный формат файла!",
	ErrAlreadyRegistered: "❌ Вы уже зарегистрированы!",
	ErrNotRegistered:     "❌ Вы не зарегистрированы!\n\nПожалуйста, сначала нажмите /start.",
	ErrNotAdmin:          "❌ У вас нет прав администратора!",
	ErrDatabaseError:     "❌ Произошла ошибка. Пожалуйста, попробуйте позже.",
	ErrUnknownCommand:    "❌ Неизвестная команда. Нажмите /help.",
	ErrTextOnly:          "❌ Пожалуйста, отправьте только текст!\n\nНельзя отправлять изображения, видео, GIF или другие файлы.",
	ErrWrongInputType:    "❌ Неправильный тип данных!\n\nПожалуйста, введите только текст.",

	// Info
	InfoProcessing:  "⏳ Обрабатывается...",
	InfoPleaseWait:  "⏳ Пожалуйста, подождите...",
}
