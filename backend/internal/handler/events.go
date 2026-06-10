package handler

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// decorate дополняет мероприятия счётчиками и статусом набора.
func (a *App) decorate(ev []Event) {
	for i := range ev {
		SyncEventStatus(&ev[i])
		ac, rc := 0, 0
		for _, p := range a.DB.Participants {
			if p.EventID == ev[i].ID {
				if p.Status == "approved" {
					ac++
				}
				if p.Status == "reserve" {
					rc++
				}
			}
		}
		ev[i].ParticipantsCount = ac
		ev[i].ReserveCount = rc
		if ev[i].Status == "ongoing" {
			ev[i].RecruitmentStatus = "going"
		} else if ev[i].MaxParticipants != nil && ac >= *ev[i].MaxParticipants && ev[i].ReserveParticipants > 0 {
			ev[i].RecruitmentStatus = "Есть резервные места"
		} else {
			ev[i].RecruitmentStatus = "В поиске волонтеров"
		}
	}
}

// eventRoutes маршрутизирует вложенные actions внутри /events.
func (a *App) eventRoutes(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 1 {
		a.auth(a.events)(w, r)
		return
	}
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}
	if len(parts) == 2 && r.Method == "GET" {
		a.auth(a.eventDetail)(w, r)
		return
	}
	action := parts[len(parts)-1]
	switch {
	case action == "register":
		a.auth(a.register)(w, r)
	case action == "unregister":
		a.auth(a.unregister)(w, r)
	case action == "review":
		a.auth(a.review)(w, r)
	case action == "manage-volunteers":
		a.roles(a.manage, "leader", "organizer")(w, r)
	case action == "export-volunteers.xlsx":
		a.roles(a.export, "leader", "organizer")(w, r)
	case len(parts) == 5 && parts[2] == "applications" && parts[4] == "status":
		a.roles(a.appStatus, "leader", "organizer")(w, r)
	default:
		http.NotFound(w, r)
	}
}

// events показывает список будущих или прошедших мероприятий.
func (a *App) events(c *AppContext) {
	tab := c.R.URL.Query().Get("tab")
	if tab == "" {
		tab = "future"
	}
	var ev []Event
	for i := range a.DB.Events {
		e := &a.DB.Events[i]
		finalize(a.DB, e)
		if (tab == "past" && e.EndDate.Before(Now())) || (tab != "past" && !e.EndDate.Before(Now())) {
			ev = append(ev, *e)
		}
	}
	if tab == "past" {
		sort.Slice(ev, func(i, j int) bool { return ev[i].EndDate.After(ev[j].EndDate) })
	} else {
		sort.Slice(ev, func(i, j int) bool { return ev[i].StartDate.Before(ev[j].StartDate) })
	}
	a.decorate(ev)
	a.DB.SaveUnlocked()
	a.render(c, "events", ViewData{"Events": ev, "Tab": tab}, 200)
}

// eventID извлекает ID мероприятия из URL.
func (a *App) eventID(r *http.Request) int {
	p := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(p) > 1 {
		return atoi(p[1])
	}
	return 0
}

// adminPathID извлекает ID ресурса из admin URL.
func adminPathID(r *http.Request) int {
	p := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(p) > 2 {
		return atoi(p[2])
	}
	return 0
}

// formInts читает repeated int-поля из multipart/form запроса.
func formInts(r *http.Request, name string) []int {
	_ = r.ParseMultipartForm(32 << 20)
	vals := r.Form[name]
	out := make([]int, 0, len(vals))
	for _, v := range vals {
		if n := atoi(v); n > 0 {
			out = append(out, n)
		}
	}
	return out
}

// canManage проверяет право управления мероприятием.
func canManage(s *Store, u *User, e *Event) bool {
	if u.Role == "leader" {
		return true
	}
	if u.Role != "organizer" {
		return false
	}
	if e.CreatedByID == u.ID {
		return true
	}
	for _, o := range s.Organizers {
		if o.EventID == e.ID && o.UserID == u.ID {
			return true
		}
	}
	return false
}

// eventDetail показывает карточку мероприятия, роли, отзывы и заявку.
func (a *App) eventDetail(c *AppContext) {
	e := a.DB.Event(a.eventID(c.R))
	if e == nil {
		http.NotFound(c.W, c.R)
		return
	}
	finalize(a.DB, e)
	org := a.DB.User(e.CreatedByID)
	var app EventParticipant
	has := false
	for _, p := range a.DB.Participants {
		if p.EventID == e.ID && p.UserID == c.User.ID {
			app = p
			has = true
		}
	}
	var roles []EventRole
	for _, r := range a.DB.EventRoles {
		if r.EventID == e.ID && r.IsActive {
			roles = append(roles, r)
		}
	}
	type RR struct {
		Username, FullName string
		Rating             int
		Comment            string
		CreatedAt          time.Time
	}
	var revs []RR
	hasRev := false
	for _, r := range a.DB.Reviews {
		if r.EventID == e.ID {
			if r.UserID == c.User.ID {
				hasRev = true
			}
			if u := a.DB.User(r.UserID); u != nil {
				revs = append(revs, RR{u.Username, u.FullName, r.Rating, r.Comment, r.CreatedAt})
			}
		}
	}
	pc := 0
	for _, p := range a.DB.Participants {
		if p.EventID == e.ID && p.Status == "approved" {
			pc++
		}
	}
	a.render(c, "event_detail", ViewData{"Event": e, "Organizer": org, "Application": app, "HasApplication": has, "IsRegistered": has && app.Status != "rejected", "CanRegister": !canManage(a.DB, c.User, e), "CanUnregister": has && time.Until(e.StartDate) > 24*time.Hour, "IsEventOrganizer": canManage(a.DB, c.User, e), "HasReview": hasRev, "Roles": roles, "Reviews": revs, "ParticipantsCount": pc, "ReviewsCount": len(revs)}, 200)
}

// register создаёт заявку пользователя на мероприятие.
func (a *App) register(c *AppContext) {
	if !a.check(c.W, c.R) {
		return
	}
	eid := a.eventID(c.R)
	a.DB.Lock()
	defer a.DB.Unlock()
	e := a.DB.Event(eid)
	if e == nil {
		return
	}
	for _, p := range a.DB.Participants {
		if p.EventID == eid && p.UserID == c.User.ID {
			a.flash(c.W, c.R, "повторная регистрация недоступна для активной заявки", "info")
			http.Redirect(c.W, c.R, fmt.Sprintf("/events/%d", eid), 302)
			return
		}
	}
	desiredRoleIDs := formInts(c.R, "desired_role_ids")
	var desiredRoleID *int
	if len(desiredRoleIDs) > 0 {
		first := desiredRoleIDs[0]
		desiredRoleID = &first
	}
	p := EventParticipant{ID: a.DB.NextIDUnlocked("participants"), EventID: eid, UserID: c.User.ID, Status: "pending", DesiredRoleID: desiredRoleID, CreatedAt: Now(), UpdatedAt: Now()}
	a.DB.Participants = append(a.DB.Participants, p)
	for _, roleID := range desiredRoleIDs {
		a.DB.RoleChoices = append(a.DB.RoleChoices, EventParticipantRoleChoice{ID: a.DB.NextIDUnlocked("role_choices"), ParticipantID: p.ID, RoleID: roleID, CreatedAt: Now()})
	}
	ids := map[int]bool{e.CreatedByID: true}
	for _, o := range a.DB.Organizers {
		if o.EventID == eid {
			ids[o.UserID] = true
		}
	}
	for uid := range ids {
		notify(a.DB, uid, "Новая заявка", fmt.Sprintf("%s хочет участвовать в «%s»", c.User.FullName, e.Title), "application", fmt.Sprintf("/events/%d/manage-volunteers", eid))
	}
	a.DB.SaveUnlocked()
	a.flash(c.W, c.R, "регистрация на ивент подтверждена", "info")
	http.Redirect(c.W, c.R, fmt.Sprintf("/events/%d", eid), 302)
}

// unregister удаляет заявку пользователя на мероприятие.
func (a *App) unregister(c *AppContext) {
	if !a.check(c.W, c.R) {
		return
	}
	eid := a.eventID(c.R)
	a.DB.Lock()
	for i, p := range a.DB.Participants {
		if p.EventID == eid && p.UserID == c.User.ID {
			a.DB.Participants = append(a.DB.Participants[:i], a.DB.Participants[i+1:]...)
			break
		}
	}
	a.DB.SaveUnlocked()
	a.DB.Unlock()
	a.flash(c.W, c.R, "регистрация отменена", "info")
	http.Redirect(c.W, c.R, fmt.Sprintf("/events/%d", eid), 302)
}

// review создаёт или обновляет отзыв участника после мероприятия.
func (a *App) review(c *AppContext) {
	if !a.check(c.W, c.R) {
		return
	}
	eid := a.eventID(c.R)
	e := a.DB.Event(eid)
	rating := atoi(c.R.FormValue("rating"))
	if e == nil || rating < 1 || rating > 10 {
		a.flash(c.W, c.R, "оценка должна быть от 1 до 10", "error")
		http.Redirect(c.W, c.R, fmt.Sprintf("/events/%d", eid), 302)
		return
	}
	approved := false
	for _, p := range a.DB.Participants {
		if p.EventID == eid && p.UserID == c.User.ID && p.Status == "approved" {
			approved = true
			break
		}
	}
	if !approved || Now().Before(e.EndDate) {
		a.flash(c.W, c.R, "отзыв можно оставить после завершения согласованному участнику", "error")
		http.Redirect(c.W, c.R, fmt.Sprintf("/events/%d", eid), 302)
		return
	}
	a.DB.Lock()
	updated := false
	for i := range a.DB.Reviews {
		if a.DB.Reviews[i].EventID == eid && a.DB.Reviews[i].UserID == c.User.ID {
			a.DB.Reviews[i].Rating = rating
			a.DB.Reviews[i].Comment = c.R.FormValue("comment")
			a.DB.Reviews[i].UpdatedAt = Now()
			updated = true
			break
		}
	}
	if !updated {
		a.DB.Reviews = append(a.DB.Reviews, Review{ID: a.DB.NextIDUnlocked("reviews"), EventID: eid, UserID: c.User.ID, Rating: rating, Comment: c.R.FormValue("comment"), CreatedAt: Now(), UpdatedAt: Now()})
	}
	a.DB.SaveUnlocked()
	a.DB.Unlock()
	a.flash(c.W, c.R, "отзыв сохранён", "info")
	http.Redirect(c.W, c.R, fmt.Sprintf("/events/%d", eid), 302)
}

// manage показывает заявки и участников мероприятия для организатора.
func (a *App) manage(c *AppContext) {
	eid := a.eventID(c.R)
	e := a.DB.Event(eid)
	type Row struct {
		P                 EventParticipant
		U                 User
		Desired, Assigned string
		CSRF              string
	}
	roleTitle := func(roleID *int) string {
		if roleID == nil {
			return ""
		}
		for _, role := range a.DB.EventRoles {
			if role.ID == *roleID {
				return role.Title
			}
		}
		return ""
	}
	groups := map[string][]Row{}
	for _, p := range a.DB.Participants {
		if p.EventID == eid {
			if u := a.DB.User(p.UserID); u != nil {
				groups[p.Status] = append(groups[p.Status], Row{P: p, U: *u, Desired: roleTitle(p.DesiredRoleID), Assigned: roleTitle(p.AssignedRoleID), CSRF: a.csrfToken(c.W, c.R)})
			}
		}
	}
	a.render(c, "manage_volunteers", ViewData{"Event": e, "Pending": groups["pending"], "Approved": groups["approved"], "Reserve": groups["reserve"], "Rejected": groups["rejected"]}, 200)
}

// appStatus меняет статус заявки на мероприятие.
func (a *App) appStatus(c *AppContext) {
	if !a.check(c.W, c.R) {
		return
	}
	parts := strings.Split(strings.Trim(c.R.URL.Path, "/"), "/")
	eid, aid := atoi(parts[1]), atoi(parts[3])
	st := c.R.FormValue("status")
	a.DB.Lock()
	for i := range a.DB.Participants {
		p := &a.DB.Participants[i]
		if p.ID == aid {
			p.Status = st
			if e := a.DB.Event(eid); e != nil {
				notify(a.DB, p.UserID, "Статус заявки изменён", fmt.Sprintf("Заявка на «%s»: %s", e.Title, st), "application", fmt.Sprintf("/events/%d", eid))
			}
			break
		}
	}
	a.DB.SaveUnlocked()
	a.DB.Unlock()
	http.Redirect(c.W, c.R, fmt.Sprintf("/events/%d/manage-volunteers", eid), 302)
}

// export отдаёт CSV/XLSX-compatible выгрузку волонтёров.
func (a *App) export(c *AppContext) {
	e := a.DB.Event(a.eventID(c.R))
	if e == nil {
		return
	}
	w := c.W
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=event_%d_volunteers.xlsx", e.ID))
	w.Write(ExportVolunteersXLSX(a.DB, *e))
}
