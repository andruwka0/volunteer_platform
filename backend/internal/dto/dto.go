package dto

import "time"

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type UserDTO struct {
	ID          int64  `json:"id"`
	Login       string `json:"login"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	MiddleName  string `json:"middle_name"`
	Telegram    string `json:"telegram"`
	Role        string `json:"role"`
	SkillPoints int64  `json:"skill_points"`
}

type UserSearchResultDTO struct {
	Users []UserDTO `json:"users"`
	Count int       `json:"count"`
}

type EventDTO struct {
	ID                   int64      `json:"id"`
	Title                string     `json:"title"`
	Description          string     `json:"description"`
	Location             string     `json:"location"`
	CoverImageURL        string     `json:"cover_image_url"`
	Status               string     `json:"status"`
	StartDate            time.Time  `json:"start_date"`
	EndDate              time.Time  `json:"end_date"`
	RegistrationDeadline *time.Time `json:"registration_deadline,omitempty"`
	MaxParticipants      *int64     `json:"max_participants,omitempty"`
	ReserveParticipants  int64      `json:"reserve_participants"`
	SkillPoints          int64      `json:"skill_points"`
	CreatedByID          int64      `json:"created_by_id"`
	ParticipantsCount    int64      `json:"participants_count"`
	ReserveCount         int64      `json:"reserve_count"`
}

type EventListDTO struct {
	Events []EventDTO `json:"events"`
	Count  int        `json:"count"`
}

type ParticipantDTO struct {
	UserID              int64  `json:"user_id"`
	FirstName           string `json:"first_name"`
	LastName            string `json:"last_name"`
	MiddleName          string `json:"middle_name"`
	Telegram            string `json:"telegram"`
	AttendanceConfirmed bool   `json:"attendance_confirmed"`
}

type ParticipantListDTO struct {
	EventID      int64            `json:"event_id"`
	Participants []ParticipantDTO `json:"participants"`
	Count        int              `json:"count"`
}

type UserEventHistoryDTO struct {
	EventID             int64     `json:"event_id"`
	Title               string    `json:"title"`
	Status              string    `json:"status"`
	JoinedAt            time.Time `json:"joined_at"`
	AttendanceConfirmed bool      `json:"attendance_confirmed"`
	SkillPoints         int64     `json:"skill_points"`
}

type UserEventHistoryListDTO struct {
	UserID int64                 `json:"user_id"`
	Events []UserEventHistoryDTO `json:"events"`
	Count  int                   `json:"count"`
}

type RewardDTO struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Cost        int64  `json:"cost"`
	ImageURL    string `json:"image_url"`
}

type UserRewardDTO struct {
	RewardID    int64  `json:"reward_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Cost        int64  `json:"cost"`
	ImageURL    string `json:"image_url"`
	Available   bool   `json:"available"`
	Claimed     bool   `json:"claimed"`
	PickedUp    bool   `json:"picked_up"`
}

type UserRewardListDTO struct {
	UserID  int64           `json:"user_id"`
	Rewards []UserRewardDTO `json:"rewards"`
}

type EventTemplateDTO struct {
	ID                  int64  `json:"id"`
	Title               string `json:"title"`
	Description         string `json:"description"`
	Location            string `json:"location"`
	CoverImageURL       string `json:"cover_image_url"`
	DurationMinutes     int    `json:"duration_minutes"`
	MaxParticipants     *int64 `json:"max_participants,omitempty"`
	ReserveParticipants int64  `json:"reserve_participants"`
	SkillPoints         int64  `json:"skill_points"`
}

type EventTemplateListDTO struct {
	Templates []EventTemplateDTO `json:"templates"`
	Count     int                `json:"count"`
}

type ImageListDTO struct {
	Images []string `json:"images"`
	Count  int      `json:"count"`
}

type MessageDTO struct {
	Message string `json:"message"`
}

type TokenDTO struct {
	Token string `json:"token"`
}

// ==================== ПОИСК С НАГРАДАМИ ====================

// UserWithRewardsDTO — юзер + его купленные награды (для выдачи мерча)
type UserWithRewardsDTO struct {
	ID          int64           `json:"id"`
	Login       string          `json:"login"`
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	MiddleName  string          `json:"middle_name"`
	Telegram    string          `json:"telegram"`
	Role        string          `json:"role"`
	SkillPoints int64           `json:"skill_points"`
	Rewards     []UserRewardDTO `json:"rewards"` // ← купленные награды
}

// UserSearchWithRewardsResultDTO — результат поиска с наградами
type UserSearchWithRewardsResultDTO struct {
	Users []UserWithRewardsDTO `json:"users"`
	Count int                  `json:"count"`
}
