package repository

import "volunteer-platform/internal/domain"

// Event ищет мероприятие по ID и возвращает указатель на запись в store.
func (s *JSONStore) Event(id int) *domain.Event {
	for i := range s.Events {
		if s.Events[i].ID == id {
			return &s.Events[i]
		}
	}
	return nil
}
