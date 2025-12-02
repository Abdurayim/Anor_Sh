-- ============================================================================
-- MIGRATION 003: MAJOR REDESIGN
-- ============================================================================
-- This migration implements a complete architectural redesign:
-- - Adds teacher role
-- - Separates students from parents
-- - Adds test results and attendance tracking
-- - Implements multi-child support for parents
-- - Adds multi-class targeting for announcements
-- ============================================================================

-- Drop all existing tables for fresh start
DROP TABLE IF EXISTS user_states;
DROP TABLE IF EXISTS announcements;
DROP TABLE IF EXISTS timetables;
DROP TABLE IF EXISTS proposals;
DROP TABLE IF EXISTS complaints;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS admins;
DROP TABLE IF EXISTS classes;

-- Drop new tables if they exist
DROP TABLE IF EXISTS announcement_classes;
DROP TABLE IF EXISTS attendance;
DROP TABLE IF EXISTS test_results;
DROP TABLE IF EXISTS teacher_classes;
DROP TABLE IF EXISTS parent_students;
DROP TABLE IF EXISTS students;
DROP TABLE IF EXISTS teachers;

-- ============================================================================
-- CORE ENTITIES
-- ============================================================================

-- Classes Table (Grade-based classes like 3A, 4B, 11-A)
CREATE TABLE classes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_name TEXT NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_classes_active ON classes(is_active);
CREATE INDEX idx_classes_name ON classes(class_name);

-- Admins Table (Max 3 admins)
CREATE TABLE admins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    phone_number TEXT NOT NULL UNIQUE,
    telegram_id INTEGER UNIQUE,
    name TEXT,
    added_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_admins_phone ON admins(phone_number);
CREATE INDEX idx_admins_telegram ON admins(telegram_id);

-- Admin limit trigger (max 3)
CREATE TRIGGER enforce_admin_limit
BEFORE INSERT ON admins
WHEN (SELECT COUNT(*) FROM admins) >= 3
BEGIN
    SELECT RAISE(ABORT, 'Maximum of 3 admins allowed');
END;

-- Teachers Table (Phone-based authentication)
CREATE TABLE teachers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    phone_number TEXT NOT NULL UNIQUE,
    telegram_id INTEGER UNIQUE,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    language TEXT NOT NULL DEFAULT 'uz' CHECK(language IN ('uz', 'ru')),
    is_active BOOLEAN NOT NULL DEFAULT 1,
    added_by_admin_id INTEGER,
    registered_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (added_by_admin_id) REFERENCES admins(id) ON DELETE SET NULL
);

CREATE INDEX idx_teachers_phone ON teachers(phone_number);
CREATE INDEX idx_teachers_telegram ON teachers(telegram_id);
CREATE INDEX idx_teachers_active ON teachers(is_active);

-- Students Table (Managed by admin/teacher)
CREATE TABLE students (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    class_id INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    added_by_admin_id INTEGER,
    added_by_teacher_id INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    FOREIGN KEY (added_by_admin_id) REFERENCES admins(id) ON DELETE SET NULL,
    FOREIGN KEY (added_by_teacher_id) REFERENCES teachers(id) ON DELETE SET NULL
);

CREATE INDEX idx_students_class ON students(class_id);
CREATE INDEX idx_students_active ON students(is_active);
CREATE INDEX idx_students_name ON students(last_name, first_name);

-- Users Table (Parents - Modified)
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    telegram_id INTEGER NOT NULL UNIQUE,
    telegram_username TEXT,
    phone_number TEXT NOT NULL UNIQUE,
    language TEXT NOT NULL DEFAULT 'uz' CHECK(language IN ('uz', 'ru')),
    current_selected_student_id INTEGER,
    registered_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (current_selected_student_id) REFERENCES students(id) ON DELETE SET NULL
);

CREATE INDEX idx_users_telegram ON users(telegram_id);
CREATE INDEX idx_users_phone ON users(phone_number);
CREATE INDEX idx_users_registered ON users(registered_at);

-- ============================================================================
-- EXISTING FEATURES (Modified)
-- ============================================================================

-- Announcements Table (Now supports teachers and multi-class targeting)
CREATE TABLE announcements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT,
    content TEXT NOT NULL,
    telegram_file_id TEXT,
    filename TEXT,
    file_type TEXT CHECK(file_type IN ('image', 'document')),
    posted_by_admin_id INTEGER,
    posted_by_teacher_id INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    FOREIGN KEY (posted_by_admin_id) REFERENCES admins(id) ON DELETE SET NULL,
    FOREIGN KEY (posted_by_teacher_id) REFERENCES teachers(id) ON DELETE SET NULL,
    CHECK ((posted_by_admin_id IS NOT NULL) OR (posted_by_teacher_id IS NOT NULL))
);

CREATE INDEX idx_announcements_created ON announcements(created_at);
CREATE INDEX idx_announcements_active ON announcements(is_active, created_at);
CREATE INDEX idx_announcements_admin ON announcements(posted_by_admin_id);
CREATE INDEX idx_announcements_teacher ON announcements(posted_by_teacher_id);

-- Timetables Table (Keep mostly same)
CREATE TABLE timetables (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL,
    telegram_file_id TEXT NOT NULL,
    filename TEXT,
    file_type TEXT NOT NULL CHECK(file_type IN ('image', 'document')),
    mime_type TEXT,
    uploaded_by_admin_id INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    FOREIGN KEY (uploaded_by_admin_id) REFERENCES admins(id) ON DELETE SET NULL
);

CREATE INDEX idx_timetables_class ON timetables(class_id, created_at);
CREATE INDEX idx_timetables_created ON timetables(created_at);

-- Complaints Table (Keep mostly same)
CREATE TABLE complaints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    complaint_text TEXT NOT NULL,
    telegram_file_id TEXT,
    filename TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'reviewed', 'archived')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_complaints_user ON complaints(user_id);
CREATE INDEX idx_complaints_created ON complaints(created_at);
CREATE INDEX idx_complaints_status ON complaints(status);
CREATE INDEX idx_complaints_user_created ON complaints(user_id, created_at);

-- Proposals Table (Keep mostly same)
CREATE TABLE proposals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    proposal_text TEXT NOT NULL,
    telegram_file_id TEXT,
    filename TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'reviewed', 'archived')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_proposals_user ON proposals(user_id);
CREATE INDEX idx_proposals_created ON proposals(created_at);
CREATE INDEX idx_proposals_status ON proposals(status);
CREATE INDEX idx_proposals_user_created ON proposals(user_id, created_at);

-- User States Table (Keep same for conversation flow)
CREATE TABLE user_states (
    telegram_id INTEGER PRIMARY KEY,
    state TEXT NOT NULL,
    data TEXT,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_states_updated ON user_states(updated_at);

-- Auto-update trigger for user_states
CREATE TRIGGER update_user_states_timestamp
AFTER UPDATE ON user_states
FOR EACH ROW
BEGIN
    UPDATE user_states SET updated_at = CURRENT_TIMESTAMP WHERE telegram_id = NEW.telegram_id;
END;

-- ============================================================================
-- JUNCTION TABLES (Many-to-Many Relationships)
-- ============================================================================

-- Parent-Student Junction (Max 4 children per parent)
CREATE TABLE parent_students (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER NOT NULL,
    student_id INTEGER NOT NULL,
    linked_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    UNIQUE(parent_id, student_id)
);

CREATE INDEX idx_parent_students_parent ON parent_students(parent_id);
CREATE INDEX idx_parent_students_student ON parent_students(student_id);

-- Trigger to enforce max 4 children per parent
CREATE TRIGGER enforce_max_children
BEFORE INSERT ON parent_students
WHEN (SELECT COUNT(*) FROM parent_students WHERE parent_id = NEW.parent_id) >= 4
BEGIN
    SELECT RAISE(ABORT, 'Maximum of 4 children allowed per parent');
END;

-- Teacher-Class Junction (Teachers can manage multiple classes)
CREATE TABLE teacher_classes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    teacher_id INTEGER NOT NULL,
    class_id INTEGER NOT NULL,
    assigned_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (teacher_id) REFERENCES teachers(id) ON DELETE CASCADE,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    UNIQUE(teacher_id, class_id)
);

CREATE INDEX idx_teacher_classes_teacher ON teacher_classes(teacher_id);
CREATE INDEX idx_teacher_classes_class ON teacher_classes(class_id);

-- Announcement-Class Junction (Announcements can target multiple classes)
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

-- ============================================================================
-- ACADEMIC TRACKING
-- ============================================================================

-- Test Results Table
CREATE TABLE test_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL,
    subject_name TEXT NOT NULL,
    score TEXT NOT NULL,
    test_date DATE NOT NULL,
    teacher_id INTEGER,
    admin_id INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    FOREIGN KEY (teacher_id) REFERENCES teachers(id) ON DELETE SET NULL,
    FOREIGN KEY (admin_id) REFERENCES admins(id) ON DELETE SET NULL,
    CHECK ((teacher_id IS NOT NULL) OR (admin_id IS NOT NULL))
);

CREATE INDEX idx_test_results_student ON test_results(student_id);
CREATE INDEX idx_test_results_date ON test_results(test_date);
CREATE INDEX idx_test_results_student_date ON test_results(student_id, test_date);

-- Attendance Table
CREATE TABLE attendance (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL,
    date DATE NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('present', 'absent')),
    marked_by_teacher_id INTEGER,
    marked_by_admin_id INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    FOREIGN KEY (marked_by_teacher_id) REFERENCES teachers(id) ON DELETE SET NULL,
    FOREIGN KEY (marked_by_admin_id) REFERENCES admins(id) ON DELETE SET NULL,
    CHECK ((marked_by_teacher_id IS NOT NULL) OR (marked_by_admin_id IS NOT NULL)),
    UNIQUE(student_id, date)
);

CREATE INDEX idx_attendance_student ON attendance(student_id);
CREATE INDEX idx_attendance_date ON attendance(date);
CREATE INDEX idx_attendance_student_date ON attendance(student_id, date);
CREATE INDEX idx_attendance_status ON attendance(status);

-- ============================================================================
-- VIEWS FOR EASY QUERYING
-- ============================================================================

-- View: Complaints with user info
CREATE VIEW v_complaints_with_user AS
SELECT
    c.id,
    c.user_id,
    c.complaint_text,
    c.telegram_file_id,
    c.filename,
    c.created_at,
    c.status,
    u.telegram_id AS user_telegram_id,
    u.telegram_username,
    u.phone_number,
    u.language
FROM complaints c
JOIN users u ON c.user_id = u.id;

-- View: Proposals with user info
CREATE VIEW v_proposals_with_user AS
SELECT
    p.id,
    p.user_id,
    p.proposal_text,
    p.telegram_file_id,
    p.filename,
    p.created_at,
    p.status,
    u.telegram_username,
    u.phone_number,
    u.language
FROM proposals p
JOIN users u ON p.user_id = u.id;

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

-- View: Parent-Student relationships
CREATE VIEW v_parent_children AS
SELECT
    ps.id,
    ps.parent_id,
    u.telegram_id,
    u.phone_number,
    ps.student_id,
    s.first_name AS student_first_name,
    s.last_name AS student_last_name,
    s.class_id,
    c.class_name,
    ps.linked_at
FROM parent_students ps
JOIN users u ON ps.parent_id = u.id
JOIN students s ON ps.student_id = s.id
JOIN classes c ON s.class_id = c.id;

-- View: Test results with student and class info
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
    tr.created_at,
    tr.updated_at
FROM test_results tr
JOIN students s ON tr.student_id = s.id
JOIN classes c ON s.class_id = c.id;

-- View: Attendance with student and class info
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
    a.created_at
FROM attendance a
JOIN students s ON a.student_id = s.id
JOIN classes c ON s.class_id = c.id;

-- ============================================================================
-- TRIGGERS FOR CASCADE HANDLING
-- ============================================================================

-- When class is deleted, notify affected parents (handled in application logic)
-- This trigger ensures all related data is properly cleaned up
CREATE TRIGGER class_deletion_cleanup
BEFORE DELETE ON classes
FOR EACH ROW
BEGIN
    -- Students will cascade delete
    -- This will trigger cascade on parent_students, test_results, attendance
    -- Timetables will cascade delete
    -- Announcement_classes will cascade delete
    -- Teacher_classes will cascade delete
    SELECT 1; -- Placeholder for potential logging
END;
