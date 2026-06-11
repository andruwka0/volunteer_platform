// internal/handler/handler.go

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
)

type Service interface {
	RegisterUser(login, password, firstname, lastname, telegram string) (string, error)
	LoginUser(login, password string) (string, error)
	PromoteUser(targetUserID int64, newRole string, requesterID int64) error
	CreateEvent(title, description, location, image string,
		startDate, endDate time.Time, registrationDeadline *time.Time,
		maxParticipants *int64, reserveParticipants, skillPoints, createdByID int64) (*domain.Event, error)
	GetEventByID(eventID int64) (*domain.Event, error)
	GetAllEvents() ([]*domain.Event, error)
	ApproveAndAwardPoints(eventID int64, adminID int64) error
	RegisterForEvent(eventID, userID int64) error
	CancelRegistration(eventID, userID int64) error
	GetEventParticipants(eventID int64) ([]int64, error)
	GetUserByID(id int64) (*domain.User, error)
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

type ErrorResponse struct {
	Code    int    `json:"code"`
	UserMsg string `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

func writeError(w http.ResponseWriter, status int, code, userMsg string, internalErr error) {
	if internalErr != nil {
		log.Printf("ошибка code=%s status=%d: %v", code, status, internalErr)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	userError := ErrorResponse{Code: status, UserMsg: userMsg}
	if err := json.NewEncoder(w).Encode(userError); err != nil {
		log.Printf("writeError: ошибка сериализации: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

// POST /register

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login     string `json:"login"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Telegram  string `json:"telegram"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат запроса", err)
		return
	}

	if req.Login == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Логин и пароль обязательны", nil)
		return
	}

	token, err := h.svc.RegisterUser(req.Login, req.Password, req.FirstName, req.LastName, req.Telegram)
	if err != nil {
		if errors.Is(err, domain.ErrUserExists) {
			writeError(w, http.StatusConflict, "USER_EXISTS", "Пользователь с таким логином уже существует", err)
			return
		}
		writeError(w, http.StatusInternalServerError, "REGISTRATION_FAILED", "Не удалось зарегистрировать пользователя", err)
		return
	}

	writeJSON(w, http.StatusCreated, TokenResponse{Token: token})
}

// POST /login

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

	writeJSON(w, http.StatusOK, TokenResponse{Token: token})
}

// POST /events

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
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат запроса", err)
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "MISSING_TITLE", "Название мероприятия обязательно", nil)
		return
	}
	if req.StartDate.IsZero() || req.EndDate.IsZero() {
		writeError(w, http.StatusBadRequest, "MISSING_DATES", "Даты начала и окончания обязательны", nil)
		return
	}
	if !req.EndDate.After(req.StartDate) {
		writeError(w, http.StatusBadRequest, "INVALID_DATES", "Дата окончания должна быть позже даты начала", nil)
		return
	}

	event, err := h.svc.CreateEvent(
		req.Title, req.Description, req.Location, req.Image,
		req.StartDate, req.EndDate, req.RegistrationDeadline,
		req.MaxParticipants, req.ReserveParticipants, req.SkillPoints, userID,
	)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidRole) {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "Недостаточно прав для создания мероприятия", err)
			return
		}
		if errors.Is(err, domain.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "Создатель не найден", err)
			return
		}
		writeError(w, http.StatusInternalServerError, "CREATE_EVENT_FAILED", "Не удалось создать мероприятие", err)
		return
	}

	writeJSON(w, http.StatusCreated, event)
}

// GET /events

func (h *Handler) GetAllEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.svc.GetAllEvents()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GET_EVENTS_FAILED", "Не удалось получить список мероприятий", err)
		return
	}

	writeJSON(w, http.StatusOK, events)
}

// GET /events/{id}

func (h *Handler) GetEventByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	eventID, err := strconv.ParseInt(idStr, 10, 64)
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

	writeJSON(w, http.StatusOK, event)
}

// POST /events/{id}/register

func (h *Handler) RegisterForEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	idStr := r.PathValue("id")
	eventID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_EVENT_ID", "Неверный ID мероприятия", err)
		return
	}

	if err := h.svc.RegisterForEvent(eventID, userID); err != nil {
		switch err {
		case domain.ErrEventNotFound:
			writeError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Мероприятие не найдено", err)
		case domain.ErrUserNotFound:
			writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "Пользователь не найден", err)
		default:
			writeError(w, http.StatusBadRequest, "REGISTRATION_FAILED", err.Error(), err)
		}
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "Вы успешно зарегистрированы на мероприятие"})
}

// DELETE /events/{id}/register

func (h *Handler) CancelRegistration(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	idStr := r.PathValue("id")
	eventID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_EVENT_ID", "Неверный ID мероприятия", err)
		return
	}

	if err := h.svc.CancelRegistration(eventID, userID); err != nil {
		switch err {
		case domain.ErrEventNotFound:
			writeError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Мероприятие не найдено", err)
		default:
			writeError(w, http.StatusBadRequest, "CANCEL_FAILED", err.Error(), err)
		}
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "Регистрация отменена"})
}

// GET /events/{id}/participants

func (h *Handler) GetEventParticipants(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	eventID, err := strconv.ParseInt(idStr, 10, 64)
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

	writeJSON(w, http.StatusOK, map[string]any{
		"event_id":     eventID,
		"participants": participants,
		"count":        len(participants),
	})
}

// POST /admin/events/{id}/approve (для Админа)

func (h *Handler) ApproveEvent(w http.ResponseWriter, r *http.Request) {
	adminID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	idStr := r.PathValue("id")
	eventID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_EVENT_ID", "Неверный ID мероприятия", err)
		return
	}

	if err := h.svc.ApproveAndAwardPoints(eventID, adminID); err != nil {
		switch err {
		case domain.ErrInvalidRole:
			writeError(w, http.StatusForbidden, "FORBIDDEN", "Только администратор может подтверждать мероприятия", err)
		case domain.ErrEventNotFound:
			writeError(w, http.StatusNotFound, "EVENT_NOT_FOUND", "Мероприятие не найдено", err)
		default:
			writeError(w, http.StatusBadRequest, "APPROVE_FAILED", err.Error(), err)
		}
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "Баллы начислены, мероприятие закрыто"})
}

// POST /admin/users/{id}/promote (для Админа)

func (h *Handler) PromoteUser(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Неверный формат запроса", err)
		return
	}

	if req.Role == "" {
		writeError(w, http.StatusBadRequest, "MISSING_ROLE", "Роль обязательна", nil)
		return
	}

	if err := h.svc.PromoteUser(targetID, req.Role, requesterID); err != nil {
		switch err {
		case domain.ErrInvalidRole:
			writeError(w, http.StatusForbidden, "FORBIDDEN", "Недостаточно прав для изменения роли", err)
		case domain.ErrUserNotFound:
			writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "Пользователь не найден", err)
		default:
			writeError(w, http.StatusBadRequest, "PROMOTE_FAILED", err.Error(), err)
		}
		return
	}

	writeJSON(w, http.StatusOK, MessageResponse{Message: "Роль успешно изменена"})
}

// GET /auth/me

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация", nil)
		return
	}

	user, err := h.svc.GetUserByID(userID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound):
			writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "Пользователь не найден", err)
		default:
			writeError(w, http.StatusInternalServerError, "GET_USER_FAILED", "Не удалось получить данные пользователя", err)
		}
		return
	}

	writeJSON(w, http.StatusOK, user)
}
