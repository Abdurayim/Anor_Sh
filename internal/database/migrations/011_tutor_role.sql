-- Migration 011: tutor role
-- role: 'teacher' (default) or 'tutor' (a teacher who can also upload/delete timetables).
-- Existing teachers stay 'teacher'; an admin promotes them from the bot.
-- Only adds columns, no existing row is changed or removed.

ALTER TABLE teachers ADD COLUMN role TEXT NOT NULL DEFAULT 'teacher' CHECK (role IN ('teacher', 'tutor'));
ALTER TABLE timetables ADD COLUMN uploaded_by_teacher_id INTEGER REFERENCES teachers(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_teachers_role ON teachers(role);
