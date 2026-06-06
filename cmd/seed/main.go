package main

import (
	"log"
	"time"

	"volunteer-platform/internal/config"
	"volunteer-platform/internal/domain"
	"volunteer-platform/internal/repository"
	"volunteer-platform/internal/service"
)

// main запускает соответствующую CLI-команду или HTTP-сервер.
func main() {
	cfg := config.Load()
	db, err := repository.OpenJSONStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	if len(db.Users) == 0 {
		leader, _ := service.CreateUser(db, "leader", "Главный организатор", "Password123", "leader", "", "@leader", "Главный", "Организатор")
		org, _ := service.CreateUser(db, "organizer", "Организатор", "Password123", "organizer", "", "@organizer", "", "")
		v, _ := service.CreateUser(db, "volunteer", "Волонтёр", "Password123", "junior_volunteer", "", "@volunteer", "", "")
		max := 30
		now := domain.Now()
		rd := now
		e := domain.Event{ID: db.NextID("events"), Title: "Демо-ивент", ShortDescription: "Первое мероприятие", FullDescription: "Описание демо-мероприятия", Location: "Москва", StartDate: now.Add(72 * time.Hour), EndDate: now.Add(96 * time.Hour), RatingDeadline: now.Add(96 * time.Hour), RegistrationDeadline: &rd, MaxParticipants: &max, ReserveParticipants: 5, SPPoints: 10, CreatedByID: leader.ID, Status: "soon", CreatedAt: now, UpdatedAt: now}
		db.Events = append(db.Events, e)
		db.Organizers = append(db.Organizers, domain.EventOrganizer{ID: db.NextID("organizers"), EventID: e.ID, UserID: org.ID, CreatedAt: now})
		db.Participants = append(db.Participants, domain.EventParticipant{ID: db.NextID("participants"), EventID: e.ID, UserID: v.ID, Status: "approved", CreatedAt: now, UpdatedAt: now})
	}
	if len(db.Rules) == 0 {
		now := domain.Now()
		db.Rules = append(db.Rules, domain.Rule{ID: db.NextID("rules"), Title: "Участвуй", Content: "Регистрируйся на мероприятия и приходи вовремя.", OrderIndex: 1, CreatedAt: now, UpdatedAt: now}, domain.Rule{ID: db.NextID("rules"), Title: "Зарабатывай SP", Content: "После завершения согласованные участники получают SP.", OrderIndex: 2, CreatedAt: now, UpdatedAt: now})
	}
	db.Save()
	log.Println("seed complete")
}
