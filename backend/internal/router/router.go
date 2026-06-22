package router

import (
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/andruwka0/volunteer_platform/internal/handler"
	"github.com/andruwka0/volunteer_platform/internal/middleware"
	"github.com/andruwka0/volunteer_platform/internal/store"
)

func getStaticDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "app/static"
	}
	dir := filepath.Dir(filename)
	projectRoot := filepath.Join(dir, "..", "..", "..")
	return filepath.Join(projectRoot, "app", "static")
}

func New(h *handler.Handler, store *store.Store) http.Handler {
	router := http.NewServeMux()

	staticDir := getStaticDir()
	router.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Редирект с корня на главную страницу
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/index.html", http.StatusMovedPermanently)
	})

	// Публичные роуты
	router.HandleFunc("POST /auth/register", h.Register)
	router.HandleFunc("POST /auth/login", h.Login)

	// GET ивенты (публичные)
	router.HandleFunc("GET /events", h.GetAllEvents)
	router.HandleFunc("GET /events/{id}", h.GetEventByID)
	router.HandleFunc("GET /events/{id}/participants", h.GetEventParticipants)

	// Защищённые роуты (нужен Auth)
	router.Handle("POST /events", middleware.Auth(http.HandlerFunc(h.CreateEvent)))
	router.Handle("POST /events/{id}/register", middleware.Auth(http.HandlerFunc(h.RegisterForEvent)))
	router.Handle("DELETE /events/{id}/register", middleware.Auth(http.HandlerFunc(h.CancelRegistration)))
	router.Handle("GET /auth/me", middleware.Auth(http.HandlerFunc(h.GetMe)))

	// Админские роуты (нужен Auth + RequireAdmin)
	router.Handle("POST /admin/events/{id}/approve",
		middleware.Auth(middleware.RequireAdmin(store)(http.HandlerFunc(h.ApproveEvent))))

	router.Handle("POST /admin/users/{id}/promote",
		middleware.Auth(middleware.RequireAdmin(store)(http.HandlerFunc(h.PromoteUser))))

	// Глобальные middleware
	var handlers http.Handler = router
	handlers = middleware.Recover(handlers)
	handlers = middleware.Logging(handlers)

	return handlers
}
