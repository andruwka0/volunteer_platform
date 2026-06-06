package repository

import "volunteer-platform/internal/domain"

type CRUD interface {
	Save()
	NextID(entity string) int
	User(id int) *domain.User
	Event(id int) *domain.Event
}
