-- ============================================================================
-- MIGRATION 006: COMPLETE SCHEMA (Fresh Start)
-- ============================================================================
-- This migration creates all tables from scratch with the final schema.
-- Use this for new deployments.
-- ============================================================================

-- Drop all existing tables for fresh start
DROP TABLE IF EXISTS user_states;
DROP TABLE IF EXISTS announcement_classes;
DROP TABLE IF EXISTS attendance;
DROP TABLE IF EXISTS test_results;
DROP TABLE IF EXISTS teacher_classes;
DROP TABLE IF EXISTS parent_students;
DROP TABLE IF EXISTS timetables;
DROP TABLE IF EXISTS announcements;
DROP TABLE IF EXISTS proposals;
DROP TABLE IF EXISTS complaints;
DROP TABLE IF EXISTS students;
DROP TABLE IF EXISTS teachers;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS admins;
DROP TABLE IF EXISTS classes;

-- Drop any leftover _new tables from failed migrations
DROP TABLE IF EXISTS users_new;
DROP TABLE IF EXISTS students_new;
DROP TABLE IF EXISTS parent_students_new;
DROP TABLE IF EXISTS test_results_new;
DROP TABLE IF EXISTS attendance_new;
DROP TABLE IF EXISTS timetables_new;
DROP TABLE IF EXISTS teacher_classes_new;
DROP TABLE IF EXISTS announcement_classes_new;

-- Drop views
DROP VIEW IF EXISTS v_complaints_with_user;
DROP VIEW IF EXISTS v_proposals_with_user;
DROP VIEW IF EXISTS v_parent_children;
DROP VIEW IF EXISTS v_students_with_parent;
DROP VIEW IF EXISTS v_teacher_classes;
DROP VIEW IF EXISTS v_test_results_export;
DROP VIEW IF EXISTS v_attendance_export;
DROP VIEW IF EXISTS v_attendance_detailed;
DROP VIEW IF EXISTS v_students_with_class;
DROP VIEW IF EXISTS v_test_results_detailed;

-- ============================================================================
-- CORE ENTITIES
-- ============================================================================

-- Classes Table
CREATE TABLE classes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_name TEXT NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_classes_active ON classes(is_active);
CREATE INDEX idx_classes_name ON classes(class_name);

-- Admins Table
CREATE TABLE admins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    phone_number TEXT NOT NULL UNIQUE,
    name TEXT,
    telegram_id INTEGER UNIQUE,
    telegram_username TEXT,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    added_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_admins_phone ON admins(phone_number);
CREATE INDEX idx_admins_telegram ON admins(telegram_id);

-- Teachers Table
CREATE TABLE teachers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    phone_number TEXT NOT NULL UNIQUE,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    telegram_id INTEGER UNIQUE,
    telegram_username TEXT,
    language TEXT NOT NULL DEFAULT 'uz' CHECK(language IN ('uz', 'ru')),
    is_active BOOLEAN NOT NULL DEFAULT 1,
    added_by_admin_id INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (added_by_admin_id) REFERENCES admins(id) ON DELETE SET NULL
);

CREATE INDEX idx_teachers_phone ON teachers(phone_number);
CREATE INDEX idx_teachers_telegram ON teachers(telegram_id);
CREATE INDEX idx_teachers_active ON teachers(is_active);

-- Students Table
CREATE TABLE students (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    class_id INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT 1,
    added_by_admin_id INTEGER,
    added_by_teacher_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    FOREIGN KEY (added_by_admin_id) REFERENCES admins(id) ON DELETE SET NULL,
    FOREIGN KEY (added_by_teacher_id) REFERENCES teachers(id) ON DELETE SET NULL
);

CREATE INDEX idx_students_class ON students(class_id);
CREATE INDEX idx_students_active ON students(is_active);
CREATE INDEX idx_students_name ON students(last_name, first_name);

-- Users Table (Parents)
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    telegram_id INTEGER NOT NULL UNIQUE,
    telegram_username TEXT,
    phone_number TEXT NOT NULL UNIQUE,
    language TEXT NOT NULL DEFAULT 'uz' CHECK(language IN ('uz', 'ru')),
    registered_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_telegram ON users(telegram_id);
CREATE INDEX idx_users_phone ON users(phone_number);
CREATE INDEX idx_users_registered ON users(registered_at);

-- Announcements Table
CREATE TABLE announcements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT,
    content TEXT NOT NULL,
    telegram_file_id TEXT,
    filename TEXT,
    file_type TEXT,
    admin_id INTEGER,
    teacher_id INTEGER,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (admin_id) REFERENCES admins(id) ON DELETE SET NULL,
    FOREIGN KEY (teacher_id) REFERENCES teachers(id) ON DELETE SET NULL
);

CREATE INDEX idx_announcements_admin ON announcements(admin_id);
CREATE INDEX idx_announcements_teacher ON announcements(teacher_id);
CREATE INDEX idx_announcements_created ON announcements(created_at);
CREATE INDEX idx_announcements_active ON announcements(is_active);

-- Timetables Table
CREATE TABLE timetables (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL,
    telegram_file_id TEXT NOT NULL,
    filename TEXT,
    file_type TEXT CHECK (file_type IN ('image', 'document')),
    mime_type TEXT,
    uploaded_by_admin_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    FOREIGN KEY (uploaded_by_admin_id) REFERENCES admins(id) ON DELETE SET NULL
);

CREATE INDEX idx_timetables_class ON timetables(class_id);
CREATE INDEX idx_timetables_created ON timetables(created_at);

-- Complaints Table
CREATE TABLE complaints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    student_id INTEGER,
    complaint_text TEXT NOT NULL,
    telegram_file_id TEXT,
    filename TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'reviewed', 'resolved')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE SET NULL
);

CREATE INDEX idx_complaints_user ON complaints(user_id);
CREATE INDEX idx_complaints_created ON complaints(created_at);
CREATE INDEX idx_complaints_status ON complaints(status);

-- Proposals Table
CREATE TABLE proposals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    student_id INTEGER,
    proposal_text TEXT NOT NULL,
    telegram_file_id TEXT,
    filename TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'reviewed', 'implemented')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE SET NULL
);

CREATE INDEX idx_proposals_user ON proposals(user_id);
CREATE INDEX idx_proposals_created ON proposals(created_at);
CREATE INDEX idx_proposals_status ON proposals(status);

-- User States Table
CREATE TABLE user_states (
    telegram_id INTEGER PRIMARY KEY,
    state TEXT NOT NULL,
    data TEXT,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_states_updated ON user_states(updated_at);

-- Parent-Students Junction Table
CREATE TABLE parent_students (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER NOT NULL,
    student_id INTEGER NOT NULL UNIQUE,
    linked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    UNIQUE(parent_id, student_id)
);

CREATE INDEX idx_parent_students_parent ON parent_students(parent_id);
CREATE INDEX idx_parent_students_student ON parent_students(student_id);

-- Teacher-Classes Junction Table
CREATE TABLE teacher_classes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    teacher_id INTEGER NOT NULL,
    class_id INTEGER NOT NULL,
    assigned_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (teacher_id) REFERENCES teachers(id) ON DELETE CASCADE,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    UNIQUE(teacher_id, class_id)
);

CREATE INDEX idx_teacher_classes_teacher ON teacher_classes(teacher_id);
CREATE INDEX idx_teacher_classes_class ON teacher_classes(class_id);

-- Announcement-Classes Junction Table
CREATE TABLE announcement_classes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    announcement_id INTEGER NOT NULL,
    class_id INTEGER NOT NULL,
    FOREIGN KEY (announcement_id) REFERENCES announcements(id) ON DELETE CASCADE,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    UNIQUE(announcement_id, class_id)
);

CREATE INDEX idx_announcement_classes_announcement ON announcement_classes(announcement_id);
CREATE INDEX idx_announcement_classes_class ON announcement_classes(class_id);

-- Test Results Table
CREATE TABLE test_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL,
    subject_name TEXT NOT NULL,
    score TEXT NOT NULL,
    test_date DATE NOT NULL,
    teacher_id INTEGER,
    admin_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    FOREIGN KEY (teacher_id) REFERENCES teachers(id) ON DELETE SET NULL,
    FOREIGN KEY (admin_id) REFERENCES admins(id) ON DELETE SET NULL
);

CREATE INDEX idx_test_results_student ON test_results(student_id);
CREATE INDEX idx_test_results_date ON test_results(test_date);
CREATE INDEX idx_test_results_student_date ON test_results(student_id, test_date);

-- Attendance Table
CREATE TABLE attendance (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL,
    date DATE NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('present', 'absent')),
    marked_by_teacher_id INTEGER,
    marked_by_admin_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    FOREIGN KEY (marked_by_teacher_id) REFERENCES teachers(id) ON DELETE SET NULL,
    FOREIGN KEY (marked_by_admin_id) REFERENCES admins(id) ON DELETE SET NULL,
    UNIQUE(student_id, date)
);

CREATE INDEX idx_attendance_student ON attendance(student_id);
CREATE INDEX idx_attendance_date ON attendance(date);
CREATE INDEX idx_attendance_student_date ON attendance(student_id, date);
CREATE INDEX idx_attendance_status ON attendance(status);

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Enforce max 4 children per parent
CREATE TRIGGER enforce_max_children
BEFORE INSERT ON parent_students
FOR EACH ROW
BEGIN
    SELECT CASE
        WHEN (SELECT COUNT(*) FROM parent_students WHERE parent_id = NEW.parent_id) >= 4
        THEN RAISE(ABORT, 'A parent cannot have more than 4 children')
    END;
END;

-- ============================================================================
-- VIEWS
-- ============================================================================

-- View: Students with class info
CREATE VIEW v_students_with_class AS
SELECT
    s.id,
    s.first_name,
    s.last_name,
    s.class_id,
    c.class_name,
    s.is_active,
    s.created_at
FROM students s
JOIN classes c ON s.class_id = c.id;

-- View: Test results with details
CREATE VIEW v_test_results_detailed AS
SELECT
    tr.id,
    tr.student_id,
    s.first_name,
    s.last_name,
    s.class_id,
    c.class_name,
    tr.subject_name,
    tr.score,
    tr.test_date,
    tr.teacher_id,
    tr.admin_id,
    tr.created_at
FROM test_results tr
JOIN students s ON tr.student_id = s.id
JOIN classes c ON s.class_id = c.id;

-- View: Attendance with details
CREATE VIEW v_attendance_detailed AS
SELECT
    a.id,
    a.student_id,
    s.first_name,
    s.last_name,
    s.class_id,
    c.class_name,
    a.date,
    a.status,
    a.marked_by_teacher_id,
    a.marked_by_admin_id,
    a.created_at
FROM attendance a
JOIN students s ON a.student_id = s.id
JOIN classes c ON s.class_id = c.id;

-- View: Complaints with user info
CREATE VIEW v_complaints_with_user AS
SELECT
    c.id,
    c.user_id,
    c.student_id,
    c.complaint_text,
    c.telegram_file_id,
    c.filename,
    c.status,
    c.created_at,
    c.updated_at,
    u.telegram_id,
    u.telegram_username,
    u.phone_number,
    u.language,
    s.first_name as student_first_name,
    s.last_name as student_last_name,
    cl.class_name
FROM complaints c
JOIN users u ON c.user_id = u.id
LEFT JOIN students s ON c.student_id = s.id
LEFT JOIN classes cl ON s.class_id = cl.id;

-- View: Proposals with user info
CREATE VIEW v_proposals_with_user AS
SELECT
    p.id,
    p.user_id,
    p.student_id,
    p.proposal_text,
    p.telegram_file_id,
    p.filename,
    p.status,
    p.created_at,
    p.updated_at,
    u.telegram_id,
    u.telegram_username,
    u.phone_number,
    u.language,
    s.first_name as student_first_name,
    s.last_name as student_last_name,
    cl.class_name
FROM proposals p
JOIN users u ON p.user_id = u.id
LEFT JOIN students s ON p.student_id = s.id
LEFT JOIN classes cl ON s.class_id = cl.id;

-- View: Parent-children info
CREATE VIEW v_parent_children AS
SELECT
    ps.id,
    ps.parent_id,
    u.telegram_id,
    u.phone_number,
    ps.student_id,
    s.first_name as student_first_name,
    s.last_name as student_last_name,
    s.class_id,
    c.class_name,
    ps.linked_at
FROM parent_students ps
JOIN users u ON ps.parent_id = u.id
JOIN students s ON ps.student_id = s.id
JOIN classes c ON s.class_id = c.id;

-- View: Students with parent info
CREATE VIEW v_students_with_parent AS
SELECT
    s.id,
    s.first_name,
    s.last_name,
    s.class_id,
    c.class_name,
    s.is_active,
    ps.parent_id,
    u.telegram_id as parent_telegram_id,
    u.phone_number as parent_phone,
    u.telegram_username as parent_username
FROM students s
JOIN classes c ON s.class_id = c.id
LEFT JOIN parent_students ps ON s.id = ps.student_id
LEFT JOIN users u ON ps.parent_id = u.id;

-- View: Teacher classes
CREATE VIEW v_teacher_classes AS
SELECT
    tc.id,
    tc.teacher_id,
    tc.class_id,
    tc.assigned_at,
    t.first_name,
    t.last_name,
    t.phone_number,
    t.telegram_id,
    c.class_name,
    c.is_active
FROM teacher_classes tc
JOIN teachers t ON tc.teacher_id = t.id
JOIN classes c ON tc.class_id = c.id;

-- View: Test results export
CREATE VIEW v_test_results_export AS
SELECT
    tr.id,
    s.first_name || ' ' || s.last_name as student_name,
    c.class_name,
    tr.subject_name,
    tr.score,
    tr.test_date,
    COALESCE(t.first_name || ' ' || t.last_name, 'N/A') as teacher_name,
    tr.created_at
FROM test_results tr
JOIN students s ON tr.student_id = s.id
JOIN classes c ON s.class_id = c.id
LEFT JOIN teachers t ON tr.teacher_id = t.id;

-- View: Attendance export
CREATE VIEW v_attendance_export AS
SELECT
    a.id,
    s.first_name || ' ' || s.last_name as student_name,
    c.class_name,
    a.date,
    a.status,
    COALESCE(t.first_name || ' ' || t.last_name, 'N/A') as marked_by_teacher,
    a.created_at
FROM attendance a
JOIN students s ON a.student_id = s.id
JOIN classes c ON s.class_id = c.id
LEFT JOIN teachers t ON a.marked_by_teacher_id = t.id;

-- Migration complete
