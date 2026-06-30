package domain

import (
	"errors"
	"time"
)

var (
	ErrUserExists           = errors.New("пользователь уже существует")
	ErrUserNotFound         = errors.New("пользователь не найден")
	ErrInvalidPassword      = errors.New("неверный пароль")
	ErrInvalidRole          = errors.New("не соответствующая роль")
	ErrEventNotFound        = errors.New("мероприятие не найдено")
	ErrInvalidDates         = errors.New("дата окончания должна быть позже даты начала")
	ErrInvalidPromotion     = errors.New("можно повысить только до Organizer или Admin")
	ErrCannotPromoteSelf    = errors.New("нельзя понизить собственную роль")
	ErrEventNotFinished     = errors.New("можно начислить баллы только для завершенных (FINISHED) ивентов")
	ErrRegistrationClosed   = errors.New("регистрация на это мероприятие закрыта")
	ErrOrganizerSelfReg     = errors.New("организатор не может регистрироваться на собственный ивент")
	ErrNotEventCreator      = errors.New("только создатель мероприятия может подтверждать посещаемость")
	ErrAlreadyConfirmed     = errors.New("посещаемость уже подтверждена")
	ErrRewardNotFound       = errors.New("награда не найдена")
	ErrInsufficientPoints   = errors.New("недостаточно баллов")
	ErrRewardAlreadyClaimed = errors.New("награда уже получена")
	ErrTemplateNotFound     = errors.New("шаблон не найден")
)

type User struct {
	ID          int64
	Login       string
	Password    string
	SkillPoints int64
	FirstName   string
	LastName    string
	MiddleName  string
	Telegram    string
	Role        string
	CreatedAt   time.Time
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
	TemplateID           *int64
	CreatedAt            time.Time
}

type UserEvent struct {
	ID                  int64
	UserID              int64
	EventID             int64
	JoinedAt            time.Time
	AttendanceConfirmed bool
}

type Reward struct {
	ID          int64
	Name        string
	Description string
	Cost        int64
	ImageURL    string
	CreatedAt   time.Time
}

type UserReward struct {
	ID        int64
	UserID    int64
	RewardID  int64
	ClaimedAt time.Time
	PickedUp  bool
}

type SkillPointTransaction struct {
	ID        int64
	UserID    int64     // для миграции на БД
	Points    int64     // +начисление / -списание
	Type      string    // "manual", "event", "reward"
	Reason    string    // описание
	EventID   *int64    // ссылка на ивент (если type = "event")
	CreatedAt time.Time // когда
}

type EventTemplate struct {
	ID                  int64
	Title               string
	Description         string
	Location            string
	CoverImageURL       string
	Duration            time.Duration
	MaxParticipants     *int64
	ReserveParticipants int64
	SkillPoints         int64
	CreatedAt           time.Time
}

// DTO-подобные структуры для ответов (используются в Service)
type ParticipantInfo struct {
	UserID              int64
	FirstName           string
	LastName            string
	MiddleName          string
	Telegram            string
	AttendanceConfirmed bool
}

type UserEventWithDetails struct {
	EventID             int64
	Title               string
	Status              string
	JoinedAt            time.Time
	AttendanceConfirmed bool
	SkillPoints         int64
}

type UserRewardInfo struct {
	RewardID    int64
	Name        string
	Description string
	Cost        int64
	ImageURL    string
	Available   bool
	Claimed     bool
	PickedUp    bool
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

const (
	TransactionTypeManual = "manual"
	TransactionTypeEvent  = "event"
	TransactionTypeReward = "reward"
)
