package handler

import (
	"net/http"
	"sort"
	"strings"
	"time"
)

// dashboard показывает ближайшие мероприятия на главной странице.
func (a *App) dashboard(c *AppContext) {
	var ev []Event
	for _, e := range a.DB.Events {
		if !e.EndDate.Before(Now()) {
			ev = append(ev, e)
		}
	}
	sort.Slice(ev, func(i, j int) bool { return ev[i].StartDate.Before(ev[j].StartDate) })
	if len(ev) > 5 {
		ev = ev[:5]
	}
	a.decorate(ev)
	a.render(c, "dashboard", ViewData{"Events": ev}, 200)
}

// people выполняет поиск людей и сортирует их по Skill Points.
func (a *App) people(c *AppContext) {
	q := strings.ToLower(c.R.URL.Query().Get("q"))
	var us []User
	for _, u := range a.DB.Users {
		if u.IsActive && (q == "" || strings.Contains(strings.ToLower(u.Username+u.FullName+u.Telegram), q)) {
			us = append(us, u)
		}
	}
	sort.Slice(us, func(i, j int) bool { return us[i].SkillPoints > us[j].SkillPoints })
	a.render(c, "people", ViewData{"Users": us, "Q": q}, 200)
}

// publicProfile рендерит публичный профиль пользователя.
func (a *App) publicProfile(c *AppContext) {
	uid := atoi(strings.TrimPrefix(c.R.URL.Path, "/people/"))
	u := a.DB.User(uid)
	if u == nil {
		http.NotFound(c.W, c.R)
		return
	}
	type H struct {
		EventID  int
		Title    string
		EndDate  time.Time
		SPPoints int
		Status   string
	}
	var h []H
	for _, p := range a.DB.Participants {
		if p.UserID == uid {
			if e := a.DB.Event(p.EventID); e != nil {
				h = append(h, H{e.ID, e.Title, e.EndDate, e.SPPoints, p.Status})
			}
		}
	}
	a.render(c, "public_profile", ViewData{"Profile": u, "History": h}, 200)
}
