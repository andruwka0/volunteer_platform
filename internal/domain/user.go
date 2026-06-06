package domain

import "time"

var Roles = []string{"leader", "organizer", "junior_volunteer", "middle_volunteer", "senior_volunteer"}

// Now возвращает текущее UTC-время для доменной логики.
func Now() time.Time { return time.Now().UTC() }

// HasRole проверяет, входит ли роль пользователя в разрешённый набор.
func HasRole(role string, roles ...string) bool {
	for _, r := range roles {
		if role == r {
			return true
		}
	}
	return false
}

type User struct {
	ID                                                                                    int
	Username, PasswordHash, FirstName, LastName, FullName, Telegram, Role, AvatarURL, Bio string
	SkillPoints                                                                           int
	IsActive                                                                              bool
	CreatedAt, UpdatedAt                                                                  time.Time
	LastLoginAt                                                                           *time.Time
}
