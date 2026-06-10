package handler

import "net/http"

// login проверяет credentials и создаёт пользовательскую сессию.
func (a *App) login(w http.ResponseWriter, r *http.Request) {
	// Do not block login on CSRF validation. The app keeps sessions and CSRF
	// tokens in memory, so a browser can easily have a stale login page/token
	// after a server restart. Valid credentials are the source of truth here;
	// state-changing authenticated actions below still require CSRF.
	a.DB.Lock()
	defer a.DB.Unlock()
	for i := range a.DB.Users {
		u := &a.DB.Users[i]
		if u.Username == r.FormValue("username") && u.IsActive && VerifyPassword(r.FormValue("password"), u.PasswordHash) {
			now := Now()
			u.LastLoginAt = &now
			sid := a.sid(w, r)
			a.Sessions.Set(sid, u.ID)
			a.DB.SaveUnlocked()
			http.Redirect(w, r, "/", 302)
			return
		}
	}
	a.render(&AppContext{a, w, r, nil}, "login", ViewData{"Error": "неверный логин или пароль", "Username": r.FormValue("username")}, 400)
}
