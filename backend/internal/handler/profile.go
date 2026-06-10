package handler

import (
	"net/http"
	"sort"
	"strings"
	"time"
)

// profileRoutes маршрутизирует действия внутри /profile.
func (a *App) profileRoutes(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) == 1 && r.Method == "GET" {
		a.auth(a.profile)(w, r)
		return
	}
	if len(parts) == 1 && r.Method == "POST" {
		a.auth(a.updateProfile)(w, r)
		return
	}
	if len(parts) > 1 && parts[1] == "account" {
		a.auth(a.updateAccount)(w, r)
		return
	}
	if len(parts) > 1 && parts[1] == "password" {
		a.auth(a.updatePassword)(w, r)
		return
	}
	http.NotFound(w, r)
}

// profile показывает приватную страницу профиля.
func (a *App) profile(c *AppContext) {
	a.render(c, "profile", a.profileData(c.User, nil), 200)
}

// profileData собирает данные профиля для template.
func (a *App) profileData(user *User, extra ViewData) ViewData {
	data := ViewData{"History": a.history(user.ID), "Points": user.SkillPoints, "IsVolunteer": !HasRole(user.Role, "leader", "organizer")}
	for k, v := range extra {
		data[k] = v
	}
	return data
}

// history собирает историю участия пользователя.
func (a *App) history(uid int) []struct {
	EventID  int
	Title    string
	EndDate  time.Time
	SPPoints int
	Status   string
} {
	var h []struct {
		EventID  int
		Title    string
		EndDate  time.Time
		SPPoints int
		Status   string
	}
	for _, p := range a.DB.Participants {
		if p.UserID == uid {
			if e := a.DB.Event(p.EventID); e != nil {
				h = append(h, struct {
					EventID  int
					Title    string
					EndDate  time.Time
					SPPoints int
					Status   string
				}{e.ID, e.Title, e.EndDate, e.SPPoints, p.Status})
			}
		}
	}
	return h
}

// updateProfile обрабатывает обновление био и аватара.
func (a *App) updateProfile(c *AppContext) {
	if !a.check(c.W, c.R) {
		return
	}
	_ = c.R.ParseMultipartForm(32 << 20)
	avatar := ""
	if cropped := c.R.FormValue("avatar_cropped_data"); cropped != "" {
		avatar, _ = SaveAvatarFromDataURL(cropped)
	}
	if avatar == "" {
		if f, h, err := c.R.FormFile("avatar_file"); err == nil {
			avatar, _ = SaveUpload(f, h, "avatar")
		}
	}
	_ = UpdateProfile(a.DB, c.User, c.R.FormValue("bio"), avatar)
	a.flash(c.W, c.R, "профиль обновлён", "info")
	http.Redirect(c.W, c.R, "/profile", 302)
}

// updateAccount обрабатывает обновление username/Telegram/ФИО.
func (a *App) updateAccount(c *AppContext) {
	if !a.check(c.W, c.R) {
		return
	}
	if err := UpdateAccountIdentity(a.DB, c.User, c.R.FormValue("username"), c.R.FormValue("telegram"), c.R.FormValue("first_name"), c.R.FormValue("last_name")); err != nil {
		a.render(c, "profile", a.profileData(c.User, ViewData{"SecurityError": err.Error(), "OpenSecurityModal": true}), 400)
		return
	}
	a.flash(c.W, c.R, "данные аккаунта обновлены", "info")
	http.Redirect(c.W, c.R, "/profile", 302)
}

// updatePassword обрабатывает смену пароля.
func (a *App) updatePassword(c *AppContext) {
	if !a.check(c.W, c.R) {
		return
	}
	if err := ChangePassword(a.DB, c.User, c.R.FormValue("current_password"), c.R.FormValue("new_password"), c.R.FormValue("confirm_new_password")); err != nil {
		a.render(c, "profile", a.profileData(c.User, ViewData{"SecurityError": err.Error(), "OpenSecurityModal": true}), 400)
		return
	}
	a.flash(c.W, c.R, "пароль обновлён", "info")
	http.Redirect(c.W, c.R, "/profile", 302)
}

type MerchLevel struct {
	Need  int
	Title string
}

// leaderboard показывает прогресс SP и уровни мерча.
func (a *App) leaderboard(c *AppContext) {
	levels := []MerchLevel{{3, "Стикерпак ЦУ"}, {10, "Бутылка / термокружка"}, {20, "Футболка волонтёра"}, {35, "Худи комьюнити"}, {50, "Большой мерч-набор"}}
	points := c.User.SkillPoints
	next := 0
	maxNeed := levels[len(levels)-1].Need
	for _, level := range levels {
		if points < level.Need {
			next = level.Need
			break
		}
	}
	progress := 100
	if points < maxNeed {
		progress = points * 100 / maxNeed
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	a.render(c, "leaderboard", ViewData{"Points": points, "IsVolunteer": !HasRole(c.User.Role, "leader", "organizer"), "NextLevel": next, "ProgressPercent": progress, "MerchLevels": levels}, 200)
}

// rules показывает правила и возможности, отсортированные по порядку.
func (a *App) rules(c *AppContext) {
	rs := append([]Rule{}, a.DB.Rules...)
	sort.Slice(rs, func(i, j int) bool { return rs[i].OrderIndex < rs[j].OrderIndex })
	a.render(c, "rules", ViewData{"Rules": rs}, 200)
}

// notifications помечает уведомления прочитанными.
func (a *App) notifications(c *AppContext) {
	if !a.check(c.W, c.R) {
		return
	}
	path := c.R.URL.Path
	a.DB.Lock()
	if strings.HasSuffix(path, "/read-all") {
		for i := range a.DB.Notifications {
			if a.DB.Notifications[i].UserID == c.User.ID {
				a.DB.Notifications[i].IsRead = true
			}
		}
	} else {
		parts := strings.Split(strings.Trim(path, "/"), "/")
		nid := atoi(parts[1])
		for i := range a.DB.Notifications {
			if a.DB.Notifications[i].ID == nid && a.DB.Notifications[i].UserID == c.User.ID {
				a.DB.Notifications[i].IsRead = true
			}
		}
	}
	a.DB.SaveUnlocked()
	a.DB.Unlock()
	next := c.R.FormValue("next")
	if next == "" {
		next = "/"
	}
	http.Redirect(c.W, c.R, next, 302)
}
