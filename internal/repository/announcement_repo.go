package repository

import (
	"database/sql"
	"fmt"

	"parent-bot/internal/models"
)

// AnnouncementRepository stores announcements. Each announcement belongs to a branch and,
// optionally, targets specific classes (announcement_classes); without targets it is for
// the whole branch.
type AnnouncementRepository struct {
	db *sql.DB
}

func NewAnnouncementRepository(db *sql.DB) *AnnouncementRepository {
	return &AnnouncementRepository{db: db}
}

const announcementColumns = `a.id, a.title, a.content, a.telegram_file_id, a.filename, a.file_type,
	a.admin_id, a.teacher_id, a.created_at, a.is_active, COALESCE(a.branch_id, 0)`

func scanAnnouncement(row interface{ Scan(...any) error }) (*models.Announcement, error) {
	var a models.Announcement
	err := row.Scan(
		&a.ID,
		&a.Title,
		&a.Content,
		&a.TelegramFileID,
		&a.Filename,
		&a.FileType,
		&a.PostedByAdminID,
		&a.PostedByTeacherID,
		&a.CreatedAt,
		&a.IsActive,
		&a.BranchID,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AnnouncementRepository) getMany(query string, args ...any) ([]*models.Announcement, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get announcements: %w", err)
	}
	defer rows.Close()

	var announcements []*models.Announcement
	for rows.Next() {
		a, err := scanAnnouncement(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan announcement: %w", err)
		}
		announcements = append(announcements, a)
	}
	return announcements, rows.Err()
}

// Create creates a new announcement and its class targets in one transaction
func (r *AnnouncementRepository) Create(req *models.CreateAnnouncementRequest) (*models.Announcement, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO announcements (title, content, telegram_file_id, filename, file_type, admin_id, teacher_id, branch_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, req.Title, req.Content, req.TelegramFileID, req.Filename, req.FileType,
		req.PostedByAdminID, req.PostedByTeacherID, req.BranchID)
	if err != nil {
		return nil, fmt.Errorf("failed to create announcement: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	for _, classID := range req.ClassIDs {
		if _, err := tx.Exec("INSERT OR IGNORE INTO announcement_classes (announcement_id, class_id) VALUES (?, ?)", id, classID); err != nil {
			return nil, fmt.Errorf("failed to link announcement class: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetByID(int(id))
}

// GetByID gets announcement by ID
func (r *AnnouncementRepository) GetByID(id int) (*models.Announcement, error) {
	a, err := scanAnnouncement(r.db.QueryRow("SELECT "+announcementColumns+" FROM announcements a WHERE a.id = ?", id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get announcement: %w", err)
	}
	return a, nil
}

// GetActive gets active announcements of a branch (branchID 0 = all branches)
func (r *AnnouncementRepository) GetActive(branchID, limit, offset int) ([]*models.Announcement, error) {
	return r.getMany(`
		SELECT `+announcementColumns+`
		FROM announcements a
		WHERE a.is_active = 1 AND (? = 0 OR a.branch_id = ?)
		ORDER BY a.created_at DESC
		LIMIT ? OFFSET ?
	`, branchID, branchID, limit, offset)
}

// GetActiveForParent gets active announcements a parent should see: branch-wide ones of
// the parent's branch plus class-targeted ones for classes of the parent's children.
func (r *AnnouncementRepository) GetActiveForParent(parentID, branchID, limit, offset int) ([]*models.Announcement, error) {
	return r.getMany(`
		SELECT `+announcementColumns+`
		FROM announcements a
		WHERE a.is_active = 1 AND a.branch_id = ?
		  AND (
			NOT EXISTS (SELECT 1 FROM announcement_classes ac WHERE ac.announcement_id = a.id)
			OR EXISTS (
				SELECT 1
				FROM announcement_classes ac
				JOIN students s ON s.class_id = ac.class_id
				JOIN parent_students ps ON ps.student_id = s.id
				WHERE ac.announcement_id = a.id AND ps.parent_id = ?
			)
		  )
		ORDER BY a.created_at DESC
		LIMIT ? OFFSET ?
	`, branchID, parentID, limit, offset)
}

// GetAll gets announcements of a branch with pagination (branchID 0 = all branches)
func (r *AnnouncementRepository) GetAll(branchID, limit, offset int) ([]*models.Announcement, error) {
	return r.getMany(`
		SELECT `+announcementColumns+`
		FROM announcements a
		WHERE (? = 0 OR a.branch_id = ?)
		ORDER BY a.created_at DESC
		LIMIT ? OFFSET ?
	`, branchID, branchID, limit, offset)
}

// GetClassIDs returns the classes an announcement targets (empty = whole branch)
func (r *AnnouncementRepository) GetClassIDs(announcementID int) ([]int, error) {
	rows, err := r.db.Query("SELECT class_id FROM announcement_classes WHERE announcement_id = ?", announcementID)
	if err != nil {
		return nil, fmt.Errorf("failed to get announcement classes: %w", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// Update updates content and attachment of an announcement
func (r *AnnouncementRepository) Update(id int, req *models.CreateAnnouncementRequest) (*models.Announcement, error) {
	_, err := r.db.Exec(`
		UPDATE announcements
		SET title = ?, content = ?, telegram_file_id = ?, filename = ?, file_type = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, req.Title, req.Content, req.TelegramFileID, req.Filename, req.FileType, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update announcement: %w", err)
	}
	return r.GetByID(id)
}

// ToggleActive toggles the active status of an announcement
func (r *AnnouncementRepository) ToggleActive(id int) error {
	query := `UPDATE announcements SET is_active = CASE WHEN is_active = 1 THEN 0 ELSE 1 END WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to toggle announcement active status: %w", err)
	}
	return nil
}

// Delete deletes an announcement
func (r *AnnouncementRepository) Delete(id int) error {
	query := `DELETE FROM announcements WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete announcement: %w", err)
	}
	return nil
}

// Count counts announcements of a branch (branchID 0 = all branches)
func (r *AnnouncementRepository) Count(branchID int) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM announcements WHERE (? = 0 OR branch_id = ?)", branchID, branchID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count announcements: %w", err)
	}
	return count, nil
}

// GetByTeacherID gets all announcements posted by a specific teacher, with target classes
func (r *AnnouncementRepository) GetByTeacherID(teacherID int, limit, offset int) ([]*models.AnnouncementWithClasses, error) {
	announcements, err := r.getMany(`
		SELECT `+announcementColumns+`
		FROM announcements a
		WHERE a.teacher_id = ?
		ORDER BY a.created_at DESC
		LIMIT ? OFFSET ?
	`, teacherID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Class names are loaded after the first result set is closed: with a single
	// SQLite connection a nested query would block forever.
	result := make([]*models.AnnouncementWithClasses, 0, len(announcements))
	for _, a := range announcements {
		item := &models.AnnouncementWithClasses{Announcement: *a}
		rows, err := r.db.Query(`
			SELECT ac.class_id, c.class_name
			FROM announcement_classes ac
			JOIN classes c ON ac.class_id = c.id
			WHERE ac.announcement_id = ?
			ORDER BY c.class_name
		`, a.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get announcement classes: %w", err)
		}
		for rows.Next() {
			var classID int
			var className string
			if err := rows.Scan(&classID, &className); err != nil {
				rows.Close()
				return nil, fmt.Errorf("failed to scan class: %w", err)
			}
			item.ClassIDs = append(item.ClassIDs, classID)
			item.ClassNames = append(item.ClassNames, className)
		}
		rows.Close()
		result = append(result, item)
	}
	return result, nil
}
