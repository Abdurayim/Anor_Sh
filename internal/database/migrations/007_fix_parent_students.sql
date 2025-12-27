-- Migration 007: Fix parent_students to allow multiple parents per student
-- Remove UNIQUE constraint on student_id to allow both mother and father to link

-- Step 1: Create new table with correct schema
CREATE TABLE parent_students_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id INTEGER NOT NULL,
    student_id INTEGER NOT NULL,  -- REMOVED UNIQUE HERE
    linked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE CASCADE,
    UNIQUE(parent_id, student_id)  -- This prevents same parent linking to same student twice
);

-- Step 2: Copy existing data
INSERT INTO parent_students_new (id, parent_id, student_id, linked_at)
SELECT id, parent_id, student_id, linked_at FROM parent_students;

-- Step 3: Drop old table
DROP TABLE parent_students;

-- Step 4: Rename new table
ALTER TABLE parent_students_new RENAME TO parent_students;

-- Step 5: Recreate indexes
CREATE INDEX idx_parent_students_parent ON parent_students(parent_id);
CREATE INDEX idx_parent_students_student ON parent_students(student_id);

-- Step 6: Recreate trigger for max 4 children per parent
CREATE TRIGGER enforce_max_children
BEFORE INSERT ON parent_students
FOR EACH ROW
BEGIN
    SELECT CASE
        WHEN (SELECT COUNT(*) FROM parent_students WHERE parent_id = NEW.parent_id) >= 4
        THEN RAISE(ABORT, 'A parent cannot have more than 4 children')
    END;
END;
