-- Migration for proposals, timetables, and announcements features

-- Proposals table (similar structure to complaints)
CREATE TABLE IF NOT EXISTS proposals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    proposal_text TEXT NOT NULL,
    telegram_file_id TEXT NOT NULL,
    filename TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'archived')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes for proposals table
CREATE INDEX IF NOT EXISTS idx_proposals_user_id ON proposals(user_id);
CREATE INDEX IF NOT EXISTS idx_proposals_created_at ON proposals(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_proposals_status ON proposals(status);
CREATE INDEX IF NOT EXISTS idx_proposals_combined ON proposals(user_id, created_at DESC);

-- View for admin dashboard
CREATE VIEW IF NOT EXISTS v_proposals_with_user AS
SELECT
    p.id,
    p.user_id,
    p.proposal_text,
    p.telegram_file_id,
    p.filename,
    p.created_at,
    p.status,
    u.telegram_id AS user_telegram_id,
    u.telegram_username,
    u.phone_number,
    u.child_name,
    u.child_class
FROM proposals p
INNER JOIN users u ON p.user_id = u.id
ORDER BY p.created_at DESC;

-- Timetables table (store timetable files per class)
-- Multiple formats supported: jpeg, jpg, heic, excel, word
CREATE TABLE IF NOT EXISTS timetables (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id INTEGER NOT NULL,
    telegram_file_id TEXT NOT NULL,
    filename TEXT NOT NULL,
    file_type TEXT NOT NULL, -- image, document
    mime_type TEXT, -- image/jpeg, application/pdf, etc.
    uploaded_by_admin_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE CASCADE,
    FOREIGN KEY (uploaded_by_admin_id) REFERENCES admins(id) ON DELETE SET NULL
);

-- Index for timetables (one timetable per class, get latest)
CREATE INDEX IF NOT EXISTS idx_timetables_class_id ON timetables(class_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_timetables_created_at ON timetables(created_at DESC);

-- Announcements table (text + optional image)
CREATE TABLE IF NOT EXISTS announcements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT,
    content TEXT NOT NULL,
    telegram_file_id TEXT, -- optional image/photo
    filename TEXT,
    file_type TEXT, -- image
    posted_by_admin_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_active INTEGER DEFAULT 1 CHECK (is_active IN (0, 1)),
    FOREIGN KEY (posted_by_admin_id) REFERENCES admins(id) ON DELETE SET NULL
);

-- Indexes for announcements
CREATE INDEX IF NOT EXISTS idx_announcements_created_at ON announcements(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_announcements_is_active ON announcements(is_active, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_announcements_admin ON announcements(posted_by_admin_id);

-- Trigger to automatically update timetables updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_timetables_updated_at
AFTER UPDATE ON timetables
FOR EACH ROW
WHEN OLD.updated_at = NEW.updated_at
BEGIN
    UPDATE timetables SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
