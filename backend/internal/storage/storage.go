package storage

import (
	"sync"
	"time"
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

	// SP по userID
	skillPoints map[int64]int64

	// nextID используется для генерации уникальных числовых ID
	nextUserID  int64
	nextEventID int64
}

func New() *Store {
	return &Store{
		users:        make(map[int64]*domain.User),
		usersByLogin: make(map[string]*domain.User),
		events:       make(map[int64]*domain.Event),
		skillPoints:  make(map[int64]int64),
		nextUserID:   1,
		nextEventID:  1,
	}
}

func (s *Store) CreateUser(login, passwordHash, firstname, lastname, telegram string) (*domain.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.usersByLogin[login]; ok {
		return nil, domain.ErrUserExists
	}
	user := &domain.User{
		Login:       login,
		Password:    passwordHash,
		FirstName:   firstname,
		LastName:    lastname,
		Telegram:    telegram,
		ID:          s.nextUserID,
		Role:        domain.RoleVolunteer,
		SkillPoints: 0,
	}
	s.users[s.nextUserID] = user
	s.usersByLogin[login] = user
	s.skillPoints[s.nextUserID] = user.SkillPoints
	s.nextUserID++
	return user, nil
}

func (s *Store) GetUserByLogin(login string) (*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if user, ok := s.usersByLogin[login]; ok {
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (s *Store) GetUserByID(userID int64) (*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if user, ok := s.users[userID]; ok {
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (s *Store) CreateEvent(title, description, location, image string, startDate, endDate time.Time, registrationDeadline *time.Time, maxParticipants *int64, reserveParticipants, skillPoints, createdByID int64) (*domain.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	event := &domain.Event{
		ID:                   s.nextEventID,
		Title:                title,
		Description:          description,
		Location:             location,
		CoverImageURL:        image,
		Status:               domain.EventRecruiting,
		StartDate:            startDate,
		EndDate:              endDate,
		RegistrationDeadline: registrationDeadline,
		MaxParticipants:      maxParticipants,
		ReserveParticipants:  reserveParticipants,
		SkillPoints:          skillPoints,
		CreatedByID:          createdByID,
		ParticipantsCount:    0,
		ReserveCount:         0,
	}
	s.events[s.nextEventID] = event
	return event, nil

}
