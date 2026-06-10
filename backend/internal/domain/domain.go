package domain

import (
	"errors"
	"time"
)

var (
	ErrUserExists        = errors.New("пользователь уже существует")
	ErrUserNotFound      = errors.New("пользователь не найден")
	ErrInvalidPassword   = errors.New("неверный пароль")
	ErrOrderExists       = errors.New("заказ уже загружен другим пользователем")
	ErrOrderOwnedByUser  = errors.New("заказ уже загружен этим пользователем")
	ErrInsufficientFunds = errors.New("недостаточно баллов")
	ErrInvalidOrder      = errors.New("неверный номер заказа")
	ErrInvalidRole       = errors.New("не соответсвующая роль")
)

type User struct {
	ID          int64
	Login       string
	Password    string
	SkillPoints int64
	FirstName   string
	LastName    string
	Telegram    string
	Role        string
}

type Event struct {
	ID                   int64
	Title                string
	Description          string
	Location             string
	CoverImageURL        string
	Status               string
	StartDate            time.Time
	EndDate              time.Time
	RegistrationDeadline *time.Time
	MaxParticipants      *int64
	ReserveParticipants  int64
	SkillPoints          int64
	CreatedByID          int64
	ParticipantsCount    int64
	ReserveCount         int64
}

const (
	RoleAdmin     = "Admin"
	RoleVolunteer = "Volunteer"
	RoleOrganizer = "Organizer"
)

const (
	EventRecruiting = "EVENT-RECRUITING"
	EventActive     = "EVENT-ACTIVE"
	EventFinished   = "EVENT-FINISHED"
	EventClosed     = "EVENT-CLOSED"
	EventCancelled  = "EVENT-CANCELLED"
)
