package repository

import "volunteer-platform/internal/domain"

// User ищет пользователя по ID и возвращает указатель на запись в store.
func (s *JSONStore) User(id int) *domain.User {
	for i := range s.Users {
		if s.Users[i].ID == id {
			return &s.Users[i]
		}
	}
	return nil
}
