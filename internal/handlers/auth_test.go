package handlers

import (
	"fmt"
	"path/filepath"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"parent-bot/internal/config"
	"parent-bot/internal/database"
	"parent-bot/internal/models"
	"parent-bot/internal/services"
)

// Telegram IDs of the fixture accounts.
const (
	tgSuperAdmin     int64 = 1000
	tgAdminOlmazor   int64 = 1001
	tgAdminSergeli   int64 = 1002
	tgTeacherOlmazor int64 = 2001
	tgParentOlmazor  int64 = 3001
	tgParentSergeli  int64 = 3002
	tgStranger       int64 = 9999
)

// fixture holds IDs created in the test database.
type fixture struct {
	classOlmazor, classSergeli     int
	studentOlmazor, studentSergeli int
	otherStudentOlmazor            int
	teacherSergeli                 int
	annOlmazor, annSergeli         int
}

// newTestBot creates a migrated SQLite database with one admin, class, student and parent per
// branch, and a BotService around it that never talks to Telegram.
func newTestBot(t *testing.T) (*services.BotService, fixture) {
	t.Helper()
	if err := database.Connect(&config.DatabaseConfig{Path: filepath.Join(t.TempDir(), "test.db")}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if err := database.Migrate(); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{Admin: config.AdminConfig{
		BranchPhones: map[string][]string{
			config.BranchOlmazor: {"+998901111111"},
			config.BranchSergeli: {"+998902222222"},
		},
		SuperAdminPhone: "+998900000000",
	}}
	bs := services.NewBotServiceWithBot(cfg, database.DB, nil)
	if err := bs.InitializeAdmins(); err != nil {
		t.Fatal(err)
	}
	mustLink := func(phone string, tg int64) {
		if a, err := bs.LinkAdminByContact(phone, tg); err != nil || a == nil {
			t.Fatalf("link admin %s: %v", phone, err)
		}
	}
	mustLink("+998900000000", tgSuperAdmin)
	mustLink("+998901111111", tgAdminOlmazor)
	mustLink("+998902222222", tgAdminSergeli)

	db := database.DB
	exec := func(q string, args ...any) int {
		res, err := db.Exec(q, args...)
		if err != nil {
			t.Fatalf("%s: %v", q, err)
		}
		id, _ := res.LastInsertId()
		return int(id)
	}

	var f fixture
	f.classOlmazor = exec("INSERT INTO classes (branch_id, class_name) VALUES (1, '5-A')")
	f.classSergeli = exec("INSERT INTO classes (branch_id, class_name) VALUES (2, '5-A')") // same name, other branch
	f.studentOlmazor = exec("INSERT INTO students (first_name, last_name, class_id) VALUES ('Aziz', 'Rahimov', ?)", f.classOlmazor)
	f.otherStudentOlmazor = exec("INSERT INTO students (first_name, last_name, class_id) VALUES ('Sardor', 'Aliyev', ?)", f.classOlmazor)
	f.studentSergeli = exec("INSERT INTO students (first_name, last_name, class_id) VALUES ('Ibrohim', 'Toirov', ?)", f.classSergeli)
	exec("INSERT INTO teachers (phone_number, first_name, last_name, telegram_id, branch_id) VALUES ('+998903333333', 'Dilnoza', 'Karimova', ?, 1)", tgTeacherOlmazor)
	f.teacherSergeli = exec("INSERT INTO teachers (phone_number, first_name, last_name, branch_id) VALUES ('+998904444444', 'Nodira', 'Sultonova', 2)")
	pO := exec("INSERT INTO users (telegram_id, phone_number, language, branch_id) VALUES (?, '+998905555555', 'uz', 1)", tgParentOlmazor)
	pS := exec("INSERT INTO users (telegram_id, phone_number, language, branch_id) VALUES (?, '+998906666666', 'ru', 2)", tgParentSergeli)
	exec("INSERT INTO parent_students (parent_id, student_id) VALUES (?, ?)", pO, f.studentOlmazor)
	exec("INSERT INTO parent_students (parent_id, student_id) VALUES (?, ?)", pS, f.studentSergeli)
	f.annOlmazor = exec("INSERT INTO announcements (content, branch_id) VALUES ('Olmazor majlisi', 1)")
	f.annSergeli = exec("INSERT INTO announcements (content, branch_id) VALUES ('Sergeli majlisi', 2)")
	exec("INSERT INTO complaints (user_id, student_id, complaint_text, telegram_file_id, filename) VALUES (?, ?, 'olmazor shikoyat', '', '')", pO, f.studentOlmazor)
	exec("INSERT INTO complaints (user_id, student_id, complaint_text, telegram_file_id, filename) VALUES (?, ?, 'sergeli shikoyat', '', '')", pS, f.studentSergeli)
	return bs, f
}

func callback(from int64, data string) *tgbotapi.CallbackQuery {
	return &tgbotapi.CallbackQuery{From: &tgbotapi.User{ID: from}, Data: data}
}

func TestAuthorizeCallback(t *testing.T) {
	bs, f := newTestBot(t)

	cases := []struct {
		name string
		from int64
		data string
		want bool
	}{
		{"admin views own class", tgAdminOlmazor, fmt.Sprintf("admin_view_class_%d", f.classOlmazor), true},
		{"admin cannot view other branch class", tgAdminOlmazor, fmt.Sprintf("admin_view_class_%d", f.classSergeli), false},
		{"sergeli admin views own class", tgAdminSergeli, fmt.Sprintf("admin_view_class_%d", f.classSergeli), true},
		{"admin cannot delete other branch class", tgAdminSergeli, fmt.Sprintf("class_delete_confirm_%d", f.classOlmazor), false},
		{"admin cannot delete other branch teacher", tgAdminOlmazor, fmt.Sprintf("admin_delete_teacher_%d", f.teacherSergeli), false},
		{"admin can delete own branch teacher", tgAdminSergeli, fmt.Sprintf("admin_delete_teacher_confirm_%d", f.teacherSergeli), true},
		{"admin cannot edit other branch announcement", tgAdminOlmazor, fmt.Sprintf("announcement_delete_%d", f.annSergeli), false},
		{"admin edits own announcement", tgAdminOlmazor, fmt.Sprintf("announcement_edit_%d", f.annOlmazor), true},
		{"parent cannot use admin buttons", tgParentOlmazor, "admin_users", false},
		{"stranger cannot use admin buttons", tgStranger, "admin_stats", false},
		{"teacher manages own branch class", tgTeacherOlmazor, fmt.Sprintf("teacher_manage_class_%d", f.classOlmazor), true},
		{"teacher cannot manage other branch class", tgTeacherOlmazor, fmt.Sprintf("teacher_manage_class_%d", f.classSergeli), false},
		{"teacher cannot mark other branch student", tgTeacherOlmazor, fmt.Sprintf("attendance_toggle_%d_%d", f.classOlmazor, f.studentSergeli), false},
		{"teacher grades own branch student", tgTeacherOlmazor, fmt.Sprintf("test_result_add_student_%d", f.studentOlmazor), true},
		{"parent cannot use teacher buttons", tgParentOlmazor, fmt.Sprintf("view_attendance_class_%d", f.classOlmazor), false},
		{"parent views own child", tgParentOlmazor, fmt.Sprintf("view_child_grades_%d", f.studentOlmazor), true},
		{"parent cannot view unlinked child", tgParentOlmazor, fmt.Sprintf("view_child_grades_%d", f.otherStudentOlmazor), false},
		{"parent cannot view other branch child", tgParentOlmazor, fmt.Sprintf("view_child_attendance_%d", f.studentSergeli), false},
		{"parent cannot complain about unlinked child", tgParentSergeli, fmt.Sprintf("complaint_select_child_%d", f.studentOlmazor), false},
		{"parent picks class in own branch", tgParentOlmazor, fmt.Sprintf("mykids_class_%d", f.classOlmazor), true},
		{"parent cannot pick class in other branch", tgParentOlmazor, fmt.Sprintf("mykids_class_%d", f.classSergeli), false},
		{"parent cannot link other branch student", tgParentSergeli, fmt.Sprintf("mykids_student_%d", f.studentOlmazor), false},
		{"neutral callback allowed", tgStranger, "lang_uz", true},
		{"super admin opens dashboard", tgSuperAdmin, "sa_dashboard", true},
		{"super admin adds admin", tgSuperAdmin, "sa_add_admin_branch_2", true},
		{"branch admin cannot use super admin panel", tgAdminOlmazor, "sa_dashboard", false},
		{"teacher cannot use super admin panel", tgTeacherOlmazor, "sa_add_admin", false},
		{"super admin is read-only for branch data", tgSuperAdmin, "admin_create_class", false},
		{"super admin cannot delete classes", tgSuperAdmin, fmt.Sprintf("class_delete_confirm_%d", f.classOlmazor), false},
		{"super admin cannot add teachers", tgSuperAdmin, "admin_add_teacher", false},
	}
	for _, c := range cases {
		if got := authorizeCallback(bs, callback(c.from, c.data)); got != c.want {
			t.Errorf("%s: authorizeCallback(%d, %q) = %v, want %v", c.name, c.from, c.data, got, c.want)
		}
	}
}

func TestBranchIsolation(t *testing.T) {
	bs, f := newTestBot(t)
	olmazor, sergeli := adminBranchID(bs, tgAdminOlmazor), adminBranchID(bs, tgAdminSergeli)
	if olmazor != 1 || sergeli != 2 {
		t.Fatalf("admin branches = %d, %d; want 1, 2", olmazor, sergeli)
	}
	if b := adminBranchID(bs, tgParentOlmazor); b != noBranch {
		t.Fatalf("parent got admin branch %d", b)
	}

	classes, _ := bs.ClassRepo.GetAll(olmazor)
	if len(classes) != 1 || classes[0].ID != f.classOlmazor {
		t.Errorf("olmazor classes = %+v", classes)
	}
	users, _ := bs.UserService.GetAllUsers(sergeli, 100, 0)
	if len(users) != 1 || users[0].TelegramID != tgParentSergeli {
		t.Errorf("sergeli parents = %+v", users)
	}
	complaints, _ := bs.ComplaintService.GetAllComplaintsWithUser(olmazor, 10, 0)
	if len(complaints) != 1 || complaints[0].ComplaintText != "olmazor shikoyat" || complaints[0].StudentName != "Aziz Rahimov" {
		t.Errorf("olmazor complaints = %+v", complaints)
	}
	if n, _ := bs.StudentService.CountStudents(sergeli); n != 1 {
		t.Errorf("sergeli students = %d, want 1", n)
	}
	if n, _ := bs.TeacherService.CountTeachers(olmazor); n != 1 {
		t.Errorf("olmazor teachers = %d, want 1", n)
	}
	ids, _ := bs.GetAdminTelegramIDs(sergeli)
	if len(ids) != 1 || ids[0] != tgAdminSergeli {
		t.Errorf("sergeli admin ids = %v", ids)
	}
	if bs.GetAdmin(tgStranger) != nil {
		t.Error("stranger is admin")
	}
}

func TestAnnouncementsForParent(t *testing.T) {
	bs, f := newTestBot(t)
	parent, _ := bs.UserService.GetUserByTelegramID(tgParentOlmazor)

	// Class-targeted announcement for a class without the parent's children must stay hidden
	otherClass, err := bs.ClassRepo.Create(1, "6-B")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bs.AnnouncementService.CreateAnnouncement(&models.CreateAnnouncementRequest{
		Content: "6-B uchun", BranchID: 1, ClassIDs: []int{otherClass.ID},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := bs.AnnouncementService.CreateAnnouncement(&models.CreateAnnouncementRequest{
		Content: "5-A uchun", BranchID: 1, ClassIDs: []int{f.classOlmazor},
	}); err != nil {
		t.Fatal(err)
	}

	got, err := bs.AnnouncementService.GetActiveAnnouncementsForParent(parent.ID, parent.BranchID, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	var contents []string
	for _, a := range got {
		contents = append(contents, a.Content)
	}
	want := map[string]bool{"Olmazor majlisi": true, "5-A uchun": true}
	if len(contents) != len(want) {
		t.Fatalf("parent sees %v, want %v", contents, want)
	}
	for _, c := range contents {
		if !want[c] {
			t.Errorf("parent should not see %q", c)
		}
	}
}

func TestAdminSync(t *testing.T) {
	bs, _ := newTestBot(t)
	// Phone removed from .env loses admin rights on next start
	bs.Config.Admin.BranchPhones[config.BranchSergeli] = nil
	if err := bs.InitializeAdmins(); err != nil {
		t.Fatal(err)
	}
	if bs.GetAdmin(tgAdminSergeli) != nil {
		t.Error("removed admin still active")
	}
	if bs.GetAdmin(tgAdminOlmazor) == nil {
		t.Error("remaining admin deactivated")
	}
}

func TestPhoneVerification(t *testing.T) {
	bs, _ := newTestBot(t)
	me := &tgbotapi.User{ID: 42}

	own := &tgbotapi.Message{From: me, Contact: &tgbotapi.Contact{PhoneNumber: "998905555555", UserID: 42}}
	if phone, ok := verifiedContactPhone(own); !ok || phone != "+998905555555" {
		t.Errorf("own contact: %q, %v", phone, ok)
	}
	forwarded := &tgbotapi.Message{From: me, Contact: &tgbotapi.Contact{PhoneNumber: "+998901111111", UserID: 7}}
	if _, ok := verifiedContactPhone(forwarded); ok {
		t.Error("someone else's contact accepted as verified")
	}
	if _, ok := verifiedContactPhone(&tgbotapi.Message{From: me, Text: "+998901111111"}); ok {
		t.Error("typed phone accepted as verified")
	}

	// Typed staff phones must be refused; parent phones are fine
	for phone, want := range map[string]bool{
		"+998901111111": true,  // Olmazor admin
		"+998902222222": true,  // Sergeli admin
		"+998903333333": true,  // teacher
		"+998905555555": false, // parent
		"+998907777777": false, // new number
	} {
		if got := isStaffPhone(bs, phone); got != want {
			t.Errorf("isStaffPhone(%s) = %v, want %v", phone, got, want)
		}
	}
}

func TestSuperAdmin(t *testing.T) {
	bs, _ := newTestBot(t)
	super := bs.GetSuperAdmin(tgSuperAdmin)
	if super == nil {
		t.Fatal("super admin not recognized")
	}
	if bs.GetAdmin(tgSuperAdmin) != nil {
		t.Error("super admin must not act as a branch admin")
	}
	if bs.GetSuperAdmin(tgAdminOlmazor) != nil {
		t.Error("branch admin recognized as super admin")
	}

	// Only the super admin adds admins
	branchAdmin := bs.GetAdmin(tgAdminOlmazor)
	if err := bs.AddBranchAdmin(branchAdmin, "+998907000001", "x", 1); err == nil {
		t.Error("branch admin could add an admin")
	}

	if err := bs.AddBranchAdmin(super, "+998907000001", "Sergeli 2", 2); err != nil {
		t.Fatalf("add admin: %v", err)
	}
	if err := bs.AddBranchAdmin(super, "+998907000001", "again", 1); err == nil {
		t.Error("duplicate admin accepted")
	}
	if err := bs.AddBranchAdmin(super, "+998903333333", "teacher", 1); err == nil {
		t.Error("teacher phone accepted as admin")
	}
	if err := bs.AddBranchAdmin(super, "+998907000002", "x", 99); err == nil {
		t.Error("unknown branch accepted")
	}

	// The new admin links by contact and gets exactly their branch
	if a, err := bs.LinkAdminByContact("+998907000001", 5555); err != nil || a == nil {
		t.Fatalf("link new admin: %v", err)
	}
	if b := adminBranchID(bs, 5555); b != 2 {
		t.Errorf("new admin branch = %d, want 2", b)
	}

	// Admins added in the bot survive the .env sync on restart
	if err := bs.InitializeAdmins(); err != nil {
		t.Fatal(err)
	}
	if bs.GetAdmin(5555) == nil {
		t.Error("bot-added admin removed by .env sync")
	}

	// Max admins per branch (Sergeli has 2 now)
	if err := bs.AddBranchAdmin(super, "+998907000003", "x", 2); err != nil {
		t.Fatal(err)
	}
	if err := bs.AddBranchAdmin(super, "+998907000004", "x", 2); err == nil {
		t.Error("more than 3 admins per branch accepted")
	}

	// Removing: .env admins cannot be removed in the bot, bot-added ones can
	envAdmin := bs.GetAdmin(tgAdminSergeli)
	if err := bs.RemoveBranchAdmin(super, envAdmin.ID); err == nil {
		t.Error(".env admin removed from the bot")
	}
	added := bs.GetAdmin(5555)
	if err := bs.RemoveBranchAdmin(super, added.ID); err != nil {
		t.Fatal(err)
	}
	if bs.GetAdmin(5555) != nil {
		t.Error("removed admin still active")
	}
	// ...and can be added again later
	if err := bs.AddBranchAdmin(super, "+998907000001", "back", 1); err != nil {
		t.Errorf("re-adding removed admin: %v", err)
	}
}

func TestBranchStats(t *testing.T) {
	bs, _ := newTestBot(t)
	all, o, s := bs.GetBranchStats(0), bs.GetBranchStats(1), bs.GetBranchStats(2)
	if o.Students != 2 || s.Students != 1 || all.Students != 3 {
		t.Errorf("students: total %d, olmazor %d, sergeli %d", all.Students, o.Students, s.Students)
	}
	if all.Parents != o.Parents+s.Parents || all.Complaints != o.Complaints+s.Complaints {
		t.Errorf("totals do not add up: %+v / %+v / %+v", all, o, s)
	}
	if all.Admins != 2 { // super admin is not counted
		t.Errorf("admins = %d, want 2", all.Admins)
	}
	if rate := o.AttendanceRate(); rate != -1 {
		t.Errorf("attendance rate without attendance = %v", rate)
	}
}
