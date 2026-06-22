package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/andruwka0/volunteer_platform/internal/auth"
	"github.com/andruwka0/volunteer_platform/internal/domain"
)

type repository interface {
	CreateUser(login, passwordHash, firstname, lastname, telegram string) (*domain.User, error)
	GetUserByLogin(login string) (*domain.User, error)
	GetUserByID(ID int64) (*domain.User, error)
	UpdateUserSkillPoints(userID int64, pointsToAdd int64) error
	UpdateUserRole(userID int64, newRole string) error

	CreateEvent(title, description, location, image string, startDate, endDate time.Time, registrationDeadline *time.Time, maxParticipants *int64, reserveParticipants, skillPoints, createdByID int64) (*domain.Event, error)
	GetEventByID(eventID int64) (*domain.Event, error)
	GetAllEvents() ([]*domain.Event, error)
	UpdateEventStatus(eventID int64, newStatus string) error

	RegisterForEvent(eventID, userID int64) error
	GetEventParticipants(eventID int64) ([]int64, error)
	AwardPointsToAllParticipants(eventID int64) error
	CancelRegistration(eventID, userID int64) error
}

type Service struct {
	repo repository
}

func New(repo repository) *Service {
	return &Service{
		repo: repo,
	}
}

func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	passwordHash := hex.EncodeToString(hash[:])
	return passwordHash
}

func (s *Service) RegisterUser(login, password, firstname, lastname, telegram string) (string, error) {
	passwordHash := HashPassword(password)
	user, err := s.repo.CreateUser(login, passwordHash, firstname, lastname, telegram)
	if err != nil {
		return "", err
	}

	token, tokenErr := auth.GenerateToken(user.ID)
	if tokenErr != nil {
		return "", fmt.Errorf("ошибка генерации токена при аутентификации: %w", tokenErr)
	}
	return token, nil
}

func (s *Service) LoginUser(login, password string) (string, error) {
	user, err := s.repo.GetUserByLogin(login)
	if err != nil {
		return "", domain.ErrUserNotFound
	}
	passwordHash := HashPassword(password)
	if passwordHash != user.Password {
		return "", domain.ErrInvalidPassword
	}
	token, tokenErr := auth.GenerateToken(user.ID)
	if tokenErr != nil {
		return "", fmt.Errorf("Ошибка генерации токена при аутентификации: %w", tokenErr)
	}
	return token, nil
}

func (s *Service) PromoteUser(targetUserID int64, newRole string, requesterID int64) error {
	requester, err := s.repo.GetUserByID(requesterID)
	if err != nil {
		return err
	}
	if requester.Role != domain.RoleAdmin {
		return domain.ErrInvalidRole
	}

	// Повышать можно только до Organizer или Admin
	if newRole != domain.RoleOrganizer && newRole != domain.RoleAdmin {
		return domain.ErrInvalidPromotion
	}

	// Нельзя понижать самого себя (защита от дурака)
	if targetUserID == requesterID && newRole != domain.RoleAdmin {
		return domain.ErrCannotPromoteSelf
	}

	if _, err := s.repo.GetUserByID(targetUserID); err != nil {
		return domain.ErrUserNotFound
	}
	return s.repo.UpdateUserRole(targetUserID, newRole)
}

// ==================== ИВЕНТЫ ====================

func (s *Service) CreateEvent(title, description, location, image string,
	startDate, endDate time.Time, registrationDeadline *time.Time,
	maxParticipants *int64, reserveParticipants, skillPoints, createdByID int64) (*domain.Event, error) {
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
		reserveParticipants, skillPoints, createdByID)
}
func (s *Service) GetEventByID(eventID int64) (*domain.Event, error) {
	return s.repo.GetEventByID(eventID)
}

func (s *Service) GetAllEvents() ([]*domain.Event, error) {
	return s.repo.GetAllEvents()
}

func (s *Service) ApproveAndAwardPoints(eventID int64, adminID int64) error {
	user, err := s.repo.GetUserByID(adminID)
	if err != nil {
		return domain.ErrUserNotFound
	}
	if user.Role != domain.RoleAdmin {
		return domain.ErrInvalidRole
	}
	event, err := s.repo.GetEventByID(eventID)
	if err != nil {
		return domain.ErrEventNotFound
	}
	if event.Status != domain.EventFinished {
		return domain.ErrEventNotFinished
	}
	err = s.repo.AwardPointsToAllParticipants(eventID)
	if err != nil {
		return fmt.Errorf("ошибка начисления баллов: %w", err)
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
	return s.repo.RegisterForEvent(eventID, userID)
}

func (s *Service) CancelRegistration(eventID, userID int64) error {
	return s.repo.CancelRegistration(eventID, userID)
}

func (s *Service) GetEventParticipants(eventID int64) ([]int64, error) {
	return s.repo.GetEventParticipants(eventID)
}

func (s *Service) GetUserByID(id int64) (*domain.User, error) {
	return s.repo.GetUserByID(id)
}
