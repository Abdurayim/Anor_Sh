package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/database"
	"parent-bot/internal/i18n"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
)

// sentRequest is one call the bot made to the (fake) Telegram API.
type sentRequest struct {
	method string
	chatID string
	text   string
	markup string
}

// fakeTelegram records everything the bot sends instead of talking to Telegram.
type fakeTelegram struct {
	mu   sync.Mutex
	sent []sentRequest
}

func (f *fakeTelegram) reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = nil
}

// messagesTo returns the text + reply markup of every message sent to a chat since reset.
func (f *fakeTelegram) messagesTo(chatID int64) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []string
	for _, r := range f.sent {
		if r.method == "sendMessage" && r.chatID == fmt.Sprint(chatID) {
			out = append(out, r.text+"\n"+r.markup)
		}
	}
	return strings.Join(out, "\n---\n")
}

// newTestBotWithTelegram is newTestBot with a working (fake) Telegram client, so whole
// conversations can be driven through HandleMessage / HandleCallbackQuery.
func newTestBotWithTelegram(t *testing.T) (*services.BotService, fixture, *fakeTelegram) {
	t.Helper()
	bs, f := newTestBot(t)

	fake := &fakeTelegram{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseMultipartForm(1 << 20)
		method := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
		fake.mu.Lock()
		fake.sent = append(fake.sent, sentRequest{
			method: method,
			chatID: r.FormValue("chat_id"),
			text:   r.FormValue("text"),
			markup: r.FormValue("reply_markup"),
		})
		fake.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if method == "getMe" {
			fmt.Fprint(w, `{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"test","username":"test_bot"}}`)
			return
		}
		fmt.Fprint(w, `{"ok":true,"result":{"message_id":1,"date":1,"chat":{"id":1,"type":"private"}}}`)
	}))
	t.Cleanup(server.Close)

	bot, err := tgbotapi.NewBotAPIWithClient("TEST", server.URL+"/bot%s/%s", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	bs.Bot = bot
	bs.TelegramService = services.NewTelegramService(bot)
	fake.reset()
	return bs, f, fake
}

func textMessage(from int64, text string) *tgbotapi.Message {
	msg := &tgbotapi.Message{From: &tgbotapi.User{ID: from}, Chat: &tgbotapi.Chat{ID: from}, Text: text}
	if strings.HasPrefix(text, "/") {
		msg.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: len(text)}}
	}
	return msg
}

// contactMessage is a contact shared by `from`; owner is the Telegram account the number belongs to.
func contactMessage(from, owner int64, phone string) *tgbotapi.Message {
	return &tgbotapi.Message{
		From:    &tgbotapi.User{ID: from},
		Chat:    &tgbotapi.Chat{ID: from},
		Contact: &tgbotapi.Contact{PhoneNumber: phone, UserID: owner},
	}
}

func press(from int64, data string) *tgbotapi.CallbackQuery {
	return &tgbotapi.CallbackQuery{
		ID:      "cb",
		From:    &tgbotapi.User{ID: from},
		Data:    data,
		Message: &tgbotapi.Message{MessageID: 7, Chat: &tgbotapi.Chat{ID: from}},
	}
}

func stateOf(t *testing.T, bs *services.BotService, telegramID int64) string {
	t.Helper()
	state, _ := bs.StateManager.Get(telegramID)
	// StateManager.Clear stores "registered", which means no conversation is in progress
	if state == nil || state.State == models.StateRegistered {
		return ""
	}
	return state.State
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// An existing teacher row keeps working after the tutor migration.
func TestExistingTeachersStayTeachers(t *testing.T) {
	bs, _ := newTestBot(t)
	teacher, err := bs.TeacherService.GetTeacherByTelegramID(tgTeacherOlmazor)
	if err != nil || teacher == nil {
		t.Fatalf("teacher not found: %v", err)
	}
	if teacher.Role != models.TeacherRoleTeacher || teacher.IsTutor() {
		t.Errorf("role = %q, want teacher", teacher.Role)
	}
}

// The reported bug: someone registers as a parent, is made a teacher afterwards and must
// get the teacher menu without anybody editing the database.
func TestParentBecomesTeacher(t *testing.T) {
	bs, _, tg := newTestBotWithTelegram(t)
	const parentPhone = "+998905555555" // tgParentOlmazor

	// Admin adds the parent's number as a teacher
	must(t, bs.StateManager.Set(tgAdminOlmazor, "awaiting_teacher_phone",
		&models.StateData{TeacherFirstName: "Aziza", TeacherLastName: "Rahimova"}))
	must(t, HandleMessage(bs, textMessage(tgAdminOlmazor, parentPhone)))
	if got := tg.messagesTo(tgParentOlmazor); !strings.Contains(got, "/start") {
		t.Errorf("parent was not told about the new role, got %q", got)
	}
	if got := tg.messagesTo(tgAdminOlmazor); !strings.Contains(got, "admin_teacher_tutor_") {
		t.Errorf("admin did not get the tutor button, got %q", got)
	}
	if teacher, _ := bs.TeacherService.GetTeacherByTelegramID(tgParentOlmazor); teacher != nil {
		t.Fatal("teacher linked without confirming the number")
	}

	// /start asks to confirm the number instead of showing the parent menu
	tg.reset()
	must(t, HandleMessage(bs, textMessage(tgParentOlmazor, "/start")))
	if got := tg.messagesTo(tgParentOlmazor); !strings.Contains(got, "request_contact") {
		t.Fatalf("no contact request on /start, got %q", got)
	}
	if got := stateOf(t, bs, tgParentOlmazor); got != models.StateAwaitingStaffContact {
		t.Fatalf("state = %q", got)
	}

	// Typing the number or forwarding someone's contact proves nothing
	must(t, HandleMessage(bs, textMessage(tgParentOlmazor, parentPhone)))
	must(t, HandleMessage(bs, textMessage(tgParentOlmazor, "/start")))
	must(t, HandleMessage(bs, contactMessage(tgParentOlmazor, tgStranger, parentPhone)))
	if teacher, _ := bs.TeacherService.GetTeacherByTelegramID(tgParentOlmazor); teacher != nil {
		t.Fatal("teacher linked by an unverified number")
	}
	if got := stateOf(t, bs, tgParentOlmazor); got != "" {
		t.Errorf("state not cleared after cancel: %q", got)
	}

	// Own contact links the account
	tg.reset()
	must(t, HandleMessage(bs, textMessage(tgParentOlmazor, "/start")))
	must(t, HandleMessage(bs, contactMessage(tgParentOlmazor, tgParentOlmazor, "998905555555")))
	teacher, _ := bs.TeacherService.GetTeacherByTelegramID(tgParentOlmazor)
	if teacher == nil {
		t.Fatal("teacher not linked after sharing own contact")
	}
	if teacher.BranchID != 1 || teacher.Language != "uz" {
		t.Errorf("teacher = %+v", teacher)
	}
	if got := tg.messagesTo(tgParentOlmazor); !strings.Contains(got, i18n.Get(i18n.BtnMarkAttendance, i18n.LanguageUzbek)) {
		t.Errorf("teacher menu not shown, got %q", got)
	}

	// From now on they are routed as a teacher, and /start no longer asks anything
	tg.reset()
	must(t, HandleMessage(bs, textMessage(tgParentOlmazor, "/start")))
	if got := tg.messagesTo(tgParentOlmazor); !strings.Contains(got, "O'qituvchi paneli") {
		t.Errorf("teacher panel not shown on /start, got %q", got)
	}
}

// Production parents registered before branches existed have no branch; the branch question
// must not block them from confirming a staff role.
func TestParentWithoutBranchBecomesTeacher(t *testing.T) {
	bs, _, tg := newTestBotWithTelegram(t)
	const tgOld int64 = 3100
	_, err := database.DB.Exec("INSERT INTO users (telegram_id, phone_number, language) VALUES (?, '+998907000000', 'ru')", tgOld)
	must(t, err)
	_, err = database.DB.Exec("INSERT INTO teachers (phone_number, first_name, last_name, branch_id) VALUES ('+998907000000', 'Olga', 'Kim', 2)")
	must(t, err)

	must(t, HandleMessage(bs, textMessage(tgOld, "/start")))
	if got := stateOf(t, bs, tgOld); got != models.StateAwaitingStaffContact {
		t.Fatalf("state = %q, want staff contact (messages: %q)", got, tg.messagesTo(tgOld))
	}
	must(t, HandleMessage(bs, contactMessage(tgOld, tgOld, "+998907000000")))
	teacher, _ := bs.TeacherService.GetTeacherByTelegramID(tgOld)
	if teacher == nil {
		t.Fatal("teacher not linked")
	}
	if teacher.Language != "ru" {
		t.Errorf("language = %q, want the parent's language ru", teacher.Language)
	}
}

// A parent who is made a branch admin afterwards.
func TestParentBecomesAdmin(t *testing.T) {
	bs, _, tg := newTestBotWithTelegram(t)
	const parentPhone = "+998906666666" // tgParentSergeli

	must(t, bs.StateManager.Set(tgSuperAdmin, stateSuperAdminAwaitingAdminPhone, &models.StateData{BranchID: 2}))
	must(t, HandleMessage(bs, textMessage(tgSuperAdmin, parentPhone)))
	if got := tg.messagesTo(tgParentSergeli); !strings.Contains(got, "/start") {
		t.Errorf("parent was not told about the admin role, got %q", got)
	}
	if bs.GetAdmin(tgParentSergeli) != nil {
		t.Fatal("admin linked without confirming the number")
	}

	must(t, HandleMessage(bs, textMessage(tgParentSergeli, "/start")))
	must(t, HandleMessage(bs, contactMessage(tgParentSergeli, tgParentSergeli, parentPhone)))
	admin := bs.GetAdmin(tgParentSergeli)
	if admin == nil || admin.BranchID != 2 {
		t.Fatalf("admin = %+v", admin)
	}
}

// /role with a number that has no staff role changes nothing.
func TestRoleCommandForPlainParent(t *testing.T) {
	bs, _, tg := newTestBotWithTelegram(t)

	must(t, HandleMessage(bs, textMessage(tgParentOlmazor, "/start")))
	if got := stateOf(t, bs, tgParentOlmazor); got != "" {
		t.Fatalf("plain parent was asked to confirm a role: state %q", got)
	}

	tg.reset()
	must(t, HandleMessage(bs, textMessage(tgParentOlmazor, "/role")))
	must(t, HandleMessage(bs, contactMessage(tgParentOlmazor, tgParentOlmazor, "+998905555555")))
	if teacher, _ := bs.TeacherService.GetTeacherByTelegramID(tgParentOlmazor); teacher != nil || bs.GetAdmin(tgParentOlmazor) != nil {
		t.Fatal("plain parent got a staff role")
	}
	if got := tg.messagesTo(tgParentOlmazor); !strings.Contains(got, i18n.Get(i18n.BtnMyChildren, i18n.LanguageUzbek)) {
		t.Errorf("parent menu not restored, got %q", got)
	}
}

func TestTutorRole(t *testing.T) {
	bs, f, tg := newTestBotWithTelegram(t)
	teacher, _ := bs.TeacherService.GetTeacherByTelegramID(tgTeacherOlmazor)
	selectOwn := fmt.Sprintf("timetable_select_%d", f.classOlmazor)
	uploadBtn := i18n.Get(i18n.BtnUploadTimetable, i18n.LanguageUzbek)

	// A plain teacher has no timetable rights
	if authorizeCallback(bs, press(tgTeacherOlmazor, selectOwn)) {
		t.Fatal("plain teacher may upload timetables")
	}
	must(t, HandleMessage(bs, textMessage(tgTeacherOlmazor, uploadBtn)))
	if got := tg.messagesTo(tgTeacherOlmazor); strings.Contains(got, "timetable_select_") {
		t.Fatalf("plain teacher got the class list for timetables: %q", got)
	}

	// Only an admin of the teacher's branch can promote
	promote := fmt.Sprintf("admin_teacher_tutor_%d", teacher.ID)
	must(t, HandleCallbackQuery(bs, press(tgAdminSergeli, promote)))
	must(t, HandleCallbackQuery(bs, press(tgTeacherOlmazor, promote)))
	if teacher, _ = bs.TeacherService.GetTeacherByTelegramID(tgTeacherOlmazor); teacher.IsTutor() {
		t.Fatal("promoted by someone who is not the branch admin")
	}
	tg.reset()
	must(t, HandleCallbackQuery(bs, press(tgAdminOlmazor, promote)))
	if teacher, _ = bs.TeacherService.GetTeacherByTelegramID(tgTeacherOlmazor); !teacher.IsTutor() {
		t.Fatal("teacher not promoted to tutor")
	}
	if got := tg.messagesTo(tgTeacherOlmazor); !strings.Contains(got, uploadBtn) {
		t.Errorf("tutor did not receive the tutor menu, got %q", got)
	}

	// Tutor: own branch only
	if !authorizeCallback(bs, press(tgTeacherOlmazor, selectOwn)) {
		t.Error("tutor cannot upload a timetable for own branch")
	}
	if authorizeCallback(bs, press(tgTeacherOlmazor, fmt.Sprintf("timetable_select_%d", f.classSergeli))) {
		t.Error("tutor may upload a timetable for another branch")
	}
	if authorizeCallback(bs, press(tgTeacherOlmazor, "admin_manage_teachers")) {
		t.Error("tutor may open the admin panel")
	}

	// Upload: menu button -> class -> file
	tg.reset()
	must(t, HandleMessage(bs, textMessage(tgTeacherOlmazor, uploadBtn)))
	if got := tg.messagesTo(tgTeacherOlmazor); !strings.Contains(got, selectOwn) || strings.Contains(got, fmt.Sprintf("timetable_select_%d", f.classSergeli)) {
		t.Fatalf("class list for tutor is wrong: %q", got)
	}
	must(t, HandleCallbackQuery(bs, press(tgTeacherOlmazor, selectOwn)))
	file := textMessage(tgTeacherOlmazor, "")
	file.Document = &tgbotapi.Document{FileID: "FILE1", FileName: "jadval.pdf", MimeType: "application/pdf"}
	must(t, HandleMessage(bs, file))

	var timetableID, uploadedBy int
	must(t, database.DB.QueryRow("SELECT id, COALESCE(uploaded_by_teacher_id, 0) FROM timetables WHERE class_id = ?", f.classOlmazor).
		Scan(&timetableID, &uploadedBy))
	if uploadedBy != teacher.ID {
		t.Errorf("uploaded_by_teacher_id = %d, want %d", uploadedBy, teacher.ID)
	}
	if got := stateOf(t, bs, tgTeacherOlmazor); got != "" {
		t.Errorf("state after upload = %q", got)
	}

	// The parent of that class sees it
	timetable, err := bs.TimetableService.GetTimetableByClassName(1, "5-A")
	if err != nil || timetable == nil || timetable.TelegramFileID != "FILE1" {
		t.Errorf("timetable for parents = %+v, %v", timetable, err)
	}

	// List and delete
	tg.reset()
	must(t, HandleMessage(bs, textMessage(tgTeacherOlmazor, i18n.Get(i18n.BtnViewTimetables, i18n.LanguageUzbek))))
	deleteData := fmt.Sprintf("timetable_delete_%d", timetableID)
	if got := tg.messagesTo(tgTeacherOlmazor); !strings.Contains(got, deleteData) || !strings.Contains(got, "teacher_back_to_main") {
		t.Fatalf("timetable list for tutor is wrong: %q", got)
	}
	must(t, HandleCallbackQuery(bs, press(tgAdminSergeli, deleteData))) // other branch: refused
	if exists, _ := bs.TimetableRepo.Exists(f.classOlmazor); !exists {
		t.Fatal("timetable deleted by another branch's admin")
	}
	must(t, HandleCallbackQuery(bs, press(tgTeacherOlmazor, deleteData)))
	if exists, _ := bs.TimetableRepo.Exists(f.classOlmazor); exists {
		t.Fatal("tutor could not delete the timetable")
	}

	// Demoted: rights are gone immediately
	must(t, HandleCallbackQuery(bs, press(tgAdminOlmazor, fmt.Sprintf("admin_teacher_plain_%d", teacher.ID))))
	if authorizeCallback(bs, press(tgTeacherOlmazor, selectOwn)) {
		t.Error("demoted tutor still has timetable rights")
	}

	// Admins keep their own timetable rights
	if !authorizeCallback(bs, press(tgAdminOlmazor, selectOwn)) {
		t.Error("admin lost timetable rights")
	}
	if authorizeCallback(bs, press(tgParentOlmazor, selectOwn)) {
		t.Error("parent may upload timetables")
	}
}

func TestTeacherLanguageSwitch(t *testing.T) {
	bs, _, tg := newTestBotWithTelegram(t)
	langBtn := i18n.Get(i18n.BtnChangeLanguage, i18n.LanguageUzbek)
	if langBtn != i18n.Get(i18n.BtnChangeLanguage, i18n.LanguageRussian) {
		t.Fatal("the language button must read the same in both languages")
	}

	must(t, HandleMessage(bs, textMessage(tgTeacherOlmazor, langBtn)))
	if got := tg.messagesTo(tgTeacherOlmazor); !strings.Contains(got, "teacher_lang_ru") {
		t.Fatalf("no language choice, got %q", got)
	}

	tg.reset()
	must(t, HandleCallbackQuery(bs, press(tgTeacherOlmazor, "teacher_lang_ru")))
	teacher, _ := bs.TeacherService.GetTeacherByTelegramID(tgTeacherOlmazor)
	if teacher.Language != "ru" {
		t.Fatalf("language = %q, want ru", teacher.Language)
	}
	if got := tg.messagesTo(tgTeacherOlmazor); !strings.Contains(got, i18n.Get(i18n.BtnMarkAttendance, i18n.LanguageRussian)) {
		t.Errorf("menu not redrawn in Russian, got %q", got)
	}

	// Parents and strangers cannot use the teacher language buttons
	if authorizeCallback(bs, press(tgParentOlmazor, "teacher_lang_ru")) || authorizeCallback(bs, press(tgStranger, "teacher_lang_uz")) {
		t.Error("non-teacher may use the teacher language buttons")
	}

	must(t, HandleCallbackQuery(bs, press(tgTeacherOlmazor, "teacher_lang_uz")))
	if teacher, _ = bs.TeacherService.GetTeacherByTelegramID(tgTeacherOlmazor); teacher.Language != "uz" {
		t.Errorf("language = %q, want uz", teacher.Language)
	}
}
