-- Migration 004: Architecture Redesign for Multi-Child, Teacher Roles, and Enhanced Features
-- This migration updates the schema for the new requirements:
-- 1. Parents can have up to 4 children (reduced from 5)
-- 2. Remove current_selected_student_id from users (no longer needed)
-- 3. Add proper CASCADE deletes for class deletion
-- 4. Optimize indexes for pagination

-- Step 0: Drop views that reference tables we're about to modify
-- (SQLite views block table drops/renames)
DROP VIEW IF EXISTS v_complaints_with_user;
DROP VIEW IF EXISTS v_proposals_with_user;
DROP VIEW IF EXISTS v_parent_children;
DROP VIEW IF EXISTS v_students_with_parent;

-- Step 1: Remove current_selected_student_id from users table
-- This field is no longer needed since parents manage multiple children
-- We need to recreate the table because SQLite doesn't support DROP COLUMN with foreign keys

CREATE TABLE users_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    telegram_id INTEGER NOT NULL UNIQUE,
    telegram_username TEXT,
    phone_number TEXT NOT NULL UNIQUE,
    language TEXT NOT NULL DEFAULT 'uz' CHECK(language IN ('uz', 'ru')),
    registered_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Copy data from old users table (excluding current_selected_student_id)
INSERT INTO users_new (id, telegram_id, telegram_username, phone_number, language, registered_at)
SELECT id, telegram_id, telegram_username, phone_number, language, registered_at FROM users;

-- Drop old table and rename new one
DROP TABLE users;
ALTER TABLE users_new RENAME TO users;

-- Recreate indexes for users
CREATE INDEX idx_users_telegram ON users(telegram_id);
CREATE INDEX idx_users_phone ON users(phone_number);
CREATE INDEX idx_users_registered ON users(registered_at);

-- Step 2: Update parent_students max children constraint from 5 to 4
-- First, drop the old trigger
DROP TRIGGER IF EXISTS enforce_max_children;

-- Recreate the trigger with new limit of 4
CREATE TRIGGER enforce_max_children
BEFORE INSERT ON parent_students
FOR EACH ROW
BEGIN
    SELECT CASE
        WHEN (SELECT COUNT(*) FROM parent_students WHERE parent_id = NEW.parent_id) >= 4
        THEN RAISE(ABORT, 'A parent cannot have more than 4 children')
    END;
END;

-- Step 3: Recreate foreign key constraints with CASCADE DELETE
-- We need to recreate tables to add ON DELETE CASCADE

-- 3.1: Recreate students table with CASCADE on class deletion
CREATE TABLE students_new (
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
    FOREIGN KEY (added_by_teacher_id) REFERENCES teachers(id) ON DELETE SET NULL,
    CHECK (added_by_admin_id IS NOT NULL OR added_by_teacher_id IS NOT NULL)
);

-- Copy data from old students table
INSERT INTO students_new SELECT * FROM students;

-- Drop old table and rename new one
DROP TABLE students;
ALTER TABLE students_new RENAME TO students;

-- Recreate indexes for students
CREATE INDEX idx_students_class ON students(class_id);
CREATE INDEX idx_students_active ON students(is_active);
CREATE INDEX idx_students_name ON students(last_name, first_name);

-- 3.2: Recreate parent_students with CASCADE on student deletion
CREATE TABLE parent_students_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER NOT NULL,
    student_id INTEGER NOT NULL UNIQUE,
    linked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    UNIQUE(parent_id, student_id)
);

-- Copy data
INSERT INTO parent_students_new SELECT * FROM parent_students;

-- Drop and rename
DROP TABLE parent_students;
ALTER TABLE parent_students_new RENAME TO parent_students;

-- Recreate indexes
CREATE INDEX idx_parent_students_parent ON parent_students(parent_id);
CREATE INDEX idx_parent_students_student ON parent_students(student_id);

-- 3.3: Recreate test_results with CASCADE on student deletion
CREATE TABLE test_results_new (
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
    FOREIGN KEY (admin_id) REFERENCES admins(id) ON DELETE SET NULL,
    CHECK (teacher_id IS NOT NULL OR admin_id IS NOT NULL)
);

-- Copy data
INSERT INTO test_results_new SELECT * FROM test_results;

-- Drop and rename
DROP TABLE test_results;
ALTER TABLE test_results_new RENAME TO test_results;

-- Recreate indexes
CREATE INDEX idx_test_results_student ON test_results(student_id);
CREATE INDEX idx_test_results_date ON test_results(test_date);
CREATE INDEX idx_test_results_student_date ON test_results(student_id, test_date);
CREATE INDEX idx_test_results_teacher ON test_results(teacher_id);
CREATE INDEX idx_test_results_admin ON test_results(admin_id);

-- 3.4: Recreate attendance with CASCADE on student deletion
CREATE TABLE attendance_new (
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
    UNIQUE(student_id, date),
    CHECK (marked_by_teacher_id IS NOT NULL OR marked_by_admin_id IS NOT NULL)
);

-- Copy data
INSERT INTO attendance_new SELECT * FROM attendance;

-- Drop and rename
DROP TABLE attendance;
ALTER TABLE attendance_new RENAME TO attendance;

-- Recreate indexes
CREATE INDEX idx_attendance_student ON attendance(student_id);
CREATE INDEX idx_attendance_date ON attendance(date);
CREATE INDEX idx_attendance_student_date ON attendance(student_id, date);
CREATE INDEX idx_attendance_status ON attendance(status);
CREATE INDEX idx_attendance_teacher ON attendance(marked_by_teacher_id);
CREATE INDEX idx_attendance_admin ON attendance(marked_by_admin_id);

-- 3.5: Recreate timetables with CASCADE on class deletion
CREATE TABLE timetables_new (
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

-- Copy data
INSERT INTO timetables_new SELECT * FROM timetables;

-- Drop and rename
DROP TABLE timetables;
ALTER TABLE timetables_new RENAME TO timetables;

-- Recreate indexes
CREATE INDEX idx_timetables_class ON timetables(class_id);
CREATE INDEX idx_timetables_created ON timetables(created_at);

-- 3.6: Recreate teacher_classes with CASCADE on class deletion
CREATE TABLE teacher_classes_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    teacher_id INTEGER NOT NULL,
    class_id INTEGER NOT NULL,
    assigned_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (teacher_id) REFERENCES teachers(id) ON DELETE CASCADE,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    UNIQUE(teacher_id, class_id)
);

-- Copy data
INSERT INTO teacher_classes_new SELECT * FROM teacher_classes;

-- Drop and rename
DROP TABLE teacher_classes;
ALTER TABLE teacher_classes_new RENAME TO teacher_classes;

-- Recreate indexes
CREATE INDEX idx_teacher_classes_teacher ON teacher_classes(teacher_id);
CREATE INDEX idx_teacher_classes_class ON teacher_classes(class_id);

-- 3.7: Recreate announcement_classes with CASCADE on class and announcement deletion
CREATE TABLE announcement_classes_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    announcement_id INTEGER NOT NULL,
    class_id INTEGER NOT NULL,
    FOREIGN KEY (announcement_id) REFERENCES announcements(id) ON DELETE CASCADE,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    UNIQUE(announcement_id, class_id)
);

-- Copy data
INSERT INTO announcement_classes_new SELECT * FROM announcement_classes;

-- Drop and rename
DROP TABLE announcement_classes;
ALTER TABLE announcement_classes_new RENAME TO announcement_classes;

-- Recreate indexes
CREATE INDEX idx_announcement_classes_announcement ON announcement_classes(announcement_id);
CREATE INDEX idx_announcement_classes_class ON announcement_classes(class_id);

-- Step 4: Add indexes for pagination support
CREATE INDEX idx_complaints_created_paginate ON complaints(created_at DESC, id);
CREATE INDEX idx_proposals_created_paginate ON proposals(created_at DESC, id);
CREATE INDEX idx_students_created_paginate ON students(created_at DESC, id);

-- Step 5: Create new views for the redesigned architecture

-- View: Parent's children with full details
DROP VIEW IF EXISTS v_parent_children;
CREATE VIEW v_parent_children AS
SELECT
    ps.parent_id,
    ps.student_id,
    ps.linked_at,
    s.first_name,
    s.last_name,
    s.class_id,
    s.is_active as student_is_active,
    c.class_name,
    c.is_active as class_is_active,
    u.telegram_id as parent_telegram_id,
    u.phone_number as parent_phone,
    u.language as parent_language
FROM parent_students ps
JOIN students s ON ps.student_id = s.id
JOIN classes c ON s.class_id = c.id
JOIN users u ON ps.parent_id = u.id;

-- View: Teacher's assigned classes with student counts
DROP VIEW IF EXISTS v_teacher_classes;
CREATE VIEW v_teacher_classes AS
SELECT
    tc.teacher_id,
    tc.class_id,
    tc.assigned_at,
    c.class_name,
    c.is_active as class_is_active,
    t.first_name as teacher_first_name,
    t.last_name as teacher_last_name,
    t.phone_number as teacher_phone,
    COUNT(DISTINCT s.id) as student_count
FROM teacher_classes tc
JOIN classes c ON tc.class_id = c.id
JOIN teachers t ON tc.teacher_id = t.id
LEFT JOIN students s ON s.class_id = c.id AND s.is_active = 1
GROUP BY tc.id;

-- View: Students with parent information
DROP VIEW IF EXISTS v_students_with_parent;
CREATE VIEW v_students_with_parent AS
SELECT
    s.id as student_id,
    s.first_name,
    s.last_name,
    s.class_id,
    s.is_active,
    s.created_at,
    c.class_name,
    ps.parent_id,
    u.phone_number as parent_phone,
    u.telegram_id as parent_telegram_id,
    u.language as parent_language
FROM students s
JOIN classes c ON s.class_id = c.id
LEFT JOIN parent_students ps ON s.id = ps.student_id
LEFT JOIN users u ON ps.parent_id = u.id;

-- View: Test results grouped by student (for export)
DROP VIEW IF EXISTS v_test_results_export;
CREATE VIEW v_test_results_export AS
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
    CASE
        WHEN tr.teacher_id IS NOT NULL THEN t.first_name || ' ' || t.last_name
        WHEN tr.admin_id IS NOT NULL THEN a.name
    END as added_by,
    tr.created_at
FROM test_results tr
JOIN students s ON tr.student_id = s.id
JOIN classes c ON s.class_id = c.id
LEFT JOIN teachers t ON tr.teacher_id = t.id
LEFT JOIN admins a ON tr.admin_id = a.id
ORDER BY c.class_name, s.last_name, s.first_name, tr.test_date DESC;

-- View: Attendance for export (current date)
DROP VIEW IF EXISTS v_attendance_export;
CREATE VIEW v_attendance_export AS
SELECT
    a.id,
    a.student_id,
    s.first_name,
    s.last_name,
    s.class_id,
    c.class_name,
    a.date,
    a.status,
    CASE
        WHEN a.marked_by_teacher_id IS NOT NULL THEN t.first_name || ' ' || t.last_name
        WHEN a.marked_by_admin_id IS NOT NULL THEN adm.name
    END as marked_by,
    a.created_at
FROM attendance a
JOIN students s ON a.student_id = s.id
JOIN classes c ON s.class_id = c.id
LEFT JOIN teachers t ON a.marked_by_teacher_id = t.id
LEFT JOIN admins adm ON a.marked_by_admin_id = adm.id
ORDER BY c.class_name, s.last_name, s.first_name;

-- Step 6: Update triggers

-- Recreate the enforce_max_children trigger (already done above, but ensuring it's active)
DROP TRIGGER IF EXISTS enforce_max_children;
CREATE TRIGGER enforce_max_children
BEFORE INSERT ON parent_students
FOR EACH ROW
BEGIN
    SELECT CASE
        WHEN (SELECT COUNT(*) FROM parent_students WHERE parent_id = NEW.parent_id) >= 4
        THEN RAISE(ABORT, 'A parent cannot have more than 4 children')
    END;
END;

-- Step 7: Recreate views that were dropped in Step 0

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

-- Migration complete
