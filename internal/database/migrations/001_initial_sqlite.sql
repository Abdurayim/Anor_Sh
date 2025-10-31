-- Initial database schema for Parent Complaint Bot (SQLite version)
-- Optimized with indexes for fast queries

-- Classes table - admins can create classes that parents choose from
CREATE TABLE IF NOT EXISTS classes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_name TEXT UNIQUE NOT NULL,
    is_active INTEGER DEFAULT 1 CHECK (is_active IN (0, 1)),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for active classes
CREATE INDEX IF NOT EXISTS idx_classes_active ON classes(is_active, class_name);

-- Users table with indexes for fast lookups
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    telegram_id INTEGER UNIQUE NOT NULL,
    telegram_username TEXT,
    phone_number TEXT UNIQUE NOT NULL,
    child_name TEXT NOT NULL,
    child_class TEXT NOT NULL,
    language TEXT DEFAULT 'uz' CHECK (language IN ('uz', 'ru')),
    registered_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for users table (for fast admin queries)
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_users_phone_number ON users(phone_number);
CREATE INDEX IF NOT EXISTS idx_users_child_class ON users(child_class);
CREATE INDEX IF NOT EXISTS idx_users_registered_at ON users(registered_at DESC);

-- Complaints table optimized for retrieval
CREATE TABLE IF NOT EXISTS complaints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    complaint_text TEXT NOT NULL,
    telegram_file_id TEXT NOT NULL,
    filename TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'archived')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Indexes for complaints table (optimized for admin dashboard)
CREATE INDEX IF NOT EXISTS idx_complaints_user_id ON complaints(user_id);
CREATE INDEX IF NOT EXISTS idx_complaints_created_at ON complaints(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_complaints_status ON complaints(status);
CREATE INDEX IF NOT EXISTS idx_complaints_combined ON complaints(user_id, created_at DESC);

-- Admins table (max 3 admins)
CREATE TABLE IF NOT EXISTS admins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    phone_number TEXT UNIQUE NOT NULL,
    telegram_id INTEGER UNIQUE,
    name TEXT,
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for admins table
CREATE INDEX IF NOT EXISTS idx_admins_phone_number ON admins(phone_number);
CREATE INDEX IF NOT EXISTS idx_admins_telegram_id ON admins(telegram_id);

-- User states table for conversation flow management
CREATE TABLE IF NOT EXISTS user_states (
    telegram_id INTEGER PRIMARY KEY,
    state TEXT NOT NULL,
    data TEXT, -- JSON data stored as TEXT
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for cleanup of old states
CREATE INDEX IF NOT EXISTS idx_user_states_updated_at ON user_states(updated_at);

-- Trigger to automatically update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_user_states_updated_at
AFTER UPDATE ON user_states
FOR EACH ROW
WHEN OLD.updated_at = NEW.updated_at
BEGIN
    UPDATE user_states SET updated_at = CURRENT_TIMESTAMP WHERE telegram_id = NEW.telegram_id;
END;

-- Trigger to enforce maximum 3 admins
CREATE TRIGGER IF NOT EXISTS enforce_max_admins
BEFORE INSERT ON admins
FOR EACH ROW
WHEN (SELECT COUNT(*) FROM admins) >= 3
BEGIN
    SELECT RAISE(ABORT, 'Maximum of 3 admins allowed');
END;

-- Create view for admin dashboard (optimized query)
CREATE VIEW IF NOT EXISTS v_complaints_with_user AS
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
    u.child_name,
    u.child_class
FROM complaints c
INNER JOIN users u ON c.user_id = u.id
ORDER BY c.created_at DESC;
