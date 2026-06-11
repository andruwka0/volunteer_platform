package store

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/andruwka0/volunteer_platform/internal/domain"
)

type Store struct {
	mu sync.RWMutex

	// users хранит пользователей по их ID
	users map[int64]*domain.User

	// usersByLogin хранит пользователей по логину - для быстрого поиска при авторизации
	usersByLogin map[string]*domain.User

	// events хранит ивенты по ID
	events map[int64]*domain.Event

	eventParticipants map[int64][]int64 // eventID -> []userID
	eventReserve      map[int64][]int64 // eventID -> []userID (лист ожидания)

	// nextID используется для генерации уникальных числовых ID
	nextUserID  int64
	nextEventID int64
}

func New() *Store {
	return &Store{
		users:             make(map[int64]*domain.User),
		usersByLogin:      make(map[string]*domain.User),
		events:            make(map[int64]*domain.Event),
		eventParticipants: make(map[int64][]int64),
		eventReserve:      make(map[int64][]int64),
		nextUserID:        1,
		nextEventID:       1,
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

func (s *Store) GetUserByID(ID int64) (*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if user, ok := s.users[ID]; ok {
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (s *Store) CreateEvent(title, description, location, image string, startDate, endDate time.Time, registrationDeadline *time.Time, maxParticipants *int64, reserveParticipants, skillPoints, createdByID int64) (*domain.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[createdByID]; !ok {
		return nil, domain.ErrUserNotFound
	}
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
	s.nextEventID++
	return event, nil

}

func (s *Store) GetEventByID(eventID int64) (*domain.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if event, ok := s.events[eventID]; ok {
		return event, nil
	}
	return nil, domain.ErrEventNotFound
}

func (s *Store) GetAllEvents() ([]*domain.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	events := make([]*domain.Event, 0, len(s.events))
	for _, event := range s.events {
		events = append(events, event)
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].StartDate.Before(events[j].StartDate)
	})
	return events, nil
}

func (s *Store) UpdateUserSkillPoints(userID int64, pointsToAdd int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return domain.ErrUserNotFound
	}

	user.SkillPoints += pointsToAdd
	return nil
}

func (s *Store) UpdateEventStatus(eventID int64, newStatus string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.events[eventID]
	if !ok {
		return domain.ErrEventNotFound
	}

	event.Status = newStatus
	return nil
}

func (s *Store) UpdateUserRole(userID int64, newRole string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return domain.ErrUserNotFound
	}

	if newRole != domain.RoleAdmin &&
		newRole != domain.RoleOrganizer &&
		newRole != domain.RoleVolunteer {
		return errors.New("недопустимая роль")
	}

	user.Role = newRole
	return nil
}

func (s *Store) RegisterForEvent(eventID, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	event, ok := s.events[eventID]
	if !ok {
		return domain.ErrEventNotFound
	}
	if event.Status != domain.EventRecruiting {
		return errors.New("регистрация на это мероприятие закрыта")
	}
	for _, id := range s.eventParticipants[eventID] {
		if id == userID {
			return errors.New("пользователь уже зарегистрирован")
		}
	}
	for _, id := range s.eventReserve[eventID] {
		if id == userID {
			return errors.New("пользователь уже в резерве")
		}
	}
	maxP := int64(0)
	if event.MaxParticipants != nil {
		maxP = *event.MaxParticipants
	}

	currentParticipants := int64(len(s.eventParticipants[eventID]))

	if currentParticipants < maxP || maxP == 0 {
		// Есть места в основном составе
		s.eventParticipants[eventID] = append(s.eventParticipants[eventID], userID)
		event.ParticipantsCount++
	} else {
		// Мест нет, идем в резерв
		s.eventReserve[eventID] = append(s.eventReserve[eventID], userID)
		event.ReserveCount++
	}

	return nil
}

func (s *Store) GetEventParticipants(eventID int64) ([]int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.events[eventID]; !ok {
		return nil, domain.ErrEventNotFound
	}
	participantIDs := s.eventParticipants[eventID]

	participants := make([]int64, len(participantIDs))
	copy(participants, participantIDs)

	return participants, nil
}

func (s *Store) AwardPointsToAllParticipants(eventID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	event, ok := s.events[eventID]
	if !ok {
		return domain.ErrEventNotFound
	}
	if event.SkillPoints <= 0 {
		return nil
	}
	for _, id := range s.eventParticipants[eventID] {
		if user, exists := s.users[id]; exists {
			user.SkillPoints += event.SkillPoints
		}
	}
	return nil
}

func (s *Store) CancelRegistration(eventID, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	event, ok := s.events[eventID]
	if !ok {
		return domain.ErrEventNotFound
	}
	if event.Status != domain.EventRecruiting {
		return errors.New("нельзя отменить регистрацию на уже начавшееся мероприятие")
	}
	participants := s.eventParticipants[eventID]
	found := false
	for i, id := range participants {
		if id == userID {
			s.eventParticipants[eventID] = append(participants[:i], participants[i+1:]...)
			event.ParticipantsCount--
			found = true
			break
		}
	}
	if !found {
		reserve := s.eventReserve[eventID]
		for i, id := range reserve {
			if id == userID {
				s.eventReserve[eventID] = append(reserve[:i], reserve[i+1:]...)
				event.ReserveCount--
				found = true
				break
			}
		}
	}
	if !found {
		return errors.New("пользователь не зарегистрирован на это мероприятие")
	}

	maxP := int64(0)
	if event.MaxParticipants != nil {
		maxP = *event.MaxParticipants
	}

	if (maxP == 0 || int64(len(s.eventParticipants[eventID])) < maxP) && len(s.eventReserve[eventID]) > 0 {
		nextInLine := s.eventReserve[eventID][0]
		s.eventReserve[eventID] = s.eventReserve[eventID][1:]
		event.ReserveCount--
		s.eventParticipants[eventID] = append(s.eventParticipants[eventID], nextInLine)
		event.ParticipantsCount++
	}

	return nil
}
