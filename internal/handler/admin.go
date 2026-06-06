package handler

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"
)

// parseDate парсит дату формы в формате YYYY-MM-DD.
func parseDate(v string) time.Time { t, _ := time.Parse("2006-01-02", v); return t.UTC() }

// pint преобразует строку в *int для nullable form-полей.
func pint(v string) *int {
	if v == "" {
		return nil
	}
	n := atoi(v)
	return &n
}

// adminRoutes маршрутизирует вложенные URL админки.
func (a *App) adminRoutes(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 1 {
		a.roles(a.adminDashboard, "leader", "organizer")(w, r)
		return
	}
	switch parts[1] {
	case "users":
		if len(parts) == 2 {
			a.roles(a.adminUsers, "leader", "organizer")(w, r)
		} else if parts[2] == "new" {
			a.roles(a.adminUserNew, "leader")(w, r)
		} else if len(parts) > 3 && parts[3] == "disable" {
			a.roles(a.adminUserDisable, "leader", "organizer")(w, r)
		} else if len(parts) > 3 && parts[3] == "delete" {
			a.roles(a.adminUserDelete, "leader", "organizer")(w, r)
		}
	case "events":
		if len(parts) == 2 {
			a.roles(a.adminEvents, "leader", "organizer")(w, r)
		} else if parts[2] == "new" {
			a.roles(a.adminEventNew, "leader", "organizer")(w, r)
		} else if len(parts) > 3 && parts[3] == "edit" {
			a.roles(a.adminEventEdit, "leader", "organizer")(w, r)
		} else if len(parts) > 3 && parts[3] == "delete" {
			a.roles(a.adminEventDelete, "leader", "organizer")(w, r)
		} else if len(parts) > 3 && parts[3] == "finish" {
			a.roles(a.adminEventFinish, "leader", "organizer")(w, r)
		}
	case "reviews":
		if len(parts) == 2 {
			a.roles(a.adminReviews, "leader", "organizer")(w, r)
		} else {
			a.roles(a.adminReviewsEvent, "leader", "organizer")(w, r)
		}
	case "roles":
		a.roles(a.adminRoles, "leader", "organizer")(w, r)
	default:
		http.NotFound(w, r)
	}
}

// adminDashboard показывает стартовую страницу админки.
func (a *App) adminDashboard(c *AppContext) {
	a.render(c, "admin_dashboard", ViewData{"UsersCount": len(a.DB.Users), "EventsCount": len(a.DB.Events), "ReviewsCount": len(a.DB.Reviews), "RolesCount": len(Roles)}, 200)
}

// adminUsers показывает список пользователей в админке.
func (a *App) adminUsers(c *AppContext) {
	a.render(c, "admin_users", ViewData{"Users": a.DB.Users}, 200)
}

// adminUserNew создаёт пользователя из админской формы.
func (a *App) adminUserNew(c *AppContext) {
	if c.R.Method == "GET" {
		a.render(c, "admin_user_form", nil, 200)
		return
	}
	if !a.check(c.W, c.R) {
		return
	}
	_, err := CreateUser(a.DB, c.R.FormValue("username"), c.R.FormValue("full_name"), c.R.FormValue("password"), c.R.FormValue("role"), "", c.R.FormValue("telegram"), c.R.FormValue("first_name"), c.R.FormValue("last_name"))
	if err != nil {
		a.render(c, "admin_user_form", ViewData{"Error": err.Error()}, 400)
		return
	}
	a.flash(c.W, c.R, "пользователь создан", "info")
	http.Redirect(c.W, c.R, "/admin/users", 302)
}

// adminUserDisable отключает пользователя без удаления данных.
func (a *App) adminUserDisable(c *AppContext) {
	if !a.check(c.W, c.R) {
		return
	}
	parts := strings.Split(strings.Trim(c.R.URL.Path, "/"), "/")
	uid := atoi(parts[2])
	if u := a.DB.User(uid); u != nil && u.ID != c.User.ID {
		a.DB.Lock()
		u.IsActive = !u.IsActive
		a.DB.SaveUnlocked()
		a.DB.Unlock()
	}
	http.Redirect(c.W, c.R, "/admin/users", 302)
}

// adminUserDelete удаляет пользователя из JSONStore.
func (a *App) adminUserDelete(c *AppContext) {
	if !a.check(c.W, c.R) {
		return
	}
	parts := strings.Split(strings.Trim(c.R.URL.Path, "/"), "/")
	uid := atoi(parts[2])
	a.DB.Lock()
	for i, u := range a.DB.Users {
		if u.ID == uid && uid != c.User.ID {
			a.DB.Users = append(a.DB.Users[:i], a.DB.Users[i+1:]...)
			break
		}
	}
	a.DB.SaveUnlocked()
	a.DB.Unlock()
	http.Redirect(c.W, c.R, "/admin/users", 302)
}

// lists добавляет в ViewData общие списки пользователей и мероприятий.
func (a *App) lists(d ViewData) {
	var v, o []User
	for _, u := range a.DB.Users {
		if HasRole(u.Role, "junior_volunteer", "middle_volunteer", "senior_volunteer") {
			v = append(v, u)
		}
		if HasRole(u.Role, "leader", "organizer") {
			o = append(o, u)
		}
	}
	d["Volunteers"] = v
	d["Organizers"] = o
}

// addEventSelectionData добавляет выбранных участников/организаторов мероприятия.
func (a *App) addEventSelectionData(d ViewData, eventID int) {
	selectedParticipants := []int{}
	selectedOrganizers := []int{}
	roles := []EventRole{}
	for _, p := range a.DB.Participants {
		if p.EventID == eventID {
			selectedParticipants = append(selectedParticipants, p.UserID)
		}
	}
	for _, o := range a.DB.Organizers {
		if o.EventID == eventID {
			selectedOrganizers = append(selectedOrganizers, o.UserID)
		}
	}
	for _, role := range a.DB.EventRoles {
		if role.EventID == eventID {
			roles = append(roles, role)
		}
	}
	sort.Slice(roles, func(i, j int) bool { return roles[i].SortOrder < roles[j].SortOrder })
	d["SelectedParticipants"] = selectedParticipants
	d["SelectedOrganizers"] = selectedOrganizers
	d["EventRoles"] = roles
}

// adminEvents показывает список мероприятий в админке.
func (a *App) adminEvents(c *AppContext) {
	a.render(c, "admin_events", ViewData{"Events": a.DB.Events}, 200)
}

// adminEventNew создаёт новое мероприятие из формы.
func (a *App) adminEventNew(c *AppContext) {
	d := ViewData{"Event": Event{}, "EventRoles": []EventRole{}, "SelectedParticipants": []int{}, "SelectedOrganizers": []int{}}
	a.lists(d)
	if c.R.Method == "GET" {
		a.render(c, "admin_event_form", d, 200)
		return
	}
	if !a.check(c.W, c.R) {
		return
	}
	e, err := a.formEvent(c, Event{ID: a.DB.NextID("events"), CreatedByID: c.User.ID, Status: "soon", CreatedAt: Now(), UpdatedAt: Now()})
	if err != nil {
		d["Error"] = err.Error()
		a.render(c, "admin_event_form", d, 400)
		return
	}
	a.DB.Lock()
	a.DB.Events = append(a.DB.Events, e)
	a.saveEventFormRelationsLocked(c, e.ID)
	a.DB.SaveUnlocked()
	a.DB.Unlock()
	a.flash(c.W, c.R, "ивент создан", "info")
	http.Redirect(c.W, c.R, "/admin/events", 302)
}

// adminEventEdit редактирует мероприятие по ID из admin URL.
func (a *App) adminEventEdit(c *AppContext) {
	eid := adminPathID(c.R)
	e := a.DB.Event(eid)
	if e == nil {
		a.render(c, "error", ViewData{"StatusCode": 404, "Message": "страница не найдена"}, 404)
		return
	}
	d := ViewData{"Event": e}
	a.addEventSelectionData(d, e.ID)
	a.lists(d)
	if c.R.Method == "GET" {
		a.render(c, "admin_event_edit", d, 200)
		return
	}
	if !a.check(c.W, c.R) {
		return
	}
	ne, err := a.formEvent(c, *e)
	if err != nil {
		d["Error"] = err.Error()
		a.render(c, "admin_event_edit", d, 400)
		return
	}
	a.DB.Lock()
	*e = ne
	a.saveEventFormRelationsLocked(c, e.ID)
	a.DB.SaveUnlocked()
	a.DB.Unlock()
	a.flash(c.W, c.R, "ивент обновлён", "info")
	http.Redirect(c.W, c.R, "/admin/events", 302)
}

// formEvent парсит форму мероприятия и возвращает обновлённую domain-модель.
func (a *App) formEvent(c *AppContext, e Event) (Event, error) {
	_ = c.R.ParseMultipartForm(32 << 20)
	title := strings.TrimSpace(c.R.FormValue("title"))
	if title == "" {
		return e, errors.New("заполни название ивента")
	}
	start := parseDate(c.R.FormValue("start_date"))
	end := parseDate(c.R.FormValue("end_date"))
	if start.IsZero() || end.IsZero() {
		return e, errors.New("укажи даты мероприятия")
	}
	if end.Before(start) {
		return e, errors.New("дата завершения не может быть раньше даты начала")
	}
	e.Title = title
	e.ShortDescription = strings.TrimSpace(c.R.FormValue("short_description"))
	e.FullDescription = strings.TrimSpace(c.R.FormValue("full_description"))
	e.Location = strings.TrimSpace(c.R.FormValue("location"))
	e.StartDate = start
	e.EndDate = end
	e.RatingDeadline = e.EndDate
	rd := parseDate(c.R.FormValue("registration_deadline"))
	if rd.IsZero() || c.R.FormValue("registration_until_start") == "on" {
		rd = e.StartDate
	}
	e.RegistrationDeadline = &rd
	e.MaxParticipants = pint(c.R.FormValue("max_participants"))
	if e.MaxParticipants != nil && *e.MaxParticipants < 0 {
		zero := 0
		e.MaxParticipants = &zero
	}
	e.ReserveParticipants = atoi(c.R.FormValue("reserve_participants"))
	if e.ReserveParticipants < 0 {
		e.ReserveParticipants = 0
	}
	e.SPPoints = atoi(c.R.FormValue("sp_points"))
	if e.SPPoints < 0 {
		e.SPPoints = 0
	}
	cropped := c.R.FormValue("event_cropped_data")
	e.CoverImageURL = a.Services.Events.ResolveCover(e.CoverImageURL, cropped, func(dataURL string) (string, error) {
		return SaveDataURL(dataURL, "event")
	}, func() (string, error) {
		f, h, err := c.R.FormFile("event_image_file")
		if err != nil {
			return "", err
		}
		defer f.Close()
		return SaveUpload(f, h, "event")
	})
	e.UpdatedAt = Now()
	return e, nil
}

// saveEventFormRelationsLocked сохраняет связи участников, организаторов и ролей.
func (a *App) saveEventFormRelationsLocked(c *AppContext, eventID int) {
	participantSeen := map[int]bool{}
	for _, uid := range formInts(c.R, "participant_ids") {
		participantSeen[uid] = true
		already := false
		for _, p := range a.DB.Participants {
			if p.EventID == eventID && p.UserID == uid {
				already = true
				break
			}
		}
		if !already {
			a.DB.Participants = append(a.DB.Participants, EventParticipant{ID: a.DB.NextIDUnlocked("participants"), EventID: eventID, UserID: uid, Status: "approved", CreatedAt: Now(), UpdatedAt: Now()})
		}
	}
	organizerSeen := map[int]bool{}
	for _, uid := range formInts(c.R, "organizer_ids") {
		organizerSeen[uid] = true
		already := false
		for _, o := range a.DB.Organizers {
			if o.EventID == eventID && o.UserID == uid {
				already = true
				break
			}
		}
		if !already {
			a.DB.Organizers = append(a.DB.Organizers, EventOrganizer{ID: a.DB.NextIDUnlocked("organizers"), EventID: eventID, UserID: uid, CreatedAt: Now()})
		}
	}
	for i := 0; i < len(a.DB.Organizers); i++ {
		if a.DB.Organizers[i].EventID == eventID && !organizerSeen[a.DB.Organizers[i].UserID] {
			a.DB.Organizers = append(a.DB.Organizers[:i], a.DB.Organizers[i+1:]...)
			i--
		}
	}
	_ = participantSeen
	if c.R.FormValue("no_roles") == "on" {
		return
	}
	titles := c.R.Form["role_titles"]
	caps := c.R.Form["role_capacities"]
	descs := c.R.Form["role_descriptions"]
	if len(titles) == 0 || strings.TrimSpace(titles[0]) == "" {
		return
	}
	// Replace role definitions from the form. Existing participant assignments remain by ID if reused elsewhere.
	for i := 0; i < len(a.DB.EventRoles); i++ {
		if a.DB.EventRoles[i].EventID == eventID {
			a.DB.EventRoles = append(a.DB.EventRoles[:i], a.DB.EventRoles[i+1:]...)
			i--
		}
	}
	for i, title := range titles {
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}
		capValue := 0
		if i < len(caps) {
			capValue = atoi(caps[i])
		}
		if capValue <= 0 {
			continue
		}
		desc := ""
		if i < len(descs) {
			desc = strings.TrimSpace(descs[i])
		}
		capCopy := capValue
		a.DB.EventRoles = append(a.DB.EventRoles, EventRole{ID: a.DB.NextIDUnlocked("event_roles"), EventID: eventID, Title: title, Description: desc, Capacity: &capCopy, IsActive: true, SortOrder: i})
	}
}

// adminEventDelete удаляет мероприятие из админки.
func (a *App) adminEventDelete(c *AppContext) {
	if !a.check(c.W, c.R) {
		return
	}
	eid := adminPathID(c.R)
	a.DB.Lock()
	for i, e := range a.DB.Events {
		if e.ID == eid {
			a.DB.Events = append(a.DB.Events[:i], a.DB.Events[i+1:]...)
			break
		}
	}
	a.DB.SaveUnlocked()
	a.DB.Unlock()
	http.Redirect(c.W, c.R, "/admin/events", 302)
}

// adminEventFinish закрывает мероприятие и запускает финализацию.
func (a *App) adminEventFinish(c *AppContext) {
	if !a.check(c.W, c.R) {
		return
	}
	if e := a.DB.Event(adminPathID(c.R)); e != nil {
		a.DB.Lock()
		e.EndDate = Now()
		finalize(a.DB, e)
		a.DB.SaveUnlocked()
		a.DB.Unlock()
	}
	http.Redirect(c.W, c.R, "/admin/events", 302)
}

// adminReviews показывает агрегированный список отзывов по мероприятиям.
func (a *App) adminReviews(c *AppContext) {
	type Row struct {
		EventID      int
		Title        string
		ReviewsCount int
	}
	var rows []Row
	for _, e := range a.DB.Events {
		c := 0
		for _, r := range a.DB.Reviews {
			if r.EventID == e.ID {
				c++
			}
		}
		rows = append(rows, Row{e.ID, e.Title, c})
	}
	a.render(c, "admin_reviews", ViewData{"Rows": rows}, 200)
}

// adminReviewsEvent показывает отзывы конкретного мероприятия.
func (a *App) adminReviewsEvent(c *AppContext) {
	e := a.DB.Event(adminPathID(c.R))
	if e == nil {
		a.render(c, "error", ViewData{"StatusCode": 404, "Message": "страница не найдена"}, 404)
		return
	}
	type Row struct {
		Username, FullName string
		Rating             int
		Comment            string
		CreatedAt          time.Time
	}
	var rows []Row
	for _, r := range a.DB.Reviews {
		if r.EventID == e.ID {
			if u := a.DB.User(r.UserID); u != nil {
				rows = append(rows, Row{Username: u.Username, FullName: u.FullName, Rating: r.Rating, Comment: r.Comment, CreatedAt: r.CreatedAt})
			}
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].CreatedAt.After(rows[j].CreatedAt) })
	a.render(c, "admin_reviews_event", ViewData{"Event": e, "Reviews": rows}, 200)
}

// adminRoles показывает справочник ролей и прав.
func (a *App) adminRoles(c *AppContext) {
	cards := []map[string]any{
		{"Name": "leader", "Can": []string{"полный доступ ко всему", "создание пользователей", "управление ролями"}, "Restrictions": []string{"нет"}},
		{"Name": "organizer", "Can": []string{"создание и завершение ивентов", "админка и отзывы", "управление заявками"}, "Restrictions": []string{"не может создавать пользователей"}},
		{"Name": "senior_volunteer", "Can": []string{"регистрация на ивенты", "комментарии", "мерч и рейтинг SP"}, "Restrictions": []string{"нет доступа в админку"}},
		{"Name": "middle_volunteer", "Can": []string{"регистрация на ивенты", "комментарии", "мерч и рейтинг SP"}, "Restrictions": []string{"нет доступа в админку"}},
		{"Name": "junior_volunteer", "Can": []string{"регистрация на ивенты", "комментарии", "мерч и рейтинг SP"}, "Restrictions": []string{"нет доступа в админку"}},
	}
	a.render(c, "admin_roles", ViewData{"Cards": cards}, 200)
}
