package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const tokenCookieName = "volunteer_token"

type Server struct {
	backend *url.URL
	client  *http.Client
	tpl     *template.Template
}

type PageData struct {
	Title     string
	Path      string
	User      *User
	Events    []EventView
	Event     *Event
	Error     string
	Message   string
	Form      map[string]string
	Merch     []MerchLevel
	NextLevel int64
	Progress  int64
}

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

type EventView struct {
	Event
	Participants []int64
	IsRegistered bool
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type ParticipantsResponse struct {
	EventID      int64   `json:"event_id"`
	Participants []int64 `json:"participants"`
	Count        int     `json:"count"`
}

type MerchLevel struct {
	Need  int64
	Title string
}

func main() {
	backendRaw := getenv("BACKEND_URL", "http://127.0.0.1:8080")
	backendURL, err := url.Parse(backendRaw)
	if err != nil {
		log.Fatalf("invalid BACKEND_URL %q: %v", backendRaw, err)
	}

	srv := &Server{
		backend: backendURL,
		client:  &http.Client{Timeout: 10 * time.Second},
		tpl:     template.Must(template.New("pages").Funcs(template.FuncMap{"fmtDate": fmtDate, "roleName": roleName, "canManage": canManage, "isAdmin": isAdmin, "joinIDs": joinIDs, "dict": dict}).Parse(pagesHTML)),
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join("static")))))
	mux.HandleFunc("/", srv.home)
	mux.HandleFunc("/login", srv.login)
	mux.HandleFunc("/register", srv.register)
	mux.HandleFunc("/logout", srv.logout)
	mux.HandleFunc("/events", srv.events)
	mux.HandleFunc("/events/", srv.eventAction)
	mux.HandleFunc("/profile", srv.profile)
	mux.HandleFunc("/leaderboard", srv.leaderboard)
	mux.HandleFunc("/rules", srv.rules)
	mux.HandleFunc("/people", srv.people)
	mux.HandleFunc("/admin", srv.admin)
	mux.HandleFunc("/admin/", srv.adminAction)

	addr := getenv("FRONTEND_ADDR", "127.0.0.1:5173")
	log.Printf("Frontend: http://%s", addr)
	log.Printf("Backend API: %s", backendURL.String())
	log.Fatal(http.ListenAndServe(addr, mux))
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/events", http.StatusSeeOther)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.render(w, r, "login", PageData{Title: "Вход", Form: map[string]string{}})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		s.render(w, r, "login", PageData{Title: "Вход", Error: "Не удалось прочитать форму"})
		return
	}
	form := map[string]string{"login": r.FormValue("login")}
	token, err := s.auth("/auth/login", map[string]string{"login": r.FormValue("login"), "password": r.FormValue("password")})
	if err != nil {
		s.render(w, r, "login", PageData{Title: "Вход", Error: err.Error(), Form: form})
		return
	}
	setToken(w, token)
	http.Redirect(w, r, "/events", http.StatusSeeOther)
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.render(w, r, "register", PageData{Title: "Регистрация", Form: map[string]string{}})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		s.render(w, r, "register", PageData{Title: "Регистрация", Error: "Не удалось прочитать форму"})
		return
	}
	payload := map[string]string{
		"login":      r.FormValue("login"),
		"password":   r.FormValue("password"),
		"first_name": r.FormValue("first_name"),
		"last_name":  r.FormValue("last_name"),
		"telegram":   r.FormValue("telegram"),
	}
	token, err := s.auth("/auth/register", payload)
	if err != nil {
		s.render(w, r, "register", PageData{Title: "Регистрация", Error: err.Error(), Form: payload})
		return
	}
	setToken(w, token)
	http.Redirect(w, r, "/events", http.StatusSeeOther)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: tokenCookieName, Value: "", Path: "/", MaxAge: -1, SameSite: http.SameSiteLaxMode})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/events/new" {
		s.eventForm(w, r)
		return
	}
	if r.URL.Path != "/events" {
		http.NotFound(w, r)
		return
	}
	user, _ := s.currentUser(r)
	events, err := s.eventViews(r, user)
	data := PageData{Title: "Ивенты", User: user, Events: events}
	if err != nil {
		data.Error = err.Error()
	}
	s.render(w, r, "events", data)
}

func (s *Server) eventAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 || parts[0] != "events" {
		http.NotFound(w, r)
		return
	}
	id := parts[1]
	var method string
	switch parts[2] {
	case "register":
		method = http.MethodPost
	case "cancel":
		method = http.MethodDelete
	default:
		http.NotFound(w, r)
		return
	}
	var resp MessageResponse
	if err := s.api(r, method, "/events/"+id+"/register", nil, &resp); err != nil {
		s.flashRedirect(w, r, "/events", "error", err.Error())
		return
	}
	s.flashRedirect(w, r, "/events", "message", resp.Message)
}

func (s *Server) eventForm(w http.ResponseWriter, r *http.Request) {
	user, err := s.requireUser(w, r)
	if err != nil {
		return
	}
	if !canManage(user) {
		http.Error(w, "создавать мероприятия могут только Admin и Organizer", http.StatusForbidden)
		return
	}
	if r.Method == http.MethodGet {
		s.render(w, r, "event_form", PageData{Title: "Создать ивент", User: user, Form: map[string]string{}})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		s.render(w, r, "event_form", PageData{Title: "Создать ивент", User: user, Error: "Не удалось прочитать форму"})
		return
	}
	form := formMap(r, "title", "description", "location", "image", "start_date", "end_date", "registration_deadline", "max_participants", "reserve_participants", "skill_points")
	payload, err := eventPayload(form)
	if err != nil {
		s.render(w, r, "event_form", PageData{Title: "Создать ивент", User: user, Error: err.Error(), Form: form})
		return
	}
	var created Event
	if err := s.api(r, http.MethodPost, "/events", payload, &created); err != nil {
		s.render(w, r, "event_form", PageData{Title: "Создать ивент", User: user, Error: err.Error(), Form: form})
		return
	}
	s.flashRedirect(w, r, "/events", "message", "Мероприятие создано")
}

func (s *Server) profile(w http.ResponseWriter, r *http.Request) {
	user, err := s.requireUser(w, r)
	if err != nil {
		return
	}
	s.render(w, r, "profile", PageData{Title: "Профиль", User: user})
}

func (s *Server) leaderboard(w http.ResponseWriter, r *http.Request) {
	user, _ := s.currentUser(r)
	levels := []MerchLevel{{3, "Стикеры"}, {10, "Футболка"}, {20, "Худи"}, {35, "Подарочный набор"}}
	var next, progress int64
	if user != nil {
		for _, level := range levels {
			if user.SkillPoints < level.Need {
				next = level.Need
				break
			}
		}
		if next > 0 {
			progress = user.SkillPoints * 100 / next
		} else {
			progress = 100
		}
	}
	s.render(w, r, "leaderboard", PageData{Title: "Мерч", User: user, Merch: levels, NextLevel: next, Progress: progress})
}

func (s *Server) rules(w http.ResponseWriter, r *http.Request) {
	user, _ := s.currentUser(r)
	s.render(w, r, "rules", PageData{Title: "Наши возможности", User: user})
}

func (s *Server) people(w http.ResponseWriter, r *http.Request) {
	user, _ := s.currentUser(r)
	s.render(w, r, "people", PageData{Title: "Поиск", User: user})
}

func (s *Server) admin(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/admin" {
		http.NotFound(w, r)
		return
	}
	user, err := s.requireUser(w, r)
	if err != nil {
		return
	}
	if !isAdmin(user) {
		http.Error(w, "Админка доступна только Admin", http.StatusForbidden)
		return
	}
	events, _ := s.eventViews(r, user)
	s.render(w, r, "admin", PageData{Title: "Админка", User: user, Events: events})
}

func (s *Server) adminAction(w http.ResponseWriter, r *http.Request) {
	user, err := s.requireUser(w, r)
	if err != nil {
		return
	}
	if !isAdmin(user) {
		http.Error(w, "Админка доступна только Admin", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if r.URL.Path == "/admin/users/promote" {
		if err := r.ParseForm(); err != nil {
			s.flashRedirect(w, r, "/admin", "error", "Не удалось прочитать форму")
			return
		}
		id := r.FormValue("user_id")
		role := r.FormValue("role")
		var resp MessageResponse
		if err := s.api(r, http.MethodPost, "/admin/users/"+id+"/promote", map[string]string{"role": role}, &resp); err != nil {
			s.flashRedirect(w, r, "/admin", "error", err.Error())
			return
		}
		s.flashRedirect(w, r, "/admin", "message", resp.Message)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) == 4 && parts[0] == "admin" && parts[1] == "events" && parts[3] == "approve" {
		var resp MessageResponse
		if err := s.api(r, http.MethodPost, "/admin/events/"+parts[2]+"/approve", nil, &resp); err != nil {
			s.flashRedirect(w, r, "/admin", "error", err.Error())
			return
		}
		s.flashRedirect(w, r, "/admin", "message", resp.Message)
		return
	}
	http.NotFound(w, r)
}

func (s *Server) auth(path string, payload any) (string, error) {
	var resp TokenResponse
	if err := s.apiWithToken("", http.MethodPost, path, payload, &resp); err != nil {
		return "", err
	}
	if resp.Token == "" {
		return "", errors.New("backend не вернул token")
	}
	return resp.Token, nil
}

func (s *Server) currentUser(r *http.Request) (*User, error) {
	var user User
	if err := s.api(r, http.MethodGet, "/auth/me", nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Server) requireUser(w http.ResponseWriter, r *http.Request) (*User, error) {
	user, err := s.currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return nil, err
	}
	return user, nil
}

func (s *Server) eventViews(r *http.Request, user *User) ([]EventView, error) {
	var events []Event
	if err := s.api(r, http.MethodGet, "/events", nil, &events); err != nil {
		return nil, err
	}
	views := make([]EventView, 0, len(events))
	for _, event := range events {
		view := EventView{Event: event}
		var participants ParticipantsResponse
		if err := s.api(r, http.MethodGet, fmt.Sprintf("/events/%d/participants", event.ID), nil, &participants); err == nil {
			view.Participants = participants.Participants
			if user != nil {
				for _, id := range participants.Participants {
					if id == user.ID {
						view.IsRegistered = true
						break
					}
				}
			}
		}
		views = append(views, view)
	}
	return views, nil
}

func (s *Server) api(r *http.Request, method, path string, payload any, out any) error {
	token := ""
	if cookie, err := r.Cookie(tokenCookieName); err == nil {
		token = cookie.Value
	}
	return s.apiWithToken(token, method, path, payload, out)
}

func (s *Server) apiWithToken(token, method, path string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(payload); err != nil {
			return err
		}
		body = buf
	}
	endpoint := s.backend.ResolveReference(&url.URL{Path: path})
	req, err := http.NewRequest(method, endpoint.String(), body)
	if err != nil {
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("backend недоступен (%s). Запустите: cd backend && go run ./cmd/server", err.Error())
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiErr APIError
		if json.Unmarshal(data, &apiErr) == nil && apiErr.Message != "" {
			return errors.New(apiErr.Message)
		}
		text := strings.TrimSpace(string(data))
		if text == "" {
			text = resp.Status
		}
		return errors.New(text)
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

func (s *Server) render(w http.ResponseWriter, r *http.Request, name string, data PageData) {
	if data.User == nil {
		data.User, _ = s.currentUser(r)
	}
	data.Path = r.URL.Path
	if data.Message == "" && r.URL.Query().Get("message") != "" {
		data.Message = r.URL.Query().Get("message")
	}
	if data.Error == "" && r.URL.Query().Get("error") != "" {
		data.Error = r.URL.Query().Get("error")
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("render %s: %v", name, err)
	}
}

func (s *Server) flashRedirect(w http.ResponseWriter, r *http.Request, target, key, value string) {
	q := url.Values{}
	q.Set(key, value)
	http.Redirect(w, r, target+"?"+q.Encode(), http.StatusSeeOther)
}

func setToken(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{Name: tokenCookieName, Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 60 * 60 * 24 * 30})
}

func formMap(r *http.Request, keys ...string) map[string]string {
	m := make(map[string]string, len(keys))
	for _, key := range keys {
		m[key] = r.FormValue(key)
	}
	return m
}

func eventPayload(form map[string]string) (map[string]any, error) {
	start, err := parseDateTime(form["start_date"])
	if err != nil {
		return nil, fmt.Errorf("неверная дата начала: %w", err)
	}
	end, err := parseDateTime(form["end_date"])
	if err != nil {
		return nil, fmt.Errorf("неверная дата окончания: %w", err)
	}
	payload := map[string]any{
		"title":                 form["title"],
		"description":           form["description"],
		"location":              form["location"],
		"image":                 form["image"],
		"start_date":            start.Format(time.RFC3339),
		"end_date":              end.Format(time.RFC3339),
		"registration_deadline": nil,
		"max_participants":      nil,
		"reserve_participants":  parseIntDefault(form["reserve_participants"], 0),
		"skill_points":          parseIntDefault(form["skill_points"], 0),
	}
	if strings.TrimSpace(form["registration_deadline"]) != "" {
		deadline, err := parseDateTime(form["registration_deadline"])
		if err != nil {
			return nil, fmt.Errorf("неверный дедлайн регистрации: %w", err)
		}
		payload["registration_deadline"] = deadline.Format(time.RFC3339)
	}
	if strings.TrimSpace(form["max_participants"]) != "" {
		payload["max_participants"] = parseIntDefault(form["max_participants"], 0)
	}
	return payload, nil
}

func parseDateTime(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02T15:04", value, time.Local)
}

func parseIntDefault(value string, fallback int64) int64 {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func fmtDate(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.Format("02.01.2006 15:04")
}

func roleName(role string) string {
	switch role {
	case "Admin":
		return "Администратор"
	case "Organizer":
		return "Организатор"
	case "Volunteer":
		return "Волонтёр"
	default:
		return role
	}
}

func canManage(user *User) bool {
	return user != nil && (user.Role == "Admin" || user.Role == "Organizer")
}

func isAdmin(user *User) bool {
	return user != nil && user.Role == "Admin"
}

func joinIDs(ids []int64) string {
	if len(ids) == 0 {
		return "пока нет"
	}
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.FormatInt(id, 10)
	}
	return strings.Join(parts, ", ")
}

func dict(values ...string) map[string]string {
	m := map[string]string{}
	for i := 0; i+1 < len(values); i += 2 {
		m[values[i]] = values[i+1]
	}
	return m
}

const pagesHTML = `
{{define "header"}}
<!doctype html><html lang="ru"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>{{.Title}} · volunteer rating platform</title><link rel="stylesheet" href="/static/css/style.css?v=go-frontend-3"></head><body>
<nav class="top-nav"><a href="/events">Ивенты</a><a href="/leaderboard">Мерч</a><a href="/rules">Наши возможности</a><a href="/people">Поиск</a>{{if .User}}<a href="/profile">Профиль</a>{{end}}{{if isAdmin .User}}<a href="/admin">Админка</a>{{end}}<div class="top-nav-actions">{{if .User}}<span class="text-muted">{{.User.FirstName}} {{.User.LastName}} · {{roleName .User.Role}}</span><form method="post" action="/logout" class="inline-form"><button class="btn btn-secondary">Выйти</button></form>{{else}}<a class="btn btn-secondary" href="/login">Войти</a>{{end}}</div></nav>
<main class="container">{{if .Message}}<div class="toast toast-info">{{.Message}}</div>{{end}}{{if .Error}}<div class="toast toast-error">{{.Error}}</div>{{end}}
{{end}}
{{define "footer"}}</main><script src="/static/js/main.js?v=go-frontend-3"></script></body></html>{{end}}

{{define "login"}}<!doctype html><html lang="ru"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>{{.Title}}</title><link rel="stylesheet" href="/static/css/style.css?v=go-frontend-3"></head><body><main class="container auth-grid"><section class="card"><h1>Вход</h1>{{if .Error}}<div class="toast toast-error">{{.Error}}</div>{{end}}<p class="text-muted">Войдите по логину и паролю. Backend token будет сохранён в защищённой cookie frontend-сервера.</p><form method="post" action="/login"><input name="login" placeholder="login" value="{{index .Form "login"}}" required><input type="password" name="password" placeholder="password" required><button class="btn btn-primary">Войти</button></form><p class="text-muted">Нет аккаунта? <a href="/register">Зарегистрироваться</a></p></section></main></body></html>{{end}}

{{define "register"}}<!doctype html><html lang="ru"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>{{.Title}}</title><link rel="stylesheet" href="/static/css/style.css?v=go-frontend-3"></head><body><main class="container auth-grid"><section class="card"><h1>Регистрация</h1>{{if .Error}}<div class="toast toast-error">{{.Error}}</div>{{end}}<p class="text-muted">Регистрация использует готовый backend route <b>POST /auth/register</b>.</p><form method="post" action="/register"><div class="form-row-2"><input name="first_name" placeholder="Имя" value="{{index .Form "first_name"}}"><input name="last_name" placeholder="Фамилия" value="{{index .Form "last_name"}}"></div><input name="telegram" placeholder="Telegram" value="{{index .Form "telegram"}}"><input name="login" placeholder="login" value="{{index .Form "login"}}" required><input type="password" name="password" placeholder="password" required><button class="btn btn-primary">Создать аккаунт</button></form><p class="text-muted">Уже есть аккаунт? <a href="/login">Войти</a></p></section></main></body></html>{{end}}

{{define "events"}}{{template "header" .}}<section class="hero-card card"><div><span class="badge">Volunteer Platform</span><h1>Ивенты</h1><p class="text-muted">Многостраничный frontend на Go: формы, страницы и действия работают через готовый backend API.</p></div>{{if canManage .User}}<a class="btn btn-primary" href="/events/new">Создать ивент</a>{{else if not .User}}<a class="btn btn-primary" href="/login">Войти для записи</a>{{end}}</section><section class="event-stack">{{range .Events}}<article class="card event-row"><img class="event-avatar" src="{{if .CoverImageURL}}{{.CoverImageURL}}{{else}}/static/img/placeholder.svg{{end}}" alt="cover"><div class="event-main"><div class="event-head"><div><h2>{{.Title}}</h2><p class="text-muted">{{.Description}}</p></div><div class="event-head-chips"><span class="badge">{{.Status}}</span><span class="badge badge-success">{{.SkillPoints}} SP</span></div></div><div class="event-meta"><span>📍 {{.Location}}</span><span>🕒 {{fmtDate .StartDate}} — {{fmtDate .EndDate}}</span><span>👥 {{.ParticipantsCount}}{{if .MaxParticipants}} / {{.MaxParticipants}}{{end}}</span><span>🧾 участники ID: {{joinIDs .Participants}}</span><span>👤 организатор ID: {{.CreatedByID}}</span></div><div class="action-row">{{if $.User}}{{if .IsRegistered}}<form method="post" action="/events/{{.ID}}/cancel"><button class="btn btn-secondary">Отменить регистрацию</button></form>{{else}}<form method="post" action="/events/{{.ID}}/register"><button class="btn btn-primary" {{if ne .Status "EVENT-RECRUITING"}}disabled{{end}}>Записаться</button></form>{{end}}{{else}}<a class="btn btn-primary" href="/login">Войти и записаться</a>{{end}}</div></div></article>{{else}}<article class="card"><p class="text-muted">Пока нет мероприятий. Организатор или админ может создать первое.</p></article>{{end}}</section>{{template "footer" .}}{{end}}

{{define "event_form"}}{{template "header" .}}<section class="card"><h1>Создать ивент</h1><p class="text-muted">Форма адаптирована под backend <b>POST /events</b>: даты уходят в JSON как RFC3339, поля — snake_case.</p><form method="post" action="/events/new"><input name="title" placeholder="Название" value="{{index .Form "title"}}" required><textarea name="description" placeholder="Описание">{{index .Form "description"}}</textarea><input name="location" placeholder="Локация" value="{{index .Form "location"}}"><input name="image" placeholder="URL обложки" value="{{index .Form "image"}}"><div class="form-row-2"><label>Начало<input name="start_date" type="datetime-local" value="{{index .Form "start_date"}}" required></label><label>Окончание<input name="end_date" type="datetime-local" value="{{index .Form "end_date"}}" required></label></div><div class="form-row-2"><label>Дедлайн регистрации<input name="registration_deadline" type="datetime-local" value="{{index .Form "registration_deadline"}}"></label><label>Максимум участников<input name="max_participants" type="number" min="0" value="{{index .Form "max_participants"}}"></label></div><div class="form-row-2"><input name="reserve_participants" type="number" min="0" placeholder="Резерв" value="{{if index .Form "reserve_participants"}}{{index .Form "reserve_participants"}}{{else}}0{{end}}"><input name="skill_points" type="number" min="0" placeholder="SP" value="{{if index .Form "skill_points"}}{{index .Form "skill_points"}}{{else}}0{{end}}"></div><button class="btn btn-primary">Создать</button><a class="btn btn-secondary" href="/events">Назад</a></form></section>{{template "footer" .}}{{end}}

{{define "profile"}}{{template "header" .}}<section class="card"><h1>Профиль</h1><div class="details-grid"><div><span class="text-muted">ID</span><b>{{.User.ID}}</b></div><div><span class="text-muted">Логин</span><b>{{.User.Login}}</b></div><div><span class="text-muted">Имя</span><b>{{.User.FirstName}} {{.User.LastName}}</b></div><div><span class="text-muted">Telegram</span><b>{{.User.Telegram}}</b></div><div><span class="text-muted">Роль</span><b>{{roleName .User.Role}}</b></div><div><span class="text-muted">SP</span><b>{{.User.SkillPoints}}</b></div></div></section>{{template "footer" .}}{{end}}

{{define "leaderboard"}}{{template "header" .}}<section class="card merch-head"><h1>Мерч за Skill Points</h1>{{if .User}}<p class="merch-points">Твои SP: <b>{{.User.SkillPoints}}</b></p><div class="merch-progress"><div class="merch-progress-fill" style="width: {{.Progress}}%"></div></div>{{else}}<p class="text-muted">Войдите, чтобы видеть свои SP.</p>{{end}}</section><section class="grid merch-grid" style="margin-top:14px">{{range .Merch}}<article class="card merch-item"><p class="merch-need">{{.Need}} SP</p><h3>{{.Title}}</h3></article>{{end}}</section>{{template "footer" .}}{{end}}

{{define "rules"}}{{template "header" .}}<h1>Ваши возможности</h1><section class="card" style="margin-bottom:14px"><h3>Усиление · 3+ баллов</h3><ul><li>Проведение экскурсии по кампусу</li><li>Запуск студенческого клуба</li><li>Точечная помощь с медиаматериалами</li></ul></section><section class="card" style="margin-bottom:14px"><h3>Вовлечение · 10+ баллов</h3><ul><li>Организация активности в рамках большого мероприятия</li><li>Длительное волонтерство на выездных событиях</li><li>Подготовка репортажей</li></ul></section><section class="card"><h3>Сотворчество · 20+ баллов</h3><ul><li>Организация масштабного мероприятия</li><li>Ведение медиаканала</li><li>Разработка полезного инструмента</li></ul></section>{{template "footer" .}}{{end}}

{{define "people"}}{{template "header" .}}<section class="card"><h1>Поиск</h1><p class="text-muted">В готовом backend нет публичного endpoint для списка пользователей, поэтому frontend не подделывает данные. Админ может менять роль пользователя по ID в админке.</p></section>{{template "footer" .}}{{end}}

{{define "admin"}}{{template "header" .}}<section class="card"><h1>Админка</h1><p class="text-muted">Доступные операции backend: назначить Organizer/Admin по ID и закрыть завершённый ивент с начислением баллов.</p><div class="admin-grid"><form method="post" action="/admin/users/promote"><h3>Изменить роль</h3><input name="user_id" type="number" min="1" placeholder="ID пользователя" required><select name="role"><option value="Organizer">Organizer</option><option value="Admin">Admin</option></select><button class="btn btn-primary">Сохранить</button></form><div><h3>Закрыть ивент</h3><p class="text-muted">Backend разрешит действие только для EVENT-FINISHED.</p></div></div></section><section class="event-stack" style="margin-top:14px">{{range .Events}}<article class="card"><div class="event-head"><div><h3>#{{.ID}} {{.Title}}</h3><p class="text-muted">{{.Status}} · {{fmtDate .StartDate}} — {{fmtDate .EndDate}}</p></div><form method="post" action="/admin/events/{{.ID}}/approve"><button class="btn btn-secondary">Начислить баллы / закрыть</button></form></div></article>{{else}}<article class="card"><p class="text-muted">Ивентов пока нет</p></article>{{end}}</section>{{template "footer" .}}{{end}}
`
