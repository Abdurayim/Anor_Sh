-- Initial database schema for Parent Complaint Bot
-- Optimized with indexes for fast queries

-- Users table with indexes for fast lookups
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    telegram_username VARCHAR(255),
    phone_number VARCHAR(20) UNIQUE NOT NULL,
    child_name VARCHAR(255) NOT NULL,
    child_class VARCHAR(50) NOT NULL,
    language VARCHAR(2) DEFAULT 'uz' CHECK (language IN ('uz', 'ru')),
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for users table (for fast admin queries)
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_users_phone_number ON users(phone_number);
CREATE INDEX IF NOT EXISTS idx_users_child_class ON users(child_class);
CREATE INDEX IF NOT EXISTS idx_users_registered_at ON users(registered_at DESC);

-- Complaints table optimized for retrieval
CREATE TABLE IF NOT EXISTS complaints (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    complaint_text TEXT NOT NULL,
    telegram_file_id VARCHAR(255) NOT NULL,
    filename VARCHAR(500) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'archived'))
);

-- Indexes for complaints table (optimized for admin dashboard)
CREATE INDEX IF NOT EXISTS idx_complaints_user_id ON complaints(user_id);
CREATE INDEX IF NOT EXISTS idx_complaints_created_at ON complaints(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_complaints_status ON complaints(status);
CREATE INDEX IF NOT EXISTS idx_complaints_combined ON complaints(user_id, created_at DESC);

-- Admins table (max 3 admins)
CREATE TABLE IF NOT EXISTS admins (
    id SERIAL PRIMARY KEY,
    phone_number VARCHAR(20) UNIQUE NOT NULL,
    telegram_id BIGINT UNIQUE,
    name VARCHAR(255),
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for admins table
CREATE INDEX IF NOT EXISTS idx_admins_phone_number ON admins(phone_number);
CREATE INDEX IF NOT EXISTS idx_admins_telegram_id ON admins(telegram_id);

-- User states table for conversation flow management
CREATE TABLE IF NOT EXISTS user_states (
    telegram_id BIGINT PRIMARY KEY,
    state VARCHAR(100) NOT NULL,
    data JSONB,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for cleanup of old states
CREATE INDEX IF NOT EXISTS idx_user_states_updated_at ON user_states(updated_at);

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for user_states
CREATE TRIGGER update_user_states_updated_at BEFORE UPDATE ON user_states
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Constraint to ensure maximum 3 admins
CREATE OR REPLACE FUNCTION check_max_admins()
RETURNS TRIGGER AS $$
BEGIN
    IF (SELECT COUNT(*) FROM admins) >= 3 THEN
        RAISE EXCEPTION 'Maximum of 3 admins allowed';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_max_admins
    BEFORE INSERT ON admins
    FOR EACH ROW EXECUTE FUNCTION check_max_admins();

-- Create view for admin dashboard (optimized query)
CREATE OR REPLACE VIEW v_complaints_with_user AS
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

-- Insert sample admins (replace with actual phone numbers)
-- This will be populated from environment variables in the application
-- INSERT INTO admins (phone_number, name) VALUES ('+998901234567', 'Admin 1');
