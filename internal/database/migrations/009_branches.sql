-- Migration 009: School branches (filiallar)
-- Every class, teacher, parent, admin and announcement belongs to one branch.
-- Students, grades, attendance, timetables inherit the branch from their class;
-- complaints/proposals inherit it from the parent.
-- Existing classes, teachers and announcements are assigned to the first branch (Olmazor).
-- Existing parents keep branch_id NULL and choose their branch on their next message.
-- Runs with foreign_keys = OFF (set by the migration runner) because classes is rebuilt.

CREATE TABLE IF NOT EXISTS branches (
    id INTEGER PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name_uz TEXT NOT NULL,
    name_ru TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO branches (id, code, name_uz, name_ru) VALUES
    (1, 'olmazor', 'Olmazor filiali', 'Филиал Олмазор'),
    (2, 'sergeli', 'Sergeli filiali', 'Филиал Сергели');

ALTER TABLE admins ADD COLUMN branch_id INTEGER REFERENCES branches(id);
ALTER TABLE teachers ADD COLUMN branch_id INTEGER REFERENCES branches(id);
ALTER TABLE users ADD COLUMN branch_id INTEGER REFERENCES branches(id);
ALTER TABLE announcements ADD COLUMN branch_id INTEGER REFERENCES branches(id);

UPDATE teachers SET branch_id = 1 WHERE branch_id IS NULL;
UPDATE announcements SET branch_id = 1 WHERE branch_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_admins_branch ON admins(branch_id);
CREATE INDEX IF NOT EXISTS idx_teachers_branch ON teachers(branch_id);
CREATE INDEX IF NOT EXISTS idx_users_branch ON users(branch_id);
CREATE INDEX IF NOT EXISTS idx_announcements_branch ON announcements(branch_id);

-- Views reference classes; drop them before rebuilding the table.
DROP VIEW IF EXISTS v_students_with_class;
DROP VIEW IF EXISTS v_test_results_detailed;
DROP VIEW IF EXISTS v_attendance_detailed;
DROP VIEW IF EXISTS v_complaints_with_user;
DROP VIEW IF EXISTS v_proposals_with_user;
DROP VIEW IF EXISTS v_parent_children;
DROP VIEW IF EXISTS v_students_with_parent;
DROP VIEW IF EXISTS v_teacher_classes;
DROP VIEW IF EXISTS v_test_results_export;
DROP VIEW IF EXISTS v_attendance_export;

-- Rebuild classes: class names are unique per branch, not globally.
CREATE TABLE classes_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    branch_id INTEGER NOT NULL DEFAULT 1 REFERENCES branches(id),
    class_name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(branch_id, class_name)
);

INSERT INTO classes_new (id, branch_id, class_name, is_active, created_at, updated_at)
SELECT id, 1, class_name, is_active, created_at, updated_at FROM classes;

DROP TABLE classes;
ALTER TABLE classes_new RENAME TO classes;

CREATE INDEX idx_classes_active ON classes(is_active);
CREATE INDEX idx_classes_name ON classes(class_name);
CREATE INDEX idx_classes_branch ON classes(branch_id);

-- ============================================================================
-- VIEWS (same as 006/008, plus branch_id)
-- ============================================================================

CREATE VIEW v_students_with_class AS
SELECT s.id, s.first_name, s.last_name, s.class_id, c.class_name, s.is_active, s.created_at, c.branch_id
FROM students s
JOIN classes c ON s.class_id = c.id;

CREATE VIEW v_test_results_detailed AS
SELECT tr.id, tr.student_id, s.first_name, s.last_name, s.class_id, c.class_name,
       tr.subject_name, tr.score, tr.test_date, tr.teacher_id, tr.admin_id, tr.created_at, c.branch_id
FROM test_results tr
JOIN students s ON tr.student_id = s.id
JOIN classes c ON s.class_id = c.id;

CREATE VIEW v_attendance_detailed AS
SELECT a.id, a.student_id, s.first_name, s.last_name, s.class_id, c.class_name,
       a.date, a.status, a.marked_by_teacher_id, a.marked_by_admin_id, a.created_at, c.branch_id
FROM attendance a
JOIN students s ON a.student_id = s.id
JOIN classes c ON s.class_id = c.id;

CREATE VIEW v_complaints_with_user AS
SELECT c.id, c.user_id, c.student_id, c.complaint_text, c.telegram_file_id, c.filename, c.status,
       c.created_at, c.updated_at, u.telegram_id, u.telegram_username, u.phone_number, u.language,
       s.first_name AS student_first_name, s.last_name AS student_last_name, cl.class_name, u.branch_id
FROM complaints c
JOIN users u ON c.user_id = u.id
LEFT JOIN students s ON c.student_id = s.id
LEFT JOIN classes cl ON s.class_id = cl.id;

CREATE VIEW v_proposals_with_user AS
SELECT p.id, p.user_id, p.student_id, p.proposal_text, p.telegram_file_id, p.filename, p.status,
       p.created_at, p.updated_at, u.telegram_id, u.telegram_username, u.phone_number, u.language,
       s.first_name AS student_first_name, s.last_name AS student_last_name, cl.class_name, u.branch_id
FROM proposals p
JOIN users u ON p.user_id = u.id
LEFT JOIN students s ON p.student_id = s.id
LEFT JOIN classes cl ON s.class_id = cl.id;

CREATE VIEW v_parent_children AS
SELECT ps.id, ps.parent_id, u.telegram_id, u.phone_number, ps.student_id,
       s.first_name AS student_first_name, s.last_name AS student_last_name,
       s.class_id, c.class_name, ps.linked_at, c.branch_id
FROM parent_students ps
JOIN users u ON ps.parent_id = u.id
JOIN students s ON ps.student_id = s.id
JOIN classes c ON s.class_id = c.id
WHERE s.is_active = 1;

CREATE VIEW v_students_with_parent AS
SELECT s.id, s.first_name, s.last_name, s.class_id, c.class_name, s.is_active,
       ps.parent_id, u.telegram_id AS parent_telegram_id, u.phone_number AS parent_phone,
       u.telegram_username AS parent_username, c.branch_id
FROM students s
JOIN classes c ON s.class_id = c.id
LEFT JOIN parent_students ps ON s.id = ps.student_id
LEFT JOIN users u ON ps.parent_id = u.id;

CREATE VIEW v_teacher_classes AS
SELECT tc.id, tc.teacher_id, tc.class_id, tc.assigned_at, t.first_name, t.last_name,
       t.phone_number, t.telegram_id, c.class_name, c.is_active, c.branch_id
FROM teacher_classes tc
JOIN teachers t ON tc.teacher_id = t.id
JOIN classes c ON tc.class_id = c.id;

CREATE VIEW v_test_results_export AS
SELECT tr.id, s.first_name || ' ' || s.last_name AS student_name, c.class_name, tr.subject_name,
       tr.score, tr.test_date, COALESCE(t.first_name || ' ' || t.last_name, 'N/A') AS teacher_name,
       tr.created_at, c.branch_id
FROM test_results tr
JOIN students s ON tr.student_id = s.id
JOIN classes c ON s.class_id = c.id
LEFT JOIN teachers t ON tr.teacher_id = t.id;

CREATE VIEW v_attendance_export AS
SELECT a.id, s.first_name || ' ' || s.last_name AS student_name, c.class_name, a.date, a.status,
       COALESCE(t.first_name || ' ' || t.last_name, 'N/A') AS marked_by_teacher, a.created_at, c.branch_id
FROM attendance a
JOIN students s ON a.student_id = s.id
JOIN classes c ON s.class_id = c.id
LEFT JOIN teachers t ON a.marked_by_teacher_id = t.id;
