package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"volunteer-platform/backend/internal/config"
)

// Запуск сервера, парсинг конфига, плавное завершение сервера
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}
	config.ServerConfig = cfg
	address := cfg.ServerHost + ":" + itoa(cfg.ServerPort)
	srv := &http.Server{
		Addr:         address,
		Handler:      nil,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		log.Printf("Сервер активен: %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	// Реализация gracefull shutdown
	quitChan := make(chan os.Signal, 1)

	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)
	<-quitChan

	log.Println("Завершение работы сервера...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Сервер завершил работу неожиданно: %v", err)
	}
	log.Println("Сервер успешно остановлен")
}

func itoa(v int) string {
	return fmtInt(v)
}

func fmtInt(v int) string {
	return fmt.Sprintf("%d", v)
}
