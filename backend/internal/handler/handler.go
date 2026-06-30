package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/andruwka0/volunteer_platform/internal/domain"
	"github.com/andruwka0/volunteer_platform/internal/dto"
)

type Service interface {
	// Аутентификация
	RegisterUser(login, password, firstname, lastname, middlename, telegram string) (string, error)
	LoginUser(login, password string) (string, error)

	// Пользователи
	GetUserByID(id int64) (*domain.User, error)
	SearchUsers(firstName, lastName, middleName string) ([]*domain.User, error)
	PromoteUser(targetUserID int64, newRole string, requesterID int64) error
	AwardUser(targetUserID int64, points int64, reason string, requesterID int64) error

	// Ивенты
	CreateEvent(title, description, location, image string,
		startDate, endDate time.Time, registrationDeadline *time.Time,
		maxParticipants *int64, reserveParticipants, skillPoints, createdByID int64,
		templateID *int64) (*domain.Event, error)
	GetEventByID(eventID int64) (*domain.Event, error)
	GetAllEvents(userID int64) ([]*domain.Event, error)
	ApproveAndAwardPoints(eventID int64, adminID int64) error

	// Регистрация на ивент
	RegisterForEvent(eventID, userID int64) error
	CancelRegistration(eventID, userID int64) error
	GetEventParticipants(eventID int64) ([]*domain.ParticipantInfo, error)
	ConfirmAttendance(eventID, userID, organizerID int64) error

	// История
	GetUserEventHistory(userID int64) ([]*domain.UserEventWithDetails, error)
	GetUserSkillPointHistory(userID int64) ([]*domain.SkillPointTransaction, error)

	// Бан-лист
	BanUser(organizerID, userID int64) error
	UnbanUser(organizerID, userID int64) error

	// Награды
	CreateReward(name, description, imageURL string, cost int64, requesterID int64) (*domain.Reward, error)
	GetAllRewards() ([]*domain.Reward, error)
	GetUserRewards(userID int64) ([]*domain.UserRewardInfo, error)
	GetUserRewardsByAdmin(targetUserID int64, adminID int64) ([]*domain.UserRewardInfo, error)
	ClaimReward(userID, rewardID int64) error
	ConfirmRewardPickup(userID, rewardID int64) error

	// Шаблоны
	CreateEventTemplate(title, description, location, imageURL string, duration time.Duration, maxParticipants *int64, reserveParticipants, skillPoints int64, requesterID int64) (*domain.EventTemplate, error)
	GetAllEventTemplates() ([]*domain.EventTemplate, error)

	// Картинки
	GetAvailableImages() []string
}

type Handler struct {
	svc Service
}

func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

type contextKey string

const CtxKeyUserID contextKey = "userID"

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(CtxKeyUserID).(int64)
	return userID, ok
}

func writeSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dto.Response{Success: true, Data: data})
}

func writeError(w http.ResponseWriter, status int, code, userMsg string, internalErr error) {
	if internalErr != nil {
		log.Printf("ошибка code=%s status=%d: %v", code, status, internalErr)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dto.Response{
		Success: false,
		Error:   &dto.ErrorInfo{Code: code, Message: userMsg},
	})
}

// ==================== АУТЕНТИФИКАЦИЯ ====================

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login      string `json:"login"`
		Password   string `json:"password"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		MiddleName string `json:"middle_name"`
		Telegram   string `json:"telegram"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат запроса", err)
		return
	}

	if req.Login == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Логин и пароль обязательны", nil)
		return
	}

	token, err := h.svc.RegisterUser(req.Login, req.Password, req.FirstName, req.LastName, req.MiddleName, req.Telegram)
	if err != nil {
		if errors.Is(err, domain.ErrUserExists) {
			writeError(w, http.StatusConflict, "USER_EXISTS", "Пользователь с таким логином уже существует", err)
			return
		}
		writeError(w, http.StatusInternalServerError, "REGISTRATION_FAILED", "Не удалось зарегистрировать пользователя", err)
		return
	}

	writeSuccess(w, http.StatusCreated, dto.TokenDTO{Token: token})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат запроса", err)
		return
	}

	if req.Login == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Логин и пароль обязательны", nil)
		return
	}

	token, err := h.svc.LoginUser(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, domain.ErrInvalidPassword) {
			writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Неверный логин или пароль", err)
			return
		}
		writeError(w, http.StatusInternalServerError, "LOGIN_FAILED", "Ошибка входа в систему", err)
		return
	}

	writeSuccess(w, http.StatusOK, dto.TokenDTO{Token: token})
}

// ==================== ПРОФИЛЬ ====================

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	user, err := h.svc.GetUserByID(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GET_USER_FAILED", "Не удалось получить данные пользователя", err)
		return
	}

	writeSuccess(w, http.StatusOK, dto.UserDTO{
		ID:          user.ID,
		Login:       user.Login,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		MiddleName:  user.MiddleName,
		Telegram:    user.Telegram,
		Role:        user.Role,
		SkillPoints: user.SkillPoints,
	})
}

// ==================== ИВЕНТЫ ====================

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	var req struct {
		Title                string     `json:"title"`
		Description          string     `json:"description"`
		Location             string     `json:"location"`
		Image                string     `json:"image"`
		StartDate            time.Time  `json:"start_date"`
		EndDate              time.Time  `json:"end_date"`
		RegistrationDeadline *time.Time `json:"registration_deadline"`
		MaxParticipants      *int64     `json:"max_participants"`
		ReserveParticipants  int64      `json:"reserve_participants"`
		SkillPoints          int64      `json:"skill_points"`
		TemplateID           *int64     `json:"template_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат запроса", err)
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "MISSING_TITLE", "Название обязательно", nil)
		return
	}
	if req.StartDate.IsZero() || req.EndDate.IsZero() {
		writeError(w, http.StatusBadRequest, "MISSING_DATES", "Даты обязательны", nil)
		return
	}

	event, err := h.svc.CreateEvent(
		req.Title, req.Description, req.Location, req.Image,
		req.StartDate, req.EndDate, req.RegistrationDeadline,
		req.MaxParticipants, req.ReserveParticipants, req.SkillPoints, userID, req.TemplateID,
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidRole):
			writeError(w, http.StatusForbidden, "FORBIDDEN", "Недостаточно прав", err)
		case errors.Is(err, domain.ErrInvalidDates):
			writeError(w, http.StatusBadRequest, "INVALID_DATES", "Дата окончания должна быть позже даты начала", err)
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера", err)
		}
		return
	}

	writeSuccess(w, http.StatusCreated, eventToDTO(event))
}

func (h *Handler) GetAllEvents(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	events, err := h.svc.GetAllEvents(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GET_EVENTS_FAILED", "Не удалось получить список мероприятий", err)
		return
	}

	eventDTOs := make([]dto.EventDTO, 0, len(events))
	for _, e := range events {
		eventDTOs = append(eventDTOs, eventToDTO(e))
	}

	writeSuccess(w, http.StatusOK, dto.EventListDTO{Events: eventDTOs, Count: len(eventDTOs)})
}

func (h *Handler) GetEventByID(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_EVENT_ID", "Неверный ID мероприятия", err)
		return
	}

	event, err := h.svc.GetEventByID(eventID)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			writeError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Мероприятие не найдено", err)
			return
		}
		writeError(w, http.StatusInternalServerError, "GET_EVENT_FAILED", "Не удалось получить мероприятие", err)
		return
	}

	writeSuccess(w, http.StatusOK, eventToDTO(event))
}

// ==================== РЕГИСТРАЦИЯ НА ИВЕНТ ====================

func (h *Handler) RegisterForEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	eventID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_EVENT_ID", "Неверный ID мероприятия", err)
		return
	}

	if errEvent := h.svc.RegisterForEvent(eventID, userID); errEvent != nil {
		switch {
		case errors.Is(errEvent, domain.ErrEventNotFound):
			writeError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Мероприятие не найдено", err)
		case errors.Is(errEvent, domain.ErrRegistrationClosed):
			writeError(w, http.StatusBadRequest, "REGISTRATION_CLOSED", "Регистрация закрыта", err)
		case errors.Is(errEvent, domain.ErrOrganizerSelfReg):
			writeError(w, http.StatusBadRequest, "ORGANIZER_SELF_REG", "Организатор не может регаться на свои ивенты", err)
		default:
			writeError(w, http.StatusBadRequest, "REGISTRATION_FAILED", errEvent.Error(), errEvent)
		}
		return
	}

	writeSuccess(w, http.StatusOK, dto.MessageDTO{Message: "Вы успешно зарегистрированы на мероприятие"})
}

func (h *Handler) CancelRegistration(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	eventID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_EVENT_ID", "Неверный ID мероприятия", err)
		return
	}

	if errCancel := h.svc.CancelRegistration(eventID, userID); errCancel != nil {
		writeError(w, http.StatusBadRequest, "CANCEL_FAILED", errCancel.Error(), errCancel)
		return
	}

	writeSuccess(w, http.StatusOK, dto.MessageDTO{Message: "Регистрация отменена"})
}

func (h *Handler) GetEventParticipants(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_EVENT_ID", "Неверный ID мероприятия", err)
		return
	}

	participants, err := h.svc.GetEventParticipants(eventID)
	if err != nil {
		if errors.Is(err, domain.ErrEventNotFound) {
			writeError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Мероприятие не найдено", err)
			return
		}
		writeError(w, http.StatusInternalServerError, "GET_PARTICIPANTS_FAILED", "Не удалось получить список участников", err)
		return
	}

	participantDTOs := make([]dto.ParticipantDTO, 0, len(participants))
	for _, p := range participants {
		participantDTOs = append(participantDTOs, dto.ParticipantDTO{
			UserID:              p.UserID,
			FirstName:           p.FirstName,
			LastName:            p.LastName,
			MiddleName:          p.MiddleName,
			Telegram:            p.Telegram,
			AttendanceConfirmed: p.AttendanceConfirmed,
		})
	}

	writeSuccess(w, http.StatusOK, dto.ParticipantListDTO{
		EventID:      eventID,
		Participants: participantDTOs,
		Count:        len(participantDTOs),
	})
}

// ==================== ПОДТВЕРЖДЕНИЕ ПОСЕЩАЕМОСТИ ====================

func (h *Handler) ConfirmAttendance(w http.ResponseWriter, r *http.Request) {
	organizerID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	eventID, err := strconv.ParseInt(r.PathValue("event_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_EVENT_ID", "Неверный ID мероприятия", err)
		return
	}

	userID, err := strconv.ParseInt(r.PathValue("user_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Неверный ID пользователя", err)
		return
	}

	if err := h.svc.ConfirmAttendance(eventID, userID, organizerID); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotEventCreator):
			writeError(w, http.StatusForbidden, "NOT_CREATOR", "Только создатель может подтверждать", err)
		case errors.Is(err, domain.ErrAlreadyConfirmed):
			writeError(w, http.StatusBadRequest, "ALREADY_CONFIRMED", "Посещаемость уже подтверждена", err)
		default:
			writeError(w, http.StatusBadRequest, "CONFIRM_FAILED", err.Error(), err)
		}
		return
	}

	writeSuccess(w, http.StatusOK, dto.MessageDTO{Message: "Посещаемость подтверждена"})
}

// ==================== АДМИНСКИЕ ДЕЙСТВИЯ ====================

func (h *Handler) ApproveEvent(w http.ResponseWriter, r *http.Request) {
	adminID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	eventID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_EVENT_ID", "Неверный ID мероприятия", err)
		return
	}

	if errPoints := h.svc.ApproveAndAwardPoints(eventID, adminID); errPoints != nil {
		switch {
		case errors.Is(errPoints, domain.ErrEventNotFinished):
			writeError(w, http.StatusBadRequest, "EVENT_NOT_FINISHED", "Начисление только для Finished events", err)
		default:
			writeError(w, http.StatusBadRequest, "APPROVE_FAILED", errPoints.Error(), errPoints)
		}
		return
	}

	writeSuccess(w, http.StatusOK, dto.MessageDTO{Message: "Баллы начислены, мероприятие закрыто"})
}

func (h *Handler) PromoteUser(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	targetID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Неверный ID пользователя", err)
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат запроса", err)
		return
	}

	if errRole := h.svc.PromoteUser(targetID, req.Role, requesterID); errRole != nil {
		switch {
		case errors.Is(errRole, domain.ErrInvalidRole):
			writeError(w, http.StatusForbidden, "FORBIDDEN", "Недостаточно прав", err)
		case errors.Is(errRole, domain.ErrInvalidPromotion):
			writeError(w, http.StatusBadRequest, "INVALID_PROMOTION", "Неверная роль", err)
		default:
			writeError(w, http.StatusBadRequest, "PROMOTE_FAILED", errRole.Error(), errRole)
		}
		return
	}

	writeSuccess(w, http.StatusOK, dto.MessageDTO{Message: "Роль успешно изменена"})
}

func (h *Handler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	firstName := r.URL.Query().Get("first_name")
	lastName := r.URL.Query().Get("last_name")
	middleName := r.URL.Query().Get("middle_name")

	if firstName == "" || lastName == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Параметры first_name и last_name обязательны", nil)
		return
	}

	users, err := h.svc.SearchUsers(firstName, lastName, middleName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "SEARCH_FAILED", "Поиск не удался", err)
		return
	}

	userDTOs := make([]dto.UserDTO, 0, len(users))
	for _, u := range users {
		userDTOs = append(userDTOs, dto.UserDTO{
			ID:          u.ID,
			Login:       u.Login,
			FirstName:   u.FirstName,
			LastName:    u.LastName,
			MiddleName:  u.MiddleName,
			Telegram:    u.Telegram,
			Role:        u.Role,
			SkillPoints: u.SkillPoints,
		})
	}

	writeSuccess(w, http.StatusOK, dto.UserSearchResultDTO{Users: userDTOs, Count: len(userDTOs)})
}

func (h *Handler) AwardUser(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	targetID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Неверный ID пользователя", err)
		return
	}

	var req struct {
		Points int64  `json:"points"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат запроса", err)
		return
	}

	if err := h.svc.AwardUser(targetID, req.Points, req.Reason, requesterID); err != nil {
		writeError(w, http.StatusBadRequest, "AWARD_FAILED", err.Error(), err)
		return
	}

	writeSuccess(w, http.StatusOK, dto.MessageDTO{Message: "Баллы успешно начислены"})
}

// GetUserRewardsByAdmin — получение купленных наград юзера (для выдачи мерча)
func (h *Handler) GetUserRewardsByAdmin(w http.ResponseWriter, r *http.Request) {
	adminID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	targetID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Неверный ID пользователя", err)
		return
	}

	rewards, err := h.svc.GetUserRewardsByAdmin(targetID, adminID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidRole):
			writeError(w, http.StatusForbidden, "FORBIDDEN", "Недостаточно прав", err)
		case errors.Is(err, domain.ErrUserNotFound):
			writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "Пользователь не найден", err)
		default:
			writeError(w, http.StatusInternalServerError, "GET_REWARDS_FAILED", err.Error(), err)
		}
		return
	}

	rewardDTOs := make([]dto.UserRewardDTO, 0, len(rewards))
	for _, r := range rewards {
		rewardDTOs = append(rewardDTOs, dto.UserRewardDTO{
			RewardID:    r.RewardID,
			Name:        r.Name,
			Description: r.Description,
			Cost:        r.Cost,
			ImageURL:    r.ImageURL,
			Available:   r.Available,
			Claimed:     r.Claimed,
			PickedUp:    r.PickedUp,
		})
	}

	writeSuccess(w, http.StatusOK, dto.UserRewardListDTO{UserID: targetID, Rewards: rewardDTOs})
}

// ==================== ИСТОРИЯ ====================

func (h *Handler) GetUserEvents(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	events, err := h.svc.GetUserEventHistory(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GET_EVENTS_FAILED", "Не удалось получить историю событий", err)
		return
	}

	eventDTOs := make([]dto.UserEventHistoryDTO, 0, len(events))
	for _, e := range events {
		eventDTOs = append(eventDTOs, dto.UserEventHistoryDTO{
			EventID:             e.EventID,
			Title:               e.Title,
			Status:              e.Status,
			JoinedAt:            e.JoinedAt,
			AttendanceConfirmed: e.AttendanceConfirmed,
			SkillPoints:         e.SkillPoints,
		})
	}

	writeSuccess(w, http.StatusOK, dto.UserEventHistoryListDTO{
		UserID: userID,
		Events: eventDTOs,
		Count:  len(eventDTOs),
	})
}

func (h *Handler) GetUserSkillPointHistory(w http.ResponseWriter, r *http.Request) {
	requesterID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	idStr := r.PathValue("id")
	targetID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Неверный ID пользователя", err)
		return
	}

	requester, err := h.svc.GetUserByID(requesterID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "USER_CHECK_FAILED", "Не удалось проверить пользователя", err)
		return
	}
	if requesterID != targetID && requester.Role != domain.RoleAdmin {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "Нет доступа к чужой истории", nil)
		return
	}

	transactions, err := h.svc.GetUserSkillPointHistory(targetID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "Пользователь не найден", err)
			return
		}
		writeError(w, http.StatusInternalServerError, "GET_HISTORY_FAILED", "Не удалось получить историю", err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{
		"user_id":      targetID,
		"transactions": transactions,
		"count":        len(transactions),
	})
}

// ==================== БАН-ЛИСТ ====================

func (h *Handler) BanUser(w http.ResponseWriter, r *http.Request) {
	organizerID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	var req struct {
		UserID int64 `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат запроса", err)
		return
	}

	if err := h.svc.BanUser(organizerID, req.UserID); err != nil {
		writeError(w, http.StatusBadRequest, "BAN_FAILED", err.Error(), err)
		return
	}

	writeSuccess(w, http.StatusOK, dto.MessageDTO{Message: "Пользователь добавлен в черный список"})
}

func (h *Handler) UnbanUser(w http.ResponseWriter, r *http.Request) {
	organizerID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	userID, err := strconv.ParseInt(r.PathValue("user_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Неверный ID пользователя", err)
		return
	}

	if err := h.svc.UnbanUser(organizerID, userID); err != nil {
		writeError(w, http.StatusBadRequest, "UNBAN_FAILED", err.Error(), err)
		return
	}

	writeSuccess(w, http.StatusOK, dto.MessageDTO{Message: "Пользователь удален из черного списка"})
}

// ==================== НАГРАДЫ ====================

func (h *Handler) CreateReward(w http.ResponseWriter, r *http.Request) {
	adminID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		ImageURL    string `json:"image_url"`
		Cost        int64  `json:"cost"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат запроса", err)
		return
	}

	reward, err := h.svc.CreateReward(req.Name, req.Description, req.ImageURL, req.Cost, adminID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_REWARD_FAILED", err.Error(), err)
		return
	}

	writeSuccess(w, http.StatusCreated, rewardToDTO(reward))
}

func (h *Handler) GetAllRewards(w http.ResponseWriter, r *http.Request) {
	rewards, err := h.svc.GetAllRewards()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GET_REWARDS_FAILED", "Не удалось получить список наград", err)
		return
	}

	rewardDTOs := make([]dto.RewardDTO, 0, len(rewards))
	for _, r := range rewards {
		rewardDTOs = append(rewardDTOs, rewardToDTO(r))
	}

	writeSuccess(w, http.StatusOK, rewardDTOs)
}

func (h *Handler) GetUserRewards(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	rewards, err := h.svc.GetUserRewards(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GET_USER_REWARDS_FAILED", "Не удалось получить награды", err)
		return
	}

	rewardDTOs := make([]dto.UserRewardDTO, 0, len(rewards))
	for _, r := range rewards {
		rewardDTOs = append(rewardDTOs, dto.UserRewardDTO{
			RewardID:    r.RewardID,
			Name:        r.Name,
			Description: r.Description,
			Cost:        r.Cost,
			ImageURL:    r.ImageURL,
			Available:   r.Available,
			Claimed:     r.Claimed,
			PickedUp:    r.PickedUp,
		})
	}

	writeSuccess(w, http.StatusOK, dto.UserRewardListDTO{UserID: userID, Rewards: rewardDTOs})
}

func (h *Handler) ClaimReward(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	rewardID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REWARD_ID", "Неверный ID награды", err)
		return
	}

	if err := h.svc.ClaimReward(userID, rewardID); err != nil {
		switch {
		case errors.Is(err, domain.ErrInsufficientPoints):
			writeError(w, http.StatusBadRequest, "INSUFFICIENT_POINTS", "Недостаточно баллов", err)
		case errors.Is(err, domain.ErrRewardAlreadyClaimed):
			writeError(w, http.StatusBadRequest, "ALREADY_CLAIMED", "Награда уже получена", err)
		default:
			writeError(w, http.StatusBadRequest, "CLAIM_FAILED", err.Error(), err)
		}
		return
	}

	writeSuccess(w, http.StatusOK, dto.MessageDTO{Message: "Награда забронирована"})
}

func (h *Handler) ConfirmRewardPickup(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("user_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Неверный ID пользователя", err)
		return
	}

	rewardID, err := strconv.ParseInt(r.PathValue("reward_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REWARD_ID", "Неверный ID награды", err)
		return
	}

	if err := h.svc.ConfirmRewardPickup(userID, rewardID); err != nil {
		writeError(w, http.StatusBadRequest, "CONFIRM_PICKUP_FAILED", err.Error(), err)
		return
	}

	writeSuccess(w, http.StatusOK, dto.MessageDTO{Message: "Выдача награды подтверждена"})
}

// ==================== ШАБЛОНЫ ====================

func (h *Handler) CreateEventTemplate(w http.ResponseWriter, r *http.Request) {
	adminID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	var req struct {
		Title               string `json:"title"`
		Description         string `json:"description"`
		Location            string `json:"location"`
		ImageURL            string `json:"image_url"`
		DurationMinutes     int    `json:"duration_minutes"`
		MaxParticipants     *int64 `json:"max_participants"`
		ReserveParticipants int64  `json:"reserve_participants"`
		SkillPoints         int64  `json:"skill_points"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат запроса", err)
		return
	}

	duration := time.Duration(req.DurationMinutes) * time.Minute

	template, err := h.svc.CreateEventTemplate(req.Title, req.Description, req.Location, req.ImageURL, duration, req.MaxParticipants, req.ReserveParticipants, req.SkillPoints, adminID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "CREATE_TEMPLATE_FAILED", err.Error(), err)
		return
	}

	writeSuccess(w, http.StatusCreated, templateToDTO(template))
}

func (h *Handler) GetAllEventTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.svc.GetAllEventTemplates()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GET_TEMPLATES_FAILED", "Не удалось получить шаблоны", err)
		return
	}

	templateDTOs := make([]dto.EventTemplateDTO, 0, len(templates))
	for _, t := range templates {
		templateDTOs = append(templateDTOs, templateToDTO(t))
	}

	writeSuccess(w, http.StatusOK, dto.EventTemplateListDTO{Templates: templateDTOs, Count: len(templateDTOs)})
}

// ==================== КАРТИНКИ ====================

func (h *Handler) GetAvailableImages(w http.ResponseWriter, r *http.Request) {
	images := h.svc.GetAvailableImages()
	writeSuccess(w, http.StatusOK, dto.ImageListDTO{Images: images, Count: len(images)})
}

// ==================== ХЕЛПЕРЫ ====================

func eventToDTO(e *domain.Event) dto.EventDTO {
	return dto.EventDTO{
		ID:                   e.ID,
		Title:                e.Title,
		Description:          e.Description,
		Location:             e.Location,
		CoverImageURL:        e.CoverImageURL,
		Status:               e.Status,
		StartDate:            e.StartDate,
		EndDate:              e.EndDate,
		RegistrationDeadline: e.RegistrationDeadline,
		MaxParticipants:      e.MaxParticipants,
		ReserveParticipants:  e.ReserveParticipants,
		SkillPoints:          e.SkillPoints,
		CreatedByID:          e.CreatedByID,
		ParticipantsCount:    e.ParticipantsCount,
		ReserveCount:         e.ReserveCount,
	}
}

func rewardToDTO(r *domain.Reward) dto.RewardDTO {
	return dto.RewardDTO{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Cost:        r.Cost,
		ImageURL:    r.ImageURL,
	}
}

func templateToDTO(t *domain.EventTemplate) dto.EventTemplateDTO {
	return dto.EventTemplateDTO{
		ID:                  t.ID,
		Title:               t.Title,
		Description:         t.Description,
		Location:            t.Location,
		CoverImageURL:       t.CoverImageURL,
		DurationMinutes:     int(t.Duration.Minutes()),
		MaxParticipants:     t.MaxParticipants,
		ReserveParticipants: t.ReserveParticipants,
		SkillPoints:         t.SkillPoints,
	}
}
