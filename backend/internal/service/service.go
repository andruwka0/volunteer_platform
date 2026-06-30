package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/andruwka0/volunteer_platform/internal/auth"
	"github.com/andruwka0/volunteer_platform/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type repository interface {
	// Пользователи
	CreateUser(login, passwordHash, firstname, lastname, middlename, telegram string) (*domain.User, error)
	GetUserByLogin(login string) (*domain.User, error)
	GetUserByID(ID int64) (*domain.User, error)
	UpdateUserRole(userID int64, newRole string) error
	AddSkillPoints(userID int64, points int64) error
	DeductSkillPoints(userID int64, points int64) error
	SearchUsers(firstName, lastName, middleName string) ([]*domain.User, error)

	// Ивенты
	CreateEvent(title, description, location, image string, startDate, endDate time.Time, registrationDeadline *time.Time, maxParticipants *int64, reserveParticipants, skillPoints, createdByID int64, templateID *int64) (*domain.Event, error)
	GetEventByID(eventID int64) (*domain.Event, error)
	GetAllEvents() ([]*domain.Event, error)
	UpdateEventStatus(eventID int64, newStatus string) error

	// Участники
	AddParticipant(eventID, userID int64) error
	AddToReserve(eventID, userID int64) error
	RemoveParticipant(eventID, userID int64) error
	RemoveFromReserve(eventID, userID int64) error
	GetParticipants(eventID int64) ([]int64, error)
	GetReserve(eventID int64) ([]int64, error)
	PromoteFromReserve(eventID int64) (int64, bool)
	IsParticipant(eventID, userID int64) bool
	IsInReserve(eventID, userID int64) bool

	// История ивентов
	AddUserEvent(userID, eventID int64) error
	RemoveUserEvent(userID, eventID int64) error
	GetUserEvents(userID int64) ([]*domain.UserEvent, error)
	ConfirmUserEventAttendance(userID, eventID int64) error
	IsAttendanceConfirmed(userID, eventID int64) bool

	// Бан-лист
	BanUser(organizerID, userID int64) error
	UnbanUser(organizerID, userID int64) error
	IsUserBanned(organizerID, userID int64) bool

	// Награды
	CreateReward(name, description, imageURL string, cost int64) (*domain.Reward, error)
	GetRewardByID(rewardID int64) (*domain.Reward, error)
	GetAllRewards() ([]*domain.Reward, error)
	AddUserReward(userID, rewardID int64) error
	GetUserRewards(userID int64) ([]*domain.UserReward, error)
	HasUserReward(userID, rewardID int64) bool
	ConfirmRewardPickup(userID, rewardID int64) error

	// Шаблоны
	CreateEventTemplate(title, description, location, imageURL string, duration time.Duration, maxParticipants *int64, reserveParticipants, skillPoints int64) (*domain.EventTemplate, error)
	GetEventTemplateByID(templateID int64) (*domain.EventTemplate, error)
	GetAllEventTemplates() ([]*domain.EventTemplate, error)

	// Картинки
	GetAvailableImages() []string

	// Транзакции SP
	AddSkillPointTransaction(userID int64, points int64, transactionType string, reason string, eventID *int64) error
	GetUserSkillPointHistory(userID int64) ([]*domain.SkillPointTransaction, error)
}

type Service struct {
	repo repository
}

func New(repo repository) *Service {
	return &Service{repo: repo}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ==================== АУТЕНТИФИКАЦИЯ ====================

func (s *Service) RegisterUser(login, password, firstname, lastname, middlename, telegram string) (string, error) {
	passwordHash, err := HashPassword(password)
	if err != nil {
		return "", fmt.Errorf("ошибка хэширования пароля: %w", err)
	}

	user, err := s.repo.CreateUser(login, passwordHash, firstname, lastname, middlename, telegram)
	if err != nil {
		return "", err
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации токена: %w", err)
	}
	return token, nil
}

func (s *Service) LoginUser(login, password string) (string, error) {
	user, err := s.repo.GetUserByLogin(login)
	if err != nil {
		return "", domain.ErrUserNotFound
	}

	if !CheckPassword(password, user.Password) {
		return "", domain.ErrInvalidPassword
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации токена: %w", err)
	}
	return token, nil
}

// ==================== ПОЛЬЗОВАТЕЛИ ====================

func (s *Service) GetUserByID(id int64) (*domain.User, error) {
	return s.repo.GetUserByID(id)
}

func (s *Service) SearchUsers(firstName, lastName, middleName string) ([]*domain.User, error) {
	if firstName == "" || lastName == "" {
		return nil, errors.New("имя и фамилия обязательны")
	}
	return s.repo.SearchUsers(firstName, lastName, middleName)
}

func (s *Service) PromoteUser(targetUserID int64, newRole string, requesterID int64) error {
	requester, err := s.repo.GetUserByID(requesterID)
	if err != nil {
		return err
	}
	if requester.Role != domain.RoleAdmin {
		return domain.ErrInvalidRole
	}
	if newRole != domain.RoleOrganizer && newRole != domain.RoleAdmin {
		return domain.ErrInvalidPromotion
	}
	if targetUserID == requesterID && newRole != domain.RoleAdmin {
		return domain.ErrCannotPromoteSelf
	}
	if _, err := s.repo.GetUserByID(targetUserID); err != nil {
		return domain.ErrUserNotFound
	}
	return s.repo.UpdateUserRole(targetUserID, newRole)
}

func (s *Service) AwardUser(targetUserID int64, points int64, reason string, requesterID int64) error {
	requester, err := s.repo.GetUserByID(requesterID)
	if err != nil {
		return err
	}
	if requester.Role != domain.RoleAdmin {
		return domain.ErrInvalidRole
	}
	if points <= 0 {
		return errors.New("количество баллов должно быть положительным")
	}
	if reason == "" {
		return errors.New("причина начисления обязательна")
	}
	if _, err := s.repo.GetUserByID(targetUserID); err != nil {
		return domain.ErrUserNotFound
	}

	if err := s.repo.AddSkillPoints(targetUserID, points); err != nil {
		return err
	}

	return s.repo.AddSkillPointTransaction(targetUserID, points, domain.TransactionTypeManual, reason, nil)
}

// ==================== ИВЕНТЫ ====================

func (s *Service) CreateEvent(title, description, location, image string,
	startDate, endDate time.Time, registrationDeadline *time.Time,
	maxParticipants *int64, reserveParticipants, skillPoints, createdByID int64,
	templateID *int64) (*domain.Event, error) {

	creator, err := s.repo.GetUserByID(createdByID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	if creator.Role != domain.RoleAdmin && creator.Role != domain.RoleOrganizer {
		return nil, domain.ErrInvalidRole
	}
	if !endDate.After(startDate) {
		return nil, domain.ErrInvalidDates
	}

	return s.repo.CreateEvent(title, description, location, image,
		startDate, endDate, registrationDeadline, maxParticipants,
		reserveParticipants, skillPoints, createdByID, templateID)
}

func (s *Service) GetEventByID(eventID int64) (*domain.Event, error) {
	return s.repo.GetEventByID(eventID)
}

func (s *Service) GetAllEvents(userID int64) ([]*domain.Event, error) {
	events, err := s.repo.GetAllEvents()
	if err != nil {
		return nil, err
	}

	filtered := make([]*domain.Event, 0, len(events))
	for _, event := range events {
		if !s.repo.IsUserBanned(event.CreatedByID, userID) {
			filtered = append(filtered, event)
		}
	}
	return filtered, nil
}

func (s *Service) ApproveAndAwardPoints(eventID int64, adminID int64) error {
	admin, err := s.repo.GetUserByID(adminID)
	if err != nil {
		return domain.ErrUserNotFound
	}
	if admin.Role != domain.RoleAdmin {
		return domain.ErrInvalidRole
	}

	event, err := s.repo.GetEventByID(eventID)
	if err != nil {
		return domain.ErrEventNotFound
	}
	if event.Status != domain.EventFinished {
		return domain.ErrEventNotFinished
	}

	participants, err := s.repo.GetParticipants(eventID)
	if err != nil {
		return err
	}

	for _, userID := range participants {
		if s.repo.IsAttendanceConfirmed(userID, eventID) {
			if err := s.repo.AddSkillPoints(userID, event.SkillPoints); err != nil {
				return fmt.Errorf("ошибка начисления баллов пользователю %d: %w", userID, err)
			}

			reason := fmt.Sprintf("Участие в мероприятии: %s", event.Title)
			eventIDCopy := eventID
			if err := s.repo.AddSkillPointTransaction(userID, event.SkillPoints, domain.TransactionTypeEvent, reason, &eventIDCopy); err != nil {
				return fmt.Errorf("ошибка создания транзакции для пользователя %d: %w", userID, err)
			}
		}
	}

	return s.repo.UpdateEventStatus(eventID, domain.EventClosed)
}

// ==================== РЕГИСТРАЦИЯ НА ИВЕНТ ====================

func (s *Service) RegisterForEvent(eventID, userID int64) error {
	if _, err := s.repo.GetUserByID(userID); err != nil {
		return domain.ErrUserNotFound
	}

	event, err := s.repo.GetEventByID(eventID)
	if err != nil {
		return domain.ErrEventNotFound
	}
	if event.Status != domain.EventRecruiting {
		return domain.ErrRegistrationClosed
	}
	if event.CreatedByID == userID {
		return domain.ErrOrganizerSelfReg
	}
	if s.repo.IsUserBanned(event.CreatedByID, userID) {
		return errors.New("регистрация на это мероприятие закрыта")
	}
	if s.repo.IsParticipant(eventID, userID) {
		return errors.New("пользователь уже зарегистрирован")
	}
	if s.repo.IsInReserve(eventID, userID) {
		return errors.New("пользователь уже в резерве")
	}

	maxP := int64(0)
	if event.MaxParticipants != nil {
		maxP = *event.MaxParticipants
	}
	participants, _ := s.repo.GetParticipants(eventID)
	currentCount := int64(len(participants))

	if currentCount < maxP || maxP == 0 {
		if err := s.repo.AddParticipant(eventID, userID); err != nil {
			return err
		}
	} else {
		if err := s.repo.AddToReserve(eventID, userID); err != nil {
			return err
		}
	}

	return s.repo.AddUserEvent(userID, eventID)
}

func (s *Service) CancelRegistration(eventID, userID int64) error {
	event, err := s.repo.GetEventByID(eventID)
	if err != nil {
		return domain.ErrEventNotFound
	}
	if event.Status != domain.EventRecruiting {
		return errors.New("нельзя отменить регистрацию на уже начавшееся мероприятие")
	}

	if s.repo.IsParticipant(eventID, userID) {
		if err := s.repo.RemoveParticipant(eventID, userID); err != nil {
			return err
		}
		s.repo.PromoteFromReserve(eventID)
	} else if s.repo.IsInReserve(eventID, userID) {
		if err := s.repo.RemoveFromReserve(eventID, userID); err != nil {
			return err
		}
	} else {
		return errors.New("пользователь не зарегистрирован на это мероприятие")
	}

	return s.repo.RemoveUserEvent(userID, eventID)
}

func (s *Service) GetEventParticipants(eventID int64) ([]*domain.ParticipantInfo, error) {
	if _, err := s.repo.GetEventByID(eventID); err != nil {
		return nil, domain.ErrEventNotFound
	}

	participantIDs, err := s.repo.GetParticipants(eventID)
	if err != nil {
		return nil, err
	}

	participants := make([]*domain.ParticipantInfo, 0, len(participantIDs))
	for _, id := range participantIDs {
		user, err := s.repo.GetUserByID(id)
		if err != nil {
			continue
		}
		participants = append(participants, &domain.ParticipantInfo{
			UserID:              id,
			FirstName:           user.FirstName,
			LastName:            user.LastName,
			MiddleName:          user.MiddleName,
			Telegram:            user.Telegram,
			AttendanceConfirmed: s.repo.IsAttendanceConfirmed(id, eventID),
		})
	}
	return participants, nil
}

// ==================== ПОДТВЕРЖДЕНИЕ ПОСЕЩАЕМОСТИ ====================

func (s *Service) ConfirmAttendance(eventID, userID, organizerID int64) error {
	event, err := s.repo.GetEventByID(eventID)
	if err != nil {
		return domain.ErrEventNotFound
	}
	if event.CreatedByID != organizerID {
		return domain.ErrNotEventCreator
	}
	if !s.repo.IsParticipant(eventID, userID) {
		return errors.New("пользователь не зарегистрирован на это мероприятие")
	}
	if s.repo.IsAttendanceConfirmed(userID, eventID) {
		return domain.ErrAlreadyConfirmed
	}
	return s.repo.ConfirmUserEventAttendance(userID, eventID)
}

// ==================== ИСТОРИЯ ИВЕНТОВ ====================

func (s *Service) GetUserEventHistory(userID int64) ([]*domain.UserEventWithDetails, error) {
	if _, err := s.repo.GetUserByID(userID); err != nil {
		return nil, domain.ErrUserNotFound
	}

	userEvents, err := s.repo.GetUserEvents(userID)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.UserEventWithDetails, 0, len(userEvents))
	for _, ue := range userEvents {
		event, err := s.repo.GetEventByID(ue.EventID)
		if err != nil {
			continue
		}
		result = append(result, &domain.UserEventWithDetails{
			EventID:             ue.EventID,
			Title:               event.Title,
			Status:              event.Status,
			JoinedAt:            ue.JoinedAt,
			AttendanceConfirmed: ue.AttendanceConfirmed,
			SkillPoints:         event.SkillPoints,
		})
	}
	return result, nil
}

// ==================== БАН-ЛИСТ ====================

func (s *Service) BanUser(organizerID, userID int64) error {
	organizer, err := s.repo.GetUserByID(organizerID)
	if err != nil {
		return domain.ErrUserNotFound
	}
	if organizer.Role != domain.RoleOrganizer && organizer.Role != domain.RoleAdmin {
		return domain.ErrInvalidRole
	}
	if _, err := s.repo.GetUserByID(userID); err != nil {
		return domain.ErrUserNotFound
	}
	return s.repo.BanUser(organizerID, userID)
}

func (s *Service) UnbanUser(organizerID, userID int64) error {
	organizer, err := s.repo.GetUserByID(organizerID)
	if err != nil {
		return domain.ErrUserNotFound
	}
	if organizer.Role != domain.RoleOrganizer && organizer.Role != domain.RoleAdmin {
		return domain.ErrInvalidRole
	}
	return s.repo.UnbanUser(organizerID, userID)
}

// ==================== НАГРАДЫ ====================

func (s *Service) CreateReward(name, description, imageURL string, cost int64, requesterID int64) (*domain.Reward, error) {
	requester, err := s.repo.GetUserByID(requesterID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	if requester.Role != domain.RoleAdmin {
		return nil, domain.ErrInvalidRole
	}
	if cost <= 0 {
		return nil, errors.New("стоимость должна быть положительной")
	}
	return s.repo.CreateReward(name, description, imageURL, cost)
}

func (s *Service) GetAllRewards() ([]*domain.Reward, error) {
	return s.repo.GetAllRewards()
}

func (s *Service) GetUserRewards(userID int64) ([]*domain.UserRewardInfo, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	rewards, err := s.repo.GetAllRewards()
	if err != nil {
		return nil, err
	}

	result := make([]*domain.UserRewardInfo, 0, len(rewards))
	for _, reward := range rewards {
		claimed := s.repo.HasUserReward(userID, reward.ID)
		var pickedUp bool
		if claimed {
			userRewards, _ := s.repo.GetUserRewards(userID)
			for _, ur := range userRewards {
				if ur.RewardID == reward.ID {
					pickedUp = ur.PickedUp
					break
				}
			}
		}
		result = append(result, &domain.UserRewardInfo{
			RewardID:    reward.ID,
			Name:        reward.Name,
			Description: reward.Description,
			Cost:        reward.Cost,
			ImageURL:    reward.ImageURL,
			Available:   user.SkillPoints >= reward.Cost,
			Claimed:     claimed,
			PickedUp:    pickedUp,
		})
	}
	return result, nil
}

// GetUserRewardsByAdmin — получение купленных наград юзера (для админа при выдаче мерча)
func (s *Service) GetUserRewardsByAdmin(targetUserID int64, adminID int64) ([]*domain.UserRewardInfo, error) {
	admin, err := s.repo.GetUserByID(adminID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	if admin.Role != domain.RoleAdmin {
		return nil, domain.ErrInvalidRole
	}

	if _, err := s.repo.GetUserByID(targetUserID); err != nil {
		return nil, domain.ErrUserNotFound
	}

	// Возвращаем ТОЛЬКО купленные награды (claimed=true)
	allRewards, err := s.GetUserRewards(targetUserID)
	if err != nil {
		return nil, err
	}

	claimedRewards := make([]*domain.UserRewardInfo, 0)
	for _, r := range allRewards {
		if r.Claimed {
			claimedRewards = append(claimedRewards, r)
		}
	}

	return claimedRewards, nil
}

func (s *Service) ClaimReward(userID, rewardID int64) error {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return domain.ErrUserNotFound
	}

	reward, err := s.repo.GetRewardByID(rewardID)
	if err != nil {
		return domain.ErrRewardNotFound
	}

	if s.repo.HasUserReward(userID, rewardID) {
		return domain.ErrRewardAlreadyClaimed
	}

	if user.SkillPoints < reward.Cost {
		return domain.ErrInsufficientPoints
	}

	if err := s.repo.DeductSkillPoints(userID, reward.Cost); err != nil {
		return err
	}

	reason := fmt.Sprintf("Получение награды: %s", reward.Name)
	if err := s.repo.AddSkillPointTransaction(userID, -reward.Cost, domain.TransactionTypeReward, reason, nil); err != nil {
		return fmt.Errorf("ошибка создания транзакции: %w", err)
	}

	return s.repo.AddUserReward(userID, rewardID)
}

func (s *Service) ConfirmRewardPickup(userID, rewardID int64) error {
	if _, err := s.repo.GetUserByID(userID); err != nil {
		return domain.ErrUserNotFound
	}
	if _, err := s.repo.GetRewardByID(rewardID); err != nil {
		return domain.ErrRewardNotFound
	}
	return s.repo.ConfirmRewardPickup(userID, rewardID)
}

// ==================== ШАБЛОНЫ ====================

func (s *Service) CreateEventTemplate(title, description, location, imageURL string, duration time.Duration, maxParticipants *int64, reserveParticipants, skillPoints int64, requesterID int64) (*domain.EventTemplate, error) {
	requester, err := s.repo.GetUserByID(requesterID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	if requester.Role != domain.RoleAdmin && requester.Role != domain.RoleOrganizer {
		return nil, domain.ErrInvalidRole
	}
	return s.repo.CreateEventTemplate(title, description, location, imageURL, duration, maxParticipants, reserveParticipants, skillPoints)
}

func (s *Service) GetAllEventTemplates() ([]*domain.EventTemplate, error) {
	return s.repo.GetAllEventTemplates()
}

// ==================== КАРТИНКИ ====================

func (s *Service) GetAvailableImages() []string {
	return s.repo.GetAvailableImages()
}

// ==================== ИСТОРИЯ ТРАНЗАКЦИЙ ====================

func (s *Service) GetUserSkillPointHistory(userID int64) ([]*domain.SkillPointTransaction, error) {
	return s.repo.GetUserSkillPointHistory(userID)
}
