package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"volunteer-platform/internal/config"
	"volunteer-platform/internal/handler"
	"volunteer-platform/internal/repository"
	"volunteer-platform/internal/router"
	"volunteer-platform/internal/service"
)

// main запускает соответствующую CLI-команду или HTTP-сервер.
func main() {
	cfg := config.Load()
	store, err := repository.OpenJSONStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	services := service.New(store)
	handlers := handler.NewWithServices(store, cfg, services)

	srv := &http.Server{Addr: cfg.Addr, Handler: router.New(handlers)}
	go func() {
		log.Println("listening", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("server stopped")
}
