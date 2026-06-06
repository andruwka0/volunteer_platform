package handler

import (
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"volunteer-platform/internal/auth"
	"volunteer-platform/internal/config"
	"volunteer-platform/internal/domain"
	"volunteer-platform/internal/middleware"
	"volunteer-platform/internal/repository"
	"volunteer-platform/internal/service"
	tpl "volunteer-platform/internal/templates"
	"volunteer-platform/internal/upload"
)

// Now возвращает текущее UTC-время для доменной логики.
func Now() time.Time { return domain.Now() }

var Roles = domain.Roles

// HasRole проверяет, входит ли роль пользователя в разрешённый набор.
func HasRole(role string, roles ...string) bool { return domain.HasRole(role, roles...) }

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

// OpenDB сохраняет старый API открытия JSON-хранилища.
func OpenDB(cfg config.Config) (*Store, error) { return repository.OpenJSONStore(cfg.DatabaseURL) }

// HashPassword создаёт salted hash пароля для хранения.
func HashPassword(p string) (string, error) { return service.HashPassword(p) }

// VerifyPassword проверяет пароль против сохранённого hash.
func VerifyPassword(p, h string) bool { return service.VerifyPassword(p, h) }

// CreateUser валидирует данные и создаёт активного пользователя.
func CreateUser(s *Store, username, full, password, role, avatar, telegram, first, last string) (*User, error) {
	return service.CreateUser(s, username, full, password, role, avatar, telegram, first, last)
}

// UpdateProfile обновляет био и аватар пользователя.
func UpdateProfile(s *Store, u *User, bio, avatar string) error {
	return service.UpdateProfile(s, u, bio, avatar)
}

// UpdateAccountIdentity обновляет username, Telegram и ФИО аккаунта.
func UpdateAccountIdentity(s *Store, u *User, username, telegram, first, last string) error {
	return service.UpdateAccountIdentity(s, u, username, telegram, first, last)
}

// ChangePassword проверяет текущий пароль и сохраняет новый hash.
func ChangePassword(s *Store, u *User, cur, nw, confirm string) error {
	return service.ChangePassword(s, u, cur, nw, confirm)
}

// SaveUpload сохраняет multipart-файл в static uploads и возвращает URL.
func SaveUpload(file multipart.File, h *multipart.FileHeader, prefix string) (string, error) {
	return upload.SaveUpload(file, h, prefix)
}

// SaveDataURL сохраняет base64 data URL в static uploads и возвращает URL.
func SaveDataURL(dataURL, prefix string) (string, error) { return upload.SaveDataURL(dataURL, prefix) }

// SaveAvatarFromDataURL сохраняет avatar data URL как upload-файл.
func SaveAvatarFromDataURL(dataURL string) (string, error) {
	return upload.SaveAvatarFromDataURL(dataURL)
}

// SyncEventStatus описывает назначение одноимённой функции.
func SyncEventStatus(e *Event) string {
	now := Now()
	if e.Status == "closed" {
		return e.Status
	}
	if now.After(e.StartDate) && now.Before(e.EndDate) {
		e.Status = "ongoing"
	} else if now.After(e.EndDate) {
		e.Status = "closed"
	} else if e.Status == "" {
		e.Status = "soon"
	}
	return e.Status
}

// finalize закрывает прошедшее мероприятие, фиксирует рейтинг и начисляет SP.
func finalize(s *Store, e *Event) {
	if Now().Before(e.EndDate) || e.FinalRatingFixedAt != nil {
		return
	}
	var sum, c int
	for _, r := range s.Reviews {
		if r.EventID == e.ID {
			sum += r.Rating
			c++
		}
	}
	if c > 0 {
		avg := float64(sum) / float64(c)
		e.FinalRating = &avg
	}
	e.FinalReviewsCount = c
	n := Now()
	e.FinalRatingFixedAt = &n
	e.Status = "closed"
	for i := range s.Participants {
		p := s.Participants[i]
		if p.EventID == e.ID && p.Status == "approved" {
			if u := s.User(p.UserID); u != nil {
				u.SkillPoints += e.SPPoints
				notify(s, u.ID, "SP начислены", fmt.Sprintf("За мероприятие «%s» начислено %d SP", e.Title, e.SPPoints), "event", fmt.Sprintf("/events/%d", e.ID))
			}
		}
	}
}

// notify создаёт уведомление в JSONStore.
func notify(s *Store, uid int, title, msg, typ, link string) {
	s.Notifications = append(s.Notifications, Notification{ID: s.NextIDUnlocked("notifications"), UserID: uid, Title: title, Message: msg, Type: typ, LinkURL: link, CreatedAt: Now()})
}

// ExportVolunteersXLSX формирует CSV, совместимый с Excel-скачиванием.
func ExportVolunteersXLSX(s *Store, e Event) []byte {
	var b strings.Builder
	w := csv.NewWriter(&b)
	_ = w.Write([]string{"ID", "ФИО", "Username", "Telegram", "Статус"})
	for _, p := range s.Participants {
		if p.EventID == e.ID {
			if u := s.User(p.UserID); u != nil {
				_ = w.Write([]string{strconv.Itoa(u.ID), u.FullName, u.Username, u.Telegram, p.Status})
			}
		}
	}
	w.Flush()
	return []byte(b.String())
}

type App struct {
	DB        *Store
	CFG       config.Config
	Services  *service.Services
	Root      string
	Templates *template.Template
	Sessions  *auth.SessionStore
	CSRF      *auth.CSRFStore
	flashes   map[string][]string
	mu        sync.Mutex
}
type ViewData map[string]any
type AppContext struct {
	App  *App
	W    http.ResponseWriter
	R    *http.Request
	User *User
}

// NewApp создаёт HTTP-приложение с дефолтным набором services.
func NewApp(db *Store, cfg config.Config) *App { return NewWithServices(db, cfg, service.New(db)) }

// NewWithServices создаёт HTTP-приложение с явно переданными services.
func NewWithServices(db *Store, cfg config.Config, services *service.Services) *App {
	root := findProjectRoot()
	if services == nil {
		services = service.New(db)
	}
	a := &App{DB: db, CFG: cfg, Services: services, Root: root, Sessions: auth.NewSessionStore(), CSRF: auth.NewCSRFStore(), flashes: map[string][]string{}}
	a.Templates = tpl.Load(root, a.funcs())
	return a
}

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

// funcs регистрирует helper-функции для Go templates.
func (a *App) funcs() template.FuncMap {
	return template.FuncMap{
		"fmtDate": func(t time.Time, l string) string {
			if t.IsZero() {
				return ""
			}
			return t.Format(l)
		},
		"fmtDatePtr": func(t *time.Time, l string) string {
			if t == nil || t.IsZero() {
				return ""
			}
			return t.Format(l)
		},
		"maxp": func(p *int) string {
			if p == nil {
				return "∞"
			}
			return strconv.Itoa(*p)
		},
		"intPtr": func(p *int) string {
			if p == nil {
				return ""
			}
			return strconv.Itoa(*p)
		},
		"fallback": func(value, fallback string) string {
			if strings.TrimSpace(value) == "" {
				return fallback
			}
			return value
		},
		"isAdmin":         func(u *User) bool { return u != nil && HasRole(u.Role, "leader", "organizer") },
		"isVolunteerRole": func(role string) bool { return !HasRole(role, "leader", "organizer") },
		"not":             func(b bool) bool { return !b },
		"sub":             func(a, b int) int { return a - b },
		"hasID": func(ids []int, id int) bool {
			for _, x := range ids {
				if x == id {
					return true
				}
			}
			return false
		},
	}
}

// randID генерирует случайный hex ID для сессий и CSRF.
func randID() string { b := make([]byte, 16); rand.Read(b); return hex.EncodeToString(b) }

// sid возвращает текущий session id или выставляет новую cookie.
func (a *App) sid(w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie(a.CFG.SessionCookieName)
	if err == nil && c.Value != "" {
		return c.Value
	}
	v := randID()
	cookie := &http.Cookie{Name: a.CFG.SessionCookieName, Value: v, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode}
	http.SetCookie(w, cookie)
	// Make the freshly-created session visible to the rest of this request.
	// Without this, rendering a page can create one session for CSRF and then
	// another for flashes/notifications, leaving the browser with a cookie that
	// does not match the hidden CSRF token in the rendered form.
	r.AddCookie(cookie)
	return v
}

// current загружает активного пользователя из session cookie.
func (a *App) current(w http.ResponseWriter, r *http.Request) (*User, bool) {
	sid := a.sid(w, r)
	uid := a.Sessions.Get(sid)
	if uid == 0 {
		return nil, false
	}
	u := a.DB.User(uid)
	return u, u != nil && u.IsActive
}

// csrfToken возвращает или создаёт CSRF-токен текущей сессии.
func (a *App) csrfToken(w http.ResponseWriter, r *http.Request) string {
	sid := a.sid(w, r)
	if token := a.CSRF.Get(sid); token != "" {
		return token
	}
	token := randID()
	a.CSRF.Set(sid, token)
	return token
}

// check валидирует CSRF-токен state-changing запроса.
func (a *App) check(w http.ResponseWriter, r *http.Request) bool {
	if r.FormValue("csrf_token") != a.csrfToken(w, r) {
		http.Error(w, "Invalid CSRF token", 403)
		return false
	}
	return true
}

// flash сохраняет одноразовое сообщение для следующего render.
func (a *App) flash(w http.ResponseWriter, r *http.Request, msg, cat string) {
	sid := a.sid(w, r)
	a.mu.Lock()
	a.flashes[sid] = append(a.flashes[sid], cat+"|"+msg)
	a.mu.Unlock()
}

// pop забирает и очищает flash-сообщения текущей сессии.
func (a *App) pop(w http.ResponseWriter, r *http.Request) []map[string]string {
	sid := a.sid(w, r)
	a.mu.Lock()
	arr := a.flashes[sid]
	a.flashes[sid] = nil
	a.mu.Unlock()
	out := []map[string]string{}
	for _, x := range arr {
		p := strings.SplitN(x, "|", 2)
		if len(p) == 2 {
			out = append(out, map[string]string{"Category": p[0], "Message": p[1]})
		}
	}
	return out
}

// render дополняет ViewData общими данными и выполняет template.
func (a *App) render(c *AppContext, name string, d ViewData, code int) {
	if d == nil {
		d = ViewData{}
	}
	d["User"] = c.User
	d["CSRF"] = a.csrfToken(c.W, c.R)
	d["Flashes"] = a.pop(c.W, c.R)
	d["Path"] = c.R.URL.Path
	if c.User != nil {
		var ns []Notification
		unread := 0
		for _, n := range a.DB.Notifications {
			if n.UserID == c.User.ID {
				ns = append(ns, n)
				if !n.IsRead {
					unread++
				}
			}
		}
		sort.Slice(ns, func(i, j int) bool { return ns[i].CreatedAt.After(ns[j].CreatedAt) })
		if len(ns) > 20 {
			ns = ns[:20]
		}
		d["Notifications"] = ns
		d["UnreadNotificationsCount"] = unread
	}
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.W.WriteHeader(code)
	if err := a.Templates.ExecuteTemplate(c.W, name, d); err != nil {
		log.Println(err)
	}
}

// auth оборачивает handler требованием авторизации.
func (a *App) auth(h func(*AppContext)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := a.current(w, r)
		if !ok {
			http.Redirect(w, r, "/login", 302)
			return
		}
		h(&AppContext{a, w, r, u})
	}
}

// roles оборачивает handler проверкой роли пользователя.
func (a *App) roles(h func(*AppContext), roles ...string) http.HandlerFunc {
	return a.auth(func(c *AppContext) {
		if !HasRole(c.User.Role, roles...) {
			a.render(c, "error", ViewData{"StatusCode": 403, "Message": "недостаточно прав"}, 403)
			return
		}
		h(c)
	})
}

// Routes собирает ServeMux приложения и подключает security headers.
func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(a.Root, "app", "static")))))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			a.login(w, r)
		} else {
			a.render(&AppContext{a, w, r, nil}, "login", nil, 200)
		}
	})
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		if a.check(w, r) {
			sid := a.sid(w, r)
			a.Sessions.Delete(sid)
			http.Redirect(w, r, "/login", 302)
		}
	})
	mux.HandleFunc("/", a.dispatch)
	return middleware.SecurityHeaders(mux)
}

// dispatch маршрутизирует вложенные URL к тематическим handler-группам.
func (a *App) dispatch(w http.ResponseWriter, r *http.Request) {
	p := strings.Trim(r.URL.Path, "/")
	parts := []string{}
	if p != "" {
		parts = strings.Split(p, "/")
	}
	if len(parts) == 0 {
		a.auth(a.dashboard)(w, r)
		return
	}
	if parts[0] == "people" {
		if len(parts) == 1 {
			a.auth(a.people)(w, r)
		} else {
			a.auth(a.publicProfile)(w, r)
		}
		return
	}
	if parts[0] == "events" {
		a.eventRoutes(w, r, parts)
		return
	}
	if parts[0] == "profile" {
		a.profileRoutes(w, r, parts)
		return
	}
	if parts[0] == "leaderboard" {
		a.auth(a.leaderboard)(w, r)
		return
	}
	if parts[0] == "rules" {
		a.auth(a.rules)(w, r)
		return
	}
	if parts[0] == "notifications" {
		a.auth(a.notifications)(w, r)
		return
	}
	if parts[0] == "admin" {
		a.adminRoutes(w, r, parts)
		return
	}
	http.NotFound(w, r)
}

// atoi безопасно преобразует строку в int с нулём по умолчанию.
func atoi(s string) int { n, _ := strconv.Atoi(s); return n }
