-- Migration 008: Fix v_parent_children view column names
-- The view had columns like 'parent_telegram_id' but code expects 'telegram_id'

-- Drop and recreate the view with correct column names
DROP VIEW IF EXISTS v_parent_children;

CREATE VIEW v_parent_children AS
SELECT
    ps.id,
    ps.parent_id,
    u.telegram_id,              -- Changed from parent_telegram_id
    u.phone_number,             -- Changed from parent_phone
    ps.student_id,
    s.first_name as student_first_name,  -- Changed from first_name
    s.last_name as student_last_name,    -- Changed from last_name
    s.class_id,
    c.class_name,
    ps.linked_at
FROM parent_students ps
JOIN users u ON ps.parent_id = u.id
JOIN students s ON ps.student_id = s.id
JOIN classes c ON s.class_id = c.id;
