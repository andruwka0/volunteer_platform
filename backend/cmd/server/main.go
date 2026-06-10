package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerHost      string        `yaml:"server_host"`
	ServerPort      string        `yaml:"server_host"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

// Запуск сервера, парсинг конфига, плавное завершение сервера
func main() {
	f, err := os.Open("config.yaml")
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		log.Fatalf("config: %v", err)
	}

	srv := &http.Server{
		Addr:         cfg.ServerHost + cfg.ServerPort,
		Handler:      nil,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		log.Printf("Сервер активен, url: %s", srv.Addr)
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
