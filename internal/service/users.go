package service

import (
	"errors"
	"regexp"
	"strings"

	"volunteer-platform/internal/domain"
	"volunteer-platform/internal/repository"
)

var usernameRE = regexp.MustCompile(`^[A-Za-z0-9_-]{3,32}$`)
var passwordRE = regexp.MustCompile(`[A-Za-z].*\d|\d.*[A-Za-z]`)

// CreateUser валидирует данные и создаёт активного пользователя.
func CreateUser(s *repository.JSONStore, username, full, password, role, avatar, telegram, first, last string) (*domain.User, error) {
	s.Lock()
	defer s.Unlock()
	username = strings.TrimSpace(username)
	if !usernameRE.MatchString(username) {
		return nil, errors.New("username: 3-32 латинских символа, цифры, _ или -")
	}
	if len(password) < 8 || !passwordRE.MatchString(password) {
		return nil, errors.New("пароль: минимум 8 символов, буква и цифра")
	}
	if !domain.HasRole(role, domain.Roles...) {
		return nil, errors.New("некорректная роль")
	}
	for _, u := range s.Users {
		if u.Username == username {
			return nil, errors.New("username уже занят")
		}
	}
	if strings.TrimSpace(full) == "" {
		full = strings.TrimSpace(first + " " + last)
	}
	if full == "" {
		return nil, errors.New("укажи имя")
	}
	hp, _ := HashPassword(password)
	u := domain.User{ID: s.NextIDUnlocked("users"), Username: username, PasswordHash: hp, FullName: full, FirstName: first, LastName: last, Telegram: telegram, Role: role, AvatarURL: avatar, IsActive: true, CreatedAt: domain.Now(), UpdatedAt: domain.Now()}
	s.Users = append(s.Users, u)
	s.SaveUnlocked()
	return &s.Users[len(s.Users)-1], nil
}

// UpdateProfile обновляет био и аватар пользователя.
func UpdateProfile(s *repository.JSONStore, u *domain.User, bio, avatar string) error {
	s.Lock()
	defer s.Unlock()
	u.Bio = bio
	if avatar != "" {
		u.AvatarURL = avatar
	}
	u.UpdatedAt = domain.Now()
	s.SaveUnlocked()
	return nil
}

// UpdateAccountIdentity обновляет username, Telegram и ФИО аккаунта.
func UpdateAccountIdentity(s *repository.JSONStore, u *domain.User, username, telegram, first, last string) error {
	s.Lock()
	defer s.Unlock()
	if !usernameRE.MatchString(username) {
		return errors.New("username: 3-32 латинских символа, цифры, _ или -")
	}
	for _, x := range s.Users {
		if x.Username == username && x.ID != u.ID {
			return errors.New("username уже занят")
		}
	}
	u.Username = username
	u.Telegram = telegram
	u.FirstName = first
	u.LastName = last
	if fn := strings.TrimSpace(first + " " + last); fn != "" {
		u.FullName = fn
	}
	u.UpdatedAt = domain.Now()
	s.SaveUnlocked()
	return nil
}

// ChangePassword проверяет текущий пароль и сохраняет новый hash.
func ChangePassword(s *repository.JSONStore, u *domain.User, cur, nw, confirm string) error {
	if !VerifyPassword(cur, u.PasswordHash) {
		return errors.New("текущий пароль неверный")
	}
	if nw != confirm {
		return errors.New("пароли не совпадают")
	}
	if len(nw) < 8 || !passwordRE.MatchString(nw) {
		return errors.New("новый пароль: минимум 8 символов, буква и цифра")
	}
	hp, _ := HashPassword(nw)
	s.Lock()
	u.PasswordHash = hp
	u.UpdatedAt = domain.Now()
	s.SaveUnlocked()
	s.Unlock()
	return nil
}
