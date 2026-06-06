package main

import (
	"flag"
	"log"

	"volunteer-platform/internal/config"
	"volunteer-platform/internal/repository"
)

// main запускает соответствующую CLI-команду или HTTP-сервер.
func main() {
	username := flag.String("username", "", "")
	flag.Parse()
	if *username == "" {
		log.Fatal("username required")
	}
	cfg := config.Load()
	db, err := repository.OpenJSONStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	for i := range db.Users {
		if db.Users[i].Username == *username {
			db.Users[i].Role = "leader"
			db.Users[i].IsActive = true
			db.Save()
			log.Println("leader", *username)
			return
		}
	}
	log.Fatal("user not found")
}
