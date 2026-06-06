package platform

import (
	"mime/multipart"
	"os"
	"path/filepath"

	"volunteer-platform/internal/config"
	"volunteer-platform/internal/domain"
	"volunteer-platform/internal/handler"
	"volunteer-platform/internal/repository"
)

type Config = config.Config
type App = handler.App
type ViewData = handler.ViewData
type AppContext = handler.AppContext

type User = domain.User
type Event = domain.Event
type Review = domain.Review
type Rule = domain.Rule
type EventParticipant = domain.EventParticipant
type EventOrganizer = domain.EventOrganizer
type EventRole = domain.EventRole
type EventParticipantRoleChoice = domain.EventParticipantRoleChoice
type Notification = domain.Notification
type Store = repository.JSONStore

var Roles = domain.Roles
var Now = domain.Now
var HasRole = domain.HasRole

// LoadConfig сохраняет старый API загрузки конфигурации.
func LoadConfig() Config { return config.Load() }

// OpenDB сохраняет старый API открытия JSON-хранилища.
func OpenDB(cfg Config) (*Store, error) { return handler.OpenDB(cfg) }

// NewApp создаёт HTTP-приложение с дефолтным набором services.
func NewApp(db *Store, cfg Config) *App { return handler.NewApp(db, cfg) }

// HashPassword создаёт salted hash пароля для хранения.
func HashPassword(p string) (string, error) { return handler.HashPassword(p) }

// VerifyPassword проверяет пароль против сохранённого hash.
func VerifyPassword(p, h string) bool { return handler.VerifyPassword(p, h) }

// CreateUser валидирует данные и создаёт активного пользователя.
func CreateUser(s *Store, username, full, password, role, avatar, telegram, first, last string) (*User, error) {
	return handler.CreateUser(s, username, full, password, role, avatar, telegram, first, last)
}

// UpdateProfile обновляет био и аватар пользователя.
func UpdateProfile(s *Store, u *User, bio, avatar string) error {
	return handler.UpdateProfile(s, u, bio, avatar)
}

// UpdateAccountIdentity обновляет username, Telegram и ФИО аккаунта.
func UpdateAccountIdentity(s *Store, u *User, username, telegram, first, last string) error {
	return handler.UpdateAccountIdentity(s, u, username, telegram, first, last)
}

// ChangePassword проверяет текущий пароль и сохраняет новый hash.
func ChangePassword(s *Store, u *User, cur, nw, confirm string) error {
	return handler.ChangePassword(s, u, cur, nw, confirm)
}

// SaveUpload сохраняет multipart-файл в static uploads и возвращает URL.
func SaveUpload(file multipart.File, h *multipart.FileHeader, prefix string) (string, error) {
	return handler.SaveUpload(file, h, prefix)
}

// SaveDataURL сохраняет base64 data URL в static uploads и возвращает URL.
func SaveDataURL(dataURL, prefix string) (string, error) { return handler.SaveDataURL(dataURL, prefix) }

// SaveAvatarFromDataURL сохраняет avatar data URL как upload-файл.
func SaveAvatarFromDataURL(dataURL string) (string, error) {
	return handler.SaveAvatarFromDataURL(dataURL)
}

// SyncEventStatus описывает назначение одноимённой функции.
func SyncEventStatus(e *Event) string { return handler.SyncEventStatus(e) }

// ExportVolunteersXLSX формирует CSV, совместимый с Excel-скачиванием.
func ExportVolunteersXLSX(s *Store, e Event) []byte { return handler.ExportVolunteersXLSX(s, e) }

// findProjectRoot ищет корень проекта по app/templates/base.html.
func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "app", "templates", "base.html")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}
