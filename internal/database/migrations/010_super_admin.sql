-- Migration 010: super admin role and admins added from the bot
-- role:   'admin' (branch admin, branch_id set) or 'super_admin' (all branches, read-only, manages admins)
-- source: 'env' (synced from .env on every start) or 'bot' (added by the super admin in the bot)

ALTER TABLE admins ADD COLUMN role TEXT NOT NULL DEFAULT 'admin' CHECK (role IN ('admin', 'super_admin'));
ALTER TABLE admins ADD COLUMN source TEXT NOT NULL DEFAULT 'env' CHECK (source IN ('env', 'bot'));
ALTER TABLE admins ADD COLUMN added_by_admin_id INTEGER REFERENCES admins(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_admins_role ON admins(role);
