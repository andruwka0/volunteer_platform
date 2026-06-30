package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/andruwka0/volunteer_platform/internal/auth"
	"github.com/andruwka0/volunteer_platform/internal/domain"
	"github.com/andruwka0/volunteer_platform/internal/handler"
	"github.com/andruwka0/volunteer_platform/internal/store"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token := strings.TrimSpace(authHeader)
		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimSpace(token[7:])
		}
		if token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		userID, err := auth.ValidateToken(token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), handler.CtxKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (str *statusRecorder) WriteHeader(status int) {
	str.status = status
	str.ResponseWriter.WriteHeader(status)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(store *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(handler.CtxKeyUserID).(int64)
			user, err := store.GetUserByID(userID)
			if err != nil || user.Role != domain.RoleAdmin {
				http.Error(w, "forbidden: admin required", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireOrganizer(store *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(handler.CtxKeyUserID).(int64)
			user, err := store.GetUserByID(userID)
			if err != nil {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			if user.Role != domain.RoleOrganizer && user.Role != domain.RoleAdmin {
				http.Error(w, "forbidden: organizer or admin required", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
