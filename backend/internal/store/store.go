package store

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/andruwka0/volunteer_platform/internal/domain"
)

type Store struct {
	mu sync.RWMutex

	// Основная сущность — доступ по её ID
	users   map[int64]*domain.User   // userID → User
	events  map[int64]*domain.Event  // eventID → Event
	rewards map[int64]*domain.Reward // rewardID → Reward

	// Вторичные индексы — для быстрого поиска по альтернативному ключу
	usersByLogin map[string]*domain.User // login → User (для логина)

	// Связи "один-ко-многим" — ключ = "владелец" списка
	eventParticipants map[int64][]int64              // eventID → []userID (кто участвует в ивенте)
	eventReserve      map[int64][]int64              // eventID → []userID (резерв ивента)
	userEvents        map[int64][]*domain.UserEvent  // userID → []UserEvent (история юзера)
	userRewards       map[int64][]*domain.UserReward // userID → []UserReward (награды юзера)

	// Матрица доступа
	organizerBlacklists map[int64]map[int64]bool // organizerID → (userID → banned)

	// Шаблоны ивентов — глобальные, все видят
	eventTemplates map[int64]*domain.EventTemplate // templateID → EventTemplate

	// Список доступных картинок (из папки app/static/images)
	availableImages []string // URL-пути к картинкам

	// История транзакций SP
	skillPointTransactions map[int64][]*domain.SkillPointTransaction // userID → []SkillPointTransaction (транзакции юзера)

	// Счётчики ID для каждой сущности
	nextUserID        int64
	nextEventID       int64
	nextRewardID      int64
	nextTemplateID    int64
	nextUserEventID   int64
	nextUserRewardID  int64
	nextTransactionID int64
}

func New() *Store {
	s := &Store{
		users:                  make(map[int64]*domain.User),
		usersByLogin:           make(map[string]*domain.User),
		events:                 make(map[int64]*domain.Event),
		eventParticipants:      make(map[int64][]int64),
		eventReserve:           make(map[int64][]int64),
		userEvents:             make(map[int64][]*domain.UserEvent),
		rewards:                make(map[int64]*domain.Reward),
		userRewards:            make(map[int64][]*domain.UserReward),
		eventTemplates:         make(map[int64]*domain.EventTemplate),
		organizerBlacklists:    make(map[int64]map[int64]bool),
		availableImages:        []string{},
		skillPointTransactions: make(map[int64][]*domain.SkillPointTransaction),
		nextUserID:             1,
		nextEventID:            1,
		nextRewardID:           1,
		nextTemplateID:         1,
		nextUserEventID:        1,
		nextUserRewardID:       1,
		nextTransactionID:      1,
	}
	if err := s.LoadAvailableImages(); err != nil {
		log.Printf("Warning: failed to load available images: %v", err)
	}
	return s
}

// ==================== ПОЛЬЗОВАТЕЛИ ====================

func (s *Store) CreateUser(login, passwordHash, firstname, lastname, middlename, telegram string) (*domain.User, error) {
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
		MiddleName:  middlename,
		Telegram:    telegram,
		ID:          s.nextUserID,
		Role:        domain.RoleVolunteer,
		SkillPoints: 0,
		CreatedAt:   time.Now(),
	}

	s.users[s.nextUserID] = user
	s.usersByLogin[login] = user
	s.userEvents[s.nextUserID] = []*domain.UserEvent{}
	s.userRewards[s.nextUserID] = []*domain.UserReward{}
	s.skillPointTransactions[s.nextUserID] = []*domain.SkillPointTransaction{}
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

func (s *Store) UpdateUserRole(userID int64, newRole string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return domain.ErrUserNotFound
	}
	user.Role = newRole
	return nil
}

func (s *Store) AddSkillPoints(userID int64, points int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return domain.ErrUserNotFound
	}
	user.SkillPoints += points
	return nil
}

func (s *Store) DeductSkillPoints(userID int64, points int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[userID]
	if !ok {
		return domain.ErrUserNotFound
	}
	user.SkillPoints -= points
	return nil
}

func (s *Store) SearchUsers(firstName, lastName, middleName string) ([]*domain.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*domain.User
	for _, user := range s.users {
		if !strings.EqualFold(user.FirstName, firstName) ||
			!strings.EqualFold(user.LastName, lastName) {
			continue
		}
		if middleName != "" && !strings.EqualFold(user.MiddleName, middleName) {
			continue
		}
		results = append(results, user)
	}
	return results, nil
}

// ==================== ИВЕНТЫ ====================

func (s *Store) CreateEvent(title, description, location, image string, startDate, endDate time.Time, registrationDeadline *time.Time, maxParticipants *int64, reserveParticipants, skillPoints, createdByID int64, templateID *int64) (*domain.Event, error) {
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
		TemplateID:           templateID,
		CreatedAt:            time.Now(),
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

// ==================== УЧАСТНИКИ ====================

func (s *Store) AddParticipant(eventID, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.eventParticipants[eventID] = append(s.eventParticipants[eventID], userID)
	s.events[eventID].ParticipantsCount++
	return nil
}

func (s *Store) AddToReserve(eventID, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.eventReserve[eventID] = append(s.eventReserve[eventID], userID)
	s.events[eventID].ReserveCount++
	return nil
}

func (s *Store) RemoveParticipant(eventID, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	participants := s.eventParticipants[eventID]
	for i, id := range participants {
		if id == userID {
			s.eventParticipants[eventID] = append(participants[:i], participants[i+1:]...)
			s.events[eventID].ParticipantsCount--
			return nil
		}
	}
	return nil
}

func (s *Store) RemoveFromReserve(eventID, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	reserve := s.eventReserve[eventID]
	for i, id := range reserve {
		if id == userID {
			s.eventReserve[eventID] = append(reserve[:i], reserve[i+1:]...)
			s.events[eventID].ReserveCount--
			return nil
		}
	}
	return nil
}

func (s *Store) GetParticipants(eventID int64) ([]int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.eventParticipants[eventID], nil
}

func (s *Store) GetReserve(eventID int64) ([]int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.eventReserve[eventID], nil
}

func (s *Store) PromoteFromReserve(eventID int64) (int64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.eventReserve[eventID]) == 0 {
		return 0, false
	}

	nextInLine := s.eventReserve[eventID][0]
	s.eventReserve[eventID] = s.eventReserve[eventID][1:]
	s.events[eventID].ReserveCount--

	s.eventParticipants[eventID] = append(s.eventParticipants[eventID], nextInLine)
	s.events[eventID].ParticipantsCount++

	return nextInLine, true
}

func (s *Store) IsParticipant(eventID, userID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, id := range s.eventParticipants[eventID] {
		if id == userID {
			return true
		}
	}
	return false
}

func (s *Store) IsInReserve(eventID, userID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, id := range s.eventReserve[eventID] {
		if id == userID {
			return true
		}
	}
	return false
}

// ==================== ИСТОРИЯ ИВЕНТОВ ====================

func (s *Store) AddUserEvent(userID, eventID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.userEvents[userID] = append(s.userEvents[userID], &domain.UserEvent{
		ID:                  s.nextUserEventID,
		UserID:              userID,
		EventID:             eventID,
		JoinedAt:            time.Now(),
		AttendanceConfirmed: false,
	})
	s.nextUserEventID++
	return nil
}

func (s *Store) RemoveUserEvent(userID, eventID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, ue := range s.userEvents[userID] {
		if ue.EventID == eventID {
			s.userEvents[userID] = append(s.userEvents[userID][:i], s.userEvents[userID][i+1:]...)
			return nil
		}
	}
	return nil
}

func (s *Store) GetUserEvents(userID int64) ([]*domain.UserEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.userEvents[userID], nil
}

func (s *Store) ConfirmUserEventAttendance(userID, eventID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ue := range s.userEvents[userID] {
		if ue.EventID == eventID {
			ue.AttendanceConfirmed = true
			return nil
		}
	}
	return nil
}

func (s *Store) IsAttendanceConfirmed(userID, eventID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, ue := range s.userEvents[userID] {
		if ue.EventID == eventID {
			return ue.AttendanceConfirmed
		}
	}
	return false
}

// ==================== БАН-ЛИСТ ====================

func (s *Store) BanUser(organizerID, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.organizerBlacklists[organizerID]; !ok {
		s.organizerBlacklists[organizerID] = make(map[int64]bool)
	}
	s.organizerBlacklists[organizerID][userID] = true
	return nil
}

func (s *Store) UnbanUser(organizerID, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if blacklist, ok := s.organizerBlacklists[organizerID]; ok {
		delete(blacklist, userID)
	}
	return nil
}

func (s *Store) IsUserBanned(organizerID, userID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if blacklist, ok := s.organizerBlacklists[organizerID]; ok {
		return blacklist[userID]
	}
	return false
}

// ==================== НАГРАДЫ ====================

func (s *Store) CreateReward(name, description, imageURL string, cost int64) (*domain.Reward, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	reward := &domain.Reward{
		ID:          s.nextRewardID,
		Name:        name,
		Description: description,
		Cost:        cost,
		ImageURL:    imageURL,
		CreatedAt:   time.Now(),
	}

	s.rewards[s.nextRewardID] = reward
	s.nextRewardID++

	return reward, nil
}

func (s *Store) GetRewardByID(rewardID int64) (*domain.Reward, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if reward, ok := s.rewards[rewardID]; ok {
		return reward, nil
	}
	return nil, domain.ErrRewardNotFound
}

func (s *Store) GetAllRewards() ([]*domain.Reward, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rewards := make([]*domain.Reward, 0, len(s.rewards))
	for _, r := range s.rewards {
		rewards = append(rewards, r)
	}
	return rewards, nil
}

func (s *Store) AddUserReward(userID, rewardID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.userRewards[userID] = append(s.userRewards[userID], &domain.UserReward{
		ID:        s.nextUserRewardID,
		UserID:    userID,
		RewardID:  rewardID,
		ClaimedAt: time.Now(),
		PickedUp:  false,
	})
	s.nextUserRewardID++
	return nil
}

func (s *Store) GetUserRewards(userID int64) ([]*domain.UserReward, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.userRewards[userID], nil
}

func (s *Store) HasUserReward(userID, rewardID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, ur := range s.userRewards[userID] {
		if ur.RewardID == rewardID {
			return true
		}
	}
	return false
}

func (s *Store) ConfirmRewardPickup(userID, rewardID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ur := range s.userRewards[userID] {
		if ur.RewardID == rewardID {
			ur.PickedUp = true
			return nil
		}
	}
	return domain.ErrRewardNotFound
}

// ==================== ШАБЛОНЫ ====================

func (s *Store) CreateEventTemplate(title, description, location, imageURL string, duration time.Duration, maxParticipants *int64, reserveParticipants, skillPoints int64) (*domain.EventTemplate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	template := &domain.EventTemplate{
		ID:                  s.nextTemplateID,
		Title:               title,
		Description:         description,
		Location:            location,
		CoverImageURL:       imageURL,
		Duration:            duration,
		MaxParticipants:     maxParticipants,
		ReserveParticipants: reserveParticipants,
		SkillPoints:         skillPoints,
		CreatedAt:           time.Now(),
	}

	s.eventTemplates[s.nextTemplateID] = template
	s.nextTemplateID++

	return template, nil
}

func (s *Store) GetEventTemplateByID(templateID int64) (*domain.EventTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if template, ok := s.eventTemplates[templateID]; ok {
		return template, nil
	}
	return nil, domain.ErrTemplateNotFound
}

func (s *Store) GetAllEventTemplates() ([]*domain.EventTemplate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	templates := make([]*domain.EventTemplate, 0, len(s.eventTemplates))
	for _, t := range s.eventTemplates {
		templates = append(templates, t)
	}
	return templates, nil
}

// ==================== КАРТИНКИ ====================

func (s *Store) LoadAvailableImages() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	staticDir := "app/static/images"

	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		s.availableImages = []string{}
		return nil
	}

	images := make([]string, 0)

	err := filepath.WalkDir(staticDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		validExtensions := map[string]bool{
			".png": true, ".jpg": true, ".jpeg": true,
			".gif": true, ".webp": true, ".svg": true,
		}

		if validExtensions[ext] {
			urlPath := filepath.ToSlash(path)
			urlPath = strings.TrimPrefix(urlPath, "app")
			urlPath = "/static" + urlPath
			images = append(images, urlPath)
		}
		return nil
	})

	if err != nil {
		return err
	}

	s.availableImages = images
	return nil
}

func (s *Store) GetAvailableImages() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.availableImages
}

// ==================== ТРАНЗАКЦИИ SP ====================

func (s *Store) AddSkillPointTransaction(userID int64, points int64, transactionType string, reason string, eventID *int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.skillPointTransactions[userID] = append(s.skillPointTransactions[userID], &domain.SkillPointTransaction{
		ID:        s.nextTransactionID,
		UserID:    userID,
		Points:    points,
		Type:      transactionType,
		Reason:    reason,
		EventID:   eventID,
		CreatedAt: time.Now(),
	})
	s.nextTransactionID++

	return nil
}

func (s *Store) GetUserSkillPointHistory(userID int64) ([]*domain.SkillPointTransaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.users[userID]; !ok {
		return nil, domain.ErrUserNotFound
	}

	return s.skillPointTransactions[userID], nil
}
