// Command seed fills the SQLite database with demo data for local testing.
//
// Usage:
//
//	go run ./cmd/seed
//
// It seeds both branches (Olmazor, Sergeli). The seed is idempotent: re-running it does
// not duplicate rows.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"parent-bot/internal/config"
	"parent-bot/internal/database"
)

// Fake Telegram IDs for demo parents. Messages to them will fail, which is expected.
const demoTelegramIDBase = 900000000

type demoStudent struct {
	first, last, class string
}

type demoTeacher struct {
	phone, first, last string
	classes            []string
}

type demoParent struct {
	phone, username, lang string
	children              []string // "First Last"
}

// demoBranch is the demo content of one school branch (branch IDs come from migration 009).
type demoBranch struct {
	id       int
	classes  []string
	teachers []demoTeacher
	students []demoStudent
	parents  []demoParent
}

var branches = []demoBranch{
	{
		id:      1, // Olmazor
		classes: []string{"5-A", "5-B", "6-A", "7-A", "9-B", "11-A"},
		teachers: []demoTeacher{
			{"+998901112233", "Dilnoza", "Karimova", []string{"5-A", "5-B"}},
			{"+998935556677", "Jasur", "Toshmatov", []string{"6-A", "7-A"}},
			{"+998998887766", "Malika", "Yusupova", []string{"9-B", "11-A"}},
		},
		students: []demoStudent{
			{"Aziz", "Rahimov", "5-A"}, {"Madina", "Rahimova", "5-A"}, {"Sardor", "Aliyev", "5-A"},
			{"Nilufar", "Qodirova", "5-A"}, {"Bekzod", "Ergashev", "5-B"}, {"Zarina", "Nazarova", "5-B"},
			{"Javohir", "Sobirov", "5-B"}, {"Shahzoda", "Umarova", "6-A"}, {"Otabek", "Mirzayev", "6-A"},
			{"Kamola", "Hasanova", "6-A"}, {"Doniyor", "Abdullayev", "7-A"}, {"Sevara", "Islomova", "7-A"},
			{"Ulug'bek", "Xolmatov", "7-A"}, {"Laylo", "Saidova", "9-B"}, {"Temur", "Raxmatullayev", "9-B"},
			{"Gulnoza", "Jo'rayeva", "11-A"}, {"Akmal", "Nurmatov", "11-A"}, {"Muxlisa", "Tursunova", "11-A"},
		},
		parents: []demoParent{
			{"+998901000001", "demo_parent_rahimov", "uz", []string{"Aziz Rahimov", "Madina Rahimova"}},
			{"+998911000002", "demo_parent_aliyev", "ru", []string{"Sardor Aliyev"}},
			{"+998931000003", "demo_parent_umarova", "uz", []string{"Shahzoda Umarova", "Laylo Saidova"}},
			{"+998941000004", "demo_parent_abdullayev", "uz", []string{"Doniyor Abdullayev"}},
		},
	},
	{
		id:      2, // Sergeli
		classes: []string{"5-A", "6-B", "8-A", "10-A"},
		teachers: []demoTeacher{
			{"+998901234500", "Nodira", "Sultonova", []string{"5-A", "6-B"}},
			{"+998977654300", "Rustam", "Qosimov", []string{"8-A", "10-A"}},
		},
		students: []demoStudent{
			{"Ibrohim", "Yo'ldoshev", "5-A"}, {"Mohinur", "Karimova", "5-A"}, {"Asadbek", "Toirov", "5-A"},
			{"Dilshoda", "Ahmedova", "6-B"}, {"Jahongir", "Usmonov", "6-B"}, {"Robiya", "Valiyeva", "8-A"},
			{"Sherzod", "Baxtiyorov", "8-A"}, {"Farangiz", "Olimova", "10-A"}, {"Bobur", "Hamidov", "10-A"},
			{"Oysha", "Nurmatova", "10-A"},
		},
		parents: []demoParent{
			{"+998901000005", "demo_parent_yuldoshev", "uz", []string{"Ibrohim Yo'ldoshev"}},
			{"+998911000006", "demo_parent_ahmedova", "ru", []string{"Dilshoda Ahmedova", "Robiya Valiyeva"}},
			{"+998931000007", "demo_parent_olimova", "uz", []string{"Farangiz Olimova"}},
		},
	},
}

var subjects = []string{"Matematika", "Ona tili", "Ingliz tili", "Informatika"}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if err := database.Connect(&cfg.Database); err != nil {
		log.Fatalf("db: %v", err)
	}
	defer database.Close()
	db := database.DB

	if err := database.Migrate(); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	days := lastSchoolDays(10)
	parentIdx := 0
	for _, br := range branches {
		seedBranch(tx, br, days, &parentIdx)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	for _, table := range []string{"branches", "classes", "teachers", "teacher_classes", "students", "users", "parent_students",
		"test_results", "attendance", "announcements", "complaints", "proposals"} {
		var n int
		mustScan(db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&n))
		fmt.Printf("%-16s %d\n", table, n)
	}
	rows, err := db.Query(`SELECT b.name_uz, COUNT(DISTINCT c.id), COUNT(DISTINCT s.id), COUNT(DISTINCT u.id)
		FROM branches b
		LEFT JOIN classes c ON c.branch_id = b.id
		LEFT JOIN students s ON s.class_id = c.id
		LEFT JOIN users u ON u.branch_id = b.id
		GROUP BY b.id ORDER BY b.id`)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var name string
		var classesN, studentsN, parentsN int
		mustScan(rows.Scan(&name, &classesN, &studentsN, &parentsN))
		fmt.Printf("  %-18s sinflar=%d o'quvchilar=%d ota-onalar=%d\n", name, classesN, studentsN, parentsN)
	}
	rows.Close()
	fmt.Println("✓ Demo data seeded")
}

// lastSchoolDays returns the last n weekdays before today, oldest first, at UTC midnight
// (same representation the bot stores after time.Parse("2006-01-02", ...)).
// seedBranch inserts classes, teachers, students, parents, grades, attendance,
// announcements, complaints and proposals of one branch.
func seedBranch(tx *sql.Tx, br demoBranch, days []time.Time, parentIdx *int) {
	adminID := sql.NullInt64{}
	_ = tx.QueryRow("SELECT id FROM admins WHERE branch_id = ? ORDER BY id LIMIT 1", br.id).Scan(&adminID)

	classIDs := map[string]int64{}
	for _, c := range br.classes {
		must(tx.Exec("INSERT OR IGNORE INTO classes (branch_id, class_name) VALUES (?, ?)", br.id, c))
		var id int64
		mustScan(tx.QueryRow("SELECT id FROM classes WHERE branch_id = ? AND class_name = ?", br.id, c).Scan(&id))
		classIDs[c] = id
	}

	teacherIDs := map[string]int64{} // class -> teacher id
	for _, t := range br.teachers {
		must(tx.Exec(`INSERT OR IGNORE INTO teachers (phone_number, first_name, last_name, added_by_admin_id, branch_id)
			VALUES (?, ?, ?, ?, ?)`, t.phone, t.first, t.last, adminID, br.id))
		var id int64
		mustScan(tx.QueryRow("SELECT id FROM teachers WHERE phone_number = ?", t.phone).Scan(&id))
		for _, c := range t.classes {
			must(tx.Exec("INSERT OR IGNORE INTO teacher_classes (teacher_id, class_id) VALUES (?, ?)", id, classIDs[c]))
			teacherIDs[c] = id
		}
	}

	studentIDs := map[string]int64{}
	for _, s := range br.students {
		var id int64
		err := tx.QueryRow("SELECT id FROM students WHERE first_name = ? AND last_name = ? AND class_id = ?",
			s.first, s.last, classIDs[s.class]).Scan(&id)
		if err == sql.ErrNoRows {
			res, err := tx.Exec(`INSERT INTO students (first_name, last_name, class_id, added_by_admin_id)
				VALUES (?, ?, ?, ?)`, s.first, s.last, classIDs[s.class], adminID)
			if err != nil {
				log.Fatal(err)
			}
			id, _ = res.LastInsertId()
		} else if err != nil {
			log.Fatal(err)
		}
		studentIDs[s.first+" "+s.last] = id
	}

	var parentIDs []int64
	for _, p := range br.parents {
		*parentIdx++
		must(tx.Exec(`INSERT OR IGNORE INTO users (telegram_id, telegram_username, phone_number, language, branch_id)
			VALUES (?, ?, ?, ?, ?)`, demoTelegramIDBase+*parentIdx, p.username, p.phone, p.lang, br.id))
		var id int64
		mustScan(tx.QueryRow("SELECT id FROM users WHERE phone_number = ?", p.phone).Scan(&id))
		parentIDs = append(parentIDs, id)
		for _, child := range p.children {
			must(tx.Exec("INSERT OR IGNORE INTO parent_students (parent_id, student_id) VALUES (?, ?)", id, studentIDs[child]))
		}
	}

	// Grades and attendance for the last 10 school days
	for idx, s := range br.students {
		sid := studentIDs[s.first+" "+s.last]
		tid := teacherIDs[s.class]
		for d, day := range days {
			status := "present"
			if (idx+d)%7 == 0 {
				status = "absent"
			}
			must(tx.Exec(`INSERT OR IGNORE INTO attendance (student_id, date, status, marked_by_teacher_id)
				VALUES (?, ?, ?, ?)`, sid, day, status, tid))
		}

		var n int
		mustScan(tx.QueryRow("SELECT COUNT(*) FROM test_results WHERE student_id = ?", sid).Scan(&n))
		if n > 0 {
			continue
		}
		for si, subj := range subjects {
			for k, day := range []time.Time{days[1], days[6]} {
				score := fmt.Sprintf("%d", 60+(idx*7+si*11+k*13+br.id*5)%41) // 60..100
				must(tx.Exec(`INSERT INTO test_results (student_id, subject_name, score, test_date, teacher_id)
					VALUES (?, ?, ?, ?, ?)`, sid, subj, score, day, tid))
			}
		}
	}

	// Announcements, complaints, proposals only on the first seed of the branch
	var annCount int
	mustScan(tx.QueryRow("SELECT COUNT(*) FROM announcements WHERE branch_id = ?", br.id).Scan(&annCount))
	if annCount > 0 || len(parentIDs) < 2 {
		return
	}

	branchName := map[int]string{1: "Olmazor", 2: "Sergeli"}[br.id]
	must(tx.Exec(`INSERT INTO announcements (title, content, admin_id, branch_id) VALUES (?, ?, ?, ?)`,
		"Ota-onalar majlisi", fmt.Sprintf("Hurmatli ota-onalar! 10-oktabr kuni soat 18:00 da %s filialining majlislar zalida umumiy ota-onalar majlisi bo'lib o'tadi.", branchName),
		adminID, br.id))
	must(tx.Exec(`INSERT INTO announcements (title, content, admin_id, branch_id) VALUES (?, ?, ?, ?)`,
		"Kuzgi ta'til", "Kuzgi ta'til 28-oktabrdan 3-noyabrgacha davom etadi. Darslar 4-noyabrdan boshlanadi.", adminID, br.id))
	firstClass := br.classes[0]
	res, err := tx.Exec(`INSERT INTO announcements (title, content, teacher_id, branch_id) VALUES (?, ?, ?, ?)`,
		firstClass+" sinfi uchun ekskursiya", "Juma kuni sinfimiz tarix muzeyiga ekskursiyaga boradi. Yo'l haqi — 20 000 so'm.",
		teacherIDs[firstClass], br.id)
	if err != nil {
		log.Fatal(err)
	}
	annID, _ := res.LastInsertId()
	must(tx.Exec("INSERT INTO announcement_classes (announcement_id, class_id) VALUES (?, ?)", annID, classIDs[firstClass]))

	firstChild := studentIDs[br.parents[0].children[0]]
	secondChild := studentIDs[br.parents[1].children[0]]
	must(tx.Exec(`INSERT INTO complaints (user_id, student_id, complaint_text, telegram_file_id, filename, status) VALUES (?, ?, ?, '', '', 'pending')`,
		parentIDs[0], firstChild, "Oshxonada ovqat sovuq beriladi, iltimos e'tibor qarating."))
	must(tx.Exec(`INSERT INTO complaints (user_id, student_id, complaint_text, telegram_file_id, filename, status) VALUES (?, ?, ?, '', '', 'reviewed')`,
		parentIDs[1], secondChild, "Уроки физкультуры часто отменяются без предупреждения."))
	must(tx.Exec(`INSERT INTO proposals (user_id, student_id, proposal_text, telegram_file_id, filename, status) VALUES (?, ?, ?, '', '', 'pending')`,
		parentIDs[0], firstChild, "Maktabda robototexnika to'garagini ochishni taklif qilaman."))
	must(tx.Exec(`INSERT INTO proposals (user_id, student_id, proposal_text, telegram_file_id, filename, status) VALUES (?, ?, ?, '', '', 'implemented')`,
		parentIDs[1], secondChild, "Kutubxonaga yangi badiiy kitoblar olib kelinsa."))
}

func lastSchoolDays(n int) []time.Time {
	now := time.Now().UTC()
	d := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	var out []time.Time
	for len(out) < n {
		d = d.AddDate(0, 0, -1)
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
			out = append([]time.Time{d}, out...)
		}
	}
	return out
}

func must(_ sql.Result, err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func mustScan(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
