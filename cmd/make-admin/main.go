package main

import (
	"flag"
	"log"

	"volunteer-platform/internal/config"
	"volunteer-platform/internal/repository"
	"volunteer-platform/internal/service"
)

// main запускает соответствующую CLI-команду или HTTP-сервер.
func main() {
	username := flag.String("username", "leader", "")
	password := flag.String("password", "Password123", "")
	full := flag.String("full-name", "Leader", "")
	flag.Parse()
	cfg := config.Load()
	db, err := repository.OpenJSONStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	for i := range db.Users {
		if db.Users[i].Username == *username {
			db.Users[i].Role = "leader"
			db.Users[i].IsActive = true
			if *password != "" {
				h, _ := service.HashPassword(*password)
				db.Users[i].PasswordHash = h
			}
			db.Save()
			log.Println("updated", *username)
			return
		}
	}
	if _, err := service.CreateUser(db, *username, *full, *password, "leader", "", "", "", ""); err != nil {
		log.Fatal(err)
	}
	log.Println("created", *username)
}
