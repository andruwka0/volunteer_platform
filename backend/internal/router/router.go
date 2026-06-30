package router

import (
	"net/http"

	"github.com/andruwka0/volunteer_platform/internal/handler"
	"github.com/andruwka0/volunteer_platform/internal/middleware"
	"github.com/andruwka0/volunteer_platform/internal/store"
)

func New(h *handler.Handler, store *store.Store) http.Handler {
	router := http.NewServeMux()

	// Health check
	router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// ==================== ПУБЛИЧНЫЕ РОУТЫ ====================
	router.HandleFunc("POST /auth/register", h.Register)
	router.HandleFunc("POST /auth/login", h.Login)

	// ==================== ЗАЩИЩЁННЫЕ РОУТЫ ====================
	// Профиль
	router.Handle("GET /auth/me", middleware.Auth(http.HandlerFunc(h.GetMe)))
	router.Handle("GET /users/me/events", middleware.Auth(http.HandlerFunc(h.GetUserEvents)))
	router.Handle("GET /users/me/rewards", middleware.Auth(http.HandlerFunc(h.GetUserRewards)))
	router.Handle("POST /users/me/rewards/{id}/claim", middleware.Auth(http.HandlerFunc(h.ClaimReward)))

	// Ивенты
	router.Handle("POST /events",
		middleware.Auth(middleware.RequireOrganizer(store)(http.HandlerFunc(h.CreateEvent))))
	router.Handle("GET /events", middleware.Auth(http.HandlerFunc(h.GetAllEvents)))
	router.Handle("GET /events/{id}", middleware.Auth(http.HandlerFunc(h.GetEventByID)))
	router.Handle("GET /events/{id}/participants", middleware.Auth(http.HandlerFunc(h.GetEventParticipants)))
	router.Handle("POST /events/{id}/register", middleware.Auth(http.HandlerFunc(h.RegisterForEvent)))
	router.Handle("DELETE /events/{id}/register", middleware.Auth(http.HandlerFunc(h.CancelRegistration)))

	// История транзакций SP
	router.Handle("GET /users/{id}/skill-points/history",
		middleware.Auth(http.HandlerFunc(h.GetUserSkillPointHistory)))

	// ==================== ОРГСКИЕ РОУТЫ ====================
	router.Handle("POST /organizer/events/{event_id}/users/{user_id}/confirm",
		middleware.Auth(middleware.RequireOrganizer(store)(http.HandlerFunc(h.ConfirmAttendance))))
	router.Handle("POST /organizer/blacklist",
		middleware.Auth(middleware.RequireOrganizer(store)(http.HandlerFunc(h.BanUser))))
	router.Handle("DELETE /organizer/blacklist/{user_id}",
		middleware.Auth(middleware.RequireOrganizer(store)(http.HandlerFunc(h.UnbanUser))))
	// Шаблоны
	router.Handle("POST /organizer/event-templates",
		middleware.Auth(middleware.RequireOrganizer(store)(http.HandlerFunc(h.CreateEventTemplate))))
	router.Handle("GET /organizer/event-templates",
		middleware.Auth(middleware.RequireOrganizer(store)(http.HandlerFunc(h.GetAllEventTemplates))))

	// ==================== АДМИНСКИЕ РОУТЫ ====================
	// Пользователи
	router.Handle("GET /admin/users",
		middleware.Auth(middleware.RequireAdmin(store)(http.HandlerFunc(h.SearchUsers))))
	router.Handle("POST /admin/users/{id}/promote",
		middleware.Auth(middleware.RequireAdmin(store)(http.HandlerFunc(h.PromoteUser))))
	router.Handle("POST /admin/users/{id}/award",
		middleware.Auth(middleware.RequireAdmin(store)(http.HandlerFunc(h.AwardUser))))

	// НОВОЕ: Награды конкретного юзера (для выдачи мерча)
	router.Handle("GET /admin/users/{id}/rewards",
		middleware.Auth(middleware.RequireAdmin(store)(http.HandlerFunc(h.GetUserRewardsByAdmin))))

	// Ивенты
	router.Handle("POST /admin/events/{id}/approve",
		middleware.Auth(middleware.RequireAdmin(store)(http.HandlerFunc(h.ApproveEvent))))

	// Награды (глобальные)
	router.Handle("POST /admin/rewards",
		middleware.Auth(middleware.RequireAdmin(store)(http.HandlerFunc(h.CreateReward))))
	router.Handle("GET /admin/rewards",
		middleware.Auth(middleware.RequireAdmin(store)(http.HandlerFunc(h.GetAllRewards))))
	router.Handle("POST /admin/users/{user_id}/rewards/{reward_id}/pickup",
		middleware.Auth(middleware.RequireAdmin(store)(http.HandlerFunc(h.ConfirmRewardPickup))))

	// Картинки
	router.Handle("GET /assets/images", middleware.Auth(http.HandlerFunc(h.GetAvailableImages)))

	// ==================== ГЛОБАЛЬНЫЕ MIDDLEWARE ====================
	var handlers http.Handler = router
	handlers = middleware.Recover(handlers)
	handlers = middleware.Logging(handlers)

	return handlers
}
