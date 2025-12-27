-- Migration 005: Fresh Start - Combined schema for clean database
-- This migration creates all tables from scratch if they don't exist

-- Create user_states table if it doesn't exist
CREATE TABLE IF NOT EXISTS user_states (
    telegram_id INTEGER PRIMARY KEY,
    state TEXT NOT NULL,
    data TEXT,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_states_updated ON user_states(updated_at);

-- Cleanup any leftover _new tables from failed migrations
DROP TABLE IF EXISTS users_new;
DROP TABLE IF EXISTS students_new;
DROP TABLE IF EXISTS parent_students_new;
DROP TABLE IF EXISTS test_results_new;
DROP TABLE IF EXISTS attendance_new;
DROP TABLE IF EXISTS timetables_new;
DROP TABLE IF EXISTS teacher_classes_new;
DROP TABLE IF EXISTS announcement_classes_new;
