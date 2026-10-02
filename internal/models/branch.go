package models

// Branch is a school branch (filial). Classes, teachers, parents, admins and
// announcements belong to exactly one branch.
type Branch struct {
	ID       int    `json:"id" db:"id"`
	Code     string `json:"code" db:"code"`
	NameUz   string `json:"name_uz" db:"name_uz"`
	NameRu   string `json:"name_ru" db:"name_ru"`
	IsActive bool   `json:"is_active" db:"is_active"`
}

// Name returns the branch name in the given language ("uz" or "ru").
func (b *Branch) Name(lang string) string {
	if lang == "ru" {
		return b.NameRu
	}
	return b.NameUz
}
