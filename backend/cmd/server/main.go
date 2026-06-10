package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/andruwka0/volunteer_platform/internal/config"
	"github.com/andruwka0/volunteer_platform/internal/domain"
	"github.com/andruwka0/volunteer_platform/internal/handler"
	"github.com/andruwka0/volunteer_platform/internal/router"
	"github.com/andruwka0/volunteer_platform/internal/service"
	"github.com/andruwka0/volunteer_platform/internal/store"
	"github.com/andruwka0/volunteer_platform/internal/worker"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются системные переменные окружения")
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Не удалось загрузить конфиг: %v", err)
	}
	str := store.New()

	adminLogin := os.Getenv("ADMIN_LOGIN")
	adminPass := os.Getenv("ADMIN_PASSWORD")

	if adminLogin != "" && adminPass != "" {
		_, err := str.CreateUser(adminLogin, service.HashPassword(adminPass), "Admin", "User", "")
		if err == nil {
			err := str.UpdateUserRole(1, domain.RoleAdmin)
			if err != nil {
				return
			}
			log.Println("Администратор инициализирован из ENV")
		}
	}
	svc := service.New(str)

	ctx, cancel := context.WithCancel(context.Background())
	go worker.StartEventStatusChecker(ctx, str, cfg.WorkerInterval)

	h := handler.New(svc)
	r := router.New(h, str)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Сервер запущен на %s", addr)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	<-stop
	log.Println("Получен сигнал остановки, начинаем graceful shutdown")
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		cfg.ShutdownTimeout,
	)
	defer shutdownCancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Ошибка при graceful shutdown: %v", err)
		if err := httpSrv.Close(); err != nil {
			log.Printf("Ошибка при принудительном закрытии: %v", err)
		}
	}

	log.Println("Сервер успешно остановлен")
}
