package models

import "time"

// Admin represents an admin user
type Admin struct {
	ID          int       `json:"id" db:"id"`
	PhoneNumber string    `json:"phone_number" db:"phone_number"`
	TelegramID  *int64    `json:"telegram_id,omitempty" db:"telegram_id"`
	Name        string    `json:"name" db:"name"`
	BranchID    int       `json:"branch_id" db:"branch_id"` // 0 for the super admin
	Role        string    `json:"role" db:"role"`
	Source      string    `json:"source" db:"source"`
	AddedAt     time.Time `json:"added_at" db:"added_at"`
}

// Admin roles
const (
	RoleAdmin      = "admin"
	RoleSuperAdmin = "super_admin"
)

// Admin sources: synced from .env or added by the super admin in the bot
const (
	AdminSourceEnv = "env"
	AdminSourceBot = "bot"
)

// IsSuperAdmin reports whether the admin has the super admin role
func (a *Admin) IsSuperAdmin() bool {
	return a.Role == RoleSuperAdmin
}

// IsAdmin checks if a phone number or telegram ID is an admin
type AdminCheck struct {
	PhoneNumber string
	TelegramID  int64
}
