package storage

import (
	"fmt"
	"sync"
	"volunteer-platform/backend/internal/domain"
)

type Store struct {
	mu sync.RWMutex
	// users хранит пользователей по их ID
	users map[int64]*domain.User

	// usersByLogin хранит пользователей по логину - для быстрого поиска при авторизации
	usersByLogin map[string]*domain.User

	// events хранит ивенты по ID
	events map[int64]*domain.Event

	// notifications по userID
	notifications map[int64]*domain.Notification

	// SP по userID
	skillPoints map[int64]int64

	// nextID используется для генерации уникальных числовых ID
	nextID int64
}

func New() *Store {
	return &Store{
		users:         make(map[int64]*domain.User),
		usersByLogin:  make(map[string]*domain.User),
		events:        make(map[int64]*domain.Event),
		notifications: make(map[int64]*domain.Notification),
		skillPoints:   make(map[int64]int64),
		nextID:        1,
	}
}

func (s *Store) CreateUser(login, passwordHash, firstname, lastname, telegram string) (*domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[s.nextID]; ok {
		return nil, fmt.Errorf("user %d already exists", s.nextID)
	}
	user := &domain.User{
		Login:       login,
		Password:    passwordHash,
		FirstName:   firstname,
		LastName:    lastname,
		Telegram:    telegram,
		ID:          s.nextID,
		Role:        domain.RoleVolunteer,
		SkillPoints: 0,
	}
	s.users[s.nextID] = user
	s.usersByLogin[login] = user
	s.skillPoints[s.nextID] = user.SkillPoints
	s.nextID++
	return user, nil
}
