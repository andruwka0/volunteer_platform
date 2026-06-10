package router

import (
	"github.com/andruwka0/volunteer_platform/internal/handler"
	"github.com/andruwka0/volunteer_platform/internal/middleware"
	"github.com/andruwka0/volunteer_platform/internal/store"
	"net/http"
)

func New(h *handler.Handler, store *store.Store) http.Handler {
	router := http.NewServeMux()

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
	router.Handle("POST /events/{id}/finish", middleware.Auth(http.HandlerFunc(h.FinishEvent)))

	// Админские роуты (нужен Auth + RequireAdmin)
	router.Handle("POST /admin/events/{id}/approve",
		middleware.RequireAdmin(store)(middleware.Auth(http.HandlerFunc(h.ApproveEvent))))

	router.Handle("POST /admin/users/{id}/promote",
		middleware.RequireAdmin(store)(middleware.Auth(http.HandlerFunc(h.PromoteUser))))

	// Глобальные middleware
	var handlers http.Handler = router
	handlers = middleware.Recover(handlers)
	handlers = middleware.Logging(handlers)

	return handlers
}
